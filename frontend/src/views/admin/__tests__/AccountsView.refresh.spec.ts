import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import AccountsView from '../AccountsView.vue'

const { list, stats } = vi.hoisted(() => ({ list: vi.fn(), stats: vi.fn() }))
vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      list,
      listWithEtag: vi.fn().mockResolvedValue({ notModified: true, data: null, etag: null }),
      getBatchTodayStats: stats,
      getUpstreamBillingProbeSettings: vi.fn().mockResolvedValue({ enabled: false })
    },
    proxies: { getAll: vi.fn().mockResolvedValue([]) },
    groups: { getAll: vi.fn().mockResolvedValue([]) }
  }
}))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn(), showInfo: vi.fn() })
}))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ token: 'test' }) }))
vi.mock('vue-router', () => ({ useRoute: () => ({ name: 'AdminAccounts' }) }))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key })
}))

const table = {
  props: ['data', 'loading', 'columns'],
  template: '<div data-test="table"><span v-if="loading">initial-loading</span><span v-else v-for="row in data" :key="row.id">{{ row.name }}</span></div>'
}
let wrapper: VueWrapper | undefined
function mountAccounts() {
  const stubs: Record<string, unknown> = {}
  for (const name of [
    'Pagination', 'ConfirmDialog', 'BaseDialog', 'AccountTableFilters', 'AccountBulkActionsBar',
    'AccountActionMenu', 'ImportDataModal', 'ReAuthAccountModal', 'AccountTestModal', 'AccountStatsModal',
    'ScheduledTestsPanel', 'SyncFromCrsModal', 'TempUnschedStatusModal', 'ErrorPassthroughRulesModal',
    'TLSFingerprintProfilesModal', 'TotpStepUpDialog', 'CreateAccountModal', 'EditAccountModal',
    'BulkEditAccountModal', 'PlatformTypeBadge', 'AccountCapacityCell', 'AccountStatusIndicator',
    'AccountTodayStatsCell', 'AccountGroupsCell', 'AccountUsageCell', 'UpstreamBillingRateCell', 'Icon'
  ]) stubs[name] = true
  wrapper = mount(AccountsView, { global: { stubs: {
    ...stubs,
    AppLayout: { template: '<div><slot /></div>' },
    TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /></div>' },
    DataTable: table,
    AccountTableActions: { emits: ['refresh'], template: '<button data-test="refresh" @click="$emit(\'refresh\')">refresh</button>' }
  } } })
  return wrapper
}
const result = (name: string) => ({
  items: [{
    id: 7, name, platform: 'openai', type: 'apikey', status: 'active', schedulable: true,
    rate_multiplier: 1, groups: [], credentials: {}, extra: {},
    created_at: '2026-11-16T00:00:00Z', updated_at: '2026-11-16T00:00:00Z'
  }],
  total: 1, pages: 1, page: 1, page_size: 20
})

describe('AccountsView refresh continuity', () => {
  beforeEach(() => {
    localStorage.clear()
    list.mockReset()
    stats.mockResolvedValue({ stats: {} })
  })
  afterEach(() => wrapper?.unmount())

  it('shows first-load skeleton, then retains rows and disables stale interactions during refresh', async () => {
    let finish!: (value: ReturnType<typeof result>) => void
    list.mockReturnValueOnce(new Promise((resolve) => { finish = resolve }))
    const view = mountAccounts()
    await flushPromises()
    expect(view.text()).toContain('initial-loading')
    finish(result('old-account'))
    await flushPromises()
    expect(view.get('[data-test="table"]').text()).toContain('old-account')
    list.mockReturnValueOnce(new Promise((resolve) => { finish = resolve }))
    await view.get('[data-test="refresh"]').trigger('click')
    await flushPromises()
    expect(view.get('[data-test="table"]').text()).toContain('old-account')
    expect(view.text()).not.toContain('initial-loading')
    expect(view.get('[aria-busy="true"]').attributes('inert')).toBeDefined()
    finish(result('new-account'))
    await flushPromises()
    expect(view.get('[data-test="table"]').text()).toContain('new-account')
    expect(view.get('[aria-busy="false"]').attributes('inert')).toBeUndefined()
  })
})
