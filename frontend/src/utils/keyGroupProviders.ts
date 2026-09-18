import type { GroupPlatform } from '@/types'

export type KeyGroupProvider = 'anthropic' | 'openai' | 'domestic' | 'other'

export const KEY_GROUP_PROVIDERS = ['anthropic', 'openai', 'domestic', 'other'] as const

// Group names are user-defined, so this must only classify the configured platform.
const PROVIDER_BY_PLATFORM: Partial<Record<GroupPlatform, KeyGroupProvider>> = {
  anthropic: 'anthropic',
  openai: 'openai',
  kimi: 'domestic',
  zhipu: 'domestic',
  deepseek: 'domestic',
  gemini: 'other',
  grok: 'other',
  antigravity: 'other'
}

export function getKeyGroupProvider(platform: GroupPlatform | string): KeyGroupProvider {
  return Object.prototype.hasOwnProperty.call(PROVIDER_BY_PLATFORM, platform)
    ? PROVIDER_BY_PLATFORM[platform as GroupPlatform]!
    : 'other'
}

export const KEY_GROUP_PROVIDER_ICONS: Record<KeyGroupProvider, GroupPlatform[]> = {
  anthropic: ['anthropic'],
  openai: ['openai'],
  domestic: ['deepseek', 'kimi'],
  other: ['gemini', 'grok']
}
