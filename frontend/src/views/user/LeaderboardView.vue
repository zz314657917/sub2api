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
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.totalCost') }}</p>
          <p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ formatCurrency(leaderboard.total_actual_cost) }}</p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.totalRequests') }}</p>
          <p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ formatNumber(leaderboard.total_requests) }}</p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.totalTokens') }}</p>
          <p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ formatNumber(leaderboard.total_tokens) }}</p>
        </div>
      </div>

      <div v-if="leaderboard" class="flex flex-col gap-2 text-sm text-gray-500 dark:text-gray-400 md:flex-row md:items-center md:justify-between">
        <span>{{ formatDateRange(leaderboard.start_date, leaderboard.end_date) }}</span>
        <span>{{ t('leaderboard.generatedAt') }}: {{ formatDateTime(leaderboard.generated_at) }}</span>
      </div>

      <div v-if="showCurrentUserSummary && leaderboard?.current_user_entry" class="card border-primary-200 bg-primary-50/60 p-4 dark:border-primary-500/30 dark:bg-primary-500/10">
        <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div class="min-w-0">
            <p class="text-sm font-semibold text-primary-700 dark:text-primary-300">{{ t('leaderboard.myRank') }}</p>
            <p class="mt-1 truncate text-xl font-bold text-gray-900 dark:text-white">#{{ leaderboard.current_user_entry.rank }} {{ leaderboard.current_user_entry.display_name }}</p>
            <p class="truncate text-sm text-gray-500 dark:text-gray-400">{{ leaderboard.current_user_entry.email_masked }}</p>
          </div>
          <div class="grid shrink-0 grid-cols-3 gap-3 text-right">
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('leaderboard.cost') }}</p>
              <p class="font-semibold text-gray-900 dark:text-white">{{ formatCurrency(leaderboard.current_user_entry.actual_cost) }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('leaderboard.requests') }}</p>
              <p class="font-semibold text-gray-900 dark:text-white">{{ formatNumber(leaderboard.current_user_entry.requests) }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('leaderboard.tokens') }}</p>
              <p class="font-semibold text-gray-900 dark:text-white">{{ formatNumber(leaderboard.current_user_entry.tokens) }}</p>
            </div>
          </div>
        </div>
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

      <div v-else-if="hasLoaded && rankingItems.length === 0" class="card p-8">
        <EmptyState :title="t('leaderboard.emptyTitle')" :description="t('leaderboard.emptyDescription')" />
      </div>

      <template v-else>
        <div class="card hidden overflow-hidden md:block">
          <div class="table-wrapper overflow-x-auto">
            <table class="w-full min-w-[760px] table-fixed">
              <thead class="bg-gray-50 dark:bg-dark-800">
                <tr>
                  <th class="w-24 px-5 py-4 text-left text-sm font-medium text-gray-600 dark:text-dark-300">{{ t('leaderboard.rank') }}</th>
                  <th class="px-5 py-4 text-left text-sm font-medium text-gray-600 dark:text-dark-300">{{ t('leaderboard.user') }}</th>
                  <th class="w-36 px-5 py-4 text-right text-sm font-medium text-gray-600 dark:text-dark-300">{{ t('leaderboard.cost') }}</th>
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
                    item.rank <= 3 ? 'top-rank-row' : '',
                  ]"
                >
                  <td class="px-5 py-4">
                    <span class="inline-flex h-8 min-w-8 items-center justify-center rounded-lg px-2 text-sm font-bold" :class="rankBadgeClass(item.rank)">
                      #{{ item.rank }}
                    </span>
                  </td>
                  <td class="px-5 py-4">
                    <div class="flex min-w-0 items-center gap-3">
                      <img v-if="item.avatar_url" :src="item.avatar_url" alt="" class="h-10 w-10 rounded-full object-cover" />
                      <div v-else class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gray-100 text-sm font-semibold text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                        {{ getInitial(item.display_name) }}
                      </div>
                      <div class="min-w-0">
                        <div class="flex items-center gap-2">
                          <span class="truncate font-semibold text-gray-900 dark:text-white">{{ item.display_name }}</span>
                          <span v-if="item.is_current_user" class="rounded-full bg-primary-100 px-2 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-500/20 dark:text-primary-300">
                            {{ t('leaderboard.currentUser') }}
                          </span>
                        </div>
                        <p class="truncate text-sm text-gray-500 dark:text-gray-400">{{ item.email_masked }}</p>
                      </div>
                    </div>
                  </td>
                  <td class="px-5 py-4 text-right font-semibold text-gray-900 dark:text-white">{{ formatCurrency(item.actual_cost) }}</td>
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
            :class="item.is_current_user ? 'border-primary-200 bg-primary-50/70 dark:border-primary-500/30 dark:bg-primary-500/10' : ''"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="flex min-w-0 items-center gap-3">
                <span class="inline-flex h-9 min-w-9 items-center justify-center rounded-lg px-2 text-sm font-bold" :class="rankBadgeClass(item.rank)">
                  #{{ item.rank }}
                </span>
                <div class="min-w-0">
                  <div class="flex min-w-0 items-center gap-2">
                    <p class="truncate font-semibold text-gray-900 dark:text-white">{{ item.display_name }}</p>
                    <span v-if="item.is_current_user" class="shrink-0 rounded-full bg-primary-100 px-2 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-500/20 dark:text-primary-300">
                      {{ t('leaderboard.currentUser') }}
                    </span>
                  </div>
                  <p class="truncate text-sm text-gray-500 dark:text-gray-400">{{ item.email_masked }}</p>
                </div>
              </div>
              <p class="shrink-0 text-right font-semibold text-gray-900 dark:text-white">{{ formatCurrency(item.actual_cost) }}</p>
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
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usageAPI } from '@/api'
import type { LeaderboardPeriod, UserLeaderboardItem, UserLeaderboardResponse } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { formatCurrency, formatDateTime, formatNumber } from '@/utils/format'

