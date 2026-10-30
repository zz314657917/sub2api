import { afterEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import AppHeader from '../AppHeader.vue'

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
  useRoute: () => ({ meta: {}, name: 'Dashboard', params: {} })
}))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/i18n', () => ({
  i18n: { global: { locale: { value: 'en' }, t: (key: string) => key } },
  availableLocales: [],
  setLocale: vi.fn()
}))
vi.mock('@/stores', () => ({
  useAppStore: () => ({
    cachedPublicSettings: { model_plaza_enabled: true },
    contactInfo: 'Support',
    docUrl: 'https://example.com/docs'
  }),
  useAuthStore: () => ({
    user: { username: 'Test', role: 'user', email: 'test@example.com', balance: 2 },
    isAdmin: false
  }),
  useOnboardingStore: () => ({ replay: vi.fn() }),
  usePelicanMetadataStore: () => ({ load: vi.fn().mockResolvedValue(undefined) })
}))
vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({ customMenuItems: [] })
}))
vi.mock('@/stores/announcements', () => ({ useAnnouncementStore: () => ({ announcements: [] }) }))

afterEach(() => {
  document.body.innerHTML = ''
})
function setup() {
  return mount(AppHeader, {
    attachTo: document.body,
    global: {
      stubs: {
        RouterLink: defineComponent({ props: ['to'], template: '<a :href="to"><slot /></a>' }),
        Icon: true,
        LocaleSwitcher: true,
        SubscriptionProgressMini: true,
        AnnouncementBell: true,
        HeaderAnnouncementCarousel: true
      }
    }
  })
}
describe('console header shortcuts and account panel', () => {
  it('exposes tutorial, models and support directly without opening a menu', async () => {
    const onSupport = vi.fn()
    window.addEventListener('sub2api:open-support-popup', onSupport)
    const wrapper = setup()
    expect(wrapper.get('.console-header-shortcuts a[href="/tutorial"]').text()).toContain('home.navTutorial')
    expect(wrapper.get('.console-header-shortcuts a[href="/models"]').text()).toContain('home.navModels')
    await wrapper.get('.console-header-shortcuts button[aria-label="common.contactSupport"]').trigger('click')
    expect(onSupport).toHaveBeenCalledOnce()
    expect(wrapper.find('[aria-controls="console-help-panel"]').exists()).toBe(false)
    window.removeEventListener('sub2api:open-support-popup', onSupport)
    wrapper.unmount()
  })
  it('preserves documentation and keyboard, Escape, outside-click and focus dismissal', async () => {
    const wrapper = setup()
    const button = wrapper.get<HTMLButtonElement>('[aria-controls="console-user-panel"]')
    await button.trigger('keydown', { key: 'ArrowDown' })
    const links = wrapper.findAll('#console-user-panel a, #console-user-panel button')
    expect(wrapper.get('#console-user-panel a[href="https://example.com/docs"]').text()).toContain('nav.docs')
    expect(document.activeElement).toBe(links[0].element)
    await links[0].trigger('keydown', { key: 'End' })
    expect(document.activeElement).toBe(links.at(-1)?.element)
    await links.at(-1)!.trigger('keydown', { key: 'Home' })
    expect(document.activeElement).toBe(links[0].element)
    await links[0].trigger('keydown', { key: 'Escape' })
    expect(button.attributes('aria-expanded')).toBe('false')
    expect(document.activeElement).toBe(button.element)
    await button.trigger('click')
    document.body.click()
    await wrapper.vm.$nextTick()
    expect(wrapper.find('#console-user-panel').exists()).toBe(false)
    await button.trigger('click')
    const outside = document.createElement('button')
    document.body.appendChild(outside)
    outside.focus()
    await wrapper.vm.$nextTick()
    expect(wrapper.find('#console-user-panel').exists()).toBe(false)
    wrapper.unmount()
  })
})
