import { afterEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import UserOrdersView from '../UserOrdersView.vue'
import Select from '@/components/common/Select.vue'
import Pagination from '@/components/common/Pagination.vue'

const getMyOrders = vi.hoisted(() => vi.fn())
vi.mock('@/api/payment', () => ({ paymentAPI: { getMyOrders, getRefundEligibleProviders: vi.fn().mockResolvedValue({ data: { provider_instance_ids: [] } }), getInvoiceSummary: vi.fn().mockResolvedValue({ data: {} }), getMyInvoices: vi.fn().mockResolvedValue({ data: { items: [] } }) } }))
vi.mock('@/stores', () => ({ useAppStore: () => ({ showError: vi.fn() }) }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key, locale: { value: 'en' } })
}))
enableAutoUnmount(afterEach)

describe('order status filtering', () => {
  it('resets to page one when changing status', async () => {
    getMyOrders.mockResolvedValue({ data: { items: [], total: 100 } })
    const wrapper = mount(UserOrdersView, { global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, OrderTable: true, BaseDialog: true, Icon: true, teleport: true } } })
    await flushPromises()
    wrapper.getComponent(Pagination).vm.$emit('update:page', 3)
    await flushPromises()
    expect(getMyOrders).toHaveBeenLastCalledWith(expect.objectContaining({ page: 3, status: undefined }))
    const select = wrapper.getComponent(Select)
    select.vm.$emit('update:modelValue', 'PENDING')
    select.vm.$emit('change')
    await flushPromises()
    expect(getMyOrders).toHaveBeenLastCalledWith(expect.objectContaining({ page: 1, status: 'PENDING' }))
  })
})
