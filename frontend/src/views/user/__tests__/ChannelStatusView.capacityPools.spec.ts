import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ChannelStatusView from '../ChannelStatusView.vue'

const { listChannelMonitors, fetchCapacityPools, showError, appStoreState } = vi.hoisted(() => ({
  listChannelMonitors: vi.fn(),
  fetchCapacityPools: vi.fn(),
  showError: vi.fn(),
  appStoreState: {
    cachedPublicSettings: {
      account_share_enabled: true,
      channel_monitor_enabled: true,
    } as Record<string, unknown> | null,
  },
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'channelStatus.capacityPools.quotaWindow') return `${params?.window} 额度`
        if (key === 'channelStatus.capacityPools.window') return `${params?.window} 窗口`
        if (key === 'channelStatus.capacityPools.percentOnly') return '剩余'
        if (key === 'channelStatus.capacityPools.accountStatus') return '账号状态'
        if (key === 'channelStatus.capacityPools.degraded') return '部分可用'
        if (key === 'channelStatus.capacityPools.schedulableSnapshot') return '可用账号'
        if (key === 'channelStatus.capacityPools.unavailableReasons.daily_quota_exceeded') return '日额度用完'
        if (key === 'channelStatus.capacityPools.unavailableReasons.rate_limited') return '限流中'
        return key
      },
    }),
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: appStoreState.cachedPublicSettings,
    showError,
  }),
}))

vi.mock('@/api/channelMonitor', () => ({
  list: listChannelMonitors,
  status: vi.fn(),
}))

vi.mock('@/api/user', () => ({
  getAccountCapacityPools: fetchCapacityPools,
}))

function mountView() {
  return mount(ChannelStatusView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        MonitorHero: { template: '<div />' },
        MonitorCardGrid: { template: '<div />' },
        MonitorDetailDialog: { template: '<div />' },
      },
    },
  })
}

