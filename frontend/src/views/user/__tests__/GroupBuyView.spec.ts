import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import GroupBuyView from '../GroupBuyView.vue'

const {
  activity,
  bindKey,
  createOrder,
  listPlans,
  myOrders,
  mySeats,
  cachedPublicSettings,
  showError,
  showSuccess,
  routerResolve,
} = vi.hoisted(() => ({
  activity: vi.fn(),
  bindKey: vi.fn(),
  createOrder: vi.fn(),
  listPlans: vi.fn(),
  myOrders: vi.fn(),
  mySeats: vi.fn(),
  cachedPublicSettings: {
    group_buy_product_name: '我的拼团',
    group_buy_description: '后台配置的我的拼团顶部说明',
  },
  showError: vi.fn(),
  showSuccess: vi.fn(),
  routerResolve: vi.fn(() => ({ href: '/payment/stripe?order_id=501' })),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      resolve: routerResolve,
    }),
  }
})

vi.mock('@/api/groupBuy', () => ({
  groupBuyAPI: {
    listPlans,
    activity,
    createOrder,
    mySeats,
    myOrders,
    bindKey,
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    cachedPublicSettings,
    showError,
    showSuccess,
  }),
}))

vi.mock('@/utils/device', () => ({
  isMobileDevice: () => false,
}))

vi.mock('@/components/payment/paymentFlow', async () => {
  const actual = await vi.importActual<typeof import('@/components/payment/paymentFlow')>('@/components/payment/paymentFlow')
  return {
    ...actual,
    writePaymentRecoverySnapshot: vi.fn(),
    clearPaymentRecoverySnapshot: vi.fn(),
  }
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const IconStub = { props: ['name'], template: '<span class="icon-stub" />' }
const PaymentStatusPanelStub = {
  props: ['orderId', 'orderType'],
  emits: ['done', 'success', 'settled'],
  template: '<section data-testid="payment-status-panel">{{ orderType }} #{{ orderId }}</section>',
}

const plan = {
  id: 11,
  title: 'Token拼拼拼 10 份团',
  description: 'Token拼拼拼 平台托管容量份额，满份后自动开通。',
  product_key: 'token_pinpinpin',
  total_shares: 10,
  seat_count: 2,
  price_per_share: 128,
  price_per_seat: 128,
  price_label: '每份 128 元',
  quota_per_share_label: '单份约 50 USD 月额度',
  quota_label: '单份约 50 USD 月额度',
  max_shares_per_user: 10,
  target_group_id: 7,
  tier_group_ids: Object.fromEntries(Array.from({ length: 10 }, (_, index) => [String(index + 1), 7])),
  tier_groups: [{ min_shares: 1, max_shares: 10, target_group_id: 7, label: '通用权益' }],
  tier_rules: [{ min_shares: 1, max_shares: 10, target_group_id: 7, label: '通用权益' }],
  validity_days: 30,
  timeout_minutes: 1440,
  launch_mode: 'auto',
  refund_mode: 'balance_credit',
  agreement_text: '我理解这是 Token拼拼拼 平台托管容量份额，不是官方 OpenAI Pro 子账号。',
  status: 'active',
  sort_order: 1,
  current_round: {
    id: 31,
    plan_id: 11,
    status: 'open',
    total_shares: 10,
    paid_shares: 4,
    reserved_shares: 0,
    available_shares: 6,
    total_seats: 2,
    paid_seats: 1,
    reserved_seats: 0,
    available_seats: 1,
    deadline_at: '2099-01-01T00:00:00Z',
    created_at: '2026-07-08T00:00:00Z',
    updated_at: '2026-07-08T00:00:00Z',
  },
  created_at: '2026-07-08T00:00:00Z',
  updated_at: '2026-07-08T00:00:00Z',
}

const activeSeat = {
  id: 41,
  round_id: 31,
  plan_id: 11,
  user_id: 5,
  order_id: 501,
  status: 'active',
  share_count: 2,
  subscription_id: 71,
  bound_api_key_id: undefined,
  expires_at: '2099-02-01T00:00:00Z',
  plan,
  round: plan.current_round,
  created_at: '2026-07-08T00:00:00Z',
  updated_at: '2026-07-08T00:00:00Z',
}

const refundSeat = {
  ...activeSeat,
  id: 42,
  status: 'refund_pending',
  subscription_id: undefined,
  bound_api_key_id: undefined,
}

function mountView() {
  return mount(GroupBuyView, {
    attachTo: document.body,
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        Icon: IconStub,
        PaymentStatusPanel: PaymentStatusPanelStub,
        Teleport: true,
        Transition: false,
      },
    },
  })
}

