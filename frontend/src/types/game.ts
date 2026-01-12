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
  queuedAt: number | null;
  opponentDisconnected: boolean;
  disconnectedAt: number | null;
  matchEnded: boolean;
  matchEndedAt: number | null;
  error: string | null;
  rematchRequested: boolean;          // Whether a rematch has been requested
  rematchRequester: string | null;    // Username of player who requested rematch
  rematchTimeout: number | null;      // Countdown for accepting rematch (seconds)
  allowRematch: boolean;              // Whether rematch button should be shown
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
  rematchRequester?: string;  // Username who requested rematch
  rematchTimeout?: number;     // Seconds remaining to respond
  allowRematch?: boolean;      // Whether rematch button should be shown
}

export interface ClientMessage {
  type: "join_queue" | "move" | "reconnect" | "rematch_request" | "rematch_response";
  jwt: string;
  gameID?: string;
  column?: number;
  difficulty?: string;
  requestRematch?: boolean;
  rematchResponse?: string;  // "accept" or "decline"
}
