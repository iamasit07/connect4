import { motion, AnimatePresence } from 'framer-motion';
import { Disk } from './Disk';

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

  
  return (
    <div className="grid grid-cols-7 gap-1 sm:gap-1.5 md:gap-2 mb-2 px-1.5 sm:px-3 md:px-4">
      {Array.from({ length: columns }).map((_, col) => {
        const canDrop = canDropInColumn(col);
        const isHovered = hoveredColumn === col;
        
        return (
          <div
            key={col}
            className={`
              aspect-square flex items-center justify-center cursor-pointer
              transition-all duration-200
              ${isMyTurn && canDrop ? 'opacity-100' : 'opacity-0'}
            `}
            onClick={() => isMyTurn && canDrop && onColumnClick(col)}
            onMouseEnter={() => isMyTurn && canDrop && onColumnHover(col)}
            onMouseLeave={() => onColumnHover(null)}
          >
            <AnimatePresence>
              {isHovered && isMyTurn && canDrop && (
                <motion.div
                  initial={{ opacity: 0, y: -20 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -20 }}
                  transition={{ type: "spring", stiffness: 300, damping: 20 }}
                  className="w-full h-full p-[10%]"
                >
                  <Disk player={currentPlayer} isNew={false} />
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        );
      })}
    </div>
  );
};
