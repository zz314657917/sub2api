<template>
  <section
    class="grid grid-cols-1 items-start gap-3"
    :class="{ 'xl:grid-cols-2': poolKeys.length === 0 }"
    data-testid="account-capacity-pools"
  >
    <article
      v-for="pool in orderedPools"
      :key="pool.key"
      class="overflow-hidden rounded-lg border border-gray-200/80 bg-white/85 shadow-sm dark:border-dark-700/70 dark:bg-dark-800/70"
    >
      <div :class="isEmptyDashboardPool(pool) ? 'p-3 sm:p-4' : 'p-4 sm:p-5'">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="flex min-w-0 items-start gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg"
              :class="poolIconClass(pool)"
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
          v-if="isEmptyDashboardPool(pool)"
          class="mt-3 rounded-lg border border-dashed border-gray-200 px-3 py-2 text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400"
        >
          {{ t('channelStatus.capacityPools.empty') }}
        </p>

        <template v-else>
          <div
            v-if="hasPrimaryPoolData(pool)"
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

          <template v-if="hasPrimaryPoolData(pool)">
            <SharedGroupGrid v-if="sharedGroups(pool).length > 0" :pool="pool" />
            <SectionFallback v-else :pool="pool" />
          </template>

          <ExternalReferencePanel
            v-if="pool.key === 'shared' && externalReferencePool"
            :pool="externalReferencePool"
            :separated="hasPrimaryPoolData(pool)"
          />
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

type CapacityPoolKey = 'mine' | 'shared'

const props = withDefaults(defineProps<{
  pools: UserAccountCapacityPools | null
  loading?: boolean
  poolKeys?: CapacityPoolKey[]
}>(), {
  loading: false,
  poolKeys: () => [],
})

const { t } = useI18n()

const orderedPools = computed<UserAccountCapacityPool[]>(() => {
  if (!props.pools) {
    return []
  }
  const requestedKeys: CapacityPoolKey[] = props.poolKeys.length ? props.poolKeys : ['mine', 'shared']
  const pools: UserAccountCapacityPool[] = requestedKeys.map((key) => props.pools![key])
  return pools.filter(hasVisibleDashboardPool)
})

const externalReferencePool = computed<UserAccountCapacityPool | null>(() => {
  const pool = props.pools?.external
  if (!hasVisiblePoolAccounts(pool)) {
    return null
  }
  const groups = sharedGroups(pool)
  if (groups.length === 0) {
    return null
  }
  return {
    ...pool,
    groups,
    sections: [],
    total_accounts: sumGroupCount(groups, 'total_accounts'),
    active_accounts: sumGroupCount(groups, 'active_accounts'),
    schedulable_accounts: sumGroupCount(groups, 'schedulable_accounts'),
    rate_limited_accounts: sumGroupCount(groups, 'rate_limited_accounts'),
    error_accounts: sumGroupCount(groups, 'error_accounts'),
    disabled_accounts: sumGroupCount(groups, 'disabled_accounts'),
    abnormal_accounts: sumGroupCount(groups, 'abnormal_accounts'),
  }
})

function poolTitle(pool: UserAccountCapacityPool): string {
  if (pool.key === 'mine') return t('channelStatus.capacityPools.mineTitle')
  if (pool.key === 'shared') return t('channelStatus.capacityPools.sharedTitle')
  if (isExternalPool(pool)) return t('channelStatus.capacityPools.externalTitle')
  return pool.title
}

function poolDescription(pool: UserAccountCapacityPool): string {
  if (pool.key === 'mine') return t('channelStatus.capacityPools.mineDescription')
  if (pool.key === 'shared') return t('channelStatus.capacityPools.sharedDescription')
  if (isExternalPool(pool)) return t('channelStatus.capacityPools.externalDescription')
  return pool.title
}

function poolIconClass(pool: UserAccountCapacityPool): string {
  if (pool.key === 'mine') {
    return 'bg-[#f3e7df] text-[#a9583e] dark:bg-[#cc785c]/15 dark:text-[#f0b89e]'
  }
  if (isExternalPool(pool)) {
    return 'bg-[#fffaf5] text-[#8e8b82] ring-1 ring-[#d8cec2] dark:bg-[#8e8b82]/15 dark:text-[#d8cec2] dark:ring-[#8e8b82]/40'
  }
  return 'bg-[#f3e7df] text-[#a9583e] dark:bg-[#cc785c]/15 dark:text-[#f0b89e]'
}

