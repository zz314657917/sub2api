import { mount } from '@vue/test-utils'
import { defineComponent, ref } from 'vue'
import { describe, expect, it } from 'vitest'
import PelicanReferenceSettings from '../PelicanReferenceSettings.vue'
import type { PelicanReferenceCatalog, PelicanReferenceSettings as Settings } from '@/api/pelicanTests'

const catalog: PelicanReferenceCatalog = { status: 'ready', synced_at: '2026-10-30T08:00:00Z', models: [
  { model: 'gpt-6.1-sol', display_name: 'GPT-6.1 Sol', efforts: ['low', 'high', 'ultra'] },
  { model: 'gpt-6-luna', display_name: 'GPT-6 Luna', efforts: ['medium', 'max'] }
] }
function editor(initial: Settings, source: PelicanReferenceCatalog | null = catalog) {
  return mount(defineComponent({
    components: { PelicanReferenceSettings },
    setup() { return { settings: ref(initial), catalog: source } },
    template: '<PelicanReferenceSettings v-model="settings" :catalog="catalog" />'
  }))
}
describe('PelicanReferenceSettings', () => {
  it('adds exact catalog combinations, toggles independently and supports ordered move/remove', async () => {
    const wrapper = editor({ enabled: false, selections: [] })
    await wrapper.get('input[aria-label="启用第三方参考评分"]').setValue(true)
    await wrapper.get('.reference-add').trigger('click')
    await wrapper.get('.reference-add').trigger('click')
    expect(wrapper.findAll('input[list]').map(node => (node.element as HTMLInputElement).value)).toEqual(['gpt-6.1-sol', 'gpt-6.1-sol'])
    expect(wrapper.findAll('select').map(node => (node.element as HTMLSelectElement).value)).toEqual(['low', 'high'])
    await wrapper.get('button[aria-label="下移参考组合 1"]').trigger('click')
    expect(wrapper.findAll('select').map(node => (node.element as HTMLSelectElement).value)).toEqual(['high', 'low'])
    await wrapper.get('button[aria-label="上移参考组合 2"]').trigger('click')
    expect(wrapper.findAll('select').map(node => (node.element as HTMLSelectElement).value)).toEqual(['low', 'high'])
    await wrapper.get('button[aria-label="移除参考组合 1"]').trigger('click')
    expect(wrapper.findAll('select').map(node => (node.element as HTMLSelectElement).value)).toEqual(['high'])
    expect(wrapper.get('input[type="checkbox"]').element).toHaveProperty('checked', true)
  })

  it('retains unknown models and unsupported saved efforts without falling back to a supported effort', async () => {
    const wrapper = editor({ enabled: true, selections: [{ model: 'unknown-custom', effort: 'ultra' }, { model: 'gpt-6-luna', effort: 'ultra' }] })
    expect(wrapper.get('input[aria-label="参考模型 1"]').element).toHaveProperty('value', 'unknown-custom')
    expect(wrapper.get('select[aria-label="参考档位 2"]').element).toHaveProperty('value', 'ultra')
    expect(wrapper.get('select[aria-label="参考档位 2"]').text()).toContain('ultra（来源未支持）')
    expect(wrapper.get('select[aria-label="参考档位 2"]').findAll('option').map(node => node.attributes('value'))).toEqual(['ultra', 'medium', 'max'])
    await wrapper.get('input[aria-label="参考模型 2"]').setValue('custom-two')
    expect(wrapper.get('select[aria-label="参考档位 2"]').element).toHaveProperty('value', 'ultra')
    expect(wrapper.get('select[aria-label="参考档位 2"]').findAll('option')).toHaveLength(6)
  })

  it('keeps saved selections editable when the catalog is unavailable', async () => {
    const wrapper = editor({ enabled: true, selections: [{ model: 'kept', effort: 'max' }] }, null)
    expect(wrapper.text()).toContain('来源目录暂不可用')
    await wrapper.get('select').setValue('ultra')
    expect(wrapper.get('input[list]').element).toHaveProperty('value', 'kept')
    expect(wrapper.get('select').element).toHaveProperty('value', 'ultra')
    await wrapper.get('.reference-add').trigger('click')
    expect(wrapper.findAll('input[list]')).toHaveLength(2)
    expect(wrapper.text()).toContain('请填写模型名称')
  })

  it('rejects exact or display-alias duplicate pairs, permits different efforts and caps sixty-four pairs', async () => {
    const wrapper = editor({ enabled: true, selections: [{ model: 'gpt-6.1-sol', effort: 'low' }, { model: 'GPT-6.1 Sol', effort: 'low' }] })
    expect(wrapper.text()).toContain('模型与档位组合不能重复')
    await wrapper.get('select[aria-label="参考档位 2"]').setValue('high')
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    const many = mount(PelicanReferenceSettings, { props: { modelValue: { enabled: true, selections: Array.from({ length: 64 }, (_, i) => ({ model: `custom-${i}`, effort: 'high' })) } } })
    expect(many.get('.reference-add').attributes('disabled')).toBeDefined()
    expect(many.find('[role="alert"]').exists()).toBe(false)
    await many.setProps({ modelValue: { enabled: true, selections: Array.from({ length: 65 }, (_, i) => ({ model: `custom-${i}`, effort: 'high' })) } })
    expect(many.text()).toContain('最多选择 64 个参考组合')
    many.unmount()
  })

  it('appends only missing catalog efforts in source order, deduplicates exact aliases, and preserves saved pairs and objects', async () => {
    const settings: Settings = { enabled: true, selections: [
      { model: 'unknown-custom', effort: 'max' },
      { model: 'GPT-6.1 Sol', effort: 'high' },
      { model: 'gpt-6-luna', effort: 'ultra' },
      { model: 'gpt-6.1-sol', effort: 'max' },
      { model: 'GPT-6.1 SOL', effort: 'low' }
    ] }
    const before = structuredClone(settings)
    const source = { ...catalog, models: [{ ...catalog.models[0]!, efforts: ['ultra', 'high', 'low', 'ultra'] }, catalog.models[1]!] }
    const wrapper = mount(PelicanReferenceSettings, { props: { modelValue: settings, catalog: source } })
    await wrapper.get('button[aria-label="添加参考模型 2 的全部档位"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[{ enabled: true, selections: [...before.selections,
      { model: 'gpt-6.1-sol', effort: 'ultra' }, { model: 'gpt-6.1-sol', effort: 'low' }
    ] }]])
    expect(settings).toEqual(before)
    const result = wrapper.emitted('update:modelValue')![0]![0] as Settings
    expect(result.selections[0]).not.toBe(settings.selections[0])
    await wrapper.setProps({ modelValue: result })
    expect(wrapper.get('button[aria-label="添加参考模型 2 的全部档位"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('button[aria-label="添加参考模型 2 的全部档位"]').attributes('title')).toContain('已包含')
  })

  it('disables add-all for unknown, unavailable, loading and disabled settings without emitting', async () => {
    const settings: Settings = { enabled: true, selections: [{ model: 'gpt-6.1-sol', effort: 'low' }, { model: 'unknown', effort: 'max' }] }
    const wrapper = mount(PelicanReferenceSettings, { props: { modelValue: settings, catalog } })
    expect(wrapper.get('button[aria-label="添加参考模型 2 的全部档位"]').attributes('disabled')).toBeDefined()
    for (const props of [{ catalog: null }, { catalog: { ...catalog, status: 'unavailable' as const } }, { catalog, loading: true }, { loading: false, disabled: true }]) {
      await wrapper.setProps(props)
      const button = wrapper.get('button[aria-label="添加参考模型 1 的全部档位"]')
      expect(button.attributes('disabled')).toBeDefined()
      await button.trigger('click')
    }
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('adds all atomically at 64 and explains an overflow without truncating', async () => {
    const selections = [{ model: 'gpt-6.1-sol', effort: 'low' }, ...Array.from({ length: 61 }, (_, i) => ({ model: `custom-${i}`, effort: 'high' }))]
    const wrapper = mount(PelicanReferenceSettings, { props: { modelValue: { enabled: false, selections }, catalog: { ...catalog, status: 'stale' } } })
    const button = () => wrapper.get('button[aria-label="添加参考模型 1 的全部档位"]')
    expect(button().attributes('disabled')).toBeUndefined()
    await button().trigger('click')
    const result = wrapper.emitted('update:modelValue')![0]![0] as Settings
    expect(result.enabled).toBe(false)
    expect(result.selections).toHaveLength(64)
    expect(result.selections.slice(-2)).toEqual([{ model: 'gpt-6.1-sol', effort: 'high' }, { model: 'gpt-6.1-sol', effort: 'ultra' }])
    await wrapper.setProps({ modelValue: { enabled: true, selections: [...selections, { model: 'custom-62', effort: 'low' }] } })
    expect(button().attributes('disabled')).toBeDefined()
    expect(button().attributes('title')).toContain('超过 64')
    expect(wrapper.get('.reference-capacity').text()).toContain('请先移除')
    await button().trigger('click')
    expect(wrapper.emitted('update:modelValue')).toHaveLength(1)
    expect(selections).toHaveLength(62)
  })

  it('keeps existing twenty-pair configurations valid and supports another manual pair', async () => {
    const wrapper = editor({ enabled: true, selections: Array.from({ length: 20 }, (_, i) => ({ model: `custom-${i}`, effort: 'high' })) })
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.get('.reference-add').attributes('disabled')).toBeUndefined()
    await wrapper.get('.reference-add').trigger('click')
    expect(wrapper.findAll('input[list]')).toHaveLength(21)
  })

  it('does not mutate the supplied settings or selection objects', async () => {
    const settings = { enabled: false, selections: [{ model: 'custom', effort: 'high' }] }
    const wrapper = mount(PelicanReferenceSettings, { props: { modelValue: settings, catalog } })
    await wrapper.get('input[list]').setValue('edited')
    expect(settings.selections[0]!.model).toBe('custom')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([{ enabled: false, selections: [{ model: 'edited', effort: 'high' }] }])
    await wrapper.setProps({ disabled: true })
    expect(wrapper.get('fieldset').attributes('disabled')).toBeDefined()
  })

  it('keeps valid custom identities exact and rejects whitespace, controls and the backend byte limit', async () => {
    const wrapper = editor({ enabled: true, selections: [{ model: 'custom-exact', effort: 'high' }] })
    const input = wrapper.get('input[list]')
    await input.setValue(' custom-exact')
    expect(wrapper.text()).toContain('模型名称不能有首尾空格')
    await input.setValue('自'.repeat(67))
    expect(wrapper.text()).toContain('不能超过 200 字节')
    await input.setValue('custom-\u007f')
    expect(wrapper.text()).toContain('控制字符')
    await input.setValue('custom-exact-Unknown')
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  })
})
