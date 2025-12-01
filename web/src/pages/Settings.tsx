import { useState, useEffect } from 'react';
import { Box, Typography, Paper, TextField, Button, Switch, FormControlLabel, Alert, CircularProgress } from '@mui/material';
import { systemAPI } from '../services/api';

// Settings type definition
interface Settings {
  apiPort: number;
  debugMode: boolean;
  logLevel: 'debug' | 'info' | 'warn' | 'error';
  autoBackup: boolean;
  backupInterval: number;
  theme: 'light' | 'dark' | 'system';
}

function Settings() {
  const [settings, setSettings] = useState<Settings>({
    apiPort: 8000,
    debugMode: false,
    logLevel: 'info',
    autoBackup: true,
    backupInterval: 24,
    theme: 'system'
  });
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');
  const [isChanged, setIsChanged] = useState<boolean>(false);

  // Fetch settings from API
  useEffect(() => {
    const fetchSettings = async () => {
      try {
        setLoading(true);
        // Note: The API method name here may need to be adjusted according to the actual situation
        // const response = await systemAPI.getSettings();
        // setSettings(response.data);
        // Since the current API may not have a getSettings method, we use mock data
        setLoading(false);
      } catch (err: any) {
        setError('Failed to fetch settings: ' + (err.response?.data?.message || err.message));
        setLoading(false);
      }
    };

    fetchSettings();
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setSettings({
      ...settings,
      [name]: type === 'checkbox' ? checked : value
    });
    setIsChanged(true);
  };

  const handleSave = async () => {
    try {
      setLoading(true);
      // Note: The API method name here may need to be adjusted according to the actual situation
      // await systemAPI.updateSettings(settings);
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      setSuccess('Settings saved successfully');
      setIsChanged(false);
      setLoading(false);
      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(''), 3000);
    } catch (err: any) {
      setError('Failed to save settings: ' + (err.response?.data?.message || err.message));
      setLoading(false);
    }
  };

  const handleReset = () => {
    setSettings({
      apiPort: 8000,
      debugMode: false,
      logLevel: 'info',
      autoBackup: true,
      backupInterval: 24,
      theme: 'system'
    });
    setIsChanged(true);
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        设置
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      {success && (
        <Alert severity="success" sx={{ mb: 3 }} onClose={() => setSuccess('')}>
          {success}
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <Paper elevation={3} sx={{ p: 3 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
            <Typography variant="h5" component="h2">
              系统设置
            </Typography>
            <Box sx={{ display: 'flex', gap: 2 }}>
              <Button 
                variant="outlined" 
                onClick={handleReset}
                disabled={loading || !isChanged}
              >
                重置
              </Button>
              <Button 
                variant="contained" 
                onClick={handleSave}
                disabled={loading || !isChanged}
              >
                {loading ? '保存中...' : '保存设置'}
              </Button>
            </Box>
          </Box>

          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 3, mt: 3 }}>
            <Box>
              <Typography variant="subtitle1" gutterBottom>
                服务器设置
              </Typography>
              <TextField
                fullWidth
                label="API端口"
                name="apiPort"
                type="number"
                value={settings.apiPort}
                onChange={handleChange}
                sx={{ mb: 2 }}
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={settings.debugMode}
                    onChange={handleChange}
                    name="debugMode"
                  />
                }
                label="调试模式"
              />
            </Box>

            <Box>
              <Typography variant="subtitle1" gutterBottom>
                日志设置
              </Typography>
              <TextField
                fullWidth
                select
                label="日志级别"
                name="logLevel"
                value={settings.logLevel}
                onChange={handleChange}
                sx={{ mb: 2 }}
                SelectProps={{
                  native: true,
                }}
              >
                <option value="debug">调试</option>
                <option value="info">信息</option>
                <option value="warn">警告</option>
                <option value="error">错误</option>
              </TextField>
            </Box>

            <Box>
              <Typography variant="subtitle1" gutterBottom>
                备份设置
              </Typography>
              <FormControlLabel
                control={
                  <Switch
                    checked={settings.autoBackup}
                    onChange={handleChange}
                    name="autoBackup"
                  />
                }
                label="自动备份"
              />
              <TextField
                fullWidth
                label="备份间隔（小时）"
                name="backupInterval"
                type="number"
                value={settings.backupInterval}
                onChange={handleChange}
                disabled={!settings.autoBackup}
                sx={{ mt: 2 }}
              />
            </Box>

            <Box>
              <Typography variant="subtitle1" gutterBottom>
                外观设置
              </Typography>
              <TextField
                fullWidth
                select
                label="主题"
                name="theme"
                value={settings.theme}
                onChange={handleChange}
              >
                <option value="light">浅色</option>
                <option value="dark">深色</option>
                <option value="system">跟随系统</option>
              </TextField>
            </Box>
          </Box>
        </Paper>
      )}
    </Box>
  );
}

export default Settings;