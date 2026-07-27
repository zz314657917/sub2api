export const LEADERBOARD_MINIMUM_ACCOUNT_AGE_DAYS_DEFAULT = 7
export const LEADERBOARD_MINIMUM_ACCOUNT_AGE_DAYS_MAX = 3650
export const LEADERBOARD_MINIMUM_ACCOUNT_AGE_MS =
  LEADERBOARD_MINIMUM_ACCOUNT_AGE_DAYS_DEFAULT * 24 * 60 * 60 * 1000

export function resolveLeaderboardMinimumAccountAgeDays(value: unknown): number {
  if (
    typeof value !== 'number'
    || !Number.isInteger(value)
    || value < 0
    || value > LEADERBOARD_MINIMUM_ACCOUNT_AGE_DAYS_MAX
  ) {
    return LEADERBOARD_MINIMUM_ACCOUNT_AGE_DAYS_DEFAULT
  }
  return value
}

export function hasLeaderboardAccountAge(
  createdAt: string | null | undefined,
  now = Date.now(),
  minimumAccountAgeDays?: unknown,
): boolean {
  if (!createdAt) return false

  const createdAtMs = Date.parse(createdAt)
  const minimumAgeMs = resolveLeaderboardMinimumAccountAgeDays(minimumAccountAgeDays) * 24 * 60 * 60 * 1000
  return Number.isFinite(createdAtMs) && now - createdAtMs >= minimumAgeMs
}
