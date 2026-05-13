<template>
  <AppLayout>
    <div class="dashboard-overview space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <!-- Row 1: Core Stats -->
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <!-- Total API Keys -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="console-stat-icon console-stat-icon--blue">
                <Icon name="key" size="md" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.apiKeys') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ stats.total_api_keys }}
                </p>
                <p class="text-xs text-green-600 dark:text-green-400">
                  {{ stats.active_api_keys }} {{ t('common.active') }}
                </p>
              </div>
            </div>
          </div>

          <!-- Service Accounts -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="console-stat-icon">
                <Icon name="server" size="md" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.accounts') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ stats.total_accounts }}
                </p>
                <p class="text-xs">
                  <span class="text-green-600 dark:text-green-400"
                    >{{ stats.normal_accounts }} {{ t('common.active') }}</span
                  >
                  <span v-if="stats.error_accounts > 0" class="ml-1 text-red-500"
                    >{{ stats.error_accounts }} {{ t('common.error') }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Today Requests -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="console-stat-icon console-stat-icon--green">
                <Icon name="trendingUp" size="md" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.todayRequests') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ stats.today_requests }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('common.total') }}: {{ formatNumber(stats.total_requests) }}
                </p>
              </div>
            </div>
          </div>

          <!-- New Users Today -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="console-stat-icon console-stat-icon--green">
                <Icon name="userPlus" size="md" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.users') }}
                </p>
                <p class="text-xl font-bold text-emerald-600 dark:text-emerald-400">
                  +{{ stats.today_new_users }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('common.total') }}: {{ formatNumber(stats.total_users) }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Row 2: Token Stats -->
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <!-- Today Tokens -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="console-stat-icon console-stat-icon--yellow">
                <Icon name="database" size="md" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.todayTokens') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatTokens(stats.today_tokens) }}
                </p>
                <p class="text-xs">
                  <span
                    class="text-green-600 dark:text-green-400"
                    :title="t('admin.dashboard.actual')"
                    >${{ formatCost(stats.today_actual_cost) }}</span
                  >
                  <span class="text-gray-400 dark:text-gray-500"> / </span>
                  <span
                    class="text-orange-500 dark:text-orange-400"
                    :title="t('admin.dashboard.accountCost')"
                    >${{ formatCost(stats.today_account_cost) }}</span
                  >
                  <span class="text-gray-400 dark:text-gray-500"> / </span>
                  <span
                    class="text-gray-400 dark:text-gray-500"
                    :title="t('admin.dashboard.standard')"
                    >${{ formatCost(stats.today_cost) }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Total Tokens -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="console-stat-icon">
                <Icon name="cube" size="md" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.totalTokens') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatTokens(stats.total_tokens) }}
                </p>
                <p class="text-xs">
                  <span
                    class="text-green-600 dark:text-green-400"
                    :title="t('admin.dashboard.actual')"
                    >${{ formatCost(stats.total_actual_cost) }}</span
                  >
                  <span class="text-gray-400 dark:text-gray-500"> / </span>
                  <span
                    class="text-orange-500 dark:text-orange-400"
                    :title="t('admin.dashboard.accountCost')"
                    >${{ formatCost(stats.total_account_cost) }}</span
                  >
                  <span class="text-gray-400 dark:text-gray-500"> / </span>
                  <span
                    class="text-gray-400 dark:text-gray-500"
                    :title="t('admin.dashboard.standard')"
                    >${{ formatCost(stats.total_cost) }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Performance (RPM/TPM) -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="console-stat-icon">
                <Icon name="bolt" size="md" />
              </div>
              <div class="flex-1">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.performance') }}
                </p>
                <div class="flex items-baseline gap-2">
                  <p class="text-xl font-bold text-gray-900 dark:text-white">
                    {{ formatTokens(stats.rpm) }}
                  </p>
                  <span class="text-xs text-gray-500 dark:text-gray-400">RPM</span>
                </div>
                <div class="flex items-baseline gap-2">
                  <p class="text-sm font-semibold text-violet-600 dark:text-violet-400">
                    {{ formatTokens(stats.tpm) }}
                  </p>
                  <span class="text-xs text-gray-500 dark:text-gray-400">TPM</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Avg Response Time -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="console-stat-icon console-stat-icon--danger">
                <Icon name="clock" size="md" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.avgResponse') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatDuration(stats.average_duration_ms) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ stats.active_users }} {{ t('admin.dashboard.activeUsers') }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Row 3: Cache Stats -->
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <!-- Today Input Cache Reuse -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="console-stat-icon console-stat-icon--blue">
                <Icon name="sparkles" size="md" />
              </div>
              <div class="min-w-0">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.todayCacheHitRate') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatCacheReuseRate(stats.today_cache_read_tokens, stats.today_input_tokens) }}
                </p>
                <p
                  class="truncate text-xs text-gray-500 dark:text-gray-400"
                  :title="cacheReadTitle(stats.today_cache_read_tokens)"
                >
                  {{ t('admin.dashboard.cacheReadTokens') }}: {{ formatTokens(stats.today_cache_read_tokens) }}
                </p>
              </div>
            </div>
          </div>

          <!-- Total Input Cache Reuse -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="console-stat-icon console-stat-icon--green">
                <Icon name="shield" size="md" />
              </div>
              <div class="min-w-0">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.totalCacheHitRate') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatCacheReuseRate(stats.total_cache_read_tokens, stats.total_input_tokens) }}
                </p>
                <p
                  class="truncate text-xs text-gray-500 dark:text-gray-400"
                  :title="cacheReadTitle(stats.total_cache_read_tokens)"
                >
                  {{ t('admin.dashboard.cacheReadTokens') }}: {{ formatTokens(stats.total_cache_read_tokens) }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Charts Section -->
        <div class="space-y-6">
          <!-- Date Range Filter -->
          <div class="dashboard-panel dashboard-panel--filters p-4">
            <div class="flex flex-wrap items-center gap-4">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300"
                  >{{ t('admin.dashboard.timeRange') }}:</span
                >
                <DateRangePicker
                  v-model:start-date="startDate"
                  v-model:end-date="endDate"
                  @change="onDateRangeChange"
                />
              </div>
              <button @click="loadDashboardStats" :disabled="chartsLoading" class="btn btn-secondary">
                {{ t('common.refresh') }}
              </button>
              <div class="ml-auto flex items-center gap-2">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300"
                  >{{ t('admin.dashboard.granularity') }}:</span
                >
                <div class="w-28">
                  <Select
                    v-model="granularity"
                    :options="granularityOptions"
                    @change="loadChartData"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Charts Grid -->
          <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <ModelDistributionChart
              :model-stats="modelStats"
              :enable-ranking-view="true"
              :ranking-items="rankingItems"
              :ranking-total-actual-cost="rankingTotalActualCost"
              :ranking-total-requests="rankingTotalRequests"
              :ranking-total-tokens="rankingTotalTokens"
              :loading="chartsLoading"
              :ranking-loading="rankingLoading"
              :ranking-error="rankingError"
              :start-date="startDate"
              :end-date="endDate"
              @ranking-click="goToUserUsage"
            />
            <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
          </div>

          <!-- User Usage Trend (Full Width) -->
          <div class="dashboard-panel p-4">
            <div class="mb-4 flex items-center justify-between gap-3">
              <div class="flex items-center gap-2">
                <span class="dashboard-title-icon dashboard-title-icon--blue">
                  <Icon name="chartBar" size="sm" />
                </span>
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                  {{ t('admin.dashboard.recentUsage') }} (Top 12)
                </h3>
              </div>
              <span class="dashboard-panel-chip">{{ t('admin.dashboard.tokens') }}</span>
            </div>
            <div class="h-72">
              <div v-if="userTrendLoading" class="flex h-full items-center justify-center">
                <LoadingSpinner size="md" />
              </div>
              <Bar v-else-if="userTrendChartData" :data="userTrendChartData" :options="barOptions" />
              <div
                v-else
                class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-gray-400"
              >
                {{ t('admin.dashboard.noDataAvailable') }}
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
import { adminAPI } from '@/api/admin'
import type {
  DashboardStats,
  TrendDataPoint,
  ModelStat,
  UserUsageTrendPoint,
  UserSpendingRankingItem
} from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Bar } from 'vue-chartjs'

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
)

const appStore = useAppStore()
const router = useRouter()
const stats = ref<DashboardStats | null>(null)
const loading = ref(false)
const chartsLoading = ref(false)
const userTrendLoading = ref(false)
const rankingLoading = ref(false)
const rankingError = ref(false)

// Chart data
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const userTrend = ref<UserUsageTrendPoint[]>([])
const rankingItems = ref<UserSpendingRankingItem[]>([])
const rankingTotalActualCost = ref(0)
const rankingTotalRequests = ref(0)
const rankingTotalTokens = ref(0)
let chartLoadSeq = 0
let usersTrendLoadSeq = 0
let rankingLoadSeq = 0
const rankingLimit = 12

// Helper function to format date in local timezone
const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const getLast24HoursRangeDates = (): { start: string; end: string } => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return {
    start: formatLocalDate(start),
    end: formatLocalDate(end)
  }
}

// Date range
const granularity = ref<'day' | 'hour'>('hour')
const defaultRange = getLast24HoursRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)

