import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import UserDashboardPerformanceStats from '../UserDashboardPerformanceStats.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const messages: Record<string, string> = {
        'dashboard.todayCacheHitRate': 'Today Input Cache Reuse',
        'dashboard.totalCacheHitRate': 'Historical Input Cache Reuse',
        'dashboard.cacheReadTokens': 'Cache Read',
        'dashboard.performance': 'Performance',
        'dashboard.avgResponse': 'Avg Response Time',
        'dashboard.averageTime': 'Average time',
        'common.notAvailable': 'N/A',
      }
      return messages[key] ?? key
    },
  }),
}))

const makeStats = (overrides: Partial<UserStatsType> = {}): UserStatsType => ({
  total_api_keys: 0,
  active_api_keys: 0,
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

describe('UserDashboardPerformanceStats', () => {
  it('renders cache reuse and average response metrics', () => {
    const wrapper = mount(UserDashboardPerformanceStats, {
      props: {
        stats: makeStats({
          today_cache_read_tokens: 300,
          today_input_tokens: 700,
          total_cache_read_tokens: 700,
          total_input_tokens: 300,
          average_duration_ms: 1500,
          rpm: 12,
          tpm: 3456,
        }),
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    expect(wrapper.findAll('.dashboard-performance-card')).toHaveLength(4)
    expect(wrapper.text()).toContain('Performance')
    expect(wrapper.text()).toContain('Today Input Cache Reuse')
    expect(wrapper.text()).toContain('Historical Input Cache Reuse')
    expect(wrapper.text()).toContain('Avg Response Time')
    expect(wrapper.text()).toContain('12')
    expect(wrapper.text()).toContain('RPM')
    expect(wrapper.text()).toContain('3,456')
    expect(wrapper.text()).toContain('TPM')
    expect(wrapper.text()).toContain('30.0%')
    expect(wrapper.text()).toContain('70.0%')
    expect(wrapper.text()).toContain('1.50s')
    expect(wrapper.text()).toContain('Cache Read: 300')
    expect(wrapper.text()).toContain('Cache Read: 700')
  })

  it('shows N/A for empty cache inputs and keeps millisecond response times', () => {
    const wrapper = mount(UserDashboardPerformanceStats, {
      props: {
        stats: makeStats({
          average_duration_ms: 250,
        }),
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('N/A')
    expect(wrapper.text()).toContain('250ms')
  })
})
