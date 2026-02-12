import { useEffect, useCallback } from 'react';
import { websocketManager } from '../services/websocketManager';
import { useGameStore } from '../store/gameStore';
import type { BotDifficulty } from '../types';
import { toast } from 'sonner';

export const useGameSocket = (onGameStart?: (gameId: string) => void) => {
  const { setConnectionStatus, setQueuing } = useGameStore();

  // Register onGameStart callback
  useEffect(() => {
    if (!onGameStart) return undefined;
    
    const unregister = websocketManager.onGameStart(onGameStart);
    return unregister;
  }, [onGameStart]);

  // Connect on mount
  useEffect(() => {
    const connect = async () => {
      try {
        setConnectionStatus('connecting');
        await websocketManager.connect();
        setConnectionStatus('connected');
      } catch (error) {
        console.error('[WebSocket] Connection failed:', error);
        setConnectionStatus('error');
        toast.error('Failed to connect to game server');
      }
    };
    
    connect();
  }, [setConnectionStatus]);

  const findMatch = useCallback(async (mode: 'pvp' | 'bot', difficulty?: BotDifficulty) => {
    setQueuing(mode, difficulty);
    
    // Ensure we are connected before sending
    await websocketManager.connect();
    
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

  const sendMessage = useCallback((message: any) => {
    websocketManager.send(message);
  }, []);

  const connect = useCallback(() => {
    return websocketManager.connect();
  }, []);

  return {
    findMatch,
    makeMove,
    surrender,
    disconnect,
    sendMessage,
    connect,
  };
};
