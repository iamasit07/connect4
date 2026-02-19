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
    isMyTurn,
    isSpectator,
    spectatorCount,
    spectatorPlayer1, 
    spectatorPlayer2,
    winner,
    winReason,
  } = useGameStore();

  if (gameStatus !== "playing" && gameStatus !== "finished") return null;

  const myColor = myPlayer === 1 ? "red" : "yellow";
  const opponentColor = myPlayer === 1 ? "yellow" : "red";
  const myTurn = isMyTurn();

  const isDraw = winner === "draw" || winReason === "draw";
  const leftSideWon = !isDraw && winner && (isSpectator ? winner === spectatorPlayer1 || winner === 'Player 1' : winner !== opponent);
  const rightSideWon = !isDraw && winner && (isSpectator ? winner === spectatorPlayer2 || winner === 'Player 2' : winner === opponent);

  const isP1Turn = (myTurn || (isSpectator && currentTurn === 1)) && !winner;
  const isP2Turn = ((!myTurn && currentTurn !== myPlayer && !isSpectator) || (isSpectator && currentTurn === 2)) && !winner;

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
    
    if (leftSideWon && !isSpectator) {
      fireWinConfetti();
    }
  }, [gameStatus, leftSideWon, isSpectator]);

  return (
    <div className="w-full max-w-[min(95vw,600px)] mx-auto mb-2 sm:mb-4 px-2 flex-shrink-0">
      {/* Spectator badge row */}
      {(spectatorCount > 0 || isSpectator) && (
        <div className="flex items-center justify-center mb-2">
          <SpectatorBadge count={spectatorCount} isSpectator={isSpectator} />
        </div>
      )}

      <div className="flex items-center justify-between gap-2 sm:gap-4 h-16 sm:h-20">
        {/* Left Side (You / Player 1) */}
        <motion.div
          animate={{
            scale: isP1Turn ? 1.02 : 1,
            boxShadow: isP1Turn ? "0 0 15px hsl(var(--primary) / 0.2)" : "none",
            opacity: gameStatus === 'finished' && !leftSideWon && !isDraw ? 0.6 : 1,
            borderColor: isP1Turn ? "hsl(var(--primary) / 0.5)" : "transparent"
          }}
          className={`
            flex-1 flex items-center gap-2 sm:gap-3 p-2 sm:p-3 rounded-xl border
            ${isP1Turn ? "bg-primary/5 ring-1 ring-primary/50" : "bg-card/50"}
            transition-all duration-300
            ${leftSideWon 
              ? (isSpectator ? "red" : myColor) === "red" 
                ? "!ring-2 !ring-red-500 !bg-red-500/10" 
                : "!ring-2 !ring-yellow-500 !bg-yellow-500/10"
              : ""}
          `}
        >
          <div className="w-8 h-8 sm:w-10 sm:h-10 relative flex-shrink-0">
            <Disk player={1} />
            {leftSideWon && (
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                className={`absolute -top-2 -left-2 bg-background rounded-full p-0.5 shadow-sm border ${
                  (isSpectator ? "red" : myColor) === "red" ? "border-red-500/50" : "border-yellow-500/50"
                }`}
              >
                <Crown className={`w-3 h-3 sm:w-4 sm:h-4 ${
                  (isSpectator ? "red" : myColor) === "red" ? "text-red-500 fill-red-500" : "text-yellow-500 fill-yellow-500"
                }`} />
              </motion.div>
            )}
          </div>
          <div className="flex-1 min-w-0 flex flex-col justify-center">
            <div className="flex items-center gap-2">
              <p className="font-semibold text-sm sm:text-base truncate">
                {isSpectator ? (spectatorPlayer1 || "Player 1") : "You"}
              </p>
              {leftSideWon && (
                <span className={`text-[10px] text-white px-1.5 py-0.5 rounded-full font-bold shadow-sm ${
                  (isSpectator ? "red" : myColor) === "red" ? "bg-red-500" : "bg-yellow-500"
                }`}>
                  WIN
                </span>
              )}
            </div>
            <div className="flex items-center gap-2 h-4">
               <p className="text-[10px] sm:text-xs text-muted-foreground capitalize leading-none">
                {isSpectator ? "Red" : myColor}
              </p>
              {!leftSideWon && isP1Turn && <TurnText active={true} />}
            </div>
          </div>
        </motion.div>

        {/* Center (VS / Result) */}
        <div className="flex flex-col items-center justify-center gap-0.5 min-w-[2.5rem] sm:min-w-[3rem]">
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

        {/* Right Side (Opponent / Player 2) */}
        <motion.div
          animate={{
            scale: isP2Turn ? 1.02 : 1,
            boxShadow: isP2Turn ? "0 0 15px hsl(var(--primary) / 0.2)" : "none",
            opacity: gameStatus === 'finished' && !rightSideWon && !isDraw ? 0.6 : 1,
            borderColor: isP2Turn ? "hsl(var(--primary) / 0.5)" : "transparent"
          }}
          className={`
            flex-1 flex items-center gap-2 sm:gap-3 p-2 sm:p-3 rounded-xl border flex-row-reverse text-right
            ${isP2Turn ? "bg-primary/5 ring-1 ring-primary/50" : "bg-card/50"}
            transition-all duration-300
            ${rightSideWon 
              ? (isSpectator ? "yellow" : opponentColor) === "red"
                ? "!ring-2 !ring-red-500 !bg-red-500/10"
                : "!ring-2 !ring-yellow-500 !bg-yellow-500/10"
              : ""}
          `}
        >
          <div className="w-8 h-8 sm:w-10 sm:h-10 relative flex-shrink-0">
            <Disk player={2} />
             {rightSideWon && (
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                className={`absolute -top-2 -right-2 bg-background rounded-full p-0.5 shadow-sm border ${
                  (isSpectator ? "yellow" : opponentColor) === "red" ? "border-red-500/50" : "border-yellow-500/50"
                }`}
              >
                <Crown className={`w-3 h-3 sm:w-4 sm:h-4 ${
                  (isSpectator ? "yellow" : opponentColor) === "red" ? "text-red-500 fill-red-500" : "text-yellow-500 fill-yellow-500"
                }`} />
              </motion.div>
            )}
          </div>
          <div className="flex-1 min-w-0 flex flex-col justify-center">
            <div className="flex items-center gap-2 justify-end">
              {rightSideWon && (
                <span className={`text-[10px] text-white px-1.5 py-0.5 rounded-full font-bold shadow-sm ${
                  (isSpectator ? "yellow" : opponentColor) === "red" ? "bg-red-500" : "bg-yellow-500"
                }`}>
                  WIN
                </span>
              )}
               <p className="font-semibold text-sm sm:text-base truncate">
                {isSpectator ? (spectatorPlayer2 || "Player 2") : opponent || "Opponent"}
              </p>
            </div>
            <div className="flex items-center gap-2 h-4 justify-end">
               {!rightSideWon && isP2Turn && <TurnText active={true} />}
               <p className="text-[10px] sm:text-xs text-muted-foreground capitalize leading-none">
                {isSpectator ? "Yellow" : opponentColor}
              </p>
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  );
};
