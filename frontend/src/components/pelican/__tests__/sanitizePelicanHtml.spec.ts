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
  it('keeps safe declarative SVG animation while stripping external targets', () => {
    const html = sanitizePelicanHtml('<svg><circle id="dot"><animate attributeName="r" values="4;8;4" dur="2s" repeatCount="indefinite" /><animateTransform attributeName="transform" type="rotate" from="0" to="360" dur="4s" repeatCount="indefinite" /><set attributeName="opacity" to="0.5" /></circle><animate attributeName="href" values="https://outside.example/x" /></svg>')
    expect(html).toContain('<animate')
    expect(html).toContain('<animateTransform')
    expect(html).toContain('<set')
    expect(html).toContain('values="4;8;4"')
    expect(html).toContain('from="0"')
    expect(html).toContain('to="360"')
    expect(html).toContain('repeatCount="indefinite"')
    expect(html).not.toContain('outside.example')
    expect(html).not.toMatch(/(?:href|xlink:href)=/i)
  })
  it('preserves motion paths but rejects resource and event attribute mutations', () => {
    const html = sanitizePelicanHtml('<svg><circle><animateMotion path="M0 0 L10 10" dur="2s" repeatCount="indefinite" /></circle><set attributeName="onclick" to="alert(1)" /><set attributeName="style" to="background:red" /><animate attributeName="fill" values="red;url(https://outside.example/x)" /><set attributeName="href" to="https://outside.example" /></svg>')
    expect(html).toContain('<animateMotion')
    expect(html).toContain('path="M0 0 L10 10"')
    expect(html).not.toMatch(/onclick|alert|background|outside\.example|<set/)
  })
  it('adds the restrictive CSP before sanitized markup', () => {
    const srcdoc = toPelicanSrcdoc('<img src="data:image/svg+xml;base64,AA==">')
    expect(srcdoc).toContain("default-src 'none'")
    expect(srcdoc).toContain('img-src data:')
    expect(srcdoc).toContain('data:image')
  })
  it('keeps the nonce script constant and removes executable artwork and supplied nonces', () => {
    const identity = { nonce: 'a'.repeat(48), token: 'b'.repeat(48) }
    const source = '<script nonce="attacker">parent.postMessage("attack","*")</script><svg onload="alert(1)"><set attributeName="nonce" to="attacker" /></svg><div nonce="attacker" onclick="alert(2)">art</div>'
    const result = toPelicanSrcdoc(source, true, identity)
    expect(result.match(/<script\b/g)).toHaveLength(1)
    expect(result).not.toMatch(/attacker|onclick|onload|alert\(/)
    expect(result).toContain(`script-src 'nonce-${identity.nonce}'`)
    expect(result.indexOf('Content-Security-Policy')).toBeLessThan(result.indexOf('<svg'))
    expect(result).toContain('height:auto!important;min-height:0!important')
    expect(toPelicanSrcdoc(source, false, identity)).not.toContain('<script')
    expect(toPelicanSrcdoc(source, true, { nonce: '\"><script>evil', token: identity.token })).not.toContain('<script')
  })
})
