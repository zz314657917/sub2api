<template>
  <section v-if="platformCards.length > 0" class="dashboard-panel dashboard-platform-panel">
    <div class="dashboard-platform-header">
      <h2 class="dashboard-panel-title">{{ t('dashboard.platformBreakdown') }}</h2>
      <span class="dashboard-platform-count">
        {{ t('dashboard.platformCount', { count: sortedPlatforms.length }) }}
      </span>
    </div>

    <div class="dashboard-platform-grid">
      <article
        v-for="item in platformCards"
        :key="item.platform"
        class="dashboard-platform-card"
        :class="{ 'dashboard-platform-card--other': item.isOther }"
      >
        <div class="dashboard-platform-card-header">
          <h3 class="dashboard-platform-name">
            {{ item.isOther ? t('dashboard.platformOther') : platformLabel(item.platform) }}
          </h3>
          <span class="dashboard-platform-total" :title="t('dashboard.actual')">
            {{ formatCost(item.total_actual_cost) }}
          </span>
        </div>

        <dl class="dashboard-platform-metrics">
          <div class="dashboard-platform-metric">
            <dt>{{ t('dashboard.todayCost') }}</dt>
            <dd>{{ formatCost(item.today_actual_cost) }}</dd>
          </div>
          <div class="dashboard-platform-metric">
            <dt>{{ t('dashboard.requests') }}</dt>
            <dd>{{ item.total_requests > 0 ? formatInteger(item.total_requests) : '-' }}</dd>
          </div>
          <div class="dashboard-platform-metric">
            <dt>{{ t('dashboard.tokens') }}</dt>
            <dd>{{ item.total_tokens > 0 ? formatTokens(item.total_tokens) : '-' }}</dd>
          </div>
        </dl>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatCreditAmount } from '@/utils/credits'
import type { PlatformDashboardStats, UserDashboardStats as UserStatsType } from '@/api/usage'

const OTHER_THRESHOLD = 0.0001

interface PlatformCard extends PlatformDashboardStats {
  isOther?: boolean
}

const props = defineProps<{
  stats: UserStatsType
}>()

const { t } = useI18n()

const PLATFORM_LABELS: Record<string, string> = {
  anthropic: 'Claude',
  openai: 'OpenAI',
  gemini: 'Gemini',
  antigravity: 'Antigravity',
}

const platformLabel = (platform: string): string => PLATFORM_LABELS[platform] ?? platform

const sortedPlatforms = computed(() => {
  const list = props.stats.by_platform ?? []
  return [...list].sort((a, b) => b.total_actual_cost - a.total_actual_cost)
})

const platformCards = computed<PlatformCard[]>(() => {
  const cards: PlatformCard[] = sortedPlatforms.value.map((platform) => ({ ...platform }))
  const sumTotal = cards.reduce((sum, item) => sum + item.total_actual_cost, 0)
  const sumToday = cards.reduce((sum, item) => sum + item.today_actual_cost, 0)
  const diffTotal = Math.max(0, (props.stats.total_actual_cost || 0) - sumTotal)
  const diffToday = Math.max(0, (props.stats.today_actual_cost || 0) - sumToday)

  if (diffTotal > OTHER_THRESHOLD || diffToday > OTHER_THRESHOLD) {
    cards.push({
      platform: '__other__',
      total_requests: 0,
      total_tokens: 0,
      total_actual_cost: diffTotal,
      today_requests: 0,
      today_tokens: 0,
      today_actual_cost: diffToday,
      isOther: true,
    })
  }

  return cards
})

const formatCost = (value: number): string =>
  formatCreditAmount(value, {
    minimumFractionDigits: 4,
    maximumFractionDigits: 4,
  })

const formatInteger = (value: number): string =>
  new Intl.NumberFormat(undefined).format(Math.trunc(Number(value) || 0))

const formatTokens = (value: number): string =>
  new Intl.NumberFormat(undefined, {
    notation: 'compact',
    maximumFractionDigits: Number(value) >= 1000 ? 1 : 0,
  }).format(Number(value) || 0)
</script>
