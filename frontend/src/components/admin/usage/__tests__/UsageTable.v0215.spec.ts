const ipGeoMocks = vi.hoisted(() => ({
  getEntry: vi.fn(() => ({ status: 'idle' as const })),
  fetchOne: vi.fn(),
  fetchBatch: vi.fn(),
}))

const appStoreMocks = vi.hoisted(() => ({
  showSuccess: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('@/utils/ipGeoLookup', () => ipGeoMocks)
vi.mock('@/stores/app', () => ({ useAppStore: () => appStoreMocks }))

import { afterEach, describe, expect, it, vi, beforeEach } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
enableAutoUnmount(afterEach)
afterEach(() => vi.restoreAllMocks())

import en from '@/i18n/locales/en/admin/usage'
import zh from '@/i18n/locales/zh/admin/usage'
import UsageTable from '../UsageTable.vue'

let locale: 'en' | 'zh' = 'en'
const localizedMessages: Record<'en' | 'zh', Record<string, string>> = {
  en: {
    'admin.usage.longContext': en.longContext,
    'admin.usage.longContextPricingTooltip': en.longContextPricingTooltip,
  },
  zh: {
    'admin.usage.longContext': zh.longContext,
    'admin.usage.longContextPricingTooltip': zh.longContextPricingTooltip,
  },
}

const messages: Record<string, string> = {
  'admin.usage.userDeletedBadge': 'Deleted',
  'usage.costDetails': 'Cost Breakdown',
  'admin.usage.inputCost': 'Input Cost',
  'admin.usage.outputCost': 'Output Cost',
  'admin.usage.cacheCreationCost': 'Cache Creation Cost',
  'admin.usage.cacheReadCost': 'Cache Read Cost',
  'usage.inputTokenPrice': 'Input price',
  'usage.outputTokenPrice': 'Output price',
  'usage.perMillionTokens': '/ 1M tokens',
  'usage.serviceTier': 'Service tier',
  'usage.serviceTierPriority': 'Fast',
  'usage.serviceTierUltrafast': 'Ultrafast',
  'usage.serviceTierFlex': 'Flex',
  'usage.serviceTierStandard': 'Standard',
  'usage.rate': 'Rate',
  'usage.accountMultiplier': 'Account rate',
  'usage.original': 'Original',
  'usage.userBilled': 'User billed',
  'usage.accountBilled': 'Account billed',
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
  'usage.stream': 'Stream',
  'usage.sync': 'Sync',
  'usage.nativeCompactionV2': 'Compaction',
  'admin.usage.billingModeToken': 'Token',
  'admin.usage.billingModePerRequest': 'Per request',
  'admin.usage.billingModeImage': 'Image',
	'admin.usage.requestIdCopied': 'Request ID copied',
	'admin.usage.upstreamRequestIdCopied': 'Upstream ID copied',
	'keys.copied': 'Copied',
	'keys.copyToClipboard': 'Copy to clipboard',
	'common.copyFailed': 'Copy failed',
	'usage.requestedModel': 'Requested',
	'usage.sentUpstreamModel': 'Sent upstream',
	'usage.upstreamResponseModel': 'Upstream response',
	'usage.modelVariant': 'Possible version variant',
	'usage.modelMismatch': 'Different model',
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => localizedMessages[locale][key] ?? messages[key] ?? key,
    }),
  }
})

const DataTableStub = {
  props: ['data'],
  template: `
    <div>
      <div v-for="row in data" :key="row.request_id" :data-request-id="row.request_id">
        <slot name="cell-model" :row="row" :value="row.model" />
        <slot name="cell-reasoning_effort" :row="row" :value="row.reasoning_effort" />
        <slot name="cell-billing_mode" :row="row" />
        <slot name="cell-tokens" :row="row" />
        <slot name="cell-cost" :row="row" />
        <slot name="cell-request_id" :row="row" />
        <slot name="cell-upstream_request_id" :row="row" />
      </div>
    </div>
  `,
}

const baseImageRow = {
  request_id: 'req-admin-image',
  model: 'gpt-image-2',
  actual_cost: 0.4,
  total_cost: 0.4,
  account_rate_multiplier: 1,
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
  cache_creation_5m_tokens: 0,
  cache_creation_1h_tokens: 0,
  cache_ttl_overridden: false,
  billing_mode: 'image',
  image_count: 2,
  image_size: '2K',
  image_input_size: null,
  image_output_size: null,
  image_size_source: null,
  image_size_breakdown: null,
}

describe('admin UsageTable tooltip', () => {
  beforeEach(() => {
    locale = 'en'
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
  })

  it.each([
    ['en', 'Long context', 'Long-context pricing was applied. Input and output rates depend on the pricing tier, not a uniform multiplier.'],
    ['zh', '长上下文', '已应用长上下文计费。输入和输出费率取决于定价档位，并非统一倍率。'],
  ] as const)('marks only applied long-context billing with localized text in %s', (language, label, tooltip) => {
    locale = language
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            ...baseImageRow,
            request_id: 'req-long-context-enabled',
            long_context_billing_applied: true,
          },
          {
            ...baseImageRow,
            request_id: 'req-long-context-disabled',
            long_context_billing_applied: false,
          },
          {
            ...baseImageRow,
            request_id: 'req-long-context-absent',
          },
        ],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.findAll('[data-testid="long-context-billing-marker"]')).toHaveLength(1)
    const marker = wrapper.get('[data-request-id="req-long-context-enabled"] [data-testid="long-context-billing-marker"]')
    expect(marker.text()).toBe(label)
    expect(marker.attributes('title')).toBe(tooltip)
    expect(wrapper.find('[data-request-id="req-long-context-disabled"] [data-testid="long-context-billing-marker"]').exists()).toBe(false)
    expect(wrapper.find('[data-request-id="req-long-context-absent"] [data-testid="long-context-billing-marker"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('x2')
  })

  it('preserves zero account and GroupID pricing multipliers in the mounted cost tooltip', async () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{
          ...baseImageRow, total_cost: 5, actual_cost: 0,
          account_rate_multiplier: 0, rate_multiplier: 4, pricing_rate_multiplier: 0,
        }],
        loading: false, columns: [],
      },
      global: { stubs: { DataTable: DataTableStub, EmptyState: true, Icon: true, Teleport: true } },
    })
    expect(wrapper.text()).toContain('A $0.000000')
    await wrapper.get('.group.relative').trigger('mouseenter')
    const pricing = wrapper.findAll('div').find(div =>
      div.element.children.length === 2 && div.element.children[0].textContent === 'usage.pricingRate')
    const account = wrapper.findAll('div').find(div =>
      div.element.children.length === 2 && div.element.children[0].textContent === 'Account rate')
    expect(pricing?.element.children[1].textContent).toBe('0.0x')
    expect(account?.element.children[1].textContent).toBe('0.0x')
  })

})
