import { describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'

import UserErrorRequestsTable from '../UserErrorRequestsTable.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      te: () => true,
    }),
  }
})

const DataTableStub = defineComponent({
  name: 'DataTable',
  props: {
    columns: { type: Array, default: () => [] },
    data: { type: Array, default: () => [] },
  },
  emits: ['sort'],
  template: `
    <div>
      <button data-test="sort-status" @click="$emit('sort', 'status', 'asc')" />
      <button data-test="sort-model" @click="$emit('sort', 'model', 'desc')" />
      <button data-test="sort-fallback" @click="$emit('sort', 'unknown', 'asc')" />
      <div v-for="row in data" :key="row.id">
        <slot name="cell-actions" :row="row" />
      </div>
    </div>
  `,
})

const UserErrorDetailModalStub = defineComponent({
  name: 'UserErrorDetailModal',
  props: {
    show: { type: Boolean, default: false },
    errorId: { type: Number, default: null },
  },
  emits: ['update:show'],
  template: '<div data-test="detail-modal" />',
})

const row = {
  id: 41,
  created_at: '2026-08-01T00:00:00Z',
  model: 'gpt-5.6-sol',
  inbound_endpoint: '/v1/responses',
  status_code: 429,
  category: 'rate_limit',
  platform: 'openai',
  message: 'rate limited',
  key_name: 'owned-key',
  key_deleted: false,
}

describe('UserErrorRequestsTable', () => {
  const mountTable = () => mount(UserErrorRequestsTable, {
    props: {
      rows: [row],
      loading: false,
      visibleColumnKeys: ['model', 'status', 'created_at'],
    },
    global: {
      stubs: {
        DataTable: DataTableStub,
        EmptyState: true,
        Icon: true,
        UserErrorDetailModal: UserErrorDetailModalStub,
      },
    },
  })

  it('maps local column keys to the server sort contract', async () => {
    const wrapper = mountTable()

    await wrapper.get('[data-test="sort-status"]').trigger('click')
    await wrapper.get('[data-test="sort-model"]').trigger('click')
    await wrapper.get('[data-test="sort-fallback"]').trigger('click')

    expect(wrapper.emitted('sort')).toEqual([
      ['status_code', 'asc'],
      ['model', 'desc'],
      ['created_at', 'asc'],
    ])
  })

  it('keeps the detail action visible and opens the selected owned row', async () => {
    const wrapper = mountTable()
    const dataTable = wrapper.getComponent(DataTableStub)

    expect((dataTable.props('columns') as Array<{ key: string }>).map((column) => column.key)).toEqual([
      'model',
      'status',
      'created_at',
      'actions',
    ])

    await wrapper.get('[data-test="user-error-detail-button"]').trigger('click')

    const modal = wrapper.getComponent(UserErrorDetailModalStub)
    expect(modal.props('errorId')).toBe(41)
    expect(modal.props('show')).toBe(true)
  })
})
