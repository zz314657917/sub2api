<template>
  <AppLayout>
    <div class="leaderboard-page space-y-6">
      <section v-if="leaderboard" class="card leaderboard-token-card leaderboard-token-summary p-5">
        <div class="leaderboard-token-summary-inner relative z-10">
          <div class="leaderboard-token-summary-main min-w-0">
            <p class="leaderboard-section-label text-sm">{{ t('leaderboard.totalTokens') }}</p>
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
            <div class="leaderboard-token-summary-meta text-sm">
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

      <div v-if="loading" class="card leaderboard-state-card flex min-h-[280px] items-center justify-center p-8">
        <div class="text-center">
          <div class="leaderboard-loading-spinner mx-auto h-8 w-8 animate-spin rounded-full border-2 border-t-transparent"></div>
          <p class="leaderboard-section-label mt-3 text-sm">{{ t('common.loading') }}</p>
        </div>
      </div>

      <div v-else-if="error" class="card leaderboard-state-card p-8 text-center">
        <h2 class="leaderboard-state-title text-lg font-semibold">{{ t('leaderboard.errorTitle') }}</h2>
        <p class="leaderboard-section-label mt-2 text-sm">{{ t('leaderboard.errorDescription') }}</p>
        <button class="btn btn-primary leaderboard-reward-claim mt-5" type="button" @click="loadLeaderboard">{{ t('leaderboard.retry') }}</button>
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
                <span
                  v-if="refreshingLeaderboard"
                  class="leaderboard-token-ranking-refreshing"
                  data-testid="leaderboard-refreshing"
                >
                  {{ t('leaderboard.refreshing') }}
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
                  <div class="leaderboard-ranking-toolbar-meta">
                    <div class="leaderboard-token-legend" aria-hidden="true">
                      <span class="leaderboard-token-legend-item">
                        <span class="leaderboard-token-legend-dot leaderboard-token-legend-dot--input"></span>
                        {{ t('leaderboard.inputTokensShort') }}
                      </span>
                      <span class="leaderboard-token-legend-item">
                        <span class="leaderboard-token-legend-dot leaderboard-token-legend-dot--output"></span>
                        {{ t('leaderboard.outputTokensShort') }}
                      </span>
                      <span class="leaderboard-token-legend-item">
                        <span class="leaderboard-token-legend-dot leaderboard-token-legend-dot--cache"></span>
                        {{ t('leaderboard.cacheTokensShort') }}
                      </span>
                    </div>
                    <span class="leaderboard-token-ranking-updated">
                      {{ t('leaderboard.generatedAt') }} {{ formatTime(leaderboard.generated_at) }}
                    </span>
                    <span
                      v-if="refreshingLeaderboard"
                      class="leaderboard-token-ranking-refreshing"
                      data-testid="leaderboard-refreshing"
                    >
                      {{ t('leaderboard.refreshing') }}
                    </span>
                  </div>
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
                        <span class="leaderboard-token-rank-index">{{ item.rank }}</span>
                        <span class="leaderboard-token-rank-avatar" data-testid="leaderboard-rank-avatar" aria-hidden="true">
                          <img
                            v-if="leaderboardAvatarUrl(item)"
                            :src="leaderboardAvatarUrl(item)"
                            alt=""
                            loading="lazy"
                          >
                          <span v-else>{{ leaderboardAvatarInitial(item) }}</span>
                        </span>
                        <span class="leaderboard-token-rank-name-line">
                          <span class="leaderboard-token-rank-name" :title="getLeaderboardDisplayName(item)">
                            {{ getLeaderboardDisplayName(item) }}
                          </span>
                          <span
                            v-if="leaderboardRankChangeLabel(item)"
                            class="leaderboard-rank-change"
                            :class="leaderboardRankChangeClass(item)"
                            :title="leaderboardRankChangeTitle(item)"
                            data-testid="leaderboard-rank-change"
                          >
                            {{ leaderboardRankChangeLabel(item) }}
                          </span>
                          <span v-if="item.is_current_user" class="leaderboard-token-current-tag">
                            {{ t('leaderboard.currentUser') }}
                          </span>
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
                        role="img"
                        tabindex="0"
                        @mouseenter="showTokenBarTooltip(item, $event)"
                        @mouseleave="hideTokenBarTooltip"
                        @focusin="showTokenBarTooltip(item, $event)"
                        @focusout="hideTokenBarTooltip"
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

          <aside class="leaderboard-side-stack xl:sticky xl:top-20 xl:self-start">
            <section class="card leaderboard-side-card p-5" data-testid="leaderboard-my-info">
              <div class="leaderboard-thursday-banner" data-testid="leaderboard-thursday-banner">
                <img :src="crazyThursdayBannerUrl" alt="" loading="lazy">
                <div class="leaderboard-thursday-banner-copy" aria-label="疯狂星期四 V你50">
                  <span>疯狂星期四</span>
                  <strong>V你50</strong>
                </div>
              </div>

              <div class="leaderboard-record-card" data-testid="leaderboard-my-record">
                <p class="leaderboard-record-kicker">{{ t('leaderboard.record.title') }}</p>
                <p class="leaderboard-record-headline">{{ myRecordHeadline }}</p>
                <p
                  class="leaderboard-record-progress"
                  :class="{ 'leaderboard-record-progress--deity': myRecordProgress.isDeity }"
                >
                  <template v-if="myRecordProgress.isDeity">
                    <strong>{{ myRecordProgress.value }}</strong>
                  </template>
                  <template v-else>
                    <span>{{ myRecordProgress.prefix }}</span>
                    <span>
                      <strong>{{ myRecordProgress.value }}</strong>
                      <span v-if="myRecordProgress.suffix" class="leaderboard-record-unit">{{ myRecordProgress.suffix }}</span>
                    </span>
                  </template>
                </p>
              </div>

              <div v-if="dailyRewards" class="leaderboard-reward-panel" data-testid="leaderboard-daily-reward">
                <div class="leaderboard-reward-head">
                  <div class="min-w-0">
                    <h2 class="leaderboard-reward-title">{{ t('leaderboard.dailyReward.title') }}</h2>
                  </div>
                  <span
                    class="leaderboard-reward-status"
                    :class="dailyRewardStatusClass"
                    :title="dailyRewardReasonText"
                    data-testid="leaderboard-daily-reward-status"
                  >
                    {{ dailyRewardStatusText }}
                  </span>
                </div>

                <div v-if="rewardTopUsers.length" class="leaderboard-weekly-winners mt-3" data-testid="leaderboard-weekly-top10">
                  <div class="leaderboard-weekly-winners-header">
                    <span>{{ t('leaderboard.dailyReward.lastWeekTopUsersTitle') }}</span>
                    <span>{{ rewardWeekRangeLabel }}</span>
                  </div>
                  <div class="leaderboard-weekly-winners-list" data-testid="leaderboard-weekly-top10-scroll">
                    <div
                      v-for="winner in rewardTopUsers"
                      :key="winner.rank"
                      class="leaderboard-weekly-winner-row"
                      :class="{ 'leaderboard-weekly-winner-row--highlighted': winner.highlighted }"
                    >
                      <span class="leaderboard-weekly-winner-rank">{{ rewardTopUserRankLabel(winner.rank) }}</span>
                      <span class="leaderboard-weekly-winner-name">{{ winner.displayName }}</span>
                      <span v-if="winner.metaText" class="leaderboard-weekly-winner-meta">{{ winner.metaText }}</span>
                      <span class="leaderboard-weekly-winner-tokens">{{ winner.tokenText }}</span>
                    </div>
                  </div>
                </div>

                <div class="leaderboard-reward-progress-card mt-3 rounded-lg p-3 text-sm">
                  <div class="flex items-center justify-between gap-3">
                    <span class="leaderboard-side-label">{{ t('leaderboard.dailyReward.lastWeekRank') }}</span>
                    <span class="leaderboard-side-value font-semibold">{{ formatRewardRankLabel(dailyRewards.current_user_rank) }}</span>
                  </div>
                  <div class="mt-2 flex items-center justify-between gap-3">
                    <span class="leaderboard-side-label">{{ t('leaderboard.dailyReward.weeklyRushProgress') }}</span>
                    <span class="leaderboard-side-value leaderboard-weekly-rush-value font-semibold">{{ weeklyRushProgressText }}</span>
                  </div>
                </div>

                <div
                  v-if="leaderboardRewardMode === 'red_packet'"
                  class="leaderboard-reward-mode-card mt-3 rounded-lg p-3 text-sm"
                  data-testid="leaderboard-red-packet-reward"
                >
                  <div class="flex items-center justify-between gap-3">
                    <span class="leaderboard-side-label">{{ t('leaderboard.dailyReward.redPacketRange') }}</span>
                    <span class="leaderboard-side-value font-semibold">{{ redPacketRangeText }}</span>
                  </div>
                  <div class="mt-2 flex items-center justify-between gap-3">
                    <span class="leaderboard-side-label">{{ t(redPacketAmountLabelKey) }}</span>
                    <span class="leaderboard-side-value font-semibold">{{ redPacketAmountText }}</span>
                  </div>
                </div>

                <div
                  v-else-if="leaderboardRewardMode === 'lottery'"
                  class="leaderboard-reward-mode-card mt-3 rounded-lg p-3 text-sm"
                  data-testid="leaderboard-lottery-reward"
                >
                  <div class="flex items-center justify-between gap-3">
                    <span class="leaderboard-side-label">{{ t('leaderboard.dailyReward.lotteryDrawTime') }}</span>
                    <span class="leaderboard-side-value font-semibold">{{ lotteryDrawTimeText }}</span>
                  </div>
                  <div class="mt-2 flex items-center justify-between gap-3">
                    <span class="leaderboard-side-label">{{ t('leaderboard.dailyReward.lotteryPrize') }}</span>
                    <span class="leaderboard-side-value font-semibold">{{ lotteryPrizeText }}</span>
                  </div>
                  <div class="mt-2 flex items-center justify-between gap-3">
                    <span class="leaderboard-side-label">{{ t('leaderboard.dailyReward.lotteryResult') }}</span>
                    <span class="leaderboard-side-value leaderboard-weekly-rush-value font-semibold">{{ lotteryResultText }}</span>
                  </div>
                </div>

                <p v-if="claimError" class="mt-3 text-sm text-red-600 dark:text-red-400">{{ claimError }}</p>

                <button
                  v-if="leaderboardRewardMode === 'red_packet'"
                  class="btn btn-primary leaderboard-reward-claim mt-3 w-full"
                  type="button"
                  :disabled="!dailyRewards.can_claim || claimingReward"
                  data-testid="leaderboard-daily-reward-claim"
                  @click="claimDailyReward"
                >
                  {{ claimButtonText }}
                </button>
              </div>
            </section>
          </aside>
        </div>

        <section class="leaderboard-calendar-card" data-testid="leaderboard-daily-champions-calendar">
          <div class="leaderboard-calendar-head">
            <h2 class="leaderboard-calendar-title">{{ t('leaderboard.calendar.title') }}</h2>
            <span class="leaderboard-calendar-meta">{{ championCalendarRangeLabel }}</span>
          </div>

          <div class="leaderboard-calendar-scroll">
            <div class="leaderboard-calendar-months">
              <section
                v-for="month in championCalendarMonths"
                :key="month.key"
                class="leaderboard-calendar-month-panel"
                :class="month.isCurrent ? 'leaderboard-calendar-month-panel--current' : 'leaderboard-calendar-month-panel--previous'"
              >
                <div class="leaderboard-calendar-month-toolbar">
                  <div class="leaderboard-calendar-month">{{ month.label }}</div>
                </div>
                <div class="leaderboard-calendar-days" :style="{ '--calendar-day-columns': String(month.columnCount) }">
                  <div
                    v-for="day in month.days"
                    :key="day.key"
                    class="leaderboard-calendar-day"
                    :class="{
                      'leaderboard-calendar-day--active': day.champion,
                      'leaderboard-calendar-day--empty': !day.champion
                    }"
                    :aria-label="championCalendarCellLabel(day)"
                    :tabindex="championCalendarCellTabIndex(day)"
                    data-testid="leaderboard-calendar-day"
                    @mouseenter="showChampionTooltip(day, $event)"
                    @mouseleave="hideChampionTooltip"
                    @focusin="showChampionTooltip(day, $event)"
                    @focusout="hideChampionTooltip"
                  >
                    <span class="leaderboard-calendar-day-number">{{ day.day }}</span>
                    <span class="leaderboard-calendar-avatar-frame" data-testid="leaderboard-calendar-avatar-frame">
                      <span v-if="day.champion" class="leaderboard-calendar-avatar" data-testid="leaderboard-calendar-avatar">
                        <img
                          v-if="dailyChampionAvatarUrl(day.champion)"
                          :src="dailyChampionAvatarUrl(day.champion)"
                          alt=""
                          loading="lazy"
                        >
                        <span v-else>{{ dailyChampionAvatarInitial(day.champion) }}</span>
                      </span>
                    </span>
                    <span v-if="day.champion" class="leaderboard-calendar-token-label">
                      {{ dailyChampionTokenMLabel(day.champion) }}
                    </span>
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

      <Teleport to="body">
        <div
          v-if="activeTokenBarTooltip"
          class="leaderboard-calendar-tooltip leaderboard-token-tooltip"
          :style="tokenBarTooltipStyle"
          role="tooltip"
          data-testid="leaderboard-token-tooltip"
        >
          <div class="leaderboard-calendar-tooltip-date">{{ activeTokenBarTooltip.dateLabel }}</div>
          <div class="leaderboard-calendar-tooltip-main">
            <span class="leaderboard-calendar-tooltip-name">{{ activeTokenBarTooltip.displayName }}</span>
            <span class="leaderboard-calendar-tooltip-tokens">{{ activeTokenBarTooltip.tokenLabel }}</span>
          </div>
          <div class="leaderboard-token-tooltip-table" role="table">
            <div
              v-for="metric in activeTokenBarTooltip.metrics"
              :key="metric.label"
              class="leaderboard-token-tooltip-row"
              role="row"
            >
              <span class="leaderboard-token-tooltip-label" role="cell">{{ metric.label }}</span>
              <span class="leaderboard-token-tooltip-value" role="cell">{{ metric.value }}</span>
              <span v-if="metric.note" class="leaderboard-token-tooltip-note" role="cell">{{ metric.note }}</span>
            </div>
          </div>
        </div>
      </Teleport>
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
import type { LeaderboardBadge, LeaderboardDailyRewardTopUser, LeaderboardDailyRewards, LeaderboardPeriod, LeaderboardRewardMode, UserLeaderboardDailyChampion, UserLeaderboardItem, UserLeaderboardResponse } from '@/api/usage'
import type { UserLeaderboardTokenTrendPoint } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { formatDateTime, formatNumber, formatTime } from '@/utils/format'
import { formatCreditAmount } from '@/utils/credits'
import crazyThursdayBannerUrl from '@/assets/leaderboard/crazy-thursday-v50.png'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler)

