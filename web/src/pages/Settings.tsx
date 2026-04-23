import { useEffect, useState } from 'react'
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import { authAPI } from '../services/api'
import { getErrorMessage } from '../services/errors'
import { useAuth } from '../services/auth'

function Settings() {
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState<{ severity: 'info' | 'success' | 'error'; text: string } | null>(null)
  const { user } = useAuth()

  const [openChangePasswordDialog, setOpenChangePasswordDialog] = useState(false)
  const [passwordForm, setPasswordForm] = useState({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  })
  const [passwordErrors, setPasswordErrors] = useState({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  })
  const [changingPassword, setChangingPassword] = useState(false)

  useEffect(() => {
    setLoading(false)
  }, [])

  const resetPasswordDialog = () => {
    setOpenChangePasswordDialog(false)
    setPasswordForm({
      oldPassword: '',
      newPassword: '',
      confirmPassword: '',
    })
    setPasswordErrors({
      oldPassword: '',
      newPassword: '',
      confirmPassword: '',
    })
  }

  const validatePasswordForm = () => {
    const errors = {
      oldPassword: '',
      newPassword: '',
      confirmPassword: '',
    }

    let isValid = true

    if (!passwordForm.oldPassword) {
      errors.oldPassword = '请输入原密码'
      isValid = false
    }

    if (!passwordForm.newPassword) {
      errors.newPassword = '请输入新密码'
      isValid = false
    } else if (passwordForm.newPassword.length < 6) {
      errors.newPassword = '新密码长度至少为6位'
      isValid = false
    }

    if (!passwordForm.confirmPassword) {
      errors.confirmPassword = '请再次确认新密码'
      isValid = false
    } else if (passwordForm.confirmPassword !== passwordForm.newPassword) {
      errors.confirmPassword = '两次输入的新密码不一致'
      isValid = false
    }

    setPasswordErrors(errors)
    return isValid
  }

  const handleChangePassword = async () => {
    if (!validatePasswordForm()) {
      return
    }

    try {
      setChangingPassword(true)
      await authAPI.updatePassword({
        current_password: passwordForm.oldPassword,
        new_password: passwordForm.newPassword,
        confirm_password: passwordForm.confirmPassword,
      })

      setMessage({ severity: 'success', text: '密码修改成功' })
      resetPasswordDialog()
    } catch (error: unknown) {
      const errorMessage = getErrorMessage(error, '密码修改失败，请稍后重试')
      setPasswordErrors((previous) => ({
        ...previous,
        oldPassword: errorMessage,
      }))
      setMessage({ severity: 'error', text: errorMessage })
    } finally {
      setChangingPassword(false)
    }
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          设置
        </Typography>
      </Box>

      {message && (
        <Alert severity={message.severity} sx={{ mb: 3 }} onClose={() => setMessage(null)}>
          {message.text}
        </Alert>
      )}

      <Alert severity="info" sx={{ mb: 3 }}>
        当前设置页只保留已经具备完整后端支持的账户能力。语言切换、账号注销、会话管理与登录设备查看等功能暂未进入主线交付，本页仅保留清晰说明，不再展示伪交互入口。
      </Alert>

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                账户设置
              </Typography>

              <Stack spacing={2.5} sx={{ mb: 3 }}>
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    当前账号
                  </Typography>
                  <Typography variant="body1">
                    {user?.username || '未知用户'}
                  </Typography>
                </Box>
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    当前角色
                  </Typography>
                  <Typography variant="body1">
                    {user?.role === 'admin' ? '管理员' : '普通用户'}
                  </Typography>
                </Box>
              </Stack>

              <Divider sx={{ mb: 3 }} />

              <Stack spacing={2}>
                <Button
                  variant="contained"
                  onClick={() => setOpenChangePasswordDialog(true)}
                  fullWidth
                >
                  修改密码
                </Button>
                <Typography variant="body2" color="text.secondary">
                  修改密码后，当前会话会继续保持。建议随后在其他设备或新会话中验证新密码是否正常生效。
                </Typography>
              </Stack>
            </CardContent>
          </Card>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                安全说明
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Tairitsu 当前已正式开放的账户安全能力是密码修改。其余安全入口会在具备明确后端支持和验证路径后再开放，不会以占位按钮形式提前暴露。
              </Typography>
              <Stack spacing={1.5}>
                <Alert severity="success">已可用：密码修改</Alert>
                <Alert severity="info">规划中：会话管理、登录设备查看</Alert>
                <Alert severity="info">规划中：账号注销与更完整的账户安全操作</Alert>
              </Stack>
            </CardContent>
          </Card>

          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                预留项
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                以下能力已明确不在当前阶段主线交付中。后续是否推进，将以实际后端能力和产品优先级为准，而不是先开放空壳入口。
              </Typography>
              <Stack spacing={1.5}>
                <Alert severity="info">语言切换：未进入当前交付范围</Alert>
                <Alert severity="info">账号注销：未进入当前交付范围</Alert>
                <Alert severity="info">系统级设置中心：暂不在本轮推进范围内</Alert>
              </Stack>
            </CardContent>
          </Card>
        </>
      )}

      <Dialog
        open={openChangePasswordDialog}
        onClose={resetPasswordDialog}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>修改密码</DialogTitle>
        <DialogContent>
          <Box sx={{ mt: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
            <TextField
              label="原密码"
              type="password"
              fullWidth
              value={passwordForm.oldPassword}
              onChange={(event) => {
                setPasswordForm({ ...passwordForm, oldPassword: event.target.value })
                if (passwordErrors.oldPassword) {
                  setPasswordErrors((previous) => ({ ...previous, oldPassword: '' }))
                }
              }}
              error={Boolean(passwordErrors.oldPassword)}
              helperText={passwordErrors.oldPassword}
              disabled={changingPassword}
            />
            <TextField
              label="新密码"
              type="password"
              fullWidth
              value={passwordForm.newPassword}
              onChange={(event) => {
                setPasswordForm({ ...passwordForm, newPassword: event.target.value })
                if (passwordErrors.newPassword) {
                  setPasswordErrors((previous) => ({ ...previous, newPassword: '' }))
                }
              }}
              error={Boolean(passwordErrors.newPassword)}
              helperText={passwordErrors.newPassword || '密码长度至少 6 位'}
              disabled={changingPassword}
            />
            <TextField
              label="再次确认新密码"
              type="password"
              fullWidth
              value={passwordForm.confirmPassword}
              onChange={(event) => {
                setPasswordForm({ ...passwordForm, confirmPassword: event.target.value })
                if (passwordErrors.confirmPassword) {
                  setPasswordErrors((previous) => ({ ...previous, confirmPassword: '' }))
                }
              }}
              error={Boolean(passwordErrors.confirmPassword)}
              helperText={passwordErrors.confirmPassword}
              disabled={changingPassword}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={resetPasswordDialog} disabled={changingPassword}>取消</Button>
          <Button
            variant="contained"
            onClick={() => { void handleChangePassword() }}
            disabled={changingPassword}
          >
            {changingPassword ? '修改中...' : '确认修改'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}

export default Settings
