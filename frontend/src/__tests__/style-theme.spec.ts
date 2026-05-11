import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const styleCss = readFileSync(resolve(process.cwd(), 'src/style.css'), 'utf8')
const tailwindConfig = readFileSync(resolve(process.cwd(), 'tailwind.config.js'), 'utf8')

describe('MC visual theme contract', () => {
  it('defines shared MC-inspired theme tokens', () => {
    expect(styleCss).toContain('--mc-grass')
    expect(styleCss).toContain('--mc-stone')
    expect(styleCss).toContain('--mc-dirt')
    expect(styleCss).toContain('.mc-pixel-grid')
    expect(tailwindConfig).toContain('#3f7f3f')
    expect(tailwindConfig).toContain('#8b6f47')
  })

  it('keeps shared components blocky and usable across pages', () => {
    expect(styleCss).toContain('box-shadow: var(--mc-shadow-block)')
    expect(styleCss).toContain('border-color: var(--mc-border)')
    expect(styleCss).toMatch(/\.card\s*\{[\s\S]*@apply bg-white dark:bg-dark-800\/80;[\s\S]*@apply rounded-lg;/)
    expect(styleCss).toMatch(/\.btn\s*\{[\s\S]*@apply rounded-md/)
    expect(styleCss).toMatch(/\.rounded-2xl,\s*[\s\S]*\.rounded-3xl\s*\{[\s\S]*border-radius: 0\.5rem;/)
  })
})
