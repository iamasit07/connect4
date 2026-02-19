import { motion, AnimatePresence } from 'framer-motion';
import { Trophy, Frown, Handshake, RotateCcw, Home } from 'lucide-react';
import { useGameStore } from '../store/gameStore';
import { Button } from '@/components/ui/button';
import { useEffect } from 'react';
import { fireWinConfetti, fireDrawConfetti } from '@/lib/confetti';
import { RematchRequest } from './RematchRequest';
import { Disk } from './Disk';

interface GameOverModalProps {
  onPlayAgain: () => void;
  onGoHome: () => void;
  onSendRematch?: () => void;
  onAcceptRematch?: () => void;
  onDeclineRematch?: () => void;
}

export const GameOverModal = ({ 
  onPlayAgain, 
  onGoHome,
  onSendRematch,
  onAcceptRematch,
  onDeclineRematch,
}: GameOverModalProps) => {
  const { 
    winner, 
    winReason, 
    gameStatus, 
    myPlayer, 
    opponent, 
    gameMode,
    rematchStatus,
    isSpectator,
    spectatorPlayer1, 
    spectatorPlayer2,
  } = useGameStore();
  
  const isWin = winner === 'You' || (myPlayer === 1 && winner === 'Player 1') || (myPlayer === 2 && winner === 'Player 2');
  const isDraw = winReason === 'draw' || winner === 'Draw';
  const isLoss = !isWin && !isDraw;
  const isPvP = gameMode === 'pvp';

  // Spectator helpers
  const spectatorWinnerIsP1 = isSpectator && (winner === spectatorPlayer1 || winner === 'Player 1');
  const spectatorWinnerIsP2 = isSpectator && (winner === spectatorPlayer2 || winner === 'Player 2');

  useEffect(() => {
    if (gameStatus === 'finished' && !isSpectator) {
      if (isWin) {
        fireWinConfetti();
      } else if (isDraw) {
        fireDrawConfetti();
      }
    } else if (gameStatus === 'finished' && isSpectator) {
        if (winner) fireWinConfetti();
        else fireDrawConfetti();
    }
  }, [gameStatus, isWin, isDraw, isSpectator, winner]);

  if (gameStatus !== 'finished') return null;

  const getIcon = () => {
    if (isSpectator) {
        if (spectatorWinnerIsP1) return <Trophy className="w-16 h-16 text-red-500" />;
        if (spectatorWinnerIsP2) return <Trophy className="w-16 h-16 text-yellow-500" />;
        if (isDraw) return <Handshake className="w-16 h-16 text-muted-foreground" />;
        return <Trophy className="w-16 h-16 text-primary" />;
    }
    if (isWin) return <Trophy className="w-16 h-16 text-yellow-500" />;
    if (isDraw) return <Handshake className="w-16 h-16 text-muted-foreground" />;
    return <Frown className="w-16 h-16 text-destructive" />;
  };

  const getTitle = () => {
    if (isSpectator) return isDraw ? 'Match Drawn' : 'Match Finished';
    if (isWin) return 'Victory!';
    if (isDraw) return "It's a Draw!";
    return 'Defeat';
  };

  const getMessage = () => {
    if (isSpectator) {
      if (isDraw) return 'The match ended in a draw.';
      if (spectatorWinnerIsP1) return `${spectatorPlayer1 || 'Player 1'} (Red) wins!`;
      if (spectatorWinnerIsP2) return `${spectatorPlayer2 || 'Player 2'} (Yellow) wins!`;
      return `${winner} wins!`;
    }
    if (isWin) {
      switch (winReason) {
        case 'connect4': return 'You connected four! Amazing!';
        case 'timeout': return 'Opponent ran out of time!';
        case 'surrender': return 'Opponent surrendered!';
        case 'disconnect': return 'Opponent disconnected!';
        default: return 'Congratulations on your victory!';
      }
    }
    if (isDraw) return 'A hard-fought battle ends in a tie!';
    switch (winReason) {
      case 'connect4': return `${opponent || 'Opponent'} connected four!`;
      case 'timeout': return 'You ran out of time!';
      case 'surrender': return 'You surrendered the game.';
      default: return 'Better luck next time!';
    }
  };

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4"
      >
        <motion.div
          initial={{ scale: 0.8, opacity: 0, y: 20 }}
          animate={{ scale: 1, opacity: 1, y: 0 }}
          exit={{ scale: 0.8, opacity: 0, y: 20 }}
          transition={{ type: 'spring', stiffness: 300, damping: 25 }}
          className={`
            bg-card rounded-2xl p-6 sm:p-8 max-w-md w-full text-center shadow-2xl
            ${(isWin && !isSpectator) || (isSpectator && !isDraw) ? 'ring-4 ring-yellow-500/50' : ''}
            ${isDraw ? 'ring-4 ring-muted/50' : ''}
            ${isLoss && !isSpectator ? 'ring-4 ring-destructive/50' : ''}
          `}
        >
          <motion.div
            initial={{ scale: 0, rotate: -180 }}
            animate={{ scale: 1, rotate: 0 }}
            transition={{ delay: 0.2, type: 'spring', stiffness: 200 }}
            className="flex justify-center mb-4"
          >
            {getIcon()}
          </motion.div>

          <motion.h2
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
            className={`
              text-2xl sm:text-3xl font-bold mb-2
              ${(isWin && !isSpectator) || (isSpectator && !isDraw) ? 'text-yellow-500' : ''}
              ${isDraw ? 'text-muted-foreground' : ''}
              ${isLoss && !isSpectator ? 'text-destructive' : ''}
            `}
          >
            {getTitle()}
          </motion.h2>

          {/* Winner/Draw details */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.4 }}
            className="mb-6 sm:mb-8 flex flex-col items-center gap-3"
          >
            {isDraw ? (
              <div className="flex flex-col items-center gap-2">
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 sm:w-12 sm:h-12 relative">
                    <Disk player={1} />
                  </div>
                  <span className="text-muted-foreground font-bold">VS</span>
                  <div className="w-10 h-10 sm:w-12 sm:h-12 relative">
                    <Disk player={2} />
                  </div>
                </div>
                <p className="text-muted-foreground font-medium text-sm sm:text-base">
                  {getMessage()}
                </p>
              </div>
            ) : (
              <div className="flex flex-col items-center gap-2">
                <span className="text-muted-foreground text-xs sm:text-sm uppercase tracking-wider font-semibold">
                  Winner
                </span>
                <div className="w-12 h-12 sm:w-16 sm:h-16 relative">
                  <Disk player={
                    isSpectator 
                      ? (spectatorWinnerIsP1 ? 1 : 2)
                      : (isWin ? myPlayer! : (myPlayer === 1 ? 2 : 1))
                  } />
                </div>
                <div className="flex flex-col items-center">
                  <span className={`text-lg sm:text-xl font-bold ${
                    isSpectator 
                      ? (spectatorWinnerIsP1 ? 'text-red-500' : 'text-yellow-500')
                      : (isWin ? 'text-yellow-500' : 'text-destructive')
                  }`}>
                    {isSpectator 
                      ? (spectatorWinnerIsP1 ? spectatorPlayer1 || 'Player 1' : spectatorPlayer2 || 'Player 2')
                      : (isWin ? 'You' : (opponent || 'Opponent'))}
                  </span>
                  {winReason && (
                    <span className="text-[10px] sm:text-xs text-muted-foreground">
                      by {winReason === 'connect4' ? 'connect 4' : winReason}
                    </span>
                  )}
                </div>
              </div>
            )}
          </motion.div>

          {/* Rematch section for PvP games */}
          {isPvP && !isSpectator && onSendRematch && onAcceptRematch && onDeclineRematch && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.5 }}
              className="mb-6"
            >
              <RematchRequest
                onSendRequest={onSendRematch}
                onAcceptRequest={onAcceptRematch}
                onDeclineRequest={onDeclineRematch}
                rematchStatus={rematchStatus}
                opponentName={opponent || 'Opponent'}
              />
            </motion.div>
          )}

          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6 }}
            className="flex flex-col sm:flex-row gap-3"
          >
            {!isPvP && (
              <Button
                onClick={onPlayAgain}
                className="flex-1 gap-2"
                size="lg"
              >
                <RotateCcw className="w-4 h-4" />
                Play Again
              </Button>
            )}
            <Button
              onClick={onGoHome}
              variant={isPvP ? 'default' : 'outline'}
              className="flex-1 gap-2"
              size="lg"
            >
              <Home className="w-4 h-4" />
              {isSpectator ? 'Leave' : 'Home'}
            </Button>
          </motion.div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
};
