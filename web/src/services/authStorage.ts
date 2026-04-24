import api, { type User, type UserSession } from './api'

export function persistAuthState(user: User, token: string, session: UserSession, rememberMe: boolean) {
  const storage = rememberMe ? localStorage : sessionStorage
  storage.setItem('user', JSON.stringify(user))
  storage.setItem('token', token)
  storage.setItem('session', JSON.stringify(session))
  api.defaults.headers.common.Authorization = `Bearer ${token}`
}

export function restoreAuthState() {
  const storedUser = localStorage.getItem('user') || sessionStorage.getItem('user')
  const storedToken = localStorage.getItem('token') || sessionStorage.getItem('token')
  const storedSession = localStorage.getItem('session') || sessionStorage.getItem('session')

  if (!storedUser || !storedToken) {
    return null
  }

  try {
    const user = JSON.parse(storedUser) as User
    const session = storedSession ? JSON.parse(storedSession) as UserSession : null
    api.defaults.headers.common.Authorization = `Bearer ${storedToken}`
    return { user, token: storedToken, session }
  } catch {
    clearPersistedAuthState()
    return null
  }
}

export function clearPersistedAuthState() {
  localStorage.removeItem('user')
  localStorage.removeItem('token')
  localStorage.removeItem('session')
  sessionStorage.removeItem('user')
  sessionStorage.removeItem('token')
  sessionStorage.removeItem('session')
  delete api.defaults.headers.common.Authorization
}
