import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import PixelCafePage from '../PixelCafePage.vue'

const overview = vi.hoisted(() => vi.fn())
const lobbyActivity = vi.hoisted(() => vi.fn())
const listRooms = vi.hoisted(() => vi.fn())
const listMyRooms = vi.hoisted(() => vi.fn())
const createOrder = vi.hoisted(() => vi.fn())
const routeQuery = vi.hoisted(() => ({} as Record<string, string>))
const cachedPublicSettings = vi.hoisted(() => ({
  pixel_cafe_title: '像素网吧',
  pixel_cafe_description: '把每个模型分组变成一间可订阅的数字包间。',
  pixel_cafe_header_visible: true,
}))

vi.mock('@/api/cafe', () => ({ cafeAPI: { overview, lobbyActivity, listRooms, listMyRooms, createOrder } }))
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
  name: 'Claude 包间 18',
  zone_key: 'claude',
  theme_key: 'warm_wood',
  scene_slot_key: 'claude-18',
  featured: true,
  plan: { id: 3, title: 'Claude Max', description: '', price_per_seat: 99, price_label: '99 CNY', validity_days: 30, total_seats: 5 },
  round: { id: 1008, status: 'open', paid_seats: 4, reserved_seats: 0, remaining_seats: 1, deadline_at: '2026-08-03T12:00:00Z' },
  seat_visuals: [
    { seat_no: 1, state: 'empty', is_mine: false },
    { seat_no: 2, state: 'paid', is_mine: false },
  ],
  purchase_state: 'available',
}

const myRoom = {
  membership_id: 892,
  room: { id: 18, code: 'C-018', name: 'Claude 包间 18', zone_key: 'claude', theme_key: 'warm_wood' },
  plan: { id: 3, title: 'Claude Max', validity_days: 30 },
  round: { id: 1008, status: 'active', paid_seats: 5, total_seats: 5 },
  seat: { id: 892, seat_no: 2, status: 'active', activated_at: '2026-08-03T01:00:00Z', expires_at: '2026-09-02T01:00:00Z' },
  account: { name: 'Claude Pro 主账号', platform: 'anthropic', email_masked: 'o***r@example.com' },
  managed_api_key: { id: 3011, name: 'Claude 包间 C-018 / 座位 2', status: 'disabled', quota: 100, quota_used: 12.3, rate_limit_5h: 10, rate_limit_1d: 20, rate_limit_7d: 80, usage_5h: 2.5, usage_7d: 18.75, protected: true as const },
}

