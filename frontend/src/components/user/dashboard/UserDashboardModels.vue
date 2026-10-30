<template>
  <section class="dashboard-panel">
    <div class="dashboard-panel-header">
      <div>
        <p class="dashboard-eyebrow">{{ t('dashboard.console.modelInsights') }}</p>
        <h2 class="dashboard-panel-title">{{ t('dashboard.console.modelSpendingRank') }}</h2>
        <p class="dashboard-panel-subtitle">{{ t('dashboard.console.modelDescription') }}</p>
      </div>
      <router-link to="/usage" class="dashboard-inline-link"
        >{{ t('dashboard.viewUsage') }} <Icon name="arrowRight" size="xs"
      /></router-link>
    </div>
    <div v-if="loading" class="dashboard-chart-state" role="status">
      <LoadingSpinner size="md" />
    </div>
    <div v-else-if="error" class="dashboard-chart-state" role="alert">
      <p>{{ t('dashboard.console.loadError') }}</p>
      <button class="btn btn-secondary" type="button" @click="$emit('retry')">
        {{ t('dashboard.console.retry') }}
      </button>
    </div>
    <div v-else-if="!models.length" class="dashboard-model-empty">
      {{ t('dashboard.console.noModels') }}
    </div>
    <div v-else class="dashboard-model-table-wrap">
      <table class="dashboard-model-table">
        <thead>
          <tr>
            <th>{{ t('dashboard.model') }}</th>
            <th>{{ t('dashboard.requests') }}</th>
            <th>{{ t('dashboard.tokens') }}</th>
            <th>{{ t('dashboard.actual') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(model, index) in rankedModels" :key="model.model">
            <td>
              <div class="dashboard-model-name">
                <span class="dashboard-model-rank">{{ String(index + 1).padStart(2, '0') }}</span
                ><span :title="model.model">{{ displayModelLabel(model.model) }}</span>
              </div>
              <div class="dashboard-model-track">
                <i :style="{ width: `${maxCost ? (model.actual_cost / maxCost) * 100 : 0}%` }" />
              </div>
            </td>
            <td>{{ model.requests.toLocaleString() }}</td>
            <td>{{ compact(model.total_tokens) }}</td>
            <td>{{ formatCreditAmount(model.actual_cost, { maximumFractionDigits: 4 }) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { displayModelLabel } from '@/utils/modelDisplay'
import { formatCreditAmount } from '@/utils/credits'
import type { ModelStat } from '@/types'
const props = defineProps<{ models: ModelStat[]; loading: boolean; error: boolean }>()
defineEmits<{ retry: [] }>()
const { t } = useI18n()
const rankedModels = computed(() =>
  [...props.models]
    .sort((a, b) => b.actual_cost - a.actual_cost || b.requests - a.requests)
    .slice(0, 6)
)
const maxCost = computed(() => Math.max(0, ...props.models.map((m) => m.actual_cost)))
const compact = (v: number) =>
  new Intl.NumberFormat(undefined, { notation: 'compact', maximumFractionDigits: 1 }).format(v)
</script>
