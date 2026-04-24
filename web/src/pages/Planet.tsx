import { useState } from 'react'
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
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
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import DownloadIcon from '@mui/icons-material/Download'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import KeyIcon from '@mui/icons-material/Key'
import SyncIcon from '@mui/icons-material/Sync'
import { planetAPI, type GeneratePlanetResponse, type SigningKeysInfoResponse } from '../services/api'
import { getErrorMessage } from '../services/errors'
import {
  findDuplicatePlanetEndpoints,
  getPlanetDownloadName,
  normalizePlanetEndpoints,
  parsePlanetIdentityPublic,
  validatePlanetEndpointValue,
  validatePlanetEndpoints,
} from '../utils/planet'

interface EndpointDraft {
  id: string
  value: string
}

interface PlanetResultState extends GeneratePlanetResponse {
  endpoint_count: number
}

const defaultIdentityPath = '/var/lib/zerotier-one'
const defaultSigningKeyPath = '/var/lib/zerotier-one'

function PlanetGenerator() {
  const [loadingIdentity, setLoadingIdentity] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [loadingSigningKeys, setLoadingSigningKeys] = useState(false)
  const [generatingSigningKeys, setGeneratingSigningKeys] = useState(false)
  const [identityPublic, setIdentityPublic] = useState('')
  const [identityPath, setIdentityPath] = useState(defaultIdentityPath)
  const [resolvedIdentityPath, setResolvedIdentityPath] = useState('')
  const [comments, setComments] = useState('')
  const [endpoints, setEndpoints] = useState<EndpointDraft[]>([{ id: '1', value: '' }])
  const [message, setMessage] = useState<{ severity: 'success' | 'error'; text: string } | null>(null)
  const [identityMessage, setIdentityMessage] = useState<{ severity: 'success' | 'error'; text: string } | null>(null)
  const [advancedModeEnabled, setAdvancedModeEnabled] = useState(false)
  const [signingKeyPath, setSigningKeyPath] = useState(defaultSigningKeyPath)
  const [signingKeysMessage, setSigningKeysMessage] = useState<{ severity: 'success' | 'error'; text: string } | null>(null)
  const [signingKeysInfo, setSigningKeysInfo] = useState<SigningKeysInfoResponse | null>(null)
  const [generatedPlanet, setGeneratedPlanet] = useState<PlanetResultState | null>(null)

  const identitySummary = parsePlanetIdentityPublic(identityPublic)
  const duplicateEndpoints = findDuplicatePlanetEndpoints(endpoints.map((endpoint) => endpoint.value))

  const loadSigningKeysInfo = async (pathOverride?: string) => {
    const nextPath = pathOverride ?? signingKeyPath
    try {
      setLoadingSigningKeys(true)
      setSigningKeysMessage(null)
      const response = await planetAPI.getSigningKeysInfo(nextPath)
      setSigningKeysInfo(response.data)
      setSigningKeysMessage({
        severity: 'success',
        text: response.data.ready
          ? '已检测到可用的 signing keys'
          : '该目录下还没有完整的 signing keys，可按需生成',
      })
    } catch (error: unknown) {
      setSigningKeysInfo(null)
      setSigningKeysMessage({ severity: 'error', text: getErrorMessage(error, '读取 signing keys 状态失败') })
    } finally {
      setLoadingSigningKeys(false)
    }
  }

  const handleGenerateSigningKeys = async () => {
    try {
      setGeneratingSigningKeys(true)
      setSigningKeysMessage(null)
      const response = await planetAPI.generateSigningKeys(signingKeyPath)
      setSigningKeysMessage({ severity: 'success', text: response.data.message })
      await loadSigningKeysInfo(signingKeyPath)
    } catch (error: unknown) {
      setSigningKeysMessage({ severity: 'error', text: getErrorMessage(error, '生成 signing keys 失败') })
    } finally {
      setGeneratingSigningKeys(false)
    }
  }

  const loadIdentity = async (pathOverride?: string) => {
    const nextPath = pathOverride ?? identityPath
    try {
      setLoadingIdentity(true)
      setIdentityMessage(null)
      const response = await planetAPI.getIdentity(nextPath)
      setIdentityPublic(response.data.identity_public)
      setResolvedIdentityPath(response.data.identity_path)
      setIdentityMessage({ severity: 'success', text: 'identity.public 读取成功' })
    } catch (error: unknown) {
      setIdentityPublic('')
      setResolvedIdentityPath('')
      setGeneratedPlanet(null)
      setIdentityMessage({ severity: 'error', text: getErrorMessage(error, '加载 identity.public 失败') })
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
        signing_key_path: advancedModeEnabled ? signingKeyPath.trim() : undefined,
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

  const getEndpointHelperText = (value: string) => {
    const trimmed = value.trim()
    if (!trimmed) {
      return '格式：IP/Port。IPv4 示例：203.0.113.1/9993；IPv6 示例：2001:db8::1/9993'
    }

    if (duplicateEndpoints.has(trimmed)) {
      return '该 stable endpoint 已重复'
    }

    const endpointError = validatePlanetEndpointValue(trimmed)
    if (endpointError) {
      return endpointError
    }

    return '该地址会作为当前 root node 的 stable endpoint 写入 planet'
  }

  const hasEndpointError = (value: string) => {
    const trimmed = value.trim()
    if (!trimmed) {
      return false
    }
    return duplicateEndpoints.has(trimmed) || validatePlanetEndpointValue(trimmed) !== null
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
        该能力当前保持实验性状态。请先在隔离环境中验证生成结果，并在替换控制器文件前完成备份。
      </Alert>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            身份加载
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            这里会从给定目录读取 `identity.public`。读取成功后，它会作为当前 Planet 的唯一 root identity。
          </Typography>

          {identityMessage && (
            <Alert severity={identityMessage.severity} sx={{ mb: 2 }} onClose={() => setIdentityMessage(null)}>
              {identityMessage.text}
            </Alert>
          )}

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
                error={Boolean(identityPublic) && !identitySummary}
                helperText={identityPublic
                  ? identitySummary
                    ? '已读取真实 identity.public，可继续填写 Planet 配置'
                    : 'identity.public 格式无效，应为 10 位地址 + :0: + 128 位公钥'
                  : '成功读取后会显示当前 root identity'}
                placeholder="格式：10hexdigits:0:publicKey"
                InputProps={{ readOnly: true }}
              />

              {identitySummary && (
                <Box sx={{ mt: 2, display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(3, 1fr)' }, gap: 2 }}>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      节点地址
                    </Typography>
                    <Typography variant="body1">{identitySummary.address}</Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      公钥长度
                    </Typography>
                    <Typography variant="body1">{identitySummary.publicKeyBytes} bytes</Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Root 模式
                    </Typography>
                    <Typography variant="body1">单 root node</Typography>
                  </Box>
                </Box>
              )}
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

          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Stable endpoint 使用 `IP/Port` 格式，用于告诉其他节点如何找到这个 root node。这里只接受可通过 API 和控制器语义表达的地址，不接受备注型文本。
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
                  error={hasEndpointError(endpoint.value)}
                  helperText={getEndpointHelperText(endpoint.value)}
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

      <Accordion sx={{ mb: 3 }} disabled={generating || loadingIdentity}>
        <AccordionSummary expandIcon={<ExpandMoreIcon />}>
          <Box>
            <Typography variant="h6">高级模式</Typography>
            <Typography variant="body2" color="text.secondary">
              默认生成模式会由服务端透明处理 signing keys。只有需要复用或管理签名文件时，才需要展开这一栏。
            </Typography>
          </Box>
        </AccordionSummary>
        <AccordionDetails>
          <Stack spacing={2}>
            <Alert severity="info">
              当前默认推荐主流程：直接读取 `identity.public` 后生成 Planet。高级模式不会引入多 root node，只额外处理 signing keys。
            </Alert>

            <Button
              variant={advancedModeEnabled ? 'contained' : 'outlined'}
              onClick={() => {
                setAdvancedModeEnabled((previous) => !previous)
                setSigningKeysMessage(null)
              }}
            >
              {advancedModeEnabled ? '已启用自定义 signing keys 模式' : '切换到自定义 signing keys 模式'}
            </Button>

            {advancedModeEnabled ? (
              <>
                <TextField
                  label="signing keys 目录"
                  fullWidth
                  value={signingKeyPath}
                  onChange={(event) => setSigningKeyPath(event.target.value)}
                  helperText="目录中应包含 previous.c25519 与 current.c25519。若不存在，可在下方直接生成。"
                  disabled={loadingSigningKeys || generatingSigningKeys || generating}
                />

                {signingKeysMessage && (
                  <Alert severity={signingKeysMessage.severity} onClose={() => setSigningKeysMessage(null)}>
                    {signingKeysMessage.text}
                  </Alert>
                )}

                <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
                  <Button
                    variant="outlined"
                    startIcon={<SyncIcon />}
                    onClick={() => { void loadSigningKeysInfo() }}
                    disabled={loadingSigningKeys || generatingSigningKeys || generating}
                  >
                    {loadingSigningKeys ? '检测中...' : '检测 signing keys'}
                  </Button>
                  <Button
                    variant="outlined"
                    color="secondary"
                    startIcon={<KeyIcon />}
                    onClick={() => { void handleGenerateSigningKeys() }}
                    disabled={loadingSigningKeys || generatingSigningKeys || generating}
                  >
                    {generatingSigningKeys ? '生成中...' : '生成 signing keys'}
                  </Button>
                </Box>

                {signingKeysInfo && (
                  <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(3, 1fr)' }, gap: 2 }}>
                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        当前模式
                      </Typography>
                      <Typography variant="body1">{signingKeysInfo.ready ? '自定义 signing keys' : '自定义目录，文件未就绪'}</Typography>
                    </Box>
                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        previous.c25519
                      </Typography>
                      <Typography variant="body1">{signingKeysInfo.previous_exists ? '已找到' : '未找到'}</Typography>
                    </Box>
                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        current.c25519
                      </Typography>
                      <Typography variant="body1">{signingKeysInfo.current_exists ? '已找到' : '未找到'}</Typography>
                    </Box>
                  </Box>
                )}
              </>
            ) : (
              <Alert severity="success">
                当前使用默认生成模式。生成时会由服务端透明处理 signing keys，不需要额外指定文件目录。
              </Alert>
            )}
          </Stack>
        </AccordionDetails>
      </Accordion>

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

            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              生成成功后可直接下载 `planet` 文件。替换到控制器前，建议先备份原文件并在目标环境完成验证。
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
