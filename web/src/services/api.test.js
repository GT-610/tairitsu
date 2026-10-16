import { describe, expect, test } from 'bun:test'
import api, { setSetupToken } from './api'

// The setup endpoints are closed to callers that cannot present the token the
// server prints, so the client has to send it on every request once the wizard
// has one -- and send nothing before that.
describe('setup token header', () => {
  async function headerFor(request) {
    let observed
    const originalAdapter = api.defaults.adapter
    api.defaults.adapter = async (config) => {
      observed = config.headers['X-Setup-Token']
      return { data: {}, status: 200, statusText: 'OK', headers: {}, config }
    }

    try {
      await request()
    } finally {
      api.defaults.adapter = originalAdapter
    }

    return observed
  }

  test('is absent before the wizard has a token', async () => {
    setSetupToken('')
    expect(await headerFor(() => api.get('/system/status'))).toBeUndefined()
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
