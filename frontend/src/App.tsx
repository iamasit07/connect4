import { BrowserRouter, Route, Routes } from "react-router-dom";
import "./App.css";
import LandingPage from "./pages/landing";
import GamePage from "./pages/game";
import LeaderboardPage from "./pages/leaderboard";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/game/:gameID" element={<GamePage />} />
        <Route path="/leaderboard" element={<LeaderboardPage />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
