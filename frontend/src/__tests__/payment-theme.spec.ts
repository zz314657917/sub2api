import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const paymentView = readFileSync(resolve(process.cwd(), 'src/views/user/PaymentView.vue'), 'utf8')

describe('payment page theme', () => {
  it('binds purchase dark-mode surfaces with scoped-css-safe selectors', () => {
    expect(paymentView).toContain(':global(.dark .purchase-pricing-page)')
    expect(paymentView).toContain(':global(.dark .purchase-pricing-page::before)')
    expect(paymentView).toContain(':global(.dark .pricing-card)')
    expect(paymentView).toContain(':global(.dark .pricing-summary)')
    expect(paymentView).toContain(':global(.dark .pricing-input)')
    expect(paymentView).not.toMatch(/:global\(\.dark\)\s+\.purchase-pricing-page/)
    expect(paymentView).not.toMatch(/:global\(\.dark\)\s+\.pricing-/)
  })

  it('keeps subscription prices compact and aligned', () => {
    expect(paymentView).toContain('formatPaymentAmountCompact')
    expect(paymentView).toContain('function formatDisplayPaymentAmount')
    expect(paymentView).toContain('pricing-plan-price-row')
    expect(paymentView).toContain('{{ formatDisplayPaymentAmount(plan.price) }}')
    expect(paymentView).toContain('{{ formatDisplayPaymentAmount(selectedPlan.price) }}')
  })
})
