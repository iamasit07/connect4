import { motion } from "framer-motion";
import { Disk } from "./Disk";
import type { CellValue } from "../types";

interface CellProps {
  value: CellValue;
  row: number;
  col: number;
  isWinning?: boolean;
  isLastMove?: boolean;
  isHovered?: boolean;
  hoverPlayer?: number | null;
  onClick?: () => void;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
}

export const Cell = ({
  value,
  row,
  col,
  isWinning = false,
  isLastMove = false,
  isHovered = false,
  hoverPlayer = null,
  onClick,
  onMouseEnter,
  onMouseLeave,
}: CellProps) => {
  return (
    <motion.div
      className="aspect-square p-0.5 sm:p-1"
      whileHover={{ scale: value === 0 ? 1.05 : 1 }}
      onClick={onClick}
      onMouseEnter={onMouseEnter}
      onMouseLeave={onMouseLeave}
    >
      <div
        className={`
          w-full h-full rounded-full 
          bg-board-slot
          flex items-center justify-center
          transition-all duration-200
        `}
        style={{
          boxShadow:
            "inset 0 4px 8px rgba(0,0,0,0.4), inset 0 -2px 4px rgba(255,255,255,0.1)",
        }}
      >
        {value !== 0 && (
          <div className="w-[92%] h-[92%]">
            <Disk
              player={value}
              isWinning={isWinning}
              isNew={isLastMove}
              row={row}
            />
          </div>
        )}
        {value === 0 && isHovered && hoverPlayer && (
          <div className="w-[92%] h-[92%] opacity-50 pointer-events-none">
            <Disk player={hoverPlayer as 1 | 2} />
          </div>
        )}
      </div>
    </motion.div>
  );
};
