<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('leaderboard.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.description') }}</p>
        </div>
        <button class="btn btn-secondary h-10 shrink-0" type="button" :disabled="loading" @click="loadLeaderboard">
          {{ t('common.refresh') }}
        </button>
      </div>

      <div class="card p-4">
        <div class="grid grid-cols-2 gap-2 sm:grid-cols-4" role="tablist" :aria-label="t('leaderboard.periodLabel')">
          <button
            v-for="option in periodOptions"
            :key="option.value"
            type="button"
            class="h-10 rounded-lg border px-3 text-sm font-medium transition-colors"
            :class="period === option.value
              ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-500/10 dark:text-primary-300'
              : 'border-gray-200 bg-white text-gray-600 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700'"
            :aria-selected="period === option.value"
            role="tab"
            @click="selectPeriod(option.value)"
          >
            {{ option.label }}
          </button>
        </div>
      </div>

      <div v-if="leaderboard" class="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.totalRequests') }}</p>
          <p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ formatNumber(leaderboard.total_requests) }}</p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.totalTokens') }}</p>
          <p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ formatNumber(leaderboard.total_tokens) }}</p>
        </div>
        <div class="card p-5" data-testid="leaderboard-cost-efficiency-summary">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.costEfficiencyKing') }}</p>
          <p class="mt-2 truncate text-2xl font-bold text-gray-900 dark:text-white">{{ costEfficiencyKing?.display_name ?? t('leaderboard.notRanked') }}</p>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ costEfficiencyPerMillionText }}</p>
        </div>
      </div>

      <div v-if="leaderboard" class="flex flex-col gap-2 text-sm text-gray-500 dark:text-gray-400 md:flex-row md:items-center md:justify-between">
        <span>{{ formatDateRange(leaderboard.start_date, leaderboard.end_date) }}</span>
        <span>{{ t('leaderboard.generatedAt') }}: {{ formatDateTime(leaderboard.generated_at) }}</span>
      </div>

      <div v-if="loading" class="card flex min-h-[280px] items-center justify-center p-8">
        <div class="text-center">
          <div class="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          <p class="mt-3 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</p>
        </div>
      </div>

      <div v-else-if="error" class="card p-8 text-center">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('leaderboard.errorTitle') }}</h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.errorDescription') }}</p>
        <button class="btn btn-primary mt-5" type="button" @click="loadLeaderboard">{{ t('leaderboard.retry') }}</button>
      </div>

      <div v-else-if="leaderboard" class="grid grid-cols-1 gap-5 xl:grid-cols-[minmax(0,1fr)_22rem]">
        <div class="min-w-0 space-y-3">
          <div v-if="rankingItems.length === 0" class="card p-8">
            <EmptyState :title="t('leaderboard.emptyTitle')" :description="t('leaderboard.emptyDescription')" />
          </div>

          <template v-else>
            <div class="card hidden overflow-hidden md:block">
              <div class="table-wrapper overflow-x-auto">
                <table class="w-full min-w-[620px] table-fixed">
                  <thead class="bg-gray-50 dark:bg-dark-800">
                    <tr>
                      <th class="w-24 px-5 py-4 text-left text-sm font-medium text-gray-600 dark:text-dark-300">{{ t('leaderboard.rank') }}</th>
                      <th class="px-5 py-4 text-left text-sm font-medium text-gray-600 dark:text-dark-300">{{ t('leaderboard.user') }}</th>
                      <th class="w-32 px-5 py-4 text-right text-sm font-medium text-gray-600 dark:text-dark-300">{{ t('leaderboard.requests') }}</th>
                      <th class="w-32 px-5 py-4 text-right text-sm font-medium text-gray-600 dark:text-dark-300">{{ t('leaderboard.tokens') }}</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-800 dark:bg-dark-900">
                    <tr
                      v-for="item in rankingItems"
                      :key="item.user_id"
                      :class="[
                        item.is_current_user ? 'bg-primary-50/70 dark:bg-primary-500/10' : '',
                        rankRowClass(item.rank),
                      ]"
                    >
                      <td class="px-5 py-4">
                        <span class="inline-flex h-8 min-w-8 items-center justify-center rounded-lg px-2 text-sm font-bold" :class="rankBadgeClass(item.rank)">
                          #{{ item.rank }}
                        </span>
                      </td>
                      <td class="px-5 py-4">
                        <div class="flex min-w-0 items-center gap-3">
                          <div :class="avatarFrameClass(item.rank)">
                            <img v-if="item.avatar_url" :src="item.avatar_url" alt="" class="h-10 w-10 rounded-full object-cover" />
                            <div v-else class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gray-100 text-sm font-semibold text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                              {{ getInitial(item.display_name) }}
                            </div>
                          </div>
                          <div class="min-w-0">
                            <div class="flex items-center gap-2">
                              <span class="truncate font-semibold text-gray-900 dark:text-white">{{ item.display_name }}</span>
                              <span
                                v-if="isCostEfficiencyKing(item)"
                                class="shrink-0 rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-800 ring-1 ring-amber-200 dark:bg-amber-500/20 dark:text-amber-200 dark:ring-amber-400/40"
                                data-testid="leaderboard-cost-efficiency-king"
                                :data-user-id="String(item.user_id)"
                              >
                                {{ t('leaderboard.costEfficiencyKing') }}
                              </span>
                              <span v-if="item.is_current_user" class="rounded-full bg-primary-100 px-2 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-500/20 dark:text-primary-300">
                                {{ t('leaderboard.currentUser') }}
                              </span>
                            </div>
                            <p class="truncate text-sm text-gray-500 dark:text-gray-400">{{ item.email_masked }}</p>
                          </div>
                        </div>
                      </td>
                      <td class="px-5 py-4 text-right text-gray-700 dark:text-gray-300">{{ formatNumber(item.requests) }}</td>
                      <td class="px-5 py-4 text-right text-gray-700 dark:text-gray-300">{{ formatNumber(item.tokens) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div class="space-y-3 md:hidden">
              <div
                v-for="item in rankingItems"
                :key="item.user_id"
                class="card p-4"
                :class="[item.is_current_user ? 'border-primary-200 bg-primary-50/70 dark:border-primary-500/30 dark:bg-primary-500/10' : '', rankRowClass(item.rank)]"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="flex min-w-0 items-center gap-3">
                    <div :class="avatarFrameClass(item.rank)">
                      <img v-if="item.avatar_url" :src="item.avatar_url" alt="" class="h-10 w-10 rounded-full object-cover" />
                      <div v-else class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gray-100 text-sm font-semibold text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                        {{ getInitial(item.display_name) }}
                      </div>
                    </div>
                    <div class="min-w-0">
                      <div class="flex min-w-0 items-center gap-2">
                        <span class="inline-flex h-7 min-w-7 items-center justify-center rounded-lg px-2 text-xs font-bold" :class="rankBadgeClass(item.rank)">
                          #{{ item.rank }}
                        </span>
                        <p class="truncate font-semibold text-gray-900 dark:text-white">{{ item.display_name }}</p>
                        <span
                          v-if="isCostEfficiencyKing(item)"
                          class="shrink-0 rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-800 ring-1 ring-amber-200 dark:bg-amber-500/20 dark:text-amber-200 dark:ring-amber-400/40"
                          data-testid="leaderboard-cost-efficiency-king"
                          :data-user-id="String(item.user_id)"
                        >
                          {{ t('leaderboard.costEfficiencyKing') }}
                        </span>
                        <span v-if="item.is_current_user" class="shrink-0 rounded-full bg-primary-100 px-2 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-500/20 dark:text-primary-300">
                          {{ t('leaderboard.currentUser') }}
                        </span>
                      </div>
                      <p class="truncate text-sm text-gray-500 dark:text-gray-400">{{ item.email_masked }}</p>
                    </div>
                  </div>
                </div>
                <div class="mt-4 grid grid-cols-2 gap-3 text-sm">
                  <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
                    <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('leaderboard.requests') }}</p>
                    <p class="font-semibold text-gray-900 dark:text-white">{{ formatNumber(item.requests) }}</p>
                  </div>
                  <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
                    <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('leaderboard.tokens') }}</p>
                    <p class="font-semibold text-gray-900 dark:text-white">{{ formatNumber(item.tokens) }}</p>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>

        <aside class="space-y-5 xl:sticky xl:top-20 xl:self-start">
          <section class="card border-primary-200 bg-primary-50/60 p-5 dark:border-primary-500/30 dark:bg-primary-500/10" data-testid="leaderboard-my-info">
            <p class="text-sm font-semibold text-primary-700 dark:text-primary-300">{{ t('leaderboard.myInfo') }}</p>
            <div class="mt-3 min-w-0">
              <p class="truncate text-xl font-bold text-gray-900 dark:text-white">
                {{ myRankLabel }} {{ myDisplayName }}
              </p>
              <p v-if="myEntry" class="truncate text-sm text-gray-500 dark:text-gray-400">{{ myEntry.email_masked }}</p>
            </div>
            <div class="mt-4 grid grid-cols-2 gap-3 text-sm">
              <div class="rounded-lg bg-white/70 p-3 dark:bg-dark-800/70">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('leaderboard.balance') }}</p>
                <p class="font-semibold text-gray-900 dark:text-white">{{ formatCurrency(myEntry?.balance ?? 0) }}</p>
              </div>
              <div class="rounded-lg bg-white/70 p-3 dark:bg-dark-800/70">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('leaderboard.requests') }}</p>
                <p class="font-semibold text-gray-900 dark:text-white">{{ formatNumber(myEntry?.requests ?? 0) }}</p>
              </div>
              <div class="rounded-lg bg-white/70 p-3 dark:bg-dark-800/70">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('leaderboard.tokens') }}</p>
                <p class="font-semibold text-gray-900 dark:text-white">{{ formatNumber(myEntry?.tokens ?? 0) }}</p>
              </div>
            </div>
          </section>

          <section v-if="dailyRewards" class="card p-5" data-testid="leaderboard-daily-reward">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('leaderboard.dailyReward.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{ t('leaderboard.dailyReward.settlementDate') }}: {{ dailyRewards.reward_date || '-' }}
                </p>
              </div>
              <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="dailyRewardStatusClass">
                {{ dailyRewardReasonText }}
              </span>
            </div>

            <div class="mt-4 grid grid-cols-3 gap-2">
              <div
                v-for="tier in rewardTiers"
                :key="tier.rank"
                class="rounded-lg border border-gray-100 bg-gray-50 p-3 text-center dark:border-dark-700 dark:bg-dark-800"
              >
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('leaderboard.dailyReward.rankReward', { rank: tier.rank }) }}</p>
                <p class="mt-1 text-sm font-bold text-gray-900 dark:text-white">{{ formatCurrency(tier.amount) }}</p>
              </div>
            </div>

            <div class="mt-4">
              <div class="flex items-center justify-between gap-3 text-xs text-gray-500 dark:text-gray-400">
                <span>{{ t('leaderboard.dailyReward.threshold') }}</span>
                <span>
                  {{ t('leaderboard.dailyReward.progress', {
                    current: formatCurrency(dailyRewards.yesterday_total_actual_cost),
                    target: formatCurrency(dailyRewards.min_total_actual_cost),
                  }) }}
                </span>
              </div>
              <div class="mt-2 h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-700">
                <div
                  class="h-full rounded-full bg-primary-500 transition-all"
                  :style="{ width: `${rewardThresholdPercent}%` }"
                ></div>
              </div>
            </div>

            <div class="mt-4 rounded-lg bg-gray-50 p-3 text-sm dark:bg-dark-800">
              <div class="flex items-center justify-between gap-3">
                <span class="text-gray-500 dark:text-gray-400">{{ t('leaderboard.myRank') }}</span>
                <span class="font-semibold text-gray-900 dark:text-white">{{ formatRewardRankLabel(dailyRewards.current_user_rank) }}</span>
              </div>
              <div class="mt-2 flex items-center justify-between gap-3">
                <span class="text-gray-500 dark:text-gray-400">{{ t('leaderboard.dailyReward.rewardAmount') }}</span>
                <span class="font-semibold text-gray-900 dark:text-white">{{ formatCurrency(dailyRewards.current_user_reward_amount) }}</span>
              </div>
            </div>

            <p v-if="claimError" class="mt-3 text-sm text-red-600 dark:text-red-400">{{ claimError }}</p>

            <button
              class="btn btn-primary mt-4 w-full"
              type="button"
              :disabled="!dailyRewards.can_claim || claimingReward"
              data-testid="leaderboard-daily-reward-claim"
              @click="claimDailyReward"
            >
              {{ claimButtonText }}
            </button>
          </section>
        </aside>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usageAPI } from '@/api'
