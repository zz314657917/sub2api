import { afterEach, describe, expect, it } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import BaseDialog from '../BaseDialog.vue'

enableAutoUnmount(afterEach)

describe('dialog accessible titles', () => {
  it('assigns each simultaneous dialog its own title id', async () => {
    mount(BaseDialog, { props: { show: true, title: 'First' }, global: { stubs: { Icon: true } } })
    mount(BaseDialog, { props: { show: true, title: 'Second' }, global: { stubs: { Icon: true } } })
    await nextTick()
    const dialogs = Array.from(document.body.querySelectorAll('[role="dialog"]'))
    const ids = dialogs.map(dialog => dialog.getAttribute('aria-labelledby'))
    expect(new Set(ids).size).toBe(2)
    expect(ids.map(id => document.getElementById(id!)?.textContent)).toEqual(['First', 'Second'])
  })
})
