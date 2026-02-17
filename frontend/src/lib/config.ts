const getBaseUrl = () => {
  if (typeof window !== 'undefined') {
    const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    if (!isLocal) return '';
  }

  if (import.meta.env.PROD) return '';
  
  if (import.meta.env.VITE_BACKEND_URL) return import.meta.env.VITE_BACKEND_URL;
  return '';
};

const getWsUrl = () => {
  if (import.meta.env.VITE_WS_URL) return import.meta.env.VITE_WS_URL;
  if (typeof window !== 'undefined') {
    return window.location.origin.replace(/^http/, 'ws');
  }
  return 'wss://localhost:8080';
};

const BASE_HTTP_URL = getBaseUrl();
const BASE_WS_URL = getWsUrl();

export const API_BASE_URL = BASE_HTTP_URL ? `${BASE_HTTP_URL}/api` : '/api';
export const WS_URL = BASE_WS_URL ? `${BASE_WS_URL}/ws` : '/ws';

export const BOARD_ROWS = 6;
export const BOARD_COLS = 7;
export const WINNING_LENGTH = 4;
export const TURN_TIME_LIMIT = 900;

export const DISK_DROP_DURATION = 500;
export const BOT_MOVE_DELAY = 800;
export const CONFETTI_DURATION = 3000;
export const BOT_DIFFICULTIES = {
  easy: { name: 'Alice', description: 'Perfect for beginners', emoji: '🧑‍🚀' },
  medium: { name: 'Bob', description: 'A balanced challenge', emoji: '🕵️‍♂️' },
  hard: { name: 'Charles', description: 'For the brave', emoji: '🧙‍♂️' },
} as const;