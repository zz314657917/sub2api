<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col gap-3">
          <div class="flex flex-col justify-between gap-3 lg:flex-row lg:items-center">
            <div class="flex flex-1 flex-wrap items-center gap-3">
              <SearchInput
                v-model="filterSearch"
                :placeholder="t('keys.searchPlaceholder')"
                class="w-full sm:w-64"
                @search="onFilterChange"
              />
              <Select
                :model-value="filterGroupId"
                class="w-40"
                :options="groupFilterOptions"
                @update:model-value="onGroupFilterChange"
              />
              <Select
                :model-value="filterStatus"
                class="w-40"
                :options="statusFilterOptions"
                @update:model-value="onStatusFilterChange"
              />
            </div>

            <div class="flex w-full flex-shrink-0 items-center justify-end gap-3 lg:w-auto">
              <button
                @click="loadApiKeys"
                :disabled="loading"
                class="btn btn-secondary"
                :title="t('common.refresh')"
              >
                <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
              </button>
              <button @click="openCreateModal" class="btn btn-primary" data-tour="keys-create-btn">
                <Icon name="plus" size="md" class="mr-2" />
                {{ t('keys.createKey') }}
              </button>
            </div>
          </div>
          <EndpointPopover
            v-if="publicSettings?.api_base_url || (publicSettings?.custom_endpoints?.length ?? 0) > 0"
            :api-base-url="publicSettings?.api_base_url || ''"
            :custom-endpoints="publicSettings?.custom_endpoints || []"
          />
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="apiKeys"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="created_at"
          default-sort-order="desc"
          @sort="handleSort"
        >
          <template #cell-key="{ value, row }">
            <div class="flex items-center gap-2">
              <code class="code text-xs">
                {{ maskApiKey(value) }}
              </code>
              <button
                @click="copyToClipboard(value, row.id)"
                class="rounded-lg p-1 transition-colors hover:bg-gray-100 dark:hover:bg-dark-700"
                :class="
                  copiedKeyId === row.id
                    ? 'text-green-500'
                    : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
                "
                :title="copiedKeyId === row.id ? t('keys.copied') : t('keys.copyToClipboard')"
              >
                <Icon
                  v-if="copiedKeyId === row.id"
                  name="check"
                  size="sm"
                  :stroke-width="2"
                />
                <Icon v-else name="clipboard" size="sm" />
              </button>
            </div>
          </template>

          <template #cell-name="{ value, row }">
            <div class="flex flex-col gap-1">
              <div class="flex items-center gap-1.5">
                <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
                <span
                  v-if="row.is_default"
                  class="inline-flex items-center rounded-md bg-blue-50 px-1.5 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-300"
                  :title="t('keys.defaultKeyHint')"
                >
                  {{ t('keys.defaultKeyBadge') }}
                </span>
                <span
                  v-if="apiKeySupportsUnifiedAccess(row)"
                  class="inline-flex items-center gap-1 rounded-md bg-emerald-50 px-1.5 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300"
                  :title="t('keys.unifiedKeyHint')"
                >
                  <Icon name="sparkles" size="xs" />
                  {{ t('keys.unifiedKeyBadge') }}
                </span>
                <Icon
                  v-if="row.ip_whitelist?.length > 0 || row.ip_blacklist?.length > 0"
                  name="shield"
                  size="sm"
                  class="text-blue-500"
                  :title="t('keys.ipRestrictionEnabled')"
                />
              </div>
              <span v-if="row.is_default" class="text-xs text-gray-500 dark:text-dark-400">
                {{ t('keys.defaultKeyStudioNote') }}
              </span>
              <div
                v-if="apiKeyCapabilityItems(row).length > 0"
                class="flex flex-wrap gap-1"
              >
                <span
                  v-for="item in apiKeyCapabilityItems(row)"
                  :key="item.key"
                  class="inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[11px] font-medium"
                  :class="item.enabled
                    ? 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-200'
                    : 'bg-amber-50 text-amber-700 dark:bg-amber-900/20 dark:text-amber-300'"
                >
                  <Icon :name="item.icon" size="xs" />
                  {{ item.label }}
                </span>
              </div>
            </div>
          </template>

          <template #cell-group="{ row }">
            <div class="group/dropdown relative">
              <div class="flex flex-col items-start gap-1.5">
                <button
                  v-if="!isSmartRoutingKey(row)"
                  :ref="(el) => setGroupButtonRef(row.id, el)"
                  @click="openGroupSelector(row)"
                  class="-mx-2 -my-1 flex cursor-pointer items-center gap-2 rounded-lg px-2 py-1 transition-all duration-200 hover:bg-gray-100 dark:hover:bg-dark-700"
                  :title="t('keys.clickToChangeGroup')"
                >
                  <GroupBadge
                    v-if="row.group"
                    :name="row.group.name"
                    :platform="row.group.platform"
                    :subscription-type="row.group.subscription_type"
                    :rate-multiplier="row.group.rate_multiplier"
                    :user-rate-multiplier="userGroupRates[row.group.id]"
                    :peak-rate-enabled="row.group.peak_rate_enabled"
                    :peak-start="row.group.peak_start"
                    :peak-end="row.group.peak_end"
                    :peak-rate-multiplier="row.group.peak_rate_multiplier"
                  />
                  <span v-else class="text-sm text-gray-400 dark:text-dark-500">{{
                    t('keys.noGroup')
                  }}</span>
                  <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('keys.selectGroup') }}</span>
                  <svg
                    class="h-3.5 w-3.5 text-gray-400 opacity-60 transition-opacity group-hover/dropdown:opacity-100"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    stroke-width="2"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M8.25 15L12 18.75 15.75 15m-7.5-6L12 5.25 15.75 9"
                    />
                  </svg>
                </button>
                <button
                  v-else
                  type="button"
                  class="-mx-2 -my-1 inline-flex items-center gap-1.5 rounded-lg px-2 py-1 text-xs font-medium text-blue-700 transition-colors hover:bg-blue-50 dark:text-blue-300 dark:hover:bg-blue-900/30"
                  :title="t('keys.clickToEditSmartRouting')"
                  @click="editKey(row)"
                >
                  <span
                    class="inline-flex items-center gap-1.5 rounded-md bg-blue-50 px-2 py-0.5 dark:bg-blue-900/30"
                  >
                    <Icon name="sparkles" size="xs" />
                    {{ t('keys.multiGroupRouteCount', { count: row.multi_group_routes.length }) }}
                  </span>
                </button>
                <span
                  v-if="row.account_pool_strategy && row.account_pool_strategy !== 'shared_only'"
                  class="inline-flex items-center rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300"
                >
                  {{ accountPoolStrategyLabel(row.account_pool_strategy) }}
                </span>
              </div>
            </div>
          </template>

          <template #cell-usage="{ row }">
            <div class="text-sm">
              <div class="flex items-center gap-1.5">
                <span class="text-gray-500 dark:text-gray-400">{{ t('keys.today') }}:</span>
                <span class="font-medium text-gray-900 dark:text-white">
                  {{ formatCreditAmount(usageStats[row.id]?.today_actual_cost ?? 0, { minimumFractionDigits: 4, maximumFractionDigits: 4 }) }}
                </span>
              </div>
              <div class="mt-0.5 flex items-center gap-1.5">
                <span class="text-gray-500 dark:text-gray-400">{{ t('keys.total') }}:</span>
                <span class="font-medium text-gray-900 dark:text-white">
                  {{ formatCreditAmount(usageStats[row.id]?.total_actual_cost ?? 0, { minimumFractionDigits: 4, maximumFractionDigits: 4 }) }}
                </span>
              </div>
              <!-- Quota progress (if quota is set) -->
              <div v-if="row.quota > 0" class="mt-1.5">
                <div class="flex items-center gap-1.5">
                  <span class="text-gray-500 dark:text-gray-400">{{ t('keys.quota') }}:</span>
                  <span :class="[
                    'font-medium',
                    row.quota_used >= row.quota ? 'text-red-500' :
                    row.quota_used >= row.quota * 0.8 ? 'text-yellow-500' :
                    'text-gray-900 dark:text-white'
                  ]">
                    {{ formatCreditAmount(row.quota_used || 0) }} / {{ formatCreditAmount(row.quota || 0) }}
                  </span>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                  <div
                    :class="[
                      'h-full rounded-full transition-all',
                      row.quota_used >= row.quota ? 'bg-red-500' :
                      row.quota_used >= row.quota * 0.8 ? 'bg-yellow-500' :
                      'bg-primary-500'
                    ]"
                    :style="{ width: Math.min((row.quota_used / row.quota) * 100, 100) + '%' }"
                  />
                </div>
              </div>
            </div>
          </template>

          <template #cell-rate_limit="{ row }">
            <div v-if="row.rate_limit_5h > 0 || row.rate_limit_1d > 0 || row.rate_limit_7d > 0" class="space-y-1.5 min-w-[140px]">
              <!-- 5h window -->
              <div v-if="row.rate_limit_5h > 0">
                <div class="flex items-center justify-between text-xs">
                  <span class="text-gray-500 dark:text-gray-400">5h</span>
                  <span :class="[
                    'font-medium tabular-nums',
                    row.usage_5h >= row.rate_limit_5h ? 'text-red-500' :
                    row.usage_5h >= row.rate_limit_5h * 0.8 ? 'text-yellow-500' :
                    'text-gray-700 dark:text-gray-300'
                  ]">
                    {{ formatCreditAmount(row.usage_5h || 0) }}/{{ formatCreditAmount(row.rate_limit_5h || 0) }}
                  </span>
                </div>
                <div class="h-1 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                  <div
                    :class="[
                      'h-full rounded-full transition-all',
                      row.usage_5h >= row.rate_limit_5h ? 'bg-red-500' :
                      row.usage_5h >= row.rate_limit_5h * 0.8 ? 'bg-yellow-500' :
                      'bg-emerald-500'
                    ]"
                    :style="{ width: Math.min((row.usage_5h / row.rate_limit_5h) * 100, 100) + '%' }"
                  />
                </div>
                <div v-if="row.reset_5h_at && formatResetTime(row.reset_5h_at)" class="text-[10px] text-gray-400 dark:text-gray-500 tabular-nums">
                  ⟳ {{ formatResetTime(row.reset_5h_at) }}
                </div>
              </div>
              <!-- 1d window -->
              <div v-if="row.rate_limit_1d > 0">
                <div class="flex items-center justify-between text-xs">
                  <span class="text-gray-500 dark:text-gray-400">1d</span>
                  <span :class="[
                    'font-medium tabular-nums',
                    row.usage_1d >= row.rate_limit_1d ? 'text-red-500' :
                    row.usage_1d >= row.rate_limit_1d * 0.8 ? 'text-yellow-500' :
                    'text-gray-700 dark:text-gray-300'
                  ]">
                    {{ formatCreditAmount(row.usage_1d || 0) }}/{{ formatCreditAmount(row.rate_limit_1d || 0) }}
                  </span>
                </div>
                <div class="h-1 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                  <div
                    :class="[
                      'h-full rounded-full transition-all',
                      row.usage_1d >= row.rate_limit_1d ? 'bg-red-500' :
                      row.usage_1d >= row.rate_limit_1d * 0.8 ? 'bg-yellow-500' :
                      'bg-emerald-500'
                    ]"
                    :style="{ width: Math.min((row.usage_1d / row.rate_limit_1d) * 100, 100) + '%' }"
                  />
                </div>
                <div v-if="row.reset_1d_at && formatResetTime(row.reset_1d_at)" class="text-[10px] text-gray-400 dark:text-gray-500 tabular-nums">
                  ⟳ {{ formatResetTime(row.reset_1d_at) }}
                </div>
              </div>
              <!-- 7d window -->
              <div v-if="row.rate_limit_7d > 0">
                <div class="flex items-center justify-between text-xs">
                  <span class="text-gray-500 dark:text-gray-400">7d</span>
                  <span :class="[
                    'font-medium tabular-nums',
                    row.usage_7d >= row.rate_limit_7d ? 'text-red-500' :
                    row.usage_7d >= row.rate_limit_7d * 0.8 ? 'text-yellow-500' :
                    'text-gray-700 dark:text-gray-300'
                  ]">
                    {{ formatCreditAmount(row.usage_7d || 0) }}/{{ formatCreditAmount(row.rate_limit_7d || 0) }}
                  </span>
                </div>
                <div class="h-1 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                  <div
                    :class="[
                      'h-full rounded-full transition-all',
                      row.usage_7d >= row.rate_limit_7d ? 'bg-red-500' :
                      row.usage_7d >= row.rate_limit_7d * 0.8 ? 'bg-yellow-500' :
                      'bg-emerald-500'
                    ]"
                    :style="{ width: Math.min((row.usage_7d / row.rate_limit_7d) * 100, 100) + '%' }"
                  />
                </div>
                <div v-if="row.reset_7d_at && formatResetTime(row.reset_7d_at)" class="text-[10px] text-gray-400 dark:text-gray-500 tabular-nums">
                  ⟳ {{ formatResetTime(row.reset_7d_at) }}
                </div>
              </div>
              <!-- Reset button -->
              <button
                v-if="row.usage_5h > 0 || row.usage_1d > 0 || row.usage_7d > 0"
                @click.stop="confirmResetRateLimitFromTable(row)"
                class="mt-0.5 inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
                :title="t('keys.resetRateLimitUsage')"
              >
                <Icon name="refresh" size="xs" />
                {{ t('keys.resetUsage') }}
              </button>
            </div>
            <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
          </template>

          <template #cell-expires_at="{ value }">
            <span v-if="value" :class="[
              'text-sm',
              new Date(value) < new Date() ? 'text-red-500 dark:text-red-400' : 'text-gray-500 dark:text-dark-400'
            ]">
              {{ formatDateTime(value) }}
            </span>
            <span v-else class="text-sm text-gray-400 dark:text-dark-500">{{ t('keys.noExpiration') }}</span>
          </template>

          <template #cell-status="{ value }">
            <span :class="[
              'badge',
              value === 'active' ? 'badge-success' :
              value === 'quota_exhausted' ? 'badge-warning' :
              value === 'expired' ? 'badge-danger' :
              'badge-gray'
            ]">
              {{ t('keys.status.' + value) }}
            </span>
          </template>

          <template #cell-last_used_at="{ value }">
            <span v-if="value" class="text-sm text-gray-500 dark:text-dark-400">
              {{ formatDateTime(value) }}
            </span>
            <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <!-- Use Key Button -->
              <button
                @click="openUseKeyModal(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/20 dark:hover:text-green-400"
              >
                <Icon name="terminal" size="sm" />
                <span class="text-xs">{{ t('keys.useKey') }}</span>
              </button>
              <!-- Import to CC Switch Button -->
              <button
                v-if="!publicSettings?.hide_ccs_import_button"
                @click="importToCcswitch(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400"
              >
                <Icon name="upload" size="sm" />
                <span class="text-xs">{{ t('keys.importToCcSwitch') }}</span>
              </button>
              <!-- Import to Cockpit Tools Button -->
              <button
                @click="importToCockpitTools(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-cyan-50 hover:text-cyan-600 dark:hover:bg-cyan-900/20 dark:hover:text-cyan-400"
                :title="t('keys.importToCockpitToolsHint')"
              >
                <Icon name="upload" size="sm" />
                <span class="text-xs">{{ t('keys.importToCockpitTools') }}</span>
              </button>
              <!-- Toggle Status Button -->
              <button
                @click="toggleKeyStatus(row)"
                :class="[
                  'flex flex-col items-center gap-0.5 rounded-lg p-1.5 transition-colors',
                  row.status === 'active'
                    ? 'text-gray-500 hover:bg-yellow-50 hover:text-yellow-600 dark:hover:bg-yellow-900/20 dark:hover:text-yellow-400'
                    : 'text-gray-500 hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/20 dark:hover:text-green-400'
                ]"
              >
                <Icon v-if="row.status === 'active'" name="ban" size="sm" />
                <Icon v-else name="checkCircle" size="sm" />
                <span class="text-xs">{{ row.status === 'active' ? t('keys.disable') : t('keys.enable') }}</span>
              </button>
              <!-- Edit Button -->
              <button
                @click="editKey(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
              >
                <Icon name="edit" size="sm" />
                <span class="text-xs">{{ t('common.edit') }}</span>
              </button>
              <!-- Delete Button -->
              <button
                @click="confirmDelete(row)"
                :disabled="row.is_default"
                :title="row.is_default ? t('keys.defaultKeyDeleteDisabled') : t('common.delete')"
                :class="[
                  'flex flex-col items-center gap-0.5 rounded-lg p-1.5 transition-colors',
                  row.is_default
                    ? 'cursor-not-allowed text-gray-300 dark:text-dark-500'
                    : 'text-gray-500 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400'
                ]"
              >
                <Icon name="trash" size="sm" />
                <span class="text-xs">{{ t('common.delete') }}</span>
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('keys.noKeysYet')"
              :description="t('keys.createFirstKey')"
              :action-text="t('keys.createKey')"
              @action="openCreateModal"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Create/Edit Modal -->
    <BaseDialog
      :show="showCreateModal || showEditModal"
      :title="showEditModal ? t('keys.editKey') : t('keys.createKey')"
      width="wide"
      @close="closeModals"
    >
      <form id="key-form" @submit.prevent="handleSubmit" class="space-y-5">
        <section class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900/40">
          <div class="mb-4 flex items-start gap-3">
            <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-blue-50 text-blue-600 dark:bg-blue-500/15 dark:text-blue-300">
              <Icon name="key" size="md" />
            </div>
            <div>
              <h4 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('keys.basicInfo') }}
              </h4>
              <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">
                {{ t('keys.basicInfoHint') }}
              </p>
            </div>
          </div>

          <div class="space-y-4">
            <div>
              <label class="input-label">{{ t('keys.nameLabel') }}</label>
              <input
                v-model="formData.name"
                type="text"
                required
                class="input"
                :placeholder="t('keys.namePlaceholder')"
                data-tour="key-form-name"
              />
            </div>

            <div>
              <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
                <label class="input-label mb-0">{{ t('keys.routingPresetLabel') }}</label>
                <span class="text-xs text-gray-400 dark:text-gray-500">
                  {{
                    routingPreset === 'manual'
                      ? t('keys.routingPreset.manual.label')
                      : t('keys.routingPresetAutoApplied')
                  }}
                </span>
              </div>
              <div class="grid gap-3 md:grid-cols-2">
                <button
                  v-for="option in routingPresetOptions"
                  :key="option.value"
                  type="button"
                  :data-testid="`routing-preset-${option.value}`"
                  :class="[
                    'flex min-h-[5rem] items-start gap-3 rounded-lg border p-3 text-left transition-colors',
                    presetToneClasses(option)
                  ]"
                  @click="applyRoutingPreset(option.value)"
                >
                  <span
                    :class="[
                      'flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg',
                      presetIconClasses(option)
                    ]"
                  >
                    <Icon :name="option.icon" size="sm" />
                  </span>
                  <span class="min-w-0">
                    <span class="block text-sm font-semibold">{{ option.title }}</span>
                    <span class="mt-1 block text-xs leading-5 opacity-80">{{ option.description }}</span>
                  </span>
                </button>
              </div>
              <div class="mt-3 rounded-lg bg-blue-50 px-3 py-2 text-xs leading-5 text-blue-700 dark:bg-blue-500/10 dark:text-blue-300">
                {{ t('keys.routingPresetHint') }}
              </div>
            </div>

            <div>
              <label class="input-label">
                {{
                  formData.enable_multi_group_routing
                    ? t('keys.defaultGroupLabel')
                    : t('keys.groupLabel')
                }}
              </label>
              <Select
                :model-value="formData.group_id"
                :options="groupOptions"
                :placeholder="t('keys.selectGroup')"
                :searchable="true"
                :search-placeholder="t('keys.searchGroup')"
                data-tour="key-form-group"
                @update:model-value="handleDefaultGroupChanged"
              >
                <template #selected="{ option }">
                  <GroupBadge
                    v-if="option"
                    :name="(option as unknown as GroupOption).label"
                    :platform="(option as unknown as GroupOption).platform"
                    :subscription-type="(option as unknown as GroupOption).subscriptionType"
                    :rate-multiplier="(option as unknown as GroupOption).rate"
                    :user-rate-multiplier="(option as unknown as GroupOption).userRate"
                    :peak-rate-enabled="(option as unknown as GroupOption).peakRateEnabled"
                    :peak-start="(option as unknown as GroupOption).peakStart"
                    :peak-end="(option as unknown as GroupOption).peakEnd"
                    :peak-rate-multiplier="(option as unknown as GroupOption).peakRateMultiplier"
                  />
                  <span v-else class="text-gray-400">{{ t('keys.selectGroup') }}</span>
                </template>
                <template #option="{ option, selected }">
                  <GroupOptionItem
                    :name="(option as unknown as GroupOption).label"
                    :platform="(option as unknown as GroupOption).platform"
                    :subscription-type="(option as unknown as GroupOption).subscriptionType"
                    :rate-multiplier="(option as unknown as GroupOption).rate"
                    :user-rate-multiplier="(option as unknown as GroupOption).userRate"
                    :peak-rate-enabled="(option as unknown as GroupOption).peakRateEnabled"
                    :peak-start="(option as unknown as GroupOption).peakStart"
                    :peak-end="(option as unknown as GroupOption).peakEnd"
                    :peak-rate-multiplier="(option as unknown as GroupOption).peakRateMultiplier"
                    :description="(option as unknown as GroupOption).description"
                    :selected="selected"
                  />
                </template>
              </Select>
              <p
                v-if="formData.enable_multi_group_routing"
                class="mt-2 text-xs leading-5 text-gray-500 dark:text-gray-400"
              >
                {{ t('keys.defaultGroupHint') }}
              </p>
            </div>

            <div v-if="formData.enable_multi_group_routing" class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-800/50">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div class="text-sm text-gray-700 dark:text-gray-200">
                  <span class="font-medium">{{ t('keys.groupPrioritySelection') }}</span>
                  <span class="ml-2 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('keys.multiGroupRouteCount', { count: formData.multi_group_routes.length }) }}
                  </span>
                </div>
                <button
                  type="button"
                  class="btn btn-secondary text-sm"
                  @click="applyRoutingPreset('manual')"
                >
                  <Icon name="menu" size="sm" class="mr-2" />
                  {{ t('keys.manualRouteSelection') }}
                </button>
              </div>
              <div class="mt-3 flex flex-wrap gap-2">
                <span
                  v-for="(route, index) in formData.multi_group_routes"
                  :key="route.client_id"
                  class="inline-flex items-center rounded-md bg-white px-2 py-1 text-xs text-gray-600 shadow-sm dark:bg-dark-900 dark:text-gray-300"
                >
                  {{ index + 1 }}.
                  {{ groupOptions.find((group) => group.value === route.group_id)?.label || t('keys.noGroup') }}
                  <span v-if="route.weight > 1" class="ml-1 text-gray-400">x{{ route.weight }}</span>
                </span>
              </div>
            </div>
          </div>
        </section>

        <section class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900/40">
          <div class="mb-4 flex items-start gap-3">
            <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-emerald-50 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-300">
              <Icon name="creditCard" size="md" />
            </div>
            <div>
              <h4 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('keys.quotaSettings') }}
              </h4>
              <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">
                {{ t('keys.quotaSettingsHint') }}
              </p>
            </div>
          </div>

          <div class="space-y-4">
            <div>
              <div class="mb-2 flex items-center justify-between gap-3">
                <label class="input-label mb-0">{{ t('keys.quotaLimit') }}</label>
                <div class="inline-flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
                  <span>{{ t('keys.unlimitedQuota') }}</span>
                  <button
                    type="button"
                    :aria-pressed="!formData.enable_quota"
                    @click="formData.enable_quota = !formData.enable_quota"
                    :class="[
                      'relative inline-flex h-7 w-12 flex-shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                      formData.enable_quota ? 'bg-gray-300 dark:bg-dark-600' : 'bg-primary-600'
                    ]"
                  >
                    <span
                      :class="[
                        'pointer-events-none inline-block h-6 w-6 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                        formData.enable_quota ? 'translate-x-0' : 'translate-x-5'
                      ]"
                    />
                  </button>
                </div>
              </div>
              <div class="relative max-w-sm">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">{{ CREDIT_SYMBOL }}</span>
                <input
                  v-model.number="formData.quota"
                  type="number"
                  step="0.01"
                  min="0"
                  class="input pl-7"
                  :disabled="!formData.enable_quota"
                  :placeholder="t('keys.quotaAmountPlaceholder')"
                />
              </div>
              <p class="input-hint">{{ t('keys.quotaAmountHint') }}</p>
            </div>

            <div v-if="showEditModal && selectedKey && selectedKey.quota > 0">
              <label class="input-label">{{ t('keys.quotaUsed') }}</label>
              <div class="flex items-center gap-2">
                <div class="flex-1 rounded-lg bg-gray-100 px-3 py-2 dark:bg-dark-700">
                  <span class="font-medium text-gray-900 dark:text-white">
                    {{ formatCreditAmount(selectedKey.quota_used || 0, { minimumFractionDigits: 4, maximumFractionDigits: 4 }) }}
                  </span>
                  <span class="mx-2 text-gray-400">/</span>
                  <span class="text-gray-500 dark:text-gray-400">
                    {{ formatCreditAmount(selectedKey.quota || 0) }}
                  </span>
                </div>
                <button
                  type="button"
                  @click="confirmResetQuota"
                  class="btn btn-secondary text-sm"
                  :title="t('keys.resetQuotaUsed')"
                >
                  {{ t('keys.reset') }}
                </button>
              </div>
            </div>
          </div>
        </section>

        <details class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900/40" open>
          <summary class="flex cursor-pointer list-none items-center justify-between gap-3">
            <span class="flex items-start gap-3">
              <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                <Icon name="cog" size="md" />
              </span>
              <span>
                <span class="block text-sm font-semibold text-gray-900 dark:text-white">
                  {{ t('keys.advancedSettings') }}
                </span>
                <span class="mt-1 block text-xs leading-5 text-gray-500 dark:text-gray-400">
                  {{ t('keys.advancedSettingsHint') }}
                </span>
              </span>
            </span>
            <Icon name="chevronDown" size="sm" class="text-gray-400" />
          </summary>

          <div class="mt-4 space-y-5">
            <div>
              <label class="input-label">{{ t('keys.accountPoolStrategyLabel') }}</label>
              <Select
                v-model="formData.account_pool_strategy"
                :options="accountPoolStrategyOptions"
              />
              <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-gray-400">
                {{ t('keys.accountPoolStrategyHint') }}
              </p>
            </div>

            <div class="space-y-3 rounded-lg border border-gray-200 bg-gray-50/70 p-4 dark:border-dark-700 dark:bg-dark-800/40">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <label class="input-label mb-1">{{ t('keys.multiGroupRouting') }}</label>
                  <p class="text-xs leading-5 text-gray-500 dark:text-gray-400">
                    {{ t('keys.multiGroupRoutingHint') }}
                  </p>
                </div>
                <button
                  type="button"
                  @click="toggleMultiGroupRouting"
                  :class="[
                    'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                    formData.enable_multi_group_routing
                      ? 'bg-primary-600'
                      : 'bg-gray-200 dark:bg-dark-600'
                  ]"
                >
                  <span
                    :class="[
                      'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                      formData.enable_multi_group_routing ? 'translate-x-5' : 'translate-x-0'
                    ]"
                  />
                </button>
              </div>

              <div v-if="formData.enable_multi_group_routing" class="space-y-3">
                <VueDraggable
                  v-model="formData.multi_group_routes"
                  :animation="200"
                  handle=".route-drag-handle"
                  class="space-y-3"
                  @end="handleRouteOrderChanged"
                >
                  <div
                    v-for="(route, index) in formData.multi_group_routes"
                    :key="route.client_id"
                    class="rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-900/40"
                  >
                    <div class="mb-3 flex items-center justify-between gap-3">
                      <div class="flex items-center gap-2">
                        <button
                          type="button"
                          class="route-drag-handle inline-flex h-8 w-8 cursor-grab items-center justify-center rounded-lg text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 active:cursor-grabbing dark:text-gray-500 dark:hover:bg-dark-700 dark:hover:text-gray-300"
                          :title="t('keys.dragRoute')"
                        >
                          <Icon name="menu" size="sm" />
                        </button>
                        <span class="flex h-6 w-6 items-center justify-center rounded-full bg-primary-50 text-xs font-semibold text-primary-600 dark:bg-primary-900/30 dark:text-primary-300">
                          {{ index + 1 }}
                        </span>
                        <span class="text-sm font-medium text-gray-700 dark:text-gray-200">{{
                          t('keys.routeConfig')
                        }}</span>
                      </div>
                      <div class="flex items-center gap-2">
                        <label class="inline-flex h-9 items-center gap-2 rounded-lg border border-gray-200 px-3 text-sm text-gray-600 dark:border-dark-600 dark:text-gray-300">
                          <input
                            v-model="route.enabled"
                            type="checkbox"
                            class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                            @change="markRoutingManual"
                          />
                          <span>{{ t('keys.routeEnabled') }}</span>
                        </label>
                        <button
                          type="button"
                          class="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-gray-200 text-gray-500 transition-colors hover:border-red-200 hover:bg-red-50 hover:text-red-600 dark:border-dark-600 dark:text-gray-400 dark:hover:border-red-900/60 dark:hover:bg-red-900/20 dark:hover:text-red-300"
                          :title="t('keys.removeRoute')"
                          @click="removeMultiGroupRoute(index)"
                        >
                          <Icon name="trash" size="sm" />
                        </button>
                      </div>
                    </div>

                    <div class="grid gap-3 md:grid-cols-[minmax(0,2fr)_minmax(7rem,1fr)_minmax(7rem,1fr)_minmax(8rem,1fr)]">
                      <div>
                        <label class="input-label">{{ t('keys.groupLabel') }}</label>
                        <Select
                          :model-value="route.group_id"
                          :options="groupOptions"
                          :placeholder="t('keys.selectGroup')"
                          :searchable="true"
                          :search-placeholder="t('keys.searchGroup')"
                          @update:model-value="updateRouteGroup(route, $event)"
                        >
                          <template #selected="{ option }">
                            <GroupBadge
                              v-if="option"
                              :name="(option as unknown as GroupOption).label"
                              :platform="(option as unknown as GroupOption).platform"
                              :subscription-type="(option as unknown as GroupOption).subscriptionType"
                              :rate-multiplier="(option as unknown as GroupOption).rate"
                              :user-rate-multiplier="(option as unknown as GroupOption).userRate"
                              :peak-rate-enabled="(option as unknown as GroupOption).peakRateEnabled"
                              :peak-start="(option as unknown as GroupOption).peakStart"
                              :peak-end="(option as unknown as GroupOption).peakEnd"
                              :peak-rate-multiplier="(option as unknown as GroupOption).peakRateMultiplier"
                            />
                            <span v-else class="text-gray-400">{{ t('keys.selectGroup') }}</span>
                          </template>
                          <template #option="{ option, selected }">
                            <GroupOptionItem
                              :name="(option as unknown as GroupOption).label"
                              :platform="(option as unknown as GroupOption).platform"
                              :subscription-type="(option as unknown as GroupOption).subscriptionType"
                              :rate-multiplier="(option as unknown as GroupOption).rate"
                              :user-rate-multiplier="(option as unknown as GroupOption).userRate"
                              :peak-rate-enabled="(option as unknown as GroupOption).peakRateEnabled"
                              :peak-start="(option as unknown as GroupOption).peakStart"
                              :peak-end="(option as unknown as GroupOption).peakEnd"
                              :peak-rate-multiplier="(option as unknown as GroupOption).peakRateMultiplier"
                              :description="(option as unknown as GroupOption).description"
                              :selected="selected"
                            />
                          </template>
                        </Select>
                      </div>
                      <div>
                        <label class="input-label">{{ t('keys.priority') }}</label>
                        <div class="flex h-10 items-center rounded-lg border border-gray-200 bg-gray-50 px-3 text-sm font-semibold text-gray-700 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200">
                          {{ index + 1 }}
                        </div>
                      </div>
                      <div>
                        <label class="input-label">{{ t('keys.weight') }}</label>
                        <input
                          v-model.number="route.weight"
                          type="number"
                          min="1"
                          class="input"
                          @input="markRoutingManual"
                        />
                      </div>
                      <div>
                        <label class="input-label">{{ t('keys.cooldownSeconds') }}</label>
                        <input
                          v-model.number="route.cooldown_seconds"
                          type="number"
                          min="0"
                          class="input"
                          @input="markRoutingManual"
                        />
                      </div>
                    </div>

                    <div class="mt-3 grid gap-3 lg:grid-cols-[minmax(0,1fr)_auto]">
                      <div>
                        <label class="input-label">{{ t('keys.modelPatterns') }}</label>
                        <textarea
                          v-model="route.model_patterns_text"
                          rows="2"
                          class="input min-h-[4.5rem] resize-y"
                          :placeholder="t('keys.modelPatternsPlaceholder')"
                          @input="markRoutingManual"
                        />
                        <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">
                          {{ t('keys.modelPatternsHint') }}
                        </p>
                      </div>
                      <div class="flex flex-wrap items-start gap-2 pt-6 lg:w-56">
                        <label class="inline-flex h-9 items-center gap-2 rounded-lg border border-gray-200 px-3 text-sm text-gray-600 dark:border-dark-600 dark:text-gray-300">
                          <input
                            v-model="route.image_only"
                            type="checkbox"
                            class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                            @change="markRoutingManual"
                          />
                          <span>{{ t('keys.routeImageOnly') }}</span>
                        </label>
                        <label class="inline-flex h-9 items-center gap-2 rounded-lg border border-gray-200 px-3 text-sm text-gray-600 dark:border-dark-600 dark:text-gray-300">
                          <input
                            v-model="route.text_only"
                            type="checkbox"
                            class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                            @change="markRoutingManual"
                          />
                          <span>{{ t('keys.routeTextOnly') }}</span>
                        </label>
                      </div>
                    </div>
                  </div>
                </VueDraggable>

                <button type="button" class="btn btn-secondary" @click="addMultiGroupRoute">
                  <Icon name="plus" size="sm" class="mr-2" />
                  {{ t('keys.addRoute') }}
                </button>
              </div>
            </div>

            <!-- Custom Key Section (only for create) -->
            <div v-if="!showEditModal" class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="input-label mb-0">{{ t('keys.customKeyLabel') }}</label>
            <button
              type="button"
              @click="formData.use_custom_key = !formData.use_custom_key"
              :class="[
                'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                formData.use_custom_key ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
              ]"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  formData.use_custom_key ? 'translate-x-4' : 'translate-x-0'
                ]"
              />
            </button>
          </div>
          <div v-if="formData.use_custom_key">
            <input
              v-model="formData.custom_key"
              type="text"
              class="input font-mono"
              :placeholder="t('keys.customKeyPlaceholder')"
              :class="{ 'border-red-500 dark:border-red-500': customKeyError }"
            />
            <p v-if="customKeyError" class="mt-1 text-sm text-red-500">{{ customKeyError }}</p>
            <p v-else class="input-hint">{{ t('keys.customKeyHint') }}</p>
          </div>
            </div>

            <div v-if="showEditModal">
              <label class="input-label">{{ t('keys.statusLabel') }}</label>
              <Select
                v-model="formData.status"
                :options="statusOptions"
                :placeholder="t('keys.selectStatus')"
              />
            </div>

            <!-- IP Restriction Section -->
            <div class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="input-label mb-0">{{ t('keys.ipRestriction') }}</label>
            <button
              type="button"
              @click="formData.enable_ip_restriction = !formData.enable_ip_restriction"
              :class="[
                'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                formData.enable_ip_restriction ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
              ]"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  formData.enable_ip_restriction ? 'translate-x-4' : 'translate-x-0'
                ]"
              />
            </button>
          </div>

          <div v-if="formData.enable_ip_restriction" class="space-y-4 pt-2">
            <div>
              <label class="input-label">{{ t('keys.ipWhitelist') }}</label>
              <textarea
                v-model="formData.ip_whitelist"
                rows="3"
                class="input font-mono text-sm"
                :placeholder="t('keys.ipWhitelistPlaceholder')"
              />
              <p class="input-hint">{{ t('keys.ipWhitelistHint') }}</p>
            </div>

            <div>
              <label class="input-label">{{ t('keys.ipBlacklist') }}</label>
              <textarea
                v-model="formData.ip_blacklist"
                rows="3"
                class="input font-mono text-sm"
                :placeholder="t('keys.ipBlacklistPlaceholder')"
              />
              <p class="input-hint">{{ t('keys.ipBlacklistHint') }}</p>
            </div>
          </div>
            </div>

            <!-- Rate Limit Section -->
            <div class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="input-label mb-0">{{ t('keys.rateLimitSection') }}</label>
            <button
              type="button"
              @click="formData.enable_rate_limit = !formData.enable_rate_limit"
              :class="[
                'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                formData.enable_rate_limit ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
              ]"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  formData.enable_rate_limit ? 'translate-x-4' : 'translate-x-0'
                ]"
              />
            </button>
          </div>

          <div v-if="formData.enable_rate_limit" class="space-y-4 pt-2">
            <p class="input-hint -mt-2">{{ t('keys.rateLimitHint') }}</p>
            <!-- 5-Hour Limit -->
            <div>
              <label class="input-label">{{ t('keys.rateLimit5h') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">{{ CREDIT_SYMBOL }}</span>
                <input
                  v-model.number="formData.rate_limit_5h"
                  type="number"
                  step="0.01"
                  min="0"
                  class="input pl-7"
                  :placeholder="'0'"
                />
              </div>
              <!-- Usage info (edit mode only) -->
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_5h > 0" class="mt-2">
                <div class="flex items-center gap-2">
                  <div class="flex-1 rounded-lg bg-gray-100 px-3 py-2 dark:bg-dark-700 text-sm">
                    <span :class="[
                      'font-medium',
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h ? 'text-red-500' :
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h * 0.8 ? 'text-yellow-500' :
                      'text-gray-900 dark:text-white'
                    ]">
                      {{ formatCreditAmount(selectedKey.usage_5h || 0, { minimumFractionDigits: 4, maximumFractionDigits: 4 }) }}
                    </span>
                    <span class="mx-2 text-gray-400">/</span>
                    <span class="text-gray-500 dark:text-gray-400">
                      {{ formatCreditAmount(selectedKey.rate_limit_5h || 0) }}
                    </span>
                  </div>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                  <div
                    :class="[
                      'h-full rounded-full transition-all',
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h ? 'bg-red-500' :
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h * 0.8 ? 'bg-yellow-500' :
                      'bg-green-500'
                    ]"
                    :style="{ width: Math.min((selectedKey.usage_5h / selectedKey.rate_limit_5h) * 100, 100) + '%' }"
                  />
                </div>
              </div>
            </div>

            <!-- Daily Limit -->
            <div>
              <label class="input-label">{{ t('keys.rateLimit1d') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">{{ CREDIT_SYMBOL }}</span>
                <input
                  v-model.number="formData.rate_limit_1d"
                  type="number"
                  step="0.01"
                  min="0"
                  class="input pl-7"
                  :placeholder="'0'"
                />
              </div>
              <!-- Usage info (edit mode only) -->
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_1d > 0" class="mt-2">
                <div class="flex items-center gap-2">
                  <div class="flex-1 rounded-lg bg-gray-100 px-3 py-2 dark:bg-dark-700 text-sm">
                    <span :class="[
                      'font-medium',
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d ? 'text-red-500' :
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d * 0.8 ? 'text-yellow-500' :
                      'text-gray-900 dark:text-white'
                    ]">
                      {{ formatCreditAmount(selectedKey.usage_1d || 0, { minimumFractionDigits: 4, maximumFractionDigits: 4 }) }}
                    </span>
                    <span class="mx-2 text-gray-400">/</span>
                    <span class="text-gray-500 dark:text-gray-400">
                      {{ formatCreditAmount(selectedKey.rate_limit_1d || 0) }}
                    </span>
                  </div>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                  <div
                    :class="[
                      'h-full rounded-full transition-all',
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d ? 'bg-red-500' :
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d * 0.8 ? 'bg-yellow-500' :
                      'bg-green-500'
                    ]"
                    :style="{ width: Math.min((selectedKey.usage_1d / selectedKey.rate_limit_1d) * 100, 100) + '%' }"
                  />
                </div>
              </div>
            </div>

            <!-- 7-Day Limit -->
            <div>
              <label class="input-label">{{ t('keys.rateLimit7d') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">{{ CREDIT_SYMBOL }}</span>
                <input
                  v-model.number="formData.rate_limit_7d"
                  type="number"
                  step="0.01"
                  min="0"
                  class="input pl-7"
                  :placeholder="'0'"
                />
              </div>
              <!-- Usage info (edit mode only) -->
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_7d > 0" class="mt-2">
                <div class="flex items-center gap-2">
                  <div class="flex-1 rounded-lg bg-gray-100 px-3 py-2 dark:bg-dark-700 text-sm">
                    <span :class="[
                      'font-medium',
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d ? 'text-red-500' :
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d * 0.8 ? 'text-yellow-500' :
                      'text-gray-900 dark:text-white'
                    ]">
                      {{ formatCreditAmount(selectedKey.usage_7d || 0, { minimumFractionDigits: 4, maximumFractionDigits: 4 }) }}
                    </span>
                    <span class="mx-2 text-gray-400">/</span>
                    <span class="text-gray-500 dark:text-gray-400">
                      {{ formatCreditAmount(selectedKey.rate_limit_7d || 0) }}
                    </span>
                  </div>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                  <div
                    :class="[
                      'h-full rounded-full transition-all',
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d ? 'bg-red-500' :
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d * 0.8 ? 'bg-yellow-500' :
                      'bg-green-500'
                    ]"
                    :style="{ width: Math.min((selectedKey.usage_7d / selectedKey.rate_limit_7d) * 100, 100) + '%' }"
                  />
                </div>
              </div>
            </div>

            <!-- Reset Rate Limit button (edit mode only) -->
            <div v-if="showEditModal && selectedKey && (selectedKey.rate_limit_5h > 0 || selectedKey.rate_limit_1d > 0 || selectedKey.rate_limit_7d > 0)">
              <button
                type="button"
                @click="confirmResetRateLimit"
                class="btn btn-secondary text-sm"
              >
                {{ t('keys.resetRateLimitUsage') }}
              </button>
            </div>
          </div>
            </div>

            <!-- Expiration Section -->
            <div class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="input-label mb-0">{{ t('keys.expiration') }}</label>
            <button
              type="button"
              @click="formData.enable_expiration = !formData.enable_expiration"
              :class="[
                'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                formData.enable_expiration ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
              ]"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  formData.enable_expiration ? 'translate-x-4' : 'translate-x-0'
                ]"
              />
            </button>
          </div>

          <div v-if="formData.enable_expiration" class="space-y-4 pt-2">
            <!-- Quick select buttons (for both create and edit mode) -->
            <div class="flex flex-wrap gap-2">
              <button
                v-for="days in ['7', '30', '90']"
                :key="days"
                type="button"
                @click="setExpirationDays(parseInt(days))"
                :class="[
                  'rounded-lg px-3 py-1.5 text-sm transition-colors',
                  formData.expiration_preset === days
                    ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-400 dark:hover:bg-dark-600'
                ]"
              >
                {{ showEditModal ? t('keys.extendDays', { days }) : t('keys.expiresInDays', { days }) }}
              </button>
              <button
                type="button"
                @click="formData.expiration_preset = 'custom'"
                :class="[
                  'rounded-lg px-3 py-1.5 text-sm transition-colors',
                  formData.expiration_preset === 'custom'
                    ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-400 dark:hover:bg-dark-600'
                ]"
              >
                {{ t('keys.customDate') }}
              </button>
            </div>

            <!-- Date picker (always show for precise adjustment) -->
            <div>
              <label class="input-label">{{ t('keys.expirationDate') }}</label>
              <input
                v-model="formData.expiration_date"
                type="datetime-local"
                class="input"
              />
              <p class="input-hint">{{ t('keys.expirationDateHint') }}</p>
            </div>

            <!-- Current expiration display (only in edit mode) -->
            <div v-if="showEditModal && selectedKey?.expires_at" class="text-sm">
              <span class="text-gray-500 dark:text-gray-400">{{ t('keys.currentExpiration') }}: </span>
              <span class="font-medium text-gray-900 dark:text-white">
                {{ formatDateTime(selectedKey.expires_at) }}
              </span>
            </div>
          </div>
            </div>
          </div>
        </details>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button @click="closeModals" type="button" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button
            form="key-form"
            type="submit"
            :disabled="submitting"
            class="btn btn-primary"
            data-tour="key-form-submit"
          >
            <svg
              v-if="submitting"
              class="-ml-1 mr-2 h-4 w-4 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{
              submitting
                ? t('keys.saving')
                : showEditModal
                  ? t('common.update')
                  : t('common.create')
            }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('keys.deleteKey')"
      :message="t('keys.deleteConfirmMessage', { name: selectedKey?.name })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="handleDelete"
      @cancel="showDeleteDialog = false"
    />

    <!-- Reset Quota Confirmation Dialog -->
    <ConfirmDialog
      :show="showResetQuotaDialog"
      :title="t('keys.resetQuotaTitle')"
      :message="t('keys.resetQuotaConfirmMessage', { name: selectedKey?.name, used: selectedKey?.quota_used?.toFixed(4) })"
      :confirm-text="t('keys.reset')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="resetQuotaUsed"
      @cancel="showResetQuotaDialog = false"
    />

    <!-- Reset Rate Limit Confirmation Dialog -->
    <ConfirmDialog
      :show="showResetRateLimitDialog"
      :title="t('keys.resetRateLimitTitle')"
      :message="t('keys.resetRateLimitConfirmMessage', { name: selectedKey?.name })"
      :confirm-text="t('keys.reset')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="resetRateLimitUsage"
      @cancel="showResetRateLimitDialog = false"
    />

    <!-- Use Key Modal -->
    <UseKeyModal
      :show="showUseKeyModal"
      :api-key="selectedKey?.key || ''"
      :base-url="publicSettings?.api_base_url || ''"
      :platform="selectedKey?.group?.platform || null"
      :allow-messages-dispatch="selectedKey?.group?.allow_messages_dispatch || false"
      :unified-access="selectedKey ? apiKeySupportsUnifiedAccess(selectedKey) : false"
      :unified-capabilities="selectedKey ? apiKeyUnifiedAccessCapabilities(selectedKey) : undefined"
      @close="closeUseKeyModal"
    />

    <!-- CCS Client Selection Dialog for Antigravity -->
    <BaseDialog
      :show="showCcsClientSelect"
      :title="t('keys.ccsClientSelect.title')"
      width="narrow"
      @close="closeCcsClientSelect"
    >
      <div class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {{ t('keys.ccsClientSelect.description') }}
	        </p>
	        <div class="grid grid-cols-2 gap-3">
	          <button
	            @click="handleCcsClientSelect('claude')"
	            class="flex flex-col items-center gap-2 p-4 rounded-xl border-2 border-gray-200 dark:border-dark-600 hover:border-primary-500 dark:hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/20 transition-all"
	          >
	            <Icon name="terminal" size="xl" class="text-gray-600 dark:text-gray-400" />
	            <span class="font-medium text-gray-900 dark:text-white">{{
	              t('keys.ccsClientSelect.claudeCode')
	            }}</span>
	            <span class="text-xs text-gray-500 dark:text-gray-400">{{
	              t('keys.ccsClientSelect.claudeCodeDesc')
	            }}</span>
	          </button>
	          <button
	            @click="handleCcsClientSelect('gemini')"
	            class="flex flex-col items-center gap-2 p-4 rounded-xl border-2 border-gray-200 dark:border-dark-600 hover:border-primary-500 dark:hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/20 transition-all"
	          >
	            <Icon name="sparkles" size="xl" class="text-gray-600 dark:text-gray-400" />
	            <span class="font-medium text-gray-900 dark:text-white">{{
	              t('keys.ccsClientSelect.geminiCli')
	            }}</span>
	            <span class="text-xs text-gray-500 dark:text-gray-400">{{
	              t('keys.ccsClientSelect.geminiCliDesc')
	            }}</span>
	          </button>
	        </div>
	      </div>
      <template #footer>
        <div class="flex justify-end">
          <button @click="closeCcsClientSelect" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Cockpit Tools Install / Fallback Dialog -->
    <BaseDialog
      :show="showCockpitToolsInstallDialog"
      :title="t('keys.cockpitToolsInstall.title')"
      width="narrow"
      @close="closeCockpitToolsInstallDialog"
    >
      <div class="space-y-4">
        <p class="text-sm leading-6 text-gray-600 dark:text-gray-400">
          {{ t('keys.cockpitToolsInstall.description') }}
        </p>
        <div class="rounded-lg border border-cyan-100 bg-cyan-50 p-3 text-sm text-cyan-800 dark:border-cyan-800 dark:bg-cyan-900/20 dark:text-cyan-200">
          {{ t('keys.cockpitToolsInstall.fallbackHint') }}
        </div>
      </div>
      <template #footer>
        <div class="flex flex-wrap justify-end gap-3">
          <button
            type="button"
            class="btn btn-secondary"
            @click="downloadPendingCockpitToolsImport"
          >
            <Icon name="download" size="sm" class="mr-2" />
            {{ t('keys.cockpitToolsInstall.downloadJson') }}
          </button>
          <a
            href="https://github.com/jlcodes99/cockpit-tools/releases/latest"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-primary"
          >
            {{ t('keys.cockpitToolsInstall.downloadApp') }}
          </a>
        </div>
      </template>
    </BaseDialog>

    <!-- Group Selector Dropdown (Teleported to body to avoid overflow clipping) -->
    <Teleport to="body">
      <div
        v-if="groupSelectorKeyId !== null && dropdownPosition"
        ref="dropdownRef"
        class="animate-in fade-in slide-in-from-top-2 fixed z-[100000020] w-max min-w-[380px] overflow-hidden rounded-xl bg-white shadow-lg ring-1 ring-black/5 duration-200 dark:bg-dark-800 dark:ring-white/10"
        style="pointer-events: auto !important;"
        :style="{
          top: dropdownPosition.top !== undefined ? dropdownPosition.top + 'px' : undefined,
          bottom: dropdownPosition.bottom !== undefined ? dropdownPosition.bottom + 'px' : undefined,
          left: dropdownPosition.left + 'px'
        }"
      >
        <!-- Search box -->
        <div class="border-b border-gray-100 p-2 dark:border-dark-700">
          <div class="relative">
            <svg class="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              v-model="groupSearchQuery"
              type="text"
              class="w-full rounded-lg border border-gray-200 bg-gray-50 py-1.5 pl-8 pr-3 text-sm text-gray-900 placeholder-gray-400 outline-none focus:border-primary-300 focus:ring-1 focus:ring-primary-300 dark:border-dark-600 dark:bg-dark-700 dark:text-white dark:placeholder-gray-500 dark:focus:border-primary-600 dark:focus:ring-primary-600"
              :placeholder="t('keys.searchGroup')"
              @click.stop
            />
          </div>
        </div>
        <!-- Group list -->
        <div class="max-h-80 overflow-y-auto p-1.5">
          <button
            v-for="option in filteredGroupOptions"
            :key="option.value ?? 'null'"
            @click="changeGroup(selectedKeyForGroup!, option.value)"
            :class="[
              'flex w-full items-center justify-between rounded-lg px-3 py-2.5 text-sm transition-colors',
              'border-b border-gray-100 last:border-0 dark:border-dark-700',
              selectedKeyForGroup?.group_id === option.value ||
              (!selectedKeyForGroup?.group_id && option.value === null)
                ? 'bg-primary-50 dark:bg-primary-900/20'
                : 'hover:bg-gray-100 dark:hover:bg-dark-700'
            ]"
            :title="option.description || undefined"
          >
            <GroupOptionItem
              :name="option.label"
              :platform="option.platform"
              :subscription-type="option.subscriptionType"
              :rate-multiplier="option.rate"
              :user-rate-multiplier="option.userRate"
              :peak-rate-enabled="option.peakRateEnabled"
              :peak-start="option.peakStart"
              :peak-end="option.peakEnd"
              :peak-rate-multiplier="option.peakRateMultiplier"
              :description="option.description"
              :selected="
                selectedKeyForGroup?.group_id === option.value ||
                (!selectedKeyForGroup?.group_id && option.value === null)
              "
            />
          </button>
          <!-- Empty state when search has no results -->
          <div v-if="filteredGroupOptions.length === 0" class="py-4 text-center text-sm text-gray-400 dark:text-gray-500">
            {{ t('keys.noGroupFound') }}
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, type ComponentPublicInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { VueDraggable } from 'vue-draggable-plus'
import { keysAPI, authAPI, usageAPI, userGroupsAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import Icon from '@/components/icons/Icon.vue'
import UseKeyModal from '@/components/keys/UseKeyModal.vue'
import EndpointPopover from '@/components/keys/EndpointPopover.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { useAppStore } from '@/stores/app'
import { useOnboardingStore } from '@/stores/onboarding'
import { useClipboard } from '@/composables/useClipboard'
import type {
  AccountPoolStrategy,
  ApiKey,
  ApiKeyMultiGroupRoute,
  ApiKeyRoutingPreset,
  Group,
  PublicSettings,
  SubscriptionType,
  GroupPlatform,
  GroupRoutingScope
} from '@/types'
import type { Column } from '@/components/common/types'
import type { BatchApiKeyUsageStats } from '@/api/usage'
import { formatDateTime } from '@/utils/format'
import { CREDIT_SYMBOL, formatCreditAmount } from '@/utils/credits'
import { maskApiKey } from '@/utils/maskApiKey'
import {
  buildCcSwitchImportDeeplink,
  buildCcSwitchUsageScript,
  type CcSwitchClientType
} from '@/utils/ccswitchImport'
import {
  apiKeySupportsUnifiedAccess,
  apiKeyUnifiedAccessCapabilities
} from '@/utils/apiKeyCapabilities'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

// Helper to format date for datetime-local input
const formatDateTimeLocal = (isoDate: string): string => {
  const date = new Date(isoDate)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const INVALID_FILE_CHARS_REGEX = new RegExp(`[<>:"/\\\\|?*\\u0000-\\u001F]`, 'g')
let routeClientIdSeed = 0

interface GroupOption {
  value: number
  label: string
  description: string | null
  rate: number
  userRate: number | null
  peakRateEnabled: boolean
  peakStart: string
  peakEnd: string
  peakRateMultiplier: number
  subscriptionType: SubscriptionType
  platform: GroupPlatform
  status: Group['status']
  routingScope: GroupRoutingScope
  allowImageGeneration: boolean
  imageRateIndependent: boolean
  imageRate: number
}

interface ApiKeyMultiGroupRouteForm {
  client_id: string
  group_id: number | null
  priority: number
  weight: number
  cooldown_seconds: number
  enabled: boolean
  model_patterns_text: string
  image_only: boolean
  text_only: boolean
}

interface RoutingPresetOption {
  value: ApiKeyRoutingPreset
  title: string
  description: string
  icon: 'sparkles' | 'dollar' | 'bolt' | 'shield' | 'menu'
  tone: 'cyan' | 'blue' | 'emerald' | 'orange' | 'rose' | 'gray'
}

const appStore = useAppStore()
const onboardingStore = useOnboardingStore()
const { copyToClipboard: clipboardCopy } = useClipboard()

const columns = computed<Column[]>(() => [
  { key: 'name', label: t('common.name'), sortable: true },
  { key: 'key', label: t('keys.apiKey'), sortable: false },
  { key: 'group', label: t('keys.group'), sortable: false },
  { key: 'usage', label: t('keys.usage'), sortable: false },
  { key: 'rate_limit', label: t('keys.rateLimitColumn'), sortable: false },
  { key: 'expires_at', label: t('keys.expiresAt'), sortable: true },
  { key: 'status', label: t('common.status'), sortable: true },
  { key: 'last_used_at', label: t('keys.lastUsedAt'), sortable: true },
  { key: 'created_at', label: t('keys.created'), sortable: true },
  { key: 'actions', label: t('common.actions'), sortable: false }
])

const apiKeys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
const loading = ref(false)
const submitting = ref(false)
const now = ref(new Date())
let resetTimer: ReturnType<typeof setInterval> | null = null
const usageStats = ref<Record<string, BatchApiKeyUsageStats>>({})
const userGroupRates = ref<Record<number, number>>({})

const pagination = ref({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})
const sortState = ref({
  sort_by: 'created_at',
  sort_order: 'desc' as 'asc' | 'desc'
})

// Filter state
const filterSearch = ref('')
const filterStatus = ref('')
const filterGroupId = ref<string | number>('')

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteDialog = ref(false)
const showResetQuotaDialog = ref(false)
const showResetRateLimitDialog = ref(false)
const showUseKeyModal = ref(false)
const showCcsClientSelect = ref(false)
const showCockpitToolsInstallDialog = ref(false)
const routingPreset = ref<ApiKeyRoutingPreset>('optimal')
const defaultGroupTouched = ref(false)
const pendingCcsRow = ref<ApiKey | null>(null)
const pendingCockpitToolsRow = ref<ApiKey | null>(null)
const selectedKey = ref<ApiKey | null>(null)
const copiedKeyId = ref<number | null>(null)
const groupSelectorKeyId = ref<number | null>(null)
const publicSettings = ref<PublicSettings | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const dropdownPosition = ref<{ top?: number; bottom?: number; left: number } | null>(null)
const groupButtonRefs = ref<Map<number, HTMLElement>>(new Map())
let abortController: AbortController | null = null

const postCreateRedirect = computed(() => sanitizeInternalRedirect(route.query.redirect))

// Get the currently selected key for group change
const selectedKeyForGroup = computed(() => {
  if (groupSelectorKeyId.value === null) return null
  return apiKeys.value.find((k) => k.id === groupSelectorKeyId.value) || null
})

const isSmartRoutingKey = (key: ApiKey) => (key.multi_group_routes?.length ?? 0) > 0

const apiKeyCapabilityItems = (key: ApiKey) => {
  const capabilities = apiKeyUnifiedAccessCapabilities(key)
  const unifiedAccess = apiKeySupportsUnifiedAccess(key)
  if (!unifiedAccess && !isSmartRoutingKey(key)) {
    return []
  }
  return [
    {
      key: 'chat',
      label: t('keys.capabilities.chat'),
      icon: 'chat' as const,
      enabled: capabilities.chat
    },
    {
      key: 'image',
      label: t('keys.capabilities.image'),
      icon: 'image' as const,
      enabled: capabilities.image
    },
    {
      key: 'video',
      label: capabilities.video
        ? t('keys.capabilities.video')
        : t('keys.capabilities.videoDisabled'),
      icon: 'sparkles' as const,
      enabled: capabilities.video
    }
  ].filter((item) => item.enabled || (item.key === 'video' && unifiedAccess))
}

const setGroupButtonRef = (keyId: number, el: Element | ComponentPublicInstance | null) => {
  if (el instanceof HTMLElement) {
    groupButtonRefs.value.set(keyId, el)
  } else {
    groupButtonRefs.value.delete(keyId)
  }
}

const formData = ref({
  name: '',
  group_id: null as number | null,
  enable_multi_group_routing: false,
  multi_group_routes: [] as ApiKeyMultiGroupRouteForm[],
  account_pool_strategy: 'shared_only' as AccountPoolStrategy,
  status: 'active' as 'active' | 'inactive',
  use_custom_key: false,
  custom_key: '',
  enable_ip_restriction: false,
  ip_whitelist: '',
  ip_blacklist: '',
  // Quota settings (empty = unlimited)
  enable_quota: false,
  quota: null as number | null,
  // Rate limit settings
  enable_rate_limit: false,
  rate_limit_5h: null as number | null,
  rate_limit_1d: null as number | null,
  rate_limit_7d: null as number | null,
  enable_expiration: false,
  expiration_preset: '30' as '7' | '30' | '90' | 'custom',
  expiration_date: ''
})

// 自定义Key验证
const customKeyError = computed(() => {
  if (!formData.value.use_custom_key || !formData.value.custom_key) {
    return ''
  }
  const key = formData.value.custom_key
  if (key.length < 16) {
    return t('keys.customKeyTooShort')
  }
  // 检查字符：只允许字母、数字、下划线、连字符
  if (!/^[a-zA-Z0-9_-]+$/.test(key)) {
    return t('keys.customKeyInvalidChars')
  }
  return ''
})

const statusOptions = computed(() => [
  { value: 'active', label: t('common.active') },
  { value: 'inactive', label: t('common.inactive') }
])

const routingPresetOptions = computed<RoutingPresetOption[]>(() => [
  {
    value: 'optimal',
    title: t('keys.routingPreset.optimal.title'),
    description: t('keys.routingPreset.optimal.description'),
    icon: 'sparkles',
    tone: 'cyan'
  },
  {
    value: 'auto',
    title: t('keys.routingPreset.auto.title'),
    description: t('keys.routingPreset.auto.description'),
    icon: 'sparkles',
    tone: 'blue'
  },
  {
    value: 'cost',
    title: t('keys.routingPreset.cost.title'),
    description: t('keys.routingPreset.cost.description'),
    icon: 'dollar',
    tone: 'emerald'
  },
  {
    value: 'speed',
    title: t('keys.routingPreset.speed.title'),
    description: t('keys.routingPreset.speed.description'),
    icon: 'bolt',
    tone: 'orange'
  },
  {
    value: 'stability',
    title: t('keys.routingPreset.stability.title'),
    description: t('keys.routingPreset.stability.description'),
    icon: 'shield',
    tone: 'rose'
  }
])

const createRouteClientId = () => `route-${Date.now()}-${routeClientIdSeed++}`

const effectiveGroupRate = (group: GroupOption) => group.userRate ?? group.rate ?? 1

const effectiveImageGroupRate = (group: GroupOption) => {
  if (group.imageRateIndependent) {
    return group.imageRate
  }
  return effectiveGroupRate(group)
}

const routeKindModelPatterns: Record<'video' | 'embedding', string> = {
  video: 'doubao-seedance-*\n*-video-*',
  embedding: '*embedding*'
}

const presetToneClasses = (option: RoutingPresetOption) => {
  const selected = routingPreset.value === option.value
  const toneClasses: Record<RoutingPresetOption['tone'], string> = {
    cyan: selected
      ? 'border-cyan-500 bg-cyan-50 text-cyan-800 ring-1 ring-cyan-500 dark:border-cyan-400 dark:bg-cyan-500/10 dark:text-cyan-200'
      : 'border-gray-200 bg-white text-gray-700 hover:border-cyan-300 hover:bg-cyan-50/50 dark:border-dark-700 dark:bg-dark-900/40 dark:text-gray-300 dark:hover:border-cyan-500/50 dark:hover:bg-cyan-500/10',
    blue: selected
      ? 'border-blue-500 bg-blue-50 text-blue-800 ring-1 ring-blue-500 dark:border-blue-400 dark:bg-blue-500/10 dark:text-blue-200'
      : 'border-gray-200 bg-white text-gray-700 hover:border-blue-300 hover:bg-blue-50/50 dark:border-dark-700 dark:bg-dark-900/40 dark:text-gray-300 dark:hover:border-blue-500/50 dark:hover:bg-blue-500/10',
    emerald: selected
      ? 'border-emerald-500 bg-emerald-50 text-emerald-800 ring-1 ring-emerald-500 dark:border-emerald-400 dark:bg-emerald-500/10 dark:text-emerald-200'
      : 'border-gray-200 bg-white text-gray-700 hover:border-emerald-300 hover:bg-emerald-50/50 dark:border-dark-700 dark:bg-dark-900/40 dark:text-gray-300 dark:hover:border-emerald-500/50 dark:hover:bg-emerald-500/10',
    orange: selected
      ? 'border-orange-500 bg-orange-50 text-orange-800 ring-1 ring-orange-500 dark:border-orange-400 dark:bg-orange-500/10 dark:text-orange-200'
      : 'border-gray-200 bg-white text-gray-700 hover:border-orange-300 hover:bg-orange-50/50 dark:border-dark-700 dark:bg-dark-900/40 dark:text-gray-300 dark:hover:border-orange-500/50 dark:hover:bg-orange-500/10',
    rose: selected
      ? 'border-rose-500 bg-rose-50 text-rose-800 ring-1 ring-rose-500 dark:border-rose-400 dark:bg-rose-500/10 dark:text-rose-200'
      : 'border-gray-200 bg-white text-gray-700 hover:border-rose-300 hover:bg-rose-50/50 dark:border-dark-700 dark:bg-dark-900/40 dark:text-gray-300 dark:hover:border-rose-500/50 dark:hover:bg-rose-500/10',
    gray: selected
      ? 'border-gray-500 bg-gray-50 text-gray-800 ring-1 ring-gray-500 dark:border-gray-400 dark:bg-dark-700 dark:text-gray-200'
      : 'border-gray-200 bg-white text-gray-700 hover:border-gray-300 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-900/40 dark:text-gray-300 dark:hover:border-gray-500/50 dark:hover:bg-dark-800'
  }
  return toneClasses[option.tone]
}

const presetIconClasses = (option: RoutingPresetOption) => {
  const selected = routingPreset.value === option.value
  const toneClasses: Record<RoutingPresetOption['tone'], string> = {
    cyan: selected
      ? 'bg-cyan-500 text-white'
      : 'bg-cyan-50 text-cyan-600 dark:bg-cyan-500/15 dark:text-cyan-300',
    blue: selected
      ? 'bg-blue-500 text-white'
      : 'bg-blue-50 text-blue-600 dark:bg-blue-500/15 dark:text-blue-300',
    emerald: selected
      ? 'bg-emerald-500 text-white'
      : 'bg-emerald-50 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-300',
    orange: selected
      ? 'bg-orange-500 text-white'
      : 'bg-orange-50 text-orange-600 dark:bg-orange-500/15 dark:text-orange-300',
    rose: selected
      ? 'bg-rose-500 text-white'
      : 'bg-rose-50 text-rose-600 dark:bg-rose-500/15 dark:text-rose-300',
    gray: selected
      ? 'bg-gray-500 text-white'
      : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
  }
  return toneClasses[option.tone]
}

const renumberMultiGroupRoutePriorities = () => {
  formData.value.multi_group_routes.forEach((route, index) => {
    route.priority = index + 1
  })
}

const createDefaultRoute = (
  groupId: number | null = formData.value.group_id,
  overrides: Partial<
    Pick<
      ApiKeyMultiGroupRouteForm,
      'weight' | 'cooldown_seconds' | 'enabled' | 'model_patterns_text' | 'image_only' | 'text_only'
    >
  > = {}
): ApiKeyMultiGroupRouteForm => ({
  client_id: createRouteClientId(),
  group_id: groupId,
  priority: formData.value.multi_group_routes.length + 1,
  weight: overrides.weight ?? 1,
  cooldown_seconds: overrides.cooldown_seconds ?? 30,
  enabled: overrides.enabled ?? true,
  model_patterns_text: overrides.model_patterns_text ?? '',
  image_only: overrides.image_only ?? false,
  text_only: overrides.text_only ?? false
})

const getNextRouteGroupId = () => {
  const used = new Set(
    formData.value.multi_group_routes
      .filter((route) => !route.model_patterns_text.trim() && !route.image_only && !route.text_only)
      .map((route) => route.group_id)
  )
  return groups.value.find((group) => !used.has(group.id))?.id ?? formData.value.group_id
}

const sortPresetGroups = (
  routeGroups: GroupOption[],
  preset: ApiKeyRoutingPreset,
  rateResolver: (group: GroupOption) => number
): GroupOption[] => {
  const available = [...routeGroups]
  if (available.length === 0) {
    return []
  }
  if (preset === 'cost') {
    return available.sort((a, b) => {
      const rateDiff = rateResolver(a) - rateResolver(b)
      return rateDiff === 0 ? a.label.localeCompare(b.label) : rateDiff
    })
  }
  return available
}

const buildPresetRouteGroups = (
  preset: ApiKeyRoutingPreset,
  kind: 'text' | 'image' | 'video' | 'embedding'
): GroupOption[] => {
  const available = groupOptions.value.filter((group) => group.status === 'active')
  const routeGroups = available.filter((group) => {
    if (kind === 'text') {
      return group.routingScope === 'inference'
    }
    if (kind === 'image') {
      return group.routingScope === 'image' && group.platform === 'openai' && group.allowImageGeneration
    }
    return group.routingScope === kind
  })
  return sortPresetGroups(routeGroups, preset, kind === 'image' ? effectiveImageGroupRate : effectiveGroupRate)
}

const buildPresetRouteForms = (
  preset: ApiKeyRoutingPreset,
  routeGroups: GroupOption[],
  scopeOverrides: Partial<Pick<ApiKeyMultiGroupRouteForm, 'image_only' | 'text_only' | 'model_patterns_text'>>
) => routeGroups.map((group, index) => {
  let weight = 1
  let cooldownSeconds = 30
  let priority = index + 1
  if (preset === 'optimal' || preset === 'auto') {
    priority = 1
  }
  if (preset === 'speed') {
    priority = 1
    weight = Math.max(routeGroups.length - index, 1)
    cooldownSeconds = 15
  } else if (preset === 'stability') {
    cooldownSeconds = 60
  }
  return {
    ...createDefaultRoute(group.value, {
      weight,
      cooldown_seconds: cooldownSeconds,
      ...scopeOverrides
    }),
    priority
  }
})

const buildPresetRoutes = (preset: ApiKeyRoutingPreset): ApiKeyMultiGroupRouteForm[] => {
  const textRoutes = buildPresetRouteForms(preset, buildPresetRouteGroups(preset, 'text'), {
    text_only: true
  })
  const imageRoutes = buildPresetRouteForms(preset, buildPresetRouteGroups(preset, 'image'), {
    image_only: true
  })
  const videoRoutes = buildPresetRouteForms(preset, buildPresetRouteGroups(preset, 'video'), {
    model_patterns_text: routeKindModelPatterns.video
  })
  const embeddingRoutes = buildPresetRouteForms(preset, buildPresetRouteGroups(preset, 'embedding'), {
    model_patterns_text: routeKindModelPatterns.embedding
  })
  return [...textRoutes, ...imageRoutes, ...videoRoutes, ...embeddingRoutes]
}

const applyRoutingPreset = (
  preset: ApiKeyRoutingPreset,
  options: { preserveTouchedDefaultGroup?: boolean } = {}
) => {
  routingPreset.value = preset
  if (preset === 'manual') {
    formData.value.enable_multi_group_routing = formData.value.multi_group_routes.length > 0
    return
  }
  const routes = buildPresetRoutes(preset)
  if (routes.length === 0) {
    formData.value.enable_multi_group_routing = false
    formData.value.multi_group_routes = []
    return
  }
  formData.value.enable_multi_group_routing = true
  formData.value.multi_group_routes = routes
  if (!options.preserveTouchedDefaultGroup || !defaultGroupTouched.value || formData.value.group_id === null) {
    formData.value.group_id = routes[0]?.group_id ?? formData.value.group_id
  }
}

const markRoutingManual = () => {
  routingPreset.value = 'manual'
}

const toggleMultiGroupRouting = () => {
  markRoutingManual()
  formData.value.enable_multi_group_routing = !formData.value.enable_multi_group_routing
  if (formData.value.enable_multi_group_routing && formData.value.multi_group_routes.length === 0) {
    formData.value.multi_group_routes = [createDefaultRoute()]
  }
  renumberMultiGroupRoutePriorities()
}

const addMultiGroupRoute = () => {
  markRoutingManual()
  formData.value.multi_group_routes.push(createDefaultRoute(getNextRouteGroupId()))
  renumberMultiGroupRoutePriorities()
}

const removeMultiGroupRoute = (index: number) => {
  markRoutingManual()
  formData.value.multi_group_routes.splice(index, 1)
  renumberMultiGroupRoutePriorities()
}

const handleRouteOrderChanged = () => {
  markRoutingManual()
  renumberMultiGroupRoutePriorities()
}

const updateRouteGroup = (route: ApiKeyMultiGroupRouteForm, value: string | number | boolean | null) => {
  markRoutingManual()
  route.group_id = typeof value === 'number' ? value : null
}

const handleDefaultGroupChanged = (value: string | number | boolean | null) => {
  defaultGroupTouched.value = true
  formData.value.group_id = typeof value === 'number' ? value : null
}

const syncCurrentRoutingPreset = () => {
  if (!showCreateModal.value || showEditModal.value || routingPreset.value === 'manual') {
    return
  }
  applyRoutingPreset(routingPreset.value, { preserveTouchedDefaultGroup: true })
}

const parseModelPatternsText = (value: string): string[] => {
  return (value || '')
    .split(/[\n,]/)
    .map((item) => item.trim())
    .filter(Boolean)
    .filter((item, index, arr) =>
      arr.findIndex((candidate) => candidate.toLowerCase() === item.toLowerCase()) === index
    )
}

const formatModelPatternsText = (patterns?: string[]) => (patterns || []).join('\n')

const normalizeRouteForm = (route: ApiKeyMultiGroupRoute): ApiKeyMultiGroupRouteForm => ({
  client_id: createRouteClientId(),
  group_id: route.group_id,
  priority: route.priority || 100,
  weight: route.weight > 0 ? route.weight : 1,
  cooldown_seconds: route.cooldown_seconds >= 0 ? route.cooldown_seconds : 30,
  enabled: route.enabled,
  model_patterns_text: formatModelPatternsText(route.model_patterns),
  image_only: Boolean(route.image_only),
  text_only: Boolean(route.text_only)
})

const normalizeRouteForms = (routes: ApiKeyMultiGroupRoute[]): ApiKeyMultiGroupRouteForm[] => {
  const normalized = routes
    .map((route, index) => ({ route, index }))
    .sort((a, b) => {
      const priorityA = a.route.priority > 0 ? a.route.priority : 100
      const priorityB = b.route.priority > 0 ? b.route.priority : 100
      return priorityA === priorityB ? a.index - b.index : priorityA - priorityB
    })
    .map(({ route }) => normalizeRouteForm(route))
  return normalized
}

const buildMultiGroupRoutes = (): ApiKeyMultiGroupRoute[] => {
  return formData.value.multi_group_routes.map((route) => {
    const modelPatterns = parseModelPatternsText(route.model_patterns_text)
    return {
      group_id: route.group_id as number,
      priority: Number.isFinite(route.priority) && route.priority > 0 ? Number(route.priority) : 100,
      weight: Number.isFinite(route.weight) ? Math.max(1, Number(route.weight)) : 1,
      cooldown_seconds: Number.isFinite(route.cooldown_seconds)
        ? Math.max(0, Number(route.cooldown_seconds))
        : 30,
      enabled: route.enabled,
      ...(modelPatterns.length > 0 ? { model_patterns: modelPatterns } : {}),
      ...(route.image_only ? { image_only: true } : {}),
      ...(route.text_only ? { text_only: true } : {})
    }
  })
}

const routeScopeKey = (route: ApiKeyMultiGroupRouteForm) => {
  const patterns = parseModelPatternsText(route.model_patterns_text)
    .map((pattern) => pattern.toLowerCase())
    .sort()
    .join(',')
  return `${route.group_id}|${route.image_only ? 'image' : ''}|${route.text_only ? 'text' : ''}|${patterns}`
}

const validateMultiGroupRoutes = (): ApiKeyMultiGroupRoute[] | null => {
  if (!formData.value.enable_multi_group_routing) {
    return []
  }
  if (formData.value.multi_group_routes.length === 0) {
    appStore.showError(t('keys.routeRequired'))
    return null
  }
  if (formData.value.multi_group_routes.some((route) => route.group_id === null)) {
    appStore.showError(t('keys.routeGroupRequired'))
    return null
  }
  if (formData.value.multi_group_routes.some((route) => route.image_only && route.text_only)) {
    appStore.showError(t('keys.routeScopeConflict'))
    return null
  }
  const routeScopes = formData.value.multi_group_routes.map(routeScopeKey)
  if (new Set(routeScopes).size !== routeScopes.length) {
    appStore.showError(t('keys.routeDuplicateGroup'))
    return null
  }
  return buildMultiGroupRoutes()
}

// Filter dropdown options
const groupFilterOptions = computed(() => [
  { value: '', label: t('keys.allGroups') },
  { value: 0, label: t('keys.noGroup') },
  ...groups.value.map((g) => ({ value: g.id, label: g.name }))
])

const statusFilterOptions = computed(() => [
  { value: '', label: t('keys.allStatus') },
  { value: 'active', label: t('keys.status.active') },
  { value: 'inactive', label: t('keys.status.inactive') },
  { value: 'quota_exhausted', label: t('keys.status.quota_exhausted') },
  { value: 'expired', label: t('keys.status.expired') }
])

const accountPoolStrategyOptions = computed(() => [
  { value: 'shared_only', label: t('keys.accountPoolStrategy.sharedOnly') },
  { value: 'private_first', label: t('keys.accountPoolStrategy.privateFirst') },
  { value: 'private_only', label: t('keys.accountPoolStrategy.privateOnly') }
])

const accountPoolStrategyLabel = (strategy: AccountPoolStrategy | string) => {
  switch (strategy) {
    case 'private_first':
      return t('keys.accountPoolStrategy.privateFirst')
    case 'private_only':
      return t('keys.accountPoolStrategy.privateOnly')
    default:
      return t('keys.accountPoolStrategy.sharedOnly')
  }
}

const onFilterChange = () => {
  pagination.value.page = 1
  loadApiKeys()
}

const onGroupFilterChange = (value: string | number | boolean | null) => {
  filterGroupId.value = value as string | number
  onFilterChange()
}

const onStatusFilterChange = (value: string | number | boolean | null) => {
  filterStatus.value = value as string
  onFilterChange()
}

// Convert groups to Select options format with rate multiplier and subscription type
const groupOptions = computed(() =>
  groups.value.map((group) => ({
    value: group.id,
    label: group.name,
    description: group.description,
    rate: group.rate_multiplier,
    userRate: userGroupRates.value[group.id] ?? null,
    peakRateEnabled: group.peak_rate_enabled,
    peakStart: group.peak_start,
    peakEnd: group.peak_end,
    peakRateMultiplier: group.peak_rate_multiplier,
    subscriptionType: group.subscription_type,
    platform: group.platform,
    status: group.status,
    routingScope: group.routing_scope || 'inference',
    allowImageGeneration: group.allow_image_generation,
    imageRateIndependent: group.image_rate_independent,
    imageRate: group.image_rate_multiplier
  }))
)

// Group dropdown search
const groupSearchQuery = ref('')
const filteredGroupOptions = computed(() => {
  const query = groupSearchQuery.value.trim().toLowerCase()
  if (!query) return groupOptions.value
  return groupOptions.value.filter((opt) => {
    return opt.label.toLowerCase().includes(query) ||
      (opt.description && opt.description.toLowerCase().includes(query))
  })
})

const copyToClipboard = async (text: string, keyId: number) => {
  const success = await clipboardCopy(text, t('keys.copied'))
  if (success) {
    copiedKeyId.value = keyId
    setTimeout(() => {
      copiedKeyId.value = null
    }, 800)
  }
}

const isAbortError = (error: unknown) => {
  if (!error || typeof error !== 'object') return false
  const { name, code } = error as { name?: string; code?: string }
  return name === 'AbortError' || code === 'ERR_CANCELED'
}

const loadApiKeys = async () => {
  abortController?.abort()
  const controller = new AbortController()
  abortController = controller
  const { signal } = controller
  loading.value = true
  try {
    // Build filters
    const filters: {
      search?: string
      status?: string
      group_id?: number | string
      sort_by?: string
      sort_order?: 'asc' | 'desc'
    } = {}
    if (filterSearch.value) filters.search = filterSearch.value
    if (filterStatus.value) filters.status = filterStatus.value
    if (filterGroupId.value !== '') filters.group_id = filterGroupId.value
    filters.sort_by = sortState.value.sort_by
    filters.sort_order = sortState.value.sort_order

    const response = await keysAPI.list(pagination.value.page, pagination.value.page_size, filters, {
      signal
    })
    if (signal.aborted) return
    apiKeys.value = response.items
    pagination.value.total = response.total
    pagination.value.pages = response.pages

    // Load usage stats for all API keys in the list
    if (response.items.length > 0) {
      const keyIds = response.items.map((k) => k.id)
      try {
        const usageResponse = await usageAPI.getDashboardApiKeysUsage(keyIds, { signal })
        if (signal.aborted) return
        usageStats.value = usageResponse.stats
      } catch (e) {
        if (!isAbortError(e)) {
          console.error('Failed to load usage stats:', e)
        }
      }
    }
  } catch (error) {
    if (isAbortError(error)) {
      return
    }
    appStore.showError(t('keys.failedToLoad'))
  } finally {
    if (abortController === controller) {
      loading.value = false
    }
  }
}

const loadGroups = async () => {
  try {
    groups.value = await userGroupsAPI.getAvailable()
    syncCurrentRoutingPreset()
  } catch (error) {
    console.error('Failed to load groups:', error)
  }
}

const loadUserGroupRates = async () => {
  try {
    userGroupRates.value = await userGroupsAPI.getUserGroupRates()
    syncCurrentRoutingPreset()
  } catch (error) {
    console.error('Failed to load user group rates:', error)
  }
}

const loadPublicSettings = async () => {
  try {
    publicSettings.value = await authAPI.getPublicSettings()
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }
}

const openUseKeyModal = (key: ApiKey) => {
  selectedKey.value = key
  showUseKeyModal.value = true
}

const closeUseKeyModal = () => {
  showUseKeyModal.value = false
  selectedKey.value = null
}

const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadApiKeys()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.page_size = pageSize
  pagination.value.page = 1
  loadApiKeys()
}

const handleSort = (key: string, order: 'asc' | 'desc') => {
  sortState.value.sort_by = key
  sortState.value.sort_order = order
  pagination.value.page = 1
  loadApiKeys()
}

const editKey = (key: ApiKey) => {
  selectedKey.value = key
  routingPreset.value = 'manual'
  defaultGroupTouched.value = true
  const hasIPRestriction = (key.ip_whitelist?.length > 0) || (key.ip_blacklist?.length > 0)
  const hasExpiration = !!key.expires_at
  const multiGroupRoutes = normalizeRouteForms(key.multi_group_routes || [])
  formData.value = {
    name: key.name,
    group_id: key.group_id,
    enable_multi_group_routing: multiGroupRoutes.length > 0,
    multi_group_routes: multiGroupRoutes,
    account_pool_strategy: key.account_pool_strategy || 'shared_only',
    status: key.status === 'quota_exhausted' || key.status === 'expired' ? 'inactive' : key.status,
    use_custom_key: false,
    custom_key: '',
    enable_ip_restriction: hasIPRestriction,
    ip_whitelist: (key.ip_whitelist || []).join('\n'),
    ip_blacklist: (key.ip_blacklist || []).join('\n'),
    enable_quota: key.quota > 0,
    quota: key.quota > 0 ? key.quota : null,
    enable_rate_limit: (key.rate_limit_5h > 0) || (key.rate_limit_1d > 0) || (key.rate_limit_7d > 0),
    rate_limit_5h: key.rate_limit_5h || null,
    rate_limit_1d: key.rate_limit_1d || null,
    rate_limit_7d: key.rate_limit_7d || null,
    enable_expiration: hasExpiration,
    expiration_preset: 'custom',
    expiration_date: key.expires_at ? formatDateTimeLocal(key.expires_at) : ''
  }
  showEditModal.value = true
}

const toggleKeyStatus = async (key: ApiKey) => {
  const newStatus = key.status === 'active' ? 'inactive' : 'active'
  try {
    await keysAPI.toggleStatus(key.id, newStatus)
    appStore.showSuccess(
      newStatus === 'active' ? t('keys.keyEnabledSuccess') : t('keys.keyDisabledSuccess')
    )
    loadApiKeys()
  } catch (error) {
    appStore.showError(t('keys.failedToUpdateStatus'))
  }
}

const openGroupSelector = (key: ApiKey) => {
  if (isSmartRoutingKey(key)) {
    groupSelectorKeyId.value = null
    dropdownPosition.value = null
    return
  }

  if (groupSelectorKeyId.value === key.id) {
    groupSelectorKeyId.value = null
    dropdownPosition.value = null
  } else {
    const buttonEl = groupButtonRefs.value.get(key.id)
    if (buttonEl) {
      const rect = buttonEl.getBoundingClientRect()
      const dropdownEstHeight = 400 // estimated max dropdown height
      const spaceBelow = window.innerHeight - rect.bottom
      const spaceAbove = rect.top

      if (spaceBelow < dropdownEstHeight && spaceAbove > spaceBelow) {
        // Not enough space below, pop upward
        dropdownPosition.value = {
          bottom: window.innerHeight - rect.top + 4,
          left: rect.left
        }
      } else {
        // Default: pop downward
        dropdownPosition.value = {
          top: rect.bottom + 4,
          left: rect.left
        }
      }
    }
    groupSelectorKeyId.value = key.id
    groupSearchQuery.value = ''
  }
}

const changeGroup = async (key: ApiKey, newGroupId: number | null) => {
  groupSelectorKeyId.value = null
  dropdownPosition.value = null
  if (key.group_id === newGroupId) return

  try {
    await keysAPI.update(key.id, { group_id: newGroupId })
    appStore.showSuccess(t('keys.groupChangedSuccess'))
    loadApiKeys()
  } catch (error) {
    appStore.showError(t('keys.failedToChangeGroup'))
  }
}

const closeGroupSelector = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  // Check if click is inside the dropdown or the trigger button
  if (!target.closest('.group\\/dropdown') && !dropdownRef.value?.contains(target)) {
    groupSelectorKeyId.value = null
    dropdownPosition.value = null
  }
}

