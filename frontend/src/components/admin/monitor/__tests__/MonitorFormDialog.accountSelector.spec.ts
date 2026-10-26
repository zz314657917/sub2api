import { defineComponent, nextTick } from 'vue'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { ChannelMonitor } from '@/api/admin/channelMonitor'
import MonitorFormDialog from '@/components/admin/monitor/MonitorFormDialog.vue'

const {
  listTemplates,
  accountsList,
  accountsGetById,
  monitorCreate,
  monitorUpdate,
} = vi.hoisted(() => ({
  listTemplates: vi.fn(),
  accountsList: vi.fn(),
  accountsGetById: vi.fn(),
  monitorCreate: vi.fn(),
  monitorUpdate: vi.fn(),
}))




vi.mock('@/api/admin', () => ({
  adminAPI: {
    channelMonitor: {
      create: monitorCreate,
      update: monitorUpdate,
    },
    channelMonitorTemplate: {
      list: listTemplates,
    },
    accounts: {
      list: (...args: unknown[]) => accountsList(...args),
      getById: (...args: unknown[]) => accountsGetById(...args),
    },
  },
}))

vi.mock('@/api/keys', () => ({
  keysAPI: { list: vi.fn() },
}))

vi.mock('@/api/groups', () => ({
  userGroupsAPI: { getUserGroupRates: vi.fn() },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: null,
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const BaseDialogStub = defineComponent({
  props: { show: { type: Boolean, default: false } },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
})

function makeMonitor(overrides: Partial<ChannelMonitor> = {}): ChannelMonitor {
  return {
    id: 42,
    name: 'primary',
    provider: 'openai',
    api_mode: 'chat_completions',
    endpoint: 'https://api.example.com',
    api_key_masked: 'sk-t***',
    primary_model: 'gpt-4o-mini',
    extra_models: [],
    group_name: '',
    enabled: true,
    interval_seconds: 60,
    jitter_seconds: 0,
    last_checked_at: null,
    created_by: 1,
    created_at: '2026-07-16T00:00:00Z',
    updated_at: '2026-07-16T00:00:00Z',
    primary_status: '',
    primary_latency_ms: null,
    availability_7d: 0,
    extra_models_status: [],
    template_id: null,
    extra_headers: {},
    body_override_mode: 'off',
    body_override: null,
    check_mode: 'probe',
    account_id: null,
    ...overrides,
  }
}

let unmountWrapper: (() => void) | undefined

function mountDialog(monitor: ChannelMonitor | null = null) {
  const wrapper = mount(MonitorFormDialog, {
    props: { show: true, monitor },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        Toggle: true,
        ModelTagInput: true,
        MonitorKeyPickerDialog: true,
        MonitorAdvancedRequestConfig: true,
      },
    },
  })
  unmountWrapper = () => wrapper.unmount()
  return wrapper
}

const accountTrigger = (wrapper: VueWrapper) =>
  wrapper.get('[data-testid="monitor-linked-account"] button')

const openAccountDropdown = async (wrapper: VueWrapper) => {
  await accountTrigger(wrapper).trigger('click')
  await nextTick()
  const dropdown = document.body.querySelector<HTMLElement>('.select-dropdown-portal')
  expect(dropdown).not.toBeNull()
  return dropdown as HTMLElement
}

const clickOption = (dropdown: HTMLElement, text: string) => {
  const option = [...dropdown.querySelectorAll('.select-option')].find((el) =>
    el.textContent?.includes(text),
  )
  expect(option, `option containing "${text}" not found`).toBeDefined()
  ;(option as HTMLElement).click()
}

const typeAccountSearch = async (wrapper: VueWrapper, query: string) => {
  await wrapper.get('[data-testid="monitor-account-search"]').setValue(query)
  await nextTick()
}

afterEach(() => {
  unmountWrapper?.()
  unmountWrapper = undefined
  document.body.innerHTML = ''
  vi.useRealTimers()
  vi.restoreAllMocks()
})

describe('MonitorFormDialog linked account selector', () => {
  beforeEach(() => {
    listTemplates.mockReset().mockResolvedValue({ items: [] })
    accountsList.mockReset().mockResolvedValue({ items: [] })
    accountsGetById.mockReset()
    monitorCreate.mockReset().mockResolvedValue({})
    monitorUpdate.mockReset().mockResolvedValue({})
  })

  it('loads the first page of provider accounts when quota mode is enabled', async () => {
    accountsList.mockResolvedValue({ items: [{ id: 1, name: 'a', platform: 'anthropic', type: 'oauth' }] })
    const wrapper = mountDialog()
    await flushPromises()
    expect(accountsList).not.toHaveBeenCalled()

    await wrapper.get('[data-testid="monitor-check-mode-quota"]').trigger('click')
    await flushPromises()

    expect(accountsList).toHaveBeenCalledTimes(1)
    expect(accountsList).toHaveBeenCalledWith(
      1,
      50,
      { platform: 'anthropic' },
      { signal: expect.any(AbortSignal) },
    )
  })

  it('debounced typing hits server-side search and aborts the previous request', async () => {
    vi.useFakeTimers()
    const wrapper = mountDialog()
    await wrapper.get('[data-testid="monitor-check-mode-quota"]').trigger('click')
    await vi.advanceTimersByTimeAsync(0)

    await typeAccountSearch(wrapper, 'hidden')
    expect(accountsList).toHaveBeenCalledTimes(1, 'debounce window must not fire a request')

    await vi.advanceTimersByTimeAsync(300)

    expect(accountsList).toHaveBeenCalledTimes(2)
    const lastCall = accountsList.mock.calls[accountsList.mock.calls.length - 1]
    expect(lastCall[2]).toMatchObject({ platform: 'anthropic', search: 'hidden' })
    const firstSignal = (accountsList.mock.calls[0][3] as { signal: AbortSignal }).signal
    expect(firstSignal.aborted).toBe(true)
  })

  it('hydrates a bound account missing from the first page and keeps the binding', async () => {
    accountsList.mockResolvedValue({ items: [{ id: 1, name: 'first page account', platform: 'anthropic', type: 'oauth' }] })
    accountsGetById.mockResolvedValue({ id: 999, name: 'hidden gem', platform: 'anthropic', type: 'oauth' })
    const wrapper = mountDialog(makeMonitor({
      provider: 'anthropic',
      check_mode: 'quota',
      account_id: 999,
      endpoint: '',
      primary_model: 'quota',
    }))
    await flushPromises()

    expect(accountsGetById).toHaveBeenCalledWith(999)
    expect(accountTrigger(wrapper).text()).toContain('hidden gem (#999)')

    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).toHaveBeenCalledWith(42, expect.objectContaining({ account_id: 999 }))
  })

  it('shows a visible hint and blocks saving when the bound account cannot be loaded', async () => {
    accountsGetById.mockRejectedValue(new Error('gone'))
    const wrapper = mountDialog(makeMonitor({
      provider: 'anthropic',
      check_mode: 'quota',
      account_id: 999,
      endpoint: '',
      primary_model: 'quota',
    }))
    await flushPromises()

    expect(wrapper.text()).toContain('admin.channelMonitor.form.linkedAccountMissing')
    expect(accountTrigger(wrapper).text()).not.toContain('hidden gem')

    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).not.toHaveBeenCalled()
  })

  it('shows the OpenAI codex probe hint only in quota mode on openai', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    expect(wrapper.text()).not.toContain('openAIQuotaProbeHint')

    await wrapper.get('[data-testid="monitor-check-mode-quota"]').trigger('click')
    await flushPromises()
    // anthropic 平台不显示。
    expect(wrapper.text()).not.toContain('openAIQuotaProbeHint')

    await wrapper.get('[data-testid="monitor-provider-openai"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('openAIQuotaProbeHint')
  })

  it('keeps a picked account visible when a later search hides it', async () => {
    vi.useFakeTimers()
    accountsList
      .mockResolvedValueOnce({ items: [
        { id: 1, name: 'alpha', platform: 'anthropic', type: 'oauth' },
        { id: 2, name: 'beta', platform: 'anthropic', type: 'oauth' },
      ] })
      .mockResolvedValueOnce({ items: [{ id: 2, name: 'beta', platform: 'anthropic', type: 'oauth' }] })
    const wrapper = mountDialog()
    await wrapper.get('[data-testid="monitor-check-mode-quota"]').trigger('click')
    await vi.advanceTimersByTimeAsync(0)

    const dropdown = await openAccountDropdown(wrapper)
    clickOption(dropdown, 'alpha (#1)')
    await nextTick()
    expect(accountTrigger(wrapper).text()).toContain('alpha (#1)')

    // 搜索 beta：结果页只剩 beta，但已选中的 alpha 仍固定展示、提交仍保留绑定。
    await typeAccountSearch(wrapper, 'beta')
    await vi.advanceTimersByTimeAsync(300)

    expect(accountTrigger(wrapper).text()).toContain('alpha (#1)')

    await wrapper.findAll('input[type="text"]')[0].setValue('my monitor')
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await vi.advanceTimersByTimeAsync(0)
    expect(monitorCreate).toHaveBeenCalledWith(expect.objectContaining({ account_id: 1 }))
  })

  it('clears the account binding when the provider switches', async () => {
    accountsList.mockResolvedValue({ items: [{ id: 1, name: 'alpha', platform: 'anthropic', type: 'oauth' }] })
    const wrapper = mountDialog()
    await wrapper.get('[data-testid="monitor-check-mode-quota"]').trigger('click')
    await flushPromises()

    const dropdown = await openAccountDropdown(wrapper)
    clickOption(dropdown, 'alpha (#1)')
    await nextTick()
    expect(accountTrigger(wrapper).text()).toContain('alpha (#1)')

    await wrapper.get('[data-testid="monitor-provider-openai"]').trigger('click')
    await flushPromises()

    expect(accountTrigger(wrapper).text()).not.toContain('alpha')
    expect(accountTrigger(wrapper).text()).toContain('linkedAccountPlaceholder')
  })

  // P2-7(b)：antigravity 强制的 quota 在切走时要对称还原，占位模型一并清掉。
  it('restores probe mode and clears the quota placeholder when switching away from antigravity', async () => {
    accountsGetById.mockRejectedValue(new Error('gone'))
    const wrapper = mountDialog(makeMonitor({
      provider: 'antigravity',
      check_mode: 'quota',
      account_id: 7,
      endpoint: '',
      primary_model: 'quota',
    }))
    await flushPromises()

    await wrapper.get('[data-testid="monitor-provider-anthropic"]').trigger('click')
    await flushPromises()

    // 占位模型已被清空：probe 模式必填校验拦下提交，而不是拿 'quota' 探活。
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).not.toHaveBeenCalled()

    await wrapper.get('[data-testid="monitor-primary-model"]').setValue('claude-sonnet-4-5')
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).toHaveBeenCalledWith(42, expect.objectContaining({
      check_mode: 'probe',
      primary_model: 'claude-sonnet-4-5',
    }))
  })

  // P2-7(c)：update 切回 probe 显式发 account_id=0 解绑存量关联（null=不动）。
  it('unbinds the linked account with account_id=0 when a quota monitor switches back to probe', async () => {
    accountsGetById.mockRejectedValue(new Error('gone'))
    const wrapper = mountDialog(makeMonitor({
      provider: 'anthropic',
      check_mode: 'quota',
      account_id: 999,
      endpoint: '',
      primary_model: 'quota',
    }))
    await flushPromises()

    await wrapper.get('[data-testid="monitor-check-mode-probe"]').trigger('click')
    await wrapper.get('[data-testid="monitor-primary-model"]').setValue('claude-sonnet-4-5')
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()

    expect(monitorUpdate).toHaveBeenCalledWith(42, expect.objectContaining({ account_id: 0 }))
  })

  // P2-7(c)：create 绝不发 0（后端会把 0 存成 &0 触发外键违约），保持 null。
  it('keeps account_id null when creating a probe monitor', async () => {
    const wrapper = mountDialog()
    await flushPromises()

    await wrapper.findAll('input[type="text"]')[0].setValue('my monitor')
    await wrapper.get('[data-testid="monitor-primary-model"]').setValue('claude-sonnet-4-5')
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()

    expect(monitorCreate).toHaveBeenCalledWith(expect.objectContaining({
      check_mode: 'probe',
      account_id: null,
    }))
  })

  it('restores an existing Responses template snapshot without clearing it or sending an empty key', async () => {
    const wrapper = mountDialog(makeMonitor({
      api_mode: 'responses',
      template_id: 8,
      extra_headers: { 'X-Trace': 'persisted' },
      body_override_mode: 'merge',
      body_override: { max_output_tokens: 32 },
      jitter_seconds: 9,
    }))
    await flushPromises()
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).toHaveBeenCalledWith(42, expect.objectContaining({
      provider: 'openai',
      api_mode: 'responses',
      template_id: 8,
      extra_headers: { 'X-Trace': 'persisted' },
      body_override_mode: 'merge',
      body_override: { max_output_tokens: 32 },
      jitter_seconds: 9,
    }))
    expect(monitorUpdate.mock.calls[0][1]).not.toHaveProperty('api_key')
    expect(monitorUpdate.mock.calls[0][1]).not.toHaveProperty('clear_template')
  })

  it('preserves a typed key on protocol switch, clears incompatible snapshots and resets protocol on provider switch', async () => {
    const wrapper = mountDialog(makeMonitor({ template_id: 8, extra_headers: { 'X-Old': 'value' } }))
    await flushPromises()
    await wrapper.get('input[type="password"]').setValue('new-openai-key')
    const responses = wrapper.findAll('button').find(b => b.text().includes('form.apiModeResponsesHint'))
    expect(responses).toBeDefined()
    await responses!.trigger('click')
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).toHaveBeenLastCalledWith(42, expect.objectContaining({
      api_mode: 'responses', api_key: 'new-openai-key', clear_template: true,
      extra_headers: {}, body_override_mode: 'off', body_override: null,
    }))
    await wrapper.get('[data-testid="monitor-provider-kimi"]').trigger('click')
    expect((wrapper.get('input[type="password"]').element as HTMLInputElement).value).toBe('')
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).toHaveBeenLastCalledWith(42, expect.objectContaining({
      provider: 'kimi', api_mode: 'chat_completions', clear_template: true,
    }))
  })

  it('offers exactly nine providers and restricts MiniMax and Antigravity to supported modes', async () => {
    const wrapper = mountDialog()
    expect(wrapper.findAll('[data-testid^="monitor-provider-"]')).toHaveLength(9)
    expect(wrapper.find('[data-testid="monitor-provider-opencode_go"]').exists()).toBe(false)
    await wrapper.get('[data-testid="monitor-check-mode-quota"]').trigger('click')
    await wrapper.get('[data-testid="monitor-provider-minimax"]').trigger('click')
    expect(wrapper.get('[data-testid="monitor-check-mode-quota"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="monitor-check-mode-quota_probe"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="monitor-check-mode-probe"]').attributes('aria-pressed')).toBe('true')
    expect(wrapper.find('[data-testid="monitor-linked-account"]').exists()).toBe(false)
    expect((wrapper.get('[data-testid="monitor-endpoint"]').element as HTMLInputElement).value).toBe('https://api.minimaxi.com')
    await wrapper.get('[data-testid="monitor-provider-antigravity"]').trigger('click')
    expect(wrapper.get('[data-testid="monitor-check-mode-probe"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="monitor-check-mode-quota_probe"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="monitor-check-mode-quota"]').attributes('aria-pressed')).toBe('true')
    expect(wrapper.find('input[type="password"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="monitor-endpoint"]').exists()).toBe(false)
  })

  it.each([
    ['anthropic', 'oauth', undefined, true],
    ['anthropic', 'setup-token', undefined, true],
    ['anthropic', 'apikey', undefined, false],
    ['openai', 'oauth', undefined, true],
    ['openai', 'apikey', undefined, false],
    ['gemini', 'oauth', undefined, true],
    ['grok', 'oauth', undefined, true],
    ['grok', 'apikey', undefined, true],
    ['antigravity', 'oauth', undefined, true],
    ['kimi', 'apikey', 'coding', true],
    ['kimi', 'apikey', 'payg', true],
    ['kimi', 'oauth', 'coding', true],
    ['kimi', 'apikey', undefined, true],
    ['kimi', 'apikey', '', true],
    ['kimi', 'apikey', 'legacy', true],
    ['kimi', 'upstream', 'payg', true],
    ['zhipu', 'apikey', 'coding', true],
    ['zhipu', 'apikey', 'payg', false],
    ['zhipu', 'apikey', undefined, false],
    ['deepseek', 'apikey', 'payg', true],
    ['deepseek', 'apikey', undefined, true],
    ['deepseek', 'apikey', '', true],
    ['deepseek', 'apikey', 'legacy', true],
    ['deepseek', 'upstream', 'payg', true],
    ['deepseek', 'apikey', 'coding', false],
    ['deepseek', 'apikey', ' coding ', false],
  ] as const)('filters local quota capability for %s %s %s', async (provider, type, mode, eligible) => {
    accountsList.mockResolvedValue({ items: [
      { id: 7, name: 'candidate-account', platform: provider, type, credentials: { account_mode: mode } },
      { id: 99, name: 'foreign-account', platform: 'foreign', type: 'oauth' },
    ] })
    const wrapper = mountDialog(makeMonitor({
      provider, check_mode: 'quota', account_id: null, primary_model: 'quota',
    }))
    await flushPromises()
    const dropdown = await openAccountDropdown(wrapper)
    expect(dropdown.textContent?.includes('candidate-account')).toBe(eligible)
    expect(dropdown.textContent).not.toContain('foreign-account')
  })

  it.each([
    ['kimi', undefined, true],
    ['kimi', '', true],
    ['kimi', '  ', true],
    ['kimi', 'https://api.kimi.com/coding/v1', true],
    ['kimi', ' https://api.moonshot.cn/v1 ', true],
    ['kimi', 'https://api.moonshot.ai/v1', true],
    ['kimi', 'https://API.KIMI.COM/coding', true],
    ['kimi', 'https://relay.example/coding', false],
    ['kimi', 'https://kimi.com.attacker.example/coding', false],
    ['kimi', 'https://notkimi.com/coding', false],
    ['kimi', 'https://api.bigmodel.cn/api/coding/paas/v4', false],
    ['kimi', 'not a URL', false],
    ['kimi', '//api.kimi.com/coding', false],
    ['zhipu', 'https://open.bigmodel.cn/api/coding/paas/v4', true],
    ['zhipu', 'https://api.z.ai/api/coding/paas/v4', true],
    ['zhipu', 'https://Z.AI/api/coding/paas/v4', true],
    ['zhipu', 'https://relay.example/api/coding/paas/v4', false],
    ['zhipu', 'https://bigmodel.cn.attacker.example/api/coding', false],
    ['zhipu', 'https://notbigmodel.cn/api/coding', false],
    ['zhipu', 'https://api.kimi.com/coding/v1', false],
    ['zhipu', 'not a URL', false],
  ] as const)('matches the backend coding official-host guard: %s %s', async (provider, base, eligible) => {
    accountsList.mockResolvedValue({ items: [{
      id: 7, name: 'coding-account', platform: provider, type: 'apikey',
      credentials: { account_mode: ' coding ', base_url: base },
    }] })
    const wrapper = mountDialog(makeMonitor({
      provider, check_mode: 'quota', account_id: null, primary_model: 'quota',
    }))
    await flushPromises()
    const dropdown = await openAccountDropdown(wrapper)
    expect(dropdown.textContent?.includes('coding-account')).toBe(eligible)
  })

  it('rejects a stored coding account with a nonofficial base URL during hydration', async () => {
    accountsGetById.mockResolvedValue({
      id: 99, name: 'unsupported-relay', platform: 'kimi', type: 'apikey',
      credentials: { account_mode: 'coding', base_url: 'https://relay.example/coding' },
    })
    const wrapper = mountDialog(makeMonitor({
      provider: 'kimi', check_mode: 'quota', account_id: 99, primary_model: 'quota',
    }))
    await flushPromises()
    expect(accountTrigger(wrapper).text()).toContain('linkedAccountMissing')
    expect(accountTrigger(wrapper).text()).toContain('#99')
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).not.toHaveBeenCalled()
  })

  it.each([
    ['grok', 'apikey', undefined],
    ['kimi', 'apikey', undefined],
    ['deepseek', 'apikey', undefined],
  ] as const)('restores and saves a supported local %s account with type %s and mode %s', async (provider, type, mode) => {
    accountsGetById.mockResolvedValue({
      id: 99, name: 'supported-local-account', platform: provider, type,
      credentials: { account_mode: mode },
    })
    const wrapper = mountDialog(makeMonitor({
      provider, check_mode: 'quota', account_id: 99, primary_model: 'quota',
    }))
    await flushPromises()
    expect(accountTrigger(wrapper).text()).toContain('supported-local-account')
    expect(wrapper.text()).not.toContain('linkedAccountMissing')
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).toHaveBeenCalledWith(42, expect.objectContaining({
      provider, check_mode: 'quota', account_id: 99,
    }))
  })

  it('retains the invalid stored account id visibly and blocks saving an ineligible account', async () => {
    accountsGetById.mockResolvedValue({ id: 99, name: 'api-only', platform: 'openai', type: 'apikey' })
    const wrapper = mountDialog(makeMonitor({
      check_mode: 'quota', account_id: 99, primary_model: 'quota',
    }))
    await flushPromises()
    expect(accountTrigger(wrapper).text()).toContain('linkedAccountMissing')
    expect(accountTrigger(wrapper).text()).toContain('#99')
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).not.toHaveBeenCalled()
  })

  it('ignores a stale list after switching providers even when the old API ignores AbortSignal', async () => {
    let resolveOld!: (value: unknown) => void
    accountsList.mockReturnValueOnce(new Promise(resolve => { resolveOld = resolve }))
      .mockResolvedValueOnce({ items: [{ id: 2, name: 'current-account', platform: 'openai', type: 'oauth' }] })
    const wrapper = mountDialog(makeMonitor({
      provider: 'anthropic', check_mode: 'quota', account_id: null, primary_model: 'quota',
    }))
    await nextTick()
    await wrapper.get('[data-testid="monitor-provider-openai"]').trigger('click')
    await flushPromises()
    resolveOld({ items: [{ id: 1, name: 'old-account', platform: 'anthropic', type: 'oauth' }] })
    await flushPromises()
    const dropdown = await openAccountDropdown(wrapper)
    expect(dropdown.textContent).toContain('current-account')
    expect(dropdown.textContent).not.toContain('old-account')
  })

  it('ignores stale account hydration when switching editing targets with the same bound id', async () => {
    let rejectOld!: (reason: Error) => void
    accountsGetById.mockReturnValueOnce(new Promise((_resolve, reject) => { rejectOld = reject }))
      .mockResolvedValueOnce({ id: 99, name: 'current-account', platform: 'anthropic', type: 'oauth' })
    const wrapper = mountDialog(makeMonitor({
      provider: 'anthropic', check_mode: 'quota', account_id: 99, primary_model: 'quota',
    }))
    await flushPromises()
    await wrapper.setProps({ monitor: makeMonitor({
      id: 43, provider: 'anthropic', check_mode: 'quota', account_id: 99, primary_model: 'quota',
    }) })
    await flushPromises()
    rejectOld(new Error('obsolete missing account'))
    await flushPromises()
    expect(accountTrigger(wrapper).text()).toContain('current-account')
    expect(wrapper.text()).not.toContain('linkedAccountMissing')
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).toHaveBeenCalledWith(43, expect.objectContaining({
      account_id: 99, check_mode: 'quota', primary_model: 'quota', endpoint: '', extra_headers: {},
    }))
  })

  it('does not let a closed dialog request stop the reopened dialog loading', async () => {
    let resolveOld!: (value: unknown) => void
    let resolveCurrent!: (value: unknown) => void
    accountsList.mockReturnValueOnce(new Promise(resolve => { resolveOld = resolve }))
      .mockReturnValueOnce(new Promise(resolve => { resolveCurrent = resolve }))
    const wrapper = mountDialog(makeMonitor({
      provider: 'anthropic', check_mode: 'quota', account_id: null, primary_model: 'quota',
    }))
    await nextTick()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    resolveOld({ items: [{ id: 1, name: 'old-account', platform: 'anthropic', type: 'oauth' }] })
    await flushPromises()
    expect(accountTrigger(wrapper).attributes('disabled')).toBeDefined()
    resolveCurrent({ items: [{ id: 2, name: 'current-account', platform: 'anthropic', type: 'oauth' }] })
    await flushPromises()
    expect(accountTrigger(wrapper).attributes('disabled')).toBeUndefined()
    const dropdown = await openAccountDropdown(wrapper)
    expect(dropdown.textContent).toContain('current-account')
    expect(dropdown.textContent).not.toContain('old-account')
  })

  it('bounds jitter by the minimum interval and preserves the configured value in payload', async () => {
    const wrapper = mountDialog(makeMonitor({ interval_seconds: 90, jitter_seconds: 25 }))
    await flushPromises()
    const numbers = wrapper.findAll('input[type="number"]')
    expect(numbers[1].attributes('max')).toBe('75')
    await numbers[0].setValue(15)
    expect(numbers[1].attributes('max')).toBe('0')
    await numbers[1].setValue(0)
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).toHaveBeenCalledWith(42, expect.objectContaining({ interval_seconds: 15, jitter_seconds: 0 }))
  })

  it('shows only templates compatible with the selected provider and protocol', async () => {
    listTemplates.mockResolvedValue({ items: [
      { id: 1, name: 'chat-template', provider: 'openai', api_mode: 'chat_completions', body_override_mode: 'off' },
      { id: 2, name: 'responses-template', provider: 'openai', api_mode: 'responses', body_override_mode: 'merge', body_override: { max_output_tokens: 40 } },
      { id: 3, name: 'foreign-template', provider: 'anthropic', api_mode: 'chat_completions', body_override_mode: 'off' },
    ] })
    const wrapper = mountDialog(makeMonitor({ api_mode: 'responses' }))
    await flushPromises()
    const picker = wrapper.findComponent({ name: 'Select' })
    expect(picker.props('options')).toEqual([
      expect.objectContaining({ value: '' }),
      expect.objectContaining({ value: '2', label: expect.stringContaining('responses-template') }),
    ])
    picker.vm.$emit('update:modelValue', '1')
    await nextTick()
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).toHaveBeenLastCalledWith(42, expect.objectContaining({ clear_template: true, api_mode: 'responses' }))
    picker.vm.$emit('update:modelValue', '2')
    await nextTick()
    await wrapper.get('#channel-monitor-form').trigger('submit')
    await flushPromises()
    expect(monitorUpdate).toHaveBeenLastCalledWith(42, expect.objectContaining({
      template_id: 2, api_mode: 'responses', body_override_mode: 'merge', body_override: { max_output_tokens: 40 },
    }))
  })
})
