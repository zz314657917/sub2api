import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import PaymentMethodSelector from '../PaymentMethodSelector.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

const methods = [
  { type: 'alipay', fee_rate: 0, available: true },
  { type: 'wxpay', fee_rate: 0.6, available: true },
  { type: 'stripe', fee_rate: 0, available: true },
  { type: 'credit_card', fee_rate: 1.2, available: false },
]

describe('PaymentMethodSelector', () => {
  it('keeps all payment methods in a bounded responsive grid with accessible labels', () => {
    const wrapper = mount(PaymentMethodSelector, {
      props: { methods, selected: 'alipay' },
    })

    const grid = wrapper.get('[data-testid="payment-method-grid"]')
    expect(grid.classes()).toEqual(expect.arrayContaining(['grid', 'grid-cols-2', 'sm:grid-cols-3', 'lg:grid-cols-4']))
    expect(grid.classes()).not.toContain('sm:flex')

    const buttons = wrapper.findAll('button')
    expect(buttons).toHaveLength(4)
    for (const button of buttons) {
      expect(button.classes()).toContain('min-w-0')
      expect(button.attributes('title')).toMatch(/^payment\.methods\./)
    }

    const labels = wrapper.findAll('[data-testid="payment-method-label"]')
    expect(labels).toHaveLength(4)
    for (const label of labels) {
      expect(label.classes()).toEqual(expect.arrayContaining(['block', 'w-full', 'truncate']))
    }
  })

  it('keeps existing selection and availability behavior', async () => {
    const wrapper = mount(PaymentMethodSelector, {
      props: { methods, selected: 'alipay' },
    })

    await wrapper.findAll('button')[1].trigger('click')
    await wrapper.findAll('button')[3].trigger('click')

    expect(wrapper.emitted('select')).toEqual([['wxpay']])
  })
})
