export interface User {
  id: string;
  username: string;
  email: string;
  rating?: number;
  wins?: number;
  losses?: number;
  draws?: number;
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface LoginCredentials {
  username: string;
  password: string;
}

export interface SignupCredentials {
  username: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface CompleteSignupRequest {
  token: string;
  username: string;
  password: string;
}
