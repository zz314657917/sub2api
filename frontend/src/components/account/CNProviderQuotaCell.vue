<template>
  <div v-if="visible" class="space-y-1">
    <div v-if="data?.success && data.tiers?.length" class="space-y-1">
      <div v-for="tier in data.tiers" :key="tier.window" class="flex items-center gap-1.5 text-[10px]">
        <span class="w-10 shrink-0 text-gray-500 dark:text-gray-400">{{ windowLabel(tier.window) }}</span>
        <div class="h-1.5 w-16 shrink-0 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
          <div class="h-full rounded-full transition-all" :class="utilizationColor(tier.used_percent)" :style="{ width: `${Math.min(100, Math.max(0, tier.used_percent))}%` }" />
        </div>
        <span :class="['shrink-0 font-medium', utilizationTextColor(tier.used_percent)]">{{ Math.round(tier.used_percent) }}%</span>
        <span v-if="tier.reset_at" class="truncate text-gray-400 dark:text-gray-500" :title="tier.reset_at">· {{ formatReset(tier.reset_at) }}</span>
      </div>
    </div>
    <div class="flex flex-wrap items-center gap-1.5">
      <button
        type="button"
        data-test="cn-provider-quota-probe"
        class="inline-flex items-center gap-0.5 whitespace-nowrap rounded px-1.5 py-0.5 text-[10px] font-medium leading-4 text-blue-600 transition-colors hover:bg-blue-50 disabled:cursor-not-allowed disabled:opacity-50 dark:text-blue-400 dark:hover:bg-blue-900/30"
        :disabled="loading"
        :title="t('admin.accounts.cnProviders.probeTooltip')"
        @click="handleProbe"
      >
        <svg class="h-2.5 w-2.5" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        {{ t('admin.accounts.cnProviders.probe') }}
      </button>
    </div>
    <div v-if="error" class="truncate text-[10px] text-red-600 dark:text-red-400" :title="error">{{ truncatedError }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { CNProviderQuotaProbeResult } from '@/api/admin/cnProviders'
import type { Account } from '@/types'
import { cnQuotaCellVisible } from './credentialsBuilder'

const props = defineProps<{ account: Account }>()
const { t } = useI18n()
const accountMode = () => {
  const mode = props.account.credentials?.account_mode
  return typeof mode === 'string' ? mode : ''
}
const visible = computed(() => cnQuotaCellVisible(props.account.platform, accountMode()))
const loading = ref(false)
const error = ref<string | null>(null)
const data = ref<CNProviderQuotaProbeResult | null>(null)
const SNAPSHOT_STALE_MS = 15 * 60 * 1000
const AUTO_PROBE_DEBOUNCE_MS = 5 * 60 * 1000
const lastAutoProbeAt = new Map<number, number>()

const readExtraNumber = (key: string): number | null => {
  const value = (props.account.extra as Record<string, unknown> | undefined)?.[key]
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}
const readExtraString = (key: string): string => {
  const value = (props.account.extra as Record<string, unknown> | undefined)?.[key]
  return typeof value === 'string' ? value : ''
}
const snapshotData = computed<CNProviderQuotaProbeResult | null>(() => {
  const platform = props.account.platform
  const used5h = readExtraNumber(`${platform}_5h_used_percent`)
  const usedWeekly = readExtraNumber(`${platform}_weekly_used_percent`)
  if (used5h == null && usedWeekly == null) return null
  const tiers: CNProviderQuotaProbeResult['tiers'] = []
  if (used5h != null) tiers.push({ window: '5h', used_percent: used5h, reset_at: readExtraString(`${platform}_5h_reset_at`) || undefined })
  if (usedWeekly != null) tiers.push({ window: 'weekly', used_percent: usedWeekly, reset_at: readExtraString(`${platform}_weekly_reset_at`) || undefined })
  return { success: true, tiers } as CNProviderQuotaProbeResult
})
const snapshotIsStale = computed(() => {
  const updatedAt = readExtraString(`${props.account.platform}_usage_updated_at`)
  if (!updatedAt) return true
  const timestamp = new Date(updatedAt).getTime()
  return Number.isNaN(timestamp) || Date.now() - timestamp > SNAPSHOT_STALE_MS
})

onMounted(() => {
  if (!visible.value) return
  data.value = snapshotData.value
  if (!snapshotIsStale.value) return
  const last = lastAutoProbeAt.get(props.account.id) ?? 0
  if (Date.now() - last < AUTO_PROBE_DEBOUNCE_MS) return
  lastAutoProbeAt.set(props.account.id, Date.now())
  handleProbe()
})

const extractError = (cause: unknown): string => {
  const err = cause as { message?: string; reason?: string; response?: { data?: { message?: string; error?: string } } }
  return err?.message || err?.reason || err?.response?.data?.message || err?.response?.data?.error || t('common.error')
}
const truncatedError = computed(() => error.value && error.value.length > 80 ? `${error.value.slice(0, 80)}...` : error.value || '')
const windowLabel = (window: string) => window === 'weekly' ? t('admin.accounts.cnProviders.windowWeekly') : t('admin.accounts.cnProviders.window5h')
const utilizationColor = (pct: number) => pct >= 90 ? 'bg-red-500' : pct >= 75 ? 'bg-amber-500' : 'bg-emerald-500'
const utilizationTextColor = (pct: number) => pct >= 90 ? 'text-red-600 dark:text-red-400' : pct >= 75 ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'
const formatReset = (iso: string) => {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return iso
  const diffMs = date.getTime() - Date.now()
  if (diffMs <= 0) return t('admin.accounts.cnProviders.resetSoon')
  if (diffMs < 3_600_000) return `${Math.max(1, Math.round(diffMs / 60_000))}m`
  const hours = Math.round(diffMs / 3_600_000)
  if (hours < 48) return `${hours}h`
  return `${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const handleProbe = async () => {
  if (loading.value) return
  loading.value = true
  error.value = null
  try {
    const result = await adminAPI.cnProviders.queryQuota(props.account.id)
    if (result.success) data.value = result
    else error.value = result.error || t('common.error')
  } catch (cause) {
    error.value = extractError(cause)
  } finally {
    loading.value = false
  }
}

watch(() => props.account.id, () => {
  data.value = null
  error.value = null
  loading.value = false
})
</script>
