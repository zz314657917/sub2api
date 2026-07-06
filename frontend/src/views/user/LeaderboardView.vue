<template>
  <AppLayout>
    <div class="space-y-6">
      <section v-if="leaderboard" class="card leaderboard-token-card leaderboard-token-summary p-5">
        <div class="leaderboard-token-summary-inner relative z-10">
          <div class="leaderboard-token-summary-main min-w-0">
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.totalTokens') }}</p>
            <div
              :key="rollingTokenAnimationKey"
              class="leaderboard-token-odometer"
              data-testid="leaderboard-total-token-odometer"
              role="text"
              aria-live="polite"
              :aria-label="formatRollingTokenNumber(displayedTotalTokens)"
            >
              <span class="sr-only">{{ formatRollingTokenNumber(displayedTotalTokens) }}</span>
              <span
                v-for="(part, index) in rollingTokenParts"
                :key="`${rollingTokenAnimationKey}-${index}-${part.type}`"
                class="leaderboard-token-part"
                :class="part.type === 'digit' ? 'leaderboard-token-reel' : 'leaderboard-token-separator'"
                :style="part.type === 'digit' ? digitReelStyle(part.value, index) : undefined"
                aria-hidden="true"
              >
                <span v-if="part.type === 'digit'" class="leaderboard-token-strip">
                  <span
                    v-for="(digit, digitIndex) in rollingTokenDigitCells"
                    :key="digitIndex"
                    class="leaderboard-token-cell"
                  >
                    {{ digit }}
                  </span>
                </span>
                <span v-else>{{ part.value }}</span>
              </span>
            </div>
            <div class="leaderboard-token-summary-meta text-sm text-gray-500 dark:text-gray-400">
              <span>{{ t('leaderboard.generatedAt') }} {{ formatTime(leaderboard.generated_at) }}</span>
            </div>
          </div>

          <div class="leaderboard-token-trend-panel" data-testid="leaderboard-recent-token-trend">
            <div class="leaderboard-token-trend-header">
              <span>{{ t('leaderboard.recentTokenTrend.title') }}</span>
              <span class="leaderboard-token-trend-legend">{{ t('leaderboard.recentTokenTrend.unit') }}</span>
            </div>
            <div v-if="recentTokenTrendChartData" class="leaderboard-token-trend-chart">
              <Line :data="recentTokenTrendChartData" :options="recentTokenTrendChartOptions" />
            </div>
            <div v-else class="leaderboard-token-trend-empty">
              {{ t('leaderboard.recentTokenTrend.empty') }}
            </div>
          </div>
        </div>
      </section>

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

      <template v-else-if="leaderboard">
        <div class="grid grid-cols-1 gap-5 xl:grid-cols-[minmax(0,1fr)_22rem]">
          <div class="min-w-0">
            <section v-if="activeRankingEmpty" class="leaderboard-token-ranking-card">
              <div class="leaderboard-ranking-card-toolbar">
                <div class="leaderboard-ranking-switch" role="tablist" :aria-label="t('leaderboard.periodLabel')">
                  <button
                    v-for="option in periodOptions"
                    :key="option.value"
                    type="button"
                    class="leaderboard-ranking-switch-button"
                    :class="period === option.value
                      ? 'leaderboard-ranking-switch-button--active'
                      : 'leaderboard-ranking-switch-button--idle'"
                    :aria-selected="period === option.value"
                    role="tab"
                    @click="selectPeriod(option.value)"
                  >
                    {{ option.label }}
                  </button>
                </div>
                <span class="leaderboard-token-ranking-updated">
                  {{ t('leaderboard.generatedAt') }} {{ formatTime(leaderboard.generated_at) }}
                </span>
              </div>

              <div class="leaderboard-ranking-empty">
                <EmptyState :title="t('leaderboard.emptyTitle')" :description="t('leaderboard.emptyDescription')" />
              </div>
            </section>

            <template v-else>
              <section class="leaderboard-token-ranking-card" data-testid="leaderboard-token-ranking">
                <div class="leaderboard-ranking-card-toolbar">
                  <div class="leaderboard-ranking-switch" role="tablist" :aria-label="t('leaderboard.periodLabel')">
                    <button
                      v-for="option in periodOptions"
                      :key="option.value"
                      type="button"
                      class="leaderboard-ranking-switch-button"
                      :class="period === option.value
                        ? 'leaderboard-ranking-switch-button--active'
                        : 'leaderboard-ranking-switch-button--idle'"
                      :aria-selected="period === option.value"
                      role="tab"
                      @click="selectPeriod(option.value)"
                    >
                      {{ option.label }}
                    </button>
                  </div>
                  <span class="leaderboard-token-ranking-updated">
                    {{ t('leaderboard.generatedAt') }} {{ formatTime(leaderboard.generated_at) }}
                  </span>
                </div>

                <div class="leaderboard-token-rank-list">
                  <article
                    v-for="item in rankingItems"
                    :key="item.user_id"
                    class="leaderboard-token-rank-row"
                    :class="item.is_current_user ? 'leaderboard-token-rank-row-current' : ''"
                    :style="tokenBarStyle(item)"
                  >
                    <div
                      class="leaderboard-token-rank-user"
                      :class="hasLeaderboardRankSecondary(item) ? '' : 'leaderboard-token-rank-user--name-only'"
                    >
                      <div class="leaderboard-token-rank-main">
                        <span class="leaderboard-token-rank-index">#{{ item.rank }}</span>
                        <span class="leaderboard-token-rank-avatar" data-testid="leaderboard-rank-avatar" aria-hidden="true">
                          <img
                            v-if="leaderboardAvatarUrl(item)"
                            :src="leaderboardAvatarUrl(item)"
                            alt=""
                            loading="lazy"
                          >
                          <span v-else>{{ leaderboardAvatarInitial(item) }}</span>
                        </span>
                        <span class="leaderboard-token-rank-name" :title="getLeaderboardDisplayName(item)">
                          {{ getLeaderboardDisplayName(item) }}
                        </span>
                        <span v-if="item.is_current_user" class="leaderboard-token-current-tag">
                          {{ t('leaderboard.currentUser') }}
                        </span>
                      </div>
                      <div v-if="visibleLeaderboardTitleBadges(item.badges).length" class="leaderboard-token-title-list">
                        <span
                          v-for="badge in visibleLeaderboardTitleBadges(item.badges)"
                          :key="badge"
                          class="leaderboard-token-title-badge"
                          :title="leaderboardBadgeTitle(badge)"
                          :aria-label="leaderboardBadgeTitle(badge)"
                          data-testid="leaderboard-rank-title"
                          :data-badge="badge"
                        >
                          {{ leaderboardTitleLabel(badge) }}
                        </span>
                        <span
                          v-if="hiddenLeaderboardTitleBadgeCount(item.badges) > 0"
                          class="leaderboard-token-title-more"
                          :title="hiddenLeaderboardBadgeTitle(item.badges)"
                          :aria-label="hiddenLeaderboardBadgeTitle(item.badges)"
                        >
                          +{{ hiddenLeaderboardTitleBadgeCount(item.badges) }}
                        </span>
                      </div>
                    </div>

                    <div class="leaderboard-token-bar-area">
                      <div
                        class="leaderboard-token-bar-track"
                        :aria-label="leaderboardTokenMetricsLabel(item)"
                        :title="leaderboardTokenMetricsLabel(item)"
                      >
                        <div
                          class="leaderboard-token-bar-fill"
                          data-testid="leaderboard-token-bar-fill"
                        >
                          <span
                            class="leaderboard-token-bar-segment leaderboard-token-bar-segment-input"
                            data-testid="leaderboard-token-segment-input"
                          ></span>
                          <span
                            class="leaderboard-token-bar-segment leaderboard-token-bar-segment-output"
                            data-testid="leaderboard-token-segment-output"
                          ></span>
                          <span
                            class="leaderboard-token-bar-segment leaderboard-token-bar-segment-cache"
                            data-testid="leaderboard-token-segment-cache"
                          ></span>
                        </div>
                        <span class="leaderboard-token-bar-value">{{ formatNumber(item.tokens) }}</span>
                      </div>
                    </div>
                  </article>
                </div>
              </section>
            </template>
          </div>

          <aside class="space-y-5 xl:sticky xl:top-20 xl:self-start">
            <section class="card p-5" data-testid="leaderboard-my-info">
            <p class="text-sm font-semibold text-primary-700 dark:text-primary-300">{{ t('leaderboard.myInfo') }}</p>
            <div v-if="myEntry?.badges?.length" class="mt-3 flex flex-wrap items-center gap-1.5">
              <span
                v-for="badge in visibleLeaderboardBadges(myEntry?.badges)"
                :key="badge"
                class="leaderboard-badge-icon"
                :class="leaderboardBadgeClass(badge)"
                :title="leaderboardBadgeTitle(badge)"
                :aria-label="leaderboardBadgeTitle(badge)"
                data-testid="leaderboard-my-badge-icon"
                :data-badge="badge"
              >
                {{ leaderboardBadgeLabel(badge) }}
              </span>
              <span
                v-if="hiddenLeaderboardBadgeCount(myEntry?.badges) > 0"
                class="leaderboard-badge-overflow"
                :title="hiddenLeaderboardBadgeTitle(myEntry?.badges)"
                :aria-label="hiddenLeaderboardBadgeTitle(myEntry?.badges)"
              >
                +{{ hiddenLeaderboardBadgeCount(myEntry?.badges) }}
              </span>
            </div>
            <div class="mt-4 min-w-0">
              <p class="truncate text-xl font-bold text-gray-900 dark:text-white">
                {{ myRankLabel }} {{ myDisplayName }}
              </p>
            </div>
            <div class="leaderboard-my-token-card mt-4 rounded-lg bg-gray-50 p-3 text-sm dark:bg-dark-800" data-testid="leaderboard-my-token">
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('leaderboard.tokens') }}</p>
              <p class="mt-1 truncate text-xl font-bold text-gray-900 dark:text-white">{{ formatNumber(myEntry?.tokens ?? 0) }}</p>
            </div>
          </section>

          <section v-if="dailyRewards" class="card p-5" data-testid="leaderboard-daily-reward">
            <div class="leaderboard-reward-head">
              <div class="min-w-0">
                <h2 class="leaderboard-reward-title">{{ t('leaderboard.dailyReward.title') }}</h2>
                <p class="leaderboard-reward-period">
                  <span>{{ t('leaderboard.dailyReward.settlementDate') }}</span>
                  <span>{{ dailyRewards.reward_date || '-' }}</span>
                </p>
              </div>
              <span class="leaderboard-reward-status" :class="dailyRewardStatusClass">
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
                <p class="mt-1 text-sm font-bold text-gray-900 dark:text-white">{{ t('leaderboard.dailyReward.rewardAmountHidden') }}</p>
              </div>
            </div>

            <div v-if="rewardTopUsers.length" class="leaderboard-weekly-winners mt-4" data-testid="leaderboard-weekly-winners">
              <div class="leaderboard-weekly-winners-header">
                <span>{{ t('leaderboard.dailyReward.lastWeekTopUsersTitle') }}</span>
                <span>{{ dailyRewards.reward_date || '-' }}</span>
              </div>
              <div class="leaderboard-weekly-winners-list">
                <div
                  v-for="winner in rewardTopUsers"
                  :key="winner.rank"
                  class="leaderboard-weekly-winner-row"
                >
                  <span class="leaderboard-weekly-winner-rank">{{ rewardTopUserRankLabel(winner.rank) }}</span>
                  <span class="leaderboard-weekly-winner-name">{{ winner.displayName }}</span>
                </div>
              </div>
            </div>

            <div class="mt-4 rounded-lg bg-gray-50 p-3 text-sm dark:bg-dark-800">
              <div class="flex items-center justify-between gap-3">
                <span class="text-gray-500 dark:text-gray-400">{{ t('leaderboard.myRank') }}</span>
                <span class="font-semibold text-gray-900 dark:text-white">{{ formatRewardRankLabel(dailyRewards.current_user_rank) }}</span>
              </div>
              <div class="mt-2 flex items-center justify-between gap-3">
                <span class="text-gray-500 dark:text-gray-400">{{ t('leaderboard.dailyReward.targetProgress') }}</span>
                <span class="font-semibold text-gray-900 dark:text-white">{{ dailyRewardGoalProgressText }}</span>
              </div>
              <div
                class="mt-3 h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700"
                :aria-label="`${t('leaderboard.dailyReward.targetProgress')} ${dailyRewardGoalProgressText}`"
              >
                <div
                  class="h-full rounded-full bg-primary-600 transition-all duration-300 dark:bg-primary-400"
                  :style="{ width: dailyRewardGoalProgressWidth }"
                ></div>
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

        <section class="leaderboard-calendar-card" data-testid="leaderboard-daily-champions-calendar">
          <div class="leaderboard-calendar-head">
            <h2 class="leaderboard-calendar-title">{{ t('leaderboard.calendar.title') }}</h2>
          </div>

          <div class="leaderboard-calendar-scroll">
            <div class="leaderboard-calendar-months">
              <section
                v-for="month in championCalendarMonths"
                :key="month.key"
                class="leaderboard-calendar-month-panel"
                :class="month.isCurrent ? 'leaderboard-calendar-month-panel--current' : 'leaderboard-calendar-month-panel--previous'"
              >
                <div class="leaderboard-calendar-month">{{ month.label }}</div>
                <div class="leaderboard-calendar-weekdays" aria-hidden="true">
                  <span
                    v-for="weekday in championCalendarWeekdays"
                    :key="`${month.key}-${weekday}`"
                    class="leaderboard-calendar-weekday"
                  >
                    {{ weekday }}
                  </span>
                </div>
                <div class="leaderboard-calendar-days">
                  <div
                    v-for="day in month.days"
                    :key="day.key"
                    class="leaderboard-calendar-day"
                    :class="{
                      'leaderboard-calendar-day--active': !day.isPlaceholder && day.champion,
                      'leaderboard-calendar-day--placeholder': day.isPlaceholder
                    }"
                    :aria-label="championCalendarCellLabel(day)"
                    :tabindex="championCalendarCellTabIndex(day)"
                    :data-testid="day.isPlaceholder ? 'leaderboard-calendar-placeholder' : 'leaderboard-calendar-day'"
                    @mouseenter="showChampionTooltip(day, $event)"
                    @mouseleave="hideChampionTooltip"
                    @focusin="showChampionTooltip(day, $event)"
                    @focusout="hideChampionTooltip"
                  >
                    <template v-if="!day.isPlaceholder">
                    <span class="leaderboard-calendar-day-number">{{ day.day }}</span>
                    <span v-if="day.champion" class="leaderboard-calendar-avatar" data-testid="leaderboard-calendar-avatar">
                      <img
                        v-if="dailyChampionAvatarUrl(day.champion)"
                        :src="dailyChampionAvatarUrl(day.champion)"
                        alt=""
                        loading="lazy"
                      >
                      <span v-else>{{ dailyChampionAvatarInitial(day.champion) }}</span>
                    </span>
                    </template>
                  </div>
                </div>
              </section>
            </div>
          </div>

          <Teleport to="body">
            <div
              v-if="activeChampionTooltip"
              class="leaderboard-calendar-tooltip"
              :style="championTooltipStyle"
              role="tooltip"
              data-testid="leaderboard-calendar-tooltip"
            >
              <div class="leaderboard-calendar-tooltip-date">{{ activeChampionTooltip.dateLabel }}</div>
              <div class="leaderboard-calendar-tooltip-main">
                <span class="leaderboard-calendar-tooltip-name">{{ activeChampionTooltip.displayName }}</span>
                <span class="leaderboard-calendar-tooltip-tokens">{{ activeChampionTooltip.tokenLabel }}</span>
              </div>
              <div class="leaderboard-calendar-tooltip-meta">{{ activeChampionTooltip.metaLabel }}</div>
            </div>
          </Teleport>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Filler,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import { usageAPI } from '@/api'
