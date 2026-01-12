import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Board from "../components/board";
import { PlayerID } from "../types/game";

interface GameReviewData {
  game_id: string;
  board: PlayerID[][];
  player1_username: string;
  player2_username: string;
  winner: string;
  result: string;
  total_moves: number;
  is_finished: boolean;
  created_at: string;
  finished_at: string;
}

const GameReview: React.FC = () => {
  const { gameId } = useParams<{ gameId: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [gameData, setGameData] = useState<GameReviewData | null>(null);

  useEffect(() => {
    if (!gameId) {
      navigate("/history");
      return;
    }

    // Fetch game board data from API
    fetch(
      `${import.meta.env.VITE_API_URL || "http://localhost:8080"}/api/game/${gameId}/board`,
      {
        credentials: "include",
      }
    )
      .then((res) => {
        if (!res.ok) {
          throw new Error("Failed to fetch game data");
        }
        return res.json();
      })
      .then((data) => {
        setGameData(data);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      });
  }, [gameId, navigate]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <p className="text-gray-600">Loading game...</p>
      </div>
    );
  }

  if (error || !gameData) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="text-center">
          <p className="text-red-600 mb-4">{error || "Game not found"}</p>
          <button
            onClick={() => navigate("/history")}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Back to History
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-lg p-6 max-w-2xl w-full">
        <div className="flex justify-between items-center mb-4">
          <h1 className="text-2xl font-bold text-gray-900">Game Review</h1>
          <button
            onClick={() => navigate("/history")}
            className="px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 transition"
          >
            ← Back to History
          </button>
        </div>

        <div className="mb-4 p-4 bg-blue-50 rounded-lg">
          <p className="text-sm text-gray-700">
            <strong>{gameData.player1_username}</strong> vs{" "}
            <strong>{gameData.player2_username}</strong>
          </p>
          <p className="text-sm text-gray-600 mt-1">
            Winner: <strong>{gameData.winner || "Draw"}</strong>
          </p>
          <p className="text-sm text-gray-600">
            Total Moves: {gameData.total_moves}
          </p>
          <p className="text-xs text-blue-600 mt-2">
            📌 This is a finished game in read-only mode. No moves can be made.
          </p>
        </div>

        <Board
          board={gameData.board}
          onColumnClick={() => {}} // No-op for read-only
          yourPlayer={null}
          currentTurn={null}
          gameOver={true}
        />
      </div>
    </div>
  );
};

export default GameReview;
