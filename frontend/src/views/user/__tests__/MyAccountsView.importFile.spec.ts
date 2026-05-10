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

function mountView() {
  return mount(MyAccountsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template: '<div><slot name="actions" /><slot name="table" /><slot name="pagination" /></div>',
        },
        DataTable: { template: '<div />' },
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
          props: ['modelValue', 'label', 'placeholder'],
          emits: ['update:modelValue'],
          template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
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
    showError.mockReset()
    showSuccess.mockReset()
    refreshUser.mockReset()
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
})
