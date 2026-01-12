import { Navigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { useEffect, useState } from "react";

interface PrivateRouteProps {
  children: React.ReactNode;
}
  
const PrivateRoute = ({ children }: PrivateRouteProps) => {
  const { isAuthenticated, initialLoading } = useAuth();
  const [isChecking, setIsChecking] = useState(true);
  const [isAuthed, setIsAuthed] = useState(false);

  useEffect(() => {
    if (initialLoading) return; // Wait for initial auth check to complete

    // Check authentication status
    const checkAuth = () => {
      const authed = isAuthenticated();
      setIsAuthed(authed);
      setIsChecking(false);
    };
    
    checkAuth();
  }, [isAuthenticated, initialLoading]);

  // Show nothing while checking (prevents flash of redirect)
  if (isChecking || initialLoading) {
    return <div className="min-h-screen flex items-center justify-center">Loading...</div>;
  }

  if (!isAuthed) {
    // Redirect to login
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

export default PrivateRoute;
