<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-12"><LoadingSpinner /></div>
      <template v-else-if="stats">
        <UserDashboardStats :stats="stats" :balance="user?.balance || 0" :is-simple="authStore.isSimpleMode" />
        <UserDashboardAccountUsage :summary="accountUsageSummary" :loading="loadingAccountUsage" :start-date="startDate" :end-date="endDate" />
        <UserDashboardCharts v-model:startDate="startDate" v-model:endDate="endDate" v-model:granularity="granularity" :loading="loadingCharts" :trend="trendData" :models="modelStats" @dateRangeChange="loadRangeData" @granularityChange="loadCharts" @refresh="refreshAll" />
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="lg:col-span-2"><UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" /></div>
          <div class="lg:col-span-1"><UserDashboardQuickActions /></div>
        </div>
      </template>
    </div>
    <UserApiKeyOnboardingDialog
      :show="showApiKeyOnboarding"
      :has-benefit="hasOnboardingBenefit"
      :benefit-kind="onboardingBenefitKind"
      :benefit-label="onboardingBenefitLabel"
      :benefit-reward-label="onboardingBenefitRewardLabel"
      @create="startApiKeyOnboardingCreate"
      @tutorial="openApiKeyTutorial"
      @skip="skipApiKeyOnboarding"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'; import { useRouter } from 'vue-router'; import { useI18n } from 'vue-i18n'; import { useAuthStore } from '@/stores/auth'; import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import { userAPI } from '@/api/user'
import { welfareAPI } from '@/api/welfare'
import AppLayout from '@/components/layout/AppLayout.vue'; import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'; import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardAccountUsage from '@/components/user/dashboard/UserDashboardAccountUsage.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'; import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import UserApiKeyOnboardingDialog from '@/components/user/dashboard/UserApiKeyOnboardingDialog.vue'
import type { UsageLog, TrendDataPoint, ModelStat, UserAccountUsageSummary, WelfareNewUserTrial } from '@/types'

const API_KEY_ONBOARDING_SKIP_DAYS = 7
const API_KEY_ONBOARDING_SKIP_MS = API_KEY_ONBOARDING_SKIP_DAYS * 24 * 60 * 60 * 1000

const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore(); const user = computed(() => authStore.user)
const stats = ref<UserStatsType | null>(null); const loading = ref(false); const loadingUsage = ref(false); const loadingCharts = ref(false); const loadingAccountUsage = ref(false)
const trendData = ref<TrendDataPoint[]>([]); const modelStats = ref<ModelStat[]>([]); const recentUsage = ref<UsageLog[]>([])
const accountUsageSummary = ref<UserAccountUsageSummary | null>(null)
const apiKeyOnboardingDismissed = ref(false)
const welfareTrial = ref<WelfareNewUserTrial | null>(null)

const formatLD = (d: Date) => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
const startDate = ref(formatLD(new Date(Date.now() - 6 * 86400000))); const endDate = ref(formatLD(new Date())); const granularity = ref('day')
const userTimezone = () => Intl.DateTimeFormat().resolvedOptions().timeZone || undefined

const loadStats = async () => { loading.value = true; try { await authStore.refreshUser(); if (!apiKeyOnboardingDismissed.value) { apiKeyOnboardingDismissed.value = readApiKeyOnboardingSkip() } stats.value = await usageAPI.getDashboardStats() } catch (error) { console.error('Failed to load dashboard stats:', error) } finally { loading.value = false } }
const loadCharts = async () => { loadingCharts.value = true; try { const res = await Promise.all([usageAPI.getDashboardTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any }), usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })]); trendData.value = res[0].trend || []; modelStats.value = res[1].models || [] } catch (error) { console.error('Failed to load charts:', error) } finally { loadingCharts.value = false } }
const loadRecent = async () => { loadingUsage.value = true; try { const res = await usageAPI.getByDateRange(startDate.value, endDate.value); recentUsage.value = res.items.slice(0, 5) } catch (error) { console.error('Failed to load recent usage:', error) } finally { loadingUsage.value = false } }
const loadAccountUsage = async () => { loadingAccountUsage.value = true; try { accountUsageSummary.value = await userAPI.getAccountUsageSummary({ start_date: startDate.value, end_date: endDate.value, timezone: userTimezone() }) } catch (error) { console.error('Failed to load account usage summary:', error); accountUsageSummary.value = null } finally { loadingAccountUsage.value = false } }
const loadWelfareTrial = async () => {
  if (user.value?.role === 'admin') {
    welfareTrial.value = null
    return
  }
  try {
    const overview = await welfareAPI.getWelfareOverview()
    welfareTrial.value = overview.enabled && overview.new_user_trial?.enabled
      ? overview.new_user_trial
      : null
  } catch (error) {
    console.error('Failed to load welfare trial:', error)
    welfareTrial.value = null
  }
}
const loadRangeData = () => { loadCharts(); loadRecent(); loadAccountUsage() }
const refreshAll = () => { loadStats(); loadWelfareTrial(); loadRangeData() }

