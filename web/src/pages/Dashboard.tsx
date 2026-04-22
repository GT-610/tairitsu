import { useState, useEffect } from 'react';
import { Box, Typography, Paper, CircularProgress, Alert, 
  Chip, LinearProgress, Button, Stack} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import { statusAPI, systemAPI, SystemStatus } from '../services/api';
import { useAuth } from '../services/auth';



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

function Dashboard() {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [lastUpdatedAt, setLastUpdatedAt] = useState<Date | null>(null);
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
  
  // 信息提示状态
  const [infoMessage] = useState<string>('');

  const fetchSystemStatus = async () => {
    setLoading(true);
    setError('');
    try {
      const statusResponse = await statusAPI.getStatus();
      setStatus(statusResponse.data);
      setLastUpdatedAt(new Date());
    } catch {
      setError('获取数据失败，请稍后重试');
    } finally {
      setLoading(false);
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
      setLastUpdatedAt(new Date());
    } catch {
      setSystemStats(prev => ({
        ...prev,
        error: '无法获取系统资源统计信息'
      }));
    }
  };

  useEffect(() => {
    void fetchSystemStatus();
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
          管理员面板
        </Typography>
        <Button
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={() => {
            void fetchSystemStatus();
            if (isAdmin) {
              void fetchSystemStats();
            }
          }}
          disabled={loading}
        >
          刷新
        </Button>
      </Box>
      
      {error && (
        <Alert
          severity="error"
          sx={{ mb: 3 }}
          action={(
            <Button color="inherit" size="small" onClick={() => { void fetchSystemStatus(); }}>
              重试
            </Button>
          )}
        >
          {error}
        </Alert>
      )}
      {infoMessage && (
        <Alert severity="info" sx={{ mb: 3 }}>
          {infoMessage}
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          <Alert severity="info" sx={{ mb: 3 }}>
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} justifyContent="space-between">
              <Box>
                <Typography variant="body2" sx={{ fontWeight: 600 }}>
                  ZeroTier: {status?.zeroTierStatus || 'unknown'}
                </Typography>
                <Typography variant="body2">
                  数据库: {status?.databaseStatus || 'unknown'}
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2">
                  运行时长: {status?.uptime !== undefined ? `${Math.floor(status.uptime)} 秒` : '未知'}
                </Typography>
                <Typography variant="body2">
                  最近更新时间: {lastUpdatedAt ? lastUpdatedAt.toLocaleTimeString() : '尚未获取'}
                </Typography>
              </Box>
            </Stack>
          </Alert>

          {/* 系统健康监控 */}
          <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
            <Typography variant="h6" component="h3" gutterBottom>
              系统健康监控
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' }, gap: 3 }}>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  CPU使用率
                </Typography>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mt: 1 }}>
                  <Typography variant="body1">
                    {systemStats.error ? '无法获取' : `${systemStats.cpuUsage?.toFixed(1) || '0'}%`}
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
                  内存使用率
                </Typography>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mt: 1 }}>
                  <Typography variant="body1">
                    {systemStats.error ? '无法获取' : `${systemStats.memoryUsage?.toFixed(1) || '0'}%`}
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
                  操作系统信息
                </Typography>
                <Typography variant="body1">
                  {systemStats.osName || 'Unknown'}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                  平台: {systemStats.platform || ''} {systemStats.platformVersion || ''} | 内核: {systemStats.kernelVersion || ''}
                </Typography>
              </Box>

            </Box>
            {systemStats.error && (
              <Alert severity="info" sx={{ mt: 3 }}>
                {systemStats.error}
              </Alert>
            )}
          </Paper>
          
          {/* 控制器详情 */}
          <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
            <Typography variant="h6" component="h3" gutterBottom>
              控制器详情
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' }, gap: 3 }}>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  控制器地址
                </Typography>
                <Typography variant="body1">
                  {status?.address || 'Unknown'}
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  版本
                </Typography>
                <Typography variant="body1">
                  {status?.version || 'Unknown'}
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  控制器状态
                </Typography>
                <Chip 
                  label={status?.zeroTierStatus || '未知'}
                  color={status?.zeroTierStatus === 'online' ? 'success' : status?.zeroTierStatus === 'error' ? 'error' : 'warning'} 
                  size="small"
                />
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  数据库状态
                </Typography>
                <Chip
                  label={status?.databaseStatus || '未知'}
                  color={status?.databaseStatus === 'connected' ? 'success' : status?.databaseStatus === 'error' ? 'error' : 'warning'}
                  size="small"
                />
              </Box>
            </Box>
          </Paper>
        </>
      )}
    </Box>
  );
}

export default Dashboard;
