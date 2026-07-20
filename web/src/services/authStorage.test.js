import { describe, expect, test } from 'bun:test'
import { persistAuthStateToStores, restoreAuthStateFromStores } from './authStorage'

class MemoryStorage {
  values = new Map()

  getItem(key) {
    return this.values.get(key) ?? null
  }

  setItem(key, value) {
    this.values.set(key, value)
  }

  removeItem(key) {
    this.values.delete(key)
  }
}

const user = {
  id: 'user-1',
  username: 'alice',
  role: 'user',
  createdAt: '2026-07-20T00:00:00Z',
  updatedAt: '2026-07-20T00:00:00Z',
}

const session = {
  id: 'session-1',
  userAgent: 'test',
  ipAddress: '203.0.113.10',
  rememberMe: true,
  lastSeenAt: '2026-07-20T00:00:00Z',
  expiresAt: '2026-07-21T00:00:00Z',
  createdAt: '2026-07-20T00:00:00Z',
  updatedAt: '2026-07-20T00:00:00Z',
  current: true,
}

describe('auth storage', () => {
  test('stores a complete auth record in only the selected storage', () => {
    const persistent = new MemoryStorage()
    const transient = new MemoryStorage()
    transient.setItem('token', 'stale-token')

    persistAuthStateToStores(user, 'current-token', session, true, persistent, transient)

    expect(restoreAuthStateFromStores(persistent, transient)).toEqual({
      user,
      token: 'current-token',
      session,
    })
    expect(transient.getItem('token')).toBeNull()
  })

  test('switching remember-me storage clears the previous record', () => {
    const persistent = new MemoryStorage()
    const transient = new MemoryStorage()

    persistAuthStateToStores(user, 'persistent-token', session, true, persistent, transient)
    persistAuthStateToStores(user, 'transient-token', { ...session, rememberMe: false }, false, persistent, transient)

    expect(persistent.getItem('token')).toBeNull()
    expect(restoreAuthStateFromStores(persistent, transient)?.token).toBe('transient-token')
  })

  test('removes a stale duplicate record when both storage areas are populated', () => {
    const persistent = new MemoryStorage()
    const transient = new MemoryStorage()
    persistent.setItem('user', JSON.stringify(user))
    persistent.setItem('token', 'persistent-token')
    persistent.setItem('session', JSON.stringify(session))
    transient.setItem('user', JSON.stringify(user))
    transient.setItem('token', 'stale-transient-token')
    transient.setItem('session', JSON.stringify(session))

    expect(restoreAuthStateFromStores(persistent, transient)?.token).toBe('persistent-token')
    expect(transient.getItem('token')).toBeNull()
  })

  test('does not combine incomplete records from different storage areas', () => {
    const persistent = new MemoryStorage()
    const transient = new MemoryStorage()
    persistent.setItem('user', JSON.stringify(user))
    transient.setItem('token', 'cross-storage-token')
    transient.setItem('session', JSON.stringify(session))

    expect(restoreAuthStateFromStores(persistent, transient)).toBeNull()
    expect(persistent.getItem('user')).toBeNull()
    expect(transient.getItem('token')).toBeNull()
  })

  test('clears corrupt stored JSON', () => {
    const persistent = new MemoryStorage()
    const transient = new MemoryStorage()
    persistent.setItem('user', '{invalid')
    persistent.setItem('token', 'token')
    persistent.setItem('session', JSON.stringify(session))

    expect(restoreAuthStateFromStores(persistent, transient)).toBeNull()
    expect(persistent.getItem('token')).toBeNull()
  })

  test('rejects a session missing required contract fields', () => {
    const persistent = new MemoryStorage()
    const transient = new MemoryStorage()
    persistent.setItem('user', JSON.stringify(user))
    persistent.setItem('token', 'token')
    persistent.setItem('session', JSON.stringify({ id: 'session-1' }))

    expect(restoreAuthStateFromStores(persistent, transient)).toBeNull()
    expect(persistent.getItem('session')).toBeNull()
  })

  test('rejects a user missing required timestamp fields', () => {
    const persistent = new MemoryStorage()
    const transient = new MemoryStorage()
    const incompleteUser = { ...user }
    delete incompleteUser.createdAt
    persistent.setItem('user', JSON.stringify(incompleteUser))
    persistent.setItem('token', 'token')
    persistent.setItem('session', JSON.stringify(session))

    expect(restoreAuthStateFromStores(persistent, transient)).toBeNull()
    expect(persistent.getItem('user')).toBeNull()
  })
})