import type { LeaderboardBadge, LeaderboardDailyRewardTopUser, LeaderboardDailyRewards, LeaderboardPeriod, UserLeaderboardDailyChampion, UserLeaderboardItem, UserLeaderboardResponse } from '@/api/usage'
import type { UserLeaderboardTokenTrendPoint } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { formatDateTime, formatNumber, formatTime } from '@/utils/format'
import { formatCreditAmount } from '@/utils/credits'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler)

const { t } = useI18n()

const period = ref<LeaderboardPeriod>('day')
const leaderboard = ref<UserLeaderboardResponse | null>(null)
const loading = ref(false)
const error = ref(false)
const claimingReward = ref(false)
const claimError = ref('')
const tokenTickerSeed = ref(0)
const visualTokenIncrement = ref(0)
const visualTokenTick = ref(0)
const leaderboardLimit = 10
const visibleBadgeLimit = 3
const visibleRankTitleLimit = 2
const visualTokenTickerIntervalMs = 3000
const visualTokenTickerSteps = [37, 54, 62, 81, 95, 128, 143, 166, 218]
const championTooltipWidth = 240
const championTooltipHeight = 98
const championTooltipGap = 10
const championTooltipViewportPadding = 12
let loadSeq = 0
let visualTokenTickerID: number | null = null

type RollingTokenPart = {
  type: 'digit' | 'separator'
  value: string
}

type RewardTopUserView = {
  rank: number
  displayName: string
}

type ChampionCalendarDay = {
  key: string
  date: string
  day: number
  champion?: UserLeaderboardDailyChampion
  streak: number
  isPlaceholder: boolean
}

type ChampionCalendarMonth = {
  key: string
  label: string
  days: ChampionCalendarDay[]
  isCurrent: boolean
}

type ChampionCalendarTooltipView = {
  dateLabel: string
  displayName: string
  metaLabel: string
  tokenLabel: string
}

const rollingTokenDigitCells = Array.from({ length: 20 }, (_, index) => String(index % 10))
const leaderboardTitleBadges: LeaderboardBadge[] = [
  'weekly_token_king',
  'monthly_token_king',
  'total_token_king',
  'night_owl',
  'burst_token_king',
  'checkin_king',
]
const devRewardTopUsers: LeaderboardDailyRewardTopUser[] = [
  { rank: 1, display_name: '落***尘' },
  { rank: 2, display_name: '138****5678' },
  { rank: 3, display_name: 't***d@example.com' },
]

const periodOptions = computed(() => [
  { value: 'day' as const, label: t('leaderboard.period.day') },
  { value: 'week' as const, label: t('leaderboard.period.week') },
  { value: 'month' as const, label: t('leaderboard.period.month') },
  { value: 'all' as const, label: t('leaderboard.period.all') },
])
const championCalendarWeekdays = computed(() => [
  t('leaderboard.calendar.weekdays.sun'),
  t('leaderboard.calendar.weekdays.mon'),
  t('leaderboard.calendar.weekdays.tue'),
  t('leaderboard.calendar.weekdays.wed'),
  t('leaderboard.calendar.weekdays.thu'),
  t('leaderboard.calendar.weekdays.fri'),
  t('leaderboard.calendar.weekdays.sat'),
])
const recentTokenTrendPoints = computed<UserLeaderboardTokenTrendPoint[]>(() => leaderboard.value?.recent_token_trend ?? [])
const dailyChampionsByDate = computed(() => {
  const champions = new Map<string, UserLeaderboardDailyChampion>()
  for (const champion of leaderboard.value?.daily_champions ?? []) {
    if (champion.date) {
      champions.set(champion.date, champion)
    }
  }
  return champions
})
const championCalendarAnchor = computed(() => parseCalendarAnchorDate(leaderboard.value?.generated_at))
const championCalendarMonths = computed<ChampionCalendarMonth[]>(() => {
  const anchor = championCalendarAnchor.value
  return [
    buildChampionCalendarMonth(new Date(anchor.getFullYear(), anchor.getMonth() - 1, 1), false),
    buildChampionCalendarMonth(new Date(anchor.getFullYear(), anchor.getMonth(), 1), true),
  ]
})
const activeChampionTooltip = ref<ChampionCalendarTooltipView | null>(null)
const championTooltipPosition = ref({ left: 0, top: 0 })
const championTooltipStyle = computed(() => ({
  left: `${championTooltipPosition.value.left}px`,
  top: `${championTooltipPosition.value.top}px`,
}))

