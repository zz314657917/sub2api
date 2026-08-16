import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const { createAccount, importCodexSession } = vi.hoisted(() => ({
  createAccount: vi.fn(),
  importCodexSession: vi.fn(),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ cachedPublicSettings: {}, showError: vi.fn(), showSuccess: vi.fn(), showWarning: vi.fn() }),
}))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ isSimpleMode: true }) }))
vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      create: createAccount,
      checkMixedChannelRisk: vi.fn().mockResolvedValue({ has_risk: false }),
      importCodexSession,
    },
    settings: { getWebSearchEmulationConfig: vi.fn().mockResolvedValue({ enabled: false, providers: [] }), getSettings: vi.fn().mockResolvedValue({}) },
    tlsFingerprintProfiles: { list: vi.fn().mockResolvedValue([]) },
  },
}))
vi.mock('@/api/admin/accounts', () => ({ getAntigravityDefaultModelMapping: vi.fn().mockResolvedValue([]) }))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

import CreateAccountModal from '../CreateAccountModal.vue'

const BaseDialogStub = defineComponent({ props: { show: Boolean }, template: '<div v-if="show"><slot /><slot name="footer" /></div>' })
const OAuthAuthorizationFlowStub = defineComponent({
  emits: ['import-codex-session'],
  template: '<button data-testid="import-codex-session" @click="$emit(\'import-codex-session\', \'session-json\')">session</button>',
})
const SelectStub = defineComponent({
  props: { modelValue: { type: String, default: '' }, options: { type: Array, default: () => [] } },
  emits: ['update:modelValue'],
  template: `
    <select v-bind="$attrs" :value="modelValue" @change="$emit('update:modelValue', $event.target.value)">
      <option v-for="option in options" :key="option.value" :value="option.value">{{ option.label }}</option>
    </select>
  `,
})

function mountModal() {
  return mount(CreateAccountModal, {
    props: { show: true, proxies: [], groups: [] },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub, OAuthAuthorizationFlow: OAuthAuthorizationFlowStub, Toggle: true, ConfirmDialog: true, Select: SelectStub, PlatformIcon: true, Icon: true,
        ProxySelector: true, GroupSelector: true, ModelWhitelistSelector: true, AccountCapabilitySelector: true,
        QuotaLimitCard: true, ShareDisplayCard: true,
      },
    },
  })
}

