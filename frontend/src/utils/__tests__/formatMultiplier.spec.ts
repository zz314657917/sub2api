import { describe, expect, it } from 'vitest'
import { formatMultiplier } from '../formatters'

describe('formatMultiplier', () => {
  it('keeps significant decimals instead of rounding to two places', () => {
    expect(formatMultiplier(0.035)).toBe('0.035')
    expect(formatMultiplier(0.125)).toBe('0.125')
    expect(formatMultiplier(0.0625)).toBe('0.0625')
  })

  it('pads round values to at least two decimals', () => {
    expect(formatMultiplier(0.3)).toBe('0.30')
    expect(formatMultiplier(1)).toBe('1.00')
    expect(formatMultiplier(1.5)).toBe('1.50')
  })

  it('handles values below four decimal places', () => {
    expect(formatMultiplier(0.001)).toBe('0.001')
    expect(formatMultiplier(0.0001)).toBe('0.0001')
    expect(formatMultiplier(0.00005)).toBe('0.000050')
  })
})
