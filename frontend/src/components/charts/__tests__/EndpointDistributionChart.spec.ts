import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import EndpointDistributionChart from '../EndpointDistributionChart.vue'

const { getUserBreakdown } = vi.hoisted(() => ({
  getUserBreakdown: vi.fn(),
}))

vi.mock('@/api/admin/dashboard', () => ({ getUserBreakdown }))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'admin.dashboard.endpointDistribution': 'Endpoint Distribution',
    'admin.dashboard.endpoint': 'Endpoint',
    'admin.dashboard.requests': 'Requests',
    'admin.dashboard.tokens': 'Tokens',
    'admin.dashboard.actual': 'Actual',
    'admin.dashboard.standard': 'Standard',
    'admin.dashboard.metricTokens': 'By Tokens',
    'admin.dashboard.metricActualCost': 'By Actual Cost',
    'admin.dashboard.noDataAvailable': 'No data available',
  }
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => messages[key] ?? key }),
  }
})

vi.mock('vue-chartjs', () => ({
  Doughnut: {
    props: ['data'],
    template: '<div class="chart-data">{{ JSON.stringify(data) }}</div>',
  },
}))

describe('EndpointDistributionChart', () => {
  it('disables admin breakdown while preserving user-safe endpoint totals', async () => {
    getUserBreakdown.mockReset()
    const wrapper = mount(EndpointDistributionChart, {
      props: {
        endpointStats: [{
          endpoint: '/v1/responses',
          requests: 5,
          total_tokens: 1500,
          cost: 0.8,
          actual_cost: 0.3,
          account_cost: 0.4,
        }],
        enableBreakdown: false,
      },
      global: {
        stubs: {
          LoadingSpinner: true,
        },
      },
    })

    const row = wrapper.get('tbody tr')
    expect(row.text()).toContain('/v1/responses')
    expect(row.text()).toContain('$0.300')

    await row.trigger('click')
    expect(getUserBreakdown).not.toHaveBeenCalled()
  })
})
