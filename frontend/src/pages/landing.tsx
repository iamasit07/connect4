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
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-md p-8 max-w-sm w-full">
        <div className="text-center mb-6">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">4 in a Row</h1>
          <p className="text-sm text-gray-600">Connect four to win</p>
        </div>

        <div className="space-y-4">
          <input
            type="text"
            placeholder="Enter your username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleStartGame()}
            className="w-full px-4 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          
          <button
            onClick={handleStartGame}
            className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
          >
            Play Game
          </button>

          <button
            onClick={() => navigate("/leaderboard")}
            className="w-full px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 transition"
          >
            Leaderboard
          </button>
        </div>
      </div>
    </div>
  );
};

export default LandingPage;
