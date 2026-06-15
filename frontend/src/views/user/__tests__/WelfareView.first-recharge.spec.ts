import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const welfareView = readFileSync(resolve(process.cwd(), 'src/views/user/WelfareView.vue'), 'utf8')

describe('WelfareView new-user first recharge gift card', () => {
  it('renders the first recharge gift card and links to the recharge page', () => {
    expect(welfareView).toContain("t('welfare.recharge.title')")
    expect(welfareView).toContain("t('welfare.recharge.description'")
    expect(welfareView).toContain("t('welfare.recharge.expiresAt'")
    expect(welfareView).toContain("t('welfare.recharge.monthlyBonusNotice')")
    expect(welfareView).toContain('data-testid="welfare-recharge-go"')
    expect(welfareView).toContain("router.push('/purchase')")
  })
})
