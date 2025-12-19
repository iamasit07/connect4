import React, { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import Board from "../components/board";
import useWebSocket from "../hooks/useWebSocket";

const GamePage: React.FC = () => {
  const { connected, gameState, joinQueue, makeMove } = useWebSocket();
  const navigate = useNavigate();
  const hasJoinedQueue = useRef(false);

  useEffect(() => {
    const username = localStorage.getItem("username");
    if (!username) {
      navigate("/");
      return;
    }

    // Only join queue once when connected
    if (connected && !hasJoinedQueue.current) {
      console.log("Joining queue with username:", username);
      hasJoinedQueue.current = true;

      // Small delay to ensure WebSocket is fully ready
      setTimeout(() => {
        joinQueue(username);
      }, 100);
    }
  }, [connected, navigate, joinQueue]);

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
