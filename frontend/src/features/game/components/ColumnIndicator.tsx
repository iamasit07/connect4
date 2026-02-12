import { motion, AnimatePresence } from 'framer-motion';
import { ChevronDown } from 'lucide-react';

interface ColumnIndicatorProps {
  columns: number;
  hoveredColumn: number | null;
  currentPlayer: 1 | 2;
  isMyTurn: boolean;
  onColumnClick: (col: number) => void;
  onColumnHover: (col: number | null) => void;
  canDropInColumn: (col: number) => boolean;
}

export const ColumnIndicator = ({
  columns,
  hoveredColumn,
  currentPlayer,
  isMyTurn,
  onColumnClick,
  onColumnHover,
  canDropInColumn,
}: ColumnIndicatorProps) => {
  const isRed = currentPlayer === 1;
  
  return (
    <div className="grid grid-cols-7 gap-1 sm:gap-1.5 md:gap-2 mb-2">
      {Array.from({ length: columns }).map((_, col) => {
        const canDrop = canDropInColumn(col);
        const isHovered = hoveredColumn === col;
        
        return (
          <div
            key={col}
            className={`
              aspect-square flex items-center justify-center cursor-pointer
              transition-all duration-200
              ${isMyTurn && canDrop ? 'opacity-100' : 'opacity-30 cursor-not-allowed'}
            `}
            onClick={() => isMyTurn && canDrop && onColumnClick(col)}
            onMouseEnter={() => isMyTurn && canDrop && onColumnHover(col)}
            onMouseLeave={() => onColumnHover(null)}
          >
            <AnimatePresence>
              {isHovered && isMyTurn && canDrop && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  className="flex flex-col items-center"
                >
                  <div 
                    className={`
                      w-8 h-8 sm:w-10 sm:h-10 md:w-12 md:h-12 rounded-full
                      ${isRed ? 'bg-disk-red/80' : 'bg-disk-yellow/80'}
                      flex items-center justify-center
                      animate-bounce-subtle
                    `}
                  >
                    <ChevronDown className="w-5 h-5 text-white" />
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        );
      })}
    </div>
  );
};
