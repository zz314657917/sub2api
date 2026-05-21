import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

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

vi.mock('vue-chartjs', () => ({
  Line: {
    name: 'Line',
    props: ['data', 'options'],
    template: '<div data-testid="leaderboard-recent-token-line">{{ data?.datasets?.[0]?.data?.join(\',\') }}</div>',
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
    'leaderboard.tokenRankingTitle': 'Token Top {count}',
    'leaderboard.tokenRankingDescription': '当前周期用量排行。',
    'leaderboard.user': '用户',
    'leaderboard.cost': '消费',
    'leaderboard.requests': '请求',
    'leaderboard.tokens': 'Token',
    'leaderboard.inputTokensShort': '输入',
    'leaderboard.outputTokensShort': '输出',
    'leaderboard.costPerMillionShort': '费用',
    'leaderboard.recentTokenTrend.title': '最近 10 天 Token',
    'leaderboard.recentTokenTrend.unit': '每日消耗',
    'leaderboard.recentTokenTrend.tokens': 'Token',
    'leaderboard.recentTokenTrend.empty': '暂无趋势数据',
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
    'leaderboard.generatedAt': '更新',
    'leaderboard.notRanked': '未上榜',
    'leaderboard.dailyReward.title': '每日排名奖励',
    'leaderboard.dailyReward.settlementDate': '结算日期',
    'leaderboard.dailyReward.threshold': '昨日总消费门槛',
    'leaderboard.dailyReward.rewardAmount': '可领额度',
    'leaderboard.dailyReward.rewardAmountHidden': '按名次发放',
    'leaderboard.dailyReward.targetProgress': '奖励目标进度',
    'leaderboard.dailyReward.progress': '{current} / {target}',
    'leaderboard.dailyReward.progressPercent': '{percent}%',
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

enableAutoUnmount(afterEach)

afterEach(() => {
  vi.useRealTimers()
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
    recent_token_trend: [
      { date: '2026-04-28', total_tokens: 120 },
      { date: '2026-04-29', total_tokens: 0 },
      { date: '2026-04-30', total_tokens: 340 },
      { date: '2026-05-01', total_tokens: 180 },
      { date: '2026-05-02', total_tokens: 260 },
      { date: '2026-05-03', total_tokens: 420 },
      { date: '2026-05-04', total_tokens: 390 },
      { date: '2026-05-05', total_tokens: 510 },
      { date: '2026-05-06', total_tokens: 470 },
      { date: '2026-05-07', total_tokens: 640 },
    ],
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
    expect(wrapper.text()).toContain('更新')
    expect(wrapper.text()).not.toContain('2026-05-01')

    await wrapper.findAll('button').find((button) => button.text() === '周榜')?.trigger('click')
    expect(wrapper.text()).not.toContain('Alice')
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
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).toContain('900')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).not.toContain('$11.00')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).not.toContain('余额')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').findAll('[data-testid="leaderboard-my-token"]')).toHaveLength(1)
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

  it('keeps the token leaderboard focused on token usage only', async () => {
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
            input_tokens: 700,
            output_tokens: 200,
            tokens: 1000,
            cost_per_1m_tokens: 10000,
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

    expect(wrapper.findAll('[data-testid="leaderboard-badge-icon"]')).toHaveLength(0)
    expect(wrapper.get('[data-testid="leaderboard-token-ranking"]').text()).toContain('1M')
    expect(wrapper.text()).not.toContain('⭐ 性价比之王')
  })

  it('renders token usage ranking bars as the main leaderboard', async () => {
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
            input_tokens: 700,
            output_tokens: 200,
            tokens: 1000,
            cost_per_1m_tokens: 10000,
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
            input_tokens: 1500,
            output_tokens: 300,
            tokens: 2000,
            cost_per_1m_tokens: 500,
            balance: 2,
            is_current_user: false,
          },
        ],
        current_user_entry: null,
        total_tokens: 3000,
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

    const ranking = wrapper.get('[data-testid="leaderboard-token-ranking"]')
    expect(ranking.text()).toContain('Token Top 2')
    expect(ranking.text()).toContain('Pricey')
    expect(ranking.text()).toContain('Efficient')
    expect(ranking.text()).toContain('2,000')
    expect(ranking.findAll('.leaderboard-token-bar-fill')).toHaveLength(2)
    const tokenBars = ranking.findAll('.leaderboard-token-bar-track')
    expect(tokenBars[0].attributes('title')).toBe('输入 700 / 输出 200 / 费用 $10,000.00 / 1M Token')
    expect(tokenBars[0].attributes('aria-label')).toBe('输入 700 / 输出 200 / 费用 $10,000.00 / 1M Token')
    expect(tokenBars[1].attributes('title')).toBe('输入 1,500 / 输出 300 / 费用 $500.00 / 1M Token')
    expect(tokenBars[1].attributes('aria-label')).toBe('输入 1,500 / 输出 300 / 费用 $500.00 / 1M Token')
    expect(ranking.findAll('.leaderboard-token-rank-row')[1].attributes('style')).toContain('--token-bar-width: 84%')
    expect(wrapper.find('[data-testid="leaderboard-cost-efficiency-summary"]').exists()).toBe(false)
  })

  it('keeps the total token display increasing visually between refreshes', async () => {
    vi.useFakeTimers()
    Object.defineProperty(document, 'hidden', { configurable: true, value: false })
    getDashboardLeaderboard.mockResolvedValue(makeResponse({ total_tokens: 1200 }))
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()

    const odometer = wrapper.get('[data-testid="leaderboard-total-token-odometer"]')
    expect(odometer.attributes('aria-label')).toBe('1,200')

    await vi.advanceTimersByTimeAsync(3000)
    await flushPromises()

    expect(odometer.attributes('aria-label')).toBe('1,237')
    expect(wrapper.get('[data-testid="leaderboard-token-ranking"]').text()).not.toContain('1,200 Token')
  })

  it('renders recent daily token trend next to the total token summary', async () => {
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()

    const trendPanel = wrapper.get('[data-testid="leaderboard-recent-token-trend"]')
    expect(trendPanel.text()).toContain('最近 10 天 Token')
    expect(trendPanel.text()).toContain('每日消耗')
    expect(wrapper.get('[data-testid="leaderboard-recent-token-line"]').text()).toBe('120,0,340,180,260,420,390,510,470,640')
  })

  it('shows lightweight rank titles without restoring old row badges', async () => {
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
            tokens: 3000,
            balance: 1,
            badges: ['weekly_token_king', 'total_token_king', 'cost_saver'],
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

    const titles = wrapper.findAll('[data-testid="leaderboard-rank-title"]')
    expect(titles.map((title) => title.text())).toEqual(['周榜王', '肝帝'])
    expect(titles.map((title) => title.attributes('data-badge'))).toEqual(['weekly_token_king', 'total_token_king'])
    expect(wrapper.findAll('[data-testid="leaderboard-badge-icon"]')).toHaveLength(0)
  })

  it('renders compact personal badge icons and collapses overflow badges in my info', async () => {
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
        current_user_entry: {
          rank: 16,
          user_id: 1,
          display_name: 'Decorated',
          email_masked: 'd***@example.com',
          avatar_url: null,
          actual_cost: 10,
          requests: 10,
          tokens: 1000,
          balance: 1,
          badges: ['weekly_token_king', 'monthly_token_king', 'total_token_king', 'night_owl', 'burst_token_king', 'checkin_king', 'cost_saver', 'cost_burner'],
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

    const badges = wrapper.findAll('[data-testid="leaderboard-my-badge-icon"]').slice(0, 3)
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
          display_name: `Ranked User ${String(index + 1).padStart(2, '0')}`,
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

    expect(wrapper.text()).toContain('Ranked User 10')
    expect(wrapper.text()).not.toContain('Ranked User 11')
    expect(wrapper.text()).not.toContain('Ranked User 12')
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
    expect(wrapper.text()).not.toContain('昨日总消费门槛')
    expect(wrapper.text()).not.toContain('$80.00 / $100.00')
    expect(wrapper.text()).toContain('第 1 名奖励')
    expect(wrapper.text()).toContain('按名次发放')
    expect(wrapper.text()).toContain('奖励目标进度')
    expect(wrapper.text()).toContain('80%')
    expect(wrapper.text()).not.toContain('$5.00')
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

  it('colors token bars by rank without row background frames', async () => {
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

    const rows = wrapper.findAll('.leaderboard-token-rank-row')
    expect(rows).toHaveLength(3)
    expect(rows[0].attributes('style')).toContain('--token-bar-color: rgb(217 119 6)')
    expect(rows[1].attributes('style')).toContain('--token-bar-color: rgb(5 150 105)')
    expect(rows[2].attributes('style')).toContain('--token-bar-color: rgb(37 99 235)')
    expect(wrapper.find('.leaderboard-avatar-frame-gold').exists()).toBe(false)
  })

  it('shows a round user avatar before each display name', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        ranking: [
          {
            rank: 1,
            user_id: 1,
            display_name: 'Alice',
            email_masked: 'a***@example.com',
            avatar_url: 'https://cdn.example.com/alice.png',
            actual_cost: 9,
            requests: 10,
            tokens: 1200,
            balance: 1,
            is_current_user: false,
          },
          {
            rank: 2,
            user_id: 2,
            display_name: 'Bob',
            email_masked: 'b***@example.com',
            avatar_url: null,
            actual_cost: 8,
            requests: 9,
            tokens: 900,
            balance: 2,
            is_current_user: true,
          },
        ],
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

    const avatars = wrapper.findAll('[data-testid="leaderboard-rank-avatar"]')
    expect(avatars).toHaveLength(2)
    expect(avatars[0].find('img').attributes('src')).toBe('https://cdn.example.com/alice.png')
    expect(avatars[1].text()).toBe('B')
    expect(wrapper.find('.leaderboard-token-bar-track [data-testid="leaderboard-rank-avatar"]').exists()).toBe(false)
  })
})
