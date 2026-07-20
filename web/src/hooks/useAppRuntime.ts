import { useCallback, useEffect, useRef, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import api, { type SetupStatus } from '../services/api'
import { useAuth } from '../services/auth'
import { hasStatus } from '../services/errors'
import { setupCompletedEvent } from '../pages/SetupWizard'

export interface SetupGateState {
  isFirstRun: boolean | null
  loading: boolean
  error: boolean
}

export function setupGateSuccess(status: SetupStatus): SetupGateState {
  return { isFirstRun: !status.initialized, loading: false, error: false }
}

export function setupGateFailure(): SetupGateState {
  return { isFirstRun: null, loading: false, error: true }
}

export function handleUnauthorizedResponse(
  error: unknown,
  clearAuth: () => void,
  navigateToLogin: () => void,
  pathname: string,
) {
  if (!hasStatus(error, 401)) return

  clearAuth()
  if (pathname !== '/login') {
    navigateToLogin()
  }
}

export function useSetupGate() {
  const [state, setState] = useState<SetupGateState>({
    isFirstRun: null,
    loading: true,
    error: false,
  })

  const refreshSetupStatus = useCallback(async () => {
    setState((current) => ({ ...current, loading: true, error: false }))
    try {
      const response = await api.get<SetupStatus>('/system/status', {
        headers: { 'Cache-Control': 'no-cache' },
      })
      setState(setupGateSuccess(response.data))
    } catch {
      setState(setupGateFailure())
    }
  }, [])

  useEffect(() => {
    void refreshSetupStatus()
  }, [refreshSetupStatus])

  useEffect(() => {
    const handleSetupComplete = () => {
      void refreshSetupStatus()
    }

    window.addEventListener(setupCompletedEvent, handleSetupComplete)
    return () => {
      window.removeEventListener(setupCompletedEvent, handleSetupComplete)
    }
  }, [refreshSetupStatus])

  return { ...state, retry: refreshSetupStatus }
}

export function useUnauthorizedRedirect() {
  const navigate = useNavigate()
  const location = useLocation()
  const { clearAuth } = useAuth()
  const locationRef = useRef(location)
  locationRef.current = location

  useEffect(() => {
    const interceptor = api.interceptors.response.use(
      (response) => response,
      (error) => {
        handleUnauthorizedResponse(
          error,
          clearAuth,
          () => { void navigate('/login') },
          locationRef.current.pathname,
        )

        return Promise.reject(error instanceof Error ? error : new Error(String(error)))
      },
    )

    return () => {
      api.interceptors.response.eject(interceptor)
    }
  }, [clearAuth, navigate])
}
