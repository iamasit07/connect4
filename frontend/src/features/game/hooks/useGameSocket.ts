import { useEffect, useCallback, useState, useRef } from 'react';
import useWebSocket, { ReadyState } from 'react-use-websocket';
import { useGameStore } from '../store/gameStore';
import { useAuthStore } from '@/features/auth/store/authStore';
import type { BotDifficulty, ServerMessage, Board, GameStateMessage } from '../types';
import { WS_URL, API_BASE_URL, BOT_MOVE_DELAY } from "@/lib/config";
import { toast } from 'sonner';

export const useGameSocket = (
  onGameStart?: (gameId: string) => void,
  onQueueTimeout?: () => void
) => {
  const shouldConnect = useGameStore(state => state.shouldConnect);
  const [token, setToken] = useState<string | null>(null);
  const onGameStartRef = useRef<((gameId: string) => void) | undefined>(onGameStart);
  const onQueueTimeoutRef = useRef<(() => void) | undefined>(onQueueTimeout);
  const disconnectToastIdRef = useRef<string | number | undefined>(undefined);
  const pendingBotMoveTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pendingGameOverRef = useRef<ServerMessage | null>(null);
  const messageQueueRef = useRef<string[]>([]);

  useEffect(() => {
    onGameStartRef.current = onGameStart;
  }, [onGameStart]);

  useEffect(() => {
    onQueueTimeoutRef.current = onQueueTimeout;
  }, [onQueueTimeout]);

  const fetchToken = useCallback(async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/me`, {
        credentials: "include",
      });
      if (response.ok) {
        const data = await response.json();
        if (data.token) {
          setToken(data.token);
          return data.token;
        }
      }
    } catch (e) {
      console.error("Failed to fetch token", e);
    }
    return null;
  }, []);

  const { sendMessage: sendWsMessage, lastMessage, readyState, getWebSocket } = useWebSocket(
    shouldConnect && token ? WS_URL : null,
    {
      share: true,
      shouldReconnect: (closeEvent) => {
        return closeEvent.code !== 1000;
      },
      reconnectAttempts: 10,
      reconnectInterval: (attemptNumber) => Math.min(Math.pow(2, attemptNumber) * 1000, 30000),
      onOpen: () => {
        useGameStore.getState().setConnectionStatus('connected');
        if (token) {
           sendWsMessage(JSON.stringify({ type: "init", jwt: token }));
           const store = useGameStore.getState();
           if (store.isSpectator && store.gameId) {
              setTimeout(() => {
                 sendWsMessage(JSON.stringify({ type: "watch_game", gameId: store.gameId }));
              }, 100);
           }

           setTimeout(() => {
             if (messageQueueRef.current.length > 0) {
               messageQueueRef.current.forEach(msg => sendWsMessage(msg));
               messageQueueRef.current = [];
             }
           }, 200);
        }
      },
      onClose: () => {
        useGameStore.getState().setConnectionStatus('disconnected');
      },
      onError: (event) => {
        console.error("WebSocket error:", event);
      },
      onMessage: (event) => {
        try {
          const message: ServerMessage = JSON.parse(event.data);
          
          if (message.type === 'queue_timeout' && onQueueTimeoutRef.current) {
            onQueueTimeoutRef.current();
          }

          handleMessage(message);
        } catch (e) {
          console.error("Failed to parse message", e);
        }
      }
    }
  );

  const processGameOver = useCallback((message: ServerMessage) => {
    useAuthStore.getState().clearActiveGameId();
    useGameStore.getState().endGame({
      winner: (message as any).winner,
      reason: (message as any).reason,
      winningCells: (message as any).winningCells,
      board: (message as any).board as Board,
      allowRematch: (message as any).allowRematch,
    });
  }, []);

  const handleMessage = useCallback((message: ServerMessage) => {
    const store = useGameStore.getState();
    switch (message.type) {
      case "queue_joined":
        toast.info("Searching for opponent...");
        break;

      case "game_start":
        store.initGame({
          gameId: message.gameId,
          board: message.board as Board,
          currentTurn: message.currentTurn,
          myPlayer: message.yourPlayer,
          opponent: message.opponent,
        });
        useAuthStore.getState().setActiveGameId(message.gameId);

        if (!window.location.pathname.startsWith('/game/')) {
          toast.success(`Game started against ${message.opponent}!`);
        }
        if (onGameStartRef.current && message.gameId) {
          onGameStartRef.current(message.gameId);
        }
        break;

      case "spectate_start":
        store.initSpectatorGame({
          gameId: message.gameId,
          board: message.board as Board,
          currentTurn: message.currentTurn,
          player1: message.player1,
          player2: message.player2,
        });

        if (!window.location.pathname.startsWith('/game/')) {
          toast.success(`Now spectating: ${message.player1} vs ${message.player2}`);
        }

        if (onGameStartRef.current && message.gameId) {
          onGameStartRef.current(message.gameId);
        }
        break;

      case "move_made": {
        const isBotGame = store.gameMode === "bot";
        const parsedLastMove = {
           column: message.column,
           row: message.row,
           player: message.player
        };

        const wasOpponentMove = parsedLastMove.player !== store.myPlayer;

        const currentTurn = message.nextTurn;

        if (isBotGame && wasOpponentMove) {
          pendingBotMoveTimeoutRef.current = setTimeout(() => {
            pendingBotMoveTimeoutRef.current = null;
            store.updateGameState({
              board: message.board as Board,
              currentTurn: currentTurn,
              lastMove: parsedLastMove,
            });
            if (pendingGameOverRef.current) {
              const queuedMsg = pendingGameOverRef.current;
              pendingGameOverRef.current = null;
              processGameOver(queuedMsg);
            }
          }, BOT_MOVE_DELAY);
        } else {
          store.updateGameState({
            board: message.board as Board,
            currentTurn: currentTurn,
            lastMove: parsedLastMove,
          });
        }
        break;
      }

      case "game_state": {
        if (store.isSpectator) {
          store.updateGameState({
            board: message.board as Board,
            currentTurn: message.currentTurn,
            lastMove: undefined,
          });
          
          if (message.winner) {
            store.endGame({
                winner: message.winner,
                reason: message.reason || "unknown",
                winningCells: message.winningCells,
                allowRematch: false,
            });
          }
        } 
        else if (message.gameId && message.yourPlayer) {
          store.initGame({
            gameId: message.gameId,
            board: message.board as Board,
            currentTurn: message.currentTurn,
            myPlayer: message.yourPlayer,
            opponent: message.opponent || store.opponent || "Opponent",
            disconnectTimeout: message.disconnectTimeout,
          });

          if (message.winner) {
            useAuthStore.getState().clearActiveGameId();
            store.endGame({
                winner: message.winner,
                reason: message.reason || "unknown",
                winningCells: message.winningCells,
                allowRematch: message.allowRematch,
            });
          }
          
          if (message.rematchRequester) {
              const currentUser = useAuthStore.getState().user;
              if (message.rematchRequester === currentUser?.username) {
                  store.setRematchStatus('sent');
              } else {
                  store.setRematchStatus('received');
              }
          }
        } else {
          store.updateGameState({
            board: message.board as Board,
            currentTurn: message.currentTurn,
            lastMove: undefined,
          });
          
          if (message.winner) {
            useAuthStore.getState().clearActiveGameId();
            store.endGame({
                winner: message.winner,
                reason: message.reason || "unknown",
                winningCells: message.winningCells,
                allowRematch: message.allowRematch,
            });
          }

          if (message.disconnectTimeout !== undefined) {
             store.setOpponentDisconnected(message.disconnectTimeout > 0, message.disconnectTimeout);
          }
        }
        break;
      }

      case "game_over":
        if (pendingBotMoveTimeoutRef.current) {
          pendingGameOverRef.current = message;
        } else {
          processGameOver(message);
        }
        break;

      case "no_active_game":
        useAuthStore.getState().clearActiveGameId();
        if (store.gameStatus === 'playing') {
          store.resetGame();
        }
        break;

      case "rematch_requested":
        if (store.isSpectator) {
           toast.info(message.message || "A rematch was requested.");
        } else if (store.gameStatus === "finished") {
           const currentUser = useAuthStore.getState().user;
           if (message.rematchRequester === currentUser?.username) {
             store.setRematchStatus("sent");
           } else {
             store.setRematchStatus("received");
           }
        }
        break;

      case "rematch_accepted":
        if (store.isSpectator) {
           toast.info(message.message || "Rematch accepted!");
        } else if (store.gameStatus === "finished") {
          store.setRematchStatus("accepted");
        }
        break;

      case "rematch_declined":
        if (store.isSpectator) {
           toast.info(message.message || "Rematch declined.");
        } else if (store.gameStatus === "finished") {
          store.setRematchStatus("declined");
          store.setAllowRematch(false);
        }
        break;

      case "rematch_timeout":
        if (store.isSpectator) {
           toast.info(message.message || "Rematch request timed out.");
        } else if (store.gameStatus === "finished") {
          store.setRematchStatus("declined");
          store.setAllowRematch(false);
        }
        break;

      case "rematch_cancelled":
        if (store.isSpectator) {
           toast.info(message.message || "Rematch request cancelled.");
        } else if (store.gameStatus === "finished") {
          store.setRematchStatus("idle");
        }
        break;

      case "opponent_disconnected":
        store.setOpponentDisconnected(true, message.disconnectTimeout || 60);
        break;

      case "opponent_reconnected":
        store.setOpponentDisconnected(false);
        if (disconnectToastIdRef.current) {
          toast.dismiss(disconnectToastIdRef.current);
          disconnectToastIdRef.current = undefined;
        }
        break;

      case "force_disconnect":
        setToken(null);
        store.setConnectionStatus("disconnected");
        toast.error(message.message || "Disconnected by the server");
        store.resetGame();
        useAuthStore.getState().clearActiveGameId();
        break;

      case "error":
        if (message.message?.toLowerCase().includes("session") || message.message?.toLowerCase().includes("token") || message.message?.toLowerCase().includes("invalidated")) {
             setToken(null);
             store.setConnectionStatus('error', message.message);
        }
        if (message.message === "not your turn") {
          toast.error("Wait for your turn");
        } else if (!window.location.pathname.startsWith('/game/')) {
          toast.error(message.message || "An error occurred");
        }
        break;

      default:
        break;
    }
  }, [processGameOver]);

  const send = useCallback((message: Record<string, unknown>) => {
     const msgStr = JSON.stringify(message);
     if (readyState === ReadyState.OPEN) {
       sendWsMessage(msgStr);
     } else {
       messageQueueRef.current.push(msgStr);
     }
  }, [readyState, sendWsMessage]);

  const connect = useCallback(async () => {
    let currentToken = token;
    if (!currentToken) {
      currentToken = await fetchToken();
    }
    if (currentToken) {
      useGameStore.getState().setShouldConnect(true);
    } else {
      console.error("Failed to connect: Could not obtain token");
    }
  }, [token, fetchToken]);

  const disconnect = useCallback(() => {
    useGameStore.getState().setShouldConnect(false);
    const ws = getWebSocket();
    if (ws && ws.readyState === WebSocket.OPEN) {
       ws.close(1000, "Client disconnect");
    }
  }, [getWebSocket]);

  const findMatch = useCallback(async (mode: 'pvp' | 'bot', difficulty?: BotDifficulty) => {
    await connect();

    useGameStore.getState().setQueuing(mode, difficulty);
    send({
      type: 'find_match',
      difficulty: mode === 'bot' ? (difficulty || 'easy') : '',
    });
  }, [connect, send]);

  const makeMove = useCallback((column: number) => {
    send({ type: 'make_move', column });
  }, [send]);

  const surrender = useCallback(() => {
    send({ type: 'abandon_game' });
  }, [send]);

  const sendMessage = useCallback((message: Record<string, unknown>) => {
    send(message);
  }, [send]);

  const spectateGame = useCallback(async (gameId: string) => {
    await connect();
    send({ type: 'watch_game', gameId });
  }, [connect, send]);

  const leaveSpectate = useCallback((gameId: string) => {
    send({ type: 'leave_spectate', gameId });
  }, [send]);

  const requestGameState = useCallback((gameId: string) => {
    send({ type: 'get_game_state', gameId });
  }, [send]);

  // Clean up on unmount
  useEffect(() => {
    return () => {
      if (pendingBotMoveTimeoutRef.current) {
        clearTimeout(pendingBotMoveTimeoutRef.current);
      }
    }
  }, []);

  return {
    connect,
    findMatch,
    makeMove,
    surrender,
    disconnect,
    sendMessage,
    spectateGame,
    leaveSpectate,
    requestGameState,
    readyState
  };
};
