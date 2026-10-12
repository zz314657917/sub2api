import { afterEach, describe, expect, it } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import BaseDialog from '../BaseDialog.vue'

enableAutoUnmount(afterEach)
const pressEscape = () => document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))

describe('stacked dialog Escape handling', () => {
  it('closes only the most recently opened dialog, then allows the parent to close', async () => {
    const parent = mount(BaseDialog, { props: { show: true, title: 'Parent' } })
    const child = mount(BaseDialog, { props: { show: true, title: 'Child' } })
    pressEscape()
    expect(child.emitted('close')).toHaveLength(1)
    expect(parent.emitted('close')).toBeUndefined()
    await child.setProps({ show: false })
    pressEscape()
    expect(parent.emitted('close')).toHaveLength(1)
  })

  it('does not dismiss a parent behind a child that disallows Escape', () => {
    const parent = mount(BaseDialog, { props: { show: true, title: 'Parent' } })
    const child = mount(BaseDialog, { props: { show: true, title: 'Child', closeOnEscape: false } })
    pressEscape()
    expect(child.emitted('close')).toBeUndefined()
    expect(parent.emitted('close')).toBeUndefined()
    child.unmount()
    pressEscape()
    expect(parent.emitted('close')).toHaveLength(1)
  })

  it('treats a reopened dialog as topmost and keeps the parent scroll lock', async () => {
    const parent = mount(BaseDialog, { props: { show: true, title: 'Parent' } })
    const child = mount(BaseDialog, { props: { show: true, title: 'Child' } })
    await parent.setProps({ show: false })
    expect(document.body.classList.contains('modal-open')).toBe(true)
    await parent.setProps({ show: true })
    pressEscape()
    expect(parent.emitted('close')).toHaveLength(1)
    expect(child.emitted('close')).toBeUndefined()
    parent.unmount()
    expect(document.body.classList.contains('modal-open')).toBe(true)
    child.unmount()
    expect(document.body.classList.contains('modal-open')).toBe(false)
  })
})
