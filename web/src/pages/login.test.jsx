import { describe, expect, test } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import Login from './Login'
import { AuthProvider } from '../services/auth'

describe('Login page', () => {
  test('shows register entry and does not expose setup entry', () => {
    const html = renderToStaticMarkup(
      <MemoryRouter>
        <AuthProvider>
          <Login />
        </AuthProvider>
      </MemoryRouter>,
    )

    expect(html).toContain('/register')
    expect(html).not.toContain('/setup')
    expect(html).toContain('没有账户? 去注册')
  })
})
