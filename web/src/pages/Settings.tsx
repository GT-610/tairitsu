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
  FormControl,
  FormControlLabel,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  Switch,
  TextField,
  Typography,
} from '@mui/material'
import { authAPI, type UserSession } from '../services/api'
import { getErrorMessage } from '../services/errors'
import { useAuth } from '../services/auth'
import { useNavigate } from 'react-router-dom'
import { formatSessionPresentation, formatSessionTime, hasDisplayableSessionTime } from '../utils/sessionPresentation'
import { getUserRoleLabel } from '../utils/userPresentation'
import { useTranslation, type LanguagePreference } from '../i18n'

function Settings() {
  const navigate = useNavigate()
  const { user, logout } = useAuth()
  const { preference, setPreference, t, translateText } = useTranslation()

  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState<{ severity: 'info' | 'success' | 'error'; text: string } | null>(null)

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
  const [logoutOtherSessionsOnPasswordChange, setLogoutOtherSessionsOnPasswordChange] = useState(true)
  const [changingPassword, setChangingPassword] = useState(false)
  const [sessions, setSessions] = useState<UserSession[]>([])
  const [loadingSessions, setLoadingSessions] = useState(false)
  const [revokingSessionId, setRevokingSessionId] = useState('')
  const [revokingOtherSessions, setRevokingOtherSessions] = useState(false)
  const visibleSessions = sessions.filter((sessionItem) => !sessionItem.revokedAt)
  const otherVisibleSessionCount = visibleSessions.filter((sessionItem) => !sessionItem.current).length

  useEffect(() => {
    const loadSettings = async () => {
      try {
        setLoading(true)
        setLoadingSessions(true)
        const sessionResponse = await authAPI.getSessions()
        setSessions(sessionResponse.data.sessions)
      } catch (error: unknown) {
        setMessage({ severity: 'error', text: getErrorMessage(error, translateText('加载设置失败')) })
      } finally {
        setLoadingSessions(false)
        setLoading(false)
      }
    }

    void loadSettings()
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
    setLogoutOtherSessionsOnPasswordChange(true)
  }

  const validatePasswordForm = () => {
    const errors = {
      oldPassword: '',
      newPassword: '',
      confirmPassword: '',
    }

    let isValid = true

    if (!passwordForm.oldPassword) {
      errors.oldPassword = translateText('请输入原密码')
      isValid = false
    }

    if (!passwordForm.newPassword) {
      errors.newPassword = translateText('请输入新密码')
      isValid = false
    } else if (passwordForm.newPassword.length < 6) {
      errors.newPassword = translateText('新密码长度至少为6位')
      isValid = false
    }

    if (!passwordForm.confirmPassword) {
      errors.confirmPassword = translateText('请再次确认新密码')
      isValid = false
    } else if (passwordForm.confirmPassword !== passwordForm.newPassword) {
      errors.confirmPassword = translateText('两次输入的新密码不一致')
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
      const response = await authAPI.updatePassword({
        current_password: passwordForm.oldPassword,
        new_password: passwordForm.newPassword,
        confirm_password: passwordForm.confirmPassword,
        logout_other_sessions: logoutOtherSessionsOnPasswordChange,
      })

      await reloadSessions()
      setMessage({
        severity: 'success',
        text: response.data.revoked_other_sessions > 0
          ? t('password.changedWithRevoked', { count: response.data.revoked_other_sessions })
          : translateText('密码修改成功'),
      })
      resetPasswordDialog()
    } catch (error: unknown) {
      const errorMessage = getErrorMessage(error, translateText('密码修改失败，请稍后重试'))
      setPasswordErrors((previous) => ({
        ...previous,
        oldPassword: errorMessage,
      }))
      setMessage({ severity: 'error', text: errorMessage })
    } finally {
      setChangingPassword(false)
    }
  }

  const reloadSessions = async () => {
    const response = await authAPI.getSessions()
    setSessions(response.data.sessions)
  }

  const handleRevokeSession = async (sessionItem: UserSession) => {
    try {
      setRevokingSessionId(sessionItem.id)
      if (sessionItem.current) {
        await logout()
        void navigate('/login', {
          replace: true,
          state: { message: translateText('当前会话已退出') },
        })
        return
      }

      await authAPI.revokeSession(sessionItem.id)
      await reloadSessions()
      setMessage({ severity: 'success', text: translateText('已移除该登录会话') })
    } catch (error: unknown) {
      setMessage({ severity: 'error', text: getErrorMessage(error, translateText('移除登录会话失败')) })
    } finally {
      setRevokingSessionId('')
    }
  }

  const handleRevokeOtherSessions = async () => {
    try {
      setRevokingOtherSessions(true)
      const response = await authAPI.revokeOtherSessions()
      await reloadSessions()
      setMessage({ severity: 'success', text: response.data.count > 0 ? t('sessions.otherRemovedWithCount', { count: response.data.count }) : translateText('没有其他可移除的会话') })
    } catch (error: unknown) {
      setMessage({ severity: 'error', text: getErrorMessage(error, translateText('移除其他会话失败')) })
    } finally {
      setRevokingOtherSessions(false)
    }
  }
  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          {translateText('设置')}
        </Typography>
      </Box>

      {message && (
        <Alert severity={message.severity} sx={{ mb: 3 }} onClose={() => setMessage(null)}>
          {message.text}
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                {t('settings.language.title')}
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                {t('settings.language.description')}
              </Typography>
              <FormControl fullWidth>
                <InputLabel id="language-preference-label">{t('settings.language.current')}</InputLabel>
                <Select<LanguagePreference>
                  labelId="language-preference-label"
                  label={t('settings.language.current')}
                  value={preference}
                  onChange={(event) => setPreference(event.target.value)}
                >
                  <MenuItem value="system">{t('language.system')}</MenuItem>
                  <MenuItem value="en">{t('language.en')}</MenuItem>
                  <MenuItem value="zh-CN">{t('language.zh-CN')}</MenuItem>
                </Select>
              </FormControl>
            </CardContent>
          </Card>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                {translateText('账户安全')}
              </Typography>
              <Stack spacing={2.5} sx={{ mb: 3 }}>
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    {translateText('当前账号')}
                  </Typography>
                  <Typography variant="body1">
                    {user?.username || translateText('未知用户')}
                  </Typography>
                </Box>
                <Box>
                  <Typography variant="body2" color="text.secondary">
                    {translateText('当前角色')}
                  </Typography>
                  <Typography variant="body1">
                    {translateText(getUserRoleLabel(user?.role))}
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
                  {translateText('修改密码')}
                </Button>
                <Stack direction="row" spacing={1.5}>
                  <Button
                    variant="outlined"
                    color="warning"
                    fullWidth
                    disabled={loadingSessions || revokingOtherSessions || otherVisibleSessionCount === 0}
                    onClick={() => { void handleRevokeOtherSessions() }}
                  >
                    {revokingOtherSessions ? translateText('移除中...') : translateText('退出其他设备')}
                  </Button>
                </Stack>
                <Typography variant="body2" color="text.secondary">
                  {translateText('你可以在这里修改密码，并管理当前账户的登录会话。退出其他设备会吊销同一账户在其他浏览器或机器上的登录状态。')}
                </Typography>
              </Stack>
            </CardContent>
          </Card>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                {translateText('登录会话')}
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                {translateText('当前页面展示的是服务端登记的登录会话。移除其他会话后，对应设备会在下一次请求时失效。')}
              </Typography>
              <Stack spacing={2}>
                {visibleSessions.length === 0 && !loadingSessions && (
                  <Alert severity="info">{translateText('当前没有可展示的登录会话。')}</Alert>
                )}
                {loadingSessions && (
                  <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
                    <CircularProgress size={24} />
                  </Box>
                )}
                {visibleSessions.map((sessionItem) => (
                  <Card key={sessionItem.id} variant="outlined">
                    <CardContent>
                      {(() => {
                        const presentation = formatSessionPresentation(sessionItem)
                        const disabledAction = presentation.status.label === '已过期'
                        return (
                          <Stack spacing={1.5}>
                            <Stack direction="row" spacing={1.5} justifyContent="space-between" alignItems="flex-start">
                              <Box sx={{ flex: 1 }}>
                                <Typography variant="subtitle1">
                                  {presentation.title}
                                </Typography>
                                <Typography variant="body2" color="text.secondary">
                                  {translateText(presentation.subtitle)}
                                </Typography>
                              </Box>
                              <Alert severity={presentation.status.severity} sx={{ py: 0 }}>
                                {translateText(presentation.status.label)}
                              </Alert>
                            </Stack>
                            {hasDisplayableSessionTime(sessionItem.lastSeenAt) && (
                              <Typography variant="body2" color="text.secondary">
                                {translateText('最近活跃：')}{formatSessionTime(sessionItem.lastSeenAt)}
                              </Typography>
                            )}
                            {hasDisplayableSessionTime(sessionItem.createdAt) && (
                              <Typography variant="body2" color="text.secondary">
                                {translateText('登录时间：')}{formatSessionTime(sessionItem.createdAt)}
                              </Typography>
                            )}
                            {hasDisplayableSessionTime(sessionItem.expiresAt) && (
                              <Typography variant="body2" color="text.secondary">
                                {translateText('到期时间：')}{formatSessionTime(sessionItem.expiresAt)}
                              </Typography>
                            )}
                            {presentation.details.map((detail) => (
                              <Typography key={detail} variant="body2" color="text.secondary">
                                {translateText(detail)}
                              </Typography>
                            ))}
                        <Stack direction="row" spacing={1.5}>
                          <Button
                            variant="outlined"
                                color={sessionItem.current ? 'warning' : 'error'}
                                disabled={disabledAction || revokingSessionId === sessionItem.id}
                            onClick={() => { void handleRevokeSession(sessionItem) }}
                          >
                                {revokingSessionId === sessionItem.id ? translateText('处理中...') : sessionItem.current ? translateText('退出当前会话') : translateText('移除此会话')}
                          </Button>
                        </Stack>
                          </Stack>
                        )
                      })()}
                    </CardContent>
                  </Card>
                ))}
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
        <DialogTitle>{translateText('修改密码')}</DialogTitle>
        <DialogContent>
          <Box sx={{ mt: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
            <TextField
              label={translateText('原密码')}
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
              label={translateText('新密码')}
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
              helperText={passwordErrors.newPassword || translateText('密码长度至少 6 位')}
              disabled={changingPassword}
            />
            <TextField
              label={translateText('再次确认新密码')}
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
            <FormControlLabel
              control={(
                <Switch
                  checked={logoutOtherSessionsOnPasswordChange}
                  onChange={(event) => setLogoutOtherSessionsOnPasswordChange(event.target.checked)}
                  disabled={changingPassword}
                />
              )}
              label={translateText('修改密码后同时退出其他设备')}
            />
            <Typography variant="body2" color="text.secondary">
              {translateText('建议开启。保存后会保留当前会话，并吊销当前账户在其他浏览器或机器上的登录状态。')}
            </Typography>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={resetPasswordDialog} disabled={changingPassword}>{translateText('取消')}</Button>
          <Button
            variant="contained"
            onClick={() => { void handleChangePassword() }}
            disabled={changingPassword}
          >
            {changingPassword ? translateText('修改中...') : translateText('确认修改')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}

export default Settings
