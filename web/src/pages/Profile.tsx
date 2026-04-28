import { Box, Typography, Card, CardContent, Avatar, Grid, Divider, Button }
from '@mui/material';
import { Link } from 'react-router-dom';
import { User } from '../services/api';
import { formatUserTime, getUserRoleLabel, hasDisplayableUserTime } from '../utils/userPresentation';
import { useTranslation } from '../i18n';

// Profile组件的props类型
interface ProfileProps {
  user: User | null;
}

function Profile({ user }: ProfileProps) {
  const { translateText } = useTranslation()

  if (!user) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body1" color="error">
          {translateText('用户信息不可用')}
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        {translateText('个人信息')}
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
                  {translateText(getUserRoleLabel(user.role))}
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid size={{ xs: 12, md: 8 }}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                {translateText('账户信息')}
              </Typography>
              <Divider sx={{ mb: 3 }} />
              
              <Box sx={{ display: 'grid', gap: 3 }}>
                {hasDisplayableUserTime(user.createdAt) && (
                  <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 2fr' }, gap: 2 }}>
                    <Typography variant="body2" color="text.secondary">
                      {translateText('创建时间')}
                    </Typography>
                    <Typography variant="body1">
                      {formatUserTime(user.createdAt)}
                    </Typography>
                  </Box>
                )}
                
                {hasDisplayableUserTime(user.updatedAt) && (
                  <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 2fr' }, gap: 2 }}>
                    <Typography variant="body2" color="text.secondary">
                      {translateText('更新时间')}
                    </Typography>
                    <Typography variant="body1">
                      {formatUserTime(user.updatedAt)}
                    </Typography>
                  </Box>
                )}
              </Box>

              <Box sx={{ mt: 4 }}>
                <Button component={Link} to="/settings" variant="outlined">
                  {translateText('打开账户设置与修改密码')}
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
