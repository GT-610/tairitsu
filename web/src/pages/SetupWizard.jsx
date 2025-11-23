import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { 
  Box, 
  Container,
  Typography, 
  Paper, 
  Stepper, 
  Step, 
  StepLabel, 
  Button, 
  CircularProgress,
  TextField
} from '@mui/material';
import ErrorAlert from '../components/ErrorAlert.jsx';
import { useNavigate } from 'react-router-dom';
import { authAPI, systemAPI } from '../services/api.js';

const SetupWizard = () => {
  const navigate = useNavigate();
  
  // 使用单个对象状态管理所有表单和UI状态
  const [state, setState] = useState({
    activeStep: 0,
    loading: false,
    error: '',
    success: '',
    adminData: {
      username: '',
      password: '',
      email: ''
    },
    dbConfig: {
      type: 'sqlite',
      path: '',
      host: '',
      port: 0,
      user: '',
      pass: '',
      name: ''
    },
    ztConfig: {
      controllerUrl: 'http://localhost:9993',
      tokenPath: '/var/lib/zerotier-one/authtoken.secret'
    },
    ztStatus: null,
    ztConnected: false
  });

  // 使用useMemo优化静态数组，避免不必要的重新创建
  const steps = useMemo(() => [
    '欢迎使用 Tairitsu',
    '配置 ZeroTier 控制器',
    '配置数据库',
    '创建管理员账户',
    '完成设置'
  ], []);

  // 在组件挂载时设置标记，表明设置向导已开始
  useEffect(() => {
    // 设置标记表示设置向导已开始
    localStorage.setItem('tairitsu_setup_started', 'true');
    
    // 组件卸载时清理标记
    return () => {
      // 只有在设置未完成时才移除标记
      if (!localStorage.getItem('tairitsu_initialized')) {
        localStorage.removeItem('tairitsu_setup_started');
      }
    };
  }, []);
  
  // 这个函数将在需要时手动调用
  const checkSystemStatus = async () => {
    try {
      const response = await systemAPI.getSetupStatus();
      if (response.data.initialized) {
        // 如果系统已初始化，重定向到登录页面
        navigate('/login');
      }
    } catch (err) {
      console.error('检查系统状态失败:', err);
      // 即使检查失败，也继续显示设置向导
    }
  };
  
  // 只有在用户点击下一步时才检查状态，避免在第一步加载时发送请求

  // 使用useCallback缓存测试ZeroTier连接函数
  const testAndInitZtConnection = useCallback(async () => {
    setState(prev => ({ 
      ...prev, 
      loading: true, 
      error: '', 
      success: '' 
    }));
    try {
      // 保存配置并同时测试连接
      const response = await systemAPI.saveZtConfig(state.ztConfig);
      // 从响应中获取ZeroTier状态信息
      setState(prev => ({
        ...prev,
        ztStatus: response.data.status,
        ztConnected: true,
        success: 'ZeroTier 连接成功！',
        loading: false
      }));
      return true;
    } catch (err) {
      setState(prev => ({
        ...prev,
        error: 'ZeroTier 连接失败: ' + (err.response?.data?.error || err.message),
        ztConnected: false,
        loading: false
      }));
      return false;
    }
  }, [state.ztConfig]);

  // 使用useCallback缓存提交处理函数
  const handleSubmit = useCallback(async () => {
    setState(prev => ({ 
      ...prev, 
      loading: true, 
      error: '', 
      success: '' 
    }));
    
    try {
      // 根据当前步骤执行不同的操作
      if (state.activeStep === 0) {
        // 欢迎页面，检查系统状态后进入下一步
        await checkSystemStatus();
        // 如果checkSystemStatus没有跳转，则继续执行
        setState(prev => ({ ...prev, activeStep: prev.activeStep + 1 }));
      } else if (state.activeStep === 1) {
        // ZeroTier 配置步骤 - 直接验证并初始化
        const success = await testAndInitZtConnection();
        if (success) {
          // 验证成功，进入下一步
          setState(prev => ({ ...prev, activeStep: prev.activeStep + 1 }));
        }
      } else if (state.activeStep === 2) {
        // 数据库配置验证
        if (!state.dbConfig.type) {
          setState(prev => ({ ...prev, error: '请选择数据库类型', loading: false }));
          return;
        }
        
        // 根据数据库类型进行不同验证
        if (state.dbConfig.type !== 'sqlite') {
          // PostgreSQL或MySQL验证
          if (!state.dbConfig.host.trim()) {
            setState(prev => ({ ...prev, error: '请输入数据库主机地址', loading: false }));
            return;
          }
          if (!state.dbConfig.port || state.dbConfig.port <= 0) {
            setState(prev => ({ ...prev, error: '请输入有效的数据库端口', loading: false }));
            return;
          }
          if (!state.dbConfig.user.trim()) {
            setState(prev => ({ ...prev, error: '请输入数据库用户名', loading: false }));
            return;
          }
          if (!state.dbConfig.name.trim()) {
            setState(prev => ({ ...prev, error: '请输入数据库名称', loading: false }));
            return;
          }
        }
        
        // 发送数据库配置到后端
        await systemAPI.configureDatabase(state.dbConfig);
        // 继续下一步
        setState(prev => ({ ...prev, activeStep: prev.activeStep + 1, loading: false }));
      } else if (state.activeStep === 3) {
        // 创建管理员账户步骤
        // 表单验证
        if (!state.adminData.username.trim()) {
          setState(prev => ({ ...prev, error: '请输入用户名', loading: false }));
          return;
        }
        if (!state.adminData.password || state.adminData.password.length < 6) {
          setState(prev => ({ ...prev, error: '密码长度至少为6位', loading: false }));
          return;
        }
        if (!state.adminData.email.trim()) {
          setState(prev => ({ ...prev, error: '请输入邮箱地址', loading: false }));
          return;
        }
        
        // 简单的邮箱格式验证
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(state.adminData.email)) {
          setState(prev => ({ ...prev, error: '请输入有效的邮箱地址', loading: false }));
          return;
        }
        
        // 创建管理员账户
        await authAPI.register(state.adminData);
        // 继续下一步
        setState(prev => ({ ...prev, activeStep: prev.activeStep + 1, loading: false }));
      } else if (state.activeStep === 4) {
        // 完成设置步骤
        // 设置系统为已初始化状态
        await systemAPI.setInitialized(true);
        // 更新localStorage标记，表示系统已初始化
        localStorage.setItem('tairitsu_initialized', 'true');
        localStorage.removeItem('tairitsu_setup_started');
        // 设置成功消息
        setState(prev => ({ ...prev, success: '系统初始化完成！即将跳转到登录页面...', loading: false }));
        // 延迟跳转到登录页面
        setTimeout(() => {
          navigate('/login');
        }, 2000);
      } else {
        // 默认情况：前进到下一步
        setState(prev => ({ ...prev, activeStep: prev.activeStep + 1, loading: false }));
      }
    } catch (err) {
      setState(prev => ({
        ...prev,
        error: '操作失败: ' + (err.response?.data?.error || err.message),
        loading: false
      }));
    }
  }, [state, checkSystemStatus, testAndInitZtConnection, navigate]);

  // 使用useCallback缓存返回处理函数
  const handleBack = useCallback(() => {
    setState(prev => ({ ...prev, activeStep: prev.activeStep - 1 }));
  }, []);

  // 使用useCallback缓存管理数据变更处理函数
  const handleAdminDataChange = useCallback((e) => {
    setState(prev => ({
      ...prev,
      adminData: {
        ...prev.adminData,
        [e.target.name]: e.target.value
      }
    }));
  }, []);

  // 使用useCallback缓存数据库配置变更处理函数
  const handleDbConfigChange = useCallback((e) => {
    setState(prev => {
      // 特殊处理数据库类型变更
      if (e.target.name === 'type') {
        const newType = e.target.value;
        return {
          ...prev,
          dbConfig: {
            type: newType,
            path: newType === 'sqlite' ? '' : prev.dbConfig.path,
            host: newType === 'sqlite' ? '' : prev.dbConfig.host,
            port: newType === 'sqlite' ? 0 : prev.dbConfig.port, // SQLite不需要端口
            user: newType === 'sqlite' ? '' : prev.dbConfig.user,
            pass: newType === 'sqlite' ? '' : prev.dbConfig.pass,
            name: newType === 'sqlite' ? '' : prev.dbConfig.name
          }
        };
      } else {
        // 处理其他字段变更
        return {
          ...prev,
          dbConfig: {
            ...prev.dbConfig,
            [e.target.name]: e.target.value
          }
        };
      }
    });
  }, []);

  // 使用useCallback缓存ZeroTier配置变更处理函数
  const handleZtConfigChange = useCallback((e) => {
    setState(prev => ({
      ...prev,
      ztConfig: {
        ...prev.ztConfig,
        [e.target.name]: e.target.value
      }
    }));
  }, []);

  // 渲染步骤内容
  const renderStepContent = (step) => {
    switch (step) {
      case 0:
        return (
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              欢迎使用 Tairitsu
            </Typography>
            <Typography variant="body1" paragraph>
              Tairitsu 是一个基于 Web 的 ZeroTier 网络管理平台，帮助您更轻松地管理和配置 ZeroTier 网络。
            </Typography>
            <Typography variant="body1" paragraph>
              在完成设置向导之前，您需要进行以下配置：
            </Typography>
            <ul>
              <li>配置 ZeroTier 控制器连接</li>
              <li>设置数据库</li>
              <li>创建管理员账户</li>
            </ul>
          </Paper>
        );
      case 1:
        return (
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              配置 ZeroTier 控制器
            </Typography>
            <Typography variant="body1" paragraph>
              请输入您的 ZeroTier 控制器信息，以便 Tairitsu 能够连接和管理您的网络。
            </Typography>
            <form onSubmit={handleSubmit}>
              <TextField
                margin="normal"
                required
                fullWidth
                id="controllerUrl"
                label="ZeroTier 控制器 URL"
                name="controllerUrl"
                autoComplete="url"
                value={ztConfig.controllerUrl}
                onChange={handleZtConfigChange}
                disabled={loading}
              />
              <TextField
                margin="normal"
                required
                fullWidth
                id="tokenPath"
                label="认证令牌文件路径"
                name="tokenPath"
                autoComplete="file-path"
                value={ztConfig.tokenPath}
                onChange={handleZtConfigChange}
                disabled={loading}
                helperText="默认为 /var/lib/zerotier-one/authtoken.secret"
              />
              {state.error && (
                <ErrorAlert 
                  message={state.error} 
                  onClose={() => setState(prev => ({ ...prev, error: '' }))} 
                />
              )}
              {state.success && (
                <ErrorAlert 
                  message={state.success} 
                  severity="success"
                  onClose={() => setState(prev => ({ ...prev, success: '' }))} 
                />
              )}
            </form>
          </Paper>
        );
      case 2:
        return (
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              配置数据库
            </Typography>
            <Typography variant="body1" paragraph>
              请选择并配置您要使用的数据库。Tairitsu 支持 SQLite、MySQL 和 PostgreSQL。
            </Typography>
            <form onSubmit={handleSubmit}>
              <TextField
                margin="normal"
                required
                fullWidth
                id="type"
                label="数据库类型"
                name="type"
                select
                SelectProps={{
                  native: true,
                }}
                value={state.dbConfig.type}
                onChange={handleDbConfigChange}
                disabled={loading}
              >
                <option value="sqlite">SQLite</option>
                <option value="mysql">MySQL</option>
                <option value="postgres">PostgreSQL</option>
              </TextField>
              
              {dbConfig.type !== 'sqlite' && (
                <>
                  <TextField
                    margin="normal"
                    required
                    fullWidth
                    id="host"
                    label="主机地址"
                    name="host"
                    autoComplete="hostname"
                    value={state.dbConfig.host}
                    onChange={handleDbConfigChange}
                    disabled={loading}
                  />
                  <TextField
                    margin="normal"
                    required
                    fullWidth
                    id="port"
                    label="端口"
                    name="port"
                    type="number"
                    value={state.dbConfig.port || ''}
                    onChange={handleDbConfigChange}
                    disabled={loading}
                    placeholder={dbConfig.type === 'mysql' ? '3306' : '5432'}
                  />
                  <TextField
                    margin="normal"
                    required
                    fullWidth
                    id="user"
                    label="用户名"
                    name="user"
                    autoComplete="username"
                    value={state.dbConfig.user}
                    onChange={handleDbConfigChange}
                    disabled={loading}
                  />
                  <TextField
                    margin="normal"
                    required
                    fullWidth
                    id="pass"
                    label="密码"
                    name="pass"
                    type="password"
                    autoComplete="current-password"
                    value={state.dbConfig.pass}
                    onChange={handleDbConfigChange}
                    disabled={loading}
                  />
                  <TextField
                    margin="normal"
                    required
                    fullWidth
                    id="name"
                    label="数据库名称"
                    name="name"
                    value={state.dbConfig.name}
                    onChange={handleDbConfigChange}
                    disabled={loading}
                  />
                </>
              )}
              
              {state.error && (
                <ErrorAlert 
                  message={state.error} 
                  onClose={() => setState(prev => ({ ...prev, error: '' }))} 
                />
              )}
              {state.success && (
                <ErrorAlert 
                  message={state.success} 
                  severity="success"
                  onClose={() => setState(prev => ({ ...prev, success: '' }))} 
                />
              )}
            </form>
          </Paper>
        );
      case 3:
        return (
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              创建管理员账户
            </Typography>
            <Typography variant="body1" paragraph>
              请创建一个管理员账户，用于登录和管理 Tairitsu 平台。
            </Typography>
            <form onSubmit={handleSubmit}>
              <TextField
                margin="normal"
                required
                fullWidth
                id="username"
                label="用户名"
                name="username"
                autoComplete="username"
                value={state.adminData.username}
                onChange={handleAdminDataChange}
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
                value={state.adminData.password}
                onChange={handleAdminDataChange}
                disabled={loading}
                helperText="密码长度至少为6位"
              />
              <TextField
                margin="normal"
                required
                fullWidth
                id="email"
                label="邮箱地址"
                name="email"
                autoComplete="email"
                value={state.adminData.email}
                onChange={handleAdminDataChange}
                disabled={loading}
              />
              {state.error && (
                <ErrorAlert 
                  message={state.error} 
                  onClose={() => setState(prev => ({ ...prev, error: '' }))} 
                />
              )}
            </form>
          </Paper>
        );
      case 4:
        return (
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              完成设置
            </Typography>
            <Typography variant="body1" paragraph>
              恭喜您！Tairitsu 平台已成功配置。
            </Typography>
            <Typography variant="body1" paragraph>
              配置概要：
            </Typography>
            <ul>
              <li>ZeroTier 控制器: {state.ztConfig.controllerUrl}</li>
              <li>数据库类型: {state.dbConfig.type}</li>
              <li>管理员账户: {state.adminData.username}</li>
            </ul>
            {state.error && (
                <ErrorAlert 
                  message={state.error} 
                  onClose={() => setState(prev => ({ ...prev, error: '' }))} 
                />
              )}
              {state.success && (
                <ErrorAlert 
                  message={state.success} 
                  severity="success"
                  onClose={() => setState(prev => ({ ...prev, success: '' }))} 
                />
              )}
          </Paper>
        );
      default:
        return <div>未知步骤</div>;
    }
  };

  return (
    <Container component="main" maxWidth="md" sx={{ mt: 8 }}>
      <Paper elevation={3} sx={{ p: 4, borderRadius: 2 }}>
        <Typography component="h1" variant="h4" align="center" gutterBottom>
          Tairitsu 初始化向导
        </Typography>
        <Stepper activeStep={state.activeStep} sx={{ mb: 4 }}>
          {steps.map((label, index) => (
            <Step key={index}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>
        <div>{renderStepContent(state.activeStep)}</div>
        <Box sx={{ mt: 4, display: 'flex', justifyContent: 'space-between' }}>
          {state.activeStep > 0 && (
            <Button 
              variant="outlined" 
              onClick={handleBack}
              disabled={state.loading || (state.activeStep === 4 && !!state.success)}
            >
              返回
            </Button>
          )}
          <Button 
            variant="contained" 
            color="primary" 
            onClick={handleSubmit}
            disabled={state.loading}
          >
            {state.loading ? <CircularProgress size={24} /> : 
              state.activeStep === 4 ? '完成设置' : '下一步'}
          </Button>
        </Box>
      </Paper>
    </Container>
  );
};

export default SetupWizard;