import { describe, expect, it } from 'vitest'
import {
  apiKeyGroups,
  apiKeyOpenAIImageGroups,
  apiKeySupportsUnifiedAccess,
  apiKeySupportsChat,
  apiKeySupportsOpenAI,
  apiKeySupportsOpenAIImageGeneration,
  apiKeySupportsVideoGeneration,
  apiKeyVideoGroups,
  primaryAPIKeyGroupName,
  primaryAPIKeyImageGroupName,
} from './apiKeyCapabilities'
import type { ApiKey, Group } from '@/types'

function group(overrides: Partial<Group>): Group {
  return {
    id: 1,
    name: 'group',
    description: null,
    platform: 'openai',
    rate_multiplier: 1,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'standard',
    routing_scope: 'inference',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    allow_image_generation: false,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    require_oauth_only: false,
    require_privacy_set: false,
    created_at: '2026-06-06T00:00:00Z',
    updated_at: '2026-06-06T00:00:00Z',
    ...overrides,
  }
}

function apiKey(overrides: Partial<ApiKey>): ApiKey {
  return {
    id: 1,
    user_id: 1,
    key: 'sk-test',
    name: 'smart key',
    group_id: null,
    multi_group_routes: [],
    account_pool_strategy: 'shared_only',
    status: 'active',
    ip_whitelist: [],
    ip_blacklist: [],
    last_used_at: null,
    quota: 0,
    quota_used: 0,
    expires_at: null,
    created_at: '2026-06-06T00:00:00Z',
    updated_at: '2026-06-06T00:00:00Z',
    rate_limit_5h: 0,
    rate_limit_1d: 0,
    rate_limit_7d: 0,
    usage_5h: 0,
    usage_1d: 0,
    usage_7d: 0,
    window_5h_start: null,
    window_1d_start: null,
    window_7d_start: null,
    reset_5h_at: null,
    reset_1d_at: null,
    reset_7d_at: null,
    ...overrides,
  }
}

