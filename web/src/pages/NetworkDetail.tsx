import { useState, useEffect } from 'react';
import { Box, Typography, Card, CardContent, CircularProgress, Alert, Button, Divider, Grid, IconButton, Tabs, Tab, Paper }
from '@mui/material';
import { ArrowBack, ContentCopy, MoreVert } from '@mui/icons-material';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { networkAPI, Network } from '../services/api';

function NetworkDetail() {
  const { id } = useParams<{ id: string }>();
  const [network, setNetwork] = useState<Network | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [activeTab, setActiveTab] = useState<number>(0);
  const navigate = useNavigate();

  useEffect(() => {
    fetchNetworkDetail();
  }, [id]);

  const fetchNetworkDetail = async () => {
    setLoading(true);
    try {
      if (!id) {
        throw new Error('网络ID不能为空');
      }
      const response = await networkAPI.getNetwork(id);
      setNetwork(response.data);
    } catch (err: any) {
      setError('获取网络详情失败');
      console.error('Fetch network detail error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setActiveTab(newValue);
  };

  if (error || !id) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error" sx={{ mb: 3 }}>
          {error || '网络不存在'}
        </Alert>
        <Button 
          variant="contained" 
          component={Link} 
          to="/networks"
        >
          返回网络列表
        </Button>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      {/* 标题栏始终显示 */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <IconButton 
            component={Link} 
            to="/networks"
            size="large"
          >
            <ArrowBack />
          </IconButton>
          <Typography variant="h4" component="h1">
            {network?.name}
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Typography variant="body1" color="text.secondary">
            网络ID: {network?.id}
          </Typography>
          <IconButton size="small">
            <ContentCopy />
          </IconButton>
        </Box>
      </Box>

      {/* 导航标签页始终显示 */}
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
        <Tabs value={activeTab} onChange={handleTabChange} aria-label="network tabs">
          <Tab label="成员设备" />
          <Tab label="设置" />
        </Tabs>
      </Box>

      {/* 加载状态显示 */}
      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        // 内容区域
        <>
          {/* 成员设备选项卡内容 */}
          {activeTab === 0 && (
            <>
              {/* 统计卡片区域 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Grid container spacing={3}>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                      <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                        <Typography variant="h6" color="text.secondary" gutterBottom>
                          设备总数
                        </Typography>
                        <Typography variant="h4">
                          {network?.members?.length || 0}
                        </Typography>
                      </CardContent>
                    </Card>
                  </Grid>
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                      <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                        <Typography variant="h6" color="text.secondary" gutterBottom>
                          已授权设备
                        </Typography>
                        <Typography variant="h4">
                          {(network?.members?.filter(member => member.authorized).length || 0)}
                        </Typography>
                      </CardContent>
                    </Card>
                  </Grid>
                </Grid>
              </Paper>

              {/* 成员设备部分 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                  <Typography variant="h5">
                    成员设备
                  </Typography>
                </Box>
                <Typography variant="body1" color="text.secondary" sx={{ textAlign: 'center', py: 5 }}>
                  暂无设备连接
                </Typography>
              </Paper>
            </>
          )}

          {/* 设置选项卡内容 */}
          {activeTab === 1 && (
            <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
              <Typography variant="h5" sx={{ mb: 3 }}>
                网络设置
              </Typography>
              <Alert severity="info" sx={{ mb: 3 }}>
                网络设置页面开发中
              </Alert>
            </Paper>
          )}
        </>
      )}
    </Box>
  );
}

export default NetworkDetail;