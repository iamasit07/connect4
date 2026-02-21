import { useState, useEffect } from "react";
import { Check, X } from "lucide-react";
import { Button } from "@/components/ui/button";

interface RematchPopupProps {
  onAcceptRequest: () => void;
  onDeclineRequest: () => void;
  opponentName: string;
}

export const RematchPopup = ({
  onAcceptRequest,
  onDeclineRequest,
  opponentName,
}: RematchPopupProps) => {
  const TIMEOUT = 10;
  const [countdown, setCountdown] = useState(TIMEOUT);

  useEffect(() => {
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
  }, []);

  // Auto-decline on timeout
  useEffect(() => {
    if (countdown === 0) {
      onDeclineRequest();
    }
  }, [countdown, onDeclineRequest]);

  return (
    <div className="absolute inset-0 z-50 bg-background/50 backdrop-blur-[2px] flex items-center justify-center">
      <div className="bg-card border shadow-xl rounded-xl p-4 md:p-6 w-[90%] max-w-sm flex flex-col items-center gap-4 animate-in fade-in zoom-in duration-200">
        <h3 className="text-lg font-semibold text-center mb-1">
          {opponentName} requested a rematch!
        </h3>
        
        <div className="flex w-full gap-3">
          <Button onClick={onAcceptRequest} className="flex-1 bg-primary text-primary-foreground hover:bg-primary/90 h-10 md:h-12 text-sm md:text-base gap-2">
            <Check className="w-4 h-4" />
            <span>Accept ({countdown}s)</span>
          </Button>
          
          <Button onClick={onDeclineRequest} variant="outline" className="flex-1 h-10 md:h-12 text-sm md:text-base gap-2">
             <X className="w-4 h-4" />
             Decline
          </Button>
        </div>
      </div>
    </div>
  );
};
