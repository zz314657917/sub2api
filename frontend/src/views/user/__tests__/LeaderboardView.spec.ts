import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
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
    'leaderboard.totalCost': '总积分消费',
    'leaderboard.totalRequests': '总请求',
    'leaderboard.totalTokens': '总 Token',
    'leaderboard.periodLabel': '排行榜周期',
    'leaderboard.tokenRankingTitle': 'Token Top {count}',
    'leaderboard.tokenRankingDescription': '当前周期用量排行。',
    'leaderboard.user': '用户',
    'leaderboard.cost': '积分消费',
    'leaderboard.requests': '请求',
    'leaderboard.tokens': 'Token',
    'leaderboard.growth': '增长',
    'leaderboard.rankChange': '排名变化',
    'leaderboard.refreshing': '后台刷新中',
    'leaderboard.participationNote': '排行榜不包含企业用户和平台用户。',
    'leaderboard.rankChangeNew': '新',
    'leaderboard.rankChangeCompared.day': '较昨日',
    'leaderboard.rankChangeCompared.week': '较上周',
    'leaderboard.rankChangeCompared.month': '较上月',
    'leaderboard.rankChangeTitle.new': '{period}新上榜',
    'leaderboard.rankChangeTitle.up': '{period}名次上升 {count}',
    'leaderboard.rankChangeTitle.down': '{period}名次下降 {count}',
    'leaderboard.rankChangeTitle.same': '{period}名次持平',
    'leaderboard.inputTokensShort': '输入',
    'leaderboard.outputTokensShort': '输出',
    'leaderboard.cacheTokensShort': '缓存',
    'leaderboard.cacheRatioShort': '缓存占比',
    'leaderboard.costPerMillionShort': '积分',
    'leaderboard.recentTokenTrend.title': '最近 10 天 Token',
    'leaderboard.recentTokenTrend.unit': '每日消耗',
    'leaderboard.recentTokenTrend.tokens': 'Token',
    'leaderboard.recentTokenTrend.empty': '暂无趋势数据',
    'leaderboard.calendar.title': '每日冠军',
    'leaderboard.calendar.previousMonth': '查看上月',
    'leaderboard.calendar.currentMonth': '回到本月',
    'leaderboard.calendar.emptyDay': '{date} 暂无冠军',
    'leaderboard.calendar.weekdays.sun': '日',
    'leaderboard.calendar.weekdays.mon': '一',
    'leaderboard.calendar.weekdays.tue': '二',
    'leaderboard.calendar.weekdays.wed': '三',
    'leaderboard.calendar.weekdays.thu': '四',
    'leaderboard.calendar.weekdays.fri': '五',
    'leaderboard.calendar.weekdays.sat': '六',
    'leaderboard.balance': '积分',
    'leaderboard.rank': '排名',
    'leaderboard.myInfo': '我的信息',
    'leaderboard.record.title': '你的战绩',
    'leaderboard.record.headlineRanked': '当前第 {rank} 名，消耗 {tokens}',
    'leaderboard.record.headlineUnranked': '暂未上榜，消耗 {tokens}',
    'leaderboard.record.distanceToBoard': '距离上榜还差',
    'leaderboard.record.waitingDraw': '等待开奖',
    'leaderboard.record.waitingDrawNextRank': '领先下一名',
    'leaderboard.record.distanceToTopThree': '距离前三还差',
    'leaderboard.record.distanceToSecond': '距离第二还差',
    'leaderboard.record.distanceToFirst': '距离第一还差',
    'leaderboard.record.deity': '掌控token的神',
    'leaderboard.costEfficiencyKing': '⭐ 积分效率之王',
    'leaderboard.badges.weeklyTokenKing': '周榜 Token 最多',
    'leaderboard.badges.monthlyTokenKing': '月榜 Token 最多',
    'leaderboard.badges.totalTokenKing': '肝帝',
    'leaderboard.badges.nightOwl': '夜猫',
    'leaderboard.badges.burstTokenKing': '爆肝王',
    'leaderboard.badges.checkinKing': '打卡王',
    'leaderboard.badges.costSaver': '1M Token 积分消耗最低',
    'leaderboard.badges.costBurner': '1M Token 积分消耗最高',
    'leaderboard.generatedAt': '更新',
    'leaderboard.notRanked': '未上榜',
    'leaderboard.dailyReward.title': '上周奖励',
    'leaderboard.dailyReward.settlementDate': '统计周期',
    'leaderboard.dailyReward.threshold': '上周积分消费门槛',
    'leaderboard.dailyReward.rewardAmount': '可领积分',
    'leaderboard.dailyReward.rewardAmountHidden': '按名次发放',
    'leaderboard.dailyReward.lastWeekRank': '上周排名',
    'leaderboard.dailyReward.targetProgress': '上周奖励门槛',
    'leaderboard.dailyReward.activationThresholdPercent': '上周开启门槛',
    'leaderboard.dailyReward.weeklyRushProgress': '本周冲榜进度',
    'leaderboard.dailyReward.weeklyRushLoading': '计算中',
    'leaderboard.dailyReward.weeklyRushNoData': '暂无进度',
    'leaderboard.dailyReward.progress': '{current} / {target}',
    'leaderboard.dailyReward.progressPercent': '{percent}%',
    'leaderboard.dailyReward.progressReached': '已达成',
    'leaderboard.dailyReward.disabled': '奖励功能暂未开启',
    'leaderboard.dailyReward.settling': '上周榜结算中，{time} 后可领取',
    'leaderboard.dailyReward.thresholdNotMet': '上周积分消费未超过最低开启门槛',
    'leaderboard.dailyReward.notTopThree': '只有上周榜前三名可以领取',
    'leaderboard.dailyReward.notTopTen': '只有上周榜前 10 名参与奖励',
    'leaderboard.dailyReward.notRanked': '你上周暂无上榜积分消费',
    'leaderboard.dailyReward.zeroReward': '当前名次奖励积分为 0',
    'leaderboard.dailyReward.eligible': '你已符合领取条件',
    'leaderboard.dailyReward.redPacketReady': '你有一个待领取红包',
    'leaderboard.dailyReward.alreadyClaimed': '上周奖励已领取',
    'leaderboard.dailyReward.statusDisabled': '未开启',
    'leaderboard.dailyReward.statusSettling': '结算中',
    'leaderboard.dailyReward.statusThresholdNotMet': '门槛未达成',
    'leaderboard.dailyReward.statusNotTopThree': '未进前三',
    'leaderboard.dailyReward.statusNotTopTen': '未进前十',
    'leaderboard.dailyReward.statusNotRanked': '未上榜',
    'leaderboard.dailyReward.statusZeroReward': '无奖励',
    'leaderboard.dailyReward.statusReady': '可领取',
    'leaderboard.dailyReward.statusClaimed': '已领取',
    'leaderboard.dailyReward.claim': '领取奖励',
    'leaderboard.dailyReward.redPacketClaim': '领取红包',
    'leaderboard.dailyReward.claiming': '领取中...',
    'leaderboard.dailyReward.claimed': '已领取',
    'leaderboard.dailyReward.claimFailed': '领取排行榜奖励失败',
    'leaderboard.dailyReward.rankLabel': '第 {rank} 名',
    'leaderboard.dailyReward.rankReward': '第 {rank} 名奖励',
    'leaderboard.dailyReward.lastWeekTopUsersTitle': '上周前 10 Token 消耗',
    'leaderboard.dailyReward.lastWeekRank1': '上周第一名',
    'leaderboard.dailyReward.lastWeekRank2': '上周第二名',
    'leaderboard.dailyReward.lastWeekRank3': '上周第三名',
    'leaderboard.dailyReward.lastWeekRankLabel': '上周第 {rank} 名',
    'leaderboard.dailyReward.noTopUser': '暂无上榜',
    'leaderboard.dailyReward.topUserTokens': '{tokens} tokens',
    'leaderboard.dailyReward.redPacketRange': '红包范围',
    'leaderboard.dailyReward.redPacketPool': '总池 {amount}',
    'leaderboard.dailyReward.redPacketPendingAmount': '待抽取金额',
    'leaderboard.dailyReward.redPacketClaimedAmount': '已抽金额',
    'leaderboard.dailyReward.redPacketPending': '点击后随机抽取',
    'leaderboard.dailyReward.lotteryRule': '玩法说明',
    'leaderboard.dailyReward.lotteryRuleText': '上榜用户随机抽1位赠送积分',
    'leaderboard.dailyReward.lotteryDrawTime': '开奖时间',
    'leaderboard.dailyReward.lotteryPrize': '抽奖金额',
    'leaderboard.dailyReward.lotteryResult': '开奖结果',
    'leaderboard.dailyReward.lastWeekLotteryWinner': '上周中奖用户',
    'leaderboard.dailyReward.lotteryPending': '等待开奖',
    'leaderboard.dailyReward.lotteryCountdownHours': '{hours}小时后开奖',
    'leaderboard.dailyReward.lotteryCountdownMinutes': '{minutes}分钟后开奖',
    'leaderboard.dailyReward.lotteryWinner': '中奖用户：{name}',
    'leaderboard.dailyReward.lotteryWon': '你已中奖 {amount}',
    'leaderboard.dailyReward.lotteryNotWon': '未中奖',
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
        input_tokens: 620,
        output_tokens: 180,
        cache_creation_tokens: 60,
        cache_read_tokens: 40,
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
    daily_champions: [],
    model_ranking: [
      {
        rank: 1,
        model: 'gpt-5.5',
        requests: 10,
        input_tokens: 700,
        output_tokens: 200,
        tokens: 1000,
        growth_percent: -77.7,
        rank_change: 1,
      },
      {
        rank: 2,
        model: 'claude-opus-4-8',
        requests: 4,
        input_tokens: 300,
        output_tokens: 100,
        tokens: 400,
        growth_percent: -87.3,
        rank_change: -1,
      },
    ],
    total_models: 2,
    daily_rewards: {
      reward_date: '2026-05-06',
      settlement_timezone: 'Asia/Shanghai',
      settlement_ready: true,
      claim_available_at: '2026-05-07T00:30:00+08:00',
      enabled: true,
      reward_mode: 'red_packet',
      red_packet_pool_amount: 20,
      red_packet_min_amount: 1,
      red_packet_max_amount: 5,
      min_total_actual_cost: 100,
      yesterday_total_actual_cost: 120,
      threshold_met: true,
      rewards: [
        { rank: 1, amount: 0 },
        { rank: 2, amount: 0 },
        { rank: 3, amount: 0 },
      ],
      top_users: [
        { rank: 1, display_name: 'A***e', email_masked: 'a***e@example.com', avatar_url: 'https://cdn.example.com/a.png', tokens: 9_000_000 },
        { rank: 2, display_name: 'B*b', email_masked: 'b***b@example.com', tokens: 8_000_000 },
        { rank: 3, display_name: 'C***l', email_masked: 'c***l@example.com', tokens: 7_000_000 },
        { rank: 4, display_name: 'D***n', email_masked: 'd***n@example.com', tokens: 6_000_000 },
        { rank: 5, display_name: 'E***a', email_masked: 'e***a@example.com', tokens: 5_000_000 },
        { rank: 6, display_name: 'F***y', email_masked: 'f***y@example.com', tokens: 4_000_000 },
        { rank: 7, display_name: 'G***o', email_masked: 'g***o@example.com', tokens: 3_000_000 },
        { rank: 8, display_name: 'H***i', email_masked: 'h***i@example.com', tokens: 2_000_000 },
        { rank: 9, display_name: 'I***n', email_masked: 'i***n@example.com', tokens: 1_000_000 },
        { rank: 10, display_name: 'J***e', email_masked: 'j***e@example.com', tokens: 500_000 },
      ],
      current_user_rank: 0,
      current_user_reward_amount: 0,
      can_claim: false,
      claimed: false,
      reason: 'not_ranked',
    },
    ...overrides,
  }
}

