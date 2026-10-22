import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import type { AdminGroup } from '@/types'
import GroupRPMOverridesModal from '../GroupRPMOverridesModal.vue'

const mocks = vi.hoisted(() => ({
  getGroupRPMOverrides: vi.fn(), batchSetGroupRPMOverrides: vi.fn(), showError: vi.fn(),
}))
vi.mock('@/api/admin', () => ({ adminAPI: { groups: mocks } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: mocks.showError, showSuccess: vi.fn() }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
beforeEach(() => vi.resetAllMocks())
afterEach(() => vi.restoreAllMocks())

function deferred() {
  let resolve!: (value: unknown) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}
const group = (id: number) => ({ id, name: `Group ${id}`, platform: 'openai' } as AdminGroup)
const entries = (id: number) => [{
  user_id: id, user_name: '', user_email: `user${id}@example.com`, user_status: 'active', rpm_override: 100,
}]

function openGroup() {
  return mount(GroupRPMOverridesModal, {
    props: { show: true, group: group(1) },
    global: { stubs: {
      BaseDialog: { props: ['show'], emits: ['close'], template: '<div v-if="show"><button class="close" @click="$emit(\'close\')" /><slot /></div>' },
      Icon: true, PlatformIcon: true, Pagination: true,
    } },
  })
}

describe('RPM request ownership', () => {
  it('loads an initially visible group', async () => {
    mocks.getGroupRPMOverrides.mockResolvedValue(entries(1))
    const wrapper = openGroup()
    await flushPromises()
    expect(mocks.getGroupRPMOverrides).toHaveBeenCalledTimes(1)
    expect(mocks.getGroupRPMOverrides).toHaveBeenCalledWith(1)
    expect(wrapper.get('tbody').text()).toContain('user1@example.com')
  })

  it('preserves the new group draft when an older success finishes last', async () => {
    const old = deferred()
    mocks.getGroupRPMOverrides.mockReturnValueOnce(old.promise).mockResolvedValueOnce(entries(2))
    const wrapper = openGroup()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true, group: group(2) })
    await flushPromises()
    await wrapper.get('tbody input[type="number"]').setValue('0')
    old.resolve(entries(1))
    await flushPromises()
    expect(wrapper.get('tbody').text()).toContain('user2@example.com')
    expect(wrapper.get('tbody').text()).not.toContain('user1@example.com')
    await wrapper.findAll('button').find(button => button.text() === 'common.save')!.trigger('click')
    expect(mocks.batchSetGroupRPMOverrides).toHaveBeenCalledWith(2, [{ user_id: 2, rpm_override: 0 }])
  })

  it.each(['success', 'error'])('old %s cannot clear current loading or report an error', async (outcome) => {
    const old = deferred()
    const current = deferred()
    mocks.getGroupRPMOverrides.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const wrapper = openGroup()
    await wrapper.setProps({ group: group(2) })
    if (outcome === 'success') old.resolve(entries(1))
    else old.reject(new Error('obsolete'))
    await flushPromises()
    expect(mocks.showError).not.toHaveBeenCalled()
    expect(wrapper.find('.animate-spin').exists()).toBe(true)
    current.resolve(entries(2))
    await flushPromises()
    expect(wrapper.get('tbody').text()).toContain('user2@example.com')
    expect(wrapper.find('.animate-spin').exists()).toBe(false)
  })

  it('clears the previous draft on a group-ID transition without closing', async () => {
    mocks.getGroupRPMOverrides.mockResolvedValueOnce(entries(1)).mockReturnValueOnce(deferred().promise)
    const wrapper = openGroup()
    await flushPromises()
    await wrapper.get('tbody input[type="number"]').setValue('200')
    await wrapper.setProps({ group: group(2) })
    expect(mocks.getGroupRPMOverrides).toHaveBeenLastCalledWith(2)
    expect(wrapper.findAll('button').some(button => button.text() === 'common.save')).toBe(false)
  })

  it.each(['hide', 'close', 'unmount', 'null group'])('invalidates pending errors on %s', async (action) => {
    const request = deferred()
    mocks.getGroupRPMOverrides.mockReturnValue(request.promise)
    const wrapper = openGroup()
    if (action === 'hide') await wrapper.setProps({ show: false })
    if (action === 'close') await wrapper.get('.close').trigger('click')
    if (action === 'unmount') wrapper.unmount()
    if (action === 'null group') await wrapper.setProps({ group: null })
    request.reject(new Error('offline'))
    await flushPromises()
    expect(mocks.showError).not.toHaveBeenCalled()
  })

  it('reports a current error and ends loading', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    mocks.getGroupRPMOverrides.mockRejectedValue(new Error('offline'))
    const wrapper = openGroup()
    await flushPromises()
    expect(mocks.showError).toHaveBeenCalledTimes(1)
    expect(mocks.showError).toHaveBeenCalledWith('admin.groups.failedToLoad')
    expect(wrapper.find('.animate-spin').exists()).toBe(false)
  })
})
