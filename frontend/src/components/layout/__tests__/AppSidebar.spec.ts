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
const onboardingStylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../styles/onboarding.css')
const onboardingStyleSource = readFileSync(onboardingStylePath, 'utf8')

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
    expect(componentSource).toContain("label: t('nav.channelManagement')")
    expect(componentSource).toContain("label: t('nav.businessOperations')")
    expect(componentSource).toContain("label: t('nav.contentRecords')")
    expect(componentSource).toContain("path: '/admin/users'")
    expect(componentSource).toContain("path: '/admin/accounts'")
    expect(componentSource).toContain("path: '/admin/shared-accounts'")
    expect(componentSource).toContain("path: '/admin/proxies'")
    expect(componentSource).toContain("path: '/admin/redeem'")
  })

  it('keeps admin usage records as a top-level sidebar entry', () => {
    const businessStart = componentSource.indexOf("label: t('nav.businessOperations')")
    const usageStart = componentSource.indexOf("path: '/admin/usage'", businessStart)
    const contentStart = componentSource.indexOf("label: t('nav.contentRecords')", usageStart)
    const contentChildrenStart = componentSource.indexOf('children: [', contentStart)
    const contentChildrenEnd = componentSource.indexOf('],', contentChildrenStart)

    expect(businessStart).toBeGreaterThan(-1)
    expect(usageStart).toBeGreaterThan(-1)
    expect(contentStart).toBeGreaterThan(-1)
    expect(contentChildrenStart).toBeGreaterThan(-1)
    expect(contentChildrenEnd).toBeGreaterThan(contentChildrenStart)
    expect(usageStart).toBeLessThan(contentStart)
    expect(componentSource.slice(contentChildrenStart, contentChildrenEnd)).not.toContain("path: '/admin/usage'")
  })

  it('places account and proxy management under basic management while channels stay standalone', () => {
    const basicStart = componentSource.indexOf("label: t('nav.basicManagement')")
    const channelStart = componentSource.indexOf("label: t('nav.channelManagement')", basicStart)
    const resourceStart = componentSource.indexOf("label: t('nav.resourceAccess')", channelStart)
    const basicBlock = componentSource.slice(basicStart, channelStart)
    const channelBlock = componentSource.slice(channelStart, resourceStart)

    expect(basicStart).toBeGreaterThan(-1)
    expect(channelStart).toBeGreaterThan(-1)
    expect(resourceStart).toBeGreaterThan(-1)
    expect(basicBlock).toContain("path: '/admin/accounts'")
    expect(basicBlock).toContain("path: '/admin/shared-accounts'")
    expect(basicBlock).toContain("path: '/admin/proxies'")
    expect(channelBlock).toContain("path: '/admin/channels/pricing'")
    expect(channelBlock).toContain("path: '/admin/channels/monitor'")
    expect(channelBlock).not.toContain("path: '/admin/accounts'")
  })

  it('keeps subscription plans beside subscription management instead of inside orders', () => {
    const businessStart = componentSource.indexOf("label: t('nav.businessOperations')")
    const businessEnd = componentSource.indexOf("path: '/admin/usage'", businessStart)
    const ordersStart = componentSource.indexOf("path: '/admin/orders'", businessStart)
    const ordersTopLevelStart = componentSource.indexOf("path: '/admin/orders'")
    const ordersChildrenStart = componentSource.indexOf('children: [', ordersTopLevelStart)
    const ordersChildrenEnd = componentSource.indexOf('],', ordersChildrenStart)
    expect(businessStart).toBeGreaterThan(-1)
    expect(businessEnd).toBeGreaterThan(businessStart)
    expect(ordersStart).toBe(-1)
    expect(ordersTopLevelStart).toBeGreaterThan(-1)
    expect(ordersTopLevelStart).toBeLessThan(businessStart)
    expect(ordersChildrenStart).toBeGreaterThan(ordersTopLevelStart)
    expect(componentSource.slice(businessStart, businessEnd)).not.toContain("label: t('nav.orderManagement')")
    expect(componentSource.indexOf("path: '/admin/subscriptions'", businessStart)).toBeGreaterThan(businessStart)
    expect(componentSource.indexOf("path: '/admin/orders/plans'", businessStart)).toBeGreaterThan(businessStart)
    expect(componentSource.slice(ordersChildrenStart, ordersChildrenEnd)).not.toContain("path: '/admin/orders/plans'")
  })

  it('keeps the order list leaf exact so payment dashboard is not double highlighted', () => {
    expect(componentSource).toContain('exactActive?: boolean')
    expect(componentSource).toContain('function isNavItemActive')
    expect(componentSource).toContain("path: '/admin/orders', label: t('nav.orderList'), icon: OrderIcon, exactActive: true")
  })

  it('shows an animated admin ticket badge from admin unread tickets', () => {
    expect(componentSource).toContain("import { adminTicketsAPI } from '@/api/admin/tickets'")
    expect(componentSource).toContain('const adminTicketUnreadTotal = ref(0)')
    expect(componentSource).toContain('function showAdminTicketAttentionBadge')
    expect(componentSource).toContain("item.path === '/admin/tickets'")
    expect(componentSource).toContain('adminTicketsAPI.list({')
    expect(componentSource).toContain('unread_only: true')
    expect(componentSource).toContain('sidebar-ticket-attention-badge')
    expect(componentSource).toContain('@keyframes sidebar-ticket-attention-pulse')
  })

  it('preserves admin onboarding anchors after grouping', () => {
    expect(componentSource).toContain('function getAdminMenuItemId')
    expect(componentSource).toContain('sidebarTourTargetGroups')
    expect(componentSource).toContain("'sidebar-group-manage'")
    expect(componentSource).toContain("'sidebar-channel-manage'")
    expect(componentSource).toContain("'sidebar-wallet'")
    expect(onboardingTourSource).toContain("'sub2api:sidebar-tour-target'")
  })

  it('keeps driver onboarding from blocking normal user clicks', () => {
    expect(onboardingTourSource).toContain('普通用户使用 dashboard 里的非阻塞提示卡')
    expect(onboardingTourSource).toContain('destroyActiveTour()')
    expect(onboardingStyleSource).toContain('.driver-active *')
    expect(onboardingStyleSource).toContain('pointer-events: auto !important;')
  })
})

