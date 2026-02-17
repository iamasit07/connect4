import { useEffect, useCallback } from 'react';
import { websocketManager } from '../services/websocketManager';
import { useGameStore } from '../store/gameStore';
import type { BotDifficulty } from '../types';
import { toast } from 'sonner';

export const useGameSocket = (
  onGameStart?: (gameId: string) => void,
  onQueueTimeout?: () => void
) => {
  const { setConnectionStatus, setQueuing } = useGameStore();

  // Register onGameStart callback
  useEffect(() => {
    if (!onGameStart) return undefined;
    
    const unregister = websocketManager.onGameStart(onGameStart);
    return unregister;
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
    try {
      setConnectionStatus('connecting');
      await websocketManager.connect();
      setConnectionStatus('connected');
    } catch (error) {
      console.error('[WebSocket] Connection failed:', error);
      setConnectionStatus('error');
      toast.error('Failed to connect to game server');
      return;
    }

    setQueuing(mode, difficulty);
    
    websocketManager.send({
      type: 'find_match',
      difficulty: mode === 'bot' ? (difficulty || 'easy') : '',
    });
  }, [setQueuing, setConnectionStatus]);

  const makeMove = useCallback((column: number) => {
    websocketManager.send({ type: 'make_move', column });
  }, []);

  const surrender = useCallback(() => {
    websocketManager.send({ type: 'abandon_game' });
  }, []);

  const disconnect = useCallback(() => {
    websocketManager.disconnect();
    setConnectionStatus('disconnected');
  }, [setConnectionStatus]);

  const sendMessage = useCallback((message: Record<string, unknown>) => {
    websocketManager.send(message as any);
  }, []);

  return {
    findMatch,
    makeMove,
    surrender,
    disconnect,
    sendMessage,
  };
};
