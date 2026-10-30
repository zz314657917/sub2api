<template>
  <section class="dashboard-stats-grid" :aria-label="t('dashboard.console.accountOverview')">
    <article
      v-for="card in cards"
      :key="card.key"
      class="dashboard-stat-card"
      :class="{ 'dashboard-stat-card--balance': card.key === 'balance' }"
    >
      <div class="dashboard-stat-header">
        <p class="dashboard-stat-label">{{ card.label }}</p>
        <Icon :name="card.icon" size="sm" />
      </div>
      <p class="dashboard-stat-value">{{ card.value }}</p>
      <p class="dashboard-stat-note">{{ card.note }}</p>
      <router-link
        v-if="card.key === 'balance'"
        :to="canPurchase ? '/purchase' : '/profile'"
        class="dashboard-inline-link"
        >{{ t(canPurchase ? 'nav.buySubscription' : 'nav.profile') }}
        <Icon name="arrowRight" size="xs"
      /></router-link>
      <router-link v-if="card.key === 'keys'" to="/keys" class="dashboard-inline-link"
        >{{ t('dashboard.apiKeys') }} <Icon name="arrowRight" size="xs"
      /></router-link>
    </article>
  </section>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { formatCreditAmount } from '@/utils/credits'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { useAuthStore } from '@/stores/auth'
import type { UserDashboardStats } from '@/api/usage'
const props = defineProps<{ stats: UserDashboardStats; balance?: number }>()
const { t } = useI18n()
const authStore = useAuthStore()
const canPurchase = computed(
  () => isFeatureFlagEnabled(FeatureFlags.payment) && !authStore.isSimpleMode
)
const cost = (value: number) =>
  formatCreditAmount(value, { minimumFractionDigits: 2, maximumFractionDigits: 4 })
const number = (value: number) => new Intl.NumberFormat().format(value)
const tokens = (value: number) =>
  new Intl.NumberFormat(undefined, { notation: 'compact', maximumFractionDigits: 1 }).format(value)
const cards = computed(() => [
  {
    key: 'balance',
    label: t('dashboard.balance'),
    icon: 'creditCard' as const,
    value: cost(props.balance ?? 0),
    note: t('dashboard.console.availableBalance')
  },
  {
    key: 'today-cost',
    label: t('dashboard.todayCost'),
    icon: 'creditCard' as const,
    value: cost(props.stats.today_actual_cost),
    note: t('dashboard.console.serverDayActual')
  },
  {
    key: 'today-tokens',
    label: t('dashboard.todayTokens'),
    icon: 'chart' as const,
    value: tokens(props.stats.today_tokens),
    note: t('dashboard.console.serverDayTokens')
  },
  {
    key: 'keys',
    label: t('dashboard.console.activeKeys'),
    icon: 'key' as const,
    value: number(props.stats.active_api_keys),
    note: t('dashboard.overview.activeApiKeys', {
      count: props.stats.active_api_keys,
      total: props.stats.total_api_keys
    })
  }
])
</script>
