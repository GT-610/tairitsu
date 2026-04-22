import { useEffect, useState, type MouseEvent, type SyntheticEvent } from 'react'
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
  Menu,
  MenuItem,
  Snackbar,
  Tab,
  Tabs,
  Typography,
} from '@mui/material'
import { Alert as MuiAlert } from '@mui/material'
import { ArrowBack, ContentCopy } from '@mui/icons-material'
import { Link, useParams } from 'react-router-dom'
import {
  type IpAssignmentPool,
  type Member as ApiMember,
  type Network,
  type NetworkConfig,
  type NetworkMetadataUpdateRequest,
  type Route,
  memberAPI,
  networkAPI,
} from '../services/api'
import { getErrorMessage } from '../services/errors'
import DeleteMemberDialog from '../components/network-detail/DeleteMemberDialog'
import DNSSettingsSection from '../components/network-detail/DNSSettingsSection'
import EditMemberDialog from '../components/network-detail/EditMemberDialog'
import IPv4AssignmentSection from '../components/network-detail/IPv4AssignmentSection'
import IPv6AssignmentSection from '../components/network-detail/IPv6AssignmentSection'
import ManagedRoutesSection from '../components/network-detail/ManagedRoutesSection'
import MulticastSettingsSection from '../components/network-detail/MulticastSettingsSection'
import NetworkBasicSettingsSection from '../components/network-detail/NetworkBasicSettingsSection'
import NetworkMembersSection from '../components/network-detail/NetworkMembersSection'
import SettingsSectionCard from '../components/network-detail/SettingsSectionCard'
import type {
  BasicSettingsDraft,
  DNSSettingsDraft,
  IPv4SettingsDraft,
  IPv6SettingsDraft,
  ManagedRoutesSettingsDraft,
  MemberFormState,
  MulticastSettingsDraft,
  NetworkMemberDevice,
} from '../components/network-detail/types'
import {
  buildAutoAssignPoolFromSubnet,
  normalizeIPv6CIDR,
  normalizeRouteTarget,
} from '../utils/networkAddressUtils'
import {
  buildMergedPools,
  buildMergedRoutes,
  getInitialBasicSettings,
  getInitialDNSSettings,
  getInitialIPv4Settings,
  getInitialIPv6Settings,
  getInitialManagedRoutesSettings,
  getInitialMulticastSettings,
} from '../utils/networkSettings'
import {
  getIPv4ConfigurationIssues,
  getIPv4PoolOverlapIssues,
  getIPv4RangeIssue,
  getIPv6ConfigurationIssues,
  getIPv6PoolOverlapIssues,
  getIPv6RangeIssue,
  getManagedRouteIssue,
  isValidDnsServer,
} from '../utils/networkValidation'

const defaultNetworkConfig: Partial<NetworkConfig> = {
  private: true,
  allowPassiveBridging: true,
  enableBroadcast: true,
  mtu: 2800,
  multicastLimit: 32,
  v4AssignMode: {
    zt: false,
  },
  v6AssignMode: {
    zt: false,
    '6plane': false,
    rfc4193: false,
  },
}

const emptyBasicSettings: BasicSettingsDraft = { name: '', description: '' }
const emptyIPv4Settings: IPv4SettingsDraft = {
  subnet: '',
  autoAssign: false,
  pools: [],
  poolStartDraft: '',
  poolEndDraft: '',
}
const emptyIPv6Settings: IPv6SettingsDraft = {
  subnet: '',
  customAssign: false,
  rfc4193: false,
  plane6: false,
  pools: [],
  poolStartDraft: '',
  poolEndDraft: '',
}
const emptyManagedRoutesSettings: ManagedRoutesSettingsDraft = {
  routes: [],
  routeDraft: { target: '', via: '' },
}
const emptyDnsSettings: DNSSettingsDraft = { domain: '', servers: [], serverDraft: '' }
const emptyMulticastSettings: MulticastSettingsDraft = { multicastLimit: 32, enableBroadcast: true }
const emptyMemberForm: MemberFormState = {
  name: '',
  authorized: false,
  activeBridge: false,
  noAutoAssignIps: false,
  ipAssignments: [''],
}

function generateRandomIPv4Subnet(): string {
  return `10.${Math.floor(Math.random() * 256)}.${Math.floor(Math.random() * 256)}.0/24`
}