describe('LeaderboardView', () => {
  beforeEach(() => {
    window.sessionStorage.clear()
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
    getDashboardLeaderboard.mockReturnValueOnce(new Promise(() => {
      // Background weekly summary remains pending so the explicit tab switch cannot hydrate from cache.
    }))
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

  it('shows cached leaderboard rows with a background refresh indicator', async () => {
    window.sessionStorage.setItem('sub2api:user-leaderboard:v1:day:10', JSON.stringify({
      savedAt: Date.now(),
      data: makeResponse({
        ranking: [
          {
            rank: 1,
            user_id: 88,
            display_name: 'Cached User',
            email_masked: 'c***@example.com',
            avatar_url: null,
            actual_cost: 1,
            requests: 1,
            input_tokens: 90,
            output_tokens: 10,
            cache_creation_tokens: 0,
            cache_read_tokens: 0,
            tokens: 100,
            cost_per_1m_tokens: 10000,
            balance: 0,
            is_current_user: false,
          },
        ],
      }),
    }))
    let resolveRefresh: (value: ReturnType<typeof makeResponse>) => void = () => {}
    getDashboardLeaderboard.mockReturnValueOnce(new Promise((resolve) => {
      resolveRefresh = resolve
    }))
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await nextTick()

    expect(wrapper.text()).toContain('Cached User')
    expect(wrapper.get('[data-testid="leaderboard-refreshing"]').text()).toBe('后台刷新中')

    resolveRefresh(makeResponse({
      ranking: [
        {
          rank: 1,
          user_id: 99,
          display_name: 'Fresh User',
          email_masked: 'f***@example.com',
          avatar_url: null,
          actual_cost: 2,
          requests: 2,
          input_tokens: 180,
          output_tokens: 20,
          cache_creation_tokens: 0,
          cache_read_tokens: 0,
          tokens: 200,
          cost_per_1m_tokens: 10000,
          balance: 0,
          is_current_user: false,
        },
      ],
    }))
    await flushPromises()

    expect(wrapper.text()).toContain('Fresh User')
    expect(wrapper.text()).not.toContain('Cached User')
    expect(wrapper.find('[data-testid="leaderboard-refreshing"]').exists()).toBe(false)
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
          {
            rank: 2,
            user_id: 3,
            display_name: 'Cora',
            email_masked: 'c***@example.com',
            avatar_url: null,
            actual_cost: 3.1,
            requests: 12,
            tokens: 1800,
            balance: 14,
            is_current_user: false,
          },
          {
            rank: 3,
            user_id: 4,
            display_name: 'Dana',
            email_masked: 'd***@example.com',
            avatar_url: null,
            actual_cost: 2.7,
            requests: 10,
            tokens: 660_000_000,
            balance: 9,
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

    const record = wrapper.get('[data-testid="leaderboard-my-record"]')
    const banner = wrapper.get('[data-testid="leaderboard-ranking-illustration"]')
    expect(banner.find('[data-testid="leaderboard-my-record"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="leaderboard-my-info"]').find('[data-testid="leaderboard-my-record"]').exists()).toBe(false)
    expect(record.text()).toContain('你的战绩')
    expect(record.text()).toContain('当前第 28 名，消耗 <0.1M')
    expect(record.text()).toContain('等待开奖')
    expect(record.text()).not.toContain('下一名距离你')
    expect(record.text()).not.toContain('距离前三还差')
    expect(record.text()).not.toContain('Me')
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

    const record = wrapper.get('[data-testid="leaderboard-my-record"]')
    expect(record.text()).toContain('你的战绩')
    expect(record.text()).toContain('当前第 1 名，消耗 <0.1M')
    expect(record.text()).toContain('等待开奖')
    expect(record.text()).not.toContain('掌控token的神')
    expect(record.text()).not.toContain('Alice')
    const banner = wrapper.get('[data-testid="leaderboard-ranking-illustration"]')
    expect(banner.get('img').attributes('alt')).toBe('')
    expect(banner.find('[data-testid="leaderboard-my-record"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="leaderboard-thursday-banner"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).not.toContain('$11.00')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).not.toContain('余额')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').findAll('[data-testid="leaderboard-my-token"]')).toHaveLength(0)
  })

  it('shows the distance to enter the visible board when current user is unranked', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        ranking: Array.from({ length: 10 }, (_, index) => ({
          rank: index + 1,
          user_id: index + 1,
          display_name: `User ${index + 1}`,
          email_masked: `u***${index + 1}@example.com`,
          avatar_url: null,
          actual_cost: 10 - index,
          requests: 10 + index,
          tokens: 2_000 - index * 100,
          balance: 1,
          is_current_user: false,
        })),
        current_user_entry: {
          rank: 0,
          user_id: 99,
          display_name: 'Hidden Me',
          email_masked: 'h***@example.com',
          avatar_url: null,
          actual_cost: 0.2,
          requests: 1,
          tokens: 718,
          balance: 1,
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

    const record = wrapper.get('[data-testid="leaderboard-my-record"]')
    expect(record.text()).toContain('暂未上榜，消耗 <0.1M')
    expect(record.text()).toContain('距离上榜还差')
    expect(record.text()).toContain('383')
    expect(record.text()).not.toContain('Hidden Me')
  })

  it('shows the next ranked user token gap after the current user is ranked', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        ranking: [
          {
            rank: 1,
            user_id: 1,
            display_name: 'First',
            email_masked: 'f***@example.com',
            avatar_url: null,
            actual_cost: 3,
            requests: 10,
            tokens: 2_000,
            balance: 1,
            is_current_user: false,
          },
          {
            rank: 2,
            user_id: 99,
            display_name: 'Me',
            email_masked: 'm***@example.com',
            avatar_url: null,
            actual_cost: 2,
            requests: 8,
            tokens: 1_500,
            balance: 1,
            is_current_user: true,
          },
          {
            rank: 3,
            user_id: 3,
            display_name: 'Third',
            email_masked: 't***@example.com',
            avatar_url: null,
            actual_cost: 1,
            requests: 6,
            tokens: 1_200,
            balance: 1,
            is_current_user: false,
          },
        ],
        current_user_entry: {
          rank: 2,
          user_id: 99,
          display_name: 'Me',
          email_masked: 'm***@example.com',
          avatar_url: null,
          actual_cost: 2,
          requests: 8,
          tokens: 1_500,
          balance: 1,
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

    const record = wrapper.get('[data-testid="leaderboard-my-record"]')
    expect(record.text()).toContain('领先下一名')
    expect(record.text()).not.toContain('等待开奖，下一名距离你')
    expect(record.text()).toContain('301')
    expect(record.text()).toContain('token')
    expect(record.text()).not.toContain('距离第一还差')
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
            cache_creation_tokens: 60,
            cache_read_tokens: 40,
            tokens: 1000,
            cost_per_1m_tokens: 10000,
            balance: 1,
            rank_change: 1,
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
    expect(wrapper.text()).not.toContain('⭐ 积分效率之王')
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
            cache_creation_tokens: 60,
            cache_read_tokens: 40,
            tokens: 1000,
            cost_per_1m_tokens: 10000,
            balance: 1,
            rank_change: 1,
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
            cache_creation_tokens: 120,
            cache_read_tokens: 80,
            tokens: 2000,
            cost_per_1m_tokens: 500,
            balance: 2,
            rank_change: -1,
            is_current_user: false,
          },
          {
            rank: 3,
            user_id: 3,
            display_name: 'Newcomer',
            email_masked: 'n***@example.com',
            avatar_url: null,
            actual_cost: 0.5,
            requests: 5,
            input_tokens: 320,
            output_tokens: 70,
            cache_creation_tokens: 5,
            cache_read_tokens: 5,
            tokens: 400,
            cost_per_1m_tokens: 1250,
            balance: 3,
            rank_new: true,
            is_current_user: false,
          },
        ],
        current_user_entry: null,
        total_tokens: 3400,
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
    expect(ranking.text()).not.toContain('Token Top 2')
    expect(ranking.text()).toContain('更新 16:00:00')
    expect(ranking.get('[data-testid="leaderboard-participation-note"]').text()).toBe('排行榜不包含企业用户和平台用户。')
    expect(ranking.get('.leaderboard-token-rank-index').text()).toBe('1')
    expect(ranking.text()).not.toContain('No.1')
    expect(ranking.text()).not.toContain('#1')
    expect(ranking.text()).toContain('Pricey')
    expect(ranking.text()).toContain('Efficient')
    const rankChanges = ranking.findAll('[data-testid="leaderboard-rank-change"]')
    expect(rankChanges.map((node) => node.text())).toEqual(['↑+1', '↓-1', '新'])
    expect(rankChanges.map((node) => node.attributes('title'))).toEqual(['较昨日名次上升 1', '较昨日名次下降 1', '较昨日新上榜'])
    expect(ranking.text()).toContain('2,000')
    expect(ranking.findAll('[data-testid="leaderboard-token-bar-fill"]')).toHaveLength(3)
    expect(ranking.findAll('[data-testid="leaderboard-token-segment-input"]')).toHaveLength(3)
    expect(ranking.findAll('[data-testid="leaderboard-token-segment-output"]')).toHaveLength(3)
    expect(ranking.findAll('[data-testid="leaderboard-token-segment-cache"]')).toHaveLength(3)
    const tokenBars = ranking.findAll('.leaderboard-token-bar-track')
    expect(tokenBars[0].attributes('title')).toBeUndefined()
    expect(tokenBars[0].attributes('aria-label')).toBe('输入 700 / 输出 200 / 缓存 100 (缓存占比 10.0%) / 积分 ✪ 10,000.00 / 1M Token')
    expect(tokenBars[1].attributes('title')).toBeUndefined()
    expect(tokenBars[1].attributes('aria-label')).toBe('输入 1,500 / 输出 300 / 缓存 200 (缓存占比 10.0%) / 积分 ✪ 500.00 / 1M Token')
    await tokenBars[0].trigger('mouseenter')
    const tokenTooltip = document.body.querySelector('[data-testid="leaderboard-token-tooltip"]')
    expect(tokenTooltip?.textContent).toContain('2026年5月7日')
    expect(tokenTooltip?.textContent).toContain('Pricey')
    expect(tokenTooltip?.textContent).toContain('1,000 tokens')
    const tokenTooltipRows = Array.from(tokenTooltip?.querySelectorAll('.leaderboard-token-tooltip-row') ?? [])
      .map((row) => Array.from(row.querySelectorAll('span')).map((cell) => cell.textContent?.trim()).filter(Boolean))
    expect(tokenTooltip?.querySelector('.leaderboard-token-tooltip-table')?.getAttribute('role')).toBe('table')
    expect(tokenTooltipRows).toEqual([
      ['输入', '700'],
      ['输出', '200'],
      ['缓存', '100', '缓存占比 10.0%'],
      ['积分', '✪ 10,000.00', '/ 1M Token'],
    ])
    expect(ranking.findAll('.leaderboard-token-rank-row')[1].attributes('style')).toContain('--token-bar-width: 84%')
    expect(ranking.findAll('.leaderboard-token-rank-row')[1].attributes('style')).toContain('--token-input-width: 75.0%')
    expect(ranking.findAll('.leaderboard-token-rank-row')[1].attributes('style')).toContain('--token-output-width: 15.0%')
    expect(ranking.findAll('.leaderboard-token-rank-row')[1].attributes('style')).toContain('--token-cache-width: 10.0%')
    expect(wrapper.find('[data-testid="leaderboard-cost-efficiency-summary"]').exists()).toBe(false)
  })

  it('keeps period tabs inside the token ranking card and hides the model ranking switch', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        period: 'day',
        model_ranking: [
          {
            rank: 1,
            model: 'gpt-5.5',
            requests: 52,
            input_tokens: 6000000000,
            output_tokens: 130000000,
            tokens: 6130000000,
            growth_percent: 28.4,
            rank_change: 1,
          },
          {
            rank: 2,
            model: 'gpt-5.4',
            requests: 10826,
            input_tokens: 1000000000,
            output_tokens: 230000000,
            tokens: 1230000000,
            growth_percent: -12.8,
            rank_change: -1,
          },
        ],
        total_models: 2,
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
    const rankingTabs = ranking.findAll('.leaderboard-ranking-switch-button')
    expect(rankingTabs.map((button) => button.text())).toEqual(['日榜', '周榜', '月榜', '总榜'])
    expect(rankingTabs[0].attributes('aria-selected')).toBe('true')
    expect(wrapper.text()).not.toContain('模型榜')
    expect(wrapper.find('[data-testid="leaderboard-model-ranking"]').exists()).toBe(false)
    expect(ranking.text()).not.toContain('gpt-5.5')

    await rankingTabs[1].trigger('click')
    await flushPromises()

    expect(getDashboardLeaderboard).toHaveBeenLastCalledWith({ period: 'week', limit: 10 })
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

  it('renders the daily champion calendar with avatar fallback and native tooltip copy', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        generated_at: '2026-07-06T00:00:00Z',
        daily_champions: [
          {
            date: '2026-06-10',
            user_id: 10,
            display_name: '155***4@qq.com',
            email_masked: '155***4@qq.com',
            avatar_url: 'https://cdn.example.com/u10.png',
            tokens: 48_163_000,
          },
          {
            date: '2026-07-06',
            user_id: 11,
            display_name: '南瓜号',
            email_masked: 'n***@example.com',
            avatar_url: null,
            tokens: 12_000,
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

    const calendar = wrapper.get('[data-testid="leaderboard-daily-champions-calendar"]')
    expect(calendar.text()).toContain('每日冠军')
    expect(calendar.text()).toContain('2026年6月 / 2026年7月')
    expect(calendar.text()).not.toContain('开荒以来')
    expect(calendar.text()).toContain('2026年7月')
    expect(calendar.text()).toContain('2026年6月')
    expect(calendar.find('[data-testid="leaderboard-calendar-prev"]').exists()).toBe(false)
    expect(calendar.find('[data-testid="leaderboard-calendar-current"]').exists()).toBe(false)
    expect(calendar.findAll('[data-testid="leaderboard-calendar-day"]')).toHaveLength(36)
    expect(calendar.findAll('[data-testid="leaderboard-calendar-placeholder"]')).toHaveLength(0)
    const julyFutureDay = calendar.findAll('[data-testid="leaderboard-calendar-day"]')
      .find((day) => day.attributes('aria-label') === '2026年7月7日 暂无冠军')
    expect(julyFutureDay).toBeUndefined()

    const championDays = calendar.findAll('[data-testid="leaderboard-calendar-day"]')
      .filter((day) => day.attributes('aria-label')?.includes('tokens'))
    expect(championDays).toHaveLength(2)

    const julyChampion = championDays.find((day) => day.attributes('aria-label')?.includes('2026年7月6日'))
    expect(julyChampion?.attributes('title')).toBeUndefined()
    expect(julyChampion?.attributes('aria-label')).toBe('2026年7月6日\n南瓜号\n1.2万 tokens')
    expect(julyChampion?.find('.leaderboard-calendar-day-number').text()).toBe('6')
    expect(julyChampion?.find('[data-testid="leaderboard-calendar-avatar-frame"] .leaderboard-calendar-day-number').exists()).toBe(false)
    expect(julyChampion?.find('[data-testid="leaderboard-calendar-avatar"]').text()).toBe('南')
    expect(julyChampion?.find('.leaderboard-calendar-token-label').text()).toBe('<0.1M')

    const emptyDay = calendar.findAll('[data-testid="leaderboard-calendar-day"]')
      .find((day) => day.attributes('aria-label') === '2026年7月1日 暂无冠军')
    expect(emptyDay?.attributes('title')).toBeUndefined()

    const juneChampion = calendar.findAll('[data-testid="leaderboard-calendar-day"]')
      .find((day) => day.attributes('aria-label')?.includes('2026年6月10日'))
    expect(juneChampion?.attributes('title')).toBeUndefined()
    expect(juneChampion?.attributes('aria-label')).toBe('2026年6月10日\n155***4@qq.com\n4816.3万 tokens')
    expect(juneChampion?.find('.leaderboard-calendar-day-number').text()).toBe('10')
    expect(juneChampion?.find('[data-testid="leaderboard-calendar-avatar-frame"]').exists()).toBe(true)
    expect(juneChampion?.find('[data-testid="leaderboard-calendar-avatar"] img').attributes('src')).toBe('https://cdn.example.com/u10.png')
    expect(juneChampion?.find('.leaderboard-calendar-token-label').text()).toBe('48.2M')
    await juneChampion?.trigger('mouseenter')
    await nextTick()
    const tooltip = document.body.querySelector('[data-testid="leaderboard-calendar-tooltip"]')
    expect(tooltip?.textContent).toContain('2026年6月10日')
    expect(tooltip?.textContent).toContain('155***4@qq.com')
    expect(tooltip?.textContent).toContain('4816.3万 tokens')
    expect(tooltip?.textContent).toContain('当日冠军')
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

  it('keeps personal badges out of the record card while rank titles still collapse on rows', async () => {
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

    const record = wrapper.get('[data-testid="leaderboard-my-record"]')
    expect(record.text()).toContain('当前第 16 名')
    expect(record.text()).not.toContain('Decorated')
    expect(wrapper.findAll('[data-testid="leaderboard-my-badge-icon"]')).toHaveLength(0)

    const titles = wrapper.findAll('[data-testid="leaderboard-rank-title"]')
    expect(titles.map((title) => title.text())).toEqual(['周榜王', '月榜王'])
    expect(titles.map((title) => title.attributes('data-badge'))).toEqual([
      'weekly_token_king',
      'monthly_token_king',
    ])
    expect(wrapper.find('.leaderboard-token-title-more').text()).toBe('+4')
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
    expect(wrapper.get('[data-testid="leaderboard-participation-note"]').text()).toBe('排行榜不包含企业用户和平台用户。')
  })

  it('always renders last week top 10 token usage and hides claim controls when rewards are disabled', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        daily_rewards: {
          reward_date: '2026-05-04~2026-05-10',
          settlement_timezone: 'Asia/Shanghai',
          settlement_ready: true,
          claim_available_at: '2026-05-07T00:30:00+08:00',
          enabled: false,
          reward_mode: 'disabled',
          min_total_actual_cost: 100,
          yesterday_total_actual_cost: 80,
          threshold_met: false,
          rewards: [],
          top_users: [
            ...Array.from({ length: 5 }, (_, index) => ({
              rank: index + 1,
              display_name: `Top ${index + 1}`,
              email_masked: `top${index + 1}@example.com`,
              avatar_url: index === 0 ? 'https://cdn.example.com/a.png' : undefined,
              tokens: 10_000_000 - index * 500_000,
            })),
          ],
          current_user_rank: 1,
          current_user_reward_amount: 0,
          can_claim: false,
          claimed: false,
          reason: 'disabled',
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
    expect(wrapper.find('[data-testid="leaderboard-thursday-banner"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="leaderboard-my-record"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('上周奖励')
    expect(wrapper.text()).not.toContain('统计周期')
    expect(wrapper.find('[data-testid="leaderboard-daily-reward-status"]').exists()).toBe(false)
    const weeklyTop10 = wrapper.get('[data-testid="leaderboard-weekly-top10"]')
    expect(weeklyTop10.text()).toContain('上周前 10 Token 消耗')
    expect(weeklyTop10.text()).toContain('05-04~05-10')
    expect(weeklyTop10.text()).not.toContain('2026-05-04')
    expect(weeklyTop10.text()).not.toContain('2026-05-10')
    expect(weeklyTop10.get('[data-testid="leaderboard-weekly-top10-scroll"]').exists()).toBe(true)
    expect(weeklyTop10.get('.leaderboard-weekly-winner-rank').text()).toBe('1')
    expect(weeklyTop10.get('[data-testid="leaderboard-weekly-winner-avatar"] img').attributes('src')).toBe('https://cdn.example.com/a.png')
    expect(weeklyTop10.element.innerHTML.indexOf('leaderboard-weekly-winner-avatar')).toBeLessThan(
      weeklyTop10.element.innerHTML.indexOf('leaderboard-weekly-winner-name'),
    )
    const weeklyRows = weeklyTop10.findAll('.leaderboard-weekly-winner-row')
    expect(weeklyRows).toHaveLength(10)
    expect(weeklyTop10.findAll('.leaderboard-weekly-winner-row--placeholder')).toHaveLength(5)
    expect(weeklyTop10.text()).not.toContain('No.1')
    expect(weeklyTop10.text()).not.toContain('#1')
    expect(weeklyTop10.text()).not.toContain('上周第一名')
    expect(weeklyRows[4].text()).toContain('5')
    expect(weeklyRows[4].text()).toContain('Top 5')
    expect(weeklyRows[9].text()).toContain('10')
    expect(weeklyRows[9].text()).toContain('暂无上榜')
    expect(wrapper.text()).toContain('1000万 tokens')
    expect(wrapper.text()).not.toContain('第 1 名奖励')
    expect(wrapper.text()).not.toContain('上周前三')
    expect(wrapper.find('[data-testid="leaderboard-daily-reward-claim"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="leaderboard-red-packet-reward"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="leaderboard-lottery-reward"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="leaderboard-threshold-progress"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('上周排名')
    expect(wrapper.text()).not.toContain('上周开启门槛')
    expect(wrapper.text()).not.toContain('本周冲榜进度')
  })

  it('shows reached text for completed last-week threshold and weekly rush distance', async () => {
    const weeklyResponse = makeResponse({
      period: 'week',
      ranking: [
        {
          rank: 1,
          user_id: 11,
          display_name: 'Top One',
          email_masked: 't***@example.com',
          avatar_url: null,
          actual_cost: 9,
          requests: 20,
          tokens: 2_000_000,
          balance: 3,
          is_current_user: false,
        },
        {
          rank: 2,
          user_id: 12,
          display_name: 'Second',
          email_masked: 's***@example.com',
          avatar_url: null,
          actual_cost: 6,
          requests: 18,
          tokens: 1_400_000,
          balance: 3,
          is_current_user: false,
        },
        {
          rank: 3,
          user_id: 13,
          display_name: 'Third',
          email_masked: 't***@example.com',
          avatar_url: null,
          actual_cost: 5,
          requests: 16,
          tokens: 800_000,
          balance: 3,
          is_current_user: false,
        },
      ],
      current_user_entry: {
        rank: 9,
        user_id: 99,
        display_name: 'Me',
        email_masked: 'm***@example.com',
        avatar_url: null,
        actual_cost: 2,
        requests: 8,
        tokens: 600_000,
        balance: 3,
        is_current_user: true,
      },
    })
    getDashboardLeaderboard
      .mockResolvedValueOnce(
        makeResponse({
          daily_rewards: {
            reward_date: '2026-05-06',
            settlement_timezone: 'Asia/Shanghai',
            settlement_ready: true,
            claim_available_at: '2026-05-07T00:30:00+08:00',
            enabled: true,
            reward_mode: 'red_packet',
            red_packet_pool_amount: 30,
            red_packet_min_amount: 1,
            red_packet_max_amount: 8,
            min_total_actual_cost: 100,
            yesterday_total_actual_cost: 120,
            threshold_met: true,
            rewards: [],
            current_user_rank: 2,
            current_user_reward_amount: 3,
            can_claim: true,
            claimed: false,
            reason: 'eligible',
          },
        })
      )
      .mockResolvedValueOnce(weeklyResponse)
    const { default: LeaderboardView } = await import('../LeaderboardView.vue')

    const wrapper = mount(LeaderboardView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('红包范围')
    expect(wrapper.text()).toContain('✪ 1.00 - ✪ 8.00')
    expect(wrapper.text()).toContain('上周开启门槛')
    expect(wrapper.text()).toContain('100%')
    expect(wrapper.get('[data-testid="leaderboard-threshold-progress"]').exists()).toBe(true)
    expect(wrapper.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('100')
    expect(wrapper.text()).toContain('本周冲榜进度')
    expect(wrapper.text()).toContain('距离前三还差 20万 token')
  })

  it('shows development preview top 10 users when the reward payload has no winners', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        daily_rewards: {
          reward_date: '2026-05-06',
          settlement_timezone: 'Asia/Shanghai',
          settlement_ready: true,
          claim_available_at: '2026-05-07T00:30:00+08:00',
          enabled: true,
          reward_mode: 'red_packet',
          red_packet_pool_amount: 20,
          red_packet_min_amount: 1,
          red_packet_max_amount: 5,
          min_total_actual_cost: 100,
          yesterday_total_actual_cost: 0,
          threshold_met: false,
          rewards: [],
          top_users: [],
          current_user_rank: 0,
          current_user_reward_amount: 0,
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

    expect(wrapper.find('[data-testid="leaderboard-weekly-top10"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('上周前 10 Token 消耗')
    expect(wrapper.text()).toContain('落***尘')
    expect(wrapper.text()).toContain('138****5678')
    expect(wrapper.text()).toContain('t***d@example.com')
    expect(wrapper.text()).toContain('last***top')
    expect(wrapper.text()).not.toContain('暂无上榜')
  })

  it('shows settling state before red packet rewards can be claimed', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        daily_rewards: {
          reward_date: '2026-05-06',
          settlement_timezone: 'Asia/Shanghai',
          settlement_ready: false,
          claim_available_at: '2026-05-07T00:30:00+08:00',
          enabled: true,
          reward_mode: 'red_packet',
          red_packet_pool_amount: 20,
          red_packet_min_amount: 1,
          red_packet_max_amount: 5,
          min_total_actual_cost: 100,
          yesterday_total_actual_cost: 120,
          threshold_met: true,
          rewards: [],
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

    const settlingStatus = wrapper.get('[data-testid="leaderboard-daily-reward-status"]')
    expect(settlingStatus.text()).toContain('结算中')
    expect(settlingStatus.attributes('title')).toContain('结算中')
    expect(wrapper.get('[data-testid="leaderboard-daily-reward-claim"]').attributes('disabled')).toBeDefined()
  })

  it('claims an eligible weekly reward and shows claimed state', async () => {
    const claimRewards = {
      reward_date: '2026-05-06',
      settlement_timezone: 'Asia/Shanghai',
      settlement_ready: true,
      claim_available_at: '2026-05-07T00:30:00+08:00',
      enabled: true,
      reward_mode: 'red_packet',
      red_packet_pool_amount: 20,
      red_packet_min_amount: 1,
      red_packet_max_amount: 5,
      min_total_actual_cost: 100,
      yesterday_total_actual_cost: 120,
      threshold_met: true,
      rewards: [],
      current_user_rank: 1,
      current_user_reward_amount: 0,
      can_claim: true,
      claimed: false,
      reason: 'eligible',
    }
    getDashboardLeaderboard.mockResolvedValue(makeResponse({ daily_rewards: claimRewards }))
    claimDashboardLeaderboardDailyReward.mockResolvedValue({
      red_packet_amount: 4.25,
      daily_rewards: {
        ...claimRewards,
        current_user_reward_amount: 4.25,
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
    const claimedStatus = wrapper.get('[data-testid="leaderboard-daily-reward-status"]')
    expect(claimedStatus.text()).toContain('已领取')
    expect(claimedStatus.attributes('title')).toBe('上周奖励已领取')
    expect(wrapper.get('[data-testid="leaderboard-daily-reward-claim"]').text()).toContain('已领取')
    expect(wrapper.get('[data-testid="leaderboard-red-packet-reward"]').text()).toContain('已抽金额')
    expect(wrapper.get('[data-testid="leaderboard-red-packet-reward"]').text()).toContain('4.25')
    expect(wrapper.find('[data-testid="leaderboard-my-info"]').text()).not.toContain('$16.00')
  })

  it('renders lottery reward draw time and not-won result without a claim button', async () => {
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        daily_rewards: {
          reward_date: '2026-05-06',
          settlement_timezone: 'Asia/Shanghai',
          settlement_ready: true,
          claim_available_at: '2026-05-07T00:30:00+08:00',
          enabled: true,
          reward_mode: 'lottery',
          lottery_amount: 12,
          lottery_cron: '0 10 * * 1',
          lottery_draw_at: '2026-05-11T10:00:00+08:00',
          lottery_winner_rank: 2,
          lottery_winner_user_id: 22,
          lottery_winner_display_name: 'Lucky User',
          min_total_actual_cost: 100,
          yesterday_total_actual_cost: 120,
          threshold_met: true,
          rewards: [],
          current_user_rank: 4,
          current_user_reward_amount: 0,
          can_claim: false,
          claimed: false,
          reason: 'not_won',
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

    const lottery = wrapper.get('[data-testid="leaderboard-lottery-reward"]')
    expect(lottery.text()).toContain('玩法说明')
    expect(lottery.text()).toContain('上榜用户随机抽1位赠送积分')
    expect(lottery.text()).toContain('开奖时间')
    expect(lottery.text()).toContain('2026')
    expect(lottery.text()).toContain('抽奖金额')
    expect(lottery.text()).toContain('12.00')
    expect(lottery.text()).toContain('开奖结果')
    expect(lottery.text()).toContain('未中奖')
    expect(lottery.text()).toContain('上周中奖用户')
    expect(lottery.text()).toContain('中奖用户：Lucky User')
    expect(wrapper.find('[data-testid="leaderboard-daily-reward-claim"]').exists()).toBe(false)
  })

  it('renders lottery pending result as a draw countdown in hours and minutes', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-05-11T08:00:00+08:00'))
    getDashboardLeaderboard.mockResolvedValue(
      makeResponse({
        daily_rewards: {
          reward_date: '2026-05-06',
          settlement_timezone: 'Asia/Shanghai',
          settlement_ready: false,
          claim_available_at: '2026-05-07T00:30:00+08:00',
          enabled: true,
          reward_mode: 'lottery',
          lottery_amount: 12,
          lottery_cron: '0 10 * * 1',
          lottery_draw_at: '2026-05-11T10:00:00+08:00',
          min_total_actual_cost: 100,
          yesterday_total_actual_cost: 120,
          threshold_met: true,
          rewards: [],
          current_user_rank: 4,
          current_user_reward_amount: 0,
          can_claim: false,
          claimed: false,
          reason: 'lottery_pending',
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

    const lottery = wrapper.get('[data-testid="leaderboard-lottery-reward"]')
    expect(lottery.text()).toContain('2小时后开奖')
    expect(lottery.text()).not.toContain('等待开奖')

    vi.setSystemTime(new Date('2026-05-11T09:14:00+08:00'))
    await vi.advanceTimersByTimeAsync(60_000)
    await flushPromises()

    expect(lottery.text()).toContain('45分钟后开奖')
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
    expect(rows[0].attributes('style')).toContain('--token-bar-color: rgb(196 111 80)')
    expect(rows[1].attributes('style')).toContain('--token-bar-color: rgb(95 143 129)')
    expect(rows[2].attributes('style')).toContain('--token-bar-color: rgb(132 118 98)')
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
