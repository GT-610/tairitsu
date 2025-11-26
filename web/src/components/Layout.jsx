
import { Outlet } from 'react-router-dom'
import { Box }
from '@mui/material'
import { useAuth } from '../services/auth.jsx'
import ResponsiveDrawer from './ResponsiveDrawer.jsx'

function Layout({ user }) {
  const { logout } = useAuth() || {}

  const handleLogout = () => {
    // 调用auth context中的logout函数
    if (typeof logout === 'function') {
      logout()
    }
  }

  return (
    <ResponsiveDrawer title="Tairitsu" user={user} onLogout={handleLogout}>
      <Outlet />
    </ResponsiveDrawer>
  )
}

export default Layout