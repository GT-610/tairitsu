import type { User } from '../services/api'

export function getUserRoleLabel(role?: User['role']): string {
  return role === 'admin' ? '管理员' : '普通用户'
}
