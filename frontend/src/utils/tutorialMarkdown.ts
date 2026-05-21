import { marked } from 'marked'
import DOMPurify from 'dompurify'

export interface TutorialTocItem {
  id: string
  text: string
  level: number
}

export interface TutorialRenderResult {
  html: string
  toc: TutorialTocItem[]
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function parseAttributes(raw: string): Record<string, string> {
  const attrs: Record<string, string> = {}
  const attrRe = /([a-zA-Z0-9_-]+)="([^"]*)"/g
  let match: RegExpExecArray | null
  while ((match = attrRe.exec(raw)) !== null) {
    attrs[match[1]] = match[2]
  }
  return attrs
}

function renderShortcodes(markdown: string): string {
  let next = markdown

  next = next.replace(
    /\[\[command([^\]]*)\]\]([\s\S]*?)\[\[\/command\]\]/g,
    (_match, rawAttrs: string, body: string) => {
      const attrs = parseAttributes(rawAttrs)
      const title = escapeHtml(attrs.title || '命令')
      const lang = escapeHtml(attrs.lang || '')
      const code = escapeHtml(body.replace(/^\n|\n$/g, ''))
      const codeClass = lang ? ` class="language-${lang}"` : ''
      return `<div class="tutorial-command-block command-block"><div class="command-block-header"><span>${title}</span><button type="button" class="copy-command-button" data-copy-code="${encodeURIComponent(body.replace(/^\n|\n$/g, ''))}">复制</button></div><pre><code${codeClass}>${code}</code></pre></div>`
    }
  )

  next = next.replace(
    /\[\[callout([^\]]*)\]\]([\s\S]*?)\[\[\/callout\]\]/g,
    (_match, rawAttrs: string, body: string) => {
      const attrs = parseAttributes(rawAttrs)
      const type = escapeHtml(attrs.type || 'tip')
      const title = escapeHtml(attrs.title || '提示')
      const content = marked.parse(body.trim()) as string
      return `<div class="tutorial-callout tutorial-callout-${type}"><strong>${title}</strong>${content}</div>`
    }
  )

  next = next.replace(
    /\[\[screenshot([^\]]*)\]\]/g,
    (_match, rawAttrs: string) => {
      const attrs = parseAttributes(rawAttrs)
      const src = escapeHtml(attrs.src || '')
      const alt = escapeHtml(attrs.alt || attrs.caption || '教程截图')
      const caption = escapeHtml(attrs.caption || alt)
      if (!src) return ''
      return `<figure class="tutorial-screenshot-card"><img src="${src}" alt="${alt}" loading="lazy" /><figcaption>${caption}</figcaption></figure>`
    }
  )

  next = next.replace(
    /\[\[link-button([^\]]*)\]\]/g,
    (_match, rawAttrs: string) => {
      const attrs = parseAttributes(rawAttrs)
      const href = escapeHtml(attrs.href || '#')
      const label = escapeHtml(attrs.label || attrs.href || '打开链接')
      return `<a class="guide-action-link" href="${href}">${label}</a>`
    }
  )

  return next
}

function generateHeadingId(text: string, index: number): string {
  const base = text
    .toLowerCase()
    .replace(/[^\w一-鿿]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return base ? `${base}-${index}` : `heading-${index}`
}

export function renderTutorialMarkdown(markdown: string): TutorialRenderResult {
  const rendered = marked.parse(renderShortcodes(markdown || '')) as string
  const sanitized = DOMPurify.sanitize(rendered, {
    ADD_ATTR: ['target', 'rel', 'loading', 'data-copy-code']
  })

  const toc: TutorialTocItem[] = []
  let headingIndex = 0
  const html = sanitized.replace(
    /<(h[1-3])[^>]*>(.*?)<\/h[1-3]>/gi,
    (_match, tag: string, content: string) => {
      const level = Number(tag[1])
      const text = content.replace(/<[^>]+>/g, '').trim()
      const id = generateHeadingId(text, headingIndex++)
      toc.push({ id, text, level })
      return `<${tag} id="${id}">${content}</${tag}>`
    }
  )

  return { html, toc }
}
