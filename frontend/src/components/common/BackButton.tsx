import { useNavigate, useLocation } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { ChevronLeft } from 'lucide-react';

export const BackButton = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const hiddenPaths = ['/', '/dashboard', '/login', '/signup', '/play'];
  
  if (hiddenPaths.includes(location.pathname)) return null;

  return (
    <Button
      variant="ghost"
      size="icon"
      className="mr-1 -ml-2" 
      onClick={() => navigate(-1)}
      title="Go back"
    >
      <ChevronLeft className="h-5 w-5" />
      <span className="sr-only">Go back</span>
    </Button>
  );
};
