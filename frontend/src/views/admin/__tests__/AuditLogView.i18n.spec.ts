import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AuditLogView from '../AuditLogView.vue'

const { list, get, clear, getStatus, showError, showSuccess } = vi.hoisted(() => ({
  list: vi.fn(),
  get: vi.fn(),
  clear: vi.fn(),
  getStatus: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

const zhMessages = {
  'admin.audit.filters.all': '全部',
  'admin.audit.roles.admin': '管理员',
  'admin.audit.roles.unknown': '未知角色',
  'admin.audit.authMethods.jwt': 'JWT',
  'admin.audit.authMethods.adminApiKey': '管理员 API Key',
  'admin.audit.actionParts.admin': '管理',
  'admin.audit.actionParts.delete': '删除'
} as Record<string, string>

const zhExactActionMessages = {
  'admin.accounts.export': '导出账号'
} as Record<string, string>

const enMessages = {
  'admin.audit.filters.all': 'All',
  'admin.audit.roles.admin': 'Administrator',
  'admin.audit.roles.unknown': 'Unknown role',
  'admin.audit.authMethods.jwt': 'JWT',
  'admin.audit.authMethods.adminApiKey': 'Admin API Key',
  'admin.audit.actionParts.admin': 'Admin',
  'admin.audit.actionParts.delete': 'Delete'
} as Record<string, string>

const enExactActionMessages = {
  'admin.accounts.export': 'Export accounts'
} as Record<string, string>

let messages = zhMessages
let exactActionMessages = zhExactActionMessages

vi.mock('@/api/admin', () => ({
  adminAPI: {
    audit: { list, get, clear }
  }
}))

vi.mock('@/api', () => ({
  totpAPI: { getStatus }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError, showSuccess })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
      te: (key: string) => key in messages,
      tm: (key: string) => (key === 'admin.audit.actions' ? exactActionMessages : {})
    })
  }
})

const DataTableStub = {
  props: ['data'],
  template: `
    <div data-test="audit-table">
      <div v-for="row in data" :key="row.id" data-test="audit-row">
        <slot name="cell-actor" :row="row" />
        <slot name="cell-action" :row="row" />
        <slot name="cell-actions" :row="row" />
      </div>
    </div>
  `
}

function mountView() {
  return mount(AuditLogView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
        DataTable: DataTableStub,
        Pagination: true,
        Select: true,
        BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' },
        ConfirmDialog: true,
        Icon: true
      }
    }
  })
}

describe('AuditLogView i18n display', () => {
  beforeEach(() => {
    messages = zhMessages
    exactActionMessages = zhExactActionMessages
    list.mockReset().mockResolvedValue({
      items: [
        {
          id: 1,
          created_at: '2026-08-02T00:00:00Z',
          actor_email: 'admin@example.com',
          actor_role: 'admin',
          auth_method: 'admin_api_key',
          action: 'admin.accounts.export',
          method: 'GET',
          path: '/api/v1/admin/accounts/data',
          client_ip: '127.0.0.1',
          status_code: 200,
          latency_ms: 5
        },
        {
          id: 2,
          created_at: '2026-08-02T00:00:00Z',
          actor_email: 'admin@example.com',
          actor_role: 'admin',
          auth_method: 'jwt',
          action: 'admin.unmapped.delete',
          method: 'DELETE',
          path: '/api/v1/admin/unmapped/1',
          client_ip: '127.0.0.1',
          status_code: 200,
          latency_ms: 5
        },
        {
          id: 3,
          created_at: '2026-08-02T00:00:00Z',
          actor_email: 'admin@example.com',
          actor_role: 'admin',
          auth_method: 'jwt',
          action: 'auth_login',
          method: 'POST',
          path: '/api/v1/auth/login',
          client_ip: '127.0.0.1',
          status_code: 200,
          latency_ms: 5
        },
        {
          id: 4,
          created_at: '2026-08-02T00:00:00Z',
          actor_email: 'admin@example.com',
          actor_role: 'custom_role',
          auth_method: 'custom_auth',
          action: 'admin.unmapped.delete',
          method: 'DELETE',
          path: '/api/v1/admin/unmapped/2',
          client_ip: '127.0.0.1',
          status_code: 200,
          latency_ms: 5
        }
      ],
      total: 4
    })
    get.mockReset().mockResolvedValue({
      id: 1,
      created_at: '2026-08-02T00:00:00Z',
      actor_email: 'admin@example.com',
      actor_role: 'admin',
      auth_method: 'admin_api_key',
      action: 'admin.accounts.export',
      method: 'GET',
      path: '/api/v1/admin/accounts/data',
      client_ip: '127.0.0.1',
      status_code: 200,
      latency_ms: 5
    })
    clear.mockReset()
    getStatus.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('localizes known audit values while preserving their raw values in titles', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('管理员 · 管理员 API Key')
    expect(wrapper.text()).toContain('导出账号')
    expect(wrapper.get('[title="admin.accounts.export"]').text()).toBe('导出账号')
    expect(wrapper.get('[title="admin · admin_api_key"]').text()).toContain('管理员 · 管理员 API Key')
  })

  it('translates known parts of new actions and keeps unmapped parts visible', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[title="admin.unmapped.delete"]').text()).toBe('管理 · unmapped · 删除')
    expect(wrapper.get('[title="auth_login"]').text()).toBe('auth_login')
    expect(wrapper.get('[title="custom_role · custom_auth"]').text()).toContain('custom_role · custom_auth')
  })

  it('keeps the raw authentication method in the detail view title', async () => {
    const wrapper = mountView()
    await flushPromises()

    const detailButton = wrapper.findAll('button').find((button) => button.text().includes('admin.audit.columns.detail'))
    await detailButton?.trigger('click')
    await flushPromises()

    expect(get).toHaveBeenCalledWith(1)
    expect(wrapper.get('[title="admin_api_key"]').text()).toBe('管理员 API Key')
    expect(wrapper.findAll('[title="admin.accounts.export"]').some((item) => item.text() === '导出账号')).toBe(true)
  })

  it('sends the raw action filter value to the API', async () => {
    const wrapper = mountView()
    await flushPromises()

    const actionFilter = wrapper.findAll('input')[2]
    await actionFilter.setValue('admin.accounts.export')
    await actionFilter.trigger('keyup.enter')
    await flushPromises()

    expect(list).toHaveBeenLastCalledWith(expect.objectContaining({ action: 'admin.accounts.export' }))
  })

  it('renders the same known values in English', async () => {
    messages = enMessages
    exactActionMessages = enExactActionMessages
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('Administrator · Admin API Key')
    expect(wrapper.get('[title="admin.accounts.export"]').text()).toBe('Export accounts')
  })
})
