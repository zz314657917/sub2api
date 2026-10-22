import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, shallowMount } from '@vue/test-utils'
import PaymentView from '../PaymentView.vue'
import { PAYMENT_RECOVERY_STORAGE_KEY, readPaymentRecoverySnapshot } from '@/components/payment/paymentFlow'

const mocks = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
  createOrder: vi.fn(), resolve: vi.fn(), push: vi.fn(), replace: vi.fn(),
  mobile: false, open: vi.fn(), navigate: vi.fn(), showError: vi.fn(), showWarning: vi.fn(),
}))
vi.mock('vue-router', async () => ({
  ...await vi.importActual<typeof import('vue-router')>('vue-router'),
  useRoute: () => ({ path: '/purchase', query: mocks.query }),
  useRouter: () => ({ resolve: mocks.resolve, push: mocks.push, replace: mocks.replace }),
}))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key }),
}))
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ user: { username: 'test', balance: 0 }, refreshUser: async () => {} }),
}))
vi.mock('@/stores/payment', () => ({ usePaymentStore: () => ({ createOrder: mocks.createOrder }) }))
vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({ activeSubscriptions: [], fetchActiveSubscriptions: async () => {} }),
}))
vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError: mocks.showError, showWarning: mocks.showWarning, showInfo: vi.fn(), showSuccess: vi.fn(),
  }),
}))
vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getCheckoutInfo: async () => ({ data: {
      methods: { wxpay: { available: true, fee_rate: 0 } },
      global_min: 0, global_max: 0, plans: [], balance_disabled: false,
      balance_recharge_multiplier: 1, recharge_fee_rate: 0, help_text: '', help_image_url: '',
    } }),
    getMembershipStatus: async () => ({ data: null }),
    getMyOrders: async () => ({ data: { items: [], total: 0 } }),
    getInvoiceSummary: async () => ({ data: { available_amount: 0 } }),
  },
}))
vi.mock('@/utils/device', () => ({ isMobileDevice: () => mocks.mobile }))
enableAutoUnmount(afterEach)
afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})
const originalWindow = window
let navigatedTo = ''
beforeEach(() => {
  vi.resetAllMocks()
  originalWindow.localStorage.clear()
  mocks.query = { wechat_resume: '1', wechat_resume_token: 'incoming-token', payment_type: 'alipay' }
  mocks.mobile = false
  navigatedTo = ''
  const location = {
    origin: originalWindow.location.origin,
    get href() { return navigatedTo },
    set href(url: string) { navigatedTo = url; mocks.navigate(url) },
  }
  vi.stubGlobal('window', new Proxy(originalWindow, {
    get(target, key) {
      if (key === 'location') return location
      if (key === 'open') return mocks.open
      return Reflect.get(target, key, target)
    },
  }))
  mocks.open.mockReturnValue({ closed: false })
  mocks.replace.mockResolvedValue(undefined)
  mocks.resolve.mockImplementation(({ path, query }: { path: string; query: Record<string, string | undefined> }) => {
    const params = new URLSearchParams()
    Object.entries(query).forEach(([key, value]) => {
      if (value !== undefined) params.set(key, value)
    })
    return { href: `${path}?${params}` }
  })
})
const stripeResult = () => ({
  order_id: 42, amount: 88, pay_amount: 88, currency: 'HKD', fee_rate: 0,
  expires_at: '2099-01-01T00:00:00Z', payment_type: 'stripe',
  client_secret: 'pi_only_in_snapshot_secret_42', resume_token: 'resume-42',
})
async function startFlow() {
  const wrapper = shallowMount(PaymentView, {
    global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, Teleport: true, Transition: false } },
  })
  await flushPromises()
  await flushPromises()
  return wrapper
}
function expectSafeRoute(method?: string) {
  const route = mocks.resolve.mock.calls.find(([value]) => value.path === '/payment/stripe')![0]
  expect(route.query).toEqual({ order_id: '42', method, resume_token: 'resume-42' })
  expect(route.query).not.toHaveProperty('client_secret')
  const href = mocks.resolve.mock.results.find(result => result.value?.href?.startsWith('/payment/stripe'))!.value.href
  expect(href).not.toContain('client_secret')
  expect(href).not.toContain('pi_only_in_snapshot_secret_42')
  return href
}

