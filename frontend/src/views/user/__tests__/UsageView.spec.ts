import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import UsageView from '../UsageView.vue'

const {
  query,
  getStatsByDateRange,
  list,
  getAvailableGroups,
  getMySubscriptions,
  routerPush,
  showError,
  showWarning,
  showSuccess,
  showInfo,
} = vi.hoisted(() => ({
  query: vi.fn(),
  getStatsByDateRange: vi.fn(),
  list: vi.fn(),
  getAvailableGroups: vi.fn(),
  getMySubscriptions: vi.fn(),
  routerPush: vi.fn(),
  showError: vi.fn(),
  showWarning: vi.fn(),
  showSuccess: vi.fn(),
  showInfo: vi.fn(),
}))

const messages: Record<string, string> = {
  'common.refresh': 'Refresh',
  'common.reset': 'Reset',
  'usage.costDetails': 'Cost Breakdown',
  'usage.routeInfo': 'Route Info',
  'usage.requestInfo': 'Request Info',
  'usage.details': 'Details',
  'usage.group': 'Billing Group',
  'usage.allGroups': 'All Groups',
  'usage.moreFilters': 'More Filters',
  'usage.columnSettings': 'Columns',
  'usage.modelFilter': 'Model Filter',
  'usage.modelFilterPlaceholder': 'Exact model name',
  'usage.allTypes': 'All Types',
  'usage.allBillingModes': 'All Billing Modes',
  'admin.usage.billingMode': 'Billing Mode',
  'admin.usage.inputCost': 'Input Cost',
  'admin.usage.outputCost': 'Output Cost',
  'admin.usage.cacheCreationCost': 'Cache Creation Cost',
  'admin.usage.cacheReadCost': 'Cache Read Cost',
  'admin.usage.inputTokens': 'Input Tokens',
  'admin.usage.outputTokens': 'Output Tokens',
  'admin.usage.cacheCreation5mTokens': 'Cache 5m Tokens',
  'admin.usage.cacheCreation1hTokens': 'Cache 1h Tokens',
  'admin.usage.cacheCreationTokens': 'Cache Creation Tokens',
  'admin.usage.cacheReadTokens': 'Cache Read Tokens',
  'usage.inputTokenPrice': 'Input price',
  'usage.outputTokenPrice': 'Output price',
  'usage.imageInputTokens': 'Image input tokens',
  'usage.imageInputTokenPrice': 'Image input price',
  'usage.imageInputCost': 'Image input cost',
  'usage.imageOutputTokens': 'Image output tokens',
  'usage.imageOutputTokenPrice': 'Image output price',
  'usage.imageOutputCost': 'Image output cost',
  'usage.perMillionTokens': '/ 1M tokens',
  'usage.serviceTier': 'Service tier',
  'usage.serviceTierPriority': 'Fast',
  'usage.serviceTierFlex': 'Flex',
  'usage.serviceTierStandard': 'Standard',
  'usage.rate': 'Rate',
  'usage.original': 'Original',
  'usage.billed': 'Billed',
  'usage.totalRequests': 'Total Requests',
  'usage.totalTokens': 'Total Tokens',
  'usage.totalCost': 'Total Cost',
  'usage.standardCost': 'Standard',
  'usage.officialReferenceCost': 'Official reference',
  'usage.actualCost': 'Actual',
  'usage.avgDuration': 'Avg Duration',
  'usage.inSelectedRange': 'Selected range',
  'usage.perRequest': 'Per request',
  'usage.allApiKeys': 'All API Keys',
  'usage.apiKeyFilter': 'API Key',
  'usage.timeRange': 'Time Range',
  'usage.exporting': 'Exporting...',
  'usage.exportCsv': 'Export CSV',
  'usage.preparingExport': 'Preparing export...',
  'usage.exportSuccess': 'Export success',
  'usage.exportFailed': 'Export failed',
  'usage.noDataToExport': 'No data',
  'usage.failedToLoad': 'Failed to load',
  'usage.noRecords': 'No records',
  'usage.model': 'Model',
  'usage.reasoningEffort': 'Reasoning Effort',
  'usage.endpoint': 'Endpoint',
  'usage.type': 'Type',
  'usage.tokens': 'Tokens',
  'usage.cost': 'Cost',
  'usage.firstToken': 'First Token',
  'usage.duration': 'Duration',
  'usage.latency': 'Latency Health',
  'usage.latencyFirstToken': 'First',
  'usage.latencyDuration': 'Total',
  'usage.time': 'Time',
  'usage.userAgent': 'User Agent',
  'usage.ws': 'WS',
  'usage.stream': 'Stream',
  'usage.sync': 'Sync',
  'usage.unknown': 'Unknown',
  'usage.in': 'In',
  'usage.out': 'Out',
  'usage.cacheRead': 'Cache Read',
  'usage.cacheWrite': 'Write',
  'usage.imageUnit': ' images',
  'usage.imageCount': 'Image count',
  'usage.imageBillingSize': 'Billing size',
  'usage.imageInputSize': 'Input size',
  'usage.imageOutputSize': 'Output size',
  'usage.imageSizeSource': 'Size source',
  'usage.imageSizeBreakdown': 'Size breakdown',
  'usage.imageSizeSourceOutput': 'Upstream output',
  'usage.imageSizeSourceInput': 'Request input',
  'usage.imageSizeSourceDefault': 'Default billing tier',
  'usage.imageSizeSourceLegacy': 'Legacy record',
  'usage.imageSizeSourceMissing': 'Not recorded',
  'usage.imageSizeNotRecorded': 'not recorded',
  'usage.imageSizeLegacyUnstandardized': 'legacy unstandardized',
  'usage.imageSizeUnknown': 'unknown',
  'usage.imageUnitPrice': 'Per-image price',
  'usage.imageTotalPrice': 'Image total price',
  'usage.unitPrice': 'Per-request price',
  'admin.usage.billingModeToken': 'Token',
  'admin.usage.billingModePerRequest': 'Per request',
  'admin.usage.billingModeImage': 'Image',
  'userSubscriptions.title': 'My Subscriptions',
  'userSubscriptions.description': 'View your subscription plans and usage',
  'userSubscriptions.noActiveSubscriptions': 'No Active Subscriptions',
  'userSubscriptions.noActiveSubscriptionsDesc':
    "You don't have any active subscriptions. Contact administrator to get one.",
  'userSubscriptions.failedToLoad': 'Failed to load subscriptions',
  'userSubscriptions.status.active': 'Active',
  'userSubscriptions.status.expired': 'Expired',
  'userSubscriptions.status.revoked': 'Revoked',
  'userSubscriptions.expires': 'Expires',
  'userSubscriptions.noExpiration': 'No expiration',
  'userSubscriptions.unlimited': 'Unlimited',
  'userSubscriptions.unlimitedDesc': 'No usage limits on this subscription',
  'userSubscriptions.daily': 'Daily',
  'userSubscriptions.weekly': 'Weekly',
  'userSubscriptions.monthly': 'Monthly',
  'userSubscriptions.daysRemaining': '{days} days remaining',
  'userSubscriptions.resetIn': 'Resets in {time}',
  'userSubscriptions.quotaEndsIn': 'Quota ends in {time}',
  'userSubscriptions.windowNotActive': 'Awaiting first use',
  'payment.planCard.rate': 'Rate',
  'payment.planCard.peakRate': 'Peak rate',
  'payment.renewNow': 'Renew Now',
  'common.today': 'today',
  'common.tomorrow': 'tomorrow',
  'common.serverTime': 'Server time',
}

