import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const viewSource = readFileSync(resolve(process.cwd(), 'src/views/admin/orders/AdminOrdersView.vue'), 'utf8')
const apiSource = readFileSync(resolve(process.cwd(), 'src/api/admin/payment.ts'), 'utf8')

describe('AdminOrdersView order date filters', () => {
  it('keeps list and stats requests on the same order filter params', () => {
    expect(apiSource).toContain("'/admin/payment/orders/stats'")
    expect(viewSource).toContain('function buildOrderQueryParams()')
    expect(viewSource).toContain('...activeDateRange()')
    expect(viewSource).toContain('adminPaymentAPI.getOrders({')
    expect(viewSource).toContain('adminPaymentAPI.getOrderStats(buildOrderQueryParams())')
  })

  it('exposes recent date presets and custom date inputs on the admin order list', () => {
    expect(viewSource).toContain("type DatePreset = 'all' | '7d' | '30d' | 'custom'")
    expect(viewSource).toContain("const datePreset = ref<DatePreset>('7d')")
    expect(viewSource).toContain("t('payment.admin.last7Days')")
    expect(viewSource).toContain("t('payment.admin.last30Days')")
    expect(viewSource).toContain('type="date"')
  })
})
