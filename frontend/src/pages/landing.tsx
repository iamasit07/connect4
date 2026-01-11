import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";

const LandingPage = () => {
  const [reconnectGameID, setReconnectGameID] = useState("");
  const navigate = useNavigate();
  const { user, logout } = useAuth();

  const handlePlayOnline = () => {
    localStorage.removeItem("gameID");
    navigate("/game/queue");
  };

  const handlePlayBot = () => {
    navigate("/bot-difficulty");
  };

  const handleReconnect = () => {
    if (reconnectGameID.trim()) {
      navigate(`/game/${reconnectGameID}`);
    } else {
      navigate("/game");
    }
  };

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-md p-8 max-w-md w-full space-y-6">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">4 in a Row</h1>
          <p className="text-xl text-gray-600">Connect four to win</p>
          {user && (
            <p className="text-sm text-gray-500 mt-2">
              Welcome back, {user.username}!
            </p>
          )}
        </div>

        {/* Start New Game Section */}
        <div className="border-2 border-blue-200 rounded-lg p-4 bg-blue-50">
          <h2 className="text-lg font-semibold text-gray-800 mb-3">
            Start New Game
          </h2>
          
          <div className="space-y-2">
            <button
              onClick={handlePlayOnline}
              className="w-full px-4 py-3 bg-blue-600 text-white rounded hover:bg-blue-700 transition font-medium flex items-center justify-center gap-2"
            >
              <span className="text-xl">👥</span>
              <span>Play Online</span>
            </button>
            <button
              onClick={handlePlayBot}
              className="w-full px-4 py-3 bg-purple-600 text-white rounded hover:bg-purple-700 transition font-medium flex items-center justify-center gap-2"
            >
              <span className="text-xl">🤖</span>
              <span>Play vs Bot</span>
            </button>
          </div>
        </div>

        {/* Reconnect Section */}
        <div className="border-2 border-green-200 rounded-lg p-4 bg-green-50">
          <h2 className="text-lg font-semibold text-gray-800 mb-2">
            Reconnect to Game
          </h2>
          <p className="text-xs text-gray-600 mb-3">
            Leave empty to auto-reconnect, or enter specific Game ID
          </p>
          <input
            type="text"
            placeholder="Game ID (optional)"
            value={reconnectGameID}
            onChange={(e) => setReconnectGameID(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleReconnect()}
            className="w-full px-4 py-2 border border-green-300 rounded focus:outline-none focus:ring-2 focus:ring-green-500 mb-3"
          />
          <button
            onClick={handleReconnect}
            className="w-full px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 transition"
          >
            Reconnect
          </button>
        </div>

        <button
          onClick={() => navigate("/leaderboard")}
          className="w-full px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 transition"
        >
          Leaderboard
        </button>

        <button
          onClick={handleLogout}
          className="w-full px-4 py-2 bg-red-100 text-red-700 rounded hover:bg-red-200 transition"
        >
          Logout
        </button>
      </div>
    </div>
  );
};

export default LandingPage;
