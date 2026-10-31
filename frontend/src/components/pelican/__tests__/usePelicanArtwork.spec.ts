import { flushPromises } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { createPelicanArtwork } from '../usePelicanArtwork'
import type { PelicanResult } from '@/api/pelicanTests'

const result = (id: number, html = '<p>作品</p>') => ({ id, status: 'success', html }) as PelicanResult
function deferred() {
  let resolve!: (result: PelicanResult) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise<PelicanResult>((yes, no) => { resolve = yes; reject = no })
  return { promise, resolve, reject }
}
describe('page artwork scheduler', () => {
  it('shares immutable IDs and keeps canceled transports in the three-slot budget until settlement', async () => {
    const jobs = Array.from({ length: 5 }, deferred)
    const fetcher = vi.fn((id: number) => jobs[id - 1]!.promise)
    const store = createPelicanArtwork(undefined, fetcher)
    const abort = new AbortController()
    const first = store.request(1, abort.signal).catch(() => null)
    const same = store.request(1)
    const rest = [2, 3, 4, 5].map(id => store.request(id))
    expect(fetcher).toHaveBeenCalledTimes(3)
    abort.abort()
    expect(fetcher).toHaveBeenCalledTimes(3)
    jobs[0]!.resolve(result(1)); await flushPromises()
    expect((await same).id).toBe(1)
    expect(fetcher).toHaveBeenCalledTimes(4)
    jobs[1]!.resolve(result(2)); jobs[2]!.resolve(result(3)); await flushPromises()
    jobs[3]!.resolve(result(4)); jobs[4]!.resolve(result(5))
    await Promise.all([first, ...rest])
    await store.request(1)
    expect(fetcher).toHaveBeenCalledTimes(5)
    store.dispose()
  })

  it('isolates filter cancellation and identity invalidation from ignored-abort late completions', async () => {
    const old = deferred()
    const fresh = deferred()
    const fetcher = vi.fn().mockReturnValueOnce(old.promise).mockReturnValueOnce(fresh.promise)
    const store = createPelicanArtwork(undefined, fetcher)
    const pending = store.request(7).catch(() => null)
    store.cancelPending()
    const newer = store.request(7)
    old.resolve(result(7, 'old private')); await flushPromises()
    expect(store.peek(7)).toBeUndefined()
    fresh.resolve(result(7, 'fresh')); await newer
    expect(store.peek(7)?.result.html).toBe('fresh')
    store.invalidate()
    expect(store.stats().entries).toBe(0)
    await expect(store.request(7)).rejects.toHaveProperty('name', 'AbortError')
    await pending
  })

  it('does not start replacement work while all three canceled transports are unsettled', async () => {
    const jobs = Array.from({ length: 4 }, deferred)
    const fetcher = vi.fn((id: number) => jobs[id - 1]!.promise)
    const store = createPelicanArtwork(undefined, fetcher)
    const abandoned = [1, 2, 3].map(id => store.request(id).catch(() => null))
    store.cancelPending()
    const replacement = store.request(4)
    expect(fetcher).toHaveBeenCalledTimes(3)
    expect(store.stats().running).toBe(3)
    jobs[0]!.resolve(result(1)); await flushPromises()
    expect(fetcher).toHaveBeenCalledTimes(4)
    expect(store.peek(1)).toBeUndefined()
    jobs[1]!.resolve(result(2)); jobs[2]!.resolve(result(3)); jobs[3]!.resolve(result(4))
    await Promise.all([...abandoned, replacement])
    store.dispose()
  })

  it('invalidates every consumer/cache on access loss and never caches a late success', async () => {
    const denied = deferred(), late = deferred()
    const onDenied = vi.fn()
    const store = createPelicanArtwork(onDenied, vi.fn().mockReturnValueOnce(denied.promise).mockReturnValueOnce(late.promise))
    const first = store.request(1).catch(() => null)
    const second = store.request(2).catch(() => null)
    denied.reject({ response: { status: 403 } }); await flushPromises()
    late.resolve(result(2)); await flushPromises()
    expect(onDenied).toHaveBeenCalledOnce()
    expect(store.allowed.value).toBe(false)
    expect(store.stats().entries).toBe(0)
    await Promise.all([first, second])
  })

  it('uses LRU with both 60-result and 6MiB string limits, including height eviction and oversized results', async () => {
    const fetcher = vi.fn(async (id: number) => result(id))
    const store = createPelicanArtwork(undefined, fetcher)
    for (let id = 1; id <= 60; id++) await store.request(id)
    store.setHeight(2, 500)
    await store.request(1)
    await store.request(61)
    expect(store.peek(2)).toBeUndefined()
    expect(store.height(2)).toBeUndefined()
    expect(store.peek(1)).toBeDefined()
    fetcher.mockImplementation(async id => result(id, 'x'.repeat(2 * 1024 * 1024)))
    await store.request(62); await store.request(63)
    expect(store.stats().bytes).toBeLessThanOrEqual(6 * 1024 * 1024)
    fetcher.mockImplementation(async id => result(id, 'x'.repeat(4 * 1024 * 1024)))
    expect((await store.request(64)).html).toHaveLength(4 * 1024 * 1024)
    expect(store.peek(64)).toBeUndefined()
    store.dispose()
  })

  it('does not retain failures or auto retry and pauses queued background work', async () => {
    const fetcher = vi.fn().mockRejectedValueOnce(new Error('network')).mockResolvedValue(result(1))
    const store = createPelicanArtwork(undefined, fetcher)
    await expect(store.request(1)).rejects.toThrow('network')
    await flushPromises()
    expect(fetcher).toHaveBeenCalledOnce()
    store.setPaused(true)
    const retry = store.request(1)
    expect(fetcher).toHaveBeenCalledOnce()
    store.setPaused(false)
    await retry
    expect(fetcher).toHaveBeenCalledTimes(2)
    store.dispose()
  })
})
