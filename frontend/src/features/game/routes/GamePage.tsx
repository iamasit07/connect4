import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Board } from '../components/Board';
import { GameInfo } from '../components/GameInfo';
import { GameControls } from '../components/GameControls';
import { GameResultBanner } from '../components/GameResultBanner';
import { GameEndActions } from '../components/GameEndActions';
import { useGameSocket } from '../hooks/useGameSocket';
import { useGameStore } from '../store/gameStore';
import { toast } from 'sonner';

const GamePage = () => {
  const { gameId } = useParams<{ gameId: string }>();
  const navigate = useNavigate();
  const { makeMove, surrender, disconnect, sendMessage } = useGameSocket();
  const { gameStatus, resetGame, setRematchStatus, gameMode, botDifficulty, gameId: storeGameId } = useGameStore();

  // Verify game ID matches
  useEffect(() => {
    if (gameStatus === 'playing' && storeGameId && storeGameId !== gameId) {
      toast.error('Game ID mismatch');
      navigate('/play');
    }
  }, [gameStatus, storeGameId, gameId, navigate]);

  const handleColumnClick = (col: number) => {
    makeMove(col);
  };

  const handleSurrender = () => {
    surrender();
  };

  const handlePlayAgain = () => {
    if (gameMode === 'bot' && botDifficulty) {
      resetGame();
      navigate('/play/bot', { state: { difficulty: botDifficulty } });
    } else {
      resetGame();
      navigate('/play/queue', { state: { from: `/game/${gameId}` } });
    }
  };

  const handleGoHome = () => {
    disconnect();
    resetGame();
    navigate('/dashboard');
  };

  const handleSendRematch = () => {
    sendMessage({ type: 'rematch_request' } as any);
    setRematchStatus('sent');
    toast.info('Rematch request sent!');
  };

  const handleAcceptRematch = () => {
    sendMessage({ type: 'rematch_accepted' } as any);
    setRematchStatus('accepted');
  };

  const handleDeclineRematch = () => {
    sendMessage({ type: 'rematch_declined' } as any);
    setRematchStatus('declined');
  };

  // Show loading if game not started yet
  if (gameStatus !== 'playing' && gameStatus !== 'finished') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Loading game...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center p-2 sm:p-4">
      <GameResultBanner />
      <GameInfo />
      <Board onColumnClick={handleColumnClick} />
      <GameControls 
        onSurrender={handleSurrender} 
        isPlaying={gameStatus === 'playing'} 
      />
      <GameEndActions
        onPlayAgain={handlePlayAgain}
        onGoHome={handleGoHome}
      />
    </div>
  );
};

export default GamePage;
