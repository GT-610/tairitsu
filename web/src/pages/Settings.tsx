import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Button,
  Alert,
  Select,
  MenuItem,
  FormControl,
  FormHelperText,
  InputLabel,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  CircularProgress,
  Stack,
  Divider,
  type SelectChangeEvent
} from '@mui/material';
import { authAPI } from '../services/api';
import { getErrorMessage } from '../services/errors';
import { useAuth } from '../services/auth';

function Settings() {
  const [loading, setLoading] = useState<boolean>(true);
  const [message, setMessage] = useState<{ severity: 'info' | 'success' | 'error'; text: string } | null>(null);
  const [language, setLanguage] = useState<string>('zh-CN');
  const { user } = useAuth();
  
  // 修改密码对话框相关状态
  const [openChangePasswordDialog, setOpenChangePasswordDialog] = useState<boolean>(false);
  const [passwordForm, setPasswordForm] = useState({
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  });
  const [passwordErrors, setPasswordErrors] = useState({
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  });
  const [changingPassword, setChangingPassword] = useState<boolean>(false);

  // 初始化加载状态
  useEffect(() => {
    // 模拟加载完成，实际应用中可能需要加载设置信息
    setLoading(false);
  }, []);

  // 显示开发中提示
  const showDevelopmentMessage = () => {
    setMessage({ severity: 'info', text: '该能力仍在继续完善中，当前版本暂未开放。' });
    setTimeout(() => {
      setMessage(null);
    }, 3000);
  };

  const resetPasswordDialog = () => {
    setOpenChangePasswordDialog(false);
    setPasswordForm({
      oldPassword: '',
      newPassword: '',
      confirmPassword: ''
    });
    setPasswordErrors({
      oldPassword: '',
      newPassword: '',
      confirmPassword: ''
    });
  };
  
  // 验证密码表单
  const validatePasswordForm = () => {
    const errors = {
      oldPassword: '',
      newPassword: '',
      confirmPassword: ''
    };
    
    let isValid = true;
    
    // 验证原密码
    if (!passwordForm.oldPassword) {
      errors.oldPassword = '请输入原密码';
      isValid = false;
    }
    
    // 验证新密码
    if (!passwordForm.newPassword) {
      errors.newPassword = '请输入新密码';
      isValid = false;
    } else if (passwordForm.newPassword.length < 6) {
      errors.newPassword = '新密码长度至少为6位';
      isValid = false;
    }
    
    // 验证确认密码
    if (!passwordForm.confirmPassword) {
      errors.confirmPassword = '请再次确认新密码';
      isValid = false;
    } else if (passwordForm.confirmPassword !== passwordForm.newPassword) {
      errors.confirmPassword = '两次输入的新密码不一致';
      isValid = false;
    }
    
    setPasswordErrors(errors);
    return isValid;
  };
  
  // 处理修改密码
  const handleChangePassword = async () => {
    if (!validatePasswordForm()) {
      return;
    }
    
    try {
      setChangingPassword(true);
      await authAPI.updatePassword({
        current_password: passwordForm.oldPassword,
        new_password: passwordForm.newPassword,
        confirm_password: passwordForm.confirmPassword
      });
      
      // 修改成功
      setMessage({ severity: 'success', text: '密码修改成功' });
      resetPasswordDialog();
    } catch (error: unknown) {
      const errorMessage = getErrorMessage(error, '密码修改失败，请稍后重试');
      setPasswordErrors(prev => ({
        ...prev,
        oldPassword: errorMessage
      }));
      setMessage({ severity: 'error', text: errorMessage });
    } finally {
      setChangingPassword(false);
    }
  };

  // 处理语言变化
  const handleLanguageChange = (event: SelectChangeEvent<string>, _child: React.ReactNode) => {
    setLanguage(event.target.value);
    showDevelopmentMessage();
  };

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
        设置页正在逐步补齐中。当前可用能力以密码修改为主，语言切换与账号注销等入口会在后续继续完善。
      </Alert>

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          {/* 个人设置 */}
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

              <Box sx={{ mb: 3 }}>
                <Typography variant="body1" sx={{ mb: 1 }}>
                  语言
                </Typography>
                <FormControl fullWidth>
                  <InputLabel id="language-select-label">选择语言</InputLabel>
                  <Select
                    labelId="language-select-label"
                    value={language}
                    label="选择语言"
                    onChange={handleLanguageChange}
                    disabled
                  >
                    <MenuItem value="zh-CN">简体中文</MenuItem>
                    <MenuItem value="en">English</MenuItem>
                  </Select>
                  <FormHelperText>语言切换功能尚未开放</FormHelperText>
                </FormControl>
              </Box>

              <Stack spacing={2} sx={{ mb: 2 }}>
                <Button
                  variant="contained"
                  onClick={() => setOpenChangePasswordDialog(true)}
                  fullWidth
                >
                  修改密码
                </Button>
                <Typography variant="body2" color="text.secondary">
                  修改密码后，当前登录状态会继续保持，但建议您在其他设备上重新验证新密码是否生效。
                </Typography>
              </Stack>

              <Box>
                <Button
                  variant="outlined"
                  color="inherit"
                  onClick={showDevelopmentMessage}
                  fullWidth
                >
                  注销账号（即将支持）
                </Button>
              </Box>
            </CardContent>
          </Card>

          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                会话与安全
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                更多账户安全能力仍在开发中。后续会补充会话管理、登录设备查看和更完整的账号操作入口。
              </Typography>
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
                <Button variant="outlined" disabled fullWidth>
                  会话管理（开发中）
                </Button>
                <Button variant="outlined" disabled fullWidth>
                  登录设备（开发中）
                </Button>
              </Stack>
            </CardContent>
          </Card>
        </>
      )}

      {/* 修改密码对话框 */}
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
              onChange={(e) => {
                setPasswordForm({ ...passwordForm, oldPassword: e.target.value });
                if (passwordErrors.oldPassword) {
                  setPasswordErrors(prev => ({ ...prev, oldPassword: '' }));
                }
              }}
              error={!!passwordErrors.oldPassword}
              helperText={passwordErrors.oldPassword}
              disabled={changingPassword}
            />
            <TextField
              label="新密码"
              type="password"
              fullWidth
              value={passwordForm.newPassword}
              onChange={(e) => {
                setPasswordForm({ ...passwordForm, newPassword: e.target.value });
                if (passwordErrors.newPassword) {
                  setPasswordErrors(prev => ({ ...prev, newPassword: '' }));
                }
              }}
              error={!!passwordErrors.newPassword}
              helperText={passwordErrors.newPassword || '密码长度至少 6 位'}
              disabled={changingPassword}
            />
            <TextField
              label="再次确认新密码"
              type="password"
              fullWidth
              value={passwordForm.confirmPassword}
              onChange={(e) => {
                setPasswordForm({ ...passwordForm, confirmPassword: e.target.value });
                if (passwordErrors.confirmPassword) {
                  setPasswordErrors(prev => ({ ...prev, confirmPassword: '' }));
                }
              }}
              error={!!passwordErrors.confirmPassword}
              helperText={passwordErrors.confirmPassword}
              disabled={changingPassword}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={resetPasswordDialog} disabled={changingPassword}>取消</Button>
          <Button 
            variant="contained" 
            onClick={() => { void handleChangePassword(); }}
            disabled={changingPassword}
          >
            {changingPassword ? '修改中...' : '确认修改'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}

export default Settings;
