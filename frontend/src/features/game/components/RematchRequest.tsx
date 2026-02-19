import { useState, useEffect, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Check, X } from "lucide-react";
import { Button } from "@/components/ui/button";

interface RematchRequestProps {
  onSendRequest: () => void;
  onAcceptRequest: () => void;
  onDeclineRequest: () => void;
  rematchStatus: 'idle' | 'sent' | 'received' | 'accepted' | 'declined';
  opponentName: string;
}

interface RematchOverlayProps {
  onAccept: () => void;
  onDecline: () => void;
  opponentName: string;
}

export const RematchOverlay = ({
  onAccept,
  onDecline,
  opponentName,
}: RematchOverlayProps) => {
  const TIMEOUT = 10;
  const [countdown, setCountdown] = useState(TIMEOUT);
  const [dismissed, setDismissed] = useState(false);

  const handleDecline = useCallback(() => {
    setDismissed(true);
    onDecline();
  }, [onDecline]);

  const handleAccept = useCallback(() => {
    setDismissed(true);
    onAccept();
  }, [onAccept]);

  // Countdown timer
  useEffect(() => {
    setCountdown(TIMEOUT);
    setDismissed(false);
    const interval = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  // Auto-decline on timeout
  useEffect(() => {
    if (countdown === 0 && !dismissed) {
      handleDecline();
    }
  }, [countdown, dismissed, handleDecline]);

  if (dismissed) return null;

  const progress = (countdown / TIMEOUT) * 100;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -20 }}
        className="absolute inset-0 z-30 flex items-center justify-center bg-black/50 backdrop-blur-sm rounded-xl sm:rounded-2xl"
      >
        <motion.div
          initial={{ scale: 0.9 }}
          animate={{ scale: 1 }}
          className="bg-card border border-border rounded-xl p-4 sm:p-6 shadow-2xl w-[85%] max-w-xs text-center space-y-3"
        >
          {/* Timer bar */}
          <div className="w-full h-1.5 bg-muted rounded-full overflow-hidden">
            <motion.div
              className="h-full bg-primary rounded-full"
              initial={{ width: "100%" }}
              animate={{ width: `${progress}%` }}
              transition={{ duration: 0.3 }}
            />
          </div>

          <p className="text-sm sm:text-base font-medium">
            <span className="text-primary">{opponentName}</span> wants a
            rematch!
          </p>
          <p className="text-xs text-muted-foreground">
            {countdown}s remaining
          </p>

          <div className="flex gap-2 justify-center pt-1">
            <Button onClick={handleAccept} size="sm" className="gap-1.5 px-4">
              <Check className="w-4 h-4" />
              Accept
            </Button>
            <Button
              onClick={handleDecline}
              variant="outline"
              size="sm"
              className="gap-1.5 px-4"
            >
              <X className="w-4 h-4" />
              Decline
            </Button>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
};

export const RematchRequest = ({
  onSendRequest,
  onAcceptRequest,
  onDeclineRequest,
  rematchStatus,
  opponentName,
}: RematchRequestProps) => {
  if (rematchStatus === 'received') {
    return (
      <div className="relative h-32 w-full">
         <RematchOverlay
           onAccept={onAcceptRequest}
           onDecline={onDeclineRequest}
           opponentName={opponentName}
         />
      </div>
    );
  }

  if (rematchStatus === 'sent') {
    return (
      <Button disabled variant="outline" className="gap-2 w-full">
        Waiting for opponent...
      </Button>
    );
  }

  if (rematchStatus === 'accepted') {
    return (
       <Button disabled className="gap-2 bg-green-600 w-full text-white">
         <Check className="w-4 h-4" />
         Rematch Accepted!
       </Button>
    );
  }

  return (
    <Button onClick={onSendRequest} className="gap-2 w-full" variant="secondary">
      Request Rematch
    </Button>
  );
};
