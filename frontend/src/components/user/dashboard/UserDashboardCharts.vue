<template>
  <section class="dashboard-charts-stack">
    <article class="dashboard-panel dashboard-panel--large">
      <div class="dashboard-panel-header">
        <div>
          <h2 class="dashboard-panel-title">{{ t('dashboard.overview.dailyCostTrend') }}</h2>
          <p class="dashboard-panel-subtitle">{{ t('dashboard.overview.dailyCostTrendDescription') }}</p>
        </div>
        <div class="dashboard-range-tabs" aria-label="Dashboard range">
          <button
            v-for="preset in rangePresets"
            :key="preset.key"
            type="button"
            class="dashboard-range-tab"
            :class="{ 'dashboard-range-tab--active': activeRange === preset.key }"
            @click="applyRange(preset.key)"
          >
            {{ preset.label }}
          </button>
        </div>
      </div>

      <div v-if="loading" class="dashboard-chart-state dashboard-chart-state--large">
        <LoadingSpinner size="md" />
      </div>
      <div v-else-if="hasCostData && costChartData" class="dashboard-main-chart">
        <Line :data="costChartData" :options="costLineOptions" />
      </div>
      <div v-else class="dashboard-chart-state dashboard-chart-state--large">
        {{ t('dashboard.noDataAvailable') }}
      </div>
    </article>

    <div class="dashboard-chart-grid">
      <article class="dashboard-panel">
        <div class="dashboard-panel-header dashboard-panel-header--compact">
          <div>
            <h3 class="dashboard-panel-title">{{ t('dashboard.overview.tokenTrend') }}</h3>
            <p class="dashboard-panel-subtitle">{{ t('dashboard.overview.tokenTrendDescription') }}</p>
          </div>
        </div>
        <div v-if="loading" class="dashboard-chart-state">
          <LoadingSpinner size="md" />
        </div>
        <div v-else-if="hasTokenData && tokenChartData" class="dashboard-small-chart">
          <Line :data="tokenChartData" :options="smallLineOptions" />
        </div>
        <div v-else class="dashboard-chart-state">
          {{ t('dashboard.noDataAvailable') }}
        </div>
      </article>

      <article class="dashboard-panel">
        <div class="dashboard-panel-header dashboard-panel-header--compact">
          <div>
            <h3 class="dashboard-panel-title">{{ t('dashboard.overview.requestTrend') }}</h3>
            <p class="dashboard-panel-subtitle">{{ t('dashboard.overview.requestTrendDescription') }}</p>
          </div>
        </div>
        <div v-if="loading" class="dashboard-chart-state">
          <LoadingSpinner size="md" />
        </div>
        <div v-else-if="hasRequestData && requestChartData" class="dashboard-small-chart">
          <Line :data="requestChartData" :options="smallLineOptions" />
        </div>
        <div v-else class="dashboard-chart-state">
          {{ t('dashboard.noDataAvailable') }}
        </div>
      </article>
    </div>

  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { TrendDataPoint } from '@/types'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

type RangePreset = '90d' | '30d' | '7d'

const props = defineProps<{
  loading: boolean
  startDate: string
  endDate: string
  granularity: string
  trend: TrendDataPoint[]
}>()

const emit = defineEmits<{
  'update:startDate': [value: string]
  'update:endDate': [value: string]
  'update:granularity': [value: string]
  dateRangeChange: [range: { startDate: string; endDate: string; preset: RangePreset }]
}>()

const { t } = useI18n()

const formatLD = (d: Date) => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const daysBetweenInclusive = computed(() => {
  const start = new Date(`${props.startDate}T00:00:00`).getTime()
  const end = new Date(`${props.endDate}T00:00:00`).getTime()
  if (!Number.isFinite(start) || !Number.isFinite(end) || end < start) return 0
  return Math.round((end - start) / 86400000) + 1
})

const activeRange = computed<RangePreset>(() => {
  const days = daysBetweenInclusive.value
  if (days <= 8) return '7d'
  if (days <= 31) return '30d'
  return '90d'
})

const rangePresets = computed(() => [
  { key: '90d' as const, label: t('dashboard.overview.last90Days') },
  { key: '30d' as const, label: t('dashboard.overview.last30Days') },
  { key: '7d' as const, label: t('dashboard.overview.last7Days') },
])

const applyRange = (preset: RangePreset) => {
  const days = preset === '90d' ? 90 : preset === '30d' ? 30 : 7
  const end = new Date()
  const start = new Date()
  start.setDate(end.getDate() - days + 1)
  const startValue = formatLD(start)
  const endValue = formatLD(end)
  emit('update:startDate', startValue)
  emit('update:endDate', endValue)
  if (props.granularity !== 'day') {
    emit('update:granularity', 'day')
  }
  emit('dateRangeChange', { startDate: startValue, endDate: endValue, preset })
}

const chartLabels = computed(() => props.trend.map((d) => d.date))
const hasCostData = computed(() => props.trend.some((d) => Number(d.actual_cost || d.cost || 0) > 0))
const hasTokenData = computed(() => props.trend.some((d) => Number(d.total_tokens || d.input_tokens || d.output_tokens || 0) > 0))
const hasRequestData = computed(() => props.trend.some((d) => Number(d.requests || 0) > 0))

