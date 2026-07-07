<template>
  <section class="dashboard-performance-grid">
    <article
      v-for="card in performanceCards"
      :key="card.key"
      class="dashboard-performance-card"
    >
      <span class="dashboard-performance-icon" :class="card.iconClass">
        <Icon :name="card.icon" size="sm" />
      </span>
      <div class="min-w-0">
        <p class="dashboard-performance-label">{{ card.label }}</p>
        <div v-if="card.kind === 'rate'" class="dashboard-performance-metrics">
          <p
            v-for="metric in card.metrics"
            :key="metric.unit"
            class="dashboard-performance-metric"
          >
            <span class="dashboard-performance-metric-value">{{ metric.value }}</span>
            <span class="dashboard-performance-metric-unit">{{ metric.unit }}</span>
          </p>
        </div>
        <template v-else>
          <p class="dashboard-performance-value">{{ card.value }}</p>
          <p class="dashboard-performance-note">{{ card.note }}</p>
        </template>
      </div>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

type PerformanceIcon = 'bolt' | 'sparkles' | 'shield' | 'clock'

interface PerformanceCardBase {
  key: string
  label: string
  icon: PerformanceIcon
  iconClass: string
}

type PerformanceCard = PerformanceCardBase & (
  | {
    kind: 'rate'
    metrics: Array<{
      value: string
      unit: string
    }>
  }
  | {
    kind: 'single'
    value: string
    note: string
  }
)

const props = defineProps<{
  stats: UserStatsType
}>()

const { t } = useI18n()

const formatNumber = (value: number): string =>
  new Intl.NumberFormat(undefined).format(Number(value) || 0)

const formatCacheReuseRate = (cacheReadTokens: number, inputTokens: number): string => {
  const inputTotal = cacheReadTokens + inputTokens
  if (inputTotal <= 0) return t('common.notAvailable')
  return `${new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1,
  }).format((cacheReadTokens / inputTotal) * 100)}%`
}

const formatDuration = (milliseconds: number): string => {
  const value = Number(milliseconds) || 0
  if (value >= 1000) {
    return `${(value / 1000).toFixed(2)}s`
  }
  return `${Math.round(value)}ms`
}

const performanceCards = computed<PerformanceCard[]>(() => [
  {
    key: 'realtime-performance',
    kind: 'rate',
    label: t('dashboard.performance'),
    metrics: [
      { value: formatNumber(props.stats.rpm || 0), unit: 'RPM' },
      { value: formatNumber(props.stats.tpm || 0), unit: 'TPM' },
    ],
    icon: 'bolt',
    iconClass: '!border-[#d8cec2] !bg-[#fffaf5] !text-[#a9583e] dark:!border-[#a9583e]/45 dark:!bg-[#a9583e]/20 dark:!text-[#cc785c]',
  },
  {
    key: 'today-cache-reuse',
    kind: 'single',
    label: t('dashboard.todayCacheHitRate'),
    value: formatCacheReuseRate(props.stats.today_cache_read_tokens || 0, props.stats.today_input_tokens || 0),
    note: `${t('dashboard.cacheReadTokens')}: ${formatNumber(props.stats.today_cache_read_tokens || 0)}`,
    icon: 'sparkles',
    iconClass: '!border-[#d8cec2] !bg-[#f5f0e8] !text-[#cc785c] dark:!border-[#cc785c]/45 dark:!bg-[#cc785c]/20 dark:!text-[#cc785c]',
  },
  {
    key: 'total-cache-reuse',
    kind: 'single',
    label: t('dashboard.totalCacheHitRate'),
    value: formatCacheReuseRate(props.stats.total_cache_read_tokens || 0, props.stats.total_input_tokens || 0),
    note: `${t('dashboard.cacheReadTokens')}: ${formatNumber(props.stats.total_cache_read_tokens || 0)}`,
    icon: 'shield',
    iconClass: '!border-[#d8cec2] !bg-[#fffaf5] !text-[#5f7f68] dark:!border-[#5f7f68]/45 dark:!bg-[#5f7f68]/20 dark:!text-[#9ab3a0]',
  },
  {
    key: 'avg-response',
    kind: 'single',
    label: t('dashboard.avgResponse'),
    value: formatDuration(props.stats.average_duration_ms || 0),
    note: t('dashboard.averageTime'),
    icon: 'clock',
    iconClass: 'dashboard-performance-icon--rose',
  },
])
</script>
