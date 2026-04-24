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
  Divider,
  IconButton,
  Stack,
  Switch,
  TextField,
  Typography,
  FormControlLabel,
} from '@mui/material'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import DownloadIcon from '@mui/icons-material/Download'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import KeyIcon from '@mui/icons-material/Key'
import SyncIcon from '@mui/icons-material/Sync'
import FolderOpenIcon from '@mui/icons-material/FolderOpen'
import EditNoteIcon from '@mui/icons-material/EditNote'
import { planetAPI, type GeneratePlanetResponse, type SigningKeysInfoResponse } from '../services/api'
import { getErrorMessage } from '../services/errors'
import {
  getPlanetDownloadName,
  normalizePlanetEndpoints,
  parsePlanetIdentityPublic,
  validatePlanetEndpointValue,
  validatePlanetEndpoints,
  validatePlanetRootNodes,
} from '../utils/planet'

type IdentitySourceMode = 'path' | 'manual'

interface EndpointDraft {
  id: string
  value: string
}

interface RootNodeDraft {
  id: string
  sourceMode: IdentitySourceMode
  identityPath: string
  resolvedIdentityPath: string
  identityPublic: string
  comments: string
  endpoints: EndpointDraft[]
  message: { severity: 'success' | 'error'; text: string } | null
  loadingIdentity: boolean
}

type PlanetResultState = GeneratePlanetResponse

const defaultIdentityPath = '/var/lib/zerotier-one'
const defaultSigningKeyPath = '/var/lib/zerotier-one'

function createEndpointDraft(value = ''): EndpointDraft {
  return { id: `${Date.now()}-${Math.random()}`, value }
}

function createRootNodeDraft(overrides?: Partial<RootNodeDraft>): RootNodeDraft {
  return {
    id: `${Date.now()}-${Math.random()}`,
    sourceMode: 'path',
    identityPath: defaultIdentityPath,
    resolvedIdentityPath: '',
    identityPublic: '',
    comments: '',
    endpoints: [createEndpointDraft()],
    message: null,
    loadingIdentity: false,
    ...overrides,
  }
}

