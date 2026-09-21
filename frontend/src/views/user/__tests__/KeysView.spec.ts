import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { nextTick } from 'vue'

import type { ApiKey } from '@/types'
import { keysAPI } from '@/api'
import KeysView from '../KeysView.vue'

const keysViewSource = readFileSync(resolve(process.cwd(), 'src/views/user/KeysView.vue'), 'utf8')

const {
  listKeys,
  getPublicSettings,
  getDashboardApiKeysUsage,
  getAvailableGroups,
  getUserGroupRates,
  showError,
  showSuccess,
  copyToClipboard,
  isCurrentStep,
  nextStep,
  routerReplace,
} = vi.hoisted(() => ({
  listKeys: vi.fn(),
  getPublicSettings: vi.fn(),
  getDashboardApiKeysUsage: vi.fn(),
  getAvailableGroups: vi.fn(),
  getUserGroupRates: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  copyToClipboard: vi.fn(),
  isCurrentStep: vi.fn(),
  nextStep: vi.fn(),
  routerReplace: vi.fn(),
}))

const messages: Record<string, string> = {
  'common.actions': 'Actions',
  'common.name': 'Name',
  'common.refresh': 'Refresh',
  'common.status': 'Status',
  'keys.apiKey': 'API Key',
  'keys.allGroups': 'All Groups',
  'keys.allStatus': 'All Status',
  'keys.columnSettings': 'Column Settings',
  'keys.createKey': 'Create API Key',
  'keys.created': 'Created',
  'keys.expiresAt': 'Expires',
  'keys.group': 'Group',
  'keys.currentConcurrency': 'Current Concurrency',
  'keys.lastUsedAt': 'Last Used',
  'keys.rateLimitColumn': 'Rate Limit',
  'keys.searchPlaceholder': 'Search name or key...',
  'keys.status.active': 'Active',
  'keys.status.expired': 'Expired',
  'keys.status.inactive': 'Inactive',
  'keys.status.quota_exhausted': 'Quota exhausted',
  'keys.usage': 'Usage',
}

