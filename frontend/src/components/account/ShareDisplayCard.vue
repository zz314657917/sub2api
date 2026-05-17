<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Input from '@/components/common/Input.vue'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  enabled: boolean
  displayName?: string | null
  displayTier?: string | null
  percentOnly?: boolean | null
  accountCount?: number | null
}>(), {
  displayName: '',
  displayTier: 'pro',
  percentOnly: true,
  accountCount: 1,
})

const emit = defineEmits<{
  'update:enabled': [value: boolean]
  'update:displayName': [value: string]
  'update:displayTier': [value: string]
  'update:percentOnly': [value: boolean]
  'update:accountCount': [value: number]
}>()

const enabledModel = computed({
  get: () => props.enabled,
  set: (value: boolean) => emit('update:enabled', value),
})

const displayNameModel = computed({
  get: () => props.displayName ?? '',
  set: (value: string) => emit('update:displayName', value),
})

const displayTierModel = computed({
  get: () => props.displayTier || 'pro',
  set: (value: string) => emit('update:displayTier', value),
})

const percentOnlyModel = computed({
  get: () => props.percentOnly !== false,
  set: (value: boolean) => emit('update:percentOnly', value),
})

const accountCountModel = computed({
  get: () => String(Math.max(1, Math.trunc(props.accountCount ?? 1))),
  set: (value: string) => {
    const parsed = Number.parseInt(value, 10)
    emit('update:accountCount', Number.isFinite(parsed) && parsed > 0 ? parsed : 1)
  },
})

const tierOptions = computed(() => [
  { value: 'plus', label: 'Plus' },
  { value: 'pro', label: 'Pro' },
  { value: 'team', label: 'Team' },
])
</script>

<template>
  <div class="rounded-lg border border-gray-200 p-4 dark:border-dark-600">
    <div class="flex items-start justify-between gap-3">
      <div>
        <h3 class="input-label mb-0 text-base font-semibold">{{ t('admin.accounts.shareDisplay.title') }}</h3>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.accounts.shareDisplay.hint') }}
        </p>
      </div>
      <button
        data-testid="share-display-toggle"
        type="button"
        @click="enabledModel = !enabledModel"
        :class="[
          'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
          enabledModel ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
        ]"
      >
        <span
          :class="[
            'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
            enabledModel ? 'translate-x-5' : 'translate-x-0'
          ]"
        />
      </button>
    </div>

    <div v-if="enabledModel" class="mt-4 grid gap-4 md:grid-cols-2">
      <Input
        data-testid="share-display-name"
        v-model="displayNameModel"
        :label="t('admin.accounts.shareDisplay.displayName')"
        :placeholder="t('admin.accounts.shareDisplay.displayNamePlaceholder')"
      />
      <div>
        <label class="input-label">{{ t('admin.accounts.shareDisplay.displayTier') }}</label>
        <select v-model="displayTierModel" class="input w-full">
          <option v-for="option in tierOptions" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
      </div>
      <Input
        data-testid="share-display-account-count"
        v-model="accountCountModel"
        type="number"
        :label="t('admin.accounts.shareDisplay.accountCount')"
        :hint="t('admin.accounts.shareDisplay.accountCountHint')"
      />
      <label class="flex items-start gap-3 rounded-lg border border-gray-200 p-3 dark:border-dark-600 md:col-span-2">
        <input v-model="percentOnlyModel" type="checkbox" class="mt-1 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
        <span>
          <span class="block text-sm font-medium text-gray-800 dark:text-gray-100">
            {{ t('admin.accounts.shareDisplay.percentOnly') }}
          </span>
          <span class="mt-1 block text-xs text-gray-500 dark:text-gray-400">
            {{ t('admin.accounts.shareDisplay.percentOnlyHint') }}
          </span>
        </span>
      </label>
    </div>
  </div>
</template>