const { t } = useI18n()

const period = ref<LeaderboardPeriod>('day')
const leaderboard = ref<UserLeaderboardResponse | null>(null)
const weekLeaderboard = ref<UserLeaderboardResponse | null>(null)
const loading = ref(false)
const refreshingLeaderboard = ref(false)
const error = ref(false)
const claimingReward = ref(false)
const claimError = ref('')
const tokenTickerSeed = ref(0)
const visualTokenIncrement = ref(0)
const visualTokenTick = ref(0)
const leaderboardLimit = 10
const visibleRankTitleLimit = 2
const visualTokenTickerIntervalMs = 3000
const visualTokenTickerSteps = [37, 54, 62, 81, 95, 128, 143, 166, 218]
const championTooltipWidth = 240
const championTooltipHeight = 98
const championTooltipGap = 10
const championTooltipViewportPadding = 12
const tokenBarTooltipHeight = 136
const tokenBarTooltipGap = 18
const leaderboardSessionCacheVersion = 'v1'
const leaderboardSessionCacheTTL = 5 * 60 * 1000
let loadSeq = 0
let visualTokenTickerID: number | null = null

type RollingTokenPart = {
  type: 'digit' | 'separator'
  value: string
}

type RewardTopUserView = {
  rank: number
  displayName: string
  tokenText: string
  metaText: string
  highlighted: boolean
}

type ChampionCalendarDay = {
  key: string
  date: string
  day: number
  champion?: UserLeaderboardDailyChampion
  streak: number
}

type ChampionCalendarMonth = {
  key: string
  label: string
  days: ChampionCalendarDay[]
  isCurrent: boolean
  columnCount: number
}

type ChampionCalendarTooltipView = {
  dateLabel: string
  displayName: string
  metaLabel: string
  tokenLabel: string
}

type TokenBarTooltipMetric = {
  label: string
  value: string
  note?: string
}

type TokenBarTooltipView = {
  dateLabel: string
  displayName: string
  metrics: TokenBarTooltipMetric[]
  tokenLabel: string
}

type LeaderboardRecordProgress = {
  prefix: string
  value: string
  suffix: string
  isDeity: boolean
}

type LeaderboardSessionCacheSnapshot = {
  savedAt: number
  data: UserLeaderboardResponse
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
  { rank: 1, display_name: '落***尘', tokens: 9_800_000 },
  { rank: 2, display_name: '138****5678', tokens: 7_600_000 },
  { rank: 3, display_name: 't***d@example.com', tokens: 6_400_000 },
  { rank: 4, display_name: 'n***a@example.com', tokens: 5_200_000 },
  { rank: 5, display_name: 'API User', tokens: 4_100_000 },
  { rank: 6, display_name: 'm***o@example.com', tokens: 3_700_000 },
  { rank: 7, display_name: 'wx***88', tokens: 3_100_000 },
  { rank: 8, display_name: 'q***n@example.com', tokens: 2_600_000 },
  { rank: 9, display_name: 'dev***user', tokens: 1_900_000 },
  { rank: 10, display_name: 'last***top', tokens: 1_200_000 },
]

