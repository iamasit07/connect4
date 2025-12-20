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
  });

  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    const wsUrl = "wss://four-in-a-row-backend-tnan.onrender.com";
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
          setGameState((prevState: GameState) => ({
            ...prevState,
            inQueue: true,
            queuedAt: Date.now(),
          }));
          break;
        case "game_start":
          // Save gameID and sessionToken to localStorage for automatic reconnection
          if (message.gameId) {
            localStorage.setItem("gameID", message.gameId);
          }
          if (message.sessionToken) {
            localStorage.setItem("sessionToken", message.sessionToken);
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
        case "move_made":
          setGameState((prevState: GameState) => ({
            ...prevState,
            board: message.board ?? prevState.board,
            currentTurn: (message.nextTurn as PlayerID) ?? null,
          }));
          break;
        case "game_over":
          // Clear gameID and sessionToken from localStorage when game ends
          localStorage.removeItem("gameID");
          localStorage.removeItem("sessionToken");
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
        case "error":
          // Handle error messages from backend
          console.error("Server error:", message.message);
          alert(`Error: ${message.message || "Failed to reconnect. Please try again."}`);
          // If it's a reconnection error, redirect to home
          if (message.message?.includes("reconnect") || message.message?.includes("game")) {
            localStorage.removeItem("gameID");
            localStorage.removeItem("sessionToken");
            localStorage.removeItem("isReconnecting");
            window.location.href = "/";
          }
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
    sendMessage({ type: "join_queue", username });
  };

  const makeMove = (column: number) => {
    if (gameState.gameId) {
      sendMessage({ type: "make_move", column });
    }
  };

  const reconnect = (username?: string, gameID?: string) => {
    const token = localStorage.getItem("sessionToken") || "";
    sendMessage({
      type: "reconnect",
      username: username || "",
      gameID: gameID || "",
      token: token,
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