describe('GroupBuyView', () => {
  beforeEach(() => {
    listPlans.mockReset().mockResolvedValue({ data: [plan] })
    activity.mockReset().mockResolvedValue({
      data: [{
        id: 1,
        plan_id: 11,
        round_id: 31,
        event_type: 'shares_paid',
        message: '份额付款成功，等待满份成团',
        created_at: '2026-07-08T00:00:00Z',
      }],
    })
    mySeats.mockReset().mockResolvedValue({
      data: {
        entitlement: {
          id: 61,
          user_id: 5,
          product_key: 'token_pinpinpin',
          status: 'active',
          active_share_count: 2,
          target_group_id: 7,
          target_group: { id: 7, name: 'Token拼拼拼 2 份档', platform: 'openai', monthly_limit_usd: 100 },
          entitlement_label: '2 份权益',
          subscription_id: 71,
          bound_api_key_id: undefined,
          expires_at: '2099-02-01T00:00:00Z',
          refreshed_at: '2026-07-08T00:00:00Z',
          created_at: '2026-07-08T00:00:00Z',
          updated_at: '2026-07-08T00:00:00Z',
        },
        seats: [activeSeat, refundSeat],
      },
    })
    myOrders.mockReset().mockResolvedValue({
      data: {
        items: [{
          id: 501,
          amount: 128,
          pay_amount: 128,
          currency: 'CNY',
          payment_type: 'stripe',
          out_trade_no: 'sub2_group_501',
          status: 'REFUND_PENDING',
          order_type: 'group_buy',
          created_at: '2026-07-08T00:00:00Z',
          expires_at: '2099-01-01T00:00:00Z',
        }],
        total: 1,
        page: 1,
        page_size: 20,
        pages: 1,
      },
    })
    createOrder.mockReset().mockResolvedValue({
      data: {
        order_id: 501,
        amount: 128,
        pay_amount: 128,
        fee_rate: 0,
        client_secret: 'cs_group_buy',
        expires_at: '2099-01-01T00:10:00.000Z',
        out_trade_no: 'sub2_group_501',
      },
    })
    bindKey.mockReset().mockResolvedValue({ data: { ...activeSeat, bound_api_key_id: 9 } })
    routerResolve.mockClear()
    showError.mockReset()
    showSuccess.mockReset()
    vi.stubGlobal('open', vi.fn(() => ({ closed: false })))
    window.localStorage.clear()
  })

  it('renders hall cards, activity, and hosted-capacity boundary copy', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(listPlans).toHaveBeenCalled()
    expect(activity).toHaveBeenCalledWith(20)
    expect(wrapper.text()).toContain('我的拼团')
    expect(wrapper.text()).toContain('Token拼拼拼 10 份团')
    expect(wrapper.text()).toContain('每份 128 元')
    expect(wrapper.text()).toContain('单份约 50 USD 月额度')
    expect(wrapper.text()).toContain('份额付款成功，等待满份成团')
    expect(wrapper.text()).toContain('后台配置的我的拼团顶部说明')
    expect(wrapper.text()).toContain('只使用自己的平台 API Key')
    expect(wrapper.text()).toContain('满份成团后按有效份额开通权益')
    expect(wrapper.text()).not.toContain('不共享官方账号或官方 API Key')
  })

  it('keeps submit disabled until the agreement is accepted and then starts group-buy payment', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('选择份额并付款'))?.trigger('click')
    await flushPromises()

    const plusButton = wrapper.findAll('button').find((button) => button.text() === '+')
    await plusButton?.trigger('click')

    const submitButton = wrapper.findAll('button').find((button) => button.text().includes('确认份额并付款'))
    expect(submitButton?.attributes('disabled')).toBeDefined()

    const agreement = wrapper.find('input[type="checkbox"]')
    await agreement.setValue(true)
    await submitButton?.trigger('click')
    await flushPromises()

    expect(createOrder).toHaveBeenCalledWith(expect.objectContaining({
      plan_id: 11,
      share_count: 2,
      payment_type: 'alipay',
      payment_source: 'hosted_redirect',
      is_mobile: false,
    }))
    expect(wrapper.find('[data-testid="payment-status-panel"]').text()).toContain('group_buy #501')
  })

  it('shows binding tab states and binds an active seat key', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('使用与绑定'))?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('已开通')
    expect(wrapper.text()).toContain('待退款')
    expect(wrapper.text()).toContain('当前有效份额')
    expect(wrapper.text()).toContain('2 份权益')

    const input = wrapper.find('input[placeholder="API Key ID"]')
    await input.setValue('9')
    await wrapper.findAll('button').find((button) => button.text().includes('绑定 Key'))?.trigger('click')
    await flushPromises()

    expect(bindKey).toHaveBeenCalledWith(41, { api_key_id: 9 })
    expect(showSuccess).toHaveBeenCalledWith('绑定成功')
  })

  it('shows group-buy order statuses without replacing the ordinary order list', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('拼团订单'))?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('这里仅展示 我的拼团 订单及份额状态，不替代普通订单列表')
    expect(wrapper.text()).toContain('#501')
    expect(wrapper.text()).not.toContain('sub2_group_501')
    expect(wrapper.text()).toContain('待退款')
  })
})
