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
  display5hLimit?: number | null
  display5hUsed?: number | null
  display7dLimit?: number | null
  display7dUsed?: number | null
}>(), {
  displayName: '',
  displayTier: 'pro',
  percentOnly: true,
  accountCount: 1,
  display5hLimit: null,
  display5hUsed: null,
  display7dLimit: null,
  display7dUsed: null,
})

const emit = defineEmits<{
  'update:enabled': [value: boolean]
  'update:displayName': [value: string]
  'update:displayTier': [value: string]
  'update:percentOnly': [value: boolean]
  'update:accountCount': [value: number]
  'update:display5hLimit': [value: number | null]
  'update:display5hUsed': [value: number | null]
  'update:display7dLimit': [value: number | null]
  'update:display7dUsed': [value: number | null]
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

function normalizeOptionalNumber(value: string): number | null {
  const trimmed = value.trim()
  if (!trimmed) return null
  const parsed = Number.parseFloat(trimmed)
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : null
}

function emitOptionalNumber(
  update: 'update:display5hLimit' | 'update:display5hUsed' | 'update:display7dLimit' | 'update:display7dUsed',
  value: number | null,
): void {
  switch (update) {
    case 'update:display5hLimit':
      emit('update:display5hLimit', value)
      return
    case 'update:display5hUsed':
      emit('update:display5hUsed', value)
      return
    case 'update:display7dLimit':
      emit('update:display7dLimit', value)
      return
    case 'update:display7dUsed':
      emit('update:display7dUsed', value)
      return
  }
}

function optionalNumberModel(
  getValue: () => number | null | undefined,
  update: 'update:display5hLimit' | 'update:display5hUsed' | 'update:display7dLimit' | 'update:display7dUsed',
) {
  return computed({
    get: () => {
      const value = getValue()
      return typeof value === 'number' && Number.isFinite(value) ? String(value) : ''
    },
    set: (value: string) => emitOptionalNumber(update, normalizeOptionalNumber(value)),
  })
}

const display5hLimitModel = optionalNumberModel(() => props.display5hLimit, 'update:display5hLimit')
const display5hUsedModel = optionalNumberModel(() => props.display5hUsed, 'update:display5hUsed')
const display7dLimitModel = optionalNumberModel(() => props.display7dLimit, 'update:display7dLimit')
const display7dUsedModel = optionalNumberModel(() => props.display7dUsed, 'update:display7dUsed')

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
      <template v-if="props.enabled">
        <Input
          data-testid="share-display-5h-limit"
          v-model="display5hLimitModel"
          type="number"
          :label="t('admin.accounts.shareDisplay.display5hLimit')"
          :placeholder="t('admin.accounts.shareDisplay.windowLimitPlaceholder')"
          :hint="t('admin.accounts.shareDisplay.display5hLimitHint')"
        />
        <Input
          data-testid="share-display-5h-used"
          v-model="display5hUsedModel"
          type="number"
          :label="t('admin.accounts.shareDisplay.display5hUsed')"
          :placeholder="t('admin.accounts.shareDisplay.windowUsedPlaceholder')"
          :hint="t('admin.accounts.shareDisplay.display5hUsedHint')"
        />
        <Input
          data-testid="share-display-7d-limit"
          v-model="display7dLimitModel"
          type="number"
          :label="t('admin.accounts.shareDisplay.display7dLimit')"
          :placeholder="t('admin.accounts.shareDisplay.windowLimitPlaceholder')"
          :hint="t('admin.accounts.shareDisplay.display7dLimitHint')"
        />
        <Input
          data-testid="share-display-7d-used"
          v-model="display7dUsedModel"
          type="number"
          :label="t('admin.accounts.shareDisplay.display7dUsed')"
          :placeholder="t('admin.accounts.shareDisplay.windowUsedPlaceholder')"
          :hint="t('admin.accounts.shareDisplay.display7dUsedHint')"
        />
      </template>
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
