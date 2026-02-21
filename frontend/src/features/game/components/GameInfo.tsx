import { useEffect } from "react";
import { motion } from "framer-motion";
import { Wifi, WifiOff, Crown } from "lucide-react";
import { fireWinConfetti } from "@/lib/confetti";
import { useGameStore } from "../store/gameStore";
import { Disk } from "./Disk";
import { SpectatorBadge } from "./SpectatorBadge";

export const GameInfo = () => {
  const {
    opponent,
    myPlayer,
    currentTurn,
    gameStatus,
    connectionStatus,
    isSpectator,
    spectatorCount,
    spectatorPlayer1, 
    spectatorPlayer2,
    winner,
    winReason,
  } = useGameStore();

  if (gameStatus !== "playing" && gameStatus !== "finished") return null;

  const isDraw = winner === "draw" || winReason === "draw";

  // Player 1 is ALWAYS Left Side, Player 2 is ALWAYS Right Side
  const isP1Turn = currentTurn === 1 && !winner;
  const isP2Turn = currentTurn === 2 && !winner;

  const player1Name = isSpectator ? (spectatorPlayer1 || "Player 1") : (myPlayer === 1 ? "You" : (opponent || "Opponent"));
  const player2Name = isSpectator ? (spectatorPlayer2 || "Player 2") : (myPlayer === 2 ? "You" : (opponent || "Opponent"));

  const p1Won = !isDraw && !!winner && (
    isSpectator ? (winner === spectatorPlayer1 || winner === 'Player 1') :
    (winner === 'Player 1' || (myPlayer === 1 ? (winner !== opponent && winner !== 'Player 2') : winner === opponent))
  );

  const p2Won = !isDraw && !!winner && (
    isSpectator ? (winner === spectatorPlayer2 || winner === 'Player 2') :
    (winner === 'Player 2' || (myPlayer === 2 ? (winner !== opponent && winner !== 'Player 1') : winner === opponent))
  );

  // Helper to render turn status text
  const TurnText = ({ active }: { active: boolean }) => {
    if (!active) return null;
    return (
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        className="text-[10px] sm:text-xs font-semibold text-primary animate-pulse"
      >
        Thinking...
      </motion.div>
    );
  };

  useEffect(() => {
    if (gameStatus !== "finished") return;
    
    // Fire confetti if "You" won
    const iWon = (myPlayer === 1 && p1Won) || (myPlayer === 2 && p2Won);
    if (iWon && !isSpectator) {
      fireWinConfetti();
    }
  }, [gameStatus, p1Won, p2Won, myPlayer, isSpectator]);

  return (
    <div className="w-full max-w-[min(95vw,600px)] mx-auto mb-2 sm:mb-4 px-2 shrink-0">
      {/* Spectator badge row */}
      {(spectatorCount > 0 || isSpectator) && (
        <div className="flex items-center justify-center mb-2">
          <SpectatorBadge count={spectatorCount} isSpectator={isSpectator} />
        </div>
      )}

      <div className="flex items-center justify-between gap-2 sm:gap-4 h-16 sm:h-20">
        {/* Left Side (Player 1 / Red) */}
        <motion.div
          animate={{
            scale: isP1Turn ? 1.02 : 1,
            boxShadow: isP1Turn ? "0 0 15px hsl(var(--primary) / 0.2)" : "none",
            opacity: gameStatus === 'finished' && !p1Won && !isDraw ? 0.6 : 1,
            borderColor: isP1Turn ? "hsl(var(--primary) / 0.5)" : "transparent"
          }}
          className={`
            flex-1 flex items-center gap-2 sm:gap-3 p-2 sm:p-3 rounded-xl border
            ${isP1Turn ? "bg-primary/5 ring-1 ring-primary/50" : "bg-card/50"}
            transition-all duration-300
            ${p1Won ? "ring-2! ring-red-500! bg-red-500/10!" : ""}
          `}
        >
          <div className="w-8 h-8 sm:w-10 sm:h-10 relative shrink-0">
            <Disk player={1} />
            {p1Won && (
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                className={`absolute -top-2 -left-2 bg-background rounded-full p-0.5 shadow-sm border border-red-500/50`}
              >
                <Crown className={`w-3 h-3 sm:w-4 sm:h-4 text-red-500 fill-red-500`} />
              </motion.div>
            )}
          </div>
          <div className="flex-1 min-w-0 flex flex-col justify-center">
            <div className="flex items-center gap-2">
              <p className="font-semibold text-sm sm:text-base truncate">
                {player1Name}
              </p>
              {p1Won && (
                <span className={`text-[10px] text-white px-1.5 py-0.5 rounded-full font-bold shadow-sm bg-red-500`}>
                  WIN
                </span>
              )}
            </div>
            <div className="flex items-center gap-2 h-4">
               <p className="text-[10px] sm:text-xs text-muted-foreground capitalize leading-none">
                Red
              </p>
              {!p1Won && isP1Turn && <TurnText active={true} />}
            </div>
          </div>
        </motion.div>

        {/* Center (VS / Result) */}
        <div className="flex flex-col items-center justify-center gap-0.5 min-w-10 sm:min-w-12">
          {isDraw ? (
             <motion.div 
               initial={{ scale: 0 }}
               animate={{ scale: 1 }}
               className="bg-muted/80 backdrop-blur-sm text-muted-foreground px-2 py-1 rounded text-[10px] font-black border tracking-wider"
             >
               DRAW
             </motion.div>
          ) : (
            <>
              <span className="text-base sm:text-xl font-black text-muted-foreground/30 leading-none">VS</span>
              <div
                className={`flex items-center gap-1 text-[10px] ${
                  connectionStatus === "connected"
                    ? "text-green-500"
                    : "text-destructive"
                }`}
              >
                {connectionStatus === "connected" ? (
                  <Wifi className="w-3 h-3" />
                ) : (
                  <WifiOff className="w-3 h-3" />
                )}
              </div>
            </>
          )}
        </div>

        {/* Right Side (Player 2 / Yellow) */}
        <motion.div
          animate={{
            scale: isP2Turn ? 1.02 : 1,
            boxShadow: isP2Turn ? "0 0 15px hsl(var(--primary) / 0.2)" : "none",
            opacity: gameStatus === 'finished' && !p2Won && !isDraw ? 0.6 : 1,
            borderColor: isP2Turn ? "hsl(var(--primary) / 0.5)" : "transparent"
          }}
          className={`
            flex-1 flex items-center gap-2 sm:gap-3 p-2 sm:p-3 rounded-xl border flex-row-reverse text-right
            ${isP2Turn ? "bg-primary/5 ring-1 ring-primary/50" : "bg-card/50"}
            transition-all duration-300
            ${p2Won ? "ring-2! ring-yellow-500! bg-yellow-500/10!" : ""}
          `}
        >
          <div className="w-8 h-8 sm:w-10 sm:h-10 relative shrink-0">
            <Disk player={2} />
             {p2Won && (
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                className={`absolute -top-2 -right-2 bg-background rounded-full p-0.5 shadow-sm border border-yellow-500/50`}
              >
                <Crown className={`w-3 h-3 sm:w-4 sm:h-4 text-yellow-500 fill-yellow-500`} />
              </motion.div>
            )}
          </div>
          <div className="flex-1 min-w-0 flex flex-col justify-center">
            <div className="flex items-center gap-2 justify-end">
              {p2Won && (
                <span className={`text-[10px] text-white px-1.5 py-0.5 rounded-full font-bold shadow-sm bg-yellow-500`}>
                  WIN
                </span>
              )}
               <p className="font-semibold text-sm sm:text-base truncate">
                {player2Name}
              </p>
            </div>
            <div className="flex items-center gap-2 h-4 justify-end">
               {!p2Won && isP2Turn && <TurnText active={true} />}
               <p className="text-[10px] sm:text-xs text-muted-foreground capitalize leading-none">
                Yellow
              </p>
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  );
};

