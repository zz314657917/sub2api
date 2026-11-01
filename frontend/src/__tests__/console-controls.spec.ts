import { afterEach, describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import postcss from 'postcss'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import Select from '@/components/common/Select.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

const read = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf8')
const controls = postcss.parse(read('src/styles/console-controls.css'))
const consoleUi = postcss.parse(read('src/styles/console-ui.css'))

function declarations(root: postcss.Root, selector: string) {
  const values: Record<string, string> = {}
  root.walkRules((rule) => {
    if (!rule.selectors.includes(selector)) return
    rule.walkDecls((decl) => { values[decl.prop] = decl.value })
  })
  return values
}

afterEach(() => {
  document.body.innerHTML = ''
  document.documentElement.classList.remove('dark')
})

describe('shared console controls contract', () => {
  it('loads from AppLayout and removes the Usage-only marker and import', () => {
    expect(read('src/components/layout/AppLayout.vue')).toContain("import '@/styles/console-controls.css'")
    const usage = read('src/views/admin/UsageView.vue')
    expect(usage).not.toContain('console-shell-usage')
    expect(usage).not.toContain('usage-theme.css')
  })

  it('gates every rule on a live console shell, including root tokens and body portals', () => {
    controls.walkRules((rule) => {
      for (const selector of rule.selectors) {
        expect(selector).toMatch(/^:root(?:\.dark)?:has\(\.console-shell\)/)
      }
    })
    const shell = document.createElement('div')
    shell.className = 'console-shell'
    document.body.append(shell)
    expect(document.documentElement.matches(':root:has(.console-shell)')).toBe(true)
    document.documentElement.classList.add('dark')
    expect(document.documentElement.matches(':root.dark:has(.console-shell)')).toBe(true)
    shell.remove()
    expect(document.documentElement.matches(':root:has(.console-shell)')).toBe(false)
    expect(document.documentElement.matches(':root.dark:has(.console-shell)')).toBe(false)
    consoleUi.walkRules((rule) => {
      expect(rule.selector).not.toContain('select-dropdown-portal')
    })
  })

  it('provides separate warm light/dark portal tokens without unscoped fallback overrides', () => {
    const light = declarations(controls, ':root:has(.console-shell)')
    const dark = declarations(controls, ':root.dark:has(.console-shell)')
    expect(light['--console-control-surface']).toBe('#fffdf9')
    expect(dark['--console-control-surface']).toBe('#211f1c')
    expect(light['--console-control-accent']).toBe('#a9583e')
    expect(dark['--console-control-accent']).toBe('#e3a58c')
    for (const source of ['Select.vue', 'DateRangePicker.vue']) {
      const fallback = read(`src/components/common/${source}`)
      expect(fallback).toContain('bg-white dark:bg-dark-800')
      expect(fallback).not.toContain('--console-control-')
    }
  })

  it('covers modal controls, date input focus, selection and the selected check icon', () => {
    const selectors: string[] = []
    controls.walkRules((rule) => { selectors.push(...rule.selectors) })
    expect(selectors.some((selector) => selector.includes('.console-modal-content') && selector.includes('.select-trigger'))).toBe(true)
    expect(declarations(controls, ':root:has(.console-shell) .date-picker-dropdown .date-picker-input:focus'))
      .toMatchObject({ 'border-color': 'var(--console-control-accent)', 'box-shadow': '0 0 0 3px var(--console-control-ring)' })
    for (const selector of [
      '.select-dropdown-portal .select-option-selected',
      '.select-dropdown-portal .select-option-selected:is(:hover, .select-option-focused):not(.select-option-disabled)',
      '.date-picker-dropdown .date-picker-preset-active:hover'
    ]) {
      expect(declarations(controls, `:root:has(.console-shell) ${selector}`)).toMatchObject({
        background: 'var(--console-control-selected)', color: 'var(--console-control-accent)'
      })
    }
    expect(declarations(controls, ':root:has(.console-shell) .select-dropdown-portal .select-option-selected .text-primary-500').color)
      .toBe('var(--console-control-accent)')
  })

  it('keeps error hover/open/focus border, outline and ring red, preserving disabled interaction styling', () => {
    const scope = ':root:has(.console-shell) :is(.console-shell, .console-modal-content)'
    expect(declarations(controls, `${scope} .select-trigger.select-trigger-error:is(:hover, :focus, :focus-visible, .select-trigger-open)`))
      .toMatchObject({ 'border-color': 'var(--console-control-error)', 'outline-color': 'var(--console-control-error)' })
    expect(declarations(controls, `${scope} .select-trigger.select-trigger-error:is(:focus, :focus-visible, .select-trigger-open)`)['box-shadow'])
      .toBe('0 0 0 3px var(--console-control-error-ring)')
    expect(declarations(controls, `${scope} .select-trigger.select-trigger-disabled`))
      .toMatchObject({ cursor: 'not-allowed', opacity: '0.6', background: 'var(--console-control-surface-soft)' })
    expect(declarations(controls, ':root:has(.console-shell) .select-dropdown-portal .select-option-disabled'))
      .toMatchObject({ cursor: 'not-allowed', opacity: '0.4' })
    expect(read('src/styles/console-controls.css')).not.toMatch(/\.btn-(?:success|danger|warning)|\.text-red-|\n\.text-orange-500\s*\{/)
  })

  it('retains actual Select error, disabled and portal selection behavior', async () => {
    const wrapper = mount(Select, {
      attachTo: document.body,
      props: {
        modelValue: 'selected', error: true, searchable: false,
        options: [
          { value: 'selected', label: 'Selected' },
          { value: 'disabled', label: 'Disabled', disabled: true },
          { value: 'next', label: 'Next' }
        ]
      }
    })
    try {
      await wrapper.get('button').trigger('click')
      await nextTick()
      expect(wrapper.get('button').classes()).toEqual(expect.arrayContaining(['select-trigger-open', 'select-trigger-error']))
      const portal = document.body.querySelector('.select-dropdown-portal')!
      expect(portal.querySelector('.select-option-selected')?.getAttribute('aria-selected')).toBe('true')
      const disabled = portal.querySelector('.select-option-disabled')!
      disabled.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      await nextTick()
      expect(wrapper.emitted('update:modelValue')).toBeUndefined()
      expect(wrapper.get('button').attributes('aria-expanded')).toBe('true')
      await wrapper.setProps({ disabled: true })
      expect(document.body.querySelector('.select-dropdown-portal')).toBeNull()
      expect(wrapper.get('button').attributes('disabled')).toBeDefined()
      expect(wrapper.get('button').classes()).toContain('select-trigger-disabled')
    } finally {
      wrapper.unmount()
    }
  })

  it('shares chart controls across routes with active and hover styles independent of lazy component CSS', () => {
    const scope = ':root:has(.console-shell) .console-shell'
    expect(declarations(controls, `${scope} :is(.dashboard-segmented, .dashboard-panel-chip)`))
      .toMatchObject({ background: 'var(--console-control-surface-soft)', 'border-color': 'var(--console-control-border)' })
    expect(declarations(controls, `${scope} .dashboard-segmented-button--active`))
      .toMatchObject({ background: 'var(--console-control-selected)', color: 'var(--console-control-accent)', 'box-shadow': 'none' })
    expect(declarations(controls, `${scope} .dashboard-segmented-button--idle:hover`).color).toBe('var(--console-control-text)')
    // These real consumers must remain connected to the shared classes.
    for (const chart of ['ModelDistributionChart', 'GroupDistributionChart', 'EndpointDistributionChart']) {
      const source = read(`src/components/charts/${chart}.vue`)
      expect(source).toContain('dashboard-segmented')
      expect(source).toContain('dashboard-segmented-button--active')
    }
    expect(read('src/components/charts/TokenUsageTrend.vue')).toContain('dashboard-panel-chip')
    expect(read('src/views/admin/UsageView.vue')).toContain('<UsageStatsCards class="console-usage-stats"')
    expect(declarations(controls, `${scope} .console-usage-stats .text-orange-500`).color)
      .toBe('var(--console-control-accent)')
    expect(declarations(controls, `${scope} td.text-orange-500`).color)
      .toBe('var(--console-control-accent)')
    const orangeSelectors: string[] = []
    controls.walkRules((rule) => {
      orangeSelectors.push(...rule.selectors.filter((selector) => selector.includes('.text-orange-500')))
    })
    expect(orangeSelectors).toEqual([
      `${scope} td.text-orange-500`,
      `${scope} .console-usage-stats .text-orange-500`
    ])
    document.body.innerHTML = '<div class="console-shell"><div class="console-usage-stats"><span class="text-orange-500"></span></div><span class="text-orange-500 warning"></span></div>'
    const summarySelector = `${scope} .console-usage-stats .text-orange-500`
    expect(document.querySelector(summarySelector)).toBe(document.querySelector('.console-usage-stats span'))
    expect(document.querySelector('.warning')?.matches(summarySelector)).toBe(false)
  })

  it('uses the tooltip panel background for both arrow directions with body-portal fallbacks', () => {
    expect(declarations(consoleUi, '.console-tooltip-surface')).toMatchObject({
      '--console-tooltip-background': '#fffaf5', background: 'var(--console-tooltip-background)'
    })
    expect(declarations(consoleUi, '.dark .console-tooltip-surface')).toMatchObject({
      '--console-tooltip-background': 'var(--console-control-surface, #211f1c)',
      background: 'var(--console-tooltip-background)'
    })
    for (const [direction, border] of [['down', 'border-top-color'], ['up', 'border-bottom-color']]) {
      expect(declarations(consoleUi, `.console-tooltip-arrow-${direction}`)[border]).toBe('var(--console-tooltip-background, #fffaf5)')
      expect(declarations(consoleUi, `.dark .console-tooltip-arrow-${direction}`)[border]).toBe('var(--console-tooltip-background, #211f1c)')
    }
  })
})
