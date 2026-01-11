import { createContext, useContext, useState, useEffect, useCallback, useRef, ReactNode } from "react";
import { apiClient, User } from "../utils/api";
import { isTokenExpired } from "../utils/cookies";

interface AuthContextType {
  user: User | null;
  loading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<boolean>;
  signup: (username: string, password: string) => Promise<boolean>;
  logout: () => Promise<void>;
  isAuthenticated: () => boolean;
  getToken: () => string | null;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const authInitialized = useRef(false);

  // Initialize user from session on mount
  useEffect(() => {
    if (authInitialized.current) return;
    authInitialized.current = true;

    const initAuth = async () => {
      try {
        const response = await apiClient.me();
        setToken(response.token);
        setUser({
          id: response.user_id,
          username: response.username,
        });
      } catch (err) {
        // No valid session - this is expected when not logged in
        console.log("No active session, user not authenticated");
        setToken(null);
        setUser(null);
      }
    };

    initAuth().catch((err) => {
      console.error("Auth initialization error:", err);
    });
  }, []);

  // Periodic token expiration check
  useEffect(() => {
    if (!token) return;

    const checkTokenExpiration = () => {
      if (isTokenExpired(token)) {
        setToken(null);
        setUser(null);
      }
    };

    const interval = setInterval(checkTokenExpiration, 60000);
    return () => clearInterval(interval);
  }, [token]);

  const login = useCallback(
    async (username: string, password: string): Promise<boolean> => {
      setLoading(true);
      setError(null);

      try {
        const response = await apiClient.login(username, password);
        setToken(response.token);
        setUser({
          id: response.user_id,
          username: response.username,
        });
        setLoading(false);
        return true;
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : "Login failed";
        setError(errorMessage);
        setLoading(false);
        return false;
      }
    },
    []
  );

  const signup = useCallback(
    async (username: string, password: string): Promise<boolean> => {
      setLoading(true);
      setError(null);

      try {
        const response = await apiClient.signup(username, password);
        setToken(response.token);
        setUser({
          id: response.user_id,
          username: response.username,
        });
        setLoading(false);
        return true;
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : "Signup failed";
        setError(errorMessage);
        setLoading(false);
        return false;
      }
    },
    []
  );

  const logout = useCallback(async () => {
    try {
      await apiClient.logout();
    } catch (err) {
      console.error("Logout error:", err);
    }

    setToken(null);
    setUser(null);
    setError(null);
  }, []);

  const isAuthenticated = useCallback((): boolean => {
    if (!token) return false;
    if (isTokenExpired(token)) return false;
    return true;
  }, [token]);

  const getToken = useCallback((): string | null => {
    return token;
  }, [token]);

  const value = {
    user,
    loading,
    error,
    login,
    signup,
    logout,
    isAuthenticated,
    getToken,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
