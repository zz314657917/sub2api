import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import CafeRoomAccountPicker from '../CafeRoomAccountPicker.vue'

const { listAccountOptions } = vi.hoisted(() => ({ listAccountOptions: vi.fn() }))

vi.mock('@/api/admin/cafeRooms', () => ({
  default: { listAccountOptions },
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((resolvePromise) => { resolve = resolvePromise })
  return { promise, resolve }
}

describe('CafeRoomAccountPicker', () => {
  it('loads only plan-compatible candidates and emits a keyboard-native single selection', async () => {
    listAccountOptions.mockReset()
    listAccountOptions.mockResolvedValue({ data: {
      items: [{ id: 41, name: 'OpenAI account', platform: 'openai', status: 'active', email_masked: 'o***i@example.com' }],
      total: 1, page: 1, page_size: 20, pages: 1,
    } })
    const wrapper = mount(CafeRoomAccountPicker, { props: { modelValue: 0, planId: 21, excludeRoomId: 7, active: true } })
    await flushPromises()

    expect(listAccountOptions).toHaveBeenCalledWith(expect.objectContaining({ plan_id: 21, exclude_room_id: 7, page: 1, page_size: 20 }))
    await wrapper.find('input[type="radio"]').trigger('change')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([41])
  })

  it('preserves multi-selection while loading another page', async () => {
    listAccountOptions.mockReset()
    listAccountOptions
      .mockResolvedValueOnce({ data: { items: [{ id: 41, name: 'First', platform: 'openai', status: 'active' }], total: 1, page: 1, page_size: 20, pages: 1 } })
      .mockResolvedValueOnce({ data: { items: [{ id: 41, name: 'First', platform: 'openai', status: 'active' }], total: 2, page: 1, page_size: 20, pages: 2 } })
      .mockResolvedValueOnce({ data: { items: [{ id: 42, name: 'Second', platform: 'openai', status: 'active' }], total: 2, page: 2, page_size: 20, pages: 2 } })
    const wrapper = mount(CafeRoomAccountPicker, { props: { modelValue: [41], multiple: true, planId: 21, active: true } })
    await flushPromises()

    expect(listAccountOptions).toHaveBeenCalledWith({ ids: [41] })
    await wrapper.find('button').trigger('click')
    await flushPromises()
    expect(listAccountOptions).toHaveBeenLastCalledWith(expect.objectContaining({ plan_id: 21, page: 2 }))
    await wrapper.findAll('input[type="checkbox"]').at(-1)?.trigger('change')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([[41, 42]])
  })

  it('keeps candidate and selected-summary responses independent when both are deferred', async () => {
    listAccountOptions.mockReset()
    const selected = deferred<{ data: { items: Array<{ id: number, name: string, platform: string, status: string }>, total: number, page: number, page_size: number, pages: number } }>()
    const candidates = deferred<{ data: { items: Array<{ id: number, name: string, platform: string, status: string }>, total: number, page: number, page_size: number, pages: number } }>()
    listAccountOptions.mockImplementationOnce(() => selected.promise).mockImplementationOnce(() => candidates.promise)
    const wrapper = mount(CafeRoomAccountPicker, { props: { modelValue: 41, planId: 21, active: true } })
    await flushPromises()

    candidates.resolve({ data: { items: [{ id: 42, name: 'Candidate account', platform: 'openai', status: 'active' }], total: 1, page: 1, page_size: 20, pages: 1 } })
    await flushPromises()
    expect(wrapper.text()).toContain('Candidate account')
    expect(wrapper.attributes('aria-busy')).toBe('false')

    selected.resolve({ data: { items: [{ id: 41, name: 'Hydrated account', platform: 'openai', status: 'active' }], total: 1, page: 1, page_size: 20, pages: 1 } })
    await flushPromises()
    expect(wrapper.text()).toContain('Hydrated account')
  })

  it('ignores a stale candidate response after the plan changes', async () => {
    listAccountOptions.mockReset()
    const oldCandidates = deferred<{ data: { items: Array<{ id: number, name: string, platform: string, status: string }>, total: number, page: number, page_size: number, pages: number } }>()
    listAccountOptions.mockImplementationOnce(() => oldCandidates.promise).mockResolvedValueOnce({ data: { items: [{ id: 52, name: 'New plan account', platform: 'anthropic', status: 'active' }], total: 1, page: 1, page_size: 20, pages: 1 } })
    const wrapper = mount(CafeRoomAccountPicker, { props: { modelValue: 0, planId: 21, active: true } })
    await flushPromises()
    await wrapper.setProps({ planId: 22 })
    await flushPromises()
    oldCandidates.resolve({ data: { items: [{ id: 51, name: 'Old plan account', platform: 'openai', status: 'active' }], total: 1, page: 1, page_size: 20, pages: 1 } })
    await flushPromises()

    expect(wrapper.text()).toContain('New plan account')
    expect(wrapper.text()).not.toContain('Old plan account')
    expect(wrapper.attributes('aria-busy')).toBe('false')
  })
})
