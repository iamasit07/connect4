import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

interface ErrorNotificationProps {
  show: boolean;
  triggeredAt: number | null;
  title?: string;
  reason?: string;
}

const ErrorNotification: React.FC<ErrorNotificationProps> = ({
  show,
  triggeredAt,
  title = "Error",
  reason,
}) => {
  const [, forceUpdate] = useState(0);
  const navigate = useNavigate();

  useEffect(() => {
    if (!show || !triggeredAt) return;

    const interval = setInterval(() => {
      const elapsed = Math.floor((Date.now() - triggeredAt) / 1000);
      const countdown = Math.max(0, 10 - elapsed);

      if (countdown === 0) {
        clearInterval(interval);
        localStorage.removeItem("gameID");
        localStorage.removeItem("username");
        localStorage.removeItem("isReconnecting");
        navigate("/");
        return;
      }

      forceUpdate((n) => n + 1);
    }, 1000);

    return () => clearInterval(interval);
  }, [show, triggeredAt, navigate]);

  if (!show || !triggeredAt) return null;

  const elapsed = Math.floor((Date.now() - triggeredAt) / 1000);
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
          <h2 className="text-2xl font-bold text-gray-900 mb-4">{title}</h2>
          {reason && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
              <p className="text-sm font-semibold text-red-800 mb-1">Reason:</p>
              <p className="text-sm text-red-700">{reason}</p>
            </div>
          )}
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

export default ErrorNotification;
