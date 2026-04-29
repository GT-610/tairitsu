import { Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, FormControlLabel, Grid, Paper, Stack, Switch, TextField, Typography } from '@mui/material'
import { useTranslation } from '../../i18n'
import type { MemberFormState, NetworkMemberDevice } from './types'

interface EditMemberDialogProps {
  open: boolean;
  saving: boolean;
  selectedMember: NetworkMemberDevice | null;
  memberForm: MemberFormState;
  onClose: () => void;
  onCopyMemberID: () => void;
  onMemberFormChange: (next: MemberFormState) => void;
  onSave: () => void;
}

function parseMemberCreationTime(value?: string | number): Date | null {
  if (value === undefined || value === null || value === '') {
    return null
  }

  const normalizedValue = typeof value === 'string' && /^\d+$/.test(value)
    ? Number(value)
    : value

  const date = typeof normalizedValue === 'number'
    ? new Date(normalizedValue < 1_000_000_000_000 ? normalizedValue * 1000 : normalizedValue)
    : new Date(normalizedValue)

  if (Number.isNaN(date.getTime()) || date.getFullYear() <= 1) {
    return null
  }

  return date
}

function formatMemberTags(member: NetworkMemberDevice | null): string {
  if (!member || member.tags.length === 0) {
    return '无'
  }

  return member.tags.map((tag) => `${tag.id}:${tag.value}`).join(', ')
}

function formatMemberCapabilities(member: NetworkMemberDevice | null): string {
  if (!member || member.capabilities.length === 0) {
    return '无'
  }

  return member.capabilities.join(', ')
}

function formatPeerLatency(value?: number): string {
  if (value === undefined || value === null || value < 0) {
    return '未知'
  }

  return `${value} ms`
}

function formatMemberVersion(member: NetworkMemberDevice | null): string {
  if (!member) {
    return '未知'
  }

  return member.peerVersion || member.clientVersion || '未知'
}

