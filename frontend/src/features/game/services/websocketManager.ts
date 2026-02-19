import { WS_URL, API_BASE_URL, BOT_MOVE_DELAY } from "@/lib/config";
import { useGameStore } from "../store/gameStore";
import type { Board, ServerMessage, ClientMessage } from "../types";
import { toast } from "sonner";

type MessageHandler = (message: ServerMessage) => void;

class WebSocketManager {
  private socket: WebSocket | null = null;
  private messageHandlers: Set<MessageHandler> = new Set();
  private connectionPromise: Promise<void> | null = null;
  private currentConnectionId: number = 0;
  private onGameStartCallbacks: Set<(gameId: string) => void> = new Set();
  private disconnectToastId: string | number | undefined;

  public onMessage(handler: MessageHandler) {
    this.messageHandlers.add(handler);
    return () => this.messageHandlers.delete(handler);
  }

  public onGameStart(callback: (gameId: string) => void) {
    this.onGameStartCallbacks.add(callback);
    return () => {
      this.onGameStartCallbacks.delete(callback);
    };
  }

  private async createConnection(myConnectionId: number): Promise<void> {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/me`, {
        credentials: "include",
      });

      // CHECK 1: If we became stale while fetching, stop immediately.
      if (myConnectionId !== this.currentConnectionId) {
        return;
      }

      if (!response.ok) {
        throw new Error(`Token fetch failed: ${response.status}`);
      }

      const data = await response.json();
      if (!data.token) {
        throw new Error("No token received");
      }

      return new Promise<void>((resolve, reject) => {
        const ws = new WebSocket(WS_URL);

        ws.onopen = () => {
          // CHECK 2: If we became stale while opening, close and exit.
          if (myConnectionId !== this.currentConnectionId) {
            ws.close();
            resolve();
            return;
          }

          ws.send(JSON.stringify({ type: "init", jwt: data.token }));

          this.socket = ws;
          this.reconnectAttempts = 0; // Reset on success
          this.isReconnecting = false;
          this.setupListeners(ws);
          resolve();
        };

        ws.onerror = (error) => {
          if (myConnectionId === this.currentConnectionId) {
            console.error("[WebSocket] Connection failed:", error);
            this.socket = null;
            this.autoReconnect(myConnectionId); 
            resolve(); 
          }
        };

        ws.onclose = () => {
          if (myConnectionId === this.currentConnectionId) {
            this.socket = null;
            this.connectionPromise = null;
            this.autoReconnect(myConnectionId);
          }
        };
      });
    } catch (error) {
      if (myConnectionId === this.currentConnectionId) {
        this.connectionPromise = null;
        this.autoReconnect(myConnectionId);
      }
    }
  }

  private resetConnectionState() {
    this.currentConnectionId++;
    this.reconnectAttempts = 0;
    this.isReconnecting = false;
    this.connectionPromise = null;
    if (this.socket) {
      this.socket.onclose = null; // Prevent triggering autoReconnect from this close
      this.socket.close(1000, "Client disconnect");
      this.socket = null;
    }
  }

  private subscribers: number = 0;

  public async connect(): Promise<void> {
    this.subscribers++;
    
    if (this.socket?.readyState === WebSocket.OPEN) {
      return;
    }

    // Use existing promise if connecting
    if (this.connectionPromise) {
      return this.connectionPromise;
    }

    // Only create new connection if we are the first subscriber or recovering
    this.currentConnectionId++; // New connection attempt ID
    this.connectionPromise = this.createConnection(this.currentConnectionId);
    return this.connectionPromise;
  }

  private isReconnecting: boolean = false;
  private reconnectAttempts: number = 0;

  public disconnect() {
    this.subscribers--;
    if (this.subscribers <= 0) {
      this.subscribers = 0; // Safety clamp
      this.resetConnectionState();
      useGameStore.getState().setConnectionStatus('disconnected');
    }
  }

  private async autoReconnect(myConnectionId: number) {
    // If we have no subscribers, don't reconnect
    if (this.subscribers === 0) return;
    if (myConnectionId !== this.currentConnectionId) return;

    // Exponential backoff: 1s, 2s, 4s, 8s, max 30s
    // Start faster (1s) to recover quick blips
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
    this.reconnectAttempts++;
    this.isReconnecting = true;
    
    useGameStore.getState().setConnectionStatus('error', `Reconnecting... (Attempt ${this.reconnectAttempts})`);

    setTimeout(async () => {
      // Check subscribers again before attempting
      if (this.subscribers === 0) return;
      if (myConnectionId !== this.currentConnectionId) return;

      try {
        await this.internalReconnect();
        // If successful, the onopen handler will reset reconnectAttempts
        if (this.socket?.readyState === WebSocket.OPEN) {
             useGameStore.getState().setConnectionStatus('connected');
             // Re-fetch game state on reconnection
             this.send({ type: 'get_game_state' });
        }
      } catch (err) {
        console.error('[WebSocket] Reconnect failed:', err);
        this.autoReconnect(this.currentConnectionId);
      }
    }, delay);
  }

  private async internalReconnect(): Promise<void> {
    this.connectionPromise = this.createConnection(this.currentConnectionId);
    return this.connectionPromise;
  }

  public send(message: ClientMessage | Record<string, unknown>) {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(message));
    } else {
      console.warn("[WebSocket] Cannot send, socket not open", message);
    }
  }

  private setupListeners(ws: WebSocket) {
    ws.onmessage = (event) => {
      try {
        const message: ServerMessage = JSON.parse(event.data);

        // Handle message internally first
        this.handleMessage(message);

        // Then notify all registered handlers
        this.messageHandlers.forEach((handler) => handler(message));
      } catch (error) {
        console.error("[WebSocket] Parse error:", error);
      }
    };
  }

  private handleMessage(message: ServerMessage) {
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

        toast.success(`Game started against ${message.opponent}!`);

        // Notify all game start callbacks
        this.onGameStartCallbacks.forEach((callback) => {
          if (message.gameId) callback(message.gameId);
        });
        break;

      case "spectate_start":
        store.initSpectatorGame({
          gameId: message.gameId,
          board: message.board as Board,
          currentTurn: message.currentTurn,
          player1: message.player1,
          player2: message.player2,
        });

        toast.success(`Now spectating: ${message.player1} vs ${message.player2}`);

        // Notify game start callbacks so we navigate to the game page
        this.onGameStartCallbacks.forEach((callback) => {
          if (message.gameId) callback(message.gameId);
        });
        break;

      case "move_made": {
        const isBotGame = store.gameMode === "bot";
        const wasOpponentMove =
          message.lastMove && message.lastMove.player !== store.myPlayer;

        const currentTurn = message.nextTurn;

        if (isBotGame && wasOpponentMove) {
          setTimeout(() => {
            store.updateGameState({
              board: message.board as Board,
              currentTurn: currentTurn,
              lastMove: message.lastMove,
            });
          }, BOT_MOVE_DELAY);
        } else {
          store.updateGameState({
            board: message.board as Board,
            currentTurn: currentTurn,
            lastMove: message.lastMove,
          });
        }
        break;
      }

      case "game_state": {
        // Reconnect scenario: game_state includes gameId + yourPlayer
        if (message.gameId && message.yourPlayer) {
          store.initGame({
            gameId: message.gameId,
            board: message.board as Board,
            currentTurn: message.currentTurn,
            myPlayer: message.yourPlayer,
            opponent: message.opponent || store.opponent || "Opponent",
          });
        } else {
          // Simple state refresh (no full reinit needed)
          store.updateGameState({
            board: message.board as Board,
            currentTurn: message.currentTurn,
            lastMove: message.lastMove,
          });
        }
        break;
      }

      case "game_over":
        store.endGame({
          winner: message.winner,
          reason: message.reason,
          winningCells: message.winningCells,
          board: message.board as Board,
          allowRematch: message.allowRematch,
        });
        break;

      case "rematch_request":
        store.setRematchStatus("received");
        break;

      case "rematch_accepted":
        store.setRematchStatus("accepted");
        toast.success("Rematch accepted! Starting new game...");
        break;

      case "rematch_declined":
        store.setRematchStatus("declined");
        store.setAllowRematch(false);
        toast.info("Rematch declined");
        break;

      case "rematch_timeout":
        store.setRematchStatus("declined");
        store.setAllowRematch(false);
        toast.info("Rematch request timed out");
        break;

      case "rematch_cancelled":
        store.setRematchStatus("idle");
        toast.info("Rematch request cancelled");
        break;

      case "opponent_disconnected":
        store.setOpponentDisconnected(true, message.disconnectTimeout || 60);
        if (this.disconnectToastId) toast.dismiss(this.disconnectToastId);
        
        this.disconnectToastId = toast.warning(message.message || "Opponent disconnected", {
          duration: (message.disconnectTimeout || 60) * 1000,
          description: "Game will end if they don't reconnect.",
        });
        break;

      case "opponent_reconnected":
        store.setOpponentDisconnected(false);
        if (this.disconnectToastId) {
          toast.dismiss(this.disconnectToastId);
          this.disconnectToastId = undefined;
        }
        toast.success("Opponent reconnected!");
        break;

      case "error":
        toast.error(message.message || "An error occurred");
        break;

      default:
        break;
    }
  }
}

export const websocketManager = new WebSocketManager();
