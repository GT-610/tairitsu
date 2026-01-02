import * as React from 'react';
import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import CssBaseline from '@mui/material/CssBaseline';
import Divider from '@mui/material/Divider';
import Drawer from '@mui/material/Drawer';
import IconButton from '@mui/material/IconButton';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import MenuIcon from '@mui/icons-material/Menu';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import DashboardIcon from '@mui/icons-material/Dashboard';
import LanguageIcon from '@mui/icons-material/Language';
import SettingsIcon from '@mui/icons-material/Settings';
import AccountCircle from '@mui/icons-material/AccountCircle';
import ImportExportIcon from '@mui/icons-material/ImportExport';
import GroupIcon from '@mui/icons-material/Group';
import PublicIcon from '@mui/icons-material/Public';
import MenuItem from '@mui/material/MenuItem';
import Menu from '@mui/material/Menu';
import MenuList from '@mui/material/MenuList';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import Button from '@mui/material/Button';
import { Link, useLocation } from 'react-router-dom';
import { User } from '../services/api';
import LogoutIcon from '@mui/icons-material/Logout';
import { ImportNetworkModal } from './ImportNetworkModal';

const drawerWidth = 240;

// Menu item type definition
interface MenuItemType {
  text: string;
  path: string;
  icon: React.ReactNode;
  isAction?: boolean;
}

const mainMenuItems: MenuItemType[] = [
  {
    text: '网络',
    path: '/networks',
    icon: <LanguageIcon />
  },
  {
    text: '设置',
    path: '/settings',
    icon: <SettingsIcon />
  }
];

const adminMenuItems: MenuItemType[] = [
  {
    text: '管理员面板',
    path: '/dashboard',
    icon: <DashboardIcon />
  },
  {
    text: '用户管理',
    path: '/user-management',
    icon: <GroupIcon />
  },
  {
    text: '导入网络',
    path: '',
    icon: <ImportExportIcon />,
    isAction: true
  },
  {
    text: '生成 planet',
    path: '',
    icon: <PublicIcon />,
    isAction: true
  }
];

// ResponsiveDrawer component props type
interface ResponsiveDrawerProps {
  window?: () => Window;
  children: React.ReactNode;
  title?: string;
  user: User | null;
  onLogout: () => void;
}

