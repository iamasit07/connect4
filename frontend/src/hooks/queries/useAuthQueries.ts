import { useQuery, useMutation } from '@tanstack/react-query';
import api from '@/lib/axios';
import type { User, LoginCredentials, AuthResponse, CompleteSignupRequest } from '@/features/auth/types';

// Query Keys
export const authKeys = {
  all: ['auth'] as const,
  user: () => [...authKeys.all, 'user'] as const,
};

// Fetch User (/me returns flat user data with token)
export const useUser = () => useQuery({
  queryKey: authKeys.user(),
  queryFn: async () => {
    const { data } = await api.get<User>('/auth/me');
    return data;
  },
  retry: false,
});

// Login (returns { token, user })
export const useLogin = () => useMutation({
  mutationFn: async (credentials: LoginCredentials) => {
    const { data } = await api.post<AuthResponse>('/auth/login', credentials);
    return data;
  },
});

// Complete Signup — handles both manual and Google flows
export const useCompleteSignup = () => useMutation({
  mutationFn: async (request: CompleteSignupRequest) => {
    // If there's a Google token, use the Google complete endpoint
    if (request.token) {
      const { data } = await api.post<AuthResponse>('/auth/google/complete', {
        token: request.token,
        username: request.username,
        password: request.password,
      });
      return data;
    }
    // Otherwise use the regular register endpoint
    const { data } = await api.post<AuthResponse>('/auth/register', {
      username: request.username,
      name: request.name,
      email: request.email,
      password: request.password,
    });
    return data;
  },
});

export const useLogout = () => useMutation({
  mutationFn: async () => {
    await api.post('/auth/logout');
  },
});

export const useUpdateProfile = () => useMutation({
  mutationFn: async (data: { name: string }) => {
    const response = await api.put<{ user: User }>('/auth/profile', data);
    return response.data;
  },
});
