import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const consoleUi = readFileSync(resolve(process.cwd(), 'src/styles/console-ui.css'), 'utf8')
const appSidebar = readFileSync(resolve(process.cwd(), 'src/components/layout/AppSidebar.vue'), 'utf8')
const appHeader = readFileSync(resolve(process.cwd(), 'src/components/layout/AppHeader.vue'), 'utf8')
const headerAnnouncementCarousel = readFileSync(resolve(process.cwd(), 'src/components/layout/HeaderAnnouncementCarousel.vue'), 'utf8')
const announcementBell = readFileSync(resolve(process.cwd(), 'src/components/common/AnnouncementBell.vue'), 'utf8')
const mainEntry = readFileSync(resolve(process.cwd(), 'src/main.ts'), 'utf8')
const homeView = readFileSync(resolve(process.cwd(), 'src/views/HomeView.vue'), 'utf8')
const keyUsageView = readFileSync(resolve(process.cwd(), 'src/views/KeyUsageView.vue'), 'utf8')
const tailwindConfig = readFileSync(resolve(process.cwd(), 'tailwind.config.js'), 'utf8')
const adminDashboard = readFileSync(resolve(process.cwd(), 'src/views/admin/DashboardView.vue'), 'utf8')
const modelDistributionChart = readFileSync(resolve(process.cwd(), 'src/components/charts/ModelDistributionChart.vue'), 'utf8')
const tokenUsageTrend = readFileSync(resolve(process.cwd(), 'src/components/charts/TokenUsageTrend.vue'), 'utf8')
const userStats = readFileSync(resolve(process.cwd(), 'src/components/user/dashboard/UserDashboardStats.vue'), 'utf8')
const userPerformanceStats = readFileSync(resolve(process.cwd(), 'src/components/user/dashboard/UserDashboardPerformanceStats.vue'), 'utf8')
const quickActions = readFileSync(resolve(process.cwd(), 'src/components/user/dashboard/UserDashboardQuickActions.vue'), 'utf8')
const recentUsage = readFileSync(resolve(process.cwd(), 'src/components/user/dashboard/UserDashboardRecentUsage.vue'), 'utf8')
const onboardingCss = readFileSync(resolve(process.cwd(), 'src/styles/onboarding.css'), 'utf8')
const adminAccountTestModal = readFileSync(resolve(process.cwd(), 'src/components/admin/account/AccountTestModal.vue'), 'utf8')
const userAccountTestModal = readFileSync(resolve(process.cwd(), 'src/components/account/AccountTestModal.vue'), 'utf8')
const adminUsageTable = readFileSync(resolve(process.cwd(), 'src/components/admin/usage/UsageTable.vue'), 'utf8')
const userUsageView = readFileSync(resolve(process.cwd(), 'src/views/user/UsageView.vue'), 'utf8')
const adminRefundDialog = readFileSync(resolve(process.cwd(), 'src/components/admin/payment/AdminRefundDialog.vue'), 'utf8')
const userOrdersView = readFileSync(resolve(process.cwd(), 'src/views/user/UserOrdersView.vue'), 'utf8')
const subscriptionProgressMini = readFileSync(resolve(process.cwd(), 'src/components/common/SubscriptionProgressMini.vue'), 'utf8')
const paymentView = readFileSync(resolve(process.cwd(), 'src/views/user/PaymentView.vue'), 'utf8')
const adminOrdersView = readFileSync(resolve(process.cwd(), 'src/views/admin/orders/AdminOrdersView.vue'), 'utf8')
const adminInvoiceRequestsView = readFileSync(resolve(process.cwd(), 'src/views/admin/orders/AdminInvoiceRequestsView.vue'), 'utf8')
const adminPaymentPlansView = readFileSync(resolve(process.cwd(), 'src/views/admin/orders/AdminPaymentPlansView.vue'), 'utf8')
const adminSubscriptionsView = readFileSync(resolve(process.cwd(), 'src/views/admin/SubscriptionsView.vue'), 'utf8')
const orderStatsCards = readFileSync(resolve(process.cwd(), 'src/components/admin/payment/OrderStatsCards.vue'), 'utf8')
const dailyRevenueChart = readFileSync(resolve(process.cwd(), 'src/components/admin/payment/DailyRevenueChart.vue'), 'utf8')
const baseDialog = readFileSync(resolve(process.cwd(), 'src/components/common/BaseDialog.vue'), 'utf8')
const accountStatusIndicator = readFileSync(resolve(process.cwd(), 'src/components/account/AccountStatusIndicator.vue'), 'utf8')
const accountUsageCell = readFileSync(resolve(process.cwd(), 'src/components/account/AccountUsageCell.vue'), 'utf8')
const createAccountModal = readFileSync(resolve(process.cwd(), 'src/components/account/CreateAccountModal.vue'), 'utf8')
const editAccountModal = readFileSync(resolve(process.cwd(), 'src/components/account/EditAccountModal.vue'), 'utf8')
const bulkEditAccountModal = readFileSync(resolve(process.cwd(), 'src/components/account/BulkEditAccountModal.vue'), 'utf8')
const accountStatsModal = readFileSync(resolve(process.cwd(), 'src/components/account/AccountStatsModal.vue'), 'utf8')
const adminAccountStatsModal = readFileSync(resolve(process.cwd(), 'src/components/admin/account/AccountStatsModal.vue'), 'utf8')
const accountReAuthModal = readFileSync(resolve(process.cwd(), 'src/components/account/ReAuthAccountModal.vue'), 'utf8')
const adminAccountReAuthModal = readFileSync(resolve(process.cwd(), 'src/components/admin/account/ReAuthAccountModal.vue'), 'utf8')
const userAdminView = readFileSync(resolve(process.cwd(), 'src/views/admin/UsersView.vue'), 'utf8')
const groupsView = readFileSync(resolve(process.cwd(), 'src/views/admin/GroupsView.vue'), 'utf8')
const riskControlView = readFileSync(resolve(process.cwd(), 'src/views/admin/RiskControlView.vue'), 'utf8')
const opsErrorLogTable = readFileSync(resolve(process.cwd(), 'src/views/admin/ops/components/OpsErrorLogTable.vue'), 'utf8')
const opsErrorDetailModal = readFileSync(resolve(process.cwd(), 'src/views/admin/ops/components/OpsErrorDetailModal.vue'), 'utf8')
const useKeyModal = readFileSync(resolve(process.cwd(), 'src/components/keys/UseKeyModal.vue'), 'utf8')
const opsLatencyChart = readFileSync(resolve(process.cwd(), 'src/views/admin/ops/components/OpsLatencyChart.vue'), 'utf8')
const opsErrorDistributionChart = readFileSync(resolve(process.cwd(), 'src/views/admin/ops/components/OpsErrorDistributionChart.vue'), 'utf8')
const opsSwitchRateTrendChart = readFileSync(resolve(process.cwd(), 'src/views/admin/ops/components/OpsSwitchRateTrendChart.vue'), 'utf8')
const opsThroughputTrendChart = readFileSync(resolve(process.cwd(), 'src/views/admin/ops/components/OpsThroughputTrendChart.vue'), 'utf8')
const opsSettingsDialog = readFileSync(resolve(process.cwd(), 'src/views/admin/ops/components/OpsSettingsDialog.vue'), 'utf8')
const opsEmailNotificationCard = readFileSync(resolve(process.cwd(), 'src/views/admin/ops/components/OpsEmailNotificationCard.vue'), 'utf8')
const adminTicketsView = readFileSync(resolve(process.cwd(), 'src/views/admin/TicketsView.vue'), 'utf8')
const userTicketsView = readFileSync(resolve(process.cwd(), 'src/views/user/TicketsView.vue'), 'utf8')
const accountCapacityPools = readFileSync(resolve(process.cwd(), 'src/components/user/monitor/AccountCapacityPools.vue'), 'utf8')
const monitorAvailabilityList = readFileSync(resolve(process.cwd(), 'src/components/user/monitor/MonitorAvailabilityList.vue'), 'utf8')
const channelMonitorFormat = readFileSync(resolve(process.cwd(), 'src/composables/useChannelMonitorFormat.ts'), 'utf8')
const groupBadge = readFileSync(resolve(process.cwd(), 'src/components/common/GroupBadge.vue'), 'utf8')
const platformTypeBadge = readFileSync(resolve(process.cwd(), 'src/components/common/PlatformTypeBadge.vue'), 'utf8')
const groupDistributionChart = readFileSync(resolve(process.cwd(), 'src/components/charts/GroupDistributionChart.vue'), 'utf8')
const endpointDistributionChart = readFileSync(resolve(process.cwd(), 'src/components/charts/EndpointDistributionChart.vue'), 'utf8')
const userKeysView = readFileSync(resolve(process.cwd(), 'src/views/user/KeysView.vue'), 'utf8')
const accountQuotaInfo = readFileSync(resolve(process.cwd(), 'src/components/account/AccountQuotaInfo.vue'), 'utf8')
const userBalanceModal = readFileSync(resolve(process.cwd(), 'src/components/admin/user/UserBalanceModal.vue'), 'utf8')
const userBalanceHistoryModal = readFileSync(resolve(process.cwd(), 'src/components/admin/user/UserBalanceHistoryModal.vue'), 'utf8')
const versionBadge = readFileSync(resolve(process.cwd(), 'src/components/common/VersionBadge.vue'), 'utf8')
const backupView = readFileSync(resolve(process.cwd(), 'src/views/admin/BackupView.vue'), 'utf8')
const dataTable = readFileSync(resolve(process.cwd(), 'src/components/common/DataTable.vue'), 'utf8')