const { t } = useI18n()

const period = ref<LeaderboardPeriod>('day')
const leaderboard = ref<UserLeaderboardResponse | null>(null)
const loading = ref(false)
const error = ref(false)
const hasLoaded = ref(false)
let loadSeq = 0

const periodOptions = computed(() => [
  { value: 'day' as const, label: t('leaderboard.period.day') },
  { value: 'week' as const, label: t('leaderboard.period.week') },
  { value: 'month' as const, label: t('leaderboard.period.month') },
  { value: 'all' as const, label: t('leaderboard.period.all') },
])

const rankingItems = computed<UserLeaderboardItem[]>(() => leaderboard.value?.ranking ?? [])
const showCurrentUserSummary = computed(() => {
  const current = leaderboard.value?.current_user_entry
  if (!current) return false
  return !rankingItems.value.some((item) => item.user_id === current.user_id)
})

async function loadLeaderboard() {
  const currentSeq = ++loadSeq
  loading.value = true
  error.value = false
  leaderboard.value = null

  try {
    const response = await usageAPI.getDashboardLeaderboard({ period: period.value, limit: 20 })
    if (currentSeq !== loadSeq) return
    leaderboard.value = response
    hasLoaded.value = true
  } catch (err) {
    if (currentSeq !== loadSeq) return
    console.error('Failed to load leaderboard:', err)
    error.value = true
    leaderboard.value = null
    hasLoaded.value = true
  } finally {
    if (currentSeq === loadSeq) {
      loading.value = false
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

function rankBadgeClass(rank: number): string {
  if (rank === 1) return 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300'
  if (rank === 2) return 'bg-slate-100 text-slate-700 dark:bg-slate-500/20 dark:text-slate-300'
  if (rank === 3) return 'bg-orange-100 text-orange-700 dark:bg-orange-500/20 dark:text-orange-300'
  return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
}

onMounted(() => {
  loadLeaderboard()
})
</script>

<style scoped>
.top-rank-row {
  box-shadow: inset 3px 0 0 rgb(245 158 11 / 0.7);
}
</style>
