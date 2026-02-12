import { motion } from 'framer-motion';
import { Disk } from './Disk';
import type { CellValue } from '../types';

interface CellProps {
  value: CellValue;
  row: number;
  col: number;
  isWinning?: boolean;
  isLastMove?: boolean;
  isHovered?: boolean;
  onClick?: () => void;
}

export const Cell = ({ 
  value, 
  row, 
  col, 
  isWinning = false, 
  isLastMove = false,
  isHovered = false,
  onClick 
}: CellProps) => {
  return (
    <motion.div
      className="aspect-square p-1 sm:p-1.5 md:p-2"
      whileHover={{ scale: value === 0 ? 1.05 : 1 }}
      onClick={onClick}
    >
      <div 
        className={`
          w-full h-full rounded-full 
          bg-board-slot
          flex items-center justify-center
          transition-all duration-200
          ${isHovered && value === 0 ? 'ring-2 ring-white/40 ring-offset-2 ring-offset-board' : ''}
        `}
        style={{
          boxShadow: 'inset 0 4px 8px rgba(0,0,0,0.4), inset 0 -2px 4px rgba(255,255,255,0.1)',
        }}
      >
        {value !== 0 && (
          <div className="w-[90%] h-[90%]">
            <Disk 
              player={value} 
              isWinning={isWinning} 
              isNew={isLastMove}
              row={row}
            />
          </div>
        )}
      </div>
    </motion.div>
  );
};
