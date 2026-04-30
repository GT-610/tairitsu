import type { IpAssignmentPool, Route } from '../services/api'
import {
  normalizeIPv4CIDR,
  normalizeIPv6CIDR,
  normalizeRouteTarget,
  parseIPv4Address,
  parseIPv4CIDR,
  parseIPv6Address,
  parseIPv6CIDR,
} from './networkAddressUtils'

function isIPv4PoolCoveredByRoutes(pool: IpAssignmentPool, routes: Route[]): boolean {
  const start = parseIPv4Address(pool.ipRangeStart)
  const end = parseIPv4Address(pool.ipRangeEnd)
  if (start === null || end === null || start > end) return false

  return routes.some((route) => {
    const parsedRoute = parseIPv4CIDR(route.target)
    if (!parsedRoute) return false
    return (start & parsedRoute.mask) === parsedRoute.network && (end & parsedRoute.mask) === parsedRoute.network
  })
}

function isIPv6PoolInsideSubnet(pool: IpAssignmentPool, subnet: string): boolean {
  const parsedSubnet = parseIPv6CIDR(subnet)
  const start = parseIPv6Address(pool.ipRangeStart)
  const end = parseIPv6Address(pool.ipRangeEnd)
  if (!parsedSubnet || start === null || end === null || start > end) return false

  const mask = parsedSubnet.prefix === 0 ? 0n : ((1n << BigInt(parsedSubnet.prefix)) - 1n) << BigInt(128 - parsedSubnet.prefix)
  return (start & mask) === parsedSubnet.network && (end & mask) === parsedSubnet.network
}

function isIPv4PoolInsideSubnet(pool: IpAssignmentPool, subnet: string): boolean {
  return isIPv4PoolCoveredByRoutes(pool, [{ target: subnet }])
}

export function getIPv4RangeIssue(ipv4Subnet: string, ipPool: IpAssignmentPool | null): string | null {
  const trimmedSubnet = ipv4Subnet.trim()
  if (trimmedSubnet === '') return 'IPv4 子网不能为空。'

  const normalizedSubnet = normalizeIPv4CIDR(trimmedSubnet)
  if (!normalizedSubnet) return 'IPv4 子网必须是有效的 CIDR，例如 10.22.2.0/24。'

  const prefix = Number(normalizedSubnet.split('/')[1])
  if (prefix > 30) return 'IPv4 子网前缀需要在 0 到 30 之间，才能用于成员地址分配。'
  if (!ipPool) return '必须提供一个有效的 IPv4 地址池范围。'

  const start = ipPool.ipRangeStart.trim()
  const end = ipPool.ipRangeEnd.trim()
  if (start === '' || end === '') return '自动分配地址池需要同时填写起始 IP 和结束 IP。'

  const parsedStart = parseIPv4Address(start)
  const parsedEnd = parseIPv4Address(end)
  if (parsedStart === null || parsedEnd === null) return '自动分配地址池不是有效的 IPv4 范围。'
  if (parsedStart > parsedEnd) return '自动分配地址池的起始 IP 不能大于结束 IP。'
  if (!isIPv4PoolInsideSubnet(ipPool, normalizedSubnet)) return '自动分配地址池必须完全落在 IPv4 子网内。'

  return null
}

export function getIPv6RangeIssue(ipv6Subnet: string, ipPool: IpAssignmentPool | null): string | null {
  const trimmedSubnet = ipv6Subnet.trim()
  if (trimmedSubnet === '') return 'IPv6 子网不能为空。'

  const normalizedSubnet = normalizeIPv6CIDR(trimmedSubnet)
  if (!normalizedSubnet) return 'IPv6 子网必须是有效的 CIDR，例如 fd00::/48。'

  const parsedSubnet = parseIPv6CIDR(normalizedSubnet)
  if (!parsedSubnet || parsedSubnet.prefix > 127) return 'IPv6 子网前缀需要在 0 到 127 之间。'
  if (!ipPool) return '必须提供一个有效的 IPv6 地址池范围。'

  const start = ipPool.ipRangeStart.trim()
  const end = ipPool.ipRangeEnd.trim()
  if (start === '' || end === '') return '自定义 IPv6 地址池需要同时填写起始 IP 和结束 IP。'

  const parsedStart = parseIPv6Address(start)
  const parsedEnd = parseIPv6Address(end)
  if (parsedStart === null || parsedEnd === null) return '自定义 IPv6 地址池不是有效的 IPv6 范围。'
  if (parsedStart > parsedEnd) return '自定义 IPv6 地址池的起始 IP 不能大于结束 IP。'
  if (!isIPv6PoolInsideSubnet(ipPool, normalizedSubnet)) return '自定义 IPv6 地址池必须完全落在 IPv6 子网内。'

  return null
}

