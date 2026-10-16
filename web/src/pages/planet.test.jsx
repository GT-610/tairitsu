import { describe, expect, mock, test } from 'bun:test'
import { render, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useTestDOM } from '../test/dom'

useTestDOM()

// The page reads the controller's identity through this module. Mocking it lets
// a test hold the request open and observe the loading branch, which a static
// render can never reach.
const apiMock = { getIdentity: mock() }

mock.module('../services/api', () => ({
  planetAPI: apiMock,
}))

const { default: Planet } = await import('./Planet')

describe('Planet page', () => {
  test('renders the default generation workflow without the removed header export', () => {
    const { container, getByRole, getByText } = render(<Planet />)

    expect(getByRole('button', { name: '读取身份' })).toBeDefined()
    expect(getByText('身份加载')).toBeDefined()
    expect(getByText('Planet 配置')).toBeDefined()
    expect(getByText('高级模式')).toBeDefined()
    expect(
      container.querySelector('input[placeholder="留空时使用控制器令牌文件所在目录"]'),
    ).not.toBeNull()

    // The C header export and both idle-feedback strings were deleted. This is
    // the default branch, which is the only state a static render can reach.
    expect(container.innerHTML).not.toContain('复制 C 头文件')
    expect(container.innerHTML).not.toContain('已读取真实 identity.public，可继续填写默认模式配置')
    expect(container.innerHTML).not.toContain('成功读取后会显示当前 root identity')
  })

  test('shows progress while the identity loads and the parsed address once it resolves', async () => {
    // user-event captures a default document when its module first loads, which
    // is before this file's DOM exists, so it has to be handed one explicitly.
    const user = userEvent.setup({ document })

    // The mock assigns this synchronously when the page issues the request, so
    // the guard below only exists to satisfy the type checker.
    /** @type {((value: unknown) => void) | undefined} */
    let resolveIdentity
    apiMock.getIdentity = mock(
      () => new Promise((resolve) => {
        resolveIdentity = (value) => resolve(value)
      }),
    )

    const { getByRole, getByText, queryByText } = render(<Planet />)
    await user.click(getByRole('button', { name: '读取身份' }))

    // The button swaps in the in-flight label, and the parsed-identity fields
    // are withheld until the response arrives.
    expect(queryByText('读取中...')).not.toBeNull()
    expect(queryByText('节点地址')).toBeNull()

    if (!resolveIdentity) {
      throw new Error('the page never called getIdentity')
    }
    resolveIdentity({
      data: {
        message: 'ok',
        identity_public: 'f76fd3000b:0:542c89e34a369c2281ed940d05beeffdbaa66930f17b875e9172e43d0ba30b6a39708507f4d64e66cde4a1040d2a995d01209d685ca6c4adb4a5c880af1e9715',
        identity_path: '/var/lib/zerotier-one/identity.public',
      },
    })

    await waitFor(() => {
      expect(getByText('f76fd3000b')).toBeDefined()
    })
    expect(getByText('读取身份')).toBeDefined()
    expect(queryByText('读取中...')).toBeNull()
  })
})
