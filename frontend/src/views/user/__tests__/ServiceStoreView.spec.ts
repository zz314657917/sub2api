import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync('src/views/user/ServiceStoreView.vue', 'utf8')
describe('ServiceStoreView', () => {
  it('uses catalogue API rather than demo products or subscription routing', () => {
    expect(source).toContain('digitalStoreAPI.getProducts')
    expect(source).toContain('digitalStoreAPI.createOrder')
    expect(source).not.toContain('Codex 接码')
    expect(source).not.toContain('to="/purchase"')
  })
  it('keeps CNY checkout and retry identity', () => {
    expect(source).toContain('limit.currency.toUpperCase() === "CNY"')
    expect(source).toContain('store.checkout.idempotency.')
    expect(source).toContain('wechat_resume_token')
  })
})
