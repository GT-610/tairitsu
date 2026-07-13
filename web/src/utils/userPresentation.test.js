import { describe, expect, test } from 'bun:test'
import { getUserRoleLabel } from './userPresentation'
import { formatDisplayableTime, hasDisplayableTime } from './timePresentation'

describe('userPresentation', () => {
  test('formats role labels consistently', () => {
    expect(getUserRoleLabel('admin')).toBe('管理员')
    expect(getUserRoleLabel('user')).toBe('普通用户')
  })

  test('hides zero-value timestamps', () => {
    expect(hasDisplayableTime('0001-01-01T00:00:00Z')).toBe(false)
    expect(formatDisplayableTime('0001-01-01T00:00:00Z')).toBe('未知')
  })
})
