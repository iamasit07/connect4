const BASE_HTTP_URL = import.meta.env.VITE_BACKEND_URL || 'https://localhost:8080';
const BASE_WS_URL = import.meta.env.VITE_WS_URL || 'wss://localhost:8080';

export const API_BASE_URL = `${BASE_HTTP_URL}/api`;
export const WS_URL = `${BASE_WS_URL}/ws`;

// Game Constants
export const BOARD_ROWS = 6;
export const BOARD_COLS = 7;
export const WINNING_LENGTH = 4;

// Animation Timings (ms)
export const DISK_DROP_DURATION = 500;
export const BOT_MOVE_DELAY = 800; // Artificial delay for bot moves to feel natural
export const CONFETTI_DURATION = 3000;

// Bot Difficulty Levels
export const BOT_DIFFICULTIES = {
  easy: { name: 'Alice', description: 'Perfect for beginners', emoji: '🧑‍🚀' },
  medium: { name: 'Bob', description: 'A balanced challenge', emoji: '🕵️‍♂️' },
  hard: { name: 'Charles', description: 'For the brave', emoji: '🧙‍♂️' },
} as const;