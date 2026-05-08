import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const { getDashboardLeaderboard, claimDashboardLeaderboardDailyReward } = vi.hoisted(() => ({
  getDashboardLeaderboard: vi.fn(),
  claimDashboardLeaderboardDailyReward: vi.fn(),
}))

vi.mock('@/api', () => ({
  usageAPI: {
    getDashboardLeaderboard,
    claimDashboardLeaderboardDailyReward,
  },
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  const messages: Record<string, string> = {
    'leaderboard.title': '排行榜',
    'leaderboard.description': '查看用户用量排名',
    'leaderboard.period.day': '日榜',
    'leaderboard.period.week': '周榜',
    'leaderboard.period.month': '月榜',
    'leaderboard.period.all': '总榜',
    'leaderboard.currentUser': '当前用户',
    'leaderboard.myRank': '我的排名',
    'leaderboard.emptyTitle': '暂无排名数据',
    'leaderboard.emptyDescription': '当前周期暂无可展示的使用记录',
    'leaderboard.totalCost': '总消费',
    'leaderboard.totalRequests': '总请求',
    'leaderboard.totalTokens': '总 Token',
    'leaderboard.user': '用户',
    'leaderboard.cost': '消费',
    'leaderboard.requests': '请求',
    'leaderboard.tokens': 'Token',
    'leaderboard.balance': '余额',
    'leaderboard.rank': '排名',
    'leaderboard.myInfo': '我的信息',
    'leaderboard.generatedAt': '生成时间',
    'leaderboard.notRanked': '未上榜',
    'leaderboard.dailyReward.title': '每日排名奖励',
    'leaderboard.dailyReward.settlementDate': '结算日期',
    'leaderboard.dailyReward.threshold': '昨日总消费门槛',
    'leaderboard.dailyReward.rewardAmount': '可领额度',
    'leaderboard.dailyReward.progress': '{current} / {target}',
    'leaderboard.dailyReward.disabled': '奖励功能暂未开启',
    'leaderboard.dailyReward.thresholdNotMet': '昨日总消费未超过最低开启门槛',
    'leaderboard.dailyReward.notTopThree': '只有昨日榜前三名可以领取',
    'leaderboard.dailyReward.notRanked': '你昨日暂无上榜消费',
    'leaderboard.dailyReward.zeroReward': '当前名次奖励额度为 0',
    'leaderboard.dailyReward.eligible': '你已符合领取条件',
    'leaderboard.dailyReward.alreadyClaimed': '昨日奖励已领取',
    'leaderboard.dailyReward.claim': '领取奖励',
    'leaderboard.dailyReward.claiming': '领取中...',
    'leaderboard.dailyReward.claimed': '已领取',
    'leaderboard.dailyReward.claimFailed': '领取排行榜奖励失败',
    'leaderboard.dailyReward.rankLabel': '第 {rank} 名',
    'leaderboard.dailyReward.rankReward': '第 {rank} 名奖励',
    'common.loading': '加载中...',
    'common.refresh': '刷新',
  }

  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) =>
        (messages[key] ?? key).replace(/\{(\w+)\}/g, (_, token) => String(params?.[token] ?? `{${token}}`)),
    }),
  }
})

const AppLayoutStub = {
  template: '<div><slot /></div>',
}

function makeResponse(overrides: Record<string, unknown> = {}) {
  return {
    period: 'day',
    start_date: '2026-05-07',
    end_date: '2026-05-07',
    generated_at: '2026-05-07T08:00:00Z',
    total_actual_cost: 3.5,
    total_requests: 12,
    total_tokens: 1200,
    ranking: [
      {
        rank: 1,
        user_id: 7,
        display_name: 'Alice',
        email_masked: 'a***@example.com',
        avatar_url: null,
        actual_cost: 2.5,
        requests: 8,
        tokens: 900,
        balance: 11,
        is_current_user: true,
      },
    ],
    current_user_entry: null,
    daily_rewards: {
      reward_date: '2026-05-06',
      settlement_timezone: 'Asia/Shanghai',
      enabled: false,
      min_total_actual_cost: 0,
      yesterday_total_actual_cost: 0,
      threshold_met: false,
      rewards: [
        { rank: 1, amount: 0 },
        { rank: 2, amount: 0 },
        { rank: 3, amount: 0 },
      ],
      current_user_rank: 0,
      current_user_reward_amount: 0,
      can_claim: false,
      claimed: false,
      reason: 'disabled',
    },
    ...overrides,
  }
}

