<template>
  <section class="dashboard-stats-grid">
    <article
      v-for="card in statCards"
      :key="card.key"
      class="dashboard-stat-card"
    >
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <p class="dashboard-stat-label">{{ card.label }}</p>
          <p class="dashboard-stat-value">{{ card.value }}</p>
        </div>
        <span class="dashboard-stat-badge" :class="card.badgeClass">
          {{ card.badge }}
        </span>
      </div>
      <p class="dashboard-stat-trend" :class="card.trendClass">
        {{ card.trend }}
        <Icon :name="card.trendIcon" size="xs" />
      </p>
      <p class="dashboard-stat-note">{{ card.note }}</p>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { formatCreditAmount } from '@/utils/credits'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

type TrendIcon = 'trendingUp' | 'arrowDown' | 'checkCircle'

const props = defineProps<{
  stats: UserStatsType
}>()

const { t } = useI18n()

const formatCompact = (value: number): string =>
  new Intl.NumberFormat(undefined, {
    notation: 'compact',
    maximumFractionDigits: value >= 1000 ? 1 : 0,
  }).format(Number(value) || 0)

const formatCost = (value: number): string =>
  formatCreditAmount(value, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 4,
  })

const formatPercent = (value: number): string =>
  `${new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 1,
  }).format(Number(value) || 0)}%`

const averageCost = computed(() => {
  const requests = props.stats.total_requests || 0
  if (requests <= 0) return 0
  return (props.stats.total_actual_cost || 0) / requests
})

const successTrend = computed(() => t('dashboard.overview.trendStable'))
const activeApiKeyNote = computed(() =>
  t('dashboard.overview.activeApiKeys', {
    count: props.stats.active_api_keys || 0,
    total: props.stats.total_api_keys || 0,
  })
)

const statCards = computed<Array<{
  key: string
  label: string
  value: string
  badge: string
  badgeClass: string
  trend: string
  trendClass: string
  trendIcon: TrendIcon
  note: string
}>>(() => [
  {
    key: 'total-cost',
    label: t('dashboard.overview.totalCost'),
    value: formatCost(props.stats.total_actual_cost || 0),
    badge: formatCost(props.stats.today_actual_cost || 0),
    badgeClass: 'dashboard-stat-badge--neutral',
    trend: successTrend.value,
    trendClass: 'dashboard-stat-trend--positive',
    trendIcon: 'trendingUp',
    note: t('dashboard.overview.recentConsumption'),
  },
  {
    key: 'total-tokens',
    label: t('dashboard.overview.totalTokens'),
    value: formatCompact(props.stats.total_tokens || 0),
    badge: formatCompact(props.stats.today_tokens || 0),
    badgeClass: 'dashboard-stat-badge--neutral',
    trend: successTrend.value,
    trendClass: 'dashboard-stat-trend--positive',
    trendIcon: 'trendingUp',
    note: activeApiKeyNote.value,
  },
  {
    key: 'total-requests',
    label: t('dashboard.overview.totalRequests'),
    value: formatCompact(props.stats.total_requests || 0),
    badge: formatCompact(props.stats.today_requests || 0),
    badgeClass: 'dashboard-stat-badge--neutral',
    trend: successTrend.value,
    trendClass: 'dashboard-stat-trend--positive',
    trendIcon: 'trendingUp',
    note: t('dashboard.overview.userRetentionNote'),
  },
  {
    key: 'failure-rate',
    label: t('dashboard.overview.failureRate'),
    value: formatPercent(0),
    badge: formatPercent(0),
    badgeClass: 'dashboard-stat-badge--neutral',
    trend: t('dashboard.overview.performanceStable'),
    trendClass: 'dashboard-stat-trend--positive',
    trendIcon: 'checkCircle',
    note: t('dashboard.overview.failureRateNote'),
  },
  {
    key: 'avg-cost',
    label: t('dashboard.overview.averageRequestCost'),
    value: formatCost(averageCost.value),
    badge: formatCost(props.stats.today_actual_cost || 0),
    badgeClass: 'dashboard-stat-badge--neutral',
    trend: successTrend.value,
    trendClass: 'dashboard-stat-trend--positive',
    trendIcon: 'trendingUp',
    note: t('dashboard.overview.averageRequestCostNote'),
  },
])
</script>