describe('apiKeyCapabilities', () => {
  it('detects image capability from route groups when the default group is text only', () => {
    const key = apiKey({
      group_id: 10,
      group: group({ id: 10, name: 'text', platform: 'anthropic' }),
      multi_group_routes: [
        { group_id: 11, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true },
      ],
      route_groups: [
        group({ id: 11, name: 'image', platform: 'openai', routing_scope: 'image', allow_image_generation: true }),
      ],
    })

    expect(apiKeyGroups(key).map((item) => item.id)).toEqual([10, 11])
    expect(apiKeySupportsChat(key)).toBe(true)
    expect(apiKeySupportsOpenAI(key)).toBe(true)
    expect(apiKeyOpenAIImageGroups(key).map((item) => item.id)).toEqual([11])
    expect(apiKeySupportsOpenAIImageGeneration(key)).toBe(true)
    expect(primaryAPIKeyGroupName(key)).toBe('text')
    expect(primaryAPIKeyImageGroupName(key)).toBe('image')
  })

  it('detects chat capability from route groups when the default group is missing', () => {
    const key = apiKey({
      group_id: null,
      group: undefined,
      multi_group_routes: [
        { group_id: 12, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true },
      ],
      route_groups: [
        group({ id: 12, name: 'gemini', platform: 'gemini' }),
      ],
    })

    expect(apiKeyGroups(key).map((item) => item.id)).toEqual([12])
    expect(apiKeySupportsChat(key)).toBe(true)
    expect(primaryAPIKeyGroupName(key)).toBe('gemini')
  })

  it('deduplicates default and route groups', () => {
    const sameGroup = group({ id: 11, name: 'image', platform: 'openai', routing_scope: 'image', allow_image_generation: true })
    const key = apiKey({
      group_id: 11,
      group: sameGroup,
      multi_group_routes: [
        { group_id: 11, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true },
      ],
      route_groups: [sameGroup],
    })

    expect(apiKeyGroups(key)).toHaveLength(1)
    expect(apiKeySupportsOpenAIImageGeneration(key)).toBe(true)
  })

  it('rejects inactive keys even when route groups support images', () => {
    const key = apiKey({
      status: 'inactive',
      multi_group_routes: [
        { group_id: 11, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true },
      ],
      route_groups: [
        group({ id: 11, platform: 'openai', routing_scope: 'image', allow_image_generation: true }),
      ],
    })

    expect(apiKeySupportsOpenAI(key)).toBe(false)
    expect(apiKeySupportsOpenAIImageGeneration(key)).toBe(false)
  })

  it('ignores disabled and text-only route groups for image capability', () => {
    const key = apiKey({
      multi_group_routes: [
        { group_id: 11, priority: 1, weight: 1, cooldown_seconds: 30, enabled: false },
        { group_id: 12, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true, text_only: true },
      ],
      route_groups: [
        group({ id: 11, platform: 'openai', allow_image_generation: true }),
        group({ id: 12, platform: 'openai', allow_image_generation: true }),
      ],
    })

    expect(apiKeyGroups(key).map((item) => item.id)).toEqual([12])
    expect(apiKeyOpenAIImageGroups(key)).toHaveLength(0)
    expect(apiKeySupportsOpenAIImageGeneration(key)).toBe(false)
  })

  it('ignores inactive route groups', () => {
    const key = apiKey({
      multi_group_routes: [
        { group_id: 11, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true },
      ],
      route_groups: [
        group({ id: 11, status: 'inactive', platform: 'openai', allow_image_generation: true }),
      ],
    })

    expect(apiKeyGroups(key)).toHaveLength(0)
    expect(apiKeySupportsOpenAIImageGeneration(key)).toBe(false)
  })

  it('ignores image-only route groups for chat capability', () => {
    const key = apiKey({
      multi_group_routes: [
        { group_id: 11, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true, image_only: true },
      ],
      route_groups: [
        group({ id: 11, name: 'image', platform: 'openai', allow_image_generation: true }),
      ],
    })

    expect(apiKeySupportsChat(key)).toBe(false)
  })

  it('detects video capability from routing scope only', () => {
    const key = apiKey({
      group_id: 10,
      group: group({ id: 10, name: 'text', platform: 'openai', routing_scope: 'inference' }),
      multi_group_routes: [
        { group_id: 10, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true, text_only: true },
        { group_id: 11, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true, model_patterns: ['sora-*'] },
        { group_id: 12, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true },
      ],
      route_groups: [
        group({ id: 11, name: 'sora', platform: 'openai', routing_scope: 'inference' }),
        group({ id: 12, name: 'video', platform: 'openai', routing_scope: 'video' }),
      ],
    })

    expect(apiKeyVideoGroups(key).map((item) => item.id)).toEqual([12])
    expect(apiKeySupportsVideoGeneration(key)).toBe(true)
  })

  it('detects video capability from a single default video group', () => {
    const key = apiKey({
      group_id: 12,
      group: group({ id: 12, name: 'video', platform: 'openai', routing_scope: 'video' }),
      multi_group_routes: [],
      route_groups: [],
    })

    expect(apiKeyVideoGroups(key).map((item) => item.id)).toEqual([12])
    expect(apiKeySupportsVideoGeneration(key)).toBe(true)
  })

  it('does not count image or video scoped groups as chat capability', () => {
    const key = apiKey({
      group_id: 11,
      group: group({ id: 11, name: 'image', platform: 'openai', routing_scope: 'image', allow_image_generation: true }),
      multi_group_routes: [
        { group_id: 12, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true, model_patterns: ['doubao-seedance-*'] },
      ],
      route_groups: [
        group({ id: 12, name: 'video', platform: 'openai', routing_scope: 'video' }),
      ],
    })

    expect(apiKeySupportsChat(key)).toBe(false)
    expect(apiKeySupportsOpenAIImageGeneration(key)).toBe(true)
    expect(apiKeySupportsVideoGeneration(key)).toBe(true)
    expect(apiKeySupportsUnifiedAccess(key)).toBe(false)
  })

  it('detects unified access when one active key covers chat image and video routes', () => {
    const key = apiKey({
      group_id: 10,
      group: group({ id: 10, name: 'text', platform: 'openai', routing_scope: 'inference' }),
      multi_group_routes: [
        { group_id: 10, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true, text_only: true },
        { group_id: 11, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true, image_only: true },
        { group_id: 12, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true, model_patterns: ['doubao-seedance-*'] },
      ],
      route_groups: [
        group({ id: 11, name: 'image', platform: 'openai', routing_scope: 'image', allow_image_generation: true }),
        group({ id: 12, name: 'video', platform: 'openai', routing_scope: 'video' }),
      ],
    })

    expect(apiKeySupportsChat(key)).toBe(true)
    expect(apiKeySupportsOpenAIImageGeneration(key)).toBe(true)
    expect(apiKeySupportsVideoGeneration(key)).toBe(true)
    expect(apiKeySupportsUnifiedAccess(key)).toBe(true)
  })
})