describe('PaymentView actual Stripe launch paths', () => {
  it.each([
    ['alipay', false, 'alipay', true],
    ['alipay', true, 'alipay', false],
    ['wxpay', true, 'wechat_pay', false],
    ['stripe', false, undefined, false],
  ] as const)('creates the %s route (mobile=%s) after persisting recovery', async (method, mobile, stripeMethod, popup) => {
    mocks.query.payment_type = method
    mocks.mobile = mobile
    mocks.createOrder.mockResolvedValue(stripeResult())
    const setItem = vi.spyOn(Storage.prototype, 'setItem')
    const navigationSnapshots: unknown[] = []
    const captureSnapshot = () => {
      navigationSnapshots.push(readPaymentRecoverySnapshot(
        originalWindow.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY),
        { resumeToken: 'resume-42' },
      ))
    }
    mocks.open.mockImplementation(() => { captureSnapshot(); return { closed: false } })
    mocks.navigate.mockImplementation(captureSnapshot)
    await startFlow()
    expect(mocks.createOrder).toHaveBeenCalledTimes(1)
    expect(mocks.createOrder).toHaveBeenCalledWith(expect.objectContaining({
      payment_type: method, wechat_resume_token: 'incoming-token', is_mobile: mobile,
    }))
    const href = expectSafeRoute(stripeMethod)
    expect(navigationSnapshots).toEqual([expect.objectContaining({
      orderId: 42, clientSecret: 'pi_only_in_snapshot_secret_42', currency: 'HKD', resumeToken: 'resume-42',
    })])
    const storageCall = setItem.mock.calls.findIndex(([key]) => key === PAYMENT_RECOVERY_STORAGE_KEY)
    expect(storageCall).toBeGreaterThanOrEqual(0)
    const navigation = popup ? mocks.open : mocks.navigate
    expect(setItem.mock.invocationCallOrder[storageCall]).toBeLessThan(navigation.mock.invocationCallOrder[0])
    if (popup) {
      expect(mocks.open).toHaveBeenCalledWith(href, 'paymentPopup', expect.any(String))
      expect(mocks.navigate).not.toHaveBeenCalled()
    } else {
      expect(mocks.navigate).toHaveBeenCalledWith(href)
      expect(mocks.open).not.toHaveBeenCalled()
    }
  })

  it('keeps popup-blocked redirect UX with a persisted snapshot and safe URL', async () => {
    mocks.createOrder.mockResolvedValue(stripeResult())
    mocks.open.mockReturnValue(null)
    await startFlow()
    expect(mocks.navigate).toHaveBeenCalledWith(expectSafeRoute('alipay'))
    expect(readPaymentRecoverySnapshot(originalWindow.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)))
      .toMatchObject({ orderId: 42, clientSecret: 'pi_only_in_snapshot_secret_42' })
  })

  it('executes the mobile fallback route construction but rejects a Stripe result without navigation', async () => {
    mocks.mobile = true
    mocks.query.payment_type = 'wxpay'
    mocks.createOrder.mockRejectedValueOnce({ reason: 'WECHAT_H5_NOT_AUTHORIZED' })
      .mockResolvedValueOnce(stripeResult())
    await startFlow()
    expect(mocks.createOrder).toHaveBeenCalledTimes(2)
    expect(mocks.createOrder).toHaveBeenNthCalledWith(2, expect.objectContaining({
      payment_type: 'wxpay', is_mobile: false, payment_source: 'hosted_redirect',
    }))
    expectSafeRoute('wechat_pay')
    expect(mocks.open).not.toHaveBeenCalled()
    expect(mocks.navigate).not.toHaveBeenCalled()
    expect(navigatedTo).toBe('')
    expect(mocks.push).not.toHaveBeenCalled()
    expect(originalWindow.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).toBeNull()
    expect(mocks.showWarning).not.toHaveBeenCalled()
    expect(mocks.showError).toHaveBeenCalled()
  })
})
