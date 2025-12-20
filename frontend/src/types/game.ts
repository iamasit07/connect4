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
  opponentDisconnected: boolean;
  disconnectedAt: number | null; // Timestamp when opponent disconnected
  matchEnded: boolean; // True when trying to reconnect to terminated session
  matchEndedAt: number | null; // Timestamp when match ended error received
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
  userToken?: string;    // Persistent user token
}

export interface ClientMessage {
  type: "join_queue" | "make_move" | "reconnect";
  username?: string;
  gameID?: string;
  column?: number;
  userToken?: string;  // Persistent user token for auth & tracking
}
