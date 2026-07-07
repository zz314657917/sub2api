<template>
  <BaseDialog
    :show="show"
    :title="t('admin.groups.accountPriority.title')"
    width="extra-wide"
    @close="handleClose"
  >
    <div v-if="group" class="space-y-4">
      <div class="flex flex-wrap items-center gap-3 rounded-lg bg-gray-50 px-4 py-2.5 text-sm dark:bg-dark-700">
        <span class="inline-flex items-center gap-1.5" :class="platformColorClass">
          <PlatformIcon :platform="group.platform" size="sm" />
          {{ t('admin.groups.platforms.' + group.platform) }}
        </span>
        <span class="text-gray-400">|</span>
        <span class="font-medium text-gray-900 dark:text-white">{{ group.name }}</span>
        <span class="text-gray-400">|</span>
        <span class="text-gray-600 dark:text-gray-400">
          {{ t('admin.groups.columns.accounts') }}: {{ rows.length }}
        </span>
        <button
          type="button"
          class="ml-auto rounded-md p-1.5 text-gray-500 transition-colors hover:bg-white hover:text-primary-600 disabled:opacity-50 dark:text-gray-400 dark:hover:bg-dark-600 dark:hover:text-primary-400"
          :disabled="loading"
          :title="t('admin.groups.accountPriority.reload')"
          @click="loadRows"
        >
          <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
        </button>
      </div>

      <p class="text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.groups.accountPriority.hint') }}
      </p>

      <div
        v-if="loading"
        class="flex items-center justify-center rounded-lg border border-dashed border-gray-200 py-10 text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
      >
        <Icon name="refresh" size="md" class="mr-2 animate-spin" />
        {{ t('admin.groups.accountPriority.loading') }}
      </div>

      <div
        v-else-if="rows.length === 0"
        class="rounded-lg border border-dashed border-gray-200 py-10 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
      >
        {{ t('admin.groups.accountPriority.empty') }}
      </div>

      <div v-else class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600">
        <div class="grid grid-cols-[3rem_minmax(0,1fr)_8rem_8rem] gap-3 border-b border-gray-200 bg-gray-50 px-3 py-2 text-xs font-medium uppercase text-gray-500 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-400">
          <span></span>
          <span>{{ t('admin.groups.accountPriority.account') }}</span>
          <span class="text-center">{{ t('admin.groups.accountPriority.groupPriority') }}</span>
          <span class="text-center">{{ t('admin.groups.accountPriority.globalPriority') }}</span>
        </div>

        <div class="max-h-[58vh] divide-y divide-gray-100 overflow-y-auto dark:divide-dark-600">
          <div
            v-for="(row, index) in rows"
            :key="row.account_id"
            class="grid grid-cols-[3rem_minmax(0,1fr)_8rem_8rem] gap-3 px-3 py-2.5 hover:bg-gray-50 dark:hover:bg-dark-700/50"
          >
            <div class="flex items-center justify-center gap-1">
              <button
                type="button"
                class="rounded p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700 disabled:opacity-35 dark:hover:bg-dark-600 dark:hover:text-gray-200"
                :disabled="index === 0"
                :title="t('admin.groups.accountPriority.moveUp')"
                @click="moveRow(index, index - 1)"
              >
                <Icon name="arrowUp" size="xs" />
              </button>
              <button
                type="button"
                class="rounded p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700 disabled:opacity-35 dark:hover:bg-dark-600 dark:hover:text-gray-200"
                :disabled="index === rows.length - 1"
                :title="t('admin.groups.accountPriority.moveDown')"
                @click="moveRow(index, index + 1)"
              >
                <Icon name="arrowDown" size="xs" />
              </button>
            </div>

            <div class="min-w-0">
              <div class="flex min-w-0 items-center gap-2">
                <PlatformIcon
                  :platform="row.platform"
                  size="xs"
                  class="text-gray-500 dark:text-gray-400"
                />
                <span class="truncate text-sm font-medium text-gray-900 dark:text-white">
                  {{ row.name }}
                </span>
              </div>
              <div class="mt-1 flex flex-wrap items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
                <span>#{{ row.account_id }}</span>
                <span>{{ row.type }}</span>
                <span :class="accountStatusClass(row.status)">
                  {{ t('admin.accounts.status.' + row.status) }}
                </span>
              </div>
            </div>

            <input
              v-model.number="row.group_priority"
              type="number"
              min="1"
              step="1"
              class="input h-9 text-center text-sm"
              @change="markDirty"
            />

            <div class="flex h-9 items-center justify-center rounded-md bg-gray-50 text-sm text-gray-500 dark:bg-dark-700 dark:text-gray-300">
              {{ row.global_priority }}
            </div>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-3 border-t border-gray-200 pt-4 dark:border-dark-600">
        <template v-if="dirty">
          <span class="text-xs text-amber-600 dark:text-amber-400">
            {{ t('admin.groups.accountPriority.unsaved') }}
          </span>
          <button
            type="button"
            class="text-xs font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
            @click="revertRows"
          >
            {{ t('admin.groups.revertChanges') }}
          </button>
        </template>

        <div class="ml-auto flex items-center gap-3">
          <button type="button" class="btn btn-secondary btn-sm px-4 py-1.5" @click="handleClose">
            {{ t('common.close') }}
          </button>
          <button
            type="button"
            class="btn btn-primary btn-sm px-4 py-1.5"
            :disabled="saving || loading || !dirty"
            @click="handleSave"
          >
            <Icon v-if="saving" name="refresh" size="sm" class="mr-1 animate-spin" />
            {{ t('common.save') }}
          </button>
        </div>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { Account, AccountPlatform, AccountType, AdminGroup } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'

