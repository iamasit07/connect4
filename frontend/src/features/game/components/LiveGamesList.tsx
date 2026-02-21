import { motion } from 'framer-motion';
import { Eye, Users, Gamepad2, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useLiveGames } from '@/hooks/queries/useGameQueries';

interface LiveGamesListProps {
  onSpectate: (gameId: string) => void;
}

export const LiveGamesList = ({ onSpectate }: LiveGamesListProps) => {
  const { data: games = [], isLoading: loading, error, refetch } = useLiveGames();

  const getTimeSinceStart = (startedAt: string) => {
    const diff = Date.now() - new Date(startedAt).getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return 'Just started';
    return `${mins}m ago`;
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="flex items-center gap-2 text-lg">
          <Gamepad2 className="w-5 h-5 text-primary" />
          Live Games
        </CardTitle>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => refetch()}
          disabled={loading}
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
        </Button>
      </CardHeader>
      <CardContent className="space-y-3">
        {loading && games.length === 0 ? (
          <>
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </>
        ) : games.length === 0 ? (
          <div className="text-center py-6 text-muted-foreground">
            <Gamepad2 className="w-10 h-10 mx-auto mb-2 opacity-50" />
            <p>No live games right now</p>
            <p className="text-xs mt-1">Check back soon!</p>
          </div>
        ) : (
          games.map((game, index) => (
            <motion.div
              key={game.gameId}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="group flex items-center justify-between p-3 rounded-lg bg-muted/50 hover:bg-muted transition-colors cursor-pointer"
              onClick={() => onSpectate(game.gameId)}
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 text-sm font-medium">
                  <span className="truncate">{game.player1.username}</span>
                  <span className="text-muted-foreground">vs</span>
                  <span className="truncate">{game.player2.username}</span>
                </div>
                <div className="flex items-center gap-3 text-xs text-muted-foreground mt-1">
                  <span>{game.moveCount} moves</span>
                  <span>•</span>
                  <span>{getTimeSinceStart(game.startedAt)}</span>
                  {game.spectatorCount > 0 && (
                    <>
                      <span>•</span>
                      <span className="flex items-center gap-1">
                        <Users className="w-3 h-3" />
                        {game.spectatorCount}
                      </span>
                    </>
                  )}
                </div>
              </div>
              <Button
                size="sm"
                variant="secondary"
                className="gap-1 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none"
              >
                <Eye className="w-3.5 h-3.5" />
                Watch
              </Button>
            </motion.div>
          ))
        )}
      </CardContent>
    </Card>
  );
};