import type { LeaderboardDailyRewards, LeaderboardPeriod, UserLeaderboardItem, UserLeaderboardResponse } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { formatCurrency, formatDateTime, formatNumber } from '@/utils/format'

const { t } = useI18n()

const period = ref<LeaderboardPeriod>('day')
const leaderboard = ref<UserLeaderboardResponse | null>(null)
const loading = ref(false)
const error = ref(false)
const claimingReward = ref(false)
const claimError = ref('')
const leaderboardLimit = 10
let loadSeq = 0

const periodOptions = computed(() => [
  { value: 'day' as const, label: t('leaderboard.period.day') },
  { value: 'week' as const, label: t('leaderboard.period.week') },
  { value: 'month' as const, label: t('leaderboard.period.month') },
  { value: 'all' as const, label: t('leaderboard.period.all') },
])

const rankingItems = computed<UserLeaderboardItem[]>(() => (leaderboard.value?.ranking ?? []).slice(0, leaderboardLimit))
const costEfficiencyKing = computed<UserLeaderboardItem | null>(() => {
  let bestItem: UserLeaderboardItem | null = null
  let bestCostPerToken = Number.POSITIVE_INFINITY

  for (const item of rankingItems.value) {
    if (item.actual_cost <= 0 || item.tokens <= 0) continue

    const costPerToken = item.actual_cost / item.tokens
    if (costPerToken < bestCostPerToken) {
      bestCostPerToken = costPerToken
      bestItem = item
    }
  }

  return bestItem
})
const costEfficiencyKingUserId = computed(() => costEfficiencyKing.value?.user_id ?? null)
const costEfficiencyPerMillionText = computed(() => {
  const item = costEfficiencyKing.value
  if (!item) return '-'
  return `1M Token = ${formatCurrency((item.actual_cost / item.tokens) * 1_000_000)}`
})
const dailyRewards = computed<LeaderboardDailyRewards | null>(() => leaderboard.value?.daily_rewards ?? null)
const myEntry = computed<UserLeaderboardItem | null>(() => {
  if (leaderboard.value?.current_user_entry) return leaderboard.value.current_user_entry
  return rankingItems.value.find((item) => item.is_current_user) ?? null
})
const myDisplayName = computed(() => myEntry.value?.display_name ?? t('leaderboard.currentUser'))
const myRankLabel = computed(() => (myEntry.value?.rank ? `#${myEntry.value.rank}` : t('leaderboard.notRanked')))