function hasVisibleDashboardPool(pool: UserAccountCapacityPool | null | undefined): pool is UserAccountCapacityPool {
  if (!pool) return false
  if (hasPoolAccounts(pool)) {
    return true
  }
  return pool.key === 'shared' && externalReferencePool.value !== null
}

function sharedGroups(pool: UserAccountCapacityPool): UserAccountCapacityPoolGroup[] {
  const groups = (pool.groups ?? []).filter(hasGroupAccounts)
  if (!isExternalPool(pool)) {
    return groups
  }
  return groups
    .filter(isVisibleExternalReferenceGroup)
    .sort((left, right) => externalReferenceGroupRank(left.group_name) - externalReferenceGroupRank(right.group_name))
}

function visibleSections(pool: UserAccountCapacityPool): UserAccountCapacityPoolSection[] {
  return (pool.sections ?? []).filter(hasSectionAccounts)
}

function hasVisiblePoolAccounts(pool: UserAccountCapacityPool | null | undefined): pool is UserAccountCapacityPool {
  if (!pool) return false
  return hasPoolAccounts(pool)
}

function hasPoolAccounts(pool: UserAccountCapacityPool): boolean {
  return positiveCount(pool.total_accounts) > 0
    || visibleSections(pool).length > 0
    || sharedGroups(pool).length > 0
}

function hasGroupAccounts(group: UserAccountCapacityPoolGroup): boolean {
  return positiveCount(group.total_accounts) > 0
}

type GroupCountKey =
  | 'total_accounts'
  | 'active_accounts'
  | 'schedulable_accounts'
  | 'rate_limited_accounts'
  | 'error_accounts'
  | 'disabled_accounts'
  | 'abnormal_accounts'

function sumGroupCount(groups: UserAccountCapacityPoolGroup[], key: GroupCountKey): number {
  return groups.reduce((total, group) => total + positiveCount(group[key] ?? 0), 0)
}

function isVisibleExternalReferenceGroup(group: UserAccountCapacityPoolGroup): boolean {
  return externalReferenceGroupRank(group.group_name) < 2
}

function externalReferenceGroupRank(groupName: string): number {
  const normalized = groupName.trim().replace(/\s+/g, '').toUpperCase()
  if (normalized.startsWith('FREE') && normalized.includes('共享号池')) return 0
  if (normalized.startsWith('PLUS') && normalized.includes('共享号池')) return 1
  return 99
}

function hasSectionAccounts(section: UserAccountCapacityPoolSection): boolean {
  return positiveCount(section.total_accounts) > 0
}

function hasPrimaryPoolData(pool: UserAccountCapacityPool): boolean {
  return hasPoolAccounts(pool)
}

function isEmptyDashboardPool(pool: UserAccountCapacityPool): boolean {
  return !hasPrimaryPoolData(pool) && !(pool.key === 'shared' && externalReferencePool.value !== null)
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
    return '0%'
  }
  return `${Math.round(value)}%`
}

function clampPercent(value: number): number {
  if (!Number.isFinite(value)) {
    return 0
  }
  return Math.min(100, Math.max(0, value))
}

function remainingPercent(value: number): number {
  return 100 - clampPercent(value)
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
  if (status === 'healthy') return 'border-[#d8cec2]/80 bg-[#fffaf5]/70 dark:border-[#cc785c]/30 dark:bg-[#cc785c]/5'
  if (status === 'unavailable') return 'border-rose-200/80 bg-rose-50/35 dark:border-rose-500/30 dark:bg-rose-500/5'
  return 'border-amber-200/90 bg-amber-50/40 dark:border-amber-500/30 dark:bg-amber-500/5'
}

function chipClass(tone: 'success' | 'warning' | 'danger' | 'neutral'): string {
  switch (tone) {
    case 'success':
      return 'bg-[#f3e7df] text-[#a9583e] dark:bg-[#cc785c]/15 dark:text-[#f0b89e]'
    case 'warning':
      return 'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300'
    case 'danger':
      return 'bg-rose-100 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300'
    case 'neutral':
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
  }
}

function windowTone(usedPercent: number): 'success' | 'warning' | 'danger' | 'neutral' {
  const used = clampPercent(usedPercent)
  if (used >= 100) return 'danger'
  if (used >= 80) return 'warning'
  return 'neutral'
}

