import { useEffect, useMemo, useState } from 'react'
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
import { authAPI, systemAPI, userAPI, type RuntimeSettings, type User, type UserSession } from '../services/api'
import { getErrorMessage } from '../services/errors'
import { useAuth } from '../services/auth'
import { useNavigate } from 'react-router-dom'
import { formatSessionPresentation } from '../utils/sessionPresentation'

function Settings() {
  const navigate = useNavigate()
  const { user, refreshUser, logout } = useAuth()
  const isAdmin = user?.role === 'admin'

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

  const [runtimeSettings, setRuntimeSettings] = useState<RuntimeSettings>({ allow_public_registration: true })
  const [initialRuntimeSettings, setInitialRuntimeSettings] = useState<RuntimeSettings>({ allow_public_registration: true })
  const [savingRuntimeSettings, setSavingRuntimeSettings] = useState(false)

  const [users, setUsers] = useState<User[]>([])
  const [targetAdminId, setTargetAdminId] = useState('')
  const [transferringAdmin, setTransferringAdmin] = useState(false)

  useEffect(() => {
    const loadSettings = async () => {
      try {
        setLoading(true)
        setLoadingSessions(true)

        if (isAdmin) {
          const [sessionResponse, settingsResponse, usersResponse] = await Promise.all([
            authAPI.getSessions(),
            systemAPI.getRuntimeSettings(),
            userAPI.getAllUsers(),
          ])

          setSessions(sessionResponse.data.sessions)
          setRuntimeSettings(settingsResponse.data)
          setInitialRuntimeSettings(settingsResponse.data)
          setUsers(usersResponse.data)
        } else {
          const sessionResponse = await authAPI.getSessions()
          setSessions(sessionResponse.data.sessions)
        }
      } catch (error: unknown) {
        setMessage({ severity: 'error', text: getErrorMessage(error, '加载设置失败') })
      } finally {
        setLoadingSessions(false)
        setLoading(false)
      }
    }

    void loadSettings()
  }, [isAdmin])

  const transferCandidates = useMemo(
    () => users.filter((candidate) => candidate.id !== user?.id && candidate.role !== 'admin'),
    [user?.id, users],
  )

  const runtimeSettingsUnsaved = runtimeSettings.allow_public_registration !== initialRuntimeSettings.allow_public_registration

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
          ? `密码修改成功，并已移除其他会话 ${response.data.revoked_other_sessions} 个`
          : '密码修改成功',
      })
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

  const handleSaveRuntimeSettings = async () => {
    try {
      setSavingRuntimeSettings(true)
      const response = await systemAPI.updateRuntimeSettings(runtimeSettings)
      setRuntimeSettings(response.data.settings)
      setInitialRuntimeSettings(response.data.settings)
      setMessage({ severity: 'success', text: '实例治理设置已保存' })
    } catch (error: unknown) {
      setMessage({ severity: 'error', text: getErrorMessage(error, '保存实例治理设置失败') })
    } finally {
      setSavingRuntimeSettings(false)
    }
  }

  const handleTransferAdmin = async () => {
    if (!targetAdminId) {
      setMessage({ severity: 'error', text: '请选择新的管理员' })
      return
    }

    try {
      setTransferringAdmin(true)
      const response = await userAPI.transferAdmin(targetAdminId)
      const nextAdmin = response.data.user
      const refreshedProfile = await authAPI.getProfile()
      refreshUser(refreshedProfile.data)
      setMessage({ severity: 'success', text: `管理员身份已转让给 ${nextAdmin.username}` })
      void navigate('/networks', {
        replace: true,
        state: { message: `管理员身份已转让给 ${nextAdmin.username}` },
      })
    } catch (error: unknown) {
      setMessage({ severity: 'error', text: getErrorMessage(error, '转让管理员身份失败') })
    } finally {
      setTransferringAdmin(false)
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
          state: { message: '当前会话已退出' },
        })
        return
      }

      await authAPI.revokeSession(sessionItem.id)
      await reloadSessions()
      setMessage({ severity: 'success', text: '已移除该登录会话' })
    } catch (error: unknown) {
      setMessage({ severity: 'error', text: getErrorMessage(error, '移除登录会话失败') })
    } finally {
      setRevokingSessionId('')
    }
  }

  const handleRevokeOtherSessions = async () => {
    try {
      setRevokingOtherSessions(true)
      const response = await authAPI.revokeOtherSessions()
      await reloadSessions()
      setMessage({ severity: 'success', text: response.data.count > 0 ? `已移除其他会话 ${response.data.count} 个` : '没有其他可移除的会话' })
    } catch (error: unknown) {
      setMessage({ severity: 'error', text: getErrorMessage(error, '移除其他会话失败') })
    } finally {
      setRevokingOtherSessions(false)
    }
  }

  const formatSessionTime = (value: string) => new Date(value).toLocaleString()

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

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                账户安全
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
                    {isAdmin ? '管理员' : '普通用户'}
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
                <Stack direction="row" spacing={1.5}>
                  <Button
                    variant="outlined"
                    color="warning"
                    fullWidth
                    disabled={loadingSessions || revokingOtherSessions || sessions.filter((sessionItem) => !sessionItem.current && !sessionItem.revokedAt).length === 0}
                    onClick={() => { void handleRevokeOtherSessions() }}
                  >
                    {revokingOtherSessions ? '移除中...' : '退出其他设备'}
                  </Button>
                </Stack>
                <Typography variant="body2" color="text.secondary">
                  你可以在这里修改密码，并管理当前账户的登录会话。退出其他设备会吊销同一账户在其他浏览器或机器上的登录状态。
                </Typography>
              </Stack>
            </CardContent>
          </Card>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                登录会话
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                当前页面展示的是服务端登记的登录会话。移除其他会话后，对应设备会在下一次请求时失效。
              </Typography>
              <Stack spacing={2}>
                {sessions.length === 0 && !loadingSessions && (
                  <Alert severity="info">当前没有可展示的登录会话。</Alert>
                )}
                {loadingSessions && (
                  <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
                    <CircularProgress size={24} />
                  </Box>
                )}
                {sessions.map((sessionItem) => (
                  <Card key={sessionItem.id} variant="outlined">
                    <CardContent>
                      {(() => {
                        const presentation = formatSessionPresentation(sessionItem)
                        const disabledAction = Boolean(sessionItem.revokedAt) || presentation.status.label === '已过期'
                        return (
                          <Stack spacing={1.5}>
                            <Stack direction="row" spacing={1.5} justifyContent="space-between" alignItems="flex-start">
                              <Box sx={{ flex: 1 }}>
                                <Typography variant="subtitle1">
                                  {presentation.title}
                                </Typography>
                                <Typography variant="body2" color="text.secondary">
                                  {presentation.subtitle}
                                </Typography>
                              </Box>
                              <Alert severity={presentation.status.severity} sx={{ py: 0 }}>
                                {presentation.status.label}
                              </Alert>
                            </Stack>
                        <Typography variant="body2" color="text.secondary">
                          最近活跃：{formatSessionTime(sessionItem.lastSeenAt)}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          登录时间：{formatSessionTime(sessionItem.createdAt)}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          到期时间：{formatSessionTime(sessionItem.expiresAt)}
                        </Typography>
                            {sessionItem.revokedAt && (
                              <Typography variant="body2" color="text.secondary">
                                移除时间：{formatSessionTime(sessionItem.revokedAt)}
                              </Typography>
                            )}
                            {presentation.details.map((detail) => (
                              <Typography key={detail} variant="body2" color="text.secondary">
                                {detail}
                              </Typography>
                            ))}
                        <Stack direction="row" spacing={1.5}>
                          <Button
                            variant="outlined"
                                color={sessionItem.current ? 'warning' : 'error'}
                                disabled={disabledAction || revokingSessionId === sessionItem.id}
                            onClick={() => { void handleRevokeSession(sessionItem) }}
                          >
                                {revokingSessionId === sessionItem.id ? '处理中...' : sessionItem.current ? '退出当前会话' : '移除此会话'}
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

          {isAdmin ? (
            <>
              <Card sx={{ mb: 3 }}>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    实例治理
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    这里控制公开发布最关键的实例级边界。当前支持配置公开注册策略，后续更完整的实例治理项会继续收敛到本区块。
                  </Typography>

                  <Stack spacing={2.5}>
                    <FormControlLabel
                      control={(
                        <Switch
                          checked={runtimeSettings.allow_public_registration}
                          onChange={(event) => setRuntimeSettings((previous) => ({
                            ...previous,
                            allow_public_registration: event.target.checked,
                          }))}
                        />
                      )}
                      label="允许公开注册"
                    />
                    <Typography variant="body2" color="text.secondary">
                      关闭后，未登录用户将不能继续公开创建账号，但 setup 阶段的首个管理员创建逻辑不受影响。
                    </Typography>
                    <Stack direction="row" spacing={1.5}>
                      <Button
                        variant="outlined"
                        disabled={!runtimeSettingsUnsaved || savingRuntimeSettings}
                        onClick={() => setRuntimeSettings(initialRuntimeSettings)}
                      >
                        重置
                      </Button>
                      <Button
                        variant="contained"
                        disabled={!runtimeSettingsUnsaved || savingRuntimeSettings}
                        onClick={() => { void handleSaveRuntimeSettings() }}
                      >
                        {savingRuntimeSettings ? '保存中...' : '保存'}
                      </Button>
                    </Stack>
                  </Stack>
                </CardContent>
              </Card>

              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    管理员职责
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    当前系统保持单管理员模型。你可以在这里把管理员身份转让给某个普通用户，转让后自己会自动降为普通用户。
                  </Typography>
                  <Stack spacing={2}>
                    <Alert severity="info">当前管理员：{user?.username || '未知用户'}</Alert>
                    <FormControl fullWidth>
                      <InputLabel id="transfer-admin-label">新的管理员</InputLabel>
                      <Select
                        labelId="transfer-admin-label"
                        value={targetAdminId}
                        label="新的管理员"
                        onChange={(event) => setTargetAdminId(event.target.value)}
                      >
                        {transferCandidates.map((candidate) => (
                          <MenuItem key={candidate.id} value={candidate.id}>
                            {candidate.username}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                    {transferCandidates.length === 0 && (
                      <Alert severity="warning">
                        当前没有可接收管理员身份的普通用户。请先创建或保留至少一个普通用户账号。
                      </Alert>
                    )}
                    <Button
                      variant="contained"
                      color="warning"
                      disabled={!targetAdminId || transferringAdmin}
                      onClick={() => { void handleTransferAdmin() }}
                    >
                      {transferringAdmin ? '转让中...' : '转让管理员身份'}
                    </Button>
                  </Stack>
                </CardContent>
              </Card>
            </>
          ) : (
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  基础说明
                </Typography>
                <Stack spacing={1.5}>
                  <Alert severity="info">实例治理设置仅对管理员开放。</Alert>
                  <Alert severity="info">管理员转让仅能由当前管理员发起。</Alert>
                  <Alert severity="info">更多账户中心能力将在后续阶段补齐。</Alert>
                </Stack>
              </CardContent>
            </Card>
          )}
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
            <FormControlLabel
              control={(
                <Switch
                  checked={logoutOtherSessionsOnPasswordChange}
                  onChange={(event) => setLogoutOtherSessionsOnPasswordChange(event.target.checked)}
                  disabled={changingPassword}
                />
              )}
              label="修改密码后同时退出其他设备"
            />
            <Typography variant="body2" color="text.secondary">
              建议开启。保存后会保留当前会话，并吊销当前账户在其他浏览器或机器上的登录状态。
            </Typography>
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
