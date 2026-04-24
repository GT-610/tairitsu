import { useEffect, useState } from 'react'
import { Alert, Box, Button, CircularProgress, Paper, Stack, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography } from '@mui/material'
import VisibilityIcon from '@mui/icons-material/Visibility'
import { Link } from 'react-router-dom'
import { networkAPI, type SharedNetworkSummary } from '../services/api'
import { getErrorMessage } from '../services/errors'

function SharedNetworks() {
  const [networks, setNetworks] = useState<SharedNetworkSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    void fetchSharedNetworks()
  }, [])

  const fetchSharedNetworks = async () => {
    setLoading(true)
    try {
      const response = await networkAPI.getSharedNetworks()
      setNetworks(Array.isArray(response.data) ? response.data : [])
      setError('')
    } catch (err: unknown) {
      setError(getErrorMessage(err, '获取共享网络列表失败'))
      setNetworks([])
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          共享给我
        </Typography>
        <Button variant="outlined" onClick={() => { void fetchSharedNetworks() }}>
          刷新
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : networks.length === 0 ? (
        <Paper sx={{ p: 5, textAlign: 'center' }}>
          <Typography variant="h6" sx={{ mb: 1 }}>
            当前没有共享给您的网络
          </Typography>
          <Typography variant="body2" color="text.secondary">
            当网络 owner 授予您只读查看权限后，相关网络会显示在这里。
          </Typography>
        </Paper>
      ) : (
        <TableContainer component={Paper} sx={{ overflowX: 'auto' }}>
          <Table sx={{ minWidth: 650 }}>
            <TableHead>
              <TableRow>
                <TableCell>名称</TableCell>
                <TableCell>网络ID</TableCell>
                <TableCell>Owner</TableCell>
                <TableCell>成员统计</TableCell>
                <TableCell>操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {networks.map((network) => (
                <TableRow key={network.id}>
                  <TableCell component="th" scope="row">
                    {network.name || '未命名网络'}
                    {network.description ? (
                      <Typography variant="body2" color="text.secondary">
                        {network.description}
                      </Typography>
                    ) : null}
                  </TableCell>
                  <TableCell>{network.id}</TableCell>
                  <TableCell>{network.owner_username || network.owner_id}</TableCell>
                  <TableCell>
                    {network.member_count} 台
                    <Typography variant="body2" color="text.secondary">
                      已授权 {network.authorized_member_count} / 待授权 {network.pending_member_count}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={1}>
                      <Button
                        component={Link}
                        to={`/shared-networks/${network.id}`}
                        variant="outlined"
                        size="small"
                        startIcon={<VisibilityIcon />}
                      >
                        查看设备
                      </Button>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Box>
  )
}

export default SharedNetworks
