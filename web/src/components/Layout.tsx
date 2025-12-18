import { Outlet } from 'react-router-dom';
import { useAuth } from '../services/auth';
import ResponsiveDrawer from './ResponsiveDrawer';
import { User } from '../services/api';

// Layout组件的props类型
interface LayoutProps {
  user: User | null;
}

function Layout({ user }: LayoutProps) {
  const { logout } = useAuth();

  const handleLogout = () => {
    // 调用auth context中的logout函数
    logout();
  };

  return (
    <ResponsiveDrawer title="Tairitsu" user={user} onLogout={handleLogout}>
      <Outlet />
    </ResponsiveDrawer>
  );
}

export default Layout;