const rewardTiers = computed(() => {
  const tiers = new Map((dailyRewards.value?.rewards ?? []).map((tier) => [tier.rank, tier.amount]))
  return [1, 2, 3].map((rank) => ({ rank, amount: tiers.get(rank) ?? 0 }))
})

const rewardThresholdPercent = computed(() => {
  const reward = dailyRewards.value
  if (!reward) return 0
  if (reward.min_total_actual_cost <= 0) return reward.threshold_met ? 100 : 0
  return Math.min(100, Math.max(0, (reward.yesterday_total_actual_cost / reward.min_total_actual_cost) * 100))
})

const dailyRewardReasonText = computed(() => {
  const reason = dailyRewards.value?.reason
  if (reason === 'eligible') return t('leaderboard.dailyReward.eligible')
  if (reason === 'already_claimed') return t('leaderboard.dailyReward.alreadyClaimed')
  if (reason === 'threshold_not_met') return t('leaderboard.dailyReward.thresholdNotMet')
  if (reason === 'not_top_three') return t('leaderboard.dailyReward.notTopThree')
  if (reason === 'not_ranked') return t('leaderboard.dailyReward.notRanked')
  if (reason === 'zero_reward') return t('leaderboard.dailyReward.zeroReward')
  return t('leaderboard.dailyReward.disabled')
})

