import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { useAuthStore } from '../store/authStore';
import api from '@/lib/axios';
import { toast } from 'sonner';
import type { User } from '../types';

const AuthCallback = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { setUser, setLoading } = useAuthStore();

  useEffect(() => {
    const handleCallback = async () => {
      const error = searchParams.get('error');

      if (error) {
        toast.error(error);
        navigate('/login');
        return;
      }

      try {
        setLoading(true);        
        const response = await api.get<User>('/auth/me');
        setUser(response.data);
        toast.success('Successfully signed in with Google!');
        navigate('/dashboard');
      } catch (err: any) {
        console.error('Auth callback failed:', err);
        const errorMessage = err.response?.data?.error || err.message || 'Failed to complete authentication';
        toast.error(errorMessage);
        setTimeout(() => navigate('/login'), 2000);
      } finally {
        setLoading(false);
      }
    };

    handleCallback();
  }, [searchParams, navigate, setUser, setLoading]);

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <Loader2 className="h-12 w-12 animate-spin text-primary mx-auto mb-4" />
        <p className="text-muted-foreground">Completing sign in...</p>
      </div>
    </div>
  );
};

export default AuthCallback;
