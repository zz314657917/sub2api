import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const groupsViewSource = readFileSync(
  resolve(process.env.S222_FRONTEND_ROOT ?? process.cwd(), 'src/views/admin/GroupsView.vue'),
  'utf8'
)

describe('GroupsView usage column', () => {
  it('renders yesterday between today and total while keeping group pricing controls', () => {
    const today = groupsViewSource.indexOf('admin.groups.usageToday')
    const yesterday = groupsViewSource.indexOf('admin.groups.usageYesterday')
    const total = groupsViewSource.indexOf('admin.groups.usageTotal')

    expect(today).toBeGreaterThanOrEqual(0)
    expect(yesterday).toBeGreaterThan(today)
    expect(total).toBeGreaterThan(yesterday)
    expect(groupsViewSource).toContain('groupPricing')
  })
})
