import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const source = readFileSync(
  resolve(process.cwd(), 'src/views/admin/ops/components/OpsDashboardHeader.vue'),
  'utf8'
)

describe('OpsDashboardHeader SLA empty window', () => {
  it('renders a neutral empty state when the selected window has no SLA requests', () => {
    expect(source).toContain(
      'if ((overview.value?.request_count_sla ?? 0) <= 0) return null'
    )
    expect(source).toContain("slaPercent == null ? '-' : `${slaPercent.toFixed(3)}%`")
    expect(source).toContain(
      'class="font-bold text-gray-900 dark:text-white"'
    )
    expect(source).not.toContain(
      'class="font-bold text-red-600 dark:text-red-400"'
    )
  })
})
