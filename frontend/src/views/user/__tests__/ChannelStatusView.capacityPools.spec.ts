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
      t: (key: string) => key,
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
        schedulable_accounts: 1,
        configured_quota: 0,
        remaining_quota: 0,
        sections: [],
      },
      shared: {
        key: 'shared',
        title: 'Shared Platform Capacity Pool',
        total_accounts: 2,
        schedulable_accounts: 2,
        configured_quota: 0,
        remaining_quota: 0,
        sections: [],
      },
    })
    showError.mockReset()
    appStoreState.cachedPublicSettings = {
      account_share_enabled: true,
      channel_monitor_enabled: true,
    }
  })

  it('loads and renders only my capacity pool when account sharing is enabled', async () => {
    const wrapper = mountView()

    await flushPromises()

    expect(fetchCapacityPools).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('channelStatus.capacityPools.mine')
    expect(wrapper.text()).not.toContain('channelStatus.capacityPools.shared')
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
})
