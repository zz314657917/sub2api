<template>
  <div class="min-w-0 grid grid-cols-2 gap-3">
    <div
      class="min-w-0"
    >
      <div
        class="flex items-center gap-1 text-[10px] font-semibold text-gray-400 whitespace-nowrap"
      >
        <Icon :name="primaryIcon" size="xs" />
        <span>{{ primaryLabel }}</span>
      </div>
      <div class="mt-1 text-base font-bold font-mono tabular-nums text-gray-900 dark:text-gray-100 whitespace-nowrap">
        {{ primaryValue }}<span class="text-xs font-normal text-gray-400 ml-0.5">{{ primaryUnit }}</span>
      </div>
    </div>
    <div
      class="min-w-0"
    >
      <div
        class="flex items-center gap-1 text-[10px] font-semibold text-gray-400 whitespace-nowrap"
      >
        <Icon :name="secondaryIcon" size="xs" />
        <span>{{ secondaryLabel }}</span>
      </div>
      <div class="mt-1 text-base font-bold font-mono tabular-nums text-gray-900 dark:text-gray-100 whitespace-nowrap">
        {{ secondaryValue }}<span class="text-xs font-normal text-gray-400 ml-0.5">{{ secondaryUnit }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  primaryLabel: string
  primaryValue: string
  primaryUnit: string
  primaryIcon: 'bolt' | 'globe' | 'clock' | 'link'
  secondaryLabel: string
  secondaryValue: string
  secondaryUnit: string
  secondaryIcon: 'bolt' | 'globe' | 'clock' | 'link'
}>()
</script>
