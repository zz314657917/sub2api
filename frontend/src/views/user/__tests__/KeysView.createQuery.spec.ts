import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import KeysView from '../KeysView.vue'

const routeState = vi.hoisted(() => ({
  path: '/keys',
  query: {} as Record<string, unknown>,
}))
const routerReplace = vi.hoisted(() => vi.fn())
const keysList = vi.hoisted(() => vi.fn())
const keysCreate = vi.hoisted(() => vi.fn())
const keysUpdate = vi.hoisted(() => vi.fn())
const availableGroups = vi.hoisted(() => vi.fn())
const userGroupRates = vi.hoisted(() => vi.fn())
const getPublicSettings = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({
      replace: routerReplace,
    }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/api', () => ({
  keysAPI: {
    list: keysList,
    create: keysCreate,
    update: keysUpdate,
  },
  authAPI: {
    getPublicSettings,
  },
  usageAPI: {
    getDashboardApiKeysUsage: vi.fn().mockResolvedValue({ stats: {} }),
  },
  userGroupsAPI: {
    getAvailable: availableGroups,
    getUserGroupRates: userGroupRates,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({
    isCurrentStep: vi.fn(() => false),
    nextStep: vi.fn(),
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true),
  }),
}))

vi.mock('@/composables/usePersistedPageSize', () => ({
  getPersistedPageSize: () => 10,
}))

function mountView() {
  return shallowMount(KeysView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /></div>' },
        DataTable: { template: '<div><slot name="empty" /></div>' },
        BaseDialog: {
          props: ['show', 'title'],
          template: '<section v-if="show" class="base-dialog">{{ title }}<slot /></section>',
        },
        Select: {
          props: ['modelValue', 'options'],
          emits: ['update:modelValue', 'change'],
          template: '<select @change="$emit(`update:modelValue`, null)" />',
        },
        VueDraggable: { template: '<div><slot /></div>' },
        SearchInput: { template: '<input />' },
        Pagination: true,
        ConfirmDialog: true,
        EmptyState: true,
        Icon: true,
        UseKeyModal: true,
        EndpointPopover: true,
        GroupBadge: true,
        GroupOptionItem: true,
        Teleport: true,
      },
    },
  })
}

