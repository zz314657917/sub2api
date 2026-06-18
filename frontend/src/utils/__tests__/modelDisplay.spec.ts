import { describe, expect, it } from 'vitest'
import { cleanModelDisplayName, displayModelChain, displayModelLabel } from '../modelDisplay'

describe('modelDisplay', () => {
  it('removes upstream brand suffixes from model labels', () => {
    expect(displayModelLabel('grok-imagine-1.5-apimart')).toBe('grok-imagine-1.5')
    expect(displayModelLabel('grok-imagine-1.5-edit-apimart')).toBe('grok-imagine-1.5-edit')
    expect(displayModelLabel('gpt-image-2 (Api Mart)')).toBe('gpt-image-2')
    expect(displayModelLabel('gemini-3-pro-image_api-mart')).toBe('gemini-3-pro-image')
  })

  it('does not leak brand-only values through fallback text', () => {
    expect(cleanModelDisplayName('apimart')).toBe('模型')
    expect(cleanModelDisplayName('', 'Api Mart')).toBe('模型')
    expect(cleanModelDisplayName('', 'grok-imagine-1.5-apimart')).toBe('grok-imagine-1.5')
  })

  it('sanitizes mapping chain display while preserving separators', () => {
    expect(displayModelChain('gpt-image-2 → grok-imagine-1.5-apimart')).toBe('gpt-image-2→grok-imagine-1.5')
  })
})
