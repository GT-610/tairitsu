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
  InputLabel
} from '@mui/material';
import { authAPI, User } from '../services/api';

function Settings() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [infoMessage, setInfoMessage] = useState<string>('');
  const [language, setLanguage] = useState<string>('zh-CN');

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
              onClick={showDevelopmentMessage}
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
                onClick={showDevelopmentMessage}
                fullWidth
              >
                对用户赋予/撤销管理员权限
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
    </Box>
  );
}

export default Settings;