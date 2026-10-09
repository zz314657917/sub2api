import { flushPromises, shallowMount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ServiceStoreView from '@/views/user/ServiceStoreView.vue'

const routeState = vi.hoisted(() => ({ query: {} as Record<string, string> }))
const routerReplace = vi.hoisted(() => vi.fn())
const api = vi.hoisted(() => ({
  getProducts: vi.fn(),
  createOrder: vi.fn(),
  getOrder: vi.fn(),
  resumeOrder: vi.fn(),
  getLimits: vi.fn(),
  launch: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return { ...actual, useRoute: () => routeState, useRouter: () => ({ replace: routerReplace }) }
})
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ user: { id: 71 } }) }))
vi.mock('@/api/digitalStore', () => ({ digitalStoreAPI: api }))
vi.mock('@/api/payment', () => ({ paymentAPI: api }))

const product = {
  id: 9,
  name: 'Fixture card',
  description: 'fixture only',
  kind: 'card' as const,
  price_cents: 1200,
  stock_available: 1,
  enabled: true,
  created_at: '',
}

function render() {
  return shallowMount(ServiceStoreView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        RouterLink: { template: '<a><slot /></a>' },
        BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' },
        StorePaymentLauncher: { template: '<div />', methods: { launch: api.launch } },
      },
    },
  })
}

function button(wrapper: ReturnType<typeof render>, text: string) {
  const found = wrapper.findAll('button').find((item) => item.text().includes(text))
  if (!found) throw new Error(`missing button: ${text}`)
  return found
}

describe('digital store checkout mount behavior', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    routeState.query = {}
    routerReplace.mockReset().mockImplementation(async ({ query }: { query: Record<string, string> }) => { routeState.query = query })
    window.localStorage.clear()
    vi.stubGlobal('crypto', { randomUUID: vi.fn(() => 'fixture-idempotency-key') })
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: false })))
    api.getProducts.mockResolvedValue({ data: { items: [product], total: 1, page: 1, page_size: 20 } })
    api.getLimits.mockResolvedValue({ data: { methods: { wxpay: { available: true, currency: 'CNY' } } } })
    api.createOrder.mockResolvedValue({ data: { order_id: 44, payment_type: 'wxpay', qr_code: 'weixin://wxpay/fixture' } })
  })

  it('submits only server-owned store context with one stable idempotency key', async () => {
    const wrapper = render()
    try {
      await flushPromises()
      await button(wrapper, '购买并付款').trigger('click')
      await flushPromises()
      await button(wrapper, '提交订单并付款').trigger('click')
      await flushPromises()

      expect(api.createOrder).toHaveBeenCalledWith(expect.objectContaining({
        product_id: 9,
        idempotency_key: 'fixture-idempotency-key',
        payment_type: 'wxpay',
        payment_source: 'store_checkout',
      }))
      expect(api.createOrder.mock.calls[0][0]).not.toHaveProperty('amount')
      expect(api.launch).toHaveBeenCalledWith(expect.objectContaining({ order_id: 44 }), 'wxpay')
      expect(window.localStorage.getItem('store.checkout.idempotency.71.9')).toBe('fixture-idempotency-key')
    } finally {
      wrapper.unmount()
    }
  })

  it('keeps the retry key and renders a visible failure for an invalid checkout response', async () => {
    api.createOrder.mockResolvedValue({ data: null })
    const wrapper = render()
    try {
      await flushPromises()
      await button(wrapper, '购买并付款').trigger('click')
      await flushPromises()
      await button(wrapper, '提交订单并付款').trigger('click')
      await flushPromises()

      const message = wrapper.get('[role="alert"]').text()
      expect(message).toContain('订单响应异常，请到我的商店订单核对状态后重试，不要重复下单。')
      expect(message).not.toContain('Cannot read properties')
      expect(window.localStorage.getItem('store.checkout.idempotency.71.9')).toBe('fixture-idempotency-key')
      expect(api.launch).not.toHaveBeenCalled()
    } finally {
      wrapper.unmount()
    }
  })

  it('generates a new key only after the stored order is terminal and the user explicitly repurchases', async () => {
    window.localStorage.setItem('store.checkout.idempotency.71.9', 'old-key')
    window.localStorage.setItem('store.checkout.idempotency.71.9.order', '43')
    api.getOrder.mockResolvedValue({ data: { payment_status: 'COMPLETED' } })
    vi.stubGlobal('crypto', { randomUUID: vi.fn(() => 'new-key') })
    const wrapper = render()
    try {
      await flushPromises()
      await button(wrapper, '购买并付款').trigger('click')
      await flushPromises()
      await button(wrapper, '提交订单并付款').trigger('click')
      await flushPromises()
      expect(button(wrapper, '再次购买').exists()).toBe(true)
      await button(wrapper, '再次购买').trigger('click')
      await flushPromises()

      expect(api.createOrder).toHaveBeenCalledWith(expect.objectContaining({ idempotency_key: 'new-key' }))
      expect(window.localStorage.getItem('store.checkout.idempotency.71.9')).toBe('new-key')
    } finally {
      wrapper.unmount()
    }
  })

  it('strips a signed WeChat resume query before explicit terminal repurchase', async () => {
    routeState.query = { wechat_resume_token: 'signed-fixture', product_id: '9', payment_type: 'wxpay', order_type: 'store' }
    window.localStorage.setItem('store.checkout.idempotency.71.9', 'old-key')
    window.localStorage.setItem('store.checkout.idempotency.71.9.order', '43')
    vi.stubGlobal('crypto', { randomUUID: vi.fn(() => 'new-key') })
    api.createOrder.mockResolvedValueOnce({ data: { order_id: 44, status: 'COMPLETED', payment_type: 'wxpay' } })
    api.getOrder.mockResolvedValue({ data: { payment_status: 'COMPLETED' } })
    const wrapper = render()
    try {
      await flushPromises()
      await button(wrapper, '购买并付款').trigger('click')
      await flushPromises()
      await button(wrapper, '提交订单并付款').trigger('click')
      await flushPromises()
      expect(api.createOrder).toHaveBeenCalledWith(expect.objectContaining({ wechat_resume_token: 'signed-fixture', idempotency_key: 'old-key' }))
      expect(button(wrapper, '再次购买').exists()).toBe(true)
      await button(wrapper, '再次购买').trigger('click')
      await flushPromises()
      expect(routerReplace).toHaveBeenCalledWith({ query: { product_id: '9' } })
      expect(api.createOrder).toHaveBeenLastCalledWith(expect.objectContaining({ idempotency_key: 'new-key', wechat_resume_token: undefined }))
    } finally {
      wrapper.unmount()
    }
  })
})
