import { useState, useEffect } from 'react';
import { Box, Typography, Card, CardContent, CircularProgress, Alert, Button, Grid, IconButton, Tabs, Tab, Paper, TextField, Switch, FormControlLabel, FormControl, RadioGroup, Radio }
from '@mui/material';
import { ArrowBack, ContentCopy, Add } from '@mui/icons-material';
import { useParams, Link } from 'react-router-dom';
import { networkAPI, Network } from '../services/api';

function NetworkDetail() {
  const { id } = useParams<{ id: string }>();
  const [network, setNetwork] = useState<Network | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [activeTab, setActiveTab] = useState<number>(0);

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
            <>
              {/* 网络基本信息 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Typography variant="h5" sx={{ mb: 2 }}>
                  网络基本信息
                </Typography>
                <Grid container spacing={3}>
                  <Grid size={{ xs: 12 }}>
                    <TextField
                      fullWidth
                      label="网络名称"
                      variant="outlined"
                      value={network?.name || ''}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12 }}>
                    <TextField
                      fullWidth
                      label="网络描述"
                      variant="outlined"
                      multiline
                      rows={3}
                      value={network?.description || ''}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12 }}>
                    <Box sx={{ display: 'flex', gap: 2 }}>
                      <Button variant="contained" color="primary">
                        保存
                      </Button>
                      <Button variant="outlined">
                        重置
                      </Button>
                    </Box>
                  </Grid>
                </Grid>
              </Paper>
              
              {/* IPv4分配 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Typography variant="h5" sx={{ mb: 2 }}>
                  IPv4分配
                </Typography>
                <Box sx={{ mb: 3, p: 2, bgcolor: 'background.paper', borderRadius: 1, border: '1px solid', borderColor: 'divider' }}>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    自动分配的IPv4地址段
                  </Typography>
                  <Typography variant="body1">
                    192.168.192.0/24
                  </Typography>
                </Box>
                
                <Box sx={{ mb: 3 }}>
                  <Typography variant="subtitle1" sx={{ mb: 2 }}>
                    IPv4地址分配池
                  </Typography>
                  <Box sx={{ mb: 2 }}>
                    <Grid container spacing={2} alignItems="center">
                      <Grid size={{ xs: 5 }}>
                        <TextField
                          fullWidth
                          label="起始IP"
                          variant="outlined"
                          size="small"
                          value="192.168.192.2"
                        />
                      </Grid>
                      <Grid size={{ xs: 5 }}>
                        <TextField
                          fullWidth
                          label="结束IP"
                          variant="outlined"
                          size="small"
                          value="192.168.192.254"
                        />
                      </Grid>
                      <Grid size={{ xs: 2 }}>
                        <Button variant="outlined" color="error" fullWidth size="small">
                          删除
                        </Button>
                      </Grid>
                    </Grid>
                  </Box>
                  <Box sx={{ display: 'flex', gap: 2 }}>
                    <Button variant="outlined" startIcon={<Add />} size="small">
                      添加地址池
                    </Button>
                    <Button variant="outlined" size="small">
                      随机分配一个地址池
                    </Button>
                  </Box>
                </Box>
                
                <Box sx={{ mb: 3 }}>
                  <Typography variant="subtitle1" sx={{ mb: 2 }}>
                    IPv4自动分配模式
                  </Typography>
                  <FormControlLabel
                    control={<Switch checked={network?.config.v4AssignMode.zt || false} />}
                    label="自动分配IPv4地址给新成员设备"
                  />
                  <Box sx={{ pl: 4, mt: -1 }}>
                    <Typography variant="body2" color="text.secondary">
                      每个成员设备将自动获取一个IPv4地址。您可以在上方定义地址池范围。
                    </Typography>
                  </Box>
                </Box>
                
                <Box sx={{ mb: 3 }}>
                  <Typography variant="subtitle1" sx={{ mb: 2 }}>
                    活跃路由
                  </Typography>
                  <Box sx={{ mb: 2 }}>
                    <Grid container spacing={2} alignItems="center">
                      <Grid size={{ xs: 5 }}>
                        <TextField
                          fullWidth
                          label="目标网络"
                          variant="outlined"
                          size="small"
                          value="192.168.192.0/24"
                        />
                      </Grid>
                      <Grid size={{ xs: 5 }}>
                        <TextField
                          fullWidth
                          label="下一跳地址"
                          variant="outlined"
                          size="small"
                          placeholder="可选"
                        />
                      </Grid>
                      <Grid size={{ xs: 2 }}>
                        <Button variant="outlined" color="error" fullWidth size="small">
                          删除
                        </Button>
                      </Grid>
                    </Grid>
                  </Box>
                  <Button variant="outlined" startIcon={<Add />} size="small">
                    添加路由
                  </Button>
                </Box>
              </Paper>
              
              {/* IPv6分配 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Typography variant="h5" sx={{ mb: 2 }}>
                  IPv6分配
                </Typography>
                
                <Box sx={{ mb: 3 }}>
                  <Typography variant="subtitle1" sx={{ mb: 2 }}>
                    IPv6地址分配
                  </Typography>
                  <FormControl component="fieldset" sx={{ mb: 2 }}>
                    <RadioGroup
                      row
                      value={network?.config.v6AssignMode.rfc4193 ? 'rfc4193' : network?.config.v6AssignMode.zt ? 'zt' : 'custom'}
                    >
                      <FormControlLabel value="none" control={<Radio />} label="不分配IPv6地址" />
                      <FormControlLabel value="rfc4193" control={<Radio />} label="分配RFC4193唯一地址" />
                      <FormControlLabel value="custom" control={<Radio />} label="从自定义IPv6范围分配" />
                    </RadioGroup>
                  </FormControl>
                  
                  <Box sx={{ pl: 4 }}>
                    <Box sx={{ mb: 2 }}>
                      <Grid container spacing={2} alignItems="center">
                        <Grid size={{ xs: 5 }}>
                          <TextField
                            fullWidth
                            label="起始IPv6"
                            variant="outlined"
                            size="small"
                            placeholder="例如: fd00::1"
                          />
                        </Grid>
                        <Grid size={{ xs: 5 }}>
                          <TextField
                            fullWidth
                            label="结束IPv6"
                            variant="outlined"
                            size="small"
                            placeholder="例如: fd00::ffff"
                          />
                        </Grid>
                        <Grid size={{ xs: 2 }}>
                          <Button variant="outlined" color="error" fullWidth size="small">
                            删除
                          </Button>
                        </Grid>
                      </Grid>
                    </Box>
                    <Button variant="outlined" startIcon={<Add />} size="small">
                      添加IPv6地址池
                    </Button>
                  </Box>
                </Box>
              </Paper>
              
              {/* 多播设置 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Typography variant="h5" sx={{ mb: 2 }}>
                  多播设置
                </Typography>
                
                <Grid container spacing={3}>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                      <Typography variant="body1">
                        多播接收者限制
                      </Typography>
                      <TextField
                        type="number"
                        size="small"
                        value={network?.config.multicastLimit || 32}
                        sx={{ width: 120 }}
                      />
                    </Box>
                  </Grid>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <FormControlLabel
                      control={<Switch checked={true} />}
                      label="启用广播"
                    />
                  </Grid>
                </Grid>
              </Paper>
              
              {/* 删除网络 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Typography variant="h5" sx={{ mb: 2, color: 'error.main' }}>
                  删除网络
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  此操作不可恢复。删除网络将断开所有连接的设备，并永久删除网络配置。
                </Typography>
                <Button variant="contained" color="error" size="large">
                  删除网络
                </Button>
              </Paper>
            </>
          )}
        </>
      )}
    </Box>
  );
}

export default NetworkDetail;