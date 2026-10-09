import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
const source = (path: string) => readFileSync(path, 'utf8')
describe('digital store API contract', () => {
  it('uses approved store routes without client amount input', () => {
    const api = source('src/api/digitalStore.ts')
    const types = source('src/types/digitalStore.ts')
    expect(api).toContain('"/store/products"')
    expect(api).toContain('"/store/orders"')
    expect(api).toContain('"/admin/store/files"')
    expect(types).not.toMatch(/CreateStoreOrderRequest[\s\S]*amount:/)
  })
})
