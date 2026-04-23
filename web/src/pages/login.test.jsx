import { describe, expect, test } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import Login from './Login'
import { AuthProvider } from '../services/auth'

describe('Login page', () => {
  test('shows register entry, hides setup entry, and does not expose forgot-password route', () => {
    const html = renderToStaticMarkup(
      <MemoryRouter>
        <AuthProvider>
          <Login />
        </AuthProvider>
      </MemoryRouter>,
    )

    expect(html).toContain('/register')
    expect(html).not.toContain('/setup')
    expect(html).not.toContain('/forgot-password')
    expect(html).toContain('忘记密码请联系管理员处理')
    expect(html).toContain('没有账户? 去注册')
  })
})
