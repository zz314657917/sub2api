import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

describe('AccountsView priority column defaults', () => {
  it('shows priority for fresh tables while preserving saved hidden-column preferences', () => {
    const source = readFileSync(resolve(process.cwd(), 'src/views/admin/AccountsView.vue'), 'utf8')
    const defaultColumns = source.match(/const DEFAULT_HIDDEN_COLUMNS = \[(.*?)\]/s)?.[1] ?? ''

    expect(defaultColumns).not.toContain("'priority'")
    expect(source).toContain("JSON.parse(saved)")
    expect(source).toContain('hiddenColumns.add(key)')
  })
})
