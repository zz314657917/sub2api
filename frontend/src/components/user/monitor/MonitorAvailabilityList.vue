<template>
  <div
    class="overflow-hidden rounded-2xl border border-gray-200/80 bg-white/80 p-4 shadow-card backdrop-blur-xl dark:border-dark-700/70 dark:bg-dark-800/70 sm:p-5"
    data-testid="monitor-availability-list"
  >
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="inline-flex min-w-0 items-center gap-2">
        <span
          class="grid h-8 w-8 flex-shrink-0 place-items-center rounded-full bg-gray-100 text-gray-700 ring-1 ring-gray-200/80 dark:bg-dark-700 dark:text-gray-200 dark:ring-dark-600"
        >
          <Icon name="server" size="sm" />
        </span>
        <h2 class="truncate text-sm font-semibold text-gray-900 dark:text-white">
          {{ t('channelStatus.availabilityPanel.title') }}
        </h2>
      </div>

      <slot name="actions" />
    </div>

    <div class="relative mt-4">
      <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
        <Icon name="search" size="sm" class="text-gray-500 dark:text-gray-400" />
      </div>
      <input
        v-model="searchQuery"
        type="search"
        class="h-8 w-full rounded-full border border-transparent bg-gray-100/90 pl-9 pr-3 text-sm text-gray-800 outline-none transition focus:border-primary-400 focus:bg-white focus:ring-2 focus:ring-primary-500/20 dark:bg-dark-700/80 dark:text-gray-100 dark:focus:border-primary-400 dark:focus:bg-dark-800"
        :placeholder="t('channelStatus.availabilityPanel.searchPlaceholder')"
        :aria-label="t('channelStatus.availabilityPanel.searchPlaceholder')"
        data-testid="monitor-availability-search"
      />
    </div>

    <div
      v-if="loading && items.length === 0"
      class="mt-4 space-y-5"
    >
      <div
        v-for="i in 5"
        :key="i"
        class="animate-pulse"
      >
        <div class="h-4 w-40 rounded bg-gray-200 dark:bg-dark-700"></div>
        <div class="mt-3 space-y-3">
          <div class="h-3 w-3/4 rounded bg-gray-100 dark:bg-dark-700"></div>
          <div class="h-1 w-full rounded bg-gray-100 dark:bg-dark-700"></div>
        </div>
      </div>
    </div>

    <EmptyState
      v-else-if="items.length === 0"
      class="py-10"
      :title="t('channelStatus.empty.title')"
      :description="t('channelStatus.empty.description')"
    />

    <EmptyState
      v-else-if="groups.length === 0"
      class="py-10"
      :title="t('channelStatus.availabilityPanel.noResultsTitle')"
      :description="t('channelStatus.availabilityPanel.noResultsDescription')"
    />

    <div
      v-else
      class="mt-4 space-y-5"
      :class="{ 'monitor-availability-list__scroll': isScrollable }"
      data-testid="monitor-availability-scroll"
    >
      <section
        v-for="group in groups"
        :key="group.key"
      >
        <h3
          class="border-b border-gray-200/80 pb-2 text-sm font-semibold text-gray-700 dark:border-dark-700 dark:text-gray-200"
        >
          {{ group.title }}
        </h3>

        <div class="divide-y divide-gray-100 dark:divide-dark-700/70">
          <button
            v-for="row in group.rows"
            :key="row.key"
            type="button"
            class="group w-full rounded-lg py-3 text-left transition hover:bg-gray-50/80 focus:outline-none focus:ring-2 focus:ring-primary-500/30 dark:hover:bg-dark-700/50"
            data-testid="monitor-availability-row"
            @click="emit('rowClick', row.monitor)"
          >
            <div class="flex items-center gap-2">
              <span
                class="h-2 w-2 flex-shrink-0 rounded-full"
                :class="dotClass(row.tone)"
              ></span>
              <span
                class="min-w-0 flex-1 truncate text-sm font-medium text-gray-900 dark:text-gray-100"
              >
                {{ displayModelLabel(row.model) }}
              </span>
              <span
                class="flex-shrink-0 text-xs tabular-nums text-gray-500 dark:text-gray-400"
              >
                {{ formatPercent(row.availability) }}
              </span>
            </div>

            <div class="mt-1.5 flex items-center gap-2 pl-4">
              <span class="w-10 flex-shrink-0 text-xs text-gray-500 dark:text-gray-400">
                {{ t('channelStatus.availabilityPanel.availabilityLabel') }}
              </span>
              <span
                class="h-1 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700"
                role="progressbar"
                :aria-valuemin="0"
                :aria-valuemax="100"
                :aria-valuenow="row.availability ?? 0"
                :aria-label="`${displayModelLabel(row.model)} ${t('channelStatus.availabilityPanel.availabilityLabel')}`"
              >
                <span
                  class="block h-full rounded-full transition-all duration-300"
                  :class="barClass(row.tone)"
                  :style="{ width: `${barWidth(row.availability)}%` }"
                ></span>
              </span>
            </div>
          </button>
        </div>
      </section>
    </div>

    <div
      class="mt-5 flex flex-wrap items-center justify-center gap-x-4 gap-y-2 text-xs text-gray-600 dark:text-gray-400"
    >
      <span
        v-for="legend in legends"
        :key="legend.key"
        class="inline-flex items-center gap-1.5"
      >
        <span
          class="h-2 w-2 rounded-full"
          :class="legend.dot"
        ></span>
        {{ legend.label }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserMonitorDetail, UserMonitorView } from '@/api/channelMonitor'
