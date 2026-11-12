import { afterEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { defineComponent, h, nextTick } from 'vue'
import Select from '../Select.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)
afterEach(() => { document.body.innerHTML = '' })

const options = [{ value: 'a', label: 'A' }, { value: 'b', label: 'B' }]

describe('Select single open dropdown', () => {
  it('closes the open dropdown when another Select inside a click-stopping container opens', async () => {
    // Mirrors BaseDialog, whose content uses @click.stop.
    const Host = defineComponent(() => () =>
      h('div', { onClick: (e: MouseEvent) => e.stopPropagation() }, [
        h(Select, { modelValue: null, options }),
        h(Select, { modelValue: null, options })
      ])
    )
    const wrapper = mount(Host, { attachTo: document.body })
    const [first, second] = wrapper.findAll('button')

    first.element.click()
    await nextTick()
    expect(first.attributes('aria-expanded')).toBe('true')

    second.element.click()
    await nextTick()
    expect(first.attributes('aria-expanded')).toBe('false')
    expect(second.attributes('aria-expanded')).toBe('true')
    expect(document.querySelectorAll('[role="listbox"]')).toHaveLength(1)
  })

  it('keeps the dropdown open when clicking inside its own search box', async () => {
    const wrapper = mount(Select, {
      attachTo: document.body,
      props: { modelValue: null, searchable: true, options }
    })
    wrapper.get('button').element.click()
    await nextTick()
    document.querySelector<HTMLInputElement>('.select-search-input')!.click()
    await nextTick()
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('true')
  })
})
