import type { Member as ApiMember } from '../services/api'
import type { NetworkMemberDevice } from '../components/network-detail/types'

export function formatNetworkMember(member: ApiMember): NetworkMemberDevice {
  const memberID = member.id || ''

  return {
    id: memberID,
    name: member.name || member.id || '未命名设备',
    description: member.description || '',
    authorized: member.config?.authorized ?? member.authorized ?? false,
    ipAssignments: member.config?.ipAssignments ?? member.ipAssignments ?? [],
    clientVersion: member.clientVersion || 'unknown',
    address: member.address || '',
    identity: member.identity || '',
    online: member.online ?? false,
    creationTime: member.creationTime,
    tags: member.tags ?? [],
    capabilities: member.capabilities ?? [],
    peerVersion: member.peerVersion || '',
    peerRole: member.peerRole || '',
    peerLatency: member.peerLatency,
    preferredPath: member.preferredPath || '',
    activeBridge: member.config?.activeBridge ?? member.activeBridge ?? false,
    noAutoAssignIps: member.config?.noAutoAssignIps ?? member.noAutoAssignIps ?? false,
  }
}
