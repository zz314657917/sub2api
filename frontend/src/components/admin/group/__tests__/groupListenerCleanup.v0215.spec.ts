import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import type { AdminGroup } from '@/types'
import GroupRPMOverridesModal from '../GroupRPMOverridesModal.vue'
import GroupRateMultipliersModal from '../GroupRateMultipliersModal.vue'

const mocks = vi.hoisted(() => ({
  getGroupRPMOverrides: vi.fn().mockResolvedValue([]),
  getGroupRateMultipliers: vi.fn().mockResolvedValue([]),
  list: vi.fn(),
}))
vi.mock('@/api/admin', () => ({ adminAPI: { groups: mocks, users: { list: mocks.list } } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn() }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
beforeEach(() => {
  mocks.getGroupRPMOverrides.mockResolvedValue([])
  mocks.getGroupRateMultipliers.mockResolvedValue([])
})
afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks(); vi.clearAllMocks() })

describe('v0215 copied baseline group listener and debounce cleanup', () => {
  it.each([
    ['RPM', GroupRPMOverridesModal],
    ['rate multiplier', GroupRateMultipliersModal],
  ] as const)('removes the %s listener and cancels pending search after unmount', async (_, component) => {
    vi.useFakeTimers()
    const add = vi.spyOn(document, 'addEventListener')
    const remove = vi.spyOn(document, 'removeEventListener')
    const wrapper = mount(component, {
      props: { show: false, group: { id: 1, name: 'Group', platform: 'openai' } as AdminGroup },
      global: { stubs: {
        BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' },
        Icon: true, PlatformIcon: true, Pagination: true,
      } },
    })
    await wrapper.setProps({ show: true })
    await flushPromises()
    const listener = add.mock.calls.find(([type]) => type === 'click')?.[1]
    expect(listener).toBeTypeOf('function')
    await wrapper.get('input[placeholder="admin.groups.searchUserPlaceholder"]').setValue('user')
    wrapper.unmount()
    expect(remove).toHaveBeenCalledWith('click', listener)
    await vi.advanceTimersByTimeAsync(500)
    expect(mocks.list).not.toHaveBeenCalled()
  })
})
