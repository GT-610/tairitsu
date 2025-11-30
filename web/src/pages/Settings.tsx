import React, { useState, useEffect } from 'react';
import { Box, Typography, Card, CardContent, Switch, Button, TextField, FormControlLabel, Alert }
from '@mui/material';
import { systemAPI } from '../services/api';

// 设置类型定义
interface SettingsType {
  autoApproveMembers: boolean;
  maxNetworkSize: number;
  keepaliveInterval: number;
  logLevel: 'debug' | 'info' | 'warn' | 'error';
}

function Settings() {
  const [settings, setSettings] = useState<SettingsType>({
    autoApproveMembers: false,
    maxNetworkSize: 50,
    keepaliveInterval: 300,
    logLevel: 'info'
  });
  const [loading, setLoading] = useState<boolean>(true);
  const [saving, setSaving] = useState<boolean>(false);
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      // 注意：这里的API方法名可能需要根据实际情况调整
      // const response = await systemAPI.getSettings();
      // setSettings(response.data);
      // 由于当前API可能没有getSettings方法，我们使用模拟数据
      setSettings({
        autoApproveMembers: false,
        maxNetworkSize: 50,
        keepaliveInterval: 300,
        logLevel: 'info'
      });
    } catch (err: any) {
      setError('获取系统设置失败');
    } finally {
      setLoading(false);
    }
  };

  const handleToggle = (name: keyof SettingsType) => (event: React.ChangeEvent<HTMLInputElement>) => {
    setSettings({
      ...settings,
      [name]: event.target.checked
    });
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type } = e.target;
    setSettings({
      ...settings,
      [name]: type === 'number' ? parseInt(value) : value
    });
  };

  const handleSave = async () => {
    setSaving(true);
    setError('');
    setSuccess('');
    
    try {
      await systemAPI.updateSettings(settings);
      setSuccess('设置已成功保存');
      setTimeout(() => setSuccess(''), 3000);
    } catch (err: any) {
      setError('保存设置失败');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body1">加载中...</Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        系统设置
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
      
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            成员管理设置
          </Typography>
          
          <FormControlLabel
            control={
              <Switch
                checked={settings.autoApproveMembers}
                onChange={handleToggle('autoApproveMembers')}
              />
            }
            label="自动批准新成员加入"
          />
          
          <TextField
            fullWidth
            label="最大网络成员数"
            name="maxNetworkSize"
            type="number"
            value={settings.maxNetworkSize}
            onChange={handleChange}
            margin="normal"
            InputProps={{
              inputProps: {
                min: 1,
                max: 1000
              }
            }}
          />
        </CardContent>
      </Card>
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            网络配置
          </Typography>
          
          <TextField
            fullWidth
            label="保活间隔 (秒)"
            name="keepaliveInterval"
            type="number"
            value={settings.keepaliveInterval}
            onChange={handleChange}
            margin="normal"
            InputProps={{
              inputProps: {
                min: 10,
                max: 3600
              }
            }}
          />
        </CardContent>
      </Card>
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            日志配置
          </Typography>
          <TextField
            fullWidth
            label="日志级别"
            name="logLevel"
            value={settings.logLevel}
            onChange={handleChange}
            margin="normal"
            select
            SelectProps={{
              native: true
            }}
          >
            <option value="debug">Debug</option>
            <option value="info">Info</option>
            <option value="warn">Warning</option>
            <option value="error">Error</option>
          </TextField>
        </CardContent>
      </Card>
      
      <Button
        variant="contained"
        onClick={handleSave}
        disabled={saving}
      >
        {saving ? '保存中...' : '保存设置'}
      </Button>
    </Box>
  );
}

export default Settings;