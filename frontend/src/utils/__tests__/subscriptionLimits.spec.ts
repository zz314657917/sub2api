import { describe, expect, it } from 'vitest'
import {
  displaySubscriptionLimit,
  hasAnySubscriptionLimit,
  hasPositiveSubscriptionLimit,
} from '../subscriptionLimits'

describe('subscriptionLimits', () => {
  it('only displays finite positive limits', () => {
    expect(displaySubscriptionLimit(25)).toBe(25)
    expect(displaySubscriptionLimit(0)).toBeNull()
    expect(displaySubscriptionLimit(-1)).toBeNull()
    expect(displaySubscriptionLimit(Number.NaN)).toBeNull()
    expect(displaySubscriptionLimit(null)).toBeNull()
    expect(displaySubscriptionLimit(undefined)).toBeNull()
  })

  it('reports whether any subscription cycle has a positive limit', () => {
    expect(hasPositiveSubscriptionLimit(1)).toBe(true)
    expect(hasPositiveSubscriptionLimit(0)).toBe(false)
    expect(hasAnySubscriptionLimit({ daily_limit_usd: 0, weekly_limit_usd: -1 })).toBe(false)
    expect(hasAnySubscriptionLimit({ daily_limit_usd: 0, weekly_limit_usd: 10 })).toBe(true)
  })
})
