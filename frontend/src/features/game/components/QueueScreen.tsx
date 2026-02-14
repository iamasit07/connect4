import { motion } from 'framer-motion';
import { Bot } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useState, useEffect } from 'react';
import { BotDifficultyDialog } from './BotDifficultyDialog';
import { WavyBackground } from '@/components/ui/wavy-background';
import type { BotDifficulty } from '../types';

interface QueueScreenProps {
  onCancel: () => void;
  onPlayBot: (difficulty: BotDifficulty) => void;
}

export const QueueScreen = ({ onCancel, onPlayBot }: QueueScreenProps) => {
  const [showBotDialog, setShowBotDialog] = useState(false);
  const [timeLeft, setTimeLeft] = useState(300); // 5 minutes in seconds

  useEffect(() => {
    const timer = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 0) return 0;
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, []);

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    // <WavyBackground 
    //   className="max-w-4xl mx-auto flex flex-col items-center justify-center"
    //   containerClassName="flex-1 h-full w-full items-center justify-center overflow-hidden"
    //   backgroundFill="hsl(var(--background))"
    //   waveOpacity={0.3}
    //   colors={['#38bdf8', '#818cf8', '#c084fc', '#e879f9', '#22d3ee']}
    // >
    <div className="flex-1 h-full w-full flex flex-col items-center justify-center overflow-hidden bg-background">
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        className="text-center w-full max-w-md relative z-10"
      >
        <div className="relative inline-flex items-center justify-center mb-10">
          <div className="text-5xl font-bold font-mono tracking-wider">
            {formatTime(timeLeft)}
          </div>
        </div>
        <h2 className="text-2xl font-bold mb-2">Finding Opponent...</h2>
        <p className="text-muted-foreground mb-8">
          Please wait while we match you with a worthy challenger
        </p>
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
    // </WavyBackground>
  );
};
