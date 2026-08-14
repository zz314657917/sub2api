import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AccountsView from '../AccountsView.vue'

const { listAccounts, listWithEtag, getBatchTodayStats, getAllProxies, getAllGroups, getUpstreamBillingProbeSettings } = vi.hoisted(() => ({
  listAccounts: vi.fn(), listWithEtag: vi.fn(), getBatchTodayStats: vi.fn(), getAllProxies: vi.fn(), getAllGroups: vi.fn(), getUpstreamBillingProbeSettings: vi.fn()
}))

vi.mock('@/api/admin', () => ({ adminAPI: { accounts: { list: listAccounts, listWithEtag, getBatchTodayStats, getUpstreamBillingProbeSettings, delete: vi.fn(), batchClearError: vi.fn(), batchRefresh: vi.fn(), toggleSchedulable: vi.fn() }, proxies: { getAll: getAllProxies }, groups: { getAll: getAllGroups } } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn(), showInfo: vi.fn() }) }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ token: 'test-token' }) }))
vi.mock('vue-i18n', async () => ({ ...(await vi.importActual<typeof import('vue-i18n')>('vue-i18n')), useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('vue-router', () => ({ useRoute: () => ({ name: 'AdminAccounts' }) }))

const DataTableStub = { props: ['data'], template: '<div><div v-for="row in data" :key="row.id"><slot name="cell-platform_type" :row="row" /></div></div>' }
const PlatformTypeBadgeStub = { props: ['planType'], template: '<span data-test="plan-type">{{ planType }}</span>' }
const account = (snapshot: Record<string, unknown>) => ({ id: 7, name: 'grok', platform: 'grok', type: 'oauth', status: 'active', schedulable: true, groups: [], credentials: {}, extra: { grok_usage_snapshot: snapshot }, created_at: '2026-08-14T00:00:00Z', updated_at: '2026-08-14T00:00:00Z' })

function mountView() {
  return mount(AccountsView, { global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, TablePageLayout: { template: '<div><slot name="table" /></div>' }, DataTable: DataTableStub, PlatformTypeBadge: PlatformTypeBadgeStub, Pagination: true, BaseDialog: true, ConfirmDialog: true, AccountTableActions: true, AccountTableFilters: true, AccountBulkActionsBar: true, AccountActionMenu: true, ImportDataModal: true, ReAuthAccountModal: true, AccountTestModal: true, AccountStatsModal: true, ScheduledTestsPanel: true, SyncFromCrsModal: true, TempUnschedStatusModal: true, ErrorPassthroughRulesModal: true, TLSFingerprintProfilesModal: true, CreateAccountModal: true, EditAccountModal: true, BulkEditAccountModal: true, AccountCapacityCell: true, AccountStatusIndicator: true, AccountTodayStatsCell: true, AccountGroupsCell: true, AccountUsageCell: true, UpstreamBillingRateCell: true, Icon: true } } })
}

describe('admin AccountsView Grok badge', () => {
  beforeEach(() => {
    listAccounts.mockReset(); listWithEtag.mockReset(); getBatchTodayStats.mockReset(); getAllProxies.mockReset(); getAllGroups.mockReset(); getUpstreamBillingProbeSettings.mockReset()
    listAccounts.mockResolvedValue({ items: [account({ subscription_tier: 'basic', used: 1 })], total: 1, page: 1, page_size: 20, pages: 1 })
    listWithEtag.mockResolvedValue({ notModified: false, etag: 'next', data: { items: [account({ subscription_tier: 'premium', used: 2 })], total: 1, pages: 1 } })
    getBatchTodayStats.mockResolvedValue({ stats: {} }); getAllProxies.mockResolvedValue([]); getAllGroups.mockResolvedValue([]); getUpstreamBillingProbeSettings.mockResolvedValue({ enabled: true, interval_minutes: 30 })
  })

  it('replaces a Grok row after canonical snapshot changes', async () => {
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.get('[data-test="plan-type"]').text()).toBe('basic')

    const refresh = (wrapper.vm as any).$?.setupState?.refreshAccountsIncrementally
    await refresh()
    await flushPromises()

    expect(wrapper.get('[data-test="plan-type"]').text()).toBe('premium')
  })
})
