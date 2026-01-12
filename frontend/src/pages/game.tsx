import React, { useEffect, useRef } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Board from "../components/board";
import DisconnectNotification from "../components/DisconnectNotification";
import ErrorNotification from "../components/ErrorNotification";
import RematchNotification from "../components/RematchNotification";
import PostGameNotification from "../components/PostGameNotification";
import useWebSocket from "../hooks/useWebSocket";

const GamePage: React.FC = () => {
  const { gameID: urlGameID } = useParams<{ gameID: string }>();
  const { connected, gameState, joinQueue, makeMove, reconnect, requestRematch, respondToRematch, justReceivedGameStart, showPostGameNotification, postGameMessage } =
    useWebSocket();
  const navigate = useNavigate();
  const hasJoinedQueue = useRef(false);
  const [, forceUpdate] = React.useState(0);
  
  // Extract difficulty from URL query parameter
  const searchParams = new URLSearchParams(window.location.search);
  const difficulty = searchParams.get("difficulty") || ""; // Empty for online matchmaking

  useEffect(() => {
    if (gameState.gameId && urlGameID !== gameState.gameId) {
      window.history.replaceState(null, "", `/game/${gameState.gameId}`);
    }
  }, [gameState.gameId, urlGameID]);



  useEffect(() => {
    if (urlGameID && urlGameID !== "queue") {
      // Only reconnect if we don't already have this game loaded
      // AND we didn't just receive a game_start message (which already has the game state)
      if (connected && !hasJoinedQueue.current && gameState.gameId !== urlGameID && !justReceivedGameStart.current) {
        hasJoinedQueue.current = true;

        setTimeout(() => {
          reconnect(urlGameID);
        }, 100);
      }
      return;
    }
    if (urlGameID === "queue") {
      if (connected && !hasJoinedQueue.current) {
        hasJoinedQueue.current = true;

        localStorage.removeItem("gameID");

        setTimeout(() => {
          joinQueue(difficulty);
        }, 100);
      }
      return;
    }
    if (!urlGameID) {
      if (connected && !hasJoinedQueue.current) {
        hasJoinedQueue.current = true;

        setTimeout(() => {
          reconnect("");
        }, 100);
      }
      return;
    }

    console.warn("Invalid game route, redirecting to home");
    navigate("/");
  }, [connected, navigate, joinQueue, reconnect, urlGameID, difficulty]);

  // Reset hasJoinedQueue when URL changes (so we can reconnect to different games)
  useEffect(() => {
    hasJoinedQueue.current = false;
  }, [urlGameID]);

  useEffect(() => {
    if (gameState.inQueue && gameState.queuedAt) {
      const interval = setInterval(() => {
        forceUpdate((n) => n + 1);
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [gameState.inQueue, gameState.queuedAt]);

  const queueCountdown =
    gameState.inQueue && gameState.queuedAt
      ? Math.max(0, 30 - Math.floor((Date.now() - gameState.queuedAt) / 1000))
      : null;

  // Redirect to bot difficulty page when matchmaking times out (only for online matchmaking)
  useEffect(() => {
    if (queueCountdown === 0 && gameState.inQueue && !difficulty) {
      // Timeout for online matchmaking - redirect to bot difficulty selection
      navigate("/bot-difficulty");
    }
  }, [queueCountdown, gameState.inQueue, difficulty, navigate]);

  const handleColumnClick = (col: number) => {
    makeMove(col);
  };

  const handlePlayAgain = () => {
    navigate("/");
  };

  const handleRematchRequest = () => {
    requestRematch();
  };

  const handleRematchAccept = () => {
    respondToRematch(true);
  };

  const handleRematchDecline = () => {
    respondToRematch(false);
  };

  const getBackgroundColor = () => {
    if (gameState.gameOver) return "bg-gray-50";
    if (gameState.currentTurn === 1) return "bg-yellow-50";
    if (gameState.currentTurn === 2) return "bg-red-50";
    return "bg-gray-50";
  };

  if (!connected && !gameState.gameOver && !gameState.matchEnded) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <p className="text-gray-600">Connecting...</p>
      </div>
    );
  }

  if (gameState.inQueue) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="text-center">
          <p className="text-lg text-gray-800">Finding opponent...</p>
          <p className="text-sm text-gray-500 mt-2">
            {queueCountdown !== null ? (
              <span>
                Bot joins in{" "}
                <span className="font-bold text-blue-600">
                  {queueCountdown} second{queueCountdown !== 1 ? "s" : ""}
                </span>
              </span>
            ) : (
              "Waiting..."
            )}
          </p>
        </div>
      </div>
    );
  }

  if (gameState.matchEnded) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <ErrorNotification
          show={gameState.matchEnded}
          triggeredAt={gameState.matchEndedAt}
          title={gameState.reason ? "Error" : "Match Ended"}
          reason={gameState.reason ?? undefined}
        />
      </div>
    );
  }

  if (!gameState.gameId) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <p className="text-gray-600">Waiting...</p>
      </div>
    );
  }

  return (
    <div
      className={`flex flex-col items-center justify-center min-h-screen ${getBackgroundColor()} gap-6 p-4`}
    >
      {gameState.gameId && (
        <div className="text-xs text-gray-500 font-mono">
          Game ID: {gameState.gameId}
        </div>
      )}

      {!gameState.gameOver && gameState.currentTurn && (
        <div
          className={`w-full max-w-md px-4 py-2 rounded text-center font-bold ${
            gameState.currentTurn === 1
              ? "bg-yellow-400 text-yellow-900"
              : "bg-red-500 text-white"
          }`}
        >
          Player {gameState.currentTurn}'s Turn
          {gameState.currentTurn === gameState.yourPlayer && " (You)"}
        </div>
      )}

      <div className="text-center">
        {gameState.gameOver ? (
          <h2 className="text-2xl font-bold text-gray-900">
            {gameState.winner === "draw"
              ? "Draw!"
              : `${gameState.winner} Wins!`}
          </h2>
        ) : (
          <div className="space-y-2">
            <div className="flex items-center justify-center gap-4 text-sm text-gray-700 bg-white px-3 py-2 rounded border border-gray-200">
              <span className="flex items-center gap-2">
                <span className="inline-block w-4 h-4 rounded-full bg-yellow-400"></span>
                <span>
                  Player 1 {gameState.yourPlayer === 1 && "(You)"}
                  {gameState.yourPlayer === 2 && `(${gameState.opponent})`}
                </span>
              </span>
              <span className="flex items-center gap-2">
                <span className="inline-block w-4 h-4 rounded-full bg-red-500"></span>
                <span>
                  Player 2 {gameState.yourPlayer === 2 && "(You)"}
                  {gameState.yourPlayer === 1 && `(${gameState.opponent})`}
                </span>
              </span>
            </div>
          </div>
        )}
      </div>

      <Board
        board={gameState.board}
        yourPlayer={gameState.yourPlayer}
        currentTurn={gameState.currentTurn}
        onColumnClick={handleColumnClick}
        gameOver={gameState.gameOver}
      />

      
      {gameState.gameOver && (
        <div className="flex gap-3">
          {!showPostGameNotification && (
            <button
              onClick={handleRematchRequest}
              className="px-6 py-2 bg-green-600 text-white rounded hover:bg-green-700 transition"
            >
              Request Rematch
            </button>
          )}
          <button
            onClick={handlePlayAgain}
            className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
          >
            Back to Home
          </button>
        </div>
      )}
      

      <DisconnectNotification
        isDisconnected={gameState.opponentDisconnected}
        disconnectedAt={gameState.disconnectedAt}
      />

      <RematchNotification
        isRequested={gameState.rematchRequested}
        requesterName={gameState.rematchRequester}
        timeout={gameState.rematchTimeout}
        onAccept={handleRematchAccept}
        onDecline={handleRematchDecline}
      />

      <ErrorNotification
        show={gameState.matchEnded}
        triggeredAt={gameState.matchEndedAt}
        title={gameState.reason ? "Error" : "Match Ended"}
        reason={gameState.reason ?? undefined}
      />

      <PostGameNotification
        show={showPostGameNotification}
        message={postGameMessage}
      />

      {gameState.error && (
        <div className="fixed top-4 left-1/2 transform -translate-x-1/2 bg-red-500 text-white px-6 py-3 rounded-lg shadow-lg z-50 max-w-md text-center">
          {gameState.error}
        </div>
      )}
    </div>
  );
};

export default GamePage;
