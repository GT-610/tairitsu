import type { Member as ApiMember } from '../services/api'
import type { NetworkMemberDevice } from '../components/network-detail/types'

export function formatNetworkMember(member: ApiMember): NetworkMemberDevice {
  const memberID = member.id || ''
  const ipAssignments = member.config?.ipAssignments ?? member.ipAssignments ?? []
  const tags = member.tags ?? []
  const capabilities = member.capabilities ?? []

  return {
    id: memberID,
    name: member.name || member.id || '',
    description: member.description || '',
    authorized: member.config?.authorized ?? member.authorized ?? false,
    ipAssignments: Array.from(ipAssignments),
    clientVersion: member.clientVersion || 'unknown',
    address: member.address || '',
    identity: member.identity || '',
    online: member.online ?? false,
    creationTime: member.creationTime,
    tags: Array.from(tags),
    capabilities: Array.from(capabilities),
    peerVersion: member.peerVersion || '',
    peerRole: member.peerRole || '',
    peerLatency: member.peerLatency,
    preferredPath: member.preferredPath || '',
    activeBridge: member.config?.activeBridge ?? member.activeBridge ?? false,
    noAutoAssignIps: member.config?.noAutoAssignIps ?? member.noAutoAssignIps ?? false,
  }
}
