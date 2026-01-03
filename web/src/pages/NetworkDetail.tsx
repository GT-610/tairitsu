import { useState, useEffect } from 'react';
import { Box, Typography, Card, CardContent, CircularProgress, Alert, Button, Grid, IconButton, Tabs, Tab, Paper, TextField, Switch, FormControlLabel, FormControl, RadioGroup, Radio, Dialog, DialogTitle, DialogContent, DialogActions, DialogContentText, Snackbar } from '@mui/material';
import { ArrowBack, ContentCopy, Add } from '@mui/icons-material';
import { useParams, Link } from 'react-router-dom';
import { networkAPI, Network, NetworkConfig, NetworkMetadataUpdateRequest, IpAssignmentPool, Route } from '../services/api';

// Default configuration for optional fields (null instead of default values)
const defaultNetworkConfig: Partial<NetworkConfig> = {
  private: true,
  allowPassiveBridging: true,
  enableBroadcast: true,
  mtu: 2800,
  multicastLimit: 32,
  v4AssignMode: {
    zt: false
  },
  v6AssignMode: {
    zt: false,
    '6plane': false,
    rfc4193: false
  }
};

function NetworkDetail() {
  const { id } = useParams<{ id: string }>();
  const [network, setNetwork] = useState<Network | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [saving, setSaving] = useState<boolean>(false);
  const [error, setError] = useState<string>('');
  const [activeTab, setActiveTab] = useState<number>(0);
  const [editingName, setEditingName] = useState<string>('');
  const [editingDescription, setEditingDescription] = useState<string>('');
  const [v4AssignMode, setV4AssignMode] = useState<boolean>(false);
  const [v6AssignMode, setV6AssignMode] = useState<'zt' | '6plane' | 'rfc4193' | 'none'>('none');
  const [multicastLimit, setMulticastLimit] = useState<number>(32);
  const [enableBroadcast, setEnableBroadcast] = useState<boolean>(true);
  const [ipPools, setIpPools] = useState<IpAssignmentPool[]>([]);
  const [routes, setRoutes] = useState<Route[]>([]);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({ open: false, message: '', severity: 'success' });
  const [basicInfoUnsaved, setBasicInfoUnsaved] = useState<boolean>(false);
  const [ipv4Unsaved, setIpv4Unsaved] = useState<boolean>(false);
  const [ipv6Unsaved, setIpv6Unsaved] = useState<boolean>(false);
  const [multicastUnsaved, setMulticastUnsaved] = useState<boolean>(false);

  useEffect(() => {
    fetchNetworkDetail();
  }, [id]);

  useEffect(() => {
    if (network) {
      setBasicInfoUnsaved(editingName !== (network.name || '') || editingDescription !== (network.db_description || network.description || ''));
    }
  }, [editingName, editingDescription, network]);

  useEffect(() => {
    if (network) {
      const poolsChanged = JSON.stringify(ipPools) !== JSON.stringify(network.config.ipAssignmentPools || []);
      const routesChanged = JSON.stringify(routes) !== JSON.stringify(network.config.routes || []);
      const v4Changed = v4AssignMode !== (network.config.v4AssignMode?.zt ?? false);
      setIpv4Unsaved(poolsChanged || routesChanged || v4Changed);
    }
  }, [ipPools, routes, v4AssignMode, network]);

  useEffect(() => {
    if (network) {
      setIpv6Unsaved(v6AssignMode !== getV6AssignModeFromConfig(network.config.v6AssignMode));
    }
  }, [v6AssignMode, network]);

  useEffect(() => {
    if (network) {
      const multicastChanged = multicastLimit !== (network.config.multicastLimit ?? 32);
      const broadcastChanged = enableBroadcast !== (network.config.enableBroadcast ?? true);
      setMulticastUnsaved(multicastChanged || broadcastChanged);
    }
  }, [multicastLimit, enableBroadcast, network]);

  const fetchNetworkDetail = async () => {
    setLoading(true);
    try {
      if (!id) {
        throw new Error('网络ID不能为空');
      }
      const response = await networkAPI.getNetwork(id);
      const networkData = response.data;
      if (networkData) {
        networkData.config = { ...defaultNetworkConfig, ...networkData.config };
        if (networkData.config.v4AssignMode) {
          networkData.config.v4AssignMode = { ...defaultNetworkConfig.v4AssignMode, ...networkData.config.v4AssignMode };
        }
        if (networkData.config.v6AssignMode) {
          networkData.config.v6AssignMode = { ...defaultNetworkConfig.v6AssignMode, ...networkData.config.v6AssignMode };
        }
        setNetwork(networkData);
        setEditingName(networkData.name || '');
        setEditingDescription(networkData.db_description || networkData.description || '');
        setV4AssignMode(networkData.config.v4AssignMode?.zt ?? false);
        setV6AssignMode(getV6AssignModeFromConfig(networkData.config.v6AssignMode));
        setMulticastLimit(networkData.config.multicastLimit ?? 32);
        setEnableBroadcast(networkData.config.enableBroadcast ?? true);
        const pools = networkData.config.ipAssignmentPools || [];
        setIpPools(pools);
        setRoutes(networkData.config.routes || []);
      }
    } catch (err: any) {
      setError('获取网络详情失败');
      console.error('Fetch network detail error:', err);
    } finally {
      setLoading(false);
    }
  };

  const getV6AssignModeFromConfig = (config: any): 'zt' | '6plane' | 'rfc4193' | 'none' => {
    if (config?.rfc4193) return 'rfc4193';
    if (config?.['6plane']) return '6plane';
    if (config?.zt) return 'zt';
    return 'none';
  };

  const showSnackbar = (message: string, severity: 'success' | 'error' | 'info' = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const handleSaveMetadata = async () => {
    if (!id || !network) return;
    setSaving(true);
    try {
      const metadataData: NetworkMetadataUpdateRequest = {
        name: editingName,
        description: editingDescription
      };
      await networkAPI.updateNetworkMetadata(id, metadataData);
      showSnackbar('名称和描述保存成功');
      setTimeout(() => fetchNetworkDetail(), 1000);
    } catch (err: any) {
      showSnackbar('保存失败: ' + (err.response?.data?.error || err.message), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteNetwork = async () => {
    if (!id) return;
    setSaving(true);
    try {
      await networkAPI.deleteNetwork(id);
      showSnackbar('网络删除成功', 'success');
      setTimeout(() => {
        window.location.href = '/networks';
      }, 1000);
    } catch (err: any) {
      showSnackbar('删除失败: ' + (err.response?.data?.error || err.message), 'error');
    } finally {
      setSaving(false);
      setDeleteDialogOpen(false);
    }
  };

  const handleSaveIPv4 = async () => {
    if (!id) return;
    setSaving(true);
    try {
      await networkAPI.updateNetwork(id, {
        ipAssignmentPools: ipPools,
        routes: routes,
        v4AssignMode: { zt: v4AssignMode }
      });
      showSnackbar('IPv4配置保存成功');
      setTimeout(() => fetchNetworkDetail(), 1000);
    } catch (err: any) {
      showSnackbar('保存失败: ' + (err.response?.data?.error || err.message), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleSaveIPv6 = async () => {
    if (!id) return;
    setSaving(true);
    try {
      const v6Mode = {
        zt: v6AssignMode === 'zt',
        '6plane': v6AssignMode === '6plane',
        rfc4193: v6AssignMode === 'rfc4193'
      };
      await networkAPI.updateNetwork(id, { v6AssignMode: v6Mode });
      showSnackbar('IPv6配置保存成功');
      setTimeout(() => fetchNetworkDetail(), 1000);
    } catch (err: any) {
      showSnackbar('保存失败: ' + (err.response?.data?.error || err.message), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleSaveMulticast = async () => {
    if (!id) return;
    setSaving(true);
    try {
      await networkAPI.updateNetwork(id, {
        multicastLimit: multicastLimit,
        enableBroadcast: enableBroadcast
      });
      showSnackbar('多播设置保存成功');
      setTimeout(() => fetchNetworkDetail(), 1000);
    } catch (err: any) {
      showSnackbar('保存失败: ' + (err.response?.data?.error || err.message), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleAddIpPool = () => {
    setIpPools([...ipPools, { ipRangeStart: '', ipRangeEnd: '' }]);
  };

  const handleRemoveIpPool = (index: number) => {
    const newPools = [...ipPools];
    newPools.splice(index, 1);
    setIpPools(newPools);
  };

  const handleUpdateIpPool = (index: number, field: 'ipRangeStart' | 'ipRangeEnd', value: string) => {
    const newPools = [...ipPools];
    newPools[index] = { ...newPools[index], [field]: value };
    setIpPools(newPools);
  };

  const handleAddRoute = () => {
    setRoutes([...routes, { target: '', via: '' }]);
  };

  const handleRemoveRoute = (index: number) => {
    const newRoutes = [...routes];
    newRoutes.splice(index, 1);
    setRoutes(newRoutes);
  };

  const handleUpdateRoute = (index: number, field: 'target' | 'via', value: string) => {
    const newRoutes = [...routes];
    newRoutes[index] = { ...newRoutes[index], [field]: value };
    setRoutes(newRoutes);
  };

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setActiveTab(newValue);
  };

  const handleCloseSnackbar = () => {
    setSnackbar({ ...snackbar, open: false });
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
              <Paper elevation={3} sx={{ p: 3, mb: 4, border: basicInfoUnsaved ? '2px solid' : 'none', borderColor: 'warning.main' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h5">
                    网络基本信息
                  </Typography>
                  {basicInfoUnsaved && (
                    <Typography variant="body2" color="warning.main">
                      未保存
                    </Typography>
                  )}
                </Box>
                <Grid container spacing={3}>
                  <Grid size={{ xs: 12 }}>
                    <TextField
                      fullWidth
                      label="网络名称"
                      variant="outlined"
                      value={editingName}
                      onChange={(e) => setEditingName(e.target.value)}
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
                      value={editingDescription}
                      onChange={(e) => setEditingDescription(e.target.value)}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12 }}>
                    <Button variant="contained" color="primary" onClick={handleSaveMetadata} disabled={saving}>
                      保存
                    </Button>
                  </Grid>
                </Grid>
              </Paper>
              
              {/* IPv4分配 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4, border: ipv4Unsaved ? '2px solid' : 'none', borderColor: 'warning.main' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h5">
                    IPv4分配
                  </Typography>
                  {ipv4Unsaved && (
                    <Typography variant="body2" color="warning.main">
                      未保存
                    </Typography>
                  )}
                </Box>
                <Box sx={{ mb: 3, p: 2, bgcolor: 'background.paper', borderRadius: 1, border: '1px solid', borderColor: 'divider' }}>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    已分配的IPv4地址段
                  </Typography>
                  {ipPools.length > 0 ? (
                    <Box>
                      {ipPools.map((pool, index) => (
                        <Typography key={index} variant="body1">
                          {pool.ipRangeStart} - {pool.ipRangeEnd}
                        </Typography>
                      ))}
                    </Box>
                  ) : (
                    <Typography variant="body1">
                      未设置
                    </Typography>
                  )}
                </Box>
                
                <Box sx={{ mb: 3 }}>
                  <Typography variant="subtitle1" sx={{ mb: 2 }}>
                    IPv4地址分配池
                  </Typography>
                  {ipPools.length > 0 ? (
                    <Box sx={{ mb: 2 }}>
                      {ipPools.map((pool, index) => (
                        <Grid container spacing={2} alignItems="center" key={index} sx={{ mb: 2 }}>
                          <Grid size={{ xs: 5 }}>
                            <TextField
                              fullWidth
                              label="起始IP"
                              variant="outlined"
                              size="small"
                              value={pool.ipRangeStart}
                              onChange={(e) => handleUpdateIpPool(index, 'ipRangeStart', e.target.value)}
                            />
                          </Grid>
                          <Grid size={{ xs: 5 }}>
                            <TextField
                              fullWidth
                              label="结束IP"
                              variant="outlined"
                              size="small"
                              value={pool.ipRangeEnd}
                              onChange={(e) => handleUpdateIpPool(index, 'ipRangeEnd', e.target.value)}
                            />
                          </Grid>
                          <Grid size={{ xs: 2 }}>
                            <Button variant="outlined" color="error" fullWidth size="small" onClick={() => handleRemoveIpPool(index)}>
                              删除
                            </Button>
                          </Grid>
                        </Grid>
                      ))}
                    </Box>
                  ) : (
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                      未设置地址池
                    </Typography>
                  )}
                  <Box sx={{ display: 'flex', gap: 2 }}>
                    <Button variant="outlined" startIcon={<Add />} size="small" onClick={handleAddIpPool}>
                      添加地址池
                    </Button>
                    <Button variant="outlined" size="small" onClick={() => {
                      const randomPool: IpAssignmentPool = {
                        ipRangeStart: `10.${Math.floor(Math.random() * 256)}.${Math.floor(Math.random() * 256)}.1`,
                        ipRangeEnd: `10.${Math.floor(Math.random() * 256)}.${Math.floor(Math.random() * 256)}.254`
                      };
                      setIpPools([...ipPools, randomPool]);
                    }}>
                      随机分配一个地址池
                    </Button>
                  </Box>
                </Box>
                
                <Box sx={{ mb: 3 }}>
                  <Typography variant="subtitle1" sx={{ mb: 2 }}>
                    IPv4自动分配模式
                  </Typography>
                  <FormControlLabel
                      control={<Switch checked={v4AssignMode} onChange={(e) => setV4AssignMode(e.target.checked)} />}
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
                  {routes.length > 0 ? (
                    <Box sx={{ mb: 2 }}>
                      {routes.map((route, index) => (
                        <Grid container spacing={2} alignItems="center" key={index} sx={{ mb: 2 }}>
                          <Grid size={{ xs: 5 }}>
                            <TextField
                              fullWidth
                              label="目标网络"
                              variant="outlined"
                              size="small"
                              value={route.target}
                              onChange={(e) => handleUpdateRoute(index, 'target', e.target.value)}
                            />
                          </Grid>
                          <Grid size={{ xs: 5 }}>
                            <TextField
                              fullWidth
                              label="下一跳地址"
                              variant="outlined"
                              size="small"
                              placeholder="可选"
                              value={route.via || ''}
                              onChange={(e) => handleUpdateRoute(index, 'via', e.target.value)}
                            />
                          </Grid>
                          <Grid size={{ xs: 2 }}>
                            <Button variant="outlined" color="error" fullWidth size="small" onClick={() => handleRemoveRoute(index)}>
                              删除
                            </Button>
                          </Grid>
                        </Grid>
                      ))}
                    </Box>
                  ) : (
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                      未设置路由
                    </Typography>
                  )}
                  <Box sx={{ display: 'flex', gap: 2 }}>
                    <Button variant="outlined" startIcon={<Add />} size="small" onClick={handleAddRoute}>
                      添加路由
                    </Button>
                  </Box>
                </Box>
                
                <Box sx={{ display: 'flex', justifyContent: 'flex-start' }}>
                  <Button variant="contained" color="primary" onClick={handleSaveIPv4} disabled={saving || !ipv4Unsaved}>
                    保存
                  </Button>
                </Box>
              </Paper>
              
              {/* IPv6分配 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4, border: ipv6Unsaved ? '2px solid' : 'none', borderColor: 'warning.main' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h5">
                    IPv6分配
                  </Typography>
                  {ipv6Unsaved && (
                    <Typography variant="body2" color="warning.main">
                      未保存
                    </Typography>
                  )}
                </Box>
                
                <Box sx={{ mb: 3 }}>
                  <Typography variant="subtitle1" sx={{ mb: 2 }}>
                    IPv6地址分配
                  </Typography>
                  <FormControl component="fieldset" sx={{ mb: 2 }}>
                    <RadioGroup
                      row
                      value={v6AssignMode}
                      onChange={(e) => setV6AssignMode(e.target.value as 'zt' | '6plane' | 'rfc4193' | 'none')}
                    >
                      <FormControlLabel value="none" control={<Radio />} label="不分配IPv6地址" />
                      <FormControlLabel value="rfc4193" control={<Radio />} label="分配RFC4193唯一地址" />
                      <FormControlLabel value="6plane" control={<Radio />} label="分配6plane地址" />
                    </RadioGroup>
                  </FormControl>
                </Box>
                
                <Box sx={{ display: 'flex', justifyContent: 'flex-start' }}>
                  <Button variant="contained" color="primary" onClick={handleSaveIPv6} disabled={saving || !ipv6Unsaved}>
                    保存
                  </Button>
                </Box>
              </Paper>
              
              {/* 多播设置 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4, border: multicastUnsaved ? '2px solid' : 'none', borderColor: 'warning.main' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h5">
                    多播设置
                  </Typography>
                  {multicastUnsaved && (
                    <Typography variant="body2" color="warning.main">
                      未保存
                    </Typography>
                  )}
                </Box>
                
                <Grid container spacing={3}>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                      <Typography variant="body1">
                        多播接收者限制
                      </Typography>
                      <TextField
                        type="number"
                        size="small"
                        value={multicastLimit}
                        onChange={(e) => setMulticastLimit(parseInt(e.target.value) || 0)}
                        sx={{ width: 120 }}
                      />
                    </Box>
                  </Grid>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <FormControlLabel
                      control={<Switch checked={enableBroadcast} onChange={(e) => setEnableBroadcast(e.target.checked)} />}
                      label="启用广播"
                    />
                  </Grid>
                </Grid>
                
                <Box sx={{ display: 'flex', justifyContent: 'flex-start', mt: 3 }}>
                  <Button variant="contained" color="primary" onClick={handleSaveMulticast} disabled={saving || !multicastUnsaved}>
                    保存
                  </Button>
                </Box>
              </Paper>
              
              {/* 删除网络 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                  <Typography variant="h5" sx={{ mb: 2, color: 'error.main' }}>
                    删除网络
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    此操作不可恢复。删除网络将断开所有连接的设备，并永久删除网络配置。
                  </Typography>
                  <Button variant="contained" color="error" size="large" onClick={() => setDeleteDialogOpen(true)}>
                    删除网络
                  </Button>
                </Paper>

                {/* 删除确认对话框 */}
                <Dialog
                  open={deleteDialogOpen}
                  onClose={() => setDeleteDialogOpen(false)}
                >
                  <DialogTitle>
                    确认删除网络
                  </DialogTitle>
                  <DialogContent>
                    <DialogContentText>
                      您确定要删除网络 "{network?.name}" 吗？此操作不可恢复，将永久删除该网络及其所有配置。
                    </DialogContentText>
                  </DialogContent>
                  <DialogActions>
                    <Button onClick={() => setDeleteDialogOpen(false)} color="primary">
                      取消
                    </Button>
                    <Button onClick={handleDeleteNetwork} color="error" autoFocus disabled={saving}>
                      {saving ? '删除中...' : '确认删除'}
                    </Button>
                  </DialogActions>
                </Dialog>
            </>
          )}
        </>
      )}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert
          onClose={handleCloseSnackbar}
          severity={snackbar.severity}
          variant="filled"
          sx={{ width: '100%' }}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}

export default NetworkDetail;