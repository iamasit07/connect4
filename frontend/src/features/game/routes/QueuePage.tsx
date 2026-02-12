import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { QueueScreen } from '../components/QueueScreen';
import { useGameSocket } from '../hooks/useGameSocket';
import { useGameStore } from '../store/gameStore';
import type { BotDifficulty } from '../types';

const QueuePage = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { findMatch, disconnect, sendMessage } = useGameSocket((gameId) => {
    navigate(`/game/${gameId}`);
  });
  const { resetGame } = useGameStore();
  const previousRoute = location.state?.from || '/play';

  useEffect(() => {
    findMatch('pvp');
  }, [findMatch]);

  const handleCancel = () => {
    sendMessage({ type: 'cancel_search' });
    disconnect();
    resetGame();
    navigate(previousRoute);
  };

  const handlePlayBot = (difficulty: BotDifficulty) => {
    disconnect();
    resetGame();
    navigate('/play', { state: { mode: 'bot', difficulty } });
  };

  return (
    <QueueScreen 
      onCancel={handleCancel}
      onPlayBot={handlePlayBot}
    />
  );
};

export default QueuePage;
