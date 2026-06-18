<template>
  <div class="dashboard-panel card p-4">
    <div class="mb-4 flex items-center justify-between gap-3">
      <div class="flex items-center gap-2">
        <span class="dashboard-title-icon">
          <Icon name="cube" size="sm" />
        </span>
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
          {{ !enableRankingView || activeView === 'model_distribution'
            ? t('admin.dashboard.modelDistribution')
            : t('admin.dashboard.spendingRankingTitle') }}
        </h3>
      </div>
      <div class="flex flex-wrap items-center justify-end gap-2">
        <div
          v-if="showSourceToggle"
          class="dashboard-segmented"
        >
          <button
            type="button"
            class="dashboard-segmented-button"
            :class="source === 'requested'
              ? 'dashboard-segmented-button--active'
              : 'dashboard-segmented-button--idle'"
            @click="emit('update:source', 'requested')"
          >
            {{ t('usage.requestedModel') }}
          </button>
          <button
            type="button"
            class="dashboard-segmented-button"
            :class="source === 'upstream'
              ? 'dashboard-segmented-button--active'
              : 'dashboard-segmented-button--idle'"
            @click="emit('update:source', 'upstream')"
          >
            {{ t('usage.upstreamModel') }}
          </button>
          <button
            type="button"
            class="dashboard-segmented-button"
            :class="source === 'mapping'
              ? 'dashboard-segmented-button--active'
              : 'dashboard-segmented-button--idle'"
            @click="emit('update:source', 'mapping')"
          >
            {{ t('usage.mapping') }}
          </button>
        </div>
        <div
          v-if="showMetricToggle"
          class="dashboard-segmented"
        >
          <button
            type="button"
            class="dashboard-segmented-button"
            :class="metric === 'tokens'
              ? 'dashboard-segmented-button--active'
              : 'dashboard-segmented-button--idle'"
            @click="emit('update:metric', 'tokens')"
          >
            {{ t('admin.dashboard.metricTokens') }}
          </button>
          <button
            type="button"
            class="dashboard-segmented-button"
            :class="metric === 'actual_cost'
              ? 'dashboard-segmented-button--active'
              : 'dashboard-segmented-button--idle'"
            @click="emit('update:metric', 'actual_cost')"
          >
            {{ t('admin.dashboard.metricActualCost') }}
          </button>
        </div>
        <div v-if="enableRankingView" class="dashboard-segmented">
          <button
            type="button"
            class="dashboard-segmented-button"
            :class="
              activeView === 'model_distribution'
                ? 'dashboard-segmented-button--active'
                : 'dashboard-segmented-button--idle'
            "
            @click="activeView = 'model_distribution'"
          >
            {{ t('admin.dashboard.viewModelDistribution') }}
          </button>
          <button
            type="button"
            class="dashboard-segmented-button"
            :class="
              activeView === 'spending_ranking'
                ? 'dashboard-segmented-button--active'
                : 'dashboard-segmented-button--idle'
            "
            @click="activeView = 'spending_ranking'"
          >
            {{ t('admin.dashboard.viewSpendingRanking') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="activeView === 'model_distribution' && loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div
      v-else-if="activeView === 'model_distribution' && displayModelStats.length > 0 && chartData"
      class="dashboard-chart-split"
    >
      <div class="dashboard-doughnut-wrap">
        <Doughnut :data="chartData" :options="doughnutOptions" />
        <div class="dashboard-doughnut-center">
          <span>{{ props.metric === 'actual_cost' ? t('admin.dashboard.actual') : t('admin.dashboard.tokens') }}</span>
          <strong>{{ props.metric === 'actual_cost' ? `$${formatCost(distributionTotal)}` : formatTokens(distributionTotal) }}</strong>
        </div>
      </div>
      <div class="dashboard-table-wrap">
        <table class="dashboard-data-table">
          <thead>
            <tr>
              <th class="pb-2 text-left">{{ t('admin.dashboard.model') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.requests') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.tokens') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.actual') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.accountCost') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.standard') }}</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="(model, index) in displayModelStats" :key="model.model">
              <tr
                class="dashboard-table-row cursor-pointer hover:bg-gray-50 dark:hover:bg-dark-700/40"
                @click="toggleBreakdown('model', model.model)"
              >
                <td
                  class="max-w-[100px] truncate py-2 font-medium text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
                  :title="displayModelLabel(model.model)"
                >
                  <span class="inline-flex items-center gap-1">
                    <span class="dashboard-color-dot" :style="{ backgroundColor: getChartColor(index) }"></span>
                    <svg v-if="expandedKey === `model-${model.model}`" class="h-3 w-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
                    <svg v-else class="h-3 w-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
                    {{ displayModelLabel(model.model) }}
                  </span>
                </td>
                <td class="py-2 text-right text-gray-600 dark:text-gray-400">
                  {{ formatNumber(model.requests) }}
                </td>
                <td class="py-2 text-right text-gray-600 dark:text-gray-400">
                  {{ formatTokens(model.total_tokens) }}
                </td>
                <td class="py-2 text-right text-green-600 dark:text-green-400">
                  ${{ formatCost(model.actual_cost) }}
                </td>
                <td class="py-2 text-right text-orange-500 dark:text-orange-400">
                  ${{ formatCost(model.account_cost) }}
                </td>
                <td class="py-2 text-right text-gray-400 dark:text-gray-500">
                  ${{ formatCost(model.cost) }}
                </td>
              </tr>
              <tr v-if="expandedKey === `model-${model.model}`">
                <td colspan="6" class="p-0">
                  <UserBreakdownSubTable
                    :items="breakdownItems"
                    :loading="breakdownLoading"
                  />
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
    <div
      v-else-if="activeView === 'model_distribution'"
      class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>

    <div v-else-if="rankingLoading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div
      v-else-if="rankingError"
      class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('admin.dashboard.failedToLoad') }}
    </div>
    <div v-else-if="rankingDisplayItems.length > 0 && rankingChartData" class="dashboard-chart-split">
      <div class="dashboard-doughnut-wrap">
        <Doughnut :data="rankingChartData" :options="rankingDoughnutOptions" />
        <div class="dashboard-doughnut-center">
          <span>{{ t('admin.dashboard.spendingRankingSpend') }}</span>
          <strong>${{ formatCost(rankingDisplayTotal) }}</strong>
        </div>
      </div>
      <div class="dashboard-table-wrap">
        <table class="dashboard-data-table">
          <thead>
            <tr>
              <th class="pb-2 text-left">{{ t('admin.dashboard.spendingRankingUser') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.spendingRankingRequests') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.spendingRankingTokens') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.spendingRankingSpend') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(item, index) in rankingDisplayItems"
              :key="item.isOther ? 'others' : `${item.user_id}-${index}`"
              class="dashboard-table-row transition-colors"
              :class="item.isOther
                ? 'bg-gray-50/70 dark:bg-dark-700/20'
                : 'cursor-pointer hover:bg-gray-50 dark:hover:bg-dark-700/40'"
              @click="item.isOther ? undefined : emit('ranking-click', item)"
            >
              <td class="py-2">
                <div class="flex min-w-0 items-center gap-2">
                  <span v-if="!item.isOther" class="shrink-0 text-[11px] font-semibold text-gray-500 dark:text-gray-400">
                    #{{ index + 1 }}
                  </span>
                  <span v-if="item.isOther" class="shrink-0 text-[11px] font-semibold text-gray-500 dark:text-gray-400">
                    --
                  </span>
                  <span
                    class="block max-w-[140px] truncate font-medium text-gray-900 dark:text-white"
                    :title="getRankingRowLabel(item)"
                  >
                    {{ getRankingRowLabel(item) }}
                  </span>
                </div>
              </td>
              <td class="py-2 text-right text-gray-600 dark:text-gray-400">
                {{ formatNumber(item.requests) }}
              </td>
              <td class="py-2 text-right text-gray-600 dark:text-gray-400">
                {{ formatTokens(item.tokens) }}
              </td>
              <td class="py-2 text-right text-green-600 dark:text-green-400">
                ${{ formatCost(item.actual_cost) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
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
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Doughnut } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import UserBreakdownSubTable from './UserBreakdownSubTable.vue'
import type { ModelStat, UserSpendingRankingItem, UserBreakdownItem } from '@/types'
import { getUserBreakdown } from '@/api/admin/dashboard'
import { displayModelLabel } from '@/utils/modelDisplay'

ChartJS.register(ArcElement, Tooltip, Legend)

const { t } = useI18n()

type DistributionMetric = 'tokens' | 'actual_cost'
type ModelSource = 'requested' | 'upstream' | 'mapping'
type RankingDisplayItem = UserSpendingRankingItem & { isOther?: boolean }
const props = withDefaults(defineProps<{
  modelStats: ModelStat[]
  upstreamModelStats?: ModelStat[]
  mappingModelStats?: ModelStat[]
  source?: ModelSource
  enableRankingView?: boolean
  rankingItems?: UserSpendingRankingItem[]
  rankingTotalActualCost?: number
  rankingTotalRequests?: number
  rankingTotalTokens?: number
  loading?: boolean
  metric?: DistributionMetric
  showSourceToggle?: boolean
  showMetricToggle?: boolean
  rankingLoading?: boolean
  rankingError?: boolean
  startDate?: string
  endDate?: string
  filters?: Record<string, any>
}>(), {
  upstreamModelStats: () => [],
  mappingModelStats: () => [],
  source: 'requested',
  enableRankingView: false,
  rankingItems: () => [],
  rankingTotalActualCost: 0,
  rankingTotalRequests: 0,
  rankingTotalTokens: 0,
  loading: false,
  metric: 'tokens',
  showSourceToggle: false,
  showMetricToggle: false,
  rankingLoading: false,
  rankingError: false
})

const expandedKey = ref<string | null>(null)
const breakdownItems = ref<UserBreakdownItem[]>([])
const breakdownLoading = ref(false)

const toggleBreakdown = async (type: string, id: string) => {
  const key = `${type}-${id}`
  if (expandedKey.value === key) {
    expandedKey.value = null
    return
  }
  expandedKey.value = key
  breakdownLoading.value = true
  breakdownItems.value = []
  try {
    const res = await getUserBreakdown({
      ...props.filters,
      start_date: props.startDate,
      end_date: props.endDate,
      model: id,
      model_source: props.source,
    })
    breakdownItems.value = res.users || []
  } catch {
    breakdownItems.value = []
  } finally {
    breakdownLoading.value = false
  }
}

const emit = defineEmits<{
  'update:metric': [value: DistributionMetric]
  'update:source': [value: ModelSource]
  'ranking-click': [item: UserSpendingRankingItem]
}>()

const enableRankingView = computed(() => props.enableRankingView)
const activeView = ref<'model_distribution' | 'spending_ranking'>('model_distribution')
const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})
const segmentBorderColor = computed(() =>
  isDarkMode.value ? 'rgba(15, 23, 42, 0.82)' : 'rgba(248, 250, 252, 0.9)'
)

