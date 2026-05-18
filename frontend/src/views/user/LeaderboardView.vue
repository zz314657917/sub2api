<template>
  <AppLayout>
    <div class="space-y-6">
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

      <section v-if="leaderboard" class="card leaderboard-token-card leaderboard-token-summary p-5">
        <div class="leaderboard-token-summary-inner relative z-10">
          <div class="mx-auto min-w-0">
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
          </div>
          <div class="leaderboard-token-summary-meta text-sm text-gray-500 dark:text-gray-400">
            <span>{{ t('leaderboard.generatedAt') }} {{ formatTime(leaderboard.generated_at) }}</span>
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

      <div v-else-if="leaderboard" class="grid grid-cols-1 gap-5 xl:grid-cols-[minmax(0,1fr)_22rem]">
        <div class="min-w-0 space-y-3">
          <div v-if="rankingItems.length === 0" class="card p-8">
            <EmptyState :title="t('leaderboard.emptyTitle')" :description="t('leaderboard.emptyDescription')" />
          </div>

          <template v-else>
            <section class="leaderboard-token-ranking-card" data-testid="leaderboard-token-ranking">
              <div class="leaderboard-token-ranking-header">
                <div class="min-w-0">
                  <h2 class="leaderboard-token-ranking-title text-base font-semibold text-gray-900 dark:text-white">
                    <span>{{ t('leaderboard.tokenRankingTitle', { count: rankingItems.length }) }}</span>
                    <span class="leaderboard-token-ranking-period">{{ currentPeriodLabel }}</span>
                  </h2>
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
                  <div class="leaderboard-token-rank-user">
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
                    <div class="leaderboard-token-bar-track" :aria-label="`${getLeaderboardDisplayName(item)} ${formatNumber(item.tokens)} Token`">
                      <div class="leaderboard-token-bar-fill"></div>
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
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usageAPI } from '@/api'
import type { LeaderboardBadge, LeaderboardDailyRewards, LeaderboardPeriod, UserLeaderboardItem, UserLeaderboardResponse } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { formatCurrency, formatDateTime, formatNumber, formatTime } from '@/utils/format'

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
let loadSeq = 0
let visualTokenTickerID: number | null = null

type RollingTokenPart = {
  type: 'digit' | 'separator'
  value: string
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

const periodOptions = computed(() => [
  { value: 'day' as const, label: t('leaderboard.period.day') },
  { value: 'week' as const, label: t('leaderboard.period.week') },
  { value: 'month' as const, label: t('leaderboard.period.month') },
  { value: 'all' as const, label: t('leaderboard.period.all') },
])
const currentPeriodLabel = computed(() => periodOptions.value.find((option) => option.value === period.value)?.label ?? '')

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

const rewardTiers = computed(() => {
  const tiers = new Map((dailyRewards.value?.rewards ?? []).map((tier) => [tier.rank, tier.amount]))
  return [1, 2, 3].map((rank) => ({ rank, amount: tiers.get(rank) ?? 0 }))
})

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
  loadLeaderboard()
}

function formatRollingTokenNumber(value: number): string {
  if (!Number.isFinite(value)) return '0'
  return Math.max(0, Math.floor(value)).toLocaleString('en-US')
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

function tokenBarStyle(item: UserLeaderboardItem): Record<string, string> {
  const palette = tokenBarPalette(item.rank)
  const widthText = tokenBarWidth(item)

  return {
    '--token-bar-width': widthText,
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
})

onUnmounted(() => {
  stopVisualTokenTicker()
})
</script>

<style scoped>
.leaderboard-token-summary {
  min-height: 9.6rem;
}

.leaderboard-token-summary-inner {
  display: grid;
  min-height: 7.35rem;
  align-content: center;
  justify-items: center;
  gap: 0.72rem;
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

.leaderboard-token-ranking-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgb(148 163 184 / 0.18);
  padding-bottom: 0.8rem;
}

.leaderboard-token-ranking-title {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.45rem;
  letter-spacing: 0;
}

.leaderboard-token-ranking-period {
  display: inline-flex;
  align-items: center;
  border-left: 1px solid rgb(148 163 184 / 0.45);
  padding-left: 0.45rem;
  color: rgb(100 116 139);
  font-size: 0.8125rem;
  font-weight: 700;
}

.leaderboard-token-ranking-updated {
  flex: 0 0 auto;
  color: rgb(100 116 139);
  font-size: 0.8125rem;
  white-space: nowrap;
}

.leaderboard-token-rank-list {
  display: grid;
  gap: 0.82rem;
  padding-top: 1rem;
}

.leaderboard-token-rank-row {
  display: grid;
  grid-template-columns: minmax(11.5rem, 13.5rem) minmax(14rem, 1fr);
  align-items: center;
  gap: 1rem;
  min-height: 2.22rem;
  padding: 0.08rem 0;
}

.leaderboard-token-rank-row-current {
  border-radius: 0.45rem;
  background: rgb(34 197 94 / 0.07);
}

.leaderboard-token-rank-user {
  display: grid;
  min-width: 0;
  grid-template-columns: 2.5rem 1.45rem minmax(0, 1fr);
  align-items: center;
  gap: 0.5rem;
}

.leaderboard-token-rank-main {
  display: contents;
}

.leaderboard-token-rank-name {
  max-width: 100%;
  min-width: 0;
  overflow: hidden;
  color: var(--token-rank-color);
  font-size: 0.875rem;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  width: var(--token-bar-width);
  height: 100%;
  min-width: 0.7rem;
  border-radius: 0.08rem 0.28rem 0.28rem 0.08rem;
  background: var(--token-bar-color);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.26),
    0 0.35rem 1.1rem var(--token-bar-glow);
  transition: width 520ms cubic-bezier(0.22, 0.72, 0.2, 1);
}

.leaderboard-token-rank-index {
  min-width: 0;
  color: rgb(100 116 139);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
  font-weight: 800;
  text-align: right;
}

.leaderboard-token-rank-avatar {
  display: inline-flex;
  flex: 0 0 auto;
  width: 1.45rem;
  height: 1.45rem;
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
  font-size: 0.7rem;
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

:global(.dark .leaderboard-token-summary-meta) {
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
  .leaderboard-token-ranking-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.35rem;
  }

  .leaderboard-token-rank-row {
    grid-template-columns: 1fr;
    gap: 0.35rem;
    padding: 0.25rem 0;
  }

  .leaderboard-token-rank-user {
    grid-template-columns: 2.35rem 1.28rem minmax(0, 1fr);
  }

  .leaderboard-token-bar-track {
    height: 1.35rem;
  }

  .leaderboard-token-bar-value {
    font-size: 0.8rem;
  }

  .leaderboard-token-rank-avatar {
    width: 1.28rem;
    height: 1.28rem;
    border-width: 1px;
    font-size: 0.62rem;
  }

  .leaderboard-token-odometer {
    font-size: clamp(1.72rem, 10.6vw, 2.45rem);
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
