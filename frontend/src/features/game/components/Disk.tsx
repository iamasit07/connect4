import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';

interface DiskProps {
  player: 1 | 2;
  isWinning?: boolean;
  isNew?: boolean;
  row?: number;
}

export const Disk = ({ player, isWinning = false, isNew = false, row = 0 }: DiskProps) => {
  const isRed = player === 1;
  
  // Calculate animation distance based on row position
  const dropDistance = -((row + 1) * 100);
  
  return (
    <motion.div
      initial={isNew ? { y: dropDistance, opacity: 0.8, scale: 0.9 } : false}
      animate={{ 
        y: 0, 
        opacity: 1, 
        scale: 1,
      }}
      transition={{
        type: 'spring',
        stiffness: 300,
        damping: 20,
        mass: 1,
        delay: isNew ? 0 : 0,
      }}
      className={cn(
        'w-full h-full rounded-full relative overflow-hidden',
        isRed ? 'bg-disk-red' : 'bg-disk-yellow',
        isWinning && 'win-glow'
      )}
      style={{
        boxShadow: isRed 
          ? 'inset 0 4px 8px rgba(255,255,255,0.25), inset 0 -4px 8px rgba(0,0,0,0.35), 0 2px 6px rgba(0,0,0,0.3)'
          : 'inset 0 4px 8px rgba(255,255,255,0.35), inset 0 -4px 8px rgba(0,0,0,0.25), 0 2px 6px rgba(0,0,0,0.3)',
      }}
    >
      {/* Flat disk ring effect - outer ring */}
      <div 
        className={cn(
          'absolute inset-[8%] rounded-full',
          isRed 
            ? 'bg-gradient-to-b from-red-400 to-red-600' 
            : 'bg-gradient-to-b from-yellow-300 to-yellow-500'
        )}
        style={{
          boxShadow: 'inset 0 2px 4px rgba(255,255,255,0.2), inset 0 -2px 4px rgba(0,0,0,0.15)',
        }}
      />
      
      {/* Inner depression - center hole effect */}
      <div 
        className={cn(
          'absolute inset-[25%] rounded-full',
          isRed 
            ? 'bg-gradient-to-b from-red-700 to-red-500' 
            : 'bg-gradient-to-b from-yellow-600 to-yellow-400'
        )}
        style={{
          boxShadow: 'inset 0 3px 6px rgba(0,0,0,0.3), inset 0 -2px 4px rgba(255,255,255,0.15)',
        }}
      />
      
      {/* Center highlight dot */}
      <div 
        className={cn(
          'absolute inset-[40%] rounded-full',
          isRed 
            ? 'bg-gradient-to-br from-red-400 to-red-600' 
            : 'bg-gradient-to-br from-yellow-300 to-yellow-500'
        )}
        style={{
          boxShadow: 'inset 0 1px 2px rgba(255,255,255,0.3)',
        }}
      />
      
      {/* Top edge highlight */}
      <div className="absolute top-0 left-[10%] right-[10%] h-[3px] bg-gradient-to-r from-transparent via-white/30 to-transparent rounded-full" />
    </motion.div>
  );
};
