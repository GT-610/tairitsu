import React, { useState, useEffect, useCallback } from 'react'
import { Box, Typography, Button, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, CircularProgress, Modal, TextField, IconButton }
from '@mui/material'
import { Link, useNavigate } from 'react-router-dom'
import { Add, Edit, Delete, Close } from '@mui/icons-material'
import { networkAPI } from '../services/api.js'
import ErrorAlert from '../components/ErrorAlert.jsx'

function Networks() {
  const navigate = useNavigate()
  
  // 合并状态为单个对象，减少渲染次数
  const [state, setState] = useState({
    networks: [],
    loading: true,
    error: '',
    openModal: false,
    editingNetwork: null,
    formData: {
      name: '',
      description: ''
    },
    submitting: false // 添加提交状态以防止重复提交
  })

  // 使用useCallback缓存fetchNetworks函数
  const fetchNetworks = useCallback(async () => {
    setState(prev => ({ ...prev, loading: true }))
    try {
      const response = await networkAPI.getAllNetworks()
      setState(prev => ({ ...prev, networks: response.data }))
    } catch (err) {
      setState(prev => ({ 
        ...prev, 
        error: '获取网络列表失败',
        loading: false 
      }))
      console.error('Fetch networks error:', err)
    } finally {
      setState(prev => ({ ...prev, loading: false }))
    }
  }, [])

  useEffect(() => {
    fetchNetworks()
  }, [fetchNetworks])

  // 使用useCallback缓存modal相关函数
  const handleOpenModal = useCallback((network = null) => {
    setState(prev => ({
      ...prev,
      editingNetwork: network,
      formData: network ? {
        name: network.name || '',
        description: network.description || ''
      } : {
        name: '',
        description: ''
      },
      openModal: true
    }))
  }, [])

  const handleCloseModal = useCallback(() => {
    setState(prev => ({
      ...prev,
      openModal: false,
      editingNetwork: null,
      formData: { name: '', description: '' }
    }))
  }, [])

  const handleChange = useCallback((e) => {
    setState(prev => ({
      ...prev,
      formData: {
        ...prev.formData,
        [e.target.name]: e.target.value
      }
    }))
  }, [])

  const handleSubmit = useCallback(async (e) => {
    e.preventDefault()
    // 防止重复提交
    if (state.submitting) return
    
    setState(prev => ({ ...prev, submitting: true, error: '' }))
    
    try {
      if (state.editingNetwork) {
        await networkAPI.updateNetwork(state.editingNetwork.id, state.formData)
      } else {
        await networkAPI.createNetwork(state.formData)
      }
      handleCloseModal()
      // 使用setTimeout轻微延迟以改善用户体验
      setTimeout(() => {
        fetchNetworks()
      }, 100)
    } catch (err) {
      setState(prev => ({
        ...prev,
        error: state.editingNetwork ? '更新网络失败' : '创建网络失败',
        submitting: false
      }))
      console.error('Network submission error:', err)
    } finally {
      setState(prev => ({ ...prev, submitting: false }))
    }
  }, [state.editingNetwork, state.formData, state.submitting, handleCloseModal, fetchNetworks])

  const handleDelete = useCallback(async (networkId) => {
    // 使用更现代的confirm对话框样式（Material-UI的Dialog在后续可以替换）
    if (window.confirm('确定要删除这个网络吗？')) {
      try {
        await networkAPI.deleteNetwork(networkId)
        // 使用setTimeout轻微延迟以改善用户体验
        setTimeout(() => {
          fetchNetworks()
        }, 100)
      } catch (err) {
        setState(prev => ({ ...prev, error: '删除网络失败' }))
        console.error('Delete network error:', err)
      }
    }
  }, [fetchNetworks])

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

      {state.error && (
        <ErrorAlert 
          message={state.error} 
          onClose={() => setState(prev => ({ ...prev, error: '' }))} 
        />
      )}

      {state.loading ? (
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
                {state.networks.map((network) => (
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

          {state.networks.length === 0 && (
            <Typography variant="body1" sx={{ textAlign: 'center', mt: 5 }} color="text.secondary">
              暂无网络，请点击"创建网络"按钮添加
            </Typography>
          )}
        </>
      )}

      {/* 创建/编辑网络弹窗 */}
      <Modal
        open={state.openModal}
        onClose={handleCloseModal}
        aria-labelledby="network-modal-title"
      >
        <Box sx={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          width: { xs: '90%', sm: 500 },
          bgcolor: '#1e1e1e',
          borderRadius: 2,
          boxShadow: 24,
          p: 4
        }}
        >
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
            <Typography id="network-modal-title" variant="h5">
              {state.editingNetwork ? '编辑网络' : '创建网络'}
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
              value={state.formData.name}
              onChange={handleChange}
            />
            <TextField
              margin="normal"
              fullWidth
              label="网络描述"
              name="description"
              multiline
              rows={3}
              value={state.formData.description}
              onChange={handleChange}
            />
            <Box sx={{ display: 'flex', gap: 2, justifyContent: 'flex-end', mt: 3 }}>
              <Button onClick={handleCloseModal}>
                取消
              </Button>
              <Button 
                type="submit" 
                variant="contained"
                disabled={state.submitting}
              >
                {state.submitting ? (
                  <CircularProgress size={16} />
                ) : (
                  state.editingNetwork ? '更新' : '创建'
                )}
              </Button>
            </Box>
          </Box>
        </Box>
      </Modal>
    </Box>
  )
}

export default Networks