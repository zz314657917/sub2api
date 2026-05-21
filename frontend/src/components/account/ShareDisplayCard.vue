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

const displayNameModel = computed({
  get: () => props.displayName ?? '',
  set: (value: string) => emit('update:displayName', value),
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
  { value: '', label: t('admin.accounts.shareDisplay.autoPool') },
  { value: 'plus', label: 'OpenAI Plus' },
  { value: 'pro', label: 'OpenAI Pro' },
])

const targetPoolModel = computed({
  get: () => props.enabled ? (props.displayTier || 'pro') : '',
  set: (value: string) => {
    emit('update:enabled', value !== '')
    emit('update:displayTier', value || 'pro')
  },
})
</script>

<template>
  <div class="rounded-lg border border-gray-200 p-4 dark:border-dark-600">
    <div>
      <h3 class="input-label mb-0 text-base font-semibold">{{ t('admin.accounts.shareDisplay.title') }}</h3>
      <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
        {{ t('admin.accounts.shareDisplay.hint') }}
      </p>
    </div>

    <div class="mt-4 grid gap-4 md:grid-cols-2">
      <div>
        <label class="input-label">{{ t('admin.accounts.shareDisplay.displayTier') }}</label>
        <select
          v-model="targetPoolModel"
          class="input w-full"
          data-testid="share-display-target-pool"
        >
          <option v-for="option in tierOptions" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
        <p class="input-hint">{{ t('admin.accounts.shareDisplay.displayTierHint') }}</p>
      </div>
      <Input
        v-if="props.enabled"
        data-testid="share-display-name"
        v-model="displayNameModel"
        :label="t('admin.accounts.shareDisplay.displayName')"
        :placeholder="t('admin.accounts.shareDisplay.displayNamePlaceholder')"
      />
      <Input
        v-if="props.enabled"
        data-testid="share-display-account-count"
        v-model="accountCountModel"
        type="number"
        :label="t('admin.accounts.shareDisplay.accountCount')"
        :hint="t('admin.accounts.shareDisplay.accountCountHint')"
      />
      <label v-if="props.enabled" class="flex items-start gap-3 rounded-lg border border-gray-200 p-3 dark:border-dark-600 md:col-span-2">
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
