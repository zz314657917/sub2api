<template>
  <section class="dashboard-charts-stack">
    <article class="dashboard-panel">
      <div class="dashboard-panel-header">
        <div>
          <p class="dashboard-eyebrow">{{ t('dashboard.console.usageAnalysis') }}</p>
          <h2 class="dashboard-panel-title">{{ t('dashboard.overview.dailyCostTrend') }}</h2>
          <p class="dashboard-panel-subtitle">{{ startDate }} — {{ endDate }} · {{ timezone }}</p>
        </div>
        <div class="dashboard-range-tabs" role="group" :aria-label="t('dashboard.timeRange')">
          <button
            v-for="days in [7, 30, 90]"
            :key="days"
            type="button"
            class="dashboard-range-tab"
            :aria-pressed="activeDays === days"
            :class="{ 'dashboard-range-tab--active': activeDays === days }"
            @click="applyRange(days)"
          >
            {{ t(`dashboard.overview.last${days}Days`) }}
          </button>
        </div>
      </div>
      <div v-if="loading" class="dashboard-chart-state" role="status">
        <LoadingSpinner size="md" />
      </div>
      <div v-else-if="error" class="dashboard-chart-state" role="alert">
        <p>{{ t('dashboard.console.loadError') }}</p>
        <button type="button" class="btn btn-secondary" @click="emit('retry')">
          {{ t('dashboard.console.retry') }}
        </button>
      </div>
      <template v-else>
        <div class="dashboard-cost-layout">
          <div class="min-w-0">
            <div class="dashboard-chart-total">
              <span>{{ t('dashboard.console.intervalActual') }}</span
              ><strong>{{ formatCost(totals.actual) }}</strong
              ><small>{{ t('dashboard.standard') }} {{ formatCost(totals.standard) }}</small>
            </div>
            <div v-if="hasRequests" class="dashboard-main-chart">
              <Bar
                :data="costChart"
                :options="barOptions"
                role="img"
                :aria-label="t('dashboard.overview.dailyCostTrend')"
              />
            </div>
            <div v-else class="dashboard-chart-state">
              <Icon name="chart" size="lg" />
              <p>
                {{
                  t(
                    firstUse
                      ? 'dashboard.console.firstUseTitle'
                      : 'dashboard.console.rangeEmptyTitle'
                  )
                }}
              </p>
              <span>{{
                t(
                  firstUse
                    ? 'dashboard.console.firstUseDescription'
                    : 'dashboard.console.rangeEmptyDescription'
                )
              }}</span
              ><router-link
                :to="firstUse ? '/tutorial/getting-started' : '/usage'"
                class="dashboard-inline-link"
                >{{ t(firstUse ? 'dashboard.console.readGuide' : 'dashboard.viewAllUsage') }}
                <Icon name="arrowRight" size="xs"
              /></router-link>
            </div>
            <p v-if="hasRequests && totals.actual === 0" class="dashboard-panel-subtitle">
              {{ t('dashboard.console.zeroCost') }}
            </p>
          </div>
          <div class="dashboard-range-summary">
            <div>
              <span>{{ t('dashboard.console.cacheRate') }}</span
              ><strong :title="t('dashboard.console.cacheRateFormula')">{{ cacheRate }}</strong
              ><small>{{ t('dashboard.console.rangeCacheRate') }}</small>
            </div>
            <div>
              <span>{{ t('dashboard.tokens') }}</span
              ><strong>{{ compact(totals.tokens) }}</strong
              ><small>{{ t('dashboard.console.reportedTokens') }}</small>
            </div>
            <router-link to="/usage" class="dashboard-inline-link"
              >{{ t('dashboard.viewUsage') }} <Icon name="arrowRight" size="xs"
            /></router-link>
          </div>
        </div>
        <div class="dashboard-token-analysis">
          <div class="dashboard-panel-header">
            <h3 class="dashboard-panel-title">{{ t('dashboard.console.tokenAnalysis') }}</h3>
            <span class="dashboard-panel-subtitle">{{
              t('dashboard.console.tokenDescription')
            }}</span>
          </div>
          <div class="dashboard-token-grid">
            <div v-for="item in tokenBreakdown" :key="item.key" :data-category="item.costKey">
              <span class="dashboard-token-dot" :style="{ background: item.color }" /><span>{{
                item.label
              }}</span
              ><strong v-if="item.value !== null">{{ compact(item.value) }} Token</strong>
              <div class="dashboard-token-spend">
                <template v-if="spendBreakdown">
                  <strong>{{ formatCost(item.spend) }}</strong>
                  <span>{{ formatShare(item.share) }}</span>
                </template>
                <span v-else>{{ t('dashboard.console.spendUnavailable') }}</span>
              </div>
              <div class="dashboard-token-track">
                <i
                  :style="{
                    width: `${item.share}%`,
                    background: item.color
                  }"
                />
              </div>
            </div>
          </div>
          <p class="dashboard-panel-subtitle dashboard-spend-explanation">
            {{ t('dashboard.console.spendAllocation') }}
          </p>
        </div>
      </template>
    </article>
  </section>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  LinearScale,
  Tooltip,
  type TooltipItem
} from 'chart.js'
import { Bar } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatCreditAmount } from '@/utils/credits'
import { calendarDayCount, calendarRange } from './dashboardRange'
import type { TrendDataPoint } from '@/types'
ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip)
const props = defineProps<{
  loading: boolean
  error?: boolean
  firstUse?: boolean
  startDate: string
  endDate: string
  granularity: string
  trend: TrendDataPoint[]
}>()
const emit = defineEmits<{
  'update:startDate': [value: string]
  'update:endDate': [value: string]
  'update:granularity': [value: string]
  dateRangeChange: [range: { startDate: string; endDate: string; preset: '7d' | '30d' | '90d' }]
  retry: []
}>()
const { t } = useI18n()
const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
const activeDays = computed(() => calendarDayCount(props.startDate, props.endDate))
function applyRange(days: number) {
  const range = calendarRange(days)
  emit('update:startDate', range.startDate)
  emit('update:endDate', range.endDate)
  emit('update:granularity', 'day')
  emit('dateRangeChange', { ...range, preset: `${days}d` as '7d' | '30d' | '90d' })
}
const totals = computed(() =>
  props.trend.reduce(
    (a, d) => ({
      actual: a.actual + Number(d.actual_cost ?? 0),
      standard: a.standard + Number(d.cost ?? 0),
      requests: a.requests + Number(d.requests ?? 0),
      tokens: a.tokens + Number(d.total_tokens ?? 0),
      input: a.input + Number(d.input_tokens ?? 0),
      cacheRead: a.cacheRead + Number(d.cache_read_tokens ?? 0)
    }),
    { actual: 0, standard: 0, requests: 0, tokens: 0, input: 0, cacheRead: 0 }
  )
)
const cacheRate = computed(() => {
  const inputTotal = totals.value.input + totals.value.cacheRead
  if (inputTotal <= 0) return '—'
  return `${new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1
  }).format((totals.value.cacheRead / inputTotal) * 100)}%`
})
const hasRequests = computed(() => totals.value.requests > 0)
const compact = (v: number) =>
  new Intl.NumberFormat(undefined, { notation: 'compact', maximumFractionDigits: 1 }).format(v)
