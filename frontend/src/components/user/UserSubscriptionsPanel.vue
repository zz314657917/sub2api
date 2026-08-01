<template>
  <section
    v-if="loading || subscriptions.length > 0"
    class="space-y-4"
    data-testid="user-subscriptions-panel"
  >
    <div v-if="loading" class="card flex justify-center py-10">
      <div
        class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
      ></div>
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-2">
      <div
        v-for="subscription in subscriptions"
        :key="subscription.id"
        class="overflow-hidden rounded-2xl border bg-white dark:bg-dark-800"
        :class="platformBorderClass(subscription.group?.platform || '')"
      >
        <div
          class="flex flex-wrap items-start justify-between gap-3 border-b border-gray-100 p-4 dark:border-dark-700"
        >
          <div class="flex min-w-0 items-start gap-3">
            <div
              :class="[
                'mt-2 h-1.5 w-1.5 shrink-0 rounded-full',
                platformAccentDotClass(subscription.group?.platform || '')
              ]"
            />
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="font-semibold text-gray-900 dark:text-white">
                  {{ subscription.group?.name || `Group #${subscription.group_id}` }}
                </h3>
                <span
                  :class="[
                    'rounded-md border px-2 py-0.5 text-[11px] font-medium',
                    platformBadgeClass(subscription.group?.platform || '')
                  ]"
                >
                  {{ platformLabel(subscription.group?.platform || '') }}
                </span>
              </div>
              <p
                v-if="subscription.group?.description"
                class="mt-0.5 text-xs text-gray-500 dark:text-dark-400"
              >
                {{ subscription.group.description }}
              </p>
              <div
                class="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-gray-400 dark:text-gray-500"
              >
                <span>{{ t('payment.planCard.rate') }}: ×{{ subscription.group?.rate_multiplier ?? 1 }}</span>
                <span
                  v-if="subscriptionHasPeakRate(subscription)"
                  class="text-amber-700 dark:text-amber-300"
                >
                  {{ t('payment.planCard.peakRate') }}: {{ subscriptionPeakRateLabel(subscription) }}
                </span>
              </div>
            </div>
          </div>
          <div class="flex shrink-0 items-center gap-2">
            <span
              :class="[
                'rounded-full px-2 py-0.5 text-xs font-medium',
                subscription.status === 'active'
                  ? 'bg-[#7f9d8a]/15 text-[#5f7f68] dark:bg-[#7f9d8a]/15 dark:text-[#9ab3a0]'
                  : subscription.status === 'expired'
                    ? 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
                    : 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'
              ]"
            >
              {{ t(`userSubscriptions.status.${subscription.status}`) }}
            </span>
            <button
              v-if="subscription.status === 'active'"
              :class="[
                'rounded-lg px-3 py-1.5 text-xs font-semibold text-white transition-colors',
                platformButtonClass(subscription.group?.platform || '')
              ]"
              @click="renewSubscription(subscription)"
            >
              {{ t('payment.renewNow') }}
            </button>
          </div>
        </div>

        <div class="space-y-4 p-4">
          <div v-if="subscription.expires_at" class="flex items-center justify-between text-sm">
            <span class="text-gray-500 dark:text-dark-400">
              {{ t('userSubscriptions.expires') }}
            </span>
            <span :class="getExpirationClass(subscription.expires_at)">
              {{ formatExpirationDate(subscription.expires_at) }}
            </span>
          </div>
          <div v-else class="flex items-center justify-between text-sm">
            <span class="text-gray-500 dark:text-dark-400">
              {{ t('userSubscriptions.expires') }}
            </span>
            <span class="text-gray-700 dark:text-gray-300">
              {{ t('userSubscriptions.noExpiration') }}
            </span>
          </div>

          <div
            v-if="displaySubscriptionLimit(subscription.group?.daily_limit_usd) != null"
            class="space-y-2"
          >
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('userSubscriptions.daily') }}
              </span>
              <span class="text-sm text-gray-500 dark:text-dark-400">
                {{ formatCreditAmount(subscription.daily_usage_usd || 0) }} /
                {{ formatLimitAmount(subscription.group?.daily_limit_usd) }}
              </span>
            </div>
            <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
              <div
                class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                :class="getProgressBarClass(subscription.daily_usage_usd, subscription.group?.daily_limit_usd)"
                :style="{ width: getProgressWidth(subscription.daily_usage_usd, subscription.group?.daily_limit_usd) }"
              ></div>
            </div>
            <p
              v-if="subscription.daily_window_start"
              class="text-xs text-gray-500 dark:text-dark-400"
            >
              {{ formatDailyUsageWindow(subscription) }}
            </p>
          </div>

          <div
            v-if="displaySubscriptionLimit(subscription.group?.weekly_limit_usd) != null"
            class="space-y-2"
          >
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('userSubscriptions.weekly') }}
              </span>
              <span class="text-sm text-gray-500 dark:text-dark-400">
                {{ formatCreditAmount(subscription.weekly_usage_usd || 0) }} /
                {{ formatLimitAmount(subscription.group?.weekly_limit_usd) }}
              </span>
            </div>
            <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
              <div
                class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                :class="getProgressBarClass(subscription.weekly_usage_usd, subscription.group?.weekly_limit_usd)"
                :style="{ width: getProgressWidth(subscription.weekly_usage_usd, subscription.group?.weekly_limit_usd) }"
              ></div>
            </div>
            <p
              v-if="subscription.weekly_window_start"
              class="text-xs text-gray-500 dark:text-dark-400"
            >
              {{
                t('userSubscriptions.resetIn', {
                  time: formatResetTime(subscription.weekly_window_start, 168)
                })
              }}
            </p>
          </div>

          <div
            v-if="displaySubscriptionLimit(subscription.group?.monthly_limit_usd) != null"
            class="space-y-2"
          >
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('userSubscriptions.monthly') }}
              </span>
              <span class="text-sm text-gray-500 dark:text-dark-400">
                {{ formatCreditAmount(subscription.monthly_usage_usd || 0) }} /
                {{ formatLimitAmount(subscription.group?.monthly_limit_usd) }}
              </span>
            </div>
            <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
              <div
                class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                :class="getProgressBarClass(subscription.monthly_usage_usd, subscription.group?.monthly_limit_usd)"
                :style="{ width: getProgressWidth(subscription.monthly_usage_usd, subscription.group?.monthly_limit_usd) }"
              ></div>
            </div>
            <p
              v-if="subscription.monthly_window_start"
              class="text-xs text-gray-500 dark:text-dark-400"
            >
              {{
                t('userSubscriptions.resetIn', {
                  time: formatResetTime(subscription.monthly_window_start, 720)
                })
              }}
            </p>
          </div>

          <div
            v-if="!hasAnySubscriptionLimit(subscription.group)"
            class="flex items-center justify-center rounded-xl border border-[#d8cec2] bg-[#faf9f5] py-6 dark:border-dark-700 dark:bg-dark-900/30"
          >
            <div class="flex items-center gap-3">
              <span class="text-4xl text-[#5f7f68] dark:text-[#9ab3a0]">∞</span>
              <div>
                <p class="text-sm font-medium text-[#5f7f68] dark:text-[#9ab3a0]">
                  {{ t('userSubscriptions.unlimited') }}
                </p>
                <p class="text-xs text-[#5f7f68]/80 dark:text-[#9ab3a0]/80">
                  {{ t('userSubscriptions.unlimitedDesc') }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import subscriptionsAPI from '@/api/subscriptions'
import { useAppStore } from '@/stores/app'
import type { UserSubscription } from '@/types'
import { formatCreditAmount } from '@/utils/credits'
import { formatDateTimeToMinute } from '@/utils/format'
import { hasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import {
  platformBadgeClass,
  platformBorderClass,
  platformButtonClass,
  platformLabel,
} from '@/utils/platformColors'
import { displaySubscriptionLimit, hasAnySubscriptionLimit } from '@/utils/subscriptionLimits'
import {
  getRemainingDurationParts,
  isOneTimeDailyQuota,
  type RemainingDurationParts,
} from '@/utils/subscriptionQuota'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const subscriptions = ref<UserSubscription[]>([])
const loading = ref(true)

function platformAccentDotClass(p: string): string {
  switch (p) {
    case 'anthropic':
      return 'bg-orange-500'
    case 'openai':
      return 'bg-emerald-500'
    case 'antigravity':
      return 'bg-purple-500'
    case 'gemini':
      return 'bg-blue-500'
    default:
      return 'bg-gray-400'
  }
}

function subscriptionHasPeakRate(subscription: UserSubscription): boolean {
  return hasPeakRate(subscription.group)
}

function subscriptionPeakRateLabel(subscription: UserSubscription): string {
  return formatPeakRateWindow(
    subscription.group,
    serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset, t('common.serverTime'))
  )
}

async function loadSubscriptions() {
  try {
    loading.value = true
    subscriptions.value = await subscriptionsAPI.getMySubscriptions()
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function renewSubscription(subscription: UserSubscription) {
  router.push({
    path: '/purchase',
    query: { tab: 'subscription', group: String(subscription.group_id) },
  })
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  const normalizedLimit = displaySubscriptionLimit(limit)
  if (normalizedLimit == null) return '0%'
  const percentage = Math.min(((used || 0) / normalizedLimit) * 100, 100)
  return `${percentage}%`
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  const normalizedLimit = displaySubscriptionLimit(limit)
  if (normalizedLimit == null) return 'bg-gray-400'
  const percentage = ((used || 0) / normalizedLimit) * 100
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-orange-500'
  return 'bg-[#7f9d8a]'
}

function formatLimitAmount(limit: number | null | undefined): string {
  return formatCreditAmount(displaySubscriptionLimit(limit) ?? 0)
}

function formatExpirationDate(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days < 0) {
    return t('userSubscriptions.status.expired')
  }

  const dateStr = formatDateTimeToMinute(expires)

  if (days === 0) {
    return `${dateStr} (${t('common.today')})`
  }
  if (days === 1) {
    return `${dateStr} (${t('common.tomorrow')})`
  }

  return `${t('userSubscriptions.daysRemaining', { days })} (${dateStr})`
}

function getExpirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days <= 0) return 'text-red-600 dark:text-red-400 font-medium'
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-700 dark:text-gray-300'
}

function formatDurationParts(parts: RemainingDurationParts): string {
  if (parts.days > 0) {
    return `${parts.days}d ${parts.hours}h`
  }
  if (parts.hours > 0) {
    return `${parts.hours}h ${parts.minutes}m`
  }
  return `${parts.minutes}m`
}

function formatDailyUsageWindow(subscription: UserSubscription): string {
  if (isOneTimeDailyQuota(subscription) && subscription.expires_at) {
    const parts = getRemainingDurationParts(subscription.expires_at)
    if (!parts) return t('userSubscriptions.windowNotActive')
    return t('userSubscriptions.quotaEndsIn', { time: formatDurationParts(parts) })
  }

  return t('userSubscriptions.resetIn', {
    time: formatResetTime(subscription.daily_window_start, 24),
  })
}

function formatResetTime(windowStart: string | null, windowHours: number): string {
  if (!windowStart) return t('userSubscriptions.windowNotActive')

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const parts = getRemainingDurationParts(end)

  return parts ? formatDurationParts(parts) : t('userSubscriptions.windowNotActive')
}

onMounted(() => {
  loadSubscriptions()
})
</script>
