<template>
  <section class="dashboard-panel dashboard-token-trend-panel">
    <div class="dashboard-panel-header">
      <div>
        <p class="dashboard-eyebrow">{{ t('dashboard.console.usageAnalysis') }}</p>
        <h2 class="dashboard-panel-title">{{ t('dashboard.console.tokenTimeTrend') }}</h2>
        <p class="dashboard-panel-subtitle">{{ t('dashboard.overview.tokenTrendDescription') }}</p>
      </div>
    </div>
    <div v-if="loading" class="dashboard-chart-state" role="status">
      <LoadingSpinner size="md" />
    </div>
    <div v-else-if="error" class="dashboard-chart-state" role="alert">
      <p>{{ t('dashboard.console.loadError') }}</p>
      <button class="btn btn-secondary" type="button" @click="$emit('retry')">
        {{ t('dashboard.console.retry') }}
      </button>
    </div>
    <div v-else-if="!hasRequests" class="dashboard-model-empty">
      {{ t('dashboard.console.rangeEmptyDescription') }}
    </div>
    <div v-else class="dashboard-token-time-chart">
      <Line
        :data="chartData"
        :options="options"
        role="img"
        :aria-label="t('dashboard.console.tokenTimeTrend')"
      />
    </div>
  </section>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip
} from 'chart.js'
import { Line } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { TrendDataPoint } from '@/types'

ChartJS.register(CategoryScale, LinearScale, LineElement, PointElement, Legend, Tooltip)
const props = defineProps<{ trend: TrendDataPoint[]; loading: boolean; error: boolean }>()
defineEmits<{ retry: [] }>()
const { t } = useI18n()
const hasRequests = computed(() => props.trend.some((day) => day.requests > 0))
const dark = ref(document.documentElement.classList.contains('dark'))
const observer = new MutationObserver(() => {
  dark.value = document.documentElement.classList.contains('dark')
})
onMounted(() =>
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
)
onBeforeUnmount(() => observer.disconnect())
const chartData = computed(() => ({
  labels: props.trend.map((day) => day.date),
  datasets: [
    {
      label: t('dashboard.input'),
      data: props.trend.map((day) => day.input_tokens),
      borderColor: dark.value ? '#e3a58c' : '#a9583e'
    },
    {
      label: t('dashboard.output'),
      data: props.trend.map((day) => day.output_tokens),
      borderColor: dark.value ? '#a4beae' : '#648374'
    },
    {
      label: t('dashboard.cacheReadTokens'),
      data: props.trend.map((day) => day.cache_read_tokens),
      borderColor: dark.value ? '#b4a5c8' : '#8b7ca0'
    },
    {
      label: t('dashboard.console.cacheWrite'),
      data: props.trend.map((day) => day.cache_creation_tokens),
      borderColor: dark.value ? '#d6b680' : '#a58449'
    }
  ].map((dataset) => ({
    ...dataset,
    borderWidth: 1.8,
    pointRadius: 0,
    pointHoverRadius: 3,
    tension: 0.25
  }))
}))
const options = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: {
    duration: window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ? 0 : 250
  },
  interaction: { mode: 'index' as const, intersect: false },
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: {
        color: dark.value ? '#b9b3a8' : '#6c6a64',
        usePointStyle: true,
        boxHeight: 5,
        boxWidth: 5,
        padding: 14,
        font: { size: 11 }
      }
    }
  },
  scales: {
    x: {
      grid: { display: false },
      border: { display: false },
      ticks: {
        color: dark.value ? '#b9b3a8' : '#6c6a64',
        maxTicksLimit: 4,
        maxRotation: 0,
        font: { size: 11 }
      }
    },
    y: {
      beginAtZero: true,
      border: { display: false },
      grid: { color: dark.value ? '#36322d' : '#ece6dd' },
      ticks: {
        color: dark.value ? '#b9b3a8' : '#6c6a64',
        maxTicksLimit: 4,
        font: { size: 11 },
        callback: (value: number | string) =>
          new Intl.NumberFormat(undefined, { notation: 'compact' }).format(Number(value))
      }
    }
  }
}))
</script>
