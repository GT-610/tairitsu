import { useState } from 'react';
import { Box, TextField, Button, Typography, Container, Paper, Alert, CircularProgress }
from '@mui/material';
import { Link, useNavigate } from 'react-router-dom';
import { authAPI } from '../services/api';

// 表单数据类型定义
interface FormData {
  username: string;
  password: string;
  confirmPassword: string;
}

function Register() {
  const [formData, setFormData] = useState<FormData>({
    username: '',
    password: '',
    confirmPassword: ''
  });
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const navigate = useNavigate();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    // 验证密码
    if (formData.password !== formData.confirmPassword) {
      setError('两次输入的密码不一致');
      return;
    }

    setLoading(true);
    try {
      await authAPI.register({
        username: formData.username,
        password: formData.password
      });
      setSuccess('注册成功，请登录');
      // 3秒后跳转到登录页
      setTimeout(() => {
        navigate('/login');
      }, 3000);
    } catch (err: any) {
      setError(err.response?.data?.message || '注册失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container component="main" maxWidth="xs" sx={{ mt: 8 }}>
      <Paper elevation={3} sx={{ p: 4, borderRadius: 2 }}>
        <Typography component="h1" variant="h5" align="center" sx={{ mb: 4 }}>
          用户注册
        </Typography>
        
        {error && (
          <Alert severity="error" sx={{ mb: 3 }}>
            {error}
          </Alert>
        )}
        
        {success && (
          <Alert severity="success" sx={{ mb: 3 }}>
            {success}
          </Alert>
        )}
        
        <Box component="form" onSubmit={handleSubmit} sx={{ mt: 1 }}>
          <TextField
            margin="normal"
            required
            fullWidth
            id="username"
            label="用户名"
            name="username"
            autoComplete="username"
            autoFocus
            value={formData.username}
            onChange={handleChange}
            disabled={loading}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            name="password"
            label="密码"
            type="password"
            id="password"
            autoComplete="new-password"
            value={formData.password}
            onChange={handleChange}
            disabled={loading}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            name="confirmPassword"
            label="确认密码"
            type="password"
            id="confirmPassword"
            value={formData.confirmPassword}
            onChange={handleChange}
            disabled={loading}
          />
          <Button
            type="submit"
            fullWidth
            variant="contained"
            sx={{ mt: 3, mb: 2 }}
            disabled={loading}
          >
            {loading ? 
              <CircularProgress size={24} color="inherit" /> : 
              '注册'
            }
          </Button>
          <Box sx={{ textAlign: 'center' }}>
            <Typography variant="body2">
              已有账号？ 
              <Link to="/login" style={{ color: '#64b5f6', textDecoration: 'none' }}>
                立即登录
              </Link>
            </Typography>
          </Box>
        </Box>
      </Paper>
    </Container>
  );
}

export default Register;