import React, { useEffect, useState } from "react";

interface DisconnectNotificationProps {
  isDisconnected: boolean;
  disconnectedAt: number | null;
}

const DisconnectNotification: React.FC<DisconnectNotificationProps> = ({
  isDisconnected,
  disconnectedAt,
}) => {
  const [, forceUpdate] = useState(0);

  useEffect(() => {
    if (!isDisconnected || !disconnectedAt) return;

    const interval = setInterval(() => {
      forceUpdate((n) => n + 1);
    }, 1000);

    return () => clearInterval(interval);
  }, [isDisconnected, disconnectedAt]);

  if (!isDisconnected || !disconnectedAt) return null;

  const elapsed = Math.floor((Date.now() - disconnectedAt) / 1000);
  const countdown = Math.max(0, 30 - elapsed);

  return (
    <div className="fixed top-4 right-4 z-50">
      <div className="bg-amber-100 border-2 border-amber-500 text-gray-900 px-4 py-3 rounded max-w-sm">
        <div>
          <h3 className="font-bold mb-1">Opponent Disconnected</h3>
          <p className="text-sm text-gray-700">Waiting for reconnection...</p>
          <p className="text-sm font-mono mt-2">
            {countdown} second{countdown !== 1 ? "s" : ""} remaining
          </p>
        </div>
      </div>
    </div>
  );
};

export default DisconnectNotification;
