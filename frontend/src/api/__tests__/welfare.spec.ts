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

import {
  claimWelfareDailyCheckin,
  claimWelfareDailyCheckinMilestone,
  getWelfareDailyCheckin,
  getWelfareOverview,
} from '@/api/welfare'

describe('welfare api', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    get.mockResolvedValue({ data: {} })
    post.mockResolvedValue({ data: {} })
  })

  it('loads welfare overview', async () => {
    await getWelfareOverview()

    expect(get).toHaveBeenCalledWith('/user/welfare/overview', {
      signal: undefined,
    })
  })

  it('passes through new user trial status fields', async () => {
    get.mockResolvedValueOnce({
      data: {
        enabled: true,
        modules: {
          daily_checkin: true,
          new_user_trial: true,
          recharge: false,
          vip: false,
        },
        new_user_trial: {
          enabled: true,
          quota_amount: 0.1,
          quota_used: 0.04,
          remaining_quota: 0.06,
          status: 'active',
          can_use: true,
          reason: 'available',
        },
      },
    })

    const overview = await getWelfareOverview()

    expect(overview.modules.new_user_trial).toBe(true)
    expect(overview.new_user_trial?.quota_amount).toBe(0.1)
    expect(overview.new_user_trial?.remaining_quota).toBe(0.06)
    expect(overview.new_user_trial?.status).toBe('active')
  })

  it('loads daily checkin status', async () => {
    await getWelfareDailyCheckin()

    expect(get).toHaveBeenCalledWith('/user/welfare/daily-checkin', {
      signal: undefined,
    })
  })

  it('claims daily checkin', async () => {
    await claimWelfareDailyCheckin()

    expect(post).toHaveBeenCalledWith('/user/welfare/daily-checkin/claim')
  })

  it('claims a daily checkin milestone', async () => {
    await claimWelfareDailyCheckinMilestone(14)

    expect(post).toHaveBeenCalledWith('/user/welfare/daily-checkin/milestones/14/claim')
  })
})