describe('ChannelStatusView capacity pools', () => {
  beforeEach(() => {
    listChannelMonitors.mockReset().mockResolvedValue({ items: [] })
    fetchCapacityPools.mockReset().mockResolvedValue({
      mine: {
        key: 'mine',
        title: 'My Account Capacity Pool',
        total_accounts: 1,
        active_accounts: 1,
        schedulable_accounts: 1,
        rate_limited_accounts: 0,
        error_accounts: 0,
        disabled_accounts: 0,
        abnormal_accounts: 0,
        configured_quota: 0,
        remaining_quota: 0,
        sections: [],
        groups: [],
      },
      shared: {
        key: 'shared',
        title: 'Shared Platform Capacity Pool',
        total_accounts: 2,
        active_accounts: 2,
        schedulable_accounts: 2,
        own_contributed_accounts: 1,
        rate_limited_accounts: 1,
        error_accounts: 0,
        disabled_accounts: 0,
        abnormal_accounts: 1,
        configured_quota: 0,
        remaining_quota: 0,
        sections: [],
        groups: [
          {
            key: 'share-display:openai:openai pro',
            group_id: undefined,
            group_name: 'OpenAI Pro',
            platform: 'openai',
            sort_order: 0,
            total_accounts: 2,
            active_accounts: 2,
            schedulable_accounts: 1,
            own_contributed_accounts: 1,
            rate_limited_accounts: 1,
            error_accounts: 0,
            disabled_accounts: 0,
            abnormal_accounts: 1,
            configured_quota: 0,
            remaining_quota: 0,
            percent_only_quota: true,
            unavailable_reasons: {
              daily_quota_exceeded: 1,
              rate_limited: 1,
            },
            status: 'degraded',
            windows: {
              '1d': {
                used_percent: 12.5,
                snapshot_accounts: 1,
                schedulable_snapshot_accounts: 1,
                remaining_units: 0.875,
              },
              '7d_quota': {
                used_percent: 50,
                snapshot_accounts: 1,
                schedulable_snapshot_accounts: 1,
                remaining_units: 0.5,
              },
              '30d': {
                used_percent: 75,
                snapshot_accounts: 1,
                schedulable_snapshot_accounts: 1,
                remaining_units: 0.25,
              },
              '5h': {
                used_percent: 40.3,
                snapshot_accounts: 2,
                schedulable_snapshot_accounts: 1,
                remaining_units: 0.6,
              },
              '7d': {
                used_percent: 20,
                snapshot_accounts: 2,
                schedulable_snapshot_accounts: 1,
                remaining_units: 0.8,
              },
            },
          },
        ],
      },
    })
    showError.mockReset()
    appStoreState.cachedPublicSettings = {
      account_share_enabled: true,
      channel_monitor_enabled: true,
    }
  })

  it('loads and renders my and shared capacity pools when account sharing is enabled', async () => {
    const wrapper = mountView()

    await flushPromises()

    expect(fetchCapacityPools).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('channelStatus.capacityPools.mine')
    expect(wrapper.text()).toContain('channelStatus.capacityPools.shared')
    expect(wrapper.text()).toContain('OpenAI Pro')
    expect(wrapper.text()).not.toContain('1d 额度')
    expect(wrapper.text()).not.toContain('7d 额度')
    expect(wrapper.text()).not.toContain('30d 额度')
    expect(wrapper.text()).toContain('5h 窗口')
    expect(wrapper.text()).toContain('7d 窗口')
    expect(wrapper.text()).toContain('账号状态')
    expect(wrapper.text()).toContain('部分可用')
    expect(wrapper.text()).not.toContain('降级')
    expect(wrapper.text()).toContain('可用账号')
    expect(wrapper.text()).toContain('channelStatus.capacityPools.schedulableRemaining')
    expect(wrapper.text()).toContain('59.7%')
    expect(wrapper.text()).toContain('80.0%')
    expect(wrapper.text()).toContain('channelStatus.capacityPools.ownContributed')
    expect(wrapper.text()).toContain('channelStatus.capacityPools.unavailableReason')
    expect(wrapper.text()).toContain('日额度用完 1')
    expect(wrapper.text()).toContain('限流中 1')
  })

  it('does not request or render capacity pools when account sharing is disabled', async () => {
    appStoreState.cachedPublicSettings = {
      account_share_enabled: false,
      channel_monitor_enabled: true,
    }

    const wrapper = mountView()

    await flushPromises()

    expect(fetchCapacityPools).not.toHaveBeenCalled()
    expect(wrapper.text()).not.toContain('channelStatus.capacityPools.mine')
    expect(wrapper.text()).not.toContain('channelStatus.capacityPools.shared')
  })

  it('hides capacity pools and groups without accounts', async () => {
    fetchCapacityPools.mockResolvedValueOnce({
      mine: {
        key: 'mine',
        title: 'My Account Capacity Pool',
        total_accounts: 0,
        active_accounts: 0,
        schedulable_accounts: 0,
        rate_limited_accounts: 0,
        error_accounts: 0,
        disabled_accounts: 0,
        abnormal_accounts: 0,
        configured_quota: 0,
        remaining_quota: 0,
        sections: [],
        groups: [],
      },
      shared: {
        key: 'shared',
        title: 'Shared Platform Capacity Pool',
        total_accounts: 1,
        active_accounts: 1,
        schedulable_accounts: 1,
        own_contributed_accounts: 0,
        rate_limited_accounts: 0,
        error_accounts: 0,
        disabled_accounts: 0,
        abnormal_accounts: 0,
        configured_quota: 0,
        remaining_quota: 0,
        sections: [
          {
            platform: 'openai',
            type: 'apikey',
            total_accounts: 0,
            schedulable_accounts: 0,
            configured_quota: 0,
            remaining_quota: 0,
          },
        ],
        groups: [
          {
            key: 'share-display:openai:openai plus',
            group_name: 'OpenAI Plus',
            platform: 'openai',
            sort_order: 0,
            total_accounts: 0,
            active_accounts: 0,
            schedulable_accounts: 0,
            rate_limited_accounts: 0,
            error_accounts: 0,
            disabled_accounts: 0,
            abnormal_accounts: 0,
            configured_quota: 0,
            remaining_quota: 0,
            status: 'unavailable',
          },
          {
            key: 'share-display:openai:openai pro',
            group_name: 'OpenAI Pro',
            platform: 'openai',
            sort_order: 1,
            total_accounts: 1,
            active_accounts: 1,
            schedulable_accounts: 1,
            rate_limited_accounts: 0,
            error_accounts: 0,
            disabled_accounts: 0,
            abnormal_accounts: 0,
            configured_quota: 0,
            remaining_quota: 0,
            status: 'healthy',
          },
        ],
      },
    })

    const wrapper = mountView()

    await flushPromises()

    expect(wrapper.text()).not.toContain('channelStatus.capacityPools.mine')
    expect(wrapper.text()).toContain('channelStatus.capacityPools.shared')
    expect(wrapper.text()).not.toContain('OpenAI Plus')
    expect(wrapper.text()).toContain('OpenAI Pro')
    expect(wrapper.text()).not.toContain('openai / apikey')
  })

  it('keeps quota windows and remaining percentages in non-plan fallback sections', async () => {
    fetchCapacityPools.mockResolvedValueOnce({
      mine: {
        key: 'mine',
        title: 'My Account Capacity Pool',
        total_accounts: 1,
        active_accounts: 1,
        schedulable_accounts: 1,
        rate_limited_accounts: 0,
        error_accounts: 0,
        disabled_accounts: 0,
        abnormal_accounts: 0,
        configured_quota: 0,
        remaining_quota: 0,
        percent_only_quota: true,
        sections: [
          {
            platform: 'openai',
            type: 'apikey',
            total_accounts: 1,
            schedulable_accounts: 1,
            configured_quota: 0,
            remaining_quota: 0,
            windows: {
              '1d': { used_percent: 25, reset_after_seconds: 3600, window_minutes: 1440 },
              '7d_quota': { used_percent: 40, reset_after_seconds: 7200, window_minutes: 10080 },
            },
          },
        ],
        groups: [],
      },
      shared: {
        key: 'shared',
        title: 'Shared Platform Capacity Pool',
        total_accounts: 0,
        active_accounts: 0,
        schedulable_accounts: 0,
        rate_limited_accounts: 0,
        error_accounts: 0,
        disabled_accounts: 0,
        abnormal_accounts: 0,
        configured_quota: 0,
        remaining_quota: 0,
        sections: [],
        groups: [],
      },
    })

    const wrapper = mountView()

    await flushPromises()

    expect(wrapper.text()).toContain('channelStatus.capacityPools.mine')
    expect(wrapper.text()).not.toContain('channelStatus.capacityPools.shared')
    expect(wrapper.text()).toContain('1d 额度 75.0%')
    expect(wrapper.text()).toContain('7d 额度 60.0%')
  })
})
