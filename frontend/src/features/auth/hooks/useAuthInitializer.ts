import { useEffect } from 'react';
import { useAuthStore } from '@/features/auth/store/authStore';
import api from '@/lib/axios';
import type { User } from '@/features/auth/types';

export const useAuthInitializer = () => {
  const { setUser, setLoading, logout } = useAuthStore();

  useEffect(() => {
    const initializeAuth = async () => {
      try {
        setLoading(true);
        const response = await api.get<User>('/auth/me');
        setUser(response.data);
      } catch {
        logout();
      } finally {
        setLoading(false);
      }
    };

    initializeAuth();
  }, [setUser, setLoading, logout]);
};
