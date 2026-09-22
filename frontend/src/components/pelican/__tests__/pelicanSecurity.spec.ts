import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import PelicanSandboxFrame from '../PelicanSandboxFrame.vue'
import { sanitizePelicanHtml, toPelicanSrcdoc } from '../sanitizePelicanHtml'

describe('Pelican untrusted artwork boundary', () => {
  it('removes external navigation and unsafe SVG mutations while retaining motion', () => {
    const clean = sanitizePelicanHtml([
      '<a href="https://outside.example">link</a>',
      '<img src="https://outside.example/p.png" srcset="https://outside.example/p2.png 2x">',
      '<svg xmlns:xlink="http://www.w3.org/1999/xlink"><image xlink:href="https://outside.example/i.svg" />',
      '<animate attributeName="href" values="https://outside.example" />',
      '<animateMotion /><animateTransform /><set /><use href="#shape" /><foreignObject><div>x</div></foreignObject></svg>',
    ].join(''))

    expect(clean).not.toMatch(/outside\.example|<a[\s>]|<use[\s>]|<foreignobject/i)
    expect(clean).toContain('<animateMotion')
    expect(clean).not.toMatch(/attributeName="href"|<set[\s>]/i)
    expect(clean).not.toMatch(/(?:href|src|xlink:href|srcset|ping)=/i)
  })

  it('drops unsafe CSS while retaining ordinary inline CSS and a data image', () => {
    const clean = sanitizePelicanHtml([
      '<style>@import "https://outside.example/a.css";</style>',
      '<div style="background:url(https://outside.example/p.png)">bad</div>',
      '<div style="color: teal; animation: bob 1s linear infinite">safe</div>',
      '<style>@keyframes bob { to { transform: translateX(1px) } }</style>',
      '<img src="data:image/png;base64,AA==">',
    ].join(''))

    expect(clean).not.toMatch(/@import|url\s*\(|outside\.example/i)
    expect(clean).toContain('color: teal')
    expect(clean).toContain('@keyframes bob')
    expect(clean).toContain('src="data:image/png;base64,AA=="')
  })

  it('places CSP before untrusted markup and renders with an empty sandbox', () => {
    const srcdoc = toPelicanSrcdoc('<svg><circle /></svg>')
    expect(srcdoc.indexOf('Content-Security-Policy')).toBeLessThan(srcdoc.indexOf('<svg>'))
    expect(srcdoc).toContain("default-src 'none'")
    expect(srcdoc).toContain('img-src data:')

    const wrapper = mount(PelicanSandboxFrame, { props: { html: '<svg><circle /></svg>', replayKey: 1 } })
    const frame = wrapper.get('iframe')
    expect(frame.attributes('sandbox')).toBe('')
    expect(frame.attributes('tabindex')).toBe('-1')
    expect(frame.attributes('srcdoc')).toContain("default-src 'none'")
    expect(frame.attributes('class')).toContain('pelican-frame')
  })

  it('keeps the enlarged artwork frame usable without relaxing the sandbox', () => {
    const wrapper = mount(PelicanSandboxFrame, { props: { html: '<svg><circle /></svg>', replayKey: 1, interactive: true } })
    const frame = wrapper.get('iframe')

    expect(frame.attributes('sandbox')).toBe('')
    expect(frame.attributes('tabindex')).toBeUndefined()
    expect(frame.attributes('aria-hidden')).toBeUndefined()
  })
})
