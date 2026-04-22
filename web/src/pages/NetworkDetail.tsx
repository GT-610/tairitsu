import { useState, useEffect } from 'react';
import { Box, Typography, Card, CardContent, CircularProgress, Alert, Button, Grid, IconButton, Tabs, Tab, Paper, TextField, Switch, FormControlLabel, FormControl, RadioGroup, Radio, Dialog, DialogTitle, DialogContent, DialogActions, DialogContentText, Snackbar, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, Menu, MenuItem, Stack } from '@mui/material';
import { ArrowBack, ContentCopy, Add, MoreHoriz } from '@mui/icons-material';
import { useParams, Link } from 'react-router-dom';
import { memberAPI, networkAPI, Network, NetworkConfig, NetworkMetadataUpdateRequest, IpAssignmentPool, Route, type Member as ApiMember } from '../services/api';
import { getErrorMessage } from '../services/errors';

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

type V6AssignMode = 'zt' | '6plane' | 'rfc4193' | 'none';

interface NetworkMemberDevice {
  id: string;
  name: string;
  description: string;
  authorized: boolean;
  ipAssignments: string[];
  clientVersion: string;
  lastSeenLabel: string;
  activeBridge: boolean;
  noAutoAssignIps: boolean;
}

function formatNetworkMember(member: ApiMember): NetworkMemberDevice {
  const memberID = member.id || member.nodeId || '';
  const clientVersion = member.clientVersion || (
    member.vMajor !== undefined && member.vMinor !== undefined && member.vRev !== undefined &&
    member.vMajor >= 0 && member.vMinor >= 0 && member.vRev >= 0
      ? `${member.vMajor}.${member.vMinor}.${member.vRev}`
      : 'unknown'
  );
  return {
    id: memberID,
    name: member.name || member.nodeId || member.id || '未命名设备',
    description: member.description || '',
    authorized: member.config?.authorized ?? member.authorized ?? false,
    ipAssignments: member.config?.ipAssignments ?? member.ipAssignments ?? [],
    clientVersion,
    lastSeenLabel: member.lastSeen ? new Date(member.lastSeen).toLocaleString() : '未知',
    activeBridge: member.config?.activeBridge ?? member.activeBridge ?? false,
    noAutoAssignIps: member.config?.noAutoAssignIps ?? member.noAutoAssignIps ?? false,
  };
}

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
  const [v6AssignMode, setV6AssignMode] = useState<V6AssignMode>('none');
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
  const [memberSearchTerm, setMemberSearchTerm] = useState<string>('');
  const [memberMenuAnchor, setMemberMenuAnchor] = useState<null | HTMLElement>(null);
  const [selectedMember, setSelectedMember] = useState<NetworkMemberDevice | null>(null);
  const [memberDialogOpen, setMemberDialogOpen] = useState<boolean>(false);
  const [memberDeleteDialogOpen, setMemberDeleteDialogOpen] = useState<boolean>(false);
  const [memberForm, setMemberForm] = useState<{ name: string; authorized: boolean; activeBridge: boolean; noAutoAssignIps: boolean; ipAssignments: string[] }>({
    name: '',
    authorized: false,
    activeBridge: false,
    noAutoAssignIps: false,
    ipAssignments: [''],
  });
  const [hidePendingBanner, setHidePendingBanner] = useState<boolean>(false);

  useEffect(() => {
    void fetchNetworkDetail();
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
        setHidePendingBanner(false);
      }
    } catch (err: unknown) {
      setError(getErrorMessage(err, '获取网络详情失败'));
      console.error('Fetch network detail error:', err);
    } finally {
      setLoading(false);
    }
  };

  const getV6AssignModeFromConfig = (config?: NetworkConfig['v6AssignMode']): V6AssignMode => {
    if (config?.rfc4193) return 'rfc4193';
    if (config?.['6plane']) return '6plane';
    if (config?.zt) return 'zt';
    return 'none';
  };

  const showSnackbar = (message: string, severity: 'success' | 'error' | 'info' = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const memberDevices = (network?.members || []).map(formatNetworkMember);
  const pendingMembers = memberDevices.filter((member) => !member.authorized);
  const authorizedMembers = memberDevices.filter((member) => member.authorized);
  const filteredMembers = memberDevices.filter((member) => {
    const query = memberSearchTerm.trim().toLowerCase();
    if (query === '') return true;
    return member.name.toLowerCase().includes(query) || member.id.toLowerCase().includes(query);
  });

  const closeMemberMenu = () => {
    setMemberMenuAnchor(null);
  };

  const handleOpenMemberMenu = (event: React.MouseEvent<HTMLElement>, member: NetworkMemberDevice) => {
    setMemberMenuAnchor(event.currentTarget);
    setSelectedMember(member);
  };

  const handleOpenEditMember = () => {
    if (!selectedMember) return;
    setMemberForm({
      name: selectedMember.name,
      authorized: selectedMember.authorized,
      activeBridge: selectedMember.activeBridge,
      noAutoAssignIps: selectedMember.noAutoAssignIps,
      ipAssignments: selectedMember.ipAssignments.length > 0 ? [...selectedMember.ipAssignments] : [''],
    });
    setMemberDialogOpen(true);
    closeMemberMenu();
  };

  const handleUpdateMemberStatus = async (member: NetworkMemberDevice, authorized: boolean) => {
    if (!id) return;
    setSaving(true);
    try {
      await memberAPI.updateMember(id, member.id, { authorized });
      showSnackbar(authorized ? '设备授权成功' : '设备已拒绝');
      await fetchNetworkDetail();
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, authorized ? '设备授权失败' : '设备拒绝失败'), 'error');
    } finally {
      setSaving(false);
      closeMemberMenu();
    }
  };

  const handleSaveMember = async () => {
    if (!id || !selectedMember) return;
    setSaving(true);
    try {
      await memberAPI.updateMember(id, selectedMember.id, {
        name: memberForm.name,
        authorized: memberForm.authorized,
        activeBridge: memberForm.activeBridge,
        noAutoAssignIps: memberForm.noAutoAssignIps,
        ipAssignments: memberForm.ipAssignments.map((ip) => ip.trim()).filter((ip) => ip !== ''),
      });
      showSnackbar('成员信息更新成功');
      setMemberDialogOpen(false);
      await fetchNetworkDetail();
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '更新成员失败'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteMember = async () => {
    if (!id || !selectedMember) return;
    setSaving(true);
    try {
      await memberAPI.deleteMember(id, selectedMember.id);
      showSnackbar('成员移除成功');
      setMemberDeleteDialogOpen(false);
      closeMemberMenu();
      await fetchNetworkDetail();
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '移除成员失败'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleMemberIpChange = (index: number, value: string) => {
    setMemberForm((prev) => {
      const next = [...prev.ipAssignments];
      next[index] = value;
      return { ...prev, ipAssignments: next };
    });
  };

  const handleAddMemberIP = () => {
    setMemberForm((prev) => ({ ...prev, ipAssignments: [...prev.ipAssignments, ''] }));
  };

  const handleRemoveMemberIP = (index: number) => {
    setMemberForm((prev) => {
      const next = prev.ipAssignments.filter((_, currentIndex) => currentIndex !== index);
      return { ...prev, ipAssignments: next.length > 0 ? next : [''] };
    });
  };

  const handleCopyMemberID = async () => {
    if (!selectedMember) return;
    try {
      await navigator.clipboard.writeText(selectedMember.id);
      showSnackbar('设备 ID 已复制');
    } catch {
      showSnackbar('复制设备 ID 失败', 'error');
    }
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
      setTimeout(() => { void fetchNetworkDetail(); }, 1000);
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error');
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
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '删除失败'), 'error');
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
      setTimeout(() => { void fetchNetworkDetail(); }, 1000);
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error');
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
      setTimeout(() => { void fetchNetworkDetail(); }, 1000);
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error');
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
      setTimeout(() => { void fetchNetworkDetail(); }, 1000);
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error');
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
              {pendingMembers.length > 0 && !hidePendingBanner && (
                <Alert
                  severity="warning"
                  sx={{ mb: 3 }}
                  action={(
                    <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
                      <Button
                        color="inherit"
                        variant="outlined"
                        size="small"
                        disabled={saving}
                        onClick={() => { void handleUpdateMemberStatus(pendingMembers[0], true); }}
                      >
                        授权第一个待审批成员
                      </Button>
                      <Button
                        color="inherit"
                        variant="text"
                        size="small"
                        disabled={saving}
                        onClick={() => { void handleUpdateMemberStatus(pendingMembers[0], false); }}
                      >
                        拒绝第一个待审批成员
                      </Button>
                      <Button color="inherit" size="small" onClick={() => setHidePendingBanner(true)}>
                        关闭
                      </Button>
                    </Stack>
                  )}
                >
                  当前有 {pendingMembers.length} 个待授权设备。你可以在此快速审批，也可以在下方成员列表中逐个处理。
                </Alert>
              )}

              {/* 统计卡片区域 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Grid container spacing={3}>
                  <Grid size={{ xs: 12, sm: 4 }}>
                    <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                      <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                        <Typography variant="h6" color="text.secondary" gutterBottom>
                          设备总数
                        </Typography>
                        <Typography variant="h4">
                          {memberDevices.length}
                        </Typography>
                      </CardContent>
                    </Card>
                  </Grid>
                  <Grid size={{ xs: 12, sm: 4 }}>
                    <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                      <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                        <Typography variant="h6" color="text.secondary" gutterBottom>
                          已授权设备
                        </Typography>
                        <Typography variant="h4">
                          {authorizedMembers.length}
                        </Typography>
                      </CardContent>
                    </Card>
                  </Grid>
                  <Grid size={{ xs: 12, sm: 4 }}>
                    <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                      <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                        <Typography variant="h6" color="text.secondary" gutterBottom>
                          待授权设备
                        </Typography>
                        <Typography variant="h4">
                          {pendingMembers.length}
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
                  <TextField
                    size="small"
                    placeholder="搜索设备名称或设备 ID"
                    value={memberSearchTerm}
                    onChange={(event) => setMemberSearchTerm(event.target.value)}
                    sx={{ width: { xs: '100%', sm: 280 } }}
                  />
                </Box>
                <TableContainer>
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>设备 ID</TableCell>
                        <TableCell>名称</TableCell>
                        <TableCell>状态</TableCell>
                        <TableCell>Managed IPs</TableCell>
                        <TableCell>ZT 版本</TableCell>
                        <TableCell>最后活动</TableCell>
                        <TableCell align="right">操作</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {filteredMembers.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={7} align="center" sx={{ py: 5, color: 'text.secondary' }}>
                            {memberSearchTerm ? '没有找到匹配的成员设备' : '暂无设备连接'}
                          </TableCell>
                        </TableRow>
                      ) : (
                        filteredMembers.map((member) => (
                          <TableRow key={member.id} hover>
                            <TableCell>{member.id}</TableCell>
                            <TableCell>{member.name}</TableCell>
                            <TableCell>
                              <Chip
                                label={member.authorized ? '已授权' : '待授权'}
                                color={member.authorized ? 'success' : 'warning'}
                                variant={member.authorized ? 'filled' : 'outlined'}
                                size="small"
                              />
                            </TableCell>
                            <TableCell>{member.ipAssignments.length > 0 ? member.ipAssignments.join(', ') : '-'}</TableCell>
                            <TableCell>{member.clientVersion}</TableCell>
                            <TableCell>{member.lastSeenLabel}</TableCell>
                            <TableCell align="right">
                              <IconButton onClick={(event) => handleOpenMemberMenu(event, member)}>
                                <MoreHoriz />
                              </IconButton>
                            </TableCell>
                          </TableRow>
                        ))
                      )}
                    </TableBody>
                  </Table>
                </TableContainer>
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
                    <Button variant="contained" color="primary" onClick={() => { void handleSaveMetadata(); }} disabled={saving}>
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
                  <Button variant="contained" color="primary" onClick={() => { void handleSaveIPv4(); }} disabled={saving || !ipv4Unsaved}>
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
                      onChange={(e) => setV6AssignMode(e.target.value as V6AssignMode)}
                    >
                      <FormControlLabel value="none" control={<Radio />} label="不分配IPv6地址" />
                      <FormControlLabel value="rfc4193" control={<Radio />} label="分配RFC4193唯一地址" />
                      <FormControlLabel value="6plane" control={<Radio />} label="分配6plane地址" />
                    </RadioGroup>
                  </FormControl>
                </Box>
                
                <Box sx={{ display: 'flex', justifyContent: 'flex-start' }}>
                  <Button variant="contained" color="primary" onClick={() => { void handleSaveIPv6(); }} disabled={saving || !ipv6Unsaved}>
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
                  <Button variant="contained" color="primary" onClick={() => { void handleSaveMulticast(); }} disabled={saving || !multicastUnsaved}>
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
                    <Button onClick={() => { void handleDeleteNetwork(); }} color="error" autoFocus disabled={saving}>
                      {saving ? '删除中...' : '确认删除'}
                    </Button>
                  </DialogActions>
                </Dialog>
            </>
          )}
        </>
      )}

      <Menu
        anchorEl={memberMenuAnchor}
        open={Boolean(memberMenuAnchor)}
        onClose={closeMemberMenu}
      >
        <MenuItem onClick={handleOpenEditMember}>编辑成员</MenuItem>
        {selectedMember && !selectedMember.authorized && (
          <MenuItem onClick={() => { void handleUpdateMemberStatus(selectedMember, true); }}>
            授权
          </MenuItem>
        )}
        {selectedMember && (
          <MenuItem onClick={() => { void handleUpdateMemberStatus(selectedMember, false); }}>
            拒绝设备
          </MenuItem>
        )}
        <MenuItem
          onClick={() => {
            setMemberDeleteDialogOpen(true);
            closeMemberMenu();
          }}
          sx={{ color: 'error.main' }}
        >
          移除成员
        </MenuItem>
      </Menu>

      <Dialog open={memberDialogOpen} onClose={() => setMemberDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>编辑网络成员</DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ mt: 1 }}>
            <Paper variant="outlined" sx={{ p: 2 }}>
              <Typography variant="h6" sx={{ mb: 2 }}>
                设备详情
              </Typography>
              <Stack spacing={2}>
                <Box sx={{ display: 'flex', gap: 1.5, alignItems: 'flex-start' }}>
                  <TextField
                    fullWidth
                    label="设备 ID"
                    value={selectedMember?.id || ''}
                    InputProps={{ readOnly: true }}
                  />
                  <Button variant="outlined" onClick={() => { void handleCopyMemberID(); }}>
                    复制
                  </Button>
                </Box>
                <TextField
                  fullWidth
                  label="设备名称"
                  value={memberForm.name}
                  onChange={(event) => setMemberForm((prev) => ({ ...prev, name: event.target.value }))}
                />
                <FormControlLabel
                  control={
                    <Switch
                      checked={memberForm.authorized}
                      onChange={(event) => setMemberForm((prev) => ({ ...prev, authorized: event.target.checked }))}
                    />
                  }
                  label={memberForm.authorized ? '设备已授权' : '设备待授权'}
                />
              </Stack>
            </Paper>

            <Paper variant="outlined" sx={{ p: 2 }}>
              <Typography variant="h6" sx={{ mb: 2 }}>
                成员元信息
              </Typography>
              <Grid container spacing={2}>
                <Grid size={{ xs: 12, sm: 6 }}>
                  <Typography variant="body2" color="text.secondary">ZeroTier 版本</Typography>
                  <Typography variant="body1">{selectedMember?.clientVersion || 'unknown'}</Typography>
                </Grid>
                <Grid size={{ xs: 12, sm: 6 }}>
                  <Typography variant="body2" color="text.secondary">最后活动</Typography>
                  <Typography variant="body1">{selectedMember?.lastSeenLabel || '未知'}</Typography>
                </Grid>
              </Grid>
            </Paper>

            <Paper variant="outlined" sx={{ p: 2 }}>
              <Typography variant="h6" sx={{ mb: 2 }}>
                Managed IPs
              </Typography>
              <Stack spacing={1.5}>
                {memberForm.ipAssignments.map((ip, index) => (
                  <Box key={`${selectedMember?.id || 'member'}-ip-${index}`} sx={{ display: 'flex', gap: 1.5 }}>
                    <TextField
                      fullWidth
                      label={`IP ${index + 1}`}
                      placeholder="例如 10.22.2.1"
                      value={ip}
                      onChange={(event) => handleMemberIpChange(index, event.target.value)}
                    />
                    <Button
                      variant="outlined"
                      color="error"
                      onClick={() => handleRemoveMemberIP(index)}
                      disabled={memberForm.ipAssignments.length === 1 && ip.trim() === ''}
                    >
                      删除
                    </Button>
                  </Box>
                ))}
                <Box>
                  <Button variant="outlined" onClick={handleAddMemberIP}>
                    添加 IP
                  </Button>
                </Box>
              </Stack>
            </Paper>

            <Paper variant="outlined" sx={{ p: 2 }}>
              <Typography variant="h6" sx={{ mb: 2 }}>
                Advanced Settings
              </Typography>
              <Stack spacing={1}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={memberForm.noAutoAssignIps}
                      onChange={(event) => setMemberForm((prev) => ({ ...prev, noAutoAssignIps: event.target.checked }))}
                    />
                  }
                  label="禁止自动分配 IP"
                />
                <FormControlLabel
                  control={
                    <Switch
                      checked={memberForm.activeBridge}
                      onChange={(event) => setMemberForm((prev) => ({ ...prev, activeBridge: event.target.checked }))}
                    />
                  }
                  label="允许以桥接设备身份接入"
                />
              </Stack>
            </Paper>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setMemberDialogOpen(false)}>取消</Button>
          <Button onClick={() => { void handleSaveMember(); }} variant="contained" disabled={saving}>
            保存
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={memberDeleteDialogOpen} onClose={() => setMemberDeleteDialogOpen(false)}>
        <DialogTitle>确认移除成员</DialogTitle>
        <DialogContent>
          <DialogContentText>
            确定要将成员 "{selectedMember?.name || selectedMember?.id}" 从网络中移除吗？此操作会删除该成员在当前网络中的记录。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setMemberDeleteDialogOpen(false)}>取消</Button>
          <Button onClick={() => { void handleDeleteMember(); }} color="error" disabled={saving}>
            移除
          </Button>
        </DialogActions>
      </Dialog>

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
