import { useState } from 'react';
import { 
  Box, 
  Typography, 
  Paper, 
  TextField, 
  Button, 
  Alert,
  CircularProgress,
  Container,
  Grid} from '@mui/material';
import { PersonAddOutlined } from '@mui/icons-material';
import { Link, useNavigate } from 'react-router-dom';
import { authAPI } from '../services/api';

// 表单数据类型定义
interface FormData {
  username: string;
  password: string;
  confirmPassword: string;
}

function Register() {
  // Form state for username, password and confirm password
  const [formData, setFormData] = useState<FormData>({
    username: '',
    password: '',
    confirmPassword: ''
  });
  // Validation errors state
  const [errors, setErrors] = useState<Partial<FormData>>({});
  // Global error message state
  const [registerError, setRegisterError] = useState<string>('');
  // Success message state
  const [registerSuccess, setRegisterSuccess] = useState<string>('');
  // Loading state for submit button
  const [loading, setLoading] = useState<boolean>(false);
  
  const navigate = useNavigate();

  // Handle input change
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    // Clear error for the corresponding field
    if (errors[name as keyof FormData]) {
      setErrors(prev => ({ ...prev, [name]: '' }));
    }
  };

  // Form validation
  const validateForm = (): boolean => {
    const newErrors: Partial<FormData> = {};
    
    if (!formData.username.trim()) {
      newErrors.username = '用户名不能为空';
    }
    
    if (!formData.password) {
      newErrors.password = '密码不能为空';
    } else if (formData.password.length < 6) {
      newErrors.password = '密码长度不能少于6位';
    }
    
    if (!formData.confirmPassword) {
      newErrors.confirmPassword = '请确认密码';
    } else if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = '两次输入的密码不一致';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setRegisterError('');
    setRegisterSuccess('');

    // Validate form
    if (!validateForm()) {
      return;
    }

    setLoading(true);
    try {
      await authAPI.register({
        username: formData.username,
        password: formData.password
      });
      setRegisterSuccess('注册成功，请登录');
      // 3秒后跳转到登录页
      setTimeout(() => {
        navigate('/login');
      }, 3000);
    } catch (err: any) {
      console.error('注册错误:', err);
      if (err.response && err.response.data && err.response.data.error) {
        setRegisterError(err.response.data.error);
      } else {
        setRegisterError('注册失败，请稍后重试');
      }
    } finally {
      setLoading(false);
    }
  };

  // Paper component styling for register form container
  const paperStyle = {
    padding: 20,
    height: 'auto',
    maxWidth: 400,
    margin: '20px auto'
  };

  // Styling for the register avatar
  const avatarStyle = {
    backgroundColor: '#1976d2',
    width: 56,
    height: 56,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    margin: '0 auto 20px auto',
    fontSize: 24
  };

  return (
    <Container component="main" maxWidth="xs">
      <Box sx={{ 
        display: 'flex', 
        flexDirection: 'column', 
        alignItems: 'center',
        minHeight: '100vh',
        justifyContent: 'center',
        py: 8
      }}>
        <Paper elevation={3} style={paperStyle}>
          <Box sx={avatarStyle}>
            <PersonAddOutlined fontSize="large" />
          </Box>
          
          <Typography variant="h5" component="h1" gutterBottom align="center">
            注册到 Tairitsu
          </Typography>
          
          {registerError && (
            <Alert severity="error" sx={{ mb: 3 }}>
              {registerError}
            </Alert>
          )}
          
          {registerSuccess && (
            <Alert severity="success" sx={{ mb: 3 }}>
              {registerSuccess}
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
              error={!!errors.username}
              helperText={errors.username}
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
              error={!!errors.password}
              helperText={errors.password}
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
              error={!!errors.confirmPassword}
              helperText={errors.confirmPassword}
              disabled={loading}
            />
            <Button
              type="submit"
              fullWidth
              variant="contained"
              sx={{ mt: 3, mb: 2 }}
              disabled={loading}
            >
              {loading ? (
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 1 }}>
                  <CircularProgress size={20} />
                  注册中...
                </Box>
              ) : (
                '注册'
              )}
            </Button>
            
            <Grid container spacing={2}>
              <Grid size={6}>
                <Button 
                  component={Link} 
                  to="/login" 
                  variant="text" 
                  fullWidth
                  sx={{ 
                    justifyContent: 'flex-start',
                    textTransform: 'none',
                    fontWeight: 'normal'
                  }}
                >
                  已有账户? 登录
                </Button>
              </Grid>
              <Grid size={6}>
                <Button 
                  component={Link} 
                  to="/" 
                  variant="text" 
                  fullWidth
                  sx={{ 
                    justifyContent: 'flex-end',
                    textTransform: 'none',
                    fontWeight: 'normal'
                  }}
                >
                  返回首页
                </Button>
              </Grid>
            </Grid>
          </Box>
        </Paper>
        
        <Typography variant="body2" color="text.secondary" align="center" sx={{ mt: 4 }}>
            © {new Date().getFullYear()} Tairitsu
          </Typography>
        </Box>
      </Container>
  );
}

export default Register;