const rankingItems = computed<UserLeaderboardItem[]>(() => {
  const visibleItems = (leaderboard.value?.ranking ?? []).slice(0, leaderboardLimit)
  const costSaverUserID = selectVisibleCostEfficiencyUserID(visibleItems, 'lowest')
  const costBurnerUserID = selectVisibleCostEfficiencyUserID(visibleItems, 'highest')

  return visibleItems.map((item) => ({
    ...item,
    badges: orderedLeaderboardBadges(item, costSaverUserID, costBurnerUserID),
  }))
})
const maxRankingTokens = computed(() => Math.max(0, ...rankingItems.value.map((item) => item.tokens)))
const activeRankingEmpty = computed(() => rankingItems.value.length === 0)
const dailyRewards = computed<LeaderboardDailyRewards | null>(() => leaderboard.value?.daily_rewards ?? null)
const myEntry = computed<UserLeaderboardItem | null>(() => {
  if (leaderboard.value?.current_user_entry) return leaderboard.value.current_user_entry
  return rankingItems.value.find((item) => item.is_current_user) ?? null
})
const myDisplayName = computed(() => (myEntry.value ? getLeaderboardDisplayName(myEntry.value) : t('leaderboard.currentUser')))
const myRankLabel = computed(() => (myEntry.value?.rank ? `#${myEntry.value.rank}` : t('leaderboard.notRanked')))
const rollingTokenParts = computed<RollingTokenPart[]>(() => {
  return formatRollingTokenNumber(displayedTotalTokens.value)
    .split('')
    .map((value) => ({
      type: /^\d$/.test(value) ? 'digit' : 'separator',
      value,
    }))
})
const displayedTotalTokens = computed(() => {
  const baseTokens = leaderboard.value?.total_tokens ?? 0
  return Math.max(0, Math.floor(baseTokens + visualTokenIncrement.value))
})
const rollingTokenAnimationKey = computed(() => `${period.value}-${tokenTickerSeed.value}-${leaderboard.value?.total_tokens ?? 0}`)

const recentTokenTrendChartColors = computed(() => {
  const isDark = document.documentElement.classList.contains('dark')
  return {
    line: isDark ? '#60a5fa' : '#2563eb',
    fill: isDark ? 'rgba(37, 99, 235, 0.24)' : 'rgba(37, 99, 235, 0.16)',
    text: isDark ? '#dbeafe' : '#1e293b',
    muted: isDark ? '#94a3b8' : '#64748b',
    grid: isDark ? 'rgba(148, 163, 184, 0.14)' : 'rgba(100, 116, 139, 0.16)',
    tooltipBg: isDark ? 'rgba(8, 13, 26, 0.96)' : 'rgba(255, 255, 255, 0.96)',
    tooltipBorder: isDark ? 'rgba(96, 165, 250, 0.34)' : 'rgba(37, 99, 235, 0.18)',
  }
})

const recentTokenTrendChartData = computed(() => {
  const points = recentTokenTrendPoints.value
  if (!points.length || !points.some((point) => point.total_tokens > 0)) return null

  return {
    labels: points.map((point) => formatTrendDateLabel(point.date)),
    datasets: [
      {
        label: t('leaderboard.recentTokenTrend.tokens'),
        data: points.map((point) => point.total_tokens),
        borderColor: recentTokenTrendChartColors.value.line,
        backgroundColor: recentTokenTrendChartColors.value.fill,
        fill: true,
        borderWidth: 2.2,
        pointRadius: 0,
        pointHoverRadius: 3,
        tension: 0.36,
      },
    ],
  }
})

const recentTokenTrendChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const,
  },
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      backgroundColor: recentTokenTrendChartColors.value.tooltipBg,
      borderColor: recentTokenTrendChartColors.value.tooltipBorder,
      borderWidth: 1,
      titleColor: recentTokenTrendChartColors.value.text,
      bodyColor: recentTokenTrendChartColors.value.text,
      displayColors: false,
      padding: 9,
      callbacks: {
        title: (items: any[]) => {
          const index = items[0]?.dataIndex
          return recentTokenTrendPoints.value[index]?.date ?? ''
        },
        label: (context: any) => `${t('leaderboard.recentTokenTrend.tokens')}: ${formatNumber(Number(context.raw ?? 0))}`,
      },
    },
  },
  scales: {
    x: {
      grid: {
        color: recentTokenTrendChartColors.value.grid,
      },
      ticks: {
        color: recentTokenTrendChartColors.value.muted,
        maxRotation: 0,
        font: {
          size: 10,
          weight: 600,
        },
      },
    },
    y: {
      beginAtZero: true,
      grid: {
        color: recentTokenTrendChartColors.value.grid,
      },
      ticks: {
        color: recentTokenTrendChartColors.value.muted,
        font: {
          size: 10,
          weight: 600,
        },
        callback: (value: string | number) => formatCompactTokens(Number(value)),
      },
    },
  },
}))

const rewardTiers = computed(() => {
  return [1, 2, 3].map((rank) => ({ rank }))
})
const rewardTopUsers = computed<RewardTopUserView[]>(() => {
  const usersByRank = new Map<number, LeaderboardDailyRewardTopUser>()
  const sourceUsers = dailyRewards.value?.top_users?.length
    ? dailyRewards.value.top_users
    : import.meta.env.DEV
      ? devRewardTopUsers
      : []
  for (const user of sourceUsers) {
    if (user.rank >= 1 && user.rank <= 3 && rewardTopUserHasName(user)) {
      usersByRank.set(user.rank, user)
    }
  }
  return [1, 2, 3]
    .map((rank) => {
      const user = usersByRank.get(rank)
      if (!user) return null
      return {
        rank,
        displayName: rewardTopUserDisplayName(user),
      }
    })
    .filter((user): user is RewardTopUserView => user != null)
})

const dailyRewardGoalProgressPercent = computed(() => {
  const reward = dailyRewards.value
  if (!reward) return 0

  const target = Number(reward.min_total_actual_cost)
  const current = Number(reward.yesterday_total_actual_cost)
  if (!Number.isFinite(target) || target <= 0) {
    return reward.threshold_met ? 100 : 0
  }
  if (reward.threshold_met) return 100
  if (!Number.isFinite(current) || current <= 0) return 0

  return Math.min(99, Math.max(0, Math.floor((current / target) * 100)))
})
const dailyRewardGoalProgressText = computed(() =>
  t('leaderboard.dailyReward.progressPercent', { percent: dailyRewardGoalProgressPercent.value })
)
const dailyRewardGoalProgressWidth = computed(() => `${dailyRewardGoalProgressPercent.value}%`)

