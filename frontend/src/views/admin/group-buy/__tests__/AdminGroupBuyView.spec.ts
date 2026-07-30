import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminGroupBuyView from '../AdminGroupBuyView.vue'

const { getAll, listPlans, listRounds, showError, showSuccess } = vi.hoisted(() => ({
  getAll: vi.fn(),
  listPlans: vi.fn(),
  listRounds: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groupBuy: { listPlans, listRounds },
    groups: { getAll },
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    cachedPublicSettings: {},
    showError,
    showSuccess,
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (_error: unknown, fallback: string) => fallback,
}))

describe('AdminGroupBuyView', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
    listPlans.mockReset().mockResolvedValue({
      data: [{
        id: 11,
        title: '10 份团',
        description: '',
        total_shares: 10,
        seat_count: 10,
        price_per_share: 12,
        price_per_seat: 12,
        price_label: '',
        quota_per_share_label: '50 USD/月',
        quota_label: '50 USD/月',
        max_shares_per_user: 10,
        target_group_id: 7,
        tier_group_ids: { '1': 7 },
        tier_groups: [],
        tier_rules: [{ min_shares: 1, max_shares: 10, target_group_id: 7, label: '默认' }],
        validity_days: 30,
        timeout_minutes: 1440,
        launch_mode: 'auto',
        refund_mode: 'balance_credit',
        agreement_text: '测试协议文案',
        status: 'active',
        sort_order: 0,
        created_at: '',
        updated_at: '',
      }],
    })
    listRounds.mockReset().mockResolvedValue({
      data: { items: [], total: 0, pages: 0 },
    })
    getAll.mockReset().mockResolvedValue([{
      id: 7,
      name: '订阅组',
      status: 'active',
      subscription_type: 'subscription',
      platform: 'openai',
    }])
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('keeps the plan editor open when clicking the backdrop', async () => {
    const wrapper = mount(AdminGroupBuyView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: { template: '<span />' },
        },
      },
    })
    await flushPromises()

    const editButton = wrapper.findAll('button').find((button) => button.text().includes('编辑'))
    expect(editButton).toBeDefined()
    await editButton!.trigger('click')

    const backdrop = document.body.querySelector<HTMLElement>('.admin-group-buy-modal-backdrop')
    expect(backdrop).not.toBeNull()
    backdrop!.click()
    await flushPromises()

    expect(document.body.textContent).toContain('编辑拼团')

    document.body.querySelector<HTMLElement>('.admin-group-buy-modal-close')!.click()
    await flushPromises()
    expect(document.body.querySelector('.admin-group-buy-modal-backdrop')).toBeNull()

    wrapper.unmount()
  })
})
