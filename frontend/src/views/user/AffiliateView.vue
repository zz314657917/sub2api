<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-8">
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <template v-else-if="detail">
        <section class="space-y-6">
          <div class="border-b border-gray-200 pb-6 dark:border-dark-700">
            <p class="inline-flex items-center gap-1.5 text-sm font-medium text-primary-600 dark:text-primary-400">
              <Icon name="gift" size="sm" />
              {{ t('affiliate.hero.eyebrow') }}
            </p>
            <div class="mt-3 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div class="max-w-3xl">
                <h1 class="text-2xl font-semibold tracking-normal text-gray-950 dark:text-white md:text-3xl">
                  {{ t('affiliate.hero.title') }}
                </h1>
                <p class="mt-3 text-sm leading-6 text-gray-600 dark:text-dark-300">
                  {{ affiliateDescription }}
                </p>
              </div>
              <div class="rounded-lg border border-primary-200 bg-primary-50 px-4 py-3 dark:border-primary-900/40 dark:bg-primary-900/20">
                <p class="text-xs text-primary-700 dark:text-primary-300">{{ t('affiliate.stats.rebateRate') }}</p>
                <p class="mt-1 text-2xl font-semibold text-primary-700 dark:text-primary-300">
                  {{ formattedRebateRate }}<span class="ml-0.5 text-base font-medium">%</span>
                </p>
              </div>
            </div>
          </div>

          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.summary.earned') }}</p>
              <p class="mt-2 text-2xl font-semibold text-gray-950 dark:text-white">
                {{ formatCurrency(detail.aff_history_quota) }}
              </p>
              <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">{{ t('affiliate.summary.earnedHint') }}</p>
            </div>
            <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.availableQuota') }}</p>
              <p class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">
                {{ formatCurrency(detail.aff_quota) }}
              </p>
              <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">{{ t('affiliate.summary.availableHint') }}</p>
            </div>
            <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.invitedUsers') }}</p>
              <p class="mt-2 text-2xl font-semibold text-gray-950 dark:text-white">
                {{ formatCount(detail.aff_count) }}
              </p>
              <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">{{ t('affiliate.summary.invitedHint') }}</p>
            </div>
            <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.summary.status') }}</p>
              <p class="mt-2 text-2xl font-semibold text-gray-950 dark:text-white">{{ t('affiliate.summary.ready') }}</p>
              <p v-if="detail.aff_frozen_quota > 0" class="mt-1 text-xs text-amber-600 dark:text-amber-400">
                {{ t('affiliate.stats.frozenQuota') }}: {{ formatCurrency(detail.aff_frozen_quota) }}
              </p>
              <p v-else class="mt-1 text-xs text-gray-400 dark:text-dark-500">{{ t('affiliate.summary.statusHint') }}</p>
            </div>
          </div>
        </section>

        <section class="space-y-3">
          <div>
            <h2 class="text-xl font-semibold text-gray-950 dark:text-white">{{ t('affiliate.share.title') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.share.description') }}</p>
          </div>

          <div class="rounded-lg border border-primary-200 bg-primary-50/60 p-4 dark:border-primary-900/40 dark:bg-primary-900/20">
            <div class="flex flex-col gap-3 lg:flex-row lg:items-center">
              <div class="flex min-w-0 flex-1 items-center gap-3 rounded-md border border-primary-200 bg-white px-4 py-3 dark:border-primary-900/50 dark:bg-dark-900">
                <Icon name="link" size="sm" class="shrink-0 text-primary-600 dark:text-primary-400" />
                <code class="min-w-0 flex-1 break-all text-sm font-semibold text-gray-950 dark:text-white">{{ inviteLink }}</code>
              </div>
              <button class="btn btn-primary shrink-0 justify-center" @click="copyInviteLink">
                <Icon name="copy" size="sm" />
                <span>{{ t('affiliate.copyLink') }}</span>
              </button>
            </div>

            <div class="mt-3 flex flex-wrap items-center gap-2 text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{ t('affiliate.share.more') }}</span>
              <a class="btn btn-secondary btn-sm" :href="mailShareHref">
                <Icon name="mail" size="sm" />
                <span>{{ t('affiliate.share.email') }}</span>
              </a>
              <button class="btn btn-secondary btn-sm" @click="copyCode">
                <Icon name="ticket" size="sm" />
                <span>{{ t('affiliate.copyCode') }}</span>
              </button>
              <span class="rounded-md bg-white px-2.5 py-1 font-mono text-xs font-semibold text-gray-700 ring-1 ring-primary-200 dark:bg-dark-900 dark:text-dark-200 dark:ring-primary-900/50">
                {{ detail.aff_code }}
              </span>
            </div>
          </div>
        </section>

        <section class="space-y-4">
          <h2 class="text-xl font-semibold text-gray-950 dark:text-white">{{ t('affiliate.steps.title') }}</h2>
          <div class="grid gap-4 md:grid-cols-3">
            <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
              <p class="font-mono text-xs font-semibold text-primary-600 dark:text-primary-400">01</p>
              <h3 class="mt-4 text-base font-semibold text-gray-950 dark:text-white">{{ t('affiliate.steps.share.title') }}</h3>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t('affiliate.steps.share.description') }}</p>
            </div>
            <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
              <p class="font-mono text-xs font-semibold text-primary-600 dark:text-primary-400">02</p>
              <h3 class="mt-4 text-base font-semibold text-gray-950 dark:text-white">{{ t('affiliate.steps.verify.title') }}</h3>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t('affiliate.steps.verify.description') }}</p>
            </div>
            <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
              <p class="font-mono text-xs font-semibold text-primary-600 dark:text-primary-400">03</p>
              <h3 class="mt-4 text-base font-semibold text-gray-950 dark:text-white">{{ t('affiliate.steps.earn.title', { amount: formattedApiCallRewardAmount }) }}</h3>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t('affiliate.steps.earn.description', { rate: `${formattedRebateRate}%` }) }}</p>
            </div>
          </div>

          <div class="rounded-lg border border-emerald-200 bg-emerald-50/60 p-4 dark:border-emerald-900/40 dark:bg-emerald-900/10">
            <div class="flex gap-3">
              <Icon name="infoCircle" size="sm" class="mt-0.5 shrink-0 text-emerald-600 dark:text-emerald-400" />
              <div>
                <p class="text-sm font-semibold text-emerald-900 dark:text-emerald-200">{{ t('affiliate.notice.title') }}</p>
                <ul class="mt-2 space-y-1 text-sm leading-6 text-emerald-800 dark:text-emerald-300">
                  <li>{{ t('affiliate.notice.line1') }}</li>
                  <li>{{ t('affiliate.notice.line2') }}</li>
                  <li>{{ t('affiliate.notice.line3') }}</li>
                  <li>{{ t('affiliate.notice.line4') }}</li>
                </ul>
              </div>
            </div>
          </div>
        </section>

        <section class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 class="text-base font-semibold text-gray-950 dark:text-white">{{ t('affiliate.transfer.title') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.transfer.description') }}</p>
            </div>
            <button
              class="btn btn-primary"
              :disabled="transferring || detail.aff_quota <= 0"
              @click="transferQuota"
            >
              <Icon v-if="transferring" name="refresh" size="sm" class="animate-spin" />
              <Icon v-else name="dollar" size="sm" />
              <span>{{ transferring ? t('affiliate.transfer.transferring') : t('affiliate.transfer.button') }}</span>
            </button>
          </div>
          <p v-if="detail.aff_quota <= 0" class="mt-3 text-sm text-amber-600 dark:text-amber-400">
            {{ t('affiliate.transfer.empty') }}
          </p>
        </section>

        <section class="space-y-4">
          <div>
            <h2 class="text-xl font-semibold text-gray-950 dark:text-white">{{ t('affiliate.recent.title') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.recent.description') }}</p>
          </div>
          <div v-if="detail.invitees.length === 0" class="rounded-lg border border-dashed border-gray-300 p-8 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ t('affiliate.recent.empty') }}
          </div>
          <div v-else class="overflow-x-auto rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
            <table class="w-full min-w-[760px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 bg-gray-50 text-gray-500 dark:border-dark-700 dark:bg-dark-900/60 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.email') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.username') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.apiStatus') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ t('affiliate.invitees.columns.rebate') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.joinedAt') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ t('affiliate.invitees.columns.action') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in detail.invitees"
                  :key="item.user_id"
                  class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                >
                  <td class="px-3 py-3 text-gray-900 dark:text-white">{{ item.email || '-' }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ item.username || '-' }}</td>
                  <td class="px-3 py-3">
                    <span
                      class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="item.api_used
                        ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
                        : 'bg-gray-100 text-gray-500 dark:bg-dark-800 dark:text-dark-400'"
                    >
                      {{ item.api_used ? t('affiliate.invitees.apiStatus.used') : t('affiliate.invitees.apiStatus.pending') }}
                    </span>
                    <p v-if="item.api_used_at" class="mt-1 text-xs text-gray-400">
                      {{ formatDateTime(item.api_used_at) }}
                    </p>
                  </td>
                  <td class="px-3 py-3 text-right font-medium text-emerald-600 dark:text-emerald-400">{{ formatCurrency(item.total_rebate) }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                  <td class="px-3 py-3 text-right">
                    <button
                      type="button"
                      class="btn btn-secondary btn-sm whitespace-nowrap"
                      :disabled="!canClaimApiCallReward(item) || claimingInviteeId === item.user_id"
                      @click="claimApiCallReward(item)"
                    >
                      <Icon
                        v-if="claimingInviteeId === item.user_id"
                        name="refresh"
                        size="sm"
                        class="animate-spin"
                      />
                      <Icon v-else name="dollar" size="sm" />
                      <span>{{ claimButtonLabel(item) }}</span>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section class="space-y-4">
          <h2 class="text-xl font-semibold text-gray-950 dark:text-white">{{ t('affiliate.faq.title') }}</h2>
          <div class="divide-y divide-gray-200 overflow-hidden rounded-lg border border-gray-200 bg-white dark:divide-dark-700 dark:border-dark-700 dark:bg-dark-800">
            <div class="p-4">
              <p class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('affiliate.faq.limit.question') }}</p>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t('affiliate.faq.limit.answer', { amount: formattedApiCallRewardAmount }) }}</p>
            </div>
            <div class="p-4">
              <p class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('affiliate.faq.when.question') }}</p>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t('affiliate.faq.when.answer') }}</p>
            </div>
            <div class="p-4">
              <p class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('affiliate.faq.expire.question') }}</p>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t('affiliate.faq.expire.answer') }}</p>
            </div>
            <div class="p-4">
              <p class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('affiliate.faq.selfInvite.question') }}</p>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300">{{ t('affiliate.faq.selfInvite.answer') }}</p>
            </div>
          </div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import userAPI from '@/api/user'
