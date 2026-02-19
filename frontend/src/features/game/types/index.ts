// ============================================
// WebSocket Client Messages (Sent by Frontend)
// ============================================

export type ClientMessage =
  | InitMessage
  | FindMatchMessage
  | MakeMoveMessage
  | AbandonMessage
  | RequestRematchMessage
  | RematchResponseMessage
  | CancelSearchMessage
  | WatchGameMessage
  | LeaveSpectateMessage;

export interface InitMessage {
  type: "init";
  jwt: string;
}

export interface FindMatchMessage {
  type: "find_match";
  difficulty: "" | "easy" | "medium" | "hard";
}

export interface MakeMoveMessage {
  type: "make_move";
  column: number; // 0-6
}

export interface AbandonMessage {
  type: "abandon_game";
}

export interface RequestRematchMessage {
  type: "request_rematch";
}

export interface RematchResponseMessage {
  type: "rematch_response";
  rematchResponse: "accept" | "decline";
}

export interface CancelSearchMessage {
  type: "cancel_search";
}

export interface WatchGameMessage {
  type: "watch_game";
  gameId: string;
}

export interface LeaveSpectateMessage {
  type: "leave_spectate";
  gameId: string;
}

// ============================================
// WebSocket Server Messages (Received from Backend)
// ============================================

export interface OpponentDisconnectedMessage {
  type: "opponent_disconnected";
  message: string;
  disconnectTimeout: number;
}

export interface OpponentReconnectedMessage {
  type: "opponent_reconnected";
}

export type ServerMessage =
  | QueueJoinedMessage
  | QueueLeftMessage
  | InitGameMessage
  | GameStateMessage
  | MoveMadeMessage
  | GameOverMessage
  | RematchRequestMessage
  | RematchAcceptedMessage
  | RematchDeclinedMessage
  | RematchTimeoutMessage
  | RematchCancelledMessage
  | SpectatorCountMessage
  | SpectateStartMessage
  | QueueTimeoutMessage
  | OpponentDisconnectedMessage
  | OpponentReconnectedMessage
  | NoActiveGameMessage
  | ErrorMessage;

export interface NoActiveGameMessage {
  type: "no_active_game";
  message: string;
}

export interface QueueTimeoutMessage {
  type: "queue_timeout";
}

export interface QueueLeftMessage {
  type: "queue_left";
}

export interface RematchRequestMessage {
  type: "rematch_request";
  rematchRequester: string;
  rematchTimeout: number;
}

export interface RematchAcceptedMessage {
  type: "rematch_accepted";
}

export interface RematchDeclinedMessage {
  type: "rematch_declined";
  allowRematch?: boolean;
}

export interface RematchTimeoutMessage {
  type: "rematch_timeout";
  message: string;
  allowRematch?: boolean;
}

export interface RematchCancelledMessage {
  type: "rematch_cancelled";
  message: string;
}

export interface SpectatorCountMessage {
  type: "spectator_count";
  count: number;
}

export interface SpectateStartMessage {
  type: "spectate_start";
  gameId: string;
  player1: string;
  player2: string;
  currentTurn: 1 | 2;
  board: number[][];
}

export interface QueueJoinedMessage {
  type: "queue_joined";
}

export interface InitGameMessage {
  type: "game_start";
  gameId: string;
  opponent: string; // Backend sends just username
  yourPlayer: 1 | 2;
  currentTurn: 1 | 2;
  board: number[][];
}

export interface GameStateMessage {
  type: "game_state";
  gameId?: string;
  yourPlayer?: 1 | 2;
  opponent?: string;
  board: number[][];
  currentTurn: 1 | 2;
  lastMove?: {
    column: number;
    row: number;
    player: number;
  };
  timeLeft?: number;
}

export interface MoveMadeMessage {
  type: "move_made";
  column: number;
  row: number;
  player: number;
  board: number[][];
  nextTurn: 1 | 2; // Backend sends nextTurn, not currentTurn
  lastMove?: {
    column: number;
    row: number;
    player: number;
  };
}

export interface GameOverMessage {
  type: "game_over";
  winner: string;
  reason: "connect4" | "timeout" | "surrender" | "disconnect";
  board?: number[][];
  newRating?: number;
  winningCells?: { row: number; col: number }[];
  allowRematch?: boolean;
}

export interface ErrorMessage {
  type: "error";
  message: string;
}

// ============================================
// REST API Response Schemas
// ============================================

export interface GameHistoryItem {
  id: string;
  opponentUsername: string;
  result: "win" | "loss" | "draw";
  endReason: string;
  createdAt: string;
  movesCount: number;
}

export interface LiveGame {
  gameId: string;
  player1: { username: string; rating: number };
  player2: { username: string; rating: number };
  spectatorCount: number;
  moveCount: number;
  startedAt: string;
}

export interface LeaderboardEntry {
  rank: number;
  username: string;
  rating: number;
  wins: number;
  losses: number;
}

export interface User {
  id: string;
  username: string;
  email: string;
  rating?: number;
  wins?: number;
  losses?: number;
  draws?: number;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// ============================================
// Game State Types
// ============================================

export type GameMode = "pvp" | "bot";
export type BotDifficulty = "easy" | "medium" | "hard";
export type PlayerNumber = 1 | 2;
export type CellValue = 0 | 1 | 2;
export type Board = CellValue[][];

export interface GameState {
  gameId: string | null;
  board: Board;
  currentTurn: PlayerNumber;
  myPlayer: PlayerNumber | null;
  opponent: { username: string; rating: number } | null;
  status: "idle" | "queuing" | "playing" | "finished";
  winner: string | null;
  winReason: string | null;
  winningCells: { row: number; col: number }[] | null;
  lastMove: { column: number; row: number; player: number } | null;
  gameMode: GameMode | null;
  botDifficulty: BotDifficulty | null;
}

export interface ConnectionState {
  status: "disconnected" | "connecting" | "connected" | "error";
  error: string | null;
}
