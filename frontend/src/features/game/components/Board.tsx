import { useState, useCallback } from 'react';
import { motion } from 'framer-motion';
import { Cell } from './Cell';
import { ColumnIndicator } from './ColumnIndicator';
import { useGameStore } from '../store/gameStore';
import { BOARD_COLS } from '@/lib/config';

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

  const handleColumnClick = useCallback((col: number) => {
    if (isMyTurn() && canDropInColumn(col)) {
      onColumnClick(col);
    }
  }, [onColumnClick, isMyTurn, canDropInColumn]);

  const isWinningCell = useCallback((row: number, col: number) => {
    if (!winningCells) return false;
    return winningCells.some(cell => cell.row === row && cell.col === col);
  }, [winningCells]);

  const isLastMoveCell = useCallback((row: number, col: number) => {
    if (!lastMove) return false;
    return lastMove.row === row && lastMove.column === col;
  }, [lastMove]);

  return (
    <div className="w-full max-w-[min(90vw,500px)] mx-auto">
      {/* Column indicators */}
      {gameStatus === 'playing' && (
        <ColumnIndicator
          columns={BOARD_COLS}
          hoveredColumn={hoveredColumn}
          currentPlayer={currentTurn}
          isMyTurn={isMyTurn()}
          onColumnClick={handleColumnClick}
          onColumnHover={setHoveredColumn}
          canDropInColumn={canDropInColumn}
        />
      )}
      
      {/* Game Board */}
      <motion.div
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ type: 'spring', stiffness: 200, damping: 20 }}
        className="bg-board rounded-xl sm:rounded-2xl p-2 sm:p-3 md:p-4 board-3d"
      >
        <div className="grid grid-cols-7 gap-1 sm:gap-1.5 md:gap-2">
          {board.map((row, rowIndex) =>
            row.map((cell, colIndex) => (
              <Cell
                key={`${rowIndex}-${colIndex}`}
                value={cell}
                row={rowIndex}
                col={colIndex}
                isWinning={isWinningCell(rowIndex, colIndex)}
                isLastMove={isLastMoveCell(rowIndex, colIndex)}
                isHovered={hoveredColumn === colIndex && cell === 0}
                onClick={() => handleColumnClick(colIndex)}
              />
            ))
          )}
        </div>
      </motion.div>
    </div>
  );
};
