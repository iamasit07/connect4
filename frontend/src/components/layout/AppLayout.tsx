import { Outlet } from 'react-router-dom';
import { Header } from './Header';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { ActiveGamePopup } from "@/features/game/components/ActiveGamePopup";

export const AppLayout = () => {
  const { logout } = useAuth();

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Header onLogout={logout} />
      <main className="flex-1 flex flex-col relative">
        <Outlet />
      </main>
    </div>
  );
};
