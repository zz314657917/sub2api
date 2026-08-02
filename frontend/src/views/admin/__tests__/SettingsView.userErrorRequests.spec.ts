import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const settingsView = readFileSync(resolve(process.cwd(), 'src/views/admin/SettingsView.vue'), 'utf8')
const settingsApi = readFileSync(resolve(process.cwd(), 'src/api/admin/settings.ts'), 'utf8')

describe('admin SettingsView user error request switch', () => {
  it('round-trips the opt-in flag from API state through the form and update payload', () => {
    expect(settingsView).toContain('allow_user_view_error_requests: false')
    expect(settingsView).toContain('v-model="form.allow_user_view_error_requests"')
    expect(settingsView).toContain('(form as Record<string, unknown>)[key] = value')
    expect(settingsView).toContain(
      'allow_user_view_error_requests: form.allow_user_view_error_requests'
    )

    const apiFieldOccurrences = settingsApi.match(/allow_user_view_error_requests\??: boolean/g) ?? []
    expect(apiFieldOccurrences.length).toBeGreaterThanOrEqual(2)
  })
})
