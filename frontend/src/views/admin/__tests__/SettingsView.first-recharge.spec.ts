import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const settingsView = readFileSync(resolve(process.cwd(), 'src/views/admin/SettingsView.vue'), 'utf8')

describe('admin SettingsView first recharge bonus setting', () => {
  it('exposes and submits the first recharge bonus amount', () => {
    expect(settingsView).toContain("t('admin.settings.features.welfare.firstRechargeBonusAmount')")
    expect(settingsView).toContain('v-model.number="form.welfare_first_recharge_bonus_amount"')
    expect(settingsView).toContain('welfare_first_recharge_bonus_amount: welfareFirstRechargeBonusAmount')
  })
})
