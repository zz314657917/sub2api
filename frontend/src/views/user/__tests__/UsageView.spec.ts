import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import UsageView from '../UsageView.vue'

const {
  query,
  getStatsByDateRange,
  list,
  getAvailableGroups,
  showError,
  showWarning,
  showSuccess,
  showInfo,
} = vi.hoisted(() => ({
  query: vi.fn(),
  getStatsByDateRange: vi.fn(),
  list: vi.fn(),
  getAvailableGroups: vi.fn(),
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

describe('user UsageView', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    query.mockReset()
    getStatsByDateRange.mockReset()
    list.mockReset()
    getAvailableGroups.mockReset()
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
    expect(costCell.text()).toContain('$0.010720')
    expect(costCell.find('.line-through').text()).toContain('$0.013400')
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
    expect(headerText()).toContain('First Token')
    expect(headerText()).not.toContain('Endpoint')
    expect(headerText()).not.toContain('Reasoning Effort')
    expect(headerText()).not.toContain('User Agent')

    await wrapper.find('button[title="Columns"]').trigger('click')
    const reasoningButton = wrapper.findAll('button').find((button) => button.text().includes('Reasoning Effort'))
    expect(reasoningButton).toBeTruthy()
    await reasoningButton!.trigger('click')
    await nextTick()

    expect(headerText()).toContain('Reasoning Effort')
    expect(window.localStorage.getItem('usage-visible-columns:v2')).toContain('reasoning_effort')
  })

  it('opens row details with group, request, user-agent, token, and cost data', async () => {
    const wrapper = await mountUsageView([
      baseUsageLog({
        request_id: 'req-detail',
        input_tokens: 1520,
        output_tokens: 959,
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
    expect(text).toContain('$0.011628')
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
    expect(text).toContain('$0.092883')
    expect(text).toContain('$5.0000 / 1M tokens')
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
