const BASE_HTTP_URL = import.meta.env.VITE_BACKEND_URL || 'https://four-in-a-row-backend-tnan.onrender.com';
const BASE_WS_URL = import.meta.env.VITE_WS_URL || 'wss://four-in-a-row-backend-tnan.onrender.com';

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

export const QUEUE_FUN_FACTS = [
  "Connect 4 was first sold in 1974 by Milton Bradley.",
  "The game has over 4 trillion possible game positions!",
  "A perfect game always results in a win for the first player.",
  "The game is also known as 'Four in a Row' or 'Captain's Mistress'.",
  "In 1988, Victor Allis solved Connect 4 - first player wins with perfect play.",
  "The standard board is 7 columns × 6 rows = 42 slots.",
  "Connect 4 is classified as a solved game in game theory.",
  "Some versions use a 8×8 or larger board for more complexity.",
];
