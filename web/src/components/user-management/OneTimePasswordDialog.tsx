import { Alert, Button, Dialog, DialogActions, DialogContent, DialogTitle, Stack, TextField, Typography } from '@mui/material'

interface OneTimePasswordDialogProps {
  open: boolean;
  username: string;
  password: string;
  subjectLabel: string;
  footerText: string;
  onClose: () => void;
}

function OneTimePasswordDialog({
  open,
  username,
  password,
  subjectLabel,
  footerText,
  onClose,
}: OneTimePasswordDialogProps) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>一次性临时密码</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <Alert severity="warning">
            该密码只会展示这一次。关闭后将无法再次查看，请立即通过其他方式安全告知用户。
          </Alert>
          <Typography variant="body2" color="text.secondary">
            {subjectLabel}：{username || '未知用户'}
          </Typography>
          <TextField
            label="临时密码"
            value={password}
            fullWidth
            InputProps={{ readOnly: true }}
          />
          <Typography variant="body2" color="text.secondary">
            {footerText}
          </Typography>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button variant="contained" onClick={onClose}>
          我已记录
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default OneTimePasswordDialog
