import { Link, useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import { Gamepad2, Trophy, History, User } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/features/auth/store/authStore";
import { useAuth } from "@/features/auth/hooks/useAuth";
import { LiveGamesList } from "@/features/game/components/LiveGamesList";
import { useGameSocket } from "@/features/game/hooks/useGameSocket";
import { useCallback, useEffect } from "react";

const Dashboard = () => {
  const { user } = useAuthStore();
  const { checkAuth } = useAuth();
  const navigate = useNavigate();

  const onGameStart = useCallback(
    (gameId: string) => {
      navigate(`/game/${gameId}`);
    },
    [navigate],
  );

  const { spectateGame, disconnect } = useGameSocket(onGameStart);
  
  useEffect(() => {
    return () => disconnect();
  }, [disconnect]);

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  const quickActions = [
    {
      title: "Game History",
      description: "View past matches",
      icon: History,
      href: "/history",
      color: "bg-secondary text-secondary-foreground",
    },
    {
      title: "Leaderboard",
      description: "See top players",
      icon: Trophy,
      href: "/leaderboard",
      color: "bg-disk-yellow/10 text-yellow-600",
    },
    {
      title: "Profile",
      description: "Manage your account",
      icon: User,
      href: "/profile",
      color: "bg-muted text-muted-foreground",
    },
  ];

  const handleSpectate = (gameId: string) => {
    spectateGame(gameId);
  };

  return (
    <div className="container py-8">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="mb-8"
      >
        <h1 className="text-3xl font-bold mb-2">
          Welcome back, {user?.username}! 👋
        </h1>
        <p className="text-muted-foreground">
          Ready for your next Connect 4 challenge?
        </p>
      </motion.div>

      {/* Stats Overview */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
        className="grid gap-4 md:grid-cols-4 mb-8"
      >
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Rating</CardDescription>
            <CardTitle className="text-3xl">{user?.rating || 1000}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Wins</CardDescription>
            <CardTitle className="text-3xl text-green-500">
              {user?.wins || 0}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Losses</CardDescription>
            <CardTitle className="text-3xl text-destructive">
              {user?.losses || 0}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Draws</CardDescription>
            <CardTitle className="text-3xl text-muted-foreground">
              {user?.draws || 0}
            </CardTitle>
          </CardHeader>
        </Card>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2 }}
        className="mb-8"
      >
        <h2 className="text-xl font-semibold mb-4">Play Now</h2>
        <div className="grid gap-4 md:grid-cols-1">
          <Link to="/play">
            <Card className="hover:ring-2 hover:ring-primary transition-all cursor-pointer group bg-gradient-to-r from-primary/10 to-secondary/10 border-2">
              <CardContent className="flex flex-col md:flex-row items-center gap-6 p-6 md:p-8 text-center md:text-left">
                <div className="p-4 rounded-full bg-primary/20 group-hover:bg-primary/30 transition-colors shrink-0">
                  <Gamepad2 className="h-10 w-10 text-primary" />
                </div>
                <div className="flex-1">
                  <h3 className="font-bold text-2xl mb-2">Start New Game</h3>
                  <p className="text-muted-foreground text-lg">
                    Play online against real players or practice with AI
                  </p>
                </div>
                <Button size="lg" className="w-full md:w-auto px-8">
                  Play Now
                </Button>
              </CardContent>
            </Card>
          </Link>
        </div>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.3 }}
      >
        <h2 className="text-xl font-semibold mb-4">Quick Actions</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {quickActions.map((action, index) => (
            <motion.div
              key={action.title}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.3 + index * 0.05 }}
            >
              <Link to={action.href}>
                <Card className="hover:ring-2 hover:ring-primary transition-all cursor-pointer h-full">
                  <CardHeader>
                    <div
                      className={`w-12 h-12 rounded-lg ${action.color} flex items-center justify-center mb-2`}
                    >
                      <action.icon className="h-6 w-6" />
                    </div>
                    <CardTitle className="text-lg">{action.title}</CardTitle>
                    <CardDescription>{action.description}</CardDescription>
                  </CardHeader>
                </Card>
              </Link>
            </motion.div>
          ))}
        </div>
      </motion.div>

      {/* Live Games Section */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4 }}
        className="mt-8"
      >
        <LiveGamesList onSpectate={handleSpectate} />
      </motion.div>
    </div>
  );
};

export default Dashboard;
