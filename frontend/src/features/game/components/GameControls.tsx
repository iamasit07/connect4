import { Flag } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";

interface GameControlsProps {
  onSurrender: () => void;
  isPlaying: boolean;
}

export const GameControls = ({ onSurrender, isPlaying }: GameControlsProps) => {
  return (
    <div className="flex justify-center gap-4 mt-8">
      {isPlaying && (
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button
              variant="destructive"
              size="sm"
              className="gap-2"
            >
              <Flag className="w-4 h-4" />
              Surrender
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Surrender the game?</AlertDialogTitle>
              <AlertDialogDescription>
                This action cannot be undone. You will lose the game and your opponent will be declared the winner.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={onSurrender}>
                Yes, Surrender
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
    </div>
  );
};
