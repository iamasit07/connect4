import { useQuery } from "@tanstack/react-query";
import api from "@/lib/axios";
import type {
  GameHistoryItem,
  LiveGame,
  LeaderboardEntry,
} from "@/features/game/types";

// Query Keys
export const gameKeys = {
  all: ["game"] as const,
  history: () => [...gameKeys.all, "history"] as const,
  live: () => [...gameKeys.all, "live"] as const,
  leaderboard: () => [...gameKeys.all, "leaderboard"] as const,
};

// Hooks
export const useGameHistory = () =>
  useQuery({
    queryKey: gameKeys.history(),
    queryFn: async () => {
      const { data } = await api.get<GameHistoryItem[]>("/history");
      return data ?? [];
    },
    staleTime: 0,
    refetchOnMount: "always",
  });

export const useLiveGames = () =>
  useQuery({
    queryKey: gameKeys.live(),
    queryFn: async (): Promise<LiveGame[]> => {
      // TODO: Implement /games/live endpoint on the backend
      return [];
    },
    refetchInterval: 5000, // Poll every 5s
  });

export const useLeaderboard = () =>
  useQuery({
    queryKey: gameKeys.leaderboard(),
    queryFn: async () => {
      const { data } = await api.get<LeaderboardEntry[]>("/leaderboard");
      return data ?? [];
    },
    staleTime: 30 * 1000,
  });