vi.mock('@/api', () => ({
  usageAPI: {
    query,
    getStatsByDateRange,
  },
  keysAPI: {
    list,
  },
  userGroupsAPI: {
    getAvailable: getAvailableGroups,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showWarning, showSuccess, showInfo }),
}))

vi.mock('@/api/subscriptions', () => ({
  default: {
    getMySubscriptions,
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: routerPush,
  }),
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

const availableGroup = {
  id: 9,
  name: 'mc',
  description: null,
  platform: 'openai',
  rate_multiplier: 1,
  is_exclusive: false,
  status: 'active',
  subscription_type: 'standard',
  daily_limit_usd: null,
  weekly_limit_usd: null,
  monthly_limit_usd: null,
  allow_image_generation: true,
  image_rate_independent: false,
  image_rate_multiplier: 1,
  image_price_1k: null,
  image_price_2k: null,
  image_price_4k: null,
  claude_code_only: false,
  fallback_group_id: null,
  fallback_group_id_on_invalid_request: null,
  require_oauth_only: false,
  require_privacy_set: false,
}

const activeSubscription = {
  id: 12,
  user_id: 42,
  group_id: 9,
  status: 'active',
  starts_at: '2026-03-01T00:00:00Z',
  daily_usage_usd: 0.25,
  weekly_usage_usd: 1.5,
  monthly_usage_usd: 4.2,
  daily_window_start: '2026-03-08T00:00:00Z',
  weekly_window_start: '2026-03-01T00:00:00Z',
  monthly_window_start: '2026-03-01T00:00:00Z',
  created_at: '2026-03-01T00:00:00Z',
  updated_at: '2026-03-08T00:00:00Z',
  expires_at: null,
  group: {
    ...availableGroup,
    daily_limit_usd: 1,
    weekly_limit_usd: 5,
    monthly_limit_usd: 20,
  },
}

const baseUsageLog = (overrides: Record<string, unknown> = {}) => ({
  id: 1,
  user_id: 42,
  api_key_id: 3,
  account_id: null,
  request_id: 'req-user-1',
  model: 'gpt-5.5',
  reasoning_effort: 'XHigh',
  inbound_endpoint: '/responses',
  upstream_endpoint: '/v1/responses',
  group_id: 9,
  group: availableGroup,
  subscription_id: null,
  actual_cost: 0.092883,
  total_cost: 0.092883,
  rate_multiplier: 1.25,
  service_tier: 'priority',
  input_cost: 0.020285,
  output_cost: 0.00303,
  cache_creation_cost: 0,
  cache_read_cost: 0.069568,
  input_tokens: 4057,
  output_tokens: 101,
  cache_creation_tokens: 0,
  cache_read_tokens: 278272,
  cache_creation_5m_tokens: 0,
  cache_creation_1h_tokens: 0,
  billing_type: 0,
  billing_mode: 'token',
  request_type: 'stream',
  stream: true,
  openai_ws_mode: false,
  image_count: 0,
  image_size: null,
  image_input_size: null,
  image_output_size: null,
  image_size_source: null,
  image_size_breakdown: null,
  image_input_tokens: 0,
  image_input_cost: 0,
  image_output_tokens: 0,
  image_output_cost: 0,
  cache_ttl_overridden: false,
  first_token_ms: 12,
  duration_ms: 345,
  created_at: '2026-03-08T00:00:00Z',
  user_agent: 'Codex Desktop/0.133.0-alpha.1 (Windows 10.0; x86_64)',
  api_key: { id: 3, name: 'demo-key' },
  ...overrides,
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const TablePageLayoutStub = {
  template: '<div><slot name="actions" /><slot name="filters" /><slot name="table" /><slot /></div>',
}
const DataTableStub = {
  name: 'DataTable',
  props: ['data', 'columns'],
  template: `
    <div>
      <div class="table-headers">
        <span v-for="column in columns" :key="column.key" class="table-header" :data-column="column.key">
          {{ column.label }}
        </span>
      </div>
      <div v-for="row in data" :key="row.request_id" class="table-row">
        <div v-for="column in columns" :key="column.key" class="table-cell" :data-column="column.key">
          <slot :name="'cell-' + column.key" :row="row" :value="row[column.key]" :expanded="false">
            {{ row[column.key] }}
          </slot>
        </div>
      </div>
      <slot v-if="!data || data.length === 0" name="empty" />
    </div>
  `,
}
const SelectStub = {
  name: 'Select',
  props: ['modelValue', 'options', 'placeholder'],
  emits: ['update:modelValue', 'change'],
  methods: {
    onChange(event: Event) {
      const raw = (event.target as HTMLSelectElement).value
      const option = (this.options || []).find((item: any) => String(item.value ?? '') === raw) || null
      const value = option?.value ?? null
      this.$emit('update:modelValue', value)
      this.$emit('change', value, option)
    },
  },
  template: `
    <select class="select-stub" :data-placeholder="placeholder" :value="modelValue ?? ''" @change="onChange">
      <option v-for="option in options" :key="String(option.value ?? '')" :value="option.value ?? ''">
        {{ option.label }}
      </option>
    </select>
  `,
}
const IconStub = { template: '<span class="icon-stub" />' }

const mountUsageView = async (items = [baseUsageLog()]) => {
  query.mockResolvedValue({
    items,
    total: items.length,
    pages: items.length > 0 ? 1 : 0,
  })
  getStatsByDateRange.mockResolvedValue({
    total_requests: items.length,
    total_tokens: 100,
    total_input_tokens: 60,
    total_output_tokens: 40,
    total_cost: 0.1,
    total_actual_cost: 0.092883,
    average_duration_ms: 345,
  })
  list.mockResolvedValue({ items: [{ id: 3, name: 'demo-key' }] })
  getAvailableGroups.mockResolvedValue([availableGroup])
  getMySubscriptions.mockResolvedValue([activeSubscription])

  const wrapper = mount(UsageView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        TablePageLayout: TablePageLayoutStub,
        Pagination: true,
        EmptyState: true,
        Select: SelectStub,
        DateRangePicker: true,
        DataTable: DataTableStub,
        Icon: IconStub,
        Teleport: true,
      },
    },
  })

  await flushPromises()
  await nextTick()
  return wrapper
}

const captureCsvExport = () => {
  let exportedBlob: Blob | null = null
  const originalCreateObjectURL = window.URL.createObjectURL
  const originalRevokeObjectURL = window.URL.revokeObjectURL
  window.URL.createObjectURL = vi.fn((blob: Blob | MediaSource) => {
    exportedBlob = blob as Blob
    return 'blob:usage-export'
  }) as typeof window.URL.createObjectURL
  window.URL.revokeObjectURL = vi.fn(() => {}) as typeof window.URL.revokeObjectURL
  const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})

  return {
    getBlob: () => exportedBlob,
    restore: () => {
      window.URL.createObjectURL = originalCreateObjectURL
      window.URL.revokeObjectURL = originalRevokeObjectURL
      clickSpy.mockRestore()
    },
    clickSpy,
  }
}

