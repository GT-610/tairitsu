import { useState, useEffect } from 'react';
import { Box, Typography, Button, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, CircularProgress, Alert, Modal, TextField, IconButton, Grid, Card, CardContent, Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Stack, Chip } from '@mui/material';
import { Link, useLocation } from 'react-router-dom';
import { Add, Delete, Close, Refresh } from '@mui/icons-material';
import { networkAPI, type NetworkSummary, type SharedNetworkSummary } from '../services/api';
import { getErrorMessage } from '../services/errors';
import { useTranslation } from '../i18n';

type DisplayNetwork = {
  id: string;
  name?: string;
  description?: string;
  owner_id: string;
  owner_username?: string;
  member_count: number;
  authorized_member_count: number;
  pending_member_count: number;
  created_at: string;
  updated_at: string;
  readOnly: boolean;
  detailPath: string;
}

function getNavigationMessage(state: unknown): string {
  if (!state || typeof state !== 'object' || !('message' in state)) {
    return ''
  }

  const { message } = state as { message?: unknown }
  return typeof message === 'string' ? message : ''
}

function Networks() {
  const location = useLocation();
  const { translateText } = useTranslation();
  const [networks, setNetworks] = useState<DisplayNetwork[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [openModal, setOpenModal] = useState<boolean>(false);
  const [editingNetwork, setEditingNetwork] = useState<DisplayNetwork | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    description: ''
  });
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);
  const [deletingNetworkId, setDeletingNetworkId] = useState<string | null>(null);
  const [deletingNetworkName, setDeletingNetworkName] = useState<string>('');
  const navigationMessage = getNavigationMessage(location.state);

  const fetchNetworks = async () => {
    setLoading(true);
    try {
      const [ownedResult, sharedResult] = await Promise.allSettled([
        networkAPI.getAllNetworks(),
        networkAPI.getSharedNetworks(),
      ])

      if (ownedResult.status !== 'fulfilled') {
        throw ownedResult.reason
      }

      const ownedNetworks = Array.isArray(ownedResult.value.data) ? ownedResult.value.data : []
      const sharedNetworks = sharedResult.status === 'fulfilled' && Array.isArray(sharedResult.value.data)
        ? sharedResult.value.data
        : []

      setNetworks([
        ...ownedNetworks.map((network: NetworkSummary): DisplayNetwork => ({
          ...network,
          readOnly: false,
          detailPath: `/networks/${network.id}`,
        })),
        ...sharedNetworks.map((network: SharedNetworkSummary): DisplayNetwork => ({
          ...network,
          readOnly: true,
          detailPath: `/networks/shared/${network.id}`,
        })),
      ])
      if (sharedResult.status !== 'fulfilled') {
        setError(translateText('共享给我的网络暂时无法加载，当前仅显示您拥有的网络'))
      } else {
        setError('')
      }
    } catch (err: unknown) {
      setError(getErrorMessage(err, translateText('获取网络列表失败')))
      setNetworks([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void fetchNetworks()
  }, [])

  const handleOpenModal = (network: DisplayNetwork | null = null) => {
    if (network?.readOnly) {
      return
    }

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
  }

  const handleCloseModal = () => {
    setOpenModal(false);
    setEditingNetwork(null);
    setFormData({
      name: '',
      description: ''
    });
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }))
  }

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      if (editingNetwork) {
        await networkAPI.updateNetworkMetadata(editingNetwork.id, {
          name: formData.name,
          description: formData.description
        });
      } else {
        await networkAPI.createNetwork({
          name: formData.name,
          description: formData.description
        });
      }
      handleCloseModal();
      void fetchNetworks()
    } catch (err: unknown) {
      setError(getErrorMessage(err, translateText(editingNetwork ? '更新网络失败' : '创建网络失败')))
    }
  }

  const handleDeleteClick = (networkId: string, networkName: string) => {
    setDeletingNetworkId(networkId);
    setDeletingNetworkName(networkName);
    setDeleteDialogOpen(true);
  }

  const handleDeleteConfirm = async () => {
    if (!deletingNetworkId) return;
    try {
      await networkAPI.deleteNetwork(deletingNetworkId);
      void fetchNetworks()
      setDeleteDialogOpen(false)
      setDeletingNetworkId(null)
      setDeletingNetworkName('')
    } catch (err: unknown) {
      setError(getErrorMessage(err, translateText('删除网络失败')))
    }
  }

  const handleDeleteCancel = () => {
    setDeleteDialogOpen(false)
    setDeletingNetworkId(null)
    setDeletingNetworkName('')
  }

  const getFilteredNetworks = () => {
    return networks.filter((network) => {
      const matchesSearch = searchQuery === '' ||
        (network.name && network.name.toLowerCase().includes(searchQuery.toLowerCase())) ||
        (network.id && network.id.toLowerCase().includes(searchQuery.toLowerCase()))

      return matchesSearch
    })
  }

  const totalNetworks = networks.length
  const ownedNetworksCount = networks.filter((network) => !network.readOnly).length
  const readOnlyNetworksCount = networks.filter((network) => network.readOnly).length
  const filteredNetworks = getFilteredNetworks()
  const isSearchMode = searchQuery.trim() !== ''
  const isEmptyState = !loading && networks.length === 0
  const isSearchEmptyState = !loading && networks.length > 0 && filteredNetworks.length === 0

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          {translateText('网络管理')}
        </Typography>
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
          <Button variant="outlined" startIcon={<Refresh />} onClick={() => { void fetchNetworks(); }} disabled={loading}>
            {translateText('刷新')}
          </Button>
          <Button variant="contained" startIcon={<Add />} onClick={() => handleOpenModal()}>
            {translateText('创建网络')}
          </Button>
        </Stack>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}
          onClose={() => setError('')}
        >
          {error}
        </Alert>
      )}

      {navigationMessage && (
        <Alert severity="success" sx={{ mb: 3 }}>
          {navigationMessage}
        </Alert>
      )}

      <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
        <Grid container spacing={3}>
          <Grid size={{ xs: 12, sm: 6, md: 4 }}>
            <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  {translateText('总网络数')}
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
                  {translateText('我拥有')}
                </Typography>
                <Typography variant="h4">
                  {ownedNetworksCount}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid size={{ xs: 12, sm: 6, md: 4 }}>
            <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  {translateText('共享给我')}
                </Typography>
                <Typography variant="h4">
                  {readOnlyNetworksCount}
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
              label={translateText('搜索')}
              variant="outlined"
              size="small"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              sx={{ width: { xs: '100%', sm: '250px' } }}
            />
            {isSearchMode && (
              <Button variant="text" onClick={() => setSearchQuery('')}>
                {translateText('清除搜索')}
              </Button>
            )}
          </Box>

          {isEmptyState ? (
            <Paper sx={{ p: 5, textAlign: 'center' }}>
              <Typography variant="h6" sx={{ mb: 1 }}>
                {translateText('还没有任何网络')}
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                {translateText('您可以直接创建一个新网络，或前往“导入网络”把控制器中已有的网络登记到当前账号。')}
              </Typography>
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} justifyContent="center">
                <Button variant="contained" startIcon={<Add />} onClick={() => handleOpenModal()}>
                  {translateText('创建网络')}
                </Button>
                <Button component={Link} to="/import-network" variant="outlined">
                  {translateText('导入现有网络')}
                </Button>
              </Stack>
            </Paper>
          ) : isSearchEmptyState ? (
            <Paper sx={{ p: 5, textAlign: 'center' }}>
              <Typography variant="h6" sx={{ mb: 1 }}>
                {translateText('没有匹配的网络')}
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                {translateText('当前搜索条件未匹配到任何网络，请尝试缩短关键字或清空搜索后重试。')}
              </Typography>
              <Button variant="outlined" onClick={() => setSearchQuery('')}>
                {translateText('清空搜索')}
              </Button>
            </Paper>
          ) : (
            <TableContainer component={Paper} sx={{ overflowX: 'auto' }}>
              <Table sx={{ minWidth: 650 }}>
                <TableHead>
                  <TableRow>
                    <TableCell>{translateText('名称')}</TableCell>
                    <TableCell>{translateText('网络ID')}</TableCell>
                    <TableCell>{translateText('成员统计')}</TableCell>
                    <TableCell>{translateText('操作')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {filteredNetworks.map((network) => (
                    <TableRow key={network.id}>
                      <TableCell component="th" scope="row">
                        <Stack direction="row" spacing={1} alignItems="center" useFlexGap flexWrap="wrap">
                          <Typography variant="body1">
                            {network.name || translateText('未命名网络')}
                          </Typography>
                          {network.readOnly && (
                            <Chip label={translateText('只读')} size="small" color="default" variant="outlined" />
                          )}
                        </Stack>
                        {network.description ? (
                          <Typography variant="body2" color="text.secondary">
                            {network.description}
                          </Typography>
                        ) : null}
                        {network.readOnly && network.owner_username ? (
                          <Typography variant="body2" color="text.secondary">
                            {translateText('共享来源：')}{network.owner_username}
                          </Typography>
                        ) : null}
                      </TableCell>
                      <TableCell>{network.id}</TableCell>
                      <TableCell>
                        {network.member_count}{translateText(' 台')}
                        <Typography variant="body2" color="text.secondary">
                          {translateText('已授权 ')}{network.authorized_member_count} / {translateText('待授权 ')}{network.pending_member_count}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', gap: 1 }}>
                          <Button
                            component={Link}
                            to={network.detailPath}
                            variant="outlined"
                            size="small"
                          >
                            {network.readOnly ? translateText('查看设备') : translateText('详情')}
                          </Button>
                          {!network.readOnly && (
                            <Button
                              variant="outlined"
                              size="small"
                              color="error"
                              startIcon={<Delete />}
                              onClick={() => handleDeleteClick(network.id, network.name || translateText('未命名网络'))}
                            >
                              {translateText('删除')}
                            </Button>
                          )}
                        </Box>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
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
              {translateText(editingNetwork ? '编辑网络' : '创建网络')}
            </Typography>
            <IconButton onClick={handleCloseModal}>
              <Close />
            </IconButton>
          </Box>
          <Box component="form" onSubmit={(event) => { void handleSubmit(event); }} sx={{ mt: 1 }}>
            <TextField
              margin="normal"
              required
              fullWidth
              label={translateText('网络名称')}
              name="name"
              value={formData.name}
              onChange={handleChange}
            />
            <TextField
              margin="normal"
              fullWidth
              label={translateText('网络描述')}
              name="description"
              multiline
              rows={3}
              value={formData.description}
              onChange={handleChange}
            />

            <Box sx={{ display: 'flex', gap: 2, justifyContent: 'flex-end', mt: 3 }}>
              <Button onClick={handleCloseModal}>
                {translateText('取消')}
              </Button>
              <Button type="submit" variant="contained">
                {translateText(editingNetwork ? '更新' : '创建')}
              </Button>
            </Box>
          </Box>
        </Box>
      </Modal>

      <Dialog
        open={deleteDialogOpen}
        onClose={handleDeleteCancel}
      >
        <DialogTitle>
          {translateText('确认删除网络')}
        </DialogTitle>
        <DialogContent>
          <DialogContentText>
            {translateText('您确定要删除网络')} "{deletingNetworkName}" {translateText('吗？此操作不可恢复，将永久删除该网络及其所有配置。')}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteCancel}>
            {translateText('取消')}
          </Button>
          <Button onClick={() => { void handleDeleteConfirm(); }} color="error">
            {translateText('确认删除')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}

export default Networks;
