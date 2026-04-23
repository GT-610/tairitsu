import { describe, expect, test } from 'bun:test'
import { getIPv4PoolOverlapIssues, getIPv6PoolOverlapIssues } from './networkValidation'

describe('networkValidation overlap checks', () => {
  test('detects overlapping IPv4 pools', () => {
    const issues = getIPv4PoolOverlapIssues([
      { ipRangeStart: '10.0.0.1', ipRangeEnd: '10.0.0.20' },
      { ipRangeStart: '10.0.0.10', ipRangeEnd: '10.0.0.30' },
    ])

    expect(issues).toHaveLength(1)
    expect(issues[0]).toContain('地址池 1 与地址池 2 存在重叠')
  })

  test('detects overlapping IPv6 pools', () => {
    const issues = getIPv6PoolOverlapIssues([
      { ipRangeStart: 'fd00::1', ipRangeEnd: 'fd00::10' },
      { ipRangeStart: 'fd00::8', ipRangeEnd: 'fd00::20' },
    ])

    expect(issues).toHaveLength(1)
    expect(issues[0]).toContain('IPv6 地址池 1 与地址池 2 存在重叠')
  })
})
