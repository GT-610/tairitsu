import { describe, expect, test } from 'bun:test'
import { buildCreateUserSuccessMessage, buildDeleteUserSuccessMessage, buildResetPasswordSuccessMessage } from './userGovernance'

describe('userGovernance utils', () => {
  test('builds create-user success message', () => {
    expect(buildCreateUserSuccessMessage({
      message: 'ok',
      user: { id: '1', username: 'alice', role: 'user', createdAt: '', updatedAt: '' },
      temporary_password: 'secret',
    })).toContain('alice')
  })

  test('builds reset-password success message with revoked session count', () => {
    expect(buildResetPasswordSuccessMessage({
      message: 'ok',
      user: { id: '1', username: 'alice', role: 'user', createdAt: '', updatedAt: '' },
      temporary_password: 'secret',
      revoked_sessions: 3,
    })).toContain('3')
  })

  test('builds delete-user success message with transferred network count', () => {
    expect(buildDeleteUserSuccessMessage({
      message: 'ok',
      user: { id: '1', username: 'alice', role: 'user', createdAt: '', updatedAt: '' },
      transferred_networks: 2,
      revoked_sessions: 1,
    })).toContain('2')
  })
})
