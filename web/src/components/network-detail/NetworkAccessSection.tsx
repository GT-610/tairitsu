import { useEffect, useMemo, useState } from 'react'
import {
  Alert,
  Autocomplete,
  Box,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  FormControl,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material'
import { useTranslation } from '../../i18n'
import { getErrorMessage } from '../../services/errors'
import { useAuth } from '../../services/auth'
import {
  networkAPI,
  type NetworkAccessCandidate,
  type NetworkAccessDetail,
  type NetworkAccessEntry,
  type NetworkRole,
} from '../../services/api'
import { assignableNetworkRoles, networkRoleColor, networkRoleLabel } from '../../utils/networkAccess'

const roleDescription: Record<Exclude<NetworkRole, 'owner'>, string> = {
  manager: '可以修改网络设置、管理设备和其他成员。',
  operator: '可以审批和管理设备，不能修改网络设置或成员。',
  viewer: '只能查看网络和设备，不能做任何修改。',
}

interface NetworkAccessSectionProps {
  networkId: string;
  access: NetworkAccessDetail | null;
  loading: boolean;
  canManage: boolean;
  onChanged: () => Promise<void>;
  onLeft: () => void;
  onNotice: (message: string, severity?: 'success' | 'error' | 'info') => void;
}

function NetworkAccessSection({
  networkId,
  access,
  loading,
  canManage,
  onChanged,
  onLeft,
  onNotice,
}: NetworkAccessSectionProps) {
  const { formatDateTime, translateText } = useTranslation()
  const { user } = useAuth()
  const [saving, setSaving] = useState(false)

  const [addOpen, setAddOpen] = useState(false)
  const [candidateQuery, setCandidateQuery] = useState('')
  const [candidates, setCandidates] = useState<NetworkAccessCandidate[]>([])
  const [candidateLoading, setCandidateLoading] = useState(false)
  const [selectedCandidate, setSelectedCandidate] = useState<NetworkAccessCandidate | null>(null)
  const [selectedRole, setSelectedRole] = useState<NetworkRole>('viewer')

  const [removeTarget, setRemoveTarget] = useState<NetworkAccessEntry | null>(null)
  const [leaveOpen, setLeaveOpen] = useState(false)

  const [transferOpen, setTransferOpen] = useState(false)
  const [transferTarget, setTransferTarget] = useState<NetworkAccessEntry | null>(null)
  const [previousOwnerRole, setPreviousOwnerRole] = useState<NetworkRole | 'none'>('manager')

  const myRole = access?.my_role
  const isOwner = myRole === 'owner'
  const assignableRoles = useMemo<NetworkRole[]>(
    () => (canManage ? assignableNetworkRoles(myRole) : []),
    [canManage, myRole],
  )

  useEffect(() => {
    if (!addOpen) return
    const query = candidateQuery.trim()
    if (query.length < 2) {
      setCandidates([])
      setCandidateLoading(false)
      return
    }

    let active = true
    const timer = window.setTimeout(() => {
      setCandidateLoading(true)
      void networkAPI.getNetworkAccessCandidates(networkId, query)
        .then((response) => {
          if (active) {
            setCandidates(Array.isArray(response.data) ? response.data : [])
          }
        })
        .catch(() => {
          if (active) setCandidates([])
        })
        .finally(() => {
          if (active) setCandidateLoading(false)
        })
    }, 300)

    return () => {
      active = false
      window.clearTimeout(timer)
    }
  }, [addOpen, candidateQuery, networkId])

  const resetAddDialog = () => {
    setAddOpen(false)
    setCandidateQuery('')
    setCandidates([])
    setSelectedCandidate(null)
    setSelectedRole('viewer')
    setCandidateLoading(false)
  }

  const handleGrant = async () => {
    if (!selectedCandidate) return
    setSaving(true)
    try {
      await networkAPI.grantNetworkAccess(networkId, selectedCandidate.id, selectedRole)
      onNotice('成员已添加')
      resetAddDialog()
      await onChanged()
    } catch (err: unknown) {
      onNotice(getErrorMessage(err, translateText('添加成员失败')), 'error')
    } finally {
      setSaving(false)
    }
  }

  const canManageTarget = (entry: NetworkAccessEntry): boolean => {
    if (!canManage || entry.user_id === access?.owner_id || entry.user_id === user?.id) return false
    if (isOwner) return true
    return entry.role === 'operator' || entry.role === 'viewer'
  }

  const handleRoleChange = async (entry: NetworkAccessEntry, role: NetworkRole) => {
    if (role === entry.role) return
    setSaving(true)
    try {
      await networkAPI.updateNetworkAccess(networkId, entry.user_id, role)
      onNotice('成员角色已更新')
      await onChanged()
    } catch (err: unknown) {
      onNotice(getErrorMessage(err, translateText('更新成员角色失败')), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleRemove = async () => {
    if (!removeTarget) return
    setSaving(true)
    try {
      await networkAPI.revokeNetworkAccess(networkId, removeTarget.user_id)
      onNotice('成员移除成功')
      setRemoveTarget(null)
      await onChanged()
    } catch (err: unknown) {
      onNotice(getErrorMessage(err, translateText('移除成员失败')), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleLeave = async () => {
    setSaving(true)
    try {
      await networkAPI.leaveNetwork(networkId)
      onNotice('已退出网络')
      setLeaveOpen(false)
      onLeft()
    } catch (err: unknown) {
      onNotice(getErrorMessage(err, translateText('退出网络失败')), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleTransfer = async () => {
    if (!transferTarget) return
    setSaving(true)
    try {
      await networkAPI.transferNetworkOwnership(networkId, transferTarget.user_id, previousOwnerRole)
      onNotice('网络所有权已转移')
      setTransferOpen(false)
      setTransferTarget(null)
      if (previousOwnerRole === 'none') {
        onLeft()
        return
      }
      await onChanged()
    } catch (err: unknown) {
      onNotice(getErrorMessage(err, translateText('转移所有权失败')), 'error')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 6 }}>
        <CircularProgress />
      </Box>
    )
  }

  if (!access) {
    return (
      <Alert severity="error" action={<Button color="inherit" size="small" onClick={() => { void onChanged() }}>{translateText('重试')}</Button>}>
        {translateText('成员与权限加载失败')}
      </Alert>
    )
  }

  return (
    <>
      <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 2, flexWrap: 'wrap', mb: 2 }}>
          <Box>
            <Typography variant="h5">{translateText('成员与权限')}</Typography>
            <Typography variant="body2" color="text.secondary">
              {translateText('成员角色只影响控制台权限，不影响设备连通性。设备仍需自行加入网络并被授权。')}
            </Typography>
          </Box>
          <Stack direction="row" spacing={1.5}>
            {canManage && (
              <Button variant="contained" onClick={() => setAddOpen(true)}>
                {translateText('添加成员')}
              </Button>
            )}
            {!isOwner && (
              <Button variant="outlined" color="warning" onClick={() => setLeaveOpen(true)}>
                {translateText('退出网络')}
              </Button>
            )}
          </Stack>
        </Box>

        <Paper variant="outlined" sx={{ p: 2, mb: 3 }}>
          <Stack direction="row" spacing={1.5} sx={{ alignItems: 'center', flexWrap: 'wrap' }}>
            <Typography variant="body1" sx={{ fontWeight: 600 }}>
              {access.owner_username || access.owner_id}
            </Typography>
            <Chip label={translateText(networkRoleLabel.owner)} color={networkRoleColor.owner} size="small" />
            <Typography variant="body2" color="text.secondary">
              {translateText('网络所有者')}
            </Typography>
            {isOwner && (
              <Button size="small" variant="text" onClick={() => setTransferOpen(true)} sx={{ ml: 'auto' }}>
                {translateText('转移所有权')}
              </Button>
            )}
          </Stack>
        </Paper>

        {access.entries.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            {translateText('当前没有其他成员。')}
          </Typography>
        ) : (
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>{translateText('用户名')}</TableCell>
                  <TableCell>{translateText('角色')}</TableCell>
                  <TableCell>{translateText('加入时间')}</TableCell>
                  <TableCell>{translateText('授权人')}</TableCell>
                  <TableCell align="right">{translateText('操作')}</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {access.entries.map((entry) => {
                  const manageable = canManageTarget(entry)
                  return (
                    <TableRow key={entry.user_id} hover>
                      <TableCell>{entry.username}</TableCell>
                      <TableCell>
                        {manageable ? (
                          <FormControl size="small" sx={{ minWidth: 120 }}>
                            <Select
                              value={entry.role}
                              onChange={(event) => { void handleRoleChange(entry, event.target.value) }}
                              disabled={saving}
                              inputProps={{ 'aria-label': `${translateText('角色')}：${entry.username}` }}
                            >
                              {assignableRoles.map((role) => (
                                <MenuItem key={role} value={role}>
                                  {translateText(networkRoleLabel[role])}
                                </MenuItem>
                              ))}
                            </Select>
                          </FormControl>
                        ) : (
                          <Chip label={translateText(networkRoleLabel[entry.role])} color={networkRoleColor[entry.role]} size="small" />
                        )}
                      </TableCell>
                      <TableCell>{formatDateTime(entry.created_at)}</TableCell>
                      <TableCell>{entry.granted_by_username || translateText('已注销用户')}</TableCell>
                      <TableCell align="right">
                        {manageable && (
                          <Button size="small" color="error" onClick={() => setRemoveTarget(entry)} disabled={saving}>
                            {translateText('移除')}
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </Paper>

      <Dialog open={addOpen} onClose={() => !saving && resetAddDialog()} maxWidth="sm" fullWidth>
        <DialogTitle>{translateText('添加成员')}</DialogTitle>
        <DialogContent>
          <Stack spacing={2.5} sx={{ mt: 1 }}>
            <Alert severity="info">
              {translateText('成员角色只影响控制台权限，不会把对方设备加入网络。')}
            </Alert>
            <Autocomplete
              options={candidates}
              getOptionLabel={(option) => option.username}
              value={selectedCandidate}
              onChange={(_event, value) => setSelectedCandidate(value)}
              inputValue={candidateQuery}
              onInputChange={(_event, value) => setCandidateQuery(value)}
              loading={candidateLoading}
              isOptionEqualToValue={(option, value) => option.id === value.id}
              noOptionsText={translateText(candidateQuery.trim().length < 2 ? '输入至少 2 个字符搜索用户' : '没有匹配的用户')}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={translateText('搜索用户名')}
                  helperText={translateText('只显示尚未加入该网络的账号')}
                />
              )}
            />
            <FormControl fullWidth>
              <InputLabel id="new-member-role-label">{translateText('角色')}</InputLabel>
              <Select
                labelId="new-member-role-label"
                label={translateText('角色')}
                value={selectedRole}
                onChange={(event) => setSelectedRole(event.target.value)}
                disabled={saving}
              >
                {assignableRoles.map((role) => (
                  <MenuItem key={role} value={role}>
                    {translateText(networkRoleLabel[role])}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            {selectedRole !== 'viewer' && (
              <Typography variant="body2" color="text.secondary">
                {translateText(roleDescription[selectedRole as Exclude<NetworkRole, 'owner'>])}
              </Typography>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={resetAddDialog} disabled={saving}>{translateText('取消')}</Button>
          <Button variant="contained" onClick={() => { void handleGrant() }} disabled={saving || !selectedCandidate}>
            {saving ? translateText('添加中...') : translateText('添加成员')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={Boolean(removeTarget)} onClose={() => !saving && setRemoveTarget(null)} maxWidth="xs" fullWidth>
        <DialogTitle>{translateText('移除成员')}</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {translateText('移除后，该用户将立即失去对此网络的访问权限。')}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRemoveTarget(null)} disabled={saving}>{translateText('取消')}</Button>
          <Button color="error" variant="contained" onClick={() => { void handleRemove() }} disabled={saving}>
            {saving ? translateText('移除中...') : translateText('确认移除')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={leaveOpen} onClose={() => !saving && setLeaveOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>{translateText('退出网络')}</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {translateText('退出后，你需要重新被邀请才能访问该网络。')}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setLeaveOpen(false)} disabled={saving}>{translateText('取消')}</Button>
          <Button color="warning" variant="contained" onClick={() => { void handleLeave() }} disabled={saving}>
            {saving ? translateText('退出中...') : translateText('确认退出')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={transferOpen} onClose={() => !saving && setTransferOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{translateText('转移所有权')}</DialogTitle>
        <DialogContent>
          <Stack spacing={2.5} sx={{ mt: 1 }}>
            <Alert severity="warning">
              {translateText('转移后，你将不再是该网络的所有者。')}
            </Alert>
            <FormControl fullWidth>
              <InputLabel id="transfer-target-label">{translateText('新的所有者')}</InputLabel>
              <Select
                labelId="transfer-target-label"
                label={translateText('新的所有者')}
                value={transferTarget?.user_id || ''}
                onChange={(event) => {
                  const next = access.entries.find((entry) => entry.user_id === event.target.value) || null
                  setTransferTarget(next)
                }}
                disabled={saving}
              >
                {access.entries.map((entry) => (
                  <MenuItem key={entry.user_id} value={entry.user_id}>
                    {entry.username}（{translateText(networkRoleLabel[entry.role])}）
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <FormControl fullWidth>
              <InputLabel id="previous-owner-role-label">{translateText('原所有者的新角色')}</InputLabel>
              <Select
                labelId="previous-owner-role-label"
                label={translateText('原所有者的新角色')}
                value={previousOwnerRole}
                onChange={(event) => setPreviousOwnerRole(event.target.value)}
                disabled={saving}
              >
                <MenuItem value="manager">{translateText('保留为管理员')}</MenuItem>
                <MenuItem value="operator">{translateText('保留为运维')}</MenuItem>
                <MenuItem value="viewer">{translateText('保留为查看者')}</MenuItem>
                <MenuItem value="none">{translateText('移除原所有者')}</MenuItem>
              </Select>
            </FormControl>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setTransferOpen(false)} disabled={saving}>{translateText('取消')}</Button>
          <Button color="warning" variant="contained" onClick={() => { void handleTransfer() }} disabled={saving || !transferTarget}>
            {saving ? translateText('转移中...') : translateText('确认转移')}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  )
}

export default NetworkAccessSection
