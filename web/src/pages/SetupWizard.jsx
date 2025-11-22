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

function SetupWizard() {
  const [activeStep, setActiveStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [adminData, setAdminData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: ''
  });
  const [dbConfig, setDbConfig] = useState({
    type: 'sqlite', // 默认选择SQLite
    path: '', // SQLite路径（由程序控制）
    host: '',
    port: 0, // 更改为数字类型，默认值为0
    user: '',
    pass: '',
    name: ''
  });
  const [ztStatus, setZtStatus] = useState(null);
  const navigate = useNavigate();

  const steps = [
    '欢迎使用 Tairitsu',
    '配置数据库',
    '创建管理员账户',
    '检测ZeroTier连接',
    '完成设置'
  ];

  // 检查是否已经完成设置
  useEffect(() => {
    // 这里可以添加检查是否已经完成设置的逻辑
    // 例如检查本地存储或向后端发送请求
  }, []);

  // 当进入第四步（ZeroTier连接检测）时，自动获取ZeroTier状态并尝试初始化
  useEffect(() => {
    if (activeStep === 3) {
      testAndInitZeroTier();
    }
  }, [activeStep]);

  // 测试并初始化ZeroTier连接
  const testAndInitZeroTier = async () => {
    setLoading(true);
    setError('');
    try {
      // 首先测试ZeroTier连接
      await systemAPI.initZeroTierClient();
      
      // 初始化成功后获取实际状态
      const statusResponse = await systemAPI.getSetupStatus();
      setZtStatus(statusResponse.data.ztStatus);
      
      // 确保状态已设置为在线
      if (statusResponse.data.ztStatus && statusResponse.data.ztStatus.online) {
        console.log('[设置向导] ZeroTier连接成功');
      }
    } catch (err) {
      // 初始化失败时，尝试获取当前状态以提供更多信息
      try {
        const statusResponse = await systemAPI.getSetupStatus();
        setZtStatus(statusResponse.data.ztStatus);
      } catch (statusErr) {
        // 如果获取状态也失败，设置为null
        setZtStatus(null);
      }
      
      const errorMsg = '无法连接到ZeroTier控制器，请检查配置';
      setError(errorMsg);
      console.error('[设置向导] ZeroTier连接失败:', err.message || err);
    } finally {
      setLoading(false);
    }
  };

  // 重试连接（保留供用户手动重试）
  const retryZeroTier = async () => {
    await testAndInitZeroTier();
  };

  const handleNext = async () => {
    setError('');
    
    // 在最后一步，保存设置并重定向到登录页面
    if (activeStep === steps.length - 1) {
      console.log('[设置向导] 完成设置，准备跳转到登录页面');
      
      // 使用replace方法导航并刷新页面，确保状态更新
      // 这会清除历史堆栈中的设置向导，防止用户返回到设置向导
      navigate('/login', { replace: true });
      
      // 强制刷新页面，确保App组件重新检查系统状态
      window.location.reload();
      return;
    }
    
    // 在配置数据库步骤，保存数据库配置
    if (activeStep === 1) {
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
    
    // 在创建管理员账户步骤，提交表单
    if (activeStep === 2) {
      // 表单验证
      if (!adminData.username.trim()) {
        setError('请输入用户名');
        return;
      }
      if (!adminData.email.trim()) {
        setError('请输入邮箱地址');
        return;
      }
      if (!adminData.password) {
        setError('请输入密码');
        return;
      }
      if (adminData.password !== adminData.confirmPassword) {
        setError('两次输入的密码不一致');
        return;
      }
      
      setLoading(true);
      try {
        // 注册管理员账户
        await authAPI.register(adminData);
        
        // 继续下一步
        setActiveStep((prevActiveStep) => prevActiveStep + 1);
      } catch (err) {
        setError('管理员账户创建失败: ' + (err.response?.data?.error || err.message));
      } finally {
        setLoading(false);
      }
      return;
    }
    
    // 在检测ZeroTier连接步骤
    if (activeStep === 3) {
      // 只有在ZeroTier状态有效且在线时才能前进
      if (!ztStatus || !ztStatus.online) {
        // 如果检测失败，保持在当前页面并显示错误提示
        setError(ztStatus ? 'ZeroTier未在线，请检查连接' : 'ZeroTier连接检测失败，请重试');
        return;
      }
      
      // 由于已经在进入此步骤时初始化了ZeroTier客户端，这里只需要验证状态
      // 验证通过，直接前进到完成步骤
      setActiveStep((prevActiveStep) => prevActiveStep + 1);
      return;
    }
    
    // 欢迎页面直接继续
    if (activeStep === 0) {
      setActiveStep((prevActiveStep) => prevActiveStep + 1);
      return;
    }
    
    // 最后一个完成步骤，确保前进逻辑清晰
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

  const getStepContent = (step) => {
    switch (step) {
      case 0:
        return (
          <Box sx={{ textAlign: 'center', py: 4 }}>
            <Typography variant="h4" gutterBottom>
              欢迎使用 Tairitsu 控制台
            </Typography>
            <Typography variant="body1" sx={{ mb: 3 }}>
              这是一个强大的ZeroTier网络管理工具，可以帮助您轻松管理和监控您的P2P网络。
            </Typography>
            <Typography variant="body1">
              我们将引导您完成初始设置过程，这只需要几分钟时间。
            </Typography>
          </Box>
        );
      case 1:
        return (
          <Box sx={{ py: 2 }}>
            <Typography variant="h5" gutterBottom>
              配置数据库
            </Typography>
            <Typography variant="body1" sx={{ mb: 3 }}>
              请选择数据库类型来存储用户数据和网络配置信息。
            </Typography>
            
            <Box sx={{ mb: 3 }}>
              <Typography variant="subtitle1" gutterBottom>
                数据库类型
              </Typography>
              <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
                <Button
                  variant={dbConfig.type === 'sqlite' ? 'contained' : 'outlined'}
                  onClick={() => handleDbConfigChange({ target: { name: 'type', value: 'sqlite' } })}
                  disabled={loading}
                >
                  SQLite (推荐)
                </Button>
                <Button
                  variant={dbConfig.type === 'postgresql' ? 'contained' : 'outlined'}
                  onClick={() => handleDbConfigChange({ target: { name: 'type', value: 'postgresql' } })}
                  disabled={loading}
                >
                  PostgreSQL (开发中)
                </Button>
                <Button
                  variant={dbConfig.type === 'mysql' ? 'contained' : 'outlined'}
                  onClick={() => handleDbConfigChange({ target: { name: 'type', value: 'mysql' } })}
                  disabled={loading}
                >
                  MySQL (开发中)
                </Button>
              </Box>
              
              {dbConfig.type === 'sqlite' && (
                <Alert severity="info" sx={{ mb: 2 }}>
                  SQLite数据库将自动创建在程序数据目录中。SQLite适用于小规模网络部署，
                  对于大型生产环境建议使用PostgreSQL或MySQL。
                </Alert>
              )}
              
              {(dbConfig.type === 'postgresql' || dbConfig.type === 'mysql') && (
                <Alert severity="warning" sx={{ mb: 2 }}>
                  注意：{dbConfig.type === 'postgresql' ? 'PostgreSQL' : 'MySQL'}支持正在开发中，
                  当前版本仅支持SQLite数据库。
                </Alert>
              )}
            </Box>
          </Box>
        );
      case 2:
        return (
          <Box sx={{ py: 2 }}>
            <Typography variant="h5" gutterBottom>
              创建管理员账户
            </Typography>
            <Typography variant="body1" sx={{ mb: 3 }}>
              请创建一个管理员账户来管理您的Tairitsu控制台。
            </Typography>
            
            <TextField
              margin="normal"
              required
              fullWidth
              id="username"
              label="用户名"
              name="username"
              autoComplete="username"
              autoFocus
              value={adminData.username}
              onChange={handleAdminDataChange}
              disabled={loading}
            />
            <TextField
              margin="normal"
              required
              fullWidth
              id="email"
              label="邮箱"
              name="email"
              type="email"
              autoComplete="email"
              value={adminData.email}
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
            />
            <TextField
              margin="normal"
              required
              fullWidth
              name="confirmPassword"
              label="确认密码"
              type="password"
              id="confirmPassword"
              value={adminData.confirmPassword}
              onChange={handleAdminDataChange}
              disabled={loading}
            />
          </Box>
        );
      case 3:
        return (
          <Box sx={{ textAlign: 'center', py: 4 }}>
            <Typography variant="h5" gutterBottom>
              检测ZeroTier连接
            </Typography>
            <Typography variant="body1" sx={{ mb: 3 }}>
              正在检测与ZeroTier控制器的连接...
            </Typography>
            
            {loading ? (
              <CircularProgress sx={{ mt: 2 }} />
            ) : ztStatus ? (
              <Box>
                {ztStatus.online ? (
                  <>
                    <Typography variant="body1" color="success.main">
                      ✓ 成功连接到ZeroTier控制器
                    </Typography>
                    <Typography variant="body2" sx={{ mt: 2 }}>
                      版本: {ztStatus.version}
                    </Typography>
                    <Typography variant="body2">
                      地址: {ztStatus.address}
                    </Typography>
                    <Typography variant="body2">
                      在线状态: 在线
                    </Typography>
                  </>
                ) : (
                  <>
                    <Typography variant="body1" color="error.main">
                      ✗ ZeroTier控制器连接失败
                    </Typography>
                    <Typography variant="body2" sx={{ mt: 2, mb: 3 }}>
                      状态: 离线
                    </Typography>
                    <Button 
                      variant="outlined" 
                      onClick={retryZeroTier}
                      color="primary"
                      disabled={loading}
                    >
                      重试连接
                    </Button>
                  </>
                )}
              </Box>
            ) : (
              <Box>
                {error && (
                  <Typography variant="body1" color="error.main" sx={{ mb: 2 }}>
                    ✗ {error}
                  </Typography>
                )}
                <Button 
                  variant="outlined" 
                  onClick={retryZeroTier}
                  color="primary"
                  disabled={loading}
                >
                  检测连接
                </Button>
                <Typography variant="body2" sx={{ mt: 3, color: 'text.secondary' }}>
                  提示: 确保ZeroTier服务已启动且配置正确
                </Typography>
              </Box>
            )}
          </Box>
        );
      case 4:
        return (
          <Box sx={{ textAlign: 'center', py: 4 }}>
            <Typography variant="h4" gutterBottom color="success.main">
              大功告成！
            </Typography>
            <Typography variant="h6" gutterBottom>
              您的Tairitsu控制台已设置完成
            </Typography>
            <Typography variant="body1" sx={{ mb: 3 }}>
              现在您可以使用刚刚创建的管理员账户登录系统，开始管理您的ZeroTier网络。
            </Typography>
            <Typography variant="body1">
              感谢您选择Tairitsu！
            </Typography>
          </Box>
        );
      default:
        return '未知步骤';
    }
  };

  return (
    <Container component="main" maxWidth="md" sx={{ mt: 4, mb: 4 }}>
      <Paper elevation={3} sx={{ p: 4, borderRadius: 2 }}>
        <Typography component="h1" variant="h4" align="center" sx={{ mb: 4 }}>
          Tairitsu 设置向导
        </Typography>
        
        <Stepper activeStep={activeStep} alternativeLabel sx={{ mb: 4 }}>
          {steps.map((label) => (
            <Step key={label}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>
        
        {error && (
          <Alert severity="error" sx={{ mb: 3 }}>
            {error}
          </Alert>
        )}
        
        {getStepContent(activeStep)}
        
        <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 4 }}>
          <Button
            disabled={activeStep === 0 || loading}
            onClick={handleBack}
          >
            返回
          </Button>
          
          <Button
            variant="contained"
            disabled={loading}
            onClick={handleNext}
            sx={{ ml: 2 }}
          >
            {loading ? (
              <CircularProgress size={24} color="inherit" />
            ) : activeStep === steps.length - 1 ? (
              '完成设置'
            ) : (
              '下一步'
            )}
          </Button>
        </Box>
      </Paper>
    </Container>
  );
}

export default SetupWizard;