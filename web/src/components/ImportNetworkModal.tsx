import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
  Box,
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
  Divider
} from '@mui/material';
import { ImportableNetworkSummary } from '../services/api';

interface ImportNetworkModalProps {
  open: boolean;
  onClose: () => void;
  onImportComplete: () => void;
}

export function ImportNetworkModal({ open, onClose, onImportComplete }: ImportNetworkModalProps) {
  const [networks, setNetworks] = useState<ImportableNetworkSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [importing, setImporting] = useState(false);
  const [selectedNetworks, setSelectedNetworks] = useState<Set<string>>(new Set());
  const [error, setError] = useState<string>('');

  useEffect(() => {
    if (open) {
      fetchImportableNetworks();
    }
  }, [open]);

  const fetchImportableNetworks = async () => {
    setLoading(true);
    setError('');
    try {
      const response = await fetch('/api/admin/networks/importable', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || sessionStorage.getItem('token')}`
        }
      });
      if (!response.ok) {
        throw new Error('获取可导入网络列表失败');
      }
      const data = await response.json();
      setNetworks(data);
    } catch (err: any) {
      setError(err.message || '获取可导入网络列表失败');
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

    setImporting(true);
    setError('');
    try {
      const response = await fetch('/api/admin/networks/import', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token') || sessionStorage.getItem('token')}`
        },
        body: JSON.stringify({ network_ids: Array.from(selectedNetworks) })
      });

      if (!response.ok) {
        throw new Error('导入网络失败');
      }

      onImportComplete();
      onClose();
    } catch (err: any) {
      setError(err.message || '导入网络失败');
    } finally {
      setImporting(false);
    }
  };

  const importableNetworks = networks.filter(n => n.is_importable);
  const nonImportableNetworks = networks.filter(n => !n.is_importable);

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Typography variant="h6">导入ZeroTier网络</Typography>
          <Chip
            label={`${importableNetworks.length}个可导入`}
            color="primary"
            size="small"
            sx={{ ml: 1 }}
          />
        </Box>
      </DialogTitle>
      <DialogContent dividers>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          以下是ZeroTier控制器上存在但尚未在Tairitsu中登记或无主的网络。选择要导入的网络，它们将被登记为您（当前管理员）所有。
        </Typography>

        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
          </Box>
        ) : networks.length === 0 ? (
          <Typography sx={{ py: 4, textAlign: 'center', color: 'text.secondary' }}>
            没有找到可导入的网络。所有网络都已在Tairitsu中登记且有所有者。
          </Typography>
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
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={importing}>
          取消
        </Button>
        <Button
          variant="contained"
          color="primary"
          onClick={handleImport}
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
      </DialogActions>
    </Dialog>
  );
}
