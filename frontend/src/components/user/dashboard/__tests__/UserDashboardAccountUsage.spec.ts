import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import UserDashboardAccountUsage from '../UserDashboardAccountUsage.vue'
import type { UserAccountUsageSummary } from '@/types'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const messages: Record<string, string> = {
        'dashboard.personalAccountUsage.title': '个人账号用量',
        'dashboard.personalAccountUsage.settlementLedger': '结算流水',
        'dashboard.personalAccountUsage.myAccounts': '我的账号',
        'dashboard.personalAccountUsage.publicApprovedInline': '公共已通过 {count}',
        'dashboard.personalAccountUsage.ownUsage': '自己使用',
        'dashboard.personalAccountUsage.sharedUsage': '别人使用',
        'dashboard.personalAccountUsage.shareIncome': '分成入账',
        'dashboard.personalAccountUsage.platformAmount': '平台留存 ${amount}',
        'dashboard.personalAccountUsage.balanceNetChange': '余额净变化',
        'dashboard.personalAccountUsage.balanceDeduction': '余额扣减 ${amount}',
        'dashboard.personalAccountUsage.privateAccounts': '私有',
        'dashboard.personalAccountUsage.publicPending': '公共待校验',
        'dashboard.personalAccountUsage.publicApproved': '公共已通过',
        'dashboard.personalAccountUsage.publicSuspended': '公共已暂停',
        'dashboard.personalAccountUsage.accountDetails': '账号明细',
        'dashboard.personalAccountUsage.accountCost': '账号成本 ${amount}',
        'dashboard.personalAccountUsage.requests': '{count} 请求',
        'dashboard.personalAccountUsage.noUsage': '当前时间范围内暂无个人账号用量',
        'dashboard.personalAccountUsage.usageSummary': '当前时间范围内共有 {count} 次个人账号请求',
      }
      let message = messages[key] ?? key
      Object.entries(params ?? {}).forEach(([paramKey, value]) => {
        message = message.replaceAll(`{${paramKey}}`, String(value))
      })
      return message
    },
  }),
}))

const makeSummary = (overrides: Partial<UserAccountUsageSummary> = {}): UserAccountUsageSummary => ({
  owner_user_id: 42,
  start_date: '2026-05-07',
  end_date: '2026-05-13',
  total_accounts: 4,
  private_accounts: 1,
  public_pending_accounts: 1,
  public_active_accounts: 2,
  public_suspended_accounts: 0,
  own_usage_cost: 1.25,
  own_usage_requests: 3,
  shared_usage_cost: 2.5,
  shared_usage_requests: 4,
  share_income: 0.75,
  platform_amount: 0.25,
  account_cost: 4.5,
  balance_deduction: 0.5,
  balance_net_change: 0.25,
  ...overrides,
})

describe('UserDashboardAccountUsage', () => {
  it('renders account usage totals and status counts', () => {
    const wrapper = mount(UserDashboardAccountUsage, {
      props: {
        summary: makeSummary(),
        loading: false,
        startDate: '2026-05-01',
        endDate: '2026-05-06',
      },
      global: {
        stubs: {
          Icon: true,
          LoadingSpinner: true,
          RouterLink: true,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('个人账号用量')
    expect(text).toContain('2026-05-07 - 2026-05-13')
    expect(text).toContain('我的账号')
    expect(text).toContain('4')
    expect(text).toContain('$1.2500')
    expect(text).toContain('$2.5000')
    expect(text).toContain('$0.7500')
    expect(text).toContain('+$0.2500')
    expect(text).toContain('账号成本 $4.5000')
    expect(text).toContain('当前时间范围内共有 7 次个人账号请求')
  })

  it('renders empty-state copy when there are no requests', () => {
    const wrapper = mount(UserDashboardAccountUsage, {
      props: {
        summary: makeSummary({
          own_usage_requests: 0,
          shared_usage_requests: 0,
          own_usage_cost: 0,
          shared_usage_cost: 0,
        }),
        loading: false,
        startDate: '2026-05-07',
        endDate: '2026-05-13',
      },
      global: {
        stubs: {
          Icon: true,
          LoadingSpinner: true,
          RouterLink: true,
        },
      },
    })

    expect(wrapper.text()).toContain('当前时间范围内暂无个人账号用量')
  })
})
