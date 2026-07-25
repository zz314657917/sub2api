import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import MyAccountsView from '../MyAccountsView.vue'

const { userAPI, showError, showSuccess, refreshUser } = vi.hoisted(() => ({
  userAPI: {
    listAccounts: vi.fn(),
    getAccountShareSummary: vi.fn(),
    importAccount: vi.fn(),
    updateAccountShareMode: vi.fn(),
    testAccount: vi.fn(),
    deleteAccount: vi.fn(),
    transferAccountShareToBalance: vi.fn(),
    createAccount: vi.fn(),
    updateAccount: vi.fn(),
    generateAccountAuthURL: vi.fn(),
    exchangeAccountOAuthCode: vi.fn(),
    importAccountSession: vi.fn(),
    listProxies: vi.fn(),
    createProxy: vi.fn(),
    updateProxy: vi.fn(),
    deleteProxy: vi.fn(),
  },
  showError: vi.fn(),
  showSuccess: vi.fn(),
  refreshUser: vi.fn(),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'myAccounts.import.fileSelected') return `Selected ${params?.name}`
        return key
      },
    }),
  }
})

vi.mock('@/api/user', () => ({
  default: userAPI,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    refreshUser,
  }),
}))

const makeAccount = (overrides: Record<string, unknown> = {}) => ({
  id: 1,
  name: 'Private Account',
  notes: null,
  platform: 'openai',
  type: 'oauth',
  credentials: {},
  extra: {},
  share_mode: 'private',
  share_status: 'not_shared',
  schedulable: true,
  current_rpm: 0,
  current_window_cost: 0,
  last_used_at: null,
  expires_at: null,
  ...overrides,
})

function mountView() {
  return mount(MyAccountsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template: '<div><slot name="actions" /><slot name="table" /><slot name="pagination" /></div>',
        },
        DataTable: {
          props: ['columns', 'data', 'loading'],
          emits: ['sort'],
          template: `
            <table>
              <thead>
                <tr>
                  <th v-for="column in columns" :key="column.key">
                    <slot :name="'header-' + column.key" :column="column">{{ column.label }}</slot>
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in data" :key="row.id">
                  <td v-for="column in columns" :key="column.key">
                    <slot :name="'cell-' + column.key" :row="row" :value="row[column.key]">
                      {{ row[column.key] }}
                    </slot>
                  </td>
                </tr>
              </tbody>
            </table>
          `,
        },
        Pagination: { template: '<div />' },
        Icon: { template: '<span />' },
        Select: {
          props: ['modelValue', 'options'],
          emits: ['update:modelValue'],
          template: `
            <select :value="modelValue" @change="$emit('update:modelValue', $event.target.value)">
              <option v-for="option in options" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          `,
        },
        Input: {
          props: ['modelValue', 'label', 'placeholder', 'dataTestid'],
          emits: ['update:modelValue'],
          template: '<input :data-testid="dataTestid" :placeholder="placeholder" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
        },
        PlatformTypeBadge: { template: '<span />' },
        AccountCapacityCell: { template: '<span />' },
        AccountStatusIndicator: { template: '<span />' },
        AccountUsageCell: { template: '<span />' },
      },
    },
  })
}

