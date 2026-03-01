import { Button } from "@/components/ui/button";
import { Home, RotateCcw } from "lucide-react";
import { useGameStore } from "../store/gameStore";
import { RematchRequest } from "./RematchRequest";

interface GameEndActionsProps {
  onPlayAgain: () => void;
  onGoHome: () => void;
  onSendRematch?: () => void;
}

export const GameEndActions = ({
  onPlayAgain,
  onGoHome,
  onSendRematch,
}: GameEndActionsProps) => {
  const { gameStatus, gameMode, rematchStatus, opponent } = useGameStore();

  if (gameStatus !== "finished") return null;
  const isBot = gameMode === "bot";

  return (
    <div className="w-full max-w-[min(90vw,500px)] mx-auto mt-4 px-4 sm:px-0 flex flex-col sm:flex-row gap-3 shrink-0">
      <Button
        onClick={onGoHome}
        variant="outline"
        className="flex-1 gap-2 h-12 text-base font-semibold border-white/10 hover:bg-muted/50 shadow-sm hover:shadow-md transition-all"
      >
        <Home className="w-4 h-4" />
        Dashboard
      </Button>
      {isBot ? (
        <Button
          onClick={onPlayAgain}
          className="flex-1 gap-2 h-12 text-base font-semibold bg-primary hover:bg-primary/90 shadow-md hover:shadow-xl hover:shadow-primary/20 transition-all"
        >
          <RotateCcw className="w-4 h-4" />
          Play Again
        </Button>
      ) : (
          <div className="flex-1">
             <RematchRequest
                onSendRequest={onSendRematch}
                rematchStatus={rematchStatus}
                opponentName={opponent || "Opponent"}
              />
          </div>
      )}
    </div>
  );
};
