import { describe, expect, test } from 'bun:test'
import { getRegistrationClosedMessage, isPublicRegistrationEnabled } from './publicRegistration'

describe('publicRegistration utils', () => {
  test('treats undefined and null as enabled by default', () => {
    expect(isPublicRegistrationEnabled(undefined)).toBe(true)
    expect(isPublicRegistrationEnabled(null)).toBe(true)
  })

  test('returns disabled state and message when public registration is off', () => {
    expect(isPublicRegistrationEnabled(false)).toBe(false)
    expect(getRegistrationClosedMessage(false)).toContain('公开注册已关闭')
  })
})
