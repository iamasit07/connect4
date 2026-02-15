import axios from "axios";
import { API_BASE_URL } from "./config";

const api = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      import("@/features/auth/store/authStore").then(({ useAuthStore }) => {
        const { logout } = useAuthStore.getState();
        logout();
      });

      const publicPaths = [
        "/",
        "/login",
        "/signup",
        "/complete-signup",
        "/auth/callback",
      ];
      if (!publicPaths.includes(window.location.pathname)) {
        window.location.href = "/login";
      }
    }
    return Promise.reject(error);
  },
);

export default api;
