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
    expect(settingsView).not.toContain(':global(.dark) .settings-tabs')
    expect(settingsView).not.toContain(':global(.dark) .settings-tab')
    expect(settingsView).toContain('.settings-tabs-shell:is(.dark *)')
    expect(settingsView).toContain('dark:bg-slate-950/95')
    expect(settingsView).toContain('rgb(15 23 42 / 0.96)')
    expect(settingsView).toContain('rgb(2 6 23 / 0.9)')
    expect(settingsView).toContain('.settings-tabs-scroll:is(.dark *)')
    expect(settingsView).toContain('rgb(2 6 23 / 0.42)')
    expect(settingsView).toContain('dark:text-slate-400')
    expect(settingsView).toContain('.settings-tab:is(.dark *)::before')
    expect(settingsView).toContain('.settings-tab-active:is(.dark *)')
    expect(settingsView).toContain('rgb(204 120 92 / 0.18)')
    expect(settingsView).toContain('.settings-tab:is(.dark *):hover')
  })
})
