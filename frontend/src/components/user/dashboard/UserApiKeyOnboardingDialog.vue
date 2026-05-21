<template>
  <BaseDialog
    :show="show"
    :title="t('dashboard.onboarding.title')"
    width="normal"
    :close-on-click-outside="true"
    @close="emit('skip')"
  >
    <div class="space-y-5">
      <div class="rounded-lg border border-blue-100 bg-blue-50 p-4 dark:border-blue-500/20 dark:bg-blue-500/10">
        <p class="text-sm leading-6 text-gray-700 dark:text-gray-300">
          {{ t('dashboard.onboarding.description') }}
        </p>
      </div>

      <div class="grid gap-3 sm:grid-cols-3">
        <div
          v-for="item in steps"
          :key="item.title"
          class="rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-800"
        >
          <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ item.title }}</p>
          <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">{{ item.description }}</p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <button type="button" class="btn btn-secondary" @click="emit('skip')">
          {{ t('dashboard.onboarding.skip') }}
        </button>
        <button type="button" class="btn btn-secondary" @click="emit('tutorial')">
          {{ t('dashboard.onboarding.viewTutorial') }}
        </button>
        <button type="button" class="btn btn-primary" @click="emit('create')">
          {{ t('dashboard.onboarding.createKey') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'

defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'create'): void
  (e: 'tutorial'): void
  (e: 'skip'): void
}>()

const { t } = useI18n()

const steps = computed(() => [
  {
    title: t('dashboard.onboarding.stepKeyTitle'),
    description: t('dashboard.onboarding.stepKeyDescription')
  },
  {
    title: t('dashboard.onboarding.stepToolTitle'),
    description: t('dashboard.onboarding.stepToolDescription')
  },
  {
    title: t('dashboard.onboarding.stepUsageTitle'),
    description: t('dashboard.onboarding.stepUsageDescription')
  }
])
</script>
