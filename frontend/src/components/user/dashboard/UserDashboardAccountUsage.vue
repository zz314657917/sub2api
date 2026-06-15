<template>
  <section class="card p-5">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('dashboard.personalAccountUsage.title') }}</h2>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ displayStartDate }} - {{ displayEndDate }}</p>
      </div>
      <router-link
        to="/my-accounts"
        class="inline-flex items-center gap-1.5 text-xs font-medium text-gray-500 transition-colors hover:text-primary-600 dark:text-gray-400 dark:hover:text-primary-300"
      >
        <Icon name="shield" size="sm" />
        {{ t('dashboard.personalAccountUsage.settlementLedger') }}
      </router-link>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <LoadingSpinner size="md" />
    </div>
    <template v-else>
      <div class="mt-5 grid gap-5 md:grid-cols-5">
        <div class="min-w-0">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.myAccounts') }}</p>
          <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ formatInteger(safeSummary.total_accounts) }}</p>
          <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
            {{ t('dashboard.personalAccountUsage.publicApprovedInline', { count: safeSummary.public_active_accounts }) }}
          </p>
        </div>

        <div class="min-w-0">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.ownUsage') }}</p>
          <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ formatCost(safeSummary.own_usage_cost) }}</p>
          <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.requests', { count: formatInteger(safeSummary.own_usage_requests) }) }}</p>
        </div>

        <div class="min-w-0">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.sharedUsage') }}</p>
          <p class="mt-1 text-2xl font-bold text-blue-600 dark:text-blue-400">{{ formatCost(safeSummary.shared_usage_cost) }}</p>
          <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.requests', { count: formatInteger(safeSummary.shared_usage_requests) }) }}</p>
        </div>

        <div class="min-w-0">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.shareIncome') }}</p>
          <p class="mt-1 text-2xl font-bold text-emerald-600 dark:text-emerald-400">{{ formatCost(safeSummary.share_income) }}</p>
          <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.platformAmount', { amount: formatCost(safeSummary.platform_amount) }) }}</p>
        </div>

        <div class="min-w-0">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.balanceNetChange') }}</p>
          <p class="mt-1 text-2xl font-bold" :class="netChangeClass">{{ formatSignedCost(safeSummary.balance_net_change) }}</p>
          <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.balanceDeduction', { amount: formatCost(safeSummary.balance_deduction) }) }}</p>
        </div>
      </div>

      <div class="mt-4 grid grid-cols-2 gap-x-5 gap-y-3 border-t border-gray-100 pt-3 text-xs dark:border-dark-700 sm:grid-cols-4">
        <div class="flex items-center justify-between gap-3">
          <span class="text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.privateAccounts') }}</span>
          <span class="font-semibold text-gray-900 dark:text-white">{{ formatInteger(safeSummary.private_accounts) }}</span>
        </div>
        <div class="flex items-center justify-between gap-3">
          <span class="text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.publicPending') }}</span>
          <span class="font-semibold text-amber-600 dark:text-amber-400">{{ formatInteger(safeSummary.public_pending_accounts) }}</span>
        </div>
        <div class="flex items-center justify-between gap-3">
          <span class="text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.publicApproved') }}</span>
          <span class="font-semibold text-emerald-600 dark:text-emerald-400">{{ formatInteger(safeSummary.public_active_accounts) }}</span>
        </div>
        <div class="flex items-center justify-between gap-3">
          <span class="text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.publicSuspended') }}</span>
          <span class="font-semibold text-rose-600 dark:text-rose-400">{{ formatInteger(safeSummary.public_suspended_accounts) }}</span>
        </div>
      </div>

      <div class="mt-4 border-t border-gray-100 pt-3 dark:border-dark-700">
        <div class="flex items-center justify-between gap-3 text-xs">
          <span class="font-medium text-gray-600 dark:text-gray-300">{{ t('dashboard.personalAccountUsage.accountDetails') }}</span>
          <span class="text-gray-500 dark:text-gray-400">{{ t('dashboard.personalAccountUsage.accountCost', { amount: formatCost(safeSummary.account_cost) }) }}</span>
        </div>
        <p v-if="totalRequests === 0" class="py-7 text-center text-sm font-medium text-gray-500 dark:text-gray-400">
          {{ t('dashboard.personalAccountUsage.noUsage') }}
        </p>
        <p v-else class="py-5 text-center text-sm text-gray-500 dark:text-gray-400">
          {{ t('dashboard.personalAccountUsage.usageSummary', { count: formatInteger(totalRequests) }) }}
        </p>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatCreditAmount } from '@/utils/credits'
import type { UserAccountUsageSummary } from '@/types'

const props = defineProps<{
  summary: UserAccountUsageSummary | null
  loading: boolean
  startDate: string
  endDate: string
}>()

const { t } = useI18n()

const emptySummary: UserAccountUsageSummary = {
  owner_user_id: 0,
  start_date: '',
  end_date: '',
  total_accounts: 0,
  private_accounts: 0,
  public_pending_accounts: 0,
  public_active_accounts: 0,
  public_suspended_accounts: 0,
  own_usage_cost: 0,
  own_usage_requests: 0,
  shared_usage_cost: 0,
  shared_usage_requests: 0,
  share_income: 0,
  platform_amount: 0,
  account_cost: 0,
  balance_deduction: 0,
  balance_net_change: 0
}

const safeSummary = computed(() => props.summary ?? emptySummary)
const displayStartDate = computed(() => safeSummary.value.start_date || props.startDate)
const displayEndDate = computed(() => safeSummary.value.end_date || props.endDate)
const totalRequests = computed(() => safeSummary.value.own_usage_requests + safeSummary.value.shared_usage_requests)
const netChangeClass = computed(() => safeSummary.value.balance_net_change >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400')

const formatCost = (value: number) => formatCreditAmount(value, { minimumFractionDigits: 4, maximumFractionDigits: 4 })
const formatInteger = (value: number) => Math.trunc(value || 0).toLocaleString()
const formatSignedCost = (value: number) => {
  const amount = Number(value || 0)
  return formatCreditAmount(amount, {
    minimumFractionDigits: 4,
    maximumFractionDigits: 4,
    signDisplay: 'always',
  })
}
</script>
