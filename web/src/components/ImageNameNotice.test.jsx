import { afterEach, beforeEach, describe, expect, test } from 'bun:test'
import { render } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useTestDOM } from '../test/dom'
import { LanguageProvider } from '../i18n'
import ImageNameNotice from './ImageNameNotice'

useTestDOM()

// MUI renders a dialog into a portal on document.body, so assertions go through
// the render's baseElement, and the language is pinned: with the default
// "system" preference the strings would depend on the test browser's locale.
beforeEach(() => {
  globalThis.localStorage.setItem('tempest-zt.language', 'zh-CN')
})

afterEach(() => {
  globalThis.localStorage.clear()
})

/**
 * A storage the test owns, so a dismissal cannot leak into another case through
 * the window's localStorage.
 *
 * @param {Record<string, string>} [initial]
 * @returns {import('./ImageNameNotice').NoticeStorage}
 */
function memoryStorage(initial = {}) {
  const entries = new Map(Object.entries(initial))
  return {
    getItem: (key) => entries.get(key) ?? null,
    setItem: (key, value) => {
      entries.set(key, value)
    },
  }
}

/**
 * @param {import('./ImageNameNotice').NoticeStorage} storage
 */
function renderNotice(storage) {
  return render(
    <LanguageProvider>
      <ImageNameNotice storage={storage} />
    </LanguageProvider>,
  )
}

describe('ImageNameNotice', () => {
  test('names both the new and the old image so the operator knows what to change', () => {
    const { baseElement } = renderNotice(memoryStorage())
    const html = baseElement.innerHTML

    expect(html).toContain('ghcr.io/gt-610/tempest-zt')
    expect(html).toContain('ghcr.io/gt-610/tairitsu')
  })

  test('closes for this visit without recording a dismissal', async () => {
    const storage = memoryStorage()
    const user = userEvent.setup({ document })
    const { baseElement, getByText } = renderNotice(storage)

    await user.click(getByText('我知道了'))

    expect(baseElement.innerHTML).not.toContain('容器镜像名称已更改')
    expect(storage.getItem('tempest-zt.image-name-notice.dismissed')).toBeNull()
  })

  test('stays hidden after the dismissal is recorded', async () => {
    const storage = memoryStorage()
    const user = userEvent.setup({ document })
    const first = renderNotice(storage)

    await user.click(first.getByText('不再提示'))

    expect(storage.getItem('tempest-zt.image-name-notice.dismissed')).toBe('true')

    first.unmount()
    const second = renderNotice(storage)
    expect(second.baseElement.innerHTML).not.toContain('容器镜像名称已更改')
  })

  test('still shows the notice when storage is unavailable', () => {
    // Private mode and blocked cookies throw on access; the notice must survive
    // that rather than breaking the page it is mounted in.
    /** @type {import('./ImageNameNotice').NoticeStorage} */
    const throwing = {
      getItem: () => {
        throw new Error('denied')
      },
      setItem: () => {
        throw new Error('denied')
      },
    }

    const { baseElement } = renderNotice(throwing)
    expect(baseElement.innerHTML).toContain('容器镜像名称已更改')
  })
})