export function getIPv4PoolOverlapIssues(ipPools: IpAssignmentPool[]): string[] {
  const parsedPools = ipPools.map((pool, index) => ({
    index,
    start: parseIPv4Address(pool.ipRangeStart),
    end: parseIPv4Address(pool.ipRangeEnd),
  }))

  const issues: string[] = []
  for (let i = 0; i < parsedPools.length; i += 1) {
    const current = parsedPools[i]
    if (current.start === null || current.end === null) continue
    for (let j = i + 1; j < parsedPools.length; j += 1) {
      const next = parsedPools[j]
      if (next.start === null || next.end === null) continue
      if (current.start <= next.end && next.start <= current.end) {
        issues.push(`地址池 ${current.index + 1} 与地址池 ${next.index + 1} 存在重叠。`)
      }
    }
  }
  return issues
}

export function getIPv6PoolOverlapIssues(ipPools: IpAssignmentPool[]): string[] {
  const parsedPools = ipPools.map((pool, index) => ({
    index,
    start: parseIPv6Address(pool.ipRangeStart),
    end: parseIPv6Address(pool.ipRangeEnd),
  }))

  const issues: string[] = []
  for (let i = 0; i < parsedPools.length; i += 1) {
    const current = parsedPools[i]
    if (current.start === null || current.end === null) continue
    for (let j = i + 1; j < parsedPools.length; j += 1) {
      const next = parsedPools[j]
      if (next.start === null || next.end === null) continue
      if (current.start <= next.end && next.start <= current.end) {
        issues.push(`IPv6 地址池 ${current.index + 1} 与地址池 ${next.index + 1} 存在重叠。`)
      }
    }
  }
  return issues
}

export function getIPv4ConfigurationIssues(ipv4Subnet: string, enabled: boolean, ipPools: IpAssignmentPool[]): string[] {
  const trimmedSubnet = ipv4Subnet.trim()
  if (trimmedSubnet === '') return ['IPv4 子网不能为空。']

  const normalizedSubnet = normalizeIPv4CIDR(trimmedSubnet)
  if (!normalizedSubnet) return ['IPv4 子网必须是有效的 CIDR，例如 10.22.2.0/24。']

  const prefix = Number(normalizedSubnet.split('/')[1])
  if (prefix > 30) return ['IPv4 子网前缀需要在 0 到 30 之间，才能用于成员地址分配。']
  if (!enabled) return []
  if (ipPools.length === 0) return ['开启自动分配后，至少需要配置一个 IPv4 地址池范围。']

  const rangeIssues = ipPools.flatMap((ipPool, index) => {
    const issue = getIPv4RangeIssue(ipv4Subnet, ipPool)
    return issue ? [`地址池 ${index + 1}：${issue}`] : []
  })
  return [...rangeIssues, ...getIPv4PoolOverlapIssues(ipPools)]
}

export function getIPv6ConfigurationIssues(ipv6Subnet: string, enabled: boolean, ipv6Pools: IpAssignmentPool[]): string[] {
  if (!enabled) return []
  if (ipv6Pools.length === 0) return ['开启自定义 IPv6 范围后，至少需要配置一个 IPv6 地址池范围。']

  const rangeIssues = ipv6Pools.flatMap((pool, index) => {
    const issue = getIPv6RangeIssue(ipv6Subnet, pool)
    return issue ? [`IPv6 地址池 ${index + 1}：${issue}`] : []
  })
  return [...rangeIssues, ...getIPv6PoolOverlapIssues(ipv6Pools)]
}

function isValidRouteVia(via: string): boolean {
  const trimmed = via.trim()
  if (trimmed === '') return true
  return parseIPv4Address(trimmed) !== null || parseIPv6Address(trimmed) !== null
}

export function getManagedRouteIssue(route: Route): string | null {
  if (!normalizeRouteTarget(route.target)) return '目标网络必须是有效的 IPv4 或 IPv6 CIDR。'
  if (!isValidRouteVia(route.via || '')) return '下一跳必须是有效的 IPv4 或 IPv6 地址。'
  return null
}

export function isValidDnsServer(server: string): boolean {
  const trimmed = server.trim()
  if (trimmed === '') return false
  return isValidRouteVia(trimmed)
}
