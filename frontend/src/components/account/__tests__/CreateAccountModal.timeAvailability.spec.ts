import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const { showError, createAccount, getWebSearchEmulationConfig, listTLSProfiles } = vi.hoisted(() => ({
  showError: vi.fn(),
  createAccount: vi.fn(),
  getWebSearchEmulationConfig: vi.fn(),
  listTLSProfiles: vi.fn()
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: { server_utc_offset: '+08:00' },
    showError,
    showSuccess: vi.fn(),
    showWarning: vi.fn()
  })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ isSimpleMode: true })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      create: createAccount,
      checkMixedChannelRisk: vi.fn().mockResolvedValue({ has_risk: false })
    },
    settings: {
      getWebSearchEmulationConfig,
      getSettings: vi.fn().mockResolvedValue({})
    },
    tlsFingerprintProfiles: {
      list: listTLSProfiles
    }
  }
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

import CreateAccountModal from '../CreateAccountModal.vue'

const BaseDialogStub = defineComponent({
  props: { show: Boolean },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})

const ToggleStub = defineComponent({
  inheritAttrs: false,
  props: { modelValue: Boolean },
  emits: ['update:modelValue'],
  template: '<button v-bind="$attrs" type="button" @click="$emit(\'update:modelValue\', !modelValue)"></button>'
})

function mountModal() {
  return mount(CreateAccountModal, {
    props: { show: true, proxies: [], groups: [] },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        Toggle: ToggleStub,
        ConfirmDialog: true,
        Select: true,
        PlatformIcon: true,
        Icon: true,
        ProxySelector: true,
        GroupSelector: true,
        ModelWhitelistSelector: true,
        AccountCapabilitySelector: true,
        QuotaLimitCard: true,
        ShareDisplayCard: true,
        OAuthAuthorizationFlow: true
      }
    }
  })
}

describe('CreateAccountModal account availability', () => {
  beforeEach(() => {
    showError.mockReset()
    createAccount.mockReset()
    getWebSearchEmulationConfig.mockResolvedValue({ enabled: false, providers: [] })
    listTLSProfiles.mockResolvedValue([])
  })

  it('blocks OAuth next before authorization when the enabled window is invalid', async () => {
    const wrapper = mountModal()
    await flushPromises()
    await wrapper.get('[data-tour="account-form-name"]').setValue('Timed OAuth')
    await wrapper.get('[data-testid="account-time-availability-enabled"]').trigger('click')
    await wrapper.get('[data-testid="account-time-availability-start"]').setValue('22:00')
    await wrapper.get('[data-testid="account-time-availability-end"]').setValue('18:00')

    await wrapper.get('#create-account-form').trigger('submit.prevent')
    await flushPromises()

    expect(showError).toHaveBeenCalledWith('admin.accounts.timeAvailability.windowInvalid')
    expect(wrapper.find('#create-account-form').exists()).toBe(true)
  })

  it('includes a valid availability window in a direct-create payload', async () => {
    createAccount.mockResolvedValue({ id: 7 })
    const wrapper = mountModal()
    await flushPromises()
    await wrapper.get('[data-tour="account-form-name"]').setValue('Timed API Key')
    const typeButtons = wrapper.findAll('[data-tour="account-form-type"] button')
    await typeButtons[1].trigger('click')
    await wrapper.get('input[type="password"]').setValue('sk-test')
    await wrapper.get('[data-testid="account-time-availability-enabled"]').trigger('click')
    await wrapper.get('[data-testid="account-time-availability-start"]').setValue('09:00')
    await wrapper.get('[data-testid="account-time-availability-end"]').setValue('18:00')

    await wrapper.get('#create-account-form').trigger('submit.prevent')
    await flushPromises()

    expect(createAccount).toHaveBeenCalledWith(expect.objectContaining({
      extra: expect.objectContaining({
        account_availability_enabled: true,
        account_availability_start: '09:00',
        account_availability_end: '18:00'
      })
    }))
  })
})
