import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AccountsView from '../AccountsView.vue'
import UpstreamBillingRateCell from '@/components/account/UpstreamBillingRateCell.vue'
import enAccounts from '@/i18n/locales/en/admin/accounts'
import zhAccounts from '@/i18n/locales/zh/admin/accounts'

const {
  listAccounts,
  listWithEtag,
  getBatchTodayStats,
  getAllProxies,
  getAllGroups,
  getUpstreamBillingProbeSettings,
  routeName
} = vi.hoisted(() => ({
  listAccounts: vi.fn(),
  listWithEtag: vi.fn(),
  getBatchTodayStats: vi.fn(),
  getAllProxies: vi.fn(),
  getAllGroups: vi.fn(),
  getUpstreamBillingProbeSettings: vi.fn(),
  routeName: { value: 'AdminAccounts' }
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      list: listAccounts,
      listWithEtag,
      getBatchTodayStats,
      getUpstreamBillingProbeSettings,
      delete: vi.fn(),
      batchClearError: vi.fn(),
      batchRefresh: vi.fn(),
      toggleSchedulable: vi.fn()
    },
    proxies: {
      getAll: getAllProxies
    },
    groups: {
      getAll: getAllGroups
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn()
  })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    token: 'test-token'
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

vi.mock('vue-router', () => ({
  useRoute: () => ({
    get name() {
      return routeName.value
    }
  })
}))

const DataTableStub = {
  props: ['columns', 'data'],
  template: `
    <div data-test="data-table">
      <template v-for="column in columns" :key="column.key">
        <div v-if="column.key === 'usage'" data-test="usage-header">
          <slot :name="'header-' + column.key" :column="column" />
        </div>
        <div v-if="column.key === 'upstream_billing_rate'" data-test="upstream-billing-header">
          <slot :name="'header-' + column.key" :column="column" />
        </div>
      </template>
      <div v-for="row in data" :key="row.id" data-test="account-rate">
        <slot name="cell-rate_multiplier" :row="row" />
        <slot name="cell-upstream_billing_rate" :row="row" />
      </div>
    </div>
  `
}

const HelpTooltipStub = {
  props: ['content', 'widthClass'],
  template: '<span data-test="usage-windows-hint">{{ content }}</span>'
}

function mountView() {
  return mount(AccountsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
        },
        DataTable: DataTableStub,
        HelpTooltip: HelpTooltipStub,
        Pagination: true,
        BaseDialog: true,
        ConfirmDialog: true,
        AccountTableActions: { template: '<div><slot name="beforeCreate" /><slot name="after" /></div>' },
        AccountTableFilters: {
          props: ['groups'],
          template: '<div data-test="account-filters" :data-group-count="groups.length"></div>'
        },
        AccountBulkActionsBar: true,
        AccountActionMenu: true,
        ImportDataModal: true,
        ReAuthAccountModal: true,
        AccountTestModal: true,
        AccountStatsModal: true,
        ScheduledTestsPanel: true,
        SyncFromCrsModal: true,
        TempUnschedStatusModal: true,
        ErrorPassthroughRulesModal: true,
        TLSFingerprintProfilesModal: true,
        CreateAccountModal: true,
        EditAccountModal: true,
        BulkEditAccountModal: true,
        PlatformTypeBadge: true,
        AccountCapacityCell: true,
        AccountStatusIndicator: true,
        AccountTodayStatsCell: true,
        AccountGroupsCell: true,
        AccountUsageCell: true,
        Icon: true
      }
    }
  })
}