const chartColors = [
  '#3b82f6',
  '#06b6d4',
  '#10b981',
  '#f59e0b',
  '#8b5cf6',
  '#14b8a6',
  '#f97316',
  '#6366f1',
  '#84cc16',
  '#eab308',
  '#a855f7',
  '#f43f5e',
  '#64748b'
]

const displayModelStats = computed(() => {
  const sourceStats = props.source === 'upstream'
    ? props.upstreamModelStats
    : props.source === 'mapping'
      ? props.mappingModelStats
      : props.modelStats
  if (!sourceStats?.length) return []

  const metricKey = props.metric === 'actual_cost' ? 'actual_cost' : 'total_tokens'
  return [...sourceStats].sort((a, b) => b[metricKey] - a[metricKey])
})

const chartData = computed(() => {
  if (!displayModelStats.value.length) return null

  return {
    labels: displayModelStats.value.map((m) => displayModelLabel(m.model)),
    datasets: [
      {
        data: displayModelStats.value.map((m) => props.metric === 'actual_cost' ? m.actual_cost : m.total_tokens),
        backgroundColor: chartColors.slice(0, displayModelStats.value.length),
        borderColor: segmentBorderColor.value,
        borderWidth: 1,
        hoverBorderWidth: 1,
        hoverOffset: 3
      }
    ]
  }
})

