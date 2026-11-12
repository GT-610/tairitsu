import type { NetworkRole } from '../services/api'

export const networkRoleLabel: Record<NetworkRole, string> = {
  owner: '所有者',
  manager: '网络管理员',
  operator: '运维',
  viewer: '查看者',
}

export const networkRoleColor: Record<NetworkRole, 'default' | 'primary' | 'secondary' | 'success'> = {
  owner: 'primary',
  manager: 'secondary',
  operator: 'success',
  viewer: 'default',
}

export function assignableNetworkRoles(actorRole: NetworkRole | undefined): NetworkRole[] {
  if (actorRole === 'owner') return ['manager', 'operator', 'viewer']
  if (actorRole === 'manager') return ['operator', 'viewer']
  return []
}
