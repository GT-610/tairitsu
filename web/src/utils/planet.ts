export function normalizePlanetEndpoints(values: string[]): string[] {
  return values.map((value) => value.trim()).filter(Boolean)
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

export function validatePlanetEndpoints(values: string[]): string | null {
  const endpoints = normalizePlanetEndpoints(values)
  if (endpoints.length === 0) {
    return '至少需要填写一个端点'
  }

  for (const endpoint of endpoints) {
    const separatorIndex = endpoint.lastIndexOf('/')
    if (separatorIndex <= 0 || separatorIndex === endpoint.length - 1) {
      return `端点格式无效：${endpoint}。应为 IP/Port 格式`
    }

    const host = endpoint.slice(0, separatorIndex)
    const portText = endpoint.slice(separatorIndex + 1)
    const port = Number(portText)
    if (!Number.isInteger(port) || port < 1 || port > 65535) {
      return `端口号无效：${portText}`
    }

    if (!isValidIPv4(host) && !isValidIPv6(host)) {
      return `IP 地址无效：${host}`
    }
  }

  return null
}

export function getPlanetDownloadName(downloadName?: string): string {
  return downloadName?.trim() || 'planet'
}
