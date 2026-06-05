<template>
  <div>
    <label v-if="title !== null" class="input-label mb-0">{{ title || t('admin.accounts.supportedCapabilities.title') }}</label>
    <p class="text-xs text-gray-500 dark:text-gray-400" :class="title !== null ? 'mt-1' : ''">
      {{ hint || t('admin.accounts.supportedCapabilities.hint') }}
    </p>
    <div class="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-4">
      <label
        v-for="option in capabilityOptions"
        :key="option.value"
        :class="[
          'flex min-h-[44px] cursor-pointer items-center gap-2 rounded-lg border px-3 py-2 text-sm transition-colors',
          isSelected(option.value)
            ? 'border-primary-300 bg-primary-50 text-primary-700 dark:border-primary-700 dark:bg-primary-900/20 dark:text-primary-300'
            : 'border-gray-200 text-gray-700 hover:border-gray-300 hover:bg-gray-50 dark:border-dark-600 dark:text-gray-200 dark:hover:border-dark-500 dark:hover:bg-dark-700',
          disabled && 'cursor-not-allowed opacity-60'
        ]"
      >
        <input
          type="checkbox"
          class="rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-dark-500"
          :checked="isSelected(option.value)"
          :disabled="disabled"
          @change="toggleCapability(option.value)"
        />
        <span>{{ option.label }}</span>
      </label>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AccountCapability } from '@/types'

const props = defineProps<{
  modelValue: AccountCapability[]
  disabled?: boolean
  title?: string | null
  hint?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: AccountCapability[]]
}>()

const { t } = useI18n()

const capabilityOptions = computed<{ value: AccountCapability; label: string }[]>(() => [
  { value: 'chat', label: t('admin.accounts.supportedCapabilities.chat') },
  { value: 'image', label: t('admin.accounts.supportedCapabilities.image') },
  { value: 'video', label: t('admin.accounts.supportedCapabilities.video') },
  { value: 'embedding', label: t('admin.accounts.supportedCapabilities.embedding') }
])

const isSelected = (capability: AccountCapability) => props.modelValue.includes(capability)

const toggleCapability = (capability: AccountCapability) => {
  if (props.disabled) return
  if (props.modelValue.includes(capability)) {
    emit('update:modelValue', props.modelValue.filter((value) => value !== capability))
    return
  }
  emit('update:modelValue', [...props.modelValue, capability])
}
</script>
