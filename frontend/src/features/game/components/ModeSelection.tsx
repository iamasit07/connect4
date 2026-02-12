import { motion } from 'framer-motion';
import { Users, Ghost } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import type { BotDifficulty } from '../types';
import { useState } from 'react';
import { BotDifficultyDialog } from './BotDifficultyDialog';

interface ModeSelectionProps {
  onSelectPvP: () => void;
  onSelectBot: (difficulty: BotDifficulty) => void;
}

export const ModeSelection = ({ onSelectPvP, onSelectBot }: ModeSelectionProps) => {
  const [showBotDialog, setShowBotDialog] = useState(false);

  return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center mb-8"
      >
        <h1 className="text-4xl sm:text-5xl font-bold text-primary mb-2">Connect 4</h1>
        <p className="text-muted-foreground">Choose your game mode</p>
      </motion.div>

      <div className="grid gap-4 sm:grid-cols-2 max-w-2xl w-full">
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.1 }}
        >
          <Card 
            className="p-6 cursor-pointer hover:ring-2 hover:ring-primary transition-all group"
            onClick={onSelectPvP}
          >
            <div className="flex flex-col items-center text-center gap-4">
              <div className="p-4 rounded-full bg-primary/10 group-hover:bg-primary/20 transition-colors">
                <Users className="w-12 h-12 text-primary" />
              </div>
              <div>
                <h2 className="text-xl font-bold mb-1">Online PvP</h2>
                <p className="text-sm text-muted-foreground">
                  Play against real players online
                </p>
              </div>
              <Button className="w-full">Find Match</Button>
            </div>
          </Card>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.2 }}
        >
          <Card 
            className="p-6 cursor-pointer hover:ring-2 hover:ring-primary transition-all group"
            onClick={() => setShowBotDialog(true)}
          >
            <div className="flex flex-col items-center text-center gap-4">
              <div className="p-4 rounded-full bg-primary/10 group-hover:bg-primary/20 transition-colors">
                <Ghost className="w-12 h-12 text-primary" />
              </div>
              <div>
                <h2 className="text-xl font-bold mb-1">VS Bot</h2>
                <p className="text-sm text-muted-foreground">
                  Practice against AI opponents
                </p>
              </div>
              <Button variant="secondary" className="w-full">Choose Difficulty</Button>
            </div>
          </Card>
        </motion.div>
      </div>

      <BotDifficultyDialog
        open={showBotDialog}
        onOpenChange={setShowBotDialog}
        onSelectDifficulty={onSelectBot}
      />
    </div>
  );
};
