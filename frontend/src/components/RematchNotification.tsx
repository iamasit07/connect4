import React from "react";

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
  if (!isRequested) return null;

  return (
    <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50">
      <div className="bg-white rounded-lg shadow-xl p-6 max-w-sm w-full mx-4 animate-fade-in-up">
        <h3 className="text-xl font-bold text-gray-900 mb-2">Rematch Request</h3>
        <p className="text-gray-600 mb-6">
          <span className="font-semibold text-blue-600">{requesterName || "Opponent"}</span> wants a rematch!
        </p>
        
        {timeout !== null && (
          <div className="mb-4">
             <div className="w-full bg-gray-200 rounded-full h-2.5">
                <div 
                  className="bg-blue-600 h-2.5 rounded-full transition-all duration-1000 ease-linear" 
                  style={{ width: `${(timeout / 30) * 100}%` }}
                ></div>
             </div>
             <p className="text-xs text-gray-500 mt-1 text-right">{timeout}s remaining</p>
          </div>
        )}

        <div className="flex gap-3 justify-end">
          <button
            onClick={onDecline}
            className="px-4 py-2 border border-gray-300 rounded text-gray-700 hover:bg-gray-50 hover:text-gray-900 transition focus:outline-none focus:ring-2 focus:ring-gray-300"
          >
            Decline
          </button>
          <button
            onClick={onAccept}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
          >
            Accept
          </button>
        </div>
      </div>
    </div>
  );
};

export default RematchNotification;
