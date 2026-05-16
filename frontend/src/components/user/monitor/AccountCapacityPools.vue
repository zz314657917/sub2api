<template>
  <section class="mb-4 grid grid-cols-1 items-start gap-3 xl:grid-cols-2" data-testid="account-capacity-pools">
    <article
      v-for="pool in orderedPools"
      :key="pool.key"
      class="overflow-hidden rounded-lg border border-gray-200/80 bg-white/85 shadow-sm dark:border-dark-700/70 dark:bg-dark-800/70"
    >
      <div :class="isEmptyPool(pool) ? 'p-3 sm:p-4' : 'p-4 sm:p-5'">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="flex min-w-0 items-start gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg"
              :class="pool.key === 'mine'
                ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
                : 'bg-cyan-100 text-cyan-700 dark:bg-cyan-500/15 dark:text-cyan-300'"
            >
              <Icon name="calculator" size="sm" />
            </div>
            <div class="min-w-0">
              <h3 class="text-base font-bold text-gray-900 dark:text-white">
                {{ poolTitle(pool) }}
              </h3>
              <p class="mt-0.5 text-xs leading-5 text-gray-500 dark:text-gray-400">
                {{ poolDescription(pool) }}
              </p>
            </div>
          </div>
        </div>

        <p
          v-if="isEmptyPool(pool)"
          class="mt-3 rounded-lg border border-dashed border-gray-200 px-3 py-2 text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400"
        >
          {{ t('channelStatus.capacityPools.empty') }}
        </p>

        <template v-else>
          <div
            :class="[
              'mt-4 grid grid-cols-2 gap-2',
              pool.key === 'shared' && positiveCount(pool.own_contributed_accounts) > 0 ? 'sm:grid-cols-5' : 'sm:grid-cols-4',
            ]"
          >
            <MetricTile :label="t('channelStatus.capacityPools.total')" :value="formatInteger(pool.total_accounts)" />
            <MetricTile
              v-if="pool.key === 'shared' && positiveCount(pool.own_contributed_accounts) > 0"
              :label="t('channelStatus.capacityPools.ownContributed')"
              :value="formatInteger(pool.own_contributed_accounts ?? 0)"
              tone="success"
            />
            <MetricTile
              :label="t('channelStatus.capacityPools.schedulable')"
              :value="formatInteger(pool.schedulable_accounts)"
              tone="success"
            />
            <MetricTile
              :label="t('channelStatus.capacityPools.rateLimited')"
              :value="formatInteger(pool.rate_limited_accounts ?? 0)"
              tone="warning"
            />
            <MetricTile
              :label="t('channelStatus.capacityPools.abnormal')"
              :value="formatInteger(pool.abnormal_accounts ?? 0)"
              tone="danger"
            />
          </div>

          <SharedGroupGrid v-if="pool.key === 'shared' && sharedGroups(pool).length > 0" :pool="pool" />
          <SectionFallback v-else :pool="pool" />
        </template>
      </div>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed, defineComponent, h } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { PropType } from 'vue'
import type {
  UserAccountCapacityPool,
  UserAccountCapacityPoolGroup,
  UserAccountCapacityPools,
  UserAccountCapacityPoolSection,
  UserAccountCapacityWindowSnapshot,
  UserAccountCapacityWindowSummary,
} from '@/types'

const props = defineProps<{
  pools: UserAccountCapacityPools | null
  loading?: boolean
}>()

const { t } = useI18n()

const orderedPools = computed<UserAccountCapacityPool[]>(() => {
  if (!props.pools) {
    return []
  }
  return [props.pools.mine, props.pools.shared].filter(Boolean)
})

function poolTitle(pool: UserAccountCapacityPool): string {
  if (pool.key === 'mine') return t('channelStatus.capacityPools.mineTitle')
  if (pool.key === 'shared') return t('channelStatus.capacityPools.sharedTitle')
  return pool.title
}

function poolDescription(pool: UserAccountCapacityPool): string {
  if (pool.key === 'mine') return t('channelStatus.capacityPools.mineDescription')
  if (pool.key === 'shared') return t('channelStatus.capacityPools.sharedDescription')
  return ''
}

function sharedGroups(pool: UserAccountCapacityPool): UserAccountCapacityPoolGroup[] {
  return pool.groups ?? []
}

