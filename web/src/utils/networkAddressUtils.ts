import type { IpAssignmentPool, Network, Route } from '../services/api'

export type ParsedIPv6CIDR = {
  network: bigint;
  prefix: number;
}

export function parseIPv4Address(value: string): number | null {
  const parts = value.trim().split('.')
  if (parts.length !== 4) return null

  let result = 0
  for (const part of parts) {
    if (!/^\d+$/.test(part)) return null
    const octet = Number(part)
    if (octet < 0 || octet > 255) return null
    result = (result << 8) + octet
  }

  return result >>> 0
}

export function parseIPv4CIDR(target: string): { network: number; mask: number } | null {
  const [address, prefixText] = target.trim().split('/')
  if (!address || !prefixText || !/^\d+$/.test(prefixText)) return null

  const prefix = Number(prefixText)
  if (prefix < 0 || prefix > 32) return null

  const ip = parseIPv4Address(address)
  if (ip === null) return null

  const mask = prefix === 0 ? 0 : ((0xffffffff << (32 - prefix)) >>> 0)
  return {
    network: ip & mask,
    mask,
  }
}

export function formatIPv4Address(value: number): string {
  return [
    (value >>> 24) & 255,
    (value >>> 16) & 255,
    (value >>> 8) & 255,
    value & 255,
  ].join('.')
}

export function parseIPv6Address(value: string): bigint | null {
  const input = value.trim().toLowerCase()
  if (input === '' || !/^[0-9a-f:]+$/.test(input)) return null
  if ((input.match(/::/g) || []).length > 1) return null

  const [leftPart, rightPart] = input.split('::')
  const left: string[] = leftPart ? leftPart.split(':').filter((part) => part !== '') : []
  const right: string[] = rightPart ? rightPart.split(':').filter((part) => part !== '') : []

  if (!input.includes('::')) {
    if (left.length !== 8) return null
  } else if (left.length + right.length >= 8) {
    return null
  }

  const missing = input.includes('::') ? 8 - (left.length + right.length) : 0
  const filledGroups: string[] = Array.from({ length: missing }, () => '0')
  const groups: string[] = input.includes('::') ? [...left, ...filledGroups, ...right] : left
  if (groups.length !== 8) return null

  let result = 0n
  for (const group of groups) {
    if (!/^[0-9a-f]{1,4}$/.test(group)) return null
    result = (result << 16n) + BigInt(parseInt(group, 16))
  }

  return result
}

export function formatIPv6Address(value: bigint): string {
  const groups: string[] = []
  for (let i = 7; i >= 0; i -= 1) {
    const group = Number((value >> BigInt(i * 16)) & 0xffffn)
    groups.push(group.toString(16))
  }
  return groups.join(':')
}

export function parseIPv6CIDR(target: string): ParsedIPv6CIDR | null {
  const [address, prefixText] = target.trim().split('/')
  if (!address || !prefixText || !/^\d+$/.test(prefixText)) return null

  const prefix = Number(prefixText)
  if (prefix < 0 || prefix > 128) return null

  const ip = parseIPv6Address(address)
  if (ip === null) return null

  const mask = prefix === 0 ? 0n : ((1n << BigInt(prefix)) - 1n) << BigInt(128 - prefix)
  return {
    network: ip & mask,
    prefix,
  }
}

export function normalizeIPv6CIDR(target: string): string | null {
  const parsed = parseIPv6CIDR(target)
  if (!parsed) return null
  return `${formatIPv6Address(parsed.network)}/${parsed.prefix}`
}

export function normalizeIPv4CIDR(target: string): string | null {
  const parsed = parseIPv4CIDR(target)
  if (!parsed) return null

  const prefix = Number(target.trim().split('/')[1])
  return `${formatIPv4Address(parsed.network)}/${prefix}`
}

export function getCIDRPrefix(target: string): number | null {
  const prefixText = target.trim().split('/')[1]
  if (!prefixText || !/^\d+$/.test(prefixText)) return null
  const prefix = Number(prefixText)
  if (prefix < 0 || prefix > 32) return null
  return prefix
}

export function getPrimaryIPv4Subnet(routes: Route[]): string {
  const primaryRoute = routes.find((route) => !route.via && parseIPv4CIDR(route.target))
  return primaryRoute?.target ?? ''
}

export function getPrimaryIPv6Subnet(routes: Route[]): string {
  const primaryRoute = routes.find((route) => !route.via && parseIPv6CIDR(route.target))
  return primaryRoute?.target ?? ''
}

export function getManagedRoutes(routes: Route[]): Route[] {
  const primaryIPv4 = getPrimaryIPv4Subnet(routes)
  const primaryIPv6 = getPrimaryIPv6Subnet(routes)
  return routes.filter((route) => route.target !== primaryIPv4 && route.target !== primaryIPv6)
}

export function buildAutoAssignPoolFromSubnet(subnet: string): IpAssignmentPool | null {
  const parsed = parseIPv4CIDR(subnet)
  const prefix = getCIDRPrefix(subnet)
  if (!parsed || prefix === null || prefix >= 31) return null

  const size = 2 ** (32 - prefix)
  const start = parsed.network + 1
  const end = parsed.network + size - 2
  if (start > end) return null

  return {
    ipRangeStart: formatIPv4Address(start >>> 0),
    ipRangeEnd: formatIPv4Address(end >>> 0),
  }
}

export function normalizeRouteTarget(target: string): string | null {
  return normalizeIPv4CIDR(target) ?? normalizeIPv6CIDR(target)
}

export function getIPv4Pools(pools: IpAssignmentPool[]): IpAssignmentPool[] {
  return pools.filter((pool) => parseIPv4Address(pool.ipRangeStart) !== null)
}

export function getIPv6Pools(pools: IpAssignmentPool[]): IpAssignmentPool[] {
  return pools.filter((pool) => parseIPv6Address(pool.ipRangeStart) !== null)
}

export function getNetworkDescription(network: Network): string {
  return network.db_description || network.description || ''
}
