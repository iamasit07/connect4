import React from "react";

interface PostGameNotificationProps {
  show: boolean;
  message: string | null;
}

const PostGameNotification: React.FC<PostGameNotificationProps> = ({
  show,
  message,
}) => {
  if (!show || !message) return null;

  return (
    <div className="fixed top-20 right-4 bg-blue-600 text-white px-6 py-4 rounded-lg shadow-lg z-50 animate-bounce">
      <div className="flex items-center gap-3">
        <span className="text-2xl">ℹ️</span>
        <p className="font-medium">{message}</p>
      </div>
    </div>
  );
};

export default PostGameNotification;
