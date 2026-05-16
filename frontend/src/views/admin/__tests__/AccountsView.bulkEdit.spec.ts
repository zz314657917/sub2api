import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AccountsView from '../AccountsView.vue'

const {
  listAccounts,
  listWithEtag,
  getBatchTodayStats,
  getAllProxies,
  getAllGroups,
  setShareStatus,
  batchSetShareStatus,
  appStore,
  routeName,
  writeText
} = vi.hoisted(() => ({
  listAccounts: vi.fn(),
  listWithEtag: vi.fn(),
  getBatchTodayStats: vi.fn(),
  getAllProxies: vi.fn(),
  getAllGroups: vi.fn(),
  setShareStatus: vi.fn(),
  batchSetShareStatus: vi.fn(),
  appStore: {
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn()
  },
  routeName: { value: 'AdminAccounts' },
  writeText: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      list: listAccounts,
      listWithEtag,
      getBatchTodayStats,
      delete: vi.fn(),
      batchClearError: vi.fn(),
      batchRefresh: vi.fn(),
      toggleSchedulable: vi.fn(),
      setShareStatus,
      batchSetShareStatus
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
  useAppStore: () => appStore
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
    <div data-test="data-table" :data-column-keys="columns.map(column => column.key).join(',')">
      <div data-test="header-select">
        <slot name="header-select" />
      </div>
      <div v-for="row in data" :key="row.id">
        <div data-test="select-cell">
          <slot name="cell-select" :row="row" :value="row.id" />
        </div>
        <span data-test="row-number">{{ row.row_number }}</span>
        <div data-test="name-cell">
          <slot name="cell-name" :row="row" :value="row.name">
            <span>{{ row.name }}</span>
          </slot>
        </div>
      </div>
    </div>
  `
}

const AccountBulkActionsBarStub = {
  props: ['selectedIds', 'showSystemActions', 'showShareReviewActions', 'loading'],
  emits: ['edit-filtered', 'share-status', 'share-status-filtered'],
  template: `
    <div
      data-test="bulk-actions"
      :data-show-system-actions="String(showSystemActions)"
      :data-show-share-review-actions="String(showShareReviewActions)"
      :data-selected-count="String(selectedIds.length)"
    >
      <button data-test="edit-filtered" @click="$emit('edit-filtered')">edit filtered</button>
      <button
        v-if="showShareReviewActions && selectedIds.length > 0"
        data-test="approve-share"
        @click="$emit('share-status', 'active')"
      >
        approve share
      </button>
      <button
        v-if="showShareReviewActions"
        data-test="approve-filtered-share"
        @click="$emit('share-status-filtered', 'active')"
      >
        approve filtered share
      </button>
    </div>
  `
}

const BulkEditAccountModalStub = {
  props: ['show', 'target'],
  template: '<div data-test="bulk-edit-modal" :data-show="String(show)" :data-target-mode="target?.mode ?? \'\'"></div>'
}

const BaseDialogStub = {
  props: ['show', 'title'],
  emits: ['close'],
  template: '<div v-if="show" data-test="bulk-share-result-dialog"><slot /><div data-test="bulk-share-result-footer"><slot name="footer" /></div></div>'
}

const ConfirmDialogStub = {
  props: ['show', 'title', 'message', 'confirmText'],
  emits: ['confirm', 'cancel'],
  template: `
    <div v-if="show" data-test="confirm-dialog" :data-title="title" :data-message="message">
      <slot />
      <button data-test="confirm-dialog-confirm" @click="$emit('confirm')">{{ confirmText }}</button>
      <button data-test="confirm-dialog-cancel" @click="$emit('cancel')">cancel</button>
    </div>
  `
}

describe('admin AccountsView bulk edit scope', () => {
  beforeEach(() => {
    localStorage.clear()

    listAccounts.mockReset()
    listWithEtag.mockReset()
    getBatchTodayStats.mockReset()
    getAllProxies.mockReset()
    getAllGroups.mockReset()
    setShareStatus.mockReset()
    batchSetShareStatus.mockReset()
    appStore.showError.mockReset()
    appStore.showSuccess.mockReset()
    appStore.showInfo.mockReset()
    writeText.mockReset()
    routeName.value = 'AdminAccounts'
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: {
        writeText
      }
    })

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
    setShareStatus.mockImplementation(async (id: number, shareStatus: string) => ({
      id,
      name: `Account ${id}`,
      platform: 'openai',
      type: 'oauth',
      status: 'active',
      schedulable: true,
      owner_user_id: 10,
      share_mode: 'public',
      share_status: shareStatus,
      groups: [],
      credentials: {},
      extra: {}
    }))
    batchSetShareStatus.mockResolvedValue({
      success: 0,
      failed: 0,
      success_ids: [],
      failed_ids: [],
      results: []
    })
  })

  it('opens bulk edit in filtered-results mode from the bulk actions dropdown', async () => {
    const wrapper = mount(AccountsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: true,
          ConfirmDialog: ConfirmDialogStub,
          BaseDialog: BaseDialogStub,
          AccountTableActions: { template: '<div><slot name="beforeCreate" /><slot name="after" /></div>' },
          AccountTableFilters: { template: '<div></div>' },
          AccountBulkActionsBar: AccountBulkActionsBarStub,
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
          BulkEditAccountModal: BulkEditAccountModalStub,
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

    await flushPromises()
    await wrapper.get('[data-test="edit-filtered"]').trigger('click')
    await flushPromises()

    expect(wrapper.get('[data-test="bulk-edit-modal"]').attributes('data-show')).toBe('true')
    expect(wrapper.get('[data-test="bulk-edit-modal"]').attributes('data-target-mode')).toBe('filtered')
  })

  it('shows a paginated row number column before account name', async () => {
    listAccounts.mockResolvedValueOnce({
      items: [
        { id: 1, name: 'Account 1', platform: 'openai', type: 'oauth', status: 'active', schedulable: true, groups: [], credentials: {}, extra: {} },
        { id: 2, name: 'Account 2', platform: 'openai', type: 'oauth', status: 'active', schedulable: true, groups: [], credentials: {}, extra: {} }
      ],
      total: 42,
      page: 1,
      page_size: 20,
      pages: 3
    }).mockResolvedValueOnce({
      items: [
        { id: 101, name: 'Account A', platform: 'openai', type: 'oauth', status: 'active', schedulable: true, groups: [], credentials: {}, extra: {} },
        { id: 102, name: 'Account B', platform: 'openai', type: 'oauth', status: 'active', schedulable: true, groups: [], credentials: {}, extra: {} }
      ],
      total: 42,
      page: 3,
      page_size: 20,
      pages: 3
    })

    const wrapper = mount(AccountsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: {
            emits: ['update:page'],
            template: '<button data-test="goto-page-3" @click="$emit(\'update:page\', 3)">go page 3</button>'
          },
          ConfirmDialog: ConfirmDialogStub,
          BaseDialog: BaseDialogStub,
          AccountTableActions: { template: '<div><slot name="beforeCreate" /><slot name="after" /></div>' },
          AccountTableFilters: { template: '<div></div>' },
          AccountBulkActionsBar: AccountBulkActionsBarStub,
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
          BulkEditAccountModal: BulkEditAccountModalStub,
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

    await flushPromises()
    await wrapper.get('[data-test="goto-page-3"]').trigger('click')
    await flushPromises()

    expect(wrapper.get('[data-test="data-table"]').attributes('data-column-keys')?.split(',').slice(0, 3)).toEqual([
      'select',
      'row_number',
      'name'
    ])
    expect(wrapper.findAll('[data-test="row-number"]').map(item => item.text())).toEqual(['41', '42'])
  })

  it('shows the shared upstream account name instead of the saved alias on the shared account page', async () => {
    routeName.value = 'AdminSharedAccounts'
    listAccounts.mockResolvedValueOnce({
      items: [
        {
          id: 1,
          name: '1',
          platform: 'openai',
          type: 'oauth',
          status: 'active',
          schedulable: true,
          owner_user_id: 10,
          share_mode: 'public',
          share_status: 'active',
          groups: [],
          credentials: { email: 'shared-openai@example.com' },
          extra: {}
        }
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })

    const wrapper = mount(AccountsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: true,
          ConfirmDialog: ConfirmDialogStub,
          BaseDialog: BaseDialogStub,
          AccountTableActions: { template: '<div><slot name="beforeCreate" /><slot name="after" /></div>' },
          AccountTableFilters: { template: '<div></div>' },
          AccountBulkActionsBar: AccountBulkActionsBarStub,
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
          BulkEditAccountModal: BulkEditAccountModalStub,
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

    await flushPromises()

    expect(wrapper.get('[data-test="name-cell"]').text()).toBe('shared-openai@example.com')
  })

  it('enables filtered bulk editing on the shared account page with shared account scope', async () => {
    routeName.value = 'AdminSharedAccounts'

    const wrapper = mount(AccountsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: true,
          ConfirmDialog: ConfirmDialogStub,
          BaseDialog: BaseDialogStub,
          AccountTableActions: { template: '<div><slot name="beforeCreate" /><slot name="after" /></div>' },
          AccountTableFilters: { template: '<div></div>' },
          AccountBulkActionsBar: AccountBulkActionsBarStub,
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
          BulkEditAccountModal: BulkEditAccountModalStub,
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

    await flushPromises()
    expect(wrapper.get('[data-test="bulk-actions"]').attributes('data-show-share-review-actions')).toBe('true')
    expect(wrapper.get('[data-test="bulk-actions"]').attributes('data-show-system-actions')).toBe('false')

    await wrapper.get('[data-test="edit-filtered"]').trigger('click')
    await flushPromises()

    const lastListCall = listAccounts.mock.calls[listAccounts.mock.calls.length - 1]
    expect(lastListCall[0]).toBe(1)
    expect(lastListCall[1]).toBe(100)
    expect(lastListCall[2]).toMatchObject({
      owner_filter: 'user',
      share_mode: 'public',
      group: ''
    })
    expect(wrapper.get('[data-test="bulk-edit-modal"]').attributes('data-show')).toBe('true')
    expect(wrapper.get('[data-test="bulk-edit-modal"]').attributes('data-target-mode')).toBe('filtered')
  })

  it('bulk approves all pending shared accounts from the current filters', async () => {
    routeName.value = 'AdminSharedAccounts'
    listAccounts.mockResolvedValueOnce({
      items: [
        {
          id: 1,
          name: 'Shared 1',
          platform: 'openai',
          type: 'oauth',
          status: 'active',
          schedulable: true,
          owner_user_id: 10,
          share_mode: 'public',
          share_status: 'pending_review',
          groups: [],
          credentials: { email: 'shared-1@example.com' },
          extra: {}
        },
        {
          id: 2,
          name: 'Shared 2',
          platform: 'openai',
          type: 'oauth',
          status: 'active',
          schedulable: true,
          owner_user_id: 11,
          share_mode: 'public',
          share_status: 'pending_review',
          groups: [],
          credentials: { email: 'shared-2@example.com' },
          extra: {}
        }
      ],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1
    }).mockResolvedValueOnce({
      items: [
        {
          id: 1,
          name: 'Shared 1',
          platform: 'openai',
          type: 'oauth',
          status: 'active',
          schedulable: true,
          owner_user_id: 10,
          share_mode: 'public',
          share_status: 'pending_review',
          groups: [],
          credentials: { email: 'shared-1@example.com' },
          extra: {}
        },
        {
          id: 2,
          name: 'Shared 2',
          platform: 'openai',
          type: 'oauth',
          status: 'active',
          schedulable: true,
          owner_user_id: 11,
          share_mode: 'public',
          share_status: 'pending_review',
          groups: [],
          credentials: { email: 'shared-2@example.com' },
          extra: {}
        }
      ],
      total: 2,
      page: 1,
      page_size: 5,
      pages: 1
    })
    batchSetShareStatus.mockResolvedValueOnce({
      success: 2,
      failed: 0,
      success_ids: [1, 2],
      failed_ids: [],
      results: [
        { account_id: 1, success: true },
        { account_id: 2, success: true }
      ]
    })

    const wrapper = mount(AccountsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: true,
          ConfirmDialog: ConfirmDialogStub,
          BaseDialog: BaseDialogStub,
          AccountTableActions: { template: '<div><slot name="beforeCreate" /><slot name="after" /></div>' },
          AccountTableFilters: { template: '<div></div>' },
          AccountBulkActionsBar: AccountBulkActionsBarStub,
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
          BulkEditAccountModal: BulkEditAccountModalStub,
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

    await flushPromises()
    expect(wrapper.get('[data-test="bulk-actions"]').attributes('data-selected-count')).toBe('0')

    await wrapper.get('[data-test="approve-filtered-share"]').trigger('click')
    await flushPromises()

    const previewCall = listAccounts.mock.calls[listAccounts.mock.calls.length - 1]
    expect(previewCall[0]).toBe(1)
    expect(previewCall[1]).toBe(5)
    expect(previewCall[2]).toMatchObject({
      owner_filter: 'user',
      share_mode: 'public',
      share_status: 'pending_review',
      group: ''
    })
    expect(batchSetShareStatus).not.toHaveBeenCalled()
    expect(wrapper.get('[data-test="confirm-dialog"]').attributes('data-message')).toBe(
      'admin.accounts.bulkActions.shareStatusConfirmMessage'
    )
    expect(wrapper.get('[data-test="confirm-dialog"]').text()).toContain('shared-1@example.com')

    await wrapper.get('[data-test="confirm-dialog-confirm"]').trigger('click')
    await flushPromises()

    expect(batchSetShareStatus).toHaveBeenCalledTimes(1)
    expect(batchSetShareStatus).toHaveBeenCalledWith({
      filters: expect.objectContaining({
        owner_filter: 'user',
        share_mode: 'public',
        share_status: 'pending_review',
        group: ''
      })
    }, 'active')
    expect(wrapper.find('[data-test="bulk-share-result-dialog"]').exists()).toBe(true)
    expect(appStore.showSuccess).toHaveBeenCalledWith(
      'admin.accounts.bulkActions.shareStatusSuccess'
    )
  })

  it('bulk approves only pending shared accounts and skips active accounts', async () => {
    routeName.value = 'AdminSharedAccounts'
    listAccounts.mockResolvedValueOnce({
      items: [
        {
          id: 1,
          name: 'Shared 1',
          platform: 'openai',
          type: 'oauth',
          status: 'active',
          schedulable: true,
          owner_user_id: 10,
          share_mode: 'public',
          share_status: 'pending_review',
          groups: [],
          credentials: {},
          extra: {}
        },
        {
          id: 2,
          name: 'Shared 2',
          platform: 'openai',
          type: 'oauth',
          status: 'active',
          schedulable: true,
          owner_user_id: 11,
          share_mode: 'public',
          share_status: 'active',
          groups: [],
          credentials: {},
          extra: {}
        }
      ],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1
    })
    batchSetShareStatus.mockResolvedValueOnce({
      success: 1,
      failed: 0,
      success_ids: [1],
      failed_ids: [],
      results: [{ account_id: 1, success: true }]
    })

    const wrapper = mount(AccountsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: true,
          ConfirmDialog: ConfirmDialogStub,
          BaseDialog: BaseDialogStub,
          AccountTableActions: { template: '<div><slot name="beforeCreate" /><slot name="after" /></div>' },
          AccountTableFilters: { template: '<div></div>' },
          AccountBulkActionsBar: AccountBulkActionsBarStub,
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
          BulkEditAccountModal: BulkEditAccountModalStub,
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

    await flushPromises()
    const checkboxes = wrapper.findAll('[data-test="select-cell"] input[type="checkbox"]')
    await checkboxes[0].trigger('change')
    await checkboxes[1].trigger('change')
    await flushPromises()

    expect(wrapper.get('[data-test="bulk-actions"]').attributes('data-selected-count')).toBe('2')
    await wrapper.get('[data-test="approve-share"]').trigger('click')
    await flushPromises()

    expect(setShareStatus).not.toHaveBeenCalled()
    expect(batchSetShareStatus).toHaveBeenCalledTimes(1)
    expect(batchSetShareStatus).toHaveBeenCalledWith([1], 'active')
    expect(appStore.showSuccess).toHaveBeenCalledWith(
      'admin.accounts.bulkActions.shareStatusSuccessWithSkipped'
    )
  })

  it('keeps failed shared account selected after a bulk review failure result', async () => {
    routeName.value = 'AdminSharedAccounts'
    listAccounts.mockResolvedValueOnce({
      items: [
        {
          id: 1,
          name: 'Shared 1',
          platform: 'openai',
          type: 'oauth',
          status: 'active',
          schedulable: true,
          owner_user_id: 10,
          share_mode: 'public',
          share_status: 'pending_review',
          groups: [],
          credentials: {},
          extra: {}
        },
        {
          id: 2,
          name: 'Shared 2',
          platform: 'openai',
          type: 'oauth',
          status: 'active',
          schedulable: true,
          owner_user_id: 11,
          share_mode: 'public',
          share_status: 'pending_review',
          groups: [],
          credentials: {},
          extra: {}
        }
      ],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1
    })
    batchSetShareStatus.mockResolvedValueOnce({
      success: 1,
      failed: 1,
      success_ids: [1],
      failed_ids: [2],
      results: [
        { account_id: 1, success: true },
        { account_id: 2, success: false, error: 'backend rejected' }
      ]
    }).mockResolvedValueOnce({
      success: 1,
      failed: 0,
      success_ids: [2],
      failed_ids: [],
      results: [
        { account_id: 2, success: true }
      ]
    })

    const wrapper = mount(AccountsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: true,
          ConfirmDialog: ConfirmDialogStub,
          BaseDialog: BaseDialogStub,
          AccountTableActions: { template: '<div><slot name="beforeCreate" /><slot name="after" /></div>' },
          AccountTableFilters: { template: '<div></div>' },
          AccountBulkActionsBar: AccountBulkActionsBarStub,
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
          BulkEditAccountModal: BulkEditAccountModalStub,
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

    await flushPromises()
    const checkboxes = wrapper.findAll('[data-test="select-cell"] input[type="checkbox"]')
    await checkboxes[0].trigger('change')
    await checkboxes[1].trigger('change')
    await flushPromises()

    await wrapper.get('[data-test="approve-share"]').trigger('click')
    await flushPromises()

    expect(batchSetShareStatus).toHaveBeenCalledWith([1, 2], 'active')
    expect(wrapper.get('[data-test="bulk-actions"]').attributes('data-selected-count')).toBe('1')
    expect(checkboxes[0].element.checked).toBe(false)
    expect(checkboxes[1].element.checked).toBe(true)
    expect(appStore.showError).toHaveBeenCalledWith(
      'admin.accounts.bulkActions.shareStatusPartial'
    )

    await wrapper
      .get('[data-test="bulk-share-result-footer"]')
      .findAll('button')
      .find(button => button.text() === 'admin.accounts.bulkActions.shareStatusResultCopyFailed')
      ?.trigger('click')
    await flushPromises()

    expect(writeText).toHaveBeenCalledTimes(1)
    expect(writeText).toHaveBeenCalledWith('2')
    expect(appStore.showSuccess).toHaveBeenCalledWith('common.copiedToClipboard')

    await wrapper
      .get('[data-test="bulk-share-result-footer"]')
      .findAll('button')
      .find(button => button.text() === 'admin.accounts.bulkActions.shareStatusResultRetryFailed')
      ?.trigger('click')
    await flushPromises()

    expect(batchSetShareStatus).toHaveBeenCalledTimes(2)
    expect(batchSetShareStatus).toHaveBeenLastCalledWith([2], 'active')
  })
})