interface PriorityRow {
  account_id: number
  name: string
  platform: AccountPlatform
  type: AccountType
  status: Account['status']
  global_priority: number
  group_priority: number
}

const props = defineProps<{
  show: boolean
  group: AdminGroup | null
}>()

const emit = defineEmits<{
  close: []
  success: []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)
const dirty = ref(false)
const serverRows = ref<PriorityRow[]>([])
const rows = ref<PriorityRow[]>([])

const platformColorClass = computed(() => {
  switch (props.group?.platform) {
    case 'anthropic':
      return 'text-orange-700 dark:text-orange-400'
    case 'openai':
      return 'text-[#a9583e] dark:text-[#f0b89e]'
    case 'antigravity':
      return 'text-purple-700 dark:text-purple-400'
    default:
      return 'text-[#6c6a64] dark:text-[#d8cec2]'
  }
})

const accountStatusClass = (status: Account['status']) => [
  'rounded px-1.5 py-0.5 font-medium',
  status === 'active'
    ? 'bg-[#eef4ef] text-[#5f7f68] dark:bg-[#7f9d8a]/15 dark:text-[#9ab3a0]'
    : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
]

const normalizePositivePriority = (value: unknown, fallback: number) => {
  const numeric = Number(value)
  if (!Number.isFinite(numeric) || numeric <= 0) {
    return fallback
  }
  return Math.trunc(numeric)
}

const getAccountGroupPriority = (account: Account, groupID: number) => {
  const accountGroup = account.account_groups?.find((item) => item.group_id === groupID)
  return normalizePositivePriority(accountGroup?.priority, account.priority || 1)
}

const toPriorityRow = (account: Account, groupID: number): PriorityRow => ({
  account_id: account.id,
  name: account.name,
  platform: account.platform,
  type: account.type,
  status: account.status,
  global_priority: normalizePositivePriority(account.priority, 1),
  group_priority: getAccountGroupPriority(account, groupID)
})

const cloneRows = (value: PriorityRow[]) => value.map((row) => ({ ...row }))

const normalizeRows = () => {
  const normalized = rows.value
    .map((row, index) => ({
      row,
      index,
      priority: normalizePositivePriority(row.group_priority, index + 1)
    }))
    .sort(
      (a, b) =>
        a.priority - b.priority ||
        a.index - b.index ||
        a.row.account_id - b.row.account_id
    )
    .map(({ row }, index) => ({
      ...row,
      group_priority: index + 1
    }))

  rows.value = normalized
  return normalized
}

const markDirty = () => {
  dirty.value = true
}

const moveRow = (fromIndex: number, toIndex: number) => {
  if (toIndex < 0 || toIndex >= rows.value.length || fromIndex === toIndex) {
    return
  }

  const nextRows = [...rows.value]
  const [row] = nextRows.splice(fromIndex, 1)
  nextRows.splice(toIndex, 0, row)
  rows.value = nextRows.map((item, index) => ({
    ...item,
    group_priority: index + 1
  }))
  dirty.value = true
}

const loadRows = async () => {
  if (!props.group) return
  loading.value = true
  try {
    const accounts = await adminAPI.groups.getGroupAccounts(props.group.id)
    const loaded = accounts
      .map((account) => toPriorityRow(account, props.group!.id))
      .sort(
        (a, b) =>
          a.group_priority - b.group_priority ||
          a.global_priority - b.global_priority ||
          a.account_id - b.account_id
      )
    serverRows.value = cloneRows(loaded)
    rows.value = loaded
    dirty.value = false
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail || t('admin.groups.accountPriority.loadFailed')
    )
    console.error('Error loading group account priorities:', error)
  } finally {
    loading.value = false
  }
}

const revertRows = () => {
  rows.value = cloneRows(serverRows.value)
  dirty.value = false
}

const resetState = () => {
  serverRows.value = []
  rows.value = []
  loading.value = false
  saving.value = false
  dirty.value = false
}

const handleSave = async () => {
  if (!props.group) return
  saving.value = true
  try {
    const normalized = normalizeRows()
    await adminAPI.groups.updateGroupAccountPriorities(
      props.group.id,
      normalized.map((row) => ({
        account_id: row.account_id,
        priority: row.group_priority
      }))
    )
    serverRows.value = cloneRows(normalized)
    dirty.value = false
    appStore.showSuccess(t('admin.groups.accountPriority.saved'))
    emit('success')
    emit('close')
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.detail || t('admin.groups.accountPriority.saveFailed')
    )
    console.error('Error saving group account priorities:', error)
  } finally {
    saving.value = false
  }
}

const handleClose = () => {
  if (dirty.value) {
    revertRows()
  }
  emit('close')
}

watch(
  () => [props.show, props.group?.id] as const,
  ([show]) => {
    if (show && props.group) {
      loadRows()
    } else if (!show) {
      resetState()
    }
  }
)
</script>
