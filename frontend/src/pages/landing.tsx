import { useState } from "react";
import { useNavigate } from "react-router-dom";

const LandingPage = () => {
  const [username, setUsername] = useState("");
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

    // Save username and navigate to game page
    localStorage.setItem("username", username);
    navigate("/game");
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-600 via-blue-700 to-purple-700 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl p-8 max-w-md w-full">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-5xl font-bold text-gray-800 mb-2">4 in a Row</h1>
          <p className="text-gray-600">Connect four discs to win!</p>
        </div>

        {/* Create/Start Game Section */}
        <div className="mb-6">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">Start a Game</h2>
          <input
            type="text"
            placeholder="Enter your username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleStartGame()}
            className="w-full px-4 py-3 border-2 border-gray-300 rounded-lg focus:outline-none focus:border-blue-500 transition"
          />
          <button
            onClick={handleStartGame}
            className="w-full mt-3 px-6 py-3 bg-blue-600 text-white font-semibold rounded-lg hover:bg-blue-700 transition shadow-md"
          >
            Create Game
          </button>
        </div>

        {/* Join Game Section */}
        <div className="mb-6">
          <h2 className="text-xl font-semibold text-gray-800 mb-3">Join a Game</h2>
          <button
            onClick={handleStartGame}
            className="w-full px-6 py-3 bg-green-600 text-white font-semibold rounded-lg hover:bg-green-700 transition shadow-md"
          >
            Join Queue
          </button>
          <p className="text-sm text-gray-500 mt-2 text-center">
            You'll be matched with another player or a bot
          </p>
        </div>

        {/* Divider */}
        <div className="border-t border-gray-300 my-6"></div>

        {/* View Leaderboard Button */}
        <button
          onClick={() => navigate("/leaderboard")}
          className="w-full px-6 py-3 bg-gray-700 text-white font-semibold rounded-lg hover:bg-gray-800 transition shadow-md"
        >
          🏆 View Leaderboard
        </button>
      </div>
    </div>
  );
};

export default LandingPage;
