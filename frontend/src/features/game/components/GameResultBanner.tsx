import { useEffect } from 'react';
import { motion } from 'framer-motion';
import { Trophy, XCircle, Minus } from 'lucide-react';
import { useGameStore } from '../store/gameStore';
import { fireWinConfetti, fireDrawConfetti } from '@/lib/confetti';

export const GameResultBanner = () => {
  const { gameStatus, winner, winReason, opponent } = useGameStore();

  const isDraw = winner === 'draw';
  const isWinner = winner !== 'draw' && winner !== opponent;

  // Fire confetti on game end
  useEffect(() => {
    if (gameStatus !== 'finished') return;
    if (isWinner) {
      fireWinConfetti();
    }
  }, [gameStatus, isWinner, isDraw]);

  if (gameStatus !== 'finished') return null;

  const getIcon = () => {
    if (isDraw) return <Minus className="w-8 h-8" />;
    if (isWinner) return <Trophy className="w-8 h-8" />;
    return <XCircle className="w-8 h-8" />;
  };

  const getMessage = () => {
    if (isDraw) return 'Draw!';
    if (isWinner) return 'You Won!';
    if (winReason === 'surrender') return 'You Surrendered';
    return 'You Lost!';
  };

  const getSubMessage = () => {
    if (isDraw) return 'A hard-fought battle ends in a tie!';
    if (isWinner) {
      switch (winReason) {
        case 'connect4': return 'You connected four!';
        case 'timeout': return `${opponent} ran out of time!`;
        case 'surrender': return `${opponent} surrendered!`;
        default: return 'Victory!';
      }
    }
    // Lost
    switch (winReason) {
      case 'connect4': return `${opponent} connected four!`;
      case 'timeout': return 'You ran out of time!';
      case 'surrender': return 'You surrendered the game.';
      default: return 'Better luck next time!';
    }
  };

  const bgColor = isDraw 
    ? 'bg-muted' 
    : isWinner 
      ? 'bg-gradient-to-r from-green-500/20 to-emerald-500/20'
      : 'bg-gradient-to-r from-red-500/20 to-rose-500/20';

  const textColor = isDraw
    ? 'text-foreground'
    : isWinner
      ? 'text-green-600 dark:text-green-400'
      : 'text-red-600 dark:text-red-400';

  return (
    <motion.div
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      className={`w-full max-w-[min(90vw,500px)] mx-auto mb-4 rounded-xl p-4 ${bgColor} border-2 ${
        isDraw ? 'border-muted-foreground/20' : isWinner ? 'border-green-500/50' : 'border-red-500/50'
      }`}
    >
      <div className="flex items-center gap-4">
        <div className={textColor}>
          {getIcon()}
        </div>
        <div className="flex-1">
          <h2 className={`text-2xl font-bold ${textColor}`}>
            {getMessage()}
          </h2>
          <p className="text-sm text-muted-foreground mt-1">
            {getSubMessage()}
          </p>
        </div>
      </div>
    </motion.div>
  );
};
