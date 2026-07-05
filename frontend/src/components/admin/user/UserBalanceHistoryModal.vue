<template>
  <BaseDialog :show="show" :title="t('admin.users.balanceHistoryTitle')" width="wide" :close-on-click-outside="true" :z-index="40" @close="$emit('close')">
    <div v-if="user" class="space-y-4">
      <!-- User header: two-row layout with full user info -->
      <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
        <!-- Row 1: avatar + email/username/created_at (left) + current balance (right) -->
        <div class="flex items-center gap-3">
          <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-900/30">
            <span class="text-lg font-medium text-primary-700 dark:text-primary-300">
              {{ user.email.charAt(0).toUpperCase() }}
            </span>
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <p class="truncate font-medium text-gray-900 dark:text-white">{{ user.email }}</p>
              <span
                v-if="user.username"
                class="flex-shrink-0 rounded bg-primary-50 px-1.5 py-0.5 text-xs text-primary-600 dark:bg-primary-900/20 dark:text-primary-400"
              >
                {{ user.username }}
              </span>
            </div>
            <p class="text-xs text-gray-400 dark:text-dark-500">
              {{ t('admin.users.createdAt') }}: {{ formatDateTime(user.created_at) }}
            </p>
          </div>
          <!-- Current balance: prominent display on the right -->
          <div class="flex-shrink-0 text-right">
            <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.users.currentBalance') }}</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              {{ formatCreditAmount(user.balance || 0) }}
            </p>
          </div>
        </div>
        <!-- Row 2: notes + total recharged -->
        <div class="mt-2.5 flex items-center justify-between border-t border-gray-200/60 pt-2.5 dark:border-dark-600/60">
          <p class="min-w-0 flex-1 truncate text-xs text-gray-500 dark:text-dark-400" :title="user.notes || ''">
            <template v-if="user.notes">{{ t('admin.users.notes') }}: {{ user.notes }}</template>
            <template v-else>&nbsp;</template>
          </p>
          <p class="ml-4 flex-shrink-0 text-xs text-gray-500 dark:text-dark-400">
            {{ t('admin.users.totalRecharged') }}: <span class="font-semibold text-emerald-600 dark:text-emerald-400">{{ formatCreditAmount(totalRecharged) }}</span>
          </p>
        </div>
      </div>

      <!-- Recent usage -->
      <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
        <div class="mb-3 flex items-center justify-between gap-3">
          <div class="min-w-0">
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('admin.users.recentUsageRecords') }}
            </h4>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
              {{ t('admin.users.recentUsageRecordsHint') }}
            </p>
          </div>
          <button
            type="button"
            :disabled="usageLoading"
            class="inline-flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 disabled:cursor-not-allowed disabled:opacity-60 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white"
            :title="t('common.refresh')"
            @click="loadRecentUsage"
          >
            <Icon name="refresh" size="sm" :class="usageLoading ? 'animate-spin' : ''" />
          </button>
        </div>

        <div v-if="usageLoading" class="flex items-center justify-center py-5">
          <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
        </div>

        <div v-else-if="usageError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-300">
          {{ usageError }}
        </div>

        <div v-else-if="recentUsage.length === 0" class="py-4 text-center">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.users.noRecentUsageRecords') }}</p>
        </div>

        <div v-else class="space-y-2">
          <div
            v-for="log in recentUsage"
            :key="log.id"
            class="rounded-lg border border-gray-100 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-700/60"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="flex min-w-0 items-start gap-3">
                <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-sky-100 text-sky-600 dark:bg-sky-900/30 dark:text-sky-300">
                  <Icon name="beaker" size="sm" />
                </div>
                <div class="min-w-0">
                  <div class="flex min-w-0 flex-wrap items-center gap-2">
                    <p class="max-w-[20rem] truncate text-sm font-medium text-gray-900 dark:text-white" :title="displayModelLabel(log.model, log.model || t('usage.unknown'))">
                      {{ displayModelLabel(log.model, log.model || t('usage.unknown')) }}
                    </p>
                    <span :class="['rounded px-1.5 py-0.5 text-[11px] font-medium', getUsageTypeBadgeClass(log)]">
                      {{ getUsageTypeLabel(log) }}
                    </span>
                  </div>
                  <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-500 dark:text-dark-400">
                    <span>{{ formatDateTime(log.created_at) }}</span>
                    <span>{{ t('usage.tokens') }}: {{ formatTokens(getUsageTotalTokens(log)) }}</span>
                    <span>{{ t('usage.duration') }}: {{ formatDuration(log.duration_ms) }}</span>
                    <span v-if="log.api_key?.name" class="max-w-[12rem] truncate" :title="log.api_key.name">
                      {{ t('usage.apiKeyFilter') }}: {{ log.api_key.name }}
                    </span>
                  </div>
                </div>
              </div>
              <div class="flex-shrink-0 text-right">
                <p class="text-sm font-semibold text-emerald-600 dark:text-emerald-400">
                  {{ formatUsageCost(log.actual_cost) }}
                </p>
                <p v-if="log.total_cost !== log.actual_cost" class="text-xs text-gray-400 line-through dark:text-dark-500">
                  {{ formatUsageCost(log.total_cost) }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Subscription payment orders -->
      <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
        <div class="mb-3 flex items-center justify-between gap-3">
          <div class="min-w-0">
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('admin.users.subscriptionPaymentOrders') }}
            </h4>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
              {{ t('admin.users.subscriptionPaymentOrdersHint') }}
            </p>
          </div>
          <button
            type="button"
            :disabled="subscriptionOrdersLoading"
            class="inline-flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 disabled:cursor-not-allowed disabled:opacity-60 dark:text-dark-300 dark:hover:bg-dark-700 dark:hover:text-white"
            :title="t('common.refresh')"
            @click="loadSubscriptionOrders"
          >
            <Icon name="refresh" size="sm" :class="subscriptionOrdersLoading ? 'animate-spin' : ''" />
          </button>
        </div>

        <div v-if="subscriptionOrdersLoading" class="flex items-center justify-center py-5">
          <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
        </div>

        <div v-else-if="subscriptionOrdersError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-300">
          {{ subscriptionOrdersError }}
        </div>

        <div v-else-if="subscriptionOrders.length === 0" class="py-4 text-center">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.users.noSubscriptionPaymentOrders') }}</p>
        </div>

        <div v-else class="space-y-2">
          <div
            v-for="order in subscriptionOrders"
            :key="order.id"
            class="rounded-lg border border-gray-100 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-700/60"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="flex min-w-0 items-start gap-3">
                <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-300">
                  <Icon name="badge" size="sm" />
                </div>
                <div class="min-w-0">
                  <div class="flex min-w-0 flex-wrap items-center gap-2">
                    <p class="font-mono text-sm font-medium text-gray-900 dark:text-white">
                      #{{ order.id }}
                    </p>
                    <OrderStatusBadge :status="order.status" />
                    <span
                      v-if="order.subscription_days"
                      class="rounded bg-purple-50 px-1.5 py-0.5 text-[11px] font-medium text-purple-700 dark:bg-purple-900/30 dark:text-purple-200"
                    >
                      {{ t('admin.users.subscriptionOrderDays', { days: order.subscription_days }) }}
                    </span>
                  </div>
                  <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-500 dark:text-dark-400">
                    <span>{{ t('payment.orders.orderNo') }}: {{ order.out_trade_no }}</span>
                    <span>{{ t('payment.orders.paymentMethod') }}: {{ t('payment.methods.' + order.payment_type, order.payment_type) }}</span>
                    <span>{{ t('payment.orders.createdAt') }}: {{ formatDateTime(order.created_at) }}</span>
                    <span v-if="order.paid_at">{{ t('admin.users.subscriptionOrderPaidAt') }}: {{ formatDateTime(order.paid_at) }}</span>
                    <span v-if="order.completed_at">{{ t('admin.users.subscriptionOrderCompletedAt') }}: {{ formatDateTime(order.completed_at) }}</span>
                    <span v-if="order.plan_id">{{ t('admin.users.subscriptionOrderPlan') }}: #{{ order.plan_id }}</span>
                  </div>
                </div>
              </div>
              <div class="flex-shrink-0 text-right">
                <p class="text-sm font-semibold text-gray-900 dark:text-white">
                  {{ formatSubscriptionOrderAmount(order) }}
                </p>
                <p v-if="order.refund_amount" class="text-xs text-red-500 dark:text-red-300">
                  {{ t('payment.admin.refundAmount') }}: {{ formatSubscriptionOrderRefund(order) }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Type filter + Action buttons -->
      <div class="flex items-center gap-3">
        <Select
          v-model="typeFilter"
          :options="typeOptions"
          class="w-56"
          @change="loadHistory(1)"
        />
        <!-- Deposit button - matches menu style -->
        <button
          v-if="!hideActions"
          @click="emit('deposit')"
          class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-700 transition-colors hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700"
        >
          <Icon name="plus" size="sm" class="text-emerald-500" :stroke-width="2" />
          {{ t('admin.users.deposit') }}
        </button>
        <!-- Withdraw button - matches menu style -->
        <button
          v-if="!hideActions"
          @click="emit('withdraw')"
          class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-700 transition-colors hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700"
        >
          <svg class="h-4 w-4 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
          </svg>
          {{ t('admin.users.withdraw') }}
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex justify-center py-8">
        <svg class="h-8 w-8 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
      </div>

      <!-- Empty state -->
      <div v-else-if="history.length === 0" class="py-8 text-center">
        <p class="text-sm text-gray-500">{{ t('admin.users.noBalanceHistory') }}</p>
      </div>

      <!-- History list -->
      <div v-else class="max-h-[28rem] space-y-3 overflow-y-auto">
        <div
          v-for="item in history"
          :key="item.id"
          class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800"
        >
          <div class="flex items-start justify-between">
            <!-- Left: type icon + description -->
            <div class="flex items-start gap-3">
              <div
                :class="[
                  'flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg',
                  getIconBg(item)
                ]"
              >
                <Icon :name="getIconName(item)" size="sm" :class="getIconColor(item)" />
              </div>
              <div>
                <p class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ getItemTitle(item) }}
                </p>
                <!-- Notes (admin adjustment reason) -->
                <p
                  v-if="item.notes"
                  class="mt-0.5 text-xs text-gray-500 dark:text-dark-400"
                  :title="item.notes"
                >
                  {{ item.notes.length > 60 ? item.notes.substring(0, 55) + '...' : item.notes }}
                </p>
                <p class="mt-0.5 text-xs text-gray-400 dark:text-dark-500">
                  {{ formatDateTime(item.used_at || item.created_at) }}
                </p>
              </div>
            </div>
            <!-- Right: value -->
            <div class="text-right">
              <p :class="['text-sm font-semibold', getValueColor(item)]">
                {{ formatValue(item) }}
              </p>
              <p
                v-if="isAdminType(item.type)"
                class="text-xs text-gray-400 dark:text-dark-500"
              >
                {{ t('redeem.adminAdjustment') }}
              </p>
              <p
                v-else
                class="font-mono text-xs text-gray-400 dark:text-dark-500"
              >
                {{ item.code.slice(0, 8) }}...
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex items-center justify-center gap-2 pt-2">
        <button
          :disabled="currentPage <= 1"
          class="btn btn-secondary px-3 py-1 text-sm"
          @click="loadHistory(currentPage - 1)"
        >
          {{ t('pagination.previous') }}
        </button>
        <span class="text-sm text-gray-500 dark:text-dark-400">
          {{ currentPage }} / {{ totalPages }}
        </span>
        <button
          :disabled="currentPage >= totalPages"
          class="btn btn-secondary px-3 py-1 text-sm"
          @click="loadHistory(currentPage + 1)"
        >
          {{ t('pagination.next') }}
        </button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI, type BalanceHistoryItem } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import { formatCreditAmount } from '@/utils/credits'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import { displayModelLabel } from '@/utils/modelDisplay'
