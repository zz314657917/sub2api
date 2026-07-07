<template>
  <Transition name="api-key-onboarding">
    <section
      v-if="show"
      class="pointer-events-none fixed inset-0 z-40 flex items-center justify-center bg-[#2f2925]/35 px-4 py-6 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-live="polite"
      :aria-label="dialogTitle"
    >
      <div class="pointer-events-auto w-full max-w-[720px] overflow-hidden rounded-xl border border-[#d8cec2] bg-[#fffaf5] text-gray-900 shadow-2xl shadow-[#a9583e]/20 dark:border-[#5f7f68]/45 dark:bg-[#2f2925] dark:text-[#f5f0e8]">
        <div class="relative min-h-[292px] overflow-hidden">
          <img
            src="/onboarding/new-user-trial-popup-header.png"
            alt=""
            aria-hidden="true"
            class="absolute inset-0 h-full w-full object-cover"
          >
          <div class="absolute inset-0 bg-gradient-to-r from-[#2f2925] via-[#2f2925]/92 to-[#2f2925]/20"></div>
          <button
            type="button"
            class="absolute right-4 top-4 z-10 rounded-lg p-2 text-[#f5f0e8]/80 transition-colors hover:bg-[#fffaf5]/15 hover:text-white"
            aria-label="Close"
            @click="emit('skip')"
          >
            <Icon name="x" size="sm" />
          </button>

          <div class="relative flex min-h-[292px] max-w-[28rem] flex-col justify-between px-5 py-5 sm:px-6 sm:py-6">
            <div>
              <div class="inline-flex items-center gap-2 rounded-full border border-[#cc785c]/50 bg-[#cc785c]/20 px-3 py-1 text-xs font-semibold text-[#fffaf5]">
                <Icon name="sparkles" size="xs" :stroke-width="2" />
                {{ badgeText }}
              </div>
              <h3 class="mt-4 text-2xl font-bold leading-tight text-white sm:text-3xl">
                {{ dialogTitle }}
              </h3>
              <p class="mt-3 text-sm leading-6 text-[#f5f0e8] sm:text-base">
                {{ dialogDescription }}
              </p>
              <p v-if="showBenefitNotice" class="mt-3 text-xs leading-5 text-[#f5f0e8]/80">
                {{ benefitNotice }}
              </p>
            </div>

            <div class="mt-5 flex flex-wrap gap-2">
              <span
                v-for="pill in pills"
                :key="pill"
                class="rounded-full border border-[#fffaf5]/20 bg-[#fffaf5]/12 px-3 py-1 text-xs font-medium text-[#fffaf5]"
              >
                {{ pill }}
              </span>
            </div>
          </div>
        </div>

        <div class="space-y-4 border-t border-[#d8cec2] bg-[#faf9f5] px-5 py-4 text-gray-900 dark:border-[#5f7f68]/40 dark:bg-[#2f2925] dark:text-[#f5f0e8]">
          <div class="grid gap-3 sm:grid-cols-3">
            <div
              v-for="item in steps"
              :key="item.title"
              class="rounded-lg border border-[#d8cec2] bg-[#fffaf5] p-3 dark:border-[#5f7f68]/40 dark:bg-[#2f2925]/80"
            >
              <div class="flex items-center gap-2 text-sm font-semibold text-gray-950 dark:text-[#f5f0e8]">
                <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-[#f5f0e8] text-[#a9583e] dark:bg-[#a9583e]/20 dark:text-[#cc785c]">
                  <Icon :name="item.icon" size="xs" :stroke-width="2" />
                </span>
                {{ item.title }}
              </div>
              <p class="mt-2 text-xs leading-5 text-gray-600 dark:text-[#f5f0e8]/75">{{ item.description }}</p>
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
              {{ primaryActionText }}
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
  hasApiKey?: boolean
  hasBenefit?: boolean
  benefitLabel?: string
  benefitRewardLabel?: string
  benefitKind?: 'reward' | 'trial' | 'wallet'
}>()

const emit = defineEmits<{
  (e: 'create'): void
  (e: 'tutorial'): void
  (e: 'skip'): void
}>()

const { t } = useI18n()

