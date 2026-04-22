import { Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, FormControlLabel, Grid, Paper, Stack, Switch, TextField, Typography } from '@mui/material'
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
  const updateIps = (ipAssignments: string[]) => onMemberFormChange({ ...memberForm, ipAssignments })

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>编辑网络成员</DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              设备详情
            </Typography>
            <Stack spacing={2}>
              <Box sx={{ display: 'flex', gap: 1.5, alignItems: 'flex-start' }}>
                <TextField fullWidth label="设备 ID" value={selectedMember?.id || ''} InputProps={{ readOnly: true }} />
                <Button variant="outlined" onClick={onCopyMemberID}>复制</Button>
              </Box>
              <TextField
                fullWidth
                label="设备名称"
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
                label={memberForm.authorized ? '设备已授权' : '设备待授权'}
              />
            </Stack>
          </Paper>

          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              成员元信息
            </Typography>
            <Grid container spacing={2}>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="body2" color="text.secondary">ZeroTier 版本</Typography>
                <Typography variant="body1">{selectedMember?.clientVersion || 'unknown'}</Typography>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="body2" color="text.secondary">最后活动</Typography>
                <Typography variant="body1">{selectedMember?.lastSeenLabel || '未知'}</Typography>
              </Grid>
            </Grid>
          </Paper>

          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              Managed IPs
            </Typography>
            <Stack spacing={1.5}>
              {memberForm.ipAssignments.map((ip, index) => (
                <Box key={`${selectedMember?.id || 'member'}-ip-${index}`} sx={{ display: 'flex', gap: 1.5 }}>
                  <TextField
                    fullWidth
                    label={`IP ${index + 1}`}
                    placeholder="例如 10.22.2.1"
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
                    删除
                  </Button>
                </Box>
              ))}
              <Box>
                <Button variant="outlined" onClick={() => updateIps([...memberForm.ipAssignments, ''])}>
                  添加 IP
                </Button>
              </Box>
            </Stack>
          </Paper>

          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              Advanced Settings
            </Typography>
            <Stack spacing={1}>
              <FormControlLabel
                control={(
                  <Switch
                    checked={memberForm.noAutoAssignIps}
                    onChange={(event) => onMemberFormChange({ ...memberForm, noAutoAssignIps: event.target.checked })}
                  />
                )}
                label="禁止自动分配 IP"
              />
              <FormControlLabel
                control={(
                  <Switch
                    checked={memberForm.activeBridge}
                    onChange={(event) => onMemberFormChange({ ...memberForm, activeBridge: event.target.checked })}
                  />
                )}
                label="允许以桥接设备身份接入"
              />
            </Stack>
          </Paper>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button onClick={onSave} variant="contained" disabled={saving}>保存</Button>
      </DialogActions>
    </Dialog>
  )
}

export default EditMemberDialog
