<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div class="min-w-0">
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('welfare.title') }}</h1>
        </div>
        <button class="btn btn-secondary h-10 shrink-0" type="button" :disabled="loading" @click="loadOverview">
          {{ t('common.refresh') }}
        </button>
      </div>

      <div v-if="loading" class="card flex min-h-[280px] items-center justify-center p-8">
        <div class="text-center">
          <div class="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          <p class="mt-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</p>
        </div>
      </div>

      <div v-else-if="error" class="card p-8 text-center">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('welfare.errorTitle') }}</h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ error }}</p>
        <button class="btn btn-primary mt-5" type="button" @click="loadOverview">{{ t('welfare.retry') }}</button>
      </div>

      <div v-else-if="overview && !overview.enabled" class="card p-8 text-center">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('welfare.disabledTitle') }}</h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ t('welfare.disabledDescription') }}</p>
      </div>

      <template v-else-if="overview">
        <section v-if="trial && trial.enabled" class="card p-5">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('welfare.trial.title') }}</h2>
                <span class="w-fit rounded-full px-3 py-1 text-xs font-medium" :class="trialStatusClass">
                  {{ trialStatusText }}
                </span>
              </div>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('welfare.trial.activationNotice') }}</p>
            </div>
            <button
              class="btn btn-primary h-10 w-full shrink-0 sm:w-auto"
              type="button"
              :disabled="!trial.can_use"
              :title="t('welfare.trial.activationNotice')"
              data-testid="welfare-trial-claim"
              @click="goToTrialActivation"
            >
              {{ t('welfare.trial.cta') }}
            </button>
          </div>

          <div class="mt-5 border-t border-gray-100 pt-4 dark:border-dark-700">
            <div class="flex items-end justify-between gap-4">
              <div>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('welfare.trial.remaining') }}</p>
                <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ formatAmount(trial.remaining_quota) }}</p>
              </div>
              <p class="shrink-0 text-xs text-gray-500 dark:text-gray-400">
                {{ t('welfare.trial.used') }} {{ formatAmount(trial.quota_used) }} / {{ t('welfare.trial.total') }} {{ formatAmount(trial.quota_amount) }}
              </p>
            </div>
            <div class="mt-3 h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-700">
              <div class="h-full rounded-full bg-primary-600 transition-all" :style="{ width: `${trialProgressPercent}%` }"></div>
            </div>
            <p class="mt-3 text-xs text-gray-500 dark:text-gray-400">{{ t('welfare.trial.walletNotice') }}</p>
          </div>
        </section>

        <section v-if="daily && daily.enabled" class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_20rem]">
          <div class="card p-5">
            <section>
              <div class="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
                <div class="flex min-w-0 gap-3">
                  <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-primary-50 text-primary-700 dark:bg-primary-500/10 dark:text-primary-300">
                    <Icon name="calendar" size="sm" :stroke-width="2" />
                  </span>
                  <div class="min-w-0">
                    <div class="flex flex-wrap items-center gap-2">
                      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('welfare.daily.title') }}</h2>
                      <span class="w-fit rounded-full px-3 py-1 text-xs font-medium" :class="dailyStatusClass">
                        {{ dailyStatusText }}
                      </span>
                    </div>
                    <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ dailyActionDescription }}</p>
                  </div>
                </div>
                <div class="flex w-full flex-col gap-3 sm:flex-row sm:items-center lg:w-auto">
                  <div class="rounded-md bg-primary-50 px-4 py-3 dark:bg-primary-500/10">
                    <p class="text-xs text-primary-700 dark:text-primary-300">{{ t('welfare.daily.todayReward') }}</p>
                    <p class="mt-1 truncate text-2xl font-bold text-primary-700 dark:text-primary-200">
                      {{ formatAmount(daily.today_reward_amount) }}
                    </p>
                  </div>
                  <button
                    class="btn h-10 w-full sm:w-auto"
                    :class="daily.can_claim_today ? 'btn-primary' : 'btn-secondary'"
                    type="button"
                    :disabled="!daily.can_claim_today || claimingDaily"
                    data-testid="welfare-daily-claim"
                    @click="claimDaily"
                  >
                    {{ dailyButtonText }}
                  </button>
                </div>
              </div>

              <div class="mt-5 grid gap-3 sm:grid-cols-3">
                <div class="rounded-md border border-gray-100 bg-white px-4 py-3 dark:border-dark-700 dark:bg-dark-800">
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('welfare.daily.currentStreak') }}</p>
                  <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">
                    {{ daily.current_streak_days }}
                    <span class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('welfare.daily.daysUnit') }}</span>
                  </p>
                </div>
                <div class="rounded-md border border-gray-100 bg-white px-4 py-3 dark:border-dark-700 dark:bg-dark-800">
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('welfare.daily.monthDays') }}</p>
                  <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">
                    {{ daily.month_checkin_days }}
                    <span class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('welfare.daily.daysUnit') }}</span>
                  </p>
                </div>
                <div class="rounded-md border border-gray-100 bg-white px-4 py-3 dark:border-dark-700 dark:bg-dark-800">
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('welfare.milestones.nextTitle') }}</p>
                  <p class="mt-1 text-sm font-medium leading-6 text-gray-900 dark:text-white">{{ nextMilestoneSummary }}</p>
                </div>
              </div>

              <p v-if="claimError" class="mt-4 text-sm text-red-600 dark:text-red-400">{{ claimError }}</p>
            </section>

            <section class="mt-6 border-t border-gray-100 pt-5 dark:border-dark-700">
              <div class="flex items-center justify-between gap-3">
                <div class="flex min-w-0 gap-3">
                  <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">
                    <Icon name="gift" size="sm" :stroke-width="2" />
                  </span>
                  <div class="min-w-0">
                    <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('welfare.milestones.title') }}</h2>
                    <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('welfare.milestones.description') }}</p>
                  </div>
                </div>
                <span class="shrink-0 text-xs text-gray-500 dark:text-gray-400">{{ daily.reward_month }}</span>
              </div>

              <div class="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                <article v-for="milestone in daily.milestones" :key="milestone.day" class="flex min-h-[126px] flex-col justify-between rounded-md border p-4" :class="milestoneCardClass(milestone)">
                  <div>
                    <div class="flex min-w-0 items-center justify-between gap-3">
                      <p class="min-w-0 truncate text-sm font-semibold text-gray-900 dark:text-white" :title="t('welfare.milestones.day', { day: milestone.day })">
                        {{ t('welfare.milestones.day', { day: milestone.day }) }}
                      </p>
                      <span class="max-w-[8rem] shrink-0 truncate rounded-full px-2.5 py-1 text-center text-xs font-medium" :class="milestoneStatusClass(milestone)" :title="milestoneStatusText(milestone)">
                        {{ milestoneStatusText(milestone) }}
                      </span>
                    </div>
                    <p class="mt-3 truncate text-xl font-bold text-gray-900 dark:text-white">
                      {{ formatAmount(milestone.amount) }}
                    </p>
                  </div>
                  <button
                    v-if="milestone.claimable"
                    class="btn mt-4 min-h-10 w-full whitespace-normal px-3 py-2 text-center leading-tight"
                    :class="milestone.claimable ? 'btn-primary' : 'btn-secondary'"
                    type="button"
                    :disabled="claimingMilestoneDay === milestone.day"
                    :data-testid="`welfare-milestone-${milestone.day}`"
                    :title="milestoneButtonText(milestone)"
                    @click="claimMilestone(milestone.day)"
                  >
                    {{ milestoneButtonText(milestone) }}
                  </button>
                </article>
              </div>
            </section>
          </div>

          <aside class="space-y-5 xl:sticky xl:top-20 xl:self-start">
            <section class="card p-5">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('welfare.calendar.title') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('welfare.calendar.description') }}</p>
              <div class="mt-4 grid grid-cols-7 gap-1.5">
                <span
                  v-for="day in monthDays"
                  :key="day.date"
                  class="flex aspect-square items-center justify-center rounded-md text-xs font-medium"
                  :class="day.checked ? 'bg-primary-600 text-white' : day.isToday ? 'border border-primary-400 text-primary-700 dark:text-primary-300' : 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'"
                  :title="day.date"
                >
                  {{ day.label }}
                </span>
              </div>
              <p class="mt-4 border-t border-gray-100 pt-3 text-xs text-gray-500 dark:border-dark-700 dark:text-gray-400">
                {{ t('welfare.notes.timezone', { timezone: daily.settlement_timezone }) }}
              </p>
            </section>
          </aside>
        </section>

        <div v-else class="card p-8 text-center">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('welfare.daily.disabledTitle') }}</h2>
          <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ t('welfare.daily.disabledDescription') }}</p>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { welfareAPI } from '@/api/welfare'
