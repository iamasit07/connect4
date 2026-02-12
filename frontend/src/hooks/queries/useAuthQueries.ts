import { useQuery, useMutation } from '@tanstack/react-query';
import api from '@/lib/axios';
import type { User, LoginCredentials, SignupCredentials, AuthResponse } from '@/features/auth/types';

// Query Keys
export const authKeys = {
  all: ['auth'] as const,
  user: () => [...authKeys.all, 'user'] as const,
};

// Fetch User
export const useUser = () => useQuery({
  queryKey: authKeys.user(),
  queryFn: async () => {
    const { data } = await api.get<User>('/auth/me');
    return data;
  },
  retry: false,
});

// Mutations
export const useLogin = () => useMutation({
  mutationFn: async (credentials: LoginCredentials) => {
    const { data } = await api.post<AuthResponse>('/auth/login', credentials);
    return data;
  },
});

export const useSignup = () => useMutation({
  mutationFn: async (credentials: SignupCredentials) => {
    const { data } = await api.post<AuthResponse>('/auth/register', credentials);
    return data;
  },
});

export const useLogout = () => useMutation({
  mutationFn: async () => {
    await api.post('/auth/logout');
  },
});

export const useUpdateProfile = () => useMutation({
  mutationFn: async (formData: FormData) => {
    const { data } = await api.put<User>('/auth/profile', formData);
    return data;
  },
});
