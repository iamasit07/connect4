import { FC } from "react";
import { PlayerID } from "../types/game";

interface BoardProps {
  board: PlayerID[][];
  yourPlayer: PlayerID | null;
  currentTurn: PlayerID | null;
  onColumnClick: (col: number) => void;
  gameOver: boolean;
}

const Board: FC<BoardProps> = ({
  board,
  yourPlayer,
  currentTurn,
  onColumnClick,
  gameOver,
}) => {
  const isYourTurn =
    !gameOver && yourPlayer !== null && currentTurn === yourPlayer;

  const getCellColor = (cellValue: PlayerID): string => {
    if (cellValue === 0) return "bg-white";
    if (cellValue === 1) return "bg-yellow-400 shadow-md"; // Yellow for player 1
    if (cellValue === 2) return "bg-red-500 shadow-md"; // Red for player 2
    return "";
  };

  const handleColumnClick = (col: number) => {
    console.log("Board column clicked:", col, "isYourTurn:", isYourTurn);

    if (!isYourTurn) {
      console.log(
        "Not your turn! Your player:",
        yourPlayer,
        "Current turn:",
        currentTurn
      );
      return;
    }

    if (board[0][col] !== 0) {
      console.log("Column is full!");
      return;
    }

    onColumnClick(col);
  };

  return (
    <div className="bg-blue-600 p-4 rounded-lg flex gap-1 shadow-xl">
      {[0, 1, 2, 3, 4, 5, 6].map((col) => (
        <div
          key={col}
          onClick={() => handleColumnClick(col)}
          className={`flex flex-col gap-1 p-1 rounded ${
            isYourTurn
              ? "cursor-pointer hover:bg-blue-500 transition"
              : "cursor-not-allowed"
          }`}
        >
          {[0, 1, 2, 3, 4, 5].map((row) => (
            <div
              key={row}
              className={`w-14 h-14 rounded-full border-2 border-blue-800 transition ${getCellColor(
                board[row][col]
              )}`}
            />
          ))}
        </div>
      ))}
    </div>
  );
};

export default Board;