const dailyRewardStatusClass = computed(() => {
  const reward = dailyRewards.value
  if (reward?.claimed) return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300'
  if (reward?.can_claim) return 'bg-primary-100 text-primary-700 dark:bg-primary-500/20 dark:text-primary-300'
  return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
})

const claimButtonText = computed(() => {
  if (claimingReward.value) return t('leaderboard.dailyReward.claiming')
  if (dailyRewards.value?.claimed) return t('leaderboard.dailyReward.claimed')
  return t('leaderboard.dailyReward.claim')
})

async function loadLeaderboard() {
  const currentSeq = ++loadSeq
  loading.value = true
  error.value = false
  claimError.value = ''
  leaderboard.value = null

  try {
    const response = await usageAPI.getDashboardLeaderboard({ period: period.value, limit: leaderboardLimit })
    if (currentSeq !== loadSeq) return
    leaderboard.value = response
  } catch (err) {
    if (currentSeq !== loadSeq) return
    console.error('Failed to load leaderboard:', err)
    error.value = true
    leaderboard.value = null
  } finally {
    if (currentSeq === loadSeq) {
      loading.value = false
    }
  }
}

async function claimDailyReward() {
  if (!dailyRewards.value?.can_claim || claimingReward.value) return
  claimingReward.value = true
  claimError.value = ''

  try {
    const result = await usageAPI.claimDashboardLeaderboardDailyReward()
    if (leaderboard.value) {
      leaderboard.value.daily_rewards = result.daily_rewards
      applyClaimedBalance(result.claimed_amount)
    }
  } catch (err) {
    console.error('Failed to claim leaderboard daily reward:', err)
    claimError.value = t('leaderboard.dailyReward.claimFailed')
  } finally {
    claimingReward.value = false
  }
}

