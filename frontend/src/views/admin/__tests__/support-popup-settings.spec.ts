import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const settingsView = readFileSync(resolve(process.cwd(), 'src/views/admin/SettingsView.vue'), 'utf8')

describe('admin support popup settings', () => {
  it('describes support popup items as configurable QR codes with captions', () => {
    expect(settingsView).toContain('客服弹窗')
    expect(settingsView).toContain('可添加多个二维码')
    expect(settingsView).toContain('添加二维码')
    expect(settingsView).toContain('客服二维码')
    expect(settingsView).toContain('二维码下方说明')
    expect(settingsView).toContain('二维码覆盖角标')
    expect(settingsView).toContain('form.support_popup_items')
  })
})

describe('admin settings tabs dark theme', () => {
  it('keeps the settings tabs on a dark surface in dark mode', () => {
    expect(settingsView).toContain(':global(.dark) .settings-tabs-shell')
    expect(settingsView).toContain('rgb(15 23 42 / 0.96)')
    expect(settingsView).toContain('rgb(2 6 23 / 0.9)')
    expect(settingsView).toContain(':global(.dark) .settings-tab-active')
    expect(settingsView).toContain('rgb(20 184 166 / 0.18)')
    expect(settingsView).toContain(':global(.dark) .settings-tab:hover')
  })
})