function windowProgressClass(usedPercent: number, percentOnlyQuota: boolean): string {
  const used = clampPercent(usedPercent)
  if (percentOnlyQuota) {
    if (used >= 100) return 'bg-rose-500'
    if (used >= 80) return 'bg-amber-500'
    return 'bg-[#cc785c]'
  }
  if (used >= 100) return 'bg-rose-500'
  if (used >= 80) return 'bg-amber-500'
  return 'bg-[#cc785c]'
}

type AccountStatusEntry = {
  key: string
  label: string
  count: number
  barClass: string
  dotClass: string
  showLegend: boolean
}

function accountStatusEntries(group: UserAccountCapacityPoolGroup): AccountStatusEntry[] {
  const schedulable = positiveCount(group.schedulable_accounts)
  const rateLimited = positiveCount(group.rate_limited_accounts)
  const error = positiveCount(group.error_accounts)
  const disabled = positiveCount(group.disabled_accounts)
  const total = positiveCount(group.total_accounts)
  const known = schedulable + rateLimited + error + disabled
  const other = Math.max(0, total - known)

  return [
    {
      key: 'schedulable',
      label: t('channelStatus.capacityPools.schedulable'),
      count: schedulable,
      barClass: 'bg-[#9c7b62]',
      dotClass: 'bg-[#9c7b62]',
      showLegend: false,
    },
    {
      key: 'rateLimited',
      label: t('channelStatus.capacityPools.rateLimited'),
      count: rateLimited,
      barClass: 'bg-amber-500',
      dotClass: 'bg-amber-500',
      showLegend: true,
    },
    {
      key: 'error',
      label: t('channelStatus.capacityPools.error'),
      count: error,
      barClass: 'bg-rose-500',
      dotClass: 'bg-rose-500',
      showLegend: true,
    },
    {
      key: 'disabled',
      label: t('channelStatus.capacityPools.disabled'),
      count: disabled,
      barClass: 'bg-slate-400',
      dotClass: 'bg-slate-400',
      showLegend: true,
    },
    {
      key: 'other',
      label: t('channelStatus.capacityPools.other'),
      count: other,
      barClass: 'bg-gray-300 dark:bg-dark-600',
      dotClass: 'bg-gray-300 dark:bg-dark-600',
      showLegend: true,
    },
  ].filter((entry) => entry.count > 0)
}

const quotaWindowOrder = ['1d', '7d_quota', '5h', '7d'] as const
const openAIPlanWindowOrder = ['5h', '7d'] as const
type CapacityWindowKey = (typeof quotaWindowOrder)[number]
const openAIPlanGroupNames = new Set(['openai plus', 'openai pro', 'openai team', 'openai free'])

function isOpenAIPlanDisplayGroup(group: UserAccountCapacityPoolGroup): boolean {
  return group.platform === 'openai'
    && group.key.startsWith('share-display:openai:')
    && openAIPlanGroupNames.has(group.group_name.trim().toLowerCase())
}

function windowSummaryKeys(group: UserAccountCapacityPoolGroup): readonly string[] {
  return isOpenAIPlanDisplayGroup(group) ? openAIPlanWindowOrder : quotaWindowOrder
}

function windowSummaries(group: UserAccountCapacityPoolGroup): Array<{ key: string; data: UserAccountCapacityWindowSummary }> {
  const windows = group.windows ?? {}
  return windowSummaryKeys(group)
    .map((key) => {
      const data = windows[key]
      return data ? { key, data } : null
    })
    .filter((item): item is { key: string; data: UserAccountCapacityWindowSummary } => item !== null)
}

function windowBadges(section: UserAccountCapacityPoolSection): Array<{ key: string; data: UserAccountCapacityWindowSnapshot }> {
  const windows = section.windows ?? {}
  return quotaWindowOrder
    .map<{ key: CapacityWindowKey; data: UserAccountCapacityWindowSnapshot } | null>((key) => {
      const data = windows[key]
      return data ? { key, data } : null
    })
    .filter((item): item is { key: CapacityWindowKey; data: UserAccountCapacityWindowSnapshot } => item !== null)
}

function windowLabel(key: string): string {
  switch (key) {
    case '1d':
      return t('channelStatus.capacityPools.quotaWindow', { window: '1d' })
    case '7d_quota':
      return t('channelStatus.capacityPools.quotaWindow', { window: '7d' })
    case '30d':
      return t('channelStatus.capacityPools.quotaWindow', { window: '30d' })
    case '5h':
      return t('channelStatus.capacityPools.window', { window: '5h' })
    case '7d':
      return t('channelStatus.capacityPools.window', { window: '7d' })
    default:
      return key
  }
}

