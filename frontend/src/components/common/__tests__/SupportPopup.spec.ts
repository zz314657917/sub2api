import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import SupportPopup from '@/components/common/SupportPopup.vue'
import { useAppStore } from '@/stores/app'

const supportPopupSource = readFileSync(resolve(process.cwd(), 'src/components/common/SupportPopup.vue'), 'utf8')

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => {
        if (key === 'common.close') return 'Close'
        if (key === 'home.contactSupport') return 'Contact Support'
        return key
      },
    }),
  }
})

describe('SupportPopup', () => {
  beforeEach(() => {
    const pinia = createPinia()
    setActivePinia(pinia)
    document.body.innerHTML = ''
    document.body.classList.remove('support-popup-open')
  })

  afterEach(() => {
    document.body.innerHTML = ''
    document.body.classList.remove('support-popup-open')
  })

  it('renders configured support images and text', async () => {
    const appStore = useAppStore()
    appStore.cachedPublicSettings = {
      support_popup_title: 'Join support group',
      support_popup_description: 'Scan a code to join',
      support_popup_footer: 'Group 2 is recommended',
      support_popup_items: [
        {
          id: 'group-1',
          title: 'Group 1',
          image_url: 'data:image/png;base64,aaa',
          caption: 'Full',
          badge: 'Full',
        },
        {
          id: 'group-2',
          title: 'Group 2',
          image_url: 'https://example.com/support.png',
          caption: '1078510185',
          badge: '',
        },
      ],
    }

    const wrapper = mount(SupportPopup, {
      attachTo: document.body,
      props: { show: true },
    })

    await nextTick()

    expect(document.body.textContent).toContain('Join support group')
    expect(document.body.textContent).toContain('Scan a code to join')
    expect(document.body.textContent).toContain('Group 1')
    expect(document.body.textContent).toContain('Group 2')
    expect(document.body.textContent).toContain('Full')
    expect(document.body.textContent).toContain('1078510185')
    expect(document.body.textContent).toContain('Group 2 is recommended')
    expect(document.body.querySelectorAll('.support-popup-card-caption')).toHaveLength(2)
    expect(document.body.querySelectorAll('.support-popup-image')).toHaveLength(2)
    expect(document.body.classList.contains('support-popup-open')).toBe(true)

    wrapper.unmount()
  })

  it('renders contact text when no image item is configured', async () => {
    const appStore = useAppStore()
    appStore.cachedPublicSettings = {
      support_popup_title: 'Join support group',
      contact_info: 'QQ: 123456789',
      support_popup_items: [
        { id: 'empty-image', title: 'Empty image', image_url: '', caption: '', badge: '' },
      ],
    }

    mount(SupportPopup, {
      attachTo: document.body,
      props: { show: true },
    })

    await nextTick()

    expect(document.body.querySelector('[role="dialog"]')).not.toBeNull()
    expect(document.body.textContent).toContain('QQ: 123456789')
    expect(document.body.classList.contains('support-popup-open')).toBe(true)
  })

  it('emits close from the close button and Escape key', async () => {
    const appStore = useAppStore()
    appStore.cachedPublicSettings = {
      support_popup_items: [
        {
          id: 'group',
          title: 'Support group',
          image_url: 'data:image/png;base64,aaa',
          caption: '',
          badge: '',
        },
      ],
    }

    const wrapper = mount(SupportPopup, {
      attachTo: document.body,
      props: { show: true },
    })

    await nextTick()

    const closeButton = document.body.querySelector('button[aria-label="Close"]')
    if (!(closeButton instanceof HTMLButtonElement)) {
      throw new Error('close button not found')
    }
    closeButton.click()
    await nextTick()

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await nextTick()

    expect(wrapper.emitted('close')).toHaveLength(2)
  })

  it('uses the warm clay popup theme instead of the old blue-slate treatment', () => {
    expect(supportPopupSource).toContain('background: #efe9de')
    expect(supportPopupSource).toContain('color: #a9583e')
    expect(supportPopupSource).toContain('rgba(204, 120, 92')
    expect(supportPopupSource).toContain('background: rgba(255, 252, 246')
    expect(supportPopupSource).toContain('border: 1px solid #d8cec2')
    expect(supportPopupSource).not.toContain('rgba(59, 130, 246')
    expect(supportPopupSource).not.toContain('rgba(37, 99, 235')
    expect(supportPopupSource).not.toContain('color: #2563eb')
    expect(supportPopupSource).not.toContain('color: #93c5fd')
  })
})
