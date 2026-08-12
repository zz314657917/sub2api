import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent } from 'vue'

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: {
      server_utc_offset: '+08:00'
    }
  })
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, values?: Record<string, string>) =>
      values?.timezone ? `${key}:${values.timezone}` : key
  })
}))

import AccountTimeAvailabilityWindow from '../AccountTimeAvailabilityWindow.vue'

const ToggleStub = defineComponent({
  props: {
    modelValue: Boolean
  },
  emits: ['update:modelValue'],
  template: '<button type="button" @click="$emit(\'update:modelValue\', !modelValue)"><slot /></button>'
})

function mountWindow(initial = { enabled: false, start: '', end: '' }) {
  return mount(AccountTimeAvailabilityWindow, {
    props: initial,
    global: {
      stubs: { Toggle: ToggleStub }
    }
  })
}

describe('AccountTimeAvailabilityWindow', () => {
  it('only shows time controls after enabling the window', async () => {
    const wrapper = mountWindow()

    expect(wrapper.find('[data-testid="account-time-availability-start"]').exists()).toBe(false)
    await wrapper.get('[data-testid="account-time-availability-enabled"]').trigger('click')

    expect(wrapper.emitted('update:enabled')?.[0]).toEqual([true])
  })

  it('reports invalid same-day windows and shows the server timezone', () => {
    const wrapper = mountWindow({ enabled: true, start: '22:00', end: '18:00' })

    expect(wrapper.text()).toContain('admin.accounts.timeAvailability.windowInvalid')
    expect(wrapper.text()).toContain('UTC+08:00')
    expect(wrapper.emitted('valid')?.at(-1)).toEqual([false])
    expect(wrapper.emitted('window-valid')?.at(-1)).toEqual([false])
  })

  it('accepts a valid window and preserves its values when disabled', async () => {
    const wrapper = mountWindow({ enabled: true, start: '18:00', end: '22:00' })

    expect(wrapper.emitted('valid')?.at(-1)).toEqual([true])
    await wrapper.get('[data-testid="account-time-availability-enabled"]').trigger('click')

    expect(wrapper.emitted('update:enabled')?.[0]).toEqual([false])
    expect(wrapper.props('start')).toBe('18:00')
    expect(wrapper.props('end')).toBe('22:00')
  })
})
