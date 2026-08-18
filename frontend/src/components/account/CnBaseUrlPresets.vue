<template>
  <div class="flex flex-wrap gap-2">
    <button
      v-for="preset in presets"
      :key="preset.mode + ':' + preset.protocol + ':' + preset.url"
      type="button"
      data-testid="cn-base-url-preset"
      :class="[
        'rounded-lg px-3 py-1 text-xs transition-colors',
        isActive(preset)
          ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
          : 'bg-gray-100 text-gray-700 hover:bg-primary-50 hover:text-primary-700 dark:bg-dark-600 dark:text-gray-300 dark:hover:bg-primary-900/30 dark:hover:text-primary-400'
      ]"
      @click="emit('select', preset)"
    >
      {{ preset.label }} ({{ displayUrl(preset.url) }})
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { CN_BASE_URL_PRESETS, type CnBaseUrlPreset } from './credentialsBuilder'

const props = defineProps<{
  platform: 'kimi' | 'zhipu' | 'deepseek'
  mode?: 'payg' | 'coding'
  protocol?: 'chat_completions' | 'anthropic' | 'responses'
  currentUrl?: string
}>()

const emit = defineEmits<{
  (e: 'select', preset: CnBaseUrlPreset): void
}>()

const presets = computed(() => {
  const all = CN_BASE_URL_PRESETS[props.platform] ?? []
  if (props.protocol == null) return all
  return all.filter(preset => preset.protocol === props.protocol)
})

const isActive = (preset: CnBaseUrlPreset) =>
  (props.mode != null && preset.mode === props.mode) || preset.url === props.currentUrl

const displayUrl = (url: string) => url.replace(/^https?:\/\//i, '')
</script>
