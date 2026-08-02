import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import OpsErrorLogTable from '../OpsErrorLogTable.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

const DataTableStub = {
  props: ['columns', 'data', 'serverSideSort'],
  emits: ['sort'],
  template: `
    <div>
      <span data-test="columns">{{ columns.map((column) => column.key).join(',') }}</span>
      <span data-test="sortable-columns">{{ columns.filter((column) => column.sortable).map((column) => column.key).join(',') }}</span>
      <span data-test="server-side-sort">{{ String(serverSideSort) }}</span>
      <button data-test="sort-status" @click="$emit('sort', 'status', 'asc')">sort</button>
      <slot name="cell-user" :row="data[0]" />
      <slot name="cell-actions" :row="data[0]" />
    </div>
  `,
}

const row = {
  id: 77,
  created_at: '2026-07-01T00:00:00Z',
  phase: 'request',
  type: 'invalid_request_error',
  error_owner: 'client',
  error_source: 'client_request',
  severity: 'warning',
  status_code: 400,
  platform: 'openai',
  model: 'gpt-5.6',
  is_retryable: false,
  retry_count: 0,
  resolved: false,
  client_request_id: 'client-1',
  request_id: 'request-1',
  message: 'invalid request',
  user_id: 9,
  user_email: 'user@test.com',
  account_name: '',
  group_name: '',
}

describe('OpsErrorLogTable', () => {
  it('honors visible columns and maps sortable status to the API key', async () => {
    const wrapper = mount(OpsErrorLogTable, {
      props: {
        rows: [row],
        total: 1,
        loading: false,
        page: 1,
        pageSize: 20,
        visibleColumnKeys: ['user', 'status', 'created_at', 'actions'],
        serverSideSort: true,
      },
      global: { stubs: { DataTable: DataTableStub, EmptyState: true, Pagination: true } },
    })

    expect(wrapper.get('[data-test="columns"]').text()).toBe('user,status,created_at,actions')
    expect(wrapper.get('[data-test="sortable-columns"]').text()).toBe('status,created_at')
    expect(wrapper.get('[data-test="server-side-sort"]').text()).toBe('true')
    await wrapper.get('[data-test="sort-status"]').trigger('click')
    expect(wrapper.emitted('sort')?.[0]).toEqual(['status_code', 'asc'])
  })

  it('keeps the existing Ops modal table non-sortable by default', () => {
    const wrapper = mount(OpsErrorLogTable, {
      props: { rows: [row], total: 1, loading: false, page: 1, pageSize: 20 },
      global: { stubs: { DataTable: DataTableStub, EmptyState: true, Pagination: true } },
    })

    expect(wrapper.get('[data-test="sortable-columns"]').text()).toBe('')
    expect(wrapper.get('[data-test="server-side-sort"]').text()).toBe('false')
  })

  it('emits user drill-down and detail actions from the row', async () => {
    const wrapper = mount(OpsErrorLogTable, {
      props: { rows: [row], total: 1, loading: false, page: 1, pageSize: 20, userClickable: true },
      global: { stubs: { DataTable: DataTableStub, EmptyState: true, Pagination: true } },
    })

    const buttons = wrapper.findAll('button')
    await buttons.find((button) => button.text().includes('user@test.com'))!.trigger('click')
    await buttons.find((button) => button.text().includes('details'))!.trigger('click')

    expect(wrapper.emitted('userClick')?.[0]).toEqual([9, 'user@test.com'])
    expect(wrapper.emitted('openErrorDetail')?.[0]).toEqual([77])
  })
})