const rankingChartData = computed(() => {
  if (!props.rankingItems?.length) return null

  const labels = props.rankingItems.map((item, index) => `#${index + 1} ${getRankingUserLabel(item)}`)
  const data = props.rankingItems.map((item) => item.actual_cost)
  const backgroundColor = chartColors.slice(0, props.rankingItems.length)

  if (otherRankingItem.value) {
    labels.push(t('admin.dashboard.spendingRankingOther'))
    data.push(otherRankingItem.value.actual_cost)
    backgroundColor.push('#94a3b8')
  }

  return {
    labels,
    datasets: [
      {
        data,
        backgroundColor,
        borderColor: segmentBorderColor.value,
        borderWidth: 1,
        hoverBorderWidth: 1,
        hoverOffset: 3
      }
    ]
  }
})

const otherRankingItem = computed<RankingDisplayItem | null>(() => {
  if (!props.rankingItems?.length) return null

  const rankedActualCost = props.rankingItems.reduce((sum, item) => sum + item.actual_cost, 0)
  const rankedRequests = props.rankingItems.reduce((sum, item) => sum + item.requests, 0)
  const rankedTokens = props.rankingItems.reduce((sum, item) => sum + item.tokens, 0)

  const otherActualCost = Math.max((props.rankingTotalActualCost || 0) - rankedActualCost, 0)
  const otherRequests = Math.max((props.rankingTotalRequests || 0) - rankedRequests, 0)
  const otherTokens = Math.max((props.rankingTotalTokens || 0) - rankedTokens, 0)

  if (otherActualCost <= 0.000001 && otherRequests <= 0 && otherTokens <= 0) return null

  return {
    user_id: 0,
    email: '',
    actual_cost: otherActualCost,
    requests: otherRequests,
    tokens: otherTokens,
    isOther: true
  }
})

