import { describe, expect, test } from 'bun:test'
import {
  detectSystemLanguage,
  normalizeLanguagePreference,
  translateMessageCode,
  translateRawText,
} from './index'

describe('i18n language resolution', () => {
  test('uses Simplified Chinese for generic and non-traditional zh locales', () => {
    expect(detectSystemLanguage(['zh'])).toBe('zh-CN')
    expect(detectSystemLanguage(['zh-CN'])).toBe('zh-CN')
    expect(detectSystemLanguage(['zh-SG'])).toBe('zh-CN')
    expect(detectSystemLanguage(['zh-Hans'])).toBe('zh-CN')
    expect(detectSystemLanguage(['zh-Hans-SG'])).toBe('zh-CN')
    expect(detectSystemLanguage(['zh-TW'])).toBe('en')
    expect(detectSystemLanguage(['zh-HK'])).toBe('en')
    expect(detectSystemLanguage(['zh-Hant'])).toBe('en')
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
    expect(translateRawText('共享来源：alice 网络ID:abc 已授权 2 台', 'en')).toBe('Shared by: alice Network ID:abc Authorized 2 devices')
    expect(translateRawText('Shared by: alice Network ID:abc Authorized 2 devices', 'zh-CN')).toBe('共享来源：alice 网络ID:abc 已授权 2 台')
    expect(translateMessageCode('auth.invalid_token', 'zh-CN')).toBe('无效的认证令牌')
    expect(translateMessageCode('auth.invalid_token', 'en')).toBe('Invalid authentication token')
    expect(translateMessageCode('user.invalid_admin_operation', 'zh-CN')).toBe('该管理员操作不允许')
    expect(translateMessageCode('network.import_empty', 'zh-CN')).toBe('网络ID列表为空')
    expect(translateMessageCode('user.required', 'zh-CN')).toBe('必须指定用户')
  })
})
