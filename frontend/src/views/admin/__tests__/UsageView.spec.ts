import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, ref } from 'vue'

import UsageView from '../UsageView.vue'

const { list, getStats, getSnapshotV2, getModelStats, getById, routeQuery } = vi.hoisted(() => {
  vi.stubGlobal('localStorage', {
    getItem: vi.fn(() => null),
    setItem: vi.fn(),
    removeItem: vi.fn(),
  })

  return {
    list: vi.fn(),
    getStats: vi.fn(),
    getSnapshotV2: vi.fn(),
    getModelStats: vi.fn(),
    getById: vi.fn(),
    routeQuery: {} as Record<string, string>,
  }
})

const messages: Record<string, string> = {
  'admin.dashboard.timeRange': 'Time Range',
  'admin.dashboard.day': 'Day',
  'admin.dashboard.hour': 'Hour',
  'admin.usage.failedToLoadUser': 'Failed to load user',
  'usage.latency': 'Latency Health',
}

const formatLocalDate = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

vi.mock('@/api/admin', () => ({
  adminAPI: {
    usage: {
      list,
      getStats,
    },
    dashboard: {
      getSnapshotV2,
      getModelStats,
    },
    users: {
      getById,
    },
  },
}))

vi.mock('@/api/admin/usage', () => ({
  adminUsageAPI: {
    list: vi.fn(),
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showWarning: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn(),
  }),
}))

vi.mock('@/utils/format', () => ({
  formatReasoningEffort: (value: string | null | undefined) => value ?? '-',
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

vi.mock('vue-router', () => ({
  useRoute: () => ({
    query: routeQuery
  })
}))

const AppLayoutStub = { template: '<div><slot /></div>' }
const UsageFiltersStub = defineComponent({
  setup(_, { expose }) {
    const userKeyword = ref('')
    let userSearchRevision = 0
    const setUserKeyword = (email: string) => {
      userKeyword.value = email
    }
    const simulateUserInput = (email: string) => {
      userSearchRevision += 1
      userKeyword.value = email
    }
    expose({
      getUserSearchRevision: () => userSearchRevision,
      setUserKeyword,
      simulateUserInput,
    })
    return { userKeyword }
  },
  template: '<div data-test="usage-filter-surface"><span data-test="user-filter-label">{{ userKeyword }}</span><slot name="after-reset" /></div>',
})
const UsageTableStub = {
  props: ['columns'],
  template: '<div data-test="usage-columns">{{ columns.map((column) => column.key).join(",") }}</div>',
}
const ModelDistributionChartStub = {
  props: ['metric'],
  emits: ['update:metric'],
  template: `
    <div data-test="model-chart">
      <span class="metric">{{ metric }}</span>
      <button class="switch-metric" @click="$emit('update:metric', 'actual_cost')">switch</button>
    </div>
  `,
}
const GroupDistributionChartStub = {
  props: ['metric'],
  emits: ['update:metric'],
  template: `
    <div data-test="group-chart">
      <span class="metric">{{ metric }}</span>
      <button class="switch-metric" @click="$emit('update:metric', 'actual_cost')">switch</button>
    </div>
  `,
}

const mountRouteFilteredUsageView = () => mount(UsageView, {
  global: {
    stubs: {
      AppLayout: AppLayoutStub,
      UsageStatsCards: true,
      UsageFilters: UsageFiltersStub,
      UsageTable: true,
      UsageExportProgress: true,
      UsageCleanupDialog: true,
      UserBalanceHistoryModal: true,
      Pagination: true,
      Select: true,
      DateRangePicker: true,
      Icon: true,
      TokenUsageTrend: true,
      ModelDistributionChart: true,
      GroupDistributionChart: true,
      EndpointDistributionChart: true,
    },
  },
})

describe('admin UsageView route filters', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    Object.keys(routeQuery).forEach((key) => delete routeQuery[key])
    list.mockReset().mockResolvedValue({ items: [], total: 0, pages: 0 })
    getStats.mockReset().mockResolvedValue({
      total_requests: 0,
      total_input_tokens: 0,
      total_output_tokens: 0,
      total_cache_tokens: 0,
      total_tokens: 0,
      total_cost: 0,
      total_actual_cost: 0,
      average_duration_ms: 0,
    })
    getSnapshotV2.mockReset().mockResolvedValue({ trend: [], models: [], groups: [] })
    getModelStats.mockReset().mockResolvedValue({ models: [] })
    getById.mockReset()
  })

  afterEach(() => {
    Object.keys(routeQuery).forEach((key) => delete routeQuery[key])
    vi.useRealTimers()
  })

  it('shows the routed user email while applying user_id to usage requests', async () => {
    routeQuery.user_id = '42'
    getById.mockResolvedValue({ id: 42, email: 'route-user@test.com' })

    const wrapper = mountRouteFilteredUsageView()
    await flushPromises()

    expect(getById).toHaveBeenCalledWith(42)
    expect(list).toHaveBeenCalledWith(expect.objectContaining({ user_id: 42 }), expect.anything())
    expect(wrapper.get('[data-test="user-filter-label"]').text()).toBe('route-user@test.com')
  })

  it('shows the routed user ID when its label lookup fails', async () => {
    routeQuery.user_id = '42'
    getById.mockRejectedValue(new Error('lookup failed'))

    const wrapper = mountRouteFilteredUsageView()
    await flushPromises()

    expect(wrapper.get('[data-test="user-filter-label"]').text()).toBe('42')
  })

  it('does not overwrite newer user input after the routed user lookup completes', async () => {
    routeQuery.user_id = '42'
    let resolveLookup!: (user: { id: number; email: string }) => void
    getById.mockReturnValue(new Promise((resolve) => { resolveLookup = resolve }))

    const wrapper = mountRouteFilteredUsageView()
    await wrapper.vm.$nextTick()
    ;(wrapper.findComponent(UsageFiltersStub).vm as any).simulateUserInput('new-search@test.com')

    resolveLookup({ id: 42, email: 'route-user@test.com' })
    await flushPromises()

    expect((wrapper.vm as any).filters.user_id).toBe(42)
    expect(wrapper.get('[data-test="user-filter-label"]').text()).toBe('new-search@test.com')
  })
})

