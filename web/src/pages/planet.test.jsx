import { describe, expect, test } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import Planet from './Planet'

describe('Planet page', () => {
  test('keeps a single experimental warning and removes old tool-oriented actions', () => {
    const html = renderToStaticMarkup(<Planet />)

    expect((html.match(/该能力当前保持实验性状态/g) || []).length).toBe(1)
    expect(html).not.toContain('复制 C 头文件')
    expect(html).toContain('身份加载')
    expect(html).toContain('Planet 配置')
    expect(html).toContain('高级模式')
    expect(html).toContain('signing keys')
    expect(html).toContain('Root Nodes')
    expect(html).toContain('读取身份')
    expect(html).toContain('placeholder="留空时使用控制器令牌文件所在目录"')
    expect(html).not.toContain('实际读取路径：')
    expect(html).not.toContain('已读取真实 identity.public，可继续填写默认模式配置')
    expect(html).not.toContain('成功读取后会显示当前 root identity')
    expect(html).not.toContain('读取中...')
  })
})
