import { describe, expect, it } from 'vitest'
import { formatPaymentAmount, formatPaymentAmountCompact } from '../currency'

describe('formatPaymentAmount', () => {
  it('uses the currency default fraction digits', () => {
    expect(formatPaymentAmount(100, 'JPY', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'KRW', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'HKD', 'en-US')).toContain('.00')
  })
})

describe('formatPaymentAmountCompact', () => {
  it('removes insignificant trailing zeros', () => {
    expect(formatPaymentAmountCompact(80, 'CNY', 'zh-CN')).not.toContain('.00')
    expect(formatPaymentAmountCompact(699.5, 'CNY', 'zh-CN')).toContain('.5')
  })
})