import type { AdminUsageLog, AdminUser } from '@/types'
import type { PaymentOrder } from '@/types/payment'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'

const props = defineProps<{ show: boolean; user: AdminUser | null; hideActions?: boolean }>()
const emit = defineEmits(['close', 'deposit', 'withdraw'])
const { t } = useI18n()

const history = ref<BalanceHistoryItem[]>([])
const loading = ref(false)
const usageLoading = ref(false)
const usageError = ref('')
const subscriptionOrdersLoading = ref(false)
const subscriptionOrdersError = ref('')
const currentPage = ref(1)
const total = ref(0)
const totalRecharged = ref(0)
const recentUsage = ref<AdminUsageLog[]>([])
const subscriptionOrders = ref<PaymentOrder[]>([])
const pageSize = 15
const recentUsageLimit = 5
const subscriptionOrdersLimit = 5
const typeFilter = ref('')

const totalPages = computed(() => Math.ceil(total.value / pageSize) || 1)

// Type filter options
const typeOptions = computed(() => [
  { value: '', label: t('admin.users.allTypes') },
  { value: 'balance', label: t('admin.users.typeBalance') },
  { value: 'affiliate_balance', label: t('admin.users.typeAffiliateBalance') },
  { value: 'leaderboard_reward', label: t('admin.users.typeLeaderboardReward') },
  { value: 'new_user_reward', label: t('admin.users.typeNewUserReward') },
  { value: 'first_recharge_bonus', label: t('admin.users.typeFirstRechargeBonus') },
  { value: 'admin_balance', label: t('admin.users.typeAdminBalance') },
  { value: 'concurrency', label: t('admin.users.typeConcurrency') },
  { value: 'admin_concurrency', label: t('admin.users.typeAdminConcurrency') },
  { value: 'subscription', label: t('admin.users.typeSubscription') }
])

