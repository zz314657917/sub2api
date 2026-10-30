<template>
  <AppLayout>
    <div class="user-dashboard-page">
      <div class="dashboard-page-heading">
        <div>
          <p class="dashboard-eyebrow">{{ t('dashboard.console.workspace') }}</p>
          <h1>{{ t('dashboard.title') }}</h1>
          <p>{{ t('dashboard.welcomeMessage') }}</p>
        </div>
        <button
          v-if="user?.role !== 'admin'"
          type="button"
          class="btn btn-secondary inline-flex items-center gap-2"
          data-testid="quick-setup"
          @click="quickSetupOpen = true"
        >
          <Icon name="bolt" size="sm" />
          {{ t('dashboard.onboarding.quickSetup') }}
        </button>
      </div>
      <div v-if="loading" class="dashboard-loading-state">
        <LoadingSpinner />
      </div>
      <div v-else-if="statsError" class="dashboard-panel dashboard-chart-state" role="alert">
        <p>{{ t('dashboard.console.loadError') }}</p>
        <button type="button" class="btn btn-secondary" @click="loadStats">
          {{ t('dashboard.console.retry') }}
        </button>
      </div>
      <template v-else-if="stats">
        <UserDashboardStats :stats="stats" :balance="user?.balance" />

        <UserDashboardCharts
          v-model:startDate="startDate"
          v-model:endDate="endDate"
          v-model:granularity="granularity"
          :loading="loadingCharts"
          :trend="trendData"
          :error="chartsError"
          :first-use="stats.total_requests === 0"
          @dateRangeChange="loadRangeData"
          @retry="loadCharts"
        />
        <div class="dashboard-analytics-grid">
          <UserDashboardModels
            :models="modelData"
            :loading="loadingModels"
            :error="modelsError"
            @retry="loadModels"
          />
          <UserDashboardTokenTrend
            :trend="trendData"
            :loading="loadingCharts"
            :error="chartsError"
            @retry="loadCharts"
          />
        </div>
        <div class="dashboard-section-heading">
          <h2>{{ t('dashboard.performance') }}</h2>
          <p>{{ t('dashboard.console.performanceDescription') }}</p>
        </div>
        <UserDashboardPerformanceStats :stats="stats" />
        <UserDashboardPlatformBreakdown :stats="stats" />

        <div class="dashboard-secondary-grid">
          <UserDashboardRecentUsage
            :data="recentUsage"
            :loading="loadingUsage"
            :error="usageError"
            :start-date="startDate"
            :end-date="endDate"
            @retry="loadRecent"
          />
          <UserDashboardQuickActions />
        </div>

        <div
          v-if="accountShareEnabled && accountUsageError"
          class="dashboard-panel dashboard-chart-state"
          role="alert"
        >
          <p>{{ t('dashboard.console.accountLoadError') }}</p>
          <button type="button" class="btn btn-secondary" @click="loadAccountUsage">
            {{ t('dashboard.console.retry') }}
          </button>
        </div>
        <UserDashboardAccountUsage
          v-if="accountShareEnabled && !accountUsageError"
          :summary="accountUsageSummary"
          :loading="loadingAccountUsage"
          :start-date="startDate"
          :end-date="endDate"
        />
      </template>
    </div>
    <UserApiKeyOnboardingDialog
      :key="user?.id"
      :show="showApiKeyOnboarding"
      :has-api-key="hasApiKeyForOnboarding"
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
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import { userAPI } from '@/api/user'
import { welfareAPI } from '@/api/welfare'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'
import UserDashboardPerformanceStats from '@/components/user/dashboard/UserDashboardPerformanceStats.vue'
import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardModels from '@/components/user/dashboard/UserDashboardModels.vue'
import UserDashboardTokenTrend from '@/components/user/dashboard/UserDashboardTokenTrend.vue'
import { calendarRange } from '@/components/user/dashboard/dashboardRange'
import UserDashboardPlatformBreakdown from '@/components/user/dashboard/UserDashboardPlatformBreakdown.vue'
import UserDashboardAccountUsage from '@/components/user/dashboard/UserDashboardAccountUsage.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'
import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import UserApiKeyOnboardingDialog from '@/components/user/dashboard/UserApiKeyOnboardingDialog.vue'
import type {
  TrendDataPoint,
  UsageLog,
  UserAccountUsageSummary,
  WelfareNewUserTrial,
  ModelStat
} from '@/types'

