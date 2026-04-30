import type { IpAssignmentPool, Network, Route } from '../services/api'
import type {
  BasicSettingsDraft,
  DNSSettingsDraft,
  IPv4SettingsDraft,
  IPv6SettingsDraft,
  ManagedRoutesSettingsDraft,
  MulticastSettingsDraft,
} from '../components/network-detail/types'
import {
  getIPv4Pools,
  getIPv6Pools,
  getManagedRoutes,
  getNetworkDescription,
  getPrimaryIPv4Subnet,
  getPrimaryIPv6Subnet,
  normalizeIPv4CIDR,
  normalizeIPv6CIDR,
  normalizeRouteTarget,
} from './networkAddressUtils'

export function getInitialBasicSettings(network: Network): BasicSettingsDraft {
  return {
    name: network.name || '',
    description: getNetworkDescription(network),
  }
}

export function getInitialIPv4Settings(network: Network): IPv4SettingsDraft {
  const pools = getIPv4Pools(network.config.ipAssignmentPools || [])
  return {
    subnet: getPrimaryIPv4Subnet(network.config.routes || []),
    autoAssign: network.config.v4AssignMode?.zt ?? false,
    pools,
    poolStartDraft: pools[0]?.ipRangeStart ?? '',
    poolEndDraft: pools[0]?.ipRangeEnd ?? '',
  }
}

export function getInitialIPv6Settings(network: Network): IPv6SettingsDraft {
  const pools = getIPv6Pools(network.config.ipAssignmentPools || [])
  return {
    subnet: getPrimaryIPv6Subnet(network.config.routes || []),
    customAssign: network.config.v6AssignMode?.zt ?? false,
    rfc4193: network.config.v6AssignMode?.rfc4193 ?? false,
    plane6: network.config.v6AssignMode?.['6plane'] ?? false,
    pools,
    poolStartDraft: pools[0]?.ipRangeStart ?? '',
    poolEndDraft: pools[0]?.ipRangeEnd ?? '',
  }
}

export function getInitialDNSSettings(network: Network): DNSSettingsDraft {
  return {
    domain: network.config.dns?.domain || '',
    servers: network.config.dns?.servers || [],
    serverDraft: '',
  }
}

export function getInitialMulticastSettings(network: Network): MulticastSettingsDraft {
  return {
    multicastLimit: network.config.multicastLimit ?? 32,
    enableBroadcast: network.config.enableBroadcast ?? true,
  }
}

export function getInitialManagedRoutesSettings(network: Network): ManagedRoutesSettingsDraft {
  return {
    routes: getManagedRoutes(network.config.routes || []),
    routeDraft: { target: '', via: '' },
  }
}

export function buildMergedRoutes(
  currentRoutes: Route[],
  ipv4Subnet: string,
  ipv6Subnet: string,
  ipv6CustomAssign: boolean,
  managedRoutes: Route[],
): Route[] {
  const currentPrimaryIPv4Subnet = getPrimaryIPv4Subnet(currentRoutes)
  const currentPrimaryIPv6Subnet = getPrimaryIPv6Subnet(currentRoutes)
  const normalizedIPv4Subnet = normalizeIPv4CIDR(ipv4Subnet)
  const normalizedIPv6Subnet = normalizeIPv6CIDR(ipv6Subnet)

  return [
    ...(normalizedIPv4Subnet ? [{ target: normalizedIPv4Subnet }] : []),
    ...(ipv6CustomAssign && normalizedIPv6Subnet ? [{ target: normalizedIPv6Subnet }] : []),
    ...managedRoutes.map((route) => ({
      target: normalizeRouteTarget(route.target) as string,
      via: route.via?.trim() ? route.via.trim() : undefined,
    })),
    ...currentRoutes.filter((route) =>
      route.via ||
      (route.target !== currentPrimaryIPv4Subnet &&
        route.target !== currentPrimaryIPv6Subnet &&
        !managedRoutes.some((managedRoute) => normalizeRouteTarget(managedRoute.target) === normalizeRouteTarget(route.target)))
    ),
  ]
}

export function buildMergedPools(
  ipv4Enabled: boolean,
  ipv4Pools: IpAssignmentPool[],
  ipv6CustomAssign: boolean,
  ipv6Pools: IpAssignmentPool[],
): IpAssignmentPool[] {
  return [
    ...(ipv4Enabled ? ipv4Pools.map((pool) => ({
      ipRangeStart: pool.ipRangeStart.trim(),
      ipRangeEnd: pool.ipRangeEnd.trim(),
    })) : []),
    ...(ipv6CustomAssign ? ipv6Pools.map((pool) => ({
      ipRangeStart: pool.ipRangeStart.trim(),
      ipRangeEnd: pool.ipRangeEnd.trim(),
    })) : []),
  ]
}