function isEmptyPool(pool: UserAccountCapacityPool): boolean {
  return pool.total_accounts <= 0 && (pool.sections?.length ?? 0) === 0 && sharedGroups(pool).length === 0
}

function formatInteger(value: number): string {
  if (!Number.isFinite(value)) {
    return '0'
  }
  return Math.round(value).toLocaleString()
}

function formatQuota(value: number): string {
  if (!Number.isFinite(value) || value <= 0) {
    return '-'
  }
  return value.toLocaleString(undefined, {
    maximumFractionDigits: value >= 10 ? 0 : 2,
  })
}

function formatPercent(value: number): string {
  if (!Number.isFinite(value)) {
    return '0.0%'
  }
  return `${value.toFixed(1)}%`
}

function formatRemainingUnits(value: number): string {
  if (!Number.isFinite(value) || value <= 0) {
    return '0'
  }
  return value.toLocaleString(undefined, {
    maximumFractionDigits: value >= 10 ? 0 : 2,
  })
}

function positiveCount(value?: number | null): number {
  if (!Number.isFinite(value ?? 0)) {
    return 0
  }
  return Math.max(0, Math.round(value ?? 0))
}

function platformLabel(platform?: string): string {
  if (!platform) return ''
  const key = `myAccounts.platforms.${platform}`
  const label = t(key)
  return label === key ? platform : label
}

function typeLabel(type: string): string {
  const key = `myAccounts.types.${type}`
  const label = t(key)
  return label === key ? type : label
}

function groupStatusTone(status: string): 'success' | 'warning' | 'danger' {
  if (status === 'healthy') return 'success'
  if (status === 'unavailable') return 'danger'
  return 'warning'
}

function groupStatusLabel(status: string): string {
  if (status === 'healthy') return t('channelStatus.capacityPools.healthy')
  if (status === 'unavailable') return t('channelStatus.capacityPools.unavailable')
  return t('channelStatus.capacityPools.degraded')
}

function groupBorderClass(status: string): string {
  if (status === 'healthy') return 'border-emerald-200/80 bg-emerald-50/35 dark:border-emerald-500/30 dark:bg-emerald-500/5'
  if (status === 'unavailable') return 'border-rose-200/80 bg-rose-50/35 dark:border-rose-500/30 dark:bg-rose-500/5'
  return 'border-amber-200/90 bg-amber-50/40 dark:border-amber-500/30 dark:bg-amber-500/5'
}

function chipClass(tone: 'success' | 'warning' | 'danger' | 'neutral'): string {
  switch (tone) {
    case 'success':
      return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
    case 'warning':
      return 'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300'
    case 'danger':
      return 'bg-rose-100 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300'
    case 'neutral':
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
  }
}

function windowSummaries(group: UserAccountCapacityPoolGroup): Array<{ key: string; data: UserAccountCapacityWindowSummary }> {
  const windows = group.windows ?? {}
  return ['1d', '7d_quota', '30d', '5h', '7d']
    .map((key) => {
      const data = windows[key]
      return data ? { key, data } : null
    })
    .filter((item): item is { key: string; data: UserAccountCapacityWindowSummary } => item !== null)
}

function windowBadges(section: UserAccountCapacityPoolSection): Array<{ key: string; data: UserAccountCapacityWindowSnapshot }> {
  const windows = section.windows ?? {}
  return ['1d', '7d_quota', '30d', '5h', '7d']
    .map((key) => {
      const data = windows[key]
      return data ? { key, data } : null
    })
    .filter((item): item is { key: string; data: UserAccountCapacityWindowSnapshot } => item !== null)
}

function windowLabel(key: string): string {
  switch (key) {
    case '1d':
      return t('channelStatus.capacityPools.quotaWindow', { window: '1d' })
    case '7d_quota':
      return t('channelStatus.capacityPools.quotaWindow', { window: '7d' })
    case '30d':
      return t('channelStatus.capacityPools.quotaWindow', { window: '30d' })
    default:
      return t('channelStatus.capacityPools.window', { window: key })
  }
}

const unavailableReasonOrder = [
  'daily_quota_exceeded',
  'weekly_quota_exceeded',
  'monthly_quota_exceeded',
  'quota_exceeded',
  'rate_limited',
  'overloaded',
  'temp_unschedulable',
  'manual_unschedulable',
  'error',
  'disabled',
  'expired',
  'unused',
  'inactive',
  'unschedulable',
]

