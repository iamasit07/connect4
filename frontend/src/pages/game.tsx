import { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import Board from "../components/board";
import useWebSocket from "../hooks/useWebSocket";

const GamePage = () => {
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
      <div className="flex items-center justify-center min-h-screen bg-gray-100">
        <p className="text-xl text-gray-700">Connecting to server...</p>
      </div>
    );
  }

  if (gameState.inQueue) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-100">
        <div className="text-center">
          <p className="text-2xl text-gray-800 mb-2">
            Searching for opponent...
          </p>
          <p className="text-sm text-gray-600">
            Bot will join if no player found in 10 seconds
          </p>
        </div>
      </div>
    );
  }

  if (!gameState.gameId) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-100">
        <p className="text-xl text-gray-700">Waiting for game to start...</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 gap-4">
      {/* Game Status */}
      <div className="text-center">
        {gameState.gameOver ? (
          <h2 className="text-3xl font-bold text-gray-800">
            {gameState.winner === "draw"
              ? "Game Draw!"
              : `${gameState.winner} Wins!`}
          </h2>
        ) : (
          <p className="text-xl text-gray-700">
            {gameState.currentTurn === gameState.yourPlayer
              ? "Your Turn"
              : `${gameState.opponent}'s Turn`}
          </p>
        )}
      </div>

      {/* Game Board */}
      <Board
        board={gameState.board}
        yourPlayer={gameState.yourPlayer}
        currentTurn={gameState.currentTurn}
        onColumnClick={handleColumnClick}
        gameOver={gameState.gameOver}
      />

      {/* Play Again Button */}
      {gameState.gameOver && (
        <button
          onClick={handlePlayAgain}
          className="px-8 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition font-semibold"
        >
          Back to Home
        </button>
      )}
    </div>
  );
};

export default GamePage;
