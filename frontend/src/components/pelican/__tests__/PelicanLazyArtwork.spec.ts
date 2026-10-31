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
  beforeEach(() => {
    top = 1200; hidden = false
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
    top = 700; root.dispatchEvent(new Event('scroll')); await flushPromises()
    expect(fetcher).toHaveBeenCalledTimes(1)
    expect(wrapper.find('iframe').exists()).toBe(false)
    top = 599; root.dispatchEvent(new Event('scroll')); await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(true)
    top = 1200; root.dispatchEvent(new Event('scroll')); await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(false)
    top = 150; root.dispatchEvent(new Event('scroll')); await flushPromises()
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
})
