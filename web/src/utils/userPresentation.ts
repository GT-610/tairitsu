import type { User } from '../services/api'

export function getUserRoleLabel(role?: User['role']): string {
  return role === 'admin' ? '管理员' : '普通用户'
}

export function formatUserTime(value?: string): string {
  if (!value) {
    return '未知'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime()) || date.getFullYear() <= 1) {
    return '未知'
  }

  return date.toLocaleString()
}
