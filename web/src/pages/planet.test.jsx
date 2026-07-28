import { describe, expect, test } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import Planet from './Planet'

const planetPageHTML = renderToStaticMarkup(<Planet />)

describe('Planet page', () => {
  test('renders the active generation workflow and directory guidance', () => {
    expect((planetPageHTML.match(/该能力当前保持实验性状态/g) || []).length).toBe(1)
    expect(planetPageHTML).toContain('身份加载')
    expect(planetPageHTML).toContain('Planet 配置')
    expect(planetPageHTML).toContain('高级模式')
    expect(planetPageHTML).toContain('signing keys')
    expect(planetPageHTML).toContain('Root Nodes')
    expect(planetPageHTML).toContain('读取身份')
    expect(planetPageHTML).toContain('placeholder="留空时使用控制器令牌文件所在目录"')
  })

  test('does not render obsolete actions or idle identity feedback', () => {
    expect(planetPageHTML).not.toContain('复制 C 头文件')
    expect(planetPageHTML).not.toContain('实际读取路径：')
    expect(planetPageHTML).not.toContain('已读取真实 identity.public，可继续填写默认模式配置')
    expect(planetPageHTML).not.toContain('成功读取后会显示当前 root identity')
    expect(planetPageHTML).not.toContain('读取中...')
  })
})
