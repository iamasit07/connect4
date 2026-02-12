import { useState } from 'react';
import { motion } from 'framer-motion';
import { User, Trophy, Target, TrendingUp, Edit2, Save, X, Loader2 } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { toast } from 'sonner';
import api from '@/lib/axios';
import { useAuthStore } from '@/features/auth/store/authStore';

const Profile = () => {
  const { user, setUser } = useAuthStore();
  const [isEditing, setIsEditing] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState({
    username: user?.username || '',
    email: user?.email || '',
  });

  const stats = [
    {
      label: 'Rating',
      value: user?.rating || 1000,
      icon: TrendingUp,
      color: 'text-primary',
    },
    {
      label: 'Wins',
      value: user?.wins || 0,
      icon: Trophy,
      color: 'text-green-500',
    },
    {
      label: 'Losses',
      value: user?.losses || 0,
      icon: Target,
      color: 'text-red-500',
    },
    {
      label: 'Draws',
      value: user?.draws || 0,
      icon: Target,
      color: 'text-muted-foreground',
    },
  ];

  const totalGames = (user?.wins || 0) + (user?.losses || 0) + (user?.draws || 0);
  const winRate = totalGames > 0 
    ? Math.round(((user?.wins || 0) / totalGames) * 100) 
    : 0;

  const handleSave = async () => {
    setIsLoading(true);
    try {
      const response = await api.put('/auth/profile', formData);
      const token = localStorage.getItem('jwt_token');
      if (token) {
        setUser(response.data, token);
      }
      toast.success('Profile updated successfully');
      setIsEditing(false);
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to update profile');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCancel = () => {
    setFormData({
      username: user?.username || '',
      email: user?.email || '',
    });
    setIsEditing(false);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="container max-w-4xl py-8"
    >
      <div className="flex items-center gap-3 mb-8">
        <div className="p-2 rounded-lg bg-primary/10">
          <User className="h-6 w-6 text-primary" />
        </div>
        <div>
          <h1 className="text-2xl font-bold">Profile</h1>
          <p className="text-muted-foreground">Manage your account</p>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        {/* Profile Card */}
        <Card className="md:col-span-2">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Account Details</CardTitle>
            {!isEditing ? (
              <Button variant="ghost" size="sm" onClick={() => setIsEditing(true)}>
                <Edit2 className="h-4 w-4 mr-2" />
                Edit
              </Button>
            ) : (
              <div className="flex gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleCancel}
                  disabled={isLoading}
                >
                  <X className="h-4 w-4 mr-2" />
                  Cancel
                </Button>
                <Button size="sm" onClick={handleSave} disabled={isLoading}>
                  {isLoading ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <Save className="h-4 w-4 mr-2" />
                  )}
                  Save
                </Button>
              </div>
            )}
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-6 mb-6">
              <Avatar className="h-20 w-20">
                <AvatarFallback className="text-2xl bg-primary/10 text-primary">
                  {user?.username?.slice(0, 2).toUpperCase() || 'U'}
                </AvatarFallback>
              </Avatar>
              <div>
                <h2 className="text-xl font-semibold">{user?.username}</h2>
                <p className="text-muted-foreground">{user?.email}</p>
              </div>
            </div>

            <Separator className="my-6" />

            {isEditing ? (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="username">Username</Label>
                  <Input
                    id="username"
                    value={formData.username}
                    onChange={(e) =>
                      setFormData({ ...formData, username: e.target.value })
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    value={formData.email}
                    onChange={(e) =>
                      setFormData({ ...formData, email: e.target.value })
                    }
                  />
                </div>
              </div>
            ) : (
              <div className="space-y-4">
                <div>
                  <Label className="text-muted-foreground">Username</Label>
                  <p className="font-medium">{user?.username}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Email</Label>
                  <p className="font-medium">{user?.email}</p>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Stats Card */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Statistics</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {stats.map((stat) => (
                <div
                  key={stat.label}
                  className="flex items-center justify-between"
                >
                  <div className="flex items-center gap-2">
                    <stat.icon className={`h-4 w-4 ${stat.color}`} />
                    <span className="text-muted-foreground">{stat.label}</span>
                  </div>
                  <span className="font-semibold">{stat.value}</span>
                </div>
              ))}

              <Separator />

              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Total Games</span>
                <span className="font-semibold">{totalGames}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Win Rate</span>
                <span className={`font-semibold ${
                  winRate >= 60 
                    ? 'text-green-500' 
                    : winRate >= 40 
                      ? 'text-foreground' 
                      : 'text-red-500'
                }`}>
                  {winRate}%
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </motion.div>
  );
};

export default Profile;
