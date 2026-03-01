import { Check, RotateCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useState, useEffect } from "react";

interface RematchRequestProps {
  onSendRequest: () => void;
  rematchStatus: 'idle' | 'sent' | 'received' | 'accepted' | 'declined';
  opponentName: string;
}

export const RematchRequest = ({
  onSendRequest,
  rematchStatus,
  opponentName,
}: RematchRequestProps) => {
  const [countdown, setCountdown] = useState(10);

  useEffect(() => {
    if (rematchStatus === 'sent') {
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

  if (rematchStatus === 'received') {
    return (
       <Button disabled variant="secondary" className="gap-2 w-full h-12 text-base font-semibold opacity-70">
         Respond to {opponentName}...
       </Button>
    );
  }

  if (rematchStatus === 'declined') {
    return (
       <Button disabled variant="outline" className="gap-2 w-full h-12 text-base font-semibold border-destructive/30 bg-destructive/10 text-destructive/90">
         Rematch Declined
       </Button>
    );
  }

  if (rematchStatus === 'sent') {
    if (countdown === 0) {
      return (
        <Button disabled variant="outline" className="gap-2 w-full h-12 text-base font-semibold border-destructive/30 bg-destructive/10 text-destructive/90">
          Request Expired
        </Button>
      );
    }
    return (
      <Button disabled variant="outline" className="gap-2 w-full h-12 text-base font-semibold bg-muted/30">
        Waiting for {opponentName} ({countdown}s)...
      </Button>
    );
  }

  if (rematchStatus === 'accepted') {
    return (
       <Button disabled className="gap-2 bg-green-600/90 text-white w-full h-12 text-base font-semibold">
         <Check className="w-4 h-4" />
         Accepted!
       </Button>
    );
  }

  return (
    <Button onClick={onSendRequest} className="gap-2 w-full h-12 text-base font-semibold bg-primary hover:bg-primary/90 shadow-md hover:shadow-xl hover:shadow-primary/20 transition-all">
      <RotateCcw className="w-4 h-4" />
      Request Rematch
    </Button>
  );
};
