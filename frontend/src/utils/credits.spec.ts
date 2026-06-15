import { describe, expect, it } from 'vitest'

import { CREDIT_SYMBOL, formatCreditAmount, formatCreditCompact, formatCreditExact } from './credits'

describe('credits formatters', () => {
  it('formats credit amounts with the fixed credit symbol', () => {
    expect(CREDIT_SYMBOL).toBe('✪')
    expect(formatCreditAmount(12.3456)).toBe('✪ 12.3456')
    expect(formatCreditAmount(12.3, { minimumFractionDigits: 2, maximumFractionDigits: 2 })).toBe('✪ 12.30')
  })

  it('normalizes non-finite inputs to zero', () => {
    expect(formatCreditAmount(null)).toBe('✪ 0.00')
    expect(formatCreditAmount(undefined)).toBe('✪ 0.00')
    expect(formatCreditAmount(Number.NaN)).toBe('✪ 0.00')
    expect(formatCreditAmount(Number.POSITIVE_INFINITY)).toBe('✪ 0.00')
  })

  it('keeps negative values signed', () => {
    expect(formatCreditAmount(-2.5, { minimumFractionDigits: 2, maximumFractionDigits: 2 })).toBe('✪ -2.50')
  })

  it('uses compact and exact precision rules', () => {
    expect(formatCreditCompact(123.456)).toBe('✪ 123.46')
    expect(formatCreditCompact(12.34567)).toBe('✪ 12.3457')
    expect(formatCreditCompact(0.1234567)).toBe('✪ 0.123457')
    expect(formatCreditExact(1.2)).toBe('✪ 1.20000000')
  })
})