function windowBadgeText(key: string, percent: number): string {
  return `${windowLabel(key)} ${formatPercent(percent)}`
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

function renderAccountStatusBar(group: UserAccountCapacityPoolGroup) {
  const entries = accountStatusEntries(group)
  const visibleEntries = entries.filter((entry) => entry.showLegend)
  if (entries.length === 0) {
    return null
  }

  return h('div', { class: 'mt-3' }, [
    h('div', { class: 'mb-2 flex items-center justify-between gap-3' }, [
      h('span', { class: 'text-xs font-semibold text-gray-900 dark:text-white' }, t('channelStatus.capacityPools.accountStatus')),
      h('span', { class: 'text-xs font-bold text-gray-900 dark:text-white' }, `${t('channelStatus.capacityPools.total')} ${formatInteger(group.total_accounts)}`),
    ]),
    h('div', { class: 'flex h-2.5 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700' }, entries.map((entry) => h('div', {
      key: entry.key,
      class: ['h-full', entry.barClass],
      style: {
        flexBasis: '0%',
        flexGrow: entry.count,
        minWidth: '4px',
      },
    }))),
    visibleEntries.length > 0
      ? h('div', { class: 'mt-2 grid grid-cols-2 gap-x-4 gap-y-2 sm:grid-cols-4' }, visibleEntries.map((entry) => h('div', {
        key: `${entry.key}-label`,
        class: 'flex min-w-0 items-center gap-2 text-xs text-gray-600 dark:text-gray-300',
      }, [
        h('span', { class: ['h-3 w-3 shrink-0 rounded', entry.dotClass] }),
        h('span', { class: 'min-w-0 truncate' }, `${entry.label} ${formatInteger(entry.count)}`),
      ])))
      : null,
  ])
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
            ? 'text-[#a9583e] dark:text-[#f0b89e]'
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
    return () => {
      const groups = sharedGroups(componentProps.pool)
      return h('div', {
        class: 'mt-4 rounded-lg border border-gray-200/80 p-3 dark:border-dark-700/70',
      }, [
        h('div', { class: 'mb-3 flex flex-wrap items-center justify-between gap-2' }, [
          h('div', [
            h('p', { class: 'text-sm font-semibold text-gray-900 dark:text-white' }, t('channelStatus.capacityPools.groupCapacity')),
            h('p', { class: 'mt-0.5 text-xs text-gray-500 dark:text-gray-400' }, t('channelStatus.capacityPools.groupCapacityHint')),
          ]),
          isExternalPool(componentProps.pool)
            ? null
            : h('div', { class: 'flex flex-wrap gap-1.5' }, [
              h('span', { class: ['rounded-md px-2 py-1 text-xs font-semibold', chipClass('success')] }, `${t('channelStatus.capacityPools.healthy')} ${groups.filter(g => g.status === 'healthy').length}`),
              h('span', { class: ['rounded-md px-2 py-1 text-xs font-semibold', chipClass('warning')] }, `${t('channelStatus.capacityPools.degraded')} ${groups.filter(g => g.status === 'degraded').length}`),
              h('span', { class: ['rounded-md px-2 py-1 text-xs font-semibold', chipClass('danger')] }, `${t('channelStatus.capacityPools.unavailable')} ${groups.filter(g => g.status === 'unavailable').length}`),
            ]),
        ]),
        h('div', { class: groupGridClass(componentProps.pool) }, groups.map((group) => {
          const reasons = unavailableReasonEntries(group.unavailable_reasons)
          return h('div', {
            key: group.key,
            class: ['rounded-lg border p-3 sm:p-4', groupBorderClass(group.status)],
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
            renderAccountStatusBar(group),
            reasons.length > 0
              ? h('div', { class: 'mt-2 flex flex-wrap items-center gap-1.5' }, [
                h('span', { class: 'text-xs font-medium text-gray-500 dark:text-gray-400' }, t('channelStatus.capacityPools.unavailableReason')),
                ...reasons.map((reason) => (
                  h('span', { key: reason.key, class: ['rounded px-2 py-0.5 text-xs', chipClass(reason.tone)] }, `${reason.label} ${formatInteger(reason.count)}`)
                )),
              ])
              : null,
            h('div', { class: 'mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2' }, windowSummaries(group).map((item) => {
              const percentOnlyQuota = Boolean(componentProps.pool.percent_only_quota || group.percent_only_quota)
              const displayPercent = percentOnlyQuota ? remainingPercent(item.data.used_percent) : clampPercent(item.data.used_percent)
              return h('div', { key: item.key, class: 'rounded-md bg-white/65 p-2 dark:bg-dark-900/35' }, [
                h('div', { class: 'flex items-center justify-between gap-2 text-xs font-semibold text-gray-800 dark:text-gray-100' }, [
                  h('span', windowLabel(item.key)),
                  h('span', formatPercent(displayPercent)),
                ]),
                h('div', { class: 'mt-1 h-1.5 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700' }, [
                  h('div', {
                    class: ['h-full rounded-full', windowProgressClass(item.data.used_percent, percentOnlyQuota)],
                    style: { width: `${displayPercent}%` },
                  }),
                ]),
                h('div', { class: 'mt-2 grid grid-cols-2 gap-2 text-xs text-gray-500 dark:text-gray-400' }, [
                  h('div', [
                    h('p', t('channelStatus.capacityPools.schedulableSnapshot')),
                    h('p', { class: 'font-medium text-gray-700 dark:text-gray-200' }, `${formatInteger(item.data.schedulable_snapshot_accounts)}/${formatInteger(item.data.snapshot_accounts)}`),
                  ]),
                  h('div', [
                    h('p', t('channelStatus.capacityPools.schedulableRemaining')),
                    percentOnlyQuota
                      ? h('p', { class: 'font-medium text-gray-700 dark:text-gray-200' }, formatPercent(displayPercent))
                      : h('p', { class: 'font-medium text-gray-700 dark:text-gray-200' }, formatRemainingUnits(item.data.remaining_units)),
                  ]),
                ]),
              ])
            })),
          ])
        })),
      ])
    }
  },
})

