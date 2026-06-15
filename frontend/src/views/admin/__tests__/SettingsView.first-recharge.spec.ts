import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const settingsView = readFileSync(resolve(process.cwd(), 'src/views/admin/SettingsView.vue'), 'utf8')

describe('admin SettingsView new-user first recharge gift settings', () => {
  it('exposes and submits the first recharge gift controls', () => {
    expect(settingsView).toContain("t('admin.settings.features.welfare.firstRechargeBonusAmount')")
    expect(settingsView).toContain('v-model.number="form.welfare_first_recharge_bonus_amount"')
    expect(settingsView).toContain("t('admin.settings.features.welfare.firstRechargeBonusValidDays')")
    expect(settingsView).toContain('v-model.number="form.welfare_first_recharge_bonus_valid_days"')
    expect(settingsView).toContain("t('admin.settings.features.welfare.firstRechargeBonusStackMonthly')")
    expect(settingsView).toContain('v-model="form.welfare_first_recharge_bonus_stack_monthly"')
    expect(settingsView).toContain('welfare_first_recharge_bonus_amount: welfareFirstRechargeBonusAmount')
    expect(settingsView).toContain('welfare_first_recharge_bonus_valid_days: welfareFirstRechargeBonusValidDays')
    expect(settingsView).toContain('welfare_first_recharge_bonus_stack_monthly:')
  })
})
