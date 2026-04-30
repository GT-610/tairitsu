import type { UserSession } from '../services/api'

interface SessionPresentation {
  title: string
  subtitle: string
  status: {
    label: string
    severity: 'success' | 'info' | 'warning' | 'error'
  }
  details: string[]
}

export function hasDisplayableSessionTime(value?: string | null): boolean {
  if (!value) {
    return false
  }

  const date = new Date(value)
  return !Number.isNaN(date.getTime()) && date.getFullYear() > 1
}

export function formatSessionTime(value: string): string {
  if (!hasDisplayableSessionTime(value)) {
    return '未知'
  }

  return new Date(value).toLocaleString()
}

function detectBrowser(userAgent: string): string {
  if (!userAgent) return '未知浏览器'
  if (/Edg\//i.test(userAgent)) return 'Edge'
  if (/OPR\//i.test(userAgent)) return 'Opera'
  if (/Chrome\//i.test(userAgent) && !/Edg\//i.test(userAgent)) return 'Chrome'
  if (/Firefox\//i.test(userAgent)) return 'Firefox'
  if (/Safari\//i.test(userAgent) && !/Chrome\//i.test(userAgent)) return 'Safari'
  return '未知浏览器'
}

function detectPlatform(userAgent: string): string {
  if (!userAgent) return '未知系统'
  if (/Windows NT/i.test(userAgent)) return 'Windows'
  if (/Android/i.test(userAgent)) return 'Android'
  if (/(iPhone|iPad|iPod)/i.test(userAgent)) return 'iOS'
  if (/Mac OS X/i.test(userAgent)) return 'macOS'
  if (/Linux/i.test(userAgent)) return 'Linux'
  return '未知系统'
}

function detectDeviceType(userAgent: string): string {
  if (!userAgent) return '未知设备'
  if (/iPad|Tablet/i.test(userAgent)) return '平板设备'
  if (/Mobile|iPhone|Android/i.test(userAgent)) return '移动设备'
  return '桌面设备'
}

function resolveSessionStatus(sessionItem: UserSession): SessionPresentation['status'] {
  if (sessionItem.current) {
    return { label: '当前会话', severity: 'success' }
  }
  if (sessionItem.revokedAt) {
    return { label: '已移除', severity: 'warning' }
  }
  if (new Date(sessionItem.expiresAt).getTime() <= Date.now()) {
    return { label: '已过期', severity: 'error' }
  }
  return { label: '其他会话', severity: 'info' }
}

export function formatSessionPresentation(sessionItem: UserSession): SessionPresentation {
  const browser = detectBrowser(sessionItem.userAgent)
  const platform = detectPlatform(sessionItem.userAgent)
  const deviceType = detectDeviceType(sessionItem.userAgent)
  const status = resolveSessionStatus(sessionItem)
  const details = [
    `IP：${sessionItem.ipAddress || '未知'}`,
    sessionItem.rememberMe ? '持久登录' : '会话登录',
    deviceType,
  ]

  return {
    title: `${browser} · ${platform}`,
    subtitle: details.join(' · '),
    status,
    details: sessionItem.userAgent ? [`客户端标识：${sessionItem.userAgent}`] : [],
  }
}
