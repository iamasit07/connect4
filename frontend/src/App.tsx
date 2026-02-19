import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { AppLayout } from "@/components/layout/AppLayout";
import { ProtectedRoute } from "@/components/layout/ProtectedRoute";
import { useAuthInitializer } from "@/features/auth/hooks/useAuthInitializer";
import LandingPage from "./pages/LandingPage";
import LoginPage from "./features/auth/pages/LoginPage";
import SignupPage from "./features/auth/pages/SignupPage";
import AuthCallback from "./features/auth/pages/AuthCallback";
import CompleteSignupPage from "./features/auth/pages/CompleteSignupPage";
import Dashboard from "./pages/Dashboard";
import PlayGame from "./features/game/routes/PlayGame";
import QueuePage from "./features/game/routes/QueuePage";
import BotLoadingPage from "./features/game/routes/BotLoadingPage";
import GamePage from "./features/game/routes/GamePage";
import GameHistory from "./pages/GameHistory";
import Leaderboard from "./pages/Leaderboard";
import Profile from "./pages/Profile";
import NotFound from "./pages/NotFound";
import { queryClient } from "@/lib/react-query";
import { useUIStore } from "@/stores/useUIStore";
import { useEffect } from "react";
import { ActiveGamePopup } from "@/features/game/components/ActiveGamePopup";

const ThemeInitializer = () => {
  const { theme } = useUIStore();

  useEffect(() => {
    const root = document.documentElement;
    root.classList.remove("light", "dark");

    if (theme === "system") {
      const systemTheme = window.matchMedia("(prefers-color-scheme: dark)")
        .matches
        ? "dark"
        : "light";
      root.classList.add(systemTheme);
    } else {
      root.classList.add(theme);
    }
  }, [theme]);

  return null;
};

const AppContent = () => {
  useAuthInitializer();

  return (
    <>
      <ThemeInitializer />
      <ActiveGamePopup />
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/signup" element={<SignupPage />} />
        <Route path="/auth/callback" element={<AuthCallback />} />
        <Route path="/complete-signup" element={<CompleteSignupPage />} />

        {/* App routes with layout */}
        <Route element={<AppLayout />}>
          <Route path="/play" element={<PlayGame />} />
          <Route path="/play/queue" element={<QueuePage />} />
          <Route path="/play/bot" element={<BotLoadingPage />} />
          <Route path="/game/:gameId" element={<GamePage />} />
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <Dashboard />
              </ProtectedRoute>
            }
          />
          <Route
            path="/history"
            element={
              <ProtectedRoute>
                <GameHistory />
              </ProtectedRoute>
            }
          />
          <Route
            path="/leaderboard"
            element={
              <ProtectedRoute>
                <Leaderboard />
              </ProtectedRoute>
            }
          />
          <Route
            path="/profile"
            element={
              <ProtectedRoute>
                <Profile />
              </ProtectedRoute>
            }
          />
        </Route>

        <Route path="*" element={<NotFound />} />
      </Routes>
    </>
  );
};

const App = () => (
  <QueryClientProvider client={queryClient}>
    <TooltipProvider>
      <Toaster />
      <Sonner />
      <BrowserRouter>
        <AppContent />
      </BrowserRouter>
    </TooltipProvider>
  </QueryClientProvider>
);

export default App;
