import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import PixelCafePage from '../PixelCafePage.vue'

const overview = vi.hoisted(() => vi.fn())
const lobbyActivity = vi.hoisted(() => vi.fn())
const listRooms = vi.hoisted(() => vi.fn())
const listMyRooms = vi.hoisted(() => vi.fn())
const createOrder = vi.hoisted(() => vi.fn())
const getCheckoutInfo = vi.hoisted(() => vi.fn())
const routeQuery = vi.hoisted(() => ({} as Record<string, string>))
const cachedPublicSettings = vi.hoisted(() => ({
  pixel_cafe_title: '像素网吧',
  pixel_cafe_description: '把每个模型分组变成一间可订阅的数字包间。',
  pixel_cafe_header_visible: true,
}))

vi.mock('@/api/cafe', () => ({ cafeAPI: { overview, lobbyActivity, listRooms, listMyRooms, createOrder } }))
vi.mock('@/api/payment', () => ({ paymentAPI: { getCheckoutInfo } }))
vi.mock('@/stores', () => ({ useAppStore: () => ({ cachedPublicSettings }) }))
vi.mock('vue-router', () => ({
  useRoute: () => ({ query: routeQuery }),
  useRouter: () => ({ resolve: () => ({ href: '/payment/stripe' }) }),
}))
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<main><slot /></main>' } }))
vi.mock('@/components/icons/Icon.vue', () => ({ default: { template: '<i />' } }))
vi.mock('@/components/payment/PaymentStatusPanel.vue', () => ({ default: { template: '<section data-testid="payment-status-panel" />' } }))
vi.mock('@/utils/apiError', () => ({ extractApiErrorMessage: (error: { message?: string }, fallback: string) => error?.message || fallback }))

const room = {
  id: 18,
  code: 'C-018',
  name: 'Plus 包间 18',
  zone_key: 'openai',
  theme_key: 'warm_wood',
  scene_slot_key: 'claude-18',
  featured: true,
  plan: { id: 3, title: 'ChatGPT Plus', description: '', price_per_share: 99, price_label: '99 CNY', validity_days: 30, subscription_tier: 'plus', total_shares: 5, max_buyers: 4, max_shares_per_user: 4, quota_per_share_label: '每份独立 Key 额度', room_key_quota_usd: 500, room_key_rate_limit_5h: 0, room_key_rate_limit_1d: 0, room_key_rate_limit_7d: 100 },
  round: { id: 1008, status: 'open', paid_shares: 2, reserved_shares: 0, remaining_shares: 3, max_buyers: 4, joined_buyers: 1, remaining_buyer_slots: 3, deadline_at: '2026-08-03T12:00:00Z' },
  member_avatars: [{ avatar_seed: 'member-one' }],
  purchase_state: 'available',
}

const myRoom = {
  membership_id: 892,
  status: 'active',
  paid_shares: 2,
  activated_at: '2026-08-25T08:00:00Z',
  expires_at: '2099-09-24T08:00:00Z',
  room: { id: 18, code: 'C-018', name: 'Plus 包间 18', zone_key: 'openai', theme_key: 'warm_wood' },
  member_avatars: [{ avatar_seed: 'my-room-member-one' }, { avatar_seed: 'my-room-member-two' }],
  plan: { id: 3, title: 'ChatGPT Plus', subscription_tier: 'plus', validity_days: 30 },
  round: { id: 1008, status: 'active', paid_shares: 5, total_shares: 5 },
  account: { name: 'ChatGPT Plus 主账号', platform: 'openai', email_masked: 'o***r@example.com', remaining_7d_percent: 62.5 },
  managed_api_key: { id: 3011, name: 'Plus 包间 C-018 / Membership', status: 'disabled', quota: 100, quota_used: 12.3, rate_limit_5h: 10, rate_limit_1d: 20, rate_limit_7d: 80, usage_5h: 2.5, usage_7d: 18.75, reset_at_5h: '2099-08-25T13:00:00Z', reset_at_7d: '2099-09-01T08:00:00Z', protected: true as const },
}

