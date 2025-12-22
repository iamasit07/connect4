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
      <div className="bg-white rounded-lg shadow-lg p-6 w-80">
        <div className="text-center space-y-3">
          <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
          
          {reason && (
            <p className="text-sm text-gray-600">{reason}</p>
          )}
          
          <div className="pt-2">
            <p className="text-xs text-gray-500 mb-1">Redirecting in</p>
            <p className="text-3xl font-bold text-blue-600">
              {countdown}s
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ErrorNotification;
