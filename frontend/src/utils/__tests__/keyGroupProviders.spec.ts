import { describe, expect, it } from 'vitest'

import { getKeyGroupProvider } from '../keyGroupProviders'

describe('getKeyGroupProvider', () => {
  it('classifies every supported platform independently of its display name', () => {
    expect(getKeyGroupProvider('anthropic')).toBe('anthropic')
    expect(getKeyGroupProvider('openai')).toBe('openai')
    expect(getKeyGroupProvider('kimi')).toBe('domestic')
    expect(getKeyGroupProvider('zhipu')).toBe('domestic')
    expect(getKeyGroupProvider('deepseek')).toBe('domestic')
    expect(getKeyGroupProvider('gemini')).toBe('other')
    expect(getKeyGroupProvider('grok')).toBe('other')
    expect(getKeyGroupProvider('antigravity')).toBe('other')
  })

  it('places unknown runtime platforms in other', () => {
    expect(getKeyGroupProvider('future-provider')).toBe('other')
    expect(getKeyGroupProvider('constructor')).toBe('other')
    expect(getKeyGroupProvider('toString')).toBe('other')
    expect(getKeyGroupProvider('__proto__')).toBe('other')
  })
})
