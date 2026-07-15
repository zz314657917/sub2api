<template>
  <div class="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-lg bg-primary-50 p-3 dark:bg-primary-900/20">
    <div class="flex flex-wrap items-center gap-2">
      <span v-if="selectedIds.length > 0" class="text-sm font-medium text-primary-900 dark:text-primary-100">
        {{ t('admin.accounts.bulkActions.selected', { count: selectedIds.length }) }}
      </span>
      <span v-else class="text-sm font-medium text-primary-900 dark:text-primary-100">
        {{ t('admin.accounts.bulkEdit.title') }}
      </span>
      <template v-if="selectedIds.length > 0">
        <button
          :disabled="loading"
          class="text-xs font-medium text-primary-700 hover:text-primary-800 disabled:cursor-not-allowed disabled:opacity-50 dark:text-primary-300 dark:hover:text-primary-200"
          @click="$emit('select-page')"
        >
          {{ t('admin.accounts.bulkActions.selectCurrentPage') }}
        </button>
        <span class="text-gray-300 dark:text-primary-800">•</span>
        <button
          :disabled="loading"
          class="text-xs font-medium text-primary-700 hover:text-primary-800 disabled:cursor-not-allowed disabled:opacity-50 dark:text-primary-300 dark:hover:text-primary-200"
          @click="$emit('clear')"
        >
          {{ t('admin.accounts.bulkActions.clear') }}
        </button>
      </template>
    </div>
    <div class="flex flex-wrap gap-2">
      <template v-if="selectedIds.length > 0">
        <button v-if="showDeleteAction" :disabled="loading" class="btn btn-danger btn-sm" @click="$emit('delete')">
          {{ t('admin.accounts.bulkActions.delete') }}
        </button>
        <template v-if="showSystemActions">
          <button :disabled="loading" class="btn btn-secondary btn-sm" @click="$emit('reset-status')">
            {{ t('admin.accounts.bulkActions.resetStatus') }}
          </button>
          <button :disabled="loading" class="btn btn-secondary btn-sm" @click="$emit('refresh-token')">
            {{ t('admin.accounts.bulkActions.refreshToken') }}
          </button>
          <button :disabled="loading" class="btn btn-secondary btn-sm" @click="$emit('probe-upstream-billing')">
            {{ t('admin.accounts.bulkActions.probeUpstreamBilling') }}
          </button>
          <button :disabled="loading" class="btn btn-success btn-sm" @click="$emit('toggle-schedulable', true)">
            {{ t('admin.accounts.bulkActions.enableScheduling') }}
          </button>
          <button :disabled="loading" class="btn btn-warning btn-sm" @click="$emit('toggle-schedulable', false)">
            {{ t('admin.accounts.bulkActions.disableScheduling') }}
          </button>
        </template>
        <template v-if="showShareReviewActions">
          <button :disabled="loading" class="btn btn-success btn-sm" @click="$emit('share-status', 'active')">
            {{ t('admin.accounts.bulkActions.approveShare') }}
          </button>
          <button :disabled="loading" class="btn btn-secondary btn-sm" @click="$emit('share-status', 'rejected')">
            {{ t('admin.accounts.bulkActions.rejectShare') }}
          </button>
          <button :disabled="loading" class="btn btn-warning btn-sm" @click="$emit('share-status', 'suspended')">
            {{ t('admin.accounts.bulkActions.suspendShare') }}
          </button>
        </template>
        <button :disabled="loading" class="btn btn-primary btn-sm" @click="$emit('edit-selected')">
          {{ t('admin.accounts.bulkActions.edit') }}
        </button>
      </template>
      <template v-if="showShareReviewActions">
        <button :disabled="loading" class="btn btn-success btn-sm" @click="$emit('share-status-filtered', 'active')">
          {{ t('admin.accounts.bulkActions.approveFilteredShare') }}
        </button>
        <button :disabled="loading" class="btn btn-secondary btn-sm" @click="$emit('share-status-filtered', 'rejected')">
          {{ t('admin.accounts.bulkActions.rejectFilteredShare') }}
        </button>
        <button :disabled="loading" class="btn btn-warning btn-sm" @click="$emit('share-status-filtered', 'suspended')">
          {{ t('admin.accounts.bulkActions.suspendFilteredShare') }}
        </button>
      </template>
      <button :disabled="loading" class="btn btn-primary btn-sm" @click="$emit('edit-filtered')">
        {{ t('admin.accounts.bulkEdit.submit') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

withDefaults(defineProps<{
  selectedIds: number[]
  showDeleteAction?: boolean
  showSystemActions?: boolean
  showShareReviewActions?: boolean
  loading?: boolean
}>(), {
  showDeleteAction: true,
  showSystemActions: true,
  showShareReviewActions: false,
  loading: false
})

defineEmits<{
  delete: []
  'edit-selected': []
  'edit-filtered': []
  clear: []
  'select-page': []
  'toggle-schedulable': [schedulable: boolean]
  'reset-status': []
  'refresh-token': []
  'probe-upstream-billing': []
  'share-status': [status: 'active' | 'rejected' | 'suspended']
  'share-status-filtered': [status: 'active' | 'rejected' | 'suspended']
}>()

const { t } = useI18n()
</script>