describe('admin UsageView distribution metric toggles', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    list.mockReset()
    getStats.mockReset()
    getSnapshotV2.mockReset()
    getModelStats.mockReset()
    getById.mockReset()
    Object.keys(routeQuery).forEach((key) => delete routeQuery[key])

    list.mockResolvedValue({
      items: [],
      total: 0,
      pages: 0,
    })
    getStats.mockResolvedValue({
      total_requests: 0,
      total_input_tokens: 0,
      total_output_tokens: 0,
      total_cache_tokens: 0,
      total_tokens: 0,
      total_cost: 0,
      total_actual_cost: 0,
      average_duration_ms: 0,
    })
    getSnapshotV2.mockResolvedValue({
      trend: [],
      models: [],
      groups: [],
    })
    getModelStats.mockResolvedValue({ models: [] })
  })

  afterEach(() => {
    Object.keys(routeQuery).forEach((key) => delete routeQuery[key])
    vi.useRealTimers()
  })

  it('keeps model and group metric toggles independent without refetching chart data', async () => {
    const wrapper = mount(UsageView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          UsageStatsCards: true,
          UsageFilters: UsageFiltersStub,
          UsageTable: UsageTableStub,
          UsageExportProgress: true,
          UsageCleanupDialog: true,
          UserBalanceHistoryModal: true,
          Pagination: true,
          Select: true,
          DateRangePicker: true,
          Icon: true,
          TokenUsageTrend: true,
          ModelDistributionChart: ModelDistributionChartStub,
          GroupDistributionChart: GroupDistributionChartStub,
        },
      },
    })

    vi.advanceTimersByTime(120)
    await flushPromises()

    const usageColumns = wrapper.find('[data-test="usage-columns"]').text()
    expect(usageColumns).toContain('latency')
    expect(usageColumns).not.toContain('first_token')
    expect(usageColumns).not.toContain('duration')

    expect(getSnapshotV2).toHaveBeenCalledTimes(1)
    const now = new Date()
    const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000)
    expect(getSnapshotV2).toHaveBeenCalledWith(expect.objectContaining({
      start_date: formatLocalDate(yesterday),
      end_date: formatLocalDate(now),
      granularity: 'hour'
    }))

    const modelChart = wrapper.find('[data-test="model-chart"]')
    const groupChart = wrapper.find('[data-test="group-chart"]')

    expect(modelChart.find('.metric').text()).toBe('tokens')
    expect(groupChart.find('.metric').text()).toBe('tokens')

    await modelChart.find('.switch-metric').trigger('click')
    await flushPromises()

    expect(modelChart.find('.metric').text()).toBe('actual_cost')
    expect(groupChart.find('.metric').text()).toBe('tokens')
    expect(getSnapshotV2).toHaveBeenCalledTimes(1)

    await groupChart.find('.switch-metric').trigger('click')
    await flushPromises()

    expect(modelChart.find('.metric').text()).toBe('actual_cost')
    expect(groupChart.find('.metric').text()).toBe('actual_cost')
    expect(getSnapshotV2).toHaveBeenCalledTimes(1)
  })

  it('elevates the filter surface only while the column settings menu is open', async () => {
    const wrapper = mount(UsageView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          UsageStatsCards: true,
          UsageFilters: UsageFiltersStub,
          UsageTable: UsageTableStub,
          UsageExportProgress: true,
          UsageCleanupDialog: true,
          UserBalanceHistoryModal: true,
          Pagination: true,
          Select: true,
          DateRangePicker: true,
          Icon: true,
          TokenUsageTrend: true,
          ModelDistributionChart: ModelDistributionChartStub,
          GroupDistributionChart: GroupDistributionChartStub,
        },
      },
    })

    vi.advanceTimersByTime(120)
    await flushPromises()

    const filterSurface = wrapper.get('[data-test="usage-filter-surface"]')
    expect(filterSurface.classes()).not.toContain('z-[221]')

    await wrapper.get('[data-test="usage-column-settings"]').trigger('click')
    expect(filterSurface.classes()).toContain('z-[221]')

    await wrapper.get('[data-test="usage-column-settings"]').trigger('click')
    expect(filterSurface.classes()).not.toContain('z-[221]')
  })
})
