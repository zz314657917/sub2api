<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('dashboard.quickActions') }}</h2>
    </div>
    <div class="space-y-3 p-4">
      <button @click="router.push('/keys')" class="console-action-row group flex w-full items-center gap-4 p-4 text-left transition-all duration-200">
        <div class="console-action-icon transition-transform group-hover:scale-105">
          <Icon name="key" size="md" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.createApiKey') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.generateNewKey') }}</p>
        </div>
        <Icon name="arrowRight" size="sm" />
      </button>

      <button @click="router.push('/usage')" class="console-action-row group flex w-full items-center gap-4 p-4 text-left transition-all duration-200">
        <div class="console-action-icon console-stat-icon--green transition-transform group-hover:scale-105">
          <Icon name="chartBar" size="md" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.viewUsage') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.checkDetailedLogs') }}</p>
        </div>
        <Icon name="arrowRight" size="sm" />
      </button>

      <button @click="router.push('/redeem')" class="console-action-row group flex w-full items-center gap-4 p-4 text-left transition-all duration-200">
        <div class="console-action-icon console-stat-icon--yellow transition-transform group-hover:scale-105">
          <Icon name="gift" size="md" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.redeemCode') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.addBalanceWithCode') }}</p>
        </div>
        <Icon name="arrowRight" size="sm" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
const router = useRouter()
const { t } = useI18n()
</script>