const periodOptions = computed(() => [
  { value: 'day' as const, label: t('leaderboard.period.day') },
  { value: 'week' as const, label: t('leaderboard.period.week') },
  { value: 'month' as const, label: t('leaderboard.period.month') },
  { value: 'all' as const, label: t('leaderboard.period.all') },
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
const championCalendarRangeLabel = computed(() => championCalendarMonths.value.map((month) => month.label).join(' / '))
const activeChampionTooltip = ref<ChampionCalendarTooltipView | null>(null)
const championTooltipPosition = ref({ left: 0, top: 0 })
const championTooltipStyle = computed(() => ({
  left: `${championTooltipPosition.value.left}px`,
  top: `${championTooltipPosition.value.top}px`,
}))
const activeTokenBarTooltip = ref<TokenBarTooltipView | null>(null)
const tokenBarTooltipPosition = ref({ left: 0, top: 0 })
const tokenBarTooltipStyle = computed(() => ({
  left: `${tokenBarTooltipPosition.value.left}px`,
  top: `${tokenBarTooltipPosition.value.top}px`,
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
const leaderboardRewardMode = computed<LeaderboardRewardMode>(() => normalizeLeaderboardRewardMode(dailyRewards.value))
const myEntry = computed<UserLeaderboardItem | null>(() => {
  if (leaderboard.value?.current_user_entry) return leaderboard.value.current_user_entry
  return rankingItems.value.find((item) => item.is_current_user) ?? null
})
const weekRankingItems = computed<UserLeaderboardItem[]>(() => {
  const visibleItems = (weekLeaderboard.value?.ranking ?? []).slice(0, leaderboardLimit)
  return visibleItems.map((item) => ({ ...item }))
})
const weekMyEntry = computed<UserLeaderboardItem | null>(() => {
  if (weekLeaderboard.value?.current_user_entry) return weekLeaderboard.value.current_user_entry
  return weekRankingItems.value.find((item) => item.is_current_user) ?? null
})
const myRankNumber = computed(() => Math.max(0, Number(myEntry.value?.rank ?? 0)))
const myTokenTotal = computed(() => Math.max(0, Math.floor(Number(myEntry.value?.tokens ?? 0))))
const myRecordHeadline = computed(() => {
  const tokens = formatRecordMillionTokens(myTokenTotal.value)
  if (myRankNumber.value > 0) {
    return t('leaderboard.record.headlineRanked', { rank: myRankNumber.value, tokens })
  }
  return t('leaderboard.record.headlineUnranked', { tokens })
})
const myRecordProgress = computed<LeaderboardRecordProgress>(() => {
  const rank = myRankNumber.value

  if (rank === 1) {
    return {
      prefix: '',
      value: t('leaderboard.record.deity'),
      suffix: '',
      isDeity: true,
    }
  }

  if (rank === 2 || rank === 3) {
    const targetRank = rank - 1
    return buildRecordProgress(
      targetRank === 1 ? t('leaderboard.record.distanceToFirst') : t('leaderboard.record.distanceToSecond'),
      findLeaderboardEntryByRank(targetRank),
    )
  }

  if (rank > 3) {
    return buildRecordProgress(t('leaderboard.record.distanceToTopThree'), findLeaderboardEntryByRank(3))
  }

  const boardEntryIndex = Math.min(leaderboardLimit, rankingItems.value.length) - 1
  const boardEntry = boardEntryIndex >= 0 ? rankingItems.value[boardEntryIndex] : null
  return buildRecordProgress(t('leaderboard.record.distanceToBoard'), boardEntry)
})
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
    line: isDark ? '#d6b79d' : '#c46f50',
    fill: isDark ? 'rgba(214, 183, 157, 0.18)' : 'rgba(196, 111, 80, 0.14)',
    text: isDark ? '#f3efe7' : '#23201c',
    muted: isDark ? '#a89f92' : '#6d675d',
    grid: isDark ? 'rgba(214, 183, 157, 0.12)' : 'rgba(84, 76, 66, 0.12)',
    tooltipBg: isDark ? 'rgba(18, 18, 16, 0.96)' : 'rgba(255, 253, 248, 0.96)',
    tooltipBorder: isDark ? 'rgba(214, 183, 157, 0.24)' : 'rgba(196, 111, 80, 0.22)',
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

const rewardTopUsers = computed<RewardTopUserView[]>(() => {
  const usersByRank = new Map<number, LeaderboardDailyRewardTopUser>()
  const sourceUsers = dailyRewards.value?.top_users?.length
    ? dailyRewards.value.top_users
    : import.meta.env.DEV
      ? devRewardTopUsers
      : []
  for (const user of sourceUsers) {
    if (user.rank >= 1 && user.rank <= leaderboardLimit && rewardTopUserHasName(user)) {
      usersByRank.set(user.rank, user)
    }
  }
  return Array.from({ length: leaderboardLimit }, (_, index) => index + 1)
    .map((rank) => {
      const user = usersByRank.get(rank)
      if (!user) return null
      return {
        rank,
        displayName: rewardTopUserDisplayName(user),
        tokenText: formatRewardTopUserTokens(user),
        metaText: rewardTopUserMetaText(user),
        highlighted: user.is_current_user === true || user.lottery_winner === true,
      }
    })
    .filter((user): user is RewardTopUserView => user != null)
})

const weeklyRushProgressText = computed(() => {
  if (!weekLeaderboard.value) return t('leaderboard.dailyReward.weeklyRushLoading')

  const progress = buildWeeklyRushProgress()
  if (progress.isDeity) return progress.value
  if (progress.value === '--') return t('leaderboard.dailyReward.weeklyRushNoData')
  return `${progress.prefix} ${progress.value}${progress.suffix ? ` ${progress.suffix}` : ''}`
})

const dailyRewardReasonText = computed(() => {
  if (leaderboardRewardMode.value === 'disabled') return t('leaderboard.dailyReward.disabled')
  const reason = dailyRewards.value?.reason
  if (reason === 'eligible') return leaderboardRewardMode.value === 'red_packet'
    ? t('leaderboard.dailyReward.redPacketReady')
    : t('leaderboard.dailyReward.eligible')
  if (reason === 'already_claimed') return t('leaderboard.dailyReward.alreadyClaimed')
  if (reason === 'settling') return t('leaderboard.dailyReward.settling', { time: formatDateTime(dailyRewards.value?.claim_available_at || '') })
  if (reason === 'threshold_not_met') return t('leaderboard.dailyReward.thresholdNotMet')
  if (reason === 'not_top_three' || reason === 'not_top_ten') return t('leaderboard.dailyReward.notTopTen')
  if (reason === 'not_ranked') return t('leaderboard.dailyReward.notRanked')
  if (reason === 'zero_reward') return t('leaderboard.dailyReward.zeroReward')
  return t('leaderboard.dailyReward.disabled')
})

const dailyRewardStatusText = computed(() => {
  if (leaderboardRewardMode.value === 'disabled') return t('leaderboard.dailyReward.statusDisabled')
  const reason = dailyRewards.value?.reason
  if (reason === 'eligible') return t('leaderboard.dailyReward.statusReady')
  if (reason === 'already_claimed') return t('leaderboard.dailyReward.statusClaimed')
  if (reason === 'settling') return t('leaderboard.dailyReward.statusSettling')
  if (reason === 'threshold_not_met') return t('leaderboard.dailyReward.statusThresholdNotMet')
  if (reason === 'not_top_three' || reason === 'not_top_ten') return t('leaderboard.dailyReward.statusNotTopTen')
  if (reason === 'not_ranked') return t('leaderboard.dailyReward.statusNotRanked')
  if (reason === 'zero_reward') return t('leaderboard.dailyReward.statusZeroReward')
  return t('leaderboard.dailyReward.statusDisabled')
})

const dailyRewardStatusClass = computed(() => {
  const reward = dailyRewards.value
  if (reward?.claimed) return 'leaderboard-reward-status--claimed'
  if (reward?.can_claim) return 'leaderboard-reward-status--ready'
  return 'leaderboard-reward-status--idle'
})

const claimButtonText = computed(() => {
  if (claimingReward.value) return t('leaderboard.dailyReward.claiming')
  if (dailyRewards.value?.claimed) return t('leaderboard.dailyReward.claimed')
  return t('leaderboard.dailyReward.redPacketClaim')
})

const redPacketRangeText = computed(() => {
  const min = Math.max(0, Number(dailyRewards.value?.red_packet_min_amount ?? 0))
  const max = Math.max(0, Number(dailyRewards.value?.red_packet_max_amount ?? 0))
  if (max > 0 && max >= min) return `${formatRewardAmount(min)} - ${formatRewardAmount(max)}`
  const pool = Math.max(0, Number(dailyRewards.value?.red_packet_pool_amount ?? 0))
  return pool > 0 ? t('leaderboard.dailyReward.redPacketPool', { amount: formatRewardAmount(pool) }) : '-'
})

const redPacketAmountLabelKey = computed(() =>
  dailyRewards.value?.claimed ? 'leaderboard.dailyReward.redPacketClaimedAmount' : 'leaderboard.dailyReward.redPacketPendingAmount'
)

const redPacketAmountText = computed(() => {
  const amount = Number(dailyRewards.value?.current_user_reward_amount ?? 0)
  if (dailyRewards.value?.claimed && amount > 0) return formatRewardAmount(amount)
  if (dailyRewards.value?.can_claim) return t('leaderboard.dailyReward.redPacketPending')
  return amount > 0 ? formatRewardAmount(amount) : '-'
})

const lotteryPrizeText = computed(() =>
  formatRewardAmount(Number(dailyRewards.value?.lottery_amount ?? 0))
)

const lotteryDrawTimeText = computed(() => {
  const drawAt = dailyRewards.value?.lottery_draw_at
  if (drawAt) return formatDateTime(drawAt)
  return dailyRewards.value?.lottery_cron?.trim() || '-'
})

const lotteryResultText = computed(() => {
  const reward = dailyRewards.value
  if (!reward) return '-'
  const winnerName = reward.lottery_winner_display_name?.trim()
    || reward.lottery_winner_email_masked?.trim()
    || ''
  if (winnerName) {
    if (reward.claimed || reward.current_user_rank === reward.lottery_winner_rank) {
      return t('leaderboard.dailyReward.lotteryWon', { amount: formatRewardAmount(Number(reward.lottery_amount ?? reward.current_user_reward_amount ?? 0)) })
    }
    return t('leaderboard.dailyReward.lotteryWinner', { name: winnerName })
  }
  if (reward.reason === 'lottery_pending' || !reward.settlement_ready) return t('leaderboard.dailyReward.lotteryPending')
  if (reward.current_user_rank > 0 && reward.current_user_rank <= leaderboardLimit) return t('leaderboard.dailyReward.lotteryNotWon')
  return t('leaderboard.dailyReward.notTopTen')
})

async function loadLeaderboard() {
  const currentSeq = ++loadSeq
  error.value = false
  claimError.value = ''
  const cached = readLeaderboardSessionCache(period.value)
  if (cached) {
    leaderboard.value = cached
    visualTokenIncrement.value = 0
    tokenTickerSeed.value += 1
  } else {
    leaderboard.value = null
  }
  loading.value = !cached
  refreshingLeaderboard.value = !!cached

  try {
    const response = await usageAPI.getDashboardLeaderboard({ period: period.value, limit: leaderboardLimit })
    if (currentSeq !== loadSeq) return
    leaderboard.value = response
    writeLeaderboardSessionCache(period.value, response)
    visualTokenIncrement.value = 0
    tokenTickerSeed.value += 1
  } catch (err) {
    if (currentSeq !== loadSeq) return
    console.error('Failed to load leaderboard:', err)
    error.value = !leaderboard.value
  } finally {
    if (currentSeq === loadSeq) {
      loading.value = false
      refreshingLeaderboard.value = false
    }
  }

  void loadWeekLeaderboardSummary()
}

async function loadWeekLeaderboardSummary() {
  if (period.value === 'week' && leaderboard.value) {
    weekLeaderboard.value = leaderboard.value
    return
  }

  const cached = readLeaderboardSessionCache('week')
  if (cached) {
    weekLeaderboard.value = cached
    return
  }

  try {
    const response = await usageAPI.getDashboardLeaderboard({ period: 'week', limit: leaderboardLimit })
    weekLeaderboard.value = response
    writeLeaderboardSessionCache('week', response)
  } catch (err) {
    console.error('Failed to load weekly leaderboard summary:', err)
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
      applyClaimedBalance(result.claimed_amount ?? result.red_packet_amount ?? result.lottery_amount ?? 0)
      writeLeaderboardSessionCache(period.value, leaderboard.value)
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

function leaderboardSessionCacheKey(value: LeaderboardPeriod): string {
  return `sub2api:user-leaderboard:${leaderboardSessionCacheVersion}:${value}:${leaderboardLimit}`
}

function readLeaderboardSessionCache(value: LeaderboardPeriod): UserLeaderboardResponse | null {
  if (typeof window === 'undefined') return null
  try {
    const raw = window.sessionStorage.getItem(leaderboardSessionCacheKey(value))
    if (!raw) return null
    const snapshot = JSON.parse(raw) as Partial<LeaderboardSessionCacheSnapshot>
    if (!snapshot || typeof snapshot.savedAt !== 'number' || !snapshot.data) return null
    if (Date.now() - snapshot.savedAt > leaderboardSessionCacheTTL) {
      window.sessionStorage.removeItem(leaderboardSessionCacheKey(value))
      return null
    }
    return snapshot.data
  } catch {
    return null
  }
}

function writeLeaderboardSessionCache(value: LeaderboardPeriod, data: UserLeaderboardResponse): void {
  if (typeof window === 'undefined') return
  try {
    const snapshot: LeaderboardSessionCacheSnapshot = {
      savedAt: Date.now(),
      data,
    }
    window.sessionStorage.setItem(leaderboardSessionCacheKey(value), JSON.stringify(snapshot))
  } catch {
    // Cache failures should never block ranking display.
  }
}

function selectPeriod(value: LeaderboardPeriod) {
  if (period.value === value) return
  period.value = value
  hideChampionTooltip()
  hideTokenBarTooltip()
  loadLeaderboard()
}

function showTokenBarTooltip(item: UserLeaderboardItem, event: Event) {
  const target = event.currentTarget
  if (!(target instanceof HTMLElement)) return

  activeTokenBarTooltip.value = {
    dateLabel: leaderboardTokenTooltipDateLabel(),
    displayName: getLeaderboardDisplayName(item),
    metrics: leaderboardTokenTooltipMetrics(item),
    tokenLabel: `${formatCompactChineseTokens(item.tokens)} tokens`,
  }
  positionTokenBarTooltip(target)
}

function leaderboardTokenTooltipDateLabel(): string {
  const generatedAt = leaderboard.value?.generated_at
  return generatedAt ? formatChampionCalendarDate(formatCalendarDateKey(new Date(generatedAt))) : t('leaderboard.tokens')
}

function leaderboardTokenTooltipMetrics(item: UserLeaderboardItem): TokenBarTooltipMetric[] {
  return [
    {
      label: t('leaderboard.inputTokensShort'),
      value: formatCompactChineseTokens(leaderboardInputTokens(item)),
    },
    {
      label: t('leaderboard.outputTokensShort'),
      value: formatCompactChineseTokens(leaderboardOutputTokens(item)),
    },
    {
      label: t('leaderboard.cacheTokensShort'),
      value: formatCompactChineseTokens(leaderboardCacheTokens(item)),
      note: `${t('leaderboard.cacheRatioShort')} ${formatLeaderboardCachePercent(item)}`,
    },
    {
      label: t('leaderboard.costPerMillionShort'),
      value: formatLeaderboardCostPerMillion(item),
      note: '/ 1M Token',
    },
  ]
}

function positionTokenBarTooltip(target: HTMLElement) {
  const rect = target.getBoundingClientRect()
  const width = 256
  const viewportWidth = window.innerWidth || document.documentElement.clientWidth || width
  const maxLeft = Math.max(championTooltipViewportPadding, viewportWidth - width - championTooltipViewportPadding)
  const preferredLeft = rect.left + window.scrollX + Math.min(24, rect.width * 0.35)
  const left = Math.min(Math.max(championTooltipViewportPadding, preferredLeft), maxLeft)
  const top = Math.max(
    championTooltipViewportPadding,
    rect.top + window.scrollY - tokenBarTooltipGap - tokenBarTooltipHeight,
  )
  tokenBarTooltipPosition.value = { left, top }
}

function hideTokenBarTooltip() {
  activeTokenBarTooltip.value = null
}

function showChampionTooltip(day: ChampionCalendarDay, event: Event) {
  if (!day.champion) {
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

function findLeaderboardEntryByRank(rank: number): UserLeaderboardItem | null {
  return rankingItems.value.find((item) => item.rank === rank) ?? null
}

function buildRecordProgress(prefix: string, target: UserLeaderboardItem | null): LeaderboardRecordProgress {
  if (!target) {
    return {
      prefix,
      value: '--',
      suffix: '',
      isDeity: false,
    }
  }

  const distance = Math.max(0, Math.floor(Number(target.tokens ?? 0)) - myTokenTotal.value + 1)
  return {
    prefix,
    value: formatRecordDistanceTokens(distance),
    suffix: 'token',
    isDeity: false,
  }
}

function buildWeeklyRushProgress(): LeaderboardRecordProgress {
  const rank = Math.max(0, Number(weekMyEntry.value?.rank ?? 0))
  const myTokens = Math.max(0, Math.floor(Number(weekMyEntry.value?.tokens ?? 0)))

  if (rank === 1) {
    return {
      prefix: '',
      value: t('leaderboard.record.deity'),
      suffix: '',
      isDeity: true,
    }
  }

  if (rank === 2 || rank === 3) {
    const targetRank = rank - 1
    return buildDistanceProgress(
      targetRank === 1 ? t('leaderboard.record.distanceToFirst') : t('leaderboard.record.distanceToSecond'),
      findWeekLeaderboardEntryByRank(targetRank),
      myTokens,
    )
  }

  if (rank > 3) {
    return buildDistanceProgress(t('leaderboard.record.distanceToTopThree'), findWeekLeaderboardEntryByRank(3), myTokens)
  }

  const boardEntryIndex = Math.min(leaderboardLimit, weekRankingItems.value.length) - 1
  const boardEntry = boardEntryIndex >= 0 ? weekRankingItems.value[boardEntryIndex] : null
  return buildDistanceProgress(t('leaderboard.record.distanceToBoard'), boardEntry, myTokens)
}

function findWeekLeaderboardEntryByRank(rank: number): UserLeaderboardItem | null {
  return weekRankingItems.value.find((item) => item.rank === rank) ?? null
}

function buildDistanceProgress(prefix: string, target: UserLeaderboardItem | null, currentTokens: number): LeaderboardRecordProgress {
  if (!target) {
    return {
      prefix,
      value: '--',
      suffix: '',
      isDeity: false,
    }
  }

  const distance = Math.max(0, Math.floor(Number(target.tokens ?? 0)) - currentTokens + 1)
  return {
    prefix,
    value: formatRecordDistanceTokens(distance),
    suffix: 'token',
    isDeity: false,
  }
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
  const anchor = championCalendarAnchor.value
  const visibleDays = isCurrent && year === anchor.getFullYear() && month === anchor.getMonth()
    ? Math.min(daysInMonth, Math.max(1, anchor.getDate()))
    : daysInMonth
  const key = formatCalendarDateKey(monthStart)
  let previousChampionUserID: number | null = null
  let streak = 0
  const days = Array.from({ length: visibleDays }, (_, index) => {
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
    }
  })

  return {
    key,
    label: `${year}年${month + 1}月`,
    isCurrent,
    days,
    columnCount: Math.min(16, Math.max(1, visibleDays)),
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

function formatTokenMillions(value: number): string {
  if (!Number.isFinite(value)) return '0M'
  const millions = Math.max(0, value) / 1_000_000
  if (millions > 0 && millions < 0.1) return '<0.1M'
  return `${formatCompactUnit(millions)}M`
}

function formatRecordMillionTokens(value: number): string {
  return formatTokenMillions(value)
}

function formatRecordDistanceTokens(value: number): string {
  if (value >= 100_000_000) return `${formatCompactUnit(value / 100_000_000)}亿`
  if (value >= 10_000) return `${formatCompactUnit(value / 10_000)}万`
  return formatNumber(value)
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

function dailyChampionTokenMLabel(champion: UserLeaderboardDailyChampion): string {
  return formatTokenMillions(champion.tokens)
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
  return day.champion ? championCalendarTooltip(day) : championCalendarEmptyLabel(day)
}

function championCalendarCellTabIndex(day: ChampionCalendarDay): number | undefined {
  return day.champion ? 0 : undefined
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
    '--token-bar-value-left': widthText,
    '--token-input-width': leaderboardTokenSegmentPercent(leaderboardInputTokens(item), item.tokens),
    '--token-output-width': leaderboardTokenSegmentPercent(leaderboardOutputTokens(item), item.tokens),
    '--token-cache-width': leaderboardTokenSegmentPercent(leaderboardCacheTokens(item), item.tokens),
    '--token-bar-color': palette.color,
    '--token-bar-glow': palette.glow,
    '--token-rank-color': palette.text,
    '--token-value-color': palette.text,
  }
}

function tokenBarPalette(rank: number): { color: string; glow: string; text: string } {
  if (rank === 1) return { color: 'rgb(196 111 80)', glow: 'rgb(196 111 80 / 0.22)', text: 'rgb(116 61 45)' }
  if (rank === 2) return { color: 'rgb(95 143 129)', glow: 'rgb(95 143 129 / 0.18)', text: 'rgb(57 98 88)' }
  if (rank === 3) return { color: 'rgb(132 118 98)', glow: 'rgb(132 118 98 / 0.16)', text: 'rgb(86 76 63)' }
  return { color: 'rgb(118 111 101)', glow: 'rgb(118 111 101 / 0.12)', text: 'rgb(74 68 59)' }
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

function leaderboardRankChangeValue(item: UserLeaderboardItem): number | null {
  const value = item.rank_change
  if (typeof value !== 'number' || !Number.isFinite(value)) return null
  return Math.trunc(value)
}

function leaderboardRankChangeLabel(item: UserLeaderboardItem): string {
  if (item.rank_new) return t('leaderboard.rankChangeNew')
  const value = leaderboardRankChangeValue(item)
  if (value == null) return ''
  if (value > 0) return `↑+${value}`
  if (value < 0) return `↓${value}`
  return '0'
}

function leaderboardRankChangeClass(item: UserLeaderboardItem): string {
  if (item.rank_new) return 'leaderboard-rank-change--new'
  const value = leaderboardRankChangeValue(item)
  if (value == null) return ''
  if (value > 0) return 'leaderboard-rank-change--up'
  if (value < 0) return 'leaderboard-rank-change--down'
  return 'leaderboard-rank-change--same'
}

function leaderboardRankChangeTitle(item: UserLeaderboardItem): string {
  const periodLabel = leaderboardRankComparedLabel()
  if (item.rank_new) return t('leaderboard.rankChangeTitle.new', { period: periodLabel })
  const value = leaderboardRankChangeValue(item)
  if (value == null) return ''
  if (value > 0) return t('leaderboard.rankChangeTitle.up', { period: periodLabel, count: value })
  if (value < 0) return t('leaderboard.rankChangeTitle.down', { period: periodLabel, count: Math.abs(value) })
  return t('leaderboard.rankChangeTitle.same', { period: periodLabel })
}

function leaderboardRankComparedLabel(): string {
  if (period.value === 'day') return t('leaderboard.rankChangeCompared.day')
  if (period.value === 'week') return t('leaderboard.rankChangeCompared.week')
  if (period.value === 'month') return t('leaderboard.rankChangeCompared.month')
  return ''
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

function formatRewardAmount(value: number): string {
  return formatCreditAmount(value, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

function formatRewardTopUserTokens(user: LeaderboardDailyRewardTopUser): string {
  const tokens = Math.max(0, Math.floor(Number(user.tokens ?? 0)))
  return tokens > 0 ? t('leaderboard.dailyReward.topUserTokens', { tokens: formatCompactChineseTokens(tokens) }) : '-'
}

const rewardWeekRangeLabel = computed(() => {
  return formatShortDateRange(dailyRewards.value?.reward_date)
})

function formatShortDateRange(value?: string): string {
  const parts = String(value || '')
    .split('~')
    .map((part) => formatShortMonthDay(part.trim()))
    .filter(Boolean)
  if (parts.length >= 2) return `${parts[0]}~${parts[1]}`
  return parts[0] || '-'
}

function formatShortMonthDay(value?: string): string {
  const [, month = '', day = ''] = String(value || '').match(/^\d{4}-(\d{2})-(\d{2})/) ?? []
  if (!month || !day) return ''
  return `${month}-${day}`
}

function rewardTopUserMetaText(user: LeaderboardDailyRewardTopUser): string {
  if (leaderboardRewardMode.value === 'red_packet') {
    const amount = Number(user.claimed_amount ?? user.red_packet_amount ?? 0)
    if (amount > 0) return t('leaderboard.dailyReward.redPacketUserClaimed', { amount: formatRewardAmount(amount) })
    if (user.claimed) return t('leaderboard.dailyReward.redPacketUserClaimedNoAmount')
  }
  if (leaderboardRewardMode.value === 'lottery' && user.lottery_winner) {
    const amount = Number(user.lottery_amount ?? dailyRewards.value?.lottery_amount ?? 0)
    return t('leaderboard.dailyReward.lotteryUserWinner', { amount: formatRewardAmount(amount) })
  }
  if (user.is_current_user) return t('leaderboard.currentUser')
  return ''
}

function normalizeLeaderboardRewardMode(reward: LeaderboardDailyRewards | null): LeaderboardRewardMode {
  const mode = reward?.reward_mode
  if (mode === 'red_packet' || mode === 'lottery' || mode === 'disabled') return mode
  return reward?.enabled ? 'red_packet' : 'disabled'
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
  return String(rank)
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

function hiddenLeaderboardBadgeTitle(badges: LeaderboardBadge[] = []): string {
  return badges
    .filter((badge) => leaderboardTitleBadges.includes(badge))
    .slice(visibleRankTitleLimit)
    .map((badge) => leaderboardBadgeTitle(badge))
    .join(' / ')
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

onMounted(() => {
  startVisualTokenTicker()
  loadLeaderboard()
  window.addEventListener('scroll', hideChampionTooltip, true)
  window.addEventListener('scroll', hideTokenBarTooltip, true)
  window.addEventListener('resize', hideChampionTooltip)
  window.addEventListener('resize', hideTokenBarTooltip)
})

onUnmounted(() => {
  stopVisualTokenTicker()
  hideChampionTooltip()
  hideTokenBarTooltip()
  window.removeEventListener('scroll', hideChampionTooltip, true)
  window.removeEventListener('scroll', hideTokenBarTooltip, true)
  window.removeEventListener('resize', hideChampionTooltip)
  window.removeEventListener('resize', hideTokenBarTooltip)
})
</script>

<style scoped>
.leaderboard-page {
  margin: -1rem;
  min-height: calc(100vh - 4rem);
  padding: 1rem;
  background:
    linear-gradient(180deg, rgb(250 249 245 / 0.96), rgb(245 241 232 / 0.88)),
    rgb(250 249 245);
  color: rgb(35 32 28);
}

.leaderboard-section-label {
  color: rgb(109 103 93);
}

.leaderboard-state-card {
  border: 1px solid rgb(214 202 186 / 0.72);
  border-radius: 0.75rem;
  background:
    linear-gradient(180deg, rgb(255 253 248 / 0.92), rgb(250 247 239 / 0.76)),
    rgb(255 253 248);
  box-shadow: 0 0.85rem 2rem rgb(60 49 36 / 0.05);
}

.leaderboard-state-title {
  color: rgb(35 32 28);
}

.leaderboard-loading-spinner {
  border-color: rgb(196 111 80);
  border-top-color: transparent;
}

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
  color: rgb(109 103 93);
  font-size: 0.8125rem;
  letter-spacing: 0;
}

.leaderboard-token-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgb(214 202 186 / 0.72);
  border-radius: 0.75rem;
  background:
    linear-gradient(135deg, rgb(255 253 248 / 0.96), rgb(247 240 229 / 0.86)),
    rgb(255 253 248);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.72),
    0 0.85rem 2rem rgb(60 49 36 / 0.06);
}

.leaderboard-token-trend-panel {
  display: grid;
  min-width: 0;
  gap: 0.52rem;
  border-left: 1px solid rgb(214 202 186 / 0.74);
  padding-left: 1.35rem;
}

.leaderboard-token-trend-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: rgb(68 62 54);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0;
}

.leaderboard-token-trend-legend {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: rgb(154 83 62);
  font-size: 0.72rem;
  white-space: nowrap;
}

.leaderboard-token-trend-legend::before {
  width: 0.55rem;
  height: 0.55rem;
  border: 1px solid currentColor;
  border-radius: 9999px;
  background: rgb(196 111 80 / 0.13);
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
  border: 1px solid rgb(214 202 186 / 0.62);
  background:
    linear-gradient(90deg, rgb(132 118 98 / 0.08) 1px, transparent 1px),
    linear-gradient(rgb(132 118 98 / 0.08) 1px, transparent 1px),
    rgb(255 253 248 / 0.58);
  background-size: 12.5% 100%, 100% 1.65rem;
  color: rgb(109 103 93);
  font-size: 0.8125rem;
}

.leaderboard-token-card::after {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(90deg, rgb(132 118 98 / 0.06) 1px, transparent 1px),
    linear-gradient(rgb(132 118 98 / 0.05) 1px, transparent 1px),
    linear-gradient(90deg, transparent, rgb(196 111 80 / 0.08), transparent);
  background-size: 5.5rem 100%, 100% 1rem, 100% 100%;
  content: "";
  opacity: 0.54;
}

.leaderboard-ranking-switch {
  display: inline-flex;
  width: fit-content;
  max-width: 100%;
  overflow-x: auto;
  border: 1px solid rgb(43 40 35 / 0.14);
  border-radius: 0.375rem;
  background: rgb(240 233 221 / 0.82);
  padding: 0.16rem;
  scrollbar-width: none;
}

.leaderboard-ranking-switch::-webkit-scrollbar {
  display: none;
}

.leaderboard-ranking-switch-button {
  flex: 0 0 auto;
  border-radius: 0.25rem;
  padding: 0.34rem 0.74rem;
  font-size: 0.8125rem;
  font-weight: 800;
  letter-spacing: 0;
  transition: background-color 0.16s ease, color 0.16s ease, box-shadow 0.16s ease;
  white-space: nowrap;
}

.leaderboard-ranking-switch-button--active {
  background: rgb(35 32 28);
  color: rgb(255 253 248);
  box-shadow: 0 0.34rem 0.9rem rgb(35 32 28 / 0.14);
}

.leaderboard-ranking-switch-button--idle {
  color: rgb(109 103 93);
}

.leaderboard-ranking-switch-button--idle:hover {
  color: rgb(35 32 28);
}

.leaderboard-ranking-card-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.leaderboard-ranking-toolbar-meta {
  display: flex;
  min-width: 0;
  flex: 1 1 auto;
  align-items: center;
  justify-content: flex-end;
  gap: 0.8rem;
}

.leaderboard-token-legend {
  display: inline-flex;
  flex: 0 1 auto;
  min-width: 0;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.45rem 0.7rem;
}

.leaderboard-token-legend-item {
  display: inline-flex;
  align-items: center;
  gap: 0.28rem;
  color: rgb(109 103 93);
  font-size: 0.75rem;
  font-weight: 800;
  line-height: 1;
  white-space: nowrap;
}

.leaderboard-token-legend-dot {
  width: 0.62rem;
  height: 0.62rem;
  border-radius: 0.18rem;
  box-shadow: inset 0 0 0 1px rgb(255 255 255 / 0.36);
}

.leaderboard-token-legend-dot--input {
  background: linear-gradient(135deg, rgb(20 20 19), rgb(64 58 50));
}

.leaderboard-token-legend-dot--output {
  background: linear-gradient(135deg, rgb(154 83 62), rgb(196 111 80));
}

.leaderboard-token-legend-dot--cache {
  background: linear-gradient(135deg, rgb(72 118 105), rgb(95 143 129));
}

.leaderboard-ranking-empty {
  display: flex;
  min-height: 12rem;
  align-items: center;
  justify-content: center;
  border-top: 1px solid rgb(214 202 186 / 0.62);
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
  color: rgb(68 41 32);
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
  border: 1px solid rgb(196 111 80 / 0.24);
  border-radius: 0.2rem;
  background:
    linear-gradient(180deg, rgb(255 253 248 / 0.98), rgb(240 233 221 / 0.78)),
    radial-gradient(circle at 50% 0%, rgb(196 111 80 / 0.14), transparent 60%);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.74),
    inset 0 -0.18rem 0 rgb(68 41 32 / 0.04),
    0 0.28rem 0.8rem rgb(154 83 62 / 0.1);
}

.leaderboard-token-ranking-card {
  overflow: hidden;
  border: 1px solid rgb(214 202 186 / 0.72);
  border-radius: 0.75rem;
  background:
    linear-gradient(180deg, rgb(255 253 248 / 0.92), rgb(250 247 239 / 0.78)),
    rgb(255 253 248);
  padding: 1.25rem;
  box-shadow: 0 0.85rem 2rem rgb(60 49 36 / 0.06);
}

.leaderboard-token-ranking-card {
  overflow: hidden;
}

.leaderboard-token-ranking-updated {
  flex: 0 0 auto;
  color: rgb(109 103 93);
  font-size: 0.8125rem;
  white-space: nowrap;
}

.leaderboard-token-ranking-refreshing {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  border: 1px solid rgb(196 111 80 / 0.26);
  border-radius: 9999px;
  background: rgb(196 111 80 / 0.08);
  padding: 0.18rem 0.52rem;
  color: rgb(116 61 45);
  font-size: 0.75rem;
  font-weight: 700;
  white-space: nowrap;
}

.leaderboard-token-rank-list {
  display: grid;
  gap: 0.42rem;
  padding-top: 0.82rem;
}

.leaderboard-token-rank-row {
  display: grid;
  grid-template-columns: minmax(13.25rem, 14.5rem) minmax(14rem, 1fr);
  align-items: center;
  gap: 0.68rem;
  min-height: 3.65rem;
  padding: 0.16rem 0;
}

.leaderboard-token-rank-row-current {
  border-radius: 0.45rem;
  background: rgb(95 143 129 / 0.1);
}

.leaderboard-token-rank-user {
  display: grid;
  min-width: 0;
  grid-template-columns: 2.5rem 2.9rem minmax(0, 1fr);
  grid-template-rows: minmax(1.2rem, auto) minmax(1.05rem, auto);
  align-items: center;
  gap: 0.6rem;
  row-gap: 0.2rem;
}

.leaderboard-token-rank-main {
  display: contents;
}

.leaderboard-token-rank-name-line {
  display: flex;
  grid-column: 3;
  grid-row: 1;
  min-width: 0;
  max-width: 100%;
  align-items: center;
  gap: 0.32rem;
}

.leaderboard-token-rank-name {
  min-width: 0;
  grid-column: 3;
  grid-row: 1;
  max-width: 100%;
  overflow: hidden;
  color: var(--token-rank-color);
  font-size: 0.875rem;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.leaderboard-rank-change {
  display: inline-flex;
  flex: 0 0 auto;
  align-self: center;
  border: 1px solid currentColor;
  border-radius: 9999px;
  background: rgb(255 253 248 / 0.78);
  padding: 0.06rem 0.32rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.66rem;
  font-weight: 900;
  line-height: 1.05;
}

.leaderboard-rank-change--up {
  color: rgb(178 67 54);
}

.leaderboard-rank-change--down {
  color: rgb(57 98 88);
}

.leaderboard-rank-change--same {
  color: rgb(118 111 101);
}

.leaderboard-rank-change--new {
  border-color: rgb(196 111 80 / 0.48);
  background: rgb(196 111 80 / 0.1);
  color: rgb(154 83 62);
}

.leaderboard-token-rank-user--name-only .leaderboard-token-rank-name-line {
  grid-row: 1 / span 2;
  align-self: center;
}

.leaderboard-token-rank-user--name-only .leaderboard-token-rank-name {
  align-self: center;
}

.leaderboard-token-bar-area {
  display: block;
  min-width: 0;
}

.leaderboard-token-bar-track {
  position: relative;
  display: grid;
  box-sizing: border-box;
  grid-template-columns: minmax(0, 1fr) 5.9rem;
  column-gap: 0.7rem;
  align-items: center;
  height: 1.32rem;
  overflow: visible;
  border-radius: 0;
  background: transparent;
}

.leaderboard-token-bar-track::before {
  position: absolute;
  top: 50%;
  right: 6.6rem;
  left: 0;
  height: 1px;
  transform: translateY(-50%);
  background: rgb(132 118 98 / 0.055);
  content: "";
}

.leaderboard-token-bar-track::after {
  position: absolute;
  top: 0;
  right: 6.6rem;
  bottom: 0;
  left: 0;
  background:
    linear-gradient(90deg, rgb(132 118 98 / 0.032) 1px, transparent 1px),
    transparent;
  background-size: 10% 100%;
  content: "";
  pointer-events: none;
}

.leaderboard-token-bar-fill {
  position: relative;
  z-index: 1;
  display: flex;
  grid-column: 1;
  grid-row: 1;
  width: var(--token-bar-width);
  height: 100%;
  min-width: 1.05rem;
  overflow: hidden;
  border-radius: 0.08rem 0.28rem 0.28rem 0.08rem;
  background: rgb(222 212 196 / 0.58);
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
  color: rgb(118 111 101);
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
  border: 2px solid rgb(255 253 248 / 0.96);
  border-radius: 9999px;
  background:
    linear-gradient(135deg, rgb(255 253 248 / 0.98), rgb(240 233 221 / 0.86)),
    var(--token-bar-color);
  box-shadow:
    0 0 0 1px rgb(35 32 28 / 0.09),
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
  background: rgb(95 143 129 / 0.14);
  padding: 0.1rem 0.45rem;
  color: rgb(57 98 88);
  font-size: 0.6875rem;
  font-weight: 700;
}

.leaderboard-token-title-list {
  display: flex;
  grid-column: 3;
  grid-row: 2;
  max-width: 100%;
  max-height: 1.45rem;
  min-width: 0;
  flex-wrap: wrap;
  justify-content: flex-start;
  gap: 0.24rem;
  overflow: hidden;
}

.leaderboard-token-title-badge,
.leaderboard-token-title-more {
  display: inline-flex;
  max-width: 5.5rem;
  align-items: center;
  overflow: hidden;
  border: 1px solid currentColor;
  border-radius: 0.25rem;
  background: rgb(255 253 248 / 0.66);
  padding: 0.1rem 0.42rem;
  color: var(--token-rank-color);
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1.1;
  opacity: 0.84;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.leaderboard-token-title-more {
  max-width: none;
  color: rgb(109 103 93);
}

.leaderboard-token-bar-value {
  position: relative;
  z-index: 2;
  grid-column: 1;
  grid-row: 1;
  justify-self: start;
  margin-left: var(--token-bar-value-left);
  min-width: 0;
  transform: translateX(0.6rem);
  color: var(--token-value-color);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.95rem;
  font-weight: 800;
  line-height: 1;
  white-space: nowrap;
}

.leaderboard-side-stack {
  min-width: 0;
}

.leaderboard-side-card {
  border: 1px solid rgb(214 202 186 / 0.46);
  border-radius: 0.75rem;
  background:
    linear-gradient(180deg, rgb(255 253 248 / 0.66), rgb(250 247 239 / 0.34)),
    rgb(255 253 248);
  box-shadow: 0 0.25rem 0.8rem rgb(60 49 36 / 0.025);
}

.leaderboard-record-card {
  min-width: 0;
  border: 1px solid rgb(214 202 186 / 0.42);
  border-radius: 0.72rem;
  background:
    linear-gradient(135deg, rgb(255 253 248 / 0.92), rgb(239 235 226 / 0.64)),
    rgb(250 247 239);
  padding: 1.05rem 1.12rem;
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.75),
    0 0.45rem 1.25rem rgb(60 49 36 / 0.04);
}

.leaderboard-thursday-banner {
  position: relative;
  height: 5.75rem;
  overflow: hidden;
  border: 1px solid rgb(218 184 132 / 0.35);
  border-radius: 0.5rem;
  background: rgb(247 240 229);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.55),
    0 0.35rem 0.9rem rgb(196 111 80 / 0.08);
}

.leaderboard-thursday-banner + .leaderboard-record-card {
  margin-top: 0.85rem;
}

.leaderboard-thursday-banner img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 56%;
}

.leaderboard-thursday-banner::after {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, rgb(112 39 27 / 0.7), rgb(112 39 27 / 0.32) 42%, transparent 76%);
  pointer-events: none;
  content: "";
}

.leaderboard-thursday-banner-copy {
  position: absolute;
  inset: 0 auto 0 0;
  z-index: 1;
  display: flex;
  width: 48%;
  min-width: 8.6rem;
  flex-direction: column;
  justify-content: center;
  padding: 0.75rem 0.8rem 0.78rem;
  color: rgb(255 253 248);
  line-height: 1;
  text-shadow: 0 1px 0 rgb(68 41 32 / 0.28);
}

.leaderboard-thursday-banner-copy span,
.leaderboard-thursday-banner-copy strong {
  display: block;
  overflow-wrap: anywhere;
  letter-spacing: 0;
}

.leaderboard-thursday-banner-copy span {
  font-size: 0.82rem;
  font-weight: 900;
}

.leaderboard-thursday-banner-copy strong {
  margin-top: 0.34rem;
  font-size: 1.72rem;
  font-weight: 950;
}

.leaderboard-record-kicker {
  color: rgb(109 103 93);
  font-size: 0.78rem;
  font-weight: 500;
  letter-spacing: 0.08em;
  line-height: 1.35;
}

.leaderboard-record-headline {
  margin-top: 0.6rem;
  color: rgb(35 32 28);
  font-size: clamp(0.98rem, 1.5vw, 1.12rem);
  font-weight: 900;
  letter-spacing: 0;
  line-height: 1.38;
}

.leaderboard-record-progress {
  display: flex;
  min-width: 0;
  align-items: baseline;
  gap: 0.42rem;
  margin-top: 0.72rem;
  color: rgb(86 80 71);
  font-size: 0.92rem;
  line-height: 1.35;
  white-space: nowrap;
}

.leaderboard-record-progress strong {
  color: rgb(178 67 54);
  font-size: clamp(1.32rem, 2.6vw, 1.72rem);
  font-weight: 900;
  letter-spacing: 0.01em;
  line-height: 1.1;
}

.leaderboard-record-unit {
  margin-left: 0.32rem;
}

.leaderboard-record-progress--deity strong {
  color: rgb(116 61 45);
  font-size: clamp(1.18rem, 2.1vw, 1.48rem);
}

.leaderboard-reward-head {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.leaderboard-reward-panel {
  margin-top: 1rem;
  border-top: 1px solid rgb(214 202 186 / 0.44);
  padding-top: 0.95rem;
}

.leaderboard-reward-title {
  color: rgb(35 32 28);
  font-size: 0.95rem;
  font-weight: 800;
  letter-spacing: 0;
  line-height: 1.3;
}

.leaderboard-reward-status {
  flex: 0 0 auto;
  display: inline-flex;
  max-width: 7.8rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  margin-top: 0.1rem;
  padding: 0.34rem 0.66rem;
  font-size: 0.7rem;
  font-weight: 700;
  line-height: 1.12;
  text-align: center;
  white-space: nowrap;
}

.leaderboard-reward-status--claimed {
  background: rgb(232 241 236);
  color: rgb(57 98 88);
}

.leaderboard-reward-status--ready {
  background: rgb(248 231 224);
  color: rgb(116 61 45);
}

.leaderboard-reward-status--idle {
  background: rgb(240 233 221);
  color: rgb(109 103 93);
}

.leaderboard-reward-progress-card,
.leaderboard-reward-mode-card {
  border: 1px solid rgb(214 202 186 / 0.34);
  background: rgb(250 247 239 / 0.34);
}

.leaderboard-reward-progress-track {
  background: rgb(222 212 196 / 0.72);
}

.leaderboard-reward-progress-fill {
  background: linear-gradient(90deg, rgb(95 143 129), rgb(196 111 80));
}

.leaderboard-weekly-rush-value {
  min-width: 0;
  overflow: hidden;
  color: rgb(116 61 45);
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.leaderboard-reward-claim {
  border-color: rgb(35 32 28);
  background: rgb(35 32 28);
  color: rgb(255 253 248);
}

.leaderboard-reward-claim:not(:disabled):hover {
  background: rgb(64 58 50);
}

.leaderboard-weekly-winners {
  display: grid;
  gap: 0.6rem;
  border: 1px solid rgb(214 202 186 / 0.4);
  border-radius: 0.5rem;
  background: rgb(250 247 239 / 0.38);
  padding: 0.65rem;
}

.leaderboard-weekly-winners-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: rgb(109 103 93);
  font-size: 0.75rem;
  font-weight: 700;
}

.leaderboard-weekly-winners-list {
  display: grid;
  gap: 0.45rem;
  max-height: 11.7rem;
  overflow-y: auto;
  padding-right: 0.18rem;
  scrollbar-color: rgb(181 166 143 / 0.7) transparent;
  scrollbar-width: thin;
}

.leaderboard-weekly-winners-list::-webkit-scrollbar {
  width: 0.38rem;
}

.leaderboard-weekly-winners-list::-webkit-scrollbar-thumb {
  border-radius: 9999px;
  background: rgb(181 166 143 / 0.72);
}

.leaderboard-weekly-winners-list::-webkit-scrollbar-track {
  background: transparent;
}

.leaderboard-weekly-winner-row {
  display: grid;
  min-width: 0;
  grid-template-columns: minmax(3.5rem, auto) minmax(0, 1fr) auto auto;
  align-items: center;
  gap: 0.65rem;
}

.leaderboard-weekly-winner-row--highlighted {
  font-weight: 800;
}

.leaderboard-weekly-winner-rank {
  color: rgb(109 103 93);
  font-size: 0.8125rem;
}

.leaderboard-weekly-winner-name {
  min-width: 0;
  overflow: hidden;
  color: rgb(35 32 28);
  font-size: 0.875rem;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.leaderboard-weekly-winner-meta,
.leaderboard-weekly-winner-tokens {
  color: rgb(109 103 93);
  font-size: 0.75rem;
  font-weight: 700;
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
  background: linear-gradient(180deg, rgb(255 253 248 / 0.96), transparent);
}

.leaderboard-token-reel::after {
  bottom: 0;
  background: linear-gradient(0deg, rgb(68 41 32 / 0.12), transparent);
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
  color: rgb(154 83 62);
  text-shadow: 0 1px 0 rgb(255 255 255 / 0.78);
}

.leaderboard-calendar-card {
  overflow: hidden;
  border: 1px solid rgb(214 202 186 / 0.62);
  border-radius: 0.75rem;
  background:
    linear-gradient(180deg, rgb(255 253 248 / 0.92), rgb(250 247 239 / 0.62)),
    rgb(255 253 248);
  padding: 1rem 1rem 1.05rem;
  box-shadow: 0 0.85rem 2rem rgb(60 49 36 / 0.06);
}

.leaderboard-calendar-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem;
  padding-bottom: 0.8rem;
}

.leaderboard-calendar-title {
  flex: 1 1 auto;
  color: rgb(35 32 28);
  font-size: 1.12rem;
  font-weight: 900;
  letter-spacing: 0;
  line-height: 1.25;
  text-align: left;
}

.leaderboard-calendar-meta {
  flex: 0 0 auto;
  color: rgb(109 103 93);
  font-size: 0.78rem;
  font-weight: 800;
  line-height: 1;
  white-space: nowrap;
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
  background: rgb(132 118 98 / 0.28);
}

.leaderboard-calendar-months {
  --calendar-day-size: 4.08rem;
  --calendar-day-label-height: 1rem;
  --calendar-day-token-height: 0.9rem;
  --calendar-day-gap-x: 0.44rem;
  --calendar-month-padding: 0.85rem;
  --calendar-month-min-width: calc(
    (var(--calendar-day-size) * 16) +
    (var(--calendar-day-gap-x) * 15) +
    (var(--calendar-month-padding) * 2)
  );
  display: grid;
  width: 100%;
  min-width: min(100%, var(--calendar-month-min-width));
  grid-template-columns: minmax(var(--calendar-month-min-width), 1fr);
  gap: 0.9rem;
  align-items: start;
}

.leaderboard-calendar-month-panel {
  border: 1px solid rgb(214 202 186 / 0.5);
  border-radius: 0.65rem;
  background:
    linear-gradient(180deg, rgb(255 253 248 / 0.78), rgb(250 247 239 / 0.46)),
    rgb(255 253 248);
  padding: var(--calendar-month-padding);
}

.leaderboard-calendar-month-panel--current {
  border-color: rgb(196 111 80 / 0.28);
  background:
    linear-gradient(180deg, rgb(255 253 248 / 0.86), rgb(247 240 229 / 0.54)),
    rgb(255 253 248);
}

.leaderboard-calendar-month-toolbar {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  margin-bottom: 0.72rem;
}

.leaderboard-calendar-month {
  color: rgb(35 32 28);
  font-size: 1rem;
  font-weight: 900;
  line-height: 1.2;
  text-align: left;
  white-space: nowrap;
}

.leaderboard-calendar-days {
  display: grid;
  grid-template-columns: repeat(var(--calendar-day-columns), var(--calendar-day-size));
  grid-auto-rows: calc(var(--calendar-day-size) + var(--calendar-day-label-height) + var(--calendar-day-token-height) + 0.2rem);
  gap: 0.58rem var(--calendar-day-gap-x);
  justify-content: start;
}

.leaderboard-calendar-day {
  position: relative;
  display: grid;
  grid-template-rows: var(--calendar-day-label-height) var(--calendar-day-size) var(--calendar-day-token-height);
  align-content: start;
  width: var(--calendar-day-size);
  min-width: 0;
  height: calc(var(--calendar-day-size) + var(--calendar-day-label-height) + var(--calendar-day-token-height) + 0.2rem);
  justify-items: center;
  overflow: visible;
  border: 0;
  background: transparent;
  transition: opacity 0.16s ease, transform 0.16s ease;
}

.leaderboard-calendar-day--active {
  cursor: help;
}

.leaderboard-calendar-day--active:hover,
.leaderboard-calendar-day--active:focus-visible {
  transform: translateY(-1px);
}

.leaderboard-calendar-day-number {
  position: relative;
  z-index: 2;
  display: inline-flex;
  width: 100%;
  min-width: 0;
  height: var(--calendar-day-label-height);
  align-items: center;
  justify-content: center;
  border-radius: 0;
  background: transparent;
  color: rgb(116 61 45);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.78rem;
  font-weight: 900;
  line-height: 1;
  box-shadow: none;
}

.leaderboard-calendar-day--empty .leaderboard-calendar-day-number {
  color: rgb(132 118 98 / 0.48);
}

.leaderboard-calendar-avatar-frame {
  position: relative;
  display: block;
  width: 100%;
  height: var(--calendar-day-size);
  overflow: hidden;
  border: 1px solid rgb(214 202 186 / 0.58);
  border-radius: 0.42rem;
  background:
    linear-gradient(180deg, rgb(255 253 248 / 0.86), rgb(250 247 239 / 0.62)),
    rgb(255 253 248 / 0.7);
  transition: border-color 0.16s ease, box-shadow 0.16s ease;
}

.leaderboard-calendar-day--active .leaderboard-calendar-avatar-frame {
  border-color: rgb(196 111 80 / 0.42);
  background:
    linear-gradient(180deg, rgb(247 240 229 / 0.8), rgb(255 253 248 / 0.64));
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.68),
    0 0.42rem 0.9rem rgb(196 111 80 / 0.08);
}

.leaderboard-calendar-day--active:hover .leaderboard-calendar-avatar-frame,
.leaderboard-calendar-day--active:focus-visible .leaderboard-calendar-avatar-frame {
  border-color: rgb(154 83 62 / 0.48);
}

.leaderboard-calendar-day--empty .leaderboard-calendar-avatar-frame {
  border-color: rgb(214 202 186 / 0.18);
  background:
    linear-gradient(180deg, rgb(255 253 248 / 0.22), rgb(250 247 239 / 0.08)),
    rgb(255 253 248 / 0.14);
  box-shadow: none;
  opacity: 0.72;
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
  border-radius: 0.42rem 0.42rem 0.2rem 0.2rem;
  background:
    radial-gradient(circle at 50% 36%, rgb(255 255 255 / 0.2), transparent 42%),
    linear-gradient(135deg, rgb(214 183 157 / 0.98), rgb(154 83 62 / 0.86)),
    rgb(154 83 62);
  color: rgb(255 253 248);
  font-size: 1.45rem;
  font-weight: 900;
  line-height: 1;
  text-shadow: 0 1px 0 rgb(68 41 32 / 0.32);
  text-transform: uppercase;
}

.leaderboard-calendar-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.leaderboard-calendar-token-label {
  display: flex;
  width: 100%;
  height: var(--calendar-day-token-height);
  align-items: center;
  justify-content: center;
  overflow: hidden;
  color: rgb(84 70 56);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.62rem;
  font-weight: 900;
  letter-spacing: 0;
  line-height: 1;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.leaderboard-calendar-day--active .leaderboard-calendar-avatar::after {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgb(35 32 28 / 0.18), transparent 42%, rgb(35 32 28 / 0.08)),
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
  border: 1px solid rgb(214 183 157 / 0.28);
  border-radius: 0.5rem;
  background:
    linear-gradient(180deg, rgb(35 32 28 / 0.98), rgb(20 20 19 / 0.98)),
    rgb(20 20 19);
  box-shadow:
    0 1rem 2rem rgb(35 32 28 / 0.22),
    0 0 0 1px rgb(214 183 157 / 0.08),
    inset 0 1px 0 rgb(255 255 255 / 0.08);
  color: rgb(255 253 248);
  padding: 0.72rem 0.78rem 0.7rem;
  pointer-events: none;
  letter-spacing: 0;
}

.leaderboard-token-tooltip {
  width: 18rem;
  min-height: 7rem;
}

.leaderboard-token-tooltip-table {
  display: grid;
  gap: 0.14rem;
  border-top: 1px solid rgb(214 183 157 / 0.14);
  padding-top: 0.3rem;
  color: rgb(168 159 145);
  font-size: 0.72rem;
  font-weight: 800;
  line-height: 1.2;
}

.leaderboard-token-tooltip-row {
  display: grid;
  min-width: 0;
  grid-template-columns: 3.2rem minmax(0, 1fr) auto;
  align-items: baseline;
  gap: 0.42rem;
}

.leaderboard-token-tooltip-label,
.leaderboard-token-tooltip-value,
.leaderboard-token-tooltip-note {
  min-width: 0;
  overflow-wrap: anywhere;
}

.leaderboard-token-tooltip-label {
  color: rgb(168 159 145);
}

.leaderboard-token-tooltip-value {
  color: rgb(255 253 248);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-weight: 900;
}

.leaderboard-token-tooltip-note {
  color: rgb(214 183 157 / 0.78);
  font-size: 0.68rem;
  text-align: right;
  white-space: nowrap;
}

.leaderboard-calendar-tooltip-date {
  color: rgb(168 159 145);
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
  color: rgb(255 253 248);
  font-family: ui-serif, Georgia, Cambria, "Times New Roman", Times, serif;
  font-size: 0.92rem;
  font-weight: 900;
  line-height: 1.18;
}

.leaderboard-calendar-tooltip-tokens {
  color: rgb(214 183 157);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
  font-weight: 900;
  line-height: 1;
  white-space: nowrap;
}

.leaderboard-calendar-tooltip-meta {
  color: rgb(168 159 145);
  font-size: 0.74rem;
  font-weight: 800;
  line-height: 1.2;
}

:global(.dark .leaderboard-page) {
  background:
    linear-gradient(180deg, rgb(22 22 20), rgb(14 14 13)),
    rgb(14 14 13);
  color: rgb(243 239 231);
}

:global(.dark .leaderboard-section-label) {
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-state-card) {
  border-color: rgb(214 183 157 / 0.18);
  background:
    linear-gradient(180deg, rgb(35 32 28 / 0.82), rgb(20 20 19 / 0.7)),
    rgb(20 20 19 / 0.84);
  box-shadow: 0 0.85rem 2rem rgb(0 0 0 / 0.18);
}

:global(.dark .leaderboard-state-title) {
  color: rgb(243 239 231);
}

:global(.dark .leaderboard-loading-spinner) {
  border-color: rgb(214 183 157);
  border-top-color: transparent;
}

:global(.dark .leaderboard-token-card) {
  border-color: rgb(214 183 157 / 0.18);
  background:
    linear-gradient(135deg, rgb(35 32 28 / 0.94), rgb(20 20 19 / 0.96)),
    rgb(20 20 19);
  box-shadow: 0 0.85rem 2rem rgb(0 0 0 / 0.2);
}

:global(.dark .leaderboard-token-card::after) {
  background:
    linear-gradient(90deg, rgb(214 183 157 / 0.07) 1px, transparent 1px),
    linear-gradient(rgb(214 183 157 / 0.06) 1px, transparent 1px),
    linear-gradient(90deg, transparent, rgb(196 111 80 / 0.08), transparent);
  background-size: 5.5rem 100%, 100% 1rem, 100% 100%;
}

:global(.dark .leaderboard-ranking-switch) {
  border-color: rgb(214 183 157 / 0.18);
  background: rgb(14 14 13 / 0.86);
}

:global(.dark .leaderboard-ranking-switch-button--active) {
  background: rgb(243 239 231);
  color: rgb(20 20 19);
  box-shadow: none;
}

:global(.dark .leaderboard-ranking-switch-button--idle) {
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-ranking-switch-button--idle:hover) {
  color: rgb(243 239 231);
}

:global(.dark .leaderboard-token-legend-item) {
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-token-legend-dot--input) {
  background: linear-gradient(135deg, rgb(160 157 150), rgb(243 239 231));
}

:global(.dark .leaderboard-token-legend-dot--output) {
  background: linear-gradient(135deg, rgb(154 83 62), rgb(196 111 80));
}

:global(.dark .leaderboard-token-legend-dot--cache) {
  background: linear-gradient(135deg, rgb(72 118 105), rgb(95 143 129));
}

:global(.dark .leaderboard-token-summary-meta) {
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-token-trend-panel) {
  border-left-color: rgb(214 183 157 / 0.18);
}

:global(.dark .leaderboard-token-trend-header) {
  color: rgb(214 183 157);
}

:global(.dark .leaderboard-token-trend-legend) {
  color: rgb(214 183 157);
}

:global(.dark .leaderboard-token-trend-legend::before) {
  background: rgb(214 183 157 / 0.14);
}

:global(.dark .leaderboard-token-trend-empty) {
  border-color: rgb(214 183 157 / 0.16);
  background:
    linear-gradient(90deg, rgb(214 183 157 / 0.08) 1px, transparent 1px),
    linear-gradient(rgb(214 183 157 / 0.08) 1px, transparent 1px),
    rgb(20 20 19 / 0.42);
  background-size: 12.5% 100%, 100% 1.65rem;
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-token-odometer) {
  color: rgb(243 239 231);
}

:global(.dark .leaderboard-token-reel) {
  border-color: rgb(214 183 157 / 0.22);
  background:
    linear-gradient(180deg, rgb(35 32 28 / 0.96), rgb(20 20 19 / 0.94)),
    radial-gradient(circle at 50% 0%, rgb(196 111 80 / 0.18), transparent 58%);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.1),
    inset 0 -0.22rem 0 rgb(0 0 0 / 0.24),
    0 0.35rem 1rem rgb(196 111 80 / 0.1);
}

:global(.dark .leaderboard-token-reel::before) {
  background: linear-gradient(180deg, rgb(35 32 28 / 0.96), transparent);
}

:global(.dark .leaderboard-token-reel::after) {
  background: linear-gradient(0deg, rgb(0 0 0 / 0.28), transparent);
}

:global(.dark .leaderboard-token-cell) {
  text-shadow: 0 0 0.55rem rgb(214 183 157 / 0.18);
}

:global(.dark .leaderboard-token-separator) {
  color: rgb(214 183 157);
  text-shadow: 0 0 0.55rem rgb(214 183 157 / 0.18);
}

:global(.dark .leaderboard-token-ranking-card) {
  border-color: rgb(214 183 157 / 0.18);
  background:
    linear-gradient(180deg, rgb(35 32 28 / 0.82), rgb(20 20 19 / 0.7)),
    rgb(20 20 19 / 0.84);
  box-shadow: 0 0.85rem 2rem rgb(0 0 0 / 0.18);
}

:global(.dark .leaderboard-token-rank-row-current) {
  background: rgb(95 143 129 / 0.14);
}

:global(.dark .leaderboard-token-title-badge),
:global(.dark .leaderboard-token-title-more) {
  background: rgb(14 14 13 / 0.36);
}

:global(.dark .leaderboard-token-title-more) {
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-side-card) {
  border-color: rgb(214 183 157 / 0.13);
  background:
    linear-gradient(180deg, rgb(35 32 28 / 0.58), rgb(20 20 19 / 0.42)),
    rgb(20 20 19 / 0.56);
  box-shadow: 0 0.35rem 1rem rgb(0 0 0 / 0.14);
}

:global(.dark .leaderboard-record-card) {
  border-color: rgb(214 183 157 / 0.14);
  background:
    linear-gradient(135deg, rgb(35 32 28 / 0.74), rgb(20 20 19 / 0.56)),
    rgb(20 20 19 / 0.72);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.04),
    0 0.45rem 1rem rgb(0 0 0 / 0.12);
}

:global(.dark .leaderboard-record-kicker) {
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-record-headline) {
  color: rgb(243 239 231);
}

:global(.dark .leaderboard-record-progress) {
  color: rgb(194 187 176);
}

:global(.dark .leaderboard-record-progress strong) {
  color: rgb(214 183 157);
}

:global(.dark .leaderboard-thursday-banner) {
  border-color: rgb(214 183 157 / 0.18);
  background: rgb(35 32 28);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.05),
    0 0.45rem 1rem rgb(0 0 0 / 0.16);
}

:global(.dark .leaderboard-thursday-banner::after) {
  background: linear-gradient(90deg, rgb(20 20 19 / 0.76), rgb(20 20 19 / 0.34) 42%, transparent 76%);
}

:global(.dark .leaderboard-side-label) {
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-side-value) {
  color: rgb(243 239 231);
}

:global(.dark .leaderboard-reward-progress-card),
:global(.dark .leaderboard-reward-mode-card) {
  border-color: rgb(214 183 157 / 0.12);
  background: rgb(14 14 13 / 0.24);
}

:global(.dark .leaderboard-reward-panel) {
  border-top-color: rgb(214 183 157 / 0.14);
}

:global(.dark .leaderboard-reward-progress-track) {
  background: rgb(64 58 50 / 0.78);
}

:global(.dark .leaderboard-reward-title) {
  color: rgb(243 239 231);
}

:global(.dark .leaderboard-reward-status--claimed) {
  background: rgb(95 143 129 / 0.16);
  color: rgb(160 191 181);
}

:global(.dark .leaderboard-reward-status--ready) {
  background: rgb(196 111 80 / 0.16);
  color: rgb(230 166 141);
}

:global(.dark .leaderboard-reward-status--idle) {
  background: rgb(118 111 101 / 0.18);
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-weekly-winners) {
  border-color: rgb(214 183 157 / 0.16);
  background: rgb(14 14 13 / 0.34);
}

:global(.dark .leaderboard-weekly-winners-header),
:global(.dark .leaderboard-weekly-winner-rank) {
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-weekly-winner-name) {
  color: rgb(243 239 231);
}

:global(.dark .leaderboard-weekly-winner-meta),
:global(.dark .leaderboard-weekly-winner-tokens) {
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-token-rank-avatar) {
  border-color: rgb(20 20 19 / 0.92);
  background:
    linear-gradient(135deg, rgb(35 32 28 / 0.96), rgb(20 20 19 / 0.9)),
    var(--token-bar-color);
  box-shadow:
    0 0 0 1px rgb(214 183 157 / 0.18),
    0 0.32rem 0.8rem var(--token-bar-glow);
}

:global(.dark .leaderboard-token-bar-track) {
  background: transparent;
}

:global(.dark .leaderboard-token-bar-track::before) {
  background: rgb(214 183 157 / 0.045);
}

:global(.dark .leaderboard-token-bar-track::after) {
  background:
    linear-gradient(90deg, rgb(214 183 157 / 0.022) 1px, transparent 1px),
    transparent;
  background-size: 10% 100%;
}

:global(.dark .leaderboard-token-bar-fill) {
  background: rgb(64 58 50 / 0.52);
}

:global(.dark .leaderboard-token-bar-segment-input) {
  background: linear-gradient(90deg, rgb(160 157 150), rgb(243 239 231));
}

:global(.dark .leaderboard-token-bar-segment-output) {
  background: linear-gradient(90deg, rgb(154 83 62), rgb(196 111 80));
}

:global(.dark .leaderboard-token-bar-segment-cache) {
  background: linear-gradient(90deg, rgb(72 118 105), rgb(95 143 129));
}

:global(.dark .leaderboard-calendar-card) {
  border-color: rgb(214 183 157 / 0.16);
  background:
    linear-gradient(180deg, rgb(35 32 28 / 0.82), rgb(20 20 19 / 0.7)),
    rgb(20 20 19 / 0.84);
  box-shadow: 0 0.85rem 2rem rgb(0 0 0 / 0.18);
}

:global(.dark .leaderboard-calendar-title),
:global(.dark .leaderboard-calendar-month) {
  color: rgb(243 239 231);
}

:global(.dark .leaderboard-calendar-meta) {
  color: rgb(168 159 145);
}

:global(.dark .leaderboard-calendar-scroll::-webkit-scrollbar-thumb) {
  background: rgb(214 183 157 / 0.24);
}

:global(.dark .leaderboard-calendar-month-panel) {
  border-color: rgb(214 183 157 / 0.12);
  background:
    linear-gradient(180deg, rgb(35 32 28 / 0.5), rgb(20 20 19 / 0.34)),
    rgb(20 20 19 / 0.42);
}

:global(.dark .leaderboard-calendar-month-panel--current) {
  border-color: rgb(214 183 157 / 0.18);
  background:
    linear-gradient(180deg, rgb(35 32 28 / 0.62), rgb(20 20 19 / 0.42)),
    rgb(20 20 19 / 0.5);
}

:global(.dark .leaderboard-calendar-avatar-frame) {
  border-color: rgb(214 183 157 / 0.16);
  background:
    linear-gradient(180deg, rgb(35 32 28 / 0.58), rgb(20 20 19 / 0.4)),
    rgb(20 20 19 / 0.34);
}

:global(.dark .leaderboard-calendar-day--active .leaderboard-calendar-avatar-frame) {
  border-color: rgb(214 183 157 / 0.3);
  background:
    linear-gradient(180deg, rgb(64 58 50 / 0.62), rgb(20 20 19 / 0.54));
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.08),
    0 0.45rem 1rem rgb(0 0 0 / 0.18);
}

:global(.dark .leaderboard-calendar-day--empty .leaderboard-calendar-avatar-frame) {
  border-color: rgb(214 183 157 / 0.05);
  background:
    linear-gradient(180deg, rgb(35 32 28 / 0.24), rgb(20 20 19 / 0.12)),
    rgb(20 20 19 / 0.14);
}

:global(.dark .leaderboard-calendar-day-number) {
  background: transparent;
  color: rgb(214 183 157);
  box-shadow: none;
}

:global(.dark .leaderboard-calendar-day--empty .leaderboard-calendar-day-number) {
  color: rgb(168 159 145 / 0.6);
}

:global(.dark .leaderboard-calendar-avatar) {
  background:
    radial-gradient(circle at 50% 36%, rgb(255 255 255 / 0.16), transparent 42%),
    linear-gradient(135deg, rgb(64 58 50 / 0.96), rgb(154 83 62 / 0.72)),
    rgb(154 83 62);
  color: rgb(243 239 231);
}

:global(.dark .leaderboard-calendar-token-label) {
  color: rgb(214 183 157);
}

:global(.dark .leaderboard-calendar-tooltip) {
  border-color: rgb(214 183 157 / 0.2);
  background:
    linear-gradient(180deg, rgb(17 17 17 / 0.98), rgb(8 8 8 / 0.98)),
    rgb(8 8 8);
  box-shadow:
    0 1rem 2rem rgb(0 0 0 / 0.34),
    0 0 0 1px rgb(214 183 157 / 0.08),
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

  .leaderboard-ranking-toolbar-meta {
    width: 100%;
    justify-content: space-between;
    gap: 0.65rem;
  }

  .leaderboard-token-legend {
    justify-content: flex-start;
  }

  .leaderboard-token-rank-row {
    grid-template-columns: 1fr;
    gap: 0.45rem;
    min-height: 3.18rem;
    padding: 0.18rem 0;
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
    border-top: 1px solid rgb(214 202 186 / 0.72);
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
    flex-wrap: wrap;
    justify-content: flex-start;
    gap: 0.45rem 0.75rem;
    padding-bottom: 0.65rem;
  }

  .leaderboard-calendar-meta {
    font-size: 0.74rem;
  }

  .leaderboard-calendar-months {
    --calendar-day-size: 3.42rem;
    --calendar-day-label-height: 0.92rem;
    --calendar-day-token-height: 0.8rem;
    --calendar-day-gap-x: 0.34rem;
    --calendar-month-padding: 0.72rem;
    gap: 0.85rem;
  }

  .leaderboard-calendar-days {
    gap: 0.48rem var(--calendar-day-gap-x);
  }

  .leaderboard-calendar-avatar {
    font-size: 1.12rem;
  }

  .leaderboard-calendar-token-label {
    font-size: 0.54rem;
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
  background: rgb(250 247 239);
  padding: 0 0.25rem;
  color: rgb(109 103 93);
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1;
  box-shadow: inset 0 0 0 1px rgb(214 202 186);
}

.leaderboard-badge-week {
  background: rgb(247 240 229);
  color: rgb(116 61 45);
}

.leaderboard-badge-month {
  background: rgb(240 233 221);
  color: rgb(86 76 63);
}

.leaderboard-badge-total {
  background: rgb(242 235 222);
  color: rgb(68 41 32);
}

.leaderboard-badge-night {
  background: rgb(238 234 226);
  color: rgb(74 68 59);
}

.leaderboard-badge-burst {
  background: rgb(248 231 224);
  color: rgb(154 83 62);
}

.leaderboard-badge-checkin {
  background: rgb(232 241 236);
  color: rgb(57 98 88);
}

.leaderboard-badge-save {
  background: rgb(232 241 236);
  color: rgb(57 98 88);
}

.leaderboard-badge-fire {
  background: rgb(248 231 224);
  color: rgb(116 61 45);
}

:global(.dark .leaderboard-badge-overflow) {
  background: rgb(20 20 19);
  color: rgb(168 159 145);
  box-shadow: inset 0 0 0 1px rgb(214 183 157 / 0.18);
}

:global(.dark .leaderboard-badge-week) {
  background: rgb(196 111 80 / 0.16);
  color: rgb(214 183 157);
}

:global(.dark .leaderboard-badge-month) {
  background: rgb(214 183 157 / 0.12);
  color: rgb(214 183 157);
}

:global(.dark .leaderboard-badge-total) {
  background: rgb(243 239 231 / 0.12);
  color: rgb(243 239 231);
}

:global(.dark .leaderboard-badge-night) {
  background: rgb(118 111 101 / 0.18);
  color: rgb(214 183 157);
}

:global(.dark .leaderboard-badge-burst) {
  background: rgb(196 111 80 / 0.16);
  color: rgb(230 166 141);
}

:global(.dark .leaderboard-badge-checkin) {
  background: rgb(95 143 129 / 0.16);
  color: rgb(160 191 181);
}

:global(.dark .leaderboard-badge-save) {
  background: rgb(95 143 129 / 0.16);
  color: rgb(160 191 181);
}

:global(.dark .leaderboard-badge-fire) {
  background: rgb(196 111 80 / 0.16);
  color: rgb(230 166 141);
}
</style>
