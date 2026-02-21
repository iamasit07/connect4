import { create } from 'zustand';
import type { User, AuthState } from '../types';

interface AuthActions {
  setUser: (user: User) => void;
  logout: () => void;
  setLoading: (loading: boolean) => void;
  setActiveGameId: (gameId: string) => void;
  clearActiveGameId: () => void;
}

export const useAuthStore = create<AuthState & AuthActions>()((set) => ({
  user: null,
  isAuthenticated: false,
  isLoading: true,

  setUser: (user) => {
    set({
      user,
      isAuthenticated: true,
      isLoading: false,
    });
  },

  logout: () => {
    set({
      user: null,
      isAuthenticated: false,
      isLoading: false,
    });
  },

  setLoading: (loading) => set({ isLoading: loading }),

  setActiveGameId: (gameId) => set((state) => ({
    user: state.user ? { ...state.user, activeGameId: gameId } : null,
  })),

  clearActiveGameId: () => set((state) => ({
    user: state.user ? { ...state.user, activeGameId: undefined } : null,
  })),
}));