const apiKeyOnboardingStorageKey = computed(() => {
  return user.value?.id ? `sub2api:user-api-key-onboarding:skip:${user.value.id}` : ''
})

const showApiKeyOnboarding = computed(() => {
  return Boolean(
    stats.value &&
    user.value?.role !== 'admin' &&
    stats.value.total_api_keys === 0 &&
    !apiKeyOnboardingDismissed.value
  )
})

const onboardingAmountFormatter = new Intl.NumberFormat(undefined, {
  minimumFractionDigits: 0,
  maximumFractionDigits: 8,
})

const formatOnboardingAmount = (value: number | null | undefined) => {
  return onboardingAmountFormatter.format(Number(value) || 0)
}

const hasTrialBenefit = computed(() => {
  const trial = welfareTrial.value
  return Boolean(trial?.enabled && trial.can_use && Number(trial.quota_amount) > 0)
})

const hasWalletBenefit = computed(() => Number(user.value?.balance) > 0)
const hasOnboardingBenefit = computed(() => hasWalletBenefit.value || hasTrialBenefit.value)
const onboardingBenefitKind = computed<'wallet' | 'trial'>(() => hasWalletBenefit.value ? 'wallet' : 'trial')

const onboardingBenefitLabel = computed(() => {
  if (hasWalletBenefit.value) {
    return t('dashboard.onboarding.balanceAmount', { amount: formatOnboardingAmount(user.value?.balance) })
  }
  const trial = welfareTrial.value
  return hasTrialBenefit.value
    ? t('dashboard.onboarding.trialQuotaAmount', { amount: formatOnboardingAmount(trial?.quota_amount) })
    : ''
})

const onboardingBenefitRewardLabel = computed(() => {
  if (hasWalletBenefit.value) {
    return ''
  }
  const trial = welfareTrial.value
  return hasTrialBenefit.value && Number(trial?.success_reward_amount) > 0
    ? formatOnboardingAmount(trial?.success_reward_amount)
    : ''
})

const readApiKeyOnboardingSkip = () => {
  const key = apiKeyOnboardingStorageKey.value
  if (!key) return false
  try {
    const raw = window.localStorage.getItem(key)
    const skippedAt = raw ? Number(raw) : 0
    return Number.isFinite(skippedAt) && Date.now() - skippedAt < API_KEY_ONBOARDING_SKIP_MS
  } catch {
    return false
  }
}

const skipApiKeyOnboarding = () => {
  apiKeyOnboardingDismissed.value = true
  const key = apiKeyOnboardingStorageKey.value
  try {
    if (key) {
      window.localStorage.setItem(key, String(Date.now()))
    }
  } catch {
    // Ignore storage failures; the dialog is still dismissed for this page session.
  }
}

const markApiKeyOnboardingDone = () => {
  apiKeyOnboardingDismissed.value = true
  const key = apiKeyOnboardingStorageKey.value
  try {
    if (key) {
      window.localStorage.setItem(key, String(Date.now()))
    }
  } catch {
    // Ignore storage failures; navigation should continue.
  }
}

const startApiKeyOnboardingCreate = async () => {
  markApiKeyOnboardingDone()
  await router.push({ path: '/keys', query: { create: '1' } })
}

const openApiKeyTutorial = async () => {
  markApiKeyOnboardingDone()
  await router.push('/tutorial/getting-started')
}

onMounted(() => {
  apiKeyOnboardingDismissed.value = readApiKeyOnboardingSkip()
  refreshAll()
})
</script>
