import React, { useState, useEffect } from 'react'
import { Box, Typography, Grid, Paper, CircularProgress }
from '@mui/material'
import { systemAPI, networkAPI } from '../services/api.js'
import StatusCard from '../components/StatusCard.jsx'
import ErrorAlert from '../components/ErrorAlert.jsx'

function Dashboard() {
  // 使用单个对象状态减少多次re-render
  const [dashboardData, setDashboardData] = useState({
    status: null,
    networks: [],
    loading: true,
    error: ''
  })

  // 使用useCallback缓存fetch函数，避免在依赖中重复创建
  const fetchData = React.useCallback(async () => {
    try {
      setDashboardData(prev => ({ ...prev, loading: true, error: '' }))
      // 获取系统状态和网络列表
      const [statusResponse, networksResponse] = await Promise.all([
        systemAPI.getStatus(),
        networkAPI.getAllNetworks()
      ])

      setDashboardData({
        status: statusResponse.data,
        networks: networksResponse.data,
        loading: false,
        error: ''
      })
    } catch (err) {
      setDashboardData(prev => ({
        ...prev,
        loading: false,
        error: '获取数据失败，请稍后重试'
      }))
      console.error('Dashboard fetch error:', err)
    }
  }, [])

  useEffect(() => {
    fetchData()
    // 每分钟刷新一次数据
    const interval = setInterval(fetchData, 60000)

    return () => clearInterval(interval)
  }, [fetchData])

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        仪表盘
      </Typography>
      
      <ErrorAlert 
        message={dashboardData.error} 
        onClose={() => setDashboardData(prev => ({ ...prev, error: '' }))} 
      />

      {dashboardData.loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          <Grid container spacing={3} sx={{ mb: 3 }}>
            <Grid item xs={12} md={4}>
            <StatusCard 
              title="活跃网络数" 
              value={dashboardData.networks.length} 
            />
          </Grid>
          <Grid item xs={12} md={4}>
            <StatusCard 
              title="ZeroTier状态" 
              value={dashboardData.status?.zerotier?.status === 'online' ? '在线' : '离线'}
              color={dashboardData.status?.zerotier?.status === 'online' ? 'success.main' : 'error.main'}
            />
          </Grid>
          <Grid item xs={12} md={4}>
            <StatusCard 
              title="总设备数" 
              value={dashboardData.status?.zerotier?.peerCount || 0}
            />
          </Grid>
          </Grid>

          <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
            <Typography variant="h5" component="h2" gutterBottom>
              系统状态
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' }, gap: 2 }}>
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
                  控制器URL
                </Typography>
                <Typography variant="body1">
                  {status?.zerotier?.controllerUrl || 'Unknown'}
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  CPU使用率
                </Typography>
                <Typography variant="body1">
                  {status?.system?.cpuUsage || '0'}%
                </Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  内存使用率
                </Typography>
                <Typography variant="body1">
                  {status?.system?.memoryUsage || '0'}%
                </Typography>
              </Box>
            </Box>
          </Paper>
        </>
      )}
    </Box>
  )
}

export default Dashboard