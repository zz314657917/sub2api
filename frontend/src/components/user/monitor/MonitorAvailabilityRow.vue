<template>
  <div class="text-right">
    <div class="text-[10px] text-gray-400 whitespace-nowrap">
      {{ windowLabel }}
    </div>
    <div class="mt-1 flex items-baseline justify-end gap-0.5">
      <span
        class="text-xl font-bold tabular-nums leading-none"
        :style="colorStyle"
      >
        {{ displayValue }}
      </span>
      <span
        class="text-xs font-semibold leading-none"
        :style="colorStyle"
      >%</span>
    </div>
    <div v-if="samplesLabel" class="mt-1 text-[10px] text-gray-400">
      {{ samplesLabel }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { hslForPct } from '@/composables/useChannelMonitorFormat'

const props = defineProps<{
  windowLabel: string
  value: number | null
  samplesLabel?: string
}>()

const { t } = useI18n()

const displayValue = computed(() => {
  if (props.value === null || Number.isNaN(props.value)) return t('monitorCommon.latencyEmpty')
  return props.value.toFixed(2)
})

const colorStyle = computed(() => {
  const colour = hslForPct(props.value)
  return colour ? { color: colour } : { color: 'rgb(156 163 175)' }
})
</script>
