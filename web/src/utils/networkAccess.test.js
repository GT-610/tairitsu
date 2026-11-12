import { describe, expect, test } from 'bun:test'
import { assignableNetworkRoles, networkRoleColor, networkRoleLabel } from './networkAccess'

describe('network access presentation', () => {
  test('labels and colors every role', () => {
    expect(Object.keys(networkRoleLabel).sort()).toEqual(['manager', 'operator', 'owner', 'viewer'])
    expect(Object.values(networkRoleLabel).every(Boolean)).toBe(true)
    expect(Object.values(networkRoleColor).every(Boolean)).toBe(true)
  })

  test('mirrors the backend assignment hierarchy', () => {
    expect(assignableNetworkRoles('owner')).toEqual(['manager', 'operator', 'viewer'])
    expect(assignableNetworkRoles('manager')).toEqual(['operator', 'viewer'])
    expect(assignableNetworkRoles('operator')).toEqual([])
    expect(assignableNetworkRoles('viewer')).toEqual([])
    expect(assignableNetworkRoles(undefined)).toEqual([])
  })
})
