import React from 'react'
import { Outlet, Link, useLocation } from 'react-router-dom'
import { AppBar, Toolbar, Drawer, List, ListItem, ListItemText, Typography, Button, Box, Avatar, IconButton }
from '@mui/material'
import { Menu, X, Home, Wifi, AccountCircle, Settings, Logout } from '@mui/icons-material'
import '../TransitionStyles.css'

function Layout({ user, onLogout }) {
  const [drawerOpen, setDrawerOpen] = React.useState(true)
  const location = useLocation()
  const isMobile = React.useMemo(() => window.innerWidth < 900, [])
  
  // 基于屏幕宽度动态计算抽屉宽度
  const drawerWidth = React.useMemo(() => {
    const screenWidth = window.innerWidth;
    if (screenWidth < 600) return '80%';
    if (screenWidth < 900) return '300px';
    return '320px';
  }, []);
  
  const toggleDrawer = () => setDrawerOpen(!drawerOpen)
  const handleLogout = onLogout

  // 使用useMemo优化数组引用，避免不必要的重渲染
  const menuItems = React.useMemo(() => [
    { text: '仪表盘', path: '/', icon: <Home /> },
    { text: '网络', path: '/networks', icon: <Wifi /> },
    { text: '个人资料', path: '/profile', icon: <AccountCircle /> },
    { text: '设置', path: '/settings', icon: <Settings /> }
  ], [])

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <AppBar position="fixed" sx={{ zIndex: 1200 }}>
        <Toolbar>
          <IconButton 
            color="inherit" 
            aria-label="菜单" 
            edge="start" 
            onClick={toggleDrawer}
            sx={{ mr: 2, display: { xs: 'flex', md: 'none' } }}
            className="animated-button"
          >
            {drawerOpen ? <X /> : <Menu />}
          </IconButton>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            Tairitsu
          </Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Typography variant="body1" sx={{ mr: 2 }}>
              欢迎，{user?.username}
            </Typography>
            <Avatar sx={{ width: 32, height: 32 }}>
              {user?.username?.[0]?.toUpperCase() || 'U'}
            </Avatar>
            {/* 退出按钮已移至侧边栏 */}
          </Box>
        </Toolbar>
      </AppBar>

      <Drawer
        variant="persistent"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        anchor="left"
        sx={{
          '& .MuiDrawer-paper': {
            width: drawerWidth,
            boxSizing: 'border-box',
            marginTop: 64,
            overflowX: 'hidden',
            transition: 'width 0.3s ease-in-out, transform 0.3s ease-in-out',
            boxShadow: drawerOpen ? '2px 0 8px rgba(0,0,0,0.1)' : 'none',
            transform: 'translateX(0)',
            zIndex: 1300
          },
        }}
      >
        <List>
          {menuItems.map((item, index) => (
            <ListItem 
              button 
              key={index}
              component={Link} 
              to={item.path}
              onClick={() => isMobile && setDrawerOpen(false)}
              className="hover-card"
              sx={{
                transition: 'all 0.2s ease-in-out',
                '&:hover': {
                  backgroundColor: 'rgba(0, 0, 0, 0.08)'
                },
                backgroundColor: location.pathname === item.path || 
                                (item.path === '/networks' && location.pathname.startsWith('/networks')) 
                                  ? 'rgba(0, 0, 0, 0.04)' : 'transparent'
              }}
            >
              <Box sx={{ minWidth: '40px', mr: 2, color: location.pathname === item.path || 
                                (item.path === '/networks' && location.pathname.startsWith('/networks')) 
                                  ? 'primary.main' : 'inherit' }}>
                {item.icon}
              </Box>
              <ListItemText primary={item.text} />
            </ListItem>
          ))}
          
          <ListItem 
            button 
            onClick={handleLogout}
            className="hover-card"
            sx={{
              color: 'error.main',
              transition: 'all 0.2s ease-in-out',
              '&:hover': {
                backgroundColor: 'rgba(255, 0, 0, 0.08)'
              },
              mt: 'auto'
            }}
          >
            <Box sx={{ minWidth: '40px', mr: 2, color: 'error.main' }}>
              <Logout />
            </Box>
            <ListItemText primary="退出登录" />
          </ListItem>
        </List>
      </Drawer>

      <Box component="main" sx={{ flexGrow: 1, p: 3, marginTop: 8 }}>
        <Outlet />
      </Box>
    </Box>
  )
}

export default Layout