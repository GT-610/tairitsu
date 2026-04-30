export function normalizePlanetEndpoints(values: string[]): string[] {
  return values.map((value) => value.trim()).filter(Boolean)
}

interface PlanetIdentitySummary {
  address: string;
  publicKey: string;
  publicKeyBytes: number;
}

interface PlanetRootNodeDraft {
  identityPublic: string;
  endpoints: string[];
}

function isValidIPv4(value: string): boolean {
  const parts = value.split('.')
  if (parts.length !== 4) {
    return false
  }

  return parts.every((part) => {
    if (!/^\d+$/.test(part)) {
      return false
    }
    const number = Number(part)
    return number >= 0 && number <= 255
  })
}

function isValidIPv6(value: string): boolean {
  return value.includes(':') && /^[0-9a-fA-F:]+$/.test(value)
}

export function parsePlanetIdentityPublic(value: string): PlanetIdentitySummary | null {
  const trimmed = value.trim()
  const parts = trimmed.split(':')
  if (parts.length !== 3) {
    return null
  }

  const [address, separator, publicKey] = parts
  if (!/^[0-9a-fA-F]{10}$/.test(address) || separator !== '0' || !/^[0-9a-fA-F]{128}$/.test(publicKey)) {
    return null
  }

  return {
    address: address.toLowerCase(),
    publicKey,
    publicKeyBytes: publicKey.length / 2,
  }
}

export function validatePlanetEndpointValue(value: string): string | null {
  const endpoint = value.trim()
  if (!endpoint) {
    return '请输入一个 stable endpoint'
  }

  const separatorIndex = endpoint.lastIndexOf('/')
  if (separatorIndex <= 0 || separatorIndex === endpoint.length - 1) {
    return '格式应为 IP/Port'
  }

  const host = endpoint.slice(0, separatorIndex)
  const portText = endpoint.slice(separatorIndex + 1)
  const port = Number(portText)
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    return '端口号必须在 1-65535 之间'
  }

  if (!isValidIPv4(host) && !isValidIPv6(host)) {
    return 'IP 地址格式无效'
  }

  return null
}

export function findDuplicatePlanetEndpoints(values: string[]): Set<string> {
  const duplicates = new Set<string>()
  const seen = new Set<string>()

  for (const value of values) {
    const trimmed = value.trim()
    if (!trimmed) {
      continue
    }

    if (seen.has(trimmed)) {
      duplicates.add(trimmed)
      continue
    }
    seen.add(trimmed)
  }

  return duplicates
}

export function validatePlanetEndpoints(values: string[]): string | null {
  const endpoints = normalizePlanetEndpoints(values)
  if (endpoints.length === 0) {
    return '至少需要填写一个端点'
  }

  for (const endpoint of endpoints) {
    const endpointError = validatePlanetEndpointValue(endpoint)
    if (endpointError) {
      return `${endpoint}：${endpointError}`
    }
  }

  const duplicates = findDuplicatePlanetEndpoints(endpoints)
  if (duplicates.size > 0) {
    const [firstDuplicate] = [...duplicates]
    return `端点重复：${firstDuplicate}`
  }

  return null
}

export function getPlanetDownloadName(downloadName?: string): string {
  return downloadName?.trim() || 'planet'
}

export function validatePlanetRootNodes(rootNodes: PlanetRootNodeDraft[]): string | null {
  if (rootNodes.length === 0) {
    return '至少需要配置一个 root node'
  }

  const seenIdentities = new Set<string>()
  for (const rootNode of rootNodes) {
    const identity = rootNode.identityPublic.trim()
    if (!identity) {
      return '每个 root node 都需要 identity.public'
    }
    if (!parsePlanetIdentityPublic(identity)) {
      return '存在格式无效的 identity.public'
    }
    if (seenIdentities.has(identity)) {
      return 'root node identity 不能重复'
    }
    seenIdentities.add(identity)

    const endpointError = validatePlanetEndpoints(rootNode.endpoints)
    if (endpointError) {
      return endpointError
    }
  }

  return null
}