// Granularity options for Select component
const granularityOptions = computed(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') }
])

// Dark mode detection
const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})

// Chart colors
const chartColors = computed(() => ({
  text: isDarkMode.value ? '#dbeafe' : '#334155',
  muted: isDarkMode.value ? '#94a3b8' : '#64748b',
  grid: isDarkMode.value ? 'rgba(148, 163, 184, 0.1)' : 'rgba(100, 116, 139, 0.1)',
  tooltipBg: isDarkMode.value ? 'rgba(8, 13, 26, 0.96)' : 'rgba(255, 255, 255, 0.96)',
  tooltipText: isDarkMode.value ? '#f8fafc' : '#0f172a',
  blue: '#3b82f6',
  cyan: '#06b6d4',
  emerald: '#10b981'
}))

// Bar chart options (for user usage ranking)
const barOptions = computed(() => ({
  indexAxis: 'y' as const,
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'nearest' as const
  },
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      backgroundColor: chartColors.value.tooltipBg,
      borderColor: isDarkMode.value ? 'rgba(59, 130, 246, 0.34)' : 'rgba(59, 130, 246, 0.18)',
      borderWidth: 1,
      titleColor: chartColors.value.tooltipText,
      bodyColor: chartColors.value.tooltipText,
      padding: 10,
      displayColors: false,
      callbacks: {
        title: (items: any[]) => {
          const index = items[0]?.dataIndex ?? 0
          const labels = userTrendChartData.value?.fullLabels || []
          return labels[index] || items[0]?.label || ''
        },
        label: (context: any) => {
          return `${t('admin.dashboard.tokens')}: ${formatTokens(context.raw)}`
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
        },
        callback: (value: string | number) => formatTokens(Number(value))
      }
    },
    y: {
      grid: {
        display: false,
        drawBorder: false
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 11,
          weight: 600
        }
      }
    }
  }
}))

