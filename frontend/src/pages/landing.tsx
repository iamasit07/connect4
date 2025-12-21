import { useState } from "react";
import { useNavigate } from "react-router-dom";

const LandingPage = () => {
  const [username, setUsername] = useState("");
  const [reconnectUsername, setReconnectUsername] = useState("");
  const [reconnectGameID, setReconnectGameID] = useState("");
  const navigate = useNavigate();

  const handleStartGame = () => {
    if (username.trim() === "") {
      alert("Please enter a username");
      return;
    }
    if (username.toUpperCase() === "BOT") {
      alert("Username 'BOT' is reserved. Please choose another username.");
      return;
    }

    // Clear any previous game/reconnect state
    localStorage.removeItem("gameID");
    localStorage.removeItem("isReconnecting");
    localStorage.setItem("username", username);
    navigate("/game/queue");
  };

  const handleReconnect = () => {
    // Require at least one: username OR gameID
    if (reconnectUsername.trim() === "" && reconnectGameID.trim() === "") {
      alert("Please enter either your username or game ID to reconnect");
      return;
    }

    // Store provided values
    if (reconnectUsername.trim() !== "") {
      localStorage.setItem("username", reconnectUsername);
    }
    if (reconnectGameID.trim() !== "") {
      localStorage.setItem("gameID", reconnectGameID);
    }
    localStorage.setItem("isReconnecting", "true");

    // Route based on what was provided
    if (reconnectGameID.trim() !== "") {
      // GameID provided - navigate to specific game
      navigate(`/game/${reconnectGameID}`);
    } else {
      // Username-only - navigate to reconnect route
      navigate("/game/reconnect");
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-md p-8 max-w-md w-full space-y-6">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">4 in a Row</h1>
          <p className="text-sm text-gray-600">Connect four to win</p>
        </div>

        {/* New Game Section */}
        <div className="border-2 border-gray-200 rounded-lg p-4">
          <h2 className="text-lg font-semibold text-gray-800 mb-3">
            Start New Game
          </h2>
          <input
            type="text"
            placeholder="Enter your username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleStartGame()}
            className="w-full px-4 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 mb-3"
          />
          <button
            onClick={handleStartGame}
            className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
          >
            Play Game
          </button>
        </div>

        {/* Reconnect Section */}
        <div className="border-2 border-green-200 rounded-lg p-4 bg-green-50">
          <h2 className="text-lg font-semibold text-gray-800 mb-2">
            Reconnect to Game
          </h2>
          <p className="text-xs text-gray-600 mb-3">
            Enter either your Game ID or username to reconnect
          </p>
          <input
            type="text"
            placeholder="Enter Game ID (from URL)"
            value={reconnectGameID}
            onChange={(e) => setReconnectGameID(e.target.value)}
            className="w-full px-4 py-2 border border-green-300 rounded focus:outline-none focus:ring-2 focus:ring-green-500 mb-2"
          />
          <input
            type="text"
            placeholder="Enter your username"
            value={reconnectUsername}
            onChange={(e) => setReconnectUsername(e.target.value)}
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
      </div>
    </div>
  );
};

export default LandingPage;