const dailyRewardReasonText = computed(() => {
  const reason = dailyRewards.value?.reason
  if (reason === 'eligible') return t('leaderboard.dailyReward.eligible')
  if (reason === 'already_claimed') return t('leaderboard.dailyReward.alreadyClaimed')
  if (reason === 'settling') return t('leaderboard.dailyReward.settling', { time: formatDateTime(dailyRewards.value?.claim_available_at || '') })
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
    visualTokenIncrement.value = 0
    tokenTickerSeed.value += 1
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
  hideChampionTooltip()
  loadLeaderboard()
}

function showChampionTooltip(day: ChampionCalendarDay, event: Event) {
  if (day.isPlaceholder || !day.champion) {
    hideChampionTooltip()
    return
  }

  const target = event.currentTarget
  if (!(target instanceof HTMLElement)) return

  activeChampionTooltip.value = {
    dateLabel: formatChampionCalendarDate(day.date),
    displayName: dailyChampionDisplayName(day.champion),
    metaLabel: championTooltipMeta(day),
    tokenLabel: `${formatCompactChineseTokens(day.champion.tokens)} tokens`,
  }
  positionChampionTooltip(target)
}

function championTooltipMeta(day: ChampionCalendarDay): string {
  if (!day.champion) return ''
  const displayName = dailyChampionDisplayName(day.champion)
  const emailMasked = day.champion.email_masked?.trim() || ''
  if (emailMasked && emailMasked !== displayName) return emailMasked
  if (day.streak > 1) return `蝉联第 ${day.streak} 天`
  return '当日冠军'
}

function positionChampionTooltip(target: HTMLElement) {
  const rect = target.getBoundingClientRect()
  const viewportWidth = window.innerWidth || document.documentElement.clientWidth || championTooltipWidth
  const maxLeft = Math.max(championTooltipViewportPadding, viewportWidth - championTooltipWidth - championTooltipViewportPadding)
  const preferredLeft = rect.left + window.scrollX - 12
  const left = Math.min(Math.max(championTooltipViewportPadding, preferredLeft), maxLeft)
  const top = Math.max(
    championTooltipViewportPadding,
    rect.top + window.scrollY - championTooltipGap - championTooltipHeight,
  )
  championTooltipPosition.value = { left, top }
}

function hideChampionTooltip() {
  activeChampionTooltip.value = null
}

function formatRollingTokenNumber(value: number): string {
  if (!Number.isFinite(value)) return '0'
  return Math.max(0, Math.floor(value)).toLocaleString('en-US')
}

function formatTrendDateLabel(value: string): string {
  const [, month = '', day = ''] = value.match(/^\d{4}-(\d{2})-(\d{2})/) ?? []
  if (!month || !day) return value
  return `${Number(month)}/${Number(day)}`
}

function formatCompactTokens(value: number): string {
  if (!Number.isFinite(value)) return '0'
  const abs = Math.abs(value)
  if (abs >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(1)}B`
  if (abs >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`
  if (abs >= 1_000) return `${(value / 1_000).toFixed(1)}K`
  return Math.round(value).toLocaleString('en-US')
}

function parseCalendarAnchorDate(value?: string): Date {
  const parsed = value ? new Date(value) : null
  if (parsed && Number.isFinite(parsed.getTime())) {
    return parsed
  }
  return new Date()
}

function buildChampionCalendarMonth(monthStart: Date, isCurrent: boolean): ChampionCalendarMonth {
  const year = monthStart.getFullYear()
  const month = monthStart.getMonth()
  const daysInMonth = new Date(year, month + 1, 0).getDate()
  const key = formatCalendarDateKey(monthStart)
  const leadingPlaceholderCount = new Date(year, month, 1).getDay()
  let previousChampionUserID: number | null = null
  let streak = 0
  const realDays = Array.from({ length: daysInMonth }, (_, index) => {
    const date = new Date(year, month, index + 1)
    const dateKey = formatCalendarDateKey(date)
    const champion = dailyChampionsByDate.value.get(dateKey)
    if (champion) {
      streak = previousChampionUserID === champion.user_id ? streak + 1 : 1
      previousChampionUserID = champion.user_id
    } else {
      streak = 0
      previousChampionUserID = null
    }
    return {
      key: dateKey,
      date: dateKey,
      day: index + 1,
      champion,
      streak,
      isPlaceholder: false,
    }
  })
  const trailingPlaceholderCount = (7 - ((leadingPlaceholderCount + daysInMonth) % 7)) % 7

  return {
    key,
    label: `${year}年${month + 1}月`,
    isCurrent,
    days: [
      ...Array.from({ length: leadingPlaceholderCount }, (_, index) => createChampionCalendarPlaceholder(`${key}-leading-${index}`)),
      ...realDays,
      ...Array.from({ length: trailingPlaceholderCount }, (_, index) => createChampionCalendarPlaceholder(`${key}-trailing-${index}`)),
    ],
  }
}

function createChampionCalendarPlaceholder(key: string): ChampionCalendarDay {
  return {
    key,
    date: '',
    day: 0,
    streak: 0,
    isPlaceholder: true,
  }
}

function formatCalendarDateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function formatChampionCalendarDate(value: string): string {
  const [, year = '', month = '', day = ''] = value.match(/^(\d{4})-(\d{2})-(\d{2})$/) ?? []
  if (!year || !month || !day) return value
  return `${year}年${Number(month)}月${Number(day)}日`
}

function formatCompactChineseTokens(value: number): string {
  if (!Number.isFinite(value)) return '0'
  const abs = Math.abs(value)
  if (abs >= 100_000_000) return `${formatCompactUnit(value / 100_000_000)}亿`
  if (abs >= 10_000) return `${formatCompactUnit(value / 10_000)}万`
  return formatNumber(Math.max(0, Math.round(value)))
}

function formatCompactUnit(value: number): string {
  return value.toFixed(1).replace(/\.0$/, '')
}

function dailyChampionDisplayName(champion: UserLeaderboardDailyChampion): string {
  return champion.display_name?.trim() || champion.email_masked?.trim() || t('leaderboard.currentUser')
}

function dailyChampionAvatarUrl(champion: UserLeaderboardDailyChampion): string {
  return champion.avatar_url?.trim() || ''
}

function dailyChampionAvatarInitial(champion: UserLeaderboardDailyChampion): string {
  return Array.from(dailyChampionDisplayName(champion).trim())[0]?.toUpperCase() || 'U'
}

function championCalendarTooltip(day: ChampionCalendarDay): string {
  if (!day.champion) return championCalendarEmptyLabel(day)
  return [
    formatChampionCalendarDate(day.date),
    dailyChampionDisplayName(day.champion),
    `${formatCompactChineseTokens(day.champion.tokens)} tokens`,
  ].join('\n')
}

function championCalendarEmptyLabel(day: ChampionCalendarDay): string {
  return t('leaderboard.calendar.emptyDay', { date: formatChampionCalendarDate(day.date) })
}

function championCalendarCellLabel(day: ChampionCalendarDay): string | undefined {
  if (day.isPlaceholder) return undefined
  return day.champion ? championCalendarTooltip(day) : championCalendarEmptyLabel(day)
}

function championCalendarCellTabIndex(day: ChampionCalendarDay): number | undefined {
  return !day.isPlaceholder && day.champion ? 0 : undefined
}

function startVisualTokenTicker() {
  stopVisualTokenTicker()
  visualTokenTickerID = window.setInterval(() => {
    if (!leaderboard.value || loading.value || error.value || document.hidden) return

    const baseTokens = Math.max(0, leaderboard.value.total_tokens)
    const step = visualTokenTickerSteps[visualTokenTick.value % visualTokenTickerSteps.length]
    const scale = baseTokens >= 10_000_000 ? 4 : baseTokens >= 1_000_000 ? 2 : 1
    visualTokenTick.value += 1
    visualTokenIncrement.value += step * scale
  }, visualTokenTickerIntervalMs)
}

function stopVisualTokenTicker() {
  if (visualTokenTickerID == null) return
  window.clearInterval(visualTokenTickerID)
  visualTokenTickerID = null
}

function tokenBarWidth(item: UserLeaderboardItem): string {
  if (maxRankingTokens.value <= 0 || item.tokens <= 0) return '0%'
  return `${Math.min(84, Math.max(4, (item.tokens / maxRankingTokens.value) * 84))}%`
}

function leaderboardCacheTokens(item: UserLeaderboardItem): number {
  return Math.max(0, (item.cache_creation_tokens ?? 0) + (item.cache_read_tokens ?? 0))
}

function leaderboardInputTokens(item: UserLeaderboardItem): number {
  return Math.max(0, item.input_tokens ?? 0)
}

function leaderboardOutputTokens(item: UserLeaderboardItem): number {
  return Math.max(0, item.output_tokens ?? 0)
}

function leaderboardCachePercent(item: UserLeaderboardItem): number {
  if (item.tokens <= 0) return 0
  return Math.min(100, (leaderboardCacheTokens(item) / item.tokens) * 100)
}

function leaderboardTokenSegmentPercent(value: number, total: number): string {
  if (total <= 0 || value <= 0) return '0%'
  return `${Math.min(100, (value / total) * 100).toFixed(1)}%`
}

function formatLeaderboardCachePercent(item: UserLeaderboardItem): string {
  return `${leaderboardCachePercent(item).toFixed(1)}%`
}

function tokenBarStyle(item: UserLeaderboardItem): Record<string, string> {
  const palette = tokenBarPalette(item.rank)
  const widthText = tokenBarWidth(item)

  return {
    '--token-bar-width': widthText,
    '--token-input-width': leaderboardTokenSegmentPercent(leaderboardInputTokens(item), item.tokens),
    '--token-output-width': leaderboardTokenSegmentPercent(leaderboardOutputTokens(item), item.tokens),
    '--token-cache-width': leaderboardTokenSegmentPercent(leaderboardCacheTokens(item), item.tokens),
    '--token-bar-value-left': `calc(${widthText} + 0.55rem)`,
    '--token-bar-value-x': '0',
    '--token-bar-color': palette.color,
    '--token-bar-glow': palette.glow,
    '--token-rank-color': palette.text,
    '--token-value-color': palette.text,
  }
}

function tokenBarPalette(rank: number): { color: string; glow: string; text: string } {
  if (rank === 1) return { color: 'rgb(217 119 6)', glow: 'rgb(217 119 6 / 0.22)', text: 'rgb(146 64 14)' }
  if (rank === 2) return { color: 'rgb(5 150 105)', glow: 'rgb(5 150 105 / 0.2)', text: 'rgb(4 120 87)' }
  if (rank === 3) return { color: 'rgb(37 99 235)', glow: 'rgb(37 99 235 / 0.18)', text: 'rgb(29 78 216)' }
  return { color: 'rgb(100 116 139)', glow: 'rgb(100 116 139 / 0.14)', text: 'rgb(71 85 105)' }
}

function digitReelStyle(value: string, _index: number): Record<string, string> {
  const digit = Number.parseInt(value, 10)
  const targetIndex = Number.isFinite(digit) ? 10 + digit : 10

  return {
    '--target-offset': `${targetIndex * -1.08}em`,
  }
}

function getLeaderboardDisplayName(item: UserLeaderboardItem): string {
  return item.display_name?.trim() || item.email_masked?.trim() || t('leaderboard.currentUser')
}

function formatLeaderboardCostPerMillion(item: UserLeaderboardItem): string {
  const value = Number.isFinite(item.cost_per_1m_tokens) && item.cost_per_1m_tokens > 0
    ? item.cost_per_1m_tokens
    : item.tokens > 0
      ? (item.actual_cost / item.tokens) * 1_000_000
      : 0
  return formatCreditAmount(value, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

function leaderboardTokenMetricsLabel(item: UserLeaderboardItem): string {
  return [
    `${t('leaderboard.inputTokensShort')} ${formatNumber(item.input_tokens ?? 0)}`,
    `${t('leaderboard.outputTokensShort')} ${formatNumber(item.output_tokens ?? 0)}`,
    `${t('leaderboard.cacheTokensShort')} ${formatNumber(leaderboardCacheTokens(item))} (${t('leaderboard.cacheRatioShort')} ${formatLeaderboardCachePercent(item)})`,
    `${t('leaderboard.costPerMillionShort')} ${formatLeaderboardCostPerMillion(item)} / 1M Token`,
  ].join(' / ')
}

function leaderboardAvatarUrl(item: UserLeaderboardItem): string {
  return item.avatar_url?.trim() || ''
}

function leaderboardAvatarInitial(item: UserLeaderboardItem): string {
  return Array.from(getLeaderboardDisplayName(item).trim())[0]?.toUpperCase() || 'U'
}

function formatRewardRankLabel(rank: number): string {
  if (!rank || rank <= 0) return t('leaderboard.notRanked')
  return t('leaderboard.dailyReward.rankLabel', { rank })
}

function rewardTopUserRankLabel(rank: number): string {
  if (rank === 1) return t('leaderboard.dailyReward.lastWeekRank1')
  if (rank === 2) return t('leaderboard.dailyReward.lastWeekRank2')
  if (rank === 3) return t('leaderboard.dailyReward.lastWeekRank3')
  return t('leaderboard.dailyReward.lastWeekRankLabel', { rank })
}

function rewardTopUserDisplayName(user?: LeaderboardDailyRewardTopUser): string {
  const displayName = user?.display_name?.trim()
  if (displayName) return displayName
  const emailMasked = user?.email_masked?.trim()
  if (emailMasked) return emailMasked
  return t('leaderboard.dailyReward.noTopUser')
}

function rewardTopUserHasName(user?: LeaderboardDailyRewardTopUser): boolean {
  return Boolean(user?.display_name?.trim() || user?.email_masked?.trim())
}

function selectVisibleCostEfficiencyUserID(items: UserLeaderboardItem[], mode: 'lowest' | 'highest'): number | null {
  let selected: UserLeaderboardItem | null = null
  let selectedCostPerToken = mode === 'lowest' ? Number.POSITIVE_INFINITY : Number.NEGATIVE_INFINITY

  for (const item of items) {
    if (item.actual_cost <= 0 || item.tokens <= 0) continue

    const costPerToken = item.actual_cost / item.tokens
    const isBetter = mode === 'lowest'
      ? costPerToken < selectedCostPerToken
      : costPerToken > selectedCostPerToken
    const isTieBreaker = selected != null
      && costPerToken === selectedCostPerToken
      && (item.tokens > selected.tokens || (item.tokens === selected.tokens && item.user_id < selected.user_id))

    if (isBetter || isTieBreaker) {
      selected = item
      selectedCostPerToken = costPerToken
    }
  }

  return selected?.user_id ?? null
}

function orderedLeaderboardBadges(item: UserLeaderboardItem, costSaverUserID: number | null, costBurnerUserID: number | null): LeaderboardBadge[] {
  const badgeSet = new Set<LeaderboardBadge>(item.badges ?? [])
  if (item.user_id === costSaverUserID) badgeSet.add('cost_saver')
  if (item.user_id === costBurnerUserID) badgeSet.add('cost_burner')

  return ([
    'weekly_token_king',
    'monthly_token_king',
    'total_token_king',
    'night_owl',
    'burst_token_king',
    'checkin_king',
    'cost_saver',
    'cost_burner',
  ] as LeaderboardBadge[]).filter((badge) => badgeSet.has(badge))
}

function visibleLeaderboardBadges(badges: LeaderboardBadge[] = []): LeaderboardBadge[] {
  return badges.slice(0, visibleBadgeLimit)
}

function hiddenLeaderboardBadges(badges: LeaderboardBadge[] = []): LeaderboardBadge[] {
  return badges.slice(visibleBadgeLimit)
}

function hiddenLeaderboardBadgeCount(badges: LeaderboardBadge[] = []): number {
  return hiddenLeaderboardBadges(badges).length
}

function hiddenLeaderboardBadgeTitle(badges: LeaderboardBadge[] = []): string {
  return hiddenLeaderboardBadges(badges).map((badge) => leaderboardBadgeTitle(badge)).join(' / ')
}

function visibleLeaderboardTitleBadges(badges: LeaderboardBadge[] = []): LeaderboardBadge[] {
  return badges.filter((badge) => leaderboardTitleBadges.includes(badge)).slice(0, visibleRankTitleLimit)
}

function hasLeaderboardRankSecondary(item: UserLeaderboardItem): boolean {
  return Boolean(item.is_current_user) || visibleLeaderboardTitleBadges(item.badges).length > 0
}

function hiddenLeaderboardTitleBadgeCount(badges: LeaderboardBadge[] = []): number {
  return badges.filter((badge) => leaderboardTitleBadges.includes(badge)).slice(visibleRankTitleLimit).length
}

function leaderboardBadgeLabel(badge: LeaderboardBadge): string {
  if (badge === 'weekly_token_king') return '周'
  if (badge === 'monthly_token_king') return '月'
  if (badge === 'total_token_king') return '肝'
  if (badge === 'night_owl') return '夜'
  if (badge === 'burst_token_king') return '爆'
  if (badge === 'checkin_king') return '勤'
  if (badge === 'cost_saver') return '省'
  if (badge === 'cost_burner') return '豪'
  return ''
}

function leaderboardTitleLabel(badge: LeaderboardBadge): string {
  if (badge === 'weekly_token_king') return '周榜王'
  if (badge === 'monthly_token_king') return '月榜王'
  if (badge === 'total_token_king') return '肝帝'
  if (badge === 'night_owl') return '夜猫'
  if (badge === 'burst_token_king') return '爆肝'
  if (badge === 'checkin_king') return '打卡王'
  return leaderboardBadgeLabel(badge)
}

function leaderboardBadgeTitle(badge: LeaderboardBadge): string {
  if (badge === 'weekly_token_king') return t('leaderboard.badges.weeklyTokenKing')
  if (badge === 'monthly_token_king') return t('leaderboard.badges.monthlyTokenKing')
  if (badge === 'total_token_king') return t('leaderboard.badges.totalTokenKing')
  if (badge === 'night_owl') return t('leaderboard.badges.nightOwl')
  if (badge === 'burst_token_king') return t('leaderboard.badges.burstTokenKing')
  if (badge === 'checkin_king') return t('leaderboard.badges.checkinKing')
  if (badge === 'cost_saver') return t('leaderboard.badges.costSaver')
  if (badge === 'cost_burner') return t('leaderboard.badges.costBurner')
  return ''
}

function leaderboardBadgeClass(badge: LeaderboardBadge): string {
  if (badge === 'weekly_token_king') return 'leaderboard-badge-week'
  if (badge === 'monthly_token_king') return 'leaderboard-badge-month'
  if (badge === 'total_token_king') return 'leaderboard-badge-total'
  if (badge === 'night_owl') return 'leaderboard-badge-night'
  if (badge === 'burst_token_king') return 'leaderboard-badge-burst'
  if (badge === 'checkin_king') return 'leaderboard-badge-checkin'
  if (badge === 'cost_saver') return 'leaderboard-badge-save'
  if (badge === 'cost_burner') return 'leaderboard-badge-fire'
  return ''
}

onMounted(() => {
  startVisualTokenTicker()
  loadLeaderboard()
  window.addEventListener('scroll', hideChampionTooltip, true)
  window.addEventListener('resize', hideChampionTooltip)
})

onUnmounted(() => {
  stopVisualTokenTicker()
  hideChampionTooltip()
  window.removeEventListener('scroll', hideChampionTooltip, true)
  window.removeEventListener('resize', hideChampionTooltip)
})
</script>

<style scoped>
.leaderboard-token-summary {
  min-height: 10.8rem;
}

.leaderboard-token-summary-inner {
  display: grid;
  min-height: 8.6rem;
  grid-template-columns: minmax(18rem, 0.92fr) minmax(20rem, 1.08fr);
  align-items: center;
  gap: 1.35rem;
  text-align: left;
}

.leaderboard-token-summary-main {
  display: grid;
  min-width: 0;
  gap: 0.28rem;
  justify-items: center;
  text-align: center;
}

.leaderboard-token-summary-meta {
  justify-items: center;
  color: rgb(100 116 139);
  font-size: 0.8125rem;
  letter-spacing: 0;
}

.leaderboard-token-card {
  position: relative;
  overflow: hidden;
  border-color: rgb(148 163 184 / 0.18);
  background:
    radial-gradient(circle at 50% 8%, rgb(251 191 36 / 0.22), transparent 34%),
    linear-gradient(180deg, rgb(255 251 235 / 0.76), rgb(255 255 255 / 0.94));
  box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.7);
}

.leaderboard-token-trend-panel {
  display: grid;
  min-width: 0;
  gap: 0.52rem;
  border-left: 1px solid rgb(148 163 184 / 0.22);
  padding-left: 1.35rem;
}

.leaderboard-token-trend-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: rgb(71 85 105);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0;
}

.leaderboard-token-trend-legend {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: rgb(37 99 235);
  font-size: 0.72rem;
  white-space: nowrap;
}

.leaderboard-token-trend-legend::before {
  width: 0.55rem;
  height: 0.55rem;
  border: 1px solid currentColor;
  border-radius: 9999px;
  background: rgb(37 99 235 / 0.12);
  content: "";
}

.leaderboard-token-trend-chart,
.leaderboard-token-trend-empty {
  height: 6.7rem;
  min-width: 0;
}

.leaderboard-token-trend-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgb(148 163 184 / 0.16);
  background:
    linear-gradient(90deg, rgb(148 163 184 / 0.08) 1px, transparent 1px),
    linear-gradient(rgb(148 163 184 / 0.08) 1px, transparent 1px);
  background-size: 12.5% 100%, 100% 1.65rem;
  color: rgb(100 116 139);
  font-size: 0.8125rem;
}

