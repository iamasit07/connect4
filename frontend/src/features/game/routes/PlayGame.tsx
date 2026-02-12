import { useNavigate, useLocation } from 'react-router-dom';
import { useCallback, useEffect, useRef } from 'react';
import { Board } from '../components/Board';
import { GameInfo } from '../components/GameInfo';
import { GameControls } from '../components/GameControls';
import { GameResultBanner } from '../components/GameResultBanner';
import { GameEndActions } from '../components/GameEndActions';
import { ModeSelection } from '../components/ModeSelection';
import { QueueScreen } from '../components/QueueScreen';
import { LiveGamesList } from '../components/LiveGamesList';
import { useGameSocket } from '../hooks/useGameSocket';
import { useGameStore } from '../store/gameStore';
import { toast } from 'sonner';
import type { BotDifficulty } from '../types';

const PlayGame = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { findMatch, makeMove, surrender, disconnect, sendMessage } = useGameSocket();
  const { gameStatus, resetGame, setRematchStatus, gameMode, botDifficulty } = useGameStore();
  const hasAutoStarted = useRef(false);

  // Check if we came here with bot mode from dashboard
  useEffect(() => {
    const state = location.state as { mode?: string; difficulty?: BotDifficulty } | null;
    
    // Only auto-start once when component mounts with bot state
    if (state?.mode === 'bot' && state?.difficulty && !hasAutoStarted.current) {
      hasAutoStarted.current = true;
      findMatch('bot', state.difficulty);
      // Clear the state so it doesn't persist in history
      navigate(location.pathname, { replace: true, state: null });
    }
  }, []); // Run only on mount

  const handleSelectPvP = () => {
    // Navigate to queue page which will auto-start matchmaking
    navigate('/play/queue', { state: { from: '/play' } });
  };

  const handleSelectBot = (difficulty: BotDifficulty) => {
    findMatch('bot', difficulty);
  };

  const handleColumnClick = (col: number) => {
    makeMove(col);
  };

  const handleSurrender = () => {
    surrender();
  };

  const handlePlayAgain = () => {
    // For bot games, restart with same difficulty
    if (gameMode === 'bot' && botDifficulty) {
      resetGame();
      setTimeout(() => findMatch('bot', botDifficulty), 100);
    } else {
      resetGame();
    }
  };

  const handleGoHome = () => {
    disconnect();
    resetGame();
    navigate('/dashboard');
  };

  const handleCancelQueue = () => {
    disconnect();
    resetGame();
  };

  const handleSendRematch = useCallback(() => {
    sendMessage({ type: 'rematch_request' } as any);
    setRematchStatus('sent');
    toast.info('Rematch request sent!');
  }, [sendMessage, setRematchStatus]);

  const handleAcceptRematch = useCallback(() => {
    sendMessage({ type: 'rematch_accept' } as any);
    setRematchStatus('accepted');
  }, [sendMessage, setRematchStatus]);

  const handleDeclineRematch = useCallback(() => {
    sendMessage({ type: 'rematch_decline' } as any);
    setRematchStatus('declined');
  }, [sendMessage, setRematchStatus]);

  const handleSpectate = (gameId: string) => {
    // Connect as spectator
    toast.info('Spectator mode coming soon!');
  };

  // Show mode selection if idle
  if (gameStatus === 'idle') {
    return (
      <>
        <ModeSelection
          onSelectPvP={handleSelectPvP}
          onSelectBot={handleSelectBot}
        />
        <div className="container pb-8 -mt-8">
          <div className="max-w-md mx-auto">
            <LiveGamesList onSpectate={handleSpectate} />
          </div>
        </div>
      </>
    );
  }

  // Show queue screen if queuing
  if (gameStatus === 'queuing') {
    return (
      <QueueScreen 
        onCancel={handleCancelQueue}
        onPlayBot={handleSelectBot}
      />
    );
  }

  // Show game
  return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center p-4">
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

export default PlayGame;
