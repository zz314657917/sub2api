import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, shallowMount } from '@vue/test-utils'
import type { Account } from '@/types'
import EditAccountModal from '../EditAccountModal.vue'

const mocks = vi.hoisted(() => ({ sync: vi.fn(), showError: vi.fn() }))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: mocks.showError, showInfo: vi.fn(), showSuccess: vi.fn() }),
}))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ isSimpleMode: true }) }))
vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: { syncUpstreamModels: mocks.sync },
    settings: {
      getWebSearchEmulationConfig: async () => ({ enabled: false, providers: [] }),
      getSettings: async () => ({}),
    },
    tlsFingerprintProfiles: { list: async () => [] },
  },
}))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({
    t: (key: string, params?: { message?: string }) => params?.message ? `${key}: ${params.message}` : key,
  }),
}))
enableAutoUnmount(afterEach)
beforeEach(() => vi.resetAllMocks())

describe('EditAccountModal upstream sync errors', () => {
  it.each([
    [{ status: 502, message: 'upstream unavailable' }, 'upstream unavailable'],
    [new Error('network failure'), 'network failure'],
    [{ status: 502 }, 'admin.accounts.syncUpstreamModelsFailed'],
  ])('extracts the error and ends syncing: %j', async (error, message) => {
    mocks.sync.mockRejectedValue(error)
    const wrapper = shallowMount(EditAccountModal, {
      props: {
        show: true,
        account: {
          id: 42, name: 'Antigravity', platform: 'antigravity', type: 'oauth',
          credentials: {}, extra: {}, group_ids: [], status: 'active', proxy_id: null,
        } as unknown as Account,
        proxies: [], groups: [],
      },
      global: { stubs: {
        BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
      } },
    })
    await flushPromises()
    const button = wrapper.findAll('button').find(button => button.text() === 'admin.accounts.syncUpstreamModels')!
    expect(button).toBeDefined()
    await button.trigger('click')
    await flushPromises()
    expect(mocks.sync).toHaveBeenCalledWith(42)
    expect(mocks.showError).toHaveBeenCalledWith(`admin.accounts.syncUpstreamModelsError: ${message}`)
    expect(button.attributes('disabled')).toBeUndefined()
  })
})
