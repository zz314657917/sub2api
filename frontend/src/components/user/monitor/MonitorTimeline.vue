<template>
  <div
    class="mt-3 min-w-0 w-full pt-3 border-t border-gray-100 dark:border-dark-700/60"
    data-testid="monitor-timeline"
  >
    <div
      class="flex justify-between text-[10px] font-semibold uppercase tracking-widest text-gray-400 mb-2"
    >
      <span>{{ t('monitorCommon.history60pts', { n: length }) }}</span>
      <span class="tabular-nums">{{ t('monitorCommon.nextUpdateIn', { n: countdownSeconds }) }}</span>
    </div>

    <div
      v-if="maintenance"
      class="flex h-5 w-full items-center justify-center rounded border border-dashed border-gray-300 dark:border-dark-600 text-[10px] uppercase tracking-widest text-gray-400"
    >
      {{ t('monitorCommon.maintenancePaused') }}
    </div>
    <div
      v-else
      class="flex items-stretch gap-[2px] h-5 w-full"
      data-testid="monitor-timeline-bars"
    >
      <div
        v-for="(bar, idx) in displayBars"
        :key="idx"
        class="flex-1 min-w-0 h-full rounded-sm transition-opacity hover:opacity-70"
        :class="bar.colorClass"
        :title="bar.title"
        data-testid="monitor-timeline-bar"
      ></div>
    </div>

    <div
      class="mt-1 flex justify-between text-[9px] uppercase tracking-widest text-gray-400"
    >
      <span>{{ t('monitorCommon.past') }}</span>
      <span>{{ t('monitorCommon.now') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MonitorTimelinePoint } from '@/api/channelMonitor'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

const props = withDefaults(defineProps<{
  buckets?: MonitorTimelinePoint[]
  countdownSeconds: number
  length?: number
  maintenance?: boolean
}>(), {
  buckets: () => [],
  length: 60,
  maintenance: false,
})

const { t } = useI18n()
const { statusLabel, formatLatency, formatRelativeTime } = useChannelMonitorFormat()

interface Bar {
  colorClass: string
  title: string
}

// Equal-height segments show probe status through colour; height is not latency.
const STATUS_COLOR: Record<string, string> = {
  operational: 'bg-emerald-500',
  degraded: 'bg-orange-500',
  failed: 'bg-red-500',
  error: 'bg-red-500',
  empty: 'bg-gray-300 dark:bg-dark-600',
}

const displayBars = computed<Bar[]>(() => {
  // Real points come newest-first; convert to oldest-first so the rightmost
  // bar represents "now". Pad the left with empty placeholders to keep the
  // bar count stable at `length`.
  const real = [...(props.buckets ?? [])]
    .slice(0, props.length)
    .reverse()

  const padCount = Math.max(0, props.length - real.length)
  const bars: Bar[] = []

  for (let i = 0; i < padCount; i += 1) {
    bars.push({
      colorClass: STATUS_COLOR.empty,
      title: '',
    })
  }

  for (const point of real) {
    const status = point.status
    const colorClass = STATUS_COLOR[status] ?? STATUS_COLOR.empty
    const latency = formatLatency(point.latency_ms)
    const relative = formatRelativeTime(point.checked_at)
    const label = statusLabel(point.status)
    bars.push({
      colorClass,
      title: `${relative} · ${label} · ${latency}ms`,
    })
  }

  return bars
})
</script>
