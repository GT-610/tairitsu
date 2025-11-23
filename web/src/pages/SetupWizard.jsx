import React, { useState, useEffect } from 'react';
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
  Alert,
  TextField
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { authAPI, statusAPI, systemAPI } from '../services/api.js';

const SetupWizard = () => {
  const navigate = useNavigate();
  const [activeStep, setActiveStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [adminData, setAdminData] = useState({
    username: '',
    password: '',
    email: ''
  });
  const [dbConfig, setDbConfig] = useState({
    type: 'sqlite',
    path: '',
    host: '',
    port: 0,
    user: '',
    pass: '',
    name: ''
  });
  const [ztConfig, setZtConfig] = useState({
    controllerUrl: 'http://localhost:9993',
    tokenPath: '/var/lib/zerotier-one/authtoken.secret'
  });
  const [ztStatus, setZtStatus] = useState(null);
  const [ztConnected, setZtConnected] = useState(false);

  // 重新定义步骤顺序
  const steps = [
    '欢迎使用 Tairitsu',
    '配置 ZeroTier 控制器',
    '配置数据库',
    '创建管理员账户',
    '完成设置'
  ];

  useEffect(() => {
    // 检查系统状态，确定是否需要显示设置向导
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

    checkSystemStatus();
  }, [navigate]);

  // 测试ZeroTier连接并初始化客户端
  const testAndInitZtConnection = async () => {
    setLoading(true);
    setError('');
    setSuccess('');
    try {
      // 保存配置并同时测试连接
      const response = await systemAPI.saveZtConfig(ztConfig);
      // 从响应中获取ZeroTier状态信息
      setZtStatus(response.data.status);
      setZtConnected(true);
      setSuccess('ZeroTier 连接成功！');
      return true;
    } catch (err) {
      setError('ZeroTier 连接失败: ' + (err.response?.data?.error || err.message));
      setZtConnected(false);
      return false;
    } finally {
      setLoading(false);
    }
  };

  // 处理表单提交
  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');
    setSuccess('');

    // 根据当前步骤执行不同的操作
    if (activeStep === 1) {
      // ZeroTier 配置步骤 - 直接验证并初始化
      const success = await testAndInitZtConnection();
      if (success) {
        // 验证成功，进入下一步
        setActiveStep((prevActiveStep) => prevActiveStep + 1);
      }
      return;
    }

    if (activeStep === 2) {
      // 数据库配置验证
      if (!dbConfig.type) {
        setError('请选择数据库类型');
        return;
      }
      
      // 根据数据库类型进行不同验证
      if (dbConfig.type === 'sqlite') {
        // SQLite 路径由程序控制，不需要用户输入
      } else {
        // PostgreSQL或MySQL验证
        if (!dbConfig.host.trim()) {
          setError('请输入数据库主机地址');
          return;
        }
        if (!dbConfig.port || dbConfig.port <= 0) {
          setError('请输入有效的数据库端口');
          return;
        }
        if (!dbConfig.user.trim()) {
          setError('请输入数据库用户名');
          return;
        }
        if (!dbConfig.name.trim()) {
          setError('请输入数据库名称');
          return;
        }
      }
      
      setLoading(true);
      try {
        // 发送数据库配置到后端
        await systemAPI.configureDatabase(dbConfig);
        // 继续下一步
        setActiveStep((prevActiveStep) => prevActiveStep + 1);
      } catch (err) {
        setError('数据库配置保存失败: ' + (err.response?.data?.error || err.message));
      } finally {
        setLoading(false);
      }
      return;
    }
    
    if (activeStep === 3) {
      // 创建管理员账户步骤
      // 表单验证
      if (!adminData.username.trim()) {
        setError('请输入用户名');
        return;
      }
      if (!adminData.password || adminData.password.length < 6) {
        setError('密码长度至少为6位');
        return;
      }
      if (!adminData.email.trim()) {
        setError('请输入邮箱地址');
        return;
      }
      
      // 简单的邮箱格式验证
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(adminData.email)) {
        setError('请输入有效的邮箱地址');
        return;
      }
      
      setLoading(true);
      try {
        // 创建管理员账户
        await authAPI.register(adminData);
        // 继续下一步
        setActiveStep((prevActiveStep) => prevActiveStep + 1);
      } catch (err) {
        setError('创建管理员账户失败: ' + (err.response?.data?.error || err.message));
      } finally {
        setLoading(false);
      }
      return;
    }
    
    if (activeStep === 4) {
      // 完成设置步骤
      setLoading(true);
      try {
        // 设置系统为已初始化状态
        await systemAPI.setInitialized(true);
        // 设置成功消息
        setSuccess('系统初始化完成！即将跳转到登录页面...');
        // 延迟跳转到登录页面
        setTimeout(() => {
          navigate('/login');
        }, 2000);
      } catch (err) {
        setError('完成设置失败: ' + (err.response?.data?.error || err.message));
      } finally {
        setLoading(false);
      }
      return;
    }
    
    // 默认情况：前进到下一步
    setActiveStep((prevActiveStep) => prevActiveStep + 1);
  };

  const handleBack = () => {
    setActiveStep((prevActiveStep) => prevActiveStep - 1);
  };

  const handleAdminDataChange = (e) => {
    setAdminData({
      ...adminData,
      [e.target.name]: e.target.value
    });
  };

  const handleDbConfigChange = (e) => {
    // 特殊处理数据库类型变更
    if (e.target.name === 'type') {
      const newType = e.target.value;
      setDbConfig({
        type: newType,
        path: newType === 'sqlite' ? '' : dbConfig.path,
        host: newType === 'sqlite' ? '' : dbConfig.host,
        port: newType === 'sqlite' ? 0 : dbConfig.port, // SQLite不需要端口
        user: newType === 'sqlite' ? '' : dbConfig.user,
        pass: newType === 'sqlite' ? '' : dbConfig.pass,
        name: newType === 'sqlite' ? '' : dbConfig.name
      });
    } else {
      // 处理其他字段变更
      setDbConfig({
        ...dbConfig,
        [e.target.name]: e.target.value
      });
    }
  };

  const handleZtConfigChange = (e) => {
    setZtConfig({
      ...ztConfig,
      [e.target.name]: e.target.value
    });
  };

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
              {error && (
                <Alert severity="error" sx={{ mt: 2 }}>
                  {error}
                </Alert>
              )}
              {success && (
                <Alert severity="success" sx={{ mt: 2 }}>
                  {success}
                </Alert>
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
                value={dbConfig.type}
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
                    value={dbConfig.host}
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
                    value={dbConfig.port || ''}
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
                    value={dbConfig.user}
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
                    value={dbConfig.pass}
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
                    value={dbConfig.name}
                    onChange={handleDbConfigChange}
                    disabled={loading}
                  />
                </>
              )}
              
              {error && (
                <Alert severity="error" sx={{ mt: 2 }}>
                  {error}
                </Alert>
              )}
              {success && (
                <Alert severity="success" sx={{ mt: 2 }}>
                  {success}
                </Alert>
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
                value={adminData.username}
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
                value={adminData.password}
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
                value={adminData.email}
                onChange={handleAdminDataChange}
                disabled={loading}
              />
              {error && (
                <Alert severity="error" sx={{ mt: 2 }}>
                  {error}
                </Alert>
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
              <li>ZeroTier 控制器: {ztConfig.controllerUrl}</li>
              <li>数据库类型: {dbConfig.type}</li>
              <li>管理员账户: {adminData.username}</li>
            </ul>
            {error && (
              <Alert severity="error" sx={{ mt: 2 }}>
                {error}
              </Alert>
            )}
            {success && (
              <Alert severity="success" sx={{ mt: 2 }}>
                {success}
              </Alert>
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
        <Stepper activeStep={activeStep} sx={{ mb: 4 }}>
          {steps.map((label, index) => (
            <Step key={index}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>
        <div>{renderStepContent(activeStep)}</div>
        <Box sx={{ mt: 4, display: 'flex', justifyContent: 'space-between' }}>
          {activeStep > 0 && (
            <Button 
              variant="outlined" 
              onClick={handleBack}
              disabled={loading || (activeStep === 4 && !!success)}
            >
              返回
            </Button>
          )}
          <Button 
            variant="contained" 
            color="primary" 
            onClick={handleSubmit}
            disabled={loading}
          >
            {loading ? <CircularProgress size={24} /> : 
              activeStep === 4 ? '完成设置' : '下一步'}
          </Button>
        </Box>
      </Paper>
    </Container>
  );
};

export default SetupWizard;