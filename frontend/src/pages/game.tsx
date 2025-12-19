import React, { useEffect, useRef } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Board from "../components/board";
import useWebSocket from "../hooks/useWebSocket";

const GamePage: React.FC = () => {
  const { gameID: urlGameID } = useParams<{ gameID: string }>();
  const { connected, gameState, joinQueue, makeMove, reconnect } =
    useWebSocket();
  const navigate = useNavigate();
  const hasJoinedQueue = useRef(false);

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
        console.log("Attempting to reconnect with URL gameID:", { username, gameID: urlGameID });
        reconnect(username, urlGameID);
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
          console.log("Attempting to reconnect with:", { username, gameID: storedGameID });
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

  const handleColumnClick = (col: number) => {
    console.log("Column clicked:", col);
    makeMove(col);
  };

  const handlePlayAgain = () => {
    localStorage.removeItem("username");
    navigate("/");
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
            Bot joins after 10 seconds
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
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-50 gap-6 p-4">
      {/* Game ID Display */}
      {gameState.gameId && (
        <div className="text-xs text-gray-500 font-mono">
          Game ID: {gameState.gameId}
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
          <div className="space-y-1">
            <p className="text-lg font-medium text-gray-900">
              {gameState.currentTurn === gameState.yourPlayer
                ? "Your Turn"
                : `${gameState.opponent}'s Turn`}
            </p>
            <div className="flex items-center justify-center gap-2 text-sm text-gray-600">
              <span className="flex items-center gap-1">
                You:{" "}
                <span className="inline-block w-4 h-4 rounded-full bg-yellow-400"></span>
              </span>
              <span className="flex items-center gap-1">
                Opponent:{" "}
                <span className="inline-block w-4 h-4 rounded-full bg-red-500"></span>
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
    </div>
  );
};

export default GamePage;
