import { useState, useEffect } from 'react';
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
  TextField,
  IconButton
} from '@mui/material';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import { useNavigate } from 'react-router-dom';
import { authAPI, systemAPI } from '../services/api.js';

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

  // Setup wizard step titles
  const steps = [
    '欢迎使用 Tairitsu',
    '配置 ZeroTier 控制器',
    '配置数据库',
    '创建管理员账户',
    '完成设置'
  ];

  // Track setup wizard state in localStorage
  useEffect(() => {
    // Mark setup wizard as started
    localStorage.setItem('tairitsu_setup_started', 'true');
    
    // Cleanup function to maintain localStorage integrity
    return () => {
      // Only remove flag if setup process was interrupted
      if (!localStorage.getItem('tairitsu_initialized')) {
        localStorage.removeItem('tairitsu_setup_started');
      }
    };
  }, []);
  
  // Verify if system is already initialized before continuing setup
  const checkSystemStatus = async () => {
    try {
      const response = await systemAPI.getSetupStatus();
      if (response.data.initialized) {
        // If system is initialized, redirect to login page
        navigate('/login');
      }
    } catch (err) {
      console.error('检查系统状态失败:', err);
      // Even if check fails, continue showing setup wizard
    }
  };
  
  // Validate ZeroTier controller connection and save configuration
  const testAndInitZtConnection = async () => {
    setLoading(true);
    setError('');
    setSuccess('');
    try {
      // Save configuration and test connection simultaneously
      const response = await systemAPI.saveZtConfig(ztConfig);
      // Get ZeroTier status information from response
      setZtStatus(response.data.status);
      setZtConnected(true);
      setSuccess('ZeroTier 连接成功！已自动前往下一步。');
      return true;
    } catch (err) {
      setError('ZeroTier 连接失败: ' + (err.response?.data?.error || err.message));
      setZtConnected(false);
      return false;
    } finally {
      setLoading(false);
    }
  };

  // Process welcome step submission
  const handleWelcomeStepSubmit = async () => {
    await checkSystemStatus();
    // If checkSystemStatus didn't redirect, continue execution
    if (activeStep === 0) {
      setActiveStep((prevActiveStep) => prevActiveStep + 1);
    }
  };

  // Validate and process ZeroTier configuration
  const handleZtStepSubmit = async () => {
    const success = await testAndInitZtConnection();
    if (success) {
      // Validation successful, proceed to next step
      setActiveStep((prevActiveStep) => prevActiveStep + 1);
    }
  };

  // Validate and submit database configuration
  const handleDbStepSubmit = async () => {
    // Basic validation for all database types
    if (!dbConfig.type) {
      setError('请选择数据库类型');
      return false;
    }
    
    // Additional validation for MySQL and PostgreSQL
    if (dbConfig.type !== 'sqlite') {
      // PostgreSQL or MySQL validation
      if (!dbConfig.host.trim()) {
        setError('请输入数据库主机地址');
        return false;
      }
      if (!dbConfig.port || dbConfig.port <= 0) {
        setError('请输入有效的数据库端口');
        return false;
      }
      if (!dbConfig.user.trim()) {
        setError('请输入数据库用户名');
        return false;
      }
      if (!dbConfig.name.trim()) {
        setError('请输入数据库名称');
        return false;
      }
    }
    
    // Send database configuration to backend
    await systemAPI.configureDatabase(dbConfig);
    return true;
  };

  // Validate administrator account creation form
  const handleAdminStepSubmit = async () => {
    // Form validation
    if (!adminData.username.trim()) {
      setError('请输入用户名');
      return false;
    }
    if (!adminData.password || adminData.password.length < 6) {
      setError('密码长度至少为6位');
      return false;
    }
    if (!adminData.email.trim()) {
      setError('请输入邮箱地址');
      return false;
    }
    
    // Simple email format validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(adminData.email)) {
      setError('请输入有效的邮箱地址');
      return false;
    }
    
    // Create admin account
    await authAPI.register(adminData);
    return true;
  };

  // Finalize system setup and mark as initialized
  const handleCompletionStepSubmit = async () => {
    // Set system to initialized state
    await systemAPI.setInitialized(true);
    // Update localStorage flag to indicate system is initialized
    localStorage.setItem('tairitsu_initialized', 'true');
    localStorage.removeItem('tairitsu_setup_started');
    // Set success message
    setSuccess('系统初始化完成！即将跳转到登录页面...');
    // Delay redirect to login page
    setTimeout(() => {
      navigate('/login');
    }, 2000);
  };

  // Main form submission handler - routes to appropriate step processor
  // 处理进入创建管理员账户步骤的初始化
  const initializeAdminStep = async () => {
    try {
      console.log('初始化管理员账户创建步骤');
      await systemAPI.initializeAdminCreation();
      console.log('管理员账户创建步骤初始化成功');
    } catch (err) {
      console.error('管理员账户创建步骤初始化失败:', err);
      // 非致命错误，不阻止用户继续
    }
  };
  
  // 监听步骤变化，当进入创建管理员账户步骤时执行初始化
  useEffect(() => {
    if (activeStep === 3) {
      initializeAdminStep();
    }
  }, [activeStep]);

  const handleSubmit = async () => {
    setLoading(true);
    setError('');
    setSuccess('');
    
    try {
      // Execute different operations based on current step
      switch (activeStep) {
        case 0:
          await handleWelcomeStepSubmit();
          break;
        case 1:
          await handleZtStepSubmit();
          break;
        case 2:
          const dbSuccess = await handleDbStepSubmit();
          if (dbSuccess) {
            setActiveStep((prevActiveStep) => prevActiveStep + 1);
          }
          break;
        case 3:
          const adminSuccess = await handleAdminStepSubmit();
          if (adminSuccess) {
            setActiveStep((prevActiveStep) => prevActiveStep + 1);
          }
          break;
        case 4:
          await handleCompletionStepSubmit();
          break;
        default:
          // Default case: proceed to next step
          setActiveStep((prevActiveStep) => prevActiveStep + 1);
      }
    } catch (err) {
      setError('操作失败: ' + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const handleBack = () => {
    setActiveStep((prevActiveStep) => prevActiveStep - 1);
  };

  // Generic form field change handler for state updates
  const handleFormChange = (setState, currentState) => (e) => {
    setState({
      ...currentState,
      [e.target.name]: e.target.value
    });
  };

  // Enhanced handler for database configuration with type-specific logic
  const handleDbConfigChange = (e) => {
    // Handle database type switching with appropriate default configurations
    if (e.target.name === 'type') {
      const newType = e.target.value;
      // Default configuration for SQLite
      const sqliteConfig = {
        type: 'sqlite',
        path: '',
        host: '',
        port: 0,
        user: '',
        pass: '',
        name: ''
      };
      
      // Handle default configurations for MySQL and PostgreSQL
      if (newType === 'sqlite') {
        // Switching to SQLite, using default configuration
        setDbConfig(sqliteConfig);
      } else {
        // Switching to MySQL or PostgreSQL, set default port
        const defaultPort = newType === 'mysql' ? 3306 : 5432;
        setDbConfig({
          ...dbConfig,
          type: newType,
          // Only set default port if current port is empty or 0
          port: dbConfig.port || dbConfig.port === 0 ? dbConfig.port : defaultPort
        });
      }
    } else {
      // Use generic function to handle other field changes
      handleFormChange(setDbConfig, dbConfig)(e);
    }
  };

  // Apply generic handler to specific form sections
  const handleAdminDataChange = handleFormChange(setAdminData, adminData);
  const handleZtConfigChange = handleFormChange(setZtConfig, ztConfig);

  // Render error and success feedback messages
  const renderMessages = () => (
    <>
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
    </>
  );
  
  // Reusable step content container component with consistent styling
  const StepContent = ({ title, description, children, showMessages = true }) => (
    <Paper sx={{ p: 3 }}>
      <Typography variant="h5" gutterBottom>
        {title}
      </Typography>
      {description && (
        <Typography variant="body1" paragraph>
          {description}
        </Typography>
      )}
      {children}
      {showMessages && renderMessages()}
    </Paper>
  );
  

  // Generic form field generator to standardize input elements
  const formField = ({ 
    name, 
    label, 
    type = 'text', 
    required = false, 
    value, 
    onChange, 
    autoComplete, 
    helperText, 
    placeholder 
  }) => (
    <TextField
      margin="normal"
      required={required}
      fullWidth
      id={name}
      label={label}
      name={name}
      type={type}
      autoComplete={autoComplete}
      value={value}
      onChange={onChange}
      disabled={loading}
      helperText={helperText}
      placeholder={placeholder}
    />
  );

  // Render appropriate content based on current step
  const renderStepContent = (step) => {
    switch (step) {
      case 0:
        return (
          <StepContent 
            title=""
            description=""
            showMessages={false}
          >
            <Box 
              sx={{ 
                display: 'flex', 
                flexDirection: 'column', 
                alignItems: 'center', 
                justifyContent: 'center', 
                minHeight: 300,
                textAlign: 'center'
              }}
            >
              <Typography variant="h3" gutterBottom>
                欢迎使用 Tairitsu
              </Typography>
              <IconButton 
                sx={{ 
                  mt: 4, 
                  width: 80, 
                  height: 80, 
                  backgroundColor: '#1976d2',
                  color: 'white',
                  '&:hover': {
                    backgroundColor: '#1565c0',
                  }
                }}
                onClick={handleSubmit}
                disabled={loading}
              >
                {loading ? <CircularProgress size={40} sx={{ color: 'white' }} /> : 
                  <ArrowForwardIcon sx={{ width: 40, height: 40 }} />}
              </IconButton>
            </Box>
          </StepContent>
        );
      case 1:
        return (
          <StepContent 
            title="配置 ZeroTier 控制器"
            description="请输入您的 ZeroTier 控制器信息，以便 Tairitsu 能够连接和管理您的网络。"
          >
            <div>
              {formField({
                name: 'controllerUrl',
                label: 'ZeroTier 控制器 URL',
                required: true,
                autoComplete: 'url',
                value: ztConfig.controllerUrl,
                onChange: handleZtConfigChange
              })}
              {formField({
                name: 'tokenPath',
                label: '认证令牌文件路径',
                required: true,
                autoComplete: 'file-path',
                value: ztConfig.tokenPath,
                onChange: handleZtConfigChange,
                helperText: '默认为 /var/lib/zerotier-one/authtoken.secret'
              })}
            </div>
          </StepContent>
        );
      case 2:
        return (
          <StepContent 
            title="配置数据库"
            description="请选择并配置您要使用的数据库。Tairitsu 支持 SQLite、MySQL 和 PostgreSQL。"
          >
            <div>
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
                  {formField({
                    name: 'host',
                    label: '主机地址',
                    required: true,
                    autoComplete: 'hostname',
                    value: dbConfig.host,
                    onChange: handleDbConfigChange
                  })}
                  {formField({
                    name: 'port',
                    label: '端口',
                    required: true,
                    type: 'number',
                    value: dbConfig.port || '',
                    onChange: handleDbConfigChange,
                    placeholder: dbConfig.type === 'mysql' ? '3306' : '5432'
                  })}
                  {formField({
                    name: 'user',
                    label: '用户名',
                    required: true,
                    autoComplete: 'username',
                    value: dbConfig.user,
                    onChange: handleDbConfigChange
                  })}
                  {formField({
                    name: 'pass',
                    label: '密码',
                    required: true,
                    type: 'password',
                    autoComplete: 'current-password',
                    value: dbConfig.pass,
                    onChange: handleDbConfigChange
                  })}
                  {formField({
                    name: 'name',
                    label: '数据库名称',
                    required: true,
                    value: dbConfig.name,
                    onChange: handleDbConfigChange
                  })}
                </>
              )}
            </div>
          </StepContent>
        );
      case 3:
        return (
          <StepContent 
            title="创建管理员账户"
            description="请创建一个管理员账户，用于登录和管理 Tairitsu 平台。"
          >
            <div>
              {formField({
                name: 'username',
                label: '用户名',
                required: true,
                autoComplete: 'username',
                value: adminData.username,
                onChange: handleAdminDataChange
              })}
              {formField({
                name: 'password',
                label: '密码',
                required: true,
                type: 'password',
                autoComplete: 'new-password',
                value: adminData.password,
                onChange: handleAdminDataChange,
                helperText: '密码长度至少为6位'
              })}
              {formField({
                name: 'email',
                label: '邮箱地址',
                required: true,
                autoComplete: 'email',
                value: adminData.email,
                onChange: handleAdminDataChange
              })}
            </div>
          </StepContent>
        );
      case 4:
        return (
          <StepContent 
            title="完成设置"
            description="恭喜您！Tairitsu 平台已成功配置。"
          >
            <Typography variant="body1" paragraph>
              配置概要：
            </Typography>
            <ul>
              <li>ZeroTier 控制器: {ztConfig.controllerUrl}</li>
              <li>数据库类型: {dbConfig.type}</li>
              <li>管理员账户: {adminData.username}</li>
            </ul>
          </StepContent>
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
          {activeStep > 0 && (
            <Button 
              variant="contained" 
              color="primary" 
              onClick={handleSubmit}
              disabled={loading}
            >
              {loading ? <CircularProgress size={24} /> : 
                activeStep === 4 ? '完成设置' : '下一步'}
            </Button>
          )}
        </Box>
      </Paper>
    </Container>
  );
};

export default SetupWizard;