import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, shallowMount } from '@vue/test-utils'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
}))
const routerPush = vi.hoisted(() => vi.fn())
const getOrder = vi.hoisted(() => vi.fn())
const paymentStore = vi.hoisted(() => ({
  config: { stripe_publishable_key: 'pk_test' } as { stripe_publishable_key?: string },
  fetchConfig: vi.fn(),
  pollOrderStatus: vi.fn(),
}))
const loadStripe = vi.hoisted(() => vi.fn())
const stripeElements = vi.hoisted(() => ({
  create: vi.fn(),
}))
const stripePaymentElement = vi.hoisted(() => ({
  mount: vi.fn(),
  on: vi.fn(),
}))
const stripeInstance = vi.hoisted(() => ({
  elements: vi.fn(),
  confirmPayment: vi.fn(),
  confirmAlipayPayment: vi.fn(),
  confirmWechatPayPayment: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: routerPush }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: 'zh-CN' },
    }),
  }
})

vi.mock('@/stores/payment', () => ({
  usePaymentStore: () => paymentStore,
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getOrder,
  },
}))

vi.mock('@stripe/stripe-js/pure', () => ({
  loadStripe,
}))
vi.mock('@/utils/device', () => ({ isMobileDevice: () => false }))

import StripePaymentView from '../StripePaymentView.vue'
import { formatPaymentAmount } from '@/components/payment/currency'
import type { PaymentOrder } from '@/types/payment'
import { PAYMENT_RECOVERY_STORAGE_KEY, type PaymentRecoverySnapshot } from '@/components/payment/paymentFlow'

enableAutoUnmount(afterEach)
afterEach(() => vi.restoreAllMocks())

function snapshot(overrides: Partial<PaymentRecoverySnapshot> = {}): PaymentRecoverySnapshot {
  return {
    orderId: 42, amount: 100, payAmount: 103, qrCode: '', expiresAt: '2099-01-01T00:00:00Z',
    paymentType: 'stripe', payUrl: '', outTradeNo: 'sub2_stripe_42', clientSecret: 'snapshot_secret_42',
    intentId: '', currency: '', countryCode: '', paymentEnv: '', orderType: 'balance',
    paymentMode: '', resumeToken: 'resume-42', createdAt: Date.now(), ...overrides,
  }
}

function saveSnapshot(overrides: Partial<PaymentRecoverySnapshot> = {}) {
  window.localStorage.setItem(PAYMENT_RECOVERY_STORAGE_KEY, JSON.stringify(snapshot(overrides)))
}

function orderFactory(overrides: Partial<PaymentOrder> = {}): PaymentOrder {
  return {
    id: 42,
    user_id: 7,
    amount: 100,
    pay_amount: 103,
    currency: 'CNY',
    fee_rate: 0.03,
    payment_type: 'stripe',
    out_trade_no: 'sub2_stripe_42',
    status: 'PENDING',
    order_type: 'balance',
    created_at: '2026-04-20T12:00:00Z',
    expires_at: '2026-04-20T12:30:00Z',
    refund_amount: 0,
    ...overrides,
  }
}

function mountView() {
  return shallowMount(StripePaymentView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: true,
      },
    },
  })
}

