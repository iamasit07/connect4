import { useEffect, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Board } from "../components/Board";
import { GameInfo } from "../components/GameInfo";
import { GameControls } from "../components/GameControls";
import { GameEndActions } from "../components/GameEndActions";
import { RematchOverlay } from "../components/RematchRequest";
import { useGameSocket } from "../hooks/useGameSocket";
import { useGameStore } from "../store/gameStore";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Home } from "lucide-react";
import { DisconnectionTimer } from "../components/DisconnectionTimer";
import { useAuthStore } from "@/features/auth/store/authStore";
import { gameService } from "../services/gameService";

const GamePage = () => {
  const { gameId } = useParams<{ gameId: string }>();
  const navigate = useNavigate();
  const { user, isLoading: isAuthLoading } = useAuthStore();

  // Handle new game starting from a rematch
  const onGameStart = useCallback(
    (newGameId: string) => {
      navigate(`/game/${newGameId}`, { replace: true });
    },
    [navigate],
  );

  const { makeMove, surrender, disconnect, sendMessage, leaveSpectate, connect } =
    useGameSocket(onGameStart);
  const {
    gameStatus,
    resetGame,
    setRematchStatus,
    gameMode,
    botDifficulty,
    rematchStatus,
    opponent,
    gameId: storeGameId,
    isSpectator,
    connectionStatus,
    loadFinishedGame,
  } = useGameStore();

  useEffect(() => {
    if (storeGameId === gameId && (gameStatus === "playing" || gameStatus === "finished")) {
      return;
    }

    const fetchGame = async () => {
      if (!gameId || !user) return;
      try {
        const details = await gameService.getGameDetails(gameId);
        
        // If game is finished
        if (details.FinishedAt && details.FinishedAt !== "0001-01-01T00:00:00Z") {
           loadFinishedGame({
             gameId: details.GameID,
             board: details.board_state,
             player1: { username: details.Player1Username, id: details.Player1ID },
             player2: { username: details.Player2Username, id: details.Player2ID },
             winner: details.WinnerUsername || null,
             reason: details.Reason,
             myUserId: parseInt(user.id),
           });
           toast.info("This game has already ended.");
        } 
      } catch (error) {
        console.error("Failed to fetch game details:", error);
      }
    };

    fetchGame();
  }, [gameId, storeGameId, gameStatus, user, loadFinishedGame]);

  useEffect(() => {
    if (gameStatus === "finished" && gameMode === "bot") return;
    if (connectionStatus === "disconnected" || connectionStatus === "error") {
      connect();
    }
    return () => {
      disconnect();
    };
  }, [connect, disconnect, connectionStatus, gameStatus, gameMode]);

  // Route Guard: Redirect to active game if trying to access another
  useEffect(() => {
    if (isAuthLoading || !user) return;

    if (user.activeGameId && user.activeGameId !== gameId && !isSpectator) {
      toast.warning("You have an active game in progress!");
      navigate(`/game/${user.activeGameId}`, { replace: true });
    }
  }, [user, gameId, isSpectator, navigate, isAuthLoading]);

  // Sync Store Game with URL
  useEffect(() => {
    if (gameStatus === "playing" && storeGameId && storeGameId !== gameId) {
      navigate(`/game/${storeGameId}`, { replace: true });
    }
  }, [gameStatus, storeGameId, gameId, navigate]);

  const handleColumnClick = (col: number) => {
    if (isSpectator) return;
    makeMove(col);
  };

  const handleSurrender = () => {
    surrender();
  };

  const handleLeaveSpectate = () => {
    if (storeGameId) {
      leaveSpectate(storeGameId);
    }
    disconnect();
    resetGame();
    navigate("/dashboard");
  };

  const handlePlayAgain = () => {
    if (gameMode === "bot" && botDifficulty) {
      disconnect();
      resetGame();
      navigate("/play/bot", { state: { difficulty: botDifficulty } });
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
        <DisconnectionTimer />
        {/* Header Area: Banner or Info */}
        <div className="flex-shrink-0 w-full min-h-[60px] flex flex-col justify-end">
          <GameInfo />
        </div>

        {/* Board Area: Flexible */}
        <div className="flex-1 w-full min-h-0 flex flex-col items-center justify-center relative">
          <Board onColumnClick={handleColumnClick} />
          
          {/* Reconnecting Overlay */}
          {(connectionStatus === 'connecting' || connectionStatus === 'error') && (
            <div className="absolute inset-0 z-40 bg-background/50 backdrop-blur-[1px] flex items-center justify-center">
              <div className="bg-card border shadow-lg rounded-full px-6 py-3 flex items-center gap-3">
                <div className="animate-spin rounded-full h-4 w-4 border-2 border-primary border-t-transparent"></div>
                <span className="font-medium text-sm">
                  {connectionStatus === 'error' ? 'Retrying connection...' : 'Reconnecting...'}
                </span>
              </div>
            </div>
          )}

          {!isSpectator && rematchStatus === "received" && (
            <RematchOverlay
              onAccept={handleAcceptRematch}
              onDecline={handleDeclineRematch}
              opponentName={opponent || "Opponent"}
            />
          )}
        </div>

        {/* Footer Area: Fixed height controls */}
        <div className="flex-shrink-0 w-full flex flex-col gap-2 pb-safe">
          {isSpectator ? (
            <div className="flex justify-center mt-2 sm:mt-4">
              <Button
                variant="outline"
                size="sm"
                className="gap-2"
                onClick={handleLeaveSpectate}
              >
                <Home className="w-4 h-4" />
                Leave Spectate
              </Button>
            </div>
          ) : (
            <>
              <GameControls
                onSurrender={handleSurrender}
                isPlaying={gameStatus === "playing"}
              />
              <GameEndActions
                onPlayAgain={handlePlayAgain}
                onGoHome={handleGoHome}
              />
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default GamePage;
