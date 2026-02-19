import { useEffect, useCallback } from 'react';
import { websocketManager } from '../services/websocketManager';
import { useGameStore } from '../store/gameStore';
import type { BotDifficulty } from '../types';

export const useGameSocket = (
  onGameStart?: (gameId: string) => void,
  onQueueTimeout?: () => void
) => {
  const setQueuing = useGameStore((state) => state.setQueuing);

  // Register onGameStart callback
  useEffect(() => {
    if (!onGameStart) return undefined;
    
    const unregister = websocketManager.onGameStart(onGameStart);
    return () => {
      unregister();
    };
  }, [onGameStart]);

  // Register onQueueTimeout callback
  useEffect(() => {
    if (!onQueueTimeout) return undefined;

    const unregister = websocketManager.onMessage((message) => {
      if (message.type === 'queue_timeout') {
        onQueueTimeout();
      }
    });
    
    return () => {
      unregister();
    };
  }, [onQueueTimeout]);

  const findMatch = useCallback(async (mode: 'pvp' | 'bot', difficulty?: BotDifficulty) => {
    await websocketManager.connect(); 

    setQueuing(mode, difficulty);
    
    websocketManager.send({
      type: 'find_match',
      difficulty: mode === 'bot' ? (difficulty || 'easy') : '',
    });
  }, [setQueuing]);

  const makeMove = useCallback((column: number) => {
    websocketManager.send({ type: 'make_move', column });
  }, []);

  const surrender = useCallback(() => {
    websocketManager.send({ type: 'abandon_game' });
  }, []);

  const disconnect = useCallback(() => {
    websocketManager.disconnect();
  }, []);

  const sendMessage = useCallback((message: Record<string, unknown>) => {
    websocketManager.send(message as any);
  }, []);

  const spectateGame = useCallback(async (gameId: string) => {
    websocketManager.connect();
    websocketManager.send({ type: 'watch_game', gameId });
  }, []);

  const leaveSpectate = useCallback((gameId: string) => {
    websocketManager.send({ type: 'leave_spectate', gameId });
  }, []);

  const connect = useCallback(async () => {
    websocketManager.connect();
  }, []);

  return {
    connect,
    findMatch,
    makeMove,
    surrender,
    disconnect,
    sendMessage,
    spectateGame,
    leaveSpectate,
  };
};
