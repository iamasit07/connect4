import { Button } from '@/components/ui/button';
import { Home, RotateCcw } from 'lucide-react';
import { useGameStore } from '../store/gameStore';

interface GameEndActionsProps {
  onPlayAgain: () => void;
  onGoHome: () => void;
}

export const GameEndActions = ({ onPlayAgain, onGoHome }: GameEndActionsProps) => {
  const { gameStatus } = useGameStore();

  if (gameStatus !== 'finished') return null;

  return (
    <div className="w-full max-w-[min(90vw,500px)] mx-auto mt-6 flex gap-3">
      <Button
        onClick={onGoHome}
        variant="outline"
        className="flex-1 gap-2"
        size="lg"
      >
        <Home className="w-4 h-4" />
        Go Home
      </Button>
      <Button
        onClick={onPlayAgain}
        className="flex-1 gap-2"
        size="lg"
      >
        <RotateCcw className="w-4 h-4" />
        Play Again
      </Button>
    </div>
  );
};
