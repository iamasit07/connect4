import { useState, useEffect } from "react";
import { Check, X } from "lucide-react";
import { Button } from "@/components/ui/button";

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
  if (rematchStatus === 'received') {
    return (
       <Button disabled className="gap-2 w-full h-9 sm:h-11 text-sm sm:text-base">
         Respond to {opponentName}...
       </Button>
    );
  }

  if (rematchStatus === 'declined') {
    return (
       <Button disabled variant="outline" className="gap-2 w-full h-9 sm:h-11 text-sm sm:text-base border-red-200 bg-red-50/50 text-red-600 dark:border-red-900/50 dark:bg-red-900/10 dark:text-red-400">
         Rematch Declined
       </Button>
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
