import { useCallback, useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import api, { type SetupStatus } from '../services/api'
import { clearPersistedAuthState } from '../services/authStorage'
import { hasStatus } from '../services/errors'
import { setupCompletedEvent } from '../pages/SetupWizard'

export function useSetupGate() {
  const [isFirstRun, setIsFirstRun] = useState<boolean | null>(null)
  const [loading, setLoading] = useState(true)

  const refreshSetupStatus = useCallback(async () => {
    try {
      const response = await api.get<SetupStatus>('/system/status', {
        headers: { 'Cache-Control': 'no-cache' },
      })
      setIsFirstRun(!response.data.initialized)
    } catch {
      setIsFirstRun(true)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refreshSetupStatus()
  }, [refreshSetupStatus])

  useEffect(() => {
    const handleSetupComplete = () => {
      setLoading(true)
      void refreshSetupStatus()
    }

    window.addEventListener(setupCompletedEvent, handleSetupComplete)
    return () => {
      window.removeEventListener(setupCompletedEvent, handleSetupComplete)
    }
  }, [refreshSetupStatus])

  return { isFirstRun, loading }
}

export function useUnauthorizedRedirect() {
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    const interceptor = api.interceptors.response.use(
      (response) => response,
      (error) => {
        if (hasStatus(error, 401)) {
          clearPersistedAuthState()
          if (location.pathname !== '/login') {
            void navigate('/login')
          }
        }

        return Promise.reject(error instanceof Error ? error : new Error(String(error)))
      },
    )

    return () => {
      api.interceptors.response.eject(interceptor)
    }
  }, [location.pathname, navigate])
}
