import { describe, expect, it } from 'vitest'
import { buildGrokUsageRefreshKey, buildOpenAIUsageRefreshKey, getAccountPlanType } from '../accountUsageRefresh'

describe('buildOpenAIUsageRefreshKey', () => {
  it('会在 codex 快照变化时生成不同 key', () => {
    const base = {
      id: 1,
      platform: 'openai',
      type: 'oauth',
      updated_at: '2026-03-07T10:00:00Z',
      last_used_at: '2026-03-07T09:59:00Z',
      extra: {
        codex_usage_updated_at: '2026-03-07T10:00:00Z',
        codex_5h_used_percent: 0,
        codex_7d_used_percent: 0
      }
    } as any

    const next = {
      ...base,
      extra: {
        ...base.extra,
        codex_usage_updated_at: '2026-03-07T10:01:00Z',
        codex_5h_used_percent: 100
      }
    }

    expect(buildOpenAIUsageRefreshKey(base)).not.toBe(buildOpenAIUsageRefreshKey(next))
  })

  it('会在 last_used_at 变化时生成不同 key', () => {
    const base = {
      id: 3,
      platform: 'openai',
      type: 'oauth',
      updated_at: '2026-03-07T10:00:00Z',
      last_used_at: '2026-03-07T10:00:00Z',
      extra: {
        codex_usage_updated_at: '2026-03-07T10:00:00Z',
        codex_5h_used_percent: 12,
        codex_7d_used_percent: 24
      }
    } as any

    const next = {
      ...base,
      last_used_at: '2026-03-07T10:02:00Z'
    }

    expect(buildOpenAIUsageRefreshKey(base)).not.toBe(buildOpenAIUsageRefreshKey(next))
  })

  it('非 OpenAI OAuth 账号返回空 key', () => {
    expect(buildOpenAIUsageRefreshKey({
      id: 2,
      platform: 'anthropic',
      type: 'oauth',
      updated_at: '2026-03-07T10:00:00Z',
      last_used_at: '2026-03-07T10:00:00Z',
      extra: {}
    } as any)).toBe('')
  })
})

describe('buildGrokUsageRefreshKey', () => {
  const grokAccount = (extra: Record<string, unknown>) => ({ platform: 'grok', extra } as any)

  it('uses canonical tier before legacy snapshot', () => {
    const base = grokAccount({
      grok_usage_snapshot: { subscription_tier: 'premium', counters: { requests: 1 } },
      grok_quota_snapshot: { subscription_tier: 'legacy', stale: true }
    })
    const legacyChanged = grokAccount({
      grok_usage_snapshot: { counters: { requests: 1 }, subscription_tier: 'premium' },
      grok_quota_snapshot: { subscription_tier: 'legacy', stale: false }
    })
    const canonicalChanged = grokAccount({
      grok_usage_snapshot: { subscription_tier: 'premium', counters: { requests: 2 } },
      grok_quota_snapshot: { subscription_tier: 'legacy', stale: true }
    })

    expect(buildGrokUsageRefreshKey(base)).toBe(buildGrokUsageRefreshKey(legacyChanged))
    expect(buildGrokUsageRefreshKey(base)).not.toBe(buildGrokUsageRefreshKey(canonicalChanged))
  })

  it('keeps array order but normalizes object key order', () => {
    const first = grokAccount({ grok_billing_snapshot: { plan: 'pro', windows: [{ reset: 2, used: 1 }] } })
    const reordered = grokAccount({ grok_billing_snapshot: { windows: [{ used: 1, reset: 2 }], plan: 'pro' } })
    const arrayChanged = grokAccount({ grok_billing_snapshot: { plan: 'pro', windows: [{ used: 1, reset: 2 }, { used: 3 }] } })

    expect(buildGrokUsageRefreshKey(first)).toBe(buildGrokUsageRefreshKey(reordered))
    expect(buildGrokUsageRefreshKey(first)).not.toBe(buildGrokUsageRefreshKey(arrayChanged))
  })

  it('returns an empty key for non-Grok accounts', () => {
    expect(buildGrokUsageRefreshKey({ platform: 'openai', extra: {} } as any)).toBe('')
  })
})

describe('getAccountPlanType', () => {
  it('uses the documented Grok tier precedence', () => {
    expect(getAccountPlanType({
      platform: 'grok',
      credentials: { subscription_tier: 'credential', plan_type: 'plan', parent_plan_type: 'parent' },
      extra: {
        grok_billing_snapshot: { plan: 'billing' },
        grok_usage_snapshot: { subscription_tier: 'usage' },
        grok_quota_snapshot: { subscription_tier: 'legacy' },
        subscription_tier: 'extra'
      }
    } as any)).toBe('credential')
  })
})