// Watch modal open or selected user switch.
watch(() => [props.show, props.user?.id] as const, ([show, id], [prevShow, prevId]) => {
  if (show && id && (!prevShow || id !== prevId)) {
    typeFilter.value = ''
    loadHistory(1)
    loadRecentUsage()
    loadSubscriptionOrders()
  }
})

const loadHistory = async (page: number) => {
  if (!props.user) return
  loading.value = true
  currentPage.value = page
  try {
    const res = await adminAPI.users.getUserBalanceHistory(
      props.user.id,
      page,
      pageSize,
      typeFilter.value || undefined
    )
    history.value = res.items || []
    total.value = res.total || 0
    totalRecharged.value = res.total_recharged || 0
  } catch (error) {
    console.error('Failed to load balance history:', error)
  } finally {
    loading.value = false
  }
}

const loadRecentUsage = async () => {
  if (!props.user) return
  usageLoading.value = true
  usageError.value = ''
  try {
    const res = await adminAPI.usage.list({
      user_id: props.user.id,
      page: 1,
      page_size: recentUsageLimit,
      sort_by: 'created_at',
      sort_order: 'desc'
    })
    recentUsage.value = res.items || []
  } catch (error) {
    console.error('Failed to load recent usage:', error)
    recentUsage.value = []
    usageError.value = t('admin.users.failedToLoadRecentUsage')
  } finally {
    usageLoading.value = false
  }
}

