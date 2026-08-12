<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import Toggle from '@/components/common/Toggle.vue'
import { serverTimezoneLabel, timeOfDayMinutes } from '@/utils/peak-rate'

const props = defineProps<{
  enabled: boolean
  start: string
  end: string
}>()

const emit = defineEmits<{
  'update:enabled': [value: boolean]
  'update:start': [value: string]
  'update:end': [value: string]
  valid: [value: boolean]
  'window-valid': [value: boolean]
}>()

const { t } = useI18n()
const appStore = useAppStore()

const serverTimezone = computed(() =>
  serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset, t('common.serverTime'))
)

const hasValidWindow = computed(() => {
  const start = timeOfDayMinutes(props.start)
  const end = timeOfDayMinutes(props.end)
  return start !== null && end !== null && start < end
})

const isValid = computed(() => !props.enabled || hasValidWindow.value)

watch(
  [isValid, hasValidWindow],
  ([valid, windowValid]) => {
    emit('valid', valid)
    emit('window-valid', windowValid)
  },
  { immediate: true }
)
</script>

<template>
  <section class="border-t border-gray-200 pt-4 dark:border-dark-600" data-testid="account-time-availability-section">
    <div class="flex items-start justify-between gap-4">
      <div>
        <label class="input-label mb-0" for="account-time-availability-enabled">
          {{ t('admin.accounts.timeAvailability.title') }}
        </label>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.accounts.timeAvailability.hint') }}
        </p>
      </div>
      <Toggle
        :model-value="enabled"
        id="account-time-availability-enabled"
        data-testid="account-time-availability-enabled"
        :aria-label="t('admin.accounts.timeAvailability.enable')"
        @update:model-value="emit('update:enabled', $event)"
      />
    </div>

    <div v-if="enabled" class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
      <div>
        <label class="input-label" for="account-time-availability-start">
          {{ t('admin.accounts.timeAvailability.start') }}
        </label>
        <input
          :value="start"
          type="time"
          class="input"
          :class="{ 'border-red-500 focus:border-red-500 focus:ring-red-500': !hasValidWindow }"
          data-testid="account-time-availability-start"
          id="account-time-availability-start"
          @input="emit('update:start', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div>
        <label class="input-label" for="account-time-availability-end">
          {{ t('admin.accounts.timeAvailability.end') }}
        </label>
        <input
          :value="end"
          type="time"
          class="input"
          :class="{ 'border-red-500 focus:border-red-500 focus:ring-red-500': !hasValidWindow }"
          data-testid="account-time-availability-end"
          id="account-time-availability-end"
          @input="emit('update:end', ($event.target as HTMLInputElement).value)"
        />
      </div>
    </div>

    <p v-if="enabled" class="mt-3 text-xs text-gray-500 dark:text-gray-400">
      {{ t('admin.accounts.timeAvailability.serverTimezone', { timezone: serverTimezone }) }}
    </p>
    <p v-if="enabled && !hasValidWindow" class="mt-2 text-xs text-red-600 dark:text-red-400" role="alert">
      {{ t('admin.accounts.timeAvailability.windowInvalid') }}
    </p>
  </section>
</template>
