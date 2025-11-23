/**
 * 登录页面组件
 * 提供用户登录功能，包括表单验证、错误处理和成功后的重定向
 */
import React, { useState } from 'react';
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
  Grid
} from '@mui/material';
import { LockOutlined } from '@mui/icons-material';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../services/auth.jsx';

/**
 * 登录页面组件
 * 
 * @component
 * @returns {React.ReactNode}
 */
function Login() {
  // 表单状态管理
  const [formData, setFormData] = useState({
    username: '',
    password: ''
  });
  
  // 验证错误状态
  const [validationErrors, setValidationErrors] = useState({});
  
  // UI状态管理
  const [isLoading, setIsLoading] = useState(false);
  const [loginError, setLoginError] = useState('');
  const [rememberMe, setRememberMe] = useState(false);
  
  // 路由和认证钩子
  const navigate = useNavigate();
  const { login } = useAuth() || {};

  /**
   * 处理表单输入变化
   * 
   * @param {React.ChangeEvent<HTMLInputElement>} e - 输入变化事件
   * @returns {void}
   */
  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    
    // 清除对应字段的验证错误
    if (validationErrors[name]) {
      setValidationErrors(prev => ({ ...prev, [name]: '' }));
    }
  };

  /**
   * 验证表单数据
   * 
   * @returns {boolean} 表单数据是否有效
   */
  const validateForm = () => {
    const newErrors = {};
    
    if (!formData.username.trim()) {
      newErrors.username = '用户名不能为空';
    }
    
    if (!formData.password) {
      newErrors.password = '密码不能为空';
    }
    
    setValidationErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  /**
   * 处理登录表单提交
   * 
   * @param {React.FormEvent} e - 表单提交事件
   * @returns {Promise<void>}
   */
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }
    
    setIsLoading(true);
    setLoginError('');
    
    try {
      // 模拟API调用延迟
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // 模拟登录验证（在实际应用中应替换为真实API调用）
      if (formData.username === 'admin' && formData.password === 'password') {
        // 创建模拟的用户数据
        const userData = {
          id: '1',
          username: 'admin',
          email: 'admin@example.com',
          role: 'admin',
          createdAt: new Date().toISOString(),
          lastLogin: new Date().toISOString()
        };
        
        // 创建模拟的认证令牌
        const authToken = 'mock_jwt_token_' + Date.now();
        
        // 根据rememberMe选项保存到对应的存储
        const storage = rememberMe ? localStorage : sessionStorage;
        storage.setItem('user', JSON.stringify(userData));
        storage.setItem('token', authToken);
        
        // 如果存在登录函数，调用它更新认证上下文
        if (typeof login === 'function') {
          login(userData, authToken);
        }
        
        // 登录成功，重定向到仪表盘
        navigate('/dashboard');
      } else {
        setLoginError('用户名或密码错误');
      }
    } catch (error) {
      console.error('登录错误:', error);
      setLoginError('登录失败，请稍后重试');
    } finally {
      setIsLoading(false);
    }
  };

  // 样式定义
  const paperStyle = {
    padding: 20,
    height: 'auto',
    maxWidth: 400,
    margin: '20px auto'
  };

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
      <Box 
        sx={{ 
          display: 'flex', 
          flexDirection: 'column', 
          alignItems: 'center',
          minHeight: '100vh',
          justifyContent: 'center',
          py: 8
        }}
      >
        {/* 登录表单卡片 */}
        <Paper elevation={3} style={paperStyle}>
          {/* 登录图标容器 */}
          <Box sx={avatarStyle}>
            <LockOutlined fontSize="large" />
          </Box>
          
          {/* 标题 */}
          <Typography variant="h5" component="h1" gutterBottom align="center">
            登录到 Tairitsu
          </Typography>
          
          {/* 登录错误提示 */}
          {loginError && (
            <Alert severity="error" sx={{ mb: 3 }}>
              {loginError}
            </Alert>
          )}
          
          {/* 登录表单 */}
          <Box component="form" onSubmit={handleSubmit} sx={{ mt: 1 }}>
            {/* 用户名输入框 */}
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
              onChange={handleInputChange}
              error={!!validationErrors.username}
              helperText={validationErrors.username}
              disabled={isLoading}
              className="form-input"
            />
            
            {/* 密码输入框 */}
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
              onChange={handleInputChange}
              error={!!validationErrors.password}
              helperText={validationErrors.password}
              disabled={isLoading}
              className="form-input"
            />
            
            {/* 记住我选项 */}
            <FormControlLabel
              control={
                <Checkbox
                  checked={rememberMe}
                  onChange={(e) => setRememberMe(e.target.checked)}
                  color="primary"
                  disabled={isLoading}
                />
              }
              label="记住我"
            />
            
            {/* 登录按钮 */}
            <Button
              type="submit"
              fullWidth
              variant="contained"
              sx={{ mt: 3, mb: 2 }}
              disabled={isLoading}
              className="animated-button"
            >
              {isLoading ? (
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 1 }}>
                  <CircularProgress size={20} />
                  登录中...
                </Box>
              ) : (
                '登录'
              )}
            </Button>
            
            {/* 其他操作链接 */}
            <Grid container>
              <Grid item xs>
                <Link to="/forgot-password" variant="body2" sx={{ textDecoration: 'none' }}>
                  忘记密码?
                </Link>
              </Grid>
              <Grid item>
                <Link to="/register" variant="body2" sx={{ textDecoration: 'none' }}>
                  没有账户? 注册
                </Link>
              </Grid>
            </Grid>
          </Box>
        </Paper>
        
        {/* 版权信息 */}
        <Typography variant="body2" color="text.secondary" align="center" sx={{ mt: 4 }}>
          © {new Date().getFullYear()} Tairitsu P2P 网络管理系统
        </Typography>
      </Box>
    </Container>
  );
}

export default Login;