.leaderboard-token-card::after {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(90deg, rgb(148 163 184 / 0.09) 1px, transparent 1px),
    linear-gradient(rgb(148 163 184 / 0.08) 1px, transparent 1px),
    linear-gradient(90deg, transparent, rgb(251 191 36 / 0.1), transparent);
  background-size: 5.5rem 100%, 100% 1rem, 100% 100%;
  content: "";
  opacity: 0.58;
}

.leaderboard-ranking-switch {
  display: inline-flex;
  width: fit-content;
  max-width: 100%;
  overflow-x: auto;
  border: 1px solid rgb(148 163 184 / 0.22);
  border-radius: 0.5rem;
  background: rgb(248 250 252 / 0.82);
  padding: 0.125rem;
  scrollbar-width: none;
}

.leaderboard-ranking-switch::-webkit-scrollbar {
  display: none;
}

.leaderboard-ranking-switch-button {
  flex: 0 0 auto;
  border-radius: 0.375rem;
  padding: 0.34rem 0.72rem;
  font-size: 0.8125rem;
  font-weight: 800;
  letter-spacing: 0;
  transition: background-color 0.16s ease, color 0.16s ease, box-shadow 0.16s ease;
  white-space: nowrap;
}

.leaderboard-ranking-switch-button--active {
  background: #ffffff;
  color: rgb(15 23 42);
  box-shadow: 0 0.45rem 1.1rem rgb(15 23 42 / 0.08);
}

.leaderboard-ranking-switch-button--idle {
  color: rgb(100 116 139);
}

.leaderboard-ranking-switch-button--idle:hover {
  color: rgb(15 23 42);
}

.leaderboard-ranking-card-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.leaderboard-ranking-empty {
  display: flex;
  min-height: 12rem;
  align-items: center;
  justify-content: center;
  border-top: 1px solid rgb(148 163 184 / 0.18);
  padding: 2rem 1rem 0.75rem;
}

.leaderboard-token-odometer {
  position: relative;
  z-index: 1;
  display: flex;
  min-height: 4.15rem;
  max-width: 100%;
  align-items: center;
  justify-content: center;
  overflow-x: auto;
  overflow-y: hidden;
  padding-top: 0.42rem;
  color: rgb(92 39 8);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: clamp(2.05rem, 6.6vw, 3.45rem);
  font-variant-numeric: tabular-nums;
  font-weight: 900;
  letter-spacing: 0;
  line-height: 1;
  scrollbar-width: none;
}

.leaderboard-token-odometer::-webkit-scrollbar {
  display: none;
}

.leaderboard-token-part {
  flex: 0 0 auto;
}

