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
  },
  authAPI: {
    getPublicSettings: vi.fn().mockResolvedValue({}),
  },
  usageAPI: {
    getDashboardApiKeysUsage: vi.fn().mockResolvedValue({ stats: {} }),
  },
  userGroupsAPI: {
    getAvailable: vi.fn().mockResolvedValue([]),
    getUserGroupRates: vi.fn().mockResolvedValue({}),
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
    keysList.mockReset().mockResolvedValue({
      items: [],
      total: 0,
      pages: 0,
    })
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

  it('submits multi-group route priorities from the current row order', async () => {
    const wrapper = mountView()
    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    setupState.openCreateModal()
    setupState.formData.group_id = 1
    setupState.formData.enable_multi_group_routing = true
    setupState.formData.multi_group_routes = [
      { client_id: 'route-b', group_id: 2, priority: 100, weight: 2, cooldown_seconds: 30, enabled: true },
      { client_id: 'route-a', group_id: 1, priority: 100, weight: 1, cooldown_seconds: 30, enabled: true },
      { client_id: 'route-c', group_id: 3, priority: 100, weight: 3, cooldown_seconds: 30, enabled: true },
    ]

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
})
