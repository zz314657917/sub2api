import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserOrdersView from '../UserOrdersView.vue'

const { paymentAPI, showError, showSuccess, push } = vi.hoisted(() => ({
  paymentAPI: {
    getMyOrders: vi.fn(),
    getRefundEligibleProviders: vi.fn(),
    getInvoiceSummary: vi.fn(),
    getMyInvoices: vi.fn(),
    createInvoice: vi.fn(),
    downloadInvoice: vi.fn(),
    cancelOrder: vi.fn(),
    requestRefund: vi.fn(),
  },
  showError: vi.fn(),
  showSuccess: vi.fn(),
  push: vi.fn(),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      locale: { value: 'zh-CN' },
      t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key),
    }),
  }
})

vi.mock('vue-router', () => ({
  useRouter: () => ({ push }),
}))

vi.mock('@/api/payment', () => ({
  paymentAPI,
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractI18nErrorMessage: () => 'common.error',
}))

vi.mock('@/components/payment/currency', () => ({
  formatPaymentAmount: (value: number, currency: string) => `${currency} ${value.toFixed(2)}`,
  normalizePaymentCurrency: (currency?: string) => currency || 'CNY',
}))

vi.mock('@/components/payment/orderUtils', () => ({
  formatOrderDateTime: (value: string) => value,
}))

const BaseDialogStub = {
  props: ['show', 'title'],
  emits: ['close'],
  template: `
    <section v-if="show" class="dialog">
      <h2>{{ title }}</h2>
      <slot />
      <slot name="footer" />
    </section>
  `,
}

function invoice(overrides: Record<string, unknown> = {}) {
  return {
    id: 7,
    user_id: 3,
    amount: 120,
    currency: 'CNY',
    invoice_type: 'vat_general',
    title: '上海示例科技有限公司',
    tax_number: '91310000TEST',
    remark: '',
    status: 'issued',
    file_name: 'invoice-7.pdf',
    created_at: '2026-07-01T10:00:00Z',
    updated_at: '2026-07-01T10:00:00Z',
    downloadable: true,
    ...overrides,
  }
}

function mountView() {
  return mount(UserOrdersView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        BaseDialog: BaseDialogStub,
        Icon: { template: '<span />' },
        Pagination: { template: '<div />' },
        OrderTable: {
          props: ['orders', 'loading'],
          template: '<div><slot v-for="order in orders" name="actions" :row="order" /></div>',
        },
        Select: {
          props: ['modelValue', 'options'],
          emits: ['update:modelValue', 'change'],
          template: `
            <select :value="modelValue" @change="$emit('update:modelValue', $event.target.value); $emit('change')">
              <option v-for="option in options" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          `,
        },
      },
    },
  })
}

describe('UserOrdersView invoices', () => {
  beforeEach(() => {
    paymentAPI.getMyOrders.mockReset().mockResolvedValue({ data: { items: [], total: 0 } })
    paymentAPI.getRefundEligibleProviders.mockReset().mockResolvedValue({ data: { provider_instance_ids: [] } })
    paymentAPI.getInvoiceSummary.mockReset().mockResolvedValue({
      data: {
        currency: 'CNY',
        eligible_amount: 300,
        requested_amount: 120,
        available_amount: 180,
      },
    })
    paymentAPI.getMyInvoices.mockReset().mockResolvedValue({
      data: {
        items: [invoice()],
        total: 1,
        page: 1,
        page_size: 20,
      },
    })
    paymentAPI.createInvoice.mockReset().mockResolvedValue({ data: invoice({ id: 8, status: 'pending', downloadable: false }) })
    paymentAPI.downloadInvoice.mockReset().mockResolvedValue({ data: new Blob(['pdf'], { type: 'application/pdf' }) })
    paymentAPI.cancelOrder.mockReset()
    paymentAPI.requestRefund.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
    push.mockReset()
    Object.defineProperty(URL, 'createObjectURL', {
      configurable: true,
      value: vi.fn(() => 'blob:invoice'),
    })
    Object.defineProperty(URL, 'revokeObjectURL', {
      configurable: true,
      value: vi.fn(),
    })
  })

  it('loads invoice summary and my invoice requests', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(paymentAPI.getInvoiceSummary).toHaveBeenCalled()
    expect(paymentAPI.getMyInvoices).toHaveBeenCalledWith({ page: 1, page_size: 20 })
    expect(wrapper.text()).toContain('payment.invoices.availableAmount')
    expect(wrapper.text()).toContain('payment.invoices.myRequests')
    expect(wrapper.text()).toContain('CNY 180.00')
    expect(wrapper.text()).toContain('payment.invoices.statuses.issued')
    expect(wrapper.text()).toContain('payment.invoices.download')
    expect(wrapper.find('.invoice-request-scroll').exists()).toBe(true)
    expect(wrapper.find('.orders-table-scroll').exists()).toBe(true)
  })

  it('submits an invoice request with amount, type, title, tax number, and remark', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find(button => button.text().includes('payment.invoices.apply'))?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('payment.invoices.applyHint')
    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('100')
    await inputs[1].setValue('杭州测试科技有限公司')
    await inputs[2].setValue('91330100TEST')
    await inputs[3].setValue('请开电子发票')
    await wrapper.findAll('button').find(button => button.text().includes('payment.invoices.types.vat_special'))?.trigger('click')
    await wrapper.findAll('button').find(button => button.text().includes('common.submit'))?.trigger('click')
    await flushPromises()

    expect(paymentAPI.createInvoice).toHaveBeenCalledWith({
      amount: 100,
      invoice_type: 'vat_special',
      title: '杭州测试科技有限公司',
      tax_number: '91330100TEST',
      remark: '请开电子发票',
    })
    expect(showSuccess).toHaveBeenCalledWith('payment.invoices.submitSuccess')
    expect(paymentAPI.getInvoiceSummary).toHaveBeenCalledTimes(2)
    expect(paymentAPI.getMyInvoices).toHaveBeenCalledTimes(2)
  })

  it('downloads an issued invoice from the authenticated API', async () => {
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined)
    const appendChild = vi.spyOn(document.body, 'appendChild')
    const removeChild = vi.spyOn(document.body, 'removeChild')

    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find(button => button.text().includes('payment.invoices.download'))?.trigger('click')
    await flushPromises()

    expect(paymentAPI.downloadInvoice).toHaveBeenCalledWith(7)
    expect(URL.createObjectURL).toHaveBeenCalled()
    expect(click).toHaveBeenCalled()
    expect(appendChild).toHaveBeenCalled()
    expect(removeChild).toHaveBeenCalled()
    expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:invoice')
  })
})