describe('KeysView create query', () => {
  beforeEach(() => {
    routeState.path = '/keys'
    routeState.query = {}
    routerReplace.mockReset()
    keysCreate.mockReset().mockResolvedValue({ id: 1 })
    keysUpdate.mockReset().mockResolvedValue({ id: 1 })
    keysList.mockReset().mockResolvedValue({
      items: [],
      total: 0,
      pages: 0,
    })
    availableGroups.mockReset().mockResolvedValue([])
    userGroupRates.mockReset().mockResolvedValue({})
    getPublicSettings.mockReset().mockResolvedValue({ account_share_enabled: true })
  })

  it.each([
    ['disabled', { account_share_enabled: false }, false],
    ['enabled', { account_share_enabled: true }, true],
    ['missing for backwards compatibility', {}, true],
  ])('controls account-pool strategy visibility when sharing is %s', async (_case, settings, expected) => {
    getPublicSettings.mockResolvedValue(settings)

    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    await flushPromises()

    expect(wrapper.find('[data-testid="account-pool-strategy"]').exists()).toBe(expected)
  })

  it('opens the create key dialog when create=1 is present', async () => {
    routeState.query = { create: '1' }

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('keys.createKey')
    expect(routerReplace).toHaveBeenCalledWith({
      path: '/keys',
      query: { create: undefined },
    })
  })

  it('derives route priority from drag order and resets legacy route knobs', async () => {
    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.group_id = 1
    setupState.formData.enable_multi_group_routing = true
    setupState.formData.multi_group_routes = [
      routeFormFixture('route-b', 2, { priority: 100, weight: 2 }),
      routeFormFixture('route-a', 1, { priority: 100, weight: 1 }),
      routeFormFixture('route-c', 3, { priority: 100, weight: 3 }),
    ]
    setupState.handleRouteOrderChanged()

    await setupState.handleSubmit()

    const routes = keysCreate.mock.calls[0][8]
    expect(routes).toEqual([
      { group_id: 2, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true },
      { group_id: 1, priority: 2, weight: 1, cooldown_seconds: 30, enabled: true },
      { group_id: 3, priority: 3, weight: 1, cooldown_seconds: 30, enabled: true },
    ])
  })

  it('loads legacy routes in priority order and drops user-managed scope/model fields', async () => {
    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.editKey({
      id: 11,
      name: 'Legacy Key',
      group_id: 1,
      multi_group_routes: [
        routeFormFixture('late', 3, {
          priority: 20,
          weight: 9,
          cooldown_seconds: 1,
          image_only: true,
          model_patterns: ['old-*'],
        }),
        routeFormFixture('early', 2, {
          priority: 2,
          weight: 4,
          cooldown_seconds: 90,
          text_only: true,
          enabled: false,
        }),
      ],
      account_pool_strategy: 'shared_only',
      status: 'active',
      ip_whitelist: [],
      ip_blacklist: [],
      quota: 0,
      quota_used: 0,
      rate_limit_5h: 0,
      rate_limit_1d: 0,
      rate_limit_7d: 0,
      expires_at: null,
    })

    await setupState.handleSubmit()

    expect(keysUpdate.mock.calls[0][1].multi_group_routes).toEqual([
      { group_id: 2, priority: 1, weight: 1, cooldown_seconds: 30, enabled: false },
      { group_id: 3, priority: 2, weight: 1, cooldown_seconds: 30, enabled: true },
    ])
  })

  it('rejects duplicate groups in the user route list', async () => {
    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.group_id = 1
    setupState.formData.enable_multi_group_routing = true
    setupState.formData.multi_group_routes = [
      routeFormFixture('duplicate-a', 2),
      routeFormFixture('duplicate-b', 2),
    ]

    await setupState.handleSubmit()

    expect(keysCreate).not.toHaveBeenCalled()
  })

  it('keeps priorities continuous when adding and removing routes', async () => {
    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.group_id = 1
    setupState.formData.enable_multi_group_routing = true
    setupState.formData.multi_group_routes = [routeFormFixture('first', 1)]
    setupState.addMultiGroupRoute()
    expect(setupState.formData.multi_group_routes.map((route: { priority: number }) => route.priority)).toEqual([1, 2])
    setupState.removeMultiGroupRoute(0)
    expect(setupState.formData.multi_group_routes.map((route: { priority: number }) => route.priority)).toEqual([1])
  })

  it('preserves enabled state while normalizing all compatibility fields', async () => {
    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.group_id = 1
    setupState.formData.enable_multi_group_routing = true
    setupState.formData.multi_group_routes = [
      routeFormFixture('disabled', 1, {
        weight: 99,
        cooldown_seconds: 999,
        enabled: false,
        image_only: true,
        text_only: true,
        model_patterns: ['should-not-send'],
      }),
    ]

    await setupState.handleSubmit()

    expect(keysCreate.mock.calls[0][8]).toEqual([
      { group_id: 1, priority: 1, weight: 1, cooldown_seconds: 30, enabled: false },
    ])
  })

  it('does not include legacy model patterns in any route payload', async () => {
    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.group_id = 1
    setupState.formData.enable_multi_group_routing = true
    setupState.formData.multi_group_routes = [routeFormFixture('pattern', 1, {
      model_patterns: ['gpt-*'],
    })]

    await setupState.handleSubmit()

    expect(keysCreate.mock.calls[0][8][0]).not.toHaveProperty('model_patterns')
  })

  it('keeps route order stable when legacy priorities tie', async () => {
    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.editKey({
      id: 12,
      name: 'Stable Key',
      group_id: 1,
      multi_group_routes: [
        { group_id: 3, priority: 1, weight: 8, cooldown_seconds: 8, enabled: true },
        { group_id: 2, priority: 1, weight: 2, cooldown_seconds: 8, enabled: true },
      ],
      account_pool_strategy: 'shared_only',
      status: 'active',
      ip_whitelist: [],
      ip_blacklist: [],
      quota: 0,
      quota_used: 0,
      rate_limit_5h: 0,
      rate_limit_1d: 0,
      rate_limit_7d: 0,
      expires_at: null,
    })

    await setupState.handleSubmit()

    expect(keysUpdate.mock.calls[0][1].multi_group_routes.map((route: { group_id: number }) => route.group_id)).toEqual([3, 2])
    expect(keysUpdate.mock.calls[0][1].multi_group_routes.every((route: { weight: number; cooldown_seconds: number }, index: number) => (
      route.weight === 1 && route.cooldown_seconds === 30 && route.priority === index + 1
    ))).toBe(true)
  })

  it('does not overwrite a touched default group when rates load after the dialog opens', async () => {
    let resolveRates!: (rates: Record<number, number>) => void
    userGroupRates.mockReturnValue(new Promise((resolve) => {
      resolveRates = resolve
    }))
    availableGroups.mockResolvedValue([
      groupFixture(1, 'A', 1),
      groupFixture(2, 'B', 2),
    ])

    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.handleDefaultGroupChanged(2)

    resolveRates({ 1: 0.5, 2: 2 })
    await flushPromises()

    expect(setupState.formData.group_id).toBe(2)
  })

  it('preserves existing routes and hidden account-pool strategy when editing and saving', async () => {
    getPublicSettings.mockResolvedValue({ account_share_enabled: false })

    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.editKey({
      id: 10,
      name: 'Weighted Key',
      key: 'sk-test',
      group_id: 1,
      group: null,
      route_groups: [],
      multi_group_routes: [
        { group_id: 1, priority: 10, weight: 3, cooldown_seconds: 30, enabled: true },
        { group_id: 2, priority: 10, weight: 1, cooldown_seconds: 30, enabled: true },
      ],
      account_pool_strategy: 'private_only',
      status: 'active',
      ip_whitelist: [],
      ip_blacklist: [],
      quota: 0,
      quota_used: 0,
      rate_limit_5h: 0,
      rate_limit_1d: 0,
      rate_limit_7d: 0,
      reset_5h_at: null,
      reset_1d_at: null,
      reset_7d_at: null,
      expires_at: null,
      created_at: '',
      updated_at: '',
      last_used_at: null,
      usage_count: 0,
    })

    await setupState.handleSubmit()

    expect(keysUpdate.mock.calls[0][1].multi_group_routes).toEqual([
      { group_id: 1, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true },
      { group_id: 2, priority: 2, weight: 1, cooldown_seconds: 30, enabled: true },
    ])
    expect(keysUpdate.mock.calls[0][1].account_pool_strategy).toBe('private_only')
  })
})

function routeFormFixture(
  clientId: string,
  groupId: number,
  overrides: Partial<{
    priority: number
    weight: number
    cooldown_seconds: number
    enabled: boolean
    image_only: boolean
    text_only: boolean
    model_patterns: string[]
  }> = {},
) {
  return {
    client_id: clientId,
    group_id: groupId,
    priority: overrides.priority ?? 1,
    weight: overrides.weight ?? 1,
    cooldown_seconds: overrides.cooldown_seconds ?? 30,
    enabled: overrides.enabled ?? true,
    image_only: overrides.image_only ?? false,
    text_only: overrides.text_only ?? false,
    model_patterns: overrides.model_patterns ?? [],
  }
}

function groupFixture(id: number, name: string, rate: number, overrides: Record<string, unknown> = {}) {
  return {
    id,
    name,
    description: null,
    platform: 'openai',
    rate_multiplier: rate,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'standard',
    routing_scope: 'inference',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    allow_image_generation: true,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    peak_rate_enabled: false,
    peak_start: '',
    peak_end: '',
    peak_rate_multiplier: 1,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    require_oauth_only: false,
    require_privacy_set: false,
    created_at: '',
    updated_at: '',
    ...overrides,
  }
}
