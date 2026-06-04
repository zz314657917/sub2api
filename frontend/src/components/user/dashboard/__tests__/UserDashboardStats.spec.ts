import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import UserDashboardStats from '../UserDashboardStats.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const messages: Record<string, string> = {
        'dashboard.overview.totalCost': 'Total Spend',
        'dashboard.overview.totalTokens': 'Total Tokens',
        'dashboard.overview.totalRequests': 'Total Requests',
        'dashboard.overview.failureRate': 'Failed Request Rate',
        'dashboard.overview.averageRequestCost': 'Avg Spend per Request',
        'dashboard.overview.trendStable': 'Stable trend',
        'dashboard.overview.performanceStable': 'Stable performance',
        'dashboard.overview.activeApiKeys': `${params?.count}/${params?.total} API keys active`,
        'dashboard.overview.recentConsumption': 'Today spend as a short-term reference',
        'dashboard.overview.userRetentionNote': 'Total requests are lifetime accumulated',
        'dashboard.overview.failureRateNote': 'Failure stats are not returned by the current API',
        'dashboard.overview.averageRequestCostNote': 'Calculated as total spend / total requests',
      }
      return messages[key] ?? key
    },
  }),
}))

const makeStats = (overrides: Partial<UserStatsType> = {}): UserStatsType => ({
  total_api_keys: 3,
  active_api_keys: 2,
  total_requests: 0,
  total_input_tokens: 0,
  total_output_tokens: 0,
  total_cache_creation_tokens: 0,
  total_cache_read_tokens: 0,
  total_tokens: 0,
  total_cost: 0,
  total_actual_cost: 0,
  today_requests: 0,
  today_input_tokens: 0,
  today_output_tokens: 0,
  today_cache_creation_tokens: 0,
  today_cache_read_tokens: 0,
  today_tokens: 0,
  today_cost: 0,
  today_actual_cost: 0,
  average_duration_ms: 0,
  rpm: 0,
  tpm: 0,
  ...overrides,
})

describe('UserDashboardStats', () => {
  it('renders the five overview metric cards', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        stats: makeStats({
          total_requests: 100,
          total_tokens: 2500000,
          total_actual_cost: 12.5,
          today_requests: 8,
          today_tokens: 1800,
          today_cost: 0.48,
          today_actual_cost: 0.42,
        }),
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    expect(wrapper.findAll('.dashboard-stat-card')).toHaveLength(5)
    expect(wrapper.text()).toContain('Total Spend')
    expect(wrapper.text()).toContain('Total Tokens')
    expect(wrapper.text()).toContain('Total Requests')
    expect(wrapper.text()).toContain('Failed Request Rate')
    expect(wrapper.text()).toContain('Avg Spend per Request')
  })

  it('formats derived values and neutral notes', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        stats: makeStats({
          total_requests: 100,
          total_tokens: 2500000,
          total_actual_cost: 12.5,
          today_cost: 0.48,
          today_actual_cost: 0.42,
        }),
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    const expectedCompactTokens = new Intl.NumberFormat(undefined, {
      notation: 'compact',
      maximumFractionDigits: 1,
    }).format(2500000)

    expect(wrapper.text()).toContain('$12.50')
    expect(wrapper.text()).toContain(expectedCompactTokens)
    expect(wrapper.text()).toContain('100')
    expect(wrapper.text()).toContain('0%')
    expect(wrapper.text()).toContain('$0.125')
    expect(wrapper.text()).toContain('2/3 API keys active')
    expect(wrapper.text()).toContain('Failure stats are not returned by the current API')
  })

  it('keeps zero-data cards stable', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        stats: makeStats({
          total_api_keys: 0,
          active_api_keys: 0,
        }),
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    expect(wrapper.findAll('.dashboard-stat-card')).toHaveLength(5)
    expect(wrapper.text()).toContain('$0.00')
    expect(wrapper.text()).toContain('0%')
    expect(wrapper.text()).toContain('$0.00')
    expect(wrapper.text()).toContain('0/0 API keys active')
  })
})
