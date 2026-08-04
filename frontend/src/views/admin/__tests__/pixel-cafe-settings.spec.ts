import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const settingsView = readFileSync(resolve(process.cwd(), 'src/views/admin/SettingsView.vue'), 'utf8')

describe('Pixel Cafe presentation settings', () => {
  it('exposes title, description and header visibility controls in the existing settings card', () => {
    expect(settingsView).toContain('data-testid="pixel-cafe-title"')
    expect(settingsView).toContain('data-testid="pixel-cafe-description"')
    expect(settingsView).toContain('data-testid="pixel-cafe-header-visible"')
    expect(settingsView).toContain('pixel_cafe_title: form.pixel_cafe_title?.trim() || \'像素网吧\'')
    expect(settingsView).toContain('pixel_cafe_description: form.pixel_cafe_description?.trim() || \'\'')
    expect(settingsView).toContain('pixel_cafe_header_visible: form.pixel_cafe_header_visible')
  })
})