describe('CreateAccountModal OpenAI billing default', () => {
  beforeEach(() => {
    createAccount.mockReset().mockResolvedValue({ id: 7 })
    importCodexSession.mockReset().mockResolvedValue({ created: 1, updated: 0, skipped: 0, failed: 0, errors: [], warnings: [] })
  })

  async function selectOpenAI(wrapper: ReturnType<typeof mountModal>) {
    await wrapper.findAll('[data-tour="account-form-platform"] button')[1].trigger('click')
  }

  async function openCodexImport(wrapper: ReturnType<typeof mountModal>, clicks = 0) {
    await selectOpenAI(wrapper)
    for (let index = 0; index < clicks; index += 1) {
      await wrapper.get('[data-testid="openai-long-context-billing-toggle"]').trigger('click')
    }
    await wrapper.get('[data-tour="account-form-name"]').setValue('Codex import')
    await wrapper.get('#create-account-form').trigger('submit.prevent')
    await flushPromises()
  }

  it('writes an explicit disabled long-context billing flag for a new OpenAI API key', async () => {
    const wrapper = mountModal()
    await flushPromises()
    await wrapper.get('[data-tour="account-form-name"]').setValue('OpenAI default off')
    await selectOpenAI(wrapper)
    await wrapper.findAll('[data-tour="account-form-type"] button')[1].trigger('click')
    await wrapper.get('input[type="password"]').setValue('sk-test')
    await wrapper.get('#create-account-form').trigger('submit.prevent')
    await flushPromises()

    expect(createAccount).toHaveBeenCalledWith(expect.objectContaining({
      platform: 'openai',
      extra: expect.objectContaining({ openai_long_context_billing_enabled: false }),
    }))
  })

  it('writes an explicit enabled long-context billing flag after opt-in', async () => {
    const wrapper = mountModal()
    await flushPromises()
    await wrapper.get('[data-tour="account-form-name"]').setValue('OpenAI opt in')
    await selectOpenAI(wrapper)
    await wrapper.findAll('[data-tour="account-form-type"] button')[1].trigger('click')
    await wrapper.get('[data-testid="openai-long-context-billing-toggle"]').trigger('click')
    await wrapper.get('input[type="password"]').setValue('sk-test')
    await wrapper.get('#create-account-form').trigger('submit.prevent')
    await flushPromises()

    expect(createAccount).toHaveBeenCalledWith(expect.objectContaining({
      platform: 'openai',
      extra: expect.objectContaining({ openai_long_context_billing_enabled: true }),
    }))
  })

  it('does not inject the OpenAI setting for another platform', async () => {
    const wrapper = mountModal()
    await flushPromises()
    await wrapper.get('[data-tour="account-form-name"]').setValue('Anthropic account')
    await wrapper.findAll('[data-tour="account-form-type"] button')[1].trigger('click')
    await wrapper.get('input[type="password"]').setValue('sk-ant-test')
    await wrapper.get('#create-account-form').trigger('submit.prevent')
    await flushPromises()

    expect(createAccount).toHaveBeenCalledWith(expect.objectContaining({
      platform: 'anthropic',
      extra: expect.not.objectContaining({ openai_long_context_billing_enabled: expect.anything() }),
    }))
  })

  it('leaves Codex session import billing ownership to the backend until the toggle is touched', async () => {
    const wrapper = mountModal()
    await openCodexImport(wrapper)
    await wrapper.get('[data-testid="import-codex-session"]').trigger('click')
    await flushPromises()

    expect(importCodexSession).toHaveBeenCalledWith(expect.objectContaining({
      extra: expect.not.objectContaining({ openai_long_context_billing_enabled: expect.anything() }),
    }))
  })

  it('writes the chosen value for Codex session import after the toggle is touched', async () => {
    const wrapper = mountModal()
    await openCodexImport(wrapper, 1)
    await wrapper.get('[data-testid="import-codex-session"]').trigger('click')
    await flushPromises()

    expect(importCodexSession).toHaveBeenCalledWith(expect.objectContaining({
      extra: expect.objectContaining({ openai_long_context_billing_enabled: true }),
    }))
  })

  it('writes false for Codex session import after an opt-in is toggled back off', async () => {
    const wrapper = mountModal()
    await openCodexImport(wrapper, 2)
    await wrapper.get('[data-testid="import-codex-session"]').trigger('click')
    await flushPromises()

    expect(importCodexSession).toHaveBeenCalledWith(expect.objectContaining({
      extra: expect.objectContaining({ openai_long_context_billing_enabled: false }),
    }))
  })

  it('keeps the OAuth fingerprint mode absent by default and persists an explicit opt-in', async () => {
    const wrapper = mountModal()
    await selectOpenAI(wrapper)
    expect(wrapper.get('[data-testid="create-codex-fingerprint-mode-select"]').element).toBeInstanceOf(HTMLSelectElement)

    await wrapper.get('[data-testid="create-codex-fingerprint-mode-select"]').setValue('full')
    await wrapper.get('[data-tour="account-form-name"]').setValue('Codex fingerprint import')
    await wrapper.get('#create-account-form').trigger('submit.prevent')
    await flushPromises()
    await wrapper.get('[data-testid="import-codex-session"]').trigger('click')
    await flushPromises()

    expect(importCodexSession).toHaveBeenCalledWith(expect.objectContaining({
      extra: expect.objectContaining({ codex_fingerprint_mode: 'full' }),
    }))
  })
})