.leaderboard-token-reel {
  position: relative;
  width: 0.7em;
  height: 1.08em;
  margin-right: 0.04em;
  overflow: hidden;
  border: 1px solid rgb(180 83 9 / 0.2);
  border-radius: 0.28rem;
  background:
    linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(254 243 199 / 0.74)),
    radial-gradient(circle at 50% 0%, rgb(253 224 71 / 0.34), transparent 60%);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.8),
    inset 0 -0.18rem 0 rgb(120 53 15 / 0.06),
    0 0.34rem 0.95rem rgb(180 83 9 / 0.12);
}

.leaderboard-token-ranking-card {
  overflow: hidden;
  border: 1px solid rgb(148 163 184 / 0.16);
  border-radius: 0.75rem;
  background:
    linear-gradient(180deg, rgb(255 255 255 / 0.58), rgb(255 255 255 / 0.2)),
    rgb(255 255 255 / 0.34);
  padding: 1.25rem;
  box-shadow: 0 1rem 2.6rem rgb(15 23 42 / 0.05);
}

.leaderboard-token-ranking-card {
  overflow: hidden;
}

.leaderboard-token-ranking-updated {
  flex: 0 0 auto;
  color: rgb(100 116 139);
  font-size: 0.8125rem;
  white-space: nowrap;
}

.leaderboard-token-rank-list {
  display: grid;
  gap: 0.58rem;
  padding-top: 1rem;
}

.leaderboard-token-rank-row {
  display: grid;
  grid-template-columns: minmax(13.25rem, 16rem) minmax(14rem, 1fr);
  align-items: center;
  gap: 1rem;
  min-height: 3.7rem;
  padding: 0.16rem 0;
}

.leaderboard-token-rank-row-current {
  border-radius: 0.45rem;
  background: rgb(34 197 94 / 0.07);
}

.leaderboard-token-rank-user {
  display: grid;
  min-width: 0;
  grid-template-columns: 2.5rem 2.9rem minmax(0, 1fr);
  grid-template-rows: minmax(1.2rem, auto) minmax(1.05rem, auto);
  align-items: center;
  gap: 0.65rem;
  row-gap: 0.22rem;
}

.leaderboard-token-rank-main {
  display: contents;
}

.leaderboard-token-rank-name {
  grid-column: 3;
  grid-row: 1;
  max-width: 100%;
  min-width: 0;
  overflow: hidden;
  color: var(--token-rank-color);
  font-size: 0.875rem;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.leaderboard-token-rank-user--name-only .leaderboard-token-rank-name {
  grid-row: 1 / span 2;
  align-self: center;
}

.leaderboard-token-bar-area {
  display: block;
  min-width: 0;
}

.leaderboard-token-bar-track {
  position: relative;
  height: 1.32rem;
  overflow: visible;
  border-radius: 0;
  background:
    linear-gradient(90deg, rgb(148 163 184 / 0.13) 1px, transparent 1px),
    transparent;
  background-size: 10% 100%;
}

.leaderboard-token-bar-fill {
  display: flex;
  width: var(--token-bar-width);
  height: 100%;
  min-width: 0.7rem;
  overflow: hidden;
  border-radius: 0.08rem 0.28rem 0.28rem 0.08rem;
  background: rgb(148 163 184 / 0.18);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.26),
    0 0.35rem 1.1rem var(--token-bar-glow);
  transition: width 520ms cubic-bezier(0.22, 0.72, 0.2, 1);
}

.leaderboard-token-bar-segment {
  display: block;
  height: 100%;
  min-width: 0;
}

.leaderboard-token-bar-segment-input {
  width: var(--token-input-width);
  background: linear-gradient(90deg, rgb(20 20 19), rgb(37 37 35));
}

.leaderboard-token-bar-segment-output {
  width: var(--token-output-width);
  background: linear-gradient(90deg, rgb(169 88 62), rgb(204 120 92));
}

.leaderboard-token-bar-segment-cache {
  width: var(--token-cache-width);
  background: linear-gradient(90deg, rgb(76 148 132), rgb(93 184 166));
}

.leaderboard-token-rank-index {
  grid-column: 1;
  grid-row: 1 / span 2;
  min-width: 0;
  color: rgb(100 116 139);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
  font-weight: 800;
  text-align: right;
}

.leaderboard-token-rank-avatar {
  grid-column: 2;
  grid-row: 1 / span 2;
  display: inline-flex;
  flex: 0 0 auto;
  width: 2.9rem;
  height: 2.9rem;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 2px solid rgb(255 255 255 / 0.9);
  border-radius: 9999px;
  background:
    linear-gradient(135deg, rgb(255 255 255 / 0.96), rgb(226 232 240 / 0.84)),
    var(--token-bar-color);
  box-shadow:
    0 0 0 1px rgb(15 23 42 / 0.08),
    0 0.32rem 0.72rem var(--token-bar-glow);
  color: var(--token-rank-color);
  font-size: 1rem;
  font-weight: 900;
  line-height: 1;
  text-transform: uppercase;
}

.leaderboard-token-rank-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.leaderboard-token-current-tag {
  flex: 0 0 auto;
  justify-self: start;
  grid-column: 3;
  grid-row: 2;
  border-radius: 9999px;
  background: rgb(34 197 94 / 0.12);
  padding: 0.1rem 0.45rem;
  color: rgb(22 163 74);
  font-size: 0.6875rem;
  font-weight: 700;
}

.leaderboard-token-title-list {
  display: flex;
  grid-column: 3;
  grid-row: 2;
  max-width: 100%;
  min-width: 0;
  flex-wrap: wrap;
  justify-content: flex-start;
  gap: 0.25rem;
}

.leaderboard-token-title-badge,
.leaderboard-token-title-more {
  display: inline-flex;
  max-width: 5.5rem;
  align-items: center;
  overflow: hidden;
  border: 1px solid currentColor;
  border-radius: 9999px;
  background: rgb(255 255 255 / 0.55);
  padding: 0.08rem 0.4rem;
  color: var(--token-rank-color);
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1.15;
  opacity: 0.84;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.leaderboard-token-title-more {
  max-width: none;
  color: rgb(100 116 139);
}

.leaderboard-token-bar-value {
  position: absolute;
  top: 50%;
  left: var(--token-bar-value-left);
  min-width: 0;
  transform: translate(var(--token-bar-value-x), -50%);
  color: var(--token-value-color);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.95rem;
  font-weight: 800;
  line-height: 1;
  white-space: nowrap;
}

.leaderboard-my-token-card {
  border: 1px solid rgb(148 163 184 / 0.16);
}

.leaderboard-reward-head {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: start;
  gap: 0.85rem;
}

.leaderboard-reward-title {
  color: rgb(17 24 39);
  font-size: 1rem;
  font-weight: 800;
  letter-spacing: 0;
  line-height: 1.3;
}

.leaderboard-reward-period {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  gap: 0.16rem 0.35rem;
  margin-top: 0.32rem;
  color: rgb(100 116 139);
  font-size: 0.75rem;
  line-height: 1.35;
}

.leaderboard-reward-status {
  display: inline-flex;
  max-width: 8.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  padding: 0.38rem 0.72rem;
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1.12;
  text-align: center;
  white-space: nowrap;
}

.leaderboard-weekly-winners {
  display: grid;
  gap: 0.6rem;
  border: 1px solid rgb(148 163 184 / 0.16);
  border-radius: 0.5rem;
  background: rgb(248 250 252);
  padding: 0.75rem;
}

.leaderboard-weekly-winners-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: rgb(100 116 139);
  font-size: 0.75rem;
  font-weight: 700;
}

.leaderboard-weekly-winners-list {
  display: grid;
  gap: 0.45rem;
}

.leaderboard-weekly-winner-row {
  display: grid;
  min-width: 0;
  grid-template-columns: minmax(5.6rem, auto) minmax(0, 1fr);
  align-items: center;
  gap: 0.65rem;
}

.leaderboard-weekly-winner-rank {
  color: rgb(100 116 139);
  font-size: 0.8125rem;
}

.leaderboard-weekly-winner-name {
  min-width: 0;
  overflow: hidden;
  color: rgb(15 23 42);
  font-size: 0.875rem;
  font-weight: 800;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.leaderboard-token-reel::before,
.leaderboard-token-reel::after {
  position: absolute;
  left: 0;
  z-index: 2;
  width: 100%;
  height: 30%;
  pointer-events: none;
  content: "";
}

.leaderboard-token-reel::before {
  top: 0;
  background: linear-gradient(180deg, rgb(255 247 237 / 0.96), transparent);
}

.leaderboard-token-reel::after {
  bottom: 0;
  background: linear-gradient(0deg, rgb(120 53 15 / 0.16), transparent);
}

.leaderboard-token-strip {
  display: flex;
  flex-direction: column;
  transform: translateY(var(--target-offset));
  transition: transform 620ms cubic-bezier(0.17, 0.84, 0.29, 1);
  will-change: transform;
}

.leaderboard-token-cell {
  display: flex;
  height: 1.08em;
  align-items: center;
  justify-content: center;
  text-shadow: 0 1px 0 rgb(255 255 255 / 0.8);
}

.leaderboard-token-separator {
  display: inline-flex;
  height: 1.08em;
  align-items: flex-end;
  padding: 0 0.026em 0.08em;
  color: rgb(180 83 9);
  text-shadow: 0 1px 0 rgb(255 255 255 / 0.78);
}

.leaderboard-calendar-card {
  overflow: hidden;
  border: 1px solid rgb(148 163 184 / 0.22);
  border-radius: 0.75rem;
  background:
    linear-gradient(180deg, rgb(255 255 255 / 0.86), rgb(248 250 252 / 0.7)),
    linear-gradient(90deg, rgb(148 163 184 / 0.1) 1px, transparent 1px),
    rgb(248 250 252 / 0.72);
  background-size: auto, 3rem 100%, auto;
  padding: 1rem;
  box-shadow: 0 1rem 2.4rem rgb(15 23 42 / 0.05);
}

.leaderboard-calendar-head {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 0.8rem;
  padding-bottom: 0.8rem;
}

.leaderboard-calendar-title {
  flex: 1 1 auto;
  color: rgb(15 23 42);
  font-size: 1.12rem;
  font-weight: 900;
  letter-spacing: 0;
  line-height: 1.25;
  text-align: left;
}

.leaderboard-calendar-scroll {
  overflow-x: auto;
  overscroll-behavior-x: contain;
  padding-bottom: 0.25rem;
}

.leaderboard-calendar-scroll::-webkit-scrollbar {
  height: 0.45rem;
}

.leaderboard-calendar-scroll::-webkit-scrollbar-thumb {
  border-radius: 9999px;
  background: rgb(148 163 184 / 0.32);
}

.leaderboard-calendar-months {
  display: grid;
  min-width: 50.75rem;
  grid-template-columns: repeat(2, minmax(24.4rem, 1fr));
  gap: 1rem;
}

.leaderboard-calendar-month-panel {
  overflow: hidden;
  border: 1px solid rgb(148 163 184 / 0.22);
  border-radius: 0.65rem;
  background:
    linear-gradient(180deg, rgb(255 255 255 / 0.9), rgb(241 245 249 / 0.64)),
    rgb(255 255 255 / 0.68);
  padding: 0.85rem;
}

.leaderboard-calendar-month-panel--current {
  border-color: rgb(59 130 246 / 0.28);
  background:
    linear-gradient(180deg, rgb(239 246 255 / 0.9), rgb(255 255 255 / 0.72)),
    rgb(255 255 255 / 0.72);
}

.leaderboard-calendar-month {
  margin-bottom: 0.6rem;
  color: rgb(15 23 42);
  font-size: 0.98rem;
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
}

.leaderboard-calendar-weekdays {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 0.4rem;
  margin-bottom: 0.42rem;
}

.leaderboard-calendar-weekday {
  color: rgb(100 116 139);
  font-size: 0.72rem;
  font-weight: 900;
  line-height: 1;
  text-align: center;
}

.leaderboard-calendar-days {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  grid-auto-rows: 3.45rem;
  gap: 0.4rem;
}

.leaderboard-calendar-day {
  position: relative;
  display: grid;
  width: 100%;
  min-width: 0;
  height: 3.45rem;
  justify-items: center;
  overflow: hidden;
  border: 1px solid rgb(148 163 184 / 0.22);
  border-radius: 0.5rem;
  background:
    linear-gradient(180deg, rgb(255 255 255 / 0.8), rgb(241 245 249 / 0.48)),
    rgb(255 255 255 / 0.56);
}

.leaderboard-calendar-day--active {
  border-color: rgb(245 158 11 / 0.42);
  background:
    radial-gradient(circle at 50% 18%, rgb(251 191 36 / 0.28), transparent 58%),
    linear-gradient(180deg, rgb(255 251 235 / 0.88), rgb(255 255 255 / 0.62));
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.7),
    0 0.45rem 1rem rgb(245 158 11 / 0.1);
}

