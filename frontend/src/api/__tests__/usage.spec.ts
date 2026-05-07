import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get } = vi.hoisted(() => ({
  get: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
  },
}))

import { getDashboardLeaderboard } from '@/api/usage'

describe('usage api', () => {
  beforeEach(() => {
    get.mockReset()
    get.mockResolvedValue({ data: { ranking: [] } })
  })

  it('loads user dashboard leaderboard with period and limit params', async () => {
    await getDashboardLeaderboard({ period: 'week', limit: 20 })

    expect(get).toHaveBeenCalledWith('/usage/dashboard/leaderboard', {
      params: {
        period: 'week',
        limit: 20,
      },
    })
  })
})
