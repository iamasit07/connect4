import { motion } from "framer-motion";
import { Wifi, WifiOff } from "lucide-react";
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
  } = useGameStore();

  if (gameStatus !== "playing") return null;

  const myColor = myPlayer === 1 ? "red" : "yellow";
  const opponentColor = myPlayer === 1 ? "yellow" : "red";
  const myTurn = isMyTurn();

  return (
    <div className="w-full max-w-[min(90vw,500px)] mx-auto mb-1 sm:mb-2 space-y-1 sm:space-y-2 flex-shrink-0">
      {/* Spectator badge row */}
      <div className="flex items-center justify-between">
        <SpectatorBadge count={spectatorCount} isSpectator={isSpectator} />
      </div>

      <div className="flex items-center justify-between gap-2 sm:gap-4">
        {/* My info */}
        <motion.div
          animate={{
            scale: myTurn ? 1.02 : 1,
            boxShadow: myTurn ? "0 0 20px hsl(var(--primary) / 0.3)" : "none",
          }}
          className={`
            flex-1 flex items-center gap-1.5 sm:gap-3 p-1.5 sm:p-3 rounded-xl
            ${myTurn ? "bg-primary/10 ring-2 ring-primary" : "bg-card"}
            ${isSpectator ? "opacity-75" : ""}
            transition-colors duration-300
          `}
        >
          <div className="w-6 h-6 sm:w-10 sm:h-10">
            <Disk player={myPlayer || 1} />
          </div>
          <div className="flex-1 min-w-0">
            <p className="font-semibold text-xs sm:text-base truncate">
              {isSpectator ? "Player 1" : "You"}
            </p>
            <p className="text-[10px] sm:text-xs text-muted-foreground capitalize">
              {myColor}
            </p>
          </div>
          {myTurn && !isSpectator && (
            <motion.div
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              className="text-xs sm:text-sm font-medium text-primary"
            >
              Your turn!
            </motion.div>
          )}
          {currentTurn === myPlayer && isSpectator && (
            <motion.div
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              className="text-xs sm:text-sm font-medium text-primary"
            >
              Playing...
            </motion.div>
          )}
        </motion.div>

        {/* VS */}
        <div className="flex flex-col items-center gap-1">
          <span className="text-lg font-bold text-muted-foreground">VS</span>
          <div
            className={`flex items-center gap-1 text-xs ${
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
        </div>

        {/* Opponent info */}
        <motion.div
          animate={{
            scale: !myTurn && currentTurn !== myPlayer ? 1.02 : 1,
            boxShadow:
              currentTurn !== myPlayer
                ? "0 0 20px hsl(var(--primary) / 0.3)"
                : "none",
          }}
          className={`
            flex-1 flex items-center gap-1.5 sm:gap-3 p-1.5 sm:p-3 rounded-xl
            ${currentTurn !== myPlayer ? "bg-primary/10 ring-2 ring-primary" : "bg-card"}
            ${isSpectator ? "opacity-75" : ""}
            transition-colors duration-300
          `}
        >
          <div className="w-6 h-6 sm:w-10 sm:h-10">
            <Disk player={myPlayer === 1 ? 2 : 1} />
          </div>
          <div className="flex-1 min-w-0">
            <p className="font-semibold text-xs sm:text-base truncate">
              {isSpectator ? "Player 2" : opponent || "Opponent"}
            </p>
            <p className="text-[10px] sm:text-xs text-muted-foreground capitalize">
              {opponentColor}
            </p>
          </div>
          {currentTurn !== myPlayer && (
            <motion.div
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              className="flex items-center gap-0.5 sm:gap-1 text-[10px] sm:text-sm font-medium text-muted-foreground"
            >
              Playing...
            </motion.div>
          )}
        </motion.div>
      </div>
    </div>
  );
};
