<template>
  <div class="flex min-h-0 flex-1 flex-col" data-test="user-error-requests-table">
    <DataTable
      :columns="columns"
      :data="rows"
      :loading="loading"
      server-side-sort
      default-sort-key="created_at"
      default-sort-order="desc"
      @sort="onSort"
    >
      <template #cell-key_name="{ row }">
        <div class="text-sm">
          <span class="text-gray-900 dark:text-white">{{ row.key_name || '-' }}</span>
          <span v-if="row.key_deleted" class="ml-1 inline-flex rounded bg-rose-100 px-1 py-px text-[10px] font-medium text-rose-600 dark:bg-rose-500/20 dark:text-rose-400">
            {{ t('usage.errors.keyDeleted') }}
          </span>
        </div>
      </template>

      <template #cell-model="{ row }">
        <span class="text-sm font-medium text-gray-900 dark:text-white">{{ row.model || '-' }}</span>
      </template>

      <template #cell-endpoint="{ row }">
        <span class="block max-w-[320px] break-all text-xs text-gray-700 dark:text-gray-300">
          {{ row.inbound_endpoint?.trim() || '-' }}
        </span>
      </template>

      <template #cell-platform="{ row }">
        <span class="text-sm text-gray-900 dark:text-white">{{ row.platform || '-' }}</span>
      </template>

      <template #cell-category="{ row }">
        <span class="text-sm text-gray-900 dark:text-white">{{ categoryLabel(row.category) }}</span>
      </template>

      <template #cell-status="{ row }">
        <span class="inline-flex rounded px-2 py-0.5 text-xs font-medium" :class="statusCodeBadgeClass(row.status_code)">
          {{ row.status_code || '-' }}
        </span>
      </template>

      <template #cell-message="{ row }">
        <span v-if="row.message" class="block max-w-[300px] truncate text-sm text-gray-600 dark:text-gray-400" :title="row.message">
          {{ row.message }}
        </span>
        <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
      </template>

      <template #cell-created_at="{ row }">
        <span class="text-sm text-gray-600 dark:text-gray-400">{{ formatDateTime(row.created_at) }}</span>
      </template>

      <template #cell-actions="{ row }">
        <button
          type="button"
          class="inline-flex h-8 w-8 items-center justify-center rounded-md text-gray-500 transition hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-white"
          :title="t('usage.details')"
          :aria-label="t('usage.details')"
          data-test="user-error-detail-button"
          @click="openDetail(row)"
        >
          <Icon name="eye" size="sm" />
        </button>
      </template>

      <template #empty>
        <EmptyState :message="t('usage.errors.empty')" />
      </template>
    </DataTable>

    <UserErrorDetailModal v-model:show="showDetail" :error-id="selectedId" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import UserErrorDetailModal from '@/components/user/UserErrorDetailModal.vue'
import { formatDateTime } from '@/utils/format'
import { mapUserErrorSortKey, statusCodeBadgeClass } from '@/utils/errorBadges'
import type { UserErrorRequest } from '@/types'
import type { Column } from '@/components/common/types'

const props = defineProps<{
  rows: UserErrorRequest[]
  loading: boolean
  visibleColumnKeys?: string[]
}>()

const emit = defineEmits<{
  (event: 'sort', sortBy: 'created_at' | 'model' | 'status_code', sortOrder: 'asc' | 'desc'): void
}>()

const { t, te } = useI18n()

const allColumns = computed<Column[]>(() => [
  { key: 'key_name', label: t('usage.errors.keyName') },
  { key: 'model', label: t('usage.errors.model'), sortable: true },
  { key: 'endpoint', label: t('usage.errors.endpoint') },
  { key: 'platform', label: t('usage.errors.platform') },
  { key: 'category', label: t('usage.errors.category') },
  { key: 'status', label: t('usage.errors.status'), sortable: true },
  { key: 'message', label: t('usage.errors.message') },
  { key: 'created_at', label: t('usage.errors.time'), sortable: true },
])

const columns = computed<Column[]>(() =>
  [
    ...(props.visibleColumnKeys
      ? allColumns.value.filter((column) => props.visibleColumnKeys?.includes(column.key))
      : allColumns.value),
    { key: 'actions', label: t('usage.details') },
  ]
)

const categoryLabel = (category: string): string => {
  const key = `usage.errors.categories.${category}`
  return te(key) ? t(key) : t('usage.errors.categories.other')
}

const showDetail = ref(false)
const selectedId = ref<number | null>(null)

const openDetail = (row: UserErrorRequest): void => {
  selectedId.value = row.id
  showDetail.value = true
}

const onSort = (key: string, order: 'asc' | 'desc'): void => {
  emit('sort', mapUserErrorSortKey(key), order)
}
</script>
