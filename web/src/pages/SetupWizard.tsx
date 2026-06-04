import { useEffect, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Container,
  FormControl,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Step,
  StepLabel,
  Stepper,
  TextField,
  Typography,
} from '@mui/material';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import { useTranslation, type LanguagePreference } from '../i18n';
import { authAPI, systemAPI, type DatabaseSetupConfig, type SetupStatus, type ZeroTierSetupConfig } from '../services/api';
import { getErrorMessage } from '../services/errors';
import { getInitialSetupWizardStep, isSetupStepSaved, setupWizardDatabaseStepCopy } from '../utils/setupWizard';

interface AdminData {
  username: string;
  password: string;
}

const setupCompletedEvent = 'tairitsu-setup-complete';

const defaultDbConfig: DatabaseSetupConfig = {
  type: 'sqlite',
  path: '',
};

const defaultZtConfig: ZeroTierSetupConfig = {
  controllerUrl: 'http://localhost:9993',
  tokenPath: '/var/lib/zerotier-one/authtoken.secret',
};

function SetupWizard() {
  const { preference, setPreference, translateText } = useTranslation();

  const steps = [
    translateText('欢迎使用 Tairitsu'),
    translateText('配置 ZeroTier 控制器'),
    translateText('配置数据库'),
    translateText('创建管理员账户'),
    translateText('完成设置'),
  ];
  const [activeStep, setActiveStep] = useState(0);
  const [status, setStatus] = useState<SetupStatus | null>(null);
  const [initialLoading, setInitialLoading] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [adminData, setAdminData] = useState<AdminData>({ username: '', password: '' });
  const [dbConfig, setDbConfig] = useState<DatabaseSetupConfig>(defaultDbConfig);
  const [ztConfig, setZtConfig] = useState<ZeroTierSetupConfig>(defaultZtConfig);

  const hydrateFromStatus = (nextStatus: SetupStatus) => {
    setStatus(nextStatus);
    setActiveStep(getInitialSetupWizardStep(nextStatus));
    setZtConfig({
      controllerUrl: nextStatus.zeroTierConfig?.controllerUrl || defaultZtConfig.controllerUrl,
      tokenPath: nextStatus.zeroTierConfig?.tokenPath || defaultZtConfig.tokenPath,
    });
    setDbConfig((previous) => ({
      ...previous,
      type: 'sqlite',
      path: nextStatus.databaseConfig?.path || '',
    }));
    if (nextStatus.adminUsername) {
      setAdminData((previous) => ({
        ...previous,
        username: nextStatus.adminUsername || previous.username,
        password: nextStatus.hasAdmin ? '' : previous.password,
      }));
    }
  };

  const fetchSetupStatus = async () => {
    const response = await systemAPI.getSetupStatus();
    hydrateFromStatus(response.data);
    return response.data;
  };

  useEffect(() => {
    const load = async () => {
      try {
        setInitialLoading(true);
        setError('');
        await fetchSetupStatus();
      } catch (err: unknown) {
        setError(getErrorMessage(err, translateText('获取初始化状态失败')));
      } finally {
        setInitialLoading(false);
      }
    };

    void load();
  }, []);

  const runStep = async () => {
    setLoading(true);
    setError('');
    setSuccess('');

    try {
      if (activeStep === 0) {
        setActiveStep(1);
        return;
      }

      if (activeStep === 1) {
        const response = await systemAPI.saveZtConfig(ztConfig);
        const nextStatus = await fetchSetupStatus();
        setSuccess(`${translateText('ZeroTier 控制器连接成功并已保存：')}${response.data.status.address || response.data.config.controllerUrl}`);
        setActiveStep(Math.max(2, getInitialSetupWizardStep(nextStatus)));
        return;
      }

      if (activeStep === 2) {
        const response = await systemAPI.configureDatabase(dbConfig);
        const nextStatus = await fetchSetupStatus();
        const savedPath = response.data.config.path || nextStatus.databaseConfig?.path || 'data/tairitsu.db';
        setSuccess(`${translateText('SQLite 配置已保存：')}${savedPath}`);
        setActiveStep(Math.max(3, getInitialSetupWizardStep(nextStatus)));
        return;
      }

      if (activeStep === 3) {
        if (!adminData.username.trim()) {
          setError(translateText('请输入用户名'));
          return;
        }
        if (!adminData.password || adminData.password.length < 6) {
          setError(translateText('密码长度至少为6位'));
          return;
        }

        if (!status?.adminCreationPrepared && !status?.hasAdmin) {
          await systemAPI.initializeAdminCreation();
        }
        await authAPI.register(adminData);
        const nextStatus = await fetchSetupStatus();
        setSuccess(`${translateText('首个管理员')} ${adminData.username.trim()} ${translateText('创建成功')}`);
        setAdminData((previous) => ({ ...previous, password: '' }));
        setActiveStep(Math.max(4, getInitialSetupWizardStep(nextStatus)));
        return;
      }

      if (activeStep === 4) {
        await systemAPI.setInitialized(true);
        const nextStatus = await fetchSetupStatus();
        if (!nextStatus.initialized) {
          throw new Error(translateText('初始化状态尚未生效，请稍后重试'));
        }

        setSuccess(translateText('系统初始化完成，正在进入登录页面...'));
        window.dispatchEvent(new Event(setupCompletedEvent));
        return;
      }
    } catch (err: unknown) {
      setError(getErrorMessage(err, translateText('操作失败')));
    } finally {
      setLoading(false);
    }
  };

  const handleBack = () => {
    setError('');
    setSuccess('');
    setActiveStep((previous) => Math.max(previous - 1, 0));
  };

  const stepSaved = status ? isSetupStepSaved(status, activeStep) : false;
  const stepStatusText = activeStep === 0 ? '' : stepSaved ? translateText('已保存') : translateText('未保存');
  const nextDisabled = loading || initialLoading || (activeStep === 4 && status?.initialized === true);

  if (initialLoading) {
    return (
      <Box sx={{ p: 3, display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '50vh' }}>
        <CircularProgress />
      </Box>
    );
  }

  const renderMessages = () => (
    <>
      {activeStep > 0 && (
        <Alert severity={stepSaved ? 'success' : 'info'} sx={{ mt: 2 }}>
          当前步骤状态：{stepStatusText}
        </Alert>
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
    </>
  );

  const renderStepContent = () => {
    switch (activeStep) {
      case 0:
        return (
          <Paper sx={{ p: 3, height: '300px', display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center' }}>
            <Typography variant="h3" gutterBottom align="center" sx={{ mb: 3 }}>
              {translateText('欢迎使用 Tairitsu')}
            </Typography>
            <Typography variant="body1" align="center" sx={{ maxWidth: 520, mb: 4 }}>
              {translateText('本向导会依次完成 ZeroTier 控制器连接、SQLite 配置、首个管理员创建，以及运行态切换。')}
            </Typography>
            <Button
              variant="contained"
              startIcon={<ArrowForwardIcon />}
              onClick={() => { void runStep(); }}
              disabled={loading}
            >
              {translateText('开始初始化')}
            </Button>
            {renderMessages()}
          </Paper>
        );
      case 1:
        return (
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              {translateText('配置 ZeroTier 控制器')}
            </Typography>
            <Typography variant="body1" paragraph>
              {translateText('这一步会测试连接并保存配置。刷新页面后，已保存的控制器地址和 token 路径会自动回显。')}
            </Typography>
            <TextField
              margin="normal"
              required
              fullWidth
              id="controllerUrl"
              label={translateText('ZeroTier 控制器 URL')}
              name="controllerUrl"
              autoComplete="url"
              value={ztConfig.controllerUrl}
              onChange={(event) => setZtConfig((previous) => ({ ...previous, controllerUrl: event.target.value }))}
              disabled={loading}
            />
            <TextField
              margin="normal"
              required
              fullWidth
              id="tokenPath"
              label={translateText('认证令牌文件路径')}
              name="tokenPath"
              autoComplete="file-path"
              value={ztConfig.tokenPath}
              onChange={(event) => setZtConfig((previous) => ({ ...previous, tokenPath: event.target.value }))}
              disabled={loading}
              helperText={translateText('例如 /var/lib/zerotier-one/authtoken.secret')}
            />
            {status?.ztStatus && (
              <Alert severity={status.ztStatus.online ? 'success' : 'warning'} sx={{ mt: 2 }}>
                {translateText('当前控制器状态：')}{status.ztStatus.online ? translateText('在线') : translateText('离线')}{status.ztStatus.address ? ` · ${status.ztStatus.address}` : ''}
              </Alert>
            )}
            {renderMessages()}
          </Paper>
        );
      case 2:
        return (
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              {translateText('配置数据库')}
            </Typography>
            <Typography variant="body1" paragraph>
              {translateText(setupWizardDatabaseStepCopy.description)}
            </Typography>
            <Alert severity="info" sx={{ mb: 2 }}>
              {translateText(setupWizardDatabaseStepCopy.supportAlert)}
            </Alert>
            <TextField
              margin="normal"
              fullWidth
              id="type"
              label={translateText('数据库类型')}
              value={dbConfig.type}
              disabled
              helperText={translateText(setupWizardDatabaseStepCopy.databaseTypeHelperText)}
            />
            <TextField
              margin="normal"
              fullWidth
              id="path"
              label={translateText('SQLite 文件路径')}
              name="path"
              value={dbConfig.path}
              onChange={(event) => setDbConfig((previous) => ({ ...previous, path: event.target.value }))}
              disabled={loading}
              helperText={translateText(setupWizardDatabaseStepCopy.databasePathHelperText)}
            />
            {renderMessages()}
          </Paper>
        );
      case 3:
        return (
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              {translateText('创建管理员账户')}
            </Typography>
            <Typography variant="body1" paragraph>
              {translateText('这一步仅用于首次部署。若你已经创建过管理员，刷新后会自动进入完成步骤，不会重复重置数据库。')}
            </Typography>
            <TextField
              margin="normal"
              required
              fullWidth
              id="username"
              label={translateText('用户名')}
              name="username"
              autoComplete="username"
              value={adminData.username}
              onChange={(event) => setAdminData((previous) => ({ ...previous, username: event.target.value }))}
              disabled={loading || !!status?.hasAdmin}
            />
            <TextField
              margin="normal"
              required
              fullWidth
              name="password"
              label={translateText('密码')}
              type="password"
              id="password"
              autoComplete="new-password"
              value={adminData.password}
              onChange={(event) => setAdminData((previous) => ({ ...previous, password: event.target.value }))}
              disabled={loading || !!status?.hasAdmin}
              helperText={translateText('密码长度至少为6位')}
            />
            {status?.adminCreationPrepared && !status?.hasAdmin && (
              <Alert severity="info" sx={{ mt: 2 }}>
                {translateText('首个管理员创建环境已准备完成，可以直接提交账号信息。')}
              </Alert>
            )}
            {renderMessages()}
          </Paper>
        );
      case 4:
        return (
          <Paper sx={{ p: 3 }}>
            <Typography variant="h5" gutterBottom>
              {translateText('完成设置')}
            </Typography>
            <Typography variant="body1" paragraph>
              {translateText('请确认以下信息。点击"完成初始化"后，系统会校验关键配置并切换到运行态。')}
            </Typography>
            <Box component="ul" sx={{ pl: 3, mb: 0 }}>
              <li>{translateText('ZeroTier 控制器：')}{status?.zeroTierConfig?.controllerUrl || ztConfig.controllerUrl}</li>
              <li>{translateText('SQLite 路径：')}{status?.databaseConfig?.path || dbConfig.path || 'data/tairitsu.db'}</li>
              <li>{translateText('首个管理员：')}{status?.adminUsername || adminData.username || translateText('尚未创建')}</li>
            </Box>
            {renderMessages()}
          </Paper>
        );
      default:
        return <div>{translateText('未知步骤')}</div>;
    }
  };

  return (
    <Container component="main" maxWidth="md" sx={{ mt: 8 }}>
      <Paper elevation={3} sx={{ p: 4, borderRadius: 2 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
          <Box sx={{ flex: 1 }} />
          <Typography component="h1" variant="h4" align="center" sx={{ flex: 2 }}>
            {translateText('Tairitsu 初始化向导')}
          </Typography>
          <Box sx={{ flex: 1, display: 'flex', justifyContent: 'flex-end' }}>
            <FormControl size="small" sx={{ minWidth: 120 }}>
              <InputLabel id="setup-language-label">Language</InputLabel>
              <Select<LanguagePreference>
                labelId="setup-language-label"
                label="Language"
                value={preference}
                onChange={(event) => setPreference(event.target.value)}
              >
                <MenuItem value="system">System</MenuItem>
                <MenuItem value="en">English</MenuItem>
                <MenuItem value="zh-CN">简体中文</MenuItem>
              </Select>
            </FormControl>
          </Box>
        </Box>
        <Stepper activeStep={activeStep} sx={{ mb: 4 }}>
          {steps.map((label, index) => (
            <Step key={index}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>
        <div>{renderStepContent()}</div>
        {activeStep !== 0 && (
          <Box sx={{ mt: 4, display: 'flex', justifyContent: 'space-between' }}>
            <Button
              variant="outlined"
              onClick={handleBack}
              disabled={loading || activeStep === 4}
            >
              {translateText('返回')}
            </Button>
            <Button
              variant="contained"
              color="primary"
              onClick={() => { void runStep(); }}
              disabled={nextDisabled}
            >
              {loading ? (
                <CircularProgress size={24} />
              ) : activeStep === 1 ? (
                translateText('测试并保存')
              ) : activeStep === 2 ? (
                translateText('保存数据库配置')
              ) : activeStep === 3 ? (
                translateText('创建首个管理员')
              ) : activeStep === 4 ? (
                translateText('完成初始化')
              ) : (
                translateText('下一步')
              )}
            </Button>
          </Box>
        )}
      </Paper>
    </Container>
  );
}

export { setupCompletedEvent };
export default SetupWizard;
