import { afterEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import ProxySelector from '../ProxySelector.vue'

const testProxy = vi.hoisted(() => vi.fn())
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/api/admin', () => ({ adminAPI: { proxies: { testProxy } } }))
enableAutoUnmount(afterEach)

describe('proxy test deduplication', () => {
  it('does not submit the same proxy twice when individual and batch tests overlap', async () => {
    let resolve!: (value: { success: boolean }) => void
    testProxy.mockReturnValue(new Promise(done => { resolve = done }))
    const wrapper = mount(ProxySelector, { props: { modelValue: null, proxies: [{ id: 1, name: 'p', protocol: 'http', host: 'localhost', port: 80 }] }, global: { stubs: { Icon: true } } })
    await wrapper.get('.select-trigger').trigger('click')
    await wrapper.get('.test-btn').trigger('click')
    await wrapper.get('.batch-test-btn').trigger('click')
    expect(testProxy).toHaveBeenCalledTimes(1)
    resolve({ success: true })
    await flushPromises()
  })
})
