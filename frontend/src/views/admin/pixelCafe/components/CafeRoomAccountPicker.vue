<template>
  <section data-testid="cafe-room-account-picker" class="space-y-3" :aria-busy="loading">
    <label class="field">
      <span class="input-label">{{ t('admin.pixelCafe.picker.search') }}</span>
      <input
        v-model="search"
        class="input"
        type="search"
        :placeholder="t('admin.pixelCafe.picker.searchPlaceholder')"
        :disabled="!planId"
        @input="scheduleSearch"
      />
    </label>

    <p v-if="!planId" class="field-hint text-amber-600 dark:text-amber-300">{{ t('admin.pixelCafe.picker.choosePlanFirst') }}</p>
    <p v-else-if="errorMessage" class="field-hint text-red-600 dark:text-red-300">{{ errorMessage }}</p>
    <p v-else-if="loading" class="field-hint text-gray-500 dark:text-dark-400">{{ t('admin.pixelCafe.picker.loading') }}</p>

    <div v-if="selectedOptions.length" class="rounded-lg border border-primary-200 bg-primary-50 p-3 dark:border-primary-800 dark:bg-primary-900/20">
      <div class="mb-2 text-xs font-medium text-primary-800 dark:text-primary-200">{{ t('admin.pixelCafe.picker.selected') }}</div>
      <div class="space-y-2">
        <label v-for="option in selectedOptions" :key="option.id" class="flex cursor-pointer items-center justify-between gap-3 text-sm">
          <span class="min-w-0">
            <span class="block truncate font-medium text-gray-900 dark:text-white">{{ option.name }}</span>
            <span class="block truncate text-xs text-gray-500 dark:text-dark-400">{{ option.platform }} · {{ option.email_masked || '-' }} · #{{ option.id }} · {{ option.status }}</span>
          </span>
          <input :checked="isSelected(option.id)" type="checkbox" class="h-4 w-4" @change="toggle(option.id)" />
        </label>
      </div>
    </div>

    <div v-if="candidates.length" class="max-h-60 space-y-1 overflow-y-auto rounded-lg border border-gray-200 p-2 dark:border-dark-700" role="listbox" :aria-multiselectable="multiple">
      <label v-for="option in candidates" :key="option.id" class="flex cursor-pointer items-center gap-3 rounded p-2 hover:bg-gray-50 dark:hover:bg-dark-800">
        <input
          :checked="isSelected(option.id)"
          :type="multiple ? 'checkbox' : 'radio'"
          :name="multiple ? undefined : 'cafe-room-account'"
          class="h-4 w-4"
          @change="toggle(option.id)"
        />
        <span class="min-w-0 text-sm">
          <span class="block truncate font-medium text-gray-900 dark:text-white">{{ option.name }}</span>
          <span class="block truncate text-xs text-gray-500 dark:text-dark-400">{{ option.platform }} · {{ option.email_masked || '-' }} · #{{ option.id }} · {{ option.status }}</span>
        </span>
      </label>
    </div>
    <p v-else-if="planId && !loading && !errorMessage" class="field-hint text-gray-500 dark:text-dark-400">{{ t('admin.pixelCafe.picker.empty') }}</p>
    <button v-if="hasMore" type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="loadMore">{{ t('admin.pixelCafe.picker.more') }}</button>
  </section>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import cafeRoomsAPI, { type CafeRoomAccountOption } from '@/api/admin/cafeRooms'
import { extractApiErrorMessage } from '@/utils/apiError'

const props = withDefaults(defineProps<{
  modelValue: number | number[]
  multiple?: boolean
  planId: number
  excludeRoomId?: number
  active?: boolean
}>(), { multiple: false, excludeRoomId: 0, active: false })

const emit = defineEmits<{ 'update:modelValue': [value: number | number[]] }>()
const { t } = useI18n()
const search = ref('')
const candidates = ref<CafeRoomAccountOption[]>([])
const selectedByID = ref<Record<number, CafeRoomAccountOption>>({})
const loading = ref(false)
const errorMessage = ref('')
const page = ref(1)
const pages = ref(1)
let candidateVersion = 0
let hydrationVersion = 0
let searchTimer: number | undefined

const selectedIDs = computed(() => (Array.isArray(props.modelValue) ? props.modelValue : props.modelValue > 0 ? [props.modelValue] : []))
const selectedOptions = computed(() => selectedIDs.value.map((id) => selectedByID.value[id] || { id, name: `#${id}`, platform: '-', status: '-' }))
const hasMore = computed(() => page.value < pages.value)

function isSelected(id: number) { return selectedIDs.value.includes(id) }
function toggle(id: number) {
  if (props.multiple) {
    emit('update:modelValue', isSelected(id) ? selectedIDs.value.filter((value) => value !== id) : [...selectedIDs.value, id])
    return
  }
  emit('update:modelValue', isSelected(id) ? 0 : id)
}

async function hydrateSelected() {
  if (!props.active || selectedIDs.value.length === 0) return
  const version = ++hydrationVersion
  const ids = [...selectedIDs.value]
  try {
    const responses = await Promise.all(Array.from({ length: Math.ceil(ids.length / 50) }, (_, index) => cafeRoomsAPI.listAccountOptions({ ids: ids.slice(index * 50, (index + 1) * 50) })))
    if (version !== hydrationVersion) return
    selectedByID.value = { ...selectedByID.value, ...Object.fromEntries(responses.flatMap((response) => response.data.items).map((item) => [item.id, item])) }
  } catch { /* candidates still provide an ID-only fallback */ }
}

async function loadCandidates(reset = true) {
  if (!props.active || !props.planId) return
  if (reset) { page.value = 1; candidates.value = [] }
  const version = ++candidateVersion
  const requestPage = page.value
  const requestSearch = search.value.trim()
  const requestPlanID = props.planId
  const requestExcludeRoomID = props.excludeRoomId
  loading.value = true
  errorMessage.value = ''
  try {
    const response = await cafeRoomsAPI.listAccountOptions({ page: requestPage, page_size: 20, search: requestSearch || undefined, plan_id: requestPlanID, exclude_room_id: requestExcludeRoomID || undefined })
    if (version !== candidateVersion) return
    const items = response.data.items
    candidates.value = reset ? items : [...candidates.value, ...items]
    selectedByID.value = { ...selectedByID.value, ...Object.fromEntries(items.map((item) => [item.id, item])) }
    page.value = response.data.page
    pages.value = response.data.pages
  } catch (error) {
    if (version === candidateVersion) errorMessage.value = extractApiErrorMessage(error, t('admin.pixelCafe.picker.error'))
  } finally {
    if (version === candidateVersion) loading.value = false
  }
}

function invalidateCandidates() {
  candidateVersion += 1
  loading.value = false
  errorMessage.value = ''
}

function scheduleSearch() {
  if (searchTimer) window.clearTimeout(searchTimer)
  invalidateCandidates()
  candidates.value = []
  page.value = 1
  pages.value = 1
  searchTimer = window.setTimeout(() => { void loadCandidates() }, 250)
}
function loadMore() { page.value += 1; void loadCandidates(false) }

watch(() => [props.active, props.planId, props.excludeRoomId] as const, () => {
  invalidateCandidates()
  candidates.value = []
  page.value = 1
  pages.value = 1
  if (props.active) { void hydrateSelected(); void loadCandidates() }
}, { immediate: true })
watch(selectedIDs, () => { void hydrateSelected() })
onUnmounted(() => { if (searchTimer) window.clearTimeout(searchTimer) })
</script>
