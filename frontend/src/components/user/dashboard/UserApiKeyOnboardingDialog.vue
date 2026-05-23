<template>
  <Transition name="api-key-onboarding">
    <section
      v-if="show"
      class="pointer-events-none fixed inset-0 z-40 flex items-center justify-center px-4 py-6"
      role="region"
      aria-live="polite"
      :aria-label="t('dashboard.onboarding.title')"
    >
      <div class="pointer-events-auto w-full max-w-[560px] overflow-hidden rounded-xl border border-blue-200 bg-white shadow-2xl shadow-blue-950/25">
        <div class="flex items-start justify-between gap-4 border-b border-blue-500/30 bg-gradient-to-r from-blue-600 to-sky-500 px-5 py-4 text-white">
          <div>
            <h3 class="text-base font-semibold text-white">
              {{ t('dashboard.onboarding.title') }}
            </h3>
            <p class="mt-1 text-sm leading-6 text-blue-50">
              {{ t('dashboard.onboarding.description') }}
            </p>
          </div>
          <button
            type="button"
            class="-mr-2 rounded-lg p-2 text-blue-100 transition-colors hover:bg-white/15 hover:text-white"
            aria-label="Close"
            @click="emit('skip')"
          >
            <Icon name="x" size="sm" />
          </button>
        </div>

        <div class="space-y-4 px-5 py-4">
          <div class="grid gap-3 sm:grid-cols-3">
            <div
              v-for="item in steps"
              :key="item.title"
              class="rounded-lg border border-blue-100 bg-blue-50 p-3"
            >
              <p class="text-sm font-semibold text-blue-950">{{ item.title }}</p>
              <p class="mt-1 text-xs leading-5 text-slate-600">{{ item.description }}</p>
            </div>
          </div>

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
        </div>
      </div>
    </section>
  </Transition>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

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

<style scoped>
.api-key-onboarding-enter-active,
.api-key-onboarding-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.api-key-onboarding-enter-from,
.api-key-onboarding-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>
