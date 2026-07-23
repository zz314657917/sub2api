import { describe, expect, it } from 'vitest'
import { planValiditySuffix } from '../validity'

const t = (key: string): string =>
  ({
    'payment.perMonth': 'month',
    'payment.days': 'days',
    'payment.weeks': 'weeks',
    'payment.months': 'months',
  })[key] ?? key

const suffix = (validity_days: number, validity_unit: string) =>
  planValiditySuffix({ validity_days, validity_unit }, t)

describe('planValiditySuffix', () => {
  it('renders singular and plural month units', () => {
    expect(suffix(1, 'months')).toBe('month')
    expect(suffix(3, 'months')).toBe('3months')
    expect(suffix(6, 'month')).toBe('6months')
  })

  it('renders singular and plural week units', () => {
    expect(suffix(2, 'weeks')).toBe('2weeks')
    expect(suffix(1, 'week')).toBe('1weeks')
  })

  it('renders day-based and unknown units as days', () => {
    expect(suffix(30, 'days')).toBe('30days')
    expect(suffix(30, 'day')).toBe('30days')
    expect(suffix(1, 'year')).toBe('1days')
    expect(suffix(365, 'unknown')).toBe('365days')
  })

  it('normalizes casing and whitespace', () => {
    expect(suffix(1, ' Months ')).toBe('month')
    expect(suffix(2, 'WEEKS')).toBe('2weeks')
  })
})