const hasBenefit = computed(() => Boolean(props.hasBenefit))
const hasApiKey = computed(() => Boolean(props.hasApiKey))
const isWalletBenefit = computed(() => props.benefitKind === 'wallet')
const isRewardBenefit = computed(() => props.benefitKind === 'reward')
const showBenefitNotice = computed(() => hasBenefit.value && !isRewardBenefit.value)
const quotaLabel = computed(() => props.benefitLabel || t(
  isRewardBenefit.value
    ? 'dashboard.onboarding.rewardFallback'
    : isWalletBenefit.value
      ? 'dashboard.onboarding.balanceFallback'
      : 'dashboard.onboarding.trialQuotaFallback'
))
const rewardLabel = computed(() => props.benefitRewardLabel || '')
const benefitNotice = computed(() => (
  isWalletBenefit.value
    ? t('dashboard.onboarding.walletBalanceNotice')
    : t('dashboard.onboarding.walletNotice')
))

const dialogTitle = computed(() => (
  hasBenefit.value
    ? t('dashboard.onboarding.trialTitle', { amount: quotaLabel.value })
    : t(hasApiKey.value ? 'dashboard.onboarding.readyTitle' : 'dashboard.onboarding.title')
))
const dialogDescription = computed(() => (
  hasBenefit.value
    ? t(
      isRewardBenefit.value
        ? hasApiKey.value
          ? 'dashboard.onboarding.rewardDescriptionWithKey'
          : 'dashboard.onboarding.rewardDescription'
        : isWalletBenefit.value
          ? hasApiKey.value
            ? 'dashboard.onboarding.balanceDescriptionWithKey'
            : 'dashboard.onboarding.balanceDescription'
          : hasApiKey.value
            ? 'dashboard.onboarding.trialDescriptionWithKey'
            : 'dashboard.onboarding.trialDescription'
    )
    : t(hasApiKey.value ? 'dashboard.onboarding.descriptionWithKey' : 'dashboard.onboarding.description')
))
const badgeText = computed(() => (
  hasBenefit.value
    ? t('dashboard.onboarding.trialBadge')
    : t('dashboard.onboarding.badge')
))

const pills = computed(() => (
  hasBenefit.value
    ? isRewardBenefit.value
      ? [
        t('dashboard.onboarding.pillGpt55'),
        t('dashboard.onboarding.pillImage2'),
        t('dashboard.onboarding.pillClaude'),
        t('dashboard.onboarding.pillOpenClaw')
      ]
      : [
        quotaLabel.value,
        t(isWalletBenefit.value ? 'dashboard.onboarding.pillBalanceDeduct' : 'dashboard.onboarding.pillAutoActivate'),
        t('dashboard.onboarding.pillNoRecharge')
      ]
    : [
      t(hasApiKey.value ? 'dashboard.onboarding.stepCopyKeyTitle' : 'dashboard.onboarding.stepKeyTitle'),
      t('dashboard.onboarding.stepToolTitle'),
      t('dashboard.onboarding.stepUsageTitle')
    ]
))

const primaryActionText = computed(() => (
  hasApiKey.value
    ? t('dashboard.onboarding.openKey')
    : t('dashboard.onboarding.createKey')
))

const steps = computed(() => [
  {
    icon: 'key' as const,
    title: t(hasApiKey.value ? 'dashboard.onboarding.stepCopyKeyTitle' : 'dashboard.onboarding.stepKeyTitle'),
    description: t(hasApiKey.value ? 'dashboard.onboarding.stepCopyKeyDescription' : 'dashboard.onboarding.stepKeyDescription')
  },
  {
    icon: 'bolt' as const,
    title: hasBenefit.value
      ? t('dashboard.onboarding.stepTrialTitle')
      : t('dashboard.onboarding.stepToolTitle'),
    description: hasBenefit.value
      ? t(
        isRewardBenefit.value
          ? 'dashboard.onboarding.stepRewardCallDescription'
          : isWalletBenefit.value
            ? 'dashboard.onboarding.stepBalanceDescription'
            : 'dashboard.onboarding.stepTrialDescription'
      )
      : t('dashboard.onboarding.stepToolDescription')
  },
  {
    icon: 'creditCard' as const,
    title: hasBenefit.value && isRewardBenefit.value
      ? t('dashboard.onboarding.stepRewardClaimTitle')
      : hasBenefit.value && !isWalletBenefit.value && rewardLabel.value
        ? t('dashboard.onboarding.stepRewardTitle', { amount: rewardLabel.value })
      : t('dashboard.onboarding.stepUsageTitle'),
    description: hasBenefit.value && isRewardBenefit.value
      ? t('dashboard.onboarding.stepRewardDescription')
      : hasBenefit.value && !isWalletBenefit.value && rewardLabel.value
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
