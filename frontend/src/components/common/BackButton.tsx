import { useNavigate, useLocation } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { ChevronLeft } from 'lucide-react';

export const BackButton = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const hiddenPaths = ['/', '/dashboard', '/login', '/signup'];
  
  if (hiddenPaths.includes(location.pathname)) return null;

  const routeHierarchy: Record<string, string> = {
    '/play': '/dashboard',
    '/play/queue': '/dashboard',
    '/play/bot': '/dashboard',
    '/leaderboard': '/dashboard',
    '/profile': '/dashboard',
    '/history': '/dashboard',
  };

  const handleBack = () => {
    if (routeHierarchy[location.pathname]) {
      navigate(routeHierarchy[location.pathname]);
      return;
    }

    if (location.pathname.startsWith('/game/')) {
      navigate('/dashboard');
      return;
    }
    navigate(-1);
  };

  return (
    <Button
      variant="ghost"
      size="icon"
      className="mr-1 -ml-2" 
      onClick={handleBack}
      title="Go back"
    >
      <ChevronLeft className="h-5 w-5" />
      <span className="sr-only">Go back</span>
    </Button>
  );
};