import type { WelfareDailyCheckin, WelfareDailyCheckinMilestone, WelfareOverview } from '@/types'

const { t } = useI18n()
const router = useRouter()

const overview = ref<WelfareOverview | null>(null)
const loading = ref(false)
const error = ref('')
const claimError = ref('')
const claimingDaily = ref(false)
const claimingMilestoneDay = ref<number | null>(null)

const daily = computed(() => overview.value?.daily_checkin ?? null)
const trial = computed(() => overview.value?.new_user_trial ?? null)

const checkedDateSet = computed(() => new Set(daily.value?.checkin_dates ?? []))

const monthDays = computed(() => {
  const state = daily.value
  if (!state?.reward_month) return []
  const [yearRaw, monthRaw] = state.reward_month.split('-')
  const year = Number(yearRaw)
  const month = Number(monthRaw)
  if (!Number.isFinite(year) || !Number.isFinite(month)) return []
  const days = new Date(year, month, 0).getDate()
  return Array.from({ length: days }, (_, index) => {
    const day = index + 1
    const date = `${state.reward_month}-${String(day).padStart(2, '0')}`
    return {
      date,
      label: String(day),
      checked: checkedDateSet.value.has(date),
      isToday: state.today === date,
    }
  })
})

const amountFormatter = new Intl.NumberFormat(undefined, {
  minimumFractionDigits: 0,
  maximumFractionDigits: 8,
})

