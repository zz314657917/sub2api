import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { usePaymentStore } from '../payment'

const getConfig = vi.hoisted(() => vi.fn())
vi.mock('@/api/payment', () => ({ paymentAPI: { getConfig } }))

describe('payment configuration requests', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    getConfig.mockReset()
  })

  it('returns the shared in-flight configuration to concurrent callers', async () => {
    let resolve!: (value: { data: { payment_enabled: boolean } }) => void
    getConfig.mockReturnValue(new Promise(done => { resolve = done }))
    const store = usePaymentStore()
    const first = store.fetchConfig()
    const second = store.fetchConfig()
    expect(getConfig).toHaveBeenCalledOnce()
    resolve({ data: { payment_enabled: true } })
    expect(await Promise.all([first, second])).toEqual([{ payment_enabled: true }, { payment_enabled: true }])
    expect(store.configLoading).toBe(false)
  })
})
