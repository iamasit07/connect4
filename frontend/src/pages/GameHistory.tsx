import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { format, isToday, isYesterday, parseISO } from 'date-fns';
import { History, Trophy, X, Minus, Loader2, RefreshCw, Check } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useGameHistory } from '@/hooks/queries/useGameQueries';
import type { GameHistoryItem } from '@/features/game/types';

const GameHistory = () => {
  const { data, isLoading, error, refetch } = useGameHistory();
  const games = data ?? [];

  const [isRefreshing, setIsRefreshing] = useState(false);
  const [showRefreshed, setShowRefreshed] = useState(false);

  const handleRefresh = async () => {
    if (isRefreshing || showRefreshed) return;
    setIsRefreshing(true);
    await refetch();
    setIsRefreshing(false);
    setShowRefreshed(true);
    setTimeout(() => {
      setShowRefreshed(false);
    }, 5000);
  };

  const groupGamesByDate = (games: GameHistoryItem[]) => {
    const groups: Record<string, GameHistoryItem[]> = {};
    
    games.forEach((game) => {
      let key = 'Unknown Date';
      
      if (game.createdAt) {
        try {
          const date = parseISO(game.createdAt);
          if (isToday(date)) {
            key = 'Today';
          } else if (isYesterday(date)) {
            key = 'Yesterday';
          } else {
            key = format(date, 'MMMM d, yyyy');
          }
        } catch (e) {
          console.warn('Invalid date:', game.createdAt);
        }
      }
    
      if (!groups[key]) {
        groups[key] = [];
      }
      groups[key].push(game);
    });
    
    return groups;
  };

  const getResultIcon = (result: string) => {
    switch (result) {
      case 'win':
        return <Trophy className="h-4 w-4 text-yellow-500" />;
      case 'loss':
        return <X className="h-4 w-4 text-destructive" />;
      default:
        return <Minus className="h-4 w-4 text-muted-foreground" />;
    }
  };

  const getResultBadge = (result: string) => {
    const variants: Record<string, string> = {
      win: 'bg-green-500/10 text-green-600 border-green-500/20',
      loss: 'bg-red-500/10 text-red-600 border-red-500/20',
      draw: 'bg-muted text-muted-foreground border-muted',
    };
    
    return (
      <Badge variant="outline" className={variants[result]}>
        {result.charAt(0).toUpperCase() + result.slice(1)}
      </Badge>
    );
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-[60vh]">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  const groupedGames = groupGamesByDate(games);

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="container max-w-4xl py-8"
    >
      <div className="flex items-center gap-3 mb-8">
        <div className="p-2 rounded-lg bg-primary/10">
          <History className="h-6 w-6 text-primary" />
        </div>
        <div>
          <h1 className="text-2xl font-bold">Game History</h1>
          <p className="text-muted-foreground">Review your past matches</p>
        </div>
        <Button
          variant="ghost"
          size={showRefreshed ? "default" : "icon"}
          className={`ml-auto transition-all ${showRefreshed ? 'bg-green-500/10 text-green-600 hover:bg-green-500/20 hover:text-green-700' : ''}`}
          onClick={handleRefresh}
          disabled={isRefreshing || showRefreshed}
          title="Refresh history"
        >
          {isRefreshing ? (
            <Loader2 className="h-5 w-5 animate-spin" />
          ) : showRefreshed ? (
            <span className="flex items-center gap-2 font-medium">
              <Check className="h-4 w-4" />
              Refreshed
            </span>
          ) : (
            <RefreshCw className="h-5 w-5" />
          )}
        </Button>
      </div>

      {error ? (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            {error.message}
          </CardContent>
        </Card>
      ) : games.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <History className="h-12 w-12 mx-auto text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">No games played yet</p>
            <p className="text-sm text-muted-foreground/70">
              Start a match to see your history here
            </p>
          </CardContent>
        </Card>
      ) : (
        <ScrollArea className="h-[calc(100vh-250px)]">
          <div className="space-y-6 pr-4">
            {Object.entries(groupedGames).map(([date, dateGames]) => (
              <motion.div
                key={date}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
              >
                <h2 className="text-sm font-medium text-muted-foreground mb-3">
                  {date}
                </h2>
                <div className="space-y-2">
                  {dateGames.map((game, index) => (
                    <motion.div
                      key={game.id}
                      initial={{ opacity: 0, x: -10 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: index * 0.05 }}
                    >
                      <Card className="hover:bg-muted/50 transition-colors">
                        <CardContent className="py-4">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-3">
                              {getResultIcon(game.result)}
                              <div>
                                <p className="font-medium">
                                  vs {game.opponentUsername}
                                </p>
                                <p className="text-sm text-muted-foreground">
                                  {game.movesCount} moves • {game.endReason}
                                </p>
                              </div>
                            </div>
                            <div className="flex items-center gap-3">
                              <span className="text-sm text-muted-foreground">
                                {format(parseISO(game.createdAt), 'h:mm a')}
                              </span>
                              {getResultBadge(game.result)}
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    </motion.div>
                  ))}
                </div>
              </motion.div>
            ))}
          </div>
        </ScrollArea>
      )}
    </motion.div>
  );
};

export default GameHistory;
