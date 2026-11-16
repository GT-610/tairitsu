import { describe, expect, test } from 'bun:test'
import {
  adoptSavedConfig,
  buildMergedPools,
  buildMergedRoutes,
  emptyNetworkConfig,
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
  access: /** @type {import('../services/api').NetworkAccessInfo} */ ({ role: 'owner', permissions: ['network.read'] }),
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

describe('adoptSavedConfig', () => {
  /**
   * A complete Network, so the fixture cannot drift from the type the page holds.
   * @param {string} id
   * @param {Partial<import('../services/api').NetworkConfig>} config
   * @returns {import('../services/api').Network}
   */
  const networkFor = (id, config = {}) => ({
    id,
    name: id,
    description: '',
    // Only private and enableBroadcast are required; the rest are optional, which is
    // exactly why adoptSavedConfig fills them in.
    config: { private: true, enableBroadcast: true, routes: [], ipAssignmentPools: [], ...config },
    members: [],
    access: { role: 'owner', permissions: ['settings.write'] },
    status: 'OK',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  })

  /**
   * @param {import('../services/api').Network | null} previous
   * @param {import('../services/api').Network} saved
   * @param {string | null} settledNetworkId
   * @returns {import('../services/api').Network}
   */
  const adoptOrFail = (previous, saved, settledNetworkId) => {
    const adopted = adoptSavedConfig(previous, saved, settledNetworkId)
    if (!adopted) throw new Error('expected the response to be adopted')
    return adopted
  }

  test('adopts the response for the network that is still on screen', () => {
    const previous = networkFor('8056c2e21c000001')
    const saved = networkFor('8056c2e21c000001', { mtu: 1400 })

    const adopted = adoptOrFail(previous, saved, '8056c2e21c000001')

    expect(adopted.config.mtu).toBe(1400)
    expect(adopted.id).toBe('8056c2e21c000001')
  })

  test('keeps the parts of the snapshot the save did not return', () => {
    const previous = {
      ...networkFor('8056c2e21c000001'),
      description: 'kept',
      members: [{ id: 'a', name: 'a member' }],
    }
    const saved = networkFor('8056c2e21c000001', { mtu: 1400 })

    const adopted = adoptOrFail(previous, saved, '8056c2e21c000001')

    expect(adopted.description).toBe('kept')
    expect(adopted.members).toHaveLength(1)
    expect(adopted.access).toEqual({ role: 'owner', permissions: ['settings.write'] })
  })

  test('fills in the defaults a section reads regardless of the response', () => {
    const previous = networkFor('8056c2e21c000001')
    const saved = networkFor('8056c2e21c000001', { v4AssignMode: { zt: true } })

    const adopted = adoptOrFail(previous, saved, '8056c2e21c000001')

    expect(adopted.config.v4AssignMode).toEqual({ zt: true })
    expect(adopted.config.v6AssignMode).toEqual(emptyNetworkConfig.v6AssignMode)
  })

  // The regression: a save for a network the page has navigated away from must not be
  // written over the network now on screen, or the next save from it submits the
  // previous network's routes and pools.
  test('drops a response that settled after navigating to another network', () => {
    const onScreenNow = networkFor('8056c2e21c000002')
    const responseForPrevious = networkFor('8056c2e21c000001', { mtu: 1400 })

    expect(adoptSavedConfig(onScreenNow, responseForPrevious, '8056c2e21c000001')).toBeNull()
  })

  test('drops a response whose own id does not match the request', () => {
    const previous = networkFor('8056c2e21c000001')
    const mismatched = networkFor('8056c2e21c000002', { mtu: 1400 })

    expect(adoptSavedConfig(previous, mismatched, '8056c2e21c000001')).toBeNull()
  })

  test('drops a response when there is no snapshot yet', () => {
    expect(adoptSavedConfig(null, networkFor('8056c2e21c000001'), '8056c2e21c000001')).toBeNull()
  })
})