// User trend chart data
const userTrendChartData = computed(() => {
  if (!userTrend.value?.length) return null

  const getDisplayName = (point: UserUsageTrendPoint): string => {
    const username = point.username?.trim()
    if (username) {
      return username
    }

    const email = point.email?.trim()
    if (email) {
      return email
    }

    return t('admin.redeem.userPrefix', { id: point.user_id })
  }

  // Group by user_id to avoid merging different users with the same display name
  const userGroups = new Map<number, { name: string; data: Map<string, number> }>()
  const allDates = new Set<string>()

  userTrend.value.forEach((point) => {
    allDates.add(point.date)
    const key = point.user_id
    if (!userGroups.has(key)) {
      userGroups.set(key, { name: getDisplayName(point), data: new Map() })
    }
    userGroups.get(key)!.data.set(point.date, point.tokens)
  })

  const totals = Array.from(userGroups.values())
    .map((group) => ({
      name: group.name,
      total: Array.from(group.data.values()).reduce((sum, value) => sum + value, 0)
    }))
    .sort((a, b) => b.total - a.total)
    .slice(0, 12)

  const fullLabels = totals.map((item) => item.name)
  const labels = fullLabels.map((label) => {
    if (label.length <= 28) return label
    return `${label.slice(0, 25)}...`
  })

  const gradients = totals.map((_, idx) => {
    const palette = [
      chartColors.value.blue,
      chartColors.value.emerald,
      chartColors.value.cyan,
      '#f59e0b',
      '#8b5cf6',
      '#f97316',
      '#14b8a6',
      '#6366f1',
      '#22c55e',
      '#eab308',
      '#0ea5e9',
      '#a855f7'
    ]
    return palette[idx % palette.length]
  })

  return {
    labels,
    fullLabels,
    datasets: [
      {
        label: t('admin.dashboard.tokens'),
        data: totals.map((item) => item.total),
        backgroundColor: gradients.map((color) => `${color}b8`),
        borderColor: gradients.map((color) => `${color}8a`),
        borderWidth: 0,
        hoverBorderWidth: 0,
        borderRadius: 7,
        borderSkipped: false,
        barThickness: 12,
        maxBarThickness: 16
      }
    ]
  }
})