const rankingDisplayItems = computed<RankingDisplayItem[]>(() => {
  if (!props.rankingItems?.length) return []
  return otherRankingItem.value
    ? [...props.rankingItems, otherRankingItem.value]
    : [...props.rankingItems]
})

const distributionTotal = computed(() => {
  return displayModelStats.value.reduce((sum, model) => {
    return sum + (props.metric === 'actual_cost' ? model.actual_cost : model.total_tokens)
  }, 0)
})

const rankingDisplayTotal = computed(() => {
  return rankingDisplayItems.value.reduce((sum, item) => sum + item.actual_cost, 0)
})

const getChartColor = (index: number): string => chartColors[index % chartColors.length]

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  cutout: '68%',
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      backgroundColor: 'rgba(8, 13, 26, 0.96)',
      borderColor: 'rgba(59, 130, 246, 0.36)',
      borderWidth: 1,
      padding: 10,
      titleColor: '#f8fafc',
      bodyColor: '#e5e7eb',
      callbacks: {
        label: (context: any) => {
          const value = context.raw as number
          const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0)
          const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0'
          const formattedValue = props.metric === 'actual_cost'
            ? `$${formatCost(value)}`
            : formatTokens(value)
          return `${context.label}: ${formattedValue} (${percentage}%)`
        }
      }
    }
  }
}))

const rankingDoughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  cutout: '68%',
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      backgroundColor: 'rgba(8, 13, 26, 0.96)',
      borderColor: 'rgba(59, 130, 246, 0.36)',
      borderWidth: 1,
      padding: 10,
      titleColor: '#f8fafc',
      bodyColor: '#e5e7eb',
      callbacks: {
        label: (context: any) => {
          const value = context.raw as number
          const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0)
          const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0'
          return `${context.label}: $${formatCost(value)} (${percentage}%)`
        }
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

