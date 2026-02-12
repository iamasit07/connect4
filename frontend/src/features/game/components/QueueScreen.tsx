import { motion } from 'framer-motion';
import { Loader2, Bot } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { QUEUE_FUN_FACTS } from '@/lib/config';
import { useState, useEffect } from 'react';
import { BotDifficultyDialog } from './BotDifficultyDialog';
import type { BotDifficulty } from '../types';

interface QueueScreenProps {
  onCancel: () => void;
  onPlayBot: (difficulty: BotDifficulty) => void;
}

export const QueueScreen = ({ onCancel, onPlayBot }: QueueScreenProps) => {
  const [factIndex, setFactIndex] = useState(0);
  const [showBotDialog, setShowBotDialog] = useState(false);

  useEffect(() => {
    const interval = setInterval(() => {
      setFactIndex((prev) => (prev + 1) % QUEUE_FUN_FACTS.length);
    }, 5000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        className="text-center max-w-md"
      >
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 2, repeat: Infinity, ease: 'linear' }}
          className="inline-block mb-6"
        >
          <Loader2 className="w-16 h-16 text-primary" />
        </motion.div>

        <h2 className="text-2xl font-bold mb-2">Finding Opponent...</h2>
        <p className="text-muted-foreground mb-8">
          Please wait while we match you with a worthy challenger
        </p>

        <motion.div
          key={factIndex}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -10 }}
          className="bg-card p-4 rounded-lg mb-8"
        >
          <p className="text-sm text-muted-foreground italic">
            💡 {QUEUE_FUN_FACTS[factIndex]}
          </p>
        </motion.div>

        <div className="flex flex-col sm:flex-row gap-3">
          <Button variant="outline" onClick={onCancel} className="flex-1">
            Cancel
          </Button>
          <Button onClick={() => setShowBotDialog(true)} className="flex-1 gap-2">
            <Bot className="w-4 h-4" />
            Play with Bot instead
          </Button>
        </div>
      </motion.div>

      <BotDifficultyDialog
        open={showBotDialog}
        onOpenChange={setShowBotDialog}
        onSelectDifficulty={onPlayBot}
      />
    </div>
  );
};
