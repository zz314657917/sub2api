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
    'leaderboard.costEfficiencyKing': '⭐ 性价比之王',
    'leaderboard.badges.weeklyTokenKing': '周榜 Token 最多',
    'leaderboard.badges.monthlyTokenKing': '月榜 Token 最多',
    'leaderboard.badges.totalTokenKing': '肝帝',
    'leaderboard.badges.nightOwl': '夜猫',
    'leaderboard.badges.burstTokenKing': '爆肝王',
    'leaderboard.badges.checkinKing': '打卡王',
    'leaderboard.badges.costSaver': '1M Token 成本最低',
    'leaderboard.badges.costBurner': '1M Token 成本最高',
    'leaderboard.generatedAt': '生成时间',
    'leaderboard.notRanked': '未上榜',
    'leaderboard.dailyReward.title': '每日排名奖励',
    'leaderboard.dailyReward.settlementDate': '结算日期',
    'leaderboard.dailyReward.threshold': '昨日总消费门槛',
    'leaderboard.dailyReward.rewardAmount': '可领额度',
    'leaderboard.dailyReward.progress': '{current} / {target}',
    'leaderboard.dailyReward.disabled': '奖励功能暂未开启',
    'leaderboard.dailyReward.settling': '昨日榜结算中，{time} 后可领取',
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
      settlement_ready: true,
      claim_available_at: '2026-05-07T00:30:00+08:00',
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
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).not.toContain('$7.00')
    expect(wrapper.text()).not.toContain('$4.25')
    expect(wrapper.text()).not.toContain('$0.50')
    expect(wrapper.text()).not.toContain('b***@example.com')
    expect(wrapper.text()).not.toContain('m***@example.com')
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
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).toContain('8')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).toContain('900')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).not.toContain('$11.00')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).not.toContain('余额')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').findAll('.rounded-lg.bg-gray-50')).toHaveLength(2)
  })

  it('hides visible leaderboard spending totals and row amounts', async () => {
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).not.toContain('$3.50')
    expect(wrapper.text()).not.toContain('$2.50')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).not.toContain('$11.00')
  })

  it('marks the lowest cost per token user as the value king', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        ranking: [
          {
            rank: 1,
            user_id: 1,
            display_name: 'Pricey',
            email_masked: 'p***@example.com',
            avatar_url: null,
            actual_cost: 10,
            requests: 10,
            tokens: 1000,
            balance: 1,
            is_current_user: false,
          },
          {
            rank: 2,
            user_id: 2,
            display_name: 'Efficient',
            email_masked: 'e***@example.com',
            avatar_url: null,
            actual_cost: 1,
            requests: 10,
            tokens: 2000,
            balance: 2,
            badges: ['cost_saver'],
            is_current_user: false,
          },
          {
            rank: 3,
            user_id: 3,
            display_name: 'Free',
            email_masked: 'f***@example.com',
            avatar_url: null,
            actual_cost: 0,
            requests: 10,
            tokens: 999999,
            balance: 3,
            is_current_user: false,
          },
        ],
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

    const badges = wrapper.findAll('[data-testid="leaderboard-badge-icon"]')
    expect(badges).toHaveLength(4)
    expect(badges.map((badge) => badge.attributes('data-user-id'))).toEqual(['1', '2', '1', '2'])
    expect(badges.map((badge) => badge.attributes('data-badge'))).toEqual(['cost_burner', 'cost_saver', 'cost_burner', 'cost_saver'])
    expect(badges.map((badge) => badge.text())).toEqual(['豪', '省', '豪', '省'])
    expect(badges[1].attributes('title')).toBe('1M Token 成本最低')
    expect(wrapper.text()).toContain('⭐ 性价比之王')
  })

  it('shows the value king summary in the top stats', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        ranking: [
          {
            rank: 1,
            user_id: 1,
            display_name: 'Pricey',
            email_masked: 'p***@example.com',
            avatar_url: null,
            actual_cost: 10,
            requests: 10,
            tokens: 1000,
            balance: 1,
            is_current_user: false,
          },
          {
            rank: 2,
            user_id: 2,
            display_name: 'Efficient',
            email_masked: 'e***@example.com',
            avatar_url: null,
            actual_cost: 1,
            requests: 10,
            tokens: 2000,
            balance: 2,
            is_current_user: false,
          },
        ],
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

    const summary = wrapper.get('[data-testid="leaderboard-cost-efficiency-summary"]')
    expect(summary.text()).toContain('Efficient')
    expect(summary.text()).toContain('1M Token = $500.00')
  })

  it('renders compact leaderboard badge icons and collapses overflow badges', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        ranking: [
          {
            rank: 1,
            user_id: 1,
            display_name: 'Decorated',
            email_masked: 'd***@example.com',
            avatar_url: null,
            actual_cost: 10,
            requests: 10,
            tokens: 1000,
            balance: 1,
            badges: ['weekly_token_king', 'monthly_token_king', 'total_token_king', 'night_owl', 'burst_token_king', 'checkin_king', 'cost_saver', 'cost_burner'],
            is_current_user: false,
          },
        ],
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

    const badges = wrapper.findAll('[data-testid="leaderboard-badge-icon"]').slice(0, 3)
    expect(badges.map((badge) => badge.text())).toEqual(['周', '月', '肝'])
    expect(badges.map((badge) => badge.attributes('data-badge'))).toEqual([
      'weekly_token_king',
      'monthly_token_king',
      'total_token_king',
    ])
    expect(wrapper.find('.leaderboard-badge-overflow').text()).toBe('+5')
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
          settlement_ready: true,
          claim_available_at: '2026-05-07T00:30:00+08:00',
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

  it('shows settling state before daily rewards can be claimed', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        daily_rewards: {
          reward_date: '2026-05-06',
          settlement_timezone: 'Asia/Shanghai',
          settlement_ready: false,
          claim_available_at: '2026-05-07T00:30:00+08:00',
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
          can_claim: false,
          claimed: false,
          reason: 'settling',
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

    expect(wrapper.text()).toContain('昨日榜结算中')
    expect(wrapper.get('[data-testid="leaderboard-daily-reward-claim"]').attributes('disabled')).toBeDefined()
  })

  it('claims an eligible daily reward and shows claimed state', async () => {
    const claimRewards = {
      reward_date: '2026-05-06',
      settlement_timezone: 'Asia/Shanghai',
      settlement_ready: true,
      claim_available_at: '2026-05-07T00:30:00+08:00',
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
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).not.toContain('$16.00')
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