.leaderboard-calendar-day--placeholder {
  border-color: transparent;
  background: transparent;
  box-shadow: none;
}

.leaderboard-calendar-day-number {
  position: absolute;
  top: 0.22rem;
  left: 0.28rem;
  z-index: 2;
  display: inline-flex;
  min-width: 1rem;
  height: 0.95rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgb(255 251 235 / 0.86);
  color: rgb(146 64 14);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.68rem;
  font-weight: 900;
  line-height: 1;
  box-shadow: 0 0 0 1px rgb(245 158 11 / 0.16);
}

.leaderboard-calendar-avatar {
  position: absolute;
  inset: 0;
  display: inline-flex;
  width: 100%;
  height: 100%;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: inherit;
  background:
    radial-gradient(circle at 50% 36%, rgb(255 255 255 / 0.26), transparent 42%),
    linear-gradient(135deg, rgb(253 230 138 / 0.98), rgb(217 119 6 / 0.82)),
    rgb(217 119 6);
  color: rgb(255 251 235);
  font-size: 1.18rem;
  font-weight: 900;
  line-height: 1;
  text-shadow: 0 1px 0 rgb(120 53 15 / 0.35);
  text-transform: uppercase;
}

.leaderboard-calendar-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.leaderboard-calendar-day--active .leaderboard-calendar-avatar::after {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgb(15 23 42 / 0.18), transparent 42%, rgb(15 23 42 / 0.08)),
    inset 0 0 0 1px rgb(255 255 255 / 0.34);
  pointer-events: none;
  content: "";
}

.leaderboard-calendar-tooltip {
  position: absolute;
  z-index: 1200;
  display: grid;
  width: 15rem;
  min-height: 6.1rem;
  gap: 0.34rem;
  border: 1px solid rgb(120 53 15 / 0.34);
  border-radius: 0.5rem;
  background:
    linear-gradient(180deg, rgb(28 25 18 / 0.98), rgb(18 16 11 / 0.98)),
    rgb(18 16 11);
  box-shadow:
    0 1rem 2rem rgb(15 23 42 / 0.22),
    0 0 0 1px rgb(253 224 71 / 0.08),
    inset 0 1px 0 rgb(255 255 255 / 0.08);
  color: rgb(245 245 244);
  padding: 0.72rem 0.78rem 0.7rem;
  pointer-events: none;
  letter-spacing: 0;
}

.leaderboard-calendar-tooltip-date {
  color: rgb(161 161 170);
  font-size: 0.68rem;
  font-weight: 900;
  line-height: 1;
}

.leaderboard-calendar-tooltip-main {
  display: grid;
  grid-template-columns: minmax(0, 1fr) max-content;
  align-items: baseline;
  gap: 0.75rem;
  min-width: 0;
}

.leaderboard-calendar-tooltip-name {
  min-width: 0;
  overflow-wrap: anywhere;
  color: rgb(250 250 249);
  font-family: ui-serif, Georgia, Cambria, "Times New Roman", Times, serif;
  font-size: 0.92rem;
  font-weight: 900;
  line-height: 1.18;
}

.leaderboard-calendar-tooltip-tokens {
  color: rgb(234 179 8);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
  font-weight: 900;
  line-height: 1;
  white-space: nowrap;
}

.leaderboard-calendar-tooltip-meta {
  color: rgb(168 162 158);
  font-size: 0.74rem;
  font-weight: 800;
  line-height: 1.2;
}

:global(.dark .leaderboard-token-card) {
  border-color: rgb(71 85 105 / 0.34);
  background:
    radial-gradient(circle at 50% 8%, rgb(251 191 36 / 0.16), transparent 34%),
    linear-gradient(180deg, rgb(30 41 59 / 0.92), rgb(15 23 42 / 0.94));
  box-shadow: none;
}

:global(.dark .leaderboard-token-card::after) {
  background:
    linear-gradient(90deg, rgb(148 163 184 / 0.12) 1px, transparent 1px),
    linear-gradient(rgb(148 163 184 / 0.1) 1px, transparent 1px),
    linear-gradient(90deg, transparent, rgb(251 191 36 / 0.08), transparent);
  background-size: 5.5rem 100%, 100% 1rem, 100% 100%;
}

:global(.dark .leaderboard-ranking-switch) {
  border-color: rgb(51 65 85 / 0.86);
  background: rgb(15 23 42 / 0.76);
}

:global(.dark .leaderboard-ranking-switch-button--active) {
  background: rgb(30 41 59 / 0.96);
  color: rgb(248 250 252);
  box-shadow: none;
}

:global(.dark .leaderboard-ranking-switch-button--idle) {
  color: rgb(148 163 184);
}

:global(.dark .leaderboard-ranking-switch-button--idle:hover) {
  color: rgb(248 250 252);
}

:global(.dark .leaderboard-token-summary-meta) {
  color: rgb(148 163 184);
}

:global(.dark .leaderboard-token-trend-panel) {
  border-left-color: rgb(71 85 105 / 0.42);
}

:global(.dark .leaderboard-token-trend-header) {
  color: rgb(203 213 225);
}

:global(.dark .leaderboard-token-trend-legend) {
  color: rgb(96 165 250);
}

:global(.dark .leaderboard-token-trend-legend::before) {
  background: rgb(96 165 250 / 0.16);
}

:global(.dark .leaderboard-token-trend-empty) {
  border-color: rgb(71 85 105 / 0.34);
  background:
    linear-gradient(90deg, rgb(148 163 184 / 0.1) 1px, transparent 1px),
    linear-gradient(rgb(148 163 184 / 0.1) 1px, transparent 1px);
  background-size: 12.5% 100%, 100% 1.65rem;
  color: rgb(148 163 184);
}

:global(.dark .leaderboard-token-odometer) {
  color: rgb(254 240 138);
}

:global(.dark .leaderboard-token-reel) {
  border-color: rgb(251 191 36 / 0.3);
  background:
    linear-gradient(180deg, rgb(30 41 59 / 0.96), rgb(15 23 42 / 0.92)),
    radial-gradient(circle at 50% 0%, rgb(251 191 36 / 0.3), transparent 58%);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.1),
    inset 0 -0.22rem 0 rgb(0 0 0 / 0.24),
    0 0.35rem 1rem rgb(251 191 36 / 0.12);
}

:global(.dark .leaderboard-token-reel::before) {
  background: linear-gradient(180deg, rgb(30 41 59 / 0.96), transparent);
}

:global(.dark .leaderboard-token-reel::after) {
  background: linear-gradient(0deg, rgb(0 0 0 / 0.28), transparent);
}

:global(.dark .leaderboard-token-cell) {
  text-shadow: 0 0 0.55rem rgb(251 191 36 / 0.24);
}

:global(.dark .leaderboard-token-separator) {
  color: rgb(252 211 77);
  text-shadow: 0 0 0.55rem rgb(251 191 36 / 0.2);
}

:global(.dark .leaderboard-token-ranking-card) {
  border-color: rgb(71 85 105 / 0.34);
  background:
    linear-gradient(180deg, rgb(15 23 42 / 0.45), rgb(15 23 42 / 0.16)),
    rgb(15 23 42 / 0.2);
  box-shadow: none;
}

:global(.dark .leaderboard-token-rank-row-current) {
  background: rgb(34 197 94 / 0.1);
}

:global(.dark .leaderboard-token-title-badge),
:global(.dark .leaderboard-token-title-more) {
  background: rgb(15 23 42 / 0.28);
}

:global(.dark .leaderboard-token-title-more) {
  color: rgb(148 163 184);
}

:global(.dark .leaderboard-reward-title) {
  color: rgb(248 250 252);
}

:global(.dark .leaderboard-reward-period) {
  color: rgb(148 163 184);
}

:global(.dark .leaderboard-weekly-winners) {
  border-color: rgb(71 85 105 / 0.34);
  background: rgb(15 23 42 / 0.32);
}

:global(.dark .leaderboard-weekly-winners-header),
:global(.dark .leaderboard-weekly-winner-rank) {
  color: rgb(148 163 184);
}