describe('MyAccountsView import file', () => {
  beforeEach(() => {
    userAPI.listAccounts.mockReset().mockResolvedValue({ items: [], total: 0, pages: 0 })
    userAPI.getAccountShareSummary.mockReset().mockResolvedValue({
      available_amount: 0,
      frozen_amount: 0,
      transferred_amount: 0,
      total_amount: 0,
    })
    userAPI.importAccount.mockReset().mockResolvedValue({
      id: 99,
      name: 'Imported Claude',
      platform: 'anthropic',
      type: 'oauth',
      credentials: {},
    })
    userAPI.updateAccountShareMode.mockReset()
    userAPI.createAccount.mockReset()
    userAPI.updateAccount.mockReset()
    userAPI.exchangeAccountOAuthCode.mockReset().mockResolvedValue(makeAccount({ id: 88 }))
    userAPI.importAccountSession.mockReset().mockResolvedValue(makeAccount({ id: 89, platform: 'anthropic' }))
    userAPI.listProxies.mockReset().mockResolvedValue([])
    userAPI.createProxy.mockReset()
    userAPI.updateProxy.mockReset()
    userAPI.deleteProxy.mockReset()
    userAPI.generateAccountAuthURL.mockReset().mockResolvedValue({
      auth_url: 'https://auth.example.test/oauth/authorize?state=state-1',
      session_id: 'session-1',
      state: 'state-1',
    })
    showError.mockReset()
    showSuccess.mockReset()
    refreshUser.mockReset()
  })

  it('recognizes compact proxy input and saves structured fields', async () => {
    userAPI.createProxy.mockResolvedValue({
      id: 7,
      name: '203.0.113.10:9004',
      protocol: 'socks5h',
      host: '203.0.113.10',
      port: 9004,
      username: 'test-user',
      password: 'test-pass',
      status: 'active',
    })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="my-accounts-open-proxies"]').trigger('click')
    await flushPromises()
    await wrapper
      .get('[data-testid="my-accounts-proxy-smart-input"]')
      .setValue('socks5h://203.0.113.10:9004:test-user:test-pass')
    await wrapper.get('[data-testid="my-accounts-proxy-smart-parse"]').trigger('click')

    const saveButton = wrapper.findAll('button').find(button => button.text().includes('新增代理'))
    expect(saveButton).toBeTruthy()
    await saveButton!.trigger('click')
    await flushPromises()

    expect(userAPI.createProxy).toHaveBeenCalledWith({
      name: '203.0.113.10:9004',
      protocol: 'socks5h',
      host: '203.0.113.10',
      port: 9004,
      username: 'test-user',
      password: 'test-pass',
      status: 'active',
    })
  })

  it('recognizes multiple proxy lines and creates each structured proxy', async () => {
    let nextId = 10
    userAPI.createProxy.mockImplementation(async (payload) => ({
      id: nextId++,
      ...payload,
    }))

    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="my-accounts-open-proxies"]').trigger('click')
    await flushPromises()
    await wrapper
      .get('[data-testid="my-accounts-proxy-smart-input"]')
      .setValue([
        'socks5h://203.0.113.10:9004:test-user:test-pass',
        'http://http-user:http-pass@proxy.example.com:8080',
      ].join('\n'))
    await wrapper.get('[data-testid="my-accounts-proxy-smart-parse"]').trigger('click')

    const batchSaveButton = wrapper.get('[data-testid="my-accounts-proxy-smart-batch-save"]')
    await batchSaveButton.trigger('click')
    await flushPromises()

    expect(userAPI.createProxy).toHaveBeenNthCalledWith(1, {
      name: '203.0.113.10:9004',
      protocol: 'socks5h',
      host: '203.0.113.10',
      port: 9004,
      username: 'test-user',
      password: 'test-pass',
    })
    expect(userAPI.createProxy).toHaveBeenNthCalledWith(2, {
      name: 'proxy.example.com:8080',
      protocol: 'http',
      host: 'proxy.example.com',
      port: 8080,
      username: 'http-user',
      password: 'http-pass',
    })
  })

  it('rejects a multi-line batch when any line is invalid', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="my-accounts-open-proxies"]').trigger('click')
    await flushPromises()
    await wrapper
      .get('[data-testid="my-accounts-proxy-smart-input"]')
      .setValue([
        'socks5h://203.0.113.10:9004:test-user:test-pass',
        'not-a-proxy',
      ].join('\n'))
    await wrapper.get('[data-testid="my-accounts-proxy-smart-parse"]').trigger('click')

    expect(wrapper.find('[data-testid="my-accounts-proxy-smart-batch-save"]').exists()).toBe(false)
    expect(userAPI.createProxy).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('myAccounts.proxy.smartInputBatchInvalid')
  })

  it('reads a selected JSON file, detects platform, and imports with that platform', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="my-accounts-open-import"]').trigger('click')

    const content = JSON.stringify({
      platform: 'anthropic',
      credentials: {
        session_key: 'claude-session-key',
      },
    })
    const file = new File([content], 'claude-session.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: vi.fn().mockResolvedValue(content),
    })

    const input = wrapper.get('[data-testid="my-accounts-import-file-input"]')
    Object.defineProperty(input.element, 'files', {
      configurable: true,
      value: [file],
    })
    await input.trigger('change')
    await flushPromises()

    const textarea = wrapper.get('[data-testid="my-accounts-import-content"]').element as HTMLTextAreaElement
    expect(textarea.value).toBe(content)
    expect(wrapper.text()).toContain('Selected claude-session.json')

    await wrapper.get('[data-testid="my-accounts-import-submit"]').trigger('click')
    await flushPromises()

    expect(userAPI.importAccount).toHaveBeenCalledWith(expect.objectContaining({
      platform: 'anthropic',
      credentials: {
        session_key: 'claude-session-key',
      },
    }))
  })

  it('generates OpenAI OAuth URLs without overriding the Codex redirect URI', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="my-accounts-open-create"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="my-accounts-next"]').trigger('click')
    await flushPromises()

    const generateButton = wrapper
      .findAll('button')
      .find(button => button.text().includes('myAccounts.generateAuthUrl'))
    expect(generateButton).toBeTruthy()

    await generateButton!.trigger('click')
    await flushPromises()

    expect(userAPI.generateAccountAuthURL).toHaveBeenCalledWith({
      platform: 'openai',
      method: 'oauth',
    })
  })

  it('passes advanced account configuration when saving an OpenAI OAuth account', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="my-accounts-open-create"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="my-accounts-name"]').setValue('Configured OpenAI')
    await wrapper.get('[data-testid="my-accounts-next"]').trigger('click')
    await flushPromises()

    const generateButton = wrapper
      .findAll('button')
      .find(button => button.text().includes('myAccounts.generateAuthUrl'))
    await generateButton!.trigger('click')
    await flushPromises()

    const oauthInput = wrapper
      .findAll('input')
      .find(input => input.attributes('placeholder') === 'myAccounts.oauthCodePlaceholder')
    expect(oauthInput).toBeTruthy()
    await oauthInput!.setValue('http://localhost:3000/auth/callback?code=oauth-code&state=state-1')

    await wrapper.get('[data-testid="my-accounts-save"]').trigger('click')
    await flushPromises()

    expect(userAPI.exchangeAccountOAuthCode).toHaveBeenCalledWith(expect.objectContaining({
      platform: 'openai',
      method: 'oauth',
      session_id: 'session-1',
      code: 'oauth-code',
      name: 'Configured OpenAI',
      concurrency: 3,
      auto_pause_on_expired: true,
      credential_extras: expect.objectContaining({
        plan_type: 'plus',
        model_mapping: expect.objectContaining({
          'gpt-5.2': 'gpt-5.2',
        }),
      }),
      extra: expect.objectContaining({
        share_display_tier: 'plus',
        share_display_5h_limit: 100,
        share_display_7d_limit: 100,
      }),
    }))
  })

  it('preserves exported sub2api account name and extra metadata on import', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="my-accounts-open-import"]').trigger('click')

    const content = JSON.stringify({
      name: 'j92wqgddr0@kairo.edu.kg-plus',
      platform: 'openai',
      type: 'oauth',
      credentials: {
        access_token: 'access-token',
        id_token: 'id-token',
        chatgpt_account_id: 'chatgpt-account-id',
      },
      extra: {
        email: 'j92wqgddr0@kairo.edu.kg',
      },
    })
    const file = new File([content], 'sub2api-4accounts-1.user-import.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: vi.fn().mockResolvedValue(content),
    })

    const input = wrapper.get('[data-testid="my-accounts-import-file-input"]')
    Object.defineProperty(input.element, 'files', {
      configurable: true,
      value: [file],
    })
    await input.trigger('change')
    await flushPromises()

    await wrapper.get('[data-testid="my-accounts-import-submit"]').trigger('click')
    await flushPromises()

    expect(userAPI.importAccount).toHaveBeenCalledWith(expect.objectContaining({
      name: 'j92wqgddr0@kairo.edu.kg-plus',
      platform: 'openai',
      type: 'oauth',
      credentials: expect.objectContaining({
        access_token: 'access-token',
        id_token: 'id-token',
        chatgpt_account_id: 'chatgpt-account-id',
      }),
      extra: {
        email: 'j92wqgddr0@kairo.edu.kg',
      },
    }))
  })

  it('reads supported files from a selected folder and imports each with inferred settings', async () => {
    userAPI.importAccount.mockImplementation(async (payload) => ({
      id: userAPI.importAccount.mock.calls.length,
      name: 'Imported',
      platform: payload.platform,
      type: payload.type,
      credentials: payload.credentials,
    }))

    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="my-accounts-open-import"]').trigger('click')

    const claudeContent = JSON.stringify({
      platform: 'anthropic',
      credentials: {
        session_key: 'claude-folder-session',
      },
    })
    const openAIContent = 'openai-refresh-token'
    const unsupportedContent = 'ignore me'

    const claudeFile = new File([claudeContent], 'claude.json', { type: 'application/json' })
    Object.defineProperty(claudeFile, 'text', { value: vi.fn().mockResolvedValue(claudeContent) })
    Object.defineProperty(claudeFile, 'webkitRelativePath', { value: 'accounts/claude.json' })

    const openAIFile = new File([openAIContent], 'openai.token', { type: 'text/plain' })
    Object.defineProperty(openAIFile, 'text', { value: vi.fn().mockResolvedValue(openAIContent) })
    Object.defineProperty(openAIFile, 'webkitRelativePath', { value: 'accounts/openai.token' })

    const ignoredFile = new File([unsupportedContent], 'readme.md', { type: 'text/markdown' })
    Object.defineProperty(ignoredFile, 'text', { value: vi.fn().mockResolvedValue(unsupportedContent) })
    Object.defineProperty(ignoredFile, 'webkitRelativePath', { value: 'accounts/readme.md' })

    const input = wrapper.get('[data-testid="my-accounts-import-folder-input"]')
    Object.defineProperty(input.element, 'files', {
      configurable: true,
      value: [claudeFile, openAIFile, ignoredFile],
    })
    await input.trigger('change')
    await flushPromises()

    expect(wrapper.text()).toContain('myAccounts.import.folderSelected')

    await wrapper.get('[data-testid="my-accounts-import-submit"]').trigger('click')
    await flushPromises()

    expect(userAPI.importAccount).toHaveBeenCalledTimes(2)
    expect(userAPI.importAccount).toHaveBeenCalledWith(expect.objectContaining({
      format: 'sub2api_oauth_json',
      platform: 'anthropic',
      credentials: {
        session_key: 'claude-folder-session',
      },
    }))
    expect(userAPI.importAccount).toHaveBeenCalledWith(expect.objectContaining({
      format: 'openai_refresh_token',
      platform: 'openai',
      credentials: {
        refresh_token: openAIContent,
      },
    }))
  })

  it('bulk applies public sharing only for selected non-public accounts', async () => {
    userAPI.listAccounts.mockResolvedValue({
      items: [
        makeAccount({ id: 1, name: 'Private One', share_mode: 'private', share_status: 'not_shared' }),
        makeAccount({ id: 2, name: 'Already Public', share_mode: 'public', share_status: 'active' }),
        makeAccount({ id: 3, name: 'Private Two', share_mode: 'private', share_status: 'not_shared' }),
      ],
      total: 3,
      pages: 1,
    })
    userAPI.updateAccountShareMode.mockImplementation(async (id, shareMode) => ({
      ...makeAccount({
        id,
        name: `Account ${id}`,
        share_mode: shareMode,
        share_status: 'pending_review',
      }),
    }))

    const wrapper = mountView()
    await flushPromises()

    const checkboxes = wrapper.findAll('input[type="checkbox"]')
    expect(checkboxes.length).toBeGreaterThanOrEqual(4)

    await checkboxes[1].setValue(true)
    await checkboxes[2].setValue(true)
    await checkboxes[3].setValue(true)

    await wrapper.get('[data-testid="my-accounts-bulk-apply-public"]').trigger('click')
    await flushPromises()

    expect(userAPI.updateAccountShareMode).toHaveBeenCalledTimes(2)
    expect(userAPI.updateAccountShareMode).toHaveBeenCalledWith(1, 'public')
    expect(userAPI.updateAccountShareMode).toHaveBeenCalledWith(3, 'public')
    expect(userAPI.updateAccountShareMode).not.toHaveBeenCalledWith(2, 'public')
    expect(showSuccess).toHaveBeenCalledWith('myAccounts.bulk.applyPublicSuccess')
  })

  it('bulk makes selected public accounts private without touching private accounts', async () => {
    userAPI.listAccounts.mockResolvedValue({
      items: [
        makeAccount({ id: 1, name: 'Private One', share_mode: 'private', share_status: 'not_shared' }),
        makeAccount({ id: 2, name: 'Public One', share_mode: 'public', share_status: 'active' }),
        makeAccount({ id: 3, name: 'Public Two', share_mode: 'public', share_status: 'pending_review' }),
      ],
      total: 3,
      pages: 1,
    })
    userAPI.updateAccountShareMode.mockImplementation(async (id, shareMode) => ({
      ...makeAccount({
        id,
        name: `Account ${id}`,
        share_mode: shareMode,
        share_status: 'not_shared',
      }),
    }))

    const wrapper = mountView()
    await flushPromises()

    const checkboxes = wrapper.findAll('input[type="checkbox"]')
    expect(checkboxes.length).toBeGreaterThanOrEqual(4)

    await checkboxes[1].setValue(true)
    await checkboxes[2].setValue(true)
    await checkboxes[3].setValue(true)

    await wrapper.get('[data-testid="my-accounts-bulk-make-private"]').trigger('click')
    await flushPromises()

    expect(userAPI.updateAccountShareMode).toHaveBeenCalledTimes(2)
    expect(userAPI.updateAccountShareMode).toHaveBeenCalledWith(2, 'private')
    expect(userAPI.updateAccountShareMode).toHaveBeenCalledWith(3, 'private')
    expect(userAPI.updateAccountShareMode).not.toHaveBeenCalledWith(1, 'private')
    expect(showSuccess).toHaveBeenCalledWith('myAccounts.bulk.makePrivateSuccess')
  })

  it('does not expose API key upload controls for normal users', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="my-accounts-open-create"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('API Key')
    expect(wrapper.find('[data-testid="my-accounts-apikey-base-url"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="my-accounts-apikey-value"]').exists()).toBe(false)
  })

  it('persists share display account count when editing an OpenAI API key account', async () => {
    userAPI.listAccounts.mockResolvedValue({
      items: [
        makeAccount({
          id: 8,
          name: 'OpenAI API key',
          type: 'apikey',
          credentials: { base_url: 'https://api.openai.com' },
          extra: {},
        }),
      ],
      total: 1,
      pages: 1,
    })
    userAPI.updateAccount.mockResolvedValue(makeAccount({
      id: 8,
      name: 'OpenAI API key',
      type: 'apikey',
      credentials: { base_url: 'https://api.openai.com' },
      extra: {
        share_display_tier: 'pro',
        share_display_percent_only: true,
        share_display_account_count: 3,
        share_display_5h_limit: 500,
        share_display_5h_used: 95.17,
        share_display_7d_limit: 2160,
        share_display_7d_used: 95.17,
      },
    }))

    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find(button => button.classes().includes('hover:text-primary-600'))?.trigger('click')
    await wrapper.get('[data-testid="share-display-target-pool"]').setValue('pro')
    await wrapper.get('[data-testid="share-display-account-count"]').setValue('3')
    await wrapper.get('[data-testid="share-display-5h-limit"]').setValue('500')
    await wrapper.get('[data-testid="share-display-5h-used"]').setValue('95.17')
    await wrapper.get('[data-testid="share-display-7d-limit"]').setValue('2160')
    await wrapper.get('[data-testid="share-display-7d-used"]').setValue('95.17')
    await wrapper.get('[data-testid="my-accounts-save"]').trigger('click')
    await flushPromises()

    expect(userAPI.updateAccount).toHaveBeenCalledWith(8, expect.objectContaining({
      extra: expect.objectContaining({
        share_display_tier: 'pro',
        share_display_percent_only: true,
        share_display_account_count: 3,
        share_display_5h_limit: 500,
        share_display_5h_used: 95.17,
        share_display_7d_limit: 2160,
        share_display_7d_used: 95.17,
      }),
    }))
  })

  it('blocks imported API key and upstream credentials before upload', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="my-accounts-open-import"]').trigger('click')
    await wrapper.get('[data-testid="my-accounts-import-content"]').setValue(JSON.stringify({
      type: 'upstream',
      credentials: {
        base_url: 'https://api.openai.com',
        api_key: 'sk-test',
      },
    }))

    await wrapper.get('[data-testid="my-accounts-import-submit"]').trigger('click')
    await flushPromises()

    expect(userAPI.importAccount).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('myAccounts.apiKeyUploadDisabled')
  })
})