const API_KEY_ONBOARDING_SKIP_DAYS = 7
const API_KEY_ONBOARDING_SKIP_MS = API_KEY_ONBOARDING_SKIP_DAYS * 24 * 60 * 60 * 1000

const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const user = computed(() => authStore.user)
const accountShareEnabled = computed(
  () => appStore.cachedPublicSettings?.account_share_enabled === true
)

const stats = ref<UserStatsType | null>(null)
const loading = ref(false)
const loadingUsage = ref(false)
const loadingCharts = ref(false)
const loadingAccountUsage = ref(false)
const loadingModels = ref(false)
const statsError = ref(false)
const chartsError = ref(false)
const modelsError = ref(false)
const usageError = ref(false)
const accountUsageError = ref(false)
const modelData = ref<ModelStat[]>([])
const trendData = ref<TrendDataPoint[]>([])
const recentUsage = ref<UsageLog[]>([])
const accountUsageSummary = ref<UserAccountUsageSummary | null>(null)
const apiKeyOnboardingDismissed = ref(false)
const quickSetupOpen = ref(false)
const welfareTrial = ref<WelfareNewUserTrial | null>(null)

const initialRange = calendarRange(7)
const startDate = ref(initialRange.startDate)
const endDate = ref(initialRange.endDate)
const granularity = ref('day')
const userTimezone = () => Intl.DateTimeFormat().resolvedOptions().timeZone || undefined
// Separate generations prevent old responses (including errors/finalizers) replacing a new range.
let chartGeneration = 0
let modelGeneration = 0
let recentGeneration = 0
let accountGeneration = 0
onBeforeUnmount(() => {
  chartGeneration++
  modelGeneration++
  recentGeneration++
  accountGeneration++
})

const loadStats = async () => {
  loading.value = true
  statsError.value = false
  try {
    await authStore.refreshUser()
    if (!apiKeyOnboardingDismissed.value) {
      apiKeyOnboardingDismissed.value = readApiKeyOnboardingSkip()
    }
    stats.value = await usageAPI.getDashboardStats()
  } catch (error) {
    statsError.value = true
    console.error('Failed to load dashboard stats:', error)
  } finally {
    loading.value = false
  }
}

const loadCharts = async () => {
  const generation = ++chartGeneration
  loadingCharts.value = true
  chartsError.value = false
  trendData.value = []
  try {
    const trend = await usageAPI.getDashboardTrend({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value as 'day' | 'hour',
      timezone: userTimezone()
    })
    if (generation === chartGeneration) trendData.value = trend.trend || []
  } catch (error) {
    if (generation === chartGeneration) chartsError.value = true
    console.error('Failed to load charts:', error)
  } finally {
    if (generation === chartGeneration) loadingCharts.value = false
  }
}

const loadModels = async () => {
  const generation = ++modelGeneration
  loadingModels.value = true
  modelsError.value = false
  modelData.value = []
  try {
    const response = await usageAPI.getDashboardModels({
      start_date: startDate.value,
      end_date: endDate.value,
      timezone: userTimezone()
    })
    if (generation === modelGeneration) modelData.value = response.models || []
  } catch (error) {
    if (generation === modelGeneration) modelsError.value = true
    console.error('Failed to load dashboard models:', error)
  } finally {
    if (generation === modelGeneration) loadingModels.value = false
  }
}

const loadRecent = async () => {
  const generation = ++recentGeneration
  loadingUsage.value = true
  usageError.value = false
  recentUsage.value = []
  try {
    const res = await usageAPI.query({
      start_date: startDate.value,
      end_date: endDate.value,
      timezone: userTimezone(),
      page: 1,
      page_size: 5
    })
    if (generation === recentGeneration) recentUsage.value = res.items.slice(0, 5)
  } catch (error) {
    if (generation === recentGeneration) usageError.value = true
    console.error('Failed to load recent usage:', error)
  } finally {
    if (generation === recentGeneration) loadingUsage.value = false
  }
}