function unavailableReasonLabel(key: string): string {
  const messageKey = `channelStatus.capacityPools.unavailableReasons.${key}`
  const label = t(messageKey)
  return label === messageKey ? key : label
}

function unavailableReasonTone(key: string): 'warning' | 'danger' | 'neutral' {
  if (key.includes('quota') || key === 'rate_limited' || key === 'overloaded' || key === 'temp_unschedulable') {
    return 'warning'
  }
  if (key === 'error' || key === 'disabled' || key === 'expired') {
    return 'danger'
  }
  return 'neutral'
}

function unavailableReasonEntries(reasons?: Record<string, number>): Array<{ key: string; label: string; count: number; tone: 'warning' | 'danger' | 'neutral' }> {
  if (!reasons) return []
  return Object.entries(reasons)
    .map(([key, value]) => ({
      key,
      label: unavailableReasonLabel(key),
      count: positiveCount(value),
      tone: unavailableReasonTone(key),
    }))
    .filter((item) => item.count > 0)
    .sort((a, b) => {
      const left = unavailableReasonOrder.indexOf(a.key)
      const right = unavailableReasonOrder.indexOf(b.key)
      const normalizedLeft = left === -1 ? unavailableReasonOrder.length : left
      const normalizedRight = right === -1 ? unavailableReasonOrder.length : right
      if (normalizedLeft !== normalizedRight) return normalizedLeft - normalizedRight
      return a.key.localeCompare(b.key)
    })
}

const MetricTile = defineComponent({
  name: 'MetricTile',
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
    tone: {
      type: String as PropType<'neutral' | 'success' | 'warning' | 'danger'>,
      default: 'neutral',
    },
  },
  setup(tileProps) {
    return () => h('div', { class: 'rounded-lg bg-gray-50 p-3 dark:bg-dark-900/40' }, [
      h('p', { class: 'text-xs text-gray-500 dark:text-gray-400' }, tileProps.label),
      h('p', {
        class: [
          'mt-1 text-xl font-bold',
          tileProps.tone === 'success'
            ? 'text-emerald-600 dark:text-emerald-300'
            : tileProps.tone === 'warning'
              ? 'text-amber-600 dark:text-amber-300'
              : tileProps.tone === 'danger'
                ? 'text-rose-600 dark:text-rose-300'
                : 'text-gray-900 dark:text-white',
        ],
      }, tileProps.value),
    ])
  },
})

