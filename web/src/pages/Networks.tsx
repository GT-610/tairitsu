import { useState, useEffect } from 'react';
import { Box, Typography, Button, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, CircularProgress, Alert, Modal, TextField, IconButton, Grid, Card, CardContent } from '@mui/material';
import { Link } from 'react-router-dom';
import { Add, Edit, Delete, Close } from '@mui/icons-material';
import { networkAPI, NetworkSummary } from '../services/api';

function Networks() {
  const [networks, setNetworks] = useState<NetworkSummary[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [openModal, setOpenModal] = useState<boolean>(false);
  const [editingNetwork, setEditingNetwork] = useState<NetworkSummary | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    description: ''
  });
  const [searchQuery, setSearchQuery] = useState<string>('');

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

  const handleOpenModal = (network: NetworkSummary | null = null) => {
    setEditingNetwork(network);
    if (network) {
      setFormData({
        name: network.name || '',
        description: network.description || ''
      });
    } else {
      setFormData({
        name: '',
        description: ''
      });
    }
    setOpenModal(true);
  };

  const handleCloseModal = () => {
    setOpenModal(false);
    setEditingNetwork(null);
    setFormData({
      name: '',
      description: ''
    });
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      if (editingNetwork) {
        await networkAPI.updateNetwork(editingNetwork.id, {
          private: true
        } as any);
      } else {
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

  const getFilteredNetworks = () => {
    return networks.filter(network => {
      const matchesSearch = searchQuery === '' ||
        (network.name && network.name.toLowerCase().includes(searchQuery.toLowerCase())) ||
        (network.id && network.id.toLowerCase().includes(searchQuery.toLowerCase()));

      return matchesSearch;
    });
  };

  const totalNetworks = networks.length;

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
                  -
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
                  -
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
          <Box sx={{ display: 'flex', gap: 2, mb: 3, flexWrap: 'wrap', alignItems: 'center' }}>
            <TextField
              label="搜索"
              variant="outlined"
              size="small"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              sx={{ width: { xs: '100%', sm: '250px' } }}
            />
          </Box>

          <TableContainer component={Paper} sx={{ overflowX: 'auto' }}>
            <Table sx={{ minWidth: 650 }}>
              <TableHead>
                <TableRow>
                  <TableCell>名称</TableCell>
                  <TableCell>网络ID</TableCell>
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
