import { useEffect, useRef, useState } from "react";
import {
  ClientMessage,
  GameState,
  PlayerID,
  ServerMessage,
} from "../types/game";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

interface UseWebSocketReturn {
  connected: boolean;
  gameState: GameState;
  joinQueue: () => void;
  makeMove: (column: number) => void;
  reconnect: (gameID?: string) => void;
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
  const navigate = useNavigate();
  const { getToken, logout } = useAuth();

  useEffect(() => {
    const wsUrl = import.meta.env.VITE_WS_URL || "ws://localhost:8080";
    ws.current = new WebSocket(`${wsUrl}/ws`);

    ws.current.onopen = () => {
      setConnected(true);
    };

    ws.current.onmessage = (event: MessageEvent) => {
      const message: ServerMessage = JSON.parse(event.data);

      switch (message.type) {
        case "queue_joined":
          setGameState((prevState: GameState) => ({
            ...prevState,
            inQueue: true,
            queuedAt: Date.now(),
          }));
          break;
        case "game_start":
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
          localStorage.removeItem("gameID");
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
        case "force_disconnect":
          // Clear gameID from localStorage
          localStorage.removeItem("gameID");
          
          // Logout to clear the HttpOnly cookie (don't await, fire and forget)
          logout();
          
          // Navigate and show alert
          navigate("/login");
          alert(message.message || "You have been logged out");
          break;
        case "error":
        case "queue_error":
        case "invalid_move":
        case "not_your_turn":
          console.error(`Server error [${message.type}]:`, message.message);
          setGameState((prevState: GameState) => ({
            ...prevState,
            error: message.message || `Error: ${message.type}`,
          }));
          setTimeout(() => {
            setGameState((prevState: GameState) => ({
              ...prevState,
              error: null,
            }));
          }, 5000);
          break;
        case "reconnect_failed":
        case "no_active_game":
        case "not_in_game":
        case "game_finished":
        case "game_not_found":
          localStorage.removeItem("gameID");
          
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
      console.log("WebSocket closed");
      setConnected(false);
    };

    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, [navigate]);

  const sendMessage = (message: ClientMessage) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message));
    } else {
      console.error(
        "Cannot send message - WebSocket not open. ReadyState:",
        ws.current?.readyState
      );
    }
  };

  const joinQueue = () => {
    const jwt = getToken() || "";
    
    if (!jwt) {
      console.error("No auth token found");
      navigate("/login");
      return;
    }
    
    sendMessage({ type: "join_queue", jwt });
  };

  const makeMove = (column: number) => {
    if (gameState.gameId) {
      const jwt = getToken() || "";
      sendMessage({ type: "move", column, jwt });
    }
  };

  const reconnect = (gameID?: string) => {
    const jwt = getToken() || "";
    
    if (!jwt) {
      setGameState((prevState: GameState) => ({
        ...prevState,
        matchEnded: true,
        matchEndedAt: Date.now(),
        gameOver: true,
        inQueue: false,
        reason: "Please log in to reconnect.",
      }));
      navigate("/login");
      return;
    }
    
    sendMessage({
      type: "reconnect",
      gameID: gameID,
      jwt: jwt,
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
