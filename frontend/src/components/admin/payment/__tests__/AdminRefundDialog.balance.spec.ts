import { afterEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import AdminRefundDialog from '../AdminRefundDialog.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
enableAutoUnmount(afterEach)

describe('refund balance warning', () => {
  it('uses the requested refund amount instead of the complete order amount', async () => {
    const wrapper = mount(AdminRefundDialog, {
      props: { show: false, userBalance: 50, order: { id: 1, amount: 100, pay_amount: 100, fee_rate: 0, payment_type: 'stripe', out_trade_no: 'o', status: 'COMPLETED', order_type: 'balance', created_at: '', expires_at: '', refund_amount: 0 } },
      global: { stubs: { BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' } } }
    })
    await wrapper.setProps({ show: true })
    expect(wrapper.text()).toContain('payment.admin.insufficientBalance')
    const amount = wrapper.get('input[type="number"]')
    await amount.setValue('20')
    expect(wrapper.text()).not.toContain('payment.admin.insufficientBalance')
    await amount.setValue('50')
    expect(wrapper.text()).not.toContain('payment.admin.insufficientBalance')
    await amount.setValue('50.01')
    expect(wrapper.text()).toContain('payment.admin.insufficientBalance')
  })
})
