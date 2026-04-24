import { describe, expect, test } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import Planet from './Planet'

describe('Planet page', () => {
  test('keeps a single experimental warning and removes old tool-oriented actions', () => {
    const html = renderToStaticMarkup(<Planet />)

    expect((html.match(/该能力当前保持实验性状态/g) || []).length).toBe(1)
    expect(html).not.toContain('复制 C 头文件')
    expect(html).not.toContain('Signing keys')
    expect(html).toContain('身份加载')
    expect(html).toContain('Planet 配置')
    expect(html).toContain('读取身份')
    expect(html).not.toContain('读取中...')
  })
})
