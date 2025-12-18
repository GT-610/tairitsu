import { Box, Typography, Card, CardContent, Avatar, Grid, Divider }
from '@mui/material';
import { User } from '../services/api';

// Profile组件的props类型
interface ProfileProps {
  user: User | null;
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
                  {user.role || '普通用户'}
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
                    用户名
                  </Typography>
                  <Typography variant="body1">
                    {user.username}
                  </Typography>
                </Box>
                
                <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 2fr' }, gap: 2 }}>
                  <Typography variant="body2" color="text.secondary">
                    创建时间
                  </Typography>
                  <Typography variant="body1">
                    {user.createdAt ? new Date(user.createdAt).toLocaleString() : '未知'}
                  </Typography>
                </Box>
                
                <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 2fr' }, gap: 2 }}>
                  <Typography variant="body2" color="text.secondary">
                    更新时间
                  </Typography>
                  <Typography variant="body1">
                    {user.updatedAt ? new Date(user.updatedAt).toLocaleString() : '未知'}
                  </Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}

export default Profile;