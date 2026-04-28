import { describe, expect, test } from 'bun:test'
import {
  detectSystemLanguage,
  normalizeLanguagePreference,
  translateMessageCode,
  translateRawText,
} from './index'

describe('i18n language resolution', () => {
  test('uses Simplified Chinese only for explicit zh-CN or zh-Hans locales', () => {
    expect(detectSystemLanguage(['zh-CN'])).toBe('zh-CN')
    expect(detectSystemLanguage(['zh-Hans'])).toBe('zh-CN')
    expect(detectSystemLanguage(['zh-Hans-SG'])).toBe('zh-CN')
    expect(detectSystemLanguage(['zh-TW'])).toBe('en')
    expect(detectSystemLanguage(['en-US'])).toBe('en')
  })

  test('normalizes stored language preferences', () => {
    expect(normalizeLanguagePreference('system')).toBe('system')
    expect(normalizeLanguagePreference('en')).toBe('en')
    expect(normalizeLanguagePreference('zh-CN')).toBe('zh-CN')
    expect(normalizeLanguagePreference('fr')).toBe('system')
  })

  test('translates raw UI copy and backend message codes', () => {
    expect(translateRawText('设置', 'en')).toBe('Settings')
    expect(translateRawText('Settings', 'zh-CN')).toBe('设置')
    expect(translateMessageCode('auth.invalid_token', 'zh-CN')).toBe('无效的认证令牌')
    expect(translateMessageCode('auth.invalid_token', 'en')).toBe('Invalid authentication token')
  })
})
