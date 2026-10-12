import { afterEach, describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import { reactive } from 'vue'
import HomeView from '../views/HomeView.vue'

const state = vi.hoisted(() => ({
  auth: { isAuthenticated: false, isAdmin: false, user: null, checkAuth: vi.fn(), refreshUser: vi.fn().mockResolvedValue(undefined) },
  app: { cachedPublicSettings: { api_base_url: '' }, publicSettingsLoaded: true }
}))
vi.mock('@/stores', () => ({ useAuthStore: () => reactive(state.auth), useAppStore: () => state.app }))
vi.mock('@/components/auth/AuthAccessPanel.vue', () => ({ default: { template: '<div />' } }))
vi.mock('../views/public/components/PublicTopNav.vue', () => ({ default: { template: '<nav />' } }))
vi.mock('../views/public/components/PublicRevealBackdrop.vue', () => ({ default: { template: '<div />' } }))
vi.mock('vue-router', () => ({ useRoute: () => ({ path: '/home' }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/api/usage', () => ({ usageAPI: { getDashboardStats: vi.fn().mockResolvedValue({}) } }))
vi.mock('@/utils/credits', () => ({ formatCreditAmount: () => '0.00' }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copied: false, copyToClipboard: vi.fn() }) }))

const wrappers: ReturnType<typeof shallowMount>[] = []
function render(loggedIn: boolean, url = 'https://example.com', admin = false) {
  state.auth.isAuthenticated = loggedIn
  state.auth.isAdmin = admin
  state.app.cachedPublicSettings.api_base_url = url
  const wrapper = shallowMount(HomeView, {
    global: { stubs: { RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' } } }
  })
  wrappers.push(wrapper)
  return wrapper
}
afterEach(() => wrappers.splice(0).forEach(wrapper => wrapper.unmount()))

describe('homepage account layout', () => {
  it('keeps the guest auth panel and hides integration guidance', () => {
    const wrapper = render(false)
    expect(wrapper.find('#home-auth-panel').exists()).toBe(true)
    expect(wrapper.find('.home-integration-preview').exists()).toBe(false)
    expect(wrapper.find('.home-account-column').exists()).toBe(false)
  })

  it.each([false, true])('places guidance below the account card and routes the dashboard (admin=%s)', admin => {
    const wrapper = render(true, 'https://example.com', admin)
    const column = wrapper.get('.home-account-column')
    expect(column.element.children[0].classList.contains('home-account-workbench')).toBe(true)
    expect(column.element.children[1].classList.contains('home-integration-preview')).toBe(true)
    const links = column.findAll('.home-account-workbench-link')
    expect(links.map(link => link.attributes('href'))).toEqual(['/purchase', admin ? '/admin/dashboard' : '/dashboard'])
    expect(wrapper.get('.home-mimo-actions').findAll('a').map(link => link.attributes('href'))).toEqual(['/tutorial/getting-started'])
    expect(wrapper.find('#home-auth-panel').exists()).toBe(false)
  })

  it.each(['https://example.com', 'https://example.com/', 'https://example.com/v1', 'https://example.com/v1/'])('normalizes the OpenAI-compatible base URL: %s', url => {
    const wrapper = render(true, url)
    expect(wrapper.get('.home-integration-preview-row code').text()).toBe('https://example.com/v1')
  })
})
