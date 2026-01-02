import { useState, useEffect } from 'react';
import { Box, Typography, Paper, CircularProgress, Alert, 
  Chip, LinearProgress} from '@mui/material';
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

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        // 获取系统状态
        const statusResponse = await statusAPI.getStatus();
        setStatus(statusResponse.data);
        
        // 获取网络列表 - 暂时不使用
        // const networksResponse = await networkAPI.getAllNetworks();
        // setNetworks(networksResponse.data);
      } catch (err) {
        setError('获取数据失败，请稍后重试');
        console.error('Dashboard fetch error:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  // 定期获取系统统计信息和系统信息（仅管理员）
  useEffect(() => {
    if (!isAdmin) return;

    // 获取系统统计信息（包含操作系统信息）
    const fetchSystemStats = async () => {
      try {
        const response = await systemAPI.getSystemStats();
        console.log('System stats response:', response.data);
        setSystemStats({
          cpuUsage: response.data.cpuUsage,
          memoryUsage: response.data.memoryUsage,
          osName: response.data.osName,
          platform: response.data.platform,
          platformVersion: response.data.platformVersion,
          kernelVersion: response.data.kernelVersion,
          error: null
        });
      } catch (err: any) {
        console.error('Failed to fetch system stats:', err);
        console.error('Error details:', err.response?.data || err.message);
        setSystemStats(prev => ({
          ...prev,
          error: '无法获取系统资源统计信息'
        }));
      }
    };

    // 立即获取一次
    fetchSystemStats();

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
        </Box>
      
      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
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
                  label="运行中" 
                  color="success" 
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