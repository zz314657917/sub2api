import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useSubscriptionStore } from '../subscriptions'

const getActiveSubscriptions = vi.hoisted(() => vi.fn())
vi.mock('@/api/subscriptions', () => ({ default: { getActiveSubscriptions } }))

describe('subscription clear', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    getActiveSubscriptions.mockReset()
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  it('clears loading immediately while an old request is still pending', async () => {
    let resolve!: (value: unknown[]) => void
    getActiveSubscriptions.mockReturnValueOnce(new Promise<unknown[]>(done => { resolve = done }))
    const store = useSubscriptionStore()
    const pending = store.fetchActiveSubscriptions()
    expect(store.loading).toBe(true)
    store.clear()
    expect(store.loading).toBe(false)
    resolve([{ id: 1 }])
    await pending
    expect(store.activeSubscriptions).toEqual([])
    expect(store.loading).toBe(false)
  })
})
