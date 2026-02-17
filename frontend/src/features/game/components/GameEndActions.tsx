import { Button } from "@/components/ui/button";
import { Home, RotateCcw } from "lucide-react";
import { useGameStore } from "../store/gameStore";

interface GameEndActionsProps {
  onPlayAgain: () => void;
  onGoHome: () => void;
}

export const GameEndActions = ({
  onPlayAgain,
  onGoHome,
}: GameEndActionsProps) => {
  const { gameStatus, gameMode } = useGameStore();

  if (gameStatus !== "finished") return null;
  const isBot = gameMode === "bot";

  return (
    <div className="w-full max-w-[min(90vw,500px)] mx-auto mt-2 sm:mt-4 flex gap-3 flex-shrink-0">
      <Button
        onClick={onGoHome}
        variant={isBot ? "outline" : "default"}
        className="flex-1 gap-1.5 sm:gap-2 h-9 sm:h-11 text-sm sm:text-base"
      >
        <Home className="w-3.5 h-3.5 sm:w-4 sm:h-4" />
        Dashboard
      </Button>
      {isBot && (
        <Button
          onClick={onPlayAgain}
          className="flex-1 gap-1.5 sm:gap-2 h-9 sm:h-11 text-sm sm:text-base"
        >
          <RotateCcw className="w-3.5 h-3.5 sm:w-4 sm:h-4" />
          Play Again
        </Button>
      )}
    </div>
  );
};
