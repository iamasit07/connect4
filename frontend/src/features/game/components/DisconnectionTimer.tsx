import React, { useEffect } from "react";
import { useGameStore } from "../store/gameStore";
import { AlertTriangle } from "lucide-react";

export const DisconnectionTimer: React.FC = () => {
  const isOpponentDisconnected = useGameStore(state => state.isOpponentDisconnected);
  const disconnectTimer = useGameStore(state => state.disconnectTimer);

  useEffect(() => {
    if (!isOpponentDisconnected || disconnectTimer <= 0) return;

    const interval = setInterval(() => {
      useGameStore.setState(state => ({
        disconnectTimer: Math.max(0, state.disconnectTimer - 1),
      }));
    }, 1000);

    return () => clearInterval(interval);
  }, [isOpponentDisconnected]);

  if (!isOpponentDisconnected) return null;

  return (
    <div className="fixed top-4 right-4 z-50 animate-in slide-in-from-right fade-in duration-300">
      <div className="bg-destructive/10 backdrop-blur-md border border-destructive/50 text-destructive px-6 py-4 rounded-xl shadow-lg flex items-center gap-4 max-w-sm">
        <div className="p-3 bg-destructive/20 rounded-full animate-pulse">
          <AlertTriangle className="w-6 h-6" />
        </div>
        <div className="flex flex-col">
          <span className="font-bold text-lg">Opponent Disconnected</span>
          <span className="text-sm opacity-90">
            Game ends in <span className="font-mono font-bold text-xl ml-1">{disconnectTimer}s</span>
          </span>
        </div>
      </div>
    </div>
  );
};
