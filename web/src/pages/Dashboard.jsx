import React, { useState, useEffect } from 'react'
import { Box, Typography, Grid, Paper, Card, CardContent, CircularProgress, Alert }
from '@mui/material'
import { statusAPI, networkAPI } from '../services/api.js'

function Dashboard() {
  const [status, setStatus] = useState(null)
  const [networks, setNetworks] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true)
      try {
        // 获取系统状态
        const statusResponse = await statusAPI.getStatus()
        setStatus(statusResponse.data)
        
        // 获取网络列表
        const networksResponse = await networkAPI.getAllNetworks()
        setNetworks(networksResponse.data)
      } catch (err) {
        setError('获取数据失败，请稍后重试')
        console.error('Dashboard fetch error:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [])

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        仪表盘
      </Typography>
      
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
          <Grid container spacing={3} sx={{ mb: 3 }}>
            <Grid size={{ xs: 12, md: 4 }}>
              <Card sx={{ backgroundColor: '#2c3e50' }}>
                <CardContent>
                  <Typography variant="h6" color="text.secondary" gutterBottom>
                    活跃网络数
                  </Typography>
                  <Typography variant="h4">
                    {networks.length}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid size={{ xs: 12, md: 4 }}>
              <Card sx={{ backgroundColor: '#2c3e50' }}>
                <CardContent>
                  <Typography variant="h6" color="text.secondary" gutterBottom>
                    ZeroTier状态
                  </Typography>
                  <Typography variant="h4"
                    color={status?.zerotier?.status === 'online' ? 'success.main' : 'error.main'}
                  >
                    {status?.zerotier?.status === 'online' ? '在线' : '离线'}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid size={{ xs: 12, md: 4 }}>
              <Card sx={{ backgroundColor: '#2c3e50' }}>
                <CardContent>
                  <Typography variant="h6" color="text.secondary" gutterBottom>
                    总设备数
                  </Typography>
                  <Typography variant="h4">
                    {status?.zerotier?.peerCount || 0}
                  </Typography>
                </CardContent>
              </Card>
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