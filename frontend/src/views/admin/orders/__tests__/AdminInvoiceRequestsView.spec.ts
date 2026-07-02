import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AdminInvoiceRequestsView from '../AdminInvoiceRequestsView.vue'

const { adminPaymentAPI, showError, showSuccess } = vi.hoisted(() => ({
  adminPaymentAPI: {
    getInvoices: vi.fn(),
    getInvoice: vi.fn(),
    approveInvoice: vi.fn(),
    rejectInvoice: vi.fn(),
    issueInvoice: vi.fn(),
  },
  showError: vi.fn(),
  showSuccess: vi.fn(),
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

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI,
  default: adminPaymentAPI,
}))

vi.mock('@/stores/app', () => ({
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
    id: 11,
    user_id: 5,
    user_email: 'buyer@example.com',
    user_name: 'buyer',
    amount: 260,
    currency: 'CNY',
    invoice_type: 'vat_general',
    title: '北京示例科技有限公司',
    tax_number: '91110000TEST',
    remark: '项目报销',
    status: 'pending',
    admin_note: '',
    invoice_no: '',
    file_name: '',
    created_at: '2026-07-01T10:00:00Z',
    updated_at: '2026-07-01T10:00:00Z',
    download_count: 0,
    downloadable: false,
    claimable: false,
    ...overrides,
  }
}

function mountView() {
  return mount(AdminInvoiceRequestsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        BaseDialog: BaseDialogStub,
        Icon: { template: '<span />' },
        Pagination: { template: '<div />' },
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

describe('AdminInvoiceRequestsView', () => {
  beforeEach(() => {
    adminPaymentAPI.getInvoices.mockReset().mockResolvedValue({
      data: {
        items: [invoice()],
        total: 1,
        page: 1,
        page_size: 20,
      },
    })
    adminPaymentAPI.getInvoice.mockReset().mockResolvedValue({ data: invoice({ admin_note: '已核对' }) })
    adminPaymentAPI.approveInvoice.mockReset().mockResolvedValue({ data: invoice({ status: 'approved' }) })
    adminPaymentAPI.rejectInvoice.mockReset().mockResolvedValue({ data: invoice({ status: 'rejected' }) })
    adminPaymentAPI.issueInvoice.mockReset().mockResolvedValue({ data: invoice({ status: 'issued', downloadable: true }) })
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('loads invoice requests and renders admin actions', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(adminPaymentAPI.getInvoices).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      status: undefined,
      user_id: undefined,
    })
    expect(wrapper.text()).toContain('buyer@example.com')
    expect(wrapper.text()).toContain('北京示例科技有限公司')
    expect(wrapper.text()).toContain('CNY 260.00')
    expect(wrapper.findAll('button').some(button => button.text().includes('payment.invoices.approve'))).toBe(true)
    expect(wrapper.findAll('button').some(button => button.text().includes('payment.invoices.reject'))).toBe(true)
    expect(wrapper.findAll('button').some(button => button.text().includes('payment.invoices.issue'))).toBe(true)
  })

  it('approves a pending invoice request', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find(button => button.text().includes('payment.invoices.approve'))?.trigger('click')
    await flushPromises()

    expect(adminPaymentAPI.approveInvoice).toHaveBeenCalledWith(11, {})
    expect(showSuccess).toHaveBeenCalledWith('common.success')
    expect(adminPaymentAPI.getInvoices).toHaveBeenCalledTimes(2)
  })

  it('uploads an issued invoice file with invoice number and admin note', async () => {
    const wrapper = mountView()
    await flushPromises()

    const issueButtons = wrapper.findAll('button').filter(button => button.text().includes('payment.invoices.issue'))
    await issueButtons[issueButtons.length - 1].trigger('click')
    await flushPromises()

    const textInputs = wrapper.findAll('input[type="text"]')
    await textInputs[0].setValue('INV-20260701-001')
    await wrapper.find('textarea').setValue('已上传电子发票')

    const file = new File(['pdf'], 'invoice.pdf', { type: 'application/pdf' })
    const fileInput = wrapper.find('input[type="file"]')
    Object.defineProperty(fileInput.element, 'files', {
      configurable: true,
      value: [file],
    })
    await fileInput.trigger('change')
    const confirmIssueButtons = wrapper.findAll('button').filter(button => button.text().includes('payment.invoices.issue'))
    await confirmIssueButtons[confirmIssueButtons.length - 1].trigger('click')
    await flushPromises()

    expect(adminPaymentAPI.issueInvoice).toHaveBeenCalledWith(11, expect.any(FormData))
    const form = adminPaymentAPI.issueInvoice.mock.calls[0][1] as FormData
    expect(form.get('file')).toBe(file)
    expect(form.get('invoice_no')).toBe('INV-20260701-001')
    expect(form.get('admin_note')).toBe('已上传电子发票')
    expect(showSuccess).toHaveBeenCalledWith('common.success')
    expect(adminPaymentAPI.getInvoices).toHaveBeenCalledTimes(2)
  })
})
