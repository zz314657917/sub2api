import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import { useAppStore } from '@/stores/app'
import type { PublicSettings } from '@/types'

import DashboardView from '../DashboardView.vue'

const mocks = vi.hoisted(() => ({
  push: vi.fn(),
  refreshUser: vi.fn(),
  getDashboardStats: vi.fn(),
  getDashboardTrend: vi.fn(),
  getDashboardModels: vi.fn(),
  getByDateRange: vi.fn(),
  getAccountUsageSummary: vi.fn(),
  getWelfareOverview: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mocks.push }),
}))

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      locale: { value: 'en' },
      setLocaleMessage: vi.fn(),
      t: (key: string) => key,
    },
  }),
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      if (params?.count !== undefined && params?.total !== undefined) {
        return `${key}:${params.count}/${params.total}`
      }
      return key
    },
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: {
      id: 1,
      role: 'user',
      balance: 0,
    },
    refreshUser: mocks.refreshUser,
  }),
}))

vi.mock('@/api/usage', () => ({
  usageAPI: {
    getDashboardStats: mocks.getDashboardStats,
    getDashboardTrend: mocks.getDashboardTrend,
    getDashboardModels: mocks.getDashboardModels,
    getByDateRange: mocks.getByDateRange,
  },
}))

vi.mock('@/api/user', () => ({
  userAPI: {
    getAccountUsageSummary: mocks.getAccountUsageSummary,
  },
}))

vi.mock('@/api/welfare', () => ({
  welfareAPI: {
    getWelfareOverview: mocks.getWelfareOverview,
  },
}))

const stubComponent = (name: string) => defineComponent({
  name,
  template: `<div data-test-stub="${name}" />`,
})

const makeStats = (overrides: Record<string, unknown> = {}) => ({
  total_api_keys: 1,
  active_api_keys: 1,
  total_requests: 10,
  total_input_tokens: 100,
  total_output_tokens: 50,
  total_cache_creation_tokens: 0,
  total_cache_read_tokens: 25,
  total_tokens: 175,
  total_cost: 1,
  total_actual_cost: 0.8,
  today_requests: 2,
  today_input_tokens: 20,
  today_output_tokens: 10,
  today_cache_creation_tokens: 0,
  today_cache_read_tokens: 5,
  today_tokens: 35,
  today_cost: 0.2,
  today_actual_cost: 0.16,
  average_duration_ms: 250,
  rpm: 1,
  tpm: 100,
  by_platform: [],
  ...overrides,
})

