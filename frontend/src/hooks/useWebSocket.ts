import { useEffect, useRef, useState } from "react";
import {
  ClientMessage,
  GameState,
  PlayerID,
  ServerMessage,
} from "../types/game";

interface UseWebSocketReturn {
  connected: boolean;
  gameState: GameState;
  joinQueue: (username: string) => void;
  makeMove: (column: number) => void;
  reconnect: (username?: string, gameID?: string) => void;
}

const useWebSocket = (): UseWebSocketReturn => {
  const [connected, setConnected] = useState(false);
  const [gameState, setGameState] = useState<GameState>({
    gameId: null,
    yourPlayer: null,
    opponent: null,
    currentTurn: null,
    board: Array(6)
      .fill(null)
      .map(() => Array(7).fill(0)) as PlayerID[][],
    gameOver: false,
    winner: null,
    reason: null,
    inQueue: false,
    queuedAt: null,
    opponentDisconnected: false,
    disconnectedAt: null,
    matchEnded: false,
    matchEndedAt: null,
    error: null,
  });

  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    const wsUrl = import.meta.env.VITE_WS_URL || "ws://localhost:8080";
    ws.current = new WebSocket(`${wsUrl}/ws`);

    ws.current.onopen = () => {
      console.log("WebSocket connected");
      setConnected(true);
    };

    ws.current.onmessage = (event: MessageEvent) => {
      const message: ServerMessage = JSON.parse(event.data);
      console.log("Received message:", message);

      switch (message.type) {
        case "queue_joined":
          // Save userToken if provided by backend
          if (message.userToken) {
            localStorage.setItem("userToken", message.userToken);
            console.log("Saved userToken:", message.userToken);
          }
          setGameState((prevState: GameState) => ({
            ...prevState,
            inQueue: true,
            queuedAt: Date.now(),
          }));
          break;
        case "game_start":
          // Save gameID to localStorage for automatic reconnection
          if (message.gameId) {
            localStorage.setItem("gameID", message.gameId);
          }
          setGameState((prevState: GameState) => ({
            ...prevState,
            gameId: message.gameId ?? null,
            yourPlayer: (message.yourPlayer as PlayerID) ?? null,
            opponent: message.opponent ?? null,
            currentTurn: (message.currentTurn as PlayerID) ?? null,
            board: message.board ?? prevState.board,
            gameOver: false,
            inQueue: false,
          }));
          break;
        case "reconnect_success":
          if (message.gameId) {
            localStorage.setItem("gameID", message.gameId);
          }
          setGameState((prevState: GameState) => ({
            ...prevState,
            gameId: message.gameId ?? null,
            yourPlayer: (message.yourPlayer as PlayerID) ?? null,
            opponent: message.opponent ?? null,
            currentTurn: (message.currentTurn as PlayerID) ?? null,
            board: message.board ?? prevState.board,
            gameOver: false,
            inQueue: false,
            matchEnded: false,
            matchEndedAt: null,
          }));
          break;
        case "move_made":
          setGameState((prevState: GameState) => ({
            ...prevState,
            board: message.board ?? prevState.board,
            currentTurn: (message.nextTurn as PlayerID) ?? null,
          }));
          break;
        case "game_over":
          // Clear gameID from localStorage when game ends
          localStorage.removeItem("gameID");
          localStorage.removeItem("isReconnecting");

          setGameState((prevState: GameState) => ({
            ...prevState,
            gameOver: true,
            winner: message.winner ?? null,
            reason: message.reason ?? null,
            board: message.board ?? prevState.board,
            opponentDisconnected: false,
            disconnectedAt: null,
          }));
          
          break;
        case "opponent_disconnected":
          setGameState((prevState: GameState) => ({
            ...prevState,
            opponentDisconnected: true,
            disconnectedAt: Date.now(),
          }));
          break;
        case "opponent_reconnected":
          setGameState((prevState: GameState) => ({
            ...prevState,
            opponentDisconnected: false,
            disconnectedAt: null,
          }));
          break;
        case "token_corrupted":
          // Token corruption detected - clear everything and redirect
          console.error("Token corruption detected, clearing session and redirecting");
          localStorage.removeItem("gameID");
          localStorage.removeItem("username");
          localStorage.removeItem("userToken");
          localStorage.removeItem("isReconnecting");
          
          setGameState((prevState: GameState) => ({
            ...prevState,
            reason: message.message || "Your authentication token was corrupted or modified. Please rejoin with correct credentials.",
            matchEnded: true,
            matchEndedAt: Date.now(),
            gameOver: true,
            inQueue: false,
          }));
          
          // Redirect to home after 10 seconds (handled by MatchEndedNotification)
          break;
        case "error":
        case "invalid_username":
        case "token_taken":
        case "queue_error":
        case "invalid_move":
        case "not_your_turn":
        case "not_in_game":
        case "game_full":
        case "already_connected":
        case "not_disconnected":
          // Handle minor error messages (temporary flash)
          console.error(`Server error [${message.type}]:`, message.message);
          setGameState((prevState: GameState) => ({
            ...prevState,
            error: message.message || `Error: ${message.type}`,
          }));
          // Clear error after 10 seconds
          setTimeout(() => {
            setGameState((prevState: GameState) => ({
              ...prevState,
              error: null,
            }));
          }, 10000);
          break;
        case "invalid_reconnect":
        case "invalid_token":
        case "username_mismatch":
        case "database_error":
        case "no_active_game":
        case "game_finished":
        case "game_not_found":
          // Fatal errors - show ErrorNotification with redirect
          console.log("Fatal error, showing ErrorNotification:", message.type);
          
          localStorage.removeItem("gameID");
          localStorage.removeItem("username");
          localStorage.removeItem("isReconnecting");
          
          setGameState((prevState: GameState) => ({
            ...prevState,
            matchEnded: true,
            matchEndedAt: Date.now(),
            gameOver: true,
            inQueue: false,
            reason: message.message || `Error: ${message.type}`,
          }));
          break;
        default:
          console.warn("Unhandled message type:", message.type);
      }
    };

    ws.current.onerror = (error: Event) => {
      console.error("WebSocket error:", error);
      setConnected(false);
    };

    ws.current.onclose = () => {
      console.log("WebSocket disconnected");
      setConnected(false);
    };

    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, []);

  const sendMessage = (message: ClientMessage) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      console.log("Sending message:", message);
      ws.current.send(JSON.stringify(message));
    } else {
      console.error(
        "Cannot send message - WebSocket not open. ReadyState:",
        ws.current?.readyState
      );
    }
  };

  const joinQueue = (username: string) => {
    console.log("joinQueue called with username:", username);
    // Get existing userToken from localStorage (if any)
    const userToken = localStorage.getItem("userToken") || undefined;
    sendMessage({ type: "join_queue", username, userToken });
  };

  const makeMove = (column: number) => {
    if (gameState.gameId) {
      const userToken = localStorage.getItem("userToken") || undefined;
      sendMessage({ type: "move", column, userToken });
    }
  };

  const reconnect = (username?: string, gameID?: string) => {
    const userToken = localStorage.getItem("userToken") || "";
    
    // UserToken is always required
    if (!userToken) {
      console.error("Reconnect failed: userToken not found in localStorage");
      // Show ErrorNotification instead of silently failing
      setGameState((prevState: GameState) => ({
        ...prevState,
        matchEnded: true,
        matchEndedAt: Date.now(),
        gameOver: true,
        inQueue: false,
        reason: "No authentication token found. Please start a new game from the home page.",
      }));
      return;
    }

    // At least username OR gameID must be provided
    if (!username && !gameID) {
      console.error("Reconnect failed: either username or gameID is required");
      // Show ErrorNotification for invalid reconnect attempt
      setGameState((prevState: GameState) => ({
        ...prevState,
        matchEnded: true,
        matchEndedAt: Date.now(),
        gameOver: true,
        inQueue: false,
        reason: "Invalid reconnection attempt. Either username or game ID is required.",
      }));
      return;
    }

    console.log("Reconnecting with:", { 
      username: username || "(not provided)", 
      gameID: gameID || "(not provided)", 
      hasToken: true 
    });
    
    sendMessage({
      type: "reconnect",
      username: username || "",
      gameID: gameID || "",
      userToken: userToken,
    });
  };

  return {
    connected,
    gameState,
    joinQueue,
    makeMove,
    reconnect,
  };
};

export default useWebSocket;
