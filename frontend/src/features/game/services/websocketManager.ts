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

  public async connect(): Promise<void> {
    const myConnectionId = ++this.currentConnectionId;

    if (this.socket?.readyState === WebSocket.OPEN) {
      console.log("[WebSocket] Already connected");
      return;
    }

    this.connectionPromise = this.createConnection(myConnectionId);
    return this.connectionPromise;
  }

  private async createConnection(myConnectionId: number): Promise<void> {
    try {
      console.log(
        `[WebSocket] Connection attempt #${myConnectionId} starting...`,
      );

      const response = await fetch(`${API_BASE_URL}/auth/me`, {
        credentials: "include",
      });

      // CHECK 1: If we became stale while fetching, stop immediately.
      if (myConnectionId !== this.currentConnectionId) {
        console.log(
          `[WebSocket] Stale attempt #${myConnectionId} aborted before socket creation.`,
        );
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
            console.log(
              `[WebSocket] Stale attempt #${myConnectionId} aborted after open.`,
            );
            ws.close();
            resolve();
            return;
          }

          console.log(
            `[WebSocket] Sending init for attempt #${myConnectionId}`,
          );
          ws.send(JSON.stringify({ type: "init", jwt: data.token }));

          this.socket = ws;
          this.setupListeners(ws);
          resolve();
        };

        ws.onerror = (error) => {
          if (myConnectionId === this.currentConnectionId) {
            console.error("[WebSocket] Connection failed:", error);
            this.socket = null;
            this.connectionPromise = null;
            reject(error);
          }
        };

        ws.onclose = (event) => {
          if (myConnectionId === this.currentConnectionId) {
            console.log("[WebSocket] Disconnected:", event.code);
            this.socket = null;
            this.connectionPromise = null;
          }
        };
      });
    } catch (error) {
      if (myConnectionId === this.currentConnectionId) {
        this.connectionPromise = null;
      }
      throw error;
    }
  }

  public disconnect() {
    this.currentConnectionId++;
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this.connectionPromise = null;
  }

  public send(message: ClientMessage | Record<string, unknown>) {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(message));
    } else {
      console.warn("[WebSocket] Cannot send, socket not open");
    }
  }

  private setupListeners(ws: WebSocket) {
    ws.onmessage = (event) => {
      try {
        const message: ServerMessage = JSON.parse(event.data);
        console.log("[WebSocket] Received message:", message.type, message);

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

      case "move_made":
      case "game_state": {
        const isBotGame = store.gameMode === "bot";
        const wasOpponentMove =
          message.lastMove && message.lastMove.player !== store.myPlayer;

        const currentTurn =
          "nextTurn" in message ? message.nextTurn : message.currentTurn;

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

      case "error":
        toast.error(message.message || "An error occurred");
        break;

      default:
        console.log("[WebSocket] Unhandled message type:", message.type);
    }
  }
}

export const websocketManager = new WebSocketManager();
