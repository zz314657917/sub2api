import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const source = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf8')

describe('store console integration', () => {
  it('keeps the store authenticated in the user console with one sidebar entry and both labels', () => {
    const router = source('src/router/index.ts')
    const sidebar = source('src/components/layout/AppSidebar.vue')
    const view = source('src/views/user/ServiceStoreView.vue')
    const zhNav = source('src/i18n/locales/zh/nav.ts')
    const enNav = source('src/i18n/locales/en/nav.ts')
    const homeView = source('src/views/HomeView.vue')
    const publicTopNav = source('src/views/public/components/PublicTopNav.vue')
    const storeRoute = router.slice(router.indexOf("path: '/store'"), router.indexOf("path: '/keys'"))
    const backendModeAllowedPaths = router.match(/const BACKEND_MODE_ALLOWED_PATHS = \[([\s\S]*?)\]/)?.[1] ?? ''

    expect(storeRoute).toMatch(/name: 'ServiceStore',[\s\S]*?views\/user\/ServiceStoreView\.vue[\s\S]*?requiresAuth: true/)
    expect(storeRoute).not.toContain('requiresAuth: false')
    expect(view).toContain('import AppLayout from "@/components/layout/AppLayout.vue"')
    expect(view).toMatch(/<AppLayout[\s\S]*service-store-page[\s\S]*<\/AppLayout/)
    expect((sidebar.match(/path: '\/store'/g) ?? [])).toHaveLength(1)
    expect(sidebar).toContain("label: t('nav.serviceStore')")
    expect(zhNav).toContain("serviceStore: '服务商店'")
    expect(enNav).toContain("serviceStore: 'Service Store'")
    const publicRoutes = router.slice(0, router.indexOf('// ==================== User Routes'))
    expect(publicRoutes).not.toContain("path: '/store'")
    expect(homeView).not.toMatch(/(?:to|:to)=["']\/store["']/)
    expect(publicTopNav).not.toMatch(/(?:to|:to)=["']\/store["']/)
    expect(backendModeAllowedPaths).not.toContain("'/store'")
  })
})
