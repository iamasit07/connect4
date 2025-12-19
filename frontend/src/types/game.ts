export type PlayerID = 0 | 1 | 2;

export interface GameState {
  gameId: string | null;
  yourPlayer: PlayerID | null;
  opponent: string | null;
  currentTurn: PlayerID | null;
  board: PlayerID[][];
  gameOver: boolean;
  winner: string | null;
  reason: string | null;
  inQueue: boolean;
  queuedAt: number | null; // Timestamp when joined queue
}

export interface ServerMessage {
  type: string;
  gameId?: string;
  yourPlayer?: number;
  opponent?: string;
  currentTurn?: number;
  board?: PlayerID[][];
  column?: number;
  row?: number;
  player?: number;
  nextTurn?: number;
  winner?: string;
  reason?: string;
  message?: string;
}

export interface ClientMessage {
  type: "join_queue" | "make_move" | "reconnect";
  username?: string;
  gameID?: string;
  column?: number;
}
