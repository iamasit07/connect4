import { useState, useEffect } from "react";
import { Check, X } from "lucide-react";
import { Button } from "@/components/ui/button";

interface RematchRequestProps {
  onSendRequest: () => void;
  onAcceptRequest: () => void;
  onDeclineRequest: () => void;
  rematchStatus: 'idle' | 'sent' | 'received' | 'accepted' | 'declined';
  opponentName: string;
}

export const RematchRequest = ({
  onSendRequest,
  onAcceptRequest,
  onDeclineRequest,
  rematchStatus,
  opponentName,
}: RematchRequestProps) => {
  const TIMEOUT = 10;
  const [countdown, setCountdown] = useState(TIMEOUT);

  useEffect(() => {
    if (rematchStatus === 'received') {
      setCountdown(TIMEOUT);
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

  // Auto-decline on timeout
  useEffect(() => {
    if (rematchStatus === 'received' && countdown === 0) {
      onDeclineRequest();
    }
  }, [countdown, rematchStatus, onDeclineRequest]);

  if (rematchStatus === 'received') {
    return (
      <div className="flex gap-2 w-full items-center">
        <Button onClick={onAcceptRequest} size="sm" className="flex-1 gap-1.5 h-9 sm:h-11 bg-primary text-primary-foreground hover:bg-primary/90 text-sm sm:text-base">
          <Check className="w-3.5 h-3.5 sm:w-4 sm:h-4" />
          <span className="truncate">Accept ({countdown}s)</span>
        </Button>
        <Button onClick={onDeclineRequest} variant="outline" size="sm" className="flex-x gap-1.5 h-9 sm:h-11 px-3">
          <X className="w-3.5 h-3.5 sm:w-4 sm:h-4" />
        </Button>
      </div>
    );
  }

  if (rematchStatus === 'sent') {
    return (
      <Button disabled variant="outline" className="gap-2 w-full h-9 sm:h-11 text-sm sm:text-base">
        Waiting for {opponentName}...
      </Button>
    );
  }

  if (rematchStatus === 'accepted') {
    return (
       <Button disabled className="gap-2 bg-green-600 w-full text-white h-9 sm:h-11 text-sm sm:text-base">
         <Check className="w-3.5 h-3.5 sm:w-4 sm:h-4" />
         Accepted!
       </Button>
    );
  }

  return (
    <Button onClick={onSendRequest} className="gap-2 w-full h-9 sm:h-11 text-sm sm:text-base" variant="default">
      Request Rematch
    </Button>
  );
};
