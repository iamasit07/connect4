import { create } from "zustand";
import { BOARD_ROWS, BOARD_COLS, TURN_TIME_LIMIT } from "@/lib/config";
import type {
  Board,
  PlayerNumber,
  GameMode,
  BotDifficulty,
  CellValue,
} from "../types";

const createEmptyBoard = (): Board => {
  return Array(BOARD_ROWS)
    .fill(null)
    .map(() => Array(BOARD_COLS).fill(0) as CellValue[]);
};

type GameStatus = "idle" | "queuing" | "playing" | "finished";
type ConnectionStatus = "disconnected" | "connecting" | "connected" | "error";
type RematchStatus = "idle" | "sent" | "received" | "accepted" | "declined";

interface GameStore {
  // Connection state
  connectionStatus: ConnectionStatus;
  connectionError: string | null;

  // Game state
  gameId: string | null;
  board: Board;
  currentTurn: PlayerNumber;
  myPlayer: PlayerNumber | null;
  opponent: string | null;
  gameStatus: GameStatus;
  winner: string | null;
  winReason: string | null;
  winningCells: { row: number; col: number }[] | null;
  lastMove: { column: number; row: number; player: number } | null;
  gameMode: GameMode | null;
  botDifficulty: BotDifficulty | null;

  // Timer state
  turnTimeLimit: number;
  timeLeft: number;

  // Spectator state
  isSpectator: boolean;
  spectatorCount: number;

  // Rematch state
  rematchStatus: RematchStatus;
  allowRematch: boolean;

  // Connection actions
  setConnectionStatus: (status: ConnectionStatus, error?: string) => void;

  // Game actions
  setQueuing: (mode: GameMode, difficulty?: BotDifficulty) => void;
  initGame: (data: {
    gameId: string;
    board: Board;
    currentTurn: PlayerNumber;
    myPlayer: PlayerNumber;
    opponent: string;
    turnTimeLimit?: number;
  }) => void;
  updateGameState: (data: {
    board: Board;
    currentTurn: PlayerNumber;
    lastMove?: { column: number; row: number; player: number };
    timeLeft?: number;
  }) => void;
  endGame: (data: {
    winner: string;
    reason: string;
    winningCells?: { row: number; col: number }[];
    board?: CellValue[][];
    allowRematch?: boolean;
  }) => void;
  resetGame: () => void;

  // Timer actions
  setTimeLeft: (time: number) => void;

  // Spectator actions
  setSpectatorMode: (isSpectator: boolean) => void;
  setSpectatorCount: (count: number) => void;
  initSpectatorGame: (data: {
    gameId: string;
    board: Board;
    currentTurn: PlayerNumber;
    player1: string;
    player2: string;
  }) => void;

  // Rematch actions
  setRematchStatus: (status: RematchStatus) => void;
  setAllowRematch: (allow: boolean) => void;

  // Helpers
  isMyTurn: () => boolean;
  getMyColor: () => "red" | "yellow" | null;
  canDropInColumn: (col: number) => boolean;
}

export const useGameStore = create<GameStore>((set, get) => ({
  // Connection state
  connectionStatus: "disconnected",
  connectionError: null,

  // Game state
  gameId: null,
  board: createEmptyBoard(),
  currentTurn: 1,
  myPlayer: null,
  opponent: null,
  gameStatus: "idle",
  winner: null,
  winReason: null,
  winningCells: null,
  lastMove: null,
  gameMode: null,
  botDifficulty: null,

  // Timer state
  turnTimeLimit: TURN_TIME_LIMIT,
  timeLeft: TURN_TIME_LIMIT,

  // Spectator state
  isSpectator: false,
  spectatorCount: 0,

  // Rematch state
  rematchStatus: "idle",
  allowRematch: true,

  setConnectionStatus: (connectionStatus, error) =>
    set({
      connectionStatus,
      connectionError: error || null,
    }),

  setQueuing: (mode, difficulty) =>
    set({
      gameStatus: "queuing",
      gameMode: mode,
      botDifficulty: difficulty || null,
      board: createEmptyBoard(),
      winner: null,
      winReason: null,
      winningCells: null,
      rematchStatus: "idle",
    }),

  initGame: ({
    gameId,
    board,
    currentTurn,
    myPlayer,
    opponent,
    turnTimeLimit,
  }) =>
    set({
      gameId,
      board,
      currentTurn,
      myPlayer,
      opponent,
      gameStatus: "playing",
      winner: null,
      winReason: null,
      winningCells: null,
      lastMove: null,
      turnTimeLimit: turnTimeLimit || TURN_TIME_LIMIT,
      timeLeft: turnTimeLimit || TURN_TIME_LIMIT,
      isSpectator: false,
      rematchStatus: "idle",
    }),

  updateGameState: ({ board, currentTurn, lastMove, timeLeft }) =>
    set((state) => ({
      board,
      currentTurn,
      lastMove: lastMove || state.lastMove,
      timeLeft: timeLeft !== undefined ? timeLeft : state.turnTimeLimit,
    })),

  endGame: ({ winner, reason, winningCells, board, allowRematch }) =>
    set((state) => ({
      gameStatus: "finished",
      winner,
      winReason: reason,
      winningCells: winningCells || null,
      board: board || state.board,
      allowRematch: allowRematch !== undefined ? allowRematch : true,
    })),

  resetGame: () =>
    set({
      connectionStatus: "disconnected",
      connectionError: null,
      gameId: null,
      board: createEmptyBoard(),
      currentTurn: 1,
      myPlayer: null,
      opponent: null,
      gameStatus: "idle",
      winner: null,
      winReason: null,
      winningCells: null,
      lastMove: null,
      gameMode: null,
      botDifficulty: null,
      timeLeft: TURN_TIME_LIMIT,
      isSpectator: false,
      spectatorCount: 0,
      rematchStatus: "idle",
      allowRematch: true,
    }),

  setTimeLeft: (time) => set({ timeLeft: time }),

  setSpectatorMode: (isSpectator) => set({ isSpectator }),

  setSpectatorCount: (count) => set({ spectatorCount: count }),

  initSpectatorGame: ({ gameId, board, currentTurn, player1, player2 }) =>
    set({
      gameId,
      board,
      currentTurn,
      myPlayer: 1, // spectators view from Player 1's perspective
      opponent: player2,
      gameStatus: "playing",
      winner: null,
      winReason: null,
      winningCells: null,
      lastMove: null,
      isSpectator: true,
      gameMode: "pvp",
      rematchStatus: "idle",
      allowRematch: false,
    }),

  setRematchStatus: (status) => set({ rematchStatus: status }),

  setAllowRematch: (allow) => set({ allowRematch: allow }),

  isMyTurn: () => {
    const { myPlayer, currentTurn, gameStatus, isSpectator } = get();
    return gameStatus === "playing" && myPlayer === currentTurn && !isSpectator;
  },

  getMyColor: () => {
    const { myPlayer } = get();
    if (!myPlayer) return null;
    return myPlayer === 1 ? "red" : "yellow";
  },

  canDropInColumn: (col) => {
    const { board, gameStatus, isSpectator } = get();
    if (gameStatus !== "playing" || isSpectator) return false;
    return board[0][col] === 0;
  },
}));
