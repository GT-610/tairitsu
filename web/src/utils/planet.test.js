import { describe, expect, test } from 'bun:test'
import {
  findDuplicatePlanetEndpoints,
  getPlanetDownloadName,
  normalizePlanetEndpoints,
  parsePlanetIdentityPublic,
  validatePlanetEndpointValue,
  validatePlanetEndpoints,
  validatePlanetRootNodes,
} from './planet'

describe('planet utils', () => {
  test('normalizes endpoint input by trimming and dropping blanks', () => {
    expect(normalizePlanetEndpoints([' 1.1.1.1/9993 ', '', '   ', '2001:db8::1/9993'])).toEqual([
      '1.1.1.1/9993',
      '2001:db8::1/9993',
    ])
  })

  test('validates required endpoints and endpoint format', () => {
    expect(validatePlanetEndpoints([])).toBe('至少需要填写一个端点')
    expect(validatePlanetEndpoints(['1.1.1.1'])).toBe('1.1.1.1：格式应为 IP/Port')
    expect(validatePlanetEndpoints(['1.1.1.1/70000'])).toBe('1.1.1.1/70000：端口号必须在 1-65535 之间')
    expect(validatePlanetEndpoints(['300.1.1.1/9993'])).toBe('300.1.1.1/9993：IP 地址格式无效')
  })

  test('accepts valid IPv4 and IPv6 endpoints and uses stable download name fallback', () => {
    expect(validatePlanetEndpoints(['203.0.113.1/9993', '2001:db8::1/9993'])).toBeNull()
    expect(getPlanetDownloadName()).toBe('planet')
    expect(getPlanetDownloadName('planet')).toBe('planet')
  })

  test('detects duplicate endpoints and validates per-field values', () => {
    expect(validatePlanetEndpointValue('')).toBe('请输入一个 stable endpoint')
    expect(validatePlanetEndpointValue('203.0.113.1/9993')).toBeNull()
    expect(validatePlanetEndpoints(['203.0.113.1/9993', '203.0.113.1/9993'])).toBe('端点重复：203.0.113.1/9993')
    expect([...findDuplicatePlanetEndpoints(['203.0.113.1/9993', '203.0.113.1/9993'])]).toEqual(['203.0.113.1/9993'])
  })

  test('parses a real identity.public string into summary fields', () => {
    expect(parsePlanetIdentityPublic('f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715')).toEqual({
      address: 'f76fd3000b',
      publicKey: '542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715',
      publicKeyBytes: 64,
    })
    expect(parsePlanetIdentityPublic('invalid')).toBeNull()
  })

  test('validates advanced root node configuration', () => {
    expect(validatePlanetRootNodes([])).toBe('至少需要配置一个 root node')
    expect(validatePlanetRootNodes([
      {
        identityPublic: 'invalid',
        endpoints: ['203.0.113.1/9993'],
      },
    ])).toBe('存在格式无效的 identity.public')
    expect(validatePlanetRootNodes([
      {
        identityPublic: 'f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715',
        endpoints: ['203.0.113.1/9993'],
      },
      {
        identityPublic: 'f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715',
        endpoints: ['203.0.113.2/9993'],
      },
    ])).toBe('root node identity 不能重复')
  })
})
