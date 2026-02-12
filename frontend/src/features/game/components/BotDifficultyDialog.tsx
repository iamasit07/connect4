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
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="text-center">Choose Your Opponent</DialogTitle>
        </DialogHeader>
        <div className="grid gap-3 mt-4">
          {(Object.entries(BOT_DIFFICULTIES) as [BotDifficulty, typeof BOT_DIFFICULTIES.easy][]).map(
            ([key, bot]) => (
              <Button
                key={key}
                variant="outline"
                className="h-auto py-4 justify-start gap-4"
                onClick={() => {
                  onOpenChange(false);
                  onSelectDifficulty(key);
                }}
              >
                <span className="text-2xl">{bot.emoji}</span>
                <div className="text-left">
                  <p className="font-semibold">{bot.name}</p>
                  <p className="text-xs text-muted-foreground">{bot.description}</p>
                </div>
              </Button>
            )
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};
