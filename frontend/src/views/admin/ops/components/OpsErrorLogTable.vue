<template>
  <div class="flex h-full min-h-0 flex-col overflow-hidden" :class="flat ? '' : 'card'">
    <DataTable
      :columns="columns"
      :data="rows"
      :loading="loading"
      :server-side-sort="serverSideSort"
      default-sort-key="created_at"
      default-sort-order="desc"
      @sort="onSort"
    >
      <template #cell-user="{ row }">
        <div v-if="row.user_id" class="text-sm">
          <button
            v-if="userClickable && row.user_email"
            type="button"
            class="font-medium text-primary-600 underline decoration-dashed underline-offset-2 hover:text-primary-700 dark:text-primary-400"
            :title="t('admin.usage.clickToViewBalance')"
            @click.stop="emit('userClick', row.user_id, row.user_email)"
          >
            {{ row.user_email }}
          </button>
          <span v-else class="font-medium text-gray-900 dark:text-white">{{ row.user_email || '-' }}</span>
          <span class="ml-1 text-gray-500 dark:text-gray-400">#{{ row.user_id }}</span>
        </div>
        <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
      </template>

      <template #cell-api_key="{ row }">
        <div v-if="row.api_key_id || row.api_key_name" class="text-sm">
          <span class="text-gray-900 dark:text-white">{{ row.api_key_name || `#${row.api_key_id}` }}</span>
          <span v-if="row.api_key_deleted" class="ml-1 rounded bg-red-100 px-1 py-0.5 text-[10px] text-red-700 dark:bg-red-500/20 dark:text-red-300">
            {{ t('usage.errors.keyDeleted') }}
          </span>
        </div>
        <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
      </template>

      <template #cell-account="{ row }">
        <span v-if="row.account_id" class="text-sm text-gray-900 dark:text-white" :title="formatAccountTooltip(row)">
          {{ formatAccountLabel(row) }}
        </span>
        <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
      </template>

      <template #cell-platform="{ row }">
        <span class="text-sm text-gray-900 dark:text-white">{{ row.platform || '-' }}</span>
      </template>

      <template #cell-model="{ row }">
        <div v-if="hasModelMapping(row)" class="max-w-72 space-y-0.5 text-xs">
          <div class="break-all font-medium text-gray-900 dark:text-white">{{ displayModelLabel(row.requested_model) }}</div>
          <div class="break-all text-gray-500 dark:text-gray-400">↳ {{ displayModelLabel(row.upstream_model) }}</div>
        </div>
        <span v-else class="text-sm font-medium text-gray-900 dark:text-white">{{ displayModel(row) || '-' }}</span>
      </template>

      <template #cell-endpoint="{ row }">
        <div class="max-w-80 space-y-1 text-xs" :title="formatEndpointTooltip(row)">
          <div class="break-all text-gray-700 dark:text-gray-300">
            <span class="font-medium text-gray-500 dark:text-gray-400">{{ t('usage.inbound') }}:</span>
            <span class="ml-1">{{ row.inbound_endpoint?.trim() || '-' }}</span>
          </div>
          <div v-if="row.upstream_endpoint" class="break-all text-gray-700 dark:text-gray-300">
            <span class="font-medium text-gray-500 dark:text-gray-400">{{ t('usage.upstream') }}:</span>
            <span class="ml-1">{{ row.upstream_endpoint }}</span>
          </div>
        </div>
      </template>

      <template #cell-group="{ row }">
        <span v-if="row.group_id" class="inline-flex max-w-40 truncate rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-700 dark:bg-dark-700 dark:text-gray-200" :title="`#${row.group_id}`">
          {{ row.group_name || `#${row.group_id}` }}
        </span>
        <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
      </template>

      <template #cell-type="{ row }">
        <span class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium" :class="getTypeBadge(row).className">
          {{ getTypeBadge(row).label }}
        </span>
      </template>

      <template #cell-category="{ row }">
        <span class="text-sm text-gray-900 dark:text-white">
          {{ t(`usage.errors.categories.${mapAdminErrorCategory(row.phase, row.type)}`) }}
        </span>
      </template>

      <template #cell-status="{ row }">
        <div class="flex flex-wrap items-center gap-1.5">
          <span class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium" :class="statusCodeBadgeClass(row.status_code)">
            {{ row.status_code }}
          </span>
          <span v-if="row.severity" class="rounded px-1.5 py-0.5 text-[10px] font-medium" :class="getSeverityClass(row.severity)">
            {{ row.severity }}
          </span>
          <span v-if="row.request_type" class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] text-gray-700 dark:bg-dark-700 dark:text-gray-200">
            {{ formatRequestType(row.request_type) }}
          </span>
        </div>
      </template>

      <template #cell-message="{ row }">
        <span class="block max-w-72 truncate text-sm text-gray-600 dark:text-gray-400" :title="row.message">
          {{ formatSmartMessage(row.message) || '-' }}
        </span>
      </template>

      <template #cell-created_at="{ row }">
        <span class="whitespace-nowrap text-sm text-gray-600 dark:text-gray-400" :title="row.request_id || row.client_request_id">
          {{ formatDateTime(row.created_at) }}
        </span>
      </template>

      <template #cell-user_agent="{ row }">
        <span class="block max-w-80 truncate text-sm text-gray-600 dark:text-gray-400" :title="row.user_agent">{{ row.user_agent || '-' }}</span>
      </template>

      <template #cell-client_ip="{ row }">
        <span class="font-mono text-sm text-gray-600 dark:text-gray-400">{{ row.client_ip || '-' }}</span>
      </template>

      <template #cell-actions="{ row }">
        <button
          type="button"
          class="btn btn-secondary px-2 py-1 text-xs"
          @click.stop="emit('openErrorDetail', row.id)"
        >
          {{ t('admin.ops.errorLog.details') }}
        </button>
      </template>

      <template #empty><EmptyState :message="t('admin.ops.errorLog.noErrors')" /></template>
    </DataTable>

    <Pagination
      v-if="total > 0"
      class="flex-shrink-0"
      :total="total"
      :page="page"
      :page-size="pageSize"
      @update:page="emit('update:page', $event)"
      @update:pageSize="emit('update:pageSize', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Pagination from '@/components/common/Pagination.vue'
