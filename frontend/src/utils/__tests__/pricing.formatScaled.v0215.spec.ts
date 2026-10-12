import { describe, expect, it } from 'vitest'
import { formatScaled } from '../pricing'

describe('formatScaled', () => {
  it.each([
    [1e-10, 1, '$1e-10'],
    [1.25e-10, 1, '$1.25e-10'],
    [1e-20, 1, '$1e-20'],
    [1e-16, 1_000_000, '$1e-10'],
    [1e10, 1, '$10000000000'],
  ])('preserves the magnitude of %s scaled by %s', (value, scale, expected) => {
    expect(formatScaled(value, scale)).toBe(expected)
  })

  it.each([
    [0.000003, 1_000_000, '$3'],
    [1.25e-8, 1_000_000, '$0.0125'],
    [5e-8, 1_000_000, '$0.05'],
    [0, 1, '$0'],
    [null, 1_000_000, '-'],
  ])('retains ordinary local pricing for %s', (value, scale, expected) => {
    expect(formatScaled(value, scale)).toBe(expected)
  })
})
