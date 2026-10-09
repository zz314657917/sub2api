import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it, vi } from 'vitest'

const appState = vi.hoisted(() => ({ cachedPublicSettings: undefined as Record<string, unknown> | undefined }))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appState,
}))

import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'

const source = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf8')

describe('service store enabled setting', () => {
  it('treats false as disabled while missing and loading public settings stay enabled', () => {
    appState.cachedPublicSettings = { service_store_enabled: false }
    expect(isFeatureFlagEnabled(FeatureFlags.serviceStore)).toBe(false)

    appState.cachedPublicSettings = {}
    expect(isFeatureFlagEnabled(FeatureFlags.serviceStore)).toBe(true)

    appState.cachedPublicSettings = undefined
    expect(isFeatureFlagEnabled(FeatureFlags.serviceStore)).toBe(true)
  })

  it('keeps false through the admin form and refreshes public settings after saving', () => {
    const view = source('src/views/admin/SettingsView.vue')
    const api = source('src/api/admin/settings.ts')

    expect(view).toContain('service_store_enabled: true,')
    expect(view).toContain('form.service_store_enabled = settings.service_store_enabled ?? true;')
    expect(view).toContain('service_store_enabled: form.service_store_enabled,')
    expect(view).toContain('data-testid="service-store-enabled"')
    expect(view).toContain('await appStore.fetchPublicSettings(true);')
    expect(api).toContain('service_store_enabled?: boolean;')
  })

  it('uses the same opt-out flag for the sidebar and authenticated route guard', () => {
    const router = source('src/router/index.ts')
    const sidebar = source('src/components/layout/AppSidebar.vue')

    expect(router).toMatch(/path: '\/store',[\s\S]*?requiresAuth: true,[\s\S]*?requiresServiceStore: true/)
    expect(router).toContain('to.meta.requiresServiceStore && !isFeatureFlagEnabled(FeatureFlags.serviceStore)')
    expect(router).toContain("next(authStore.isAdmin ? '/admin/dashboard' : '/dashboard')")
    expect(sidebar).toContain('const flagServiceStore = makeSidebarFlag(FeatureFlags.serviceStore)')
    expect(sidebar).toContain("path: '/store', label: t('nav.serviceStore'), icon: PriceTagIcon, featureFlag: flagServiceStore")
  })
})