const readBlobText = (blob: Blob) =>
  new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result))
    reader.onerror = () => reject(reader.error)
    reader.readAsText(blob)
  })

const parseSimpleCsvLine = (line: string): string[] => {
  const cells: string[] = []
  let current = ''
  let quoted = false

  for (let i = 0; i < line.length; i += 1) {
    const char = line[i]
    const next = line[i + 1]

    if (char === '"' && quoted && next === '"') {
      current += '"'
      i += 1
    } else if (char === '"') {
      quoted = !quoted
    } else if (char === ',' && !quoted) {
      cells.push(current)
      current = ''
    } else {
      current += char
    }
  }

  cells.push(current)
  return cells
}

describe('user UsageView', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    query.mockReset()
    getStatsByDateRange.mockReset()
    list.mockReset()
    getAvailableGroups.mockReset()
    getMySubscriptions.mockReset()
    routerPush.mockReset()
    showError.mockReset()
    showWarning.mockReset()
    showSuccess.mockReset()
    showInfo.mockReset()
    window.localStorage.clear()

    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      writable: true,
      value: vi.fn().mockImplementation(() => ({
        matches: true,
        media: '',
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    })

    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
      x: 0,
      y: 0,
      top: 20,
      left: 20,
      right: 120,
      bottom: 40,
      width: 100,
      height: 20,
      toJSON: () => ({}),
    } as DOMRect)

    ;(globalThis as any).ResizeObserver = class {
      observe() {}
      disconnect() {}
    }
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders subscriptions above usage records and keeps renewal routing', async () => {
    const wrapper = await mountUsageView()

    expect(wrapper.find('[data-testid="user-subscriptions-panel"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('My Subscriptions')
    expect(wrapper.text()).not.toContain('Usage Records')
    expect(wrapper.text()).toContain('mc')
    expect(wrapper.text()).toContain('Active')
    expect(wrapper.find('.table-headers').exists()).toBe(true)

    const renewButton = wrapper.findAll('button').find((button) => button.text().includes('Renew Now'))
    expect(renewButton).toBeTruthy()
    await renewButton!.trigger('click')

    expect(routerPush).toHaveBeenCalledWith({
      path: '/purchase',
      query: { tab: 'subscription', group: '9' },
    })
  })

  it('shows subscription empty state on the combined usage page', async () => {
    getMySubscriptions.mockResolvedValueOnce([])
    const wrapper = await mountUsageView()

    expect(wrapper.text()).toContain('No Active Subscriptions')
    expect(wrapper.find('.table-headers').exists()).toBe(true)
  })

  it('renders the actual billing group with the row rate multiplier', async () => {
    const wrapper = await mountUsageView([baseUsageLog({ rate_multiplier: 1.25 })])

    const groupCell = wrapper.find('.table-cell[data-column="group"]')
    expect(groupCell.exists()).toBe(true)
    expect(groupCell.text()).toContain('mc')
    expect(groupCell.text()).toContain('1.25x')
  })

  it('shows original and billed cost in the cost column', async () => {
    const wrapper = await mountUsageView([
      baseUsageLog({
        total_cost: 0.0134,
        actual_cost: 0.01072,
      }),
    ])

    const costCell = wrapper.find('.table-cell[data-column="cost"]')
    expect(costCell.text()).toContain('$0.013400')
    expect(costCell.text()).toContain('✪ 0.010720')
    expect(costCell.find('.line-through').text()).toContain('$0.013400')
  })

  it('renders compact scope totals without restoring duplicated stat cards', async () => {
    const wrapper = await mountUsageView()

    const text = wrapper.text()

    expect(text).toContain('All API Keys / All Groups')
    expect(text).toContain('Total Tokens100')
    expect(text).toContain('Total Cost✪ 0.09')
    expect(text).not.toContain('Total Requests0In selected range')
    expect(text).not.toContain('Average Duration')
  })

  it('shows cache read percentage in the cache read column', async () => {
    const wrapper = await mountUsageView([
      baseUsageLog({
        input_tokens: 100,
        output_tokens: 200,
        cache_creation_tokens: 0,
        cache_read_tokens: 700,
      }),
    ])

    const cacheReadCell = wrapper.find('.table-cell[data-column="cache_read"]')
    expect(cacheReadCell.text()).toContain('700')
    expect(cacheReadCell.text()).toContain('87.5%')
    expect(cacheReadCell.find('[title]').attributes('title')).toBe('700 (87.5%)')
  })

  it('renders studio bridge rows with null timing values', async () => {
    const wrapper = await mountUsageView([
      baseUsageLog({
        request_id: 'studio-bridge-null-duration',
        inbound_endpoint: '/studio-bridge/chat',
        input_tokens: 0,
        output_tokens: 0,
        cache_read_tokens: 0,
        cache_creation_tokens: 0,
        first_token_ms: null,
        duration_ms: null,
        total_cost: 0.001,
        actual_cost: 0.001,
      }),
    ])

    expect(wrapper.find('.table-row').exists()).toBe(true)
    expect(wrapper.find('.table-cell[data-column="model"]').text()).toBe('gpt-5.5 (XHigh)')
    expect(wrapper.find('.table-cell[data-column="latency"]').text()).toContain('-')
    expect(wrapper.find('.table-cell[data-column="cost"]').text()).toContain('✪ 0.001000')
  })

  it('omits the model suffix when no reasoning effort was recorded', async () => {
    const wrapper = await mountUsageView([baseUsageLog({ reasoning_effort: null })])

    expect(wrapper.find('.table-cell[data-column="model"]').text()).toBe('gpt-5.5')
  })

  it('passes group filter to usage list and stats requests', async () => {
    const wrapper = await mountUsageView()

    await wrapper.find('select[data-placeholder="All Groups"]').setValue('9')
    await flushPromises()

    const listParams = query.mock.calls.at(-1)?.[0] as Record<string, unknown>
    expect(listParams.group_id).toBe(9)
    expect(getStatsByDateRange.mock.calls.at(-1)?.[2]).toEqual(
      expect.objectContaining({ group_id: 9 })
    )
  })

  it('passes model, request type, and billing mode from more filters', async () => {
    const wrapper = await mountUsageView()

    const moreFiltersButton = wrapper.findAll('button').find((button) => button.text().includes('More Filters'))
    expect(moreFiltersButton).toBeTruthy()
    await moreFiltersButton!.trigger('click')
    await wrapper.find('input[placeholder="Exact model name"]').setValue('gpt-5.5')
    await wrapper.find('input[placeholder="Exact model name"]').trigger('keydown.enter')
    await wrapper.find('select[data-placeholder="All Types"]').setValue('stream')
    await wrapper.find('select[data-placeholder="All Billing Modes"]').setValue('image')
    await flushPromises()

    const listParams = query.mock.calls.at(-1)?.[0] as Record<string, unknown>
    expect(listParams).toEqual(
      expect.objectContaining({
        model: 'gpt-5.5',
        request_type: 'stream',
        billing_mode: 'image',
      })
    )
    expect(getStatsByDateRange.mock.calls.at(-1)?.[2]).toEqual(
      expect.objectContaining({
        model: 'gpt-5.5',
        request_type: 'stream',
        billing_mode: 'image',
      })
    )
  })

  it('hides low-frequency columns by default and can toggle them on', async () => {
    const wrapper = await mountUsageView()
    const headerText = () => wrapper.find('.table-headers').text()

    expect(headerText()).toContain('Billing Group')
    expect(headerText()).toContain('Cache Read')
    expect(headerText()).toContain('Latency Health')
    expect(headerText()).not.toContain('First Token')
    expect(headerText()).not.toContain('Endpoint')
    expect(headerText()).not.toContain('Reasoning Effort')
    expect(headerText()).not.toContain('User Agent')

    await wrapper.find('button[title="Columns"]').trigger('click')
    const reasoningButton = wrapper.findAll('button').find((button) => button.text().includes('Reasoning Effort'))
    expect(reasoningButton).toBeTruthy()
    await reasoningButton!.trigger('click')
    await nextTick()

    expect(headerText()).toContain('Reasoning Effort')
    expect(window.localStorage.getItem('usage-visible-columns:v3')).toContain('reasoning_effort')
  })

  it('elevates the filter surface only while the column settings menu is open', async () => {
    const wrapper = await mountUsageView()
    const filterSurface = wrapper.get('[data-test="user-usage-filter-surface"]')
    const settingsButton = wrapper.get('[data-test="user-usage-column-settings"]')

    expect(filterSurface.classes()).not.toContain('relative')
    expect(filterSurface.classes()).not.toContain('z-[221]')

    await settingsButton.trigger('click')
    expect(filterSurface.classes()).toContain('relative')
    expect(filterSurface.classes()).toContain('z-[221]')

    await settingsButton.trigger('click')
    expect(filterSurface.classes()).not.toContain('relative')
    expect(filterSurface.classes()).not.toContain('z-[221]')
  })

  it('migrates v2 timing columns to the combined latency column', async () => {
    window.localStorage.setItem('usage-visible-columns:v2', JSON.stringify([
      'api_key', 'model', 'first_token', 'duration', 'created_at', 'actions',
    ]))

    const wrapper = await mountUsageView()
    const headerText = wrapper.find('.table-headers').text()

    expect(headerText).toContain('Latency Health')
    expect(headerText).not.toContain('First Token')
    expect(window.localStorage.getItem('usage-visible-columns:v3')).toContain('latency')
  })

  it('opens row details with group, request, user-agent, token, and cost data', async () => {
    const wrapper = await mountUsageView([
      baseUsageLog({
        request_id: 'req-detail',
        input_tokens: 1520,
        output_tokens: 959,
        cache_read_tokens: 380,
        actual_cost: 0.011628,
        total_cost: 0.012,
      }),
    ])

    await wrapper.find('button[title="Details"]').trigger('click')
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Route Info')
    expect(text).toContain('Request Info')
    expect(text).toContain('mc')
    expect(text).toContain('req-detail')
    expect(text).toContain('Codex Desktop/0.133.0-alpha.1')
    expect(text).toContain('1,520')
    expect(text).toContain('959')
    expect(text).toContain('380 (20.0%)')
    expect(text).toContain('✪ 0.011628')
  })

  it('exports csv with group, request id, user-agent, and current filters', async () => {
    const log = baseUsageLog({
      request_id: 'req-user-export',
      group_id: 9,
      group: availableGroup,
      user_agent: 'Codex Desktop CSV',
    })
    const wrapper = await mountUsageView([log])
    await wrapper.find('select[data-placeholder="All Groups"]').setValue('9')
    await flushPromises()

    const csvExport = captureCsvExport()
    try {
      const setupState = (wrapper.vm as any).$?.setupState
      await setupState.exportToCSV()

      expect(csvExport.getBlob()).not.toBeNull()
      const csv = await readBlobText(csvExport.getBlob() as Blob)
      expect(csv).toContain('Group Name')
      expect(csv).toContain('Group ID')
      expect(csv).toContain('Request ID')
      expect(csv).toContain('User-Agent')
      expect(csv).toContain('mc')
      expect(csv).toContain('9')
      expect(csv).toContain('req-user-export')
      expect(csv).toContain('Codex Desktop CSV')
      const [headerLine, dataLine] = csv.split('\n')
      const headers = parseSimpleCsvLine(headerLine)
      const values = parseSimpleCsvLine(dataLine)
      const row = Object.fromEntries(headers.map((header, index) => [header, values[index]]))
      expect(row['Billed Cost']).toBe('0.09288300')
      expect(row['Original Cost']).toBe('0.09288300')
      expect(row['Billed Cost']).not.toContain('$')
      expect(row['Billed Cost']).not.toContain('✪')
      expect(row['Original Cost']).not.toContain('$')
      expect(row['Original Cost']).not.toContain('✪')
      expect(
        query.mock.calls.some((call) => {
          const params = call[0] as Record<string, unknown> | undefined
          return params?.page_size === 100 && params?.group_id === 9
        })
      ).toBe(true)
      expect(csvExport.clickSpy).toHaveBeenCalled()
      expect(showSuccess).toHaveBeenCalled()
    } finally {
      csvExport.restore()
    }
  })

  it('shows fast service tier and unit prices in user tooltip', async () => {
    const wrapper = await mountUsageView()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.tooltipData = baseUsageLog({
      request_id: 'req-user-1',
      actual_cost: 0.092883,
      total_cost: 0.092883,
      rate_multiplier: 1,
      service_tier: 'priority',
      input_cost: 0.020285,
      output_cost: 0.00303,
      cache_creation_cost: 0,
      cache_read_cost: 0.069568,
      input_tokens: 4057,
      output_tokens: 101,
    })
    setupState.tooltipVisible = true
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Service tier')
    expect(text).toContain('Fast')
    expect(text).toContain('Rate')
    expect(text).toContain('1.00x')
    expect(text).toContain('Billed')
    expect(text).toContain('✪ 0.092883')
    expect(text).toContain('Official reference')
    expect(text).toContain('$0.092883')
    expect(text).toContain('$5.0000 / 1M tokens')
    expect(text).toContain('$30.0000 / 1M tokens')
  })

  it('shows cache read percentage in the token tooltip', async () => {
    const wrapper = await mountUsageView()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.tokenTooltipData = baseUsageLog({
      input_tokens: 4057,
      output_tokens: 101,
      cache_creation_tokens: 0,
      cache_read_tokens: 943,
    })
    setupState.tokenTooltipVisible = true
    await nextTick()

    expect(wrapper.text()).toContain('943 (18.9%)')
  })

  it('shows separate text and image input/output usage', async () => {
    const row = baseUsageLog({
      image_count: 1,
      billing_mode: 'token',
      input_tokens: 371,
      image_input_tokens: 352,
      output_tokens: 439,
      image_output_tokens: 400,
      input_cost: 0.000095,
      image_input_cost: 0.002816,
      output_cost: 0.00039,
      image_output_cost: 0.012,
      total_cost: 0.015301,
      actual_cost: 0.015301,
    })
    const wrapper = await mountUsageView([row])
    const setupState = (wrapper.vm as any).$?.setupState

    setupState.tokenTooltipData = row
    setupState.tokenTooltipVisible = true
    setupState.tooltipData = row
    setupState.tooltipVisible = true
    await nextTick()

    const text = wrapper.text()
    expect(wrapper.find('.table-cell[data-column="billing_mode"]').text()).toContain('Token')
    expect(text).toContain('Image input tokens')
    expect(text).toContain('352')
    expect(text).toContain('Image output tokens')
    expect(text).toContain('400')
    expect(text).toContain('Image input cost')
    expect(text).toContain('$0.002816')
    expect(text).toContain('Image output cost')
    expect(text).toContain('$0.012000')
    expect(text).toContain('$8.0000 / 1M tokens')
    expect(text).toContain('$30.0000 / 1M tokens')
  })

  it('exports historical image rows with image billing mode derived from image_count', async () => {
    const wrapper = await mountUsageView([
      baseUsageLog({
        request_id: 'req-user-export-legacy-image',
        actual_cost: 0.2,
        total_cost: 0.2,
        input_cost: 0,
        output_cost: 0,
        cache_creation_cost: 0,
        cache_read_cost: 0,
        input_tokens: 0,
        output_tokens: 0,
        cache_creation_tokens: 0,
        cache_read_tokens: 0,
        image_count: 1,
        image_size: null,
        billing_mode: null,
        first_token_ms: null,
        duration_ms: 345,
        model: 'gpt-image-2',
        reasoning_effort: null,
      }),
    ])

    const csvExport = captureCsvExport()
    try {
      const setupState = (wrapper.vm as any).$?.setupState
      await setupState.exportToCSV()

      expect(csvExport.getBlob()).not.toBeNull()
      const csv = await readBlobText(csvExport.getBlob() as Blob)
      expect(csv).toContain('Billing Mode')
      expect(csv).toContain('Image')
      expect(csv).not.toContain(',Token,0,0,0,0,')
    } finally {
      csvExport.restore()
    }
  })

  it('does not display a 2K fallback for historical image rows with missing size', async () => {
    const wrapper = await mountUsageView([
      baseUsageLog({
        request_id: 'req-user-legacy-missing-image',
        actual_cost: 0.2,
        total_cost: 0.2,
        input_cost: 0,
        output_cost: 0,
        cache_creation_cost: 0,
        cache_read_cost: 0,
        input_tokens: 0,
        output_tokens: 0,
        cache_creation_tokens: 0,
        cache_read_tokens: 0,
        image_count: 1,
        image_size: null,
        image_input_size: null,
        image_output_size: null,
        image_size_source: null,
        image_size_breakdown: null,
        billing_mode: null,
        first_token_ms: null,
        duration_ms: 1,
        model: 'gpt-image-2',
      }),
    ])

    const text = wrapper.text()
    expect(text).toContain('Image')
    expect(text).toContain('not recorded')
    expect(text).not.toContain('(2K)')
  })

  it('shows image billing metadata in the user cost tooltip', async () => {
    const wrapper = await mountUsageView([])

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.tooltipData = baseUsageLog({
      request_id: 'req-user-output-image',
      actual_cost: 0.8,
      total_cost: 0.8,
      rate_multiplier: 1,
      service_tier: null,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
      cache_creation_tokens: 0,
      cache_read_tokens: 0,
      billing_mode: null,
      image_count: 2,
      image_size: '4K',
      image_input_size: '1024x1024',
      image_output_size: '3840x2160',
      image_size_source: 'output',
      image_size_breakdown: { '4K': 2 },
    })
    setupState.tooltipVisible = true
    await nextTick()

    const text = wrapper.text()
    expect(text).toContain('Image count')
    expect(text).toContain('Billing size')
    expect(text).toContain('4K')
    expect(text).toContain('Size source')
    expect(text).toContain('Upstream output')
    expect(text).toContain('Input size')
    expect(text).toContain('1024x1024')
    expect(text).toContain('Output size')
    expect(text).toContain('3840x2160')
    expect(text).toContain('4K x 2')
  })
})
