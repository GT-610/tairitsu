import { Outlet } from 'react-router-dom';
import { Box } from '@mui/material';
import { useAuth } from '../services/auth';
import ResponsiveDrawer from './ResponsiveDrawer';
import { User } from '../services/api';

// Layout component props type
interface LayoutProps {
  user: User | null;
}

function Layout({ user }: LayoutProps) {
  const { logout } = useAuth();

  const handleLogout = () => {
    // Call logout function from auth context
    logout();
  };

  return (
    <ResponsiveDrawer title="Tairitsu" user={user} onLogout={handleLogout}>
      <Outlet />
    </ResponsiveDrawer>
  );
}

export default Layout;