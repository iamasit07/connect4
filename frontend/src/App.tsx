import { BrowserRouter, Route, Routes } from "react-router-dom";
import "./App.css";
import LandingPage from "./pages/landing";
import LoginPage from "./pages/login";
import SignupPage from "./pages/signup";
import CompleteSignupPage from "./pages/complete_signup";
import GamePage from "./pages/game";
import BotDifficulty from "./pages/BotDifficulty";
import LeaderboardPage from "./pages/leaderboard";
import GameHistory from "./pages/GameHistory";
import GameReview from "./pages/GameReview";
import PrivateRoute from "./components/PrivateRoute";
import { AuthProvider } from "./contexts/AuthContext";

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/complete-signup" element={<CompleteSignupPage />} />
          <Route
            path="/"
            element={
              <PrivateRoute>
                <LandingPage />
              </PrivateRoute>
            }
          />
          <Route
            path="/game"
            element={
              <PrivateRoute>
                <GamePage />
              </PrivateRoute>
            }
          />
          <Route
            path="/game/:gameID"
            element={
              <PrivateRoute>
                <GamePage />
              </PrivateRoute>
            }
          />
          <Route
            path="/game/review/:gameId"
            element={
              <PrivateRoute>
                <GameReview />
              </PrivateRoute>
            }
          />
          <Route
            path="/bot-difficulty"
            element={
              <PrivateRoute>
                <BotDifficulty />
              </PrivateRoute>
            }
          />
          <Route
            path="/leaderboard"
            element={
              <PrivateRoute>
                <LeaderboardPage />
              </PrivateRoute>
            }
          />
          <Route
            path="/history"
            element={
              <PrivateRoute>
                <GameHistory />
              </PrivateRoute>
            }
          />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;