import type { Column } from '@/components/common/types'
import type { OpsErrorLog } from '@/api/admin/ops'
import { displayModelLabel } from '@/utils/modelDisplay'
import { mapAdminErrorCategory, mapAdminErrorSortKey, statusCodeBadgeClass } from '@/utils/errorBadges'
import { formatDateTime, getSeverityClass } from '../utils/opsFormatters'

const props = defineProps<{
  rows: OpsErrorLog[]
  total: number
  loading: boolean
  page: number
  pageSize: number
  userClickable?: boolean
  visibleColumnKeys?: string[]
  flat?: boolean
  serverSideSort?: boolean
}>()

const emit = defineEmits<{
  (event: 'openErrorDetail', id: number): void
  (event: 'update:page', value: number): void
  (event: 'update:pageSize', value: number): void
  (event: 'sort', sortBy: string, sortOrder: 'asc' | 'desc'): void
  (event: 'userClick', userId: number, email?: string): void
}>()

const { t } = useI18n()
const allColumns = computed<Column[]>(() => [
  { key: 'user', label: t('admin.ops.errorLog.user') },
  { key: 'api_key', label: t('usage.errors.keyName') },
  { key: 'account', label: t('admin.ops.errorLog.account') },
  { key: 'platform', label: t('admin.ops.errorLog.platform') },
  { key: 'model', label: t('admin.ops.errorLog.model'), sortable: props.serverSideSort },
  { key: 'endpoint', label: t('admin.ops.errorLog.endpoint') },
  { key: 'group', label: t('admin.ops.errorLog.group') },
  { key: 'type', label: t('admin.ops.errorLog.type') },
  { key: 'category', label: t('usage.errors.category') },
  { key: 'status', label: t('admin.ops.errorLog.status'), sortable: props.serverSideSort },
  { key: 'message', label: t('admin.ops.errorLog.message') },
  { key: 'created_at', label: t('admin.ops.errorLog.time'), sortable: props.serverSideSort },
  { key: 'user_agent', label: t('usage.userAgent') },
  { key: 'client_ip', label: t('admin.usage.ipAddress') },
  { key: 'actions', label: t('admin.ops.errorLog.action') }
])
const columns = computed(() => props.visibleColumnKeys
  ? allColumns.value.filter((column) => props.visibleColumnKeys!.includes(column.key))
  : allColumns.value)

const isUpstreamRow = (row: OpsErrorLog) => row.phase === 'upstream' && row.error_owner === 'provider'
const hasModelMapping = (row: OpsErrorLog) => Boolean(row.requested_model && row.upstream_model && row.requested_model !== row.upstream_model)
const displayModel = (row: OpsErrorLog) => displayModelLabel(row.upstream_model || row.requested_model || row.model || '', '')
const formatEndpointTooltip = (row: OpsErrorLog) => [row.inbound_endpoint, row.upstream_endpoint].filter(Boolean).join('\n')
const formatAccountLabel = (row: OpsErrorLog) => row.account_name || row.account_notes || `#${row.account_id}`
const formatAccountTooltip = (row: OpsErrorLog) => [row.account_name, row.account_notes, row.account_id ? `#${row.account_id}` : ''].filter(Boolean).join('\n')

const getTypeBadge = (row: OpsErrorLog) => {
  if (isUpstreamRow(row)) return { label: t('admin.ops.errorLog.typeUpstream'), className: 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300' }
  if (row.phase === 'request') return { label: t('admin.ops.errorLog.typeRequest'), className: 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300' }
  if (row.phase === 'auth') return { label: t('admin.ops.errorLog.typeAuth'), className: 'bg-orange-100 text-orange-700 dark:bg-orange-500/20 dark:text-orange-300' }
  if (row.phase === 'routing') return { label: t('admin.ops.errorLog.typeRouting'), className: 'bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300' }
  if (row.phase === 'internal') return { label: t('admin.ops.errorLog.typeInternal'), className: 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-200' }
  return { label: row.phase || row.error_owner || t('common.unknown'), className: 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-200' }
}

const formatRequestType = (value: number) => {
  if (value === 1) return t('admin.ops.errorLog.requestTypeSync')
  if (value === 2) return t('admin.ops.errorLog.requestTypeStream')
  if (value === 3) return t('admin.ops.errorLog.requestTypeWs')
  return ''
}

const formatSmartMessage = (message: string) => {
  if (!message) return ''
  if (message.startsWith('{') || message.startsWith('[')) {
    try {
      const value = JSON.parse(message)
      return String(value?.error?.message || value?.message || value?.detail || message)
    } catch {
      // Keep the stored text when it is not valid JSON.
    }
  }
  return message.length > 200 ? `${message.slice(0, 200)}...` : message
}

const onSort = (key: string, order: 'asc' | 'desc') => emit('sort', mapAdminErrorSortKey(key), order)
</script>
