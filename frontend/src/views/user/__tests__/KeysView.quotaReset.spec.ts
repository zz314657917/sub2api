import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import type { ApiKey } from '@/types'
import KeysView from '../KeysView.vue'

const { update, list, showError, showSuccess } = vi.hoisted(() => ({
  update: vi.fn(),
  list: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ path: '/keys', query: {} }),
  useRouter: () => ({ replace: vi.fn() }),
}))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key }),
}))
vi.mock('@/api', () => ({
  keysAPI: { list, update },
  authAPI: { getPublicSettings: vi.fn().mockResolvedValue({}) },
  usageAPI: { getDashboardApiKeysUsage: vi.fn().mockResolvedValue({ stats: {} }) },
  userGroupsAPI: {
    getAvailable: vi.fn().mockResolvedValue([]),
    getUserGroupRates: vi.fn().mockResolvedValue({}),
  },
}))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, showSuccess }) }))
vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({ isCurrentStep: () => false, nextStep: vi.fn() }),
}))
vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn() }),
}))
vi.mock('@/composables/usePersistedPageSize', () => ({ getPersistedPageSize: () => 10 }))

function key(id = 1, status: ApiKey['status'] = 'quota_exhausted'): ApiKey {
  return {
    id, name: `key-${id}`, status, quota: 100, quota_used: 100,
    group_id: 61, ip_whitelist: [], ip_blacklist: [],
    rate_limit_5h: 10, rate_limit_1d: 20, rate_limit_7d: 30,
    usage_5h: 7, usage_1d: 8, usage_7d: 9,
    multi_group_routes: [],
  } as unknown as ApiKey
}

type ViewState = {
  selectedKey: ApiKey | null
  formData: { status: string }
  editKey: (key: ApiKey) => void
  resetQuotaUsed: () => Promise<void>
  handleSubmit: () => Promise<void>
}

const wrappers: VueWrapper[] = []
async function mountedView() {
  const wrapper = mount(KeysView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /></div>' },
        DataTable: true,
        BaseDialog: {
          props: ['show'],
          template: '<section v-if="show"><slot /><slot name="footer" /></section>',
        },
        Select: true,
        VueDraggable: { template: '<div><slot /></div>' },
        SearchInput: true,
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
  wrappers.push(wrapper)
  await flushPromises()
  // Execute the actual mounted component handlers and refs, never a reimplemented
  // handler or source-string probe.
  const state = (wrapper.vm as unknown as { $: { setupState: ViewState } }).$.setupState
  state.editKey(key())
  await flushPromises()
  return state
}

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((done) => { resolve = done })
  return { promise, resolve }
}

describe('KeysView quota reset', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
    update.mockReset()
    list.mockResolvedValue({ items: [], total: 0, pages: 0 })
  })
  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount())
  })

  it('restores exhausted status and uses active in the subsequent edit save', async () => {
    const state = await mountedView()
    expect(state.formData.status).toBe('inactive')
    update.mockResolvedValue({ ...key(), quota_used: 0, status: 'active' })
    await state.resetQuotaUsed()
    expect(update).toHaveBeenNthCalledWith(1, 1, { reset_quota: true })
    expect(state.selectedKey?.status).toBe('active')
    expect(state.selectedKey?.quota_used).toBe(0)
    expect(state.formData.status).toBe('active')
    await state.handleSubmit()
    expect(update).toHaveBeenNthCalledWith(2, 1, expect.objectContaining({ status: 'active' }))
  })

  it('uses returned nonzero quota and leaves rate limit usage unchanged', async () => {
    const state = await mountedView()
    update.mockResolvedValue({ ...key(), quota_used: 2.75, status: 'active', usage_5h: 0 })
    await state.resetQuotaUsed()
    expect(state.selectedKey?.quota_used).toBe(2.75)
    expect(state.selectedKey?.usage_5h).toBe(7)
    expect(state.selectedKey?.usage_1d).toBe(8)
    expect(state.selectedKey?.usage_7d).toBe(9)
  })

  it('updates the captured target when selection changes during the request', async () => {
    const state = await mountedView()
    const target = state.selectedKey!
    const pending = deferred<ApiKey>()
    update.mockReturnValue(pending.promise)
    const request = state.resetQuotaUsed()
    state.editKey(key(2, 'inactive'))
    const selected = state.selectedKey!
    pending.resolve({ ...key(), quota_used: 3, status: 'active' })
    await request
    expect(target.quota_used).toBe(3)
    expect(target.status).toBe('active')
    expect(selected.quota_used).toBe(100)
    expect(selected.status).toBe('inactive')
    expect(state.formData.status).toBe('inactive')
    expect(update).toHaveBeenCalledWith(1, { reset_quota: true })
  })

  it('updates the target safely after selection is cleared', async () => {
    const state = await mountedView()
    const target = state.selectedKey!
    const pending = deferred<ApiKey>()
    update.mockReturnValue(pending.promise)
    const request = state.resetQuotaUsed()
    state.selectedKey = null
    pending.resolve({ ...key(), quota_used: 4, status: 'active' })
    await request
    expect(target.quota_used).toBe(4)
    expect(target.status).toBe('active')
    expect(state.selectedKey).toBeNull()
    expect(state.formData.status).toBe('inactive')
    expect(showError).not.toHaveBeenCalled()
  })

  it.each(['inactive', 'quota_exhausted', 'expired'] as const)(
    'preserves returned %s and keeps the editable form inactive',
    async (status) => {
      const state = await mountedView()
      update.mockResolvedValue({ ...key(), quota_used: 6, status })
      await state.resetQuotaUsed()
      expect(state.selectedKey?.status).toBe(status)
      expect(state.selectedKey?.quota_used).toBe(6)
      expect(state.formData.status).toBe('inactive')
    },
  )

  it('does not mutate the target or form when the API fails', async () => {
    const state = await mountedView()
    const original = { ...state.selectedKey }
    update.mockRejectedValue({ response: { data: { detail: 'fixture reset failure' } } })
    await state.resetQuotaUsed()
    expect(state.selectedKey).toEqual(original)
    expect(state.formData.status).toBe('inactive')
    expect(showError).toHaveBeenCalledWith('fixture reset failure')
    expect(showSuccess).not.toHaveBeenCalled()
  })

  it.each([
    ['active', 'inactive'],
    ['inactive', 'active'],
  ] as const)(
    'preserves the unsaved %s to %s choice when reset leaves server status unchanged',
    async (serverStatus, selectedStatus) => {
      const state = await mountedView()
      state.editKey(key(1, serverStatus))
      const target = state.selectedKey!
      state.formData.status = selectedStatus
      update.mockResolvedValue({ ...key(1, serverStatus), quota_used: 2.75 })

      await state.resetQuotaUsed()

      expect(update).toHaveBeenNthCalledWith(1, 1, { reset_quota: true })
      expect(target.quota_used).toBe(2.75)
      expect(target.status).toBe(serverStatus)
      expect(state.formData.status).toBe(selectedStatus)
      await state.handleSubmit()
      expect(update).toHaveBeenNthCalledWith(
        2, 1, expect.objectContaining({ status: selectedStatus }),
      )
    },
  )
})
