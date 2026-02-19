import api from "@/lib/axios";
import { Board } from "../types";

export interface GameDetails {
  GameID: string;
  Player1ID: number;
  Player1Username: string;
  Player2ID: number | null;
  Player2Username: string;
  WinnerID: number | null;
  WinnerUsername: string;
  Reason: string;
  TotalMoves: number;
  DurationSeconds: number;
  CreatedAt: string;
  FinishedAt: string;
  board_state: Board;
}

export const gameService = {
  getGameDetails: async (gameId: string): Promise<GameDetails> => {
    const response = await api.get<GameDetails>(`/history/${gameId}`);
    return response.data;
  },
};
