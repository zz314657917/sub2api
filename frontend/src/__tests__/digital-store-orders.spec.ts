import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import StoreOrdersView from '@/views/user/StoreOrdersView.vue'

const api = vi.hoisted(() => ({ getOrders: vi.fn(), resumeOrder: vi.fn(), getDelivery: vi.fn(), download: vi.fn(), cancelOrder: vi.fn(), launch: vi.fn() }))
vi.mock('@/api/digitalStore', () => ({ digitalStoreAPI: api }))
vi.mock('@/api/payment', () => ({ paymentAPI: api }))

const order = { order_id: 42, product_name: '测试卡密', kind: 'card', price_cents: 100, payment_status: 'PENDING', delivery_status: 'reserved', created_at: '' }
function render() {
  return mount(StoreOrdersView, { global: { stubs: {
    AppLayout: { template: '<div><slot /></div>' },
    RouterLink: { template: '<a><slot /></a>' },
    BaseDialog: { template: '<div />' },
    StorePaymentLauncher: { template: '<div />', methods: { launch: api.launch } },
  } } })
}
function button(wrapper: ReturnType<typeof render>, text: string) {
  return wrapper.findAll('button').find(item => item.text() === text)!
}
beforeEach(() => {
  vi.resetAllMocks()
  api.getOrders.mockResolvedValue({ data: { items: [order], total: 25 } })
  api.resumeOrder.mockResolvedValue({ data: { order_id: 42, pay_url: 'fixture' } })
})
describe('商店订单交互', () => {
  it('分页请求真实页码', async () => {
    const wrapper = render()
    await flushPromises()
    await button(wrapper, '下一页').trigger('click')
    await flushPromises()
    expect(api.getOrders).toHaveBeenLastCalledWith({ page: 2, page_size: 20 })
    wrapper.unmount()
  })
  it('继续付款仅恢复原订单并防止重复点击', async () => {
    let finish!: (value: unknown) => void
    api.resumeOrder.mockImplementation(() => new Promise(resolve => { finish = resolve }))
    const wrapper = render()
    await flushPromises()
    await button(wrapper, '继续付款').trigger('click')
    await button(wrapper, '继续付款').trigger('click')
    expect(api.resumeOrder).toHaveBeenCalledTimes(1)
    expect(api.resumeOrder).toHaveBeenCalledWith(42)
    finish({ data: { order_id: 42 } })
    await flushPromises()
    expect(api.launch).toHaveBeenCalledWith({ order_id: 42 })
    wrapper.unmount()
  })
  it('请求失败显示错误并保留可重试按钮', async () => {
    api.resumeOrder.mockRejectedValue(new Error('network'))
    const wrapper = render()
    await flushPromises()
    await button(wrapper, '继续付款').trigger('click')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('不要重复下单')
    expect(button(wrapper, '继续付款').attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })
  it('取消经过确认后刷新原订单列表', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    const wrapper = render()
    await flushPromises()
    await button(wrapper, '取消订单').trigger('click')
    await flushPromises()
    expect(api.cancelOrder).toHaveBeenCalledWith(42)
    expect(api.getOrders).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })
})
