import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')
const consoleUiPath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../styles/console-ui.css')
const consoleUiSource = readFileSync(consoleUiPath, 'utf8')
const onboardingTourPath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../composables/useOnboardingTour.ts')
const onboardingTourSource = readFileSync(onboardingTourPath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar admin navigation groups', () => {
  it('keeps admin management entries inside grouped sidebar sections', () => {
    expect(componentSource).toContain("label: t('nav.basicManagement')")
    expect(componentSource).toContain("label: t('nav.resourceAccess')")
    expect(componentSource).toContain("label: t('nav.businessOperations')")
    expect(componentSource).toContain("label: t('nav.contentRecords')")
    expect(componentSource).toContain("path: '/admin/users'")
    expect(componentSource).toContain("path: '/admin/accounts'")
    expect(componentSource).toContain("path: '/admin/redeem'")
  })

  it('keeps subscription plans beside subscription management instead of inside orders', () => {
    const businessStart = componentSource.indexOf("label: t('nav.businessOperations')")
    const ordersStart = componentSource.indexOf("path: '/admin/orders'", businessStart)
    const ordersChildrenStart = componentSource.indexOf('children: [', ordersStart)
    const ordersChildrenEnd = componentSource.indexOf('],', ordersChildrenStart)
    expect(businessStart).toBeGreaterThan(-1)
    expect(ordersStart).toBeGreaterThan(-1)
    expect(ordersChildrenStart).toBeGreaterThan(-1)
    expect(componentSource.indexOf("path: '/admin/subscriptions'", businessStart)).toBeLessThan(ordersStart)
    expect(componentSource.indexOf("path: '/admin/orders/plans'", businessStart)).toBeLessThan(ordersStart)
    expect(componentSource.slice(ordersChildrenStart, ordersChildrenEnd)).not.toContain("path: '/admin/orders/plans'")
  })

  it('preserves admin onboarding anchors after grouping', () => {
    expect(componentSource).toContain('function getAdminMenuItemId')
    expect(componentSource).toContain('sidebarTourTargetGroups')
    expect(componentSource).toContain("'sidebar-group-manage'")
    expect(componentSource).toContain("'sidebar-channel-manage'")
    expect(componentSource).toContain("'sidebar-wallet'")
    expect(onboardingTourSource).toContain("'sub2api:sidebar-tour-target'")
  })
})

describe('AppSidebar creation center navigation', () => {
  it('shows creation center as a section title without a collapsible parent', () => {
    expect(componentSource).toContain("t('nav.creationCenter')")
    expect(componentSource).not.toContain("path: '/studio'")
    expect(componentSource).toContain("label: t('nav.chatCreator')")
    expect(componentSource).toContain("path: '/chat'")
    expect(componentSource).toContain("path: '/image-creator'")
    expect(componentSource.indexOf("path: '/chat'")).toBeLessThan(
      componentSource.indexOf("path: '/image-creator'")
    )
  })

  it('keeps creation links aligned with regular sidebar links in the console shell', () => {
    const sidebarLinkBlock = consoleUiSource.match(/\.console-sidebar \.sidebar-link\s*\{[\s\S]*?\n\}/)?.[0] ?? ''
    const sidebarSectionBlock = consoleUiSource.match(/\.console-sidebar \.sidebar-section\s*\{[\s\S]*?\n\}/)?.[0] ?? ''

    expect(sidebarLinkBlock).toContain('width: 100%;')
    expect(sidebarLinkBlock).toContain('box-sizing: border-box;')
    expect(sidebarSectionBlock).toContain('width: 100%;')
    expect(componentSource).toContain("path: '/image-creator'")
    expect(componentSource).toContain("path: '/dashboard'")
  })
})

describe('AppSidebar self navigation groups', () => {
  it('keeps account entries grouped while usage/status links stay available', () => {
    expect(componentSource).toContain("label: t('nav.accountCenter')")
    expect(componentSource).toContain("t('nav.myAccount')")
    expect(componentSource).not.toContain("label: t('nav.usageAndStatus')")
    expect(componentSource).toContain("const welfareItem: NavItem = { path: '/welfare'")
    expect(componentSource).toContain('const primarySelfItems')
    const primaryStart = componentSource.indexOf('const primarySelfItems')
    const accountStart = componentSource.indexOf('const accountItems')
    const usageStart = componentSource.indexOf('const usageStatusItems')
    const primaryEnd = componentSource.indexOf('if (authStore.isSimpleMode)', primaryStart)
    expect(primaryStart).toBeGreaterThan(-1)
    expect(accountStart).toBeGreaterThan(-1)
    expect(usageStart).toBeGreaterThan(-1)
    expect(primaryEnd).toBeGreaterThan(primaryStart)
    expect(componentSource.slice(primaryStart, primaryEnd)).toContain('welfareItem')
    expect(componentSource.slice(accountStart, usageStart)).not.toContain("path: '/welfare'")
    expect(componentSource).toContain('...usageStatusItems')
    expect(componentSource).toContain("path: '/keys'")
    expect(componentSource).toContain("path: '/usage'")
    expect(componentSource).toContain("path: '/purchase'")
    expect(componentSource).toContain("path: '/leaderboard'")
    expect(componentSource).toContain("path: '/available-channels'")
    expect(componentSource).toContain("path: '/monitor'")
    expect(componentSource).toContain("path: '/my-accounts'")
    expect(componentSource).toContain("path: '/orders'")
    expect(componentSource).toContain("path: '/profile'")
    expect(componentSource).toContain("path: '/subscriptions'")
  })

  it('keeps the API key tour target available inside personal groups', () => {
    expect(componentSource).toContain('function getSelfMenuItemTour')
    expect(componentSource).toContain("'[data-tour=\"sidebar-my-keys\"]'")
  })
})
