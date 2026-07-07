<template>
  <div class="dashboard-panel card p-4">
    <div class="mb-4 flex items-center justify-between gap-3">
      <div class="flex items-center gap-2">
        <span class="dashboard-title-icon">
          <Icon name="trendingUp" size="sm" />
        </span>
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
          {{ t('admin.dashboard.tokenUsageTrend') }}
        </h3>
      </div>
      <span class="dashboard-panel-chip">Input / Output</span>
    </div>
    <div v-if="loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="trendData.length > 0 && chartData" class="h-56">
      <Line :data="chartData" :options="lineOptions" />
    </div>
    <div
      v-else
      class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import type { TrendDataPoint } from '@/types'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const { t } = useI18n()

const props = defineProps<{
  trendData: TrendDataPoint[]
  loading?: boolean
}>()

const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})

const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e8ded4' : '#3d3a34',
  muted: isDarkMode.value ? '#b4aaa0' : '#7a736b',
  grid: isDarkMode.value ? 'rgba(216, 206, 194, 0.11)' : 'rgba(160, 153, 144, 0.16)',
  tooltipBg: isDarkMode.value ? 'rgba(20, 20, 19, 0.96)' : 'rgba(255, 250, 245, 0.98)',
  tooltipText: isDarkMode.value ? '#fffaf5' : '#26251e',
  input: '#cc785c',
  output: '#8f8072',
  cacheCreation: '#d9a441',
  cacheRead: '#7f9d8a',
  cacheHitRate: '#6f6a5f'
}))

const chartData = computed(() => {
  if (!props.trendData?.length) return null

  return {
    labels: props.trendData.map((d) => d.date),
    datasets: [
      {
        label: 'Input',
        data: props.trendData.map((d) => d.input_tokens),
        borderColor: chartColors.value.input,
        backgroundColor: `${chartColors.value.input}24`,
        fill: true,
        borderWidth: 1.8,
        pointRadius: 2,
        pointHoverRadius: 4,
        tension: 0.35
      },
      {
        label: 'Output',
        data: props.trendData.map((d) => d.output_tokens),
        borderColor: chartColors.value.output,
        backgroundColor: `${chartColors.value.output}20`,
        fill: true,
        borderWidth: 1.8,
        pointRadius: 2,
        pointHoverRadius: 4,
        tension: 0.35
      },
      {
        label: 'Cache Creation',
        data: props.trendData.map((d) => d.cache_creation_tokens),
        borderColor: chartColors.value.cacheCreation,
        backgroundColor: `${chartColors.value.cacheCreation}18`,
        fill: false,
        borderWidth: 1.2,
        pointRadius: 1.5,
        pointHoverRadius: 3,
        tension: 0.35
      },
      {
        label: 'Cache Read',
        data: props.trendData.map((d) => d.cache_read_tokens),
        borderColor: chartColors.value.cacheRead,
        backgroundColor: `${chartColors.value.cacheRead}18`,
        fill: false,
        borderWidth: 1.2,
        pointRadius: 1.5,
        pointHoverRadius: 3,
        tension: 0.35
      },
      {
        label: 'Cache Hit Rate',
        data: props.trendData.map((d) => {
          const totalPromptTokens = d.input_tokens + d.cache_read_tokens + d.cache_creation_tokens
          return totalPromptTokens > 0 ? (d.cache_read_tokens / totalPromptTokens) * 100 : 0
        }),
        borderColor: chartColors.value.cacheHitRate,
        backgroundColor: `${chartColors.value.cacheHitRate}20`,
        borderDash: [5, 5],
        borderWidth: 1.2,
        pointRadius: 0,
        pointHoverRadius: 3,
        fill: false,
        tension: 0.35,
        yAxisID: 'yPercent'
      }
    ]
  }
})

const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        boxWidth: 8,
        boxHeight: 8,
        padding: 14,
        font: {
          size: 11,
          weight: 600
        }
      }
    },
    tooltip: {
      backgroundColor: chartColors.value.tooltipBg,
      borderColor: isDarkMode.value ? 'rgba(240, 184, 158, 0.32)' : 'rgba(204, 120, 92, 0.24)',
      borderWidth: 1,
      titleColor: chartColors.value.tooltipText,
      bodyColor: chartColors.value.tooltipText,
      footerColor: chartColors.value.muted,
      padding: 10,
      boxPadding: 4,
      callbacks: {
        label: (context: any) => {
          if (context.dataset.yAxisID === 'yPercent') {
            return `${context.dataset.label}: ${context.raw.toFixed(1)}%`
          }
          return `${context.dataset.label}: ${formatTokens(context.raw)}`
        },
        footer: (tooltipItems: any) => {
          const dataIndex = tooltipItems[0]?.dataIndex
          if (dataIndex !== undefined && props.trendData[dataIndex]) {
            const data = props.trendData[dataIndex]
            return `Actual: $${formatCost(data.actual_cost)} | Standard: $${formatCost(data.cost)}`
          }
          return ''
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        color: chartColors.value.grid,
        drawBorder: false
      },
      ticks: {
        color: chartColors.value.muted,
        font: {
          size: 10,
          weight: 500
        }
      }
    },
    y: {
      grid: {
        color: chartColors.value.grid,
        drawBorder: false
      },
      ticks: {
        color: chartColors.value.muted,
        font: {
          size: 10,
          weight: 500
        },
        callback: (value: string | number) => formatTokens(Number(value))
      }
    },
    yPercent: {
      position: 'right' as const,
      min: 0,
      max: 100,
      grid: {
        drawOnChartArea: false
      },
      ticks: {
        color: chartColors.value.cacheHitRate,
        font: {
          size: 10,
          weight: 600
        },
        callback: (value: string | number) => `${value}%`
      }
    }
  }
}))

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

const formatCost = (value: number): string => {
  if (value >= 1000) {
    return (value / 1000).toFixed(2) + 'K'
  } else if (value >= 1) {
    return value.toFixed(2)
  } else if (value >= 0.01) {
    return value.toFixed(3)
  }
  return value.toFixed(4)
}
</script>

<style scoped>
.dashboard-title-icon {
  display: inline-flex;
  width: 1.75rem;
  height: 1.75rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(204, 120, 92, 0.24);
  border-radius: 0.5rem;
  background: rgba(204, 120, 92, 0.1);
  color: #a9583e;
}

.dark .dashboard-title-icon {
  border-color: rgba(240, 184, 158, 0.3);
  background: rgba(204, 120, 92, 0.16);
  color: #f0b89e;
}

.dashboard-panel-chip {
  border: 1px solid rgba(216, 206, 194, 0.74);
  border-radius: 999px;
  background: rgba(255, 250, 245, 0.72);
  padding: 0.25rem 0.625rem;
  color: #7a736b;
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1;
}

.dark .dashboard-panel-chip {
  border-color: rgba(71, 85, 105, 0.74);
  background: rgba(15, 23, 42, 0.64);
  color: #94a3b8;
}
</style>
