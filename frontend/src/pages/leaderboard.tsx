import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

interface PlayerStats {
  username: string;
  games_played: number;
  games_won: number;
  win_rate: number;
}

const LeaderboardPage = () => {
  const [leaderboard, setLeaderboard] = useState<PlayerStats[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const navigate = useNavigate();

  useEffect(() => {
    fetchLeaderboard();
  }, []);

  const fetchLeaderboard = async () => {
    try {
      const apiUrl = import.meta.env.VITE_BACKEND_URL || "http://localhost:8080";
      const response = await fetch(`${apiUrl}/api/leaderboard`);
      if (!response.ok) {
        throw new Error("Failed to fetch leaderboard");
      }

      const data = await response.json();
      setLeaderboard(data || []);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <p className="text-gray-600">Loading...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen bg-gray-50 gap-4">
        <p className="text-red-600">Error: {error}</p>
        <button
          onClick={fetchLeaderboard}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center min-h-screen py-8 bg-gray-50 px-4">
      <h1 className="text-2xl font-bold mb-6 text-gray-900">Leaderboard</h1>

      <div className="w-full max-w-2xl">
        {leaderboard.length === 0 ? (
          <p className="text-center text-gray-500">No games yet</p>
        ) : (
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <table className="w-full">
              <thead className="bg-gray-100 border-b">
                <tr>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-700">
                    Rank
                  </th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-700">
                    Player
                  </th>
                  <th className="px-4 py-3 text-center text-sm font-medium text-gray- 700">
                    Played
                  </th>
                  <th className="px-4 py-3 text-center text-sm font-medium text-gray-700">
                    Won
                  </th>
                  <th className="px-4 py-3 text-center text-sm font-medium text-gray-700">
                    Win %
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {leaderboard.map((player, index) => (
                  <tr key={player.username} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      {index === 0 && "🥇"}
                      {index === 1 && "🥈"}
                      {index === 2 && "🥉"}
                      {index > 2 && `#${index + 1}`}
                    </td>
                    <td className="px-4 py-3 font-medium text-gray-900">
                      {player.username}
                    </td>
                    <td className="px-4 py-3 text-center text-gray-600">
                      {player.games_played}
                    </td>
                    <td className="px-4 py-3 text-center text-gray-600">
                      {player.games_won}
                    </td>
                    <td className="px-4 py-3 text-center font-medium text-blue-600">
                      {player.win_rate}%
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <button
        onClick={() => navigate("/")}
        className="mt-6 px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
      >
        Back to Home
      </button>
    </div>
  );
};

export default LeaderboardPage;
