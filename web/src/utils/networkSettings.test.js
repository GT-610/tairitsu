import { describe, expect, test } from 'bun:test'
import {
  buildMergedPools,
  buildMergedRoutes,
  getInitialManagedRoutesSettings,
} from './networkSettings'

const baseNetwork = {
  id: 'f76fd3000b86b177',
  name: 'test-network',
  config: {
    private: true,
    enableBroadcast: true,
    routes: [
      { target: '10.0.0.0/24' },
      { target: 'fd00::/64' },
      { target: '10.1.0.0/24', via: '10.0.0.1' },
    ],
    ipAssignmentPools: [
      { ipRangeStart: '10.0.0.10', ipRangeEnd: '10.0.0.20' },
      { ipRangeStart: 'fd00::10', ipRangeEnd: 'fd00::20' },
    ],
  },
  members: [],
  status: 'OK',
  createdAt: '',
  updatedAt: '',
}

describe('networkSettings mapping', () => {
  test('extracts managed routes without touching primary subnets', () => {
    const managed = getInitialManagedRoutesSettings(baseNetwork)
    expect(managed.routes).toEqual([{ target: '10.1.0.0/24', via: '10.0.0.1' }])
  })

  test('merges routes without dropping unrelated managed routes', () => {
    const routes = buildMergedRoutes(
      baseNetwork.config.routes || [],
      '10.2.0.0/24',
      'fd00:2::/64',
      true,
      [{ target: '10.3.0.0/24', via: '10.2.0.1' }],
    )

    expect(routes).toEqual([
      { target: '10.2.0.0/24' },
      { target: 'fd00:2:0:0:0:0:0:0/64' },
      { target: '10.3.0.0/24', via: '10.2.0.1' },
      { target: '10.1.0.0/24', via: '10.0.0.1' },
    ])
  })

  test('merges IPv4 and IPv6 pools without overwriting either side', () => {
    const pools = buildMergedPools(
      true,
      [{ ipRangeStart: '10.2.0.10', ipRangeEnd: '10.2.0.20' }],
      true,
      [{ ipRangeStart: 'fd00:2::10', ipRangeEnd: 'fd00:2::20' }],
    )

    expect(pools).toEqual([
      { ipRangeStart: '10.2.0.10', ipRangeEnd: '10.2.0.20' },
      { ipRangeStart: 'fd00:2::10', ipRangeEnd: 'fd00:2::20' },
    ])
  })
})