function overviewPayload(rooms = [room]) {
  return {
    data: {
      api_version: 'cafe.v1',
      server_time: '2026-08-03T12:00:00Z',
      zones: [
        { key: 'featured', name: '精选大厅', room_count: 1, open_share_count: 1 },
        { key: 'openai', name: 'OpenAI 区', room_count: 1, open_share_count: 1 },
      ],
      rooms,
      lobby: {
        available: true,
        date: '2026-08-03',
        timezone: 'Asia/Shanghai',
        label: '今日使用用户',
        unique_users: 2,
        successful_requests: 5,
        display_max: 50,
        avatars: [{ avatar_seed: 'abcdef1234567890', seat_index: 1, activity: 'recent' }],
      },
    },
  }
}

function mountPage() {
  return mount(PixelCafePage, {
    global: {
      stubs: {
        RouterLink: { template: '<a><slot /></a>' },
        teleport: true,
      },
    },
  })
}

describe('PixelCafePage', () => {
  beforeEach(() => {
    Object.keys(routeQuery).forEach(key => delete routeQuery[key])
    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: vi.fn(() => ({ matches: false })),
    })
    overview.mockReset().mockResolvedValue(overviewPayload())
    lobbyActivity.mockReset().mockResolvedValue({ data: { available: true, date: '2026-08-03', timezone: 'Asia/Shanghai', label: '今日使用用户', unique_users: 2, successful_requests: 5, display_max: 50, avatars: [{ avatar_seed: 'abcdef1234567890', seat_index: 1, activity: 'recent' }] } })
    listRooms.mockReset().mockResolvedValue({ data: { items: [room], total: 1, page: 1, page_size: 24, pages: 1 } })
    listMyRooms.mockReset().mockResolvedValue({ data: { items: [myRoom], total: 1, page: 1, page_size: 20, pages: 1 } })
    createOrder.mockReset().mockResolvedValue({
      data: {
        order_id: 901,
        amount: 99,
        pay_amount: 99,
        payment_type: 'alipay',
        qr_code: 'https://pay.example.com/qr/901',
        expires_at: '2026-08-03T14:00:00Z',
        room_id: room.id,
        round_id: room.round.id,
        share_count: 1,
      },
    })
    getCheckoutInfo.mockReset().mockResolvedValue({
      data: {
        methods: {
          alipay: { available: true },
          wxpay: { available: false },
          wxpay_direct: { available: false },
          stripe: { available: true },
          airwallex: { available: false },
        },
      },
    })
  })

  it('renders the real overview room data and selected-room status', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(overview).toHaveBeenCalledWith({ room_limit: 24 })
    const roomLists = wrapper.findAll('[data-testid="pixel-cafe-room-list"]')
    expect(roomLists).toHaveLength(1)
    expect(wrapper.find('.pixel-cafe-scene > [data-testid="pixel-cafe-room-list"]').exists()).toBe(true)
    expect(roomLists[0].text()).toContain('Plus 包间 18')
    expect(wrapper.text()).toContain('已售 2/5 份')
    expect(wrapper.find('.pixel-cafe-room-card-badges').text()).toContain('GPTPLUS')
    expect(wrapper.find('.pixel-cafe-room-card').text()).not.toContain('ChatGPT Plus · 5 人共享')
    expect(wrapper.find('.pixel-cafe-room-card-action').exists()).toBe(false)
    expect(wrapper.find('[data-testid="pixel-cafe-room-card-members"]').attributes('aria-label')).toBe('1 人已加入')
    expect(wrapper.findAll('[data-testid="pixel-cafe-room-member-avatar"]')).toHaveLength(1)
    expect(wrapper.find('[data-testid="pixel-cafe-my-room-members"]').attributes('aria-label')).toBe('2 人已加入')
    expect(wrapper.findAll('[data-testid="pixel-cafe-my-room-member-avatar"]')).toHaveLength(2)
    expect(wrapper.find('[data-testid="pixel-cafe-room-dialog"]').exists()).toBe(false)
    expect(wrapper.find('.pixel-cafe-scene .pixel-cafe-inspector').exists()).toBe(false)
    await wrapper.find('.pixel-cafe-room-card').trigger('click')
    expect(wrapper.find('[data-testid="pixel-cafe-room-dialog"]').exists()).toBe(true)
    expect(wrapper.find('.pixel-cafe-inspector').attributes('aria-modal')).toBe('true')
    expect(wrapper.text()).toContain('周期')
    expect(wrapper.text()).toContain('30 天')
    const limits = wrapper.get('[data-testid="pixel-cafe-purchase-limits"]')
    expect(limits.text()).toContain('每份独立 Key 额度')
    expect(limits.text()).toContain('总额度500.00 USD')
    expect(limits.text()).toContain('5H 限额不限')
    expect(limits.text()).toContain('1D 限额不限')
    expect(limits.text()).toContain('7D 限额100.00 USD')
    expect(limits.text()).not.toContain('月限额')
    expect(wrapper.find('[data-testid="pixel-cafe-lobby-activity"]').exists()).toBe(false)
    expect(lobbyActivity).not.toHaveBeenCalled()
    expect(wrapper.text()).not.toContain('abcdef1234567890')
    await wrapper.find('.pixel-cafe-dialog-close').trigger('click')
    expect(wrapper.find('[data-testid="pixel-cafe-room-dialog"]').exists()).toBe(false)
    await wrapper.find('.pixel-cafe-room-card').trigger('click')
    await wrapper.find('.pixel-cafe-dialog-close').trigger('keydown', { key: 'Escape' })
    expect(wrapper.find('[data-testid="pixel-cafe-room-dialog"]').exists()).toBe(false)
  })

  it('shows up to five anonymous member avatars and hides the row for empty rooms', async () => {
    const crowdedRoom = {
      ...room,
      id: 19,
      code: 'C-019',
      name: '拥挤包间',
      plan: { ...room.plan, total_shares: 8, max_buyers: 8 },
      round: { ...room.round, paid_shares: 7, remaining_shares: 1, max_buyers: 8, joined_buyers: 7, remaining_buyer_slots: 1 },
      member_avatars: Array.from({ length: 7 }, (_, index) => ({ avatar_seed: `member-${index + 1}` })),
    }
    const emptyRoom = {
      ...room,
      id: 20,
      code: 'C-020',
      name: '空包间',
      round: { ...room.round, paid_shares: 0, remaining_shares: 5, joined_buyers: 0 },
      member_avatars: [],
    }
    overview.mockResolvedValueOnce(overviewPayload([crowdedRoom, emptyRoom]))
    const wrapper = mountPage()
    await flushPromises()

    const cards = wrapper.findAll('.pixel-cafe-room-card')
    expect(cards[0].findAll('[data-testid="pixel-cafe-room-member-avatar"]')).toHaveLength(5)
    expect(cards[0].find('[data-testid="pixel-cafe-room-card-members"]').text()).toContain('+2')
    expect(cards[1].find('[data-testid="pixel-cafe-room-card-members"]').exists()).toBe(false)
  })

  it('shows ten local-only demo rooms and fifty ambient users without calling the room API', async () => {
    routeQuery.demo = '1'
    const wrapper = mountPage()
    await flushPromises()

    expect(overview).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="pixel-cafe-demo-badge"]').text()).toContain('本地演示数据')
    expect(wrapper.find('[data-testid="pixel-cafe-room-list"]').findAll('.pixel-cafe-room-card')).toHaveLength(10)
    expect(wrapper.findAll('[data-testid="pixel-cafe-lobby-avatar"]')).toHaveLength(0)
    expect(wrapper.findAll('[data-testid="pixel-cafe-fallback-avatar"]')).toHaveLength(16)
    expect(wrapper.text()).toContain('深夜 Pro 包间')
    expect(wrapper.text()).toContain('可购买')
    expect(wrapper.find('.pixel-cafe-scene-art').exists()).toBe(true)
    expect(wrapper.find('.pixel-cafe-front-desk').text()).toContain('共享房间')
    expect(wrapper.find('.pixel-cafe-workbench').exists()).toBe(true)

    await wrapper.find('.pixel-cafe-room-card').trigger('click')
    expect(wrapper.text()).toContain('本地演示不创建订单')
    expect((wrapper.find('.pixel-cafe-primary').element as HTMLButtonElement).disabled).toBe(true)
  })

  it('ignores temporary zone categories and renders one neutral room-list title', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.find('.pixel-cafe-zones').exists()).toBe(false)
    expect(wrapper.find('.pixel-cafe-room-list-heading h2').text()).toBe('全部包间')
    expect(wrapper.text()).not.toContain('精选大厅')
    expect(listRooms).not.toHaveBeenCalled()
  })

  it('uses occupied room seats as anonymous ambient avatars when lobby activity is unavailable', async () => {
    overview.mockResolvedValueOnce({ data: {
      ...overviewPayload().data,
      lobby: { available: false, date: '2026-08-03', timezone: 'Asia/Shanghai', label: '', unique_users: 0, successful_requests: 0, display_max: 0, avatars: [] },
    } })
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.findAll('[data-testid="pixel-cafe-lobby-avatar"]')).toHaveLength(0)
  })

  it('renders configured header copy and can hide the entire header block', async () => {
    cachedPublicSettings.pixel_cafe_title = '模型包间'
    cachedPublicSettings.pixel_cafe_description = '按模型选择独立房间。'
    cachedPublicSettings.pixel_cafe_header_visible = true
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.find('h1').text()).toBe('模型包间')
    expect(wrapper.find('[data-testid="pixel-cafe-description"]').text()).toBe('按模型选择独立房间。')
    expect(wrapper.find('.pixel-cafe-subtitle').exists()).toBe(false)
    const descriptionPosition = wrapper.find('[data-testid="pixel-cafe-description"]').element.compareDocumentPosition(
      wrapper.find('[data-testid="pixel-cafe-room-list"]').element,
    )
    expect(descriptionPosition & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    wrapper.unmount()

    cachedPublicSettings.pixel_cafe_header_visible = false
    const hiddenWrapper = mountPage()
    await flushPromises()
    expect(hiddenWrapper.find('.pixel-cafe-header > div').exists()).toBe(false)
    expect(hiddenWrapper.find('h1').exists()).toBe(false)
    expect(hiddenWrapper.find('[data-testid="pixel-cafe-description"]').exists()).toBe(false)
    hiddenWrapper.unmount()
  })

  it('does not render the legacy group-buy entry', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.find('.pixel-cafe-legacy').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('旧版拼团')
  })

  it('contains horizontal overflow only while the cafe route is mounted', () => {
    document.documentElement.style.overflowX = 'visible'
    document.body.style.overflowX = 'visible'
    const wrapper = mountPage()

    expect(document.documentElement.style.overflowX).toBe('hidden')
    expect(document.body.style.overflowX).toBe('hidden')

    wrapper.unmount()

    expect(document.documentElement.style.overflowX).toBe('visible')
    expect(document.body.style.overflowX).toBe('visible')
  })

  it('loads and displays the current user my-room projection without key material', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(listMyRooms).toHaveBeenCalledWith({ page: 1, page_size: 20, status: 'active,waiting' })
    expect(wrapper.find('[data-testid="pixel-cafe-my-rooms-list"]').text()).toContain('Plus 包间 18')
    expect(wrapper.text()).toContain('绑定账号：ChatGPT Plus 主账号')
    expect(wrapper.find('[data-testid="pixel-cafe-my-room-lifetime"]').text()).toMatch(/^剩余时间：/)
    expect(wrapper.findAll('[data-testid="pixel-cafe-my-room-usage"]')).toHaveLength(1)
    expect(wrapper.find('[data-testid="pixel-cafe-account-limit"]').text()).toContain('账号 7D 剩余')
    expect(wrapper.find('[data-testid="pixel-cafe-account-limit"]').text()).toContain('62.50%')
    expect(wrapper.find('[data-testid="pixel-cafe-account-limit"] [role="progressbar"]').attributes('aria-valuenow')).toBe('62.5')
    expect(wrapper.find('[data-testid="pixel-cafe-account-limit"] [role="progressbar"] > span').attributes('style')).toContain('width: 62.5%')
    expect(wrapper.find('[data-testid="pixel-cafe-my-limit"]').text()).toContain('我的限额')
    expect(wrapper.find('[data-testid="pixel-cafe-my-limit"]').text()).toContain('18.75 / 80.00')
    expect(wrapper.find('[data-testid="pixel-cafe-my-limit"] [role="progressbar"]').attributes('aria-valuenow')).toBe('18.75')
    expect(wrapper.find('[data-testid="pixel-cafe-my-limit"] [role="progressbar"] > span').attributes('style')).toContain('width: 23.4375%')
    expect(wrapper.find('[data-testid="pixel-cafe-my-rooms-list"]').text()).not.toContain('C-018')
    expect(wrapper.text()).not.toContain('我的份额')
    expect(wrapper.text()).not.toContain('o***r@example.com')
    expect(wrapper.text()).not.toContain('5H 限额')
    expect(wrapper.text()).not.toContain('7D 限额')
    expect(wrapper.text()).not.toContain('下次刷新')
    expect(wrapper.text()).not.toContain('总额度')
    expect(wrapper.text()).not.toContain('使用中')
    expect(wrapper.text()).not.toContain('sk-cafe-my-rooms-private')
    await wrapper.findAll('.pixel-cafe-my-rooms-tab')[1].trigger('click')
    await flushPromises()
    expect(listMyRooms).toHaveBeenLastCalledWith({ page: 1, page_size: 20, status: 'history' })
  })

  it('shows safe empty states and unlimited windows for incomplete key projections', async () => {
    listMyRooms.mockResolvedValueOnce({
      data: {
        items: [{
          ...myRoom,
          account: myRoom.account,
          managed_api_key: {
            ...myRoom.managed_api_key,
            quota: 0,
            quota_used: 0,
            rate_limit_5h: 0,
            rate_limit_7d: 0,
            usage_5h: 0,
            usage_7d: 0,
          },
        }],
      },
    })
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('绑定账号：ChatGPT Plus 主账号')
    expect(wrapper.find('[data-testid="pixel-cafe-my-limit"]').text()).toContain('0.00 / 不限')
    expect(wrapper.find('[data-testid="pixel-cafe-limit-5h"]').exists()).toBe(false)

    listMyRooms.mockResolvedValueOnce({ data: { items: [{ ...myRoom, account: undefined, managed_api_key: null }] } })
    await wrapper.findAll('.pixel-cafe-my-rooms-tab')[1].trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('绑定账号：账号信息暂不可用')
    expect(wrapper.text()).toContain('我的限额：暂不可用')
  })

  it('hides account usage bars until the room is assigned and active', async () => {
    listMyRooms.mockResolvedValueOnce({
      data: {
        items: [{
          ...myRoom,
          status: 'paid',
          account: null,
          managed_api_key: null,
          round: { ...myRoom.round, status: 'awaiting_account' },
        }],
      },
    })
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.find('[data-testid="pixel-cafe-my-room-usage"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('绑定账号：已成团，预计 24 小时内开通')
    expect(wrapper.text()).toContain('剩余时间：开通后开始计算')
    expect(wrapper.text()).toContain('我的限额：开通后显示')
  })

  it('clears the page-owned countdown timer on unmount', async () => {
    const setIntervalSpy = vi.spyOn(window, 'setInterval')
    const clearIntervalSpy = vi.spyOn(window, 'clearInterval')
    const wrapper = mountPage()
    await flushPromises()

    const timerCallIndex = setIntervalSpy.mock.calls.findIndex(([, delay]) => delay === 30_000)
    expect(timerCallIndex).toBeGreaterThanOrEqual(0)
    const timerID = setIntervalSpy.mock.results[timerCallIndex]?.value
    wrapper.unmount()
    expect(clearIntervalSpy).toHaveBeenCalledWith(timerID)
    setIntervalSpy.mockRestore()
    clearIntervalSpy.mockRestore()
  })

  it('shows a retryable my-room error without replacing room discovery', async () => {
    listMyRooms.mockRejectedValueOnce(new Error('my rooms unavailable'))
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.find('[data-testid="pixel-cafe-my-rooms-error"]').text()).toContain('my rooms unavailable')
    expect(wrapper.find('[data-testid="pixel-cafe-room-list"]').exists()).toBe(true)
    await wrapper.find('.pixel-cafe-my-rooms-retry').trigger('click')
    await flushPromises()
    expect(listMyRooms).toHaveBeenCalledTimes(2)
  })

  it('renders an error and retries overview loading', async () => {
    overview.mockRejectedValueOnce(new Error('Cafe unavailable'))
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.find('[data-testid="pixel-cafe-error"]').text()).toContain('Cafe unavailable')
    await wrapper.find('.pixel-cafe-retry').trigger('click')
    await flushPromises()
    expect(overview).toHaveBeenCalledTimes(2)
  })

  it('renders an empty state when the overview has no rooms', async () => {
    overview.mockResolvedValueOnce(overviewPayload([]))
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.find('[data-testid="pixel-cafe-empty"]').exists()).toBe(true)
    expect(wrapper.find('.pixel-cafe-scene-art').exists()).toBe(true)
  })

  it('submits a multi-share purchase with agreement and opens payment waiting state', async () => {
    const wrapper = mountPage()
    await flushPromises()
    await wrapper.find('.pixel-cafe-room-card').trigger('click')

    const submit = wrapper.find('.pixel-cafe-primary')
    expect(submit.attributes('disabled')).toBeDefined()
    await wrapper.findAll('.pixel-cafe-seat-button')[1].trigger('click')
    expect(wrapper.find('[data-testid="pixel-cafe-share-count"]').text()).toContain('2 份')
    await wrapper.find('.pixel-cafe-agreement input').setValue(true)
    await wrapper.vm.$nextTick()
    await wrapper.find('.pixel-cafe-primary').trigger('click')
    await flushPromises()

    expect(createOrder).toHaveBeenCalledWith(room.id, expect.objectContaining({
      share_count: 2,
      payment_type: 'alipay',
      agreement_accepted: true,
    }), expect.any(String))
    expect(wrapper.find('[data-testid="payment-status-panel"]').exists()).toBe(true)
  })

  it('only shows payment methods enabled by the shared Sub2API checkout config', async () => {
    const wrapper = mountPage()
    await flushPromises()
    await wrapper.find('.pixel-cafe-room-card').trigger('click')

    const options = wrapper.findAll('.pixel-cafe-payment-select option').map(option => option.text())
    expect(options).toEqual(['支付宝', 'Stripe'])
    expect(options).not.toContain('微信支付')
    expect(options).not.toContain('Airwallex')
  })

  it('supports a one-share room through the same share purchase contract', async () => {
    const singleRoom = {
      ...room,
      id: 19,
      code: 'C-019',
      name: 'Plus 单份包间 19',
      plan: { ...room.plan, total_shares: 1, max_buyers: 1, max_shares_per_user: 1 },
      round: { ...room.round, id: 1009, paid_shares: 0, remaining_shares: 1, max_buyers: 1, joined_buyers: 0, remaining_buyer_slots: 1 },
      member_avatars: [],
    }
    overview.mockResolvedValueOnce(overviewPayload([singleRoom]))
    const wrapper = mountPage()
    await flushPromises()
    await wrapper.find('.pixel-cafe-room-card').trigger('click')

    expect(wrapper.text()).toContain('已售 0/1 份')
    expect(wrapper.findAll('.pixel-cafe-seat-button')).toHaveLength(2)
    expect(wrapper.findAll('.pixel-cafe-seat-button').every(button => (button.element as HTMLButtonElement).disabled)).toBe(true)
    expect(wrapper.find('.pixel-cafe-single-room-note').text()).toContain('每份 99 CNY')

    await wrapper.find('.pixel-cafe-agreement input').setValue(true)
    await wrapper.find('.pixel-cafe-primary').trigger('click')
    await flushPromises()
    expect(createOrder).toHaveBeenCalledWith(singleRoom.id, expect.objectContaining({ share_count: 1 }), expect.any(String))
  })

  it('keeps the selector open and shows the order failure', async () => {
    createOrder.mockRejectedValueOnce(new Error('seat unavailable'))
    const wrapper = mountPage()
    await flushPromises()
    await wrapper.find('.pixel-cafe-room-card').trigger('click')
    await wrapper.find('.pixel-cafe-seat-button').trigger('click')
    await wrapper.find('.pixel-cafe-agreement input').setValue(true)
    await wrapper.vm.$nextTick()
    await wrapper.find('.pixel-cafe-primary').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="pixel-cafe-order-error"]').text()).toContain('seat unavailable')
  })
})