describe('LeaderboardView', () => {
  beforeEach(() => {
    getDashboardLeaderboard.mockReset()
    claimDashboardLeaderboardDailyReward.mockReset()
    getDashboardLeaderboard.mockResolvedValue(makeResponse())
  })

  it('loads day leaderboard first and reloads when switching period', async () => {
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()

    expect(getDashboardLeaderboard).toHaveBeenCalledWith({ period: 'day', limit: 10 })

    await wrapper.findAll('button').find((button) => button.text() === '周榜')?.trigger('click')
    await flushPromises()

    expect(getDashboardLeaderboard).toHaveBeenLastCalledWith({ period: 'week', limit: 10 })
  })

  it('clears stale rows while loading the next period', async () => {
    let resolveWeek: (value: ReturnType<typeof makeResponse>) => void = () => {}
    getDashboardLeaderboard.mockResolvedValueOnce(makeResponse({ period: 'day', start_date: '2026-05-01', end_date: '2026-05-01' }))
    getDashboardLeaderboard.mockReturnValueOnce(new Promise((resolve) => {
      resolveWeek = resolve
    }))
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()
    expect(wrapper.text()).toContain('Alice')
    expect(wrapper.text()).toContain('2026-05-01')

    await wrapper.findAll('button').find((button) => button.text() === '周榜')?.trigger('click')
    expect(wrapper.text()).not.toContain('Alice')
    expect(wrapper.text()).not.toContain('2026-05-01')
    expect(wrapper.text()).toContain('加载中')

    resolveWeek(makeResponse({
      period: 'week',
      ranking: [
        {
          rank: 1,
          user_id: 8,
          display_name: 'Weekly User',
          email_masked: 'w***@example.com',
          avatar_url: null,
          actual_cost: 5,
          requests: 9,
          tokens: 1000,
          balance: 3,
          is_current_user: false,
        },
      ],
    }))
    await flushPromises()

    expect(wrapper.text()).toContain('Weekly User')
  })

  it('highlights the current user row and shows my rank summary outside the top list', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        ranking: [
          {
            rank: 1,
            user_id: 2,
            display_name: 'Bob',
            email_masked: 'b***@example.com',
            avatar_url: null,
            actual_cost: 4.25,
            requests: 16,
            tokens: 2400,
            balance: 20,
            is_current_user: false,
          },
        ],
        current_user_entry: {
          rank: 28,
          user_id: 9,
          display_name: 'Me',
          email_masked: 'm***@example.com',
          avatar_url: null,
          actual_cost: 0.5,
          requests: 2,
          tokens: 120,
          balance: 7,
          is_current_user: true,
        },
      })
    )
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('我的信息')
    expect(wrapper.text()).toContain('#28')
    expect(wrapper.text()).toContain('Me')
    expect(wrapper.text()).toContain('$7.00')
  })

  it('shows my info even when the current user is in the top list', async () => {
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('[data-testid="leaderboard-my-info"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).toContain('Alice')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).toContain('$11.00')
  })

  it('renders at most 10 ranking items', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        ranking: Array.from({ length: 12 }, (_, index) => ({
          rank: index + 1,
          user_id: index + 1,
          display_name: `User ${index + 1}`,
          email_masked: `u***${index + 1}@example.com`,
          avatar_url: null,
          actual_cost: 12 - index,
          requests: 10 + index,
          tokens: 1000 + index,
          balance: index,
          is_current_user: false,
        })),
        current_user_entry: null,
      })
    )
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('User 10')
    expect(wrapper.text()).not.toContain('User 11')
    expect(wrapper.text()).not.toContain('User 12')
  })

  it('renders an empty state when there are no ranking items', async () => {
    getDashboardLeaderboard.mockResolvedValue(makeResponse({ ranking: [], current_user_entry: null }))
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('暂无排名数据')
    expect(wrapper.text()).toContain('当前周期暂无可展示的使用记录')
  })

  it('renders daily rewards with threshold-not-met copy', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        daily_rewards: {
          reward_date: '2026-05-06',
          settlement_timezone: 'Asia/Shanghai',
          enabled: true,
          min_total_actual_cost: 100,
          yesterday_total_actual_cost: 80,
          threshold_met: false,
          rewards: [
            { rank: 1, amount: 5 },
            { rank: 2, amount: 3 },
            { rank: 3, amount: 1 },
          ],
          current_user_rank: 1,
          current_user_reward_amount: 5,
          can_claim: false,
          claimed: false,
          reason: 'threshold_not_met',
        },
      })
    )
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('[data-testid="leaderboard-daily-reward"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('每日排名奖励')
    expect(wrapper.text()).toContain('昨日总消费未超过最低开启门槛')
    expect(wrapper.text()).toContain('第 1 名奖励')
    expect(wrapper.text()).toContain('$5.00')
  })

  it('claims an eligible daily reward and shows claimed state', async () => {
    const claimRewards = {
      reward_date: '2026-05-06',
      settlement_timezone: 'Asia/Shanghai',
      enabled: true,
      min_total_actual_cost: 100,
      yesterday_total_actual_cost: 120,
      threshold_met: true,
      rewards: [
        { rank: 1, amount: 5 },
        { rank: 2, amount: 3 },
        { rank: 3, amount: 1 },
      ],
      current_user_rank: 1,
      current_user_reward_amount: 5,
      can_claim: true,
      claimed: false,
      reason: 'eligible',
    }
    getDashboardLeaderboard.mockResolvedValue(makeResponse({ daily_rewards: claimRewards }))
    claimDashboardLeaderboardDailyReward.mockResolvedValue({
      claimed_amount: 5,
      daily_rewards: {
        ...claimRewards,
        can_claim: false,
        claimed: true,
        reason: 'already_claimed',
      },
    })
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()
    await wrapper.get('[data-testid="leaderboard-daily-reward-claim"]').trigger('click')
    await flushPromises()

    expect(claimDashboardLeaderboardDailyReward).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('昨日奖励已领取')
    expect(wrapper.get('[data-testid="leaderboard-daily-reward-claim"]').text()).toContain('已领取')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).toContain('$16.00')
  })

  it('adds top-three avatar frame classes', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        ranking: [1, 2, 3].map((rank) => ({
          rank,
          user_id: rank,
          display_name: `Top ${rank}`,
          email_masked: `t***${rank}@example.com`,
          avatar_url: null,
          actual_cost: 10 - rank,
          requests: 10,
          tokens: 1000,
          balance: rank,
          is_current_user: rank === 1,
        })),
      })
    )
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('.leaderboard-avatar-frame-gold').exists()).toBe(true)
    expect(wrapper.find('.leaderboard-avatar-frame-silver').exists()).toBe(true)
    expect(wrapper.find('.leaderboard-avatar-frame-bronze').exists()).toBe(true)
  })
})
