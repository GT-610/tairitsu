import { useState, useEffect } from 'react';
import { Box, Typography, Paper, CircularProgress, Alert, 
  Chip, LinearProgress, Button } from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import { networkAPI, systemAPI, type NetworkSummary, type RuntimeStatus } from '../services/api';
import { useAuth } from '../services/auth';
import { useTranslation } from '../i18n';



// 系统统计信息类型扩展
interface ExtendedSystemStats {
  cpuUsage: number | null;
  memoryUsage: number | null;
  error: string | null;
  osName?: string;
  platform?: string;
  platformVersion?: string;
  kernelVersion?: string;
}

interface OverviewStats {
  networkCount: number;
  memberCount: number;
  authorizedMemberCount: number;
  pendingMemberCount: number;
}

function zeroTierStatusLabel(status?: RuntimeStatus['zeroTierStatus']) {
  switch (status) {
    case 'online':
      return '在线'
    case 'offline':
      return '离线'
    case 'error':
      return '错误'
    default:
      return '未知'
  }
}

function databaseStatusLabel(status?: RuntimeStatus['databaseStatus']) {
  switch (status) {
    case 'connected':
      return '已连接'
    case 'disconnected':
      return '未连接'
    case 'error':
      return '错误'
    default:
      return '未知'
  }
}

