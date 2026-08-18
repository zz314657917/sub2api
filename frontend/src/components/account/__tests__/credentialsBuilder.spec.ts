import { describe, it, expect } from 'vitest'
import {
  applyInterceptWarmup,
  cnBalanceCellVisible,
  cnQuotaCellVisible,
  defaultCNBaseUrl,
} from '../credentialsBuilder'

describe('applyInterceptWarmup', () => {
  it('create + enabled=true: should set intercept_warmup_requests to true', () => {
    const creds: Record<string, unknown> = { access_token: 'tok' }
    applyInterceptWarmup(creds, true, 'create')
    expect(creds.intercept_warmup_requests).toBe(true)
  })

  it('create + enabled=false: should not add the field', () => {
    const creds: Record<string, unknown> = { access_token: 'tok' }
    applyInterceptWarmup(creds, false, 'create')
    expect('intercept_warmup_requests' in creds).toBe(false)
  })

  it('edit + enabled=true: should set intercept_warmup_requests to true', () => {
    const creds: Record<string, unknown> = { api_key: 'sk' }
    applyInterceptWarmup(creds, true, 'edit')
    expect(creds.intercept_warmup_requests).toBe(true)
  })

  it('edit + enabled=false + field exists: should delete the field', () => {
    const creds: Record<string, unknown> = { api_key: 'sk', intercept_warmup_requests: true }
    applyInterceptWarmup(creds, false, 'edit')
    expect('intercept_warmup_requests' in creds).toBe(false)
  })

  it('edit + enabled=false + field absent: should not throw', () => {
    const creds: Record<string, unknown> = { api_key: 'sk' }
    applyInterceptWarmup(creds, false, 'edit')
    expect('intercept_warmup_requests' in creds).toBe(false)
  })

  it('should not affect other fields', () => {
    const creds: Record<string, unknown> = {
      api_key: 'sk',
      base_url: 'url',
      intercept_warmup_requests: true
    }
    applyInterceptWarmup(creds, false, 'edit')
    expect(creds.api_key).toBe('sk')
    expect(creds.base_url).toBe('url')
    expect('intercept_warmup_requests' in creds).toBe(false)
  })
})

describe('CN provider credential presets', () => {
  it('resolves provider/mode/protocol defaults and rejects illegal combinations', () => {
    expect(defaultCNBaseUrl('kimi', 'payg', 'chat_completions')).toBe('https://api.moonshot.cn/v1')
    expect(defaultCNBaseUrl('kimi', 'coding', 'anthropic')).toBe('https://api.kimi.com/coding')
    expect(defaultCNBaseUrl('deepseek', 'payg', 'responses')).toBe('https://api.deepseek.com')
    expect(defaultCNBaseUrl('kimi', 'payg', 'responses')).toBe('')
    expect(defaultCNBaseUrl('deepseek', 'coding', 'chat_completions')).toBe('')
  })

  it('exposes quota and balance cells only for their owning plan types', () => {
    expect(cnQuotaCellVisible('kimi', 'coding')).toBe(true)
    expect(cnQuotaCellVisible('zhipu', 'coding')).toBe(true)
    expect(cnQuotaCellVisible('deepseek', 'coding')).toBe(false)
    expect(cnBalanceCellVisible('kimi', 'payg')).toBe(true)
    expect(cnBalanceCellVisible('deepseek', 'payg')).toBe(true)
    expect(cnBalanceCellVisible('zhipu', 'payg')).toBe(false)
  })
})
