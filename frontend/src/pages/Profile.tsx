import { useRef, useState } from "react";
import { motion } from "framer-motion";
import {
  User,
  Trophy,
  Target,
  TrendingUp,
  Edit2,
  Save,
  X,
  Loader2,
  Camera,
  Trash2,
  LogOut,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";
import { useAuthStore } from "@/features/auth/store/authStore";
import {
  useUpdateProfile,
  useUploadAvatar,
  useRemoveAvatar,
} from "@/hooks/queries/useAuthQueries";
import { useAuth } from "@/features/auth/hooks/useAuth";

const Profile = () => {
  const { user, setUser } = useAuthStore();
  const { logout } = useAuth();
  const updateProfileMutation = useUpdateProfile();
  const uploadAvatarMutation = useUploadAvatar();
  const removeAvatarMutation = useRemoveAvatar();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({
    name: user?.name || "",
  });

  const stats = [
    {
      label: "Rating",
      value: user?.rating || 1000,
      icon: TrendingUp,
      color: "text-primary",
    },
    {
      label: "Wins",
      value: user?.wins || 0,
      icon: Trophy,
      color: "text-green-500",
    },
    {
      label: "Losses",
      value: user?.losses || 0,
      icon: Target,
      color: "text-red-500",
    },
    {
      label: "Draws",
      value: user?.draws || 0,
      icon: Target,
      color: "text-muted-foreground",
    },
  ];

  const totalGames =
    (user?.wins || 0) + (user?.losses || 0) + (user?.draws || 0);
  const winRate =
    totalGames > 0 ? Math.round(((user?.wins || 0) / totalGames) * 100) : 0;

  const handleSave = async () => {
    try {
      const response = await updateProfileMutation.mutateAsync({
        name: formData.name,
      });
      setUser(response.user);
      toast.success("Profile updated successfully");
      setIsEditing(false);
    } catch (error) {
      toast.error(error.response?.data?.message || "Failed to update profile");
    }
  };

  const handleCancel = () => {
    setFormData({
      name: user?.name || "",
    });
    setIsEditing(false);
  };

  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file size (2MB)
    if (file.size > 2 * 1024 * 1024) {
      toast.error("Image must be smaller than 2MB");
      return;
    }

    // Validate file type
    if (!["image/jpeg", "image/png", "image/webp"].includes(file.type)) {
      toast.error("Only JPEG, PNG, and WebP images are allowed");
      return;
    }

    try {
      const response = await uploadAvatarMutation.mutateAsync(file);
      if (user) {
        setUser({ ...user, avatar_url: response.avatar_url });
      }
      toast.success("Avatar updated successfully");
    } catch (error) {
      toast.error(error.response?.data?.message || "Failed to upload avatar");
    }

    // Reset file input
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const handleRemoveAvatar = async () => {
    try {
      await removeAvatarMutation.mutateAsync();
      if (user) {
        setUser({ ...user, avatar_url: "" });
      }
      toast.success("Avatar removed");
    } catch (error) {
      toast.error("Failed to remove avatar");
    }
  };

  const displayName = user?.name || user?.username || "";

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
              <div className="flex gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setIsEditing(true)}
                >
                  <Edit2 className="h-4 w-4 mr-2" />
                  Edit
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={logout}
                >
                  <LogOut className="h-4 w-4 mr-2" />
                  Logout
                </Button>
              </div>
            ) : (
              <div className="flex gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleCancel}
                  disabled={updateProfileMutation.isPending}
                >
                  <X className="h-4 w-4 mr-2" />
                  Cancel
                </Button>
                <Button
                  size="sm"
                  onClick={handleSave}
                  disabled={updateProfileMutation.isPending}
                >
                  {updateProfileMutation.isPending ? (
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
              <div className="relative group">
                <Avatar className="h-20 w-20">
                  {user?.avatar_url && (
                    <AvatarImage src={user.avatar_url} alt={displayName} />
                  )}
                  <AvatarFallback className="text-2xl bg-primary/10 text-primary">
                    {displayName.slice(0, 2).toUpperCase() || "U"}
                  </AvatarFallback>
                </Avatar>
                {isEditing && (
                  <div className="absolute inset-0 flex items-center justify-center gap-1">
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept="image/jpeg,image/png,image/webp"
                      className="hidden"
                      onChange={handleAvatarUpload}
                    />
                    <Button
                      size="icon"
                      variant="secondary"
                      className="h-8 w-8 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
                      onClick={() => fileInputRef.current?.click()}
                      disabled={uploadAvatarMutation.isPending}
                    >
                      {uploadAvatarMutation.isPending ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <Camera className="h-4 w-4" />
                      )}
                    </Button>
                    {user?.avatar_url && (
                      <Button
                        size="icon"
                        variant="destructive"
                        className="h-8 w-8 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
                        onClick={handleRemoveAvatar}
                        disabled={removeAvatarMutation.isPending}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                )}
              </div>
              <div>
                <h2 className="text-xl font-semibold">{displayName}</h2>
                <p className="text-sm text-muted-foreground">
                  @{user?.username}
                </p>
                <p className="text-sm text-muted-foreground">{user?.email}</p>
              </div>
            </div>

            <Separator className="my-6" />

            {isEditing ? (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Display Name</Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) =>
                      setFormData({ ...formData, name: e.target.value })
                    }
                    placeholder="Enter your display name"
                  />
                </div>
                <div className="space-y-2">
                  <Label className="text-muted-foreground">Username</Label>
                  <p className="font-medium text-sm">{user?.username}</p>
                </div>
                <div className="space-y-2">
                  <Label className="text-muted-foreground">Email</Label>
                  <p className="font-medium text-sm">{user?.email}</p>
                </div>
              </div>
            ) : (
              <div className="space-y-4">
                <div>
                  <Label className="text-muted-foreground">Display Name</Label>
                  <p className="font-medium">{user?.name || "—"}</p>
                </div>
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
                <span
                  className={`font-semibold ${
                    winRate >= 60
                      ? "text-green-500"
                      : winRate >= 40
                        ? "text-foreground"
                        : "text-red-500"
                  }`}
                >
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
