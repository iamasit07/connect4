import { RefObject, useEffect, useRef, useState } from "react";
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
  joinQueue: (difficulty?: string) => void;
  makeMove: (column: number) => void;
  reconnect: (gameID?: string) => void;
  requestRematch: () => void;
  respondToRematch: (accept: boolean) => void;
  justReceivedGameStart: RefObject<boolean>;
  showPostGameNotification: boolean;
  postGameMessage: string;
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
    rematchRequested: false,
    rematchRequester: null,
    rematchTimeout: null,
    allowRematch: true, // Default true, backend will override
  });
  
  const [showPostGameNotification, setShowPostGameNotification] = useState(false);
  const [postGameMessage, setPostGameMessage] = useState("");

  const ws = useRef<WebSocket | null>(null);
  const navigate = useNavigate();
  const { getToken, logout } = useAuth();
  
  // Use refs to avoid stale closures while keeping WebSocket stable
  const navigateRef = useRef(navigate);
  const logoutRef = useRef(logout);
  const justReceivedGameStart = useRef(false);
  
  // Update refs on every render
  useEffect(() => {
    navigateRef.current = navigate;
    logoutRef.current = logout;
  });

  useEffect(() => {
    const wsUrl = import.meta.env.VITE_WS_URL || "ws://localhost:8080";
    ws.current = new WebSocket(`${wsUrl}/ws`);

    ws.current.onopen = () => {
      setConnected(true);
    };

    ws.current.onmessage = (event: MessageEvent) => {
      const message: ServerMessage = JSON.parse(event.data);

      switch (message.type) {
        case "game_start":
          if (message.gameId) {
            sessionStorage.setItem("gameID", message.gameId);
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
          // Set flag to prevent reconnect when navigating
          justReceivedGameStart.current = true;
          // Navigate to the new game route
          if (message.gameId) {
            navigateRef.current(`/game/${message.gameId}`);
          }
          // Reset flag after navigation
          setTimeout(() => {
            justReceivedGameStart.current = false;
          }, 500);
          break;
        case "reconnect_success":
          if (message.gameId) {
            sessionStorage.setItem("gameID", message.gameId);
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
          sessionStorage.removeItem("gameID"); // Clear finished game
          setGameState((prevState: GameState) => ({
            ...prevState,
            gameOver: true,
            winner: message.winner ?? null,
            reason: message.reason ?? null,
            board: message.board ?? prevState.board,
            opponentDisconnected: false,
            disconnectedAt: null,
            allowRematch: message.allowRematch ?? true, // Use backend value
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
          sessionStorage.removeItem("gameID");
          
          // Logout to clear the HttpOnly cookie (don't await, fire and forget)
          logoutRef.current();
          
          // Navigate and show alert
          navigateRef.current("/login");
          alert(message.message || "You have been logged out");
          break;
        case "session_invalidated":
          // Session was invalidated (logged in from another device)
          sessionStorage.removeItem("gameID");
          logoutRef.current();
          navigateRef.current("/login");
          alert(message.message || "Your session has been invalidated. Please log in again.");
          break;
        case "session_expired":
          // Session expired (30 days passed)
          sessionStorage.removeItem("gameID");
          logoutRef.current();
          navigateRef.current("/login");
          alert("Your session has expired. Please log in again.");
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
          sessionStorage.removeItem("gameID");
          
          setGameState((prevState: GameState) => ({
            ...prevState,
            matchEnded: true,
            matchEndedAt: Date.now(),
            gameOver: true,
            inQueue: false,
            reason: message.message || `Error: ${message.type}`,
          }));
          break;
        case "rematch_request":
          setGameState((prevState: GameState) => ({
            ...prevState,
            rematchRequested: true,
            rematchRequester: message.rematchRequester ?? null,
            rematchTimeout: message.rematchTimeout ?? null,
          }));
          break;
        case "rematch_accepted":
          // Brief confirmation message, then wait for game_start
          setGameState((prevState: GameState) => ({
            ...prevState,
            rematchRequested: false,
            rematchRequester: null,
            rematchTimeout: null,
          }));
          break;
        case "rematch_declined":
        case "rematch_timeout":
        case "rematch_cancelled":
          // Clear rematch state and show notification
          setGameState((prevState: GameState) => ({
            ...prevState,
            rematchRequested: false,
            rematchRequester: null,
            rematchTimeout: null,
          }));
          // Show post-game notification
          const notificationMsg = message.type === "rematch_declined" 
            ? "Rematch declined" 
            : message.type === "rematch_timeout"
            ? "Rematch request timed out"
            : "Rematch cancelled";
          setPostGameMessage(notificationMsg);
          setShowPostGameNotification(true);
          
          // Update allowRematch based on backend
          setGameState((prevState:  GameState) => ({
            ...prevState,
            allowRematch: message.allowRematch ?? false,
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
  }, []); // Empty dependencies - only create WebSocket once

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

  const joinQueue = (difficulty?: string) => {
    const jwt = getToken() || "";
    
    if (!jwt) {
      console.error("No auth token found");
      navigateRef.current("/login");
      return;
    }
    
    sendMessage({ 
      type: "join_queue", 
      jwt,
      difficulty: difficulty || ""
    });
    
    if (!difficulty) {
      setGameState((prevState: GameState) => ({
        ...prevState,
        inQueue: true,
        queuedAt: Date.now(),
      }));
    }
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
      navigateRef.current("/login");
      return;
    }
    
    sendMessage({
      type: "reconnect",
      gameID: gameID,
      jwt: jwt,
    });
  };

  const requestRematch = () => {
    const jwt = getToken() || "";
    if (!jwt) {
      console.error("No auth token found");
      return;
    }
    sendMessage({ type: "rematch_request", jwt });
  };

  const respondToRematch = (accept: boolean) => {
    const jwt = getToken() || "";
    if (!jwt) {
      console.error("No auth token found");
      return;
    }
    sendMessage({ 
      type: "rematch_response", 
      jwt,
      rematchResponse: accept ? "accept" : "decline"
    });
  };

  return {
    connected,
    gameState,
    joinQueue,
    makeMove,
    reconnect,
    requestRematch,
    respondToRematch,
    justReceivedGameStart,
    showPostGameNotification,
    postGameMessage,
  };
};

export default useWebSocket;
