import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

interface GameHistoryItem {
  game_id: string;
  player1: {
    id?: number;
    username: string;
  };
  player2: {
    id?: number;
    username: string;
  };
  result: "won" | "lost" | "draw";
  reason: string;
  total_moves: number;
  duration_seconds: number;
  created_at: string;
  finished_at: string;
}

interface GroupedGames {
  [key: string]: GameHistoryItem[];
}

const GameHistory: React.FC = () => {
  const [games, setGames] = useState<GameHistoryItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    fetchGameHistory();
  }, []);

  const fetchGameHistory = async () => {
    try {
      const response = await fetch(
        `${import.meta.env.VITE_API_URL || "http://localhost:8080"}/api/games/history`,
        {
          credentials: "include",
        }
      );

      if (!response.ok) {
        throw new Error("Failed to fetch game history");
      }

      const data = await response.json();
      setGames(data.games || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  const groupGamesByDate = (games: GameHistoryItem[]): GroupedGames => {
    const grouped: GroupedGames = {};
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    games.forEach((game) => {
      const gameDate = new Date(game.finished_at);
      gameDate.setHours(0, 0, 0, 0);

      let dateKey: string;
      if (gameDate.getTime() === today.getTime()) {
        dateKey = "Today";
      } else if (gameDate.getTime() === yesterday.getTime()) {
        dateKey = "Yesterday";
      } else {
        dateKey = gameDate.toLocaleDateString("en-US", {
          month: "long",
          day: "numeric",
          year: "numeric",
        });
      }

      if (!grouped[dateKey]) {
        grouped[dateKey] = [];
      }
      grouped[dateKey].push(game);
    });

    return grouped;
  };

  const getResultBadgeClass = (result: string) => {
    switch (result) {
      case "won":
        return "bg-green-100 text-green-800";
      case "lost":
        return "bg-red-100 text-red-800";
      case "draw":
        return "bg-gray-100 text-gray-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  const handleViewBoard = (gameId: string) => {
    navigate(`/game/review/${gameId}`);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <p className="text-gray-600">Loading game history...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="text-center">
          <p className="text-red-600 mb-4">{error}</p>
          <button
            onClick={() => navigate("/")}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Back to Home
          </button>
        </div>
      </div>
    );
  }

  const groupedGames = groupGamesByDate(games);
  const dateKeys = Object.keys(groupedGames);

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-6xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold text-gray-900">Game History</h1>
          <button
            onClick={() => navigate("/")}
            className="px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 transition"
          >
            ← Back to Home
          </button>
        </div>

        {games.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-8 text-center">
            <p className="text-gray-600">No games played yet. Start your first game!</p>
          </div>
        ) : (
          dateKeys.map((dateKey) => (
            <div key={dateKey} className="mb-8">
              <h2 className="text-xl font-semibold text-gray-800 mb-3">{dateKey}</h2>
              <div className="bg-white rounded-lg shadow overflow-hidden">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        You
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Opponent
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Game ID
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Result
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {groupedGames[dateKey].map((game) => (
                      <tr key={game.game_id} className="hover:bg-gray-50">
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                          {game.player1.username}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                          {game.player2.username}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 font-mono">
                          {game.game_id.substring(0, 8)}...
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getResultBadgeClass(
                              game.result
                            )}`}
                          >
                            {game.result.toUpperCase()}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm">
                          <button
                            onClick={() => handleViewBoard(game.game_id)}
                            className="text-blue-600 hover:text-blue-900 font-medium"
                          >
                            View Board →
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default GameHistory;