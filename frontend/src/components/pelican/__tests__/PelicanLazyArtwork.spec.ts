import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import PelicanLazyArtwork from '../PelicanLazyArtwork.vue'
import { createPelicanArtwork, providePelicanArtwork } from '../usePelicanArtwork'
import type { PelicanResult } from '@/api/pelicanTests'

const rectangle = (top: number, bottom: number) => ({ top, bottom, left: 0, right: 700, width: 700, height: bottom - top, x: 0, y: top, toJSON: () => ({}) })
describe('lazy artwork real geometry', () => {
  let top = 1200
  let hidden = false
  let frameId = 0
  const frames = new Map<number, FrameRequestCallback>()
  async function flushFrame() {
    await flushPromises()
    const callbacks = [...frames.values()]
    frames.clear()
    callbacks.forEach(callback => callback(16))
    await flushPromises()
  }
  beforeEach(() => {
    top = 1200; hidden = false; frameId = 0; frames.clear()
    vi.stubGlobal('requestAnimationFrame', vi.fn((callback: FrameRequestCallback) => { frames.set(++frameId, callback); return frameId }))
    vi.stubGlobal('cancelAnimationFrame', vi.fn((id: number) => frames.delete(id)))
    vi.stubGlobal('IntersectionObserver', undefined)
    vi.spyOn(document, 'hidden', 'get').mockImplementation(() => hidden)
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function () {
      return this.classList.contains('lazy-artwork') ? rectangle(top, top + 220) : rectangle(100, 600)
    })
  })
  afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals() })
  function setup(fetcher = vi.fn(async (id: number) => ({ id, status: 'success', html: '<p>ok</p>' }) as PelicanResult)) {
    const store = createPelicanArtwork(undefined, fetcher)
    const root = document.createElement('div')
    const wrapper = mount(defineComponent({
      setup() { providePelicanArtwork(store); return () => h(PelicanLazyArtwork, { resultId: 1, root }) }
    }), { global: { stubs: { PelicanSandboxFrame: { props: ['html'], template: '<iframe :srcdoc="html" />' } } } })
    return { wrapper, store, fetcher, root }
  }
  it('fallback prefetches within 240px of actual scroll root but mounts iframe only on actual intersection', async () => {
    const { wrapper, store, fetcher, root } = setup()
    await flushPromises()
    expect(fetcher).not.toHaveBeenCalled()
    top = 700; root.dispatchEvent(new Event('scroll')); await flushFrame()
    expect(fetcher).toHaveBeenCalledTimes(1)
    expect(wrapper.find('iframe').exists()).toBe(false)
    top = 599; root.dispatchEvent(new Event('scroll')); await flushFrame()
    expect(wrapper.find('iframe').exists()).toBe(true)
    top = 1200; root.dispatchEvent(new Event('scroll')); await flushFrame()
    expect(wrapper.find('iframe').exists()).toBe(false)
    top = 150; root.dispatchEvent(new Event('scroll')); await flushFrame()
    expect(wrapper.find('iframe').exists()).toBe(true)
    expect(fetcher).toHaveBeenCalledTimes(1)
    wrapper.unmount(); store.dispose()
    top = 1200; root.dispatchEvent(new Event('scroll')); window.dispatchEvent(new Event('resize'))
    expect(fetcher).toHaveBeenCalledTimes(1)
  })
  it('suspends hidden documents and preserves a local failure until explicit retry', async () => {
    top = 150
    const fetcher = vi.fn().mockRejectedValueOnce(new Error('offline')).mockResolvedValue({ id: 1, status: 'success', html: 'ok' })
    const { wrapper, store, root } = setup(fetcher)
    await flushPromises()
    expect(wrapper.text()).toContain('作品加载失败')
    root.dispatchEvent(new Event('scroll')); await flushPromises()
    expect(fetcher).toHaveBeenCalledTimes(1)
    await wrapper.get('button[aria-label="重新加载作品"]').trigger('click'); await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(true)
    hidden = true; document.dispatchEvent(new Event('visibilitychange')); await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(false)
    hidden = false; document.dispatchEvent(new Event('visibilitychange')); await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(true)
    expect(fetcher).toHaveBeenCalledTimes(2)
    wrapper.unmount(); store.dispose()
  })
  it('rejects late detail and releases visible HTML when the page context loses authorization', async () => {
    top = 150
    const { wrapper, store, fetcher, root } = setup()
    await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(true)
    store.invalidate(); await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(false)
    root.dispatchEvent(new Event('scroll')); await flushPromises()
    expect(fetcher).toHaveBeenCalledTimes(1)
    wrapper.unmount(); store.dispose()
  })

  it('coalesces root, capture scroll, resize and both IO callbacks into one geometry pass per frame', async () => {
    const observers: IntersectionObserverCallback[] = []
    const disconnect = vi.fn()
    vi.stubGlobal('IntersectionObserver', class {
      constructor(callback: IntersectionObserverCallback) { observers.push(callback) }
      observe() {}
      disconnect = disconnect
    })
    const resizeCallbacks: ResizeObserverCallback[] = []
    vi.stubGlobal('ResizeObserver', class {
      constructor(callback: ResizeObserverCallback) { resizeCallbacks.push(callback) }
      observe() {}
      disconnect = disconnect
    })
    const { wrapper, store, root, fetcher } = setup()
    await flushFrame()
    const geometry = vi.mocked(HTMLElement.prototype.getBoundingClientRect)
    geometry.mockClear()
    top = 700
    for (let i = 0; i < 12; i++) {
      root.dispatchEvent(new Event('scroll'))
      window.dispatchEvent(new Event('scroll'))
      window.dispatchEvent(new Event('resize'))
      observers.forEach(callback => callback([], {} as IntersectionObserver))
      resizeCallbacks.forEach(callback => callback([], {} as ResizeObserver))
    }
    expect(frames.size).toBe(1)
    expect(geometry).not.toHaveBeenCalled()
    await flushFrame()
    expect(geometry.mock.instances.filter(element => element === wrapper.get('.lazy-artwork').element)).toHaveLength(1)
    expect(fetcher).toHaveBeenCalledTimes(1)
    expect(wrapper.find('iframe').exists()).toBe(false)
    wrapper.unmount(); store.dispose()
    expect(disconnect).toHaveBeenCalled()
  })

  it('cancels a queued geometry callback on unmount, including a late observer delivery', async () => {
    const callbacks: IntersectionObserverCallback[] = []
    vi.stubGlobal('IntersectionObserver', class {
      constructor(callback: IntersectionObserverCallback) { callbacks.push(callback) }
      observe() {}
      disconnect() {}
    })
    const { wrapper, store, root, fetcher } = setup()
    await flushFrame()
    top = 150
    root.dispatchEvent(new Event('scroll'))
    const queued = [...frames.values()][0]!
    expect(frames.size).toBe(1)
    wrapper.unmount()
    expect(cancelAnimationFrame).toHaveBeenCalled()
    expect(frames.size).toBe(0)
    const geometry = vi.mocked(HTMLElement.prototype.getBoundingClientRect)
    geometry.mockClear()
    queued(16)
    callbacks.forEach(callback => callback([], {} as IntersectionObserver))
    await flushFrame()
    expect(geometry).not.toHaveBeenCalled()
    expect(fetcher).not.toHaveBeenCalled()
    store.dispose()
  })

  it.each(['hidden', 'authorization'])('clears mounted artwork immediately before a queued frame (%s)', async reason => {
    top = 150
    const { wrapper, store, root } = setup()
    await flushFrame()
    expect(wrapper.find('iframe').exists()).toBe(true)
    root.dispatchEvent(new Event('scroll'))
    const queued = [...frames.values()][0]!
    if (reason === 'hidden') {
      hidden = true
      document.dispatchEvent(new Event('visibilitychange'))
    } else store.invalidate()
    await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(false)
    expect(frames.size).toBe(0)
    queued(16)
    await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(false)
    wrapper.unmount(); store.dispose()
  })
})