const trialProgressPercent = computed(() => {
  const state = trial.value
  if (!state || state.quota_amount <= 0) return 0
  return Math.max(0, Math.min(100, (state.quota_used / state.quota_amount) * 100))
})

const nextMilestone = computed(() => {
  const state = daily.value
  if (!state) return null
  return state.milestones.find((item) => !item.claimed && item.day >= state.current_streak_days) ?? state.milestones.find((item) => !item.claimed) ?? null
})

const nextMilestoneSummary = computed(() => {
  const state = daily.value
  const milestone = nextMilestone.value
  if (!state || !milestone) return t('welfare.milestones.allDone')
  if (milestone.claimable) {
    return t('welfare.milestones.readySummary', {
      day: milestone.day,
      amount: formatAmount(milestone.amount),
    })
  }
  return t('welfare.milestones.nextSummary', {
    days: Math.max(0, milestone.day - state.current_streak_days),
    day: milestone.day,
    amount: formatAmount(milestone.amount),
  })
})

const dailyStatusText = computed(() => {
  const state = daily.value
  if (!state) return t('welfare.notOpen')
  if (state.checked_today) return t('welfare.daily.checked')
  if (state.can_claim_today) return t('welfare.daily.canClaim')
  return reasonText(state.reason)
})

const dailyActionDescription = computed(() => {
  const state = daily.value
  if (!state) return t('welfare.daily.description')
  if (state.checked_today) {
    return t('welfare.daily.claimedDescription', {
      amount: formatAmount(state.today_reward_amount),
      streak: state.current_streak_days,
    })
  }
  if (state.can_claim_today) return t('welfare.daily.description')
  return reasonText(state.reason)
})

const dailyStatusClass = computed(() => {
  const state = daily.value
  if (state?.checked_today) return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
  if (state?.can_claim_today) return 'bg-primary-50 text-primary-700 dark:bg-primary-500/10 dark:text-primary-300'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
})

const dailyButtonText = computed(() => {
  if (claimingDaily.value) return t('welfare.daily.claiming')
  if (daily.value?.checked_today) return t('welfare.daily.claimed')
  return daily.value?.can_claim_today ? t('welfare.daily.claim') : reasonText(daily.value?.reason)
})

const trialStatusText = computed(() => {
  const state = trial.value
  if (!state) return t('welfare.notOpen')
  switch (state.status) {
    case 'available':
      return t('welfare.trial.status.available')
    case 'active':
      return t('welfare.trial.status.active')
    case 'in_progress':
      return t('welfare.trial.status.inProgress')
    case 'exhausted':
      return t('welfare.trial.status.exhausted')
    default:
      return reasonText(state.reason)
  }
})

