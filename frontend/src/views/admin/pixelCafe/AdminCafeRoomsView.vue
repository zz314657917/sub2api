<template>
  <AppLayout>
    <div class="space-y-4">
      <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="min-w-56 flex-1 sm:max-w-72">
            <input
              v-model="search"
              class="input"
              type="search"
              :placeholder="t('admin.pixelCafe.searchPlaceholder')"
              @input="scheduleReload"
            />
          </div>
          <Select
            v-model="filters.status"
            :options="statusOptions"
            class="w-40"
            @change="resetAndLoad"
          />
          <Select
            v-model="filters.zone"
            :options="zoneOptions"
            class="w-40"
            @change="resetAndLoad"
          />
          <div class="ml-auto flex flex-wrap items-center gap-2">
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="loading"
              :title="t('admin.pixelCafe.refresh')"
              @click="loadRooms"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
              <span class="sr-only">{{ t('admin.pixelCafe.refresh') }}</span>
            </button>
            <button type="button" class="btn btn-secondary" @click="openBulkDialog">
              <Icon name="copy" size="md" class="mr-1" />
              {{ t('admin.pixelCafe.bulkCreate') }}
            </button>
            <button type="button" class="btn btn-primary" @click="openCreateDialog">
              <Icon name="plus" size="md" class="mr-1" />
              {{ t('admin.pixelCafe.createRoom') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="rooms"
          :loading="loading"
          :sticky-first-column="true"
          :sticky-actions-column="true"
          :expandable-actions="false"
          row-key="id"
        >
          <template #cell-room="{ row }">
            <div class="min-w-44">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.name }}</div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ row.code }} · #{{ row.id }}</div>
            </div>
          </template>
          <template #cell-zone="{ row }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ row.zone_key || 'featured' }}</span>
          </template>
          <template #cell-plan="{ row }">
            <div class="min-w-44">
              <div class="font-medium text-gray-800 dark:text-gray-200">{{ planTitle(row) }}</div>
              <div class="text-xs text-gray-500 dark:text-dark-400">
                #{{ row.plan_id }} · {{ planMode(row) }}
              </div>
            </div>
          </template>
          <template #cell-account="{ row }">
            <div v-if="accountFor(row.account_id)" class="min-w-40">
              <div class="font-medium text-gray-800 dark:text-gray-200">{{ accountFor(row.account_id)?.name }}</div>
              <div class="text-xs text-gray-500 dark:text-dark-400">
                #{{ row.account_id }} · {{ accountPlatform(row.account_id) }}
              </div>
            </div>
            <span v-else class="text-sm text-gray-500 dark:text-dark-400">
              {{ row.account_id ? t('admin.pixelCafe.unknownAccount', { id: row.account_id }) : '-' }}
            </span>
          </template>
          <template #cell-status="{ row }">
            <span class="status-badge" :class="statusClass(row.status)">
              {{ t(`admin.pixelCafe.status.${row.status}`) }}
            </span>
          </template>
          <template #cell-featured="{ row }">
            <span class="text-sm" :class="row.featured ? 'text-amber-600 dark:text-amber-300' : 'text-gray-500 dark:text-dark-400'">
              {{ row.featured ? t('admin.pixelCafe.featured') : t('admin.pixelCafe.notFeatured') }}
            </span>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex min-w-56 flex-wrap items-center gap-2">
              <button type="button" class="btn btn-ghost btn-sm" @click="openEditDialog(row)">
                <Icon name="edit" size="sm" class="mr-1" />
                {{ t('admin.pixelCafe.actions.edit') }}
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="row.status !== 'enabled' || openingRoundId === row.id"
                @click="openRound(row)"
              >
                <Icon name="play" size="sm" class="mr-1" />
                {{ openingRoundId === row.id ? t('admin.pixelCafe.actions.openingRound') : t('admin.pixelCafe.actions.openRound') }}
              </button>
              <button
                type="button"
                class="btn btn-ghost btn-sm text-red-600 hover:text-red-700 dark:text-red-300"
                :disabled="row.status === 'enabled' || deletingId === row.id"
                @click="askDelete(row)"
              >
                <Icon name="trash" size="sm" class="mr-1" />
                {{ t('admin.pixelCafe.actions.delete') }}
              </button>
            </div>
          </template>
          <template #empty>
            <EmptyState
              :title="t('admin.pixelCafe.noRooms')"
              :description="loadError || t('admin.pixelCafe.noRooms')"
              :action-text="t('admin.pixelCafe.createRoom')"
              @action="openCreateDialog"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="changePage"
          @update:page-size="changePageSize"
        />
      </template>
      </TablePageLayout>

      <AdminGroupBuyView embedded />
    </div>

    <BaseDialog
      :show="roomDialogOpen"
      :title="editingRoom ? t('admin.pixelCafe.form.editTitle') : t('admin.pixelCafe.form.createTitle')"
      width="wide"
      @close="closeRoomDialog"
    >
      <form id="cafe-room-form" class="space-y-4" @submit.prevent="saveRoom">
        <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(20rem,0.9fr)]">
          <div class="grid gap-4 sm:grid-cols-2">
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.code') }}</span>
            <input v-model.trim="roomForm.code" class="input" required maxlength="64" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.name') }}</span>
            <input v-model.trim="roomForm.name" class="input" required maxlength="120" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.plan') }}</span>
            <select v-model.number="roomForm.plan_id" class="input" required>
              <option :value="0" disabled>{{ t('admin.pixelCafe.form.choosePlan') }}</option>
              <option v-for="plan in roomPlans" :key="plan.id" :value="plan.id">
                {{ plan.title }} · #{{ plan.id }} · {{ plan.target_group_id }}
              </option>
            </select>
            <span v-if="roomPlans.length === 0" class="field-hint text-amber-600 dark:text-amber-300">
              {{ t('admin.pixelCafe.noRoomPlans') }}
            </span>
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.zone') }}</span>
            <input v-model.trim="roomForm.zone_key" class="input" maxlength="32" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.theme') }}</span>
            <input v-model.trim="roomForm.theme_key" class="input" maxlength="64" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.sceneSlot') }}</span>
            <input v-model.trim="roomForm.scene_slot_key" class="input" maxlength="120" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.status') }}</span>
            <select v-model="roomForm.status" class="input">
              <option v-for="option in statusOptions.slice(1)" :key="String(option.value)" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.sortOrder') }}</span>
            <input v-model.number="roomForm.sort_order" class="input" type="number" min="0" />
          </label>
          </div>
          <div class="rounded-lg border border-gray-200 p-4 dark:border-dark-700">
            <span class="input-label mb-2 block">{{ t('admin.pixelCafe.form.account') }}</span>
            <CafeRoomAccountPicker
              v-model="roomForm.account_id"
              :plan-id="roomForm.plan_id"
              :exclude-room-id="editingRoom?.id || 0"
              :active="roomDialogOpen"
            />
          </div>
        </div>
        <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
          <input v-model="roomForm.featured" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600" />
          {{ t('admin.pixelCafe.form.featured') }}
        </label>
        <p v-if="dependencyLoading" class="text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.pixelCafe.loadingDependencies') }}
        </p>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeRoomDialog">{{ t('common.cancel') }}</button>
          <button type="submit" form="cafe-room-form" class="btn btn-primary" :disabled="saving || !roomForm.plan_id || !roomForm.account_id">
            <Icon v-if="saving" name="refresh" size="sm" class="mr-1 animate-spin" />
            {{ saving ? t('admin.pixelCafe.form.saving') : t('admin.pixelCafe.form.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="bulkDialogOpen"
      :title="t('admin.pixelCafe.bulk.title')"
      width="wide"
      @close="closeBulkDialog"
    >
      <form id="cafe-room-bulk-form" class="space-y-4" @submit.prevent="submitBulkCreate">
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.plan') }}</span>
            <select v-model.number="bulkForm.plan_id" class="input" required>
              <option :value="0" disabled>{{ t('admin.pixelCafe.form.choosePlan') }}</option>
              <option v-for="plan in roomPlans" :key="plan.id" :value="plan.id">{{ plan.title }} · #{{ plan.id }}</option>
            </select>
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.bulk.codePrefix') }}</span>
            <input v-model.trim="bulkForm.code_prefix" class="input" maxlength="40" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.bulk.startNumber') }}</span>
            <input v-model.number="bulkForm.start_number" class="input" type="number" min="1" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.zone') }}</span>
            <input v-model.trim="bulkForm.zone_key" class="input" maxlength="32" />
          </label>
          <label class="field sm:col-span-2">
            <span class="input-label">{{ t('admin.pixelCafe.form.theme') }}</span>
            <input v-model.trim="bulkForm.theme_key" class="input" maxlength="64" />
          </label>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-dark-700">
          <span class="input-label mb-2 block">{{ t('admin.pixelCafe.bulk.accounts') }}</span>
          <CafeRoomAccountPicker v-model="bulkForm.account_ids" multiple :plan-id="bulkForm.plan_id" :active="bulkDialogOpen" />
        </div>
        <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
          <input v-model="bulkForm.create_open_round" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600" />
          {{ t('admin.pixelCafe.bulk.createOpenRound') }}
        </label>
        <div v-if="bulkResult" class="space-y-3 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900/50">
          <div class="flex flex-wrap gap-3 text-sm">
            <span class="text-emerald-700 dark:text-emerald-300">{{ t('admin.pixelCafe.bulk.created', { count: bulkResult.created.length }) }}</span>
            <span class="text-red-700 dark:text-red-300">{{ t('admin.pixelCafe.bulk.failed', { count: bulkResult.failed.length }) }}</span>
          </div>
          <ul v-if="bulkResult.failed.length" class="space-y-1 text-sm text-red-700 dark:text-red-300">
            <li v-for="failure in bulkResult.failed" :key="`${failure.account_id}-${failure.error_code}`">
              #{{ failure.account_id }} · {{ failure.error_code }} · {{ failure.message }}
            </li>
          </ul>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeBulkDialog">{{ t('common.close') }}</button>
          <button type="submit" form="cafe-room-bulk-form" class="btn btn-primary" :disabled="bulkSaving || !bulkForm.plan_id || bulkForm.account_ids.length === 0">
            <Icon v-if="bulkSaving" name="refresh" size="sm" class="mr-1 animate-spin" />
            {{ bulkSaving ? t('admin.pixelCafe.bulk.submitting') : t('admin.pixelCafe.bulk.submit') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="Boolean(roomToDelete)"
      :title="t('admin.pixelCafe.confirmDeleteTitle')"
      :message="roomToDelete ? t('admin.pixelCafe.confirmDeleteMessage', { name: roomToDelete.name }) : ''"
      :danger="true"
      @cancel="roomToDelete = null"
      @confirm="deleteRoom"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { GroupBuyPlan } from '@/types/groupBuy'
import type { Column } from '@/components/common/types'
import type { CafeRoom, CafeRoomBulkResult, CafeRoomInput, CafeRoomStatus } from '@/types/pixelCafe'

import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import AdminGroupBuyView from '@/views/admin/group-buy/AdminGroupBuyView.vue'
import CafeRoomAccountPicker from './components/CafeRoomAccountPicker.vue'
import type { CafeRoomAccountOption } from '@/api/admin/cafeRooms'

const { t } = useI18n()
const appStore = useAppStore()

const rooms = ref<CafeRoom[]>([])
const plans = ref<GroupBuyPlan[]>([])
const accountOptionsByID = ref<Record<number, CafeRoomAccountOption>>({})
const loading = ref(false)
const dependencyLoading = ref(false)
const loadError = ref('')
const search = ref('')
const filters = reactive({ status: '', zone: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0, pages: 1 })

const roomDialogOpen = ref(false)
const bulkDialogOpen = ref(false)
const editingRoom = ref<CafeRoom | null>(null)
const saving = ref(false)
const bulkSaving = ref(false)
const openingRoundId = ref<number | null>(null)
const deletingId = ref<number | null>(null)
const roomToDelete = ref<CafeRoom | null>(null)
const bulkResult = ref<CafeRoomBulkResult | null>(null)

const roomForm = reactive<CafeRoomInput>({
  code: '',
  name: '',
  plan_id: 0,
  account_id: 0,
  zone_key: 'featured',
  theme_key: 'warm_wood',
  scene_slot_key: '',
  status: 'draft',
  featured: false,
  sort_order: 0,
})

const bulkForm = reactive({
  plan_id: 0,
  account_ids: [] as number[],
  code_prefix: 'ROOM-',
  start_number: 1,
  zone_key: 'featured',
  theme_key: 'warm_wood',
  create_open_round: false,
})

const columns = computed<Column[]>(() => [
  { key: 'room', label: t('admin.pixelCafe.columns.room'), sortable: true },
  { key: 'zone', label: t('admin.pixelCafe.columns.zone'), sortable: true },
  { key: 'plan', label: t('admin.pixelCafe.columns.plan') },
  { key: 'account', label: t('admin.pixelCafe.columns.account') },
  { key: 'status', label: t('admin.pixelCafe.columns.status'), sortable: true },
  { key: 'featured', label: t('admin.pixelCafe.columns.featured') },
  { key: 'actions', label: t('admin.pixelCafe.columns.actions') },
])

const statusOptions = computed(() => [
  { value: '', label: t('admin.pixelCafe.allStatus') },
  ...(['draft', 'enabled', 'maintenance', 'disabled'] as CafeRoomStatus[]).map((status) => ({
    value: status,
    label: t(`admin.pixelCafe.status.${status}`),
  })),
])

const zoneOptions = computed(() => [
  { value: '', label: t('admin.pixelCafe.allZones') },
  ...Array.from(new Set(rooms.value.map((room) => room.zone_key).filter(Boolean))).map((zone) => ({ value: zone, label: zone })),
])

const roomPlans = computed(() => plans.value.filter((plan) => plan.fulfillment_mode === 'room_subscription'))
function resetRoomForm() {
  Object.assign(roomForm, {
    code: '', name: '', plan_id: roomPlans.value[0]?.id ?? 0, account_id: 0,
    zone_key: 'featured', theme_key: 'warm_wood', scene_slot_key: '', status: 'draft', featured: false, sort_order: 0,
  })
}

function openCreateDialog() {
  editingRoom.value = null
  resetRoomForm()
  roomDialogOpen.value = true
}

function openEditDialog(room: CafeRoom) {
  editingRoom.value = room
  Object.assign(roomForm, {
    code: room.code, name: room.name, plan_id: room.plan_id, account_id: room.account_id ?? 0,
    zone_key: room.zone_key, theme_key: room.theme_key, scene_slot_key: room.scene_slot_key,
    status: room.status, featured: room.featured, sort_order: room.sort_order,
  })
  roomDialogOpen.value = true
}

function closeRoomDialog() {
  if (saving.value) return
  roomDialogOpen.value = false
}

function openBulkDialog() {
  bulkResult.value = null
  bulkForm.plan_id = roomPlans.value[0]?.id ?? 0
  bulkForm.account_ids = []
  bulkDialogOpen.value = true
}

function closeBulkDialog() {
  if (bulkSaving.value) return
  bulkDialogOpen.value = false
}

function statusClass(status: string) {
  return {
    draft: 'status-badge-muted',
    enabled: 'status-badge-success',
    maintenance: 'status-badge-warning',
    disabled: 'status-badge-danger',
  }[status] || 'status-badge-muted'
}

function planTitle(room: CafeRoom) {
  return room.plan?.title || plans.value.find((plan) => plan.id === room.plan_id)?.title || `#${room.plan_id}`
}

function planMode(room: CafeRoom) {
  return room.plan?.fulfillment_mode || plans.value.find((plan) => plan.id === room.plan_id)?.fulfillment_mode || 'aggregate_tier'
}

function accountFor(id: number | null | undefined) {
  return id ? accountOptionsByID.value[id] : undefined
}

function accountPlatform(id: number | null | undefined) {
  return id ? accountFor(id)?.platform || '-' : '-'
}

async function loadDependencies() {
  dependencyLoading.value = true
  try {
    const planResponse = await adminAPI.groupBuy.listPlans()
    plans.value = planResponse.data
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.dependencies')))
  } finally {
    dependencyLoading.value = false
  }
}

let searchTimer: number | null = null
function scheduleReload() {
  if (searchTimer) window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(resetAndLoad, 300)
}

function resetAndLoad() {
  pagination.page = 1
  void loadRooms()
}

async function loadRooms() {
  loading.value = true
  loadError.value = ''
  try {
    const response = await adminAPI.cafeRooms.list({
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
      zone: filters.zone || undefined,
      search: search.value.trim() || undefined,
      sort_by: 'sort_order',
      sort_order: 'asc',
    })
    rooms.value = response.data.items
    await hydrateRoomAccounts(response.data.items)
    pagination.total = response.data.total
    pagination.pages = response.data.pages
    pagination.page = response.data.page
    pagination.page_size = response.data.page_size
  } catch (error) {
    loadError.value = extractApiErrorMessage(error, t('admin.pixelCafe.errors.load'))
    appStore.showError(loadError.value)
  } finally {
    loading.value = false
  }
}

async function hydrateRoomAccounts(currentRooms: CafeRoom[]) {
  const ids = [...new Set(currentRooms.map((room) => room.account_id).filter((id): id is number => Boolean(id && id > 0)))]
  if (ids.length === 0) return
  try {
    const responses = await Promise.all(Array.from({ length: Math.ceil(ids.length / 50) }, (_, index) => adminAPI.cafeRooms.listAccountOptions({ ids: ids.slice(index * 50, (index + 1) * 50) })))
    accountOptionsByID.value = { ...accountOptionsByID.value, ...Object.fromEntries(responses.flatMap((response) => response.data.items).map((account) => [account.id, account])) }
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.accounts')))
  }
}

function changePage(page: number) {
  pagination.page = page
  void loadRooms()
}

function changePageSize(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadRooms()
}

async function saveRoom() {
  if (!roomForm.plan_id || !roomForm.account_id) return
  saving.value = true
  try {
    if (editingRoom.value) {
      await adminAPI.cafeRooms.update(editingRoom.value.id, { ...roomForm })
      appStore.showSuccess(t('admin.pixelCafe.success.updated'))
    } else {
      await adminAPI.cafeRooms.create({ ...roomForm })
      appStore.showSuccess(t('admin.pixelCafe.success.created'))
    }
    roomDialogOpen.value = false
    await loadRooms()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.save')))
  } finally {
    saving.value = false
  }
}

function askDelete(room: CafeRoom) {
  roomToDelete.value = room
}

async function deleteRoom() {
  if (!roomToDelete.value) return
  const room = roomToDelete.value
  deletingId.value = room.id
  try {
    await adminAPI.cafeRooms.remove(room.id)
    roomToDelete.value = null
    appStore.showSuccess(t('admin.pixelCafe.success.deleted'))
    if (rooms.value.length === 1 && pagination.page > 1) pagination.page -= 1
    await loadRooms()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.delete')))
  } finally {
    deletingId.value = null
  }
}

async function openRound(room: CafeRoom) {
  openingRoundId.value = room.id
  try {
    await adminAPI.cafeRooms.openRound(room.id)
    appStore.showSuccess(t('admin.pixelCafe.success.roundOpened'))
    await loadRooms()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.openRound')))
  } finally {
    openingRoundId.value = null
  }
}

async function submitBulkCreate() {
  if (!bulkForm.plan_id || bulkForm.account_ids.length === 0) {
    appStore.showError(t('admin.pixelCafe.bulk.noneSelected'))
    return
  }
  bulkSaving.value = true
  bulkResult.value = null
  try {
    const response = await adminAPI.cafeRooms.bulkCreate({ ...bulkForm })
    bulkResult.value = response.data
    appStore.showSuccess(t('admin.pixelCafe.success.bulkCreated', {
      created: response.data.created.length,
      failed: response.data.failed.length,
    }))
    await loadRooms()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.bulk')))
  } finally {
    bulkSaving.value = false
  }
}

onMounted(() => {
  void Promise.all([loadDependencies(), loadRooms()])
})

onUnmounted(() => {
  if (searchTimer) window.clearTimeout(searchTimer)
})
</script>

<style scoped>
.field {
  @apply block space-y-1;
}

.field-hint {
  @apply mt-1 block text-xs;
}

.status-badge {
  @apply inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-medium;
}

.status-badge-muted {
  @apply border-gray-200 bg-gray-50 text-gray-700 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300;
}

.status-badge-success {
  @apply border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-900/20 dark:text-emerald-300;
}

.status-badge-warning {
  @apply border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-300;
}

.status-badge-danger {
  @apply border-red-200 bg-red-50 text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300;
}
</style>
