import { motion } from 'framer-motion';
import { Eye, Users } from 'lucide-react';
import { cn } from '@/lib/utils';

interface SpectatorBadgeProps {
  count: number;
  isSpectator?: boolean;
}

export const SpectatorBadge = ({ count, isSpectator = false }: SpectatorBadgeProps) => {
  if (count === 0 && !isSpectator) return null;

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      className={cn(
        'flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium',
        isSpectator 
          ? 'bg-purple-500/20 text-purple-600 dark:text-purple-400'
          : 'bg-muted text-muted-foreground'
      )}
    >
      {isSpectator ? (
        <>
          <Eye className="w-3.5 h-3.5" />
          <span>Spectating</span>
        </>
      ) : (
        <>
          <Users className="w-3.5 h-3.5" />
          <span>{count} watching</span>
        </>
      )}
    </motion.div>
  );
};
