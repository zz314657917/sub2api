<template>
  <AppLayout>
    <div class="space-y-6">
      <UserSubscriptionsPanel id="subscriptions" />

      <section class="space-y-4" data-test="user-usage-analytics">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('usage.analyticsTitle') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('usage.analyticsDescription') }}</p>
          </div>
          <div class="inline-flex items-center gap-2">
            <span class="text-sm font-medium text-gray-600 dark:text-gray-300">{{ t('usage.granularity') }}</span>
            <div class="w-28">
              <Select v-model="granularity" :options="granularityOptions" @change="loadChartData" />
            </div>
          </div>
        </div>

        <UsageStatsCards
          :stats="usageStats"
          :show-account-cost="false"
          :strike-standard-cost="true"
        />

        <div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
          <ModelDistributionChart
            v-model:metric="modelDistributionMetric"
            :model-stats="requestedModelStats"
            :loading="modelStatsLoading"
            :show-source-toggle="false"
            :show-metric-toggle="true"
            :enable-breakdown="false"
            :show-account-cost="false"
            :start-date="startDate"
            :end-date="endDate"
          />
          <GroupDistributionChart
            v-model:metric="groupDistributionMetric"
            :group-stats="groupStats"
            :loading="chartsLoading"
            :show-metric-toggle="true"
            :enable-breakdown="false"
            :show-account-cost="false"
            :start-date="startDate"
            :end-date="endDate"
          />
          <EndpointDistributionChart
            v-model:metric="endpointDistributionMetric"
            :endpoint-stats="endpointStats"
            :loading="endpointStatsLoading"
            :show-source-toggle="false"
            :show-metric-toggle="true"
            :enable-breakdown="false"
            :title="t('usage.endpointDistribution')"
            :start-date="startDate"
            :end-date="endDate"
          />
          <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
        </div>
      </section>

      <div v-if="errorViewEnabled" class="flex gap-1 border-b border-gray-200 dark:border-dark-700" data-test="user-usage-tabs">
        <button
          type="button"
          class="border-b-2 px-4 py-2 text-sm font-medium transition"
          :class="activeTab === 'usage' ? 'border-primary-500 text-primary-600 dark:text-primary-400' : 'border-transparent text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'"
          @click="activeTab = 'usage'"
        >
          {{ t('usage.tabs.usage') }}
        </button>
        <button
          type="button"
          class="border-b-2 px-4 py-2 text-sm font-medium transition"
          :class="activeTab === 'errors' ? 'border-primary-500 text-primary-600 dark:text-primary-400' : 'border-transparent text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'"
          @click="switchToErrors"
        >
          {{ t('usage.tabs.errors') }}
        </button>
      </div>

      <TablePageLayout>
      <template #actions>
        <div v-if="activeTab === 'usage'" class="flex w-full flex-wrap items-center gap-x-4 gap-y-1 rounded-lg border border-gray-200 bg-white/70 px-4 py-2 text-xs text-gray-600 shadow-sm dark:border-dark-700 dark:bg-dark-900/70 dark:text-gray-400">
          <span class="min-w-0 flex-1 truncate">{{ activeScopeSummary }}</span>
          <span class="flex items-center gap-1 whitespace-nowrap border-l border-gray-200 pl-4 text-gray-500 dark:border-dark-700 dark:text-gray-500">
            <span>{{ t('usage.totalTokens') }}</span>
            <span class="font-semibold text-gray-900 dark:text-white">{{ activeScopeTotalTokens }}</span>
          </span>
          <span class="flex items-center gap-1 whitespace-nowrap text-gray-500 dark:text-gray-500">
            <span>{{ t('usage.totalCost') }}</span>
            <span class="font-semibold text-[#a9583e] dark:text-[#f0b89e]" :title="activeScopeExactCost">{{ activeScopeTotalCost }}</span>
          </span>
        </div>
        <div v-else class="flex w-full items-center rounded-lg border border-gray-200 bg-white/70 px-4 py-2 text-xs text-gray-600 shadow-sm dark:border-dark-700 dark:bg-dark-900/70 dark:text-gray-400">
          {{ t('usage.errors.total', { count: errorTotal }) }}
        </div>
      </template>

      <template #filters>
        <div
          data-test="user-usage-filter-surface"
          class="card"
          :class="{ 'relative z-[221]': showColumnMenu }"
        >
          <div class="px-6 py-4">
          <div v-if="activeTab === 'usage'" class="flex flex-wrap items-end gap-4">
            <!-- API Key Filter -->
            <div class="min-w-[180px]">
              <label class="input-label">{{ t('usage.apiKeyFilter') }}</label>
              <Select
                v-model="filters.api_key_id"
                :options="apiKeyOptions"
                :placeholder="t('usage.allApiKeys')"
                @change="applyFilters"
              />
            </div>

            <!-- Group Filter -->
            <div class="min-w-[190px]">
              <label class="input-label">{{ t('usage.group') }}</label>
              <Select
                v-model="filters.group_id"
                :options="groupOptions"
                :placeholder="t('usage.allGroups')"
                @change="applyFilters"
              >
                <template #selected="{ option }">
                  <span class="truncate">{{ option?.label || t('usage.allGroups') }}</span>
                </template>
              </Select>
            </div>

            <!-- Date Range Filter -->
            <div>
              <label class="input-label">{{ t('usage.timeRange') }}</label>
              <DateRangePicker
                v-model:start-date="startDate"
                v-model:end-date="endDate"
                @change="onDateRangeChange"
              />
            </div>

            <!-- Actions -->
            <div class="ml-auto flex items-center gap-3">
              <button @click="showMoreFilters = !showMoreFilters" class="btn btn-secondary">
                <Icon name="filter" size="sm" class="mr-1.5" />
                {{ t('usage.moreFilters') }}
              </button>
              <button @click="applyFilters" :disabled="loading" class="btn btn-secondary">
                {{ t('common.refresh') }}
              </button>
              <button @click="resetFilters" class="btn btn-secondary">
                {{ t('common.reset') }}
              </button>
              <button @click="exportToCSV" :disabled="exporting" class="btn btn-primary">
                <svg
                  v-if="exporting"
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
                {{ exporting ? t('usage.exporting') : t('usage.exportCsv') }}
              </button>
              <div class="relative" ref="columnMenuRef">
                <button
                  data-test="user-usage-column-settings"
                  @click="showColumnMenu = !showColumnMenu"
                  class="btn btn-secondary px-2"
                  :title="t('usage.columnSettings')"
                >
                  <Icon name="cog" size="sm" />
                </button>
                <div
                  v-if="showColumnMenu"
                  class="absolute right-0 top-full z-50 mt-2 max-h-80 w-52 overflow-y-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-600 dark:bg-dark-800"
                >
                  <button
                    v-for="column in toggleableColumns"
                    :key="column.key"
                    @click="toggleColumn(column.key)"
                    class="flex w-full items-center justify-between px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
                  >
                    <span>{{ column.label }}</span>
                    <Icon v-if="isColumnVisible(column.key)" name="check" size="sm" class="text-primary-500" />
                  </button>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="flex flex-wrap items-end gap-4" data-test="user-error-filters">
            <div class="min-w-[180px]">
              <label class="input-label">{{ t('usage.errors.keyName') }}</label>
              <Select
                v-model="errorFilter.api_key_id"
                :options="errorKeyOptions"
                :placeholder="t('usage.errors.allKeys')"
                @change="applyErrorFilters"
              />
            </div>
            <div class="min-w-[200px]">
              <label class="input-label">{{ t('usage.errors.model') }}</label>
              <input
                v-model.trim="errorFilter.model"
                type="text"
                class="input"
                :placeholder="t('usage.errors.modelPlaceholder')"
                @keydown.enter="applyErrorFilters"
              />
            </div>
            <div class="min-w-[190px]">
              <label class="input-label">{{ t('usage.errors.category') }}</label>
              <Select
                v-model="errorFilter.category"
                :options="errorCategoryOptions"
                :placeholder="t('usage.errors.allCategories')"
                @change="applyErrorFilters"
              />
            </div>
            <div class="min-w-[170px]">
              <label class="input-label">{{ t('usage.errors.status') }}</label>
              <Select
                v-model="errorFilter.status_code"
                :options="errorStatusOptions"
                :placeholder="t('usage.errors.allStatuses')"
                @change="applyErrorFilters"
              />
            </div>
            <div>
              <label class="input-label">{{ t('usage.timeRange') }}</label>
              <DateRangePicker
                v-model:start-date="startDate"
                v-model:end-date="endDate"
                @change="onDateRangeChange"
              />
            </div>
            <div class="ml-auto flex items-center gap-3">
              <button type="button" class="btn btn-secondary" :disabled="errorLoading" @click="loadErrors">
                {{ t('common.refresh') }}
              </button>
              <button type="button" class="btn btn-secondary" @click="resetErrorFilters">
                {{ t('common.reset') }}
              </button>
              <div class="relative">
                <button
                  type="button"
                  class="btn btn-secondary px-2"
                  :title="t('usage.columnSettings')"
                  data-test="user-error-column-settings"
                  @click="showColumnMenu = !showColumnMenu"
                >
                  <Icon name="cog" size="sm" />
                </button>
                <div
                  v-if="showColumnMenu"
                  class="absolute right-0 top-full z-50 mt-2 max-h-80 w-52 overflow-y-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-600 dark:bg-dark-800"
                >
                  <button
                    v-for="column in errorToggleableColumns"
                    :key="column.key"
                    type="button"
                    class="flex w-full items-center justify-between px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
                    @click="toggleErrorColumn(column.key)"
                  >
                    <span>{{ column.label }}</span>
                    <Icon v-if="isErrorColumnVisible(column.key)" name="check" size="sm" class="text-primary-500" />
                  </button>
                </div>
              </div>
            </div>
          </div>
          <div v-if="activeTab === 'usage' && showMoreFilters" class="mt-4 grid gap-4 border-t border-gray-200 pt-4 dark:border-dark-700 md:grid-cols-3">
            <div>
              <label class="input-label">{{ t('usage.modelFilter') }}</label>
              <input
                v-model.trim="filters.model"
                type="text"
                class="input"
                :placeholder="t('usage.modelFilterPlaceholder')"
                @keydown.enter="applyFilters"
              />
            </div>
            <div>
              <label class="input-label">{{ t('usage.type') }}</label>
              <Select
                v-model="filters.request_type"
                :options="requestTypeOptions"
                :placeholder="t('usage.allTypes')"
                @change="applyFilters"
              />
            </div>
            <div>
              <label class="input-label">{{ t('admin.usage.billingMode') }}</label>
              <Select
                v-model="filters.billing_mode"
                :options="billingModeOptions"
                :placeholder="t('usage.allBillingModes')"
                @change="applyFilters"
              />
            </div>
          </div>
        </div>
        </div>
      </template>

      <template #table>
        <DataTable
          v-if="activeTab === 'usage'"
          :columns="visibleColumns"
          :data="usageLogs"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="created_at"
          default-sort-order="desc"
          @sort="handleSort"
        >
          <template #cell-api_key="{ row }">
            <span class="text-sm text-gray-900 dark:text-white">{{
              row.api_key?.name || '-'
            }}</span>
          </template>

          <template #cell-group="{ row }">
            <GroupBadge
              v-if="row.group"
              class="max-w-[180px]"
              :name="row.group.name"
              :platform="row.group.platform"
              :subscription-type="row.group.subscription_type"
              :rate-multiplier="pricingRateMultiplier(row)"
            />
            <span v-else-if="row.group_id" class="text-sm text-gray-500 dark:text-gray-400">#{{ row.group_id }}</span>
            <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
          </template>

          <template #cell-model="{ value, row }">
            <span class="font-medium text-gray-900 dark:text-white">{{ displayModelWithReasoningEffort(value, row.reasoning_effort) }}</span>
          </template>

          <template #cell-reasoning_effort="{ row }">
            <span class="text-sm text-gray-900 dark:text-white">
              {{ formatReasoningEffort(row.reasoning_effort) }}
            </span>
          </template>

          <template #cell-endpoint="{ row }">
            <span class="text-sm text-gray-600 dark:text-gray-300 block max-w-[320px] whitespace-normal break-all">
              {{ formatUsageEndpoints(row) }}
            </span>
          </template>

          <template #cell-stream="{ row }">
            <span
              class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium"
              :class="getRequestTypeBadgeClass(row)"
            >
              {{ getRequestTypeLabel(row) }}
            </span>
          </template>

          <template #cell-billing_mode="{ row }">
            <span class="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium"
                  :class="getBillingModeBadgeClass(getDisplayBillingMode(row))">
              {{ getBillingModeLabel(getDisplayBillingMode(row), t) }}
            </span>
          </template>

          <template #cell-tokens="{ row }">
            <!-- 图片生成请求 -->
            <div v-if="isImageUsage(row)" class="flex items-center gap-1.5">
              <svg
                class="h-4 w-4 text-[#a9583e] dark:text-[#f0b89e]"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                />
              </svg>
              <span class="font-medium text-gray-900 dark:text-white">{{ row.image_count }}{{ t('usage.imageUnit') }}</span>
              <span class="text-[#6c6a64] dark:text-gray-400">({{ formatImageBillingSize(row, t) }})</span>
            </div>
            <!-- Token 请求 -->
            <div v-else class="flex items-center gap-1.5">
              <div class="space-y-1.5 text-sm">
                <!-- Input / Output Tokens -->
                <div class="flex items-center gap-2">
                  <!-- Input -->
                  <div class="inline-flex items-center gap-1">
                    <Icon name="arrowDown" size="sm" class="text-[#8e8b82] dark:text-[#a9583e]" />
                    <span class="font-medium text-gray-900 dark:text-white">{{
                      formatNumber(row.input_tokens)
                    }}</span>
                  </div>
                  <!-- Output -->
                  <div class="inline-flex items-center gap-1">
                    <Icon name="arrowUp" size="sm" class="text-[#a9583e] dark:text-[#f0b89e]" />
                    <span class="font-medium text-gray-900 dark:text-white">{{
                      formatNumber(row.output_tokens)
                    }}</span>
                  </div>
                </div>
                <!-- Cache Write Tokens -->
                <div
                  v-if="hasPositiveNumber(row.cache_creation_tokens)"
                  class="flex items-center gap-2"
                >
                  <!-- Cache Write -->
                  <div class="inline-flex items-center gap-1">
                    <Icon name="edit" size="sm" class="text-amber-500" />
                    <span class="font-medium text-amber-600 dark:text-amber-400">{{
                      formatCacheTokens(toFiniteNumber(row.cache_creation_tokens))
                    }}</span>
                    <span v-if="hasPositiveNumber(row.cache_creation_1h_tokens)" class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-orange-100 text-orange-600 ring-1 ring-inset ring-orange-200 dark:bg-orange-500/20 dark:text-orange-400 dark:ring-orange-500/30">1h</span>
                    <span v-if="row.cache_ttl_overridden" :title="t('usage.cacheTtlOverriddenHint')" class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-rose-100 text-rose-600 ring-1 ring-inset ring-rose-200 dark:bg-rose-500/20 dark:text-rose-400 dark:ring-rose-500/30 cursor-help">R</span>
                  </div>
                </div>
                <div v-if="hasImageInputTokens(row)" class="flex items-center gap-2">
                  <div class="inline-flex items-center gap-1 text-fuchsia-600 dark:text-fuchsia-400">
                    <Icon name="image" size="sm" />
                    <span class="font-medium">{{ formatNumber(row.image_input_tokens) }}</span>
                  </div>
                </div>
                <div v-if="hasImageOutputTokens(row)" class="flex items-center gap-2">
                  <div class="inline-flex items-center gap-1 text-pink-600 dark:text-pink-400">
                    <Icon name="image" size="sm" />
                    <span class="font-medium">{{ formatNumber(row.image_output_tokens) }}</span>
                  </div>
                </div>
              </div>
              <!-- Token Detail Tooltip -->
              <div
                class="group relative"
                @mouseenter="showTokenTooltip($event, row)"
                @mouseleave="hideTokenTooltip"
              >
                <div
                  class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-[#f5f0e8] transition-colors group-hover:bg-[#f3e7df] dark:bg-gray-700 dark:group-hover:bg-[#cc785c]/15"
                >
                  <Icon
                    name="infoCircle"
                    size="xs"
                    class="text-[#8e8b82] group-hover:text-[#a9583e] dark:text-gray-500 dark:group-hover:text-[#f0b89e]"
                  />
                </div>
              </div>
            </div>
          </template>

          <template #cell-cache_read="{ row }">
            <div
              v-if="!isImageUsage(row) && hasPositiveNumber(row.cache_read_tokens)"
              class="inline-flex flex-col items-start gap-0.5 text-sm leading-tight"
              :title="`${formatNumber(row.cache_read_tokens)} (${formatCacheReadPercent(row)})`"
            >
              <div class="inline-flex items-center gap-1">
                <Icon name="database" size="sm" class="h-3.5 w-3.5 text-[#8e8b82] dark:text-[#f0b89e]" />
                <span class="font-medium text-[#6c6a64] dark:text-[#f0b89e]">
                  {{ formatCacheTokens(toFiniteNumber(row.cache_read_tokens)) }}
                </span>
              </div>
              <span class="text-xs font-medium text-[#8e8b82] dark:text-[#d8cec2]/80">
                {{ formatCacheReadPercent(row) }}
              </span>
            </div>
            <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
          </template>

          <template #cell-cost="{ row }">
            <div class="flex items-center gap-1.5 text-sm">
              <div class="flex flex-col items-start leading-tight">
                <span
                  class="text-xs text-gray-400 line-through dark:text-gray-500"
                  :title="`${t('usage.officialReferenceCost')}: ${formatOfficialReferenceCost(row.total_cost)}`"
                >
                  {{ formatOfficialReferenceCost(row.total_cost) }}
                </span>
                <span class="font-medium text-[#a9583e] dark:text-[#f0b89e]">
                  {{ formatCostFixed(row.actual_cost) }}
                </span>
              </div>
              <!-- Cost Detail Tooltip -->
              <div
                class="group relative"
                @mouseenter="showTooltip($event, row)"
                @mouseleave="hideTooltip"
              >
                <div
                  class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-[#f5f0e8] transition-colors group-hover:bg-[#f3e7df] dark:bg-gray-700 dark:group-hover:bg-[#cc785c]/15"
                >
                  <Icon
                    name="infoCircle"
                    size="xs"
                    class="text-[#8e8b82] group-hover:text-[#a9583e] dark:text-gray-500 dark:group-hover:text-[#f0b89e]"
                  />
                </div>
              </div>
            </div>
          </template>

          <template #cell-latency="{ row }">
            <div class="flex items-stretch gap-2">
              <span
                class="w-1 shrink-0 rounded-full"
                :class="row.first_token_ms != null
                  ? ['bg-gradient-to-b from-40% to-60%', LATENCY_BAR_FROM_CLASSES[firstTokenSeverity(row.first_token_ms)], LATENCY_BAR_TO_CLASSES[durationSeverity(row.duration_ms ?? 0)]]
                  : LATENCY_BAR_CLASSES[durationSeverity(row.duration_ms ?? 0)]"
                aria-hidden="true"
              ></span>
              <div class="grid grid-cols-[max-content_max-content] items-baseline gap-x-2 gap-y-0.5 text-xs">
                <span class="text-gray-400 dark:text-gray-500">{{ t('usage.latencyFirstToken') }}</span>
                <span v-if="row.first_token_ms != null" class="font-medium tabular-nums" :class="LATENCY_TEXT_CLASSES[firstTokenSeverity(row.first_token_ms)]">{{ formatDuration(row.first_token_ms) }}</span>
                <span v-else class="text-gray-400 dark:text-gray-500">-</span>
                <span class="text-gray-400 dark:text-gray-500">{{ t('usage.latencyDuration') }}</span>
                <span class="font-medium tabular-nums" :class="LATENCY_TEXT_CLASSES[durationSeverity(row.duration_ms ?? 0)]">{{ formatDuration(row.duration_ms) }}</span>
              </div>
            </div>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-600 dark:text-gray-400">{{
              formatDateTime(value)
            }}</span>
          </template>

          <template #cell-user_agent="{ row }">
            <span v-if="row.user_agent" class="block max-w-[260px] truncate text-sm text-gray-600 dark:text-gray-400" :title="row.user_agent">{{ formatUserAgent(row.user_agent) }}</span>
            <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
          </template>

          <template #cell-actions="{ row }">
            <button
              type="button"
              class="inline-flex h-8 w-8 items-center justify-center rounded-md text-gray-500 transition hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-white"
              :title="t('usage.details')"
              @click="openUsageDetails(row)"
            >
              <Icon name="eye" size="sm" />
            </button>
          </template>

          <template #empty>
            <EmptyState :message="t('usage.noRecords')" />
          </template>
        </DataTable>
        <UserErrorRequestsTable
          v-else-if="errorViewEnabled"
          :rows="errorRows"
          :loading="errorLoading"
          :visible-column-keys="errorVisibleColumnKeys"
          @sort="handleErrorSort"
        />
      </template>

      <template #pagination>
        <Pagination
          v-if="activeTab === 'usage' && pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
        <Pagination
          v-else-if="activeTab === 'errors' && errorTotal > 0"
          :page="errorPagination.page"
          :total="errorTotal"
          :page-size="errorPagination.page_size"
          @update:page="handleErrorPageChange"
          @update:pageSize="handleErrorPageSizeChange"
        />
      </template>
      </TablePageLayout>
    </div>
  </AppLayout>

  <!-- Token Tooltip Portal -->
  <Teleport to="body">
    <div
      v-if="tokenTooltipVisible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{
        left: tokenTooltipPosition.x + 'px',
        top: tokenTooltipPosition.y + 'px'
      }"
    >
      <div
        class="whitespace-nowrap rounded-lg border border-[#d8cec2] bg-[#fffaf5] px-3 py-2.5 text-xs text-[#141413] shadow-xl dark:border-gray-600 dark:bg-gray-800 dark:text-white"
      >
        <div class="space-y-1.5">
          <!-- Token Breakdown -->
          <div>
            <div class="text-xs font-semibold text-[#504f49] mb-1 dark:text-gray-300">{{ t('usage.tokenDetails') }}</div>
            <div v-if="tokenTooltipData && tokenTooltipData.input_tokens > 0 && !hasImageInputTokens(tokenTooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.inputTokens') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ formatNumber(tokenTooltipData.input_tokens) }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageInputTokens(tokenTooltipData) && textInputTokens(tokenTooltipData) > 0" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.inputTokens') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ formatNumber(textInputTokens(tokenTooltipData)) }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageInputTokens(tokenTooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageInputTokens') }}</span>
              <span class="font-medium text-fuchsia-600 dark:text-fuchsia-300">{{ formatNumber(tokenTooltipData.image_input_tokens) }}</span>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.output_tokens > 0 && !hasImageOutputTokens(tokenTooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.outputTokens') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ formatNumber(tokenTooltipData.output_tokens) }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageOutputTokens(tokenTooltipData) && textOutputTokens(tokenTooltipData) > 0" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.outputTokens') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ formatNumber(textOutputTokens(tokenTooltipData)) }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageOutputTokens(tokenTooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageOutputTokens') }}</span>
              <span class="font-medium text-pink-600 dark:text-pink-300">{{ formatNumber(tokenTooltipData.image_output_tokens) }}</span>
            </div>
            <div v-if="tokenTooltipData && hasPositiveNumber(tokenTooltipData.cache_creation_tokens)">
              <!-- 有 5m/1h 明细时，展开显示 -->
              <template v-if="hasPositiveNumber(tokenTooltipData.cache_creation_5m_tokens) || hasPositiveNumber(tokenTooltipData.cache_creation_1h_tokens)">
                <div v-if="hasPositiveNumber(tokenTooltipData.cache_creation_5m_tokens)" class="flex items-center justify-between gap-4">
                  <span class="text-[#6c6a64] flex items-center gap-1.5 dark:text-gray-400">
                    {{ t('admin.usage.cacheCreation5mTokens') }}
                    <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-amber-500/20 text-amber-400 ring-1 ring-inset ring-amber-500/30">5m</span>
                  </span>
                  <span class="font-medium text-[#141413] dark:text-white">{{ formatNumber(tokenTooltipData.cache_creation_5m_tokens) }}</span>
                </div>
                <div v-if="hasPositiveNumber(tokenTooltipData.cache_creation_1h_tokens)" class="flex items-center justify-between gap-4">
                  <span class="text-[#6c6a64] flex items-center gap-1.5 dark:text-gray-400">
                    {{ t('admin.usage.cacheCreation1hTokens') }}
                    <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-orange-500/20 text-orange-400 ring-1 ring-inset ring-orange-500/30">1h</span>
                  </span>
                  <span class="font-medium text-[#141413] dark:text-white">{{ formatNumber(tokenTooltipData.cache_creation_1h_tokens) }}</span>
                </div>
              </template>
              <!-- 无明细时，只显示聚合值 -->
              <div v-else class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.cacheCreationTokens') }}</span>
                <span class="font-medium text-[#141413] dark:text-white">{{ formatNumber(tokenTooltipData.cache_creation_tokens) }}</span>
              </div>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.cache_ttl_overridden" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] flex items-center gap-1.5 dark:text-gray-400">
                {{ t('usage.cacheTtlOverriddenLabel') }}
                <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-rose-500/20 text-rose-400 ring-1 ring-inset ring-rose-500/30">R-{{ hasPositiveNumber(tokenTooltipData.cache_creation_1h_tokens) ? '5m' : '1H' }}</span>
              </span>
              <span class="font-medium text-rose-400">{{ hasPositiveNumber(tokenTooltipData.cache_creation_1h_tokens) ? t('usage.cacheTtlOverridden1h') : t('usage.cacheTtlOverridden5m') }}</span>
            </div>
            <div v-if="tokenTooltipData && hasPositiveNumber(tokenTooltipData.cache_read_tokens)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.cacheReadTokens') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">
                {{ formatNumber(tokenTooltipData.cache_read_tokens) }} ({{ formatCacheReadPercent(tokenTooltipData) }})
              </span>
            </div>
          </div>
          <!-- Total -->
          <div class="flex items-center justify-between gap-6 border-t border-[#d8cec2] pt-1.5 dark:border-gray-700">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.totalTokens') }}</span>
            <span class="font-semibold text-[#a9583e] dark:text-[#f0b89e]">{{ formatNumber((tokenTooltipData?.input_tokens || 0) + (tokenTooltipData?.output_tokens || 0) + (tokenTooltipData?.cache_creation_tokens || 0) + (tokenTooltipData?.cache_read_tokens || 0)) }}</span>
          </div>
        </div>
        <!-- Tooltip Arrow (left side) -->
        <div
          class="absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-r-[#fffaf5] border-t-transparent dark:border-r-gray-800"
        ></div>
      </div>
    </div>
  </Teleport>

  <!-- Tooltip Portal -->
  <Teleport to="body">
    <div
      v-if="tooltipVisible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{
        left: tooltipPosition.x + 'px',
        top: tooltipPosition.y + 'px'
      }"
    >
      <div
        class="whitespace-nowrap rounded-lg border border-[#d8cec2] bg-[#fffaf5] px-3 py-2.5 text-xs text-[#141413] shadow-xl dark:border-gray-600 dark:bg-gray-800 dark:text-white"
      >
        <div class="space-y-1.5">
          <!-- Cost Breakdown -->
          <div class="mb-2 border-b border-[#d8cec2] pb-1.5 dark:border-gray-700">
            <div class="text-xs font-semibold text-[#504f49] mb-1 dark:text-gray-300">{{ t('usage.costDetails') }}</div>
            <div v-if="tooltipData && hasPositiveNumber(tooltipData.input_cost)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.inputCost') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ formatOfficialReferenceCost(tooltipData.input_cost) }}</span>
            </div>
            <div v-if="tooltipData && hasImageInputCost(tooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageInputCost') }}</span>
              <span class="font-medium text-fuchsia-600 dark:text-fuchsia-300">{{ formatOfficialReferenceCost(tooltipData.image_input_cost) }}</span>
            </div>
            <div v-if="tooltipData && hasPositiveNumber(tooltipData.output_cost)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.outputCost') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ formatOfficialReferenceCost(tooltipData.output_cost) }}</span>
            </div>
            <div v-if="tooltipData && hasImageOutputCost(tooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageOutputCost') }}</span>
              <span class="font-medium text-pink-600 dark:text-pink-300">{{ formatOfficialReferenceCost(tooltipData.image_output_cost) }}</span>
            </div>
            <!-- Per-image billing: show image metadata and unit price -->
            <template v-if="tooltipData && isImageUsage(tooltipData)">
              <div class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageCount') }}</span>
                <span class="font-medium text-[#141413] dark:text-white">{{ tooltipData.image_count }}{{ t('usage.imageUnit') }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageBillingSize') }}</span>
                <span class="font-medium text-[#141413] dark:text-white">{{ formatImageBillingSize(tooltipData, t) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageSizeSource') }}</span>
                <span class="font-medium text-[#141413] dark:text-white">{{ formatImageSizeSource(tooltipData, t) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageInputSize') }}</span>
                <span class="font-medium text-[#141413] dark:text-white">{{ formatImageInputSize(tooltipData, t) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageOutputSize') }}</span>
                <span class="font-medium text-[#141413] dark:text-white">{{ formatImageOutputSize(tooltipData, t) }}</span>
              </div>
              <div v-if="formatImageSizeBreakdown(tooltipData)" class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageSizeBreakdown') }}</span>
                <span class="font-medium text-[#141413] dark:text-white">{{ formatImageSizeBreakdown(tooltipData) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageUnitPrice') }}</span>
                <span class="font-medium text-[#a9583e] dark:text-[#f0b89e]">{{ formatOfficialReferenceCost(imageUnitPrice(tooltipData)) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageTotalPrice') }}</span>
                <span class="font-medium text-[#141413] dark:text-white">{{ formatOfficialReferenceCost(tooltipData.total_cost) }}</span>
              </div>
            </template>
            <!-- Token billing: show unit prices per 1M tokens -->
            <template v-else-if="!getDisplayBillingMode(tooltipData) || getDisplayBillingMode(tooltipData) === BILLING_MODE_TOKEN">
              <div v-if="tooltipData && textInputTokens(tooltipData) > 0" class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.inputTokenPrice') }}</span>
                <span class="font-medium text-[#6c6a64] dark:text-[#f0b89e]">{{ formatTokenPricePerMillion(tooltipData.input_cost, textInputTokens(tooltipData)) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && hasImageInputTokens(tooltipData)" class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageInputTokenPrice') }}</span>
                <span class="font-medium text-fuchsia-600 dark:text-fuchsia-300">{{ formatTokenPricePerMillion(tooltipData.image_input_cost, tooltipData.image_input_tokens) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && tooltipData.output_cost > 0 && textOutputTokens(tooltipData) > 0" class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.outputTokenPrice') }}</span>
                <span class="font-medium text-[#a9583e] dark:text-[#f0b89e]">{{ formatTokenPricePerMillion(tooltipData.output_cost, textOutputTokens(tooltipData)) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
              <div v-if="tooltipData && hasImageOutputTokens(tooltipData)" class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageOutputTokenPrice') }}</span>
                <span class="font-medium text-pink-600 dark:text-pink-300">{{ formatTokenPricePerMillion(tooltipData.image_output_cost, tooltipData.image_output_tokens) }} {{ t('usage.perMillionTokens') }}</span>
              </div>
            </template>
            <div v-else class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.unitPrice') }}</span>
              <span class="font-medium text-[#a9583e] dark:text-[#f0b89e]">{{ formatOfficialReferenceCost(tooltipData?.total_cost) }}</span>
            </div>
            <div v-if="tooltipData && hasPositiveNumber(tooltipData.cache_creation_cost)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.cacheCreationCost') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ formatOfficialReferenceCost(tooltipData.cache_creation_cost) }}</span>
            </div>
            <div v-if="tooltipData && hasPositiveNumber(tooltipData.cache_read_cost)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.cacheReadCost') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ formatOfficialReferenceCost(tooltipData.cache_read_cost) }}</span>
            </div>
          </div>
          <!-- Rate and Summary -->
          <div class="flex items-center justify-between gap-6">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.serviceTier') }}</span>
            <span class="font-semibold text-[#6c6a64] dark:text-[#f0b89e]">{{ getUsageServiceTierLabel(tooltipData?.service_tier, t) }}</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.pricingRate') }}</span>
            <span class="font-semibold text-[#a9583e] dark:text-[#f0b89e]"
              >{{ formatMultiplier(pricingRateMultiplier(tooltipData)) }}x</span
            >
          </div>
          <div v-if="balanceConversionMultiplier(tooltipData) !== 1" class="flex items-center justify-between gap-6">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.balanceConversion') }}</span>
            <span class="font-semibold text-[#a9583e] dark:text-[#f0b89e]">{{ formatMultiplier(balanceConversionMultiplier(tooltipData)) }}x</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.officialReferenceCost') }}</span>
            <span class="font-medium text-[#141413] dark:text-white">{{ formatOfficialReferenceCost(tooltipData?.total_cost) }}</span>
          </div>
          <div class="flex items-center justify-between gap-6 border-t border-[#d8cec2] pt-1.5 dark:border-gray-700">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.billed') }}</span>
            <span class="font-semibold text-[#a9583e] dark:text-[#f0b89e]"
              >{{ formatCostFixed(tooltipData?.actual_cost) }}</span
            >
          </div>
        </div>
        <!-- Tooltip Arrow (left side) -->
        <div
          class="absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-r-[#fffaf5] border-t-transparent dark:border-r-gray-800"
        ></div>
      </div>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="detailsVisible && selectedUsageLog"
      class="fixed inset-0 z-[9998] flex items-stretch justify-end bg-black/40"
      @click.self="closeUsageDetails"
    >
      <aside class="h-full w-full max-w-2xl overflow-y-auto border-l border-gray-200 bg-white shadow-2xl dark:border-dark-700 dark:bg-dark-900">
        <div class="sticky top-0 z-10 flex items-start justify-between border-b border-gray-200 bg-white px-6 py-4 dark:border-dark-700 dark:bg-dark-900">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('usage.details') }}</h2>
            <p class="mt-1 max-w-xl break-all text-xs text-gray-500 dark:text-gray-400">
              {{ selectedUsageLog.request_id || '-' }}
            </p>
          </div>
          <button
            type="button"
            class="rounded-md p-2 text-gray-500 transition hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-white"
            @click="closeUsageDetails"
          >
            <Icon name="x" size="sm" />
          </button>
        </div>

        <div class="space-y-5 px-6 py-5">
          <section class="space-y-3">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('usage.routeInfo') }}</h3>
            <div class="grid gap-3 rounded-lg border border-gray-200 p-4 text-sm dark:border-dark-700 md:grid-cols-2">
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.apiKeyFilter') }}</div>
                <div class="mt-1 font-medium text-gray-900 dark:text-white">{{ selectedUsageLog.api_key?.name || '-' }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.group') }}</div>
                <div class="mt-1">
                  <GroupBadge
                    v-if="selectedUsageLog.group"
                    class="max-w-full"
                    :name="selectedUsageLog.group.name"
                    :platform="selectedUsageLog.group.platform"
                    :subscription-type="selectedUsageLog.group.subscription_type"
                    :rate-multiplier="pricingRateMultiplier(selectedUsageLog)"
                  />
                  <span v-else-if="selectedUsageLog.group_id" class="text-gray-600 dark:text-gray-300">#{{ selectedUsageLog.group_id }}</span>
                  <span v-else class="text-gray-400">-</span>
                </div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">Group ID</div>
                <div class="mt-1 font-mono text-gray-900 dark:text-white">{{ selectedUsageLog.group_id ?? '-' }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.pricingRate') }}</div>
                <div class="mt-1 font-medium text-gray-900 dark:text-white">{{ formatMultiplier(pricingRateMultiplier(selectedUsageLog)) }}x</div>
              </div>
              <div v-if="balanceConversionMultiplier(selectedUsageLog) !== 1">
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.balanceConversion') }}</div>
                <div class="mt-1 font-medium text-gray-900 dark:text-white">{{ formatMultiplier(balanceConversionMultiplier(selectedUsageLog)) }}x</div>
              </div>
            </div>
          </section>

          <section class="space-y-3">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('usage.requestInfo') }}</h3>
            <div class="grid gap-3 rounded-lg border border-gray-200 p-4 text-sm dark:border-dark-700 md:grid-cols-2">
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">Request ID</div>
                <div class="mt-1 break-all font-mono text-gray-900 dark:text-white">{{ selectedUsageLog.request_id || '-' }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.model') }}</div>
                <div class="mt-1 break-all font-medium text-gray-900 dark:text-white">{{ displayModelWithReasoningEffort(selectedUsageLog.model, selectedUsageLog.reasoning_effort) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.endpoint') }}</div>
                <div class="mt-1 break-all text-gray-900 dark:text-white">{{ formatUsageEndpoints(selectedUsageLog) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.type') }}</div>
                <div class="mt-1">{{ getRequestTypeLabel(selectedUsageLog) }}</div>
              </div>
              <div class="md:col-span-2">
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.userAgent') }}</div>
                <div class="mt-1 break-all text-gray-900 dark:text-white">{{ selectedUsageLog.user_agent || '-' }}</div>
              </div>
            </div>
          </section>

          <section class="space-y-3">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('usage.tokenDetails') }}</h3>
            <div class="grid gap-3 rounded-lg border border-gray-200 p-4 text-sm dark:border-dark-700 md:grid-cols-2">
              <div>{{ t('usage.in') }}: <span class="font-medium">{{ formatNumber(textInputTokens(selectedUsageLog)) }}</span></div>
              <div>{{ t('usage.out') }}: <span class="font-medium">{{ formatNumber(textOutputTokens(selectedUsageLog)) }}</span></div>
              <div v-if="hasImageInputTokens(selectedUsageLog)">{{ t('usage.imageInputTokens') }}: <span class="font-medium">{{ formatNumber(selectedUsageLog.image_input_tokens) }}</span></div>
              <div v-if="hasImageOutputTokens(selectedUsageLog)">{{ t('usage.imageOutputTokens') }}: <span class="font-medium">{{ formatNumber(selectedUsageLog.image_output_tokens) }}</span></div>
              <div>
                {{ t('usage.cacheRead') }}:
                <span class="font-medium">
                  {{ formatNumber(selectedUsageLog.cache_read_tokens) }} ({{ formatCacheReadPercent(selectedUsageLog) }})
                </span>
              </div>
              <div>{{ t('usage.cacheWrite') }}: <span class="font-medium">{{ formatNumber(selectedUsageLog.cache_creation_tokens) }}</span></div>
              <div v-if="isImageUsage(selectedUsageLog)">
                {{ t('usage.imageCount') }}: <span class="font-medium">{{ selectedUsageLog.image_count }}{{ t('usage.imageUnit') }}</span>
              </div>
              <div v-if="isImageUsage(selectedUsageLog)">
                {{ t('usage.imageBillingSize') }}: <span class="font-medium">{{ formatImageBillingSize(selectedUsageLog, t) }}</span>
              </div>
            </div>
          </section>

          <section class="space-y-3">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('usage.costDetails') }}</h3>
            <div class="grid gap-3 rounded-lg border border-gray-200 p-4 text-sm dark:border-dark-700 md:grid-cols-2">
              <div>{{ t('usage.billed') }}: <span class="font-medium text-[#a9583e] dark:text-[#f0b89e]">{{ formatCostFixed(selectedUsageLog.actual_cost) }}</span></div>
              <div>{{ t('usage.officialReferenceCost') }}: <span class="font-medium">{{ formatOfficialReferenceCost(selectedUsageLog.total_cost) }}</span></div>
              <div>{{ t('admin.usage.inputCost') }}: <span class="font-medium">{{ formatOfficialReferenceCost(selectedUsageLog.input_cost) }}</span></div>
              <div>{{ t('admin.usage.outputCost') }}: <span class="font-medium">{{ formatOfficialReferenceCost(selectedUsageLog.output_cost) }}</span></div>
              <div v-if="hasImageInputCost(selectedUsageLog)">{{ t('usage.imageInputCost') }}: <span class="font-medium">{{ formatOfficialReferenceCost(selectedUsageLog.image_input_cost) }}</span></div>
              <div v-if="hasImageOutputCost(selectedUsageLog)">{{ t('usage.imageOutputCost') }}: <span class="font-medium">{{ formatOfficialReferenceCost(selectedUsageLog.image_output_cost) }}</span></div>
              <div>{{ t('usage.firstToken') }}: <span class="font-medium">{{ selectedUsageLog.first_token_ms != null ? formatDuration(selectedUsageLog.first_token_ms) : '-' }}</span></div>
              <div>{{ t('usage.duration') }}: <span class="font-medium">{{ formatDuration(selectedUsageLog.duration_ms) }}</span></div>
            </div>
          </section>
        </div>
      </aside>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { usageAPI, keysAPI, userGroupsAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import UserSubscriptionsPanel from '@/components/user/UserSubscriptionsPanel.vue'
import UserErrorRequestsTable from '@/components/user/UserErrorRequestsTable.vue'
import UsageStatsCards from '@/components/admin/usage/UsageStatsCards.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import GroupDistributionChart from '@/components/charts/GroupDistributionChart.vue'
import EndpointDistributionChart from '@/components/charts/EndpointDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import type {
  UsageLog,
  ApiKey,
  Group,
  UsageQueryParams,
  UsageStatsResponse,
  TrendDataPoint,
  ModelStat,
  GroupStat,
  EndpointStat,
  UserErrorRequest,
  UserErrorListParams,
} from '@/types'
import type { Column } from '@/components/common/types'
import { formatDateTime, formatReasoningEffort } from '@/utils/format'
import { formatCreditAmount } from '@/utils/credits'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { formatCacheTokens, formatMultiplier } from '@/utils/formatters'
import { formatTokenPricePerMillion } from '@/utils/usagePricing'
import {
  LATENCY_BAR_CLASSES,
  LATENCY_BAR_FROM_CLASSES,
  LATENCY_BAR_TO_CLASSES,
  LATENCY_TEXT_CLASSES,
  durationSeverity,
  firstTokenSeverity,
} from '@/utils/latencyHealth'
import { getUsageServiceTierLabel } from '@/utils/usageServiceTier'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import {
  BILLING_MODE_IMAGE,
  BILLING_MODE_PER_REQUEST,
  BILLING_MODE_TOKEN,
  getBillingModeBadgeClass,
  getBillingModeLabel,
} from '@/utils/billingMode'
import {
  formatImageBillingSize,
  formatImageInputSize,
  formatImageOutputSize,
  formatImageSizeBreakdown,
  formatImageSizeSource,
  hasImageInputTokens,
  textInputTokens,
  hasImageInputCost,
  hasImageOutputTokens,
  textOutputTokens,
  hasImageOutputCost,
} from '@/utils/imageUsage'
import { displayModelLabel, displayModelWithReasoningEffort } from '@/utils/modelDisplay'
import { COMMON_ERROR_STATUS_CODES } from '@/utils/errorBadges'

const { t } = useI18n()
const appStore = useAppStore()

let abortController: AbortController | null = null
let errorRequestSequence = 0
let statsRequestSequence = 0
let modelRequestSequence = 0
let chartRequestSequence = 0

// Tooltip state
const tooltipVisible = ref(false)
const tooltipPosition = ref({ x: 0, y: 0 })
const tooltipData = ref<UsageLog | null>(null)

// Token tooltip state
const tokenTooltipVisible = ref(false)
const tokenTooltipPosition = ref({ x: 0, y: 0 })
const tokenTooltipData = ref<UsageLog | null>(null)

// Usage stats from API
const usageStats = ref<UsageStatsResponse | null>(null)
const trendData = ref<TrendDataPoint[]>([])
const requestedModelStats = ref<ModelStat[]>([])
const groupStats = ref<GroupStat[]>([])
const endpointStats = ref<EndpointStat[]>([])
const chartsLoading = ref(false)
const modelStatsLoading = ref(false)
const endpointStatsLoading = ref(false)
const modelDistributionMetric = ref<'tokens' | 'actual_cost'>('tokens')
const groupDistributionMetric = ref<'tokens' | 'actual_cost'>('tokens')
const endpointDistributionMetric = ref<'tokens' | 'actual_cost'>('tokens')
const granularity = ref<'day' | 'hour'>('day')
const activeTab = ref<'usage' | 'errors'>('usage')
const errorViewEnabled = computed(
  () => appStore.cachedPublicSettings?.allow_user_view_error_requests === true
)

const errorRows = ref<UserErrorRequest[]>([])
const errorLoading = ref(false)
const errorTotal = ref(0)
const errorPagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
})
const errorSort = reactive<{
  sort_by: 'created_at' | 'model' | 'status_code'
  sort_order: 'asc' | 'desc'
}>({
  sort_by: 'created_at',
  sort_order: 'desc',
})
const errorFilter = ref<{
  api_key_id: number | null
  model: string
  category: string | null
  status_code: number | null
}>({
  api_key_id: null,
  model: '',
  category: null,
  status_code: null,
})

watch(errorViewEnabled, (enabled) => {
  if (enabled) return
  activeTab.value = 'usage'
  errorRequestSequence += 1
  errorRows.value = []
  errorTotal.value = 0
  errorLoading.value = false
})

const COLUMN_VISIBILITY_KEY = 'usage-visible-columns:v3'
const LEGACY_COLUMN_VISIBILITY_KEYS = ['usage-visible-columns:v2', 'usage-visible-columns:v1']
const MIGRATED_DEFAULT_VISIBLE_COLUMNS = ['cache_read', 'latency']
const DEFAULT_VISIBLE_COLUMNS = [
  'api_key',
  'group',
  'model',
  'stream',
  'billing_mode',
  'tokens',
  'cache_read',
  'cost',
  'latency',
  'created_at',
  'actions'
]
const ALWAYS_VISIBLE_COLUMNS = ['api_key', 'created_at', 'actions']

const allColumns = computed<Column[]>(() => [
  { key: 'api_key', label: t('usage.apiKeyFilter'), sortable: false },
  { key: 'group', label: t('usage.group'), sortable: false },
  { key: 'model', label: t('usage.model'), sortable: true },
  { key: 'reasoning_effort', label: t('usage.reasoningEffort'), sortable: false },
  { key: 'endpoint', label: t('usage.endpoint'), sortable: false },
  { key: 'stream', label: t('usage.type'), sortable: false },
  { key: 'billing_mode', label: t('admin.usage.billingMode'), sortable: false },
  { key: 'tokens', label: t('usage.tokens'), sortable: false },
  { key: 'cache_read', label: t('usage.cacheRead'), sortable: false },
  { key: 'cost', label: t('usage.cost'), sortable: false },
  { key: 'latency', label: t('usage.latency'), sortable: false },
  { key: 'created_at', label: t('usage.time'), sortable: true },
  { key: 'user_agent', label: t('usage.userAgent'), sortable: false },
  { key: 'actions', label: t('usage.details'), sortable: false }
])

const visibleColumnKeys = ref<Set<string>>(new Set(DEFAULT_VISIBLE_COLUMNS))
const showColumnMenu = ref(false)
const columnMenuRef = ref<HTMLElement | null>(null)
const showMoreFilters = ref(false)
const selectedUsageLog = ref<UsageLog | null>(null)
const detailsVisible = ref(false)

const visibleColumns = computed<Column[]>(() =>
  allColumns.value.filter((column) => visibleColumnKeys.value.has(column.key))
)

const toggleableColumns = computed(() =>
  allColumns.value.filter((column) => !ALWAYS_VISIBLE_COLUMNS.includes(column.key))
)

const ERROR_COLUMN_VISIBILITY_KEY = 'user-error-visible-columns:v1'
const ERROR_DEFAULT_VISIBLE_COLUMNS = [
  'key_name',
  'model',
  'endpoint',
  'platform',
  'category',
  'status',
  'message',
  'created_at',
]
const ERROR_ALWAYS_VISIBLE_COLUMNS = ['status', 'created_at']
const errorVisibleColumnSet = ref<Set<string>>(new Set(ERROR_DEFAULT_VISIBLE_COLUMNS))
const errorColumns = computed<Column[]>(() => [
  { key: 'key_name', label: t('usage.errors.keyName'), sortable: false },
  { key: 'model', label: t('usage.errors.model'), sortable: true },
  { key: 'endpoint', label: t('usage.errors.endpoint'), sortable: false },
  { key: 'platform', label: t('usage.errors.platform'), sortable: false },
  { key: 'category', label: t('usage.errors.category'), sortable: false },
  { key: 'status', label: t('usage.errors.status'), sortable: true },
  { key: 'message', label: t('usage.errors.message'), sortable: false },
  { key: 'created_at', label: t('usage.errors.time'), sortable: true },
])
const errorToggleableColumns = computed(() =>
  errorColumns.value.filter((column) => !ERROR_ALWAYS_VISIBLE_COLUMNS.includes(column.key))
)
const errorVisibleColumnKeys = computed(() =>
  errorColumns.value
    .filter((column) => errorVisibleColumnSet.value.has(column.key))
    .map((column) => column.key)
)

const isErrorColumnVisible = (key: string): boolean => errorVisibleColumnSet.value.has(key)

const persistErrorVisibleColumns = (): void => {
  try {
    localStorage.setItem(ERROR_COLUMN_VISIBILITY_KEY, JSON.stringify([...errorVisibleColumnSet.value]))
  } catch (error) {
    console.error('Failed to save user error visible columns:', error)
  }
}

const toggleErrorColumn = (key: string): void => {
  if (ERROR_ALWAYS_VISIBLE_COLUMNS.includes(key)) return
  const next = new Set(errorVisibleColumnSet.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  ERROR_ALWAYS_VISIBLE_COLUMNS.forEach((column) => next.add(column))
  errorVisibleColumnSet.value = next
  persistErrorVisibleColumns()
}

const loadErrorVisibleColumns = (): void => {
  try {
    const raw = localStorage.getItem(ERROR_COLUMN_VISIBILITY_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw) as string[]
    const valid = new Set(errorColumns.value.map((column) => column.key))
    const next = new Set(parsed.filter((key) => valid.has(key)))
    ERROR_ALWAYS_VISIBLE_COLUMNS.forEach((key) => next.add(key))
    errorVisibleColumnSet.value = next
  } catch (error) {
    console.error('Failed to load user error visible columns:', error)
  }
}

const usageLogs = ref<UsageLog[]>([])
const apiKeys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
const loading = ref(false)
const exporting = ref(false)

const apiKeyOptions = computed(() => {
  return [
    { value: null, label: t('usage.allApiKeys') },
    ...apiKeys.value.map((key) => ({
      value: key.id,
      label: key.name
    }))
  ]
})

const errorKeyOptions = computed(() => [
  { value: null, label: t('usage.errors.allKeys') },
  ...apiKeys.value.map((key) => ({ value: key.id, label: key.name })),
])

const errorCategoryOptions = computed(() => [
  { value: null, label: t('usage.errors.allCategories') },
  ...[
    'auth',
    'rate_limit',
    'quota',
    'invalid_request',
    'service_unavailable',
    'upstream',
    'internal',
    'cyber',
  ].map((category) => ({
    value: category,
    label: t(`usage.errors.categories.${category}`),
  })),
])

const errorStatusOptions = computed(() => [
  { value: null, label: t('usage.errors.allStatuses') },
  ...COMMON_ERROR_STATUS_CODES.map((status) => ({ value: status, label: String(status) })),
])

const groupOptions = computed(() => [
  { value: null, label: t('usage.allGroups') },
  ...groups.value.map((group) => ({
    value: group.id,
    label: group.name
  }))
])

const requestTypeOptions = computed(() => [
  { value: null, label: t('usage.allTypes') },
  { value: 'sync', label: t('usage.sync') },
  { value: 'stream', label: t('usage.stream') },
  { value: 'ws_v2', label: t('usage.ws') },
  { value: 'live', label: t('usage.live') },
  { value: 'cyber', label: t('usage.cyber') }
])

const billingModeOptions = computed(() => [
  { value: null, label: t('usage.allBillingModes') },
  { value: BILLING_MODE_TOKEN, label: getBillingModeLabel(BILLING_MODE_TOKEN, t) },
  { value: BILLING_MODE_PER_REQUEST, label: getBillingModeLabel(BILLING_MODE_PER_REQUEST, t) },
  { value: BILLING_MODE_IMAGE, label: getBillingModeLabel(BILLING_MODE_IMAGE, t) }
])

const granularityOptions = computed(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') },
])

// Helper function to format date in local timezone
const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

// Initialize date range immediately
const now = new Date()
const weekAgo = new Date(now)
weekAgo.setDate(weekAgo.getDate() - 6)

// Date range state
const startDate = ref(formatLocalDate(weekAgo))
const endDate = ref(formatLocalDate(now))

const filters = ref<UsageQueryParams>({
  api_key_id: undefined,
  group_id: undefined,
  model: undefined,
  request_type: undefined,
  billing_mode: undefined,
  start_date: undefined,
  end_date: undefined
})

// Initialize filters with date range
filters.value.start_date = startDate.value
filters.value.end_date = endDate.value

// Handle date range change from DateRangePicker
const onDateRangeChange = (range: {
  startDate: string
  endDate: string
  preset: string | null
}) => {
  filters.value.start_date = range.startDate
  filters.value.end_date = range.endDate
  applyFilters()
}

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})
const sortState = reactive({
  sort_by: 'created_at',
  sort_order: 'desc' as 'asc' | 'desc'
})

const toFiniteNumber = (value: unknown, fallback = 0): number => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : fallback
}

const hasPositiveNumber = (value: unknown): boolean => toFiniteNumber(value) > 0

const formatNumber = (value: unknown): string => toFiniteNumber(value).toLocaleString()

const usageInputTokenTotal = (
  row: Pick<UsageLog, 'input_tokens' | 'cache_creation_tokens' | 'cache_read_tokens'>
): number =>
  toFiniteNumber(row.input_tokens)
  + toFiniteNumber(row.cache_creation_tokens)
  + toFiniteNumber(row.cache_read_tokens)

const formatCacheReadPercent = (
  row: Pick<UsageLog, 'input_tokens' | 'output_tokens' | 'cache_creation_tokens' | 'cache_read_tokens'>
): string => {
  const total = usageInputTokenTotal(row)
  if (total <= 0) return '0%'
  const percent = (toFiniteNumber(row.cache_read_tokens) / total) * 100
  if (percent >= 99.95 && percent < 100) return '99.9%'
  return `${percent.toFixed(1)}%`
}

const formatCostNumberFixed = (value: unknown, digits = 6): string => toFiniteNumber(value).toFixed(digits)
const formatCostFixed = (value: unknown, digits = 6): string =>
  formatCreditAmount(toFiniteNumber(value), {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  })

const formatOfficialReferenceCost = (value: unknown, digits = 6): string =>
  `$${formatCostNumberFixed(value, digits)}`

const formatDuration = (ms: number | null | undefined): string => {
  if (ms == null) return '-'
  const safeMs = toFiniteNumber(ms)
  if (safeMs < 1000) return `${safeMs.toFixed(0)}ms`
  if (safeMs < 60_000) return `${(safeMs / 1000).toFixed(2)}s`
  const totalSec = Math.round(safeMs / 1000)
  if (totalSec < 3600) return `${Math.floor(totalSec / 60)}m ${totalSec % 60}s`
  return `${Math.floor(totalSec / 3600)}h ${Math.floor((totalSec % 3600) / 60)}m`
}

const imageUnitPrice = (row: UsageLog | null): number => {
  const imageCount = toFiniteNumber(row?.image_count)
  if (!row || imageCount <= 0) return 0
  const price = toFiniteNumber(row.total_cost) / imageCount
  return Number.isFinite(price) ? price : 0
}

const getDisplayBillingMode = (row: Pick<UsageLog, 'billing_mode' | 'image_count'> | null | undefined): string | null | undefined => {
  if (row?.billing_mode) {
    return row.billing_mode
  }
  if (hasPositiveNumber(row?.image_count)) {
    return BILLING_MODE_IMAGE
  }
  return row?.billing_mode
}

const isImageUsage = (row: Pick<UsageLog, 'billing_mode' | 'image_count'> | null | undefined): boolean => {
  const mode = getDisplayBillingMode(row)
  return mode === BILLING_MODE_IMAGE || mode === BILLING_MODE_PER_REQUEST
}

const formatUserAgent = (ua: string): string => {
  return ua
}

const selectedAPIKeyLabel = computed(() => {
  const id = filters.value.api_key_id
  if (!id) return t('usage.allApiKeys')
  return apiKeys.value.find((key) => key.id === Number(id))?.name || `#${id}`
})

const selectedGroupLabel = computed(() => {
  const id = filters.value.group_id
  if (!id) return t('usage.allGroups')
  return groups.value.find((group) => group.id === Number(id))?.name || `#${id}`
})

const selectedRequestTypeLabel = computed(() => {
  const requestType = filters.value.request_type
  if (!requestType) return t('usage.allTypes')
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
})

const activeScopeSummary = computed(() => {
  const parts = [
    `${startDate.value} - ${endDate.value}`,
    selectedAPIKeyLabel.value,
    selectedGroupLabel.value
  ]
  if (filters.value.model) parts.push(displayModelLabel(filters.value.model))
  if (filters.value.request_type) parts.push(selectedRequestTypeLabel.value)
  if (filters.value.billing_mode) parts.push(getBillingModeLabel(filters.value.billing_mode, t))
  return parts.join(' / ')
})

const activeScopeTotalTokens = computed(() => formatNumber(usageStats.value?.total_tokens ?? 0))

const activeScopeExactCost = computed(() => formatCostFixed(usageStats.value?.total_actual_cost ?? 0, 6))

const activeScopeTotalCost = computed(() =>
  formatCreditAmount(usageStats.value?.total_actual_cost ?? 0, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
)

const isColumnVisible = (key: string): boolean => visibleColumnKeys.value.has(key)

const normalizeVisibleColumnKeys = (keys: Iterable<string>): Set<string> => {
  const valid = new Set(allColumns.value.map((column) => column.key))
  const next = new Set<string>()
  for (const key of keys) {
    if (valid.has(key)) {
      next.add(key)
    }
  }
  ALWAYS_VISIBLE_COLUMNS.forEach((key) => {
    if (valid.has(key)) {
      next.add(key)
    }
  })
  return next.size > 0 ? next : new Set(DEFAULT_VISIBLE_COLUMNS)
}

const persistVisibleColumns = () => {
  try {
    localStorage.setItem(COLUMN_VISIBILITY_KEY, JSON.stringify([...visibleColumnKeys.value]))
  } catch (error) {
    console.error('Failed to save usage visible columns:', error)
  }
}

const loadVisibleColumns = () => {
  try {
    let raw = localStorage.getItem(COLUMN_VISIBILITY_KEY)
    let migratedFromLegacy = false
    if (!raw) {
      for (const key of LEGACY_COLUMN_VISIBILITY_KEYS) {
        const legacyRaw = localStorage.getItem(key)
        if (legacyRaw) {
          raw = legacyRaw
          break
        }
      }
      migratedFromLegacy = Boolean(raw)
    }
    if (!raw) return
    const parsed = JSON.parse(raw) as string[]
    const valid = new Set(allColumns.value.map((column) => column.key))
    const next = parsed.filter((key) => valid.has(key))
    if (next.length > 0) {
      const nextKeys = normalizeVisibleColumnKeys(next)
      if (migratedFromLegacy) {
        MIGRATED_DEFAULT_VISIBLE_COLUMNS.forEach((key) => {
          if (valid.has(key)) {
            nextKeys.add(key)
          }
        })
      }
      visibleColumnKeys.value = nextKeys
      if (migratedFromLegacy) {
        persistVisibleColumns()
      }
    }
  } catch (error) {
    console.error('Failed to load usage visible columns:', error)
  }
}

const toggleColumn = (key: string) => {
  if (ALWAYS_VISIBLE_COLUMNS.includes(key)) return
  const next = new Set(visibleColumnKeys.value)
  if (next.has(key)) {
    next.delete(key)
  } else {
    next.add(key)
  }
  visibleColumnKeys.value = normalizeVisibleColumnKeys(next)
  persistVisibleColumns()
}

const getRequestTypeLabel = (log: UsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'cyber') return t('usage.cyber')
  if (requestType === 'live') return t('usage.live')
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
}

const getRequestTypeBadgeClass = (log: UsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'cyber') return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  if (requestType === 'live') return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200'
  if (requestType === 'ws_v2') return 'bg-[#f3e7df] text-[#a9583e] dark:bg-[#cc785c]/15 dark:text-[#f0b89e]'
  if (requestType === 'stream') return 'bg-[#fffaf5] text-[#a9583e] ring-1 ring-[#d8cec2] dark:bg-[#cc785c]/15 dark:text-[#f0b89e] dark:ring-[#cc785c]/30'
  if (requestType === 'sync') return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
  return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200'
}


const getRequestTypeExportText = (log: UsageLog): string => {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'cyber') return 'Cyber'
  if (requestType === 'live') return 'Live'
  if (requestType === 'ws_v2') return 'WS'
  if (requestType === 'stream') return 'Stream'
  if (requestType === 'sync') return 'Sync'
  return 'Unknown'
}

const formatUsageEndpoints = (log: UsageLog): string => {
  const inbound = log.inbound_endpoint?.trim()
  return inbound || '-'
}

type UsageTableQueryParams = UsageQueryParams & {
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

const buildUsageQueryParams = (page: number, pageSize: number): UsageTableQueryParams => {
  const params: UsageTableQueryParams = {
    page,
    page_size: pageSize,
    start_date: filters.value.start_date,
    end_date: filters.value.end_date,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    sort_by: sortState.sort_by,
    sort_order: sortState.sort_order
  }

  if (filters.value.api_key_id) {
    params.api_key_id = Number(filters.value.api_key_id)
  }
  if (filters.value.group_id) {
    params.group_id = Number(filters.value.group_id)
  }
  if (filters.value.model) {
    params.model = filters.value.model
  }
  if (filters.value.request_type) {
    params.request_type = filters.value.request_type
  }
  if (filters.value.billing_mode) {
    params.billing_mode = filters.value.billing_mode
  }

  return params
}

const loadUsageLogs = async () => {
  if (abortController) {
    abortController.abort()
  }
  const currentAbortController = new AbortController()
  abortController = currentAbortController
  const { signal } = currentAbortController
  loading.value = true
  try {
    const response = await usageAPI.query(
      buildUsageQueryParams(pagination.page, pagination.page_size),
      { signal }
    )
    if (signal.aborted) {
      return
    }
    usageLogs.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
  } catch (error) {
    if (signal.aborted) {
      return
    }
    const abortError = error as { name?: string; code?: string }
    if (abortError?.name === 'AbortError' || abortError?.code === 'ERR_CANCELED') {
      return
    }
    appStore.showError(t('usage.failedToLoad'))
  } finally {
    if (abortController === currentAbortController) {
      loading.value = false
    }
  }
}

const loadApiKeys = async () => {
  try {
    const response = await keysAPI.list(1, 100)
    apiKeys.value = response.items
  } catch (error) {
    console.error('Failed to load API keys:', error)
  }
}

const loadGroups = async () => {
  try {
    groups.value = await userGroupsAPI.getAvailable()
  } catch (error) {
    console.error('Failed to load groups:', error)
  }
}

const loadUsageStats = async () => {
  const sequence = ++statsRequestSequence
  endpointStatsLoading.value = true
  try {
    const stats = await usageAPI.getStatsByDateRange(
      filters.value.start_date || startDate.value,
      filters.value.end_date || endDate.value,
      {
        api_key_id: filters.value.api_key_id ? Number(filters.value.api_key_id) : undefined,
        group_id: filters.value.group_id ? Number(filters.value.group_id) : undefined,
        model: filters.value.model || undefined,
        request_type: filters.value.request_type,
        billing_mode: filters.value.billing_mode || undefined,
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      }
    )
    if (sequence !== statsRequestSequence) return
    usageStats.value = stats
    endpointStats.value = stats.endpoints || []
  } catch (error) {
    if (sequence !== statsRequestSequence) return
    console.error('Failed to load usage stats:', error)
    endpointStats.value = []
  } finally {
    if (sequence === statsRequestSequence) endpointStatsLoading.value = false
  }
}

const buildAnalyticsFilters = (): UsageQueryParams => ({
  start_date: filters.value.start_date || startDate.value,
  end_date: filters.value.end_date || endDate.value,
  timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  api_key_id: filters.value.api_key_id ? Number(filters.value.api_key_id) : undefined,
  group_id: filters.value.group_id ? Number(filters.value.group_id) : undefined,
  model: filters.value.model || undefined,
  request_type: filters.value.request_type,
  billing_mode: filters.value.billing_mode || undefined,
})

const loadModelStats = async (): Promise<void> => {
  const sequence = ++modelRequestSequence
  modelStatsLoading.value = true
  try {
    const response = await usageAPI.getDashboardModels({
      ...buildAnalyticsFilters(),
      model_source: 'requested',
    })
    if (sequence !== modelRequestSequence) return
    requestedModelStats.value = response.models || []
  } catch (error) {
    if (sequence !== modelRequestSequence) return
    console.error('Failed to load user model stats:', error)
    requestedModelStats.value = []
  } finally {
    if (sequence === modelRequestSequence) modelStatsLoading.value = false
  }
}

const loadChartData = async (): Promise<void> => {
  const sequence = ++chartRequestSequence
  chartsLoading.value = true
  try {
    const response = await usageAPI.getDashboardSnapshotV2({
      ...buildAnalyticsFilters(),
      granularity: granularity.value,
      include_trend: true,
      include_model_stats: false,
      include_group_stats: true,
    })
    if (sequence !== chartRequestSequence) return
    trendData.value = response.trend || []
    groupStats.value = response.groups || []
  } catch (error) {
    if (sequence !== chartRequestSequence) return
    console.error('Failed to load user usage charts:', error)
    trendData.value = []
    groupStats.value = []
  } finally {
    if (sequence === chartRequestSequence) chartsLoading.value = false
  }
}

const applyFilters = () => {
  pagination.page = 1
  void loadUsageLogs()
  void loadUsageStats()
  void loadModelStats()
  void loadChartData()
  if (activeTab.value === 'errors') {
    errorPagination.page = 1
    void loadErrors()
  }
}

const resetFilters = () => {
  filters.value = {
    api_key_id: undefined,
    group_id: undefined,
    model: undefined,
    request_type: undefined,
    billing_mode: undefined,
    start_date: undefined,
    end_date: undefined
  }
  // Reset date range to default (last 7 days)
  const now = new Date()
  const weekAgo = new Date(now)
  weekAgo.setDate(weekAgo.getDate() - 6)
  startDate.value = formatLocalDate(weekAgo)
  endDate.value = formatLocalDate(now)
  filters.value.start_date = startDate.value
  filters.value.end_date = endDate.value
  granularity.value = 'day'
  pagination.page = 1
  void loadUsageLogs()
  void loadUsageStats()
  void loadModelStats()
  void loadChartData()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadUsageLogs()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadUsageLogs()
}

const handleSort = (key: string, order: 'asc' | 'desc') => {
  sortState.sort_by = key
  sortState.sort_order = order
  pagination.page = 1
  loadUsageLogs()
}

const loadErrors = async (): Promise<void> => {
  if (!errorViewEnabled.value) return
  const sequence = ++errorRequestSequence
  errorLoading.value = true
  const params: UserErrorListParams = {
    page: errorPagination.page,
    page_size: errorPagination.page_size,
    start_date: startDate.value,
    end_date: endDate.value,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    api_key_id: errorFilter.value.api_key_id || undefined,
    model: errorFilter.value.model.trim() || undefined,
    category: errorFilter.value.category || undefined,
    status_code: errorFilter.value.status_code || undefined,
    sort_by: errorSort.sort_by,
    sort_order: errorSort.sort_order,
  }
  try {
    const response = await usageAPI.listMyErrorRequests(params)
    if (sequence !== errorRequestSequence) return
    errorRows.value = response.items
    errorTotal.value = response.total
  } catch (error) {
    if (sequence !== errorRequestSequence) return
    console.error('Failed to load user error requests:', error)
    appStore.showError(t('usage.errors.failedToLoad'))
  } finally {
    if (sequence === errorRequestSequence) errorLoading.value = false
  }
}

const applyErrorFilters = (): void => {
  errorPagination.page = 1
  void loadErrors()
}

const resetErrorFilters = (): void => {
  errorFilter.value = {
    api_key_id: null,
    model: '',
    category: null,
    status_code: null,
  }
  errorPagination.page = 1
  void loadErrors()
}

const handleErrorSort = (
  sortBy: 'created_at' | 'model' | 'status_code',
  sortOrder: 'asc' | 'desc'
): void => {
  errorSort.sort_by = sortBy
  errorSort.sort_order = sortOrder
  errorPagination.page = 1
  void loadErrors()
}

const handleErrorPageChange = (page: number): void => {
  errorPagination.page = page
  void loadErrors()
}

const handleErrorPageSizeChange = (pageSize: number): void => {
  errorPagination.page_size = pageSize
  errorPagination.page = 1
  void loadErrors()
}

const switchToErrors = (): void => {
  if (!errorViewEnabled.value) return
  activeTab.value = 'errors'
  showColumnMenu.value = false
  void loadErrors()
}

/**
 * Escape CSV value to prevent injection and handle special characters
 */
const escapeCSVValue = (value: unknown): string => {
  if (value == null) return ''

  const str = String(value)
  const escaped = str.replace(/"/g, '""')

  // Prevent formula injection by prefixing dangerous characters with single quote
  if (/^[=+\-@\t\r]/.test(str)) {
    return `"\'${escaped}"`
  }

  // Escape values containing comma, quote, or newline
  if (/[,"\n\r]/.test(str)) {
    return `"${escaped}"`
  }

  return str
}

const exportToCSV = async () => {
  if (pagination.total === 0) {
    appStore.showWarning(t('usage.noDataToExport'))
    return
  }

  exporting.value = true
  appStore.showInfo(t('usage.preparingExport'))

  try {
    const allLogs: UsageLog[] = []
    const pageSize = 100 // Use a larger page size for export to reduce requests
    const totalRequests = Math.ceil(pagination.total / pageSize)

    for (let page = 1; page <= totalRequests; page++) {
      const response = await usageAPI.query(buildUsageQueryParams(page, pageSize))
      allLogs.push(...response.items)
    }

    if (allLogs.length === 0) {
      appStore.showWarning(t('usage.noDataToExport'))
      return
    }

    const headers = [
      'Time',
      'API Key Name',
      'Group Name',
      'Group ID',
      'Request ID',
      'Model',
      'Reasoning Effort',
      'Inbound Endpoint',
      'Type',
      'Billing Mode',
      'Input Tokens',
      'Output Tokens',
      'Cache Read Tokens',
      'Cache Creation Tokens',
      'Pricing Rate Multiplier',
      'Balance Conversion Multiplier',
      'Composite Rate Multiplier',
      'Billed Cost',
      'Original Cost',
      'First Token (ms)',
      'Duration (ms)',
      'User-Agent'
    ]
    const rows = allLogs.map((log) =>
      [
        log.created_at,
        log.api_key?.name || '',
        log.group?.name || '',
        log.group_id ?? '',
        log.request_id,
        log.model,
        formatReasoningEffort(log.reasoning_effort),
        log.inbound_endpoint || '',
        getRequestTypeExportText(log),
        getBillingModeLabel(getDisplayBillingMode(log), t),
        toFiniteNumber(log.input_tokens),
        toFiniteNumber(log.output_tokens),
        toFiniteNumber(log.cache_read_tokens),
        toFiniteNumber(log.cache_creation_tokens),
        pricingRateMultiplier(log),
        balanceConversionMultiplier(log),
        toFiniteNumber(log.rate_multiplier, 1),
        formatCostNumberFixed(log.actual_cost, 8),
        formatCostNumberFixed(log.total_cost, 8),
        log.first_token_ms ?? '',
        log.duration_ms ?? '',
        log.user_agent || ''
      ].map(escapeCSVValue)
    )

    const csvContent = [
      headers.map(escapeCSVValue).join(','),
      ...rows.map((row) => row.join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `usage_${filters.value.start_date}_to_${filters.value.end_date}.csv`
    link.click()
    window.URL.revokeObjectURL(url)

    appStore.showSuccess(t('usage.exportSuccess'))
  } catch (error) {
    appStore.showError(t('usage.exportFailed'))
    console.error('CSV Export failed:', error)
  } finally {
    exporting.value = false
  }
}

// Tooltip functions
const showTooltip = (event: MouseEvent, row: UsageLog) => {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()

  tooltipData.value = row
  // Position to the right of the icon, vertically centered
  tooltipPosition.value.x = rect.right + 8
  tooltipPosition.value.y = rect.top + rect.height / 2
  tooltipVisible.value = true
}

const hideTooltip = () => {
  tooltipVisible.value = false
  tooltipData.value = null
}

const pricingRateMultiplier = (log: Pick<UsageLog, 'rate_multiplier' | 'pricing_rate_multiplier'> | null | undefined): number =>
  log?.pricing_rate_multiplier ?? log?.rate_multiplier ?? 1

const balanceConversionMultiplier = (log: Pick<UsageLog, 'balance_conversion_multiplier'> | null | undefined): number =>
  log?.balance_conversion_multiplier ?? 1

// Token tooltip functions
const showTokenTooltip = (event: MouseEvent, row: UsageLog) => {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()

  tokenTooltipData.value = row
  tokenTooltipPosition.value.x = rect.right + 8
  tokenTooltipPosition.value.y = rect.top + rect.height / 2
  tokenTooltipVisible.value = true
}

const hideTokenTooltip = () => {
  tokenTooltipVisible.value = false
  tokenTooltipData.value = null
}

const openUsageDetails = (row: UsageLog) => {
  selectedUsageLog.value = row
  detailsVisible.value = true
}

const closeUsageDetails = () => {
  detailsVisible.value = false
  selectedUsageLog.value = null
}

onMounted(() => {
  loadVisibleColumns()
  loadErrorVisibleColumns()
  void loadApiKeys()
  void loadGroups()
  void loadUsageLogs()
  void loadUsageStats()
  void loadModelStats()
  void loadChartData()
})

onUnmounted(() => {
  abortController?.abort()
  errorRequestSequence += 1
  statsRequestSequence += 1
  modelRequestSequence += 1
  chartRequestSequence += 1
})
</script>
