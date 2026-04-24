import { describe, expect, test } from 'bun:test'
import { formatUserTime, getUserRoleLabel } from './userPresentation'

describe('userPresentation', () => {
  test('formats role labels consistently', () => {
    expect(getUserRoleLabel('admin')).toBe('管理员')
    expect(getUserRoleLabel('user')).toBe('普通用户')
  })

  test('hides zero-value timestamps', () => {
    expect(formatUserTime('0001-01-01T00:00:00Z')).toBe('未知')
  })
})
