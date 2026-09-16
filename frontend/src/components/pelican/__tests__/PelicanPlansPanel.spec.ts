import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import PelicanPlansPanel from '../PelicanPlansPanel.vue'

describe('PelicanPlansPanel', () => {
  it('renders actual plan fields and emits its selected action', async () => {
    const wrapper = mount(PelicanPlansPanel, { props: { loading: false, groups: [{ id: 3, name: '授权分组' }], plans: [{ id: 7, group_id: 3, group_name: '授权分组', model_id: 'gpt-test', interval_minutes: 30, enabled: false, max_results: 20, min_chars: 9366, created_at: '', updated_at: '' }] } })
    expect(wrapper.text()).toContain('授权分组 · gpt-test')
    expect(wrapper.text()).toContain('已停用')
    await wrapper.findAll('button').find(button => button.text() === '立即运行')!.trigger('click')
    expect(wrapper.emitted('run')?.[0]).toEqual([7])
  })
  it('emits editable form fields for a new disabled-by-default plan', async () => {
    const wrapper = mount(PelicanPlansPanel, { props: { loading: false, groups: [{ id: 3, name: '授权分组' }], plans: [] } })
    await wrapper.get('select').setValue('3')
    await wrapper.get('input[aria-label="模型 ID"]').setValue('gpt-test')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[0]).toEqual([null, expect.objectContaining({ group_id: 3, model_id: 'gpt-test', enabled: false, interval_minutes: 60, max_results: 20, min_chars: 9366 })])
  })
})
