import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const consoleUi = readFileSync(resolve(process.cwd(), 'src/styles/console-ui.css'), 'utf8')
const appSidebar = readFileSync(resolve(process.cwd(), 'src/components/layout/AppSidebar.vue'), 'utf8')
const appHeader = readFileSync(resolve(process.cwd(), 'src/components/layout/AppHeader.vue'), 'utf8')
const mainEntry = readFileSync(resolve(process.cwd(), 'src/main.ts'), 'utf8')
const homeView = readFileSync(resolve(process.cwd(), 'src/views/HomeView.vue'), 'utf8')
const keyUsageView = readFileSync(resolve(process.cwd(), 'src/views/KeyUsageView.vue'), 'utf8')
const tailwindConfig = readFileSync(resolve(process.cwd(), 'tailwind.config.js'), 'utf8')
const adminDashboard = readFileSync(resolve(process.cwd(), 'src/views/admin/DashboardView.vue'), 'utf8')
const userStats = readFileSync(resolve(process.cwd(), 'src/components/user/dashboard/UserDashboardStats.vue'), 'utf8')
const quickActions = readFileSync(resolve(process.cwd(), 'src/components/user/dashboard/UserDashboardQuickActions.vue'), 'utf8')
const recentUsage = readFileSync(resolve(process.cwd(), 'src/components/user/dashboard/UserDashboardRecentUsage.vue'), 'utf8')

const consoleComponents = [
  appSidebar,
  appHeader,
  adminDashboard,
  userStats,
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

  it('defines distinct light and dark console surfaces', () => {
    expect(consoleUi).toContain('.console-shell {')
    expect(consoleUi).toContain('linear-gradient(180deg, #f8fafc')
    expect(consoleUi).toContain('.dark .console-shell {')
    expect(consoleUi).toContain('linear-gradient(180deg, #020617')
    expect(consoleUi).toContain('.dark .console-sidebar')
    expect(consoleUi).toContain('.dark .console-header')
  })

  it('defaults new visitors to dark mode unless they explicitly chose light', () => {
    expect(mainEntry).toContain("const shouldUseDark = savedTheme !== 'light'")
    expect(appSidebar).toContain("if (savedTheme !== 'light')")
    expect(homeView).toContain("if (savedTheme !== 'light')")
    expect(keyUsageView).toContain("if (savedTheme !== 'light')")
    expect(homeView).not.toContain("prefers-color-scheme: dark")
    expect(keyUsageView).not.toContain("prefers-color-scheme: dark")
  })

  it('keeps console dark palette on slate surfaces instead of the old warm browns', () => {
    expect(tailwindConfig).toContain("800: '#1e293b'")
    expect(tailwindConfig).toContain("900: '#0f172a'")
    expect(tailwindConfig).toContain("950: '#020617'")
    expect(consoleUi).toContain('background: rgba(8, 15, 29, 0.82)')
    expect(consoleUi).toContain('background: rgba(8, 15, 29, 0.84)')
    expect(consoleUi).toContain('background: rgba(10, 18, 32, 0.86)')
    expect(consoleUi).toContain('.dark .console-shell .select-trigger')
    expect(consoleUi).toContain('.dark .select-dropdown-portal')
    expect(consoleUi).not.toContain('#33251b')
    expect(consoleUi).not.toContain('#463225')
    expect(consoleUi).not.toContain('#1d130d')
    expect(consoleUi).not.toContain('rgba(139, 111, 71')
    expect(consoleUi).not.toContain('rgba(91, 78, 60')
  })
})
