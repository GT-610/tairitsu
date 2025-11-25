import React from 'react'
import { Outlet, Link, useNavigate } from 'react-router-dom'
import { AppBar, Toolbar, Drawer, List, ListItem, ListItemText, Typography, Button, Box, Avatar }
from '@mui/material'
import { useAuth } from '../services/auth.jsx'

function Layout({ user }) {
  const [drawerOpen, setDrawerOpen] = React.useState(false)
  const navigate = useNavigate()
  const { logout } = useAuth() || {}

  const handleLogout = () => {
    // 调用auth context中的logout函数
    if (typeof logout === 'function') {
      logout()
    }
    // 导航到登录页面
    navigate('/login')
  }

  const menuItems = [
    { text: '仪表盘', path: '/dashboard' },
    { text: '网络管理', path: '/networks' },
    { text: '个人设置', path: '/profile' }
  ]

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <AppBar position="fixed" sx={{ zIndex: 1200 }}>
        <Toolbar>
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
            <Button color="inherit" onClick={handleLogout}>
              退出
            </Button>
          </Box>
        </Toolbar>
      </AppBar>

      <Drawer
        variant="permanent"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        sx={{
          width: 240,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: 240,
            boxSizing: 'border-box',
            marginTop: 64
          }
        }}
      >
        <List>
          {menuItems.map((item, index) => (
            <ListItem 
              button 
              key={index}
              component={Link} 
              to={item.path}
              onClick={() => setDrawerOpen(false)}
            >
              <ListItemText primary={item.text} />
            </ListItem>
          ))}
        </List>
      </Drawer>

      <Box component="main" sx={{ flexGrow: 1, p: 3, marginTop: 8 }}>
        <Outlet />
      </Box>
    </Box>
  )
}

export default Layout