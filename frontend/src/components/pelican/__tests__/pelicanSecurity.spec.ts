import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
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

  it('places CSP before artwork and permits only the trusted opaque preview bridge', () => {
    const srcdoc = toPelicanSrcdoc('<svg><circle /></svg>')
    expect(srcdoc.indexOf('Content-Security-Policy')).toBeLessThan(srcdoc.indexOf('<svg>'))
    expect(srcdoc).toContain("default-src 'none'")
    expect(srcdoc).toContain('img-src data:')

    const wrapper = mount(PelicanSandboxFrame, { props: { html: '<svg><circle /></svg>', replayKey: 1 } })
    const frame = wrapper.get('iframe')
    expect(frame.attributes('sandbox')).toBe('allow-scripts')
    expect(frame.attributes('tabindex')).toBe('-1')
    expect(frame.attributes('srcdoc')).toContain("default-src 'none'")
    expect(frame.attributes('class')).toContain('pelican-frame')
    expect(frame.attributes('srcdoc')).toMatch(/script-src 'nonce-[a-f0-9]{48}'/)
    wrapper.unmount()
  })

  it('keeps the enlarged artwork frame usable without relaxing the sandbox', () => {
    const wrapper = mount(PelicanSandboxFrame, { props: { html: '<svg><circle /></svg>', replayKey: 1, interactive: true } })
    const frame = wrapper.get('iframe')

    expect(frame.attributes('sandbox')).toBe('')
    expect(frame.attributes('tabindex')).toBeUndefined()
    expect(frame.attributes('aria-hidden')).toBeUndefined()
    expect(frame.attributes('srcdoc')).not.toContain('<script')
    wrapper.unmount()
  })

  it('accepts only current-frame numeric measurements and clamps long artwork', async () => {
    const wrapper = mount(PelicanSandboxFrame, { attachTo: document.body, props: { html: '<p>short</p>', replayKey: 0 } })
    const iframe = wrapper.get('iframe').element as HTMLIFrameElement
    const token = iframe.srcdoc.match(/data-token="([a-f0-9]+)"/)![1]
    const send = async (data: Record<string, unknown>, source: Window | null = iframe.contentWindow) => {
      window.dispatchEvent(new MessageEvent('message', { source, data: { type: 'pelican-preview-height', token, sequence: 1, height: 70, ...data } }))
      await wrapper.vm.$nextTick()
    }
    expect(iframe.style.height).toBe('220px')
    for (const data of [{ token: 'foreign' }, { height: NaN }, { height: Infinity }, { height: '70' }, { height: -1 }, { height: 100001 }, { sequence: 0 }, { type: 'other' }]) await send(data)
    await send({}, window)
    expect(iframe.style.height).toBe('220px')
    await send({})
    expect(iframe.style.height).toBe('70px')
    await send({ height: 400 }) // Replay of the same sequence is ignored.
    expect(iframe.style.height).toBe('70px')
    await send({ height: 1200, sequence: 2 })
    expect(iframe.style.height).toBe('580px')
    await wrapper.setProps({ replayKey: 1 })
    const nextFrame = wrapper.get('iframe').element as HTMLIFrameElement
    expect(nextFrame).not.toBe(iframe)
    expect(nextFrame.srcdoc).not.toContain(token)
    await send({ height: 80, sequence: 3 }, nextFrame.contentWindow)
    expect(nextFrame.style.height).toBe('220px')
    const remove = vi.spyOn(window, 'removeEventListener')
    wrapper.unmount()
    expect(remove).toHaveBeenCalledWith('message', expect.any(Function))
    remove.mockRestore()
  })

  it('reuses the outer document CSP nonce when embedding the trusted bridge', () => {
    const script = document.createElement('script')
    script.setAttribute('nonce', 'outer-server-nonce-1234567890')
    document.head.appendChild(script)
    const wrapper = mount(PelicanSandboxFrame, { props: { html: '<p>short</p>', replayKey: 0 } })
    expect(wrapper.get('iframe').attributes('srcdoc')).toContain("script-src 'nonce-outer-server-nonce-1234567890'")
    wrapper.unmount()
    script.remove()
  })
})
