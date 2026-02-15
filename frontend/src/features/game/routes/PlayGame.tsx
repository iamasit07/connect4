import { useNavigate, useLocation } from "react-router-dom";
import { useEffect } from "react";
import { ModeSelection } from "../components/ModeSelection";
import type { BotDifficulty } from "../types";

const PlayGame = () => {
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const state = location.state as {
      mode?: string;
      difficulty?: BotDifficulty;
    } | null;

    if (state?.mode === "bot" && state?.difficulty) {
      navigate("/play/bot", {
        state: { difficulty: state.difficulty },
        replace: true,
      });
    }
  }, [location, navigate]);

  const handleSelectPvP = () => {
    navigate("/play/queue", { state: { from: "/play" } });
  };

  const handleSelectBot = (difficulty: BotDifficulty) => {
    navigate("/play/bot", { state: { difficulty } });
  };

  return (
    <div className="flex-1 w-full bg-background flex flex-col items-center justify-center">
      <ModeSelection
        onSelectPvP={handleSelectPvP}
        onSelectBot={handleSelectBot}
      />
    </div>
  );
};

export default PlayGame;
