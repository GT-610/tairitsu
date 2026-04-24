import { useEffect, useMemo, useState } from 'react'
import { Alert, Box, Button, CircularProgress, IconButton, Snackbar, Typography } from '@mui/material'
import { ArrowBack, ContentCopy } from '@mui/icons-material'
import { Link, useParams } from 'react-router-dom'
import { Alert as MuiAlert } from '@mui/material'
import { memberAPI, networkAPI, type SharedNetworkSummary } from '../services/api'
import { getErrorMessage } from '../services/errors'
import NetworkMembersSection from '../components/network-detail/NetworkMembersSection'
import type { NetworkMemberDevice } from '../components/network-detail/types'
import { formatNetworkMember } from '../utils/memberUtils'

function SharedNetworkMembers() {
  const { id } = useParams<{ id: string }>()
  const [network, setNetwork] = useState<SharedNetworkSummary | null>(null)
  const [members, setMembers] = useState<NetworkMemberDevice[]>([])
  const [memberSearchTerm, setMemberSearchTerm] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  })

  useEffect(() => {
    void fetchSharedNetwork()
  }, [id])

  const fetchSharedNetwork = async () => {
    setLoading(true)
    try {
      if (!id) {
        throw new Error('网络ID不能为空')
      }

      const [sharedNetworksResponse, membersResponse] = await Promise.all([
        networkAPI.getSharedNetworks(),
        memberAPI.getMembers(id),
      ])

      const sharedNetworks = Array.isArray(sharedNetworksResponse.data) ? sharedNetworksResponse.data : []
      const currentNetwork = sharedNetworks.find((item) => item.id === id) || null
      if (!currentNetwork) {
        throw new Error('共享网络不存在或已失去访问权限')
      }

      setNetwork(currentNetwork)
      setMembers((Array.isArray(membersResponse.data) ? membersResponse.data : []).map(formatNetworkMember))
      setError('')
    } catch (err: unknown) {
      setError(getErrorMessage(err, '获取共享网络成员失败'))
      setNetwork(null)
      setMembers([])
    } finally {
      setLoading(false)
    }
  }

  const filteredMembers = useMemo(() => {
    const query = memberSearchTerm.trim().toLowerCase()
    if (!query) return members
    return members.filter((member) => member.name.toLowerCase().includes(query) || member.id.toLowerCase().includes(query))
  }, [memberSearchTerm, members])

  const pendingMembers = members.filter((member) => !member.authorized)
  const authorizedMembers = members.filter((member) => member.authorized)

  const handleCopyNetworkID = async () => {
    if (!network?.id) return
    try {
      await navigator.clipboard.writeText(network.id)
      setSnackbar({ open: true, message: '网络 ID 已复制', severity: 'success' })
    } catch {
      setSnackbar({ open: true, message: '复制网络 ID 失败', severity: 'error' })
    }
  }

  if (error || !id) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error" sx={{ mb: 3 }}>
          {error || '共享网络不存在'}
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
          <Box>
            <Typography variant="h4" component="h1">
              {network?.name}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              由 {network?.owner_username || network?.owner_id} 授予只读查看权限
            </Typography>
          </Box>
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

      {network?.description ? (
        <Alert severity="info" sx={{ mb: 3 }}>
          {network.description}
        </Alert>
      ) : null}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        <NetworkMembersSection
          memberDevices={members}
          pendingMembers={pendingMembers}
          authorizedMembers={authorizedMembers}
          filteredMembers={filteredMembers}
          memberSearchTerm={memberSearchTerm}
          saving={false}
          hidePendingBanner
          onMemberSearchTermChange={setMemberSearchTerm}
          onHidePendingBanner={() => {}}
          onQuickApprove={() => {}}
          onQuickReject={() => {}}
          readOnly
        />
      )}

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar((previous) => ({ ...previous, open: false }))}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <MuiAlert
          onClose={() => setSnackbar((previous) => ({ ...previous, open: false }))}
          severity={snackbar.severity}
          variant="filled"
          sx={{ width: '100%' }}
        >
          {snackbar.message}
        </MuiAlert>
      </Snackbar>
    </Box>
  )
}

export default SharedNetworkMembers
