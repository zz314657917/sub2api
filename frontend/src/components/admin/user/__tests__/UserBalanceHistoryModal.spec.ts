import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserBalanceHistoryModal from '../UserBalanceHistoryModal.vue'

const { getUserBalanceHistory, listUsage, getOrders } = vi.hoisted(() => ({
  getUserBalanceHistory: vi.fn(),
  listUsage: vi.fn(),
  getOrders: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: { getUserBalanceHistory },
    usage: { list: listUsage },
    payment: { getOrders },
  },
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const user = {
  id: 12,
  email: 'user@example.com',
  username: 'user',
  balance: 0,
  notes: '',
  created_at: '2026-06-01T11:00:00Z',
} as any

function mountModal() {
  return mount(UserBalanceHistoryModal, {
    props: { show: false, user },
    global: {
      stubs: {
        BaseDialog: { template: '<div><slot /></div>' },
        Select: true,
        Icon: true,
        OrderStatusBadge: true,
      },
    },
  })
}

describe('UserBalanceHistoryModal', () => {
  beforeEach(() => {
    getUserBalanceHistory.mockReset().mockResolvedValue({ items: [], total: 0, total_recharged: 0 })
    listUsage.mockReset().mockResolvedValue({
      items: [{
        id: 1,
        model: 'gpt-5',
        request_type: 'sync',
        created_at: '2026-06-01T12:00:00Z',
        input_tokens: 10,
        output_tokens: 20,
        cache_creation_tokens: 0,
        cache_read_tokens: 0,
        duration_ms: 120,
        total_cost: 0.01,
        actual_cost: 0.01,
      }],
    })
    getOrders.mockReset().mockResolvedValue({ data: { items: [] } })
  })

  it('requests the latest 30 usage records in a fixed-height scroll container', async () => {
    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(listUsage).toHaveBeenCalledWith({
      user_id: 12,
      page: 1,
      page_size: 30,
      sort_by: 'created_at',
      sort_order: 'desc',
    })

    const usageList = wrapper.get('[data-test="recent-usage-list"]')
    expect(usageList.classes()).toContain('max-h-[28rem]')
    expect(usageList.classes()).toContain('overflow-y-auto')
  })
})
