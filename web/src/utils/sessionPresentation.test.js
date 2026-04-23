import { describe, expect, test } from 'bun:test'
import { formatSessionPresentation } from './sessionPresentation'

describe('sessionPresentation', () => {
  test('formats readable browser and platform labels', () => {
    const result = formatSessionPresentation({
      id: 's1',
      userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36',
      ipAddress: '127.0.0.1',
      rememberMe: true,
      lastSeenAt: new Date().toISOString(),
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      current: false,
    })

    expect(result.title).toBe('Chrome · macOS')
    expect(result.subtitle).toContain('持久登录')
    expect(result.subtitle).toContain('桌面设备')
    expect(result.status.label).toBe('其他会话')
  })

  test('marks revoked and current sessions distinctly', () => {
    const revoked = formatSessionPresentation({
      id: 's2',
      userAgent: '',
      ipAddress: '',
      rememberMe: false,
      lastSeenAt: new Date().toISOString(),
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      revokedAt: new Date().toISOString(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      current: false,
    })
    const current = formatSessionPresentation({
      id: 's3',
      userAgent: '',
      ipAddress: '',
      rememberMe: false,
      lastSeenAt: new Date().toISOString(),
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      current: true,
    })

    expect(revoked.status.label).toBe('已移除')
    expect(current.status.label).toBe('当前会话')
  })
})
