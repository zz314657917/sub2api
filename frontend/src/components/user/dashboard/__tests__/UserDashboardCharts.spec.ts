import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import UserDashboardCharts from '../UserDashboardCharts.vue'
import type { TrendDataPoint } from '@/types'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('vue-chartjs', () => ({
  Bar: { name: 'Bar', props: ['data', 'options'], template: '<div data-testid="bar" />' }
}))
const props = {
  loading: false,
  startDate: '2026-10-24',
  endDate: '2026-10-30',
  granularity: 'day',
  trend: [] as TrendDataPoint[]
}
const global = { stubs: { Icon: true, RouterLink: true, LoadingSpinner: true } }
function spendPoint(overrides: Partial<TrendDataPoint> = {}): TrendDataPoint {
  return {
    date: '2026-10-30', requests: 2, input_tokens: 9000, output_tokens: 100,
    cache_read_tokens: 10, cache_creation_tokens: 20, total_tokens: 9130,
    cost: 999, actual_cost: 10,
    actual_cost_breakdown: { input: 1, output: 6, cache_read: 1, cache_creation: 1, other: 1 },
    ...overrides
  }
}
describe('UserDashboardCharts', () => {
  it('uses actual spend shares, retains token counts and includes Other in the denominator', () => {
    const wrapper = mount(UserDashboardCharts, { props: { ...props, trend: [spendPoint()] }, global })
    const input = wrapper.get('[data-category="input"]')
    const output = wrapper.get('[data-category="output"]')
    expect(input.text()).toContain(`${new Intl.NumberFormat(undefined, { notation: 'compact', maximumFractionDigits: 1 }).format(9000)} Token`)
    expect(input.get('.dashboard-token-spend').text()).toContain('✪ 1.00')
    expect(input.get('.dashboard-token-spend').text()).toContain('10%')
    expect(input.get('i').attributes('style')).toContain('width: 10%')
    expect(output.get('i').attributes('style')).toContain('width: 60%')
    expect(wrapper.get('[data-category="other"]').text()).toContain('10%')
    expect(wrapper.text()).toContain('dashboard.console.spendAllocation')
    wrapper.unmount()
  })
  it.each([
    undefined,
    { input: 1 },
    { input: 1, output: 6, cache_read: 1, cache_creation: 1, other: NaN },
    { input: -1, output: 8, cache_read: 1, cache_creation: 1, other: 1 },
    { input: 1, output: 6, cache_read: 1, cache_creation: 1, other: 0 },
    { input: 1, output: 6, cache_read: 1, cache_creation: 1, other: Infinity }
  ])('shows unavailable for missing, partial or invalid spend metadata (%j)', (breakdown) => {
    const point = spendPoint({ actual_cost_breakdown: breakdown as TrendDataPoint['actual_cost_breakdown'] })
    const wrapper = mount(UserDashboardCharts, { props: { ...props, trend: [point] }, global })
    expect(wrapper.findAll('.dashboard-token-spend').every((item) => item.text() === 'dashboard.console.spendUnavailable')).toBe(true)
    expect(wrapper.findAll('.dashboard-token-track i').every((item) => item.attributes('style')?.includes('width: 0%'))).toBe(true)
    wrapper.unmount()
  })
  it('does not normalize a partially covered range and refreshes shares when the range data changes', async () => {
    const wrapper = mount(UserDashboardCharts, { props: { ...props, trend: [spendPoint(), spendPoint({ actual_cost_breakdown: undefined })] }, global })
    expect(wrapper.get('.dashboard-token-spend').text()).toBe('dashboard.console.spendUnavailable')
    await wrapper.setProps({ startDate: '2026-10-01', trend: [spendPoint({ actual_cost: 20, actual_cost_breakdown: { input: 10, output: 10, cache_read: 0, cache_creation: 0, other: 0 } })] })
    expect(wrapper.get('[data-category="input"] i').attributes('style')).toContain('width: 50%')
    expect(wrapper.find('[data-category="other"]').exists()).toBe(false)
    expect(wrapper.get('[data-category="input"] .dashboard-token-spend').text()).toContain('✪ 10.00')
    wrapper.unmount()
  })
  it('shows valid zero spend as 0% and tiny nonzero spend honestly', async () => {
    const wrapper = mount(UserDashboardCharts, { props: { ...props, trend: [spendPoint({ actual_cost: 0, actual_cost_breakdown: { input: 0, output: 0, cache_read: 0, cache_creation: 0, other: 0 } })] }, global })
    expect(wrapper.get('.dashboard-token-spend').text()).toContain('0%')
    expect(wrapper.get('.dashboard-token-spend').text()).toContain('✪ 0.00')
    expect(wrapper.text()).toContain('dashboard.console.zeroCost')
    await wrapper.setProps({ trend: [spendPoint({ actual_cost: 1, actual_cost_breakdown: { input: 0.00001, output: 0.99999, cache_read: 0, cache_creation: 0, other: 0 } })] })
    expect(wrapper.get('[data-category="input"] .dashboard-token-spend').text()).toContain('<0.1%')
    expect(wrapper.get('[data-category="input"] i').attributes('style')).toContain('width: 0.001%')
    wrapper.unmount()
  })
  it('accepts floating point residue without adjusting the actual denominator', () => {
    const wrapper = mount(UserDashboardCharts, { props: { ...props, trend: [spendPoint({ actual_cost: .3, actual_cost_breakdown: { input: .1, output: .2, cache_read: 0, cache_creation: 0, other: 0 } })] }, global })
    expect(wrapper.get('[data-category="input"] .dashboard-token-spend').text()).toContain('33.3%')
    wrapper.unmount()
  })
  it('distinguishes first-use, empty range and load failure with retry', async () => {
    const wrapper = mount(UserDashboardCharts, { props, global })
    expect(wrapper.text()).toContain('dashboard.console.rangeEmptyTitle')
    await wrapper.setProps({ firstUse: true })
    expect(wrapper.text()).toContain('dashboard.console.firstUseTitle')
    await wrapper.setProps({ error: true })
    expect(wrapper.text()).toContain('dashboard.console.loadError')
    expect(wrapper.text()).not.toContain('dashboard.console.firstUseTitle')
    expect(wrapper.text()).not.toContain('✪ 0')
    await wrapper.get('.dashboard-chart-state button').trigger('click')
    expect(wrapper.emitted('retry')).toHaveLength(1)
  })
  it('retains zero-cost requests and never substitutes standard spend for actual', () => {
    const wrapper = mount(UserDashboardCharts, {
      props: {
        ...props,
        trend: [
          {
            date: '2026-10-30',
            requests: 4,
            cost: 5,
            actual_cost: 0,
            input_tokens: 0,
            output_tokens: 0,
            cache_creation_tokens: 0,
            cache_read_tokens: 0,
            total_tokens: 0
          }
        ]
      },
      global
    })
    expect(wrapper.get('[data-testid="bar"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('dashboard.console.zeroCost')
    expect(wrapper.get('.dashboard-chart-total strong').text()).toBe('✪ 0.00')
    expect(wrapper.findComponent({ name: 'Bar' }).props('data').datasets[0].data).toEqual([0])
    expect(wrapper.text()).toContain('✪ 5.00')
  })
  it('updates a single inclusive date range for all consumers', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 9, 30, 23, 30))
    const wrapper = mount(UserDashboardCharts, { props, global })
    await wrapper.findAll('.dashboard-range-tab')[1].trigger('click')
    expect(wrapper.emitted('dateRangeChange')?.[0]).toEqual([
      { startDate: '2026-10-01', endDate: '2026-10-30', preset: '30d' }
    ])
    wrapper.unmount()
    vi.useRealTimers()
  })
  it('calculates cache reuse from aggregate input counts instead of daily averages or total tokens', async () => {
    const wrapper = mount(UserDashboardCharts, {
      props: {
        ...props,
        trend: [
          { date: '2026-10-29', requests: 1, input_tokens: 10, cache_read_tokens: 90, cache_creation_tokens: 9000, output_tokens: 5000, total_tokens: 14100 },
          { date: '2026-10-30', requests: 1, input_tokens: 890, cache_read_tokens: 10, cache_creation_tokens: 1000, output_tokens: 5000, total_tokens: 6900 }
        ] as TrendDataPoint[]
      },
      global
    })
    const summary = wrapper.get('.dashboard-range-summary')
    expect(summary.text()).toContain('dashboard.console.cacheRate')
    expect(summary.get('strong').text()).toBe('10.0%')
    expect(summary.text()).not.toContain('dashboard.requests')
    await wrapper.setProps({
      trend: [{ date: '2026-10-30', requests: 1, input_tokens: 0, cache_read_tokens: 100 } as TrendDataPoint]
    })
    expect(summary.get('strong').text()).toBe('100.0%')
    wrapper.unmount()
  })
  it.each([
    { input_tokens: 100, cache_read_tokens: 0, expected: '0.0%' },
    { input_tokens: 0, cache_read_tokens: 0, expected: '—' }
  ])('distinguishes no cached input from unavailable input ($expected)', ({ expected, ...tokens }) => {
    const wrapper = mount(UserDashboardCharts, {
      props: {
        ...props,
        trend: [{ date: '2026-10-30', requests: 1, ...tokens } as TrendDataPoint]
      },
      global
    })
    expect(wrapper.get('.dashboard-range-summary strong').text()).toBe(expected)
    wrapper.unmount()
  })
  it('refreshes the palette on a live theme toggle', async () => {
    const wrapper = mount(UserDashboardCharts, {
      props: {
        ...props,
        trend: [{ requests: 1, date: '2026-10-30', actual_cost: 1 } as TrendDataPoint]
      },
      global
    })
    const bar = wrapper.findComponent({ name: 'Bar' })
    const before = bar.props('data').datasets[0].backgroundColor
    document.documentElement.classList.add('dark')
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(bar.props('data').datasets[0].backgroundColor).not.toBe(before)
    document.documentElement.classList.remove('dark')
    wrapper.unmount()
  })
})
