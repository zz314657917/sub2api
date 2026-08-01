<template>
  <section class="card overflow-hidden" data-testid="user-token-ranking">
    <div class="flex flex-wrap items-center justify-end gap-3 border-b border-gray-200 px-4 py-3 dark:border-dark-700 sm:px-6">
      <div class="flex items-center gap-3">
        <span v-if="!loading && items.length" class="text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.usage.tokenRanking.userCount', { count: items.length }) }}
        </span>
        <div class="w-40 md:hidden">
          <Select v-model="sortBy" :options="sortOptions" @change="load" />
        </div>
        <div class="w-28">
          <Select v-model="limit" :options="limitOptions" @change="load" />
        </div>
      </div>
    </div>

    <div v-if="loading" class="flex min-h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="loadFailed" class="px-4 py-12 text-center text-sm text-red-600 dark:text-red-400">
      {{ t('admin.usage.tokenRanking.loadFailed') }}
    </div>
    <div v-else-if="items.length === 0" class="px-4 py-12 text-center text-sm text-gray-500 dark:text-gray-400">
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>

    <template v-else>
      <div class="hidden overflow-x-auto md:block">
        <table class="w-full min-w-[860px] divide-y divide-gray-200 dark:divide-dark-700">
          <thead class="bg-gray-50 dark:bg-dark-800">
            <tr>
              <th class="w-16 px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 sm:px-6">#</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.usage.tokenRanking.columns.user') }}</th>
              <th v-for="column in sortableColumns" :key="column.key" class="whitespace-nowrap px-4 py-3 text-right">
                <button
                  type="button"
                  class="text-xs font-medium text-gray-500 hover:text-primary-600 dark:text-gray-400 dark:hover:text-primary-400"
                  :class="sortBy === column.key ? 'text-primary-600 dark:text-primary-400' : ''"
                  @click="setSort(column.key)"
                >
                  {{ t(column.label) }}<span v-if="sortBy === column.key" class="ml-1" aria-hidden="true">↓</span>
                </button>
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white dark:divide-dark-700 dark:bg-dark-900">
            <tr
              v-for="(item, index) in items"
              :key="item.user_id"
              class="cursor-pointer transition-colors hover:bg-gray-50 dark:hover:bg-dark-800"
              @click="emit('select-user', item.user_id, item.email)"
            >
              <td class="px-4 py-3 sm:px-6"><span :class="rankClass(index)">{{ index + 1 }}</span></td>
              <td class="max-w-64 px-4 py-3 text-sm font-medium text-gray-800 dark:text-gray-200">
                <span class="block truncate" :title="item.email">{{ userLabel(item) }}</span>
                <span class="text-xs font-normal text-gray-400">#{{ item.user_id }}</span>
              </td>
              <td class="ranking-number">{{ item.requests.toLocaleString() }}</td>
              <td class="ranking-number">{{ formatCompactNumber(item.input_tokens) }}</td>
              <td class="ranking-number">{{ formatCompactNumber(item.output_tokens) }}</td>
              <td class="ranking-number">{{ formatCompactNumber(item.cache_tokens) }}</td>
              <td class="ranking-number font-semibold text-gray-900 dark:text-gray-100">{{ formatCompactNumber(item.total_tokens) }}</td>
              <td class="ranking-number font-semibold text-green-600 dark:text-green-400">${{ formatCostFixed(item.actual_cost, 4) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="space-y-3 p-3 md:hidden">
        <button
          v-for="(item, index) in items"
          :key="item.user_id"
          type="button"
          class="w-full rounded-lg border border-gray-200 bg-white p-4 text-left dark:border-dark-700 dark:bg-dark-900"
          @click="emit('select-user', item.user_id, item.email)"
        >
          <div class="flex items-center gap-3">
            <span :class="rankClass(index)">{{ index + 1 }}</span>
            <div class="min-w-0 flex-1">
              <div class="truncate text-sm font-semibold text-gray-900 dark:text-gray-100">{{ userLabel(item) }}</div>
              <div class="text-xs text-gray-400">#{{ item.user_id }}</div>
            </div>
            <div class="text-right">
              <div class="text-sm font-semibold text-gray-900 dark:text-gray-100">{{ formatCompactNumber(item.total_tokens) }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.usage.tokenRanking.columns.totalTokens') }}</div>
            </div>
          </div>
          <div class="mt-3 grid grid-cols-2 gap-2 border-t border-gray-100 pt-3 text-xs dark:border-dark-700">
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.usage.tokenRanking.columns.requests') }}: {{ item.requests.toLocaleString() }}</span>
            <span class="text-right font-medium text-green-600 dark:text-green-400">${{ formatCostFixed(item.actual_cost, 4) }}</span>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.usage.tokenRanking.columns.inputTokens') }}: {{ formatCompactNumber(item.input_tokens) }}</span>
            <span class="text-right text-gray-500 dark:text-gray-400">{{ t('admin.usage.tokenRanking.columns.outputTokens') }}: {{ formatCompactNumber(item.output_tokens) }}</span>
          </div>
        </button>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getUserBreakdown, type UserBreakdownParams } from '@/api/admin/dashboard'
import { formatCompactNumber, formatCostFixed } from '@/utils/format'
import type { UserBreakdownItem } from '@/types'
import Select from '@/components/common/Select.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const props = defineProps<{
  startDate: string
  endDate: string
  filters: Record<string, unknown>
  model?: string
}>()

const emit = defineEmits<{ (event: 'select-user', userId: number, email: string): void }>()
const { t } = useI18n()

type SortKey = NonNullable<UserBreakdownParams['sort_by']>
const sortableColumns: { key: SortKey; label: string }[] = [
  { key: 'requests', label: 'admin.usage.tokenRanking.columns.requests' },
  { key: 'input_tokens', label: 'admin.usage.tokenRanking.columns.inputTokens' },
  { key: 'output_tokens', label: 'admin.usage.tokenRanking.columns.outputTokens' },
  { key: 'cache_tokens', label: 'admin.usage.tokenRanking.columns.cacheTokens' },
  { key: 'total_tokens', label: 'admin.usage.tokenRanking.columns.totalTokens' },
  { key: 'actual_cost', label: 'admin.usage.tokenRanking.columns.cost' }
]
const limitOptions = [20, 50, 100, 200].map((value) => ({ value, label: `Top ${value}` }))
const sortOptions = sortableColumns.map((column) => ({ value: column.key, label: t(column.label) }))
const rankClasses = [
  'inline-flex h-7 w-7 items-center justify-center rounded-full bg-amber-100 text-xs font-semibold text-amber-700 dark:bg-amber-500/20 dark:text-amber-400',
  'inline-flex h-7 w-7 items-center justify-center rounded-full bg-gray-200 text-xs font-semibold text-gray-600 dark:bg-gray-500/20 dark:text-gray-300',
  'inline-flex h-7 w-7 items-center justify-center rounded-full bg-orange-100 text-xs font-semibold text-orange-700 dark:bg-orange-500/20 dark:text-orange-400'
]

const items = ref<UserBreakdownItem[]>([])
const loading = ref(false)
const loadFailed = ref(false)
const sortBy = ref<SortKey>('total_tokens')
const limit = ref(50)
let requestSequence = 0

const rankClass = (index: number) => rankClasses[index] || 'inline-flex h-7 w-7 items-center justify-center text-xs tabular-nums text-gray-400'
const userLabel = (item: UserBreakdownItem) => item.email || `User #${item.user_id}`

const setSort = (key: SortKey) => {
  if (sortBy.value === key) return
  sortBy.value = key
  void load()
}

const load = async () => {
  const sequence = ++requestSequence
  loading.value = true
  loadFailed.value = false
  try {
    const response = await getUserBreakdown({
      ...props.filters,
      start_date: props.startDate,
      end_date: props.endDate,
      model: props.model || undefined,
      sort_by: sortBy.value,
      limit: limit.value
    })
    if (sequence !== requestSequence) return
    items.value = response.users || []
  } catch {
    if (sequence !== requestSequence) return
    items.value = []
    loadFailed.value = true
  } finally {
    if (sequence === requestSequence) loading.value = false
  }
}

watch(
  () => [props.startDate, props.endDate, props.model, JSON.stringify(props.filters)],
  () => void load(),
  { immediate: true }
)

defineExpose({ reload: load })
</script>

<style scoped>
.ranking-number {
  white-space: nowrap;
  padding: 0.75rem 1rem;
  text-align: right;
  font-size: 0.875rem;
  font-variant-numeric: tabular-nums;
  color: rgb(107 114 128);
}
</style>
