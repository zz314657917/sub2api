import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import UserDashboardStats from '../UserDashboardStats.vue'
import type { UserDashboardStats as Stats } from '@/api/usage'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))
const settings = vi.hoisted(() => ({
  payment_enabled: true as boolean | undefined,
  isSimpleMode: false
}))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ cachedPublicSettings: settings })
}))
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    get isSimpleMode() {
      return settings.isSimpleMode
    }
  })
}))

describe('UserDashboardStats', () => {
  beforeEach(() => {
    settings.payment_enabled = true
    settings.isSimpleMode = false
  })
  it('shows balance, actual server-day spend, today tokens and active keys without fabricated rates', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        balance: 8.25,
        stats: {
          today_actual_cost: 0,
          today_cost: 19,
          today_requests: 12,
          today_tokens: 17600,
          active_api_keys: 2,
          total_api_keys: 3
        } as Stats
      },
      global: { stubs: { Icon: true, RouterLink: true } }
    })
    expect(wrapper.findAll('.dashboard-stat-card')).toHaveLength(4)
    const cards = wrapper.findAll('.dashboard-stat-card')
    expect(cards[0].text()).toContain('✪ 8.25')
    expect(cards[1].text()).toContain('✪ 0.00')
    expect(cards[1].text()).not.toContain('19')
    expect(cards[1].text()).toContain('dashboard.console.serverDayActual')
    expect(cards[2].text()).toContain('dashboard.todayTokens')
    expect(cards[2].text()).toContain(new Intl.NumberFormat(undefined, {
      notation: 'compact', maximumFractionDigits: 1
    }).format(17600))
    expect(cards[2].text()).toContain('dashboard.console.serverDayTokens')
    expect(cards[2].text()).not.toContain('dashboard.todayRequests')
    expect(cards[3].text()).toContain('2')
    expect(wrapper.text()).not.toContain('%')
    expect(wrapper.text()).not.toContain('trendStable')
  })
  it.each([
    { enabled: true, simple: false, route: '/purchase', label: 'nav.buySubscription' },
    { enabled: false, simple: false, route: '/profile', label: 'nav.profile' },
    { enabled: true, simple: true, route: '/profile', label: 'nav.profile' },
    { enabled: false, simple: true, route: '/profile', label: 'nav.profile' },
    { enabled: undefined, simple: false, route: '/purchase', label: 'nav.buySubscription' }
  ])(
    'uses the registered payment/simple-mode semantics ($enabled/$simple)',
    ({ enabled, simple, route, label }) => {
      settings.payment_enabled = enabled
      settings.isSimpleMode = simple
      const wrapper = mount(UserDashboardStats, {
        props: {
          balance: 8,
          stats: {
            today_actual_cost: 0,
            today_requests: 0,
            today_tokens: 0,
            active_api_keys: 0,
            total_api_keys: 0
          } as Stats
        },
        global: {
          stubs: {
            Icon: true,
            RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' }
          }
        }
      })
      const balance = wrapper.get('.dashboard-stat-card--balance')
      expect(balance.get('a').attributes('href')).toBe(route)
      expect(balance.text()).toContain(label)
      wrapper.unmount()
    }
  )
})
