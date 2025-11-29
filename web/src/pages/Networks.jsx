import { useState, useEffect } from 'react'
import { Box, Typography, Button, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, CircularProgress, Alert, Modal, TextField, IconButton, Switch, FormControlLabel, Divider } from '@mui/material'
import { Link, useNavigate } from 'react-router-dom'
import { Add, Edit, Delete, Close } from '@mui/icons-material'
import { networkAPI } from '../services/api.js'

function Networks() {
  const [networks, setNetworks] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [openModal, setOpenModal] = useState(false)
  const [editingNetwork, setEditingNetwork] = useState(null)
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    config: {
      private: true,
      allowPassivePortForwarding: true,
      ipAssignmentPools: [{ ipRangeStart: '192.168.192.1', ipRangeEnd: '192.168.192.254' }],
      routes: [{ target: '192.168.192.0/24' }],
      v4AssignMode: { zt: true },
      v6AssignMode: { zt: false }
    }
  })
  const navigate = useNavigate()

  const fetchNetworks = async () => {
    setLoading(true)
    try {
      const response = await networkAPI.getAllNetworks()
      setNetworks(response.data)
    } catch (err) {
      setError('获取网络列表失败')
      console.error('Fetch networks error:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchNetworks()
  }, [])

  const handleOpenModal = (network = null) => {
    setEditingNetwork(network)
    if (network) {
      setFormData({
        name: network.name || '',
        description: network.description || '',
        config: {
          private: network.config?.private ?? true,
          allowPassivePortForwarding: network.config?.allowPassivePortForwarding ?? true,
          ipAssignmentPools: network.config?.ipAssignmentPools ?? [{ ipRangeStart: '192.168.192.1', ipRangeEnd: '192.168.192.254' }],
          routes: network.config?.routes ?? [{ target: '192.168.192.0/24' }],
          v4AssignMode: network.config?.v4AssignMode ?? { zt: true },
          v6AssignMode: network.config?.v6AssignMode ?? { zt: false }
        }
      })
    } else {
      setFormData({
        name: '',
        description: '',
        config: {
          private: true,
          allowPassivePortForwarding: true,
          ipAssignmentPools: [{ ipRangeStart: '192.168.192.1', ipRangeEnd: '192.168.192.254' }],
          routes: [{ target: '192.168.192.0/24' }],
          v4AssignMode: { zt: true },
          v6AssignMode: { zt: false }
        }
      })
    }
    setOpenModal(true)
  }

  const handleCloseModal = () => {
    setOpenModal(false)
    setEditingNetwork(null)
    setFormData({
      name: '',
      description: '',
      config: {
        private: true,
        allowPassivePortForwarding: true,
        ipAssignmentPools: [{ ipRangeStart: '192.168.192.1', ipRangeEnd: '192.168.192.254' }],
        routes: [{ target: '192.168.192.0/24' }],
        v4AssignMode: { zt: true },
        v6AssignMode: { zt: false }
      }
    })
  }

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    if (name.startsWith('config.')) {
      const configField = name.split('.')[1]
      setFormData({
        ...formData,
        config: {
          ...formData.config,
          [configField]: type === 'checkbox' ? checked : value
        }
      })
    } else {
      setFormData({
        ...formData,
        [name]: value
      })
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    try {
      if (editingNetwork) {
        await networkAPI.updateNetwork(editingNetwork.id, formData)
      } else {
        await networkAPI.createNetwork(formData)
      }
      handleCloseModal()
      fetchNetworks()
    } catch (err) {
      setError(editingNetwork ? '更新网络失败' : '创建网络失败')
    }
  }

  const handleDelete = async (networkId) => {
    if (window.confirm('确定要删除这个网络吗？')) {
      try {
        await networkAPI.deleteNetwork(networkId)
        fetchNetworks()
      } catch (err) {
        setError('删除网络失败')
      }
    }
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          网络管理
        </Typography>
        <Button variant="contained" startIcon={<Add />} onClick={() => handleOpenModal()}>
          创建网络
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}
          onClose={() => setError('')}
        >
          {error}
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          <TableContainer component={Paper}>
            <Table sx={{ minWidth: 650 }}>
              <TableHead>
                <TableRow>
                  <TableCell>网络ID</TableCell>
                  <TableCell>名称</TableCell>
                  <TableCell>描述</TableCell>
                  <TableCell>成员数</TableCell>
                  <TableCell>操作</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {networks.map((network) => (
                  <TableRow key={network.id}>
                    <TableCell component="th" scope="row">
                      {network.id}
                    </TableCell>
                    <TableCell>{network.name}</TableCell>
                    <TableCell>{network.description || '-'}</TableCell>
                    <TableCell>{network.memberCount || 0}</TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', gap: 1 }}>
                        <Button 
                          component={Link} 
                          to={`/networks/${network.id}`}
                          variant="outlined" 
                          size="small"
                        >
                          详情
                        </Button>
                        <Button 
                          variant="outlined" 
                          size="small"
                          startIcon={<Edit />}
                          onClick={() => handleOpenModal(network)}
                        >
                          编辑
                        </Button>
                        <Button 
                          variant="outlined" 
                          size="small"
                          color="error"
                          startIcon={<Delete />}
                          onClick={() => handleDelete(network.id)}
                        >
                          删除
                        </Button>
                      </Box>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>

          {networks.length === 0 && (
            <Typography variant="body1" sx={{ textAlign: 'center', mt: 5 }} color="text.secondary">
              暂无网络，请点击"创建网络"按钮添加
            </Typography>
          )}
        </>
      )}

      {/* 创建/编辑网络弹窗 */}
      <Modal
        open={openModal}
        onClose={handleCloseModal}
        aria-labelledby="network-modal-title"
      >
        <Box sx={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          width: { xs: '90%', sm: 600 },
          bgcolor: '#1e1e1e',
          borderRadius: 2,
          boxShadow: 24,
          p: 4,
          maxHeight: '80vh',
          overflow: 'auto'
        }}
        >
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
            <Typography id="network-modal-title" variant="h5">
              {editingNetwork ? '编辑网络' : '创建网络'}
            </Typography>
            <IconButton onClick={handleCloseModal}>
              <Close />
            </IconButton>
          </Box>
          <Box component="form" onSubmit={handleSubmit} sx={{ mt: 1 }}>
            <TextField
              margin="normal"
              required
              fullWidth
              label="网络名称"
              name="name"
              value={formData.name}
              onChange={handleChange}
            />
            <TextField
              margin="normal"
              fullWidth
              label="网络描述"
              name="description"
              multiline
              rows={3}
              value={formData.description}
              onChange={handleChange}
            />
            
            <Divider sx={{ my: 3 }} />
            <Typography variant="h6" gutterBottom>
              网络配置
            </Typography>
            
            <FormControlLabel
              control={
                <Switch
                  checked={formData.config.private}
                  onChange={(e) => setFormData({
                    ...formData,
                    config: {
                      ...formData.config,
                      private: e.target.checked
                    }
                  })}
                />
              }
              label="私有网络（需要授权）"
            />
            
            <FormControlLabel
              control={
                <Switch
                  checked={formData.config.allowPassivePortForwarding}
                  onChange={(e) => setFormData({
                    ...formData,
                    config: {
                      ...formData.config,
                      allowPassivePortForwarding: e.target.checked
                    }
                  })}
                />
              }
              label="允许被动端口转发"
            />
            
            <FormControlLabel
              control={
                <Switch
                  checked={formData.config.v4AssignMode.zt}
                  onChange={(e) => setFormData({
                    ...formData,
                    config: {
                      ...formData.config,
                      v4AssignMode: { zt: e.target.checked }
                    }
                  })}
                />
              }
              label="启用IPv4自动分配"
            />
            
            <FormControlLabel
              control={
                <Switch
                  checked={formData.config.v6AssignMode.zt}
                  onChange={(e) => setFormData({
                    ...formData,
                    config: {
                      ...formData.config,
                      v6AssignMode: { zt: e.target.checked }
                    }
                  })}
                />
              }
              label="启用IPv6自动分配"
            />
            
            <Box sx={{ display: 'flex', gap: 2, justifyContent: 'flex-end', mt: 3 }}>
              <Button onClick={handleCloseModal}>
                取消
              </Button>
              <Button type="submit" variant="contained">
                {editingNetwork ? '更新' : '创建'}
              </Button>
            </Box>
          </Box>
        </Box>
      </Modal>
    </Box>
  )
}

export default Networks