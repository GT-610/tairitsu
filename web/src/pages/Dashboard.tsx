import React, { useState, useEffect } from 'react';
import { Box, Typography, Grid, Paper, Card, CardContent, CircularProgress, Alert, 
  Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Divider,
  Chip, LinearProgress, Button } from '@mui/material';
import { statusAPI, networkAPI, Network } from '../services/api';

// 设备类型定义
interface Device {
  id: string;
  address: string;
  name: string;
  connected: boolean;
  lastSeen: string;
}

// 系统状态类型定义
interface SystemStatus {
  online: boolean;
  zerotier?: {
    peerCount: number;
    controllerUrl: string;
  };
  system?: {
    cpuUsage: string;
    memoryUsage: string;
  };
  version: string;
}

function Dashboard() {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [networks, setNetworks] = useState<Network[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  // 模拟用户权限状态，实际应该从认证系统获取
  const [isAdmin, setIsAdmin] = useState<boolean>(false);
  
  // 模拟最近连接的设备数据
  const [recentDevices, setRecentDevices] = useState<Device[]>([
    { id: '1', address: 'zt-5678', name: 'Device-1', connected: true, lastSeen: '2分钟前' },
    { id: '2', address: 'zt-1234', name: 'Device-2', connected: true, lastSeen: '5分钟前' },
    { id: '3', address: 'zt-9876', name: 'Device-3', connected: false, lastSeen: '30分钟前' },
  ]);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        // 获取系统状态
        const statusResponse = await statusAPI.getStatus();
        setStatus(statusResponse.data);
        
        // 获取网络列表
        const networksResponse = await networkAPI.getAllNetworks();
        setNetworks(networksResponse.data);
      } catch (err) {
        setError('获取数据失败，请稍后重试');
        console.error('Dashboard fetch error:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          仪表盘
        </Typography>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Chip 
            label={status?.online ? '在线' : '离线'} 
            color={status?.online ? 'success' : 'error'}
          />
          <Button 
            variant="outlined" 
            onClick={() => setIsAdmin(!isAdmin)}
            size="small"
          >
            {isAdmin ? '切换为普通用户' : '切换为管理员'}
          </Button>
        </Box>
      </Box>
      
      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          {/* 公共信息区域 - 所有用户可见 */}
          <Typography variant="h5" component="h2" gutterBottom sx={{ mt: 2 }}>
            系统概览
          </Typography>
          
          {/* 概览卡片作为一个整体 */}
          <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
            <Grid container spacing={3}>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                  <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                    <Typography variant="h6" color="text.secondary" gutterBottom>
                      活跃网络数
                    </Typography>
                    <Typography variant="h4">
                      {networks.length}
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                  <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                    <Typography variant="h6" color="text.secondary" gutterBottom>
                      总设备数
                    </Typography>
                    <Typography variant="h4">
                      {status?.zerotier?.peerCount || 0}
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                  <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                    <Typography variant="h6" color="text.secondary" gutterBottom>
                      连接质量
                    </Typography>
                    <Box sx={{ mt: 1 }}>
                      <Typography variant="h4">优</Typography>
                      <LinearProgress 
                        variant="determinate" 
                        value={85} 
                        color="success" 
                        sx={{ mt: 1 }}
                      />
                    </Box>
                  </CardContent>
                </Card>
              </Grid>
              <Grid size={{ xs: 12, sm: 6, md: 3 }}>
                <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                  <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                    <Typography variant="h6" color="text.secondary" gutterBottom>
                      我的连接状态
                    </Typography>
                    <Typography variant="h4" color="success.main">
                      已连接
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                      IP: 10.147.17.234
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>
            </Grid>
          </Paper>
          
          {/* 最近连接的设备 */}
          <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
            <Typography variant="h5" component="h2" gutterBottom>
              最近连接的设备
            </Typography>
            <TableContainer component={Paper}>
              <Table sx={{ minWidth: 650 }} aria-label="recent devices table">
                <TableHead>
                  <TableRow>
                    <TableCell>设备名称</TableCell>
                    <TableCell>设备地址</TableCell>
                    <TableCell>连接状态</TableCell>
                    <TableCell>最后在线时间</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {recentDevices.map((device) => (
                    <TableRow
                      key={device.id}
                      sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                    >
                      <TableCell component="th" scope="row">
                        {device.name}
                      </TableCell>
                      <TableCell>{device.address}</TableCell>
                      <TableCell>
                        <Chip 
                          label={device.connected ? '在线' : '离线'} 
                          size="small"
                          color={device.connected ? 'success' : 'default'}
                        />
                      </TableCell>
                      <TableCell>{device.lastSeen}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </Paper>
          
          {/* 网络概览 */}
          <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
            <Typography variant="h5" component="h2" gutterBottom>
              网络概览
            </Typography>
            <TableContainer component={Paper}>
              <Table sx={{ minWidth: 650 }} aria-label="networks overview table">
                <TableHead>
                  <TableRow>
                    <TableCell>网络名称</TableCell>
                    <TableCell>网络ID</TableCell>
                    <TableCell>在线设备数</TableCell>
                    <TableCell>操作</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {networks.length > 0 ? networks.map((network, index) => (
                    <TableRow
                      key={network.id || index}
                      sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                    >
                      <TableCell component="th" scope="row">
                        {network.name || '未命名网络'}
                      </TableCell>
                      <TableCell>{network.id || '未知ID'}</TableCell>
                      <TableCell>{network.members.length || 0}</TableCell>
                      <TableCell>详情</TableCell>
                    </TableRow>
                  )) : (
                    <TableRow>
                      <TableCell colSpan={4} align="center">暂无网络</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          </Paper>
          
          {/* 管理员信息区域 - 仅管理员可见 */}
          {isAdmin && (
            <>
              <Divider sx={{ my: 4 }} />
              <Typography variant="h5" component="h2" gutterBottom>
                管理员区域
              </Typography>
              
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
                        {status?.system?.cpuUsage || '0'}%
                      </Typography>
                      <LinearProgress 
                        variant="determinate" 
                        value={parseInt(status?.system?.cpuUsage || '0')} 
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
                        {status?.system?.memoryUsage || '0'}%
                      </Typography>
                      <LinearProgress 
                        variant="determinate" 
                        value={parseInt(status?.system?.memoryUsage || '0')} 
                        sx={{ flexGrow: 1 }}
                      />
                    </Box>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      系统版本
                    </Typography>
                    <Typography variant="body1">
                      {status?.version || 'Unknown'}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      运行时间
                    </Typography>
                    <Typography variant="body1">
                      12天4小时
                    </Typography>
                  </Box>
                </Box>
              </Paper>
              
              {/* 控制器详情 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Typography variant="h6" component="h3" gutterBottom>
                  控制器详情
                </Typography>
                <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' }, gap: 3 }}>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      控制器URL
                    </Typography>
                    <Typography variant="body1">
                      {status?.zerotier?.controllerUrl || 'Unknown'}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      API端口
                    </Typography>
                    <Typography variant="body1">
                      9993
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
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      安全状态
                    </Typography>
                    <Chip 
                      label="正常" 
                      color="success" 
                      size="small"
                    />
                  </Box>
                </Box>
              </Paper>
              
              {/* 高级网络统计 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Typography variant="h6" component="h3" gutterBottom>
                  高级网络统计
                </Typography>
                <Alert severity="info">
                  该功能正在开发中
                </Alert>
              </Paper>
              
              {/* 安全监控 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Typography variant="h6" component="h3" gutterBottom>
                  安全监控
                </Typography>
                <Alert severity="info">
                  该功能正在开发中
                </Alert>
              </Paper>
            </>
          )}
        </>
      )}
    </Box>
  );
}

export default Dashboard;