const chartColors = computed(() => {
  const isDark = document.documentElement.classList.contains('dark')
  return {
    grid: isDark ? 'rgba(148, 163, 184, 0.13)' : 'rgba(148, 163, 184, 0.22)',
    text: isDark ? '#cbd5e1' : '#64748b',
    title: isDark ? '#f8fafc' : '#0f172a',
    tooltipBg: isDark ? 'rgba(15, 23, 42, 0.98)' : 'rgba(255, 255, 255, 0.98)',
    tooltipBorder: isDark ? 'rgba(148, 163, 184, 0.28)' : 'rgba(100, 116, 139, 0.18)',
    actual: '#14b8a6',
    standard: '#f59e0b',
    input: '#a78bfa',
    output: '#22d3ee',
    request: '#60a5fa',
  }
})

const costChartData = computed(() => {
  if (!hasCostData.value) return null
  return {
    labels: chartLabels.value,
    datasets: [
      {
        label: t('dashboard.actual'),
        data: props.trend.map((d) => d.actual_cost),
        borderColor: chartColors.value.actual,
        backgroundColor: `${chartColors.value.actual}30`,
        fill: true,
        borderWidth: 1.8,
        pointRadius: 0,
        pointHoverRadius: 3,
        tension: 0.38,
      },
      {
        label: t('dashboard.standard'),
        data: props.trend.map((d) => d.cost),
        borderColor: chartColors.value.standard,
        backgroundColor: `${chartColors.value.standard}24`,
        fill: true,
        borderWidth: 1.4,
        pointRadius: 0,
        pointHoverRadius: 3,
        tension: 0.38,
      },
    ],
  }
})

const tokenChartData = computed(() => {
  if (!hasTokenData.value) return null
  return {
    labels: chartLabels.value,
    datasets: [
      {
        label: t('dashboard.input'),
        data: props.trend.map((d) => d.input_tokens + d.cache_read_tokens + d.cache_creation_tokens),
        borderColor: chartColors.value.input,
        backgroundColor: `${chartColors.value.input}24`,
        fill: true,
        borderWidth: 1.5,
        pointRadius: 0,
        tension: 0.36,
      },
      {
        label: t('dashboard.output'),
        data: props.trend.map((d) => d.output_tokens),
        borderColor: chartColors.value.output,
        backgroundColor: `${chartColors.value.output}20`,
        fill: true,
        borderWidth: 1.5,
        pointRadius: 0,
        tension: 0.36,
      },
    ],
  }
})

const requestChartData = computed(() => {
  if (!hasRequestData.value) return null
  return {
    labels: chartLabels.value,
    datasets: [
      {
        label: t('dashboard.requests'),
        data: props.trend.map((d) => d.requests),
        borderColor: chartColors.value.request,
        backgroundColor: `${chartColors.value.request}26`,
        fill: true,
        borderWidth: 1.6,
        pointRadius: 0,
        tension: 0.34,
      },
    ],
  }
})

const sharedLineOptions = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const,
  },
}

const costLineOptions = computed(() => ({
  ...sharedLineOptions,
  plugins: {
    legend: {
      display: true,
      position: 'top' as const,
      align: 'end' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        boxWidth: 7,
        boxHeight: 7,
        font: { size: 11, weight: 600 },
      },
    },
    tooltip: tooltipOptions.value,
  },
  scales: axisOptions.value,
}))

const smallLineOptions = computed(() => ({
  ...sharedLineOptions,
  plugins: {
    legend: { display: false },
    tooltip: tooltipOptions.value,
  },
  scales: axisOptions.value,
}))

const tooltipOptions = computed(() => ({
  backgroundColor: chartColors.value.tooltipBg,
  borderColor: chartColors.value.tooltipBorder,
  borderWidth: 1,
  titleColor: chartColors.value.title,
  bodyColor: chartColors.value.title,
  padding: 10,
  callbacks: {
    label: (context: any) => {
      const value = Number(context.raw) || 0
      if (context.dataset.label === t('dashboard.actual') || context.dataset.label === t('dashboard.standard')) {
        return `${context.dataset.label}: $${formatCost(value)}`
      }
      return `${context.dataset.label}: ${formatCompact(value)}`
    },
  },
}))

const axisOptions = computed(() => ({
  x: {
    grid: {
      color: chartColors.value.grid,
      drawBorder: false,
    },
    ticks: {
      color: chartColors.value.text,
      maxTicksLimit: 9,
      font: { size: 10, weight: 500 },
    },
  },
  y: {
    grid: {
      color: chartColors.value.grid,
      drawBorder: false,
    },
    ticks: {
      color: chartColors.value.text,
      font: { size: 10, weight: 500 },
      callback: (value: string | number) => formatCompact(Number(value)),
    },
  },
}))

const formatCompact = (value: number): string =>
  new Intl.NumberFormat(undefined, {
    notation: 'compact',
    maximumFractionDigits: Number(value) >= 1000 ? 1 : 0,
  }).format(Number(value) || 0)

const formatCost = (value: number): string =>
  new Intl.NumberFormat(undefined, {
    minimumFractionDigits: Number(value) >= 1 ? 2 : 4,
    maximumFractionDigits: 4,
  }).format(Number(value) || 0)
</script>