function isExternalPool(pool: UserAccountCapacityPool): boolean {
  return pool.key === 'public_shared_capacity_reference' || pool.key === 'external_ai_pixel'
}

function groupGridClass(pool: UserAccountCapacityPool): string {
  return isExternalPool(pool)
    ? 'grid grid-cols-1 gap-3 md:grid-cols-2'
    : 'grid grid-cols-1 gap-2 2xl:grid-cols-2'
}

const ExternalReferencePanel = defineComponent({
  name: 'ExternalReferencePanel',
  props: {
    pool: { type: Object as PropType<UserAccountCapacityPool>, required: true },
    separated: { type: Boolean, default: false },
  },
  setup(componentProps) {
    return () => h('div', {
      class: [
        componentProps.separated ? 'mt-4 border-t border-gray-200/80 pt-4 dark:border-dark-700/70' : 'mt-4',
      ],
    }, [
      h('div', { class: 'flex flex-wrap items-start justify-between gap-3' }, [
        h('div', { class: 'min-w-0' }, [
          h('p', { class: 'text-sm font-bold text-gray-900 dark:text-white' }, poolTitle(componentProps.pool)),
          h('p', { class: 'mt-0.5 text-xs leading-5 text-gray-500 dark:text-gray-400' }, poolDescription(componentProps.pool)),
        ]),
      ]),
      sharedGroups(componentProps.pool).length > 0
        ? h(SharedGroupGrid, { pool: componentProps.pool })
        : h(SectionFallback, { pool: componentProps.pool }),
    ])
  },
})

const SectionFallback = defineComponent({
  name: 'SectionFallback',
  props: {
    pool: { type: Object as PropType<UserAccountCapacityPool>, required: true },
  },
  setup(componentProps) {
    return () => {
      const sections = visibleSections(componentProps.pool)
      return h('div', { class: 'mt-4 space-y-2' }, [
      ...sections.map((section) => h('div', {
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
          h('div', { class: 'flex flex-wrap justify-end gap-1.5' }, windowBadges(section).map((item) => {
            const percentOnlyQuota = Boolean(componentProps.pool.percent_only_quota || section.percent_only_quota)
            const displayPercent = percentOnlyQuota ? remainingPercent(item.data.used_percent) : clampPercent(item.data.used_percent)
            return h('span', { key: item.key, class: ['rounded-full px-2 py-1 text-xs font-semibold', chipClass(windowTone(item.data.used_percent))] }, windowBadgeText(item.key, displayPercent))
          })),
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
      sections.length === 0
        ? h('p', { class: 'rounded-lg border border-dashed border-gray-200 p-4 text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400' }, t('channelStatus.capacityPools.empty'))
        : null,
      sections.length > 0 && componentProps.pool.configured_quota > 0 && !componentProps.pool.percent_only_quota
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
    }
  },
})
</script>
