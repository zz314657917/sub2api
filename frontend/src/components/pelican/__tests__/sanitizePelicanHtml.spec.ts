import { describe, expect, it } from 'vitest'
import { sanitizePelicanHtml, toPelicanSrcdoc } from '../sanitizePelicanHtml'

describe('pelican HTML isolation', () => {
  it('preserves local SVG gradients without permitting remote CSS resources', () => {
    const html = sanitizePelicanHtml('<style>circle{fill:url(#sea);animation:spin 2s linear infinite}@keyframes spin{to{transform:rotate(360deg)}}</style><svg><defs><linearGradient id="sea" /></defs><circle /></svg>')
    expect(html).toContain('url(#sea)')
    expect(html).toContain('@keyframes spin')
    expect(sanitizePelicanHtml('<style>circle{fill:url(https://example.test/x)}</style>')).not.toContain('example.test')
  })
  it('removes executable and external content while retaining inline style and SVG', () => {
    const html = sanitizePelicanHtml('<style>svg{color:red}</style><svg><circle /></svg><script>alert(1)</script><img src="https://example.test/a.png"><a href="https://example.test">x</a>')
    expect(html).toContain('<style>')
    expect(html).toContain('<svg>')
    expect(html).not.toMatch(/script|example\.test|href=/)
  })
  it('adds the restrictive CSP before sanitized markup', () => {
    const srcdoc = toPelicanSrcdoc('<img src="data:image/svg+xml;base64,AA==">')
    expect(srcdoc).toContain("default-src 'none'")
    expect(srcdoc).toContain('img-src data:')
    expect(srcdoc).toContain('data:image')
  })
})
