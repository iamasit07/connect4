import { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { API_BASE_URL } from '@/lib/config';
import { useAuthStore } from '../store/authStore';
import type { LoginCredentials, CompleteSignupRequest } from '../types';
import { useLogin, useCompleteSignup, useLogout, useUser } from '@/hooks/queries/useAuthQueries';

export const useAuth = () => {
  const navigate = useNavigate();
  const { user, isAuthenticated, isLoading, setUser, logout: storeLogout, setLoading } = useAuthStore();
  
  const loginMutation = useLogin();
  const completeSignupMutation = useCompleteSignup();
  const logoutMutation = useLogout();
  const { refetch: refetchUser } = useUser();

  const checkAuth = useCallback(async () => {
    try {
      const { data, isError } = await refetchUser();
      
      if (data && !isError) {
        setUser(data);
      } else {
        storeLogout();
      }
    } catch (error) {
      storeLogout();
    } finally {
      setLoading(false);
    }
  }, [setUser, storeLogout, setLoading, refetchUser]);

  const login = useCallback(async (credentials: LoginCredentials) => {
    try {
      setLoading(true);
      const response = await loginMutation.mutateAsync(credentials);
      setUser(response.user);
      toast.success('Welcome back!');
      navigate('/dashboard');
    } catch (error: any) {
      const message = error.response?.data?.message || 'Login failed';
      toast.error(message);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [navigate, setUser, setLoading, loginMutation]);

  const completeSignup = useCallback(async (request: CompleteSignupRequest) => {
    try {
      setLoading(true);
      const response = await completeSignupMutation.mutateAsync(request);
      setUser(response.user);
      toast.success('Account created successfully!');
      navigate('/dashboard');
    } catch (error: any) {
      const message = error.response?.data?.error || error.response?.data?.message || 'Signup failed';
      toast.error(message);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [navigate, setUser, setLoading, completeSignupMutation]);

  const loginWithGoogle = useCallback(() => {
    window.location.href = `${API_BASE_URL}/auth/google/login`;
  }, []);

  const handleLogout = useCallback(async () => {
    try {
      await logoutMutation.mutateAsync();
    } catch (error) {
      console.error('Logout request failed:', error);
    } finally {
      storeLogout();
      toast.success('Logged out successfully');
      navigate('/');
    }
  }, [storeLogout, navigate, logoutMutation]);

  return {
    user,
    isAuthenticated,
    isLoading,
    login,
    completeSignup,
    loginWithGoogle,
    logout: handleLogout,
    checkAuth,
  };
};
