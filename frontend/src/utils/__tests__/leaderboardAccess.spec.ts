import { describe, expect, it } from 'vitest'
import {
  hasLeaderboardAccountAge,
  LEADERBOARD_MINIMUM_ACCOUNT_AGE_MS,
  resolveLeaderboardMinimumAccountAgeDays,
} from '@/utils/leaderboardAccess'

describe('leaderboard account age access', () => {
  const now = Date.parse('2026-07-27T21:00:00.000Z')

  it('fails closed for missing or invalid registration timestamps', () => {
    expect(hasLeaderboardAccountAge(undefined, now)).toBe(false)
    expect(hasLeaderboardAccountAge('', now)).toBe(false)
    expect(hasLeaderboardAccountAge('not-a-date', now)).toBe(false)
  })

  it('uses a complete seven-day boundary', () => {
    expect(hasLeaderboardAccountAge(new Date(now - LEADERBOARD_MINIMUM_ACCOUNT_AGE_MS + 1).toISOString(), now)).toBe(false)
    expect(hasLeaderboardAccountAge(new Date(now - LEADERBOARD_MINIMUM_ACCOUNT_AGE_MS).toISOString(), now)).toBe(true)
  })

  it('uses the configured day count and keeps zero valid', () => {
    const createdAt = new Date(now - 2 * 24 * 60 * 60 * 1000).toISOString()
    expect(hasLeaderboardAccountAge(createdAt, now, 2)).toBe(true)
    expect(hasLeaderboardAccountAge(createdAt, now, 3)).toBe(false)
    expect(hasLeaderboardAccountAge(new Date(now).toISOString(), now, 0)).toBe(true)
  })

  it('falls back to seven days for invalid configuration', () => {
    expect(resolveLeaderboardMinimumAccountAgeDays(undefined)).toBe(7)
    expect(resolveLeaderboardMinimumAccountAgeDays(null)).toBe(7)
    expect(resolveLeaderboardMinimumAccountAgeDays(false)).toBe(7)
    expect(resolveLeaderboardMinimumAccountAgeDays('0')).toBe(7)
    expect(resolveLeaderboardMinimumAccountAgeDays(-1)).toBe(7)
    expect(resolveLeaderboardMinimumAccountAgeDays(3651)).toBe(7)
    expect(hasLeaderboardAccountAge(new Date(now - 6 * 24 * 60 * 60 * 1000).toISOString(), now, 'invalid')).toBe(false)
  })
})