const SharedGroupGrid = defineComponent({
  name: 'SharedGroupGrid',
  props: {
    pool: { type: Object as PropType<UserAccountCapacityPool>, required: true },
  },
  setup(componentProps) {
    return () => h('div', {
      class: 'mt-4 rounded-lg border border-gray-200/80 p-3 dark:border-dark-700/70',
    }, [
      h('div', { class: 'mb-3 flex flex-wrap items-center justify-between gap-2' }, [
        h('div', [
          h('p', { class: 'text-sm font-semibold text-gray-900 dark:text-white' }, t('channelStatus.capacityPools.groupCapacity')),
          h('p', { class: 'mt-0.5 text-xs text-gray-500 dark:text-gray-400' }, t('channelStatus.capacityPools.groupCapacityHint')),
        ]),
        h('div', { class: 'flex flex-wrap gap-1.5' }, [
          h('span', { class: ['rounded-md px-2 py-1 text-xs font-semibold', chipClass('success')] }, `${t('channelStatus.capacityPools.healthy')} ${componentProps.pool.groups?.filter(g => g.status === 'healthy').length ?? 0}`),
          h('span', { class: ['rounded-md px-2 py-1 text-xs font-semibold', chipClass('warning')] }, `${t('channelStatus.capacityPools.degraded')} ${componentProps.pool.groups?.filter(g => g.status === 'degraded').length ?? 0}`),
          h('span', { class: ['rounded-md px-2 py-1 text-xs font-semibold', chipClass('danger')] }, `${t('channelStatus.capacityPools.unavailable')} ${componentProps.pool.groups?.filter(g => g.status === 'unavailable').length ?? 0}`),
        ]),
      ]),
      h('div', { class: 'grid grid-cols-1 gap-2 2xl:grid-cols-2' }, (componentProps.pool.groups ?? []).map((group) => (
        h('div', {
          key: group.key,
          class: ['rounded-lg border p-3', groupBorderClass(group.status)],
        }, [
          h('div', { class: 'flex items-start justify-between gap-3' }, [
            h('div', { class: 'min-w-0' }, [
              h('div', { class: 'flex min-w-0 flex-wrap items-center gap-1.5' }, [
                h('p', { class: 'min-w-0 truncate text-sm font-bold text-gray-900 dark:text-white' }, group.group_name),
                componentProps.pool.key === 'shared' && positiveCount(group.own_contributed_accounts) > 0
                  ? h('span', { class: ['shrink-0 rounded px-2 py-0.5 text-xs font-semibold', chipClass('success')] }, `${t('channelStatus.capacityPools.ownContributed')} ${formatInteger(group.own_contributed_accounts ?? 0)}`)
                  : null,
              ]),
              h('p', { class: 'mt-0.5 text-xs text-gray-500 dark:text-gray-400' }, [
                platformLabel(group.platform),
                platformLabel(group.platform) ? ' · ' : '',
                `${t('channelStatus.capacityPools.total')} ${formatInteger(group.total_accounts)}`,
                `, ${t('channelStatus.capacityPools.active')} ${formatInteger(group.active_accounts)}`,
                `, ${t('channelStatus.capacityPools.schedulable')} ${formatInteger(group.schedulable_accounts)}`,
              ]),
            ]),
            h('span', {
              class: ['shrink-0 rounded-md px-2 py-1 text-xs font-semibold', chipClass(groupStatusTone(group.status))],
            }, groupStatusLabel(group.status)),
          ]),
          h('div', { class: 'mt-2 flex flex-wrap gap-1.5' }, [
            group.rate_limited_accounts > 0
              ? h('span', { class: ['rounded px-2 py-0.5 text-xs', chipClass('warning')] }, `${t('channelStatus.capacityPools.rateLimited')} ${formatInteger(group.rate_limited_accounts)}`)
              : null,
            group.error_accounts > 0
              ? h('span', { class: ['rounded px-2 py-0.5 text-xs', chipClass('danger')] }, `${t('channelStatus.capacityPools.error')} ${formatInteger(group.error_accounts)}`)
              : null,
            group.disabled_accounts > 0
              ? h('span', { class: ['rounded px-2 py-0.5 text-xs', chipClass('neutral')] }, `${t('channelStatus.capacityPools.disabled')} ${formatInteger(group.disabled_accounts)}`)
              : null,
          ]),
          unavailableReasonEntries(group.unavailable_reasons).length > 0
            ? h('div', { class: 'mt-2 flex flex-wrap items-center gap-1.5' }, [
              h('span', { class: 'text-xs font-medium text-gray-500 dark:text-gray-400' }, t('channelStatus.capacityPools.unavailableReason')),
              ...unavailableReasonEntries(group.unavailable_reasons).map((reason) => (
                h('span', { key: reason.key, class: ['rounded px-2 py-0.5 text-xs', chipClass(reason.tone)] }, `${reason.label} ${formatInteger(reason.count)}`)
              )),
            ])
            : null,
          h('div', { class: 'mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2' }, windowSummaries(group).map((item) => (
            h('div', { key: item.key, class: 'rounded-md bg-white/65 p-2 dark:bg-dark-900/35' }, [
              h('div', { class: 'flex items-center justify-between gap-2 text-xs font-semibold text-gray-800 dark:text-gray-100' }, [
                h('span', windowLabel(item.key)),
                h('span', formatPercent(item.data.used_percent)),
              ]),
              h('div', { class: 'mt-1 h-1.5 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700' }, [
                h('div', {
                  class: 'h-full rounded-full bg-emerald-500',
                  style: { width: `${Math.min(100, Math.max(0, item.data.used_percent))}%` },
                }),
              ]),
              h('div', { class: 'mt-2 grid grid-cols-2 gap-2 text-xs text-gray-500 dark:text-gray-400' }, [
                h('div', [
                  h('p', t('channelStatus.capacityPools.schedulableSnapshot')),
                  h('p', { class: 'font-medium text-gray-700 dark:text-gray-200' }, `${formatInteger(item.data.schedulable_snapshot_accounts)}/${formatInteger(item.data.snapshot_accounts)}`),
                ]),
                h('div', [
              componentProps.pool.percent_only_quota || group.percent_only_quota
                ? h('p', t('channelStatus.capacityPools.percentOnly'))
                : h('p', t('channelStatus.capacityPools.schedulableRemaining')),
              componentProps.pool.percent_only_quota || group.percent_only_quota
                ? h('p', { class: 'font-medium text-gray-700 dark:text-gray-200' }, formatPercent(item.data.used_percent))
                : h('p', { class: 'font-medium text-gray-700 dark:text-gray-200' }, formatRemainingUnits(item.data.remaining_units)),
            ]),
          ]),
        ])
          ))),
        ])
      ))),
    ])
  },
})

