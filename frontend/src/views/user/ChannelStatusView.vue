<template>
  <AppLayout>
    <div class="space-y-4">
      <div
        class="grid grid-cols-1 items-start gap-4"
        :class="{ 'xl:grid-cols-2': showCapacityColumn }"
        data-testid="channel-status-layout"
      >
        <div
          v-if="showCapacityColumn"
          class="space-y-4"
          data-testid="capacity-column"
        >
          <AccountCapacityPools
            v-if="showSharedCapacityPool"
            :pools="visibleCapacityPools"
            :loading="capacityLoading"
            :pool-keys="['shared']"
          />

          <AccountCapacityPools
            v-if="showMineCapacityPool"
            :pools="visibleCapacityPools"
            :loading="capacityLoading"
            :pool-keys="['mine']"
          />
        </div>

        <section class="space-y-3" data-testid="channel-monitor-panel">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="min-w-0">
              <h2 class="text-base font-bold text-gray-900 dark:text-white">
                {{ t('channelStatus.monitorTitle') }}
              </h2>
              <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
                {{ t('channelStatus.monitorDescription') }}
              </p>
            </div>

            <MonitorHero
              :overall-status="overallStatus"
              :window="currentWindow"
              :loading="loading"
              :auto-refresh="autoRefresh"
              @update:window="handleWindowChange"
              @refresh="refresh()"
            />
          </div>

          <MonitorCardGrid
            :items="items"
            :window="currentWindow"
            :countdown-seconds="autoRefresh.countdown.value"
            :loading="loading"
            :detail-cache="detailCache"
            @card-click="openDetail"
          />
        </section>
      </div>
    </div>

    <MonitorDetailDialog
      :show="showDetail"
      :monitor-id="detailTarget?.id ?? null"
      :title="detailTitle"
      @close="closeDetail"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  list as listChannelMonitorViews,
  status as fetchChannelMonitorDetail,
  type UserMonitorView,
  type UserMonitorDetail,
} from '@/api/channelMonitor'
import { getAccountCapacityPools } from '@/api/user'
import type { UserAccountCapacityPool, UserAccountCapacityPools } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import MonitorHero, {
  type MonitorWindow,
  type OverallStatus,
} from '@/components/user/monitor/MonitorHero.vue'
import AccountCapacityPools from '@/components/user/monitor/AccountCapacityPools.vue'
import MonitorCardGrid from '@/components/user/monitor/MonitorCardGrid.vue'
import MonitorDetailDialog from '@/components/user/MonitorDetailDialog.vue'
import { DEFAULT_INTERVAL_SECONDS, STATUS_OPERATIONAL } from '@/constants/channelMonitor'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const { t } = useI18n()
const appStore = useAppStore()

// ── State ──
const items = ref<UserMonitorView[]>([])
const loading = ref(false)
const capacityLoading = ref(false)
const capacityPools = ref<UserAccountCapacityPools | null>(null)
const currentWindow = ref<MonitorWindow>('7d')
const detailCache = reactive<Record<number, UserMonitorDetail>>({})
const showDetail = ref(false)
const detailTarget = ref<UserMonitorView | null>(null)

let abortController: AbortController | null = null
let capacityAbortController: AbortController | null = null
let refreshVersion = 0
let unmounted = false
const pendingDetails = new Map<number, { version?: number; promise: Promise<void> }>()

function isCurrentRefresh(version: number) {
  return !unmounted && version === refreshVersion
}

const autoRefresh = useAutoRefresh({
  storageKey: 'channel-status-auto-refresh',
  intervals: [30, 60, 120] as const,
  defaultInterval: DEFAULT_INTERVAL_SECONDS,
  onRefresh: () => refresh(true),
  shouldPause: () => document.hidden || loading.value || capacityLoading.value,
})

// ── Computed ──
const overallStatus = computed<OverallStatus>(() => {
  if (items.value.length === 0) return 'operational'
  for (const it of items.value) {
    if (it.primary_status === 'failed' || it.primary_status === 'error') return 'degraded'
    if (it.primary_status !== STATUS_OPERATIONAL) return 'degraded'
  }
  return 'operational'
})

const detailTitle = computed(() => {
  return detailTarget.value?.name || t('channelStatus.detailTitle')
})

const accountShareEnabled = computed(() => appStore.cachedPublicSettings?.account_share_enabled !== false)
const accountShareChannelStatusVisible = computed(
  () => accountShareEnabled.value && appStore.cachedPublicSettings?.account_share_channel_status_visible !== false,
)
const externalCapacityReferenceFeatureEnabled = false
const externalCapacityReferenceEnabled = computed(
  () => externalCapacityReferenceFeatureEnabled
    && appStore.cachedPublicSettings?.external_capacity_reference_enabled === true,
)
const visibleCapacityPools = computed<UserAccountCapacityPools | null>(() => {
  if (!capacityPools.value) return null
  if (externalCapacityReferenceEnabled.value) return capacityPools.value
  return {
    ...capacityPools.value,
    external: null,
  }
})
const showSharedCapacityPool = computed(() => accountShareChannelStatusVisible.value && hasVisibleCapacityPool(visibleCapacityPools.value?.shared))
const showMineCapacityPool = computed(() => accountShareChannelStatusVisible.value && hasVisibleCapacityPool(visibleCapacityPools.value?.mine))
const showCapacityColumn = computed(() => showSharedCapacityPool.value || showMineCapacityPool.value)

