import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { welfareAPI } from '@/api/welfare'
import type { WelfareDailyCheckin, WelfareNewUserTrial, WelfareOverview } from '@/types'

const CACHE_TTL_MS = 60_000

export const useWelfareStore = defineStore('welfare', () => {
  const overview = ref<WelfareOverview | null>(null)
  const loading = ref(false)
  const loaded = ref(false)
  const lastFetchedAt = ref<number | null>(null)

  let activePromise: Promise<WelfareOverview | null> | null = null
  let mutationVersion = 0

  const hasClaimableReward = computed(() => {
    const current = overview.value
    if (!current?.enabled) return false

    const daily = current.daily_checkin
    const hasDailyCheckin = Boolean(daily?.enabled && daily.can_claim_today)
    const hasMilestone = Boolean(daily?.enabled && daily.milestones.some((item) => item.claimable))

    const trial = current.new_user_trial
    const hasTrialReward = Boolean(
      trial?.enabled &&
      trial.success_reward_claimable &&
      !trial.success_reward_claimed
    )

    return hasDailyCheckin || hasMilestone || hasTrialReward
  })

  async function fetchOverview(force = false): Promise<WelfareOverview | null> {
    const now = Date.now()
    if (
      !force &&
      loaded.value &&
      lastFetchedAt.value &&
      now - lastFetchedAt.value < CACHE_TTL_MS
    ) {
      return overview.value
    }

    if (activePromise) {
      return activePromise
    }

    const requestVersion = mutationVersion
    loading.value = true
    const request = welfareAPI
      .getWelfareOverview()
      .then((data) => {
        if (requestVersion === mutationVersion) {
          setOverview(data)
        }
        return data
      })
      .catch((error: unknown) => {
        lastFetchedAt.value = null
        console.error('Failed to fetch welfare overview:', error)
        return null
      })
      .finally(() => {
        if (activePromise === request) {
          activePromise = null
          loading.value = false
        }
      })

    activePromise = request
    return request
  }

  function setOverview(nextOverview: WelfareOverview): void {
    mutationVersion++
    overview.value = nextOverview
    loaded.value = true
    lastFetchedAt.value = Date.now()
  }

  function updateDaily(nextDaily: WelfareDailyCheckin): void {
    if (!overview.value) return
    setOverview({
      ...overview.value,
      daily_checkin: nextDaily,
    })
  }

  function updateTrial(nextTrial: WelfareNewUserTrial): void {
    if (!overview.value) return
    setOverview({
      ...overview.value,
      new_user_trial: nextTrial,
    })
  }

  function reset(): void {
    overview.value = null
    loading.value = false
    loaded.value = false
    lastFetchedAt.value = null
    activePromise = null
    mutationVersion++
  }

  return {
    overview,
    loading,
    loaded,
    hasClaimableReward,
    fetchOverview,
    setOverview,
    updateDaily,
    updateTrial,
    reset,
  }
})