function applyClaimedBalance(amount: number) {
  const current = myEntry.value
  if (!current || amount <= 0) return
  current.balance = (current.balance ?? 0) + amount
  for (const item of leaderboard.value?.ranking ?? []) {
    if (item.user_id === current.user_id && item !== current) {
      item.balance = (item.balance ?? 0) + amount
    }
  }
}

function selectPeriod(value: LeaderboardPeriod) {
  if (period.value === value) return
  period.value = value
  loadLeaderboard()
}

function formatDateRange(start: string, end: string): string {
  if (!start && !end) return ''
  if (!start) return dateLabel(end)
  if (!end || start === end) return dateLabel(start)
  return `${dateLabel(start)} - ${dateLabel(end)}`
}

function dateLabel(value: string): string {
  return /^\d{4}-\d{2}-\d{2}$/.test(value) ? value : value
}

function getInitial(name: string): string {
  return (name || '?').trim().slice(0, 1).toUpperCase()
}

function formatRewardRankLabel(rank: number): string {
  if (!rank || rank <= 0) return t('leaderboard.notRanked')
  return t('leaderboard.dailyReward.rankLabel', { rank })
}

function isCostEfficiencyKing(item: UserLeaderboardItem): boolean {
  return costEfficiencyKingUserId.value === item.user_id
}

function rankBadgeClass(rank: number): string {
  if (rank === 1) return 'bg-amber-100 text-amber-800 ring-1 ring-amber-300 dark:bg-amber-500/20 dark:text-amber-200 dark:ring-amber-400/50'
  if (rank === 2) return 'bg-slate-100 text-slate-800 ring-1 ring-slate-300 dark:bg-slate-400/20 dark:text-slate-100 dark:ring-slate-300/50'
  if (rank === 3) return 'bg-orange-100 text-orange-800 ring-1 ring-orange-300 dark:bg-orange-500/20 dark:text-orange-200 dark:ring-orange-400/50'
  return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
}

function rankRowClass(rank: number): string {
  if (rank === 1) return 'top-rank-row top-rank-row-gold'
  if (rank === 2) return 'top-rank-row top-rank-row-silver'
  if (rank === 3) return 'top-rank-row top-rank-row-bronze'
  return ''
}

function avatarFrameClass(rank: number): string {
  if (rank === 1) return 'leaderboard-avatar-frame leaderboard-avatar-frame-gold'
  if (rank === 2) return 'leaderboard-avatar-frame leaderboard-avatar-frame-silver'
  if (rank === 3) return 'leaderboard-avatar-frame leaderboard-avatar-frame-bronze'
  return 'shrink-0'
}

onMounted(() => {
  loadLeaderboard()
})
</script>

<style scoped>
.top-rank-row {
  box-shadow: inset 3px 0 0 rgb(148 163 184 / 0.4);
}

.top-rank-row-gold {
  box-shadow: inset 3px 0 0 rgb(245 158 11 / 0.85);
}

.top-rank-row-silver {
  box-shadow: inset 3px 0 0 rgb(148 163 184 / 0.85);
}

.top-rank-row-bronze {
  box-shadow: inset 3px 0 0 rgb(194 120 54 / 0.85);
}

.leaderboard-avatar-frame {
  position: relative;
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  padding: 3px;
}

.leaderboard-avatar-frame::after {
  position: absolute;
  inset: -3px;
  border-radius: inherit;
  content: "";
  opacity: 0.55;
  filter: blur(5px);
}

.leaderboard-avatar-frame > * {
  position: relative;
  z-index: 1;
}

.leaderboard-avatar-frame-gold {
  background: linear-gradient(135deg, #fef3c7, #f59e0b 45%, #92400e);
  box-shadow: 0 0 0 1px rgb(245 158 11 / 0.35), 0 8px 22px rgb(245 158 11 / 0.24);
}

.leaderboard-avatar-frame-gold::after {
  background: rgb(245 158 11 / 0.65);
}

.leaderboard-avatar-frame-silver {
  background: linear-gradient(135deg, #f8fafc, #94a3b8 45%, #475569);
  box-shadow: 0 0 0 1px rgb(148 163 184 / 0.35), 0 8px 22px rgb(148 163 184 / 0.22);
}

.leaderboard-avatar-frame-silver::after {
  background: rgb(148 163 184 / 0.62);
}

.leaderboard-avatar-frame-bronze {
  background: linear-gradient(135deg, #ffedd5, #c27836 45%, #7c2d12);
  box-shadow: 0 0 0 1px rgb(194 120 54 / 0.35), 0 8px 22px rgb(194 120 54 / 0.2);
}

.leaderboard-avatar-frame-bronze::after {
  background: rgb(194 120 54 / 0.58);
}
</style>
