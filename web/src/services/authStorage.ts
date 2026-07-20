import api, { type User, type UserSession } from './api'

interface AuthStorage {
  getItem(key: string): string | null
  setItem(key: string, value: string): void
  removeItem(key: string): void
}

interface RestoredAuthState {
  user: User
  token: string
  session: UserSession
}

const authStorageKeys = ['user', 'token', 'session'] as const

function clearAuthStorage(storage: AuthStorage) {
  for (const key of authStorageKeys) {
    storage.removeItem(key)
  }
}

function isUser(value: unknown): value is User {
  if (typeof value !== 'object' || value === null) return false
  const user = value as Partial<User>
  return typeof user.id === 'string'
    && typeof user.username === 'string'
    && (user.role === 'admin' || user.role === 'user')
    && typeof user.createdAt === 'string'
    && typeof user.updatedAt === 'string'
}

function isUserSession(value: unknown): value is UserSession {
  if (typeof value !== 'object' || value === null) return false
  const session = value as Partial<UserSession>
  return typeof session.id === 'string'
    && session.id.trim() !== ''
    && typeof session.userAgent === 'string'
    && typeof session.ipAddress === 'string'
    && typeof session.rememberMe === 'boolean'
    && typeof session.lastSeenAt === 'string'
    && typeof session.expiresAt === 'string'
    && (session.revokedAt === undefined || session.revokedAt === null || typeof session.revokedAt === 'string')
    && typeof session.createdAt === 'string'
    && typeof session.updatedAt === 'string'
    && typeof session.current === 'boolean'
}

function readAuthStorage(storage: AuthStorage): RestoredAuthState | null {
  const storedUser = storage.getItem('user')
  const storedToken = storage.getItem('token')
  const storedSession = storage.getItem('session')
  const hasStoredAuth = storedUser !== null || storedToken !== null || storedSession !== null

  if (!hasStoredAuth) return null

  if (!storedUser || !storedToken || !storedSession) {
    clearAuthStorage(storage)
    return null
  }

  try {
    const user: unknown = JSON.parse(storedUser)
    const session: unknown = JSON.parse(storedSession)
    if (!isUser(user) || !isUserSession(session)) {
      clearAuthStorage(storage)
      return null
    }
    return { user, token: storedToken, session }
  } catch {
    clearAuthStorage(storage)
    return null
  }
}

export function persistAuthStateToStores(
  user: User,
  token: string,
  session: UserSession,
  rememberMe: boolean,
  persistentStorage: AuthStorage,
  transientStorage: AuthStorage,
) {
  clearAuthStorage(persistentStorage)
  clearAuthStorage(transientStorage)

  const targetStorage = rememberMe ? persistentStorage : transientStorage
  targetStorage.setItem('user', JSON.stringify(user))
  targetStorage.setItem('token', token)
  targetStorage.setItem('session', JSON.stringify(session))
}

export function restoreAuthStateFromStores(
  persistentStorage: AuthStorage,
  transientStorage: AuthStorage,
): RestoredAuthState | null {
  const persistentState = readAuthStorage(persistentStorage)
  if (persistentState) {
    clearAuthStorage(transientStorage)
    return persistentState
  }

  const transientState = readAuthStorage(transientStorage)
  if (transientState) {
    clearAuthStorage(persistentStorage)
  }
  return transientState
}

export function persistAuthState(user: User, token: string, session: UserSession, rememberMe: boolean) {
  persistAuthStateToStores(user, token, session, rememberMe, localStorage, sessionStorage)
  api.defaults.headers.common.Authorization = `Bearer ${token}`
}

export function restoreAuthState() {
  const restored = restoreAuthStateFromStores(localStorage, sessionStorage)
  if (restored) {
    api.defaults.headers.common.Authorization = `Bearer ${restored.token}`
  }
  return restored
}

export function clearPersistedAuthState() {
  clearAuthStorage(localStorage)
  clearAuthStorage(sessionStorage)
  delete api.defaults.headers.common.Authorization
}