describe('AppSidebar chat image navigation', () => {
  it('uses a single chat entry and hides the image manager from the sidebar', () => {
    expect(componentSource).not.toContain("t('nav.creationCenter')")
    expect(componentSource).not.toContain("label: t('nav.chatCreator')")
    expect(componentSource).not.toContain("label: t('nav.imageCreator')")
    expect(componentSource).not.toContain("path: '/chat'")
    expect(componentSource).not.toContain("path: '/image-creator'")
    expect(componentSource).toContain("label: t('nav.chatImageCreator')")
    expect(componentSource).toContain("path: '/chat-images'")
    expect(componentSource).not.toContain("label: t('nav.imageManager')")
    expect(componentSource).not.toContain("path: '/image-manager'")
  })

  it('keeps creation links aligned with regular sidebar links in the console shell', () => {
    const sidebarLinkBlock = consoleUiSource.match(/\.console-sidebar \.sidebar-link\s*\{[\s\S]*?\n\}/)?.[0] ?? ''
    const sidebarSectionBlock = consoleUiSource.match(/\.console-sidebar \.sidebar-section\s*\{[\s\S]*?\n\}/)?.[0] ?? ''

    expect(sidebarLinkBlock).toContain('width: 100%;')
    expect(sidebarLinkBlock).toContain('box-sizing: border-box;')
    expect(sidebarSectionBlock).toContain('width: 100%;')
    expect(componentSource).toContain("path: '/chat-images'")
    expect(componentSource).toContain("path: '/dashboard'")
  })

  it('keeps chat image launch from also navigating the current page', () => {
    expect(componentSource).toContain("action: 'studioBridgeLaunch'")
    expect(componentSource).toContain('function navItemElement')
    expect(componentSource).toContain("return item.action ? 'button' : 'router-link'")
    expect(componentSource).toContain("if (item.action) return { type: 'button' }")
    expect(componentSource).toContain('v-bind="navItemAttrs(item)"')
  })
})

describe('AppSidebar self navigation groups', () => {
  it('keeps account entries ungrouped at the bottom while usage/status links stay available', () => {
    expect(componentSource).not.toContain("label: t('nav.accountCenter')")
    expect(componentSource).not.toContain("path: '/self/account-center'")
    expect(componentSource).toContain("t('nav.myAccount')")
    expect(componentSource).not.toContain("label: t('nav.usageAndStatus')")
    expect(componentSource).toContain("const welfareItem: NavItem = { path: '/welfare'")
    expect(componentSource).toContain('showWelfareClaimBadge(item)')
    expect(componentSource).toContain('showTicketUnreadBadge(item)')
    expect(componentSource).toContain('showInvoiceClaimBadge(item)')
    expect(componentSource).toContain("t('nav.claimQuota')")
    expect(componentSource).toContain("t('nav.claimInvoice')")
    expect(componentSource).toContain('ticketsAPI.unreadSummary()')
    expect(componentSource).toContain('paymentAPI.getInvoiceClaimSummary()')
    expect(componentSource).toContain('const invoiceClaimTotal = ref(0)')
    expect(componentSource).toContain('function refreshInvoiceClaimBadge')
    expect(componentSource).toContain('.sidebar-unread-badge')
    expect(componentSource).toContain('.sidebar-invoice-claim-badge')
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
    expect(componentSource.slice(primaryStart, primaryEnd)).toContain("path: '/affiliate'")
    const accountItemsBlock = componentSource.slice(accountStart, usageStart)
    expect(accountItemsBlock).not.toContain("path: '/welfare'")
    expect(accountItemsBlock).not.toContain("path: '/affiliate'")
    expect(componentSource).toContain('...usageStatusItems')
    expect(componentSource).toContain('...accountItems')
    expect(componentSource).toMatch(/items\.push\(\s*\.\.\.primarySelfItems,\s*\.\.\.usageStatusItems,\s*\.\.\.accountItems,\s*\)/)
    expect(componentSource).toContain("path: '/keys'")
    expect(componentSource).toContain("path: '/chat-images'")
    expect(componentSource).toContain("path: '/usage'")
    expect(componentSource).toContain("path: '/tickets'")
    expect(componentSource).toContain("path: '/purchase'")
    expect(componentSource).toContain("path: '/leaderboard'")
    expect(componentSource).toContain("path: '/available-channels'")
    expect(componentSource).toContain("path: '/monitor'")
    expect(componentSource).toContain("path: '/my-accounts'")
    expect(accountItemsBlock).not.toContain("path: '/orders'")
    expect(componentSource).toContain("path: '/profile'")
    expect(componentSource).toContain("path: '/usage'")
    expect(componentSource).toContain("label: t('nav.usageAndSubscriptions')")
    expect(accountItemsBlock).not.toContain("path: '/redeem'")
  })

  it('keeps the API key tour target available inside personal groups', () => {
    expect(componentSource).toContain('function getSelfMenuItemTour')
    expect(componentSource).toContain("'[data-tour=\"sidebar-my-keys\"]'")
  })
})