import type { AffiliateInvitee, UserAffiliateDetail } from '@/types'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const transferring = ref(false)
const claimingInviteeId = ref<number | null>(null)
const detail = ref<UserAffiliateDetail | null>(null)

const inviteLink = computed(() => {
  if (!detail.value) return ''
  if (typeof window === 'undefined') return `/register?aff=${encodeURIComponent(detail.value.aff_code)}`
  return `${window.location.origin}/register?aff=${encodeURIComponent(detail.value.aff_code)}`
})

// Rebate rate is a percentage in the range [0, 100]; backend already clamps it.
// We trim trailing zeros (e.g. 20.00 → "20", 12.50 → "12.5") for a cleaner UI.
const formattedRebateRate = computed(() => {
  const v = detail.value?.effective_rebate_rate_percent ?? 0
  const rounded = Math.round(v * 100) / 100
  return Number.isInteger(rounded) ? String(rounded) : rounded.toString()
})

const formattedApiCallRewardAmount = computed(() => {
  const amount = detail.value?.api_call_reward_amount ?? 0
  return formatCurrency(amount)
})

const affiliateDescription = computed(() => {
  return t('affiliate.descriptionWithReward', { amount: formattedApiCallRewardAmount.value })
})

const mailShareHref = computed(() => {
  const subject = t('affiliate.share.emailSubject')
  const body = t('affiliate.share.emailBody', { link: inviteLink.value })
  return `mailto:?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`
})

