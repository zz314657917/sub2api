import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useWelfareStore } from '@/stores/welfare'
import { welfareAPI } from '@/api/welfare'
import type { WelfareOverview } from '@/types'

vi.mock('@/api/welfare', () => ({
  welfareAPI: {
    getWelfareOverview: vi.fn(),
  },
}))

const baseOverview = (overrides: Partial<WelfareOverview> = {}): WelfareOverview => ({
  enabled: true,
  modules: {
    daily_checkin: true,
    new_user_trial: true,
    recharge: true,
    vip: false,
  },
  recharge: {
    enabled: true,
    first_bonus_amount: 5,
    first_bonus_claimed: false,
    reason: 'available',
  },
  daily_checkin: {
    enabled: true,
    today: '2026-06-01',
    reward_month: '2026-06',
    checked_today: true,
    today_reward_amount: 0,
    reward_min: 0,
    reward_max: 0,
    current_streak_days: 1,
    month_checkin_days: 1,
    checkin_dates: ['2026-06-01'],
    milestones: [],
    can_claim_today: false,
    reason: 'already_checked',
    settlement_timezone: 'Asia/Shanghai',
  },
  new_user_trial: {
    enabled: true,
    quota_amount: 0.1,
    quota_used: 0,
    remaining_quota: 0.1,
    success_reward_amount: 1,
    success_reward_claimable: false,
    success_reward_claimed: false,
    success_reward_reason: 'not_reached',
    status: 'available',
    can_use: true,
    reason: 'available',
  },
  ...overrides,
})

describe('useWelfareStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.mocked(welfareAPI.getWelfareOverview).mockReset()
  })

  it('marks daily checkin as claimable', () => {
    const store = useWelfareStore()
    store.setOverview(baseOverview({
      daily_checkin: {
        ...baseOverview().daily_checkin!,
        checked_today: false,
        can_claim_today: true,
        reason: 'available',
      },
    }))

    expect(store.hasClaimableReward).toBe(true)
  })

  it('marks milestone and new user success reward as claimable', () => {
    const store = useWelfareStore()
    store.setOverview(baseOverview({
      daily_checkin: {
        ...baseOverview().daily_checkin!,
        milestones: [{ day: 7, amount: 1, claimed: false, claimable: true, reason: 'available' }],
      },
    }))
    expect(store.hasClaimableReward).toBe(true)

    store.setOverview(baseOverview({
      daily_checkin: {
        ...baseOverview().daily_checkin!,
        milestones: [],
      },
      new_user_trial: {
        ...baseOverview().new_user_trial!,
        success_reward_claimable: true,
      },
    }))
    expect(store.hasClaimableReward).toBe(true)
  })

  it('does not mark available trial quota itself as claimable balance', () => {
    const store = useWelfareStore()
    store.setOverview(baseOverview())

    expect(store.hasClaimableReward).toBe(false)
  })

  it('does not mark first recharge bonus as a manual claimable reward', () => {
    const store = useWelfareStore()
    store.setOverview(baseOverview({
      recharge: {
        enabled: true,
        first_bonus_amount: 5,
        first_bonus_claimed: false,
        reason: 'available',
      },
    }))

    expect(store.hasClaimableReward).toBe(false)
  })

  it('fetches and caches welfare overview', async () => {
    const store = useWelfareStore()
    vi.mocked(welfareAPI.getWelfareOverview).mockResolvedValue(baseOverview({
      daily_checkin: {
        ...baseOverview().daily_checkin!,
        can_claim_today: true,
      },
    }))

    await store.fetchOverview()
    await store.fetchOverview()

    expect(welfareAPI.getWelfareOverview).toHaveBeenCalledTimes(1)
    expect(store.hasClaimableReward).toBe(true)
  })
})
