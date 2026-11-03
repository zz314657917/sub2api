import { describe, expect, it } from 'vitest'
import { normalizeGroupOpenAIFast, supportsGroupOpenAIFast } from '../groupsOpenAIFast'
import zh from '@/i18n/locales/zh/admin/groups'
import en from '@/i18n/locales/en/admin/groups'

describe('group OpenAI Fast platform normalization', () => {
  it.each(['openai', 'composite'])('supports %s and defaults missing flags to false', (platform) => {
    expect(supportsGroupOpenAIFast(platform)).toBe(true)
    expect(normalizeGroupOpenAIFast(platform, true)).toBe(true)
    expect(normalizeGroupOpenAIFast(platform, false)).toBe(false)
    expect(normalizeGroupOpenAIFast(platform, undefined)).toBe(false)
  })

  it.each(['anthropic', 'gemini', 'antigravity', 'grok', 'kimi', 'zhipu', 'deepseek', '', 'unknown'])(
    'clears the flag on unsupported platform %s',
    (platform) => {
      expect(supportsGroupOpenAIFast(platform)).toBe(false)
      expect(normalizeGroupOpenAIFast(platform, true)).toBe(false)
      expect(normalizeGroupOpenAIFast(platform, false)).toBe(false)
    },
  )

  it('explains global policy precedence and WebSocket reconnection in both locales', () => {
    expect(zh.openaiFast.force).toBe('强制使用 Fast（priority）')
    expect(zh.openaiFast.hint).toContain('全局 Fast/Flex 策略仍可过滤或拦截')
    expect(zh.openaiFast.hint).toContain('WebSocket 会话需重连')
    expect(en.openaiFast.hint).toContain('global Fast/Flex policy can still filter or block')
    expect(en.openaiFast.hint).toContain('WebSocket sessions must reconnect')
  })
})
