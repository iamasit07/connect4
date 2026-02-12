import { motion } from 'framer-motion';
import { Trophy, Medal, Crown, Loader2 } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useLeaderboard } from '@/hooks/queries/useGameQueries';
import { useAuthStore } from '@/features/auth/store/authStore';

const Leaderboard = () => {
  const { data: entries = [], isLoading, error } = useLeaderboard();
  const { user } = useAuthStore();

  const getRankIcon = (rank: number) => {
    switch (rank) {
      case 1:
        return <Crown className="h-5 w-5 text-yellow-500" />;
      case 2:
        return <Medal className="h-5 w-5 text-gray-400" />;
      case 3:
        return <Medal className="h-5 w-5 text-amber-600" />;
      default:
        return (
          <span className="w-5 text-center text-muted-foreground font-medium">
            {rank}
          </span>
        );
    }
  };

  const getRowStyle = (rank: number, username: string) => {
    const isCurrentUser = user?.username === username;
    
    if (rank === 1) return 'bg-yellow-500/5 border-yellow-500/20';
    if (rank === 2) return 'bg-gray-500/5 border-gray-500/20';
    if (rank === 3) return 'bg-amber-500/5 border-amber-500/20';
    if (isCurrentUser) return 'bg-primary/5 border-primary/20';
    return '';
  };

  if (isLoading) {
    return (
      <div className="container max-w-4xl py-8">
        <div className="flex items-center gap-3 mb-8">
          <Skeleton className="h-10 w-10 rounded-lg" />
          <div>
            <Skeleton className="h-6 w-32 mb-2" />
            <Skeleton className="h-4 w-40" />
          </div>
        </div>
        <Card>
          <div className="p-4 space-y-4">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="flex items-center gap-4">
                <Skeleton className="h-8 w-8" />
                <Skeleton className="h-8 w-8 rounded-full" />
                <Skeleton className="h-4 flex-1" />
                <Skeleton className="h-4 w-16" />
                <Skeleton className="h-4 w-16" />
              </div>
            ))}
          </div>
        </Card>
      </div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="container max-w-4xl py-8"
    >
      <div className="flex items-center gap-3 mb-8">
        <div className="p-2 rounded-lg bg-yellow-500/10">
          <Trophy className="h-6 w-6 text-yellow-500" />
        </div>
        <div>
          <h1 className="text-2xl font-bold">Leaderboard</h1>
          <p className="text-muted-foreground">Top Connect 4 players</p>
        </div>
      </div>

      {error ? (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            {error.message || 'Failed to load leaderboard'}
          </CardContent>
        </Card>
      ) : entries.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Trophy className="h-12 w-12 mx-auto text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">No players ranked yet</p>
            <p className="text-sm text-muted-foreground/70">
              Play some games to appear on the leaderboard
            </p>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-16">Rank</TableHead>
                <TableHead>Player</TableHead>
                <TableHead className="text-right">Rating</TableHead>
                <TableHead className="text-right">W/L</TableHead>
                <TableHead className="text-right">Win Rate</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {entries.map((entry, index) => {
                const totalGames = entry.wins + entry.losses;
                const winRate = totalGames > 0 
                  ? Math.round((entry.wins / totalGames) * 100) 
                  : 0;

                return (
                  <motion.tr
                    key={entry.username}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: index * 0.03 }}
                    className={`border-l-2 ${getRowStyle(entry.rank, entry.username)}`}
                  >
                    <TableCell>
                      <div className="flex items-center justify-center">
                        {getRankIcon(entry.rank)}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <Avatar className="h-8 w-8">
                          <AvatarFallback className="text-xs">
                            {entry.username.slice(0, 2).toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                        <span className={
                          user?.username === entry.username 
                            ? 'font-semibold text-primary' 
                            : 'font-medium'
                        }>
                          {entry.username}
                          {user?.username === entry.username && ' (You)'}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="text-right font-mono font-semibold">
                      {entry.rating}
                    </TableCell>
                    <TableCell className="text-right text-muted-foreground">
                      <span className="text-green-600">{entry.wins}</span>
                      {' / '}
                      <span className="text-red-600">{entry.losses}</span>
                    </TableCell>
                    <TableCell className="text-right">
                      <span className={
                        winRate >= 60 
                          ? 'text-green-600' 
                          : winRate >= 40 
                            ? 'text-muted-foreground' 
                            : 'text-red-600'
                      }>
                        {winRate}%
                      </span>
                    </TableCell>
                  </motion.tr>
                );
              })}
            </TableBody>
          </Table>
        </Card>
      )}
    </motion.div>
  );
};

export default Leaderboard;
