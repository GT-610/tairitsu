import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Stepper,
  Step,
  StepLabel,
  Paper,
  Button,
  CircularProgress,
  Alert,
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { systemAPI } from '../services/api';

// Admin account data type
interface AdminAccount {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}

// Database configuration type
interface DatabaseConfig {
  type: 'sqlite' | 'mysql' | 'postgresql';
  host?: string;
  port?: number;
  username?: string;
  password?: string;
  database?: string;
  sslMode?: boolean;
}

// ZeroTier configuration type
interface ZeroTierConfig {
  controllerUrl: string;
  apiToken: string;
  enable: boolean;
}

// ZeroTier status type
interface ZeroTierStatus {
  online: boolean;
  peerCount: number;
  address: string;
}

/**
 * SetupWizard Component
 * Guides users through the initial system setup process
 */
function SetupWizard() {
  // Admin account data
  const [adminAccount, setAdminAccount] = useState<AdminAccount>({
    username: '',
    email: '',
    password: '',
    confirmPassword: ''
  });

  // Database configuration
  const [databaseConfig, setDatabaseConfig] = useState<DatabaseConfig>({
    type: 'sqlite',
    sslMode: false
  });

  // ZeroTier configuration
  const [zeroTierConfig, setZeroTierConfig] = useState<ZeroTierConfig>({
    controllerUrl: 'http://localhost:9993',
    apiToken: '',
    enable: true
  });

  // Setup wizard step titles
  const steps = ['欢迎', 'ZeroTier配置', '数据库配置', '管理员账户', '完成'];

  // Current step (0-based index)
  const [activeStep, setActiveStep] = useState<number>(0);
  // Loading state
  const [loading, setLoading] = useState<boolean>(false);
  // Error message state
  const [error, setError] = useState<string>('');
  // Success message state
  const [success, setSuccess] = useState<string>('');
  // ZeroTier connection status
  const [zeroTierStatus, setZeroTierStatus] = useState<ZeroTierStatus | null>(null);
  // Navigation hook
  const navigate = useNavigate();

  // Track setup wizard state in localStorage
  useEffect(() => {
    // Mark setup wizard as started
    localStorage.setItem('setupWizardStarted', 'true');

    // Cleanup function to maintain localStorage integrity
    return () => {
      // Only remove flag if setup process was interrupted
      if (activeStep < steps.length - 1) {
        localStorage.removeItem('setupWizardStarted');
      }
    };
  }, [activeStep]);

  // Validate ZeroTier controller connection and save configuration
  const validateZeroTierConnection = async () => {
    try {
      setLoading(true);
      setError('');

      // Save configuration and test connection simultaneously
      const response = await systemAPI.initZeroTierClient();

      // Get ZeroTier status information from response
      const status: ZeroTierStatus = {
        online: response.data.online || false,
        peerCount: response.data.peerCount || 0,
        address: response.data.address || ''
      };

      setZeroTierStatus(status);
      setLoading(false);
      return true;
    } catch (err: any) {
      setError('ZeroTier connection failed: ' + (err.response?.data?.message || err.message));
      setLoading(false);
      return false;
    }
  };

  // Handle form submission for each step
  const handleNext = async () => {
    setError('');
    setSuccess('');

    // Execute different actions based on current step
    switch (activeStep) {
      case 0:
        // Welcome page - proceed to next step directly
        setActiveStep((prevActiveStep) => prevActiveStep + 1);
        break;

      case 1:
        // ZeroTier configuration step - validate and initialize
        if (zeroTierConfig.enable) {
          const isValid = await validateZeroTierConnection();
          if (isValid) {
            // Validation successful, proceed to next step
            setActiveStep((prevActiveStep) => prevActiveStep + 1);
          }
        } else {
          // ZeroTier disabled, proceed to next step
          setActiveStep((prevActiveStep) => prevActiveStep + 1);
        }
        break;

      case 2:
        // Database configuration validation
        if (!databaseConfig.type) {
          setError('Please select a database type');
          return;
        }

        // Validate based on database type
        if (databaseConfig.type === 'sqlite') {
          // SQLite path is controlled by the program, no user input required
        } else {
          // Validate PostgreSQL or MySQL configuration
          if (!databaseConfig.host || !databaseConfig.port || !databaseConfig.username || !databaseConfig.database) {
            setError('Please fill in complete database connection information');
            return;
          }

          // Basic port validation
          if (databaseConfig.port < 1 || databaseConfig.port > 65535) {
            setError('Please enter a valid port number (1-65535)');
            return;
          }
        }

        try {
          setLoading(true);
          // Send database configuration to backend
          await systemAPI.configureDatabase(databaseConfig);

          // Reload routes after database configuration
          await systemAPI.reloadRoutes();

          // Proceed to next step
          setActiveStep((prevActiveStep) => prevActiveStep + 1);
        } catch (err: any) {
          setError('Database configuration failed: ' + (err.response?.data?.message || err.message));
        } finally {
          setLoading(false);
        }
        break;

      case 3:
        // Create administrator account step
        // Form validation
        if (!adminAccount.username || !adminAccount.email || !adminAccount.password || !adminAccount.confirmPassword) {
          setError('Please fill in complete administrator account information');
          return;
        }

        if (adminAccount.password !== adminAccount.confirmPassword) {
          setError('Passwords do not match');
          return;
        }

        if (adminAccount.password.length < 6) {
          setError('Password length cannot be less than 6 characters');
          return;
        }

        // Simple email format validation
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(adminAccount.email)) {
          setError('Please enter a valid email address');
          return;
        }

        try {
          setLoading(true);
          // Create administrator account
          await systemAPI.initializeAdminCreation();

          // Proceed to next step
          setActiveStep((prevActiveStep) => prevActiveStep + 1);
        } catch (err: any) {
          setError('Administrator account creation failed: ' + (err.response?.data?.message || err.message));
        } finally {
          setLoading(false);
        }
        break;

      case 4:
        // Finalize setup step
        try {
          setLoading(true);
          // Mark system as initialized
          await systemAPI.setInitialized(true);

          // Reload routes to ensure proper initialization
          await systemAPI.reloadRoutes();

          // Update localStorage flag to indicate system initialization
          localStorage.removeItem('setupWizardStarted');
          localStorage.setItem('setupCompleted', 'true');

          // Set success message
          setSuccess('系统初始化完成，即将跳转到登录页面');

          // Delay page refresh to show success message
          setTimeout(() => {
            navigate('/login');
          }, 2000);
        } catch (err: any) {
          setError('System initialization failed: ' + (err.response?.data?.message || err.message));
        } finally {
          setLoading(false);
        }
        break;

      default:
        // Default case: proceed to next step
        setActiveStep((prevActiveStep) => prevActiveStep + 1);
        break;
    }
  };

  // Navigate to previous step
  const handleBack = () => {
    setActiveStep((prevActiveStep) => prevActiveStep - 1);
    setError('');
  };

  // Handle changes to admin account data
  const handleAdminAccountChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setAdminAccount({
      ...adminAccount,
      [name]: value
    });
  };

  // Handle changes to database configuration
  const handleDatabaseConfigChange = (e: React.ChangeEvent<HTMLInputElement | { name?: string; value: unknown }>) => {
    const { name, value, type, checked } = e.target as any;

    // Special handling for database type changes
    if (name === 'type') {
      // Default configuration for SQLite
      const defaultConfig: DatabaseConfig = {
        type: value as 'sqlite' | 'mysql' | 'postgresql',
        sslMode: false
      };

      // Handle switching between database types
      if (value === 'sqlite') {
        // Switch to SQLite with default configuration
        setDatabaseConfig(defaultConfig);
      } else {
        // Switch to MySQL or PostgreSQL with default ports
        defaultConfig.host = 'localhost';
        defaultConfig.port = value === 'mysql' ? 3306 : 5432;
        defaultConfig.username = '';
        defaultConfig.password = '';
        defaultConfig.database = '';
        setDatabaseConfig(defaultConfig);
      }
    } else {
      // Handle changes to other fields
      setDatabaseConfig({
        ...databaseConfig,
        [name]: type === 'checkbox' ? checked : value
      });
    }
  };

  // Helper function to render error and success messages
  const renderMessages = () => {
    return (
      <>
        {error && (
          <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError('')}>
            {error}
          </Alert>
        )}
        {success && (
          <Alert severity="success" sx={{ mb: 3 }}>
            {success}
          </Alert>
        )}
      </>
    );
  };

  // Render content for each step
  const renderStepContent = (step: number) => {
    // Implementation of step content rendering (omitted for brevity)
    return <div>Step {step + 1} content</div>;
  };

  return (
    <Box sx={{ p: 3 }}>
      <Paper elevation={3} sx={{ p: 4, maxWidth: 800, mx: 'auto', mt: 8 }}>
        <Typography variant="h4" component="h1" gutterBottom align="center">
          系统初始化向导
        </Typography>
        <Typography variant="body1" align="center" sx={{ mb: 4, color: 'text.secondary' }}>
          欢迎使用 Tairitsu！请按照向导完成系统初始化配置
        </Typography>

        <Stepper activeStep={activeStep} sx={{ mb: 4 }}>
          {steps.map((label) => (
            <Step key={label}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>

        {renderMessages()}

        <Box sx={{ mt: 2, mb: 3 }}>
          {renderStepContent(activeStep)}
        </Box>

        <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 4 }}>
          <Button
            disabled={activeStep === 0 || loading}
            onClick={handleBack}
            variant="outlined"
          >
            上一步
          </Button>
          <Button
            variant="contained"
            onClick={handleNext}
            disabled={loading}
          >
            {loading ? (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <CircularProgress size={20} color="inherit" />
                处理中...
              </Box>
            ) : activeStep === steps.length - 1 ? (
              '完成'
            ) : (
              '下一步'
            )}
          </Button>
        </Box>
      </Paper>
    </Box>
  );
}

export default SetupWizard;