describe('admin AccountsView usage windows hint', () => {
  beforeEach(() => {
    localStorage.clear()

    listAccounts.mockReset()
    listWithEtag.mockReset()
    getBatchTodayStats.mockReset()
    getAllProxies.mockReset()
    getAllGroups.mockReset()
    getUpstreamBillingProbeSettings.mockReset()
    routeName.value = 'AdminAccounts'

    listAccounts.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 0
    })
    listWithEtag.mockResolvedValue({
      notModified: true,
      etag: null,
      data: null
    })
    getBatchTodayStats.mockResolvedValue({ stats: {} })
    getAllProxies.mockResolvedValue([])
    getAllGroups.mockResolvedValue([])
    getUpstreamBillingProbeSettings.mockResolvedValue({ enabled: true, interval_minutes: 30 })
  })

  it('keeps groups available when loading proxies fails', async () => {
    getAllProxies.mockRejectedValue(new Error('proxy service unavailable'))
    getAllGroups.mockResolvedValue([{ id: 7, name: 'production' }])

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[data-test="account-filters"]').attributes('data-group-count')).toBe('1')
  })

  it('renders an explanatory tooltip next to the usage windows column header', async () => {
    const wrapper = mountView()
    await flushPromises()

    const header = wrapper.find('[data-test="usage-header"]')
    expect(header.exists()).toBe(true)
    expect(header.text()).toContain('admin.accounts.columns.usageWindows')

    const hint = wrapper.find('[data-test="usage-windows-hint"]')
    expect(hint.exists()).toBe(true)
    expect(hint.text()).toBe('admin.accounts.usageWindowsHint')
  })

  it('renders the upstream billing trust warning next to the declared-rate column', async () => {
    const wrapper = mountView()
    await flushPromises()

    const header = wrapper.find('[data-test="upstream-billing-header"]')
    expect(header.exists()).toBe(true)
    expect(header.text()).toContain('admin.accounts.columns.upstreamBillingRate')
    expect(wrapper.findAll('[data-test="usage-windows-hint"]').some(node =>
      node.text() === 'admin.accounts.upstreamBilling.trustWarning'
    )).toBe(true)
    const columns = wrapper.getComponent(DataTableStub).props('columns') as Array<{ key: string; sortable: boolean }>
    expect(columns.find(column => column.key === 'upstream_billing_rate')?.sortable).toBe(false)
  })

  it('shows account multipliers with enough precision to match declared rates', async () => {
    listAccounts.mockResolvedValueOnce({
      items: [{
        id: 7,
        name: 'precision-account',
        platform: 'gemini',
        type: 'apikey',
        status: 'active',
        schedulable: true,
        rate_multiplier: 0.065,
        extra: {
          upstream_billing_probe_enabled: true,
          upstream_billing_rate_sync_enabled: true
        },
        created_at: '2026-07-13T00:00:00Z',
        updated_at: '2026-07-13T00:00:00Z'
      }],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[data-test="account-rate"]').text()).toBe('0.065x')
    const indicator = wrapper.get('[data-testid="account-rate-sync-indicator"]')
    expect(indicator.attributes('title')).toBe('admin.accounts.upstreamBilling.syncedRateTooltip')
  })

  it('passes the disabled global probe state to the upstream rate cell on the shared account page', async () => {
    routeName.value = 'AdminSharedAccounts'
    getUpstreamBillingProbeSettings.mockResolvedValueOnce({ enabled: false, interval_minutes: 30 })
    listAccounts.mockResolvedValueOnce({
      items: [{
        id: 8,
        name: 'probe-disabled-account',
        platform: 'openai',
        type: 'apikey',
        status: 'active',
        schedulable: true,
        rate_multiplier: 1,
        extra: { upstream_billing_probe_enabled: true },
        created_at: '2026-07-13T00:00:00Z',
        updated_at: '2026-07-13T00:00:00Z'
      }],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })

    const wrapper = mountView()
    await flushPromises()

    const rateCell = wrapper.getComponent(UpstreamBillingRateCell)
    expect(rateCell.props('globalProbeEnabled')).toBe(false)
    expect(rateCell.vm.$attrs).not.toHaveProperty('interval-minutes')
  })

  it('defines every upstream probe settings message in both locales', () => {
    const keys = ['autoProbeSettings', 'intervalMinutes', 'settingsSaved', 'settingsFailed'] as const
    for (const key of keys) {
      expect(zhAccounts.upstreamBilling[key]).toBeTruthy()
      expect(enAccounts.upstreamBilling[key]).toBeTruthy()
    }
  })
})
