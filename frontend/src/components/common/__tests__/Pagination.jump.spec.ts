import { afterEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import Pagination from '../Pagination.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)

describe('pagination jump', () => {
  it('accepts a numeric jump value emitted by the input component', async () => {
    const wrapper = mount(Pagination, { props: { total: 200, page: 1, pageSize: 20, showJump: true }, global: { stubs: { Icon: true, Select: true } } })
    const input = wrapper.get('input')
    await input.setValue(5 as unknown as string)
    await input.trigger('keyup.enter')
    expect(wrapper.emitted('update:page')).toEqual([[5]])
  })
})
