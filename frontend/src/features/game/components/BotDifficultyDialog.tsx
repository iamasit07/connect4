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
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="text-center">Choose Your Opponent</DialogTitle>
        </DialogHeader>
        <div className="grid grid-cols-3 gap-3 mt-4">
          {(Object.entries(BOT_DIFFICULTIES) as [BotDifficulty, typeof BOT_DIFFICULTIES.easy][]).map(
            ([key, bot]) => (
              <Button
                key={key}
                variant="ghost" 
                className="w-full h-auto flex flex-col items-center gap-2 p-4 rounded-xl border border-border bg-card hover:bg-accent hover:border-primary/50 transition-all cursor-pointer whitespace-normal"
                onClick={() => {
                  onOpenChange(false);
                  onSelectDifficulty(key);
                }}
              >
                <span className="text-4xl">{bot.emoji}</span>
                <p className="font-semibold text-sm">{bot.name}</p>
                <p className="text-xs text-muted-foreground capitalize">{key}</p>
              </Button>
            )
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};