function formatCount(value: number): string {
  return value.toLocaleString()
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) {
    loading.value = true
  }
  try {
    detail.value = await userAPI.getAffiliateDetail()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.loadFailed')))
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

async function copyCode(): Promise<void> {
  if (!detail.value?.aff_code) return
  await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied'))
}

async function copyInviteLink(): Promise<void> {
  if (!inviteLink.value) return
  await copyToClipboard(inviteLink.value, t('affiliate.linkCopied'))
}

async function transferQuota(): Promise<void> {
  if (!detail.value || detail.value.aff_quota <= 0 || transferring.value) return
  transferring.value = true
  try {
    const resp = await userAPI.transferAffiliateQuota()
    appStore.showSuccess(t('affiliate.transfer.success', { amount: formatCurrency(resp.transferred_quota) }))
    await Promise.all([
      loadAffiliateDetail(true),
      authStore.refreshUser().catch(() => undefined),
    ])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.transferFailed')))
  } finally {
    transferring.value = false
  }
}

function canClaimApiCallReward(item: AffiliateInvitee): boolean {
  return Boolean(
    detail.value &&
    detail.value.api_call_reward_amount > 0 &&
    item.api_used &&
    !item.api_call_reward_claimed
  )
}

function claimButtonLabel(item: AffiliateInvitee): string {
  if (claimingInviteeId.value === item.user_id) {
    return t('affiliate.invitees.actions.claiming')
  }
  if (item.api_call_reward_claimed) {
    return t('affiliate.invitees.actions.claimed')
  }
  if ((detail.value?.api_call_reward_amount ?? 0) <= 0) {
    return t('affiliate.invitees.actions.notConfigured')
  }
  if (!item.api_used) {
    return t('affiliate.invitees.actions.waiting')
  }
  return t('affiliate.invitees.actions.claim')
}

async function claimApiCallReward(item: AffiliateInvitee): Promise<void> {
  if (!canClaimApiCallReward(item) || claimingInviteeId.value !== null) return
  claimingInviteeId.value = item.user_id
  try {
    const resp = await userAPI.claimAffiliateApiCallReward(item.user_id)
    appStore.showSuccess(t('affiliate.invitees.actions.claimSuccess', { amount: formatCurrency(resp.reward_amount) }))
    await Promise.all([
      loadAffiliateDetail(true),
      authStore.refreshUser().catch(() => undefined),
    ])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.invitees.actions.claimFailed')))
  } finally {
    claimingInviteeId.value = null
  }
}

onMounted(() => {
  void loadAffiliateDetail()
})
</script>
