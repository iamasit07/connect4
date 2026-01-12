import React, { useEffect, useState } from "react";

interface RematchNotificationProps {
  isRequested: boolean;
  requesterName: string | null;
  timeout: number | null;
  onAccept: () => void;
  onDecline: () => void;
}

const RematchNotification: React.FC<RematchNotificationProps> = ({
  isRequested,
  requesterName,
  timeout,
  onAccept,
  onDecline,
}) => {
  const [countdown, setCountdown] = useState(timeout || 0);

  useEffect(() => {
    if (isRequested && timeout) {
      setCountdown(timeout);
      const interval = setInterval(() => {
        setCountdown((prev) => Math.max(0, prev - 1));
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [isRequested, timeout]);

  if (!isRequested || !requesterName) {
    return null;
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
        <h2 className="text-2xl font-bold text-gray-900 mb-4 text-center">
          Rematch Request
        </h2>
        <p className="text-gray-700 text-center mb-4">
          <span className="font-semibold text-blue-600">{requesterName}</span>{" "}
          wants a rematch!
        </p>
        <p className="text-sm text-gray-500 text-center mb-6">
          Time remaining:{" "}
          <span className="font-bold text-red-600">{countdown}s</span>
        </p>
        <div className="flex gap-3">
          <button
            onClick={onDecline}
            className="flex-1 px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300 transition font-medium"
          >
            Decline
          </button>
          <button
            onClick={onAccept}
            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition font-medium"
          >
            Accept
          </button>
        </div>
      </div>
    </div>
  );
};

export default RematchNotification;
