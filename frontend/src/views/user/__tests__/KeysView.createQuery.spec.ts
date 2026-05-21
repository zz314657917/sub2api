import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import KeysView from '../KeysView.vue'

const routeState = vi.hoisted(() => ({
  path: '/keys',
  query: {} as Record<string, unknown>,
}))
const routerReplace = vi.hoisted(() => vi.fn())
const keysList = vi.hoisted(() => vi.fn())

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
})