function Dashboard() {
  const { translateText } = useTranslation()
  const [status, setStatus] = useState<RuntimeStatus | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [overviewStats, setOverviewStats] = useState<OverviewStats>({
    networkCount: 0,
    memberCount: 0,
    authorizedMemberCount: 0,
    pendingMemberCount: 0,
  });
  const [overviewError, setOverviewError] = useState<string>('');
  // 系统统计信息状态
  const [systemStats, setSystemStats] = useState<ExtendedSystemStats>({
    cpuUsage: null,
    memoryUsage: null,
    error: null
  });
  // 从认证上下文获取用户信息
  const { user } = useAuth();
  
  // 根据用户角色判断是否为管理员
  const isAdmin = user?.role === 'admin';

  const buildOverviewStats = (networks: NetworkSummary[]): OverviewStats => {
    return networks.reduce<OverviewStats>((summary, network) => ({
      networkCount: summary.networkCount + 1,
      memberCount: summary.memberCount + (network.member_count || 0),
      authorizedMemberCount: summary.authorizedMemberCount + (network.authorized_member_count || 0),
      pendingMemberCount: summary.pendingMemberCount + (network.pending_member_count || 0),
    }), {
      networkCount: 0,
      memberCount: 0,
      authorizedMemberCount: 0,
      pendingMemberCount: 0,
    });
  };

  const fetchSystemStatus = async () => {
    setLoading(true);
    setError('');
    try {
      const statusResponse = await systemAPI.getStatus();
      setStatus(statusResponse.data);
    } catch {
      setError('获取数据失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  const fetchOverviewStats = async () => {
    try {
      setOverviewError('');
      const response = await networkAPI.getAllNetworks();
      const networks = Array.isArray(response.data) ? response.data : [];
      setOverviewStats(buildOverviewStats(networks));
    } catch {
      setOverviewError('无法获取网络总览统计信息');
    }
  };

  const fetchSystemStats = async () => {
    try {
      const response = await systemAPI.getSystemStats();
      setSystemStats({
        cpuUsage: response.data.cpuUsage,
        memoryUsage: response.data.memoryUsage,
        osName: response.data.osName,
        platform: response.data.platform,
        platformVersion: response.data.platformVersion,
        kernelVersion: response.data.kernelVersion,
        error: null
      });
    } catch {
      setSystemStats(prev => ({
        ...prev,
        error: '无法获取系统资源统计信息'
      }));
    }
  };

  useEffect(() => {
    void fetchSystemStatus();
    void fetchOverviewStats();
  }, []);

  // 定期获取系统统计信息和系统信息（仅管理员）
  useEffect(() => {
    if (!isAdmin) return;

    // 立即获取一次
    void fetchSystemStats();

    // 每5秒获取一次系统统计信息
    const interval = setInterval(fetchSystemStats, 5000);

    // 清除定时器
    return () => clearInterval(interval);
  }, [isAdmin]);

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          {translateText('管理员面板')}
        </Typography>
        <Button
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={() => {
            void fetchSystemStatus();
            void fetchOverviewStats();
            if (isAdmin) {
              void fetchSystemStats();
            }
          }}
          disabled={loading}
        >
          {translateText('刷新')}
        </Button>
      </Box>
      
      {error && (
        <Alert
          severity="error"
          sx={{ mb: 3 }}
          action={(
            <Button color="inherit" size="small" onClick={() => { void fetchSystemStatus(); }}>
              {translateText('重试')}
            </Button>
          )}
        >
          {translateText(error)}
        </Alert>
      )}
      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
            <Typography variant="h6" component="h3" gutterBottom>
              {translateText('控制器总览')}
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(4, 1fr)' }, gap: 3 }}>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {translateText('网络总数')}
                </Typography>
                <Typography variant="h4">
                  {overviewStats.networkCount}
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {translateText('成员总数')}
                </Typography>
                <Typography variant="h4">
                  {overviewStats.memberCount}
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {translateText('已授权成员')}
                </Typography>
                <Typography variant="h4">
                  {overviewStats.authorizedMemberCount}
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {translateText('待授权成员')}
                </Typography>
                <Typography variant="h4">
                  {overviewStats.pendingMemberCount}
                </Typography>
              </Box>
            </Box>
            {overviewError && (
              <Alert severity="info" sx={{ mt: 3 }}>
                {translateText(overviewError)}
              </Alert>
            )}
          </Paper>

          {/* 系统健康监控 */}
          <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
            <Typography variant="h6" component="h3" gutterBottom>
              {translateText('系统健康监控')}
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' }, gap: 3 }}>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {translateText('CPU使用率')}
                </Typography>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mt: 1 }}>
                  <Typography variant="body1">
                    {systemStats.error ? translateText('无法获取') : `${systemStats.cpuUsage?.toFixed(1) || '0'}%`}
                  </Typography>
                  <LinearProgress 
                    variant="determinate" 
                    value={systemStats.cpuUsage || 0} 
                    sx={{ flexGrow: 1 }}
                  />
                </Box>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {translateText('内存使用率')}
                </Typography>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mt: 1 }}>
                  <Typography variant="body1">
                    {systemStats.error ? translateText('无法获取') : `${systemStats.memoryUsage?.toFixed(1) || '0'}%`}
                  </Typography>
                  <LinearProgress 
                    variant="determinate" 
                    value={systemStats.memoryUsage || 0} 
                    sx={{ flexGrow: 1 }}
                  />
                </Box>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {translateText('操作系统信息')}
                </Typography>
                <Typography variant="body1">
                  {systemStats.osName || translateText('Unknown')}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                  {translateText('平台:')} {systemStats.platform || translateText('Unknown')} {systemStats.platformVersion || ''} | {translateText('内核:')} {systemStats.kernelVersion || translateText('Unknown')}
                </Typography>
              </Box>

            </Box>
            {systemStats.error && (
              <Alert severity="info" sx={{ mt: 3 }}>
                {translateText(systemStats.error)}
              </Alert>
            )}
          </Paper>
          
          {/* 控制器详情 */}
          <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
            <Typography variant="h6" component="h3" gutterBottom>
              {translateText('控制器详情')}
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' }, gap: 3 }}>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {translateText('控制器地址')}
                </Typography>
                <Typography variant="body1">
                  {status?.address || translateText('未知')}
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {translateText('版本')}
                </Typography>
                <Typography variant="body1">
                  {status?.version || translateText('未知')}
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {translateText('控制器状态')}
                </Typography>
                <Chip 
                  label={translateText(zeroTierStatusLabel(status?.zeroTierStatus))}
                  color={status?.zeroTierStatus === 'online' ? 'success' : status?.zeroTierStatus === 'error' ? 'error' : 'warning'} 
                  size="small"
                />
                {status?.zeroTierError && (
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                    {status.zeroTierError}
                  </Typography>
                )}
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {translateText('数据库状态')}
                </Typography>
                <Chip
                  label={translateText(databaseStatusLabel(status?.databaseStatus))}
                  color={status?.databaseStatus === 'connected' ? 'success' : status?.databaseStatus === 'error' ? 'error' : 'warning'}
                  size="small"
                />
                {status?.databaseError && (
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                    {status.databaseError}
                  </Typography>
                )}
              </Box>
            </Box>
          </Paper>
        </>
      )}
    </Box>
  );
}

export default Dashboard;
