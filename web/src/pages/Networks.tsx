import { useState, useEffect } from 'react';
import { Box, Typography, Button, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, CircularProgress, Alert, Modal, TextField, IconButton, Grid, Card, CardContent, Select, MenuItem, FormControl, InputLabel, Chip } from '@mui/material';
import { Link } from 'react-router-dom';
import { Add, Edit, Delete, Close } from '@mui/icons-material';
import { networkAPI, Network, NetworkConfig } from '../services/api';

// 表单数据类型定义
interface FormData {
  name: string;
  description: string;
  config: NetworkConfig;
}

function Networks() {
  const [networks, setNetworks] = useState<Network[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [openModal, setOpenModal] = useState<boolean>(false);
  const [editingNetwork, setEditingNetwork] = useState<Network | null>(null);
  const [formData, setFormData] = useState<FormData>({
    name: '',
    description: '',
    config: {
      private: true,
      allowPassiveBridging: true,
      ipAssignmentPools: [{ ipRangeStart: '192.168.192.1', ipRangeEnd: '192.168.192.254' }],
      routes: [{ target: '192.168.192.0/24' }],
      v4AssignMode: { zt: true, rfc4193: false, user: false },
      v6AssignMode: { zt: false, rfc4193: false, user: false },
      multicastLimit: 32
    }
  });
  // 搜索和筛选状态
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');


  const fetchNetworks = async () => {
    setLoading(true);
    try {
      const response = await networkAPI.getAllNetworks();
      setNetworks(response.data);
    } catch (err: any) {
      setError('获取网络列表失败');
      console.error('Fetch networks error:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNetworks();
  }, []);

  const handleOpenModal = (network: Network | null = null) => {
    setEditingNetwork(network);
    if (network) {
      setFormData({
        name: network.name || '',
        description: network.description || '',
        config: {
          private: network.config?.private ?? true,
          allowPassiveBridging: network.config?.allowPassiveBridging ?? true,
          ipAssignmentPools: network.config?.ipAssignmentPools ?? [{ ipRangeStart: '192.168.192.1', ipRangeEnd: '192.168.192.254' }],
          routes: network.config?.routes ?? [{ target: '192.168.192.0/24' }],
          v4AssignMode: network.config?.v4AssignMode ?? { zt: true, rfc4193: false, user: false },
          v6AssignMode: network.config?.v6AssignMode ?? { zt: false, rfc4193: false, user: false },
          multicastLimit: network.config?.multicastLimit ?? 32
        }
      });
    } else {
      setFormData({
        name: '',
        description: '',
        config: {
          private: true,
          allowPassiveBridging: true,
          ipAssignmentPools: [{ ipRangeStart: '192.168.192.1', ipRangeEnd: '192.168.192.254' }],
          routes: [{ target: '192.168.192.0/24' }],
          v4AssignMode: { zt: true, rfc4193: false, user: false },
          v6AssignMode: { zt: false, rfc4193: false, user: false },
          multicastLimit: 32
        }
      });
    }
    setOpenModal(true);
  };

  const handleCloseModal = () => {
    setOpenModal(false);
    setEditingNetwork(null);
    setFormData({
      name: '',
      description: '',
      config: {
        private: true,
        allowPassiveBridging: true,
        ipAssignmentPools: [{ ipRangeStart: '192.168.192.1', ipRangeEnd: '192.168.192.254' }],
        routes: [{ target: '192.168.192.0/24' }],
        v4AssignMode: { zt: true, rfc4193: false, user: false },
        v6AssignMode: { zt: false, rfc4193: false, user: false },
        multicastLimit: 32
      }
    });
  };

  // 统一处理表单变化，支持嵌套对象
  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const target = e.target as HTMLInputElement;
    const { name, value, type, checked } = target;
    
    if (!name) return;
    
    if (name.startsWith('config.')) {
      const parts = name.split('.');
      
      if (parts.length === 2) {
        // 处理二级嵌套，如 config.private
        const [, child] = parts;
        setFormData(prev => ({
          ...prev,
          config: {
            ...prev.config,
            [child]: type === 'checkbox' ? checked : value
          }
        }));
      } else if (parts.length === 3) {
        // 处理三级嵌套，如 config.v4AssignMode.zt
        const [, child, grandchild] = parts;
        setFormData(prev => ({
          ...prev,
          config: {
            ...prev.config,
            [child]: {
              ...prev.config[child as keyof typeof prev.config] as any,
              [grandchild]: type === 'checkbox' ? checked : value
            }
          }
        }));
      }
    } else {
      // 处理一级字段，如 name, description
      setFormData(prev => ({
        ...prev,
        [name]: value
      }));
    }
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      if (editingNetwork) {
        // updateNetwork只需要NetworkConfig的部分属性，不需要name和description
        await networkAPI.updateNetwork(editingNetwork.id, formData.config);
      } else {
        // createNetwork只需要name和description
        await networkAPI.createNetwork({
          name: formData.name,
          description: formData.description
        });
      }
      handleCloseModal();
      fetchNetworks();
    } catch (err: any) {
      setError(editingNetwork ? '更新网络失败' : '创建网络失败');
    }
  };

  const handleDelete = async (networkId: string) => {
    if (window.confirm('确定要删除这个网络吗？')) {
      try {
        await networkAPI.deleteNetwork(networkId);
        fetchNetworks();
      } catch (err: any) {
        setError('删除网络失败');
      }
    }
  };

  // 过滤网络列表的函数
  const getFilteredNetworks = () => {
    return networks.filter(network => {
      // 搜索过滤
      const matchesSearch = searchQuery === '' || 
        (network.name && network.name.toLowerCase().includes(searchQuery.toLowerCase())) ||
        (network.id && network.id.toLowerCase().includes(searchQuery.toLowerCase()));
      
      // 状态过滤（暂时不实现实际状态判断，因为后端还没提供）
      // const matchesStatus = statusFilter === '' || network.status === statusFilter;
      const matchesStatus = statusFilter === '';
      
      return matchesSearch && matchesStatus;
    });
  };

  // 计算统计数据
  const totalNetworks = networks.length;
  
  const totalMembers = networks.reduce((sum, network) => sum + (network.members?.length || 0), 0);
  
  const authorizedMembers = networks.reduce((sum, network) => {
    return sum + (network.members?.filter(member => member.authorized).length || 0);
  }, 0);
  
  const unauthorizedMembers = totalMembers - authorizedMembers;

  // 获取网络状态的显示信息
  const getNetworkStatusInfo = (status: string) => {
    switch (status) {
      case 'OK':
        return { label: '正常运行', color: 'success' as const };
      case 'REQUESTING_CONFIGURATION':
        return { label: '等待配置', color: 'warning' as const };
      case 'ACCESS_DENIED':
        return { label: '访问被拒绝', color: 'error' as const };
      case 'NOT_FOUND':
        return { label: '网络不存在', color: 'error' as const };
      case 'PORT_ERROR':
        return { label: '端口错误', color: 'error' as const };
      case 'CLIENT_TOO_OLD':
        return { label: '客户端版本过旧', color: 'warning' as const };
      case 'AUTHENTICATION_REQUIRED':
        return { label: '需要认证', color: 'warning' as const };
      default:
        return { label: status || '未知', color: 'default' as const };
    }
  };

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

      {/* 统计卡片区域 */}
      <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
        <Grid container spacing={3}>
          <Grid size={{ xs: 12, sm: 6, md: 4 }}>
            <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  总网络数
                </Typography>
                <Typography variant="h4">
                  {totalNetworks}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid size={{ xs: 12, sm: 6, md: 4 }}>
            <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  已认证设备数
                </Typography>
                <Typography variant="h4">
                  {authorizedMembers}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid size={{ xs: 12, sm: 6, md: 4 }}>
            <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  待认证设备数
                </Typography>
                <Typography variant="h4">
                  {unauthorizedMembers}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      </Paper>

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          {/* 搜索和筛选区域 */}
          <Box sx={{ display: 'flex', gap: 2, mb: 3, flexWrap: 'wrap', alignItems: 'center' }}>
            <TextField
              label="搜索"
              variant="outlined"
              size="small"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              sx={{ width: { xs: '100%', sm: '250px' } }}
            />
            <FormControl variant="outlined" size="small" sx={{ width: { xs: '100%', sm: '150px' } }}>
              <InputLabel>状态</InputLabel>
              <Select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                label="状态"
              >
                <MenuItem value="">全部</MenuItem>
                <MenuItem value="online">在线</MenuItem>
                <MenuItem value="offline">离线</MenuItem>
                <MenuItem value="warning">警告</MenuItem>
              </Select>
            </FormControl>
          </Box>

          <TableContainer component={Paper} sx={{ overflowX: 'auto' }}>
            <Table sx={{ minWidth: 650 }}>
              <TableHead>
                <TableRow>
                  <TableCell>名称</TableCell>
                  <TableCell>网络ID</TableCell>
                  <TableCell>设备数</TableCell>
                  <TableCell>状态</TableCell>
                  <TableCell>操作</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {getFilteredNetworks().map((network) => (
                  <TableRow key={network.id}>
                    <TableCell component="th" scope="row">
                      {network.name || '未命名网络'}
                    </TableCell>
                    <TableCell>{network.id}</TableCell>
                    <TableCell>{(network.members?.length || 0)}</TableCell>
                    <TableCell>
                      <Chip 
                        label={getNetworkStatusInfo(network.status).label}
                        color={getNetworkStatusInfo(network.status).color}
                        size="small"
                      />
                    </TableCell>
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

          {getFilteredNetworks().length === 0 && (
            <Typography variant="body1" sx={{ textAlign: 'center', mt: 5 }} color="text.secondary">
              暂无网络或未匹配到搜索结果
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
  );
}

export default Networks;