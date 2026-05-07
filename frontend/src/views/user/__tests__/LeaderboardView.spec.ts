import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const { getDashboardLeaderboard } = vi.hoisted(() => ({
  getDashboardLeaderboard: vi.fn(),
}))

vi.mock('@/api', () => ({
  usageAPI: {
    getDashboardLeaderboard,
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
    'leaderboard.rank': '排名',
    'leaderboard.generatedAt': '生成时间',
    'common.loading': '加载中...',
    'common.refresh': '刷新',
  }

  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
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
        is_current_user: true,
      },
    ],
    current_user_entry: null,
    ...overrides,
  }
}

describe('LeaderboardView', () => {
  beforeEach(() => {
    getDashboardLeaderboard.mockReset()
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

    expect(wrapper.text()).toContain('我的排名')
    expect(wrapper.text()).toContain('#28')
    expect(wrapper.text()).toContain('Me')
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
})
