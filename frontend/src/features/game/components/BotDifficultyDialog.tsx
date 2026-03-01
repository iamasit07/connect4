import { Button } from '@/components/ui/button';
import { BOT_DIFFICULTIES } from '@/lib/config';
import type { BotDifficulty } from '../types';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface BotDifficultyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelectDifficulty: (difficulty: BotDifficulty) => void;
}

export const BotDifficultyDialog = ({
  open,
  onOpenChange,
  onSelectDifficulty,
}: BotDifficultyDialogProps) => {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle className="text-center">Choose Your Opponent</DialogTitle>
        </DialogHeader>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mt-6">
          {(Object.entries(BOT_DIFFICULTIES) as [BotDifficulty, typeof BOT_DIFFICULTIES.easy][]).map(
            ([key, bot]) => {
              // Assign a gentle theme color based on difficulty
              const colorClass = 
                key === 'easy' ? 'hover:border-green-500/50 hover:bg-green-500/10 hover:shadow-green-500/20' :
                key === 'medium' ? 'hover:border-yellow-500/50 hover:bg-yellow-500/10 hover:shadow-yellow-500/20' :
                'hover:border-red-500/50 hover:bg-red-500/10 hover:shadow-red-500/20';

              return (
                <button
                  key={key}
                  className={`w-full flex sm:flex-col items-center gap-4 sm:gap-3 p-5 rounded-2xl border border-border/50 bg-card/50 backdrop-blur-sm shadow-sm hover:shadow-xl transition-all duration-300 hover:-translate-y-1 cursor-pointer text-left sm:text-center group text-foreground dark:text-foreground ${colorClass}`}
                  onClick={() => {
                    onOpenChange(false);
                    onSelectDifficulty(key);
                  }}
                >
                  <div className="text-4xl sm:text-5xl bg-background rounded-full w-14 h-14 sm:w-16 sm:h-16 flex items-center justify-center shadow-inner group-hover:scale-110 transition-transform duration-300 shrink-0">
                    {bot.emoji}
                  </div>
                  <div className="flex-1">
                    <p className="font-bold text-lg mb-0.5">{bot.name}</p>
                    <p className="text-sm text-foreground/70 capitalize">{key}</p>
                  </div>
                </button>
              );
            }
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};
