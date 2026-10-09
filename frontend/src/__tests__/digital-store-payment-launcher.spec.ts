import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import StorePaymentLauncher from '@/components/store/StorePaymentLauncher.vue'

const toCanvas = vi.hoisted(() => vi.fn().mockResolvedValue(undefined))
vi.mock('qrcode', () => ({ default: { toCanvas } }))
vi.mock('vue-router', () => ({ useRouter: () => ({ resolve: vi.fn(() => ({ href: '/fixture-payment' })) }) }))
vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return { ...actual, useI18n: () => ({ t: (key: string) => ({ 'payment.qr.scanWxpay': '扫码微信支付' }[key] || key), locale: 'zh' }) }
})
vi.mock('@/stores/payment', () => ({ usePaymentStore: () => ({ pollOrderStatus: vi.fn() }) }))
vi.mock('@/stores', () => ({ useAppStore: () => ({ showError: vi.fn() }) }))

describe('StorePaymentLauncher QR first mount', () => {
  beforeEach(() => {
    toCanvas.mockClear()
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: false })))
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(null)
  })

  it('renders a returned WeChat URI into the mounted QR canvas on first launch', async () => {
    const wrapper = mount(StorePaymentLauncher, {
      global: {
        stubs: {
          BaseDialog: { props: ['show', 'title'], template: '<div v-if="show"><h3>{{ title }}</h3><slot /><slot name="footer" /></div>' },
          Icon: true,
        },
      },
    })
    try {
      await (wrapper.vm as unknown as { launch: (result: object, method: string) => Promise<void> }).launch({
        order_id: 44,
        payment_type: 'wxpay',
        qr_code: 'weixin://wxpay/fixture-uri',
        expires_at: '2099-01-01T00:00:00Z',
      }, 'wxpay')
      await flushPromises()
      expect(wrapper.text()).toContain('扫码微信支付')
      expect(toCanvas).toHaveBeenCalledWith(expect.any(HTMLCanvasElement), 'weixin://wxpay/fixture-uri', expect.any(Object))
    } finally {
      wrapper.unmount()
    }
  })
})
