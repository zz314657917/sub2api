import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import type { ApiKey, OpenAIFastPolicy } from '@/types'
import zhKeys from '@/i18n/locales/zh/keys'
import enKeys from '@/i18n/locales/en/keys'
import KeysView from '../KeysView.vue'

const api = vi.hoisted(() => ({
  list: vi.fn(),
  create: vi.fn(),
  update: vi.fn(),
  toggleStatus: vi.fn(),
  getPublicSettings: vi.fn(),
  getAvailable: vi.fn(),
  getUserGroupRates: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

// Follow the existing KeysView tests' API/store/router setup and layout stubs.
// The form and its Select components remain mounted and interactive.
vi.mock('@/api', () => ({
  keysAPI: api,
  authAPI: { getPublicSettings: api.getPublicSettings },
  usageAPI: { getDashboardApiKeysUsage: vi.fn().mockResolvedValue({ stats: {} }) },
  userGroupsAPI: { getAvailable: api.getAvailable, getUserGroupRates: api.getUserGroupRates }
}))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: api.showError, showSuccess: api.showSuccess })
}))
vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({ isCurrentStep: () => false, nextStep: vi.fn() })
}))
vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn() })
}))
vi.mock('@/composables/usePersistedPageSize', () => ({ getPersistedPageSize: () => 10 }))
vi.mock('vue-router', () => ({
  useRoute: () => ({ path: '/keys', query: {} }),
  useRouter: () => ({ replace: vi.fn() })
}))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: translate })
}))

let locale: 'zh' | 'en' = 'zh'
const translate = (key: string): string => {
  const parts = key.split('.')
  if (parts.shift() !== 'keys') return key
  let value: unknown = locale === 'zh' ? zhKeys : enKeys
  for (const part of parts) value = (value as Record<string, unknown>)?.[part]
  return typeof value === 'string' ? value : key
}

const fixture = (policy?: OpenAIFastPolicy): ApiKey => ({
  id: 17,
  user_id: 1,
  key: 'sk-test-key',
  name: 'test-key',
  is_default: false,
  group_id: 2,
  multi_group_routes: [],
  account_pool_strategy: 'shared_only',
  ...(policy ? { openai_fast_policy: policy } : {}),
  status: 'active',
  ip_whitelist: [],
  ip_blacklist: [],
  last_used_at: null,
  quota: 0,
  quota_used: 0,
  expires_at: null,
  created_at: '2026-11-05T00:00:00Z',
  updated_at: '2026-11-05T00:00:00Z',
  current_concurrency: 0,
  rate_limit_5h: 0,
  rate_limit_1d: 0,
  rate_limit_7d: 0,
  token_multiplier_cap: 0,
  usage_5h: 0,
  usage_1d: 0,
  usage_7d: 0,
  window_5h_start: null,
  window_1d_start: null,
  window_7d_start: null,
  reset_5h_at: null,
  reset_1d_at: null,
  reset_7d_at: null
})

const policies: OpenAIFastPolicy[] = ['passthrough', 'force', 'filter']
const wrappers: VueWrapper[] = []
const mountView = async (key = fixture()) => {
  api.list.mockResolvedValue({ items: [key], total: 1, page: 1, page_size: 10, pages: 1 })
  const wrapper = mount(KeysView, {
    attachTo: document.body,
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template: '<div><slot name="filters" /><slot name="actions" /><slot name="table" /></div>'
        },
        DataTable: {
          props: ['data'],
          template: '<div><div v-for="row in data" :key="row.id" data-test="key-row"><slot name="cell-actions" :row="row" /></div></div>'
        },
        BaseDialog: {
          props: ['show'],
          emits: ['close'],
          template: '<div v-if="show" role="dialog"><button data-test="close-dialog" @click="$emit(\'close\')">Close</button><slot /><slot name="footer" /></div>'
        },
        Pagination: true,
        ConfirmDialog: true,
        EmptyState: true,
        Icon: true,
        PlatformIcon: true,
        UseKeyModal: true,
        EndpointPopover: true,
        GroupBadge: true,
        GroupOptionItem: true,
        SearchInput: true,
        VueDraggable: { template: '<div><slot /></div>' }
      }
    }
  })
  wrappers.push(wrapper)
  await flushPromises()
  return wrapper
}