const formatNumber = (value: number): string => {
  return value.toLocaleString()
}

const getRankingUserLabel = (item: UserSpendingRankingItem): string => {
  if (item.email) return item.email
  return t('admin.redeem.userPrefix', { id: item.user_id })
}

const getRankingRowLabel = (item: RankingDisplayItem): string => {
  if (item.isOther) return t('admin.dashboard.spendingRankingOther')
  return getRankingUserLabel(item)
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
.dashboard-chart-split {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.dashboard-doughnut-wrap {
  position: relative;
  width: 12rem;
  height: 12rem;
  flex: 0 0 12rem;
}

.dashboard-doughnut-center {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 7.25rem;
  transform: translate(-50%, -50%);
  text-align: center;
  pointer-events: none;
}

.dashboard-doughnut-center span {
  display: block;
  color: #64748b;
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1.15;
}

.dark .dashboard-doughnut-center span {
  color: #94a3b8;
}

.dashboard-doughnut-center strong {
  display: block;
  margin-top: 0.25rem;
  overflow: hidden;
  color: #0f172a;
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dark .dashboard-doughnut-center strong {
  color: #f8fafc;
}

.dashboard-table-wrap {
  max-height: 12.25rem;
  flex: 1 1 auto;
  overflow: auto;
  border-radius: 0.625rem;
}

.dashboard-data-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 0.75rem;
}

.dashboard-data-table th {
  padding-bottom: 0.5rem;
  color: #64748b;
  font-weight: 700;
  text-align: right;
}

.dashboard-data-table th:first-child {
  text-align: left;
}

.dark .dashboard-data-table th {
  color: #94a3b8;
}

.dashboard-table-row {
  border-top: 1px solid rgba(148, 163, 184, 0.14);
  transition: background-color 0.16s ease, color 0.16s ease;
}

.dark .dashboard-table-row {
  border-top-color: rgba(51, 65, 85, 0.78);
}

.dashboard-color-dot {
  width: 0.5rem;
  height: 0.5rem;
  flex: 0 0 auto;
  border-radius: 999px;
  box-shadow: 0 0 0 3px rgba(148, 163, 184, 0.12);
}

.dashboard-title-icon {
  display: inline-flex;
  width: 1.75rem;
  height: 1.75rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(16, 185, 129, 0.22);
  border-radius: 0.5rem;
  background: rgba(16, 185, 129, 0.09);
  color: #059669;
}

.dark .dashboard-title-icon {
  border-color: rgba(52, 211, 153, 0.28);
  background: rgba(16, 185, 129, 0.14);
  color: #86efac;
}

.dashboard-segmented {
  display: inline-flex;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 0.5rem;
  background: rgba(248, 250, 252, 0.78);
  padding: 0.125rem;
}

.dark .dashboard-segmented {
  border-color: rgba(51, 65, 85, 0.86);
  background: rgba(15, 23, 42, 0.76);
}

.dashboard-segmented-button {
  border-radius: 0.375rem;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 700;
  transition: background-color 0.16s ease, color 0.16s ease;
}

.dashboard-segmented-button--active {
  background: #ffffff;
  color: #0f172a;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.08);
}

.dark .dashboard-segmented-button--active {
  background: rgba(30, 41, 59, 0.96);
  color: #f8fafc;
  box-shadow: none;
}

.dashboard-segmented-button--idle {
  color: #64748b;
}

.dashboard-segmented-button--idle:hover {
  color: #0f172a;
}

.dark .dashboard-segmented-button--idle {
  color: #94a3b8;
}

.dark .dashboard-segmented-button--idle:hover {
  color: #f8fafc;
}

@media (max-width: 767px) {
  .dashboard-chart-split {
    align-items: stretch;
    flex-direction: column;
  }

  .dashboard-doughnut-wrap {
    align-self: center;
  }
}
</style>
