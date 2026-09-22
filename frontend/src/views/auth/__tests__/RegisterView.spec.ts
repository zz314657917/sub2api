import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import RegisterView from '../RegisterView.vue'

const getPublicSettings = vi.hoisted(() => vi.fn())
const appStore = vi.hoisted(() => ({ cachedPublicSettings: { promo_code_enabled: false }, showError: vi.fn(), showSuccess: vi.fn(), showWarning: vi.fn() }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }), useRoute: () => ({ query: {} }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key, locale: { value: 'en' } })
}))
vi.mock('@/stores', () => ({ useAuthStore: () => ({ register: vi.fn() }), useAppStore: () => appStore }))
vi.mock('@/api/auth', () => ({ getPublicSettings }))

describe('registration promo visibility', () => {
  it('keeps a disabled promo field hidden before asynchronous settings resolve', () => {
    getPublicSettings.mockReturnValue(new Promise(() => {}))
    const wrapper = mount(RegisterView, { global: { stubs: { AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' }, Icon: true, TurnstileWidget: true, LoginAgreementPrompt: true, EmailOAuthButtons: true, LinuxDoOAuthSection: true, WechatOAuthSection: true, OidcOAuthSection: true, RouterLink: true, transition: false } } })
    expect(wrapper.find('#promo_code').exists()).toBe(false)
  })
})
