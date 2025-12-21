// errorMessages.ts - User-friendly error message mappings

export const ERROR_MESSAGES: Record<string, string> = {
  // Authentication errors
  invalid_username: "Please enter a valid username",
  token_taken: "This session is already in use. Please try again.",
  not_authenticated: "Please join a game first",
  
  // Game errors
  no_active_game: "You don't have an active game",
  game_not_found: "Game session not found",
  game_finished: "This game has already ended",
  invalid_move: "Invalid move. Please try again.",
  not_your_turn: "It's not your turn yet",
  
  // Queue errors
  queue_error: "Failed to join matchmaking. Please try again.",
  
  // Reconnection errors
  game_full: "This game is already full",
  already_connected: "You are already connected to this game",
  not_disconnected: "You weren't disconnected from this game",
  invalid_token: "Invalid session token",
  not_in_game: "You are not a player in this game",
  
  // Generic
  unknown_message_type: "An unexpected error occurred",
  default: "An error occurred. Please try again.",
};

export function getErrorMessage(errorType: string): string {
  return ERROR_MESSAGES[errorType] || ERROR_MESSAGES.default;
}
