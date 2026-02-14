export interface User {
  id: string;
  username: string;
  name: string;
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
  name: string;
  email: string;
}

export interface CompleteSignupRequest {
  token?: string;
  name: string;
  email: string;
  username: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}
