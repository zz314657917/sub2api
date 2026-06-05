import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import AccountCapacityPools from '../AccountCapacityPools.vue'
import type { UserAccountCapacityPools } from '@/types'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const messages: Record<string, string> = {
        'channelStatus.capacityPools.mineTitle': '我的账号容量池',
        'channelStatus.capacityPools.mineDescription': '展示我的账号容量。',
        'channelStatus.capacityPools.sharedTitle': '平台共享容量池',
        'channelStatus.capacityPools.sharedDescription': '展示平台共享容量。',
        'channelStatus.capacityPools.total': '总数',
        'channelStatus.capacityPools.active': '活跃',
        'channelStatus.capacityPools.schedulable': '可调度',
        'channelStatus.capacityPools.rateLimited': '限流',
        'channelStatus.capacityPools.abnormal': '异常',
        'channelStatus.capacityPools.groupCapacity': '分组容量',
        'channelStatus.capacityPools.groupCapacityHint': '按共享分组展示。',
        'channelStatus.capacityPools.healthy': '正常',
        'channelStatus.capacityPools.degraded': '部分可用',
        'channelStatus.capacityPools.unavailable': '不可用',
        'channelStatus.capacityPools.accountStatus': '账号状态',
        'channelStatus.capacityPools.schedulableSnapshot': '可用账号',
        'channelStatus.capacityPools.schedulableRemaining': '剩余',
        'channelStatus.capacityPools.window': `${params?.window} 窗口`,
        'myAccounts.platforms.openai': 'OpenAI',
      }
      return messages[key] ?? key
    },
  }),
}))

const pools: UserAccountCapacityPools = {
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
    sections: [
      {
        platform: 'openai',
        type: 'apikey',
        total_accounts: 1,
        schedulable_accounts: 1,
        configured_quota: 0,
        remaining_quota: 0,
      },
    ],
    groups: [],
  },
  shared: {
    key: 'shared',
    title: 'Shared Platform Capacity Pool',
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
    sections: [],
    groups: [
      {
        key: 'share-display:openai:openai pro',
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
        status: 'degraded',
        windows: {
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
  external: null,
}

describe('AccountCapacityPools', () => {
  it('can render only the shared capacity pool', () => {
    const wrapper = mount(AccountCapacityPools, {
      props: {
        pools,
        poolKeys: ['shared'],
      },
    })

    expect(wrapper.text()).toContain('平台共享容量池')
    expect(wrapper.text()).toContain('OpenAI Pro')
    expect(wrapper.text()).not.toContain('我的账号容量池')
  })

  it('can render only my account capacity pool', () => {
    const wrapper = mount(AccountCapacityPools, {
      props: {
        pools,
        poolKeys: ['mine'],
      },
    })

    expect(wrapper.text()).toContain('我的账号容量池')
    expect(wrapper.text()).toContain('OpenAI / apikey')
    expect(wrapper.text()).not.toContain('平台共享容量池')
    expect(wrapper.text()).not.toContain('OpenAI Pro')
  })
})
