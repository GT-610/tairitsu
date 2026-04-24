import { describe, expect, test } from 'bun:test'
import { getPlanetDownloadName, normalizePlanetEndpoints, validatePlanetEndpoints } from './planet'

describe('planet utils', () => {
  test('normalizes endpoint input by trimming and dropping blanks', () => {
    expect(normalizePlanetEndpoints([' 1.1.1.1/9993 ', '', '   ', '2001:db8::1/9993'])).toEqual([
      '1.1.1.1/9993',
      '2001:db8::1/9993',
    ])
  })

  test('validates required endpoints and endpoint format', () => {
    expect(validatePlanetEndpoints([])).toBe('至少需要填写一个端点')
    expect(validatePlanetEndpoints(['1.1.1.1'])).toBe('端点格式无效：1.1.1.1。应为 IP/Port 格式')
    expect(validatePlanetEndpoints(['1.1.1.1/70000'])).toBe('端口号无效：70000')
    expect(validatePlanetEndpoints(['300.1.1.1/9993'])).toBe('IP 地址无效：300.1.1.1')
  })

  test('accepts valid IPv4 and IPv6 endpoints and uses stable download name fallback', () => {
    expect(validatePlanetEndpoints(['203.0.113.1/9993', '2001:db8::1/9993'])).toBeNull()
    expect(getPlanetDownloadName()).toBe('planet')
    expect(getPlanetDownloadName('planet')).toBe('planet')
  })
})
