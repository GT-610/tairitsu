import { useEffect, useMemo, useState } from 'react'
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Checkbox,
  Chip,
  CircularProgress,
  FormControl,
  InputLabel,
  List,
  ListItem,
  ListItemText,
  MenuItem,
  Paper,
  Select,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material'
import RefreshIcon from '@mui/icons-material/Refresh'
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline'
import { Link as RouterLink } from 'react-router-dom'
import {
  networkAPI,
  userAPI,
  type ImportNetworksResponse,
  type ImportableNetworkCandidate,
  type ImportableNetworksResponse,
  type User,
} from '../services/api'
import { getErrorMessage } from '../services/errors'
import { buildImportResultFeedback, groupImportCandidates } from '../utils/importNetwork'

function statusChipColor(status: ImportableNetworkCandidate['status']) {
  switch (status) {
    case 'available':
      return 'success'
    case 'managed':
      return 'info'
    default:
      return 'warning'
  }
}

function statusLabel(status: ImportableNetworkCandidate['status']) {
  switch (status) {
    case 'available':
      return '可接管'
    case 'managed':
      return '已接管'
    default:
      return '需处理'
  }
}

function candidateSecondary(candidate: ImportableNetworkCandidate) {
  const parts = [
    candidate.reason_message,
    candidate.owner_username ? `Owner: ${candidate.owner_username}` : '',
    typeof candidate.member_count === 'number' ? `成员 ${candidate.member_count}` : '',
    candidate.controller_status ? `状态 ${candidate.controller_status}` : '',
  ].filter(Boolean)

  return parts.join(' · ')
}

