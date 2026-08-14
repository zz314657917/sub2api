import type { Account } from '@/types'

const normalizeUsageRefreshValue = (value: unknown): string => {
  if (value == null) return ''
  return String(value)
}

const asRecord = (value: unknown): Record<string, unknown> => {
  return value != null && typeof value === 'object' && !Array.isArray(value)
    ? value as Record<string, unknown>
    : {}
}

const nonBlankString = (value: unknown): string => {
  return typeof value === 'string' && value.trim() ? value.trim() : ''
}

const stableSnapshotValue = (value: unknown): string => {
  if (value == null || typeof value !== 'object') return JSON.stringify(value)
  if (Array.isArray(value)) return `[${value.map(stableSnapshotValue).join(',')}]`

  const record = value as Record<string, unknown>
  return `{${Object.keys(record).sort().map(key => `${JSON.stringify(key)}:${stableSnapshotValue(record[key])}`).join(',')}}`
}

export const buildOpenAIUsageRefreshKey = (account: Pick<Account, 'id' | 'platform' | 'type' | 'updated_at' | 'last_used_at' | 'rate_limit_reset_at' | 'extra'>): string => {
  if (account.platform !== 'openai' || account.type !== 'oauth') {
    return ''
  }

  const extra = account.extra ?? {}
  return [
    account.id,
    account.updated_at,
    account.last_used_at,
    account.rate_limit_reset_at,
    extra.codex_usage_updated_at,
    extra.codex_5h_used_percent,
    extra.codex_5h_reset_at,
    extra.codex_5h_reset_after_seconds,
    extra.codex_5h_window_minutes,
    extra.codex_7d_used_percent,
    extra.codex_7d_reset_at,
    extra.codex_7d_reset_after_seconds,
    extra.codex_7d_window_minutes
  ].map(normalizeUsageRefreshValue).join('|')
}

export const getAccountPlanType = (account: Pick<Account, 'platform' | 'credentials' | 'extra'>): string => {
  const credentials = asRecord(account.credentials)
  if (account.platform !== 'grok') return nonBlankString(credentials.plan_type)

  const extra = asRecord(account.extra)
  const billingSnapshot = asRecord(extra.grok_billing_snapshot)
  const usageSnapshot = asRecord(extra.grok_usage_snapshot)
  const legacySnapshot = asRecord(extra.grok_quota_snapshot)
  return nonBlankString(credentials.subscription_tier) ||
    nonBlankString(billingSnapshot.plan) ||
    nonBlankString(usageSnapshot.subscription_tier) ||
    nonBlankString(legacySnapshot.subscription_tier) ||
    nonBlankString(extra.subscription_tier) ||
    nonBlankString(credentials.plan_type) ||
    nonBlankString(credentials.parent_plan_type)
}

export const buildGrokUsageRefreshKey = (account: Pick<Account, 'platform' | 'extra'>): string => {
  if (account.platform !== 'grok') return ''

  const extra = asRecord(account.extra)
  const billingSnapshot = asRecord(extra.grok_billing_snapshot)
  const usageSnapshot = asRecord(extra.grok_usage_snapshot)
  const legacySnapshot = asRecord(extra.grok_quota_snapshot)
  const canonicalTier = nonBlankString(billingSnapshot.plan) || nonBlankString(usageSnapshot.subscription_tier)
  const snapshots = [billingSnapshot, usageSnapshot]
  if (!canonicalTier) snapshots.push(legacySnapshot)
  return snapshots.map(stableSnapshotValue).join('|')
}