function formatNetworkMember(member: ApiMember): NetworkMemberDevice {
  const memberID = member.id || member.nodeId || ''
  const clientVersion = member.clientVersion || (
    member.vMajor !== undefined && member.vMinor !== undefined && member.vRev !== undefined &&
    member.vMajor >= 0 && member.vMinor >= 0 && member.vRev >= 0
      ? `${member.vMajor}.${member.vMinor}.${member.vRev}`
      : 'unknown'
  )

  return {
    id: memberID,
    name: member.name || member.nodeId || member.id || '未命名设备',
    description: member.description || '',
    authorized: member.config?.authorized ?? member.authorized ?? false,
    ipAssignments: member.config?.ipAssignments ?? member.ipAssignments ?? [],
    clientVersion,
    lastSeenLabel: member.lastSeen ? new Date(member.lastSeen).toLocaleString() : '未知',
    activeBridge: member.config?.activeBridge ?? member.activeBridge ?? false,
    noAutoAssignIps: member.config?.noAutoAssignIps ?? member.noAutoAssignIps ?? false,
  }
}

function NetworkDetail() {
  const { id } = useParams<{ id: string }>()
  const [network, setNetwork] = useState<Network | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [activeTab, setActiveTab] = useState(0)

  const [initialBasicSettings, setInitialBasicSettings] = useState<BasicSettingsDraft>(emptyBasicSettings)
  const [basicSettings, setBasicSettings] = useState<BasicSettingsDraft>(emptyBasicSettings)
  const [initialIPv4Settings, setInitialIPv4Settings] = useState<IPv4SettingsDraft>(emptyIPv4Settings)
  const [ipv4Settings, setIPv4Settings] = useState<IPv4SettingsDraft>(emptyIPv4Settings)
  const [initialIPv6Settings, setInitialIPv6Settings] = useState<IPv6SettingsDraft>(emptyIPv6Settings)
  const [ipv6Settings, setIPv6Settings] = useState<IPv6SettingsDraft>(emptyIPv6Settings)
  const [initialManagedRoutesSettings, setInitialManagedRoutesSettings] = useState<ManagedRoutesSettingsDraft>(emptyManagedRoutesSettings)
  const [managedRoutesSettings, setManagedRoutesSettings] = useState<ManagedRoutesSettingsDraft>(emptyManagedRoutesSettings)
  const [initialDnsSettings, setInitialDnsSettings] = useState<DNSSettingsDraft>(emptyDnsSettings)
  const [dnsSettings, setDnsSettings] = useState<DNSSettingsDraft>(emptyDnsSettings)
  const [initialMulticastSettings, setInitialMulticastSettings] = useState<MulticastSettingsDraft>(emptyMulticastSettings)
  const [multicastSettings, setMulticastSettings] = useState<MulticastSettingsDraft>(emptyMulticastSettings)

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({
    open: false,
    message: '',
    severity: 'success',
  })

  const [memberSearchTerm, setMemberSearchTerm] = useState('')
  const [hidePendingBanner, setHidePendingBanner] = useState(false)
  const [memberMenuAnchor, setMemberMenuAnchor] = useState<HTMLElement | null>(null)
  const [selectedMember, setSelectedMember] = useState<NetworkMemberDevice | null>(null)
  const [memberDialogOpen, setMemberDialogOpen] = useState(false)
  const [memberDeleteDialogOpen, setMemberDeleteDialogOpen] = useState(false)
  const [memberForm, setMemberForm] = useState<MemberFormState>(emptyMemberForm)

  useEffect(() => {
    void fetchNetworkDetail()
  }, [id])

  const showSnackbar = (message: string, severity: 'success' | 'error' | 'info' = 'success') => {
    setSnackbar({ open: true, message, severity })
  }

  const scheduleRefetch = () => {
    window.setTimeout(() => {
      void fetchNetworkDetail()
    }, 1000)
  }

  const fetchNetworkDetail = async () => {
    setLoading(true)
    try {
      if (!id) {
        throw new Error('网络ID不能为空')
      }

      const response = await networkAPI.getNetwork(id)
      const networkData = response.data
      if (!networkData) return

      networkData.config = { ...defaultNetworkConfig, ...networkData.config }
      if (networkData.config.v4AssignMode) {
        networkData.config.v4AssignMode = { ...defaultNetworkConfig.v4AssignMode, ...networkData.config.v4AssignMode }
      }
      if (networkData.config.v6AssignMode) {
        networkData.config.v6AssignMode = { ...defaultNetworkConfig.v6AssignMode, ...networkData.config.v6AssignMode }
      }

      setNetwork(networkData)

      const nextBasic = getInitialBasicSettings(networkData)
      const nextIPv4 = getInitialIPv4Settings(networkData)
      const nextIPv6 = getInitialIPv6Settings(networkData)
      const nextManagedRoutes = getInitialManagedRoutesSettings(networkData)
      const nextDns = getInitialDNSSettings(networkData)
      const nextMulticast = getInitialMulticastSettings(networkData)

      setInitialBasicSettings(nextBasic)
      setBasicSettings(nextBasic)
      setInitialIPv4Settings(nextIPv4)
      setIPv4Settings(nextIPv4)
      setInitialIPv6Settings(nextIPv6)
      setIPv6Settings(nextIPv6)
      setInitialManagedRoutesSettings(nextManagedRoutes)
      setManagedRoutesSettings(nextManagedRoutes)
      setInitialDnsSettings(nextDns)
      setDnsSettings(nextDns)
      setInitialMulticastSettings(nextMulticast)
      setMulticastSettings(nextMulticast)
      setHidePendingBanner(false)
      setError('')
    } catch (err: unknown) {
      setError(getErrorMessage(err, '获取网络详情失败'))
      console.error('Fetch network detail error:', err)
    } finally {
      setLoading(false)
    }
  }

  const memberDevices = (network?.members || []).map(formatNetworkMember)
  const pendingMembers = memberDevices.filter((member) => !member.authorized)
  const authorizedMembers = memberDevices.filter((member) => member.authorized)
  const filteredMembers = memberDevices.filter((member) => {
    const query = memberSearchTerm.trim().toLowerCase()
    if (query === '') return true
    return member.name.toLowerCase().includes(query) || member.id.toLowerCase().includes(query)
  })

  const ipv4RangeDraftIssue = getIPv4RangeIssue(ipv4Settings.subnet, {
    ipRangeStart: ipv4Settings.poolStartDraft,
    ipRangeEnd: ipv4Settings.poolEndDraft,
  })
  const ipv4ConfigurationIssues = getIPv4ConfigurationIssues(ipv4Settings.subnet, ipv4Settings.autoAssign, ipv4Settings.pools)
  const ipv6RangeDraftIssue = getIPv6RangeIssue(ipv6Settings.subnet, {
    ipRangeStart: ipv6Settings.poolStartDraft,
    ipRangeEnd: ipv6Settings.poolEndDraft,
  })
  const ipv6ConfigurationIssues = getIPv6ConfigurationIssues(ipv6Settings.subnet, ipv6Settings.customAssign, ipv6Settings.pools)

  const closeMemberMenu = () => {
    setMemberMenuAnchor(null)
  }

  const handleOpenMemberMenu = (event: MouseEvent<HTMLElement>, member: NetworkMemberDevice) => {
    setMemberMenuAnchor(event.currentTarget)
    setSelectedMember(member)
  }

  const handleOpenEditMember = () => {
    if (!selectedMember) return
    setMemberForm({
      name: selectedMember.name,
      authorized: selectedMember.authorized,
      activeBridge: selectedMember.activeBridge,
      noAutoAssignIps: selectedMember.noAutoAssignIps,
      ipAssignments: selectedMember.ipAssignments.length > 0 ? [...selectedMember.ipAssignments] : [''],
    })
    setMemberDialogOpen(true)
    closeMemberMenu()
  }

  const handleUpdateMemberStatus = async (member: NetworkMemberDevice, authorized: boolean) => {
    if (!id) return
    setSaving(true)
    try {
      await memberAPI.updateMember(id, member.id, { authorized })
      showSnackbar(authorized ? '设备授权成功' : '设备已拒绝')
      await fetchNetworkDetail()
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, authorized ? '设备授权失败' : '设备拒绝失败'), 'error')
    } finally {
      setSaving(false)
      closeMemberMenu()
    }
  }

  const handleSaveMember = async () => {
    if (!id || !selectedMember) return
    setSaving(true)
    try {
      await memberAPI.updateMember(id, selectedMember.id, {
        name: memberForm.name,
        authorized: memberForm.authorized,
        activeBridge: memberForm.activeBridge,
        noAutoAssignIps: memberForm.noAutoAssignIps,
        ipAssignments: memberForm.ipAssignments.map((ip) => ip.trim()).filter((ip) => ip !== ''),
      })
      showSnackbar('成员信息更新成功')
      setMemberDialogOpen(false)
      await fetchNetworkDetail()
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '更新成员失败'), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleDeleteMember = async () => {
    if (!id || !selectedMember) return
    setSaving(true)
    try {
      await memberAPI.deleteMember(id, selectedMember.id)
      showSnackbar('成员移除成功')
      setMemberDeleteDialogOpen(false)
      closeMemberMenu()
      await fetchNetworkDetail()
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '移除成员失败'), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleCopyMemberID = async () => {
    if (!selectedMember) return
    try {
      await navigator.clipboard.writeText(selectedMember.id)
      showSnackbar('设备 ID 已复制')
    } catch {
      showSnackbar('复制设备 ID 失败', 'error')
    }
  }

  const handleCopyNetworkID = async () => {
    if (!network?.id) return
    try {
      await navigator.clipboard.writeText(network.id)
      showSnackbar('网络 ID 已复制')
    } catch {
      showSnackbar('复制网络 ID 失败', 'error')
    }
  }

  const handleSaveMetadata = async () => {
    if (!id) return
    setSaving(true)
    try {
      const metadataData: NetworkMetadataUpdateRequest = {
        name: basicSettings.name,
        description: basicSettings.description,
      }
      await networkAPI.updateNetworkMetadata(id, metadataData)
      showSnackbar('名称和描述保存成功')
      scheduleRefetch()
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleDeleteNetwork = async () => {
    if (!id) return
    setSaving(true)
    try {
      await networkAPI.deleteNetwork(id)
      showSnackbar('网络删除成功', 'success')
      window.setTimeout(() => {
        window.location.href = '/networks'
      }, 1000)
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '删除失败'), 'error')
    } finally {
      setSaving(false)
      setDeleteDialogOpen(false)
    }
  }

  const handleSaveIPv4 = async () => {
    if (!id || !network) return
    if (ipv4ConfigurationIssues.length > 0) {
      showSnackbar(ipv4ConfigurationIssues[0], 'error')
      return
    }

    setSaving(true)
    try {
      await networkAPI.updateNetwork(id, {
        routes: buildMergedRoutes(
          network.config.routes || [],
          ipv4Settings.subnet,
          ipv6Settings.subnet,
          ipv6Settings.customAssign,
          managedRoutesSettings.routes,
        ),
        ipAssignmentPools: buildMergedPools(
          ipv4Settings.autoAssign,
          ipv4Settings.pools,
          ipv6Settings.customAssign,
          ipv6Settings.pools,
        ),
        v4AssignMode: { zt: ipv4Settings.autoAssign },
      })
      showSnackbar('IPv4配置保存成功')
      scheduleRefetch()
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleSaveIPv6 = async () => {
    if (!id || !network) return
    if (ipv6ConfigurationIssues.length > 0) {
      showSnackbar(ipv6ConfigurationIssues[0], 'error')
      return
    }

    setSaving(true)
    try {
      await networkAPI.updateNetwork(id, {
        routes: buildMergedRoutes(
          network.config.routes || [],
          ipv4Settings.subnet,
          ipv6Settings.subnet,
          ipv6Settings.customAssign,
          managedRoutesSettings.routes,
        ),
        ipAssignmentPools: buildMergedPools(
          ipv4Settings.autoAssign,
          ipv4Settings.pools,
          ipv6Settings.customAssign,
          ipv6Settings.pools,
        ),
        v6AssignMode: {
          zt: ipv6Settings.customAssign,
          '6plane': ipv6Settings.plane6,
          rfc4193: ipv6Settings.rfc4193,
        },
      })
      showSnackbar('IPv6配置保存成功')
      scheduleRefetch()
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleSaveManagedRoutes = async () => {
    if (!id || !network) return
    setSaving(true)
    try {
      await networkAPI.updateNetwork(id, {
        routes: buildMergedRoutes(
          network.config.routes || [],
          ipv4Settings.subnet,
          ipv6Settings.subnet,
          ipv6Settings.customAssign,
          managedRoutesSettings.routes,
        ),
      })
      showSnackbar('托管路由保存成功')
      scheduleRefetch()
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleSaveDNS = async () => {
    if (!id) return
    setSaving(true)
    try {
      await networkAPI.updateNetwork(id, {
        dns: {
          domain: dnsSettings.domain.trim(),
          servers: dnsSettings.servers,
        },
      })
      showSnackbar('DNS 设置保存成功')
      scheduleRefetch()
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleSaveMulticast = async () => {
    if (!id) return
    setSaving(true)
    try {
      await networkAPI.updateNetwork(id, {
        multicastLimit: multicastSettings.multicastLimit,
        enableBroadcast: multicastSettings.enableBroadcast,
      })
      showSnackbar('多播设置保存成功')
      scheduleRefetch()
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleGenerateSubnetAndRange = () => {
    const subnet = generateRandomIPv4Subnet()
    const pool = buildAutoAssignPoolFromSubnet(subnet)
    setIPv4Settings((prev) => ({
      ...prev,
      subnet,
      ...(pool ? { poolStartDraft: pool.ipRangeStart, poolEndDraft: pool.ipRangeEnd } : {}),
    }))
  }

  const handleSetIPv4Range = () => {
    const nextPool: IpAssignmentPool = {
      ipRangeStart: ipv4Settings.poolStartDraft.trim(),
      ipRangeEnd: ipv4Settings.poolEndDraft.trim(),
    }
    const issue = getIPv4RangeIssue(ipv4Settings.subnet, nextPool)
    if (issue) {
      showSnackbar(issue, 'error')
      return
    }
    if (ipv4Settings.pools.some((pool) => pool.ipRangeStart === nextPool.ipRangeStart && pool.ipRangeEnd === nextPool.ipRangeEnd)) {
      showSnackbar('该地址池范围已存在', 'info')
      return
    }
    const overlapIssues = getIPv4PoolOverlapIssues([...ipv4Settings.pools, nextPool])
    if (overlapIssues.length > 0) {
      showSnackbar(overlapIssues[0], 'error')
      return
    }
    setIPv4Settings((prev) => ({
      ...prev,
      pools: [...prev.pools, nextPool],
      poolStartDraft: '',
      poolEndDraft: '',
    }))
  }

  const handleSetIPv6Range = () => {
    const nextPool: IpAssignmentPool = {
      ipRangeStart: ipv6Settings.poolStartDraft.trim(),
      ipRangeEnd: ipv6Settings.poolEndDraft.trim(),
    }
    const issue = getIPv6RangeIssue(ipv6Settings.subnet, nextPool)
    if (issue) {
      showSnackbar(issue, 'error')
      return
    }
    if (ipv6Settings.pools.some((pool) => pool.ipRangeStart === nextPool.ipRangeStart && pool.ipRangeEnd === nextPool.ipRangeEnd)) {
      showSnackbar('该 IPv6 地址池范围已存在', 'info')
      return
    }
    const overlapIssues = getIPv6PoolOverlapIssues([...ipv6Settings.pools, nextPool])
    if (overlapIssues.length > 0) {
      showSnackbar(overlapIssues[0], 'error')
      return
    }
    setIPv6Settings((prev) => ({
      ...prev,
      pools: [...prev.pools, nextPool],
      poolStartDraft: '',
      poolEndDraft: '',
    }))
  }

  const handleAddManagedRoute = () => {
    const nextRoute: Route = {
      target: managedRoutesSettings.routeDraft.target.trim(),
      via: managedRoutesSettings.routeDraft.via?.trim() || undefined,
    }
    const issue = getManagedRouteIssue(nextRoute)
    if (issue) {
      showSnackbar(issue, 'error')
      return
    }
    const normalizedTarget = normalizeRouteTarget(nextRoute.target)
    if (managedRoutesSettings.routes.some((route) => normalizeRouteTarget(route.target) === normalizedTarget)) {
      showSnackbar('该托管路由目标已存在', 'info')
      return
    }
    setManagedRoutesSettings((prev) => ({
      routes: [...prev.routes, { target: normalizedTarget as string, via: nextRoute.via }],
      routeDraft: { target: '', via: '' },
    }))
  }

  const handleAddDnsServer = () => {
    const nextServer = dnsSettings.serverDraft.trim()
    if (!isValidDnsServer(nextServer)) {
      showSnackbar('DNS 服务器必须是有效的 IPv4 或 IPv6 地址', 'error')
      return
    }
    if (dnsSettings.servers.includes(nextServer)) {
      showSnackbar('该 DNS 服务器已存在', 'info')
      return
    }
    setDnsSettings((prev) => ({
      ...prev,
      servers: [...prev.servers, nextServer],
      serverDraft: '',
    }))
  }

  const handleTabChange = (_event: SyntheticEvent, newValue: number) => {
    setActiveTab(newValue)
  }

  const handleCloseSnackbar = () => {
    setSnackbar((prev) => ({ ...prev, open: false }))
  }

  if (error || !id) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error" sx={{ mb: 3 }}>
          {error || '网络不存在'}
        </Alert>
        <Button variant="contained" component={Link} to="/networks">
          返回网络列表
        </Button>
      </Box>
    )
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <IconButton component={Link} to="/networks" size="large">
            <ArrowBack />
          </IconButton>
          <Typography variant="h4" component="h1">
            {network?.name}
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Typography variant="body1" color="text.secondary">
            网络ID: {network?.id}
          </Typography>
          <IconButton size="small" onClick={() => { void handleCopyNetworkID() }}>
            <ContentCopy />
          </IconButton>
        </Box>
      </Box>

      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
        <Tabs value={activeTab} onChange={handleTabChange} aria-label="network tabs">
          <Tab label="成员设备" />
          <Tab label="设置" />
        </Tabs>
      </Box>

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          {activeTab === 0 && (
            <NetworkMembersSection
              memberDevices={memberDevices}
              pendingMembers={pendingMembers}
              authorizedMembers={authorizedMembers}
              filteredMembers={filteredMembers}
              memberSearchTerm={memberSearchTerm}
              saving={saving}
              hidePendingBanner={hidePendingBanner}
              onMemberSearchTermChange={setMemberSearchTerm}
              onHidePendingBanner={() => setHidePendingBanner(true)}
              onQuickApprove={() => { void handleUpdateMemberStatus(pendingMembers[0], true) }}
              onQuickReject={() => { void handleUpdateMemberStatus(pendingMembers[0], false) }}
              onOpenMemberMenu={handleOpenMemberMenu}
            />
          )}

          {activeTab === 1 && (
            <>
              <NetworkBasicSettingsSection
                saving={saving}
                initialValue={initialBasicSettings}
                draftValue={basicSettings}
                onChange={setBasicSettings}
                onReset={() => setBasicSettings(initialBasicSettings)}
                onSave={() => { void handleSaveMetadata() }}
              />

              <IPv4AssignmentSection
                saving={saving}
                initialValue={initialIPv4Settings}
                draftValue={ipv4Settings}
                draftRangeIssue={ipv4RangeDraftIssue}
                configurationIssues={ipv4ConfigurationIssues}
                onChange={(next) => {
                  let updated = next
                  if (updated.autoAssign && updated.pools.length === 0 && (!updated.poolStartDraft || !updated.poolEndDraft)) {
                    const generatedPool = buildAutoAssignPoolFromSubnet(updated.subnet)
                    if (generatedPool) {
                      updated = {
                        ...updated,
                        poolStartDraft: generatedPool.ipRangeStart,
                        poolEndDraft: generatedPool.ipRangeEnd,
                      }
                    }
                  }
                  setIPv4Settings(updated)
                }}
                onReset={() => setIPv4Settings(initialIPv4Settings)}
                onSave={() => { void handleSaveIPv4() }}
                onGenerateSubnetAndRange={handleGenerateSubnetAndRange}
                onSetRange={handleSetIPv4Range}
                onRemoveRange={(index) => setIPv4Settings((prev) => ({
                  ...prev,
                  pools: prev.pools.filter((_, currentIndex) => currentIndex !== index),
                }))}
              />

              <IPv6AssignmentSection
                saving={saving}
                initialValue={initialIPv6Settings}
                draftValue={ipv6Settings}
                draftRangeIssue={ipv6RangeDraftIssue}
                configurationIssues={ipv6ConfigurationIssues}
                subnetValid={Boolean(normalizeIPv6CIDR(ipv6Settings.subnet))}
                onChange={setIPv6Settings}
                onReset={() => setIPv6Settings(initialIPv6Settings)}
                onSave={() => { void handleSaveIPv6() }}
                onSetRange={handleSetIPv6Range}
                onRemoveRange={(index) => setIPv6Settings((prev) => ({
                  ...prev,
                  pools: prev.pools.filter((_, currentIndex) => currentIndex !== index),
                }))}
              />

              <ManagedRoutesSection
                saving={saving}
                initialValue={initialManagedRoutesSettings}
                draftValue={managedRoutesSettings}
                onChange={setManagedRoutesSettings}
                onReset={() => setManagedRoutesSettings(initialManagedRoutesSettings)}
                onSave={() => { void handleSaveManagedRoutes() }}
                onAddRoute={handleAddManagedRoute}
                onRemoveRoute={(index) => setManagedRoutesSettings((prev) => ({
                  ...prev,
                  routes: prev.routes.filter((_, currentIndex) => currentIndex !== index),
                }))}
              />

              <DNSSettingsSection
                saving={saving}
                initialValue={initialDnsSettings}
                draftValue={dnsSettings}
                onChange={setDnsSettings}
                onReset={() => setDnsSettings(initialDnsSettings)}
                onSave={() => { void handleSaveDNS() }}
                onAddServer={handleAddDnsServer}
                onRemoveServer={(index) => setDnsSettings((prev) => ({
                  ...prev,
                  servers: prev.servers.filter((_, currentIndex) => currentIndex !== index),
                }))}
              />

              <MulticastSettingsSection
                saving={saving}
                initialValue={initialMulticastSettings}
                draftValue={multicastSettings}
                onChange={setMulticastSettings}
                onReset={() => setMulticastSettings(initialMulticastSettings)}
                onSave={() => { void handleSaveMulticast() }}
              />

              <SettingsSectionCard title="删除网络">
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  此操作不可恢复。删除网络将断开所有连接的设备，并永久删除网络配置。
                </Typography>
                <Button variant="contained" color="error" size="large" onClick={() => setDeleteDialogOpen(true)}>
                  删除网络
                </Button>
              </SettingsSectionCard>

              <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
                <DialogTitle>确认删除网络</DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    您确定要删除网络 "{network?.name}" 吗？此操作不可恢复，将永久删除该网络及其所有配置。
                  </DialogContentText>
                </DialogContent>
                <DialogActions>
                  <Button onClick={() => setDeleteDialogOpen(false)} color="primary">取消</Button>
                  <Button onClick={() => { void handleDeleteNetwork() }} color="error" autoFocus disabled={saving}>
                    {saving ? '删除中...' : '确认删除'}
                  </Button>
                </DialogActions>
              </Dialog>
            </>
          )}
        </>
      )}

      <Menu anchorEl={memberMenuAnchor} open={Boolean(memberMenuAnchor)} onClose={closeMemberMenu}>
        <MenuItem onClick={handleOpenEditMember}>编辑成员</MenuItem>
        {selectedMember && !selectedMember.authorized && (
          <MenuItem onClick={() => { void handleUpdateMemberStatus(selectedMember, true) }}>
            授权
          </MenuItem>
        )}
        {selectedMember && (
          <MenuItem onClick={() => { void handleUpdateMemberStatus(selectedMember, false) }}>
            拒绝设备
          </MenuItem>
        )}
        <MenuItem
          onClick={() => {
            setMemberDeleteDialogOpen(true)
            closeMemberMenu()
          }}
          sx={{ color: 'error.main' }}
        >
          移除成员
        </MenuItem>
      </Menu>

      <EditMemberDialog
        open={memberDialogOpen}
        saving={saving}
        selectedMember={selectedMember}
        memberForm={memberForm}
        onClose={() => setMemberDialogOpen(false)}
        onCopyMemberID={() => { void handleCopyMemberID() }}
        onMemberFormChange={setMemberForm}
        onSave={() => { void handleSaveMember() }}
      />

      <DeleteMemberDialog
        open={memberDeleteDialogOpen}
        saving={saving}
        selectedMember={selectedMember}
        onClose={() => setMemberDeleteDialogOpen(false)}
        onConfirm={() => { void handleDeleteMember() }}
      />

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <MuiAlert onClose={handleCloseSnackbar} severity={snackbar.severity} variant="filled" sx={{ width: '100%' }}>
          {snackbar.message}
        </MuiAlert>
      </Snackbar>
    </Box>
  )
}

export default NetworkDetail
