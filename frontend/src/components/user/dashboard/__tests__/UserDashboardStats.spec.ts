import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import UserDashboardStats from '../UserDashboardStats.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const messages: Record<string, string> = {
        'dashboard.todayCacheHitRate': 'Today Cache Hit Rate',
        'dashboard.totalCacheHitRate': 'Historical Cache Hit Rate',
        'dashboard.cacheReadTokens': 'Cache Read',
        'common.notAvailable': 'N/A',
      }
      return messages[key] ?? key
    },
  }),
}))

const makeStats = (overrides: Partial<UserStatsType> = {}): UserStatsType => ({
  total_api_keys: 1,
  active_api_keys: 1,
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
  it('renders today and historical cache hit rate cards', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        stats: makeStats({
          today_cache_read_tokens: 300,
          today_cache_creation_tokens: 100,
          total_cache_read_tokens: 900,
          total_cache_creation_tokens: 100,
        }),
        balance: 0,
        isSimple: true,
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('Today Cache Hit Rate')
    expect(wrapper.text()).toContain('Historical Cache Hit Rate')
    expect(wrapper.text()).toContain('75.0%')
    expect(wrapper.text()).toContain('90.0%')
    expect(wrapper.text()).toContain('Cache Read: 300')
    expect(wrapper.text()).toContain('Cache Read: 900')
  })

  it('renders N/A when there are no cache tokens', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        stats: makeStats(),
        balance: 0,
        isSimple: true,
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('N/A')
  })
})
