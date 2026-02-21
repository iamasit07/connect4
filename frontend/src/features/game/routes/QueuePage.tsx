import { useEffect, useCallback, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { QueueScreen } from '../components/QueueScreen';
import { useGameSocket } from '../hooks/useGameSocket';
import { useGameStore } from '../store/gameStore';
import type { BotDifficulty } from '../types';

const QueuePage = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const gameFound = useRef(false);

  const onGameStart = useCallback((gameId: string) => {
    gameFound.current = true;
    navigate(`/game/${gameId}`);
  }, [navigate]);

  const { findMatch, disconnect, sendMessage } = useGameSocket(onGameStart);
  const { resetGame } = useGameStore();
  const previousRoute = location.state?.from || '/play';

  useEffect(() => {
    findMatch('pvp');
    
    return () => {
      if (!gameFound.current) {
        resetGame();
      }
    };
  }, []);

  const handleCancel = () => {
    sendMessage({ type: 'cancel_search' });
    disconnect();
    resetGame();
    navigate(previousRoute);
  };

  const handlePlayBot = (difficulty: BotDifficulty) => {
    disconnect();
    resetGame();
    navigate('/play/bot', { state: { difficulty } });
  };

  return (
    <QueueScreen 
      onCancel={handleCancel}
      onPlayBot={handlePlayBot}
    />
  );
};

export default QueuePage;
