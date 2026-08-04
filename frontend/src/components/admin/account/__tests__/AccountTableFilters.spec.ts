import { defineComponent } from 'vue'
import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import AccountTableFilters from '../AccountTableFilters.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

const AccountTableSelectStub = defineComponent({
  name: 'AccountTableSelect',
  props: {
    modelValue: { type: [String, Number, Boolean], default: '' },
    options: { type: Array, default: () => [] }
  },
  emits: ['update:modelValue', 'change'],
  template: '<div />'
})

describe('AccountTableFilters plan type filter', () => {
  it('exposes all plan categories and emits the selected value', async () => {
    const wrapper = mount(AccountTableFilters, {
      props: {
        searchQuery: '',
        filters: {
          platform: '',
          type: '',
          status: '',
          privacy_mode: '',
          plan_type: '',
          owner_filter: '',
          share_mode: '',
          share_status: '',
          group: ''
        }
      },
      global: {
        stubs: {
          Select: AccountTableSelectStub,
          SearchInput: true
        }
      }
    })

    const planSelect = wrapper.findAllComponents(AccountTableSelectStub).find((select) =>
      (select.props('options') as Array<{ value: string }>).some((option) => option.value === 'k12')
    )
    expect(planSelect).toBeDefined()
    expect((planSelect?.props('options') as Array<{ value: string }>).map((option) => option.value)).toEqual([
      '', 'plus', 'pro', 'k12', 'team', 'free', 'other', 'unrecognized'
    ])

    planSelect?.vm.$emit('update:modelValue', 'k12')
    await wrapper.vm.$nextTick()

    expect(wrapper.emitted('update:filters')?.at(-1)?.[0]).toMatchObject({ plan_type: 'k12' })
  })
})