const loadSubscriptionOrders = async () => {
  if (!props.user) return
  subscriptionOrdersLoading.value = true
  subscriptionOrdersError.value = ''
  try {
    const res = await adminAPI.payment.getOrders({
      user_id: props.user.id,
      order_type: 'subscription',
      page: 1,
      page_size: subscriptionOrdersLimit,
    })
    subscriptionOrders.value = res.data.items || []
  } catch (error) {
    console.error('Failed to load subscription payment orders:', error)
    subscriptionOrders.value = []
    subscriptionOrdersError.value = t('admin.users.failedToLoadSubscriptionOrders')
  } finally {
    subscriptionOrdersLoading.value = false
  }
}

const toFiniteNumber = (value: unknown): number => {
  const numeric = Number(value)
  return Number.isFinite(numeric) ? numeric : 0
}

const formatUsageCost = (value: unknown): string =>
  formatCreditAmount(toFiniteNumber(value), {
    minimumFractionDigits: 4,
    maximumFractionDigits: 4,
  })

const formatSubscriptionOrderAmount = (order: PaymentOrder): string =>
  formatPaymentAmount(toFiniteNumber(order.pay_amount || order.amount), normalizePaymentCurrency(order.currency))

const formatSubscriptionOrderRefund = (order: PaymentOrder): string =>
  formatPaymentAmount(toFiniteNumber(order.refund_amount), normalizePaymentCurrency(order.currency))

