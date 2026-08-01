<template>
  <BaseDialog
    :show="show"
    :title="t('usage.errors.detail.title')"
    width="wide"
    @close="emit('update:show', false)"
  >
    <div v-if="loading" class="flex justify-center py-10">
      <svg class="h-7 w-7 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 0 1 8-8V0C5.373 0 0 5.373 0 12h4Zm2 5.291A7.962 7.962 0 0 1 4 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647Z" />
      </svg>
    </div>

    <div v-else-if="loadError" class="py-8 text-center text-sm text-red-500">
      {{ t('usage.errors.detail.loadFailed') }}
    </div>

    <div v-else-if="detail" class="space-y-5 text-sm" data-test="user-error-detail">
      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <span class="font-medium text-gray-500 dark:text-dark-400">{{ t('usage.errors.time') }}</span>
          <p class="mt-1 text-gray-900 dark:text-dark-100">{{ formatDateTime(detail.created_at) }}</p>
        </div>
        <div>
          <span class="font-medium text-gray-500 dark:text-dark-400">{{ t('usage.errors.keyName') }}</span>
          <p class="mt-1 text-gray-900 dark:text-dark-100">
            {{ detail.key_name || '-' }}
            <span v-if="detail.key_deleted" class="ml-1 text-xs text-rose-600 dark:text-rose-400">
              {{ t('usage.errors.keyDeleted') }}
            </span>
          </p>
        </div>
        <div>
          <span class="font-medium text-gray-500 dark:text-dark-400">{{ t('usage.errors.model') }}</span>
          <p class="mt-1 break-all text-gray-900 dark:text-dark-100">{{ detail.model || '-' }}</p>
        </div>
        <div>
          <span class="font-medium text-gray-500 dark:text-dark-400">{{ t('usage.errors.status') }}</span>
          <p class="mt-1">
            <span class="inline-flex rounded px-2 py-0.5 text-xs font-medium" :class="statusCodeBadgeClass(detail.status_code)">
              {{ detail.status_code || '-' }}
            </span>
          </p>
        </div>
        <div class="sm:col-span-2">
          <span class="font-medium text-gray-500 dark:text-dark-400">{{ t('usage.errors.endpoint') }}</span>
          <p class="mt-1 break-all text-gray-900 dark:text-dark-100">{{ detail.inbound_endpoint || '-' }}</p>
        </div>
        <div>
          <span class="font-medium text-gray-500 dark:text-dark-400">{{ t('usage.errors.category') }}</span>
          <p class="mt-1 text-gray-900 dark:text-dark-100">{{ categoryLabel(detail.category) }}</p>
        </div>
        <div>
          <span class="font-medium text-gray-500 dark:text-dark-400">{{ t('usage.errors.platform') }}</span>
          <p class="mt-1 text-gray-900 dark:text-dark-100">{{ detail.platform || '-' }}</p>
        </div>
        <div v-if="detail.upstream_status_code != null">
          <span class="font-medium text-gray-500 dark:text-dark-400">{{ t('usage.errors.detail.upstreamStatus') }}</span>
          <p class="mt-1 text-gray-900 dark:text-dark-100">{{ detail.upstream_status_code }}</p>
        </div>
      </div>

      <div v-if="detail.message">
        <span class="font-medium text-gray-500 dark:text-dark-400">{{ t('usage.errors.message') }}</span>
        <p class="mt-1 break-all text-gray-900 dark:text-dark-100">{{ detail.message }}</p>
      </div>

      <div v-if="detail.error_body">
        <span class="font-medium text-gray-500 dark:text-dark-400">{{ t('usage.errors.detail.responseBody') }}</span>
        <pre class="mt-1 max-h-[40vh] overflow-auto whitespace-pre-wrap break-all rounded-md border border-gray-200 bg-gray-50 p-3 text-xs text-gray-800 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-200">{{ detail.error_body }}</pre>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { usageAPI } from '@/api'
import { formatDateTime } from '@/utils/format'
import { statusCodeBadgeClass } from '@/utils/errorBadges'
import type { UserErrorRequestDetail } from '@/types'

const props = defineProps<{
  show: boolean
  errorId: number | null
}>()

const emit = defineEmits<{
  (event: 'update:show', value: boolean): void
}>()

const { t, te } = useI18n()
const loading = ref(false)
const loadError = ref(false)
const detail = ref<UserErrorRequestDetail | null>(null)
let requestSequence = 0

const categoryLabel = (category: string): string => {
  const key = `usage.errors.categories.${category}`
  return te(key) ? t(key) : t('usage.errors.categories.other')
}

async function fetchDetail(id: number): Promise<void> {
  const sequence = ++requestSequence
  loading.value = true
  loadError.value = false
  detail.value = null
  try {
    const response = await usageAPI.getMyErrorDetail(id)
    if (sequence === requestSequence) detail.value = response
  } catch (error) {
    if (sequence !== requestSequence) return
    console.error('[UserErrorDetailModal] Failed to load error detail:', error)
    loadError.value = true
  } finally {
    if (sequence === requestSequence) loading.value = false
  }
}

watch(
  () => [props.show, props.errorId] as const,
  ([show, id]) => {
    if (show && id != null) {
      void fetchDetail(id)
      return
    }
    requestSequence += 1
    detail.value = null
    loadError.value = false
    loading.value = false
  }
)
</script>
