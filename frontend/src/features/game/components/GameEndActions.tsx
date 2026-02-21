import { Button } from "@/components/ui/button";
import { Home, RotateCcw } from "lucide-react";
import { useGameStore } from "../store/gameStore";
import { RematchRequest } from "./RematchRequest";

interface GameEndActionsProps {
  onPlayAgain: () => void;
  onGoHome: () => void;
  onSendRematch?: () => void;
  onAcceptRematch?: () => void;
  onDeclineRematch?: () => void;
}

export const GameEndActions = ({
  onPlayAgain,
  onGoHome,
  onSendRematch,
  onAcceptRematch,
  onDeclineRematch,
}: GameEndActionsProps) => {
  const { gameStatus, gameMode, rematchStatus, opponent } = useGameStore();

  if (gameStatus !== "finished") return null;
  const isBot = gameMode === "bot";

  return (
    <div className="w-full max-w-[min(90vw,500px)] mx-auto mt-2 sm:mt-4 flex gap-3 shrink-0">
      <Button
        onClick={onGoHome}
        variant={isBot ? "outline" : "default"}
        className="flex-1 gap-1.5 sm:gap-2 h-9 sm:h-11 text-sm sm:text-base"
      >
        <Home className="w-3.5 h-3.5 sm:w-4 sm:h-4" />
        Dashboard
      </Button>
      {isBot ? (
        <Button
          onClick={onPlayAgain}
          className="flex-1 gap-1.5 sm:gap-2 h-9 sm:h-11 text-sm sm:text-base"
        >
          <RotateCcw className="w-3.5 h-3.5 sm:w-4 sm:h-4" />
          Play Again
        </Button>
      ) : (
          <div className="flex-1">
             <RematchRequest
                onSendRequest={onSendRematch}
                onAcceptRequest={onAcceptRematch}
                onDeclineRequest={onDeclineRematch}
                rematchStatus={rematchStatus}
                opponentName={opponent || "Opponent"}
              />
          </div>
      )}
    </div>
  );
};