const confirmDelete = (key: ApiKey) => {
  if (key.is_default) {
    appStore.showInfo(t('keys.defaultKeyDeleteDisabled'))
    return
  }
  selectedKey.value = key
  showDeleteDialog.value = true
}

const openCreateModal = () => {
  selectedKey.value = null
  showEditModal.value = false
  closeModals()
  showCreateModal.value = true
  applyRoutingPreset('optimal')
}

const handleSubmit = async () => {
  // Validate group_id is required
  if (formData.value.group_id === null) {
    appStore.showError(t('keys.groupRequired'))
    return
  }
  const multiGroupRoutes = validateMultiGroupRoutes()
  if (multiGroupRoutes === null) {
    return
  }

  // Validate custom key if enabled
  if (!showEditModal.value && formData.value.use_custom_key) {
    if (!formData.value.custom_key) {
      appStore.showError(t('keys.customKeyRequired'))
      return
    }
    if (customKeyError.value) {
      appStore.showError(customKeyError.value)
      return
    }
  }

  // Parse IP lists only if IP restriction is enabled
  const parseIPList = (text: string): string[] =>
    text.split('\n').map(ip => ip.trim()).filter(ip => ip.length > 0)
  const ipWhitelist = formData.value.enable_ip_restriction ? parseIPList(formData.value.ip_whitelist) : []
  const ipBlacklist = formData.value.enable_ip_restriction ? parseIPList(formData.value.ip_blacklist) : []

  // Calculate quota value (null/empty/0 = unlimited, stored as 0)
  const quota = formData.value.enable_quota && formData.value.quota && formData.value.quota > 0
    ? formData.value.quota
    : 0

  // Calculate expiration
  let expiresInDays: number | undefined
  let expiresAt: string | null | undefined
  if (formData.value.enable_expiration && formData.value.expiration_date) {
    if (!showEditModal.value) {
      // Create mode: calculate days from date
      const expDate = new Date(formData.value.expiration_date)
      const now = new Date()
      const diffDays = Math.ceil((expDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
      expiresInDays = diffDays > 0 ? diffDays : 1
    } else {
      // Edit mode: use custom date directly
      expiresAt = new Date(formData.value.expiration_date).toISOString()
    }
  } else if (showEditModal.value) {
    // Edit mode: if expiration disabled or date cleared, send empty string to clear
    expiresAt = ''
  }

  // Calculate rate limit values (send 0 when toggle is off)
  const rateLimitData = formData.value.enable_rate_limit ? {
    rate_limit_5h: formData.value.rate_limit_5h && formData.value.rate_limit_5h > 0 ? formData.value.rate_limit_5h : 0,
    rate_limit_1d: formData.value.rate_limit_1d && formData.value.rate_limit_1d > 0 ? formData.value.rate_limit_1d : 0,
    rate_limit_7d: formData.value.rate_limit_7d && formData.value.rate_limit_7d > 0 ? formData.value.rate_limit_7d : 0,
  } : { rate_limit_5h: 0, rate_limit_1d: 0, rate_limit_7d: 0 }

  submitting.value = true
  try {
    if (showEditModal.value && selectedKey.value) {
      await keysAPI.update(selectedKey.value.id, {
        name: formData.value.name,
        group_id: formData.value.group_id,
        multi_group_routes: multiGroupRoutes,
        account_pool_strategy: formData.value.account_pool_strategy,
        status: formData.value.status,
        ip_whitelist: ipWhitelist,
        ip_blacklist: ipBlacklist,
        quota: quota,
        expires_at: expiresAt,
        rate_limit_5h: rateLimitData.rate_limit_5h,
        rate_limit_1d: rateLimitData.rate_limit_1d,
        rate_limit_7d: rateLimitData.rate_limit_7d,
      })
      appStore.showSuccess(t('keys.keyUpdatedSuccess'))
    } else {
      const customKey = formData.value.use_custom_key ? formData.value.custom_key : undefined
      await keysAPI.create(
        formData.value.name,
        formData.value.group_id,
        customKey,
        ipWhitelist,
        ipBlacklist,
        quota,
        expiresInDays,
        rateLimitData,
        multiGroupRoutes,
        formData.value.account_pool_strategy
      )
      appStore.showSuccess(t('keys.keyCreatedSuccess'))
      // Only advance tour if active, on submit step, and creation succeeded
      if (onboardingStore.isCurrentStep('[data-tour="key-form-submit"]')) {
        onboardingStore.nextStep(500)
      }
      if (postCreateRedirect.value) {
        closeModals()
        await router.push(postCreateRedirect.value)
        return
      }
    }
    closeModals()
    loadApiKeys()
  } catch (error: any) {
    const errorMsg = error.response?.data?.detail || t('keys.failedToSave')
    appStore.showError(errorMsg)
    // Don't advance tour on error
  } finally {
    submitting.value = false
  }
}

function sanitizeInternalRedirect(value: unknown): string {
  const raw = Array.isArray(value) ? value[0] : value
  if (typeof raw !== 'string') {
    return ''
  }
  const trimmed = raw.trim()
  if (!trimmed || !trimmed.startsWith('/') || trimmed.startsWith('//')) {
    return ''
  }
  return trimmed
}

/**
 * 处理删除 API Key 的操作
 * 优化：错误处理改进，优先显示后端返回的具体错误消息（如权限不足等），
 * 若后端未返回消息则显示默认的国际化文本
 */
const handleDelete = async () => {
  if (!selectedKey.value) return
  if (selectedKey.value.is_default) {
    appStore.showInfo(t('keys.defaultKeyDeleteDisabled'))
    showDeleteDialog.value = false
    return
  }

  try {
    await keysAPI.delete(selectedKey.value.id)
    appStore.showSuccess(t('keys.keyDeletedSuccess'))
    showDeleteDialog.value = false
    loadApiKeys()
  } catch (error: any) {
    // 优先使用后端返回的错误消息，提供更具体的错误信息给用户
    const errorMsg = error?.message || t('keys.failedToDelete')
    appStore.showError(errorMsg)
  }
}

const closeModals = () => {
  showCreateModal.value = false
  showEditModal.value = false
  selectedKey.value = null
  routingPreset.value = 'optimal'
  defaultGroupTouched.value = false
  formData.value = {
    name: '',
    group_id: null,
    enable_multi_group_routing: false,
    multi_group_routes: [],
    account_pool_strategy: 'shared_only',
    status: 'active',
    use_custom_key: false,
    custom_key: '',
    enable_ip_restriction: false,
    ip_whitelist: '',
    ip_blacklist: '',
    enable_quota: false,
    quota: null,
    enable_rate_limit: false,
    rate_limit_5h: null,
    rate_limit_1d: null,
    rate_limit_7d: null,
    enable_expiration: false,
    expiration_preset: '30',
    expiration_date: ''
  }
}

// Show reset quota confirmation dialog
const confirmResetQuota = () => {
  showResetQuotaDialog.value = true
}

// Set expiration date based on quick select days
const setExpirationDays = (days: number) => {
  formData.value.expiration_preset = days.toString() as '7' | '30' | '90'
  const expDate = new Date()
  expDate.setDate(expDate.getDate() + days)
  formData.value.expiration_date = formatDateTimeLocal(expDate.toISOString())
}

// Reset quota used for an API key
const resetQuotaUsed = async () => {
  if (!selectedKey.value) return
  showResetQuotaDialog.value = false
  try {
    await keysAPI.update(selectedKey.value.id, { reset_quota: true })
    appStore.showSuccess(t('keys.quotaResetSuccess'))
    // Update local state
    if (selectedKey.value) {
      selectedKey.value.quota_used = 0
    }
  } catch (error: any) {
    const errorMsg = error.response?.data?.detail || t('keys.failedToResetQuota')
    appStore.showError(errorMsg)
  }
}

// Show reset rate limit confirmation dialog (from edit modal)
const confirmResetRateLimit = () => {
  showResetRateLimitDialog.value = true
}

// Show reset rate limit confirmation dialog (from table row)
const confirmResetRateLimitFromTable = (row: ApiKey) => {
  selectedKey.value = row
  showResetRateLimitDialog.value = true
}

// Reset rate limit usage for an API key
const resetRateLimitUsage = async () => {
  if (!selectedKey.value) return
  showResetRateLimitDialog.value = false
  try {
    await keysAPI.update(selectedKey.value.id, { reset_rate_limit_usage: true })
    appStore.showSuccess(t('keys.rateLimitResetSuccess'))
    // Refresh key data
    await loadApiKeys()
    // Update the editing key with fresh data
    const refreshedKey = apiKeys.value.find(k => k.id === selectedKey.value!.id)
    if (refreshedKey) {
      selectedKey.value = refreshedKey
    }
  } catch (error: any) {
    const errorMsg = error.response?.data?.detail || t('keys.failedToResetRateLimit')
    appStore.showError(errorMsg)
  }
}

const importToCcswitch = (row: ApiKey) => {
  const platform = row.group?.platform || 'anthropic'

  // For antigravity platform, show client selection dialog
  if (platform === 'antigravity') {
    pendingCcsRow.value = row
    showCcsClientSelect.value = true
    return
  }

  // For other platforms, execute directly
  executeCcsImport(row, platform === 'gemini' ? 'gemini' : 'claude')
}

const executeCcsImport = (row: ApiKey, clientType: CcSwitchClientType) => {
  const baseUrl = resolveApiBaseUrl()
  const platform = row.group?.platform || 'anthropic'
  const usageScript = buildCcSwitchUsageScript()
  const providerName = (publicSettings.value?.site_name || 'sub2api').trim() || 'sub2api'
  const deeplink = buildCcSwitchImportDeeplink({
    baseUrl,
    platform,
    clientType,
    providerName,
    apiKey: row.key,
    usageScript
  })

  try {
    window.open(deeplink, '_self')

    // Check if the protocol handler worked by detecting if we're still focused
    setTimeout(() => {
      if (document.hasFocus()) {
        // Still focused means the protocol handler likely failed
        appStore.showError(t('keys.ccSwitchNotInstalled'))
      }
    }, 100)
  } catch (error) {
    appStore.showError(t('keys.ccSwitchNotInstalled'))
  }
}

const handleCcsClientSelect = (clientType: CcSwitchClientType) => {
  if (pendingCcsRow.value) {
    executeCcsImport(pendingCcsRow.value, clientType)
  }
  showCcsClientSelect.value = false
  pendingCcsRow.value = null
}

const closeCcsClientSelect = () => {
  showCcsClientSelect.value = false
  pendingCcsRow.value = null
}

function resolveApiBaseUrl(): string {
  return (publicSettings.value?.api_base_url || window.location.origin).trim().replace(/\/+$/, '')
}

function sanitizeFileNameSegment(input: string | undefined, fallback: string): string {
  const normalized = (input || '')
    .trim()
    .replace(INVALID_FILE_CHARS_REGEX, '_')
    .replace(/\s+/g, '_')
    .replace(/_+/g, '_')
    .replace(/^_+|_+$/g, '')
  return normalized || fallback
}

function buildCockpitToolsCodexImportPayload(row: ApiKey) {
  const baseUrl = resolveApiBaseUrl()
  const siteName = (publicSettings.value?.site_name || 'Sub2API').trim() || 'Sub2API'
  const accountName = row.name?.trim() || siteName

  return {
    auth_mode: 'apikey',
    OPENAI_API_KEY: row.key,
    api_base_url: baseUrl,
    api_provider_mode: 'custom',
    api_provider_id: sanitizeFileNameSegment(siteName.toLowerCase(), 'sub2api'),
    api_provider_name: siteName,
    email: `api-key-${row.id}@sub2api.local`,
    account_name: accountName,
    account_note: t('keys.cockpitToolsImportNote', { siteName }),
    plan_type: 'API Key',
    created_at: Math.floor(Date.now() / 1000),
    last_used: Math.floor(Date.now() / 1000),
  }
}

function buildCockpitToolsImportDeeplink(row: ApiKey): string {
  const params = new URLSearchParams([
    ['provider', 'codex'],
    ['payload', JSON.stringify(buildCockpitToolsCodexImportPayload(row))],
    ['auto_import', 'true'],
    ['activate', 'true'],
    ['source', 'sub2api']
  ])
  return `cockpit-tools://import?${params.toString()}`
}

function downloadJsonFile(fileName: string, payload: unknown) {
  const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

function downloadCockpitToolsImportFile(row: ApiKey) {
  const name = sanitizeFileNameSegment(row.name, `key-${row.id}`)
  downloadJsonFile(
    `cockpit-tools-codex-${name}.json`,
    buildCockpitToolsCodexImportPayload(row)
  )
}

function downloadPendingCockpitToolsImport() {
  if (!pendingCockpitToolsRow.value) return
  downloadCockpitToolsImportFile(pendingCockpitToolsRow.value)
}

function closeCockpitToolsInstallDialog() {
  showCockpitToolsInstallDialog.value = false
  pendingCockpitToolsRow.value = null
}

function importToCockpitTools(row: ApiKey) {
  try {
    window.open(buildCockpitToolsImportDeeplink(row), '_self')
    setTimeout(() => {
      if (document.hasFocus()) {
        pendingCockpitToolsRow.value = row
        showCockpitToolsInstallDialog.value = true
      }
    }, 300)
  } catch (error) {
    appStore.showError(t('keys.cockpitToolsImportFailed'))
  }
}

function formatResetTime(resetAt: string | null): string {
  if (!resetAt) return ''
  const diff = new Date(resetAt).getTime() - now.value.getTime()
  if (diff <= 0) return t('keys.resetNow')
  const days = Math.floor(diff / 86400000)
  const hours = Math.floor((diff % 86400000) / 3600000)
  const mins = Math.floor((diff % 3600000) / 60000)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

onMounted(() => {
  loadApiKeys()
  loadGroups()
  loadUserGroupRates()
  loadPublicSettings()
  if (route.query.create === '1') {
    openCreateModal()
    router.replace({ path: route.path, query: { ...route.query, create: undefined } })
  }
  document.addEventListener('click', closeGroupSelector)
  resetTimer = setInterval(() => { now.value = new Date() }, 60000)
})

onUnmounted(() => {
  document.removeEventListener('click', closeGroupSelector)
  if (resetTimer) clearInterval(resetTimer)
})
</script>
