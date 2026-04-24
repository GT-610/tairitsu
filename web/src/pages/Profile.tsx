import { Box, Typography, Card, CardContent, Avatar, Grid, Divider, Button }
from '@mui/material';
import { Link } from 'react-router-dom';
import { User } from '../services/api';

// Profile组件的props类型
interface ProfileProps {
  user: User | null;
}

function formatProfileTime(value?: string): string {
  if (!value) {
    return '未记录'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime()) || date.getFullYear() <= 1) {
    return '未记录'
  }

  return date.toLocaleString()
}

function getRoleLabel(role?: string): string {
  if (role === 'admin') {
    return '管理员'
  }
  if (role === 'user') {
    return '普通用户'
  }
  return '未知角色'
}

function hasValidProfileTime(value?: string): boolean {
  if (!value) {
    return false
  }

  const date = new Date(value)
  return !Number.isNaN(date.getTime()) && date.getFullYear() > 1
}

function Profile({ user }: ProfileProps) {
  if (!user) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body1" color="error">
          用户信息不可用
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        个人信息
      </Typography>
      
      <Grid container spacing={3}>
        <Grid size={{ xs: 12, md: 4 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3 }}>
                <Avatar sx={{ width: 120, height: 120, fontSize: 48 }}>
                  {user.username?.[0]?.toUpperCase() || 'U'}
                </Avatar>
                <Typography variant="h5">
                  {user.username}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {getRoleLabel(user.role)}
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid size={{ xs: 12, md: 8 }}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                账户信息
              </Typography>
              <Divider sx={{ mb: 3 }} />
              
              <Box sx={{ display: 'grid', gap: 3 }}>
                <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 2fr' }, gap: 2 }}>
                  <Typography variant="body2" color="text.secondary">
                    创建时间
                  </Typography>
                  <Typography variant="body1">
                    {formatProfileTime(user.createdAt)}
                  </Typography>
                </Box>
                
                {hasValidProfileTime(user.updatedAt) && (
                  <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 2fr' }, gap: 2 }}>
                    <Typography variant="body2" color="text.secondary">
                      更新时间
                    </Typography>
                    <Typography variant="body1">
                      {formatProfileTime(user.updatedAt)}
                    </Typography>
                  </Box>
                )}
              </Box>

              <Box sx={{ mt: 4 }}>
                <Button component={Link} to="/settings" variant="outlined">
                  打开账户设置与修改密码
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}

export default Profile;
