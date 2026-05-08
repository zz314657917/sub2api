import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    post,
  },
}))

import { claimDashboardLeaderboardDailyReward, getDashboardLeaderboard } from '@/api/usage'

describe('usage api', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    get.mockResolvedValue({ data: { ranking: [] } })
    post.mockResolvedValue({ data: { daily_rewards: null, claimed_amount: 0 } })
  })

  it('loads user dashboard leaderboard with period and limit params', async () => {
    await getDashboardLeaderboard({ period: 'week', limit: 10 })

    expect(get).toHaveBeenCalledWith('/usage/dashboard/leaderboard', {
      params: {
        period: 'week',
        limit: 10,
      },
    })
  })

  it('claims leaderboard daily reward', async () => {
    await claimDashboardLeaderboardDailyReward()

    expect(post).toHaveBeenCalledWith('/usage/dashboard/leaderboard/daily-reward/claim')
  })
})
