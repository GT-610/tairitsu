import React, { useState, useEffect } from 'react'
import { Box, Typography, Card, CardContent, CircularProgress, Alert, Button, Divider, Grid, Chip }
from '@mui/material'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { networkAPI } from '../services/api.js'

function NetworkDetail() {
  const { id } = useParams()
  const [network, setNetwork] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const navigate = useNavigate()

  useEffect(() => {
    fetchNetworkDetail()
  }, [id])

  const fetchNetworkDetail = async () => {
    setLoading(true)
    try {
      const response = await networkAPI.getNetwork(id)
      setNetwork(response.data)
    } catch (err) {
      setError('获取网络详情失败')
      console.error('Fetch network detail error:', err)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <Box sx={{ p: 3, display: 'flex', justifyContent: 'center', mt: 10 }}>
        <CircularProgress />
      </Box>
    )
  }

  if (error || !network) {
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
    )
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', gap: 2, alignItems: 'center', mb: 3 }}>
        <Button 
          variant="outlined" 
          component={Link} 
          to="/networks"
        >
          ← 返回列表
        </Button>
        <Typography variant="h4" component="h1">
          网络详情
        </Typography>
      </Box>

      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h5" gutterBottom>
                {network.name}
              </Typography>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                网络ID: {network.id}
              </Typography>
              <Divider sx={{ my: 2 }} />
              <Typography variant="body1" paragraph>
                {network.description || '暂无描述'}
              </Typography>
              <Box sx={{ display: 'flex', gap: 2, mt: 2 }}>
                <Button 
                  variant="contained" 
                  component={Link} 
                  to={`/networks/${network.id}/members`}
                >
                  管理成员
                </Button>
              </Box>
            </CardContent>
          </Card>

          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                网络配置
              </Typography>
              <Box sx={{ display: 'grid', gap: 2 }}>
                <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 1 }}>
                  <Typography variant="body2" color="text.secondary">
                    允许默认路由
                  </Typography>
                  <Typography variant="body1">
                    {network.config?.allowDefault || false ? '是' : '否'}
                  </Typography>
                </Box>
                <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 1 }}>
                  <Typography variant="body2" color="text.secondary">
                    允许DNS
                  </Typography>
                  <Typography variant="body1">
                    {network.config?.allowDNS || false ? '是' : '否'}
                  </Typography>
                </Box>
                <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 1 }}>
                  <Typography variant="body2" color="text.secondary">
                    私有网络
                  </Typography>
                  <Typography variant="body1">
                    {network.config?.private || true ? '是' : '否'}
                  </Typography>
                </Box>
                {network.config?.v4AssignMode?.zt && (
                    <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 1 }}>
                      <Typography variant="body2" color="text.secondary">
                        IPv4 CIDR
                      </Typography>
                      <Typography variant="body1">
                        {network.config?.v4AssignMode?.zt?.cidr || '未设置'}
                      </Typography>
                    </Box>
                  </>
                ) : null}
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                统计信息
              </Typography>
              <Box sx={{ display: 'grid', gap: 2 }}>
                <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 1 }}>
                  <Typography variant="body2" color="text.secondary">
                    成员数
                  </Typography>
                  <Typography variant="body1">
                    {network.memberCount || 0}
                  </Typography>
                </Box>
                <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 1 }}>
                  <Typography variant="body2" color="text.secondary">
                    已授权
                  </Typography>
                  <Typography variant="body1">
                    {network.authorizedMemberCount || 0}
                  </Typography>
                </Box>
                <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 1 }}>
                  <Typography variant="body2" color="text.secondary">
                    在线成员
                  </Typography>
                  <Typography variant="body1">
                    {network.onlineMemberCount || 0}
                  </Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>

          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                访问控制
              </Typography>
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mt: 1 }}>
                {network.config?.private ? (
                    <>
                    <Chip label="需要授权" color="primary" size="small" />
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                      新设备需要管理员批准才能加入网络
                    </Typography>
                  </>
                ) : (
                    <>
                    <Chip label="开放加入" color="success" size="small" />
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                      任何设备都可以加入此网络
                    </Typography>
                  </>
                )}
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  )
}

export default NetworkDetail