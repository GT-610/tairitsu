import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Button,
  Checkbox,
  FormControlLabel,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  ListItemSecondaryAction,
  Chip,
  Alert,
  CircularProgress,
  Divider,
  Paper,
  Stack
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import { Link as RouterLink } from 'react-router-dom';
import { ImportableNetworkSummary, networkAPI } from '../services/api';
import { getErrorMessage } from '../services/errors';

function ImportNetwork() {
  const [networks, setNetworks] = useState<ImportableNetworkSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [importing, setImporting] = useState(false);
  const [selectedNetworks, setSelectedNetworks] = useState<Set<string>>(new Set());
  const [error, setError] = useState<string>('');
  const [successMessage, setSuccessMessage] = useState<string>('');
  const [importResult, setImportResult] = useState<{
    importedIds: string[];
    failed: Array<{ network_id: string; reason: string }>;
  } | null>(null);

  useEffect(() => {
    void fetchImportableNetworks();
  }, []);

  const fetchImportableNetworks = async () => {
    setLoading(true);
    setError('');
    try {
      const response = await networkAPI.getImportableNetworks();
      setNetworks(Array.isArray(response.data) ? response.data : []);
    } catch (err: unknown) {
      setError(getErrorMessage(err, '获取可导入网络列表失败'));
      setNetworks([]);
    } finally {
      setLoading(false);
    }
  };

  const handleToggle = (networkId: string) => {
    const newSelected = new Set(selectedNetworks);
    if (newSelected.has(networkId)) {
      newSelected.delete(networkId);
    } else {
      newSelected.add(networkId);
    }
    setSelectedNetworks(newSelected);
  };

  const handleSelectAll = () => {
    const importableNetworks = networks.filter(n => n.is_importable);
    if (selectedNetworks.size === importableNetworks.length) {
      setSelectedNetworks(new Set());
    } else {
      setSelectedNetworks(new Set(importableNetworks.map(n => n.network_id)));
    }
  };

  const handleImport = async () => {
    if (selectedNetworks.size === 0) return;

    const importCount = selectedNetworks.size;
    setImporting(true);
    setError('');
    setSuccessMessage('');
    setImportResult(null);
    try {
      const response = await networkAPI.importNetworks(Array.from(selectedNetworks));
      const importedIds = Array.isArray(response.data.imported_ids) ? response.data.imported_ids : [];
      const failed = Array.isArray(response.data.failed) ? response.data.failed : [];

      setSelectedNetworks(new Set());
      setImportResult({
        importedIds,
        failed,
      });

      if (importedIds.length === importCount && failed.length === 0) {
        setSuccessMessage(`已成功导入 ${importedIds.length} 个网络，列表已刷新。`);
      } else if (importedIds.length > 0) {
        setSuccessMessage(`已导入 ${importedIds.length} 个网络，另有 ${failed.length} 个网络未完成导入。`);
      } else {
        setError(response.data.message || '未成功导入任何网络');
      }

      await fetchImportableNetworks();
    } catch (err: unknown) {
      setError(getErrorMessage(err, '导入网络失败'));
    } finally {
      setImporting(false);
    }
  };

  const importableNetworks = networks.filter(n => n.is_importable);
  const nonImportableNetworks = networks.filter(n => !n.is_importable);

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          导入 ZeroTier 网络
        </Typography>
        <Stack direction="row" spacing={1} alignItems="center">
          <Chip
            label={`${importableNetworks.length}个可导入`}
            color="primary"
            size="small"
          />
          <Button
            variant="outlined"
            size="small"
            startIcon={<RefreshIcon />}
            onClick={() => { void fetchImportableNetworks(); }}
            disabled={loading || importing}
          >
            刷新
          </Button>
        </Stack>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      {successMessage && (
        <Alert
          severity="success"
          sx={{ mb: 3 }}
          icon={<CheckCircleOutlineIcon fontSize="inherit" />}
          onClose={() => setSuccessMessage('')}
          action={(
            <Button component={RouterLink} to="/networks" color="inherit" size="small">
              查看网络
            </Button>
          )}
        >
          {successMessage}
        </Alert>
      )}

      {importResult && importResult.failed.length > 0 && (
        <Alert severity={importResult.importedIds.length > 0 ? 'warning' : 'error'} sx={{ mb: 3 }}>
          <Typography variant="body2" sx={{ fontWeight: 600, mb: 1 }}>
            导入结果摘要
          </Typography>
          <Typography variant="body2" sx={{ mb: 1 }}>
            成功 {importResult.importedIds.length} 个，失败 {importResult.failed.length} 个。
          </Typography>
          <List dense sx={{ py: 0 }}>
            {importResult.failed.map((item) => (
              <ListItem key={`${item.network_id}-${item.reason}`} disablePadding sx={{ py: 0.25 }}>
                <ListItemText
                  primary={item.network_id}
                  secondary={item.reason}
                  primaryTypographyProps={{ variant: 'body2' }}
                  secondaryTypographyProps={{ variant: 'caption' }}
                />
              </ListItem>
            ))}
          </List>
        </Alert>
      )}

      <Alert severity="info" sx={{ mb: 3 }}>
        该页面用于把控制器中已存在、但尚未在 Tairitsu 中登记的网络纳入当前管理员账号。导入前请先确认网络归属关系，避免误接管其他已使用中的网络。
      </Alert>

      <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          以下是 ZeroTier 控制器上存在但尚未在 Tairitsu 中登记或无主的网络。选择要导入的网络，它们将被登记为您（当前管理员）所有。
        </Typography>

        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
          </Box>
        ) : networks.length === 0 ? (
          <Box sx={{ py: 5, textAlign: 'center' }}>
            <Typography sx={{ color: 'text.secondary', mb: 2 }}>
              没有找到可导入的网络。所有网络都已在 Tairitsu 中登记且有所有者。
            </Typography>
            <Button
              variant="outlined"
              startIcon={<RefreshIcon />}
              onClick={() => { void fetchImportableNetworks(); }}
            >
              重新检查
            </Button>
          </Box>
        ) : (
          <>
            {importableNetworks.length > 0 && (
              <>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={selectedNetworks.size === importableNetworks.length}
                        indeterminate={selectedNetworks.size > 0 && selectedNetworks.size < importableNetworks.length}
                        onChange={handleSelectAll}
                      />
                    }
                    label="全选可导入的网络"
                  />
                  <Typography variant="body2" color="text.secondary">
                    ({selectedNetworks.size}/{importableNetworks.length})
                  </Typography>
                </Box>

                <List dense>
                  {importableNetworks.map((item) => (
                    <ListItem
                      key={item.network_id}
                      disablePadding
                      sx={{ mb: 0.5 }}
                    >
                      <ListItemButton
                        onClick={() => handleToggle(item.network_id)}
                        sx={{ borderRadius: 1, bgcolor: 'action.hover' }}
                      >
                        <ListItemIcon>
                          <Checkbox
                            edge="start"
                            checked={selectedNetworks.has(item.network_id)}
                            tabIndex={-1}
                          />
                        </ListItemIcon>
                        <ListItemText
                          primary={item.network_id}
                          secondary={item.reason}
                        />
                        <ListItemSecondaryAction>
                          <Chip
                            label={item.reason}
                            size="small"
                            color="success"
                            variant="outlined"
                          />
                        </ListItemSecondaryAction>
                      </ListItemButton>
                    </ListItem>
                  ))}
                </List>
              </>
            )}

            {nonImportableNetworks.length > 0 && (
              <>
                <Divider sx={{ my: 2 }} />
                <Typography variant="subtitle2" color="text.secondary" sx={{ mb: 1 }}>
                  不可导入的网络
                </Typography>
                <List dense>
                  {nonImportableNetworks.map((item) => (
                    <ListItem
                      key={item.network_id}
                      disablePadding
                      sx={{ mb: 0.5, opacity: 0.7 }}
                    >
                      <ListItemButton disabled>
                        <ListItemIcon>
                          <Checkbox
                            edge="start"
                            checked={false}
                            tabIndex={-1}
                          />
                        </ListItemIcon>
                        <ListItemText
                          primary={item.network_id}
                          secondary={item.reason}
                        />
                        <ListItemSecondaryAction>
                          <Chip
                            label={item.reason}
                            size="small"
                            variant="outlined"
                          />
                        </ListItemSecondaryAction>
                      </ListItemButton>
                    </ListItem>
                  ))}
                </List>
              </>
            )}

            <Box sx={{ mt: 3, display: 'flex', justifyContent: 'flex-end', gap: 2 }}>
              <Button
                variant="contained"
                color="primary"
                onClick={() => { void handleImport(); }}
                disabled={selectedNetworks.size === 0 || importing}
              >
                {importing ? (
                  <>
                    <CircularProgress size={20} sx={{ mr: 1 }} />
                    导入中...
                  </>
                ) : (
                  `导入所选网络 (${selectedNetworks.size})`
                )}
              </Button>
              <Button component={RouterLink} to="/networks" variant="outlined">
                返回网络列表
              </Button>
            </Box>
          </>
        )}
      </Paper>
    </Box>
  );
}

export default ImportNetwork;