:global(.dark .leaderboard-weekly-winner-name) {
  color: rgb(248 250 252);
}

:global(.dark .leaderboard-token-rank-avatar) {
  border-color: rgb(15 23 42 / 0.92);
  background:
    linear-gradient(135deg, rgb(30 41 59 / 0.96), rgb(15 23 42 / 0.9)),
    var(--token-bar-color);
  box-shadow:
    0 0 0 1px rgb(148 163 184 / 0.18),
    0 0.32rem 0.8rem var(--token-bar-glow);
}

:global(.dark .leaderboard-token-bar-track) {
  background:
    linear-gradient(90deg, rgb(148 163 184 / 0.1) 1px, transparent 1px),
    transparent;
  background-size: 10% 100%;
}

:global(.dark .leaderboard-token-bar-fill) {
  background: rgb(51 65 85 / 0.42);
}

:global(.dark .leaderboard-token-bar-segment-input) {
  background: linear-gradient(90deg, rgb(160 157 150), rgb(250 249 245));
}

:global(.dark .leaderboard-token-bar-segment-output) {
  background: linear-gradient(90deg, rgb(169 88 62), rgb(204 120 92));
}

:global(.dark .leaderboard-token-bar-segment-cache) {
  background: linear-gradient(90deg, rgb(76 148 132), rgb(93 184 166));
}

:global(.dark .leaderboard-calendar-card) {
  border-color: rgb(71 85 105 / 0.34);
  background:
    linear-gradient(180deg, rgb(15 23 42 / 0.58), rgb(15 23 42 / 0.28)),
    linear-gradient(90deg, rgb(148 163 184 / 0.08) 1px, transparent 1px),
    rgb(15 23 42 / 0.3);
  background-size: auto, 3rem 100%, auto;
  box-shadow: none;
}

:global(.dark .leaderboard-calendar-title),
:global(.dark .leaderboard-calendar-month) {
  color: rgb(248 250 252);
}

:global(.dark .leaderboard-calendar-scroll::-webkit-scrollbar-thumb) {
  background: rgb(148 163 184 / 0.34);
}

:global(.dark .leaderboard-calendar-month-panel) {
  border-color: rgb(71 85 105 / 0.34);
  background:
    linear-gradient(180deg, rgb(30 41 59 / 0.68), rgb(15 23 42 / 0.5)),
    rgb(15 23 42 / 0.42);
}

:global(.dark .leaderboard-calendar-month-panel--current) {
  border-color: rgb(96 165 250 / 0.36);
  background:
    linear-gradient(180deg, rgb(30 41 59 / 0.78), rgb(15 23 42 / 0.56)),
    rgb(15 23 42 / 0.48);
}

:global(.dark .leaderboard-calendar-weekday) {
  color: rgb(148 163 184);
}

:global(.dark .leaderboard-calendar-day) {
  border-color: rgb(71 85 105 / 0.4);
  background:
    linear-gradient(180deg, rgb(30 41 59 / 0.66), rgb(15 23 42 / 0.36)),
    rgb(15 23 42 / 0.32);
}

:global(.dark .leaderboard-calendar-day--active) {
  border-color: rgb(252 211 77 / 0.42);
  background:
    radial-gradient(circle at 50% 18%, rgb(251 191 36 / 0.22), transparent 58%),
    linear-gradient(180deg, rgb(30 41 59 / 0.84), rgb(15 23 42 / 0.6));
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.08),
    0 0.45rem 1rem rgb(251 191 36 / 0.1);
}

:global(.dark .leaderboard-calendar-day--placeholder) {
  border-color: transparent;
  background: transparent;
  box-shadow: none;
}

:global(.dark .leaderboard-calendar-day-number) {
  background: rgb(15 23 42 / 0.72);
  color: rgb(253 224 71);
  box-shadow: 0 0 0 1px rgb(252 211 77 / 0.12);
}

:global(.dark .leaderboard-calendar-avatar) {
  background:
    radial-gradient(circle at 50% 36%, rgb(255 255 255 / 0.16), transparent 42%),
    linear-gradient(135deg, rgb(30 41 59 / 0.96), rgb(217 119 6 / 0.7)),
    rgb(217 119 6);
  color: rgb(253 224 71);
}

:global(.dark .leaderboard-calendar-tooltip) {
  border-color: rgb(252 211 77 / 0.2);
  background:
    linear-gradient(180deg, rgb(17 17 17 / 0.98), rgb(8 8 8 / 0.98)),
    rgb(8 8 8);
  box-shadow:
    0 1rem 2rem rgb(0 0 0 / 0.34),
    0 0 0 1px rgb(252 211 77 / 0.08),
    inset 0 1px 0 rgb(255 255 255 / 0.08);
}

@media (prefers-reduced-motion: reduce) {
  .leaderboard-token-strip {
    transition: none;
    will-change: auto;
  }

  .leaderboard-token-bar-fill {
    transition: none;
  }
}

@media (max-width: 767px) {
  .leaderboard-ranking-card-toolbar {
    flex-wrap: wrap;
    margin-bottom: 0.7rem;
  }

  .leaderboard-token-rank-row {
    grid-template-columns: 1fr;
    gap: 0.45rem;
    min-height: 3.35rem;
    padding: 0.26rem 0;
  }

  .leaderboard-token-rank-user {
    grid-template-columns: 2.35rem 2.5rem minmax(0, 1fr);
    grid-template-rows: minmax(1.1rem, auto) minmax(1rem, auto);
    row-gap: 0.18rem;
  }

  .leaderboard-token-bar-track {
    height: 1.35rem;
  }

  .leaderboard-token-bar-value {
    font-size: 0.8rem;
  }

  .leaderboard-token-rank-avatar {
    width: 2.5rem;
    height: 2.5rem;
    border-width: 1px;
    font-size: 0.9rem;
  }

  .leaderboard-token-odometer {
    justify-content: center;
    font-size: clamp(1.72rem, 10.6vw, 2.45rem);
  }

  .leaderboard-token-summary-inner {
    grid-template-columns: 1fr;
    gap: 1rem;
    text-align: center;
  }

  .leaderboard-token-summary-main {
    justify-items: center;
  }

  .leaderboard-token-summary-meta {
    justify-items: center;
  }

  .leaderboard-token-trend-panel {
    width: 100%;
    border-left: 0;
    border-top: 1px solid rgb(148 163 184 / 0.22);
    padding-top: 1rem;
    padding-left: 0;
  }

  .leaderboard-token-trend-chart,
  .leaderboard-token-trend-empty {
    height: 6.1rem;
  }

  .leaderboard-calendar-card {
    padding: 0.95rem;
  }

  .leaderboard-calendar-head {
    padding-bottom: 0.65rem;
  }

  .leaderboard-calendar-months {
    min-width: 44.7rem;
    grid-template-columns: repeat(2, 21.85rem);
    gap: 0.75rem;
  }

  .leaderboard-calendar-month-panel {
    padding: 0.72rem;
  }

  .leaderboard-calendar-weekdays,
  .leaderboard-calendar-days {
    gap: 0.48rem;
  }

  .leaderboard-calendar-days {
    grid-auto-rows: 3.05rem;
  }

  .leaderboard-calendar-day {
    height: 3.05rem;
  }

  .leaderboard-calendar-avatar {
    font-size: 1rem;
  }

  .leaderboard-reward-head {
    grid-template-columns: 1fr;
    gap: 0.65rem;
  }

  .leaderboard-reward-status {
    justify-self: start;
    max-width: 100%;
  }
}

.leaderboard-badge-icon {
  display: inline-flex;
  width: 1.25rem;
  height: 1.25rem;
  align-items: center;
  justify-content: center;
  border: 1px solid currentColor;
  border-radius: 0.25rem;
  font-size: 0.6875rem;
  font-weight: 800;
  line-height: 1;
  box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.45);
}

.leaderboard-badge-overflow {
  display: inline-flex;
  height: 1.25rem;
  min-width: 1.25rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.25rem;
  background: rgb(248 250 252);
  padding: 0 0.25rem;
  color: rgb(71 85 105);
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1;
  box-shadow: inset 0 0 0 1px rgb(203 213 225);
}

.leaderboard-badge-week {
  background: rgb(239 246 255);
  color: rgb(37 99 235);
}

.leaderboard-badge-month {
  background: rgb(245 243 255);
  color: rgb(124 58 237);
}

.leaderboard-badge-total {
  background: rgb(254 249 195);
  color: rgb(161 98 7);
}

.leaderboard-badge-night {
  background: rgb(238 242 255);
  color: rgb(67 56 202);
}

.leaderboard-badge-burst {
  background: rgb(255 241 242);
  color: rgb(225 29 72);
}

.leaderboard-badge-checkin {
  background: rgb(240 253 250);
  color: rgb(13 148 136);
}

.leaderboard-badge-save {
  background: rgb(240 253 244);
  color: rgb(22 163 74);
}

.leaderboard-badge-fire {
  background: rgb(255 247 237);
  color: rgb(234 88 12);
}

:global(.dark .leaderboard-badge-overflow) {
  background: rgb(30 41 59);
  color: rgb(203 213 225);
  box-shadow: inset 0 0 0 1px rgb(71 85 105);
}

:global(.dark .leaderboard-badge-week) {
  background: rgb(37 99 235 / 0.16);
  color: rgb(147 197 253);
}

:global(.dark .leaderboard-badge-month) {
  background: rgb(124 58 237 / 0.16);
  color: rgb(196 181 253);
}

:global(.dark .leaderboard-badge-total) {
  background: rgb(202 138 4 / 0.16);
  color: rgb(253 224 71);
}

:global(.dark .leaderboard-badge-night) {
  background: rgb(79 70 229 / 0.16);
  color: rgb(165 180 252);
}

:global(.dark .leaderboard-badge-burst) {
  background: rgb(225 29 72 / 0.16);
  color: rgb(253 164 175);
}

:global(.dark .leaderboard-badge-checkin) {
  background: rgb(13 148 136 / 0.16);
  color: rgb(94 234 212);
}

:global(.dark .leaderboard-badge-save) {
  background: rgb(22 163 74 / 0.16);
  color: rgb(134 239 172);
}

:global(.dark .leaderboard-badge-fire) {
  background: rgb(234 88 12 / 0.16);
  color: rgb(253 186 116);
}
</style>