function ImportNetwork() {
  const [response, setResponse] = useState<ImportableNetworksResponse | null>(null)
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [importing, setImporting] = useState(false)
  const [selectedNetworks, setSelectedNetworks] = useState<Set<string>>(new Set())
  const [selectedOwnerId, setSelectedOwnerId] = useState('')
  const [error, setError] = useState('')
  const [successMessage, setSuccessMessage] = useState('')
  const [importResult, setImportResult] = useState<ImportNetworksResponse | null>(null)

  useEffect(() => {
    void fetchImportConsole()
  }, [])

  const fetchImportConsole = async (options?: { clearError?: boolean }) => {
    const { clearError = true } = options ?? {}
    setLoading(true)
    if (clearError) {
      setError('')
    }

    try {
      const [networkResponse, userResponse] = await Promise.all([
        networkAPI.getImportableNetworks(),
        userAPI.getAllUsers(),
      ])

      const nextResponse = networkResponse.data
      const userList = Array.isArray(userResponse.data) ? userResponse.data : []
      setResponse(nextResponse)
      setUsers(userList)
      setSelectedOwnerId((previous) => previous || userList[0]?.id || '')

      const availableIds = new Set(
        (nextResponse.candidates || [])
          .filter((candidate) => candidate.can_import)
          .map((candidate) => candidate.network_id),
      )
      setSelectedNetworks((previous) => new Set([...previous].filter((id) => availableIds.has(id))))
    } catch (err: unknown) {
      setError(getErrorMessage(err, '获取控制器接管台信息失败'))
      setResponse(null)
      setUsers([])
    } finally {
      setLoading(false)
    }
  }

  const groupedCandidates = useMemo(
    () => groupImportCandidates(response?.candidates || []),
    [response],
  )
  const availableCandidates = groupedCandidates.available
  const managedCandidates = groupedCandidates.managed
  const blockedCandidates = groupedCandidates.blocked

  const handleToggle = (networkId: string) => {
    setSelectedNetworks((previous) => {
      const next = new Set(previous)
      if (next.has(networkId)) {
        next.delete(networkId)
      } else {
        next.add(networkId)
      }
      return next
    })
  }

  const handleSelectAll = () => {
    if (selectedNetworks.size === availableCandidates.length) {
      setSelectedNetworks(new Set())
      return
    }

    setSelectedNetworks(new Set(availableCandidates.map((candidate) => candidate.network_id)))
  }

  const handleImport = async () => {
    if (selectedNetworks.size === 0) {
      setError('请至少选择一个可接管网络')
      return
    }
    if (!selectedOwnerId) {
      setError('请选择目标 owner')
      return
    }

    setImporting(true)
    setError('')
    setSuccessMessage('')
    setImportResult(null)

    try {
      const importResponse = await networkAPI.importNetworks(Array.from(selectedNetworks), selectedOwnerId)
      const result = importResponse.data
      const feedback = buildImportResultFeedback(result)

      setImportResult(result)
      setSuccessMessage(feedback.text)
      setSelectedNetworks(new Set())
      await fetchImportConsole({ clearError: feedback.severity !== 'error' })
    } catch (err: unknown) {
      setError(getErrorMessage(err, '导入网络失败'))
    } finally {
      setImporting(false)
    }
  }

  const firstImportedNetworkId = importResult?.imported[0]?.network_id || ''

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          控制器接管台
        </Typography>
        <Button
          variant="outlined"
          size="small"
          startIcon={<RefreshIcon />}
          onClick={() => { void fetchImportConsole() }}
          disabled={loading || importing}
        >
          刷新
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      {successMessage && (
        <Alert
          severity={importResult && (importResult.summary.failed > 0 || importResult.summary.skipped > 0) ? 'warning' : 'success'}
          sx={{ mb: 3 }}
          icon={<CheckCircleOutlineIcon fontSize="inherit" />}
          onClose={() => setSuccessMessage('')}
          action={firstImportedNetworkId ? (
            <Button component={RouterLink} to={`/networks/${firstImportedNetworkId}`} color="inherit" size="small">
              查看首个网络
            </Button>
          ) : (
            <Button component={RouterLink} to="/networks" color="inherit" size="small">
              查看网络
            </Button>
          )}
        >
          {successMessage}
        </Alert>
      )}

      <Alert severity="info" sx={{ mb: 3 }}>
        该页面用于接管控制器中已存在但尚未纳入 Tairitsu 管理的网络，并将其分配给指定 owner。已接管网络会保留在控制器中，仅补齐 Tairitsu 侧登记和归属关系。
      </Alert>

      <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(4, 1fr)' }, gap: 2, mb: 3 }}>
        <Card>
          <CardContent>
            <Typography variant="body2" color="text.secondary">控制器网络总数</Typography>
            <Typography variant="h4">{response?.summary.total ?? 0}</Typography>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <Typography variant="body2" color="text.secondary">可接管</Typography>
            <Typography variant="h4">{response?.summary.available ?? 0}</Typography>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <Typography variant="body2" color="text.secondary">已接管</Typography>
            <Typography variant="h4">{response?.summary.managed ?? 0}</Typography>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <Typography variant="body2" color="text.secondary">需人工处理</Typography>
            <Typography variant="h4">{response?.summary.blocked ?? 0}</Typography>
          </CardContent>
        </Card>
      </Box>

      <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          接管设置
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          先选择导入后的 owner。只有“可接管”状态的网络才会进入当前批次，已被其他 owner 接管或控制器读取异常的网络不会被强行导入。
        </Typography>
        <FormControl fullWidth disabled={loading || importing || users.length === 0}>
          <InputLabel id="owner-select-label">目标 owner</InputLabel>
          <Select
            labelId="owner-select-label"
            value={selectedOwnerId}
            label="目标 owner"
            onChange={(event) => setSelectedOwnerId(event.target.value)}
          >
            {users.map((user) => (
              <MenuItem key={user.id} value={user.id}>
                {user.username} ({user.role})
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Paper>

      {importResult && (
        <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            最近一次接管结果
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            请求 {importResult.summary.requested} 个，成功 {importResult.summary.imported} 个，失败 {importResult.summary.failed} 个，跳过 {importResult.summary.skipped} 个。
          </Typography>

          {importResult.imported.length > 0 && (
            <Box sx={{ mb: 2 }}>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>已成功接管</Typography>
              <List dense>
                {importResult.imported.map((item) => (
                  <ListItem key={`imported-${item.network_id}`} disablePadding sx={{ py: 0.5 }}>
                    <ListItemText
                      primary={item.name || item.network_id}
                      secondary={`${item.network_id} · 已分配给 ${item.owner_username || '目标 owner'}`}
                    />
                  </ListItem>
                ))}
              </List>
            </Box>
          )}

          {importResult.skipped.length > 0 && (
            <Box sx={{ mb: 2 }}>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>已跳过</Typography>
              <List dense>
                {importResult.skipped.map((item) => (
                  <ListItem key={`skipped-${item.network_id}-${item.reason_code}`} disablePadding sx={{ py: 0.5 }}>
                    <ListItemText
                      primary={item.name || item.network_id}
                      secondary={`${item.network_id} · ${item.reason_message || '已跳过'}`}
                    />
                  </ListItem>
                ))}
              </List>
            </Box>
          )}

          {importResult.failed.length > 0 && (
            <Box>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>失败项</Typography>
              <List dense>
                {importResult.failed.map((item) => (
                  <ListItem key={`failed-${item.network_id}-${item.reason_code}`} disablePadding sx={{ py: 0.5 }}>
                    <ListItemText
                      primary={item.name || item.network_id}
                      secondary={`${item.network_id} · ${item.reason_message || '导入失败'}`}
                    />
                  </ListItem>
                ))}
              </List>
            </Box>
          )}
        </Paper>
      )}

      <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Box>
            <Typography variant="h6">可接管网络</Typography>
            <Typography variant="body2" color="text.secondary">
              这些网络尚未纳入 Tairitsu 或尚未分配 owner，可直接选择并分配给目标 owner。
            </Typography>
          </Box>
          <Chip label={`已选 ${selectedNetworks.size} / ${availableCandidates.length}`} color="primary" size="small" />
        </Box>

        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
          </Box>
        ) : availableCandidates.length === 0 ? (
          <Alert severity="success">当前没有待接管网络，控制器中的网络都已经被 Tairitsu 管理或需要人工处理。</Alert>
        ) : (
          <>
            <Stack direction="row" spacing={1.5} sx={{ mb: 2 }}>
              <Button variant="outlined" onClick={handleSelectAll}>
                {selectedNetworks.size === availableCandidates.length ? '取消全选' : '全选可接管网络'}
              </Button>
              <Button
                variant="contained"
                onClick={() => { void handleImport() }}
                disabled={selectedNetworks.size === 0 || importing || selectedOwnerId === ''}
              >
                {importing ? '接管中...' : `接管所选网络 (${selectedNetworks.size})`}
              </Button>
              <Button component={RouterLink} to="/networks" variant="outlined">
                返回网络列表
              </Button>
            </Stack>

            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell padding="checkbox" />
                  <TableCell>网络</TableCell>
                  <TableCell>状态</TableCell>
                  <TableCell>成员数</TableCell>
                  <TableCell>控制器状态</TableCell>
                  <TableCell>说明</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {availableCandidates.map((candidate) => (
                  <TableRow key={candidate.network_id} hover>
                    <TableCell padding="checkbox">
                      <Checkbox
                        checked={selectedNetworks.has(candidate.network_id)}
                        onChange={() => handleToggle(candidate.network_id)}
                      />
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                        {candidate.name || candidate.network_id}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {candidate.network_id}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip label={statusLabel(candidate.status)} size="small" color={statusChipColor(candidate.status)} />
                    </TableCell>
                    <TableCell>{typeof candidate.member_count === 'number' ? candidate.member_count : '-'}</TableCell>
                    <TableCell>{candidate.controller_status || '-'}</TableCell>
                    <TableCell>
                      <Typography variant="body2">{candidate.reason_message}</Typography>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </>
        )}
      </Paper>

      <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          已被 Tairitsu 接管
        </Typography>
        {managedCandidates.length === 0 ? (
          <Typography variant="body2" color="text.secondary">当前没有已接管网络。</Typography>
        ) : (
          <List dense>
            {managedCandidates.map((candidate) => (
              <ListItem key={candidate.network_id} disablePadding sx={{ py: 0.75 }}>
                <ListItemText
                  primary={candidate.name || candidate.network_id}
                  secondary={candidateSecondary(candidate)}
                />
              </ListItem>
            ))}
          </List>
        )}
      </Paper>

      <Paper elevation={3} sx={{ p: 3 }}>
        <Typography variant="h6" gutterBottom>
          需人工处理
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          这些网络不会出现在当前导入批次中，通常是因为控制器详情读取失败或状态异常。请先排查控制器连通性，再刷新列表重试。
        </Typography>
        {blockedCandidates.length === 0 ? (
          <Typography variant="body2" color="text.secondary">当前没有需要人工处理的网络。</Typography>
        ) : (
          <List dense>
            {blockedCandidates.map((candidate) => (
              <ListItem key={candidate.network_id} disablePadding sx={{ py: 0.75 }}>
                <ListItemText
                  primary={candidate.name || candidate.network_id}
                  secondary={candidateSecondary(candidate)}
                />
              </ListItem>
            ))}
          </List>
        )}
      </Paper>
    </Box>
  )
}

export default ImportNetwork