const SectionFallback = defineComponent({
  name: 'SectionFallback',
  props: {
    pool: { type: Object as PropType<UserAccountCapacityPool>, required: true },
  },
  setup(componentProps) {
    return () => h('div', { class: 'mt-4 space-y-2' }, [
      ...(componentProps.pool.sections ?? []).map((section) => h('div', {
        key: `${componentProps.pool.key}-${section.platform}-${section.type}`,
        class: 'rounded-lg border border-gray-100 bg-white/70 p-3 dark:border-dark-700/70 dark:bg-dark-900/30',
      }, [
        h('div', { class: 'flex items-center justify-between gap-3' }, [
          h('div', [
            h('div', { class: 'flex flex-wrap items-center gap-1.5' }, [
              h('p', { class: 'text-sm font-semibold text-gray-900 dark:text-white' }, `${platformLabel(section.platform)} / ${typeLabel(section.type)}`),
              componentProps.pool.key === 'shared' && positiveCount(section.own_contributed_accounts) > 0
                ? h('span', { class: ['rounded px-2 py-0.5 text-xs font-semibold', chipClass('success')] }, `${t('channelStatus.capacityPools.ownContributed')} ${formatInteger(section.own_contributed_accounts ?? 0)}`)
                : null,
            ]),
            h('p', { class: 'mt-0.5 text-xs text-gray-500 dark:text-gray-400' }, `${section.schedulable_accounts}/${section.total_accounts} ${t('channelStatus.capacityPools.schedulable')}`),
          ]),
          h('div', { class: 'flex flex-wrap justify-end gap-1.5' }, windowBadges(section).map((item) => (
            h('span', { key: item.key, class: 'rounded-full bg-gray-100 px-2 py-1 text-xs font-semibold text-gray-700 dark:bg-dark-700 dark:text-gray-200' }, `${windowLabel(item.key)} ${Math.round(item.data.used_percent)}%`)
          ))),
        ]),
        unavailableReasonEntries(section.unavailable_reasons).length > 0
          ? h('div', { class: 'mt-2 flex flex-wrap items-center gap-1.5' }, [
            h('span', { class: 'text-xs font-medium text-gray-500 dark:text-gray-400' }, t('channelStatus.capacityPools.unavailableReason')),
            ...unavailableReasonEntries(section.unavailable_reasons).map((reason) => (
              h('span', { key: reason.key, class: ['rounded px-2 py-0.5 text-xs', chipClass(reason.tone)] }, `${reason.label} ${formatInteger(reason.count)}`)
            )),
          ])
          : null,
      ])),
      componentProps.pool.sections.length === 0
        ? h('p', { class: 'rounded-lg border border-dashed border-gray-200 p-4 text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400' }, t('channelStatus.capacityPools.empty'))
        : null,
      componentProps.pool.sections.length > 0 && componentProps.pool.configured_quota > 0 && !componentProps.pool.percent_only_quota
        ? h('div', { class: 'grid grid-cols-2 gap-2 pt-1' }, [
          h('div', { class: 'rounded-lg bg-gray-50 p-3 dark:bg-dark-900/40' }, [
            h('p', { class: 'text-xs text-gray-500 dark:text-gray-400' }, t('channelStatus.capacityPools.configuredQuota')),
            h('p', { class: 'mt-1 text-lg font-bold text-gray-900 dark:text-white' }, formatQuota(componentProps.pool.configured_quota)),
          ]),
          h('div', { class: 'rounded-lg bg-gray-50 p-3 dark:bg-dark-900/40' }, [
            h('p', { class: 'text-xs text-gray-500 dark:text-gray-400' }, t('channelStatus.capacityPools.remainingQuota')),
            h('p', { class: 'mt-1 text-lg font-bold text-gray-900 dark:text-white' }, formatQuota(componentProps.pool.remaining_quota)),
          ]),
        ])
        : null,
    ])
  },
})
</script>
