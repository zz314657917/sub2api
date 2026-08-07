import { describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

const { updateAccountMock, checkMixedChannelRiskMock } = vi.hoisted(() => ({
  updateAccountMock: vi.fn(),
  checkMixedChannelRiskMock: vi.fn()
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn()
  })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    isSimpleMode: true
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      update: updateAccountMock,
      checkMixedChannelRisk: checkMixedChannelRiskMock
    },
    settings: {
      getWebSearchEmulationConfig: vi.fn().mockResolvedValue({ enabled: false, providers: [] }),
      getSettings: vi.fn().mockResolvedValue({})
    },
    tlsFingerprintProfiles: {
      list: vi.fn().mockResolvedValue([])
    }
  }
}))

vi.mock('@/api/admin/accounts', () => ({
  getAntigravityDefaultModelMapping: vi.fn()
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

import EditAccountModal from '../EditAccountModal.vue'
import ShareDisplayCard from '../ShareDisplayCard.vue'

const BaseDialogStub = defineComponent({
  name: 'BaseDialog',
  props: {
    show: {
      type: Boolean,
      default: false
    }
  },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})

const ModelWhitelistSelectorStub = defineComponent({
  name: 'ModelWhitelistSelector',
  props: {
    modelValue: {
      type: Array,
      default: () => []
    }
  },
  emits: ['update:modelValue'],
  template: `
    <div>
      <button
        type="button"
        data-testid="rewrite-to-snapshot"
        @click="$emit('update:modelValue', ['gpt-5.2-2025-12-11'])"
      >
        rewrite
      </button>
      <span data-testid="model-whitelist-value">
        {{ Array.isArray(modelValue) ? modelValue.join(',') : '' }}
      </span>
    </div>
  `
})

const SelectStub = defineComponent({
  name: 'SelectStub',
  props: {
    modelValue: {
      type: [String, Number, Boolean, null],
      default: ''
    },
    options: {
      type: Array,
      default: () => []
    },
    disabled: {
      type: Boolean,
      default: false
    }
  },
  emits: ['update:modelValue'],
  template: `
    <select
      v-bind="$attrs"
      :value="modelValue"
      :disabled="disabled"
      @change="$emit('update:modelValue', $event.target.value)"
    >
      <option v-for="option in options" :key="option.value" :value="option.value">
        {{ option.label }}
      </option>
    </select>
  `
})

function buildAccount() {
  return {
    id: 1,
    name: 'OpenAI Key',
    notes: '',
    platform: 'openai',
    type: 'apikey',
    credentials: {
      base_url: 'https://api.openai.com',
      model_mapping: {
        'gpt-5.2': 'gpt-5.2'
      }
    },
    credentials_status: {
      has_api_key: true
    },
    extra: {},
    proxy_id: null,
    concurrency: 1,
    priority: 1,
    rate_multiplier: 1,
    status: 'active',
    group_ids: [],
    expires_at: null,
    auto_pause_on_expired: false
  } as any
}

function buildOpenAIAccount(type: 'apikey' | 'oauth' | 'setup-token') {
  const account = buildAccount()
  account.type = type
  if (type !== 'apikey') {
    account.name = type === 'oauth' ? 'OpenAI OAuth' : 'OpenAI Setup Token'
    account.credentials = {
      access_token: 'access-token',
      refresh_token: 'refresh-token'
    }
    delete account.credentials_status
  }
  return account
}

function buildVertexAccount() {
  return {
    id: 2,
    name: 'Vertex SA',
    notes: '',
    platform: 'gemini',
    type: 'service_account',
    credentials: {
      service_account_json: '{"type":"service_account","client_email":"sa@example.iam.gserviceaccount.com","private_key":"-----BEGIN PRIVATE KEY-----\\nMIIE\\n-----END PRIVATE KEY-----\\n"}',
      project_id: 'demo-project',
      client_email: 'sa@example.iam.gserviceaccount.com',
      location: 'us-central1',
      tier_id: 'vertex'
    },
    extra: {},
    proxy_id: null,
    concurrency: 1,
    priority: 1,
    rate_multiplier: 1,
    status: 'active',
    group_ids: [],
    expires_at: null,
    auto_pause_on_expired: false
  } as any
}

function mountModal(account = buildAccount()) {
  return mount(EditAccountModal, {
    props: {
      show: true,
      account,
      proxies: [],
      groups: []
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        Select: SelectStub,
        Icon: true,
        ProxySelector: true,
        GroupSelector: true,
        ModelWhitelistSelector: ModelWhitelistSelectorStub,
        Input: {
          props: ['modelValue', 'label', 'placeholder', 'dataTestid', 'type', 'hint'],
          emits: ['update:modelValue'],
          template: '<input v-bind="$attrs" :data-testid="dataTestid" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />'
        }
      }
    }
  })
}

describe('EditAccountModal', () => {
  it('reopening the same account rehydrates the OpenAI whitelist from props', async () => {
    const account = buildAccount()
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    expect(wrapper.get('[data-testid="model-whitelist-value"]').text()).toBe('gpt-5.2')

    await wrapper.get('[data-testid="rewrite-to-snapshot"]').trigger('click')
    expect(wrapper.get('[data-testid="model-whitelist-value"]').text()).toBe('gpt-5.2-2025-12-11')

    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })

    expect(wrapper.get('[data-testid="model-whitelist-value"]').text()).toBe('gpt-5.2')

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.credentials?.model_mapping).toEqual({
      'gpt-5.2': 'gpt-5.2'
    })
  })

  it('preserves model mappings when editing the whitelist', async () => {
    const account = buildAccount()
    account.credentials.model_mapping = {
      'gpt-5.2': 'gpt-5.2',
      'gpt-latest': 'gpt-5.2'
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    expect(wrapper.get('[data-testid="model-whitelist-value"]').text()).toBe('gpt-5.2')

    await wrapper.get('[data-testid="rewrite-to-snapshot"]').trigger('click')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.credentials?.model_mapping).toEqual({
      'gpt-5.2-2025-12-11': 'gpt-5.2-2025-12-11',
      'gpt-latest': 'gpt-5.2'
    })
  })

  it('submits OpenAI compact mode and compact-only model mapping', async () => {
    const account = buildAccount()
    account.extra = {
      openai_compact_mode: 'force_on'
    }
    account.credentials = {
      ...account.credentials,
      compact_model_mapping: {
        'gpt-5.4': 'gpt-5.4-openai-compact'
      }
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra?.openai_compact_mode).toBe('force_on')
    expect(updateAccountMock.mock.calls[0]?.[1]?.credentials?.compact_model_mapping).toEqual({
      'gpt-5.4': 'gpt-5.4-openai-compact'
    })
  })

  it('submits OpenAI APIKey Responses support override mode', async () => {
    const account = buildAccount()
    account.extra = {
      openai_responses_mode: 'force_chat_completions',
      openai_responses_supported: false
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('[data-testid="openai-responses-mode-select"]').setValue('force_responses')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra?.openai_responses_mode).toBe('force_responses')
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra?.openai_responses_supported).toBe(false)
  })

  it('submits the account upstream billing auto-probe setting', async () => {
    const account = buildAccount()
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)
    const toggle = wrapper.get('[data-testid="upstream-billing-auto-probe"]')
    expect(toggle.attributes('aria-checked')).toBe('false')

    await toggle.trigger('click')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.upstream_billing_probe_enabled).toBe(true)
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra).not.toHaveProperty(
      'upstream_billing_probe_enabled'
    )
  })

  it('exposes the upstream billing auto-probe toggle for non-OpenAI API-key accounts', async () => {
    // 探测已放宽到全部 API-key 平台：grok 账号同样能开启并保存。
    const account = buildAccount()
    account.platform = 'grok'
    account.name = 'grok-relay'
    account.credentials = { api_key: 'sk-grok', base_url: 'https://relay.example/v1' }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)
    const toggle = wrapper.get('[data-testid="upstream-billing-auto-probe"]')
    expect(toggle.attributes('aria-checked')).toBe('false')

    await toggle.trigger('click')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.upstream_billing_probe_enabled).toBe(true)
  })

  it('enabling rate sync also enables probing and stops submitting a manual rate', async () => {
    const account = buildAccount()
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)
    const syncToggle = wrapper.get('[data-testid="upstream-billing-rate-sync"]')
    const probeToggle = wrapper.get('[data-testid="upstream-billing-auto-probe"]')
    const rateInput = wrapper.get<HTMLInputElement>('[data-testid="account-rate-multiplier"]')
    expect(syncToggle.attributes('aria-checked')).toBe('false')
    expect(probeToggle.attributes('aria-checked')).toBe('false')
    expect(rateInput.element.disabled).toBe(false)
    expect(wrapper.text()).toContain('admin.accounts.billingRateMultiplierHint')
    expect(wrapper.text()).not.toContain('admin.accounts.upstreamBilling.syncRateManagedHint')

    await syncToggle.trigger('click')
    expect(syncToggle.attributes('aria-checked')).toBe('true')
    expect(probeToggle.attributes('aria-checked')).toBe('true')
    expect(rateInput.element.disabled).toBe(true)
    expect(wrapper.text()).toContain('admin.accounts.upstreamBilling.syncRateManagedHint')
    expect(wrapper.text()).not.toContain('admin.accounts.billingRateMultiplierHint')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    const payload = updateAccountMock.mock.calls[0]?.[1]
    expect(payload?.upstream_billing_probe_enabled).toBe(true)
    expect(payload?.upstream_billing_rate_sync_enabled).toBe(true)
    expect(payload).not.toHaveProperty('rate_multiplier')
  })

  it('disabling probing also disables rate sync and restores manual rate editing', async () => {
    const account = buildAccount()
    account.extra = {
      upstream_billing_probe_enabled: true,
      upstream_billing_rate_sync_enabled: true
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)
    const syncToggle = wrapper.get('[data-testid="upstream-billing-rate-sync"]')
    const probeToggle = wrapper.get('[data-testid="upstream-billing-auto-probe"]')
    const rateInput = wrapper.get<HTMLInputElement>('[data-testid="account-rate-multiplier"]')
    expect(syncToggle.attributes('aria-checked')).toBe('true')
    expect(rateInput.element.disabled).toBe(true)

    await probeToggle.trigger('click')
    expect(probeToggle.attributes('aria-checked')).toBe('false')
    expect(syncToggle.attributes('aria-checked')).toBe('false')
    expect(rateInput.element.disabled).toBe(false)
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    const payload = updateAccountMock.mock.calls[0]?.[1]
    expect(payload?.upstream_billing_probe_enabled).toBe(false)
    expect(payload?.upstream_billing_rate_sync_enabled).toBe(false)
    expect(payload?.rate_multiplier).toBe(1)
  })

  it('disabling only rate sync keeps automatic probing enabled', async () => {
    const account = buildAccount()
    account.rate_multiplier = 0.0655
    account.extra = {
      upstream_billing_probe_enabled: true,
      upstream_billing_rate_sync_enabled: true
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)
    await wrapper.get('[data-testid="upstream-billing-rate-sync"]').trigger('click')
    expect(wrapper.get('[data-testid="upstream-billing-auto-probe"]').attributes('aria-checked')).toBe(
      'true'
    )
    const rateInput = wrapper.get<HTMLInputElement>('[data-testid="account-rate-multiplier"]')
    expect(rateInput.attributes('step')).toBe('0.0001')
    expect(rateInput.element.validity.valid).toBe(true)
    const form = wrapper.get<HTMLFormElement>('form#edit-account-form')
    const submitButton = wrapper.get<HTMLButtonElement>('button[type="submit"][form="edit-account-form"]')
    form.element.requestSubmit(submitButton.element)
    await flushPromises()

    const payload = updateAccountMock.mock.calls[0]?.[1]
    expect(payload?.upstream_billing_probe_enabled).toBe(true)
    expect(payload?.upstream_billing_rate_sync_enabled).toBe(false)
    expect(payload?.rate_multiplier).toBe(0.0655)
  })

  it('clears OpenAI APIKey Responses override when set back to auto', async () => {
    const account = buildAccount()
    account.extra = {
      openai_responses_mode: 'force_chat_completions',
      openai_responses_supported: true
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('[data-testid="openai-responses-mode-select"]').setValue('auto')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra).not.toHaveProperty('openai_responses_mode')
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra?.openai_responses_supported).toBe(true)
  })

  it('submits OpenAI APIKey endpoint capability overrides', async () => {
    const account = buildAccount()
    account.extra = {
      openai_responses_supported: true
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('[data-testid="openai-endpoint-capability-chat_completions"]').setValue(false)
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.credentials?.openai_capabilities).toEqual(['embeddings'])
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra).not.toHaveProperty('openai_responses_mode')
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra?.openai_responses_supported).toBe(true)
    expect(wrapper.get('[data-testid="openai-responses-mode-select"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="openai-responses-mode-not-applicable"]').exists()).toBe(true)
  })

  it('omits OpenAI endpoint capability override when both defaults are selected', async () => {
    const account = buildAccount()
    account.credentials = {
      ...account.credentials,
      openai_capabilities: ['embeddings']
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('[data-testid="openai-endpoint-capability-chat_completions"]').setValue(true)
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.credentials).not.toHaveProperty('openai_capabilities')
  })

  it('submits account-level Codex image generation bridge override', async () => {
    const account = buildAccount()
    account.extra = {
      codex_image_generation_bridge: false,
      codex_image_generation_bridge_enabled: true
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('button[data-testid="codex-image-bridge-enabled"]').trigger('click')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra?.codex_image_generation_bridge).toBe(true)
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra).not.toHaveProperty('codex_image_generation_bridge_enabled')
  })

  for (const accountType of ['apikey', 'oauth', 'setup-token'] as const) {
    it(`loads and clears Codex image tool strip policy for OpenAI ${accountType}`, async () => {
      const account = buildOpenAIAccount(accountType)
      account.extra = {
        codex_image_generation_explicit_tool_policy: 'strip',
        unknown_policy_neighbor: `${accountType}-keep`
      }
      updateAccountMock.mockReset()
      checkMixedChannelRiskMock.mockReset()
      checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
      updateAccountMock.mockResolvedValue(account)

      const wrapper = mountModal(account)

      expect(wrapper.get('[data-testid="codex-image-tool-policy-strip"]').attributes('aria-pressed')).toBe('true')
      await wrapper.get('[data-testid="codex-image-tool-policy-allow"]').trigger('click')
      await wrapper.get('form#edit-account-form').trigger('submit.prevent')

      expect(updateAccountMock).toHaveBeenCalledTimes(1)
      expect(updateAccountMock.mock.calls[0]?.[1]?.extra?.unknown_policy_neighbor).toBe(`${accountType}-keep`)
      expect(updateAccountMock.mock.calls[0]?.[1]?.extra).not.toHaveProperty('codex_image_generation_explicit_tool_policy')
    })

    it(`defaults to allow and saves Codex image tool strip policy for OpenAI ${accountType}`, async () => {
      const account = buildOpenAIAccount(accountType)
      account.extra = {
        codex_image_generation_explicit_tool_policy: 'unexpected',
        unknown_policy_neighbor: `${accountType}-keep`,
        openai: {
          codex_image_generation_explicit_tool_policy: 'strip',
          nested_policy_neighbor: `${accountType}-nested-keep`
        }
      }
      updateAccountMock.mockReset()
      checkMixedChannelRiskMock.mockReset()
      checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
      updateAccountMock.mockResolvedValue(account)

      const wrapper = mountModal(account)

      expect(wrapper.get('[data-testid="codex-image-tool-policy-allow"]').attributes('aria-pressed')).toBe('true')
      await wrapper.get('[data-testid="codex-image-tool-policy-strip"]').trigger('click')
      await wrapper.get('form#edit-account-form').trigger('submit.prevent')

      expect(updateAccountMock).toHaveBeenCalledTimes(1)
      expect(updateAccountMock.mock.calls[0]?.[1]?.extra).toEqual(expect.objectContaining({
        codex_image_generation_explicit_tool_policy: 'strip',
        unknown_policy_neighbor: `${accountType}-keep`,
        openai: {
          nested_policy_neighbor: `${accountType}-nested-keep`
        }
      }))
    })
  }

  it('loads a nested-only Codex image tool strip policy', () => {
    const account = buildOpenAIAccount('oauth')
    account.extra = {
      openai: {
        codex_image_generation_explicit_tool_policy: 'strip'
      }
    }

    const wrapper = mountModal(account)

    expect(wrapper.get('[data-testid="codex-image-tool-policy-strip"]').attributes('aria-pressed')).toBe('true')
  })

  it('clears top-level and nested Codex image tool policies while preserving neighboring extra keys', async () => {
    const account = buildOpenAIAccount('setup-token')
    account.extra = {
      codex_image_generation_explicit_tool_policy: 'strip',
      top_level_neighbor: 'top-keep',
      openai: {
        codex_image_generation_explicit_tool_policy: 'strip',
        nested_neighbor: 'nested-keep'
      }
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('[data-testid="codex-image-tool-policy-allow"]').trigger('click')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    const submittedExtra = updateAccountMock.mock.calls[0]?.[1]?.extra
    expect(submittedExtra).not.toHaveProperty('codex_image_generation_explicit_tool_policy')
    expect(submittedExtra).toEqual(expect.objectContaining({
      top_level_neighbor: 'top-keep',
      openai: {
        nested_neighbor: 'nested-keep'
      }
    }))
  })

  it('prefers an unknown top-level Codex image tool policy over nested strip', () => {
    const account = buildOpenAIAccount('apikey')
    account.extra = {
      codex_image_generation_explicit_tool_policy: 'unexpected',
      openai: {
        codex_image_generation_explicit_tool_policy: 'strip'
      }
    }

    const wrapper = mountModal(account)

    expect(wrapper.get('[data-testid="codex-image-tool-policy-allow"]').attributes('aria-pressed')).toBe('true')
    expect(wrapper.get('[data-testid="codex-image-tool-policy-strip"]').attributes('aria-pressed')).toBe('false')
  })

  it('normalizes a top-level REMOVE Codex image tool policy alias and saves canonical strip', async () => {
    const account = buildOpenAIAccount('oauth')
    account.extra = {
      codex_image_generation_explicit_tool_policy: ' REMOVE ',
      unknown_policy_neighbor: 'keep'
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    expect(wrapper.get('[data-testid="codex-image-tool-policy-strip"]').attributes('aria-pressed')).toBe('true')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock.mock.calls[0]?.[1]?.extra).toEqual(expect.objectContaining({
      codex_image_generation_explicit_tool_policy: 'strip',
      unknown_policy_neighbor: 'keep'
    }))
  })

  it('normalizes a nested Drop Codex image tool policy alias and saves canonical strip', async () => {
    const account = buildOpenAIAccount('setup-token')
    account.extra = {
      openai: {
        codex_image_generation_explicit_tool_policy: 'Drop',
        nested_neighbor: 'keep'
      }
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    expect(wrapper.get('[data-testid="codex-image-tool-policy-strip"]').attributes('aria-pressed')).toBe('true')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock.mock.calls[0]?.[1]?.extra).toEqual(expect.objectContaining({
      codex_image_generation_explicit_tool_policy: 'strip',
      openai: {
        nested_neighbor: 'keep'
      }
    }))
  })

  it('keeps setup-token Codex image policy controls isolated from other OpenAI settings', () => {
    const wrapper = mountModal(buildOpenAIAccount('setup-token'))

    expect(wrapper.get('[data-testid="codex-image-tool-policy-allow"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="codex-image-bridge-inherit"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="openai-responses-mode-select"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="openai-image-input-object-url-toggle"]').exists()).toBe(false)
  })

  it('submits OpenAI APIKey image input object URL marker', async () => {
    const account = buildAccount()
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('[data-testid="openai-image-input-object-url-toggle"]').trigger('click')
    await wrapper.get('[data-testid="openai-image-upload-limit-bytes"]').setValue('1048576')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra).toEqual(expect.objectContaining({
      image_input_transport: 'object_url',
      image_upload_limit_bytes: 1048576,
      image_url_fields_supported: true
    }))
  })

  it('clears OpenAI APIKey image input object URL marker when disabled', async () => {
    const account = buildAccount()
    account.extra = {
      image_input_transport: 'object_url',
      image_upload_limit_bytes: 1048576,
      image_url_fields_supported: true
    }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('[data-testid="openai-image-input-object-url-toggle"]').trigger('click')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra).not.toHaveProperty('image_input_transport')
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra).not.toHaveProperty('image_upload_limit_bytes')
    expect(updateAccountMock.mock.calls[0]?.[1]?.extra).not.toHaveProperty('image_url_fields_supported')
  })

  it('moves a shared OpenAI OAuth account to the selected capacity pool', async () => {
    const account = {
      ...buildAccount(),
      id: 9,
      name: 'openai oauth Account',
      type: 'oauth',
      owner_user_id: 10,
      share_mode: 'public',
      share_status: 'active',
      credentials: {
        access_token: 'token',
        refresh_token: 'refresh'
      },
      extra: {}
    } as any
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()

    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    expect(wrapper.getComponent(ShareDisplayCard).exists()).toBe(true)

    await wrapper.get('[data-testid="share-display-target-pool"]').setValue('plus')
    await wrapper.get('[data-testid="share-display-5h-limit"]').setValue('500')
    await wrapper.get('[data-testid="share-display-5h-used"]').setValue('95.17')
    await wrapper.get('[data-testid="share-display-7d-limit"]').setValue('2160')
    await wrapper.get('[data-testid="share-display-7d-used"]').setValue('95.17')
    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock).toHaveBeenCalledWith(9, expect.objectContaining({
      extra: expect.objectContaining({
        share_display_tier: 'plus',
        share_display_percent_only: true,
        share_display_account_count: 1,
        share_display_5h_limit: 500,
        share_display_5h_used: 95.17,
        share_display_7d_limit: 2160,
        share_display_7d_used: 95.17
      })
    }))
  })

  it('rehydrates OpenAI OAuth capacity pool settings from flattened response fields', async () => {
    const account = {
      ...buildAccount(),
      id: 10,
      name: 'openai oauth Account',
      type: 'oauth',
      credentials: {
        access_token: 'token',
        refresh_token: 'refresh'
      },
      share_display_tier: 'plus',
      share_display_percent_only: true,
      share_display_account_count: 2,
      share_display_5h_limit: 500,
      share_display_5h_used: 95.17,
      share_display_7d_limit: 2160,
      share_display_7d_used: 95.17,
      extra: {}
    } as any
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    expect((wrapper.get('[data-testid="share-display-target-pool"]').element as HTMLSelectElement).value).toBe('plus')
    expect((wrapper.get('[data-testid="share-display-account-count"]').element as HTMLInputElement).value).toBe('2')
    expect((wrapper.get('[data-testid="share-display-5h-limit"]').element as HTMLInputElement).value).toBe('500')
    expect((wrapper.get('[data-testid="share-display-5h-used"]').element as HTMLInputElement).value).toBe('95.17')
    expect((wrapper.get('[data-testid="share-display-7d-limit"]').element as HTMLInputElement).value).toBe('2160')
    expect((wrapper.get('[data-testid="share-display-7d-used"]').element as HTMLInputElement).value).toBe('95.17')

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledWith(10, expect.objectContaining({
      extra: expect.objectContaining({
        share_display_tier: 'plus',
        share_display_account_count: 2,
        share_display_5h_limit: 500,
        share_display_5h_used: 95.17,
        share_display_7d_limit: 2160,
        share_display_7d_used: 95.17
      })
    }))
  })

  it('allows saving apikey account when backend redacted api_key but credentials_status reports it exists', async () => {
    // 新前端 + 新后端：响应已脱敏，credentials 里没有 api_key，credentials_status.has_api_key=true
    const account = buildAccount()
    account.credentials = {
      base_url: 'https://api.openai.com',
      model_mapping: { 'gpt-5.2': 'gpt-5.2' }
    }
    account.credentials_status = { has_api_key: true }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    // 用户未输入新 key 时，payload 不应带 api_key，由后端合并保留旧值
    expect(updateAccountMock.mock.calls[0]?.[1]?.credentials).not.toHaveProperty('api_key')
  })

  it('allows saving apikey account against legacy backend without credentials_status', async () => {
    // 新前端 + 旧后端：credentials_status 缺失，但 credentials.api_key 仍是明文，应允许保存
    const account = buildAccount()
    account.credentials = {
      ...account.credentials,
      api_key: 'sk-test'
    }
    delete account.credentials_status
    // 显式确保没有 credentials_status
    expect(account.credentials_status).toBeUndefined()
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    // 旧后端响应未脱敏，原 api_key 会随 currentCredentials 一起传回去（旧行为，等价于无操作）
    expect(updateAccountMock.mock.calls[0]?.[1]?.credentials?.api_key).toBe('sk-test')
  })

  it('blocks apikey save when neither credentials_status nor legacy api_key indicates existence', async () => {
    const account = buildAccount()
    account.credentials = {
      base_url: 'https://api.openai.com'
    }
    delete account.credentials_status
    // 既没有 credentials_status 也没有旧的 api_key
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })

    const wrapper = mountModal(account)

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).not.toHaveBeenCalled()
  })

  it('allows saving Vertex SA account when backend redacted service_account_json but credentials_status reports it exists', async () => {
    // 新前端 + 新后端：响应已脱敏，credentials 里没有 service_account_json，credentials_status.has_service_account_json=true
    const account = buildVertexAccount()
    account.credentials = {
      project_id: 'demo-project',
      client_email: 'sa@example.iam.gserviceaccount.com',
      location: 'us-central1',
      tier_id: 'vertex'
    }
    account.credentials_status = { has_service_account_json: true }
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
    expect(updateAccountMock.mock.calls[0]?.[1]?.credentials?.project_id).toBe('demo-project')
  })

  it('allows saving Vertex SA account against legacy backend without credentials_status', async () => {
    // 新前端 + 旧后端：credentials_status 缺失，但 credentials.service_account_json 仍是明文，应允许保存
    const account = buildVertexAccount()
    expect(account.credentials_status).toBeUndefined()
    expect(account.credentials.service_account_json).toBeTruthy()
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })
    updateAccountMock.mockResolvedValue(account)

    const wrapper = mountModal(account)

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).toHaveBeenCalledTimes(1)
  })

  it('blocks Vertex SA save when neither credentials_status nor legacy json indicates existence', async () => {
    const account = buildVertexAccount()
    account.credentials = {
      project_id: 'demo-project',
      client_email: 'sa@example.iam.gserviceaccount.com',
      location: 'us-central1',
      tier_id: 'vertex'
    }
    // 既没有 credentials_status 也没有旧的 service_account_json
    updateAccountMock.mockReset()
    checkMixedChannelRiskMock.mockReset()
    checkMixedChannelRiskMock.mockResolvedValue({ has_risk: false })

    const wrapper = mountModal(account)

    await wrapper.get('form#edit-account-form').trigger('submit.prevent')

    expect(updateAccountMock).not.toHaveBeenCalled()
  })
})