describe('User DashboardView data loading', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
    useAppStore().cachedPublicSettings = { account_share_enabled: true } as PublicSettings
    window.localStorage.clear()
    mocks.refreshUser.mockResolvedValue(undefined)
    mocks.getDashboardStats.mockResolvedValue(makeStats())
    mocks.getDashboardTrend.mockResolvedValue({ trend: [] })
    mocks.getDashboardModels.mockResolvedValue({ models: [] })
    mocks.getByDateRange.mockResolvedValue({ items: [] })
    mocks.getAccountUsageSummary.mockResolvedValue(null)
    mocks.getWelfareOverview.mockResolvedValue({ enabled: false })
  })

  it.each([true, false, undefined])('loads account usage only when sharing is enabled (%s)', async (enabled) => {
    const appStore = useAppStore()
    appStore.cachedPublicSettings = enabled === undefined
      ? null
      : { account_share_enabled: enabled } as PublicSettings
    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: defineComponent({
            name: 'AppLayout',
            template: '<main><slot /></main>',
          }),
          LoadingSpinner: true,
          UserDashboardStats: stubComponent('UserDashboardStats'),
          UserDashboardPerformanceStats: stubComponent('UserDashboardPerformanceStats'),
          UserDashboardCharts: stubComponent('UserDashboardCharts'),
          UserDashboardPlatformBreakdown: stubComponent('UserDashboardPlatformBreakdown'),
          UserDashboardAccountUsage: stubComponent('UserDashboardAccountUsage'),
          UserDashboardRecentUsage: stubComponent('UserDashboardRecentUsage'),
          UserDashboardQuickActions: stubComponent('UserDashboardQuickActions'),
          UserApiKeyOnboardingDialog: stubComponent('UserApiKeyOnboardingDialog'),
        },
      },
    })

    await flushPromises()

    expect(mocks.refreshUser).toHaveBeenCalledTimes(1)
    expect(mocks.getDashboardStats).toHaveBeenCalledTimes(1)
    expect(mocks.getDashboardTrend).toHaveBeenCalledTimes(1)
    expect(mocks.getByDateRange).toHaveBeenCalledTimes(1)
    expect(mocks.getAccountUsageSummary).toHaveBeenCalledTimes(enabled === true ? 1 : 0)
    expect(wrapper.find('[data-test-stub="UserDashboardAccountUsage"]').exists()).toBe(enabled === true)
    expect(mocks.getWelfareOverview).toHaveBeenCalledTimes(1)
    expect(mocks.getDashboardModels).not.toHaveBeenCalled()

    // Public settings may arrive after the dashboard mounts or be refreshed later.
    appStore.cachedPublicSettings = { account_share_enabled: true } as PublicSettings
    await flushPromises()
    expect(wrapper.find('[data-test-stub="UserDashboardAccountUsage"]').exists()).toBe(true)
    expect(mocks.getAccountUsageSummary).toHaveBeenCalledTimes(1)

    appStore.cachedPublicSettings = { account_share_enabled: false } as PublicSettings
    await flushPromises()
    expect(wrapper.find('[data-test-stub="UserDashboardAccountUsage"]').exists()).toBe(false)
    wrapper.findComponent({ name: 'UserDashboardCharts' }).vm.$emit('dateRangeChange')
    await flushPromises()
    expect(mocks.getAccountUsageSummary).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })

  it('shows onboarding for unused default key and opens the key page', async () => {
    mocks.getDashboardStats.mockResolvedValue(makeStats({
      total_api_keys: 1,
      active_api_keys: 1,
      total_requests: 0,
    }))

    const onboardingStub = defineComponent({
      name: 'UserApiKeyOnboardingDialog',
      props: {
        show: Boolean,
        hasApiKey: Boolean,
      },
      emits: ['create'],
      template: `
        <button
          data-test="onboarding-primary"
          :data-show="String(show)"
          :data-has-api-key="String(hasApiKey)"
          @click="$emit('create')"
        >
          onboarding
        </button>
      `,
    })

    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: defineComponent({
            name: 'AppLayout',
            template: '<main><slot /></main>',
          }),
          LoadingSpinner: true,
          UserDashboardStats: stubComponent('UserDashboardStats'),
          UserDashboardPerformanceStats: stubComponent('UserDashboardPerformanceStats'),
          UserDashboardCharts: stubComponent('UserDashboardCharts'),
          UserDashboardPlatformBreakdown: stubComponent('UserDashboardPlatformBreakdown'),
          UserDashboardAccountUsage: stubComponent('UserDashboardAccountUsage'),
          UserDashboardRecentUsage: stubComponent('UserDashboardRecentUsage'),
          UserDashboardQuickActions: stubComponent('UserDashboardQuickActions'),
          UserApiKeyOnboardingDialog: onboardingStub,
        },
      },
    })

    await flushPromises()

    const onboarding = wrapper.get('[data-test="onboarding-primary"]')
    expect(onboarding.attributes('data-show')).toBe('true')
    expect(onboarding.attributes('data-has-api-key')).toBe('true')

    await onboarding.trigger('click')
    expect(mocks.push).toHaveBeenCalledWith({ path: '/keys' })
    expect(window.localStorage.getItem('sub2api:user-api-key-onboarding:skip:1')).toBeNull()
  })

  it('can reopen quick setup after a skip and after requests have started', async () => {
    window.localStorage.setItem('sub2api:user-api-key-onboarding:skip:1', String(Date.now()))
    const onboardingStub = defineComponent({
      props: { show: Boolean },
      emits: ['skip'],
      template: '<button data-test="setup-state" :data-show="String(show)" @click="$emit(\'skip\')">close</button>',
    })
    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: defineComponent({ template: '<main><slot /></main>' }),
          LoadingSpinner: true,
          Icon: true,
          UserDashboardStats: true,
          UserDashboardPerformanceStats: true,
          UserDashboardCharts: true,
          UserDashboardPlatformBreakdown: true,
          UserDashboardAccountUsage: true,
          UserDashboardRecentUsage: true,
          UserDashboardQuickActions: true,
          UserApiKeyOnboardingDialog: onboardingStub,
        },
      },
    })
    await flushPromises()
    expect(wrapper.get('[data-test="setup-state"]').attributes('data-show')).toBe('false')
    await wrapper.get('[data-testid="quick-setup"]').trigger('click')
    expect(wrapper.get('[data-test="setup-state"]').attributes('data-show')).toBe('true')
    await wrapper.get('[data-test="setup-state"]').trigger('click')
    expect(wrapper.get('[data-test="setup-state"]').attributes('data-show')).toBe('false')
    await wrapper.get('[data-testid="quick-setup"]').trigger('click')
    expect(wrapper.get('[data-test="setup-state"]').attributes('data-show')).toBe('true')
  })
})