import {
  STATUS_DEGRADED,
  STATUS_ERROR,
  STATUS_FAILED,
  STATUS_OPERATIONAL,
} from '@/constants/channelMonitor'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { cleanModelDisplayName, displayModelLabel } from '@/utils/modelDisplay'

type MonitorWindow = '7d' | '15d' | '30d'
type Tone = 'normal' | 'warning' | 'danger' | 'maintenance'

interface DisplayRow {
  key: string
  monitor: UserMonitorView
  model: string
  status: string
  availability: number | null
  tone: Tone
}

interface DisplayGroup {
  key: string
  title: string
  rows: DisplayRow[]
}

const props = defineProps<{
  items: UserMonitorView[]
  window: MonitorWindow
  loading: boolean
  detailCache: Record<number, UserMonitorDetail>
}>()

const emit = defineEmits<{
  (e: 'rowClick', item: UserMonitorView): void
}>()

const { t } = useI18n()
const { providerLabel, formatPercent } = useChannelMonitorFormat()
const searchQuery = ref('')
const MAX_INLINE_ROWS = 12

const legends = computed(() => [
  {
    key: 'danger',
    label: t('channelStatus.availabilityPanel.legend.abnormal'),
    dot: 'bg-red-500',
  },
  {
    key: 'normal',
    label: t('channelStatus.availabilityPanel.legend.normal'),
    dot: 'bg-emerald-500',
  },
  {
    key: 'warning',
    label: t('channelStatus.availabilityPanel.legend.highLatency'),
    dot: 'bg-amber-500',
  },
  {
    key: 'maintenance',
    label: t('channelStatus.availabilityPanel.legend.maintenance'),
    dot: 'bg-sky-500',
  },
])

const groups = computed<DisplayGroup[]>(() => {
  const map = new Map<string, DisplayGroup>()
  const query = searchQuery.value.trim().toLocaleLowerCase()

  for (const item of props.items) {
    const key = item.group_name || item.provider || 'default'
    const group = map.get(key) ?? {
      key,
      title: groupTitle(item),
      rows: [],
    }
    const rows = modelRowsFor(item).filter(row => matchesQuery(row, query))
    if (rows.length === 0) continue
    group.rows.push(...rows)
    if (group.rows.length > 0) {
      map.set(key, group)
    }
  }

  return [...map.values()]
})