// Format helpers
const formatTokens = (value: number | undefined): string => {
  if (value === undefined || value === null) return '0'
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

const formatNumber = (value: number): string => {
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

const formatDuration = (ms: number): string => {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`
  }
  return `${Math.round(ms)}ms`
}

const formatCacheReuseRate = (cacheReadTokens: number, inputTokens: number): string => {
  const inputTotal = cacheReadTokens + inputTokens
  if (inputTotal <= 0) return t('common.notAvailable')
  return `${((cacheReadTokens / inputTotal) * 100).toFixed(1)}%`
}

const cacheReadTitle = (cacheReadTokens: number): string =>
  `${t('admin.dashboard.cacheReadTokens')}: ${formatNumber(cacheReadTokens)}`

const goToUserUsage = (item: UserSpendingRankingItem) => {
  void router.push({
    path: '/admin/usage',
    query: {
      user_id: String(item.user_id),
      start_date: startDate.value,
      end_date: endDate.value
    }
  })
}

// Date range change handler
const onDateRangeChange = (range: {
  startDate: string
  endDate: string
  preset: string | null
}) => {
  // Auto-select granularity based on date range
  const start = new Date(range.startDate)
  const end = new Date(range.endDate)
  const daysDiff = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))

  // If range is 1 day, use hourly granularity
  if (daysDiff <= 1) {
    granularity.value = 'hour'
  } else {
    granularity.value = 'day'
  }

  loadChartData()
}

// Load data
const loadDashboardSnapshot = async (includeStats: boolean) => {
  const currentSeq = ++chartLoadSeq
  if (includeStats && !stats.value) {
    loading.value = true
  }
  chartsLoading.value = true
  try {
    const response = await adminAPI.dashboard.getSnapshotV2({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      include_stats: includeStats,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false
    })
    if (currentSeq !== chartLoadSeq) return
    if (includeStats && response.stats) {
      stats.value = response.stats
    }
    trendData.value = response.trend || []
    modelStats.value = response.models || []
  } catch (error) {
    if (currentSeq !== chartLoadSeq) return
    appStore.showError(t('admin.dashboard.failedToLoad'))
    console.error('Error loading dashboard snapshot:', error)
  } finally {
    if (currentSeq === chartLoadSeq) {
      loading.value = false
      chartsLoading.value = false
    }
  }
}

const loadUsersTrend = async () => {
  const currentSeq = ++usersTrendLoadSeq
  userTrendLoading.value = true
  try {
    const response = await adminAPI.dashboard.getUserUsageTrend({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      limit: 12
    })
    if (currentSeq !== usersTrendLoadSeq) return
    userTrend.value = response.trend || []
  } catch (error) {
    if (currentSeq !== usersTrendLoadSeq) return
    console.error('Error loading users trend:', error)
    userTrend.value = []
  } finally {
    if (currentSeq === usersTrendLoadSeq) {
      userTrendLoading.value = false
    }
  }
}

const loadUserSpendingRanking = async () => {
  const currentSeq = ++rankingLoadSeq
  rankingLoading.value = true
  rankingError.value = false
  try {
    const response = await adminAPI.dashboard.getUserSpendingRanking({
      start_date: startDate.value,
      end_date: endDate.value,
      limit: rankingLimit
    })
    if (currentSeq !== rankingLoadSeq) return
    rankingItems.value = response.ranking || []
    rankingTotalActualCost.value = response.total_actual_cost || 0
    rankingTotalRequests.value = response.total_requests || 0
    rankingTotalTokens.value = response.total_tokens || 0
  } catch (error) {
    if (currentSeq !== rankingLoadSeq) return
    console.error('Error loading user spending ranking:', error)
    rankingItems.value = []
    rankingTotalActualCost.value = 0
    rankingTotalRequests.value = 0
    rankingTotalTokens.value = 0
    rankingError.value = true
  } finally {
    if (currentSeq === rankingLoadSeq) {
      rankingLoading.value = false
    }
  }
}

const loadDashboardStats = async () => {
  await Promise.all([
    loadDashboardSnapshot(true),
    loadUsersTrend(),
    loadUserSpendingRanking()
  ])
}

const loadChartData = async () => {
  await Promise.all([
    loadDashboardSnapshot(false),
    loadUsersTrend(),
    loadUserSpendingRanking()
  ])
}

onMounted(() => {
  loadDashboardStats()
})
</script>

<style scoped>
.dashboard-overview :deep(.dashboard-panel),
.dashboard-panel {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 0.75rem;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.92), rgba(248, 250, 252, 0.82)),
    radial-gradient(circle at 100% 0%, rgba(59, 130, 246, 0.08), transparent 18rem);
  box-shadow: 0 18px 44px rgba(15, 23, 42, 0.07);
  backdrop-filter: blur(16px);
}

.dark .dashboard-overview :deep(.dashboard-panel),
.dark .dashboard-panel {
  border-color: rgba(55, 65, 81, 0.9);
  background:
    linear-gradient(135deg, rgba(13, 20, 33, 0.96), rgba(9, 14, 26, 0.9)),
    radial-gradient(circle at 100% 0%, rgba(6, 182, 212, 0.08), transparent 19rem);
  box-shadow: 0 18px 48px rgba(0, 0, 0, 0.25);
}

.dashboard-panel::before {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(90deg, rgba(59, 130, 246, 0.16), transparent 18rem),
    linear-gradient(rgba(148, 163, 184, 0.08) 1px, transparent 1px);
  background-size: auto, 100% 2.75rem;
  opacity: 0.28;
  content: '';
}

.dark .dashboard-panel::before {
  background:
    linear-gradient(90deg, rgba(6, 182, 212, 0.12), transparent 18rem),
    linear-gradient(rgba(148, 163, 184, 0.12) 1px, transparent 1px);
  opacity: 0.2;
}

.dashboard-panel > * {
  position: relative;
  z-index: 1;
}

.dashboard-panel--filters {
  border-radius: 0.625rem;
}

.dashboard-title-icon {
  display: inline-flex;
  width: 1.75rem;
  height: 1.75rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 0.5rem;
  background: rgba(59, 130, 246, 0.09);
  color: #2563eb;
}

.dark .dashboard-title-icon {
  border-color: rgba(96, 165, 250, 0.28);
  background: rgba(59, 130, 246, 0.14);
  color: #93c5fd;
}

.dashboard-title-icon--blue {
  border-color: rgba(6, 182, 212, 0.22);
  background: rgba(6, 182, 212, 0.09);
  color: #0891b2;
}

.dark .dashboard-title-icon--blue {
  border-color: rgba(34, 211, 238, 0.26);
  background: rgba(8, 145, 178, 0.16);
  color: #67e8f9;
}

.dashboard-panel-chip {
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.62);
  padding: 0.25rem 0.625rem;
  color: #64748b;
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
