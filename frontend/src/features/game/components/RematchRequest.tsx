import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { RotateCcw, Check, X, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

interface RematchRequestProps {
  onSendRequest: () => void;
  onAcceptRequest: () => void;
  onDeclineRequest: () => void;
  rematchStatus: "idle" | "sent" | "received" | "accepted" | "declined";
  opponentName: string;
}

export const RematchRequest = ({
  onSendRequest,
  onAcceptRequest,
  onDeclineRequest,
  rematchStatus,
  opponentName,
}: RematchRequestProps) => {
  const [countdown, setCountdown] = useState(10);

  useEffect(() => {
    if (rematchStatus === "sent" || rematchStatus === "received") {
      setCountdown(10);
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
    }
  }, [rematchStatus]);

  useEffect(() => {
    if (countdown === 0 && rematchStatus === "received") {
      toast.info("Rematch request expired");
    }
  }, [countdown, rematchStatus]);

  if (rematchStatus === "idle") {
    return (
      <Button
        onClick={onSendRequest}
        variant="outline"
        className="gap-2"
        size="lg"
      >
        <RotateCcw className="w-4 h-4" />
        Request Rematch
      </Button>
    );
  }

  if (rematchStatus === "sent") {
    if (countdown === 0) {
      return (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="text-muted-foreground text-sm"
        >
          Rematch request timed out
        </motion.div>
      );
    }

    return (
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex items-center gap-3 px-4 py-2 rounded-lg bg-muted"
      >
        <Loader2 className="w-4 h-4 animate-spin text-primary" />
        <span className="text-sm">
          Waiting for {opponentName}... ({countdown}s)
        </span>
      </motion.div>
    );
  }

  if (rematchStatus === "received") {
    const expired = countdown === 0;

    return (
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        className="flex flex-col items-center gap-2 sm:gap-3 p-2 sm:p-4 rounded-xl bg-primary/10 ring-2 ring-primary"
      >
        <p className="text-xs sm:text-sm font-medium">
          {expired
            ? "Rematch request expired"
            : `${opponentName} wants a rematch! (${countdown}s)`}
        </p>
        {!expired && (
          <div className="flex gap-2">
            <Button
              onClick={onAcceptRequest}
              size="sm"
              className="h-7 px-2 sm:h-9 sm:px-4 gap-1 text-xs sm:text-sm"
            >
              <Check className="w-3 h-3 sm:w-4 sm:h-4" />
              Accept
            </Button>
            <Button
              onClick={onDeclineRequest}
              variant="outline"
              size="sm"
              className="h-7 px-2 sm:h-9 sm:px-4 gap-1 text-xs sm:text-sm"
            >
              <X className="w-3 h-3 sm:w-4 sm:h-4" />
              Decline
            </Button>
          </div>
        )}
      </motion.div>
    );
  }

  if (rematchStatus === "accepted") {
    return (
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="flex items-center gap-2 text-green-500"
      >
        <Check className="w-5 h-5" />
        <span>Rematch starting...</span>
      </motion.div>
    );
  }

  if (rematchStatus === "declined") {
    return (
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="text-muted-foreground text-sm"
      >
        Rematch declined
      </motion.div>
    );
  }

  return null;
};
