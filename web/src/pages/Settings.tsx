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
  InputLabel,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField
} from '@mui/material';
import { Link } from 'react-router-dom';
import { authAPI, User } from '../services/api';

function Settings() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [infoMessage, setInfoMessage] = useState<string>('');
  const [language, setLanguage] = useState<string>('zh-CN');
  
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

  // 获取当前用户信息
  useEffect(() => {
    const fetchUserInfo = async () => {
      try {
        const response = await authAPI.getProfile();
        setUser(response.data);
      } catch (err) {
        console.error('获取用户信息失败:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchUserInfo();
  }, []);

  // 显示开发中提示
  const showDevelopmentMessage = () => {
    setInfoMessage('功能正在开发中');
    setTimeout(() => {
      setInfoMessage('');
    }, 3000);
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
        oldPassword: passwordForm.oldPassword,
        newPassword: passwordForm.newPassword
      });
      
      // 修改成功
      setInfoMessage('密码修改成功');
      setOpenChangePasswordDialog(false);
      
      // 重置表单
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
    } catch (error) {
      // 修改失败
      console.error('修改密码失败:', error);
      setPasswordErrors(prev => ({
        ...prev,
        oldPassword: '原密码错误或修改失败'
      }));
    } finally {
      setChangingPassword(false);
    }
  };

  // 处理语言变化
  const handleLanguageChange = (event: React.ChangeEvent<{ value: unknown }>) => {
    setLanguage(event.target.value as string);
    showDevelopmentMessage();
  };

  if (loading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body1">加载中...</Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        设置
      </Typography>

      {infoMessage && (
        <Alert severity="info" sx={{ mb: 3 }}>
          {infoMessage}
        </Alert>
      )}

      {/* 个人设置 */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            个人设置
          </Typography>

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
              >
                <MenuItem value="zh-CN">简体中文</MenuItem>
                <MenuItem value="en">English</MenuItem>
              </Select>
            </FormControl>
          </Box>

          <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
            <Button
              variant="contained"
              onClick={() => setOpenChangePasswordDialog(true)}
              fullWidth
            >
              修改密码
            </Button>
          </Box>

          <Box>
            <Button
              variant="contained"
              color="error"
              onClick={showDevelopmentMessage}
              fullWidth
            >
              注销账号
            </Button>
          </Box>
        </CardContent>
      </Card>

      {/* 管理员设置 - 仅对管理员显示 */}
      {user?.role === 'admin' && (
        <Card sx={{ mb: 3 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              管理员设置
            </Typography>

            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <Button
                variant="contained"
                onClick={showDevelopmentMessage}
                fullWidth
              >
                导入已有 ZeroTier 网络
              </Button>

              <Button
        variant="contained"
        component={Link}
        to="/settings/user-management"
        fullWidth
      >
        用户管理
      </Button>

              <Button
                variant="contained"
                onClick={showDevelopmentMessage}
                fullWidth
              >
                生成 planet 文件
              </Button>
            </Box>
          </CardContent>
        </Card>
      )}

      {/* 修改密码对话框 */}
      <Dialog
        open={openChangePasswordDialog}
        onClose={() => setOpenChangePasswordDialog(false)}
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
              onChange={(e) => setPasswordForm({ ...passwordForm, oldPassword: e.target.value })}
              error={!!passwordErrors.oldPassword}
              helperText={passwordErrors.oldPassword}
            />
            <TextField
              label="新密码"
              type="password"
              fullWidth
              value={passwordForm.newPassword}
              onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })}
              error={!!passwordErrors.newPassword}
              helperText={passwordErrors.newPassword}
            />
            <TextField
              label="再次确认新密码"
              type="password"
              fullWidth
              value={passwordForm.confirmPassword}
              onChange={(e) => setPasswordForm({ ...passwordForm, confirmPassword: e.target.value })}
              error={!!passwordErrors.confirmPassword}
              helperText={passwordErrors.confirmPassword}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenChangePasswordDialog(false)}>取消</Button>
          <Button 
            variant="contained" 
            onClick={() => handleChangePassword()}
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