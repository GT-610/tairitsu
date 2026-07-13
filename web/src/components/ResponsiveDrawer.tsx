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
import VolunteerActivismIcon from '@mui/icons-material/VolunteerActivism';
import BrightnessAutoIcon from '@mui/icons-material/BrightnessAuto';
import DarkModeIcon from '@mui/icons-material/DarkMode';
import LightModeIcon from '@mui/icons-material/LightMode';
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
import { Link, useLocation} from 'react-router-dom';
import { User } from '../services/api';
import LogoutIcon from '@mui/icons-material/Logout';
import { useTranslation } from '../i18n';

const drawerWidth = 240;

// Menu item type definition
interface MenuItemType {
  text: string;
  path: string;
  icon: React.ReactNode;
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
    path: '/import-network',
    icon: <ImportExportIcon />
  },
  {
    text: '生成 Planet（实验性）',
    path: '/planet',
    icon: <PublicIcon />
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
  const { cycleThemePreference, t, themePreference, translateText } = useTranslation();
  const [mobileOpen, setMobileOpen] = React.useState<boolean>(false);
  const [isClosing, setIsClosing] = React.useState<boolean>(false);
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const [openConfirmDialog, setOpenConfirmDialog] = React.useState<boolean>(false);
  const [openSponsorDialog, setOpenSponsorDialog] = React.useState<boolean>(false);
  const location = useLocation();
  const ThemeIcon = themePreference === 'system'
    ? BrightnessAutoIcon
    : themePreference === 'light'
      ? LightModeIcon
      : DarkModeIcon;
  const currentThemeLabel = themePreference === 'system'
    ? t('theme.system')
    : themePreference === 'light'
      ? t('theme.light')
      : t('theme.dark');
  const themeToggleLabel = themePreference === 'system'
    ? t('theme.toggleToLight')
    : themePreference === 'light'
      ? t('theme.toggleToDark')
      : t('theme.toggleToSystem');
  const themeButtonLabel = `${currentThemeLabel}. ${themeToggleLabel}`;

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
              <ListItemText primary={translateText(item.text)} />
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
                <ListItemButton
                  component={Link}
                  to={item.path}
                  selected={location.pathname.startsWith(item.path)}
                  onClick={handleDrawerClose}
                >
                  <ListItemIcon>
                    {item.icon}
                  </ListItemIcon>
                  <ListItemText primary={translateText(item.text)} />
                </ListItemButton>
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
            aria-label={t('navigation.openDrawer')}
            edge="start"
            onClick={handleDrawerToggle}
            sx={{ mr: 2, display: { sm: 'none' } }}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1 }}>
            {title}
          </Typography>
          <IconButton
            color="inherit"
            aria-label={t('sponsor.button')}
            title={t('sponsor.button')}
            onClick={() => setOpenSponsorDialog(true)}
            size="large"
          >
            <VolunteerActivismIcon />
          </IconButton>
          {user && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
              <IconButton
                color="inherit"
                aria-label={themeButtonLabel}
                title={themeButtonLabel}
                onClick={cycleThemePreference}
                size="large"
              >
                <ThemeIcon />
              </IconButton>
              <Button
                aria-label={t('user.accountMenu')}
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
                    <ListItemText>{translateText('退出登录')}</ListItemText>
                  </MenuItem>
                </MenuList>
              </Menu>
            </Box>
          )}
        </Toolbar>
      </AppBar>
      <Box
        component="nav"
        sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}
        aria-label={t('navigation.menu')}
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
          {translateText('退出登录')}
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="confirm-logout-description">
            {translateText('确定要退出登录吗？')}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCancelLogout} color="primary">
            {translateText('取消')}
          </Button>
          <Button onClick={handleConfirmLogout} color="primary" autoFocus>
            {translateText('确认')}
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog
        open={openSponsorDialog}
        onClose={() => setOpenSponsorDialog(false)}
        aria-labelledby="sponsor-dialog-title"
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle id="sponsor-dialog-title">
          {translateText('支持 Tairitsu')}
        </DialogTitle>
        <DialogContent>
          <DialogContentText component="div" sx={{ whiteSpace: 'pre-line' }}>
            {translateText('Tairitsu 是我开发的第一个开源生产环境项目。项目的发展离不开社区和用户的支持。')}
            {'\n\n'}
            {translateText('如果你觉得 Tairitsu 对你有帮助，欢迎通过以下方式支持项目：')}
            {'\n\n'}
            {translateText('• 前往 GitHub 仓库点击 Star，让更多人发现这个项目')}
            {'\n'}
            {translateText('• 通过 GitHub 仓库上的 Sponsor 按钮赞助我，支持项目的持续开发')}
            {'\n\n'}
            {translateText('使用过程中遇到问题或有好的建议，也欢迎提交 Issue 或 Pull Request！')}
            {'\n\n'}
            <Box component="span">
              {translateText('GitHub 仓库：')}
              <Box
                component="a"
                href="https://github.com/GT-610/tairitsu"
                target="_blank"
                rel="noopener noreferrer"
                sx={{ color: 'primary.main', textDecoration: 'underline' }}
              >
                https://github.com/GT-610/tairitsu
              </Box>
            </Box>
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenSponsorDialog(false)} color="primary" autoFocus>
            {translateText('确定')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
