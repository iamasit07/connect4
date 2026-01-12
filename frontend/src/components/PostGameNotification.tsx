import React, { useEffect, useState } from "react";

interface PostGameNotificationProps {
  show: boolean;
  message: string;
  onClose?: () => void;
}

const PostGameNotification: React.FC<PostGameNotificationProps> = ({
  show,
  message,
  onClose,
}) => {
  const [visible, setVisible] = useState(show);

  useEffect(() => {
    if (show) {
      setVisible(true);
      // Auto-dismiss after 5 seconds
      const timer = setTimeout(() => {
        setVisible(false);
        if (onClose) onClose();
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [show, onClose]);

  if (!visible) return null;

  return (
    <div className="fixed top-4 left-1/2 transform -translate-x-1/2 z-50 animate-fade-in">
      <div className="bg-blue-500 text-white px-6 py-4 rounded-lg shadow-lg max-w-md">
        <div className="flex items-center gap-3">
          <div className="flex-shrink-0">
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>
          <div className="flex-1">
            <p className="font-medium">{message}</p>
            <p className="text-sm text-blue-100 mt-1">Board is now read-only</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PostGameNotification;