vi.mock('@/api', () => ({
  keysAPI: {
    list: listKeys,
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    toggleStatus: vi.fn(),
  },
  authAPI: {
    getPublicSettings,
  },
  usageAPI: {
    getDashboardApiKeysUsage,
  },
  userGroupsAPI: {
    getAvailable: getAvailableGroups,
    getUserGroupRates,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({
    isCurrentStep,
    nextStep,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
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

vi.mock('vue-router', () => ({
  useRoute: () => ({
    path: '/keys',
    query: {},
  }),
  useRouter: () => ({
    replace: routerReplace,
  }),
}))

const createApiKey = (): ApiKey => ({
  id: 1,
  user_id: 1,
  key: 'sk-test-key',
  name: 'test-key',
  group_id: null,
  status: 'active',
  ip_whitelist: [],
  ip_blacklist: [],
  last_used_at: null,
  quota: 0,
  quota_used: 0,
  expires_at: null,
  created_at: '2026-06-27T00:00:00Z',
  updated_at: '2026-06-27T00:00:00Z',
  current_concurrency: 3,
  rate_limit_5h: 0,
  rate_limit_1d: 0,
  rate_limit_7d: 0,
  usage_5h: 0,
  usage_1d: 0,
  usage_7d: 0,
  window_5h_start: null,
  window_1d_start: null,
  window_7d_start: null,
  reset_5h_at: null,
  reset_1d_at: null,
  reset_7d_at: null,
})

const AppLayoutStub = {
  template: '<div><slot /></div>',
}

const TablePageLayoutStub = {
  template: `
    <div>
      <slot name="filters" />
      <slot name="actions" />
      <slot name="table" />
      <slot name="pagination" />
    </div>
  `,
}

const DataTableStub = {
  props: ['columns', 'data'],
  emits: ['sort'],
  template: `
    <div>
      <div data-test="columns">{{ columns.map((col) => col.key).join(',') }}</div>
      <div v-for="row in data" :key="row.id">
        <slot name="cell-name" :value="row.name" :row="row" />
        <div data-test="current-concurrency">
          <slot name="cell-current_concurrency" :value="row.current_concurrency" :row="row" />
        </div>
        <slot name="cell-actions" :row="row" />
      </div>
      <slot name="empty" />
    </div>
  `,
}

const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue'],
  template: '<select :value="modelValue" @change="$emit(\'update:modelValue\', $event.target.value)"></select>',
}

const SearchInputStub = {
  props: ['modelValue'],
  emits: ['update:modelValue', 'search'],
  template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
}

const IconStub = {
  props: ['name'],
  template: '<span data-test="icon">{{ name }}</span>',
}

const mountedWrappers: VueWrapper[] = []

const mountView = async () => {
  const wrapper = mount(KeysView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        TablePageLayout: TablePageLayoutStub,
        DataTable: DataTableStub,
        Pagination: true,
        BaseDialog: {
          props: ['show'],
          emits: ['close'],
          template: '<div v-if="show" role="dialog"><button data-test="close-dialog" @click="$emit(\'close\')">Close</button><slot /><slot name="footer" /></div>',
        },
        ConfirmDialog: true,
        EmptyState: true,
        Select: SelectStub,
        SearchInput: SearchInputStub,
        Icon: IconStub,
        UseKeyModal: true,
        EndpointPopover: true,
        GroupBadge: true,
        GroupOptionItem: true,
        Teleport: true,
      },
    },
  })
  await flushPromises()
  await nextTick()
  mountedWrappers.push(wrapper)
  return wrapper
}

afterEach(() => {
  mountedWrappers.splice(0).forEach((wrapper) => wrapper.unmount())
})

const visibleColumnKeys = (wrapper: VueWrapper) =>
  wrapper.get('[data-test="columns"]').text().split(',').filter(Boolean)

describe('user KeysView column settings', () => {
  beforeEach(() => {
    localStorage.clear()

    listKeys.mockReset()
    getPublicSettings.mockReset()
    getDashboardApiKeysUsage.mockReset()
    getAvailableGroups.mockReset()
    getUserGroupRates.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
    copyToClipboard.mockReset()
    isCurrentStep.mockReset()
    nextStep.mockReset()
    routerReplace.mockReset()
    vi.mocked(keysAPI.create).mockReset()
    vi.mocked(keysAPI.update).mockReset()

    listKeys.mockResolvedValue({
      items: [createApiKey()],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    getPublicSettings.mockResolvedValue({})
    getDashboardApiKeysUsage.mockResolvedValue({ stats: {} })
    getAvailableGroups.mockResolvedValue([])
    getUserGroupRates.mockResolvedValue({})
    isCurrentStep.mockReturnValue(false)
  })

  it('includes the current concurrency column', async () => {
    const wrapper = await mountView()

    expect(visibleColumnKeys(wrapper)).toContain('current_concurrency')
  })

  it('renders the current concurrency value', async () => {
    const wrapper = await mountView()

    expect(wrapper.get('[data-test="current-concurrency"]').text()).toBe('3')
  })
})

describe('user KeysView editor layout', () => {
  it('uses a full-width dialog with a bounded multi-group route editor in the desktop right column', () => {
    expect(keysViewSource).toContain('width="full"')
    expect(keysViewSource).toContain('<aside class="key-route-panel')
    expect(keysViewSource).toContain('lg:grid-cols-[minmax(0,1fr)_minmax(32rem,44rem)]')
    expect(keysViewSource).toContain('lg:col-start-2 lg:row-start-1 lg:row-span-3')
    expect(keysViewSource).toContain('lg:max-h-[calc(100vh-20rem)] lg:overflow-y-auto')
    expect(keysViewSource).toContain('key-route-heading flex min-w-0 items-center')
    expect(keysViewSource).toContain("truncate text-xs text-gray-500")
    expect(keysViewSource).toContain('key-route-list divide-y')
    expect(keysViewSource).toContain('key-route-row flex flex-wrap')
    expect(keysViewSource).toContain(':options="getRouteGroupOptions(route)"')
    expect(keysViewSource).toContain(':disabled="!canAddMultiGroupRoute"')
    expect(keysViewSource).not.toContain("t('keys.routeConfig')")
    expect(keysViewSource).not.toContain('routingPresetOptions')
    expect(keysViewSource).not.toContain('v-model.number="route.weight"')
    expect(keysViewSource).not.toContain('v-model.number="route.cooldown_seconds"')
    expect(keysViewSource).not.toContain('v-model="route.image_only"')
    expect(keysViewSource).not.toContain('v-model="route.text_only"')
  })
})

describe('user KeysView create provider selection', () => {
  const availableGroups = [
    { id: 1, name: 'Misleading GPT name', platform: 'anthropic' },
    { id: 2, name: 'Misleading Claude name', platform: 'openai' },
    { id: 3, name: 'Shared group', platform: 'deepseek' },
    { id: 4, name: 'Shared group', platform: 'gemini' },
    { id: 5, name: 'Unknown named group', platform: 'unknown-platform' },
  ]

  const formSelect = (wrapper: VueWrapper) => wrapper.findComponent('[data-tour="key-form-group"]')
  const optionIds = (wrapper: VueWrapper) => formSelect(wrapper).props('options').map((option: { value: number }) => option.value)
  const openCreate = async () => {
    const wrapper = await mountView()
    await wrapper.get('[data-tour="keys-create-btn"]').trigger('click')
    return wrapper
  }

  beforeEach(() => {
    localStorage.clear()
    listKeys.mockReset()
    getPublicSettings.mockReset()
    getDashboardApiKeysUsage.mockReset()
    getAvailableGroups.mockReset()
    getUserGroupRates.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
    copyToClipboard.mockReset()
    isCurrentStep.mockReset()
    nextStep.mockReset()
    routerReplace.mockReset()
    vi.mocked(keysAPI.create).mockReset()
    vi.mocked(keysAPI.update).mockReset()

    listKeys.mockResolvedValue({ items: [createApiKey()], total: 1, page: 1, page_size: 20, pages: 1 })
    getPublicSettings.mockResolvedValue({})
    getDashboardApiKeysUsage.mockResolvedValue({ stats: {} })
    getAvailableGroups.mockResolvedValue(availableGroups)
    getUserGroupRates.mockResolvedValue({})
    isCurrentStep.mockReturnValue(false)
  })

  it('filters only the create default group by platform, including unknown platforms', async () => {
    const wrapper = await openCreate()
    expect(wrapper.findAll('input[name="key-provider"]')).toHaveLength(4)
    expect(optionIds(wrapper)).toEqual([1])

    await wrapper.get('input[value="openai"]').setValue()
    expect(optionIds(wrapper)).toEqual([2])
    await wrapper.get('input[value="domestic"]').setValue()
    expect(optionIds(wrapper)).toEqual([3])
    await wrapper.get('input[value="other"]').setValue()
    expect(optionIds(wrapper)).toEqual([4, 5])
  })

  it('clears a stale default group when switching provider and does not put provider in create payload', async () => {
    const wrapper = await openCreate()
    await wrapper.get('[data-tour="key-form-name"]').setValue('My key')
    await formSelect(wrapper).vm.$emit('update:modelValue', 1)
    await wrapper.get('input[value="domestic"]').setValue()
    expect(formSelect(wrapper).props('modelValue')).toBeNull()

    await formSelect(wrapper).vm.$emit('update:modelValue', 3)
    vi.mocked(keysAPI.create).mockResolvedValue(createApiKey())
    await wrapper.get('#key-form').trigger('submit')
    await flushPromises()
    expect(vi.mocked(keysAPI.create)).toHaveBeenCalledWith(
      'My key', 3, undefined, [], [], 0, undefined,
      { rate_limit_5h: 0, rate_limit_1d: 0, rate_limit_7d: 0 }, [], 'shared_only', 0
    )
  })

  it('selects the first non-empty provider after delayed group loading and disables empty providers', async () => {
    let resolveGroups!: (groups: typeof availableGroups) => void
    getAvailableGroups.mockReturnValue(new Promise((resolve) => { resolveGroups = resolve }))
    const wrapper = await openCreate()
    expect(wrapper.get<HTMLInputElement>('input[value="anthropic"]').element.disabled).toBe(true)
    resolveGroups([availableGroups[1]])
    await flushPromises()
    expect(wrapper.get<HTMLInputElement>('input[value="openai"]').element.checked).toBe(true)
    expect(optionIds(wrapper)).toEqual([2])
  })

  it('resets the create provider and default group when the dialog reopens', async () => {
    const wrapper = await openCreate()
    await wrapper.get('input[value="domestic"]').setValue()
    await formSelect(wrapper).vm.$emit('update:modelValue', 3)
    await wrapper.get('[data-test="close-dialog"]').trigger('click')
    await wrapper.get('[data-tour="keys-create-btn"]').trigger('click')

    expect(wrapper.get<HTMLInputElement>('input[value="anthropic"]').element.checked).toBe(true)
    expect(formSelect(wrapper).props('modelValue')).toBeNull()
    expect(optionIds(wrapper)).toEqual([1])
  })

  it('keeps existing create routes and their first-response timeouts when the default provider changes', async () => {
    const wrapper = await openCreate()
    await wrapper.get('[data-tour="key-form-name"]').setValue('Routed key')
    await formSelect(wrapper).vm.$emit('update:modelValue', 1)
    await wrapper.get('button[aria-label="keys.multiGroupRouting"]').trigger('click')
    await nextTick()
    const setupFormData = (wrapper.vm as any).$?.setupState.formData
    expect(setupFormData.enable_multi_group_routing).toBe(true)
    expect(setupFormData.multi_group_routes).toHaveLength(1)
    const routeTimeout = wrapper.find('input[aria-label="keys.firstResponseTimeoutRouteSeconds"]')
    expect(routeTimeout.exists()).toBe(true)
    await routeTimeout.setValue('12')
    await wrapper.get('input[value="domestic"]').setValue()
    await formSelect(wrapper).vm.$emit('update:modelValue', 3)
    expect(setupFormData.enable_multi_group_routing).toBe(true)
    expect(setupFormData.multi_group_routes).toHaveLength(1)
    expect(setupFormData.multi_group_routes[0]).toMatchObject({ group_id: 1, first_response_timeout_seconds: 12 })

    vi.mocked(keysAPI.create).mockResolvedValue(createApiKey())
    await wrapper.get('#key-form').trigger('submit')
    await flushPromises()
    const createPayload = vi.mocked(keysAPI.create).mock.calls[0]
    expect(createPayload[1]).toBe(3)
    expect(createPayload[8]).toEqual([
      expect.objectContaining({ group_id: 1, priority: 1, first_response_timeout_seconds: 12 }),
    ])
  })

  it('keeps every authorized group and route timeout when editing', async () => {
    const key = {
      ...createApiKey(),
      group_id: 1,
      route_groups: availableGroups.slice(0, 4),
      multi_group_routes: [
        { group_id: 1, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true, first_response_timeout_seconds: 12 },
        { group_id: 4, priority: 2, weight: 1, cooldown_seconds: 30, enabled: true, first_response_timeout_seconds: 0 },
      ],
    }
    listKeys.mockResolvedValue({ items: [key], total: 1, page: 1, page_size: 20, pages: 1 })
    const wrapper = await mountView()
    ;(wrapper.vm as unknown as { editKey: (value: typeof key) => void }).editKey(key)
    await nextTick()
    expect(wrapper.find('[data-tour="key-form-provider"]').exists()).toBe(false)
    expect(optionIds(wrapper)).toEqual([1, 2, 3, 4, 5])
    await wrapper.get('#key-form').trigger('submit')
    await flushPromises()
    const [keyId, payload] = vi.mocked(keysAPI.update).mock.calls[0]
    expect(keyId).toBe(1)
    expect(payload).toMatchObject({ group_id: 1 })
    expect(payload.multi_group_routes).toEqual([
      expect.objectContaining({ group_id: 1, priority: 1, first_response_timeout_seconds: 12 }),
      expect.objectContaining({ group_id: 4, priority: 2 }),
    ])
  })
})
