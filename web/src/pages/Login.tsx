import { useState } from 'react';
import { 
  Box, 
  Typography, 
  Paper, 
  TextField, 
  Button, 
  FormControlLabel, 
  Checkbox, 
  Alert,
  CircularProgress,
  Container,
  Grid} from '@mui/material';
import { LockOutlined } from '@mui/icons-material';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../services/auth';
import { authAPI } from '../services/api';

/**
 * Login Component
 * Renders login form with username, password fields and login functionality
 */
function Login() {
  // Form data type definition
  interface FormData {
    username: string;
    password: string;
  }

  // Form state for username and password
  const [formData, setFormData] = useState<FormData>({
    username: '',
    password: ''
  });
  // Validation errors state
  const [errors, setErrors] = useState<Partial<FormData>>({});
  // Loading state for submit button
  const [loading, setLoading] = useState<boolean>(false);
  // Global login error message state
  const [loginError, setLoginError] = useState<string>('');
  // Remember me checkbox state
  const [rememberMe, setRememberMe] = useState<boolean>(false);
  
  // Navigation hook for redirect after login
  const navigate = useNavigate();
  // Auth service hook for login functionality
  const { login } = useAuth();

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
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Handle login submission
  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }
    
    setLoading(true);
    setLoginError('');
    
    try {
      // Call backend login API
      const response = await authAPI.login({
        username: formData.username,
        password: formData.password
      });
      
      // Extract user data and token from response
      const { user, token } = response.data;
      
      // Save to localStorage or sessionStorage
      if (rememberMe) {
        localStorage.setItem('user', JSON.stringify(user));
        localStorage.setItem('token', token);
      } else {
        sessionStorage.setItem('user', JSON.stringify(user));
        sessionStorage.setItem('token', token);
      }
      
      // Call login function from auth context
      login(user, token);
      
      // Login successful, redirect to dashboard
      navigate('/dashboard');
    } catch (error: any) {
      console.error('登录错误:', error);
      if (error.response && error.response.data && error.response.data.error) {
        setLoginError(error.response.data.error);
      } else if (error.response && error.response.status === 401) {
        setLoginError('用户名或密码错误');
      } else {
        setLoginError('登录失败，请稍后重试');
      }
    } finally {
      setLoading(false);
    }
  };

  // Paper component styling for login form container
  const paperStyle = {
    padding: 20,
    height: 'auto',
    maxWidth: 400,
    margin: '20px auto'
  };

  // Styling for the login lock icon avatar
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

  // Render the login form UI
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
            <LockOutlined fontSize="large" />
          </Box>
          
          <Typography variant="h5" component="h1" gutterBottom align="center">
            登录到 Tairitsu
          </Typography>
          
          {loginError && (
            <Alert severity="error" sx={{ mb: 3 }}>
              {loginError}
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
              autoComplete="current-password"
              value={formData.password}
              onChange={handleChange}
              error={!!errors.password}
              helperText={errors.password}
              disabled={loading}
            />
            
            <FormControlLabel
              control={
                <Checkbox
                  checked={rememberMe}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => setRememberMe(e.target.checked)}
                  color="primary"
                  disabled={loading}
                />
              }
              label="记住我"
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
                  登录中...
                </Box>
              ) : (
                '登录'
              )}
            </Button>
            
            <Grid container spacing={2}>
              <Grid size={6}>
                <Button 
                  component={Link} 
                  to="/forgot-password" 
                  variant="text" 
                  fullWidth
                  sx={{ 
                    justifyContent: 'flex-start',
                    textTransform: 'none',
                    fontWeight: 'normal'
                  }}
                >
                  忘记密码?
                </Button>
              </Grid>
              <Grid size={6}>
                <Button 
                  component={Link} 
                  to="/register" 
                  variant="text" 
                  fullWidth
                  sx={{ 
                    justifyContent: 'flex-end',
                    textTransform: 'none',
                    fontWeight: 'normal'
                  }}
                >
                  没有账户? 注册
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

export default Login;