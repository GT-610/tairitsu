import { describe, expect, test } from 'bun:test'
import {
  handleUnauthorizedResponse,
  setupGateFailure,
  setupGateSuccess,
} from './useAppRuntime'

describe('app runtime state', () => {
  test('keeps setup state unknown when the status request fails', () => {
    expect(setupGateFailure()).toEqual({
      isFirstRun: null,
      loading: false,
      error: true,
    })
  })

  test('derives first-run state only from a successful backend response', () => {
    expect(setupGateSuccess({ initialized: false })).toEqual({
      isFirstRun: true,
      loading: false,
      error: false,
    })
    expect(setupGateSuccess({ initialized: true })).toEqual({
      isFirstRun: false,
      loading: false,
      error: false,
    })
  })

  test('clears auth and redirects on unauthorized responses', () => {
    let cleared = 0
    let redirected = 0
    const unauthorized = { isAxiosError: true, response: { status: 401 } }

    handleUnauthorizedResponse(unauthorized, () => { cleared += 1 }, () => { redirected += 1 }, '/networks')

    expect(cleared).toBe(1)
    expect(redirected).toBe(1)
  })

  test('does not redirect again when already on the login page', () => {
    let cleared = 0
    let redirected = 0
    const unauthorized = { isAxiosError: true, response: { status: 401 } }

    handleUnauthorizedResponse(unauthorized, () => { cleared += 1 }, () => { redirected += 1 }, '/login')

    expect(cleared).toBe(1)
    expect(redirected).toBe(0)
  })
})
