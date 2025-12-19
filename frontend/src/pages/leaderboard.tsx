import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

interface PlayerStats {
  username: string;
  gamesPlayed: number;
  wins: number;
  winRate: number;
}

const LeaderboardPage = () => {
  const [leaderboard, setLeaderboard] = useState<PlayerStats[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const navigate = useNavigate();

  useEffect(() => {
    fetchleaderboard();
  }, []);

  const fetchleaderboard = async () => {
    try {
      const response = await fetch("/api/leaderboard?limit=10");
      if (!response.ok) {
        throw new Error("Failed to fetch leaderboard");
      }

      const data = await response.json();
      setLeaderboard(data);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div>Loading leaderboard...</div>;
  }

  if (error) {
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <p className="text-red-500">Error: {error}</p>
      <button
        onClick={fetchleaderboard}
        className="px-4 py-2 bg-blue-500 text-white rounded"
      >
        Retry
      </button>
    </div>;
  }

  return (
    <div className="leaderboard-page">
      <h1>Leaderboard</h1>
      <table>
        <thead>
          <tr>
            <th>Rank</th>
            <th>Username</th>
            <th>Games Played</th>
            <th>Wins</th>
            <th>Win Rate</th>
          </tr>
        </thead>
        <tbody>
          {leaderboard.map((player: PlayerStats, index: number) => (
            <tr key={player.username}>
              <td>{index + 1}</td>
              <td>{player.username}</td>
              <td>{player.gamesPlayed}</td>
              <td>{player.wins}</td>
              <td>{(player.winRate * 100).toFixed(2)}%</td>
            </tr>
          ))}
        </tbody>
      </table>
      <button
        onClick={() => navigate("/")}
        className="mt-4 px-4 py-2 bg-blue-500 text-white rounded"
      >
        Back to Home
      </button>
    </div>
  );
};

export default LeaderboardPage;
