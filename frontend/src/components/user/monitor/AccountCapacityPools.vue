<template>
  <section class="mb-5 grid gap-4" data-testid="account-capacity-pools">
    <article
      v-for="pool in orderedPools"
      :key="pool.key"
      class="relative overflow-hidden rounded-3xl border border-gray-200/80 dark:border-dark-700/70 bg-white/80 dark:bg-dark-800/70 shadow-sm"
    >
      <div class="absolute inset-x-0 top-0 h-1" :class="pool.key === 'mine' ? 'bg-sky-500' : 'bg-emerald-500'"></div>
      <div class="p-5">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-400">
              {{ t(`channelStatus.capacityPools.${pool.key}`) }}
            </p>
            <h3 class="mt-1 text-lg font-bold text-gray-900 dark:text-white">
              {{ poolTitle(pool) }}
            </h3>
          </div>
          <span
            class="rounded-full px-2.5 py-1 text-xs font-semibold"
            :class="pool.schedulable_accounts > 0
              ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
              : 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'"
          >
            {{ pool.schedulable_accounts }}/{{ pool.total_accounts }}
            {{ t('channelStatus.capacityPools.schedulable') }}
          </span>
        </div>

        <div class="mt-4 grid grid-cols-3 gap-2">
          <div class="rounded-2xl bg-gray-50 dark:bg-dark-900/40 p-3">
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('channelStatus.capacityPools.total') }}</p>
            <p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">{{ pool.total_accounts }}</p>
          </div>
          <div class="rounded-2xl bg-gray-50 dark:bg-dark-900/40 p-3">
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('channelStatus.capacityPools.configuredQuota') }}</p>
            <p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">{{ formatQuota(pool.configured_quota) }}</p>
          </div>
          <div class="rounded-2xl bg-gray-50 dark:bg-dark-900/40 p-3">
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('channelStatus.capacityPools.remainingQuota') }}</p>
            <p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">{{ formatQuota(pool.remaining_quota) }}</p>
          </div>
        </div>

        <div class="mt-4 space-y-2">
          <div
            v-for="section in pool.sections"
            :key="`${pool.key}-${section.platform}-${section.type}`"
            class="rounded-2xl border border-gray-100 dark:border-dark-700/70 bg-white/70 dark:bg-dark-900/30 p-3"
          >
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-sm font-semibold text-gray-900 dark:text-white">
                  {{ platformLabel(section.platform) }} / {{ typeLabel(section.type) }}
                </p>
                <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ section.schedulable_accounts }}/{{ section.total_accounts }}
                  {{ t('channelStatus.capacityPools.schedulable') }}
                </p>
              </div>
              <div class="flex flex-wrap justify-end gap-1.5">
                <span
                  v-for="window in windowBadges(section)"
                  :key="window.key"
                  class="rounded-full bg-gray-100 px-2 py-1 text-xs font-semibold text-gray-700 dark:bg-dark-700 dark:text-gray-200"
                >
                  {{ window.label }}
                </span>
              </div>
            </div>
          </div>

          <p
            v-if="pool.sections.length === 0"
            class="rounded-2xl border border-dashed border-gray-200 dark:border-dark-700 p-4 text-sm text-gray-500 dark:text-gray-400"
          >
            {{ t('channelStatus.capacityPools.empty') }}
          </p>
        </div>
      </div>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserAccountCapacityPool, UserAccountCapacityPools, UserAccountCapacityPoolSection } from '@/types'

const props = defineProps<{
  pools: UserAccountCapacityPools | null
  loading?: boolean
}>()

const { t } = useI18n()

const orderedPools = computed<UserAccountCapacityPool[]>(() => {
  if (!props.pools) {
    return []
  }
  return [props.pools.mine].filter(Boolean)
})

function poolTitle(pool: UserAccountCapacityPool): string {
  if (pool.key === 'mine') return t('channelStatus.capacityPools.mineTitle')
  if (pool.key === 'shared') return t('channelStatus.capacityPools.sharedTitle')
  return pool.title
}

function formatQuota(value: number): string {
  if (!Number.isFinite(value) || value <= 0) {
    return '-'
  }
  return value.toLocaleString(undefined, {
    maximumFractionDigits: value >= 10 ? 0 : 2,
  })
}

function platformLabel(platform: string): string {
  const key = `myAccounts.platforms.${platform}`
  const label = t(key)
  return label === key ? platform : label
}

function typeLabel(type: string): string {
  const key = `myAccounts.types.${type}`
  const label = t(key)
  return label === key ? type : label
}

function windowBadges(section: UserAccountCapacityPoolSection): Array<{ key: string; label: string }> {
  const windows = section.windows ?? {}
  return ['5h', '7d']
    .map((key) => {
      const snapshot = windows[key]
      if (!snapshot) return null
      return {
        key,
        label: `${key} ${Math.round(snapshot.used_percent)}%`,
      }
    })
    .filter((item): item is { key: string; label: string } => item !== null)
}
</script>
