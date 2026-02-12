import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { useGameSocket } from '../hooks/useGameSocket';
import { useGameStore } from '../store/gameStore';
import type { BotDifficulty } from '../types';

const BotLoadingPage = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { findMatch, disconnect } = useGameSocket((gameId) => {
    // Navigate to game when it starts
    console.log('[BotLoadingPage] Game started with ID:', gameId, '- Navigating to /game/' + gameId);
    navigate(`/game/${gameId}`);
  });
  const { resetGame } = useGameStore();
  const difficulty = (location.state as { difficulty?: BotDifficulty })?.difficulty;

  // Auto-start bot game when component mounts
  useEffect(() => {
    if (difficulty) {
      findMatch('bot', difficulty);
    } else {
      // No difficulty specified, go back to play
      navigate('/play');
    }
    // Don't disconnect on unmount - WebSocket should persist to GamePage
  }, [difficulty, findMatch, navigate]);

  const handleCancel = () => {
    disconnect();
    resetGame();
    navigate('/play');
  };

  return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center p-4">
      <div className="text-center space-y-6">
        <Loader2 className="w-16 h-16 animate-spin text-primary mx-auto" />
        <div>
          <h2 className="text-2xl font-bold mb-2">Starting Game...</h2>
          <p className="text-muted-foreground">
            Preparing your match against the {difficulty} bot
          </p>
        </div>
        <button
          onClick={handleCancel}
          className="text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          Cancel
        </button>
      </div>
    </div>
  );
};

export default BotLoadingPage;