function PlanetGenerator() {
  const [loadingIdentity, setLoadingIdentity] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [loadingSigningKeys, setLoadingSigningKeys] = useState(false)
  const [generatingSigningKeys, setGeneratingSigningKeys] = useState(false)
  const [identityPublic, setIdentityPublic] = useState('')
  const [identityPath, setIdentityPath] = useState(defaultIdentityPath)
  const [resolvedIdentityPath, setResolvedIdentityPath] = useState('')
  const [comments, setComments] = useState('')
  const [endpoints, setEndpoints] = useState<EndpointDraft[]>([createEndpointDraft()])
  const [message, setMessage] = useState<{ severity: 'success' | 'error'; text: string } | null>(null)
  const [identityMessage, setIdentityMessage] = useState<{ severity: 'success' | 'error'; text: string } | null>(null)
  const [advancedModeEnabled, setAdvancedModeEnabled] = useState(false)
  const [recommendValues, setRecommendValues] = useState(true)
  const [planetId, setPlanetId] = useState('')
  const [birthTime, setBirthTime] = useState('')
  const [downloadName, setDownloadName] = useState('planet')
  const [useCustomSigningKeys, setUseCustomSigningKeys] = useState(false)
  const [signingKeyPath, setSigningKeyPath] = useState(defaultSigningKeyPath)
  const [signingKeysMessage, setSigningKeysMessage] = useState<{ severity: 'success' | 'error'; text: string } | null>(null)
  const [signingKeysInfo, setSigningKeysInfo] = useState<SigningKeysInfoResponse | null>(null)
  const [rootNodes, setRootNodes] = useState<RootNodeDraft[]>([createRootNodeDraft()])
  const [generatedPlanet, setGeneratedPlanet] = useState<PlanetResultState | null>(null)

  const identitySummary = parsePlanetIdentityPublic(identityPublic)

  const syncMainFlowIntoFirstRootNode = () => {
    setRootNodes((previous) => {
      const [first, ...rest] = previous.length > 0 ? previous : [createRootNodeDraft()]
      return [
        {
          ...first,
          sourceMode: 'path',
          identityPath,
          resolvedIdentityPath,
          identityPublic: identityPublic.trim(),
          comments: comments.trim(),
          endpoints: endpoints.length > 0 ? endpoints.map((endpoint) => createEndpointDraft(endpoint.value)) : [createEndpointDraft()],
        },
        ...rest,
      ]
    })
  }

  const loadSigningKeysInfo = async (pathOverride?: string) => {
    const nextPath = pathOverride ?? signingKeyPath
    try {
      setLoadingSigningKeys(true)
      setSigningKeysMessage(null)
      const response = await planetAPI.getSigningKeysInfo(nextPath)
      setSigningKeysInfo(response.data)
      setSigningKeysMessage({
        severity: 'success',
        text: response.data.ready ? '已检测到完整的 signing keys' : '目录可用，但还没有完整的 signing keys',
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

  const loadRootNodeIdentity = async (rootNodeId: string) => {
    const rootNode = rootNodes.find((item) => item.id === rootNodeId)
    if (!rootNode) {
      return
    }

    try {
      setRootNodes((previous) => previous.map((item) => (
        item.id === rootNodeId ? { ...item, loadingIdentity: true, message: null } : item
      )))
      const response = await planetAPI.getIdentity(rootNode.identityPath)
      setRootNodes((previous) => previous.map((item) => (
        item.id === rootNodeId
          ? {
              ...item,
              loadingIdentity: false,
              identityPublic: response.data.identity_public,
              resolvedIdentityPath: response.data.identity_path,
              message: { severity: 'success', text: 'identity.public 读取成功' },
            }
          : item
      )))
    } catch (error: unknown) {
      setRootNodes((previous) => previous.map((item) => (
        item.id === rootNodeId
          ? {
              ...item,
              loadingIdentity: false,
              identityPublic: '',
              resolvedIdentityPath: '',
              message: { severity: 'error', text: getErrorMessage(error, '读取 identity.public 失败') },
            }
          : item
      )))
    }
  }

  const handleAddEndpoint = () => {
    setEndpoints((previous) => [...previous, createEndpointDraft()])
  }

  const handleRemoveEndpoint = (id: string) => {
    setEndpoints((previous) => (previous.length > 1 ? previous.filter((endpoint) => endpoint.id !== id) : previous))
  }

  const handleEndpointChange = (id: string, value: string) => {
    setEndpoints((previous) => previous.map((endpoint) => (endpoint.id === id ? { ...endpoint, value } : endpoint)))
  }

  const updateRootNode = (rootNodeId: string, updater: (rootNode: RootNodeDraft) => RootNodeDraft) => {
    setRootNodes((previous) => previous.map((rootNode) => (rootNode.id === rootNodeId ? updater(rootNode) : rootNode)))
  }

  const addRootNode = () => {
    setRootNodes((previous) => [...previous, createRootNodeDraft()])
  }

  const removeRootNode = (rootNodeId: string) => {
    setRootNodes((previous) => (previous.length > 1 ? previous.filter((rootNode) => rootNode.id !== rootNodeId) : previous))
  }

  const addRootNodeEndpoint = (rootNodeId: string) => {
    updateRootNode(rootNodeId, (rootNode) => ({
      ...rootNode,
      endpoints: [...rootNode.endpoints, createEndpointDraft()],
    }))
  }

  const removeRootNodeEndpoint = (rootNodeId: string, endpointId: string) => {
    updateRootNode(rootNodeId, (rootNode) => ({
      ...rootNode,
      endpoints: rootNode.endpoints.length > 1 ? rootNode.endpoints.filter((endpoint) => endpoint.id !== endpointId) : rootNode.endpoints,
    }))
  }

  const updateRootNodeEndpoint = (rootNodeId: string, endpointId: string, value: string) => {
    updateRootNode(rootNodeId, (rootNode) => ({
      ...rootNode,
      endpoints: rootNode.endpoints.map((endpoint) => (endpoint.id === endpointId ? { ...endpoint, value } : endpoint)),
    }))
  }

  const buildMainFlowRootNode = () => ({
    identity_public: identityPublic.trim(),
    comments: comments.trim(),
    endpoints: normalizePlanetEndpoints(endpoints.map((endpoint) => endpoint.value)),
  })

  const buildAdvancedRootNodes = () => rootNodes.map((rootNode) => ({
    identity_public: rootNode.identityPublic.trim(),
    comments: rootNode.comments.trim(),
    endpoints: normalizePlanetEndpoints(rootNode.endpoints.map((endpoint) => endpoint.value)),
  }))

  const handleGenerate = async () => {
    if (!advancedModeEnabled && !identityPublic.trim()) {
      setMessage({ severity: 'error', text: '请先成功加载 identity.public' })
      setGeneratedPlanet(null)
      return
    }

    if (!advancedModeEnabled) {
      const endpointValues = endpoints.map((endpoint) => endpoint.value)
      const endpointError = validatePlanetEndpoints(endpointValues)
      if (endpointError) {
        setMessage({ severity: 'error', text: endpointError })
        setGeneratedPlanet(null)
        return
      }
    } else {
      const rootNodeError = validatePlanetRootNodes(rootNodes.map((rootNode) => ({
        identityPublic: rootNode.identityPublic,
        endpoints: rootNode.endpoints.map((endpoint) => endpoint.value),
      })))
      if (rootNodeError) {
        setMessage({ severity: 'error', text: rootNodeError })
        setGeneratedPlanet(null)
        return
      }

      if (!recommendValues) {
        if (!planetId.trim()) {
          setMessage({ severity: 'error', text: '关闭推荐值后，需要填写 Planet ID' })
          setGeneratedPlanet(null)
          return
        }
        if (!birthTime.trim()) {
          setMessage({ severity: 'error', text: '关闭推荐值后，需要填写 Birth Time' })
          setGeneratedPlanet(null)
          return
        }
        if (!Number.isInteger(Number(planetId)) || Number(planetId) <= 0) {
          setMessage({ severity: 'error', text: 'Planet ID 必须是正整数' })
          setGeneratedPlanet(null)
          return
        }
        if (!Number.isInteger(Number(birthTime)) || Number(birthTime) <= 0) {
          setMessage({ severity: 'error', text: 'Birth Time 必须是正整数毫秒时间戳' })
          setGeneratedPlanet(null)
          return
        }
      }
    }

    try {
      setGenerating(true)
      setMessage(null)

      const response = await planetAPI.generatePlanet({
        root_nodes: advancedModeEnabled ? buildAdvancedRootNodes() : [buildMainFlowRootNode()],
        signing_key_path: advancedModeEnabled && useCustomSigningKeys ? signingKeyPath.trim() : undefined,
        planet_id: advancedModeEnabled && !recommendValues ? Number(planetId) : undefined,
        birth_time: advancedModeEnabled && !recommendValues ? Number(birthTime) : undefined,
        recommend_values: advancedModeEnabled ? recommendValues : true,
        download_name: advancedModeEnabled ? downloadName.trim() : 'planet',
      })

      setGeneratedPlanet(response.data)
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

  const renderEndpointFields = (
    endpointDrafts: EndpointDraft[],
    onChange: (endpointId: string, value: string) => void,
    onRemove: (endpointId: string) => void,
    onAdd: () => void,
    disabled: boolean,
  ) => (
    <Stack spacing={2}>
      {endpointDrafts.map((endpoint, index) => (
        <Box key={endpoint.id} sx={{ display: 'flex', gap: 1, alignItems: 'flex-start' }}>
          <TextField
            label={`端点 ${index + 1}`}
            placeholder="IP/Port，例如：203.0.113.1/9993"
            value={endpoint.value}
            onChange={(event) => onChange(endpoint.id, event.target.value)}
            fullWidth
            error={Boolean(endpoint.value.trim()) && validatePlanetEndpointValue(endpoint.value) !== null}
            helperText={endpoint.value.trim()
              ? validatePlanetEndpointValue(endpoint.value) ?? '该地址会作为 stable endpoint 写入 planet'
              : '格式：IP/Port。IPv4 示例：203.0.113.1/9993；IPv6 示例：2001:db8::1/9993'}
            disabled={disabled}
          />
          {endpointDrafts.length > 1 && (
            <IconButton color="error" onClick={() => onRemove(endpoint.id)} disabled={disabled} sx={{ mt: 1 }}>
              <DeleteIcon />
            </IconButton>
          )}
        </Box>
      ))}
      <Box>
        <Button startIcon={<AddIcon />} onClick={onAdd} disabled={disabled}>
          添加端点
        </Button>
      </Box>
    </Stack>
  )

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
            默认模式会从给定目录读取一个 `identity.public`，并将其作为单 root node 的 identity。
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
                    ? '已读取真实 identity.public，可继续填写默认模式配置'
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
            placeholder="例如：Primary root node"
            helperText="默认模式下，这会作为当前 root node 的注释信息"
            disabled={generating}
          />

          <Typography variant="subtitle2" sx={{ mb: 1 }}>
            端点列表
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Stable endpoint 使用 `IP/Port` 格式，用于告诉其他节点如何找到这个 root node。
          </Typography>

          {renderEndpointFields(
            endpoints,
            (endpointId, value) => handleEndpointChange(endpointId, value),
            (endpointId) => handleRemoveEndpoint(endpointId),
            handleAddEndpoint,
            generating,
          )}
        </CardContent>
      </Card>

      <Accordion
        sx={{ mb: 3 }}
        disabled={generating || loadingIdentity}
        onChange={(_, expanded) => {
          if (expanded && !advancedModeEnabled) {
            syncMainFlowIntoFirstRootNode()
          }
          setAdvancedModeEnabled(expanded)
        }}
      >
        <AccordionSummary expandIcon={<ExpandMoreIcon />}>
          <Box>
            <Typography variant="h6">高级模式</Typography>
            <Typography variant="body2" color="text.secondary">
              补全多 root node、手工元数据和自定义 signing keys，但默认主流程仍保持简洁。
            </Typography>
          </Box>
        </AccordionSummary>
        <AccordionDetails>
          <Stack spacing={3}>
            <Alert severity="info">
              启用高级模式后，将以高级配置区作为唯一提交真相。默认主流程中的 identity、comments 和 endpoints 会自动带入第一个 root node。
            </Alert>

            <Card variant="outlined">
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Planet 元数据
                </Typography>

                <FormControlLabel
                  control={(
                    <Switch
                      checked={recommendValues}
                      onChange={(event) => setRecommendValues(event.target.checked)}
                    />
                  )}
                  label="使用推荐值"
                />

                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  推荐值会由系统自动生成更安全的 `planet_id` 与 `birth_time`。关闭后可手工指定，但不建议随意修改。
                </Typography>

                <Stack spacing={2}>
                  <TextField
                    label="Planet ID"
                    value={planetId}
                    onChange={(event) => setPlanetId(event.target.value)}
                    disabled={recommendValues || generating}
                    helperText="关闭推荐值后可手工填写，不能使用保留值"
                  />
                  <TextField
                    label="Birth Time"
                    value={birthTime}
                    onChange={(event) => setBirthTime(event.target.value)}
                    disabled={recommendValues || generating}
                    helperText="使用 Unix 毫秒时间戳"
                  />
                  <TextField
                    label="下载文件名"
                    value={downloadName}
                    onChange={(event) => setDownloadName(event.target.value)}
                    disabled={generating}
                    helperText="仅影响浏览器下载文件名，不会写入服务器路径"
                  />
                </Stack>
              </CardContent>
            </Card>

            <Card variant="outlined">
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Signing Keys
                </Typography>

                <FormControlLabel
                  control={(
                    <Switch
                      checked={useCustomSigningKeys}
                      onChange={(event) => setUseCustomSigningKeys(event.target.checked)}
                    />
                  )}
                  label="使用自定义 signing keys 目录"
                />

                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  默认模式下由服务端透明处理 signing keys；只有需要复用已有 keys 或显式管理时才启用自定义目录。
                </Typography>

                {useCustomSigningKeys ? (
                  <Stack spacing={2}>
                    <TextField
                      label="signing keys 目录"
                      fullWidth
                      value={signingKeyPath}
                      onChange={(event) => setSigningKeyPath(event.target.value)}
                      disabled={loadingSigningKeys || generatingSigningKeys || generating}
                      helperText="目录中应包含 previous.c25519 与 current.c25519"
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
                          <Typography variant="body2" color="text.secondary">当前模式</Typography>
                          <Typography variant="body1">{signingKeysInfo.ready ? '自定义 signing keys' : '自定义目录，文件未就绪'}</Typography>
                        </Box>
                        <Box>
                          <Typography variant="body2" color="text.secondary">previous.c25519</Typography>
                          <Typography variant="body1">{signingKeysInfo.previous_exists ? '已找到' : '未找到'}</Typography>
                        </Box>
                        <Box>
                          <Typography variant="body2" color="text.secondary">current.c25519</Typography>
                          <Typography variant="body1">{signingKeysInfo.current_exists ? '已找到' : '未找到'}</Typography>
                        </Box>
                      </Box>
                    )}
                  </Stack>
                ) : (
                  <Alert severity="success">当前使用默认 signing keys 模式。</Alert>
                )}
              </CardContent>
            </Card>

            <Card variant="outlined">
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2, gap: 2, flexWrap: 'wrap' }}>
                  <Box>
                    <Typography variant="h6">Root Nodes</Typography>
                    <Typography variant="body2" color="text.secondary">
                      每个 root node 都可选择“从目录读取 identity.public”或“手工粘贴 identity.public”。
                    </Typography>
                  </Box>
                  <Button startIcon={<AddIcon />} onClick={addRootNode} disabled={generating}>
                    添加 Root Node
                  </Button>
                </Box>

                <Stack spacing={2}>
                  {rootNodes.map((rootNode, index) => {
                    const rootIdentitySummary = parsePlanetIdentityPublic(rootNode.identityPublic)
                    return (
                      <Card key={rootNode.id} variant="outlined">
                        <CardContent>
                          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2, gap: 2 }}>
                            <Typography variant="subtitle1">Root Node {index + 1}</Typography>
                            {rootNodes.length > 1 && (
                              <IconButton color="error" onClick={() => removeRootNode(rootNode.id)} disabled={generating}>
                                <DeleteIcon />
                              </IconButton>
                            )}
                          </Box>

                          {rootNode.message && (
                            <Alert severity={rootNode.message.severity} sx={{ mb: 2 }} onClose={() => updateRootNode(rootNode.id, (item) => ({ ...item, message: null }))}>
                              {rootNode.message.text}
                            </Alert>
                          )}

                          <Box sx={{ display: 'flex', gap: 1, mb: 2, flexWrap: 'wrap' }}>
                            <Button
                              variant={rootNode.sourceMode === 'path' ? 'contained' : 'outlined'}
                              startIcon={<FolderOpenIcon />}
                              onClick={() => updateRootNode(rootNode.id, (item) => ({ ...item, sourceMode: 'path' }))}
                              disabled={generating}
                            >
                              从目录读取
                            </Button>
                            <Button
                              variant={rootNode.sourceMode === 'manual' ? 'contained' : 'outlined'}
                              startIcon={<EditNoteIcon />}
                              onClick={() => updateRootNode(rootNode.id, (item) => ({ ...item, sourceMode: 'manual' }))}
                              disabled={generating}
                            >
                              手工粘贴
                            </Button>
                          </Box>

                          <Stack spacing={2}>
                            {rootNode.sourceMode === 'path' ? (
                              <>
                                <TextField
                                  label="ZeroTier 数据目录路径"
                                  fullWidth
                                  value={rootNode.identityPath}
                                  onChange={(event) => updateRootNode(rootNode.id, (item) => ({ ...item, identityPath: event.target.value }))}
                                  disabled={rootNode.loadingIdentity || generating}
                                />
                                <Button
                                  variant="outlined"
                                  onClick={() => { void loadRootNodeIdentity(rootNode.id) }}
                                  disabled={rootNode.loadingIdentity || generating}
                                >
                                  {rootNode.loadingIdentity ? '读取中...' : '读取该 Root 的身份'}
                                </Button>
                                {rootNode.resolvedIdentityPath && (
                                  <Typography variant="body2" color="text.secondary">
                                    实际读取路径：{rootNode.resolvedIdentityPath}
                                  </Typography>
                                )}
                              </>
                            ) : null}

                            <TextField
                              label="identity.public"
                              fullWidth
                              multiline
                              rows={2}
                              value={rootNode.identityPublic}
                              onChange={(event) => updateRootNode(rootNode.id, (item) => ({ ...item, identityPublic: event.target.value }))}
                              placeholder="格式：10hexdigits:0:publicKey"
                              helperText={rootNode.identityPublic.trim()
                                ? rootIdentitySummary
                                  ? '该 identity.public 将作为此 root node 的 identity'
                                  : 'identity.public 格式无效'
                                : '可读取后自动填入，也可直接手工粘贴'}
                              error={Boolean(rootNode.identityPublic.trim()) && !rootIdentitySummary}
                              disabled={rootNode.loadingIdentity || generating}
                            />

                            <TextField
                              label="Root Node 备注"
                              fullWidth
                              value={rootNode.comments}
                              onChange={(event) => updateRootNode(rootNode.id, (item) => ({ ...item, comments: event.target.value }))}
                              helperText="用于标识该 root node 的注释信息"
                              disabled={generating}
                            />

                            <Divider />

                            <Typography variant="subtitle2">Root Node 端点</Typography>
                            {renderEndpointFields(
                              rootNode.endpoints,
                              (endpointId, value) => updateRootNodeEndpoint(rootNode.id, endpointId, value),
                              (endpointId) => removeRootNodeEndpoint(rootNode.id, endpointId),
                              () => addRootNodeEndpoint(rootNode.id),
                              generating || rootNode.loadingIdentity,
                            )}
                          </Stack>
                        </CardContent>
                      </Card>
                    )
                  })}
                </Stack>
              </CardContent>
            </Card>
          </Stack>
        </AccordionDetails>
      </Accordion>

      <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
        <Button
          variant="contained"
          onClick={() => { void handleGenerate() }}
          disabled={loadingIdentity || generating}
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
                <Typography variant="body2" color="text.secondary">Planet ID</Typography>
                <Typography variant="body1">{generatedPlanet.planet_id}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">生成时间</Typography>
                <Typography variant="body1">{new Date(generatedPlanet.birth_time).toLocaleString()}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Root Node 数</Typography>
                <Typography variant="body1">{generatedPlanet.root_node_count}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">总端点数</Typography>
                <Typography variant="body1">{generatedPlanet.endpoint_count}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">下载文件名</Typography>
                <Typography variant="body1">{generatedPlanet.download_name}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">推荐值</Typography>
                <Typography variant="body1">{generatedPlanet.used_recommended_values ? '已使用' : '未使用'}</Typography>
              </Box>
            </Box>

            <Button variant="contained" color="success" startIcon={<DownloadIcon />} onClick={handleDownloadPlanet}>
              下载 Planet
            </Button>
          </CardContent>
        </Card>
      )}
    </Box>
  )
}

export default PlanetGenerator