const field = (wrapper: VueWrapper) => wrapper.get('[data-testid="key-openai-fast-policy"]')
const policyLabel = (policy: OpenAIFastPolicy) => translate(`keys.openaiFastPolicy.${policy}`)
const openCreate = async (wrapper: VueWrapper) => {
  await wrapper.get('[data-tour="keys-create-btn"]').trigger('click')
  await flushPromises()
}
const edit = async (wrapper: VueWrapper) => {
  const button = wrapper.get('[data-test="key-row"]').findAll('button')
    .find((button) => button.text() === 'common.edit')
  expect(button).toBeDefined()
  await button!.trigger('click')
  await flushPromises()
}
const choosePolicy = async (wrapper: VueWrapper, policy: OpenAIFastPolicy) => {
  await field(wrapper).get('button').trigger('click')
  await flushPromises()
  const options = Array.from(document.querySelectorAll<HTMLElement>('[role="option"]'))
  expect(options.map((option) => option.textContent?.trim())).toEqual(policies.map(policyLabel))
  const target = options.find((option) => option.textContent?.trim() === policyLabel(policy))
  expect(target).toBeDefined()
  target!.click()
  await flushPromises()
  expect(field(wrapper).get('button').text()).toBe(policyLabel(policy))
}
const completeCreate = async (wrapper: VueWrapper) => {
  await wrapper.get('[data-tour="key-form-name"]').setValue('new-key')
  await wrapper.get('[data-tour="key-form-group"] button').trigger('click')
  await flushPromises()
  const group = document.querySelector<HTMLElement>('[role="option"]')
  expect(group).not.toBeNull()
  group!.click()
  await flushPromises()
}
const submit = async (wrapper: VueWrapper) => {
  await wrapper.get('#key-form').trigger('submit')
  await flushPromises()
}

