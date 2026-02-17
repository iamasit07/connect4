import { useState, useCallback } from "react";
import { motion } from "framer-motion";
import { Cell } from "./Cell";
import { ColumnIndicator } from "./ColumnIndicator";
import { useGameStore } from "../store/gameStore";
import { BOARD_COLS, BOARD_ROWS } from "@/lib/config";

interface BoardProps {
  onColumnClick: (col: number) => void;
}

export const Board = ({ onColumnClick }: BoardProps) => {
  const [hoveredColumn, setHoveredColumn] = useState<number | null>(null);

  const {
    board,
    currentTurn,
    myPlayer,
    lastMove,
    winningCells,
    gameStatus,
    canDropInColumn,
    isMyTurn,
  } = useGameStore();

  const getLowestEmptyRow = useCallback(
    (col: number) => {
      for (let row = BOARD_ROWS - 1; row >= 0; row--) {
        if (board[row][col] === 0) {
          return row;
        }
      }
      return -1;
    },
    [board],
  );

  const ghostDiskRow = hoveredColumn !== null ? getLowestEmptyRow(hoveredColumn) : -1;

  const handleColumnClick = useCallback(
    (col: number) => {
      if (isMyTurn() && canDropInColumn(col)) {
        onColumnClick(col);
      }
    },
    [onColumnClick, isMyTurn, canDropInColumn],
  );

  const isWinningCell = useCallback(
    (row: number, col: number) => {
      if (!winningCells) return false;
      return winningCells.some((cell) => cell.row === row && cell.col === col);
    },
    [winningCells],
  );

  const isLastMoveCell = useCallback(
    (row: number, col: number) => {
      if (!lastMove) return false;
      return lastMove.row === row && lastMove.column === col;
    },
    [lastMove],
  );

  return (
    <div className="w-full h-full flex flex-col items-center justify-center min-h-0">
      {/* Column indicators */}
      {gameStatus === "playing" && (
        <div className="w-full max-w-[min(90vw,500px)] h-8 sm:h-10 mb-1 flex-shrink-0">
          <ColumnIndicator
            columns={BOARD_COLS}
            hoveredColumn={hoveredColumn}
            currentPlayer={currentTurn}
            isMyTurn={isMyTurn()}
            onColumnClick={handleColumnClick}
            onColumnHover={setHoveredColumn}
            canDropInColumn={canDropInColumn}
          />
        </div>
      )}

      {/* Game Board */}
      <div className="relative w-full max-w-[min(90vw,500px)] aspect-[7/6] flex-shrink-1 min-h-0 max-h-full">
        <motion.div
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ type: "spring", stiffness: 200, damping: 20 }}
          className="bg-board rounded-xl sm:rounded-2xl p-1.5 sm:p-3 md:p-4 board-3d shadow-xl w-full h-full"
        >
          <div className="grid grid-cols-7 gap-1 sm:gap-1.5 md:gap-2 h-full">
            {board.map((row, rowIndex) =>
              row.map((cell, colIndex) => (
                <Cell
                  key={`${rowIndex}-${colIndex}`}
                  value={cell}
                  row={rowIndex}
                  col={colIndex}
                  isWinning={isWinningCell(rowIndex, colIndex)}
                  isLastMove={isLastMoveCell(rowIndex, colIndex)}
                  isHovered={
                    hoveredColumn === colIndex &&
                    rowIndex === ghostDiskRow &&
                    cell === 0
                  }
                  hoverPlayer={currentTurn}
                  onClick={() => handleColumnClick(colIndex)}
                  onMouseEnter={() => setHoveredColumn(colIndex)}
                  onMouseLeave={() => setHoveredColumn(null)}
                />
              )),
            )}
          </div>
        </motion.div>
      </div>
    </div>
  );
};