const formatCost = (v: number) =>
  formatCreditAmount(v, { minimumFractionDigits: 2, maximumFractionDigits: 4 })
const spendKeys = ['input', 'output', 'cache_read', 'cache_creation', 'other'] as const
const spendBreakdown = computed(() => {
  const sums = { input: 0, output: 0, cache_read: 0, cache_creation: 0, other: 0 }
  for (const point of props.trend) {
    const breakdown = point.actual_cost_breakdown
    if (!breakdown || !Number.isFinite(point.actual_cost) || point.actual_cost < 0) return null
    let sum = 0
    for (const key of spendKeys) {
      const value = breakdown[key]
      if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) return null
      sum += value
      sums[key] += value
    }
    if (Math.abs(sum - point.actual_cost) > Math.max(1e-8, Math.abs(point.actual_cost) * 1e-9)) return null
  }
  const total = totals.value.actual
  const sum = spendKeys.reduce((value, key) => value + sums[key], 0)
  if (!Number.isFinite(total) || !Number.isFinite(sum) ||
      spendKeys.some((key) => !Number.isFinite(sums[key])) ||
      Math.abs(sum - total) > Math.max(1e-8, Math.abs(total) * 1e-9)) return null
  return sums
})
const formatShare = (share: number) => share > 0 && share < 0.1
  ? '<0.1%'
  : `${new Intl.NumberFormat(undefined, { maximumFractionDigits: 1 }).format(share)}%`
