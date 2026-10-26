<template>
  <button
    type="button"
    class="monitor-card group text-left p-4 rounded-xl h-full min-w-0 w-full bg-white/70 backdrop-blur-xl border border-gray-200/80 shadow-card dark:bg-dark-800/60 dark:border-dark-700/70 hover:shadow-card-hover dark:hover:border-primary-500/30 hover:border-gray-300 transition-colors duration-200 flex flex-col"
    @click="emit('click')"
  >
    <div class="monitor-row-summary">
      <!-- Channel identity stays separate from the compact metric columns. -->
      <div class="grid grid-cols-[2.25rem_minmax(0,1fr)_auto] items-start gap-x-3 gap-y-1 w-full min-w-0">
        <span
          class="row-span-2 w-9 h-9 rounded-xl ring-1 ring-black/5 dark:ring-white/10 grid place-items-center"
          :class="[providerGradient(item.provider), providerTintClass]"
        >
          <ProviderIcon :provider="item.provider" :size="20" />
        </span>
        <div class="min-w-0 text-sm font-semibold truncate text-gray-900 dark:text-gray-100" :title="item.name">
          {{ item.name }}
        </div>
        <span
          class="px-2 py-0.5 rounded-full text-[11px] font-semibold whitespace-nowrap"
          :class="statusBadgeClass(item.primary_status)"
        >
          {{ statusLabel(item.primary_status) }}
        </span>
        <div class="col-start-2 col-span-2 flex items-center gap-1.5 min-w-0">
          <span
            class="inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium flex-shrink-0"
            :class="providerBadgeClass(item.provider)"
          >
            {{ providerLabel(item.provider) }}
          </span>
          <span class="min-w-0 flex-1 font-mono text-[10px] truncate text-gray-500 dark:text-gray-400" :title="displayModelLabel(item.primary_model)">
            {{ displayModelLabel(item.primary_model) }}
          </span>
          <span
            v-if="item.group_name"
            class="max-w-[30%] truncate rounded-md px-1.5 py-0.5 text-[10px] font-medium bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300"
            :title="item.group_name"
          >
            {{ item.group_name }}
          </span>
          <span v-if="extraModelsCountLabel" class="flex-shrink-0 text-[10px] text-gray-400">
            {{ extraModelsCountLabel }}
          </span>
        </div>
      </div>

      <div class="flex items-center gap-4 min-w-0">
        <MonitorMetricPair
          class="flex-1"
          primary-icon="bolt"
          :primary-label="t('monitorCommon.dialogLatency')"
          :primary-value="formatLatency(item.primary_latency_ms)"
          primary-unit="ms"
          secondary-icon="globe"
          :secondary-label="t('monitorCommon.endpointPing')"
          :secondary-value="formatLatency(item.primary_ping_latency_ms)"
          secondary-unit="ms"
        />
        <MonitorAvailabilityRow
          class="shrink-0 border-l border-gray-200/80 dark:border-dark-700/60 pl-4"
          :window-label="availabilityLabel"
          :value="availabilityValue"
        />
      </div>
    </div>

    <!-- Timeline -->
    <div class="mt-auto w-full min-w-0">
      <MonitorTimeline
        :buckets="item.timeline"
        :countdown-seconds="countdownSeconds"
      />
    </div>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserMonitorView } from '@/api/channelMonitor'
import {
  useChannelMonitorFormat,
  providerGradient,
} from '@/composables/useChannelMonitorFormat'
import ProviderIcon from './ProviderIcon.vue'
import MonitorMetricPair from './MonitorMetricPair.vue'
import MonitorAvailabilityRow from './MonitorAvailabilityRow.vue'
import MonitorTimeline from './MonitorTimeline.vue'
import { displayModelLabel } from '@/utils/modelDisplay'

const PROVIDER_TINT: Record<string, string> = {
  openai: 'text-[#a9583e] dark:text-[#f0b89e]',
  anthropic: 'text-orange-600 dark:text-orange-300',
  gemini: 'text-[#6c6a64] dark:text-[#d8cec2]',
}

const props = defineProps<{
  item: UserMonitorView
  window: '7d' | '15d' | '30d'
  availabilityValue: number | null
  countdownSeconds: number
}>()

const emit = defineEmits<{
  (e: 'click'): void
}>()

const { t } = useI18n()
const {
  statusLabel,
  statusBadgeClass,
  providerLabel,
  providerBadgeClass,
  formatLatency,
} = useChannelMonitorFormat()

const providerTintClass = computed(() =>
  PROVIDER_TINT[props.item.provider] ?? 'text-gray-500 dark:text-gray-300'
)

const availabilityLabel = computed(() => {
  const win = t(`channelStatus.windowTab.${props.window}`)
  return `${t('monitorCommon.availabilityPrefix')} · ${win}`
})

const extraModelsCountLabel = computed(() => {
  const count = props.item.extra_models?.length ?? 0
  if (count === 0) return undefined
  return t('monitorCommon.extraModelsCount', { n: count })
})
</script>

<style scoped>
.monitor-card {
  container-type: inline-size;
}

.monitor-row-summary {
  display: grid;
  align-items: center;
  gap: 1rem;
}

@container (min-width: 560px) {
  .monitor-row-summary {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.15fr);
  }
}
</style>
