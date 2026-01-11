import { useNavigate } from "react-router-dom";

const BotDifficulty = () => {
  const navigate = useNavigate();

  const handleDifficultySelect = (difficulty: string) => {
    navigate(`/game/queue?difficulty=${difficulty}`);
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-md p-8 max-w-md w-full">
        <h1 className="text-2xl font-bold text-gray-900 text-center mb-2">
          Select Bot Difficulty
        </h1>
        <p className="text-sm text-gray-500 text-center mb-6">
          Choose your opponent's skill level
        </p>

        <div className="space-y-3">
          <button
            onClick={() => handleDifficultySelect("easy")}
            className="w-full px-6 py-4 bg-green-500 text-white rounded-lg hover:bg-green-600 transition font-medium text-lg flex items-center justify-center gap-3"
          >
            <span className="text-2xl">🟢</span>
            <div className="text-left">
              <div className="font-bold">Easy</div>
              <div className="text-sm opacity-90">Simple bot, great for beginners</div>
            </div>
          </button>

          <button
            onClick={() => handleDifficultySelect("medium")}
            className="w-full px-6 py-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition font-medium text-lg flex items-center justify-center gap-3"
          >
            <span className="text-2xl">🔵</span>
            <div className="text-left">
              <div className="font-bold">Medium</div>
              <div className="text-sm opacity-90">Balanced challenge</div>
            </div>
          </button>

          <button
            onClick={() => handleDifficultySelect("hard")}
            className="w-full px-6 py-4 bg-red-600 text-white rounded-lg hover:bg-red-700 transition font-medium text-lg flex items-center justify-center gap-3"
          >
            <span className="text-2xl">🔴</span>
            <div className="text-left">
              <div className="font-bold">Hard</div>
              <div className="text-sm opacity-90">Extremely challenging</div>
            </div>
          </button>
        </div>

        <button
          onClick={() => navigate("/")}
          className="w-full mt-6 px-4 py-2 text-sm text-gray-600 hover:text-gray-800 transition"
        >
          ← Back to Home
        </button>
      </div>
    </div>
  );
};

export default BotDifficulty;