const totalVisibleRows = computed(() => groups.value.reduce((total, group) => total + group.rows.length, 0))
const isScrollable = computed(() => totalVisibleRows.value > MAX_INLINE_ROWS)

function groupTitle(item: UserMonitorView): string {
  const base = item.group_name || providerLabel(item.provider)
  const suffix = t('channelStatus.availabilityPanel.groupSuffix')
  return base.includes(suffix) ? base : `${base} ${suffix}`
}

function modelRowsFor(item: UserMonitorView): DisplayRow[] {
  const rows: DisplayRow[] = []
  rows.push({
    key: `${item.id}:${item.primary_model}`,
    monitor: item,
    model: item.primary_model || item.name || '-',
    status: item.primary_status,
    availability: resolveAvailability(item, item.primary_model, item.availability_7d),
    tone: toneFor(item.primary_status),
  })

  for (const extra of item.extra_models ?? []) {
    rows.push({
      key: `${item.id}:${extra.model}`,
      monitor: item,
      model: extra.model,
      status: extra.status,
      availability: resolveAvailability(item, extra.model, extra.availability_7d ?? null),
      tone: toneFor(extra.status),
    })
  }

  return rows
}

function matchesQuery(row: DisplayRow, query: string): boolean {
  if (!query) return true
  const searchable = [
    cleanModelDisplayName(row.model, ''),
    row.monitor.name,
    row.monitor.group_name,
    providerLabel(row.monitor.provider),
  ]
    .filter(Boolean)
    .join(' ')
    .toLocaleLowerCase()
  return searchable.includes(query)
}

function resolveAvailability(
  item: UserMonitorView,
  model: string,
  fallback7d: number | null,
): number | null {
  const detail = props.detailCache[item.id]
  const modelDetail = detail?.models.find(m => m.model === model)
  if (props.window === '7d') {
    return fallback7d ?? modelDetail?.availability_7d ?? null
  }
  if (!modelDetail) return null
  return props.window === '15d'
    ? modelDetail.availability_15d ?? null
    : modelDetail.availability_30d ?? null
}

function toneFor(status: string): Tone {
  switch (status) {
    case STATUS_OPERATIONAL:
      return 'normal'
    case STATUS_DEGRADED:
      return 'warning'
    case STATUS_FAILED:
    case STATUS_ERROR:
      return 'danger'
    default:
      return 'maintenance'
  }
}

function dotClass(tone: Tone): string {
  switch (tone) {
    case 'danger':
      return 'bg-red-500'
    case 'warning':
      return 'bg-amber-500'
    case 'maintenance':
      return 'bg-sky-500'
    case 'normal':
    default:
      return 'bg-emerald-500'
  }
}

function barClass(tone: Tone): string {
  switch (tone) {
    case 'danger':
      return 'bg-red-500'
    case 'warning':
      return 'bg-amber-500'
    case 'maintenance':
      return 'bg-sky-500'
    case 'normal':
    default:
      return 'bg-emerald-500'
  }
}

function barWidth(value: number | null): number {
  if (value === null || Number.isNaN(value)) return 0
  return Math.max(0, Math.min(100, value))
}
</script>

<style scoped>
.monitor-availability-list__scroll {
  max-height: clamp(24rem, calc(100vh - 18rem), 40rem);
  overflow-y: auto;
  overscroll-behavior: contain;
  padding-right: 0.25rem;
  scrollbar-color: rgb(148 163 184 / 0.7) transparent;
  scrollbar-width: thin;
}

.monitor-availability-list__scroll::-webkit-scrollbar {
  width: 0.375rem;
}

.monitor-availability-list__scroll::-webkit-scrollbar-thumb {
  background-color: rgb(148 163 184 / 0.55);
  border-radius: 999px;
}
</style>
