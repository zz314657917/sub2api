export interface SubscriptionLimitSource {
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
}

export function displaySubscriptionLimit(value: number | null | undefined): number | null {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : null
}

export function hasPositiveSubscriptionLimit(value: number | null | undefined): boolean {
  return displaySubscriptionLimit(value) != null
}

export function hasAnySubscriptionLimit(source: SubscriptionLimitSource | null | undefined): boolean {
  return (
    hasPositiveSubscriptionLimit(source?.daily_limit_usd) ||
    hasPositiveSubscriptionLimit(source?.weekly_limit_usd) ||
    hasPositiveSubscriptionLimit(source?.monthly_limit_usd)
  )
}
