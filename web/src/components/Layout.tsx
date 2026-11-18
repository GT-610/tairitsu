import { Outlet } from 'react-router-dom';
import { useAuth } from '../services/auth';
import ResponsiveDrawer from './ResponsiveDrawer';
import ImageNameNotice from './ImageNameNotice';
import { User } from '../services/api';

// Layout组件的props类型
interface LayoutProps {
  user: User | null;
}

function Layout({ user }: LayoutProps) {
  const { logout } = useAuth();

  const handleLogout = () => {
    // 调用auth context中的logout函数
    void logout();
  };

  return (
    <ResponsiveDrawer title="Tempest-ZT" user={user} onLogout={handleLogout}>
      {/* The administrator is the one who pulls the image, so the notice only
          concerns that account. */}
      {user?.role === 'admin' && <ImageNameNotice />}
      <Outlet />
    </ResponsiveDrawer>
  );
}

export default Layout;
