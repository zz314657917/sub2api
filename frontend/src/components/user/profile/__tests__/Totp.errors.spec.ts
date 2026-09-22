import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import TotpDisableDialog from '../TotpDisableDialog.vue'

const api = vi.hoisted(() => ({ getVerificationMethod: vi.fn(), sendVerifyCode: vi.fn(), disable: vi.fn() }))
const showError = vi.hoisted(() => vi.fn())
vi.mock('@/api', () => ({ totpAPI: api }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, showSuccess: vi.fn() }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)

describe('TOTP normalized errors', () => {
  beforeEach(() => { vi.resetAllMocks(); api.getVerificationMethod.mockResolvedValue({ method: 'password' }) })

  it('shows a normalized standard Error instead of an undefined response field', async () => {
    api.disable.mockRejectedValueOnce(new Error('Incorrect password'))
    const wrapper = mount(TotpDisableDialog)
    await flushPromises()
    await wrapper.get('input[type="password"]').setValue('bad')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(showError).toHaveBeenCalledWith('Incorrect password')
  })

  it('shows a normalized legacy Axios response message', async () => {
    api.disable.mockRejectedValueOnce({ response: { data: { detail: 'Legacy response detail' } } })
    const wrapper = mount(TotpDisableDialog)
    await flushPromises()
    await wrapper.get('input[type="password"]').setValue('bad')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(showError).toHaveBeenCalledWith('Legacy response detail')
  })
})
