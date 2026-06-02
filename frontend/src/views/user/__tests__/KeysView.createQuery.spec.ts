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
    getPublicSettings: vi.fn().mockResolvedValue({}),
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

  it('renumbers multi-group route priorities after route order changes', async () => {
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
    expect(routes.map((route: { group_id: number; priority: number }) => ({
      group_id: route.group_id,
      priority: route.priority,
    }))).toEqual([
      { group_id: 2, priority: 1 },
      { group_id: 1, priority: 2 },
      { group_id: 3, priority: 3 },
    ])
  })

  it('auto preset enables multi-group routing and generates routes from available groups', async () => {
    availableGroups.mockResolvedValue([
      groupFixture(1, 'OpenAI', 1),
      groupFixture(2, 'Gemini', 1.2),
    ])

    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.name = 'Auto Key'

    await setupState.handleSubmit()

    const routes = keysCreate.mock.calls[0][8]
    expect(routes.map((route: { group_id: number; priority: number; weight: number }) => ({
      group_id: route.group_id,
      priority: route.priority,
      weight: route.weight,
      text_only: route.text_only,
      image_only: route.image_only,
    }))).toEqual([
      { group_id: 1, priority: 1, weight: 1, text_only: true, image_only: undefined },
      { group_id: 2, priority: 1, weight: 1, text_only: true, image_only: undefined },
      { group_id: 1, priority: 1, weight: 1, text_only: undefined, image_only: true },
      { group_id: 2, priority: 1, weight: 1, text_only: undefined, image_only: true },
    ])
  })

  it('speed preset keeps routes in one priority pool and applies descending weights', async () => {
    availableGroups.mockResolvedValue([
      groupFixture(1, 'A', 1),
      groupFixture(2, 'B', 1),
      groupFixture(3, 'C', 1),
    ])

    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.name = 'Speed Key'
    setupState.applyRoutingPreset('speed')

    await setupState.handleSubmit()

    const routes = keysCreate.mock.calls[0][8]
    expect(routes.map((route: { group_id: number; priority: number; weight: number; cooldown_seconds: number; text_only?: boolean; image_only?: boolean }) => ({
      group_id: route.group_id,
      priority: route.priority,
      weight: route.weight,
      cooldown_seconds: route.cooldown_seconds,
      text_only: route.text_only,
      image_only: route.image_only,
    }))).toEqual([
      { group_id: 1, priority: 1, weight: 3, cooldown_seconds: 15, text_only: true, image_only: undefined },
      { group_id: 2, priority: 1, weight: 2, cooldown_seconds: 15, text_only: true, image_only: undefined },
      { group_id: 3, priority: 1, weight: 1, cooldown_seconds: 15, text_only: true, image_only: undefined },
      { group_id: 1, priority: 1, weight: 3, cooldown_seconds: 15, text_only: undefined, image_only: true },
      { group_id: 2, priority: 1, weight: 2, cooldown_seconds: 15, text_only: undefined, image_only: true },
      { group_id: 3, priority: 1, weight: 1, cooldown_seconds: 15, text_only: undefined, image_only: true },
    ])
  })

  it('cost preset orders routes by effective user multiplier', async () => {
    availableGroups.mockResolvedValue([
      groupFixture(1, 'Standard', 1.2),
      groupFixture(2, 'Discount', 1),
      groupFixture(3, 'User Discount', 1.5),
    ])
    userGroupRates.mockResolvedValue({ 3: 0.6 })

    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.name = 'Cost Key'
    setupState.applyRoutingPreset('cost')

    await setupState.handleSubmit()

    const routes = keysCreate.mock.calls[0][8]
    expect(routes
      .filter((route: { text_only?: boolean }) => route.text_only)
      .map((route: { group_id: number; priority: number }) => ({
      group_id: route.group_id,
      priority: route.priority,
    }))).toEqual([
      { group_id: 3, priority: 1 },
      { group_id: 2, priority: 2 },
      { group_id: 1, priority: 3 },
    ])
  })

  it('manual routes are not overwritten after group/rate reload sync', async () => {
    availableGroups.mockResolvedValue([
      groupFixture(1, 'A', 1),
      groupFixture(2, 'B', 2),
    ])

    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.name = 'Manual Key'
    setupState.formData.group_id = 2
    setupState.formData.enable_multi_group_routing = true
    setupState.formData.multi_group_routes = [
      routeFormFixture('manual', 2, { priority: 1, weight: 7, cooldown_seconds: 44 }),
    ]
    setupState.markRoutingManual()
    setupState.syncCurrentRoutingPreset()

    await setupState.handleSubmit()

    const routes = keysCreate.mock.calls[0][8]
    expect(routes).toEqual([
      { group_id: 2, priority: 1, weight: 7, cooldown_seconds: 44, enabled: true },
    ])
  })

  it('preserves manual model patterns and route scope flags in payload', async () => {
    availableGroups.mockResolvedValue([
      groupFixture(1, 'Image', 1),
    ])

    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.name = 'Pattern Key'
    setupState.formData.group_id = 1
    setupState.formData.enable_multi_group_routing = true
    setupState.formData.multi_group_routes = [
      routeFormFixture('pattern', 1, {
        model_patterns_text: 'gpt-image-*\nclaude-*',
        image_only: true,
      }),
    ]
    setupState.markRoutingManual()

    await setupState.handleSubmit()

    const routes = keysCreate.mock.calls[0][8]
    expect(routes).toEqual([
      {
        group_id: 1,
        priority: 1,
        weight: 1,
        cooldown_seconds: 30,
        enabled: true,
        model_patterns: ['gpt-image-*', 'claude-*'],
        image_only: true,
      },
    ])
  })

  it('cost preset orders image routes by independent image multiplier', async () => {
    availableGroups.mockResolvedValue([
      groupFixture(1, 'Text Cheap Image Expensive', 0.5, { image_rate_independent: true, image_rate_multiplier: 2 }),
      groupFixture(2, 'Text Expensive Image Cheap', 2, { image_rate_independent: true, image_rate_multiplier: 0.3 }),
      groupFixture(3, 'Shared Middle', 1),
    ])

    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.name = 'Image Cost Key'
    setupState.applyRoutingPreset('cost')

    await setupState.handleSubmit()

    const routes = keysCreate.mock.calls[0][8]
    const imageRoutes = routes.filter((route: { image_only?: boolean }) => route.image_only)
    expect(imageRoutes.map((route: { group_id: number; priority: number }) => ({
      group_id: route.group_id,
      priority: route.priority,
    }))).toEqual([
      { group_id: 2, priority: 1 },
      { group_id: 3, priority: 2 },
      { group_id: 1, priority: 3 },
    ])
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

  it('preserves existing equal-priority weighted routes when editing and saving', async () => {
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
      account_pool_strategy: 'shared_only',
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
      { group_id: 1, priority: 10, weight: 3, cooldown_seconds: 30, enabled: true },
      { group_id: 2, priority: 10, weight: 1, cooldown_seconds: 30, enabled: true },
    ])
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
    model_patterns_text: string
    image_only: boolean
    text_only: boolean
  }> = {},
) {
  return {
    client_id: clientId,
    group_id: groupId,
    priority: overrides.priority ?? 1,
    weight: overrides.weight ?? 1,
    cooldown_seconds: overrides.cooldown_seconds ?? 30,
    enabled: overrides.enabled ?? true,
    model_patterns_text: overrides.model_patterns_text ?? '',
    image_only: overrides.image_only ?? false,
    text_only: overrides.text_only ?? false,
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
    created_at: '',
    updated_at: '',
    ...overrides,
  }
}
