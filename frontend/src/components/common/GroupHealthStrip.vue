<template>
  <div class="flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1 text-[11px] leading-4" data-testid="group-health-strip">
    <div
      v-if="bars.length"
      class="flex h-2.5 items-stretch gap-[2px]"
      :style="{ width: `${bars.length * 6}px` }"
      aria-hidden="true"
    >
      <span
        v-for="(bar, idx) in bars"
        :key="idx"
        class="flex-1 rounded-[2px]"
        :class="bar.colorClass"
        :title="bar.title"
      />
    </div>
    <span v-if="health.availability !== null" class="whitespace-nowrap text-gray-500 dark:text-gray-400">
      {{ t('keys.groupHealth.availability') }}
      <span class="font-semibold tabular-nums" :style="{ color: hslForPct(health.availability) }">{{ formatPct(health.availability) }}</span>
    </span>
    <span
      v-if="health.cacheRate !== null"
      class="whitespace-nowrap text-gray-500 dark:text-gray-400"
      :title="t('keys.groupHealth.cacheRateHint')"
    >
      {{ t('keys.groupHealth.cacheRate') }}
      <span class="font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">{{ formatPct(health.cacheRate) }}</span>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { GroupHealth } from '@/composables/useGroupHealth'
import { hslForPct, useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

const props = withDefaults(defineProps<{
  health: GroupHealth
  /** Max probe points shown; newest at the right. */
  points?: number
}>(), {
  points: 20
})

const { t } = useI18n()
const { statusLabel, formatRelativeTime } = useChannelMonitorFormat()

const STATUS_COLOR: Record<string, string> = {
  operational: 'bg-emerald-500',
  degraded: 'bg-amber-400',
  failed: 'bg-red-500',
  error: 'bg-red-500'
}

const bars = computed(() =>
  props.health.timeline
    .slice(0, props.points)
    .reverse()
    .map((point) => ({
      colorClass: STATUS_COLOR[point.status] ?? 'bg-gray-300 dark:bg-dark-600',
      title: `${formatRelativeTime(point.checked_at)} · ${statusLabel(point.status)}`
    }))
)

const formatPct = (value: number) => `${value >= 99.95 ? '100' : value.toFixed(1)}%`
</script>
