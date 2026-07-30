import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AdminGroupBuyView from '../AdminGroupBuyView.vue'

const {
  listPlans,
  listRounds,
  listRoundSeats,
  processRefunds,
  updatePlan,
  getAll,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  listPlans: vi.fn(),
  listRounds: vi.fn(),
  listRoundSeats: vi.fn(),
  processRefunds: vi.fn(),
  updatePlan: vi.fn(),
  getAll: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groupBuy: { listPlans, listRounds, listRoundSeats, processRefunds, updatePlan },
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
    listPlans.mockReset().mockResolvedValue({ data: [{
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
      agreement_text: '',
      status: 'active',
      sort_order: 0,
      created_at: '',
      updated_at: '',
    }] })
    listRounds.mockReset().mockResolvedValue({ data: {
      items: [{
        id: 31,
        plan_id: 11,
        status: 'cancelled',
        total_shares: 10,
        paid_shares: 2,
        reserved_shares: 0,
        available_shares: 8,
        total_seats: 10,
        paid_seats: 2,
        reserved_seats: 0,
        available_seats: 8,
        deadline_at: '2026-07-23T00:00:00Z',
        created_at: '',
        updated_at: '',
        refund_summary: { total: 2, pending: 1, processing: 0, pending_provider: 0, succeeded: 1, failed: 0, needs_review: 0, amount: 24 },
      }],
      total: 1,
      pages: 1,
    } })
    listRoundSeats.mockReset().mockResolvedValue({ data: [{
      id: 41,
      round_id: 31,
      plan_id: 11,
      user_id: 9,
      status: 'refunded',
      share_count: 2,
      user: { id: 9, email: 'buyer@example.com', username: 'buyer' },
      order: { id: 51, amount: 24, pay_amount: 24, currency: 'CNY', payment_type: 'stripe', out_trade_no: 'GB-51', status: 'REFUNDED', order_type: 'group_buy', created_at: '', expires_at: '' },
      refund: { id: 61, seat_id: 41, user_id: 9, mode: 'balance_credit', status: 'succeeded', amount: 24, created_at: '', updated_at: '' },
      created_at: '',
      updated_at: '',
    }] })
    processRefunds.mockReset().mockResolvedValue({ data: { processed: 1, succeeded: 1, pending: 0, failed: 0, failures: [] } })
    updatePlan.mockReset().mockResolvedValue({ data: {} })
    getAll.mockReset().mockResolvedValue([{ id: 7, name: '订阅组', status: 'active', subscription_type: 'subscription', platform: 'openai' }])
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('allows cancelled-round refunds and exposes participant/order/refund details', async () => {
    const wrapper = mount(AdminGroupBuyView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: { template: '<span />' },
        },
      },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('1 已退 / 1 待处理 / 0 异常')
    const refundButton = wrapper.findAll('button').find((button) => button.text().includes('处理退款'))
    expect(refundButton?.attributes('disabled')).toBeUndefined()

    const detailButton = wrapper.findAll('button').find((button) => button.text().includes('详情'))
    await detailButton?.trigger('click')
    await flushPromises()
    expect(listRoundSeats).toHaveBeenCalledWith(31)
    expect(document.body.textContent).toContain('buyer@example.com')
    expect(document.body.textContent).toContain('已退款')

    await refundButton?.trigger('click')
    await flushPromises()
    expect(processRefunds).toHaveBeenCalledWith(31)
    expect(showSuccess).toHaveBeenCalledWith('退款处理完成：成功 1，待确认 0，失败 0')

    wrapper.unmount()
  })

  it('keeps the plan editor open and submits an editable agreement', async () => {
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

    const textareas = document.body.querySelectorAll<HTMLTextAreaElement>('textarea')
    const agreement = textareas.item(textareas.length - 1)
    agreement.value = '自定义退款协议'
    agreement.dispatchEvent(new Event('input', { bubbles: true }))

    const saveButton = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((button) =>
      button.textContent?.includes('保存拼团'),
    )
    expect(saveButton).toBeDefined()
    saveButton!.click()
    await flushPromises()

    expect(updatePlan).toHaveBeenCalledWith(
      11,
      expect.objectContaining({
        agreement_text: '自定义退款协议',
      }),
    )
    wrapper.unmount()
  })
})