describe('KeysView OpenAI Fast policy form', () => {
  beforeEach(() => {
    locale = 'zh'
    localStorage.clear()
    vi.clearAllMocks()
    api.create.mockReset().mockResolvedValue(fixture())
    api.update.mockReset().mockResolvedValue(fixture())
    api.getPublicSettings.mockResolvedValue({ account_share_enabled: false })
    api.getAvailable.mockResolvedValue([{ id: 2, name: 'OpenAI group', platform: 'openai' }])
    api.getUserGroupRates.mockResolvedValue({})
  })

  afterEach(() => {
    wrappers.splice(0).forEach((wrapper) => wrapper.unmount())
    document.body.innerHTML = ''
  })

  it.each(['zh', 'en'] as const)('renders a labelled three-option field and policy guidance in %s', async (language) => {
    locale = language
    const wrapper = await mountView()
    await openCreate(wrapper)
    expect(field(wrapper).element.tagName).toBe('FIELDSET')
    expect(field(wrapper).get('legend').text()).toBe(translate('keys.openaiFastPolicyLabel'))
    expect(field(wrapper).get('[aria-labelledby]').attributes('aria-labelledby')).toBe('key-openai-fast-policy-label')
    const hint = field(wrapper).get('#key-openai-fast-policy-hint').text()
    expect(hint).toBe(translate('keys.openaiFastPolicyHint'))
    for (const term of ['OpenAI', 'Fast', 'priority', 'WebSocket', 'WS']) expect(hint).toContain(term)
    for (const term of language === 'zh' ? ['分组', '全局', '价格', '重新连接'] : ['group', 'global', 'prices', 'Reconnect']) {
      expect(hint).toContain(term)
    }
    expect(field(wrapper).get('button').text()).toBe(policyLabel('passthrough'))
    await choosePolicy(wrapper, 'filter')
  })

  it.each(policies)('creates with %s and resets the form after success', async (policy) => {
    const wrapper = await mountView()
    await openCreate(wrapper)
    await completeCreate(wrapper)
    await choosePolicy(wrapper, policy)
    await submit(wrapper)

    expect(api.create).toHaveBeenCalledTimes(1)
    const args = api.create.mock.calls[0]
    expect(args.slice(0, 2)).toEqual(['new-key', 2])
    if (policy === 'passthrough') {
      expect(args).toHaveLength(11)
    } else {
      expect(args).toHaveLength(12)
      expect(args[11]).toBe(policy)
    }
    expect(wrapper.find('#key-form').exists()).toBe(false)
    await openCreate(wrapper)
    expect(field(wrapper).get('button').text()).toBe(policyLabel('passthrough'))
  })

  it.each([...policies, undefined])('loads a saved %s policy, including old responses without the field', async (policy) => {
    const wrapper = await mountView(fixture(policy))
    await edit(wrapper)
    expect(field(wrapper).get('button').text()).toBe(policyLabel(policy ?? 'passthrough'))
    expect(api.update).not.toHaveBeenCalled()
  })

  it.each([
    ['passthrough', 'force'],
    ['force', 'filter'],
    ['filter', 'passthrough']
  ] as const)('saves a change from %s to %s and displays the persisted result', async (previous, policy) => {
    const wrapper = await mountView(fixture(previous))
    await edit(wrapper)
    await choosePolicy(wrapper, policy)
    api.list.mockResolvedValue({ items: [fixture(policy)], total: 1, pages: 1 })
    api.update.mockResolvedValueOnce(fixture(policy))
    await submit(wrapper)
    expect(api.update).toHaveBeenCalledWith(17, expect.objectContaining({ openai_fast_policy: policy }))
    expect(api.create).not.toHaveBeenCalled()
    await edit(wrapper)
    expect(field(wrapper).get('button').text()).toBe(policyLabel(policy))
  })

  it.each([...policies, undefined])('omits unchanged %s policy from an ordinary edit', async (policy) => {
    const wrapper = await mountView(fixture(policy))
    await edit(wrapper)
    await wrapper.get('[data-tour="key-form-name"]').setValue('renamed')
    await submit(wrapper)
    expect(api.update).toHaveBeenCalledTimes(1)
    expect(api.update.mock.calls[0][1]).not.toHaveProperty('openai_fast_policy')
  })

  it.each(['force', 'filter'] as const)('resets a cancelled %s selection before the next create', async (policy) => {
    const wrapper = await mountView(fixture(policy))
    await edit(wrapper)
    await wrapper.get('[data-test="close-dialog"]').trigger('click')
    await openCreate(wrapper)
    expect(field(wrapper).get('button').text()).toBe(policyLabel('passthrough'))
    await choosePolicy(wrapper, policy)
    await wrapper.get('[data-test="close-dialog"]').trigger('click')
    await openCreate(wrapper)
    expect(field(wrapper).get('button').text()).toBe(policyLabel('passthrough'))
    expect(api.create).not.toHaveBeenCalled()
    expect(api.update).not.toHaveBeenCalled()
  })

  it('keeps the selection after a create error so the user can retry', async () => {
    const wrapper = await mountView()
    await openCreate(wrapper)
    await completeCreate(wrapper)
    await choosePolicy(wrapper, 'force')
    api.create.mockRejectedValueOnce(new Error('save failed'))
    await submit(wrapper)
    expect(field(wrapper).get('button').text()).toBe(policyLabel('force'))
    expect(api.showError).toHaveBeenCalled()
    await submit(wrapper)
    expect(api.create.mock.calls[1][11]).toBe('force')
  })

  it('preserves a managed-key denial without falling back to another write', async () => {
    const wrapper = await mountView()
    await edit(wrapper)
    await choosePolicy(wrapper, 'force')
    api.update.mockRejectedValueOnce({
      response: { data: { detail: 'CAFE_MANAGED_KEY_PROTECTED' } }
    })
    await submit(wrapper)
    expect(api.showError).toHaveBeenCalledWith('CAFE_MANAGED_KEY_PROTECTED')
    expect(field(wrapper).get('button').text()).toBe(policyLabel('force'))
    expect(api.update).toHaveBeenCalledTimes(1)
    expect(api.create).not.toHaveBeenCalled()
    expect(api.toggleStatus).not.toHaveBeenCalled()
  })
})
