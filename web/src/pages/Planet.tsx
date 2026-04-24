import { useEffect, useState } from 'react'
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  IconButton,
  List,
  ListItem,
  ListItemSecondaryAction,
  TextField,
  Typography,
} from '@mui/material'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import DownloadIcon from '@mui/icons-material/Download'
import { planetAPI, type GeneratePlanetResponse } from '../services/api'
import { getErrorMessage } from '../services/errors'
import { getPlanetDownloadName, normalizePlanetEndpoints, validatePlanetEndpoints } from '../utils/planet'

interface EndpointDraft {
  id: string
  value: string
}

interface PlanetResultState extends GeneratePlanetResponse {
  endpoint_count: number
}

const defaultIdentityPath = '/var/lib/zerotier-one'

function PlanetGenerator() {
  const [loadingIdentity, setLoadingIdentity] = useState(true)
  const [generating, setGenerating] = useState(false)
  const [identityPublic, setIdentityPublic] = useState('')
  const [identityPath, setIdentityPath] = useState(defaultIdentityPath)
  const [resolvedIdentityPath, setResolvedIdentityPath] = useState('')
  const [comments, setComments] = useState('')
  const [endpoints, setEndpoints] = useState<EndpointDraft[]>([{ id: '1', value: '' }])
  const [message, setMessage] = useState<{ severity: 'success' | 'error'; text: string } | null>(null)
  const [generatedPlanet, setGeneratedPlanet] = useState<PlanetResultState | null>(null)

  useEffect(() => {
    void loadIdentity(defaultIdentityPath)
  }, [])

  const loadIdentity = async (pathOverride?: string) => {
    const nextPath = pathOverride ?? identityPath
    try {
      setLoadingIdentity(true)
      setMessage(null)
      const response = await planetAPI.getIdentity(nextPath)
      setIdentityPublic(response.data.identity_public)
      setResolvedIdentityPath(response.data.identity_path)
    } catch (error: unknown) {
      setIdentityPublic('')
      setResolvedIdentityPath('')
      setGeneratedPlanet(null)
      setMessage({ severity: 'error', text: getErrorMessage(error, '加载 identity.public 失败') })
    } finally {
      setLoadingIdentity(false)
    }
  }

  const handleAddEndpoint = () => {
    setEndpoints((previous) => [...previous, { id: Date.now().toString(), value: '' }])
  }

  const handleRemoveEndpoint = (id: string) => {
    setEndpoints((previous) => (
      previous.length > 1 ? previous.filter((endpoint) => endpoint.id !== id) : previous
    ))
  }

  const handleEndpointChange = (id: string, value: string) => {
    setEndpoints((previous) => previous.map((endpoint) => (
      endpoint.id === id ? { ...endpoint, value } : endpoint
    )))
  }

  const handleGenerate = async () => {
    if (!identityPublic.trim()) {
      setMessage({ severity: 'error', text: '请先成功加载 identity.public' })
      setGeneratedPlanet(null)
      return
    }

    const endpointValues = endpoints.map((endpoint) => endpoint.value)
    const endpointError = validatePlanetEndpoints(endpointValues)
    if (endpointError) {
      setMessage({ severity: 'error', text: endpointError })
      setGeneratedPlanet(null)
      return
    }

    const normalizedEndpoints = normalizePlanetEndpoints(endpointValues)

    try {
      setGenerating(true)
      setMessage(null)

      const response = await planetAPI.generatePlanet({
        identity_public: identityPublic.trim(),
        endpoints: normalizedEndpoints,
        comments: comments.trim(),
      })

      setGeneratedPlanet({
        ...response.data,
        endpoint_count: normalizedEndpoints.length,
      })
      setMessage({ severity: 'success', text: response.data.message })
    } catch (error: unknown) {
      setGeneratedPlanet(null)
      setMessage({ severity: 'error', text: getErrorMessage(error, '生成 Planet 失败') })
    } finally {
      setGenerating(false)
    }
  }

  const handleDownloadPlanet = () => {
    if (!generatedPlanet) {
      return
    }

    const blob = new Blob([new Uint8Array(generatedPlanet.planet_data)], { type: 'application/octet-stream' })
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = getPlanetDownloadName(generatedPlanet.download_name)
    document.body.appendChild(anchor)
    anchor.click()
    document.body.removeChild(anchor)
    URL.revokeObjectURL(url)
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          生成 Planet（实验性）
        </Typography>
      </Box>

      {message && (
        <Alert severity={message.severity} sx={{ mb: 3 }} onClose={() => setMessage(null)}>
          {message.text}
        </Alert>
      )}

      <Alert severity="warning" sx={{ mb: 3 }}>
        该能力当前保持实验性状态。生成的 planet 文件请仅在隔离环境中验证，并在替换到控制器前自行完成备份。
      </Alert>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            身份加载
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            输入 ZeroTier 数据目录后，系统会读取其中的 `identity.public`，并将其作为当前 Planet 的唯一 root identity。
          </Typography>

          <TextField
            label="ZeroTier 数据目录路径"
            fullWidth
            value={identityPath}
            onChange={(event) => setIdentityPath(event.target.value)}
            sx={{ mb: 2 }}
            helperText="默认为 /var/lib/zerotier-one"
            disabled={loadingIdentity || generating}
          />

          <Button
            variant="outlined"
            onClick={() => { void loadIdentity() }}
            disabled={loadingIdentity || generating}
            sx={{ mb: 2 }}
          >
            {loadingIdentity ? '读取中...' : '读取身份'}
          </Button>

          {loadingIdentity ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
              <CircularProgress size={24} />
            </Box>
          ) : (
            <>
              {resolvedIdentityPath && (
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  实际读取路径：{resolvedIdentityPath}
                </Typography>
              )}
              <TextField
                label="identity.public"
                fullWidth
                multiline
                rows={2}
                value={identityPublic}
                helperText="成功读取后会显示当前 root identity"
                placeholder="格式：10hexdigits:0:publicKey"
                InputProps={{ readOnly: true }}
              />
            </>
          )}
        </CardContent>
      </Card>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Planet 配置
          </Typography>

          <TextField
            label="备注"
            fullWidth
            value={comments}
            onChange={(event) => setComments(event.target.value)}
            sx={{ mb: 2 }}
            placeholder="例如：Private Planet - example.com"
            helperText="可选，仅用于标识这次生成的用途"
            disabled={generating}
          />

          <Typography variant="subtitle2" sx={{ mb: 1 }}>
            端点列表
          </Typography>

          <List>
            {endpoints.map((endpoint, index) => (
              <ListItem key={endpoint.id} sx={{ px: 0 }}>
                <TextField
                  label={`端点 ${index + 1}`}
                  placeholder="IP/Port，例如：203.0.113.1/9993"
                  value={endpoint.value}
                  onChange={(event) => handleEndpointChange(endpoint.id, event.target.value)}
                  fullWidth
                  helperText="支持 IPv4 和 IPv6，格式统一为 IP/Port"
                  disabled={generating}
                />
                <ListItemSecondaryAction>
                  {endpoints.length > 1 && (
                    <IconButton edge="end" onClick={() => handleRemoveEndpoint(endpoint.id)} color="error" disabled={generating}>
                      <DeleteIcon />
                    </IconButton>
                  )}
                </ListItemSecondaryAction>
              </ListItem>
            ))}
          </List>

          <Button startIcon={<AddIcon />} onClick={handleAddEndpoint} sx={{ mt: 1 }} disabled={generating}>
            添加端点
          </Button>
        </CardContent>
      </Card>

      <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
        <Button
          variant="contained"
          onClick={() => { void handleGenerate() }}
          disabled={loadingIdentity || generating || !identityPublic.trim()}
          startIcon={generating ? <CircularProgress size={20} color="inherit" /> : undefined}
        >
          {generating ? '生成中...' : '生成 Planet'}
        </Button>
      </Box>

      {generatedPlanet && (
        <Card sx={{ mb: 3 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              生成结果
            </Typography>

            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(3, 1fr)' }, gap: 2, mb: 3 }}>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  Planet ID
                </Typography>
                <Typography variant="body1">{generatedPlanet.planet_id}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  生成时间
                </Typography>
                <Typography variant="body1">{new Date(generatedPlanet.birth_time).toLocaleString()}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  端点数量
                </Typography>
                <Typography variant="body1">{generatedPlanet.endpoint_count}</Typography>
              </Box>
            </Box>

            <Button
              variant="contained"
              color="success"
              startIcon={<DownloadIcon />}
              onClick={handleDownloadPlanet}
            >
              下载 Planet
            </Button>
          </CardContent>
        </Card>
      )}
    </Box>
  )
}

export default PlanetGenerator