const tokenBreakdown = computed(() => {
  const categories = [
    { key: 'input_tokens', costKey: 'input' as const, label: t('dashboard.input'), color: '#ad684f' },
    { key: 'output_tokens', costKey: 'output' as const, label: t('dashboard.output'), color: '#648374' },
    { key: 'cache_read_tokens', costKey: 'cache_read' as const, label: t('dashboard.cacheReadTokens'), color: '#8b7ca0' },
    { key: 'cache_creation_tokens', costKey: 'cache_creation' as const, label: t('dashboard.console.cacheWrite'), color: '#b18b4f' }
  ]
  const items = categories.map((item) => ({
    ...item,
    value: props.trend.reduce((sum, d) => sum + Number(d[item.key as keyof TrendDataPoint] ?? 0), 0),
    spend: spendBreakdown.value?.[item.costKey] ?? 0
  }))
  const other = spendBreakdown.value?.other ?? 0
  const all: { key: string; costKey: string; label: string; color: string; value: number | null; spend: number }[] = items
  if (other > 0) all.push({ key: 'other', costKey: 'other', label: t('dashboard.console.otherSpend'), color: '#7c858e', value: null, spend: other })
  return all.map((item) => ({
    ...item,
    share: totals.value.actual > 0 ? Math.min(100, Math.max(0, item.spend / totals.value.actual * 100)) : 0
  }))
})
const dark = ref(document.documentElement.classList.contains('dark'))
const observer = new MutationObserver(() => {
  dark.value = document.documentElement.classList.contains('dark')
})
onMounted(() =>
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
)
onBeforeUnmount(() => observer.disconnect())
const costChart = computed(() => ({
  labels: props.trend.map((d) => d.date),
  datasets: [
    {
      label: t('dashboard.actual'),
      data: props.trend.map((d) => Number(d.actual_cost ?? 0)),
      backgroundColor: dark.value ? '#d79a7e' : '#bd775e',
      hoverBackgroundColor: '#a9583e',
      borderRadius: 3,
      maxBarThickness: 28
    }
  ]
}))
const barOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: {
    duration: window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ? 0 : 250
  },
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (ctx: TooltipItem<'bar'>) =>
          `${t('dashboard.actual')}: ${formatCost(Number(ctx.raw))}`
      }
    }
  },
  scales: {
    x: {
      grid: { display: false },
      ticks: {
        color: dark.value ? '#b9b3a8' : '#6c6a64',
        maxTicksLimit: 7,
        maxRotation: 0,
        font: { size: 11 }
      },
      border: { display: false }
    },
    y: {
      beginAtZero: true,
      grid: { color: dark.value ? '#36322d' : '#ece6dd' },
      border: { display: false },
      ticks: {
        color: dark.value ? '#b9b3a8' : '#6c6a64',
        maxTicksLimit: 4,
        callback: (v: string | number) => formatCost(Number(v)),
        font: { size: 11 }
      }
    }
  }
}))
</script>