describe('StripePaymentView', () => {
  beforeEach(() => {
    routeState.query = {
      order_id: '42',
      client_secret: 'pi_secret_42',
    }
    routerPush.mockReset()
    getOrder.mockReset()
    paymentStore.config = { stripe_publishable_key: 'pk_test' }
    paymentStore.fetchConfig.mockReset().mockResolvedValue(undefined)
    paymentStore.pollOrderStatus.mockReset()
    loadStripe.mockReset().mockResolvedValue(stripeInstance)
    stripeElements.create.mockReset().mockReturnValue(stripePaymentElement)
    stripePaymentElement.mount.mockReset()
    stripePaymentElement.on.mockReset().mockImplementation((event: string, callback: () => void) => {
      if (event === 'ready') callback()
    })
    stripeInstance.elements.mockReset().mockReturnValue(stripeElements)
    stripeInstance.confirmPayment.mockReset()
    stripeInstance.confirmAlipayPayment.mockReset().mockResolvedValue({ error: { message: 'test alipay response' } })
    stripeInstance.confirmWechatPayPayment.mockReset().mockResolvedValue({ error: { message: 'test wechat response' } })
    window.localStorage.clear()
    getOrder.mockResolvedValue({ data: orderFactory() })
  })

  it('有效同单快照 currency 为空时使用订单接口返回的 Stripe 币种展示金额', async () => {
    saveSnapshot({ currency: '' })
    getOrder.mockResolvedValue({
      data: orderFactory({ currency: 'HKD', pay_amount: 103 }),
    })

    const wrapper = mountView()
    await flushPromises()
    await flushPromises()

    expect(getOrder).toHaveBeenCalledWith(42)
    expect(loadStripe).toHaveBeenCalledWith('pk_test')
    expect(wrapper.text()).toContain(formatPaymentAmount(103, 'HKD', 'zh-CN'))
  })

  it('falls back to snapshot currency when the order has none', async () => {
    saveSnapshot({ currency: 'USD' })
    getOrder.mockResolvedValue({ data: orderFactory({ currency: '' }) })
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain(formatPaymentAmount(103, 'USD', 'zh-CN'))
  })

  it.each(['alipay', 'wechat_pay', ''])('initializes %s only with the matching snapshot secret and no URL secret', async (method) => {
    saveSnapshot()
    routeState.query = { order_id: '42', method, resume_token: 'resume-42' }
    mountView()
    await flushPromises()
    await flushPromises()
    expect(loadStripe).toHaveBeenCalledTimes(1)
    expect(loadStripe).toHaveBeenCalledWith('pk_test')
    if (method === 'alipay') {
      expect(stripeInstance.confirmAlipayPayment).toHaveBeenCalledWith('snapshot_secret_42', expect.any(Object))
    } else if (method === 'wechat_pay') {
      expect(stripeInstance.confirmWechatPayPayment).toHaveBeenCalledWith('snapshot_secret_42', expect.any(Object))
    } else {
      expect(stripeInstance.elements).toHaveBeenCalledWith(expect.objectContaining({ clientSecret: 'snapshot_secret_42' }))
      expect(stripePaymentElement.mount).toHaveBeenCalledWith('#stripe-payment-element')
    }
  })

  it.each(['alipay', 'wechat_pay', ''])('ignores a conflicting URL secret for %s', async (method) => {
    saveSnapshot()
    routeState.query.method = method
    routeState.query.client_secret = 'untrusted_url_secret'
    mountView()
    await flushPromises()
    await flushPromises()
    if (method === 'alipay') {
      expect(stripeInstance.confirmAlipayPayment).toHaveBeenCalledWith('snapshot_secret_42', expect.any(Object))
    } else if (method === 'wechat_pay') {
      expect(stripeInstance.confirmWechatPayPayment).toHaveBeenCalledWith('snapshot_secret_42', expect.any(Object))
    } else {
      expect(stripeInstance.elements).toHaveBeenCalledWith(expect.objectContaining({ clientSecret: 'snapshot_secret_42' }))
    }
  })

  it.each([
    ['missing', null],
    ['wrong order', JSON.stringify(snapshot({ orderId: 99 }))],
    ['empty secret', JSON.stringify(snapshot({ clientSecret: '' }))],
    ['whitespace secret', JSON.stringify(snapshot({ clientSecret: '  ' }))],
    ['expired', JSON.stringify(snapshot({ expiresAt: '2020-01-01T00:00:00Z' }))],
    ['invalid expiry', JSON.stringify(snapshot({ expiresAt: 'invalid' }))],
    ['bad JSON', '{'],
    ['malformed', JSON.stringify({ orderId: 42, clientSecret: 'some_secret' })],
    ['null JSON', 'null'],
    ['mismatched token', JSON.stringify(snapshot({ resumeToken: 'other-token' }))],
  ])('fails closed for %s despite a URL secret', async (_case, raw) => {
    if (raw !== null) window.localStorage.setItem(PAYMENT_RECOVERY_STORAGE_KEY, raw)
    routeState.query.resume_token = 'resume-42'
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('payment.stripeMissingParams')
    expect(wrapper.find('.animate-spin').exists()).toBe(false)
    expect(getOrder).not.toHaveBeenCalled()
    expect(loadStripe).not.toHaveBeenCalled()
    expect(stripeInstance.confirmAlipayPayment).not.toHaveBeenCalled()
    expect(stripeInstance.confirmWechatPayPayment).not.toHaveBeenCalled()
    expect(stripeInstance.confirmPayment).not.toHaveBeenCalled()
    expect(stripeInstance.elements).not.toHaveBeenCalled()
    expect(stripePaymentElement.mount).not.toHaveBeenCalled()
    expect(window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)).toBe(raw)
    await wrapper.get('button').trigger('click')
    expect(routerPush).toHaveBeenCalledTimes(1)
    expect(routerPush).toHaveBeenCalledWith('/purchase')
  })

  it.each(['getter', 'getItem'])('handles localStorage %s exceptions without SDK or unhandled rejection', async (source) => {
    if (source === 'getter') {
      vi.spyOn(window, 'localStorage', 'get').mockImplementation(() => { throw new Error('Storage unavailable') })
    } else {
      vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => { throw new Error('Storage unavailable') })
    }
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('payment.stripeLoadFailed')
    expect(wrapper.find('.animate-spin').exists()).toBe(false)
    expect(getOrder).not.toHaveBeenCalled()
    expect(loadStripe).not.toHaveBeenCalled()
    expect(stripeInstance.elements).not.toHaveBeenCalled()
    expect(stripeInstance.confirmAlipayPayment).not.toHaveBeenCalled()
    expect(stripeInstance.confirmWechatPayPayment).not.toHaveBeenCalled()
    expect(stripeInstance.confirmPayment).not.toHaveBeenCalled()
    expect(stripePaymentElement.mount).not.toHaveBeenCalled()
  })

  it.each([
    ['missing order', undefined, 'resume-42'],
    ['wrong order', '99', 'resume-42'],
    ['non-numeric order', 'invalid', 'resume-42'],
    ['non-integer order', '42.5', 'resume-42'],
    ['negative order', '-42', 'resume-42'],
    ['non-finite order', 'Infinity', 'resume-42'],
    ['empty token', '42', ''],
    ['array token', '42', ['resume-42']],
    ['null token', '42', null],
  ])('rejects invalid route binding: %s', async (_case, orderId, token) => {
    saveSnapshot()
    routeState.query.order_id = orderId
    routeState.query.resume_token = token
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('payment.stripeMissingParams')
    expect(wrapper.find('.animate-spin').exists()).toBe(false)
    expect(getOrder).not.toHaveBeenCalled()
    expect(loadStripe).not.toHaveBeenCalled()
    expect(stripeInstance.confirmAlipayPayment).not.toHaveBeenCalled()
    expect(stripeInstance.confirmWechatPayPayment).not.toHaveBeenCalled()
    expect(stripeInstance.confirmPayment).not.toHaveBeenCalled()
    expect(stripeInstance.elements).not.toHaveBeenCalled()
    expect(stripePaymentElement.mount).not.toHaveBeenCalled()
  })
})
