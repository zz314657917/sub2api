import { marked, Renderer } from 'marked'
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

export interface TutorialRenderOptions {
  skipTitle?: string
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function normalizeCodeLanguage(value: string): string {
  return value.trim().split(/\s+/, 1)[0] || 'text'
}

function renderCodeBlock(code: string, rawLanguage = '', title = ''): string {
  const language = normalizeCodeLanguage(rawLanguage)
  const escapedLanguage = escapeHtml(language)
  const escapedTitle = escapeHtml(title.trim())
  const titleMarkup = escapedTitle
    ? `<span class="command-block-title">${escapedTitle}</span><span class="command-block-label-separator" aria-hidden="true"> · </span>`
    : ''

  return `<div class="tutorial-command-block command-block"><div class="command-block-header"><span class="command-block-label">${titleMarkup}<span class="command-block-language">${escapedLanguage}</span></span><button type="button" class="copy-command-button" data-copy-code="${encodeURIComponent(code)}">复制</button></div><pre><code class="language-${escapedLanguage}">${escapeHtml(code)}</code></pre></div>`
}

function parseMarkdown(markdown: string): string {
  const renderer = new Renderer()
  renderer.code = ({ text, lang }) => renderCodeBlock(text, lang || '')
  return marked.parse(markdown, { renderer }) as string
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
  const blocks: string[] = []
  const preserveBlock = (html: string): string => {
    const token = `\n\n@@TUTORIAL_SHORTCODE_${blocks.length}@@\n\n`
    blocks.push(html)
    return token
  }

  next = next.replace(
    /\[\[command([^\]]*)\]\]([\s\S]*?)\[\[\/command\]\]/g,
    (_match, rawAttrs: string, body: string) => {
      const attrs = parseAttributes(rawAttrs)
      const code = body.replace(/^\n|\n$/g, '')
      return preserveBlock(renderCodeBlock(code, attrs.lang || '', attrs.title || '命令'))
    }
  )

  next = next.replace(
    /\[\[callout([^\]]*)\]\]([\s\S]*?)\[\[\/callout\]\]/g,
    (_match, rawAttrs: string, body: string) => {
      const attrs = parseAttributes(rawAttrs)
      const type = escapeHtml(attrs.type || 'tip')
      const title = escapeHtml(attrs.title || '提示')
      const content = parseMarkdown(body.trim())
      return preserveBlock(`<div class="tutorial-callout tutorial-callout-${type}"><strong>${title}</strong>${content}</div>`)
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
      return preserveBlock(
        `<figure class="tutorial-screenshot-card"><img src="${src}" alt="${alt}" loading="lazy" /><figcaption>${caption}</figcaption></figure>`
      )
    }
  )

  next = next.replace(
    /\[\[link-button([^\]]*)\]\]/g,
    (_match, rawAttrs: string) => {
      const attrs = parseAttributes(rawAttrs)
      const href = escapeHtml(attrs.href || '#')
      const label = escapeHtml(attrs.label || attrs.href || '打开链接')
      return preserveBlock(`<a class="guide-action-link" href="${href}">${label}</a>`)
    }
  )

  return next.replace(/@@TUTORIAL_SHORTCODE_(\d+)@@/g, (_match, index: string) => blocks[Number(index)] || '')
}

function generateHeadingId(text: string, index: number): string {
  const base = text
    .toLowerCase()
    .replace(/[^\w一-鿿]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return base ? `${base}-${index}` : `heading-${index}`
}

export function renderTutorialMarkdown(markdown: string, options: TutorialRenderOptions = {}): TutorialRenderResult {
  const rendered = parseMarkdown(renderShortcodes(markdown || ''))
  const sanitized = DOMPurify.sanitize(rendered, {
    ADD_ATTR: ['target', 'rel', 'loading', 'data-copy-code']
  })

  const toc: TutorialTocItem[] = []
  let headingIndex = 0
  let skippedPageTitle = false
  const skipTitle = (options.skipTitle || '').trim()
  const html = sanitized.replace(
    /<(h[1-3])[^>]*>(.*?)<\/h[1-3]>/gi,
    (_match, tag: string, content: string) => {
      const level = Number(tag[1])
      const text = content.replace(/<[^>]+>/g, '').trim()
      if (!skippedPageTitle && skipTitle && level <= 2 && text === skipTitle) {
        skippedPageTitle = true
        return ''
      }
      const id = generateHeadingId(text, headingIndex++)
      toc.push({ id, text, level })
      return `<${tag} id="${id}">${content}</${tag}>`
    }
  )

  return { html, toc }
}
