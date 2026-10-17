import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import type { AdminGroup } from '@/types'
import GroupRateMultipliersModal from '../GroupRateMultipliersModal.vue'

const mocks = vi.hoisted(() => ({ getGroupRateMultipliers: vi.fn(), batchSetGroupRateMultipliers: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { groups: mocks } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess: vi.fn(), showError: vi.fn() }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
afterEach(() => vi.restoreAllMocks())
beforeEach(() => {
  vi.resetAllMocks()
  mocks.getGroupRateMultipliers.mockResolvedValue([
    { user_id: 7, user_email: 'user@example.com', user_name: '', user_status: 'active', rate_multiplier: 2, rpm_override: null }
  ])
})

async function openEditor() {
  const wrapper = mount(GroupRateMultipliersModal, {
    props: { show: false, group: { id: 1, name: 'Group', platform: 'openai', rate_multiplier: 1 } as AdminGroup },
    global: { stubs: {
      BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' },
      Icon: true, PlatformIcon: true, Pagination: true
    } }
  })
  await wrapper.setProps({ show: true })
  await flushPromises()
  return wrapper
}

describe('group rate multiplier edits', () => {
  it.each(['-1', '-0.001', 'NaN', 'Infinity', '-Infinity', '1e309', 'invalid'])('does not stage invalid custom rate %s', async (value) => {
    const wrapper = await openEditor()
    const input = wrapper.get<HTMLInputElement>('tbody input[type="number"]')
    // number inputs sanitize nonfinite text to ''. Supply the raw change value
    // so this tests rejection rather than accidentally testing override clearing.
    const valueGetter = vi.spyOn(input.element, 'value', 'get').mockReturnValue(value)
    await input.trigger('change')
    valueGetter.mockRestore()
    expect(wrapper.findAll('button').some(button => button.text() === 'common.save')).toBe(false)
    expect(mocks.batchSetGroupRateMultipliers).not.toHaveBeenCalled()

    // Invalid edits must also preserve a valid, unsaved override.
    await input.setValue('0.5')
    const draftGetter = vi.spyOn(input.element, 'value', 'get').mockReturnValue(value)
    await input.trigger('change')
    draftGetter.mockRestore()
    await wrapper.findAll('button').find(button => button.text() === 'common.save')!.trigger('click')
    expect(mocks.batchSetGroupRateMultipliers).toHaveBeenCalledWith(1, [{ user_id: 7, rate_multiplier: 0.5 }])
  })

  it.each([['0', 0], ['0.5', 0.5], ['1e-3', 0.001]])('saves nonnegative rate %s', async (value, expected) => {
    const wrapper = await openEditor()
    await wrapper.get('tbody input[type="number"]').setValue(value)
    await wrapper.findAll('button').find(button => button.text() === 'common.save')!.trigger('click')
    expect(mocks.batchSetGroupRateMultipliers).toHaveBeenCalledWith(1, [{ user_id: 7, rate_multiplier: expected }])
  })

  it('still removes an override when its input is cleared', async () => {
    const wrapper = await openEditor()
    await wrapper.get('tbody input[type="number"]').setValue('')
    await wrapper.findAll('button').find(button => button.text() === 'common.save')!.trigger('click')
    expect(mocks.batchSetGroupRateMultipliers).toHaveBeenCalledWith(1, [])
  })
})
