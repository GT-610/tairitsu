import type { User } from '../services/api'

export function getUserRoleLabel(role?: User['role']): string {
  return role === 'admin' ? '管理员' : '普通用户'
}

export function hasDisplayableUserTime(value?: string): boolean {
  if (!value) {
    return false
  }

  const date = new Date(value)
  return !Number.isNaN(date.getTime()) && date.getFullYear() > 1
}

export function formatUserTime(value?: string): string {
  if (!hasDisplayableUserTime(value)) {
    return '未知'
  }

  return new Date(value as string).toLocaleString()
}