const trialStatusClass = computed(() => {
  const state = trial.value
  if (state?.status === 'exhausted') return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
  if (state?.status === 'in_progress') return 'bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300'
  if (state?.can_use) return 'bg-primary-50 text-primary-700 dark:bg-primary-500/10 dark:text-primary-300'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
})

function formatAmount(value: number | null | undefined): string {
  return amountFormatter.format(Number(value) || 0)
}

function reasonText(reason?: string): string {
  switch (reason) {
    case 'already_checked':
    case 'already_claimed':
      return t('welfare.reason.alreadyClaimed')
    case 'not_reached':
      return t('welfare.reason.notReached')
    case 'zero_reward':
      return t('welfare.reason.zeroReward')
    case 'disabled':
      return t('welfare.reason.disabled')
    case 'not_configured':
      return t('welfare.reason.notConfigured')
    case 'in_progress':
      return t('welfare.reason.inProgress')
    case 'exhausted':
      return t('welfare.reason.exhausted')
    case 'daily_limit':
      return t('welfare.reason.dailyLimit')
    default:
      return t('welfare.reason.unavailable')
  }
}

function milestoneStatusText(milestone: WelfareDailyCheckinMilestone): string {
  if (milestone.claimed) return t('welfare.milestones.claimed')
  if (milestone.claimable) return t('welfare.milestones.canClaim')
  return reasonText(milestone.reason)
}

function milestoneStatusClass(milestone: WelfareDailyCheckinMilestone): string {
  if (milestone.claimed) return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
  if (milestone.claimable) return 'bg-primary-50 text-primary-700 dark:bg-primary-500/10 dark:text-primary-300'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
}

function milestoneCardClass(milestone: WelfareDailyCheckinMilestone): string {
  if (milestone.claimable) return 'border-primary-200 bg-primary-50/60 dark:border-primary-500/30 dark:bg-primary-500/10'
  if (milestone.claimed) return 'border-emerald-200 bg-emerald-50/50 dark:border-emerald-500/30 dark:bg-emerald-500/10'
  return 'border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800'
}

function milestoneButtonText(milestone: WelfareDailyCheckinMilestone): string {
  if (claimingMilestoneDay.value === milestone.day) return t('welfare.milestones.claiming')
  if (milestone.claimed) return t('welfare.milestones.claimed')
  return milestone.claimable ? t('welfare.milestones.claim') : reasonText(milestone.reason)
}

function goToTrialActivation(): void {
  void router.push('/keys')
}

function extractErrorMessage(err: unknown, fallback: string): string {
  if (err && typeof err === 'object' && 'message' in err) {
    const message = String((err as { message?: unknown }).message || '')
    if (message) return message
  }
  return fallback
}

async function loadOverview(): Promise<void> {
  loading.value = true
  error.value = ''
  claimError.value = ''
  try {
    overview.value = await welfareAPI.getWelfareOverview()
  } catch (err) {
    error.value = extractErrorMessage(err, t('welfare.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function claimDaily(): Promise<void> {
  if (!daily.value?.can_claim_today || claimingDaily.value) return
  claimingDaily.value = true
  claimError.value = ''
  try {
    const result = await welfareAPI.claimWelfareDailyCheckin()
    updateDaily(result.daily_checkin)
  } catch (err) {
    claimError.value = extractErrorMessage(err, t('welfare.daily.claimFailed'))
  } finally {
    claimingDaily.value = false
  }
}

async function claimMilestone(day: number): Promise<void> {
  if (claimingMilestoneDay.value !== null) return
  const milestone = daily.value?.milestones.find((item) => item.day === day)
  if (!milestone?.claimable) return
  claimingMilestoneDay.value = day
  claimError.value = ''
  try {
    const result = await welfareAPI.claimWelfareDailyCheckinMilestone(day)
    updateDaily(result.daily_checkin)
  } catch (err) {
    claimError.value = extractErrorMessage(err, t('welfare.milestones.claimFailed'))
  } finally {
    claimingMilestoneDay.value = null
  }
}

function updateDaily(nextDaily: WelfareDailyCheckin): void {
  if (!overview.value) return
  overview.value = {
    ...overview.value,
    daily_checkin: nextDaily,
  }
}

onMounted(() => {
  void loadOverview()
})
</script>
