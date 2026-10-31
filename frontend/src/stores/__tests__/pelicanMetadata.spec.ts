import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { usePelicanMetadataStore } from '@/stores/pelicanMetadata'
import { getPelicanTestMetadata } from '@/api/pelicanTests'

vi.mock('@/api/pelicanTests', () => ({ getPelicanTestMetadata: vi.fn() }))

describe('usePelicanMetadataStore', () => {
  beforeEach(() => { setActivePinia(createPinia()); vi.clearAllMocks() })

  it('coalesces pending forced loads but performs a fresh check on the next navigation', async () => {
    let resolve!: (value: { display_name: string; enabled: boolean }) => void
    vi.mocked(getPelicanTestMetadata).mockImplementationOnce(() => new Promise(inner => { resolve = inner }))
    const store = usePelicanMetadataStore()
    const guard = store.load(true)
    const duplicate = store.load(true)
    const mounted = store.load()
    expect(getPelicanTestMetadata).toHaveBeenCalledTimes(1)
    resolve({ display_name: '当前名称', enabled: true })
    await Promise.all([guard, duplicate, mounted])
    await store.load()
    expect(getPelicanTestMetadata).toHaveBeenCalledTimes(1)
    vi.mocked(getPelicanTestMetadata).mockResolvedValueOnce({ display_name: '导航新检查', enabled: false })
    await store.load(true)
    expect(getPelicanTestMetadata).toHaveBeenCalledTimes(2)
    expect(store.enabled).toBe(false)
  })

  it('does not let an older metadata response replace a saved name', async () => {
    let resolve!: (value: { display_name: string; enabled: boolean }) => void
    vi.mocked(getPelicanTestMetadata).mockImplementationOnce(() => new Promise(inner => { resolve = inner }))
    const store = usePelicanMetadataStore()
    const pending = store.load()
    store.applySavedName('保存后的名称')
    resolve({ display_name: '陈旧响应名称', enabled: true })
    await pending
    expect(store.displayName).toBe('保存后的名称')
    expect(store.loading).toBe(false)
  })

  it('does not let an older enabled response overwrite a saved disable', async () => {
    let resolve!: (value: { display_name: string; enabled: boolean }) => void
    vi.mocked(getPelicanTestMetadata).mockImplementationOnce(() => new Promise(inner => { resolve = inner }))
    const store = usePelicanMetadataStore()
    const pending = store.load()
    store.applySavedSettings('已关闭', false)
    resolve({ display_name: '旧响应', enabled: true })
    await pending
    expect(store.enabled).toBe(false)
    expect(store.displayName).toBe('已关闭')
  })

  it('ignores an old identity response after reset', async () => {
    let resolve!: (value: { display_name: string; enabled: boolean }) => void
    vi.mocked(getPelicanTestMetadata).mockImplementationOnce(() => new Promise(inner => { resolve = inner }))
    const store = usePelicanMetadataStore()
    const pending = store.load()
    store.reset()
    resolve({ display_name: '旧身份', enabled: true })
    await pending
    expect(store.loaded).toBe(false)
    expect(store.enabled).toBe(false)
  })

  it('fails closed after a metadata request failure', async () => {
    vi.mocked(getPelicanTestMetadata).mockRejectedValueOnce(new Error('unavailable'))
    const store = usePelicanMetadataStore()
    store.applySavedSettings('可用', true)
    await expect(store.load(true)).rejects.toThrow('unavailable')
    expect(store.enabled).toBe(false)
    expect(store.loaded).toBe(false)
  })
})