const consoleComponents = [
  appSidebar,
  appHeader,
  adminDashboard,
  userStats,
  userPerformanceStats,
  quickActions,
  recentUsage,
]

describe('console visual direction', () => {
  it('uses linear icons instead of homepage pixel icons in console chrome and dashboards', () => {
    for (const source of consoleComponents) {
      expect(source).toContain("from '@/components/icons/Icon.vue'")
      expect(source).not.toContain("from '@/components/icons/PixelIcon.vue'")
      expect(source).not.toContain('<PixelIcon')
    }
  })

  it('keeps console icon frames transparent and line-icon friendly', () => {
    expect(consoleUi).toContain('.console-nav-icon-frame {')
    expect(consoleUi).toContain('background: transparent')
    expect(consoleUi).toContain('.console-nav-icon-frame svg')
    expect(consoleUi).not.toContain('pixel-glyph')
  })

  it('defines warm light console surfaces and keeps slate dark mode available', () => {
    expect(consoleUi).toContain('.console-shell {')
    expect(consoleUi).toContain('--console-bg: #faf9f5')
    expect(consoleUi).toContain('--console-accent: #cc785c')
    expect(consoleUi).toContain('--console-text: #141413')
    expect(consoleUi).toContain('linear-gradient(180deg, #faf9f5 0%, #f5f0e8 52%, #faf9f5 100%)')
    expect(consoleUi).toContain('.console-shell .btn-primary')
    expect(consoleUi).not.toContain('.console-shell .btn-primary,\n.console-shell .btn-success')
    expect(consoleUi).toContain('.dark .console-shell {')
    expect(consoleUi).toContain('linear-gradient(180deg, #020617')
    expect(consoleUi).toContain('.dark .console-sidebar')
    expect(consoleUi).toContain('.dark .console-header')
  })

  it('keeps console data tables on warm surfaces instead of cold gray sticky panels', () => {
    expect(consoleUi).toContain('--table-header-bg: rgba(250, 249, 245, 0.97)')
    expect(consoleUi).toContain('--table-sticky-bg: rgba(255, 252, 246, 0.98)')
    expect(consoleUi).toContain('border-color: rgba(216, 206, 194, 0.5)')
    expect(consoleUi).toContain('color: #504f49')
    expect(dataTable).toContain('--table-header-bg: rgba(250, 249, 245, 0.96)')
    expect(dataTable).toContain('--table-sticky-bg: rgba(255, 252, 246, 0.98)')
    expect(dataTable).toContain('background-color: rgba(216, 206, 194, 0.22)')
    expect(dataTable).toContain('background-color: rgba(160, 153, 144, 0.78)')
    expect(dataTable).not.toContain('background-color: rgb(249 250 251)')
    expect(dataTable).not.toContain('background-color: white;')
  })

  it('keeps decorative console backdrop layers behind content', () => {
    expect(consoleUi).toContain('isolation: isolate')
    expect(consoleUi).toContain('z-index: -1')
    expect(consoleUi).toContain('z-index: -2')
    expect(consoleUi).toContain('z-index: 0')
  })

  it('defaults new visitors to light mode unless they explicitly chose dark', () => {
    expect(mainEntry).toContain("const shouldUseDark = savedTheme === 'dark'")
    expect(appSidebar).toContain("if (savedTheme === 'dark')")
    expect(homeView).toContain("if (savedTheme === 'dark')")
    expect(keyUsageView).toContain("if (savedTheme === 'dark')")
    expect(homeView).not.toContain("prefers-color-scheme: dark")
    expect(keyUsageView).not.toContain("prefers-color-scheme: dark")
  })

  it('keeps console dark palette on slate surfaces without overriding semantic success color', () => {
    expect(tailwindConfig).toContain("800: '#1e293b'")
    expect(tailwindConfig).toContain("900: '#0f172a'")
    expect(tailwindConfig).toContain("950: '#020617'")
    expect(consoleUi).toContain('background: rgba(8, 15, 29, 0.82)')
    expect(consoleUi).toContain('background: rgba(8, 15, 29, 0.84)')
    expect(consoleUi).toContain('background: rgba(10, 18, 32, 0.86)')
    expect(consoleUi).toContain('.dark .console-shell .select-trigger')
    expect(consoleUi).toContain('.dark .select-dropdown-portal')
    expect(consoleUi).toContain('--dashboard-success: #a9583e')
    expect(consoleUi).toContain('--dashboard-success: #f0b89e')
    expect(consoleUi).not.toContain('.dark .console-shell .btn-primary,\n.dark .console-shell .btn-success')
    expect(consoleUi).not.toContain('#33251b')
    expect(consoleUi).not.toContain('#463225')
    expect(consoleUi).not.toContain('#1d130d')
    expect(consoleUi).not.toContain('rgba(139, 111, 71')
    expect(consoleUi).not.toContain('rgba(91, 78, 60')
  })

  it('maps legacy purple and indigo utility accents onto the warm console palette', () => {
    expect(consoleUi).toContain('.console-shell .bg-purple-50')
    expect(consoleUi).toContain('.console-shell .text-indigo-600')
    expect(consoleUi).toContain('background-color: rgba(204, 120, 92, 0.1)')
    expect(consoleUi).toContain('color: #a9583e')
  })

  it('uses warm onboarding tour chrome instead of the old teal action color', () => {
    expect(onboardingCss).toContain('background-color: #141413')
    expect(onboardingCss).toContain('outline: 2px solid rgba(204, 120, 92, 0.34)')
    expect(onboardingCss).toContain('border-left-color: #cc785c')
    expect(onboardingCss).not.toContain('background-color: #14b8a6')
  })

  it('keeps account test terminals in the warm light command surface', () => {
    for (const source of [adminAccountTestModal, userAccountTestModal]) {
      expect(source).toContain('bg-[#f5f0e8]')
      expect(source).toContain('border border-[#d8cec2]')
      expect(source).toContain("text-[#a9583e] dark:text-[#f0b89e]")
      expect(source).not.toContain('border border-gray-700 bg-gray-900')
    }
  })

  it('uses warm light usage table tooltips across admin and user usage views', () => {
    for (const source of [adminUsageTable, userUsageView]) {
      expect(source).toContain('bg-[#fffaf5]')
      expect(source).toContain('text-[#141413]')
      expect(source).toContain('border-r-[#fffaf5]')
      expect(source).not.toContain('border border-gray-700 bg-gray-900')
    }
  })

  it('keeps refund and subscription controls out of the old violet SaaS accent', () => {
    expect(adminRefundDialog).toContain('border border-[#d8cec2] bg-[#fffaf5]')
    expect(adminRefundDialog).not.toContain('bg-violet-50')
    expect(userOrdersView).toContain('text-[#a9583e] hover:bg-[#f3e7df]')
    expect(userOrdersView).toContain('bg-[#141413] text-white')
    expect(subscriptionProgressMini).toContain('border border-[#d8cec2] bg-[#fffaf5]')
    expect(subscriptionProgressMini).not.toContain('bg-purple-50 px-3')
  })

  it('applies warm modal and tooltip primitives to deep console surfaces', () => {
    expect(baseDialog).toContain('console-modal-content')
    expect(consoleUi).toContain('.console-tooltip-surface')
    expect(consoleUi).toContain('background: #fffaf5')
    expect(consoleUi).toContain('.console-note-accent')
    expect(consoleUi).toContain('.console-badge-accent')
    expect(consoleUi).toContain("input[type='checkbox']")
    expect(consoleUi).toContain('accent-color: #cc785c')
  })

  it('removes old black tooltip surfaces from account and admin user management', () => {
    for (const source of [accountStatusIndicator, accountUsageCell, createAccountModal, editAccountModal, userAdminView]) {
      expect(source).toContain('console-tooltip-surface')
      expect(source).not.toContain('rounded bg-gray-900 px')
      expect(source).not.toContain('bg-gray-900 px-3 py-2 text-xs text-white')
    }
  })

  it('keeps account modals and group badges on the warm accent instead of purple SaaS accents', () => {
    for (const source of [createAccountModal, editAccountModal, bulkEditAccountModal, groupsView]) {
      expect(source).toContain('console-badge-accent')
      expect(source).not.toContain('bg-purple-100 text-purple-700')
      expect(source).not.toContain('bg-purple-50 p-3')
    }

    for (const source of [accountStatsModal, adminAccountStatsModal, accountReAuthModal, adminAccountReAuthModal]) {
      expect(source).toContain('#cc785c')
      expect(source).not.toContain('from-emerald-50')
      expect(source).not.toContain('from-blue-50')
      expect(source).not.toContain('from-green-500')
      expect(source).not.toContain('from-blue-500')
    }
  })

  it('warms risk and ops routing accents without changing semantic danger states', () => {
    expect(riskControlView).toContain("bg-[#f3e7df] text-[#a9583e]")
    for (const source of [opsErrorLogTable, opsErrorDetailModal]) {
      expect(source).toContain("ring-[#cc785c]/25")
      expect(source).not.toContain('bg-purple-50 text-purple-700')
    }
  })

  it('keeps purchase and payment order surfaces on warm light accents', () => {
    expect(paymentView).toContain('linear-gradient(180deg, rgba(250, 249, 245, 0.98)')
    expect(paymentView).toContain('rgba(204, 120, 92, 0.12)')
    expect(paymentView).toContain("bg-[#141413] text-white shadow-lg shadow-black/10")
    expect(paymentView).not.toContain('linear-gradient(180deg, rgba(248, 250, 252, 0.96)')
    expect(paymentView).not.toContain('bg-indigo-500 px-4')
    expect(paymentView).not.toContain('shadow-blue-500/20')

    for (const source of [adminOrdersView, adminInvoiceRequestsView, adminPaymentPlansView, adminSubscriptionsView]) {
      expect(source).toContain('hover:bg-[#f3e7df]')
      expect(source).toContain('text-[#a9583e]')
      expect(source).not.toContain('hover:bg-purple-50')
      expect(source).not.toContain('text-purple-600')
      expect(source).not.toContain('bg-purple-100')
    }

    expect(orderStatsCards).toContain('bg-[#f3e7df]')
    expect(orderStatsCards).not.toContain('bg-purple-100')
    expect(dailyRevenueChart).toContain("borderColor: '#cc785c'")
    expect(dailyRevenueChart).not.toContain("borderColor: 'rgb(59, 130, 246)'")
  })

  it('keeps admin dashboard first-screen charts on warm muted accents', () => {
    for (const source of [adminDashboard, modelDistributionChart, tokenUsageTrend]) {
      expect(source).toContain('#cc785c')
      expect(source).toContain('rgba(204, 120, 92')
      expect(source).not.toContain('#3b82f6')
      expect(source).not.toContain('#06b6d4')
      expect(source).not.toContain('#8b5cf6')
      expect(source).not.toContain('rgba(59, 130, 246')
      expect(source).not.toContain('text-violet-600')
      expect(source).not.toContain('text-blue-600')
    }

    expect(adminDashboard).toContain("text-[#a9583e] dark:text-[#f0b89e]")
    expect(modelDistributionChart).toContain("hover:bg-[#f8efe8]")
    expect(tokenUsageTrend).toContain("cacheHitRate: '#6f6a5f'")
  })

  it('warms header announcement, balance, and claim badge accents', () => {
    expect(headerAnnouncementCarousel).toContain('color: #a9583e')
    expect(announcementBell).toContain("text-[#a9583e] dark:text-[#f0b89e]")
    expect(announcementBell).toContain('from-[#cc785c] to-[#a9583e]')
    expect(announcementBell).not.toContain('from-blue-500 to-indigo-600')
    expect(announcementBell).not.toContain('text-blue-600')

    expect(consoleUi).toContain('.console-balance')
    expect(consoleUi).toContain('background: rgba(255, 250, 245, 0.9)')
    expect(consoleUi).toContain('background: #cc785c')
    expect(appSidebar).toContain('background: #f3e7df')
    expect(appSidebar).toContain('color: #a9583e')
    expect(appSidebar).not.toContain('rgb(220 252 231)')
  })

  it('keeps screenshot-visible ops and ticket surfaces off old blue/teal accents', () => {
    for (const source of [
      useKeyModal,
      opsLatencyChart,
      opsErrorDistributionChart,
      opsSwitchRateTrendChart,
      opsThroughputTrendChart,
      opsSettingsDialog,
      opsEmailNotificationCard,
      adminTicketsView,
      userTicketsView,
      accountCapacityPools,
      monitorAvailabilityList,
      channelMonitorFormat,
    ]) {
      expect(source).not.toContain('#3b82f6')
      expect(source).not.toContain('#14b8a6')
      expect(source).not.toContain('rgb(37 99 235)')
    }

    expect(useKeyModal).toContain('dark:bg-[#cc785c]/10')
    expect(opsLatencyChart).toContain("blue: '#cc785c'")
    expect(opsErrorDistributionChart).toContain("blue: '#8e8b82'")
    expect(opsSwitchRateTrendChart).toContain("teal: '#cc785c'")
    expect(opsThroughputTrendChart).toContain("blue: '#cc785c'")
    expect(opsSettingsDialog).toContain('dark:bg-[#cc785c]/15')
    expect(opsEmailNotificationCard).toContain('dark:bg-[#cc785c]/15')
    expect(accountCapacityPools).toContain("barClass: 'bg-[#9c7b62]'")
    expect(monitorAvailabilityList).toContain("dot: 'bg-[#9c7b62]'")
    expect(channelMonitorFormat).toContain("return '#a9583e'")
    expect(channelMonitorFormat).toContain('from-[#fffaf5] to-[#f3e7df]')
    expect(channelMonitorFormat).not.toContain('bg-emerald-100 text-emerald-700')
    expect(channelMonitorFormat).not.toContain('from-sky-50 to-indigo-100')
    expect(adminTicketsView).toContain('background: rgb(20 20 19)')
    expect(userTicketsView).toContain('background: rgb(20 20 19)')
    expect(opsSwitchRateTrendChart).not.toContain('text-teal-500')
  })

  it('keeps shared account badges and distribution charts off old blue and emerald accents', () => {
    for (const source of [groupBadge, platformTypeBadge, accountQuotaInfo]) {
      expect(source).toContain('text-[#a9583e]')
      expect(source).not.toContain('bg-emerald-100 text-emerald-700')
      expect(source).not.toContain('bg-blue-100 text-blue-700')
      expect(source).not.toContain('bg-blue-100 text-blue-600')
      expect(source).not.toContain('bg-purple-100 text-purple-700')
    }

    for (const source of [groupDistributionChart, endpointDistributionChart]) {
      expect(source).toContain("'#cc785c'")
      expect(source).toContain("'#8e8b82'")
      expect(source).not.toContain("'#3b82f6'")
      expect(source).not.toContain("'#10b981'")
      expect(source).not.toContain("'#14b8a6'")
      expect(source).not.toContain('text-blue-600')
      expect(source).not.toContain('text-green-600')
    }
  })

  it('warms key, balance, version, and backup utility surfaces', () => {
    for (const source of [userKeysView, userBalanceModal, userBalanceHistoryModal, versionBadge, backupView]) {
      expect(source).toContain('#d8cec2')
      expect(source).not.toContain('dark:bg-blue-900/30')
      expect(source).not.toContain('dark:bg-blue-500/10')
      expect(source).not.toContain('bg-blue-50')
      expect(source).not.toContain('text-emerald-600')
    }
  })

  it('keeps the balance modal user identity readable in dark mode', () => {
    expect(userBalanceModal).toContain('text-gray-900 dark:text-gray-100')
    expect(userBalanceModal).toContain('text-gray-500 dark:text-gray-400')
  })
})
