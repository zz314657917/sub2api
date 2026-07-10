import { describe, expect, it } from 'vitest'
import { renderTutorialMarkdown, type TutorialRenderOptions } from '../tutorialMarkdown'

function render(markdown: string, options?: TutorialRenderOptions) {
  const result = renderTutorialMarkdown(markdown, options)
  const root = document.createElement('div')
  root.innerHTML = result.html
  return { ...result, root }
}

function decodeCopyCode(button: HTMLButtonElement): string {
  return decodeURIComponent(button.dataset.copyCode || '')
}

describe('renderTutorialMarkdown', () => {
  it('renders fenced code with a language label and one usable copy button', () => {
    const code = 'npm.cmd install\nnpm.cmd run dev'
    const { root } = render(`\`\`\`powershell\n${code}\n\`\`\``)
    const block = root.querySelector<HTMLElement>('.tutorial-command-block')
    const button = block?.querySelector<HTMLButtonElement>('button[data-copy-code]')

    expect(root.querySelectorAll('.tutorial-command-block')).toHaveLength(1)
    expect(block?.querySelector('.command-block-language')?.textContent).toBe('powershell')
    expect(block?.querySelector('pre > code.language-powershell')?.textContent).toBe(code)
    expect(block?.querySelectorAll('button[data-copy-code]')).toHaveLength(1)
    expect(button?.type).toBe('button')
    expect(decodeCopyCode(button!)).toBe(code)
  })

  it('uses the same single-copy markup for command shortcodes', () => {
    const code = 'curl https://api.example.com/v1/models?group=default&enabled=true'
    const { root } = render(`[[command title="检查模型" lang="bash"]]\n${code}\n[[/command]]`)
    const block = root.querySelector<HTMLElement>('.tutorial-command-block')
    const button = block?.querySelector<HTMLButtonElement>('button[data-copy-code]')

    expect(root.querySelectorAll('.tutorial-command-block')).toHaveLength(1)
    expect(block?.querySelectorAll('.tutorial-command-block')).toHaveLength(0)
    expect(block?.querySelector('.command-block-title')?.textContent).toBe('检查模型')
    expect(block?.querySelector('.command-block-language')?.textContent).toBe('bash')
    expect(block?.querySelector('pre > code.language-bash')?.textContent).toBe(code)
    expect(block?.querySelectorAll('button[data-copy-code]')).toHaveLength(1)
    expect(decodeCopyCode(button!)).toBe(code)
  })

  it('adds a text language fallback and decorates fenced code inside callouts', () => {
    const { root } = render(`\`\`\`\nplain text\n\`\`\`\n\n[[callout type="tip" title="提示"]]\n\`\`\`json\n{"ok":true}\n\`\`\`\n[[/callout]]`)
    const blocks = Array.from(root.querySelectorAll<HTMLElement>('.tutorial-command-block'))

    expect(blocks).toHaveLength(2)
    expect(blocks.map((block) => block.querySelector('.command-block-language')?.textContent)).toEqual(['text', 'json'])
    expect(blocks.every((block) => block.querySelectorAll('button[data-copy-code]').length === 1)).toBe(true)
    expect(root.querySelector('.tutorial-callout .tutorial-command-block')).not.toBeNull()
  })

  it('preserves headings, TOC, shortcodes, and DOMPurify cleaning', () => {
    const { root, toc } = render(`# 教程\n\n## 快速开始\n\n[[screenshot src="/guide.png" alt="配置截图" caption="填写配置"]]\n\n[[callout type="warning" title="注意"]]\n<strong onclick="alert(1)">检查配置</strong>\n[[/callout]]\n\n[[link-button href="/models" label="模型广场"]]\n\n<script>alert('xss')</script>`, {
      skipTitle: '教程'
    })

    expect(root.querySelector('h1')).toBeNull()
    expect(root.querySelector('h2')?.id).toBe('快速开始-0')
    expect(toc).toEqual([{ id: '快速开始-0', text: '快速开始', level: 2 }])
    expect(root.querySelector<HTMLImageElement>('.tutorial-screenshot-card img')?.getAttribute('src')).toBe('/guide.png')
    expect(root.querySelector('.tutorial-callout-warning strong')?.textContent).toBe('注意')
    expect(root.querySelector<HTMLAnchorElement>('.guide-action-link')?.getAttribute('href')).toBe('/models')
    expect(root.querySelector('script')).toBeNull()
    expect(root.querySelector('[onclick]')).toBeNull()
  })
})
