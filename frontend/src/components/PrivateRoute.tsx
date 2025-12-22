import { Navigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { useEffect, useState } from "react";

interface PrivateRouteProps {
  children: React.ReactNode;
}

const PrivateRoute = ({ children }: PrivateRouteProps) => {
  const { isAuthenticated, logout } = useAuth();
  const [isChecking, setIsChecking] = useState(true);
  const [isAuthed, setIsAuthed] = useState(false);

  useEffect(() => {
    // Check authentication status
    const checkAuth = () => {
      const authed = isAuthenticated();
      setIsAuthed(authed);
      
      if (!authed) {
        logout(); // Clear any stale credentials
      }
      
      setIsChecking(false);
    };
    
    checkAuth();
  }, [isAuthenticated, logout]);

  // Show nothing while checking (prevents flash of redirect)
  if (isChecking) {
    return null;
  }

  if (!isAuthed) {
    // Redirect to login
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

export default PrivateRoute;
