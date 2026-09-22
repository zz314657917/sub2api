import { afterEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import ModelTagInput from '../ModelTagInput.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)

describe('model tag keyboard navigation', () => {
  it('allows Tab to leave an empty input but commits a non-empty model', async () => {
    const wrapper = mount(ModelTagInput, { props: { models: [] } })
    const input = wrapper.get('input')
    const empty = new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true })
    input.element.dispatchEvent(empty)
    expect(empty.defaultPrevented).toBe(false)
    await input.setValue('gpt-test')
    const filled = new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true })
    input.element.dispatchEvent(filled)
    expect(filled.defaultPrevented).toBe(true)
    expect(wrapper.emitted('update:models')).toEqual([[['gpt-test']]])
  })
})
