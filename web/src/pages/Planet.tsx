import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Button,
  Alert,
  TextField,
  CircularProgress,
  List,
  ListItem,
  ListItemSecondaryAction,
  IconButton,
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import DownloadIcon from '@mui/icons-material/Download';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import { planetAPI } from '../services/api';

interface Endpoint {
  id: string;
  value: string;
}

function PlanetGenerator() {
  const [loading, setLoading] = useState<boolean>(true);
  const [generating, setGenerating] = useState<boolean>(false);
  const [identityPublic, setIdentityPublic] = useState<string>('');
  const [identityPath, setIdentityPath] = useState<string>('/var/lib/zerotier-one');
  const [comments, setComments] = useState<string>('');
  const [endpoints, setEndpoints] = useState<Endpoint[]>([
    { id: '1', value: '' }
  ]);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [generatedPlanet, setGeneratedPlanet] = useState<{
    cHeader: string;
    planetData: number[];
  } | null>(null);
  const [copyDialogOpen, setCopyDialogOpen] = useState(false);

  useEffect(() => {
    loadIdentity();
  }, []);

  const loadIdentity = async () => {
    try {
      setLoading(true);
      const response = await planetAPI.getIdentity(identityPath);
      if (response.data.success) {
        setIdentityPublic(response.data.identityPublic || '');
      } else {
        setMessage({ type: 'error', text: response.data.message });
      }
    } catch (error: any) {
      setMessage({
        type: 'error',
        text: error.response?.data?.message || 'Failed to load identity.public'
      });
    } finally {
      setLoading(false);
    }
  };

  const handleAddEndpoint = () => {
    setEndpoints([...endpoints, { id: Date.now().toString(), value: '' }]);
  };

  const handleRemoveEndpoint = (id: string) => {
    if (endpoints.length > 1) {
      setEndpoints(endpoints.filter(ep => ep.id !== id));
    }
  };

  const handleEndpointChange = (id: string, value: string) => {
    setEndpoints(endpoints.map(ep =>
      ep.id === id ? { ...ep, value } : ep
    ));
  };

  const validateEndpoints = (): string | null => {
    const filledEndpoints = endpoints.filter(ep => ep.value.trim() !== '');
    if (filledEndpoints.length === 0) {
      return 'At least one endpoint is required';
    }
    for (const ep of filledEndpoints) {
      const value = ep.value.trim();
      if (!value.includes('/')) {
        return `Invalid endpoint format: ${value}. Expected format: IP/Port`;
      }
      const parts = value.split('/');
      const port = parseInt(parts[1], 10);
      if (isNaN(port) || port < 1 || port > 65535) {
        return `Invalid port number: ${parts[1]}`;
      }
    }
    return null;
  };

  const handleGenerate = async () => {
    if (!identityPublic) {
      setMessage({ type: 'error', text: 'identity.public is required' });
      return;
    }

    const endpointError = validateEndpoints();
    if (endpointError) {
      setMessage({ type: 'error', text: endpointError });
      return;
    }

    const filledEndpoints = endpoints
      .filter(ep => ep.value.trim() !== '')
      .map(ep => ep.value.trim());

    try {
      setGenerating(true);
      setMessage(null);

      const response = await planetAPI.generatePlanet({
        identityPublic,
        endpoints: filledEndpoints,
        comments
      });

      if (response.data.success) {
        setGeneratedPlanet({
          cHeader: response.data.cHeader || '',
          planetData: Array.from(response.data.planetData || [])
        });
        setMessage({ type: 'success', text: 'Planet file generated successfully!' });
      } else {
        setMessage({ type: 'error', text: response.data.message || 'Failed to generate planet' });
      }
    } catch (error: any) {
      setMessage({
        type: 'error',
        text: error.response?.data?.message || 'Failed to generate planet'
      });
    } finally {
      setGenerating(false);
    }
  };

  const handleDownloadPlanet = () => {
    if (!generatedPlanet?.planetData) return;

    const blob = new Blob([new Uint8Array(generatedPlanet.planetData)], { type: 'application/octet-stream' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'planet';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const handleCopyCHeader = () => {
    if (generatedPlanet?.cHeader) {
      navigator.clipboard.writeText(generatedPlanet.cHeader);
      setMessage({ type: 'success', text: 'C header copied to clipboard!' });
    }
    setCopyDialogOpen(false);
  };

  const openCHeaderDialog = () => {
    setCopyDialogOpen(true);
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          Planet 生成
        </Typography>
      </Box>

      {message && (
        <Alert severity={message.type} sx={{ mb: 3 }} onClose={() => setMessage(null)}>
          {message.text}
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          <Alert severity="warning" sx={{ mb: 3 }}>
            <Typography variant="subtitle2" gutterBottom>
              实验性功能
            </Typography>
            <Typography variant="body2">
              此功能仍处于实验性阶段，不保证可用。生成的 planet 文件可能存在兼容性问题，
              请在生产环境使用前进行充分测试。使用此功能前，请确保已备份原有的 planet 文件。
            </Typography>
          </Alert>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Alert severity="warning" sx={{ mb: 2 }}>
                <Typography variant="body2">
                  生成自定义 planet 会创建一个独立的网络根服务器。替换 planet 文件后，
                  所有节点需要重新连接到此新根服务器，原有的网络和成员关系不会受影响，
                  但请确保所有节点使用相同的 planet 文件。
                </Typography>
              </Alert>

              <Typography variant="h6" gutterBottom>
                控制器身份信息
              </Typography>

              <TextField
                label="ZeroTier 数据目录路径"
                fullWidth
                value={identityPath}
                onChange={(e) => setIdentityPath(e.target.value)}
                sx={{ mb: 2 }}
                helperText="默认为 /var/lib/zerotier-one"
              />

              <Button
                variant="outlined"
                onClick={loadIdentity}
                sx={{ mb: 2 }}
              >
                重新加载身份
              </Button>

              <TextField
                label="identity.public"
                fullWidth
                multiline
                rows={2}
                value={identityPublic}
                onChange={(e) => setIdentityPublic(e.target.value)}
                helperText="从 identity.public 文件读取的节点身份"
                placeholder="格式: 10hexdigits:0:publicKey"
              />
            </CardContent>
          </Card>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Planet 配置
              </Typography>

              <TextField
                label="备注 (Comments)"
                fullWidth
                value={comments}
                onChange={(e) => setComments(e.target.value)}
                sx={{ mb: 2 }}
                placeholder="例如: My Custom Planet - example.com"
                helperText="可选，仅用于标识这个 planet 的用途"
              />

              <Typography variant="subtitle2" sx={{ mb: 1 }}>
                端点 (Endpoints)
              </Typography>

              <List>
                {endpoints.map((endpoint, index) => (
                  <ListItem key={endpoint.id} sx={{ px: 0 }}>
                    <TextField
                      label={`端点 ${index + 1}`}
                      placeholder="IP/Port 例如: 192.168.1.1/9993"
                      value={endpoint.value}
                      onChange={(e) => handleEndpointChange(endpoint.id, e.target.value)}
                      fullWidth
                      helperText="格式: IP地址/端口号"
                    />
                    <ListItemSecondaryAction>
                      {endpoints.length > 1 && (
                        <IconButton
                          edge="end"
                          onClick={() => handleRemoveEndpoint(endpoint.id)}
                          color="error"
                        >
                          <DeleteIcon />
                        </IconButton>
                      )}
                    </ListItemSecondaryAction>
                  </ListItem>
                ))}
              </List>

              <Button
                startIcon={<AddIcon />}
                onClick={handleAddEndpoint}
                sx={{ mt: 1 }}
              >
                添加端点
              </Button>

              <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                建议至少添加一个 IPv4 和一个 IPv6 端点。每个端点格式为 IP/Port，例如: 203.0.113.1/9993
              </Typography>
            </CardContent>
          </Card>

          <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
            <Button
              variant="contained"
              onClick={handleGenerate}
              disabled={generating || !identityPublic}
              startIcon={generating ? <CircularProgress size={20} color="inherit" /> : undefined}
            >
              {generating ? '生成中...' : '生成 Planet'}
            </Button>
          </Box>

          {generatedPlanet && (
            <Card sx={{ mb: 3 }}>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  生成的 Planet 文件
                </Typography>

                <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
                  <Button
                    variant="contained"
                    color="success"
                    startIcon={<DownloadIcon />}
                    onClick={handleDownloadPlanet}
                  >
                    下载 Planet 文件
                  </Button>
                  <Button
                    variant="outlined"
                    startIcon={<ContentCopyIcon />}
                    onClick={openCHeaderDialog}
                  >
                    复制 C 头文件
                  </Button>
                </Box>

                <Paper variant="outlined" sx={{ p: 2, bgcolor: 'grey.50' }}>
                  <Typography variant="body2" component="pre" sx={{ fontFamily: 'monospace', fontSize: '0.75rem', overflow: 'auto' }}>
                    {generatedPlanet.cHeader}
                  </Typography>
                </Paper>

                <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                  提示: 将下载的 planet 文件复制到 ZeroTier 控制器的数据目录 (例如: /var/lib/zerotier-one/planet)，
                  然后重启 ZeroTier 服务使更改生效。
                </Typography>
              </CardContent>
            </Card>
          )}

          <Dialog open={copyDialogOpen} onClose={() => setCopyDialogOpen(false)} maxWidth="md" fullWidth>
            <DialogTitle>复制 C 头文件</DialogTitle>
            <DialogContent>
              <Paper variant="outlined" sx={{ p: 2, bgcolor: 'grey.50', maxHeight: 400, overflow: 'auto' }}>
                <Typography component="pre" sx={{ fontFamily: 'monospace', fontSize: '0.75rem', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                  {generatedPlanet?.cHeader}
                </Typography>
              </Paper>
            </DialogContent>
            <DialogActions>
              <Button onClick={() => setCopyDialogOpen(false)}>关闭</Button>
              <Button variant="contained" onClick={handleCopyCHeader}>
                复制到剪贴板
              </Button>
            </DialogActions>
          </Dialog>
        </>
      )}
    </Box>
  );
}

export default PlanetGenerator;