function overviewPayload(rooms = [room]) {
  return {
    data: {
      api_version: 'cafe.v1',
      server_time: '2026-08-03T12:00:00Z',
      zones: [
        { key: 'featured', name: '精选大厅', room_count: 1, open_seat_count: 1 },
        { key: 'claude', name: 'Claude 区', room_count: 1, open_seat_count: 1 },
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
        seat_no: 1,
      },
    })
  })

  it('renders the real overview room data and selected-room status', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(overview).toHaveBeenCalledWith({ room_limit: 8 })
    expect(wrapper.find('[data-testid="pixel-cafe-room-list"]').text()).toContain('Claude 包间 18')
    expect(wrapper.text()).toContain('1/5 空位')
    await wrapper.find('.pixel-cafe-room-card').trigger('click')
    expect(wrapper.text()).toContain('周期')
    expect(wrapper.text()).toContain('30 天')
    expect(wrapper.find('[data-testid="pixel-cafe-lobby-activity"]').exists()).toBe(false)
    expect(lobbyActivity).not.toHaveBeenCalled()
    expect(wrapper.text()).not.toContain('abcdef1234567890')
  })

  it('shows five local-only demo rooms with waiting groups without calling the room API', async () => {
    routeQuery.demo = '1'
    const wrapper = mountPage()
    await flushPromises()

    expect(overview).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="pixel-cafe-demo-badge"]').text()).toContain('本地演示数据')
    expect(wrapper.find('[data-testid="pixel-cafe-room-list"]').findAll('.pixel-cafe-room-card')).toHaveLength(5)
    expect(wrapper.findAll('[data-testid="pixel-cafe-lobby-avatar"]')).toHaveLength(0)
    expect(wrapper.text()).toContain('Claude 深夜包间')
    expect(wrapper.text()).toContain('可加入')
    expect(wrapper.find('.pixel-cafe-scene-art').exists()).toBe(true)
    expect(wrapper.find('.pixel-cafe-front-desk').text()).toContain('共享房间')
    expect(wrapper.find('.pixel-cafe-workbench').exists()).toBe(true)

    await wrapper.find('.pixel-cafe-room-card').trigger('click')
    expect(wrapper.text()).toContain('本地演示不创建订单')
    expect((wrapper.find('.pixel-cafe-primary').element as HTMLButtonElement).disabled).toBe(true)
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
    expect(wrapper.find('.pixel-cafe-subtitle').text()).toBe('按模型选择独立房间。')
    wrapper.unmount()

    cachedPublicSettings.pixel_cafe_header_visible = false
    const hiddenWrapper = mountPage()
    await flushPromises()
    expect(hiddenWrapper.find('.pixel-cafe-header > div').exists()).toBe(false)
    expect(hiddenWrapper.find('h1').exists()).toBe(false)
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
    expect(wrapper.find('[data-testid="pixel-cafe-my-rooms-list"]').text()).toContain('Claude 包间 C-018 / 座位 2')
    expect(wrapper.text()).toContain('绑定账号：Claude Pro 主账号')
    expect(wrapper.text()).toContain('o***r@example.com')
    expect(wrapper.text()).toContain('5H 2.50 / 10.00')
    expect(wrapper.text()).toContain('7D 18.75 / 80.00')
    expect(wrapper.text()).toContain('使用中')
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
          account: null,
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

    expect(wrapper.text()).toContain('未绑定账号')
    expect(wrapper.text()).toContain('5H 0.00 / 不限')
    expect(wrapper.text()).toContain('7D 0.00 / 不限')

    listMyRooms.mockResolvedValueOnce({ data: { items: [{ ...myRoom, account: undefined, managed_api_key: undefined }] } })
    await wrapper.findAll('.pixel-cafe-my-rooms-tab')[1].trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('未绑定账号')
    expect(wrapper.text()).toContain('暂无受管 Key')
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

  it('loads a selected zone from the room API', async () => {
    const wrapper = mountPage()
    await flushPromises()
    await wrapper.findAll('.pixel-cafe-zone')[1].trigger('click')
    await flushPromises()

    expect(listRooms).toHaveBeenCalledWith({ page: 1, page_size: 24, zone: 'claude' })
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

  it('renders an empty state when the selected zone has no rooms', async () => {
    listRooms.mockResolvedValueOnce({ data: { items: [], total: 0, page: 1, page_size: 24, pages: 0 } })
    const wrapper = mountPage()
    await flushPromises()
    await wrapper.findAll('.pixel-cafe-zone')[1].trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="pixel-cafe-empty"]').exists()).toBe(true)
    expect(wrapper.find('.pixel-cafe-scene-art').exists()).toBe(true)
  })

  it('submits only a selected empty seat with agreement and opens payment waiting state', async () => {
    const wrapper = mountPage()
    await flushPromises()
    await wrapper.find('.pixel-cafe-room-card').trigger('click')

    const submit = wrapper.find('.pixel-cafe-primary')
    expect(submit.attributes('disabled')).toBeDefined()
    await wrapper.find('.pixel-cafe-seat-button').trigger('click')
    await wrapper.find('.pixel-cafe-agreement input').setValue(true)
    await wrapper.vm.$nextTick()
    await wrapper.find('.pixel-cafe-primary').trigger('click')
    await flushPromises()

    expect(createOrder).toHaveBeenCalledWith(room.id, expect.objectContaining({
      seat_no: 1,
      payment_type: 'alipay',
      agreement_accepted: true,
    }), expect.any(String))
    expect(wrapper.find('[data-testid="payment-status-panel"]').exists()).toBe(true)
  })

  it('treats a one-seat room as an exclusive room without rendering a seat picker', async () => {
    const singleRoom = {
      ...room,
      id: 19,
      code: 'C-019',
      name: 'Claude 独享包间 19',
      plan: { ...room.plan, total_seats: 1 },
      round: { ...room.round, id: 1009, paid_seats: 0, remaining_seats: 1, total_seats: 1 },
      seat_visuals: [{ seat_no: 1, state: 'empty', is_mine: false }],
    }
    overview.mockResolvedValueOnce(overviewPayload([singleRoom]))
    const wrapper = mountPage()
    await flushPromises()
    await wrapper.find('.pixel-cafe-room-card').trigger('click')

    expect(wrapper.text()).toContain('独享房间')
    expect(wrapper.find('.pixel-cafe-seat-button').exists()).toBe(false)
    expect(wrapper.find('.pixel-cafe-single-room-note').text()).toContain('独享房间')
    expect(wrapper.find('.pixel-cafe-primary').text()).toContain('预订独享房间')

    await wrapper.find('.pixel-cafe-agreement input').setValue(true)
    await wrapper.find('.pixel-cafe-primary').trigger('click')
    await flushPromises()
    expect(createOrder).toHaveBeenCalledWith(singleRoom.id, expect.objectContaining({ seat_no: 1 }), expect.any(String))
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
