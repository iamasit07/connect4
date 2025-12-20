import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

interface MatchEndedNotificationProps {
  matchEnded: boolean;
  matchEndedAt: number | null;
}

const MatchEndedNotification: React.FC<MatchEndedNotificationProps> = ({
  matchEnded,
  matchEndedAt,
}) => {
  const [, forceUpdate] = useState(0);
  const navigate = useNavigate();

  useEffect(() => {
    if (!matchEnded || !matchEndedAt) return;

    const interval = setInterval(() => {
      const elapsed = Math.floor((Date.now() - matchEndedAt) / 1000);
      const countdown = Math.max(0, 10 - elapsed);

      // Check if countdown has reached 0
      if (countdown === 0) {
        clearInterval(interval);
        localStorage.removeItem("gameID");
        localStorage.removeItem("username");
        localStorage.removeItem("isReconnecting");
        navigate("/");
        return;
      }

      // Force re-render to update countdown display
      forceUpdate((n) => n + 1);
    }, 1000);

    return () => clearInterval(interval);
  }, [matchEnded, matchEndedAt, navigate]);

  if (!matchEnded || !matchEndedAt) return null;

  const elapsed = Math.floor((Date.now() - matchEndedAt) / 1000);
  const countdown = Math.max(0, 10 - elapsed);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl p-8 max-w-md mx-4">
        <div className="text-center">
          <div className="mb-4">
            <svg
              className="mx-auto h-16 w-16 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">Match Ended</h2>
          <p className="text-gray-600 mb-6">
            This game session has ended or no longer exists. You will be
            redirected to the home page.
          </p>
          <div className="bg-gray-100 rounded-lg p-4">
            <p className="text-sm text-gray-700 mb-2">Redirecting in:</p>
            <p className="text-4xl font-bold text-blue-600">
              {countdown} second{countdown !== 1 ? "s" : ""}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default MatchEndedNotification;