const formatTokens = (value: number): string => {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  return value.toLocaleString()
}

const formatDuration = (ms: number | null | undefined): string => {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const getUsageTotalTokens = (log: AdminUsageLog): number => (
  toFiniteNumber(log.input_tokens) +
  toFiniteNumber(log.output_tokens) +
  toFiniteNumber(log.cache_creation_tokens) +
  toFiniteNumber(log.cache_read_tokens)
)

const getUsageTypeLabel = (log: AdminUsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
}

const getUsageTypeBadgeClass = (log: AdminUsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'ws_v2') return 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-200'
  if (requestType === 'stream') return 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-200'
  if (requestType === 'sync') return 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-200'
  return 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-200'
}

// Helper: check if admin type
const isAdminType = (type: string) => type === 'admin_balance' || type === 'admin_concurrency'

// Helper: check if balance type (includes admin_balance)
const isBalanceType = (type: string) => type === 'balance' || type === 'admin_balance' || type === 'affiliate_balance' || type === 'leaderboard_reward' || type === 'new_user_reward' || type === 'first_recharge_bonus'

// Helper: check if subscription type
const isSubscriptionType = (type: string) => type === 'subscription'

// Icon name based on type
const getIconName = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) return 'dollar'
  if (isSubscriptionType(item.type)) return 'badge'
  return 'bolt' // concurrency
}

// Icon background color
const getIconBg = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    return item.value >= 0
      ? 'bg-emerald-100 dark:bg-emerald-900/30'
      : 'bg-red-100 dark:bg-red-900/30'
  }
  if (isSubscriptionType(item.type)) return 'bg-purple-100 dark:bg-purple-900/30'
  return item.value >= 0
    ? 'bg-blue-100 dark:bg-blue-900/30'
    : 'bg-orange-100 dark:bg-orange-900/30'
}

// Icon text color
const getIconColor = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    return item.value >= 0
      ? 'text-emerald-600 dark:text-emerald-400'
      : 'text-red-600 dark:text-red-400'
  }
  if (isSubscriptionType(item.type)) return 'text-purple-600 dark:text-purple-400'
  return item.value >= 0
    ? 'text-blue-600 dark:text-blue-400'
    : 'text-orange-600 dark:text-orange-400'
}

// Value text color
const getValueColor = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    return item.value >= 0
      ? 'text-emerald-600 dark:text-emerald-400'
      : 'text-red-600 dark:text-red-400'
  }
  if (isSubscriptionType(item.type)) return 'text-purple-600 dark:text-purple-400'
  return item.value >= 0
    ? 'text-blue-600 dark:text-blue-400'
    : 'text-orange-600 dark:text-orange-400'
}

// Item title
const getItemTitle = (item: BalanceHistoryItem) => {
  switch (item.type) {
    case 'balance':
      return t('redeem.balanceAddedRedeem')
    case 'affiliate_balance':
      return t('redeem.balanceAddedAffiliate')
    case 'leaderboard_reward':
      return t('redeem.balanceAddedLeaderboardReward')
    case 'new_user_reward':
      return t('redeem.balanceAddedNewUserReward')
    case 'first_recharge_bonus':
      return t('redeem.balanceAddedFirstRechargeBonus')
    case 'admin_balance':
      return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
    case 'concurrency':
      return t('redeem.concurrencyAddedRedeem')
    case 'admin_concurrency':
      return item.value >= 0 ? t('redeem.concurrencyAddedAdmin') : t('redeem.concurrencyReducedAdmin')
    case 'subscription':
      return t('redeem.subscriptionAssigned')
    default:
      return t('common.unknown')
  }
}

// Format display value
const formatValue = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    return formatCreditAmount(item.value, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
      signDisplay: 'always',
    })
  }
  if (isSubscriptionType(item.type)) {
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}d - ${groupName}` : `${days}d`
  }
  // concurrency types
  const sign = item.value >= 0 ? '+' : ''
  return `${sign}${item.value}`
}
</script>