// ── Loaders ──
async function reload(silent = false) {
  if (abortController) abortController.abort()
  const ctrl = new AbortController()
  abortController = ctrl
  if (!silent) loading.value = true
  try {
    const res = await listChannelMonitorViews({ signal: ctrl.signal })
    if (ctrl.signal.aborted || abortController !== ctrl) return
    items.value = res.items || []
    return true
  } catch (err: unknown) {
    const e = err as { name?: string; code?: string }
    if (e?.name === 'AbortError' || e?.code === 'ERR_CANCELED') return
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.loadError')))
  } finally {
    if (abortController === ctrl) {
      if (!silent) loading.value = false
      autoRefresh.countdown.value = DEFAULT_INTERVAL_SECONDS
      abortController = null
    }
  }
}

async function loadCapacityPools(silent = false) {
  if (!accountShareChannelStatusVisible.value) {
    if (capacityAbortController) capacityAbortController.abort()
    capacityPools.value = null
    capacityLoading.value = false
    return
  }
  if (capacityAbortController) capacityAbortController.abort()
  const ctrl = new AbortController()
  capacityAbortController = ctrl
  if (!silent) capacityLoading.value = true
  try {
    const res = await getAccountCapacityPools({ signal: ctrl.signal })
    if (ctrl.signal.aborted || capacityAbortController !== ctrl) return
    capacityPools.value = res
  } catch (err: unknown) {
    const e = err as { name?: string; code?: string }
    if (e?.name === 'AbortError' || e?.code === 'ERR_CANCELED') return
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.capacityPools.loadError')))
  } finally {
    if (capacityAbortController === ctrl) {
      if (!silent) capacityLoading.value = false
      capacityAbortController = null
    }
  }
}

async function refresh(silent = false) {
  if (unmounted) return
  const version = ++refreshVersion
  const [reloaded] = await Promise.all([reload(silent), loadCapacityPools(silent)])
  if (!reloaded || !isCurrentRefresh(version)) return
  // After base reload, refresh any cached detail records so non-7d availability
  // values stay in sync without forcing the user to switch tabs again.
  if (currentWindow.value !== '7d') {
    await Promise.all(items.value.map(it => loadDetail(it.id, true, version)))
  }
}

async function loadDetail(id: number, force = false, version?: number) {
  const isActive = () => !unmounted && (version === undefined || isCurrentRefresh(version))
  if (!isActive()) return
  if (!force && detailCache[id]) return
  // Ordinary window details survive refreshes (including 7d extra models).
  // Only details owned by a refresh expire when that refresh is superseded.
  const pending = pendingDetails.get(id)
  if (!force && pending && (pending.version === undefined || isCurrentRefresh(pending.version))) {
    return pending.promise
  }
  const request = { version, promise: Promise.resolve() }
  pendingDetails.set(id, request)
  const isCurrentDetail = () => isActive() && pendingDetails.get(id) === request
  request.promise = (async () => {
    try {
      const detail = await fetchChannelMonitorDetail(id)
      if (!isCurrentDetail()) return
      detailCache[id] = detail
    } catch (err: unknown) {
      if (!isCurrentDetail()) return
      appStore.showError(extractApiErrorMessage(err, t('channelStatus.detailLoadError')))
    } finally {
      if (pendingDetails.get(id) === request) pendingDetails.delete(id)
    }
  })()
  return request.promise
}

async function ensureDetailsForWindow() {
  const shouldLoadDetails = currentWindow.value !== '7d'
    || items.value.some(it => (it.extra_models?.length ?? 0) > 0)
  if (!shouldLoadDetails) return
  await Promise.all(items.value.map(it => loadDetail(it.id)))
}

// ── Handlers ──
async function handleWindowChange(value: MonitorWindow) {
  currentWindow.value = value
  await ensureDetailsForWindow()
}

function openDetail(row: UserMonitorView) {
  detailTarget.value = row
  showDetail.value = true
}

function closeDetail() {
  showDetail.value = false
  detailTarget.value = null
}

function hasVisibleCapacityPool(pool: UserAccountCapacityPool | null | undefined): boolean {
  if (!pool) return false
  return positiveCount(pool.total_accounts) > 0
    || (pool.sections ?? []).some(section => positiveCount(section.total_accounts) > 0)
    || (pool.groups ?? []).some(group => positiveCount(group.total_accounts) > 0)
}

function positiveCount(value?: number | null): number {
  if (!Number.isFinite(value ?? 0)) return 0
  return Math.max(0, Math.round(value ?? 0))
}

watch(items, () => {
  void ensureDetailsForWindow()
})

watch(
  () => appStore.cachedPublicSettings?.channel_monitor_enabled,
  (enabled) => {
    if (enabled === false) autoRefresh.stop()
    else if (autoRefresh.enabled.value) autoRefresh.start()
  },
)

watch(accountShareChannelStatusVisible, (enabled) => {
  if (enabled) {
    void loadCapacityPools(false)
  } else {
    if (capacityAbortController) capacityAbortController.abort()
    capacityPools.value = null
    capacityLoading.value = false
  }
})

onMounted(() => {
  void reload(false)
  void loadCapacityPools(false)
  if (appStore.cachedPublicSettings?.channel_monitor_enabled !== false) {
    autoRefresh.setEnabled(true)
  }
})

onBeforeUnmount(() => {
  unmounted = true
  refreshVersion++
  if (abortController) abortController.abort()
  if (capacityAbortController) capacityAbortController.abort()
})
</script>
