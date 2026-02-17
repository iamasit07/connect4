import { useEffect, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Board } from "../components/Board";
import { GameInfo } from "../components/GameInfo";
import { GameControls } from "../components/GameControls";
import { GameResultBanner } from "../components/GameResultBanner";
import { GameEndActions } from "../components/GameEndActions";
import { RematchOverlay } from "../components/RematchRequest";
import { useGameSocket } from "../hooks/useGameSocket";
import { useGameStore } from "../store/gameStore";
import { toast } from "sonner";

const GamePage = () => {
  const { gameId } = useParams<{ gameId: string }>();
  const navigate = useNavigate();

  // Handle new game starting from a rematch
  const onGameStart = useCallback(
    (newGameId: string) => {
      navigate(`/game/${newGameId}`, { replace: true });
    },
    [navigate],
  );

  const { makeMove, surrender, disconnect, sendMessage } =
    useGameSocket(onGameStart);
  const {
    gameStatus,
    resetGame,
    setRematchStatus,
    gameMode,
    botDifficulty,
    rematchStatus,
    allowRematch,
    opponent,
    gameId: storeGameId,
  } = useGameStore();

  useEffect(() => {
    if (gameStatus === "playing" && storeGameId && storeGameId !== gameId) {
      navigate(`/game/${storeGameId}`, { replace: true });
    }
  }, [gameStatus, storeGameId, gameId, navigate]);

  const handleColumnClick = (col: number) => {
    makeMove(col);
  };

  const handleSurrender = () => {
    surrender();
  };

  const handlePlayAgain = () => {
    if (gameMode === "bot" && botDifficulty) {
      sendMessage({ type: "request_rematch" });
    } else {
      // For PvP without rematch: go back to queue
      disconnect();
      resetGame();
      navigate("/play/queue", { state: { from: `/game/${gameId}` } });
    }
  };

  const handleGoHome = () => {
    disconnect();
    resetGame();
    navigate("/dashboard");
  };

  const handleSendRematch = () => {
    sendMessage({ type: "request_rematch" });
    setRematchStatus("sent");
    toast.info("Rematch request sent!");
  };

  const handleAcceptRematch = () => {
    sendMessage({ type: "rematch_response", rematchResponse: "accept" });
    setRematchStatus("accepted");
  };

  const handleDeclineRematch = () => {
    sendMessage({ type: "rematch_response", rematchResponse: "decline" });
    setRematchStatus("declined");
  };

  const isPvP = gameMode === "pvp";

  // Show loading if game not started yet
  if (gameStatus !== "playing" && gameStatus !== "finished") {
    return (
      <div className="flex-1 bg-background flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Loading game...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 bg-background flex items-center justify-center overflow-hidden h-full">
      <div className="w-full max-w-[min(90vw,500px)] flex flex-col h-full py-4 gap-4">
        {/* Header Area: Banner or Info */}
        <div className="flex-shrink-0 w-full min-h-[60px] flex flex-col justify-end">
          <GameResultBanner />
          <GameInfo />
        </div>

        {/* Board Area: Flexible */}
        <div className="flex-1 w-full min-h-0 flex flex-col items-center justify-center relative">
          <Board onColumnClick={handleColumnClick} />
          {rematchStatus === "received" && (
            <RematchOverlay
              onAccept={handleAcceptRematch}
              onDecline={handleDeclineRematch}
              opponentName={opponent || "Opponent"}
            />
          )}
        </div>

        {/* Footer Area: Fixed height controls */}
        <div className="flex-shrink-0 w-full flex flex-col gap-2 pb-safe">
          <GameControls
            onSurrender={handleSurrender}
            isPlaying={gameStatus === "playing"}
          />
          <GameEndActions
            onPlayAgain={handlePlayAgain}
            onGoHome={handleGoHome}
          />
        </div>
      </div>
    </div>
  );
};

export default GamePage;
