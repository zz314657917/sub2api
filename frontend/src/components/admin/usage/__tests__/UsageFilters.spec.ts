import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UsageFilters from '../UsageFilters.vue'

const {
  searchUsers,
  searchApiKeys,
  listGroups,
  getModelStats,
  listAccounts,
} = vi.hoisted(() => ({
  searchUsers: vi.fn(),
  searchApiKeys: vi.fn(),
  listGroups: vi.fn(),
  getModelStats: vi.fn(),
  listAccounts: vi.fn(),
}))

const messages: Record<string, string> = {
  'admin.usage.userFilter': 'User',
  'usage.apiKeyFilter': 'API Key',
  'usage.model': 'Model',
  'admin.usage.account': 'Account',
  'usage.type': 'Type',
  'admin.usage.billingType': 'Billing Type',
  'admin.usage.billingMode': 'Billing Mode',
  'admin.usage.group': 'Group',
  'admin.usage.searchUserPlaceholder': 'Search user by email...',
  'admin.usage.searchApiKeyPlaceholder': 'Search API key by name...',
  'admin.usage.searchAccountPlaceholder': 'Search account by name...',
  'admin.usage.filterByUserId': 'Filter by user ID',
  'admin.usage.allModels': 'All Models',
  'admin.usage.allGroups': 'All Groups',
  'admin.usage.allTypes': 'All Types',
  'admin.usage.allBillingTypes': 'All Billing Types',
  'admin.usage.billingTypeBalance': 'Credits',
  'admin.usage.billingTypeSubscription': 'Subscription',
  'admin.usage.allBillingModes': 'All Billing Modes',
  'admin.usage.billingModeToken': 'Token',
  'admin.usage.billingModePerRequest': 'Per Request',
  'admin.usage.billingModeImage': 'Image',
  'usage.ws': 'WS',
  'usage.stream': 'Stream',
  'usage.sync': 'Sync',
  'common.loading': 'Loading...',
  'common.noOptionsFound': 'No options found',
  'common.refresh': 'Refresh',
  'common.reset': 'Reset',
  'admin.usage.cleanup.button': 'Cleanup',
  'usage.exportExcel': 'Export',
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

vi.mock('@/api/admin', () => ({
  adminAPI: {
    usage: {
      searchUsers,
      searchApiKeys,
    },
    groups: {
      list: listGroups,
    },
    dashboard: {
      getModelStats,
    },
    accounts: {
      list: listAccounts,
    },
  },
}))

const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue', 'change'],
  template: '<button type="button" class="select-stub">{{ options?.[0]?.label }}</button>',
}

const mountFilters = () => mount(UsageFilters, {
  props: {
    modelValue: {
      user_id: undefined,
      api_key_id: undefined,
      account_id: undefined,
    },
    exporting: false,
    startDate: '2026-06-01',
    endDate: '2026-06-01',
  },
  global: {
    stubs: {
      Select: SelectStub,
    },
  },
})

describe('UsageFilters dropdown visibility', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    searchUsers.mockResolvedValue([])
    searchApiKeys.mockResolvedValue([])
    listGroups.mockResolvedValue({ items: [] })
    getModelStats.mockResolvedValue({ models: [] })
    listAccounts.mockResolvedValue({ items: [] })
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
    vi.clearAllMocks()
  })

  it('shows a selectable user-id dropdown item for numeric user input', async () => {
    const wrapper = mountFilters()
    await flushPromises()

    const userInput = wrapper.get('input[placeholder="Search user by email..."]')
    await userInput.setValue('3059214355')
    await userInput.trigger('focus')

    expect(wrapper.text()).toContain('Filter by user ID')
    expect(wrapper.text()).toContain('#3059214355')

    const userIdButton = wrapper.findAll('button')
      .find((button) => button.text().includes('Filter by user ID'))
    expect(userIdButton).toBeDefined()

    await userIdButton!.trigger('click')
    await flushPromises()

    expect(searchApiKeys).toHaveBeenCalledWith(3059214355, '')
    expect(wrapper.emitted('change')).toHaveLength(1)
    expect((wrapper.props('modelValue') as Record<string, unknown>).user_id).toBe(3059214355)
  })

  it('keeps the api-key dropdown visible when focused with no results', async () => {
    const wrapper = mountFilters()
    await flushPromises()

    const apiKeyInput = wrapper.get('input[placeholder="Search API key by name..."]')
    await apiKeyInput.trigger('focus')

    expect(wrapper.text()).toContain('No options found')

    await vi.advanceTimersByTimeAsync(300)
    await flushPromises()

    expect(searchApiKeys).toHaveBeenCalledWith(undefined, '')
    expect(wrapper.text()).toContain('No options found')
  })
})
