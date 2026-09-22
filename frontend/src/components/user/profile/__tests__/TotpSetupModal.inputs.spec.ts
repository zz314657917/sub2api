import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import TotpSetupModal from '../TotpSetupModal.vue'

const api = vi.hoisted(() => ({ getVerificationMethod: vi.fn(), initiateSetup: vi.fn(), enable: vi.fn() }))
vi.mock('@/api', () => ({ totpAPI: api }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn() }) }))
vi.mock('@/utils/apiError', () => ({ extractApiErrorMessage: (error: { message?: string }, fallback: string) => error.message || fallback }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('qrcode', () => ({ default: { toDataURL: vi.fn() } }))
enableAutoUnmount(afterEach)

describe('TOTP setup input synchronization', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    api.getVerificationMethod.mockResolvedValue({ method: 'password' })
    api.initiateSetup.mockResolvedValue({ secret: 's', qr_code_url: '', setup_token: 't' })
  })

  it('clears the displayed digits after verify failure', async () => {
    api.enable.mockRejectedValueOnce(new Error('invalid'))
    const wrapper = mount(TotpSetupModal)
    await flushPromises()
    await wrapper.get('input[type="password"]').setValue('password')
    await wrapper.get('.btn-primary').trigger('click')
    await flushPromises()
    await wrapper.get('.btn-primary').trigger('click')
    for (const input of wrapper.findAll('input[maxlength="1"]')) await input.setValue('1')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.findAll('input[maxlength="1"]').map(input => (input.element as HTMLInputElement).value)).toEqual(['', '', '', '', '', ''])
  })
})
