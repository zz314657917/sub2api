<template>
  <Transition name="api-key-onboarding">
    <section
      v-if="show"
      class="pointer-events-none fixed inset-0 z-40 flex items-center justify-center bg-slate-950/35 px-4 py-6 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-live="polite"
      :aria-label="dialogTitle"
    >
      <div class="pointer-events-auto w-full max-w-[720px] overflow-hidden rounded-xl border border-slate-700 bg-slate-950 text-white shadow-2xl shadow-cyan-950/30">
        <div class="relative min-h-[292px] overflow-hidden">
          <img
            src="/onboarding/new-user-trial-popup-header.png"
            alt=""
            aria-hidden="true"
            class="absolute inset-0 h-full w-full object-cover"
          >
          <div class="absolute inset-0 bg-gradient-to-r from-slate-950 via-slate-950/92 to-slate-950/20"></div>
          <button
            type="button"
            class="absolute right-4 top-4 z-10 rounded-lg p-2 text-slate-300 transition-colors hover:bg-white/10 hover:text-white"
            aria-label="Close"
            @click="emit('skip')"
          >
            <Icon name="x" size="sm" />
          </button>

          <div class="relative flex min-h-[292px] max-w-[28rem] flex-col justify-between px-5 py-5 sm:px-6 sm:py-6">
            <div>
              <div class="inline-flex items-center gap-2 rounded-full border border-cyan-300/30 bg-cyan-300/10 px-3 py-1 text-xs font-semibold text-cyan-100">
                <Icon name="sparkles" size="xs" :stroke-width="2" />
                {{ badgeText }}
              </div>
              <h3 class="mt-4 text-2xl font-bold leading-tight text-white sm:text-3xl">
                {{ dialogTitle }}
              </h3>
              <p class="mt-3 text-sm leading-6 text-slate-200 sm:text-base">
                {{ dialogDescription }}
              </p>
              <p v-if="hasBenefit" class="mt-3 text-xs leading-5 text-cyan-100/80">
                {{ benefitNotice }}
              </p>
            </div>

            <div class="mt-5 flex flex-wrap gap-2">
              <span
                v-for="pill in pills"
                :key="pill"
                class="rounded-full border border-white/10 bg-white/10 px-3 py-1 text-xs font-medium text-slate-100"
              >
                {{ pill }}
              </span>
            </div>
          </div>
        </div>

        <div class="space-y-4 border-t border-slate-800 bg-white px-5 py-4 text-slate-900 dark:bg-dark-900 dark:text-white">
          <div class="grid gap-3 sm:grid-cols-3">
            <div
              v-for="item in steps"
              :key="item.title"
              class="rounded-lg border border-slate-200 bg-slate-50 p-3 dark:border-dark-700 dark:bg-dark-800"
            >
              <div class="flex items-center gap-2 text-sm font-semibold text-slate-950 dark:text-white">
                <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-cyan-50 text-cyan-700 dark:bg-cyan-500/10 dark:text-cyan-300">
                  <Icon :name="item.icon" size="xs" :stroke-width="2" />
                </span>
                {{ item.title }}
              </div>
              <p class="mt-2 text-xs leading-5 text-slate-600 dark:text-slate-300">{{ item.description }}</p>
            </div>
          </div>

          <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <button type="button" class="btn btn-secondary" @click="emit('skip')">
              {{ t('dashboard.onboarding.skip') }}
            </button>
            <button type="button" class="btn btn-secondary" @click="openSupportPopup">
              {{ t('dashboard.onboarding.joinGroup') }}
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
import { openSupportPopup } from '@/utils/supportPopup'

const props = defineProps<{
  show: boolean
  hasBenefit?: boolean
  benefitLabel?: string
  benefitRewardLabel?: string
  benefitKind?: 'wallet' | 'trial'
}>()

const emit = defineEmits<{
  (e: 'create'): void
  (e: 'tutorial'): void
  (e: 'skip'): void
}>()

const { t } = useI18n()

const hasBenefit = computed(() => Boolean(props.hasBenefit))
const isWalletBenefit = computed(() => props.benefitKind === 'wallet')
const quotaLabel = computed(() => props.benefitLabel || t(isWalletBenefit.value ? 'dashboard.onboarding.balanceFallback' : 'dashboard.onboarding.trialQuotaFallback'))
const rewardLabel = computed(() => props.benefitRewardLabel || '')
const benefitNotice = computed(() => (
  isWalletBenefit.value
    ? t('dashboard.onboarding.walletBalanceNotice')
    : t('dashboard.onboarding.walletNotice')
))

const dialogTitle = computed(() => (
  hasBenefit.value
    ? t('dashboard.onboarding.trialTitle', { amount: quotaLabel.value })
    : t('dashboard.onboarding.title')
))
const dialogDescription = computed(() => (
  hasBenefit.value
    ? t(isWalletBenefit.value ? 'dashboard.onboarding.balanceDescription' : 'dashboard.onboarding.trialDescription')
    : t('dashboard.onboarding.description')
))
const badgeText = computed(() => (
  hasBenefit.value
    ? t('dashboard.onboarding.trialBadge')
    : t('dashboard.onboarding.badge')
))

const pills = computed(() => (
  hasBenefit.value
    ? [
      quotaLabel.value,
      t(isWalletBenefit.value ? 'dashboard.onboarding.pillBalanceDeduct' : 'dashboard.onboarding.pillAutoActivate'),
      t('dashboard.onboarding.pillNoRecharge')
    ]
    : [
      t('dashboard.onboarding.stepKeyTitle'),
      t('dashboard.onboarding.stepToolTitle'),
      t('dashboard.onboarding.stepUsageTitle')
    ]
))

const steps = computed(() => [
  {
    icon: 'key' as const,
    title: t('dashboard.onboarding.stepKeyTitle'),
    description: t('dashboard.onboarding.stepKeyDescription')
  },
  {
    icon: 'bolt' as const,
    title: hasBenefit.value
      ? t('dashboard.onboarding.stepTrialTitle')
      : t('dashboard.onboarding.stepToolTitle'),
    description: hasBenefit.value
      ? t(isWalletBenefit.value ? 'dashboard.onboarding.stepBalanceDescription' : 'dashboard.onboarding.stepTrialDescription')
      : t('dashboard.onboarding.stepToolDescription')
  },
  {
    icon: 'creditCard' as const,
    title: hasBenefit.value && !isWalletBenefit.value && rewardLabel.value
      ? t('dashboard.onboarding.stepRewardTitle', { amount: rewardLabel.value })
      : t('dashboard.onboarding.stepUsageTitle'),
    description: hasBenefit.value && !isWalletBenefit.value && rewardLabel.value
      ? t('dashboard.onboarding.stepRewardDescription')
      : t('dashboard.onboarding.stepUsageDescription')
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
