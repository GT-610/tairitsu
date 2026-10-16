import { describe, expect, test } from 'bun:test'
import api, { setSetupToken } from './api'

// The setup endpoints are closed to callers that cannot present the token the
// server prints, so the client has to send it on every request once the wizard
// has one -- and send nothing before that.
describe('setup token header', () => {
  /**
   * Runs one request through a stub adapter that reports the setup token the
   * interceptor placed on the outgoing headers.
   *
   * @param {() => Promise<unknown>} request
   * @returns {Promise<unknown>}
   */
  async function headerFor(request) {
    /** @type {unknown} */
    let observed
    // The stub returns a minimal response object, which axios's adapter type is
    // free to reject, so the swap goes through the untyped surface.
    const defaults = /** @type {any} */ (api.defaults)
    const originalAdapter = defaults.adapter
    defaults.adapter = async (/** @type {import('axios').AxiosRequestConfig} */ config) => {
      const headers = /** @type {import('axios').AxiosHeaders | undefined} */ (config.headers)
      observed = headers?.get('X-Setup-Token')
      return { data: {}, status: 200, statusText: 'OK', headers: {}, config }
    }

    try {
      await request()
    } finally {
      defaults.adapter = originalAdapter
    }

    return observed
  }

  test('is absent before the wizard has a token', async () => {
    setSetupToken('')
    expect(await headerFor(() => api.get('/system/status'))).toBeFalsy()
  })

  test('is attached after the wizard accepts one', async () => {
    setSetupToken('  token-value  ')
    try {
      expect(await headerFor(() => api.get('/system/status'))).toBe('token-value')
    } finally {
      setSetupToken('')
    }
  })
})
