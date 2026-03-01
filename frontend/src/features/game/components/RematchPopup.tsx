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
    <div className="absolute inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="relative bg-card/95 backdrop-blur-xl border border-white/10 shadow-2xl rounded-3xl p-6 md:p-8 w-full max-w-sm flex flex-col items-center gap-5 sm:gap-6 animate-in fade-in zoom-in duration-300 overflow-hidden ring-1 ring-primary/20">
        
        {/* Subtle Background Glow */}
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[150%] h-24 bg-primary/20 blur-[50px] -z-10 rounded-full" />

        <div className="text-center space-y-1">
          <h3 className="text-2xl font-bold tracking-tight text-primary drop-shadow-[0_2px_10px_rgba(var(--primary),0.2)]">
            Rematch?
          </h3>
          <p className="text-muted-foreground text-sm font-medium">
            <span className="text-foreground">{opponentName}</span> wants to play again!
          </p>
        </div>
        
        <div className="flex w-full gap-3">
          <Button onClick={onDeclineRequest} variant="outline" className="flex-1 h-11 md:h-12 text-sm md:text-base gap-2 border-primary/20 hover:bg-destructive/10 hover:text-destructive hover:border-destructive/30">
             <X className="w-4 h-4" />
             Decline
          </Button>
          
          <Button onClick={onAcceptRequest} className="flex-1 bg-primary text-primary-foreground shadow-lg hover:shadow-primary/25 h-11 md:h-12 text-sm md:text-base gap-2">
            <Check className="w-4 h-4" />
            <span>Accept ({countdown}s)</span>
          </Button>
        </div>
      </div>
    </div>
  );
};