function EditMemberDialog({
  open,
  saving,
  selectedMember,
  memberForm,
  onClose,
  onCopyMemberID,
  onMemberFormChange,
  onSave,
}: EditMemberDialogProps) {
  const { formatDateTime, translateText } = useTranslation()
  const updateIps = (ipAssignments: string[]) => onMemberFormChange({ ...memberForm, ipAssignments })
  const memberCreationTime = parseMemberCreationTime(selectedMember?.creationTime)

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{translateText('编辑网络成员')}</DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              {translateText('设备详情')}
            </Typography>
            <Stack spacing={2}>
              <Box sx={{ display: 'flex', gap: 1.5, alignItems: 'flex-start' }}>
                <TextField fullWidth label={translateText('设备 ID')} value={selectedMember?.id || ''} InputProps={{ readOnly: true }} />
                <Button variant="outlined" onClick={onCopyMemberID}>{translateText('复制')}</Button>
              </Box>
              <TextField
                fullWidth
                label={translateText('设备名称')}
                value={memberForm.name}
                onChange={(event) => onMemberFormChange({ ...memberForm, name: event.target.value })}
              />
              <FormControlLabel
                control={(
                  <Switch
                    checked={memberForm.authorized}
                    onChange={(event) => onMemberFormChange({ ...memberForm, authorized: event.target.checked })}
                  />
                )}
                label={translateText(memberForm.authorized ? '设备已授权' : '设备待授权')}
              />
            </Stack>
          </Paper>

          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              {translateText('成员元信息')}
            </Typography>
            <Grid container spacing={2}>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="body2" color="text.secondary">{translateText('节点地址')}</Typography>
                <Typography variant="body1">{selectedMember?.address || translateText('未知')}</Typography>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="body2" color="text.secondary">{translateText('在线状态')}</Typography>
                <Typography variant="body1">{selectedMember ? translateText(selectedMember.online ? '在线' : '离线') : translateText('未知')}</Typography>
              </Grid>
              <Grid size={{ xs: 12 }}>
                <Typography variant="body2" color="text.secondary">{translateText('身份标识')}</Typography>
                <Typography variant="body1" sx={{ wordBreak: 'break-all' }}>{selectedMember?.identity || translateText('未知')}</Typography>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="body2" color="text.secondary">{translateText('加入时间')}</Typography>
                <Typography variant="body1">{memberCreationTime ? formatDateTime(memberCreationTime) : translateText('未知')}</Typography>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="body2" color="text.secondary">{translateText('ZeroTier 版本')}</Typography>
                <Typography variant="body1">{translateText(formatMemberVersion(selectedMember))}</Typography>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="body2" color="text.secondary">{translateText('桥接模式')}</Typography>
                <Typography variant="body1">{selectedMember ? translateText(selectedMember.activeBridge ? '已启用' : '未启用') : translateText('未知')}</Typography>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="body2" color="text.secondary">{translateText('自动分配 IP')}</Typography>
                <Typography variant="body1">{selectedMember ? translateText(selectedMember.noAutoAssignIps ? '已禁止' : '允许') : translateText('未知')}</Typography>
              </Grid>
              <Grid size={{ xs: 12 }}>
                <Typography variant="body2" color="text.secondary">{translateText('能力')}</Typography>
                <Typography variant="body1">{translateText(formatMemberCapabilities(selectedMember))}</Typography>
              </Grid>
              <Grid size={{ xs: 12 }}>
                <Typography variant="body2" color="text.secondary">{translateText('标签')}</Typography>
                <Typography variant="body1">{translateText(formatMemberTags(selectedMember))}</Typography>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="body2" color="text.secondary">{translateText('Peer 角色')}</Typography>
                <Typography variant="body1">{selectedMember?.peerRole || translateText('未知')}</Typography>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="body2" color="text.secondary">{translateText('当前路径')}</Typography>
                <Typography variant="body1" sx={{ wordBreak: 'break-all' }}>{selectedMember?.preferredPath || translateText('未知')}</Typography>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="body2" color="text.secondary">{translateText('延迟')}</Typography>
                <Typography variant="body1">{translateText(formatPeerLatency(selectedMember?.peerLatency))}</Typography>
              </Grid>
            </Grid>
          </Paper>

          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              {translateText('托管 IP')}
            </Typography>
            <Stack spacing={1.5}>
              {memberForm.ipAssignments.map((ip, index) => (
                <Box key={`${selectedMember?.id || 'member'}-ip-${index}`} sx={{ display: 'flex', gap: 1.5 }}>
                  <TextField
                    fullWidth
                    label={`IP ${index + 1}`}
                    placeholder={translateText('例如 10.22.2.1')}
                    value={ip}
                    onChange={(event) => {
                      const next = [...memberForm.ipAssignments]
                      next[index] = event.target.value
                      updateIps(next)
                    }}
                  />
                  <Button
                    variant="outlined"
                    color="error"
                    onClick={() => {
                      const next = memberForm.ipAssignments.filter((_, currentIndex) => currentIndex !== index)
                      updateIps(next.length > 0 ? next : [''])
                    }}
                    disabled={memberForm.ipAssignments.length === 1 && ip.trim() === ''}
                  >
                    {translateText('删除')}
                  </Button>
                </Box>
              ))}
              <Box>
                <Button variant="outlined" onClick={() => updateIps([...memberForm.ipAssignments, ''])}>
                  {translateText('添加 IP')}
                </Button>
              </Box>
            </Stack>
          </Paper>

          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              {translateText('高级设置')}
            </Typography>
            <Stack spacing={1}>
              <FormControlLabel
                control={(
                  <Switch
                    checked={memberForm.noAutoAssignIps}
                    onChange={(event) => onMemberFormChange({ ...memberForm, noAutoAssignIps: event.target.checked })}
                  />
                )}
                label={translateText('禁止自动分配 IP')}
              />
              <FormControlLabel
                control={(
                  <Switch
                    checked={memberForm.activeBridge}
                    onChange={(event) => onMemberFormChange({ ...memberForm, activeBridge: event.target.checked })}
                  />
                )}
                label={translateText('允许以桥接设备身份接入')}
              />
            </Stack>
          </Paper>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{translateText('取消')}</Button>
        <Button onClick={onSave} variant="contained" disabled={saving}>{translateText('保存')}</Button>
      </DialogActions>
    </Dialog>
  )
}

export default EditMemberDialog