export default function ResponsiveDrawer({ window, children, title = 'Tairitsu', user, onLogout }: ResponsiveDrawerProps) {
  const [mobileOpen, setMobileOpen] = React.useState<boolean>(false);
  const [isClosing, setIsClosing] = React.useState<boolean>(false);
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const [openConfirmDialog, setOpenConfirmDialog] = React.useState<boolean>(false);
  const [openImportModal, setOpenImportModal] = React.useState<boolean>(false);
  const location = useLocation();

  const handleDrawerClose = () => {
    setIsClosing(true);
    setMobileOpen(false);
  };

  const handleDrawerTransitionEnd = () => {
    setIsClosing(false);
  };

  const handleDrawerToggle = () => {
    if (!isClosing) {
      setMobileOpen(!mobileOpen);
    }
  };

  const handleMenu = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLogoutClick = () => {
    handleClose();
    setOpenConfirmDialog(true);
  };

  const handleConfirmLogout = () => {
    setOpenConfirmDialog(false);
    onLogout();
  };

  const handleCancelLogout = () => {
    setOpenConfirmDialog(false);
  };

  const handleOpenImportModal = () => {
    setOpenImportModal(true);
  };

  const handleCloseImportModal = () => {
    setOpenImportModal(false);
  };

  const handleImportComplete = () => {
    if (window) {
      window().location.reload();
    }
  };

  const drawer = (
    <div>
      <Toolbar />
      <Divider />
      {/* 主菜单 */}
      <List>
        {mainMenuItems.map((item) => (
          <ListItem key={item.text} disablePadding>
            <ListItemButton
              component={Link}
              to={item.path}
              selected={location.pathname.startsWith(item.path)}
              onClick={handleDrawerClose}
            >
              <ListItemIcon>
                {item.icon}
              </ListItemIcon>
              <ListItemText primary={item.text} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
      
      {/* 管理员菜单，仅管理员可见 */}
      {user && user.role === 'admin' && (
        <>
          <Divider />
          <List>
            {adminMenuItems.map((item) => (
              <ListItem key={item.text} disablePadding>
                {item.isAction ? (
                  <ListItemButton onClick={() => { handleDrawerClose(); item.text === '导入网络' && handleOpenImportModal(); item.text === '生成 planet' && alert('功能正在开发中'); }}>
                    <ListItemIcon>
                      {item.icon}
                    </ListItemIcon>
                    <ListItemText primary={item.text} />
                  </ListItemButton>
                ) : (
                  <ListItemButton
                    component={Link}
                    to={item.path}
                    selected={location.pathname.startsWith(item.path)}
                    onClick={handleDrawerClose}
                  >
                    <ListItemIcon>
                      {item.icon}
                    </ListItemIcon>
                    <ListItemText primary={item.text} />
                  </ListItemButton>
                )}
              </ListItem>
            ))}
          </List>
        </>
      )}
    </div>
  );

  const container = window !== undefined ? () => window().document.body : undefined;

  return (
    <Box sx={{ display: 'flex' }}>
      <CssBaseline />
      <AppBar
        position="fixed"
        sx={{
          width: { xs: '100%', sm: `calc(100% - ${drawerWidth}px)` },
          ml: { xs: 0, sm: `${drawerWidth}px` },
        }}
      >
        <Toolbar>
          <IconButton
            color="inherit"
            aria-label="open drawer"
            edge="start"
            onClick={handleDrawerToggle}
            sx={{ mr: 2, display: { sm: 'none' } }}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1 }}>
            {title}
          </Typography>
          {user && (
            <div>
              <Button
                aria-label="account of current user"
                aria-controls="menu-appbar"
                aria-haspopup="true"
                onClick={handleMenu}
                color="inherit"
                startIcon={<AccountCircle />}
              >
                {user.username}
              </Button>
              <Menu
                id="menu-appbar"
                anchorEl={anchorEl}
                keepMounted
                open={Boolean(anchorEl)}
                onClose={handleClose}
              >
                <MenuList>
                  <MenuItem onClick={handleLogoutClick}>
                    <ListItemIcon>
                      <LogoutIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>退出登录</ListItemText>
                  </MenuItem>
                </MenuList>
              </Menu>
            </div>
          )}
        </Toolbar>
      </AppBar>
      <Box
        component="nav"
        sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}
        aria-label="navigation menu"
      >
        {/* The implementation can be swapped with js to avoid SEO duplication of links. */}
        <Drawer
          container={container}
          variant="temporary"
          open={mobileOpen}
          onTransitionEnd={handleDrawerTransitionEnd}
          onClose={handleDrawerClose}
          sx={{
            display: { xs: 'block', sm: 'none' },
            '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
          }}
          slotProps={{
            root: {
              keepMounted: true, // Better open performance on mobile.
            },
          }}
        >
          {drawer}
        </Drawer>
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: 'none', sm: 'block' },
            '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
          }}
          open
        >
          {drawer}
        </Drawer>
      </Box>
      <Box
        component="main"
        sx={{ flexGrow: 1, p: 3, width: { xs: '100%', sm: `calc(100% - ${drawerWidth}px)` } }}
      >
        <Toolbar />
        {children}
      </Box>
      <Dialog
        open={openConfirmDialog}
        onClose={handleCancelLogout}
        aria-labelledby="confirm-logout-title"
        aria-describedby="confirm-logout-description"
      >
        <DialogTitle id="confirm-logout-title">
          退出登录
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="confirm-logout-description">
            确定要退出登录吗？
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCancelLogout} color="primary">
            取消
          </Button>
          <Button onClick={handleConfirmLogout} color="primary" autoFocus>
            确认
          </Button>
        </DialogActions>
      </Dialog>

      <ImportNetworkModal
        open={openImportModal}
        onClose={handleCloseImportModal}
        onImportComplete={handleImportComplete}
      />
    </Box>
  );
}