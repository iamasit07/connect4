import { Navigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { useEffect, useState } from "react";

interface PrivateRouteProps {
  children: React.ReactNode;
}
  
const PrivateRoute = ({ children }: PrivateRouteProps) => {
  const { isAuthenticated } = useAuth();
  const [isChecking, setIsChecking] = useState(true);
  const [isAuthed, setIsAuthed] = useState(false);

  useEffect(() => {
    // Check authentication status
    const checkAuth = () => {
      const authed = isAuthenticated();
      setIsAuthed(authed);
      setIsChecking(false);
    };
    
    checkAuth();
  }, [isAuthenticated]);

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
