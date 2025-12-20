import React, { useEffect, useRef } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Board from "../components/board";
import DisconnectNotification from "../components/DisconnectNotification";
import useWebSocket from "../hooks/useWebSocket";

const GamePage: React.FC = () => {
  const { gameID: urlGameID } = useParams<{ gameID: string }>();
  const { connected, gameState, joinQueue, makeMove, reconnect } =
    useWebSocket();
  const navigate = useNavigate();
  const hasJoinedQueue = useRef(false);
  const [, forceUpdate] = React.useState(0);

  // Navigate to correct URL when game starts
  useEffect(() => {
    if (gameState.gameId && urlGameID !== gameState.gameId) {
      navigate(`/game/${gameState.gameId}`, { replace: true });
    }
  }, [gameState.gameId, urlGameID, navigate]);

  useEffect(() => {
    const username = localStorage.getItem("username");
    const storedGameID = localStorage.getItem("gameID");
    const isReconnecting = localStorage.getItem("isReconnecting") === "true";

    // If we have gameID in URL, treat it as reconnection
    if (urlGameID && urlGameID !== "queue") {
      if (!username) {
        navigate("/");
        return;
      }

      if (connected && !hasJoinedQueue.current) {
        hasJoinedQueue.current = true;
        
        // Add small delay to ensure WebSocket is fully ready
        setTimeout(() => {
          console.log("Attempting to reconnect with URL gameID:", {
            username,
            gameID: urlGameID,
          });
          reconnect(username, urlGameID);
        }, 100);
      }
      return;
    }

    // Handle normal flow
    if (!username && !storedGameID) {
      navigate("/");
      return;
    }

    if (connected && !hasJoinedQueue.current) {
      hasJoinedQueue.current = true;

      setTimeout(() => {
        if (isReconnecting && storedGameID) {
          console.log("Attempting to reconnect with:", {
            username,
            gameID: storedGameID,
          });
          localStorage.removeItem("isReconnecting");
          localStorage.removeItem("gameID");
          reconnect(username || undefined, storedGameID);
        } else {
          if (!username) {
            navigate("/");
            return;
          }
          console.log("Joining queue with username:", username);
          joinQueue(username);
        }
      }, 100);
    }
  }, [connected, navigate, joinQueue, reconnect, urlGameID]);

  // Countdown timer for matchmaking queue
  useEffect(() => {
    if (gameState.inQueue && gameState.queuedAt) {
      const interval = setInterval(() => {
        forceUpdate((n) => n + 1);
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [gameState.inQueue, gameState.queuedAt]);

  const queueCountdown =
    gameState.inQueue && gameState.queuedAt
      ? Math.max(0, 10 - Math.floor((Date.now() - gameState.queuedAt) / 1000))
      : null;

  const handleColumnClick = (col: number) => {
    console.log("Column clicked:", col);
    makeMove(col);
  };

  const handlePlayAgain = () => {
    localStorage.removeItem("username");
    navigate("/");
  };

  // Determine background color based on turn
  const getBackgroundColor = () => {
    if (gameState.gameOver) return "bg-gray-50";
    if (gameState.currentTurn === 1) return "bg-yellow-50";
    if (gameState.currentTurn === 2) return "bg-red-50";
    return "bg-gray-50";
  };

  if (!connected) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <p className="text-gray-600">Connecting...</p>
      </div>
    );
  }

  if (gameState.inQueue) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="text-center">
          <p className="text-lg text-gray-800">Finding opponent...</p>
          <p className="text-sm text-gray-500 mt-2">
            {queueCountdown !== null ? (
              <span>
                Bot joins in{" "}
                <span className="font-bold text-blue-600">
                  {queueCountdown} second{queueCountdown !== 1 ? "s" : ""}
                </span>
              </span>
            ) : (
              "Waiting..."
            )}
          </p>
        </div>
      </div>
    );
  }

  if (!gameState.gameId) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <p className="text-gray-600">Waiting...</p>
      </div>
    );
  }

  return (
    <div
      className={`flex flex-col items-center justify-center min-h-screen ${getBackgroundColor()} gap-6 p-4`}
    >
      {/* Game ID Display */}
      {gameState.gameId && (
        <div className="text-xs text-gray-500 font-mono">
          Game ID: {gameState.gameId}
        </div>
      )}

      {/* Turn Indicator Banner */}
      {!gameState.gameOver && gameState.currentTurn && (
        <div
          className={`w-full max-w-md px-4 py-2 rounded text-center font-bold ${
            gameState.currentTurn === 1
              ? "bg-yellow-400 text-yellow-900"
              : "bg-red-500 text-white"
          }`}
        >
          Player {gameState.currentTurn}'s Turn
          {gameState.currentTurn === gameState.yourPlayer && " (You)"}
        </div>
      )}

      <div className="text-center">
        {gameState.gameOver ? (
          <h2 className="text-2xl font-bold text-gray-900">
            {gameState.winner === "draw"
              ? "Draw!"
              : `${gameState.winner} Wins!`}
          </h2>
        ) : (
          <div className="space-y-2">
            {/* Color Legend */}
            <div className="flex items-center justify-center gap-4 text-sm text-gray-700 bg-white px-3 py-2 rounded border border-gray-200">
              <span className="flex items-center gap-2">
                <span className="inline-block w-4 h-4 rounded-full bg-yellow-400"></span>
                <span>
                  Player 1 {gameState.yourPlayer === 1 && "(You)"}
                  {gameState.yourPlayer === 2 && `(${gameState.opponent})`}
                </span>
              </span>
              <span className="flex items-center gap-2">
                <span className="inline-block w-4 h-4 rounded-full bg-red-500"></span>
                <span>
                  Player 2 {gameState.yourPlayer === 2 && "(You)"}
                  {gameState.yourPlayer === 1 && `(${gameState.opponent})`}
                </span>
              </span>
            </div>
          </div>
        )}
      </div>

      <Board
        board={gameState.board}
        yourPlayer={gameState.yourPlayer}
        currentTurn={gameState.currentTurn}
        onColumnClick={handleColumnClick}
        gameOver={gameState.gameOver}
      />

      {gameState.gameOver && (
        <button
          onClick={handlePlayAgain}
          className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
        >
          Back to Home
        </button>
      )}

      {/* Disconnect Notification */}
      <DisconnectNotification
        isDisconnected={gameState.opponentDisconnected}
        disconnectedAt={gameState.disconnectedAt}
      />
    </div>
  );
};

export default GamePage;