const loadAccountUsage = async () => {
  const generation = ++accountGeneration
  accountUsageSummary.value = null
  accountUsageError.value = false
  if (!accountShareEnabled.value) {
    loadingAccountUsage.value = false
    return
  }
  loadingAccountUsage.value = true
  try {
    const summary = await userAPI.getAccountUsageSummary({
      start_date: startDate.value,
      end_date: endDate.value,
      timezone: userTimezone()
    })
    if (generation === accountGeneration)
      accountUsageSummary.value = accountShareEnabled.value ? summary : null
  } catch (error) {
    console.error('Failed to load account usage summary:', error)
    if (generation === accountGeneration) accountUsageError.value = true
  } finally {
    if (generation === accountGeneration) loadingAccountUsage.value = false
  }
}

watch(accountShareEnabled, () => {
  void loadAccountUsage()
})

const loadWelfareTrial = async () => {
  if (user.value?.role === 'admin') {
    welfareTrial.value = null
    return
  }
  try {
    const overview = await welfareAPI.getWelfareOverview()
    welfareTrial.value =
      overview.enabled && overview.new_user_trial?.enabled ? overview.new_user_trial : null
  } catch (error) {
    console.error('Failed to load welfare trial:', error)
    welfareTrial.value = null
  }
}

const loadRangeData = () => {
  loadCharts()
  loadModels()
  loadRecent()
  loadAccountUsage()
}

const refreshAll = () => {
  loadStats()
  loadWelfareTrial()
  loadRangeData()
}

const apiKeyOnboardingStorageKey = computed(() => {
  return user.value?.id ? `sub2api:user-api-key-onboarding:skip:${user.value.id}` : ''
})

const showApiKeyOnboarding = computed(() => {
  return Boolean(
    user.value?.role !== 'admin' &&
      (quickSetupOpen.value ||
        (stats.value && stats.value.total_requests === 0 && !apiKeyOnboardingDismissed.value))
  )
})

const hasApiKeyForOnboarding = computed(() => Number(stats.value?.total_api_keys || 0) > 0)

const onboardingAmountFormatter = new Intl.NumberFormat(undefined, {
  minimumFractionDigits: 0,
  maximumFractionDigits: 8
})

const formatOnboardingAmount = (value: number | null | undefined) => {
  return onboardingAmountFormatter.format(Number(value) || 0)
}

const hasTrialBenefit = computed(() => {
  const trial = welfareTrial.value
  return Boolean(trial?.enabled && trial.can_use && Number(trial.quota_amount) > 0)
})

const hasRewardBenefit = computed(() => {
  const trial = welfareTrial.value
  return Boolean(
    trial?.enabled && !trial.success_reward_claimed && Number(trial.success_reward_amount) > 0
  )
})

const hasOnboardingBenefit = computed(() => hasRewardBenefit.value || hasTrialBenefit.value)
const onboardingBenefitKind = computed<'reward' | 'trial'>(() =>
  hasRewardBenefit.value ? 'reward' : 'trial'
)

const onboardingBenefitLabel = computed(() => {
  const trial = welfareTrial.value
  if (hasRewardBenefit.value) {
    return t('dashboard.onboarding.rewardAmount', {
      amount: formatOnboardingAmount(trial?.success_reward_amount)
    })
  }
  return hasTrialBenefit.value
    ? t('dashboard.onboarding.trialQuotaAmount', {
        amount: formatOnboardingAmount(trial?.quota_amount)
      })
    : ''
})

const onboardingBenefitRewardLabel = computed(() => {
  if (hasRewardBenefit.value) {
    return onboardingBenefitLabel.value
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
  quickSetupOpen.value = false
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

const closeApiKeyOnboardingForNavigation = () => {
  quickSetupOpen.value = false
  apiKeyOnboardingDismissed.value = true
}

const startApiKeyOnboardingCreate = async () => {
  closeApiKeyOnboardingForNavigation()
  await router.push(
    hasApiKeyForOnboarding.value ? { path: '/keys' } : { path: '/keys', query: { create: '1' } }
  )
}

const openApiKeyTutorial = async () => {
  closeApiKeyOnboardingForNavigation()
  await router.push('/tutorial/getting-started')
}

onMounted(() => {
  apiKeyOnboardingDismissed.value = readApiKeyOnboardingSkip()
  refreshAll()
})
</script>
