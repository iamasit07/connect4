import { useState, useEffect, useCallback } from "react";
import { apiClient, User } from "../utils/api";

interface UseAuthReturn {
  user: User | null;
  loading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<boolean>;
  signup: (username: string, password: string) => Promise<boolean>;
  logout: () => void;
  isAuthenticated: () => boolean;
  getToken: () => string | null;
}

// Decode JWT to extract user info and check expiration
function parseJWT(token: string): User | null {
  try {
    const base64Url = token.split(".")[1];
    const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split("")
        .map((c) => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
        .join("")
    );
    const payload = JSON.parse(jsonPayload);
    return {
      id: payload.user_id,
      username: payload.username,
    };
  } catch (error) {
    console.error("Failed to parse JWT:", error);
    return null;
  }
}

// Check if JWT token is expired
function isTokenExpired(token: string): boolean {
  try {
    const base64Url = token.split(".")[1];
    const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split("")
        .map((c) => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
        .join("")
    );
    const payload = JSON.parse(jsonPayload);

    // Check if exp field exists and if it's expired
    if (!payload.exp) {
      return true; // No expiration means invalid token
    }

    // exp is in seconds, Date.now() is in milliseconds
    const currentTime = Date.now() / 1000;
    return payload.exp < currentTime;
  } catch (error) {
    console.error("Failed to check token expiration:", error);
    return true; // If we can't parse it, consider it expired
  }
}

export function useAuth(): UseAuthReturn {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Initialize user from stored token on mount
  useEffect(() => {
    const token = localStorage.getItem("authToken");
    if (token) {
      // Check if token is expired
      if (isTokenExpired(token)) {
        localStorage.removeItem("authToken");
        setUser(null);
        return;
      }

      const userData = parseJWT(token);
      if (userData) {
        setUser(userData);
      } else {
        // Invalid token, clear it
        localStorage.removeItem("authToken");
      }
    }
  }, []);

  // Periodic check for token expiration
  useEffect(() => {
    const checkTokenExpiration = () => {
      const token = localStorage.getItem("authToken");
      if (token && isTokenExpired(token)) {
        localStorage.removeItem("authToken");
        setUser(null);
      }
    };

    // Check every minute
    const interval = setInterval(checkTokenExpiration, 60000);

    return () => clearInterval(interval);
  }, []);

  const login = useCallback(
    async (username: string, password: string): Promise<boolean> => {
      setLoading(true);
      setError(null);

      try {
        const response = await apiClient.login(username, password);

        // Store token
        localStorage.setItem("authToken", response.token);

        // Set user
        setUser({
          id: response.user_id,
          username: response.username,
        });

        setLoading(false);
        return true;
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : "Login failed";
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

        // Store token
        localStorage.setItem("authToken", response.token);

        // Set user
        setUser({
          id: response.user_id,
          username: response.username,
        });

        setLoading(false);
        return true;
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : "Signup failed";
        setError(errorMessage);
        setLoading(false);
        return false;
      }
    },
    []
  );

  const logout = useCallback(() => {
    localStorage.removeItem("authToken");
    setUser(null);
    setError(null);
  }, []);

  const isAuthenticated = useCallback((): boolean => {
    const token = localStorage.getItem("authToken");
    if (!token) {
      return false;
    }

    // Check if token is expired
    if (isTokenExpired(token)) {
      return false;
    }

    return true;
  }, []);

  const getToken = useCallback((): string | null => {
    return localStorage.getItem("authToken");
  }, []);

  return {
    user,
    loading,
    error,
    login,
    signup,
    logout,
    isAuthenticated,
    getToken,
  };
}
