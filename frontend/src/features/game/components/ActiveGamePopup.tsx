import React, { useState, useEffect } from "react";
import { useAuthStore } from "@/features/auth/store/authStore";
import { useGameStore } from "@/features/game/store/gameStore";
import { useLocation, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { AlertCircle } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useGameSocket } from "../hooks/useGameSocket";

export const ActiveGamePopup: React.FC = () => {
  const { user } = useAuthStore();
  const { 
    gameId: activeGameId, 
    isActiveGamePopupDismissed, 
    dismissActiveGamePopup 
  } = useGameStore(); 
  
  const location = useLocation();
  const navigate = useNavigate();
  const { surrender, connect } = useGameSocket(); 
  
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    if (!user?.activeGameId || isActiveGamePopupDismissed) return;

    const isOnGamePage = location.pathname === `/game/${user.activeGameId}`;
    if (isOnGamePage) return;

    const authPages = ["/login", "/signup", "/auth/callback", "/complete-signup"];
    if (authPages.includes(location.pathname)) return;

    setIsOpen(true);
  }, [user?.activeGameId, location.pathname, isActiveGamePopupDismissed]);

  const handleReconnect = () => {
    const targetGameId = user?.activeGameId || activeGameId;
    if (targetGameId) {
      setIsOpen(false);
      navigate(`/game/${targetGameId}`);
    }
  };

  const handleAbandon = async () => {
    try {
        await connect();
        setTimeout(() => {
            surrender();
            dismissActiveGamePopup();
            setIsOpen(false);
        }, 100);
    } catch (error) {
        console.error("Failed to connect for abandon:", error);
        dismissActiveGamePopup();
        setIsOpen(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={() => {}}>
      <DialogContent className="sm:max-w-md [&>button]:hidden" onInteractOutside={(e) => e.preventDefault()} onEscapeKeyDown={(e) => e.preventDefault()}>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-amber-600">
            <AlertCircle className="w-5 h-5" />
            Active Game in Progress
          </DialogTitle>
          <DialogDescription>
            You have an active game running. Leaving it will result in a forfeit.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2 sm:justify-end">
          <Button variant="destructive" onClick={handleAbandon}>
            Abandon Game
          </Button>
          <Button onClick={handleReconnect} className="bg-amber-600 hover:bg-amber-700 text-white">
            Reconnect
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
