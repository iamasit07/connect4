import React, { useEffect, useState } from "react";

interface DisconnectNotificationProps {
  isDisconnected: boolean;
  disconnectedAt: number | null;
}

const DisconnectNotification: React.FC<DisconnectNotificationProps> = ({
  isDisconnected,
  disconnectedAt,
}) => {
  const [countdown, setCountdown] = useState<number | null>(null);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    if (isDisconnected && disconnectedAt) {
      setIsVisible(true);
      const updateCountdown = () => {
        const elapsed = Math.floor((Date.now() - disconnectedAt) / 1000);
        const remaining = Math.max(0, 30 - elapsed);
        setCountdown(remaining);
      };
      
      updateCountdown(); // Initial update
      const interval = setInterval(updateCountdown, 1000);
      return () => clearInterval(interval);
    } else {
      setIsVisible(false);
      setCountdown(null);
    }
  }, [isDisconnected, disconnectedAt]);

  if (!isVisible) return null;

  return (
    <div className="fixed top-4 right-4 z-50">
      <div className="bg-amber-100 border-2 border-amber-500 text-gray-900 px-4 py-3 rounded max-w-sm">
        <div>
          <h3 className="font-bold mb-1">Opponent Disconnected</h3>
          <p className="text-sm text-gray-700">
            Waiting for reconnection...
          </p>
          {countdown !== null && (
            <p className="text-sm font-mono mt-2">
              {countdown} second{countdown !== 1 ? "s" : ""} remaining
            </p>
          )}
        </div>
      </div>
    </div>
  );
};

export default DisconnectNotification;
