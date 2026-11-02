import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import PelicanGroupOrderSettings from '../PelicanGroupOrderSettings.vue'

describe('PelicanGroupOrderSettings', () => {
  const groups = [{ id: 1, name: '分组一' }, { id: 2, name: '分组二' }, { id: 3, name: '分组三' }]

  it('excludes selected groups and adds a new priority without changing the input', async () => {
    const initial = [2]
    const wrapper = mount(PelicanGroupOrderSettings, { props: { modelValue: initial, groups } })
    expect(wrapper.findAll('option').map(option => option.attributes('value'))).toEqual(['', '1', '3'])
    await wrapper.get('select').setValue('3')
    await wrapper.findAll('button').find(button => button.text() === '添加')!.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[[2, 3]]])
    expect(initial).toEqual([2])
    wrapper.unmount()
  })

  it('moves, removes and clears priorities, retaining unavailable group IDs', async () => {
    const wrapper = mount(PelicanGroupOrderSettings, { props: { modelValue: [99, 2, 1], groups } })
    expect(wrapper.text()).toContain('分组 #99（暂不可用）')
    expect(wrapper.get('button[aria-label="上移 分组 #99（暂不可用）"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('button[aria-label="下移 分组一"]').attributes('disabled')).toBeDefined()
    await wrapper.get('button[aria-label="上移 分组二"]').trigger('click')
    await wrapper.get('button[aria-label="下移 分组二"]').trigger('click')
    await wrapper.get('button[aria-label="移除 分组二"]').trigger('click')
    await wrapper.findAll('button').find(button => button.text() === '恢复默认排序')!.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[[2, 99, 1]], [[99, 1, 2]], [[99, 1]], [[]]])
    wrapper.unmount()
  })

  it('keeps saved IDs when the directory fails and allows retry', async () => {
    const wrapper = mount(PelicanGroupOrderSettings, { props: { modelValue: [99], groups: [], error: true } })
    expect(wrapper.get('.group-order-list').text()).toContain('#99')
    expect(wrapper.get('[role="alert"]').text()).toContain('已保存的排序仍保留')
    await wrapper.findAll('button').find(button => button.text() === '重试')!.trigger('click')
    expect(wrapper.emitted('retry')).toHaveLength(1)
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    wrapper.unmount()
  })
})
