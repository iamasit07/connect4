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
  reconnect: (username: string) => void;
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
  });

  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    ws.current = new WebSocket("ws://localhost:8080/ws");

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
          }));
          break;
        case "game_start":
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
          setGameState((prevState: GameState) => ({
            ...prevState,
            gameOver: true,
            winner: message.winner ?? null,
            reason: message.reason ?? null,
            board: message.board ?? prevState.board,
          }));
          break;
        case "opponent_disconnected":
          alert("Your opponent has disconnected. Waiting for connection...");
          break;
        case "opponent_reconnected":
          alert("Your opponent has reconnected. The game will resume.");
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

  const reconnect = (username: string) => {
    sendMessage({ type: "reconnect", username });
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
