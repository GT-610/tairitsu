import { expect, test } from 'bun:test'
import { getErrorMessage } from './errors'

test('displays the original setup detail returned by the server', () => {
  const detail = 'failed to read token file: open /var/lib/zerotier-one/authtoken.secret: no such file or directory'
  const message = getErrorMessage({
    isAxiosError: true,
    response: {
      data: {
        error_code: 'setup.zerotier_config_save_failed',
        detail,
      },
    },
  }, 'fallback')

  expect(message).toContain(detail)
})
