<template>
  <div class="card overflow-hidden">
    <div class="overflow-auto">
      <DataTable
        :columns="columns"
        :data="data"
        :loading="loading"
        :server-side-sort="serverSideSort"
        :default-sort-key="defaultSortKey"
        :default-sort-order="defaultSortOrder"
        @sort="(key, order) => $emit('sort', key, order)"
      >
        <template #cell-user="{ row }">
          <div class="text-sm">
            <button
              v-if="row.user?.email"
              class="font-medium text-primary-600 underline decoration-dashed underline-offset-2 transition-colors hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
              @click="$emit('userClick', row.user_id, row.user?.email)"
              :title="t('admin.usage.clickToViewBalance')"
            >
              {{ row.user.email }}
            </button>
            <span v-else class="font-medium text-gray-900 dark:text-white">-</span>
            <span class="ml-1 text-gray-500 dark:text-gray-400">#{{ row.user_id }}</span>
          </div>
        </template>

        <template #cell-api_key="{ row }">
          <span class="text-sm text-gray-900 dark:text-white">{{ row.api_key?.name || '-' }}</span>
        </template>

        <template #cell-account="{ row }">
          <span class="text-sm text-gray-900 dark:text-white">{{ row.account?.name || '-' }}</span>
        </template>

        <template #cell-model="{ row }">
          <div v-if="row.model_mapping_chain && row.model_mapping_chain.includes('→')" class="space-y-0.5 text-xs">
            <div v-for="(step, i) in displayModelChain(row.model_mapping_chain).split('→')" :key="i"
                 class="break-all"
                 :class="i === 0 ? 'font-medium text-gray-900 dark:text-white' : 'text-gray-500 dark:text-gray-400'"
                 :style="i > 0 ? `padding-left: ${i * 0.75}rem` : ''">
              <span v-if="i > 0" class="mr-0.5">↳</span>{{ i === 0 ? displayModelWithReasoningEffort(step, row.reasoning_effort) : step }}
            </div>
          </div>
          <div v-else-if="row.upstream_model && row.upstream_model !== row.model" class="space-y-0.5 text-xs">
            <div class="break-all font-medium text-gray-900 dark:text-white">
              {{ displayModelWithReasoningEffort(row.model, row.reasoning_effort) }}
            </div>
            <div class="break-all text-gray-500 dark:text-gray-400">
              <span class="mr-0.5">↳</span>{{ displayModelLabel(row.upstream_model) }}
            </div>
          </div>
          <span v-else class="font-medium text-gray-900 dark:text-white">{{ displayModelWithReasoningEffort(row.model, row.reasoning_effort) }}</span>
        </template>

        <template #cell-reasoning_effort="{ row }">
          <span class="text-sm text-gray-900 dark:text-white">
            {{ formatReasoningEffort(row.reasoning_effort) }}
          </span>
        </template>

        <template #cell-endpoint="{ row }">
          <div class="max-w-[320px] space-y-1 text-xs">
            <div class="break-all text-gray-700 dark:text-gray-300">
              <span class="font-medium text-gray-500 dark:text-gray-400">{{ t('usage.inbound') }}:</span>
              <span class="ml-1">{{ row.inbound_endpoint?.trim() || '-' }}</span>
            </div>
            <div class="break-all text-gray-700 dark:text-gray-300">
              <span class="font-medium text-gray-500 dark:text-gray-400">{{ t('usage.upstream') }}:</span>
              <span class="ml-1">{{ row.upstream_endpoint?.trim() || '-' }}</span>
            </div>
          </div>
        </template>

        <template #cell-group="{ row }">
          <span v-if="row.group" class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium bg-[#f3e7df] text-[#a9583e] dark:bg-[#cc785c]/15 dark:text-[#f0b89e]">
            {{ row.group.name }}
          </span>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #cell-stream="{ row }">
          <span class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium" :class="getRequestTypeBadgeClass(row)">
            {{ getRequestTypeLabel(row) }}
          </span>
        </template>

        <template #cell-billing_mode="{ row }">
          <span class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium" :class="getBillingModeBadgeClass(getDisplayBillingMode(row))">
            {{ getBillingModeLabel(getDisplayBillingMode(row), t) }}
          </span>
        </template>

        <template #cell-tokens="{ row }">
          <!-- 图片生成请求（仅按次计费时显示图片格式） -->
          <div v-if="isImageUsage(row)" class="flex items-center gap-1.5">
            <svg class="h-4 w-4 text-[#a9583e] dark:text-[#f0b89e]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <span class="font-medium text-gray-900 dark:text-white">{{ row.image_count }}{{ t('usage.imageUnit') }}</span>
            <span class="text-[#6c6a64] dark:text-gray-400">({{ formatImageBillingSize(row, t) }})</span>
          </div>
          <!-- Token 请求 -->
          <div v-else class="flex items-center gap-1.5">
            <div class="space-y-1 text-sm">
              <div class="flex items-center gap-2">
                <div class="inline-flex items-center gap-1">
                  <Icon name="arrowDown" size="sm" class="h-3.5 w-3.5 text-[#a9583e] dark:text-[#a9583e]" />
                  <span class="font-medium text-gray-900 dark:text-white">{{ row.input_tokens?.toLocaleString() || 0 }}</span>
                </div>
                <div class="inline-flex items-center gap-1">
                  <Icon name="arrowUp" size="sm" class="h-3.5 w-3.5 text-[#a9583e] dark:text-[#f0b89e]" />
                  <span class="font-medium text-gray-900 dark:text-white">{{ row.output_tokens?.toLocaleString() || 0 }}</span>
                </div>
              </div>
              <div v-if="row.cache_creation_tokens > 0" class="flex items-center gap-2">
                <div class="inline-flex items-center gap-1">
                  <svg class="h-3.5 w-3.5 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                  <span class="font-medium text-amber-600 dark:text-amber-400">{{ formatCacheTokens(row.cache_creation_tokens) }}</span>
                  <span v-if="row.cache_creation_1h_tokens > 0" class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-orange-100 text-orange-600 ring-1 ring-inset ring-orange-200 dark:bg-orange-500/20 dark:text-orange-400 dark:ring-orange-500/30">1h</span>
                  <span v-if="row.cache_ttl_overridden" :title="t('usage.cacheTtlOverriddenHint')" class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-rose-100 text-rose-600 ring-1 ring-inset ring-rose-200 dark:bg-rose-500/20 dark:text-rose-400 dark:ring-rose-500/30 cursor-help">R</span>
                </div>
              </div>
              <div v-if="hasImageInputTokens(row)" class="flex items-center gap-2">
                <div class="inline-flex items-center gap-1 text-fuchsia-600 dark:text-fuchsia-400">
                  <Icon name="image" size="sm" class="h-3.5 w-3.5" />
                  <span class="font-medium">{{ row.image_input_tokens.toLocaleString() }}</span>
                </div>
              </div>
              <div v-if="hasImageOutputTokens(row)" class="flex items-center gap-2">
                <div class="inline-flex items-center gap-1 text-pink-600 dark:text-pink-400">
                  <Icon name="image" size="sm" class="h-3.5 w-3.5" />
                  <span class="font-medium">{{ row.image_output_tokens.toLocaleString() }}</span>
                </div>
              </div>
            </div>
            <!-- Token Detail Tooltip -->
            <div
              class="group relative"
              @mouseenter="showTokenTooltip($event, row)"
              @mouseleave="hideTokenTooltip"
            >
              <div class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-gray-100 transition-colors group-hover:bg-[#f3e7df] dark:bg-gray-700 dark:group-hover:bg-[#cc785c]/15">
                <Icon name="infoCircle" size="xs" class="text-gray-400 group-hover:text-[#a9583e] dark:text-gray-500 dark:group-hover:text-[#f0b89e]" />
              </div>
            </div>
          </div>
        </template>

        <template #cell-cache_read="{ row }">
          <div
            v-if="!isImageUsage(row) && row.cache_read_tokens > 0"
            class="inline-flex items-center gap-1 text-sm"
            :title="`${row.cache_read_tokens.toLocaleString()} (${formatCacheReadPercent(row)})`"
          >
            <Icon name="database" size="sm" class="h-3.5 w-3.5 text-[#a9583e]" />
            <span class="font-medium text-[#a9583e] dark:text-[#f0b89e]">
              {{ formatCacheTokens(row.cache_read_tokens) }}
            </span>
            <span data-testid="cache-read-percent" class="text-xs font-medium text-[#8e8b82] dark:text-[#d8cec2]/80">
              {{ formatCacheReadPercent(row) }}
            </span>
          </div>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #cell-cost="{ row }">
          <div class="text-sm">
            <div class="flex items-center gap-1.5">
              <span class="font-medium text-[#a9583e] dark:text-[#f0b89e]">${{ row.actual_cost?.toFixed(6) || '0.000000' }}</span>
              <span
                v-if="row.long_context_billing_applied"
                data-testid="long-context-billing-marker"
                class="inline-flex items-center rounded px-1 py-px text-[10px] font-semibold leading-tight bg-amber-100 text-amber-700 ring-1 ring-inset ring-amber-200 dark:bg-amber-500/20 dark:text-amber-300 dark:ring-amber-500/30"
              >x2</span>
              <!-- Cost Detail Tooltip -->
              <div
                class="group relative"
                @mouseenter="showTooltip($event, row)"
                @mouseleave="hideTooltip"
              >
                <div class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-gray-100 transition-colors group-hover:bg-[#f3e7df] dark:bg-gray-700 dark:group-hover:bg-[#cc785c]/15">
                  <Icon name="infoCircle" size="xs" class="text-gray-400 group-hover:text-[#a9583e] dark:text-gray-500 dark:group-hover:text-[#f0b89e]" />
                </div>
              </div>
            </div>
            <div v-if="row.account_rate_multiplier != null" class="mt-0.5 text-[11px] text-orange-500 dark:text-orange-400">
              A ${{ accountBilled(row).toFixed(6) }}
            </div>
          </div>
        </template>

        <!-- 合并首字/总耗时的健康度列：色条上端表示首字档，下端表示总耗时档。 -->
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
          <span class="text-sm text-gray-600 dark:text-gray-400">{{ formatDateTime(value) }}</span>
        </template>

        <template #cell-user_agent="{ row }">
          <span v-if="row.user_agent" class="text-sm text-gray-600 dark:text-gray-400 block max-w-[320px] truncate" :title="row.user_agent">{{ formatUserAgent(row.user_agent) }}</span>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #cell-ip_address="{ row }">
          <span v-if="row.ip_address" class="text-sm font-mono text-gray-600 dark:text-gray-400">{{ row.ip_address }}</span>
          <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
        </template>

        <template #empty><EmptyState :message="t('usage.noRecords')" /></template>
      </DataTable>
    </div>
  </div>

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
      <div class="whitespace-nowrap rounded-lg border border-[#d8cec2] bg-[#fffaf5] px-3 py-2.5 text-xs text-[#141413] shadow-xl dark:border-gray-600 dark:bg-gray-800 dark:text-white">
        <div class="space-y-1.5">
          <div>
            <div class="text-xs font-semibold text-[#504f49] mb-1 dark:text-gray-300">{{ t('usage.tokenDetails') }}</div>
            <div v-if="tokenTooltipData && tokenTooltipData.input_tokens > 0 && !hasImageInputTokens(tokenTooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.inputTokens') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ tokenTooltipData.input_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageInputTokens(tokenTooltipData) && textInputTokens(tokenTooltipData) > 0" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.inputTokens') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ textInputTokens(tokenTooltipData).toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageInputTokens(tokenTooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageInputTokens') }}</span>
              <span class="font-medium text-fuchsia-600 dark:text-fuchsia-300">{{ tokenTooltipData.image_input_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.output_tokens > 0 && !hasImageOutputTokens(tokenTooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.outputTokens') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ tokenTooltipData.output_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageOutputTokens(tokenTooltipData) && textOutputTokens(tokenTooltipData) > 0" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.outputTokens') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ textOutputTokens(tokenTooltipData).toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && hasImageOutputTokens(tokenTooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageOutputTokens') }}</span>
              <span class="font-medium text-pink-600 dark:text-pink-300">{{ tokenTooltipData.image_output_tokens.toLocaleString() }}</span>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.cache_creation_tokens > 0">
              <!-- 有 5m/1h 明细时，展开显示 -->
              <template v-if="tokenTooltipData.cache_creation_5m_tokens > 0 || tokenTooltipData.cache_creation_1h_tokens > 0">
                <div v-if="tokenTooltipData.cache_creation_5m_tokens > 0" class="flex items-center justify-between gap-4">
                  <span class="text-[#6c6a64] flex items-center gap-1.5 dark:text-gray-400">
                    {{ t('admin.usage.cacheCreation5mTokens') }}
                    <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-amber-500/20 text-amber-400 ring-1 ring-inset ring-amber-500/30">5m</span>
                  </span>
                  <span class="font-medium text-[#141413] dark:text-white">{{ tokenTooltipData.cache_creation_5m_tokens.toLocaleString() }}</span>
                </div>
                <div v-if="tokenTooltipData.cache_creation_1h_tokens > 0" class="flex items-center justify-between gap-4">
                  <span class="text-[#6c6a64] flex items-center gap-1.5 dark:text-gray-400">
                    {{ t('admin.usage.cacheCreation1hTokens') }}
                    <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-orange-500/20 text-orange-400 ring-1 ring-inset ring-orange-500/30">1h</span>
                  </span>
                  <span class="font-medium text-[#141413] dark:text-white">{{ tokenTooltipData.cache_creation_1h_tokens.toLocaleString() }}</span>
                </div>
              </template>
              <!-- 无明细时，只显示聚合值 -->
              <div v-else class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.cacheCreationTokens') }}</span>
                <span class="font-medium text-[#141413] dark:text-white">{{ tokenTooltipData.cache_creation_tokens.toLocaleString() }}</span>
              </div>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.cache_ttl_overridden" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] flex items-center gap-1.5 dark:text-gray-400">
                {{ t('usage.cacheTtlOverriddenLabel') }}
                <span class="inline-flex items-center rounded px-1 py-px text-[10px] font-medium leading-tight bg-rose-500/20 text-rose-400 ring-1 ring-inset ring-rose-500/30">R-{{ tokenTooltipData.cache_creation_1h_tokens > 0 ? '5m' : '1H' }}</span>
              </span>
              <span class="font-medium text-rose-400">{{ tokenTooltipData.cache_creation_1h_tokens > 0 ? t('usage.cacheTtlOverridden1h') : t('usage.cacheTtlOverridden5m') }}</span>
            </div>
            <div v-if="tokenTooltipData && tokenTooltipData.cache_read_tokens > 0" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.cacheReadTokens') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">{{ tokenTooltipData.cache_read_tokens.toLocaleString() }} ({{ formatCacheReadPercent(tokenTooltipData) }})</span>
            </div>
          </div>
          <div class="flex items-center justify-between gap-6 border-t border-[#d8cec2] pt-1.5 dark:border-gray-700">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.totalTokens') }}</span>
            <span class="font-semibold text-[#a9583e] dark:text-[#f0b89e]">{{ ((tokenTooltipData?.input_tokens || 0) + (tokenTooltipData?.output_tokens || 0) + (tokenTooltipData?.cache_creation_tokens || 0) + (tokenTooltipData?.cache_read_tokens || 0)).toLocaleString() }}</span>
          </div>
        </div>
        <div class="absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-r-[#fffaf5] border-t-transparent dark:border-r-gray-800"></div>
      </div>
    </div>
  </Teleport>

  <!-- Cost Tooltip Portal -->
  <Teleport to="body">
    <div
      v-if="tooltipVisible"
      class="fixed z-[9999] pointer-events-none -translate-y-1/2"
      :style="{
        left: tooltipPosition.x + 'px',
        top: tooltipPosition.y + 'px'
      }"
    >
      <div class="whitespace-nowrap rounded-lg border border-[#d8cec2] bg-[#fffaf5] px-3 py-2.5 text-xs text-[#141413] shadow-xl dark:border-gray-600 dark:bg-gray-800 dark:text-white">
        <div class="space-y-1.5">
          <!-- Cost Breakdown -->
          <div class="mb-2 border-b border-[#d8cec2] pb-1.5 dark:border-gray-700">
            <div class="text-xs font-semibold text-[#504f49] mb-1 dark:text-gray-300">{{ t('usage.costDetails') }}</div>
            <div v-if="tooltipData && tooltipData.input_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.inputCost') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">${{ tooltipData.input_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && hasImageInputCost(tooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageInputCost') }}</span>
              <span class="font-medium text-fuchsia-600 dark:text-fuchsia-300">${{ tooltipData.image_input_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.output_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.outputCost') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">${{ tooltipData.output_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && hasImageOutputCost(tooltipData)" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageOutputCost') }}</span>
              <span class="font-medium text-pink-600 dark:text-pink-300">${{ tooltipData.image_output_cost.toFixed(6) }}</span>
            </div>
            <!-- Token billing: show unit prices per 1M tokens -->
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
                <span class="font-medium text-[#f0b89e]">${{ imageUnitPrice(tooltipData).toFixed(6) }}</span>
              </div>
              <div class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.imageTotalPrice') }}</span>
                <span class="font-medium text-[#141413] dark:text-white">${{ tooltipData.total_cost?.toFixed(6) || '0.000000' }}</span>
              </div>
            </template>
            <template v-else-if="!tooltipData?.billing_mode || tooltipData.billing_mode === BILLING_MODE_TOKEN">
              <div v-if="tooltipData && textInputTokens(tooltipData) > 0" class="flex items-center justify-between gap-4">
                <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.inputTokenPrice') }}</span>
                <span class="font-medium text-[#f0b89e]">{{ formatTokenPricePerMillion(tooltipData.input_cost, textInputTokens(tooltipData)) }} {{ t('usage.perMillionTokens') }}</span>
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
              <span class="font-medium text-[#f0b89e]">${{ tooltipData?.total_cost?.toFixed(6) || '0.000000' }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.cache_creation_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.cacheCreationCost') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">${{ tooltipData.cache_creation_cost.toFixed(6) }}</span>
            </div>
            <div v-if="tooltipData && tooltipData.cache_read_cost > 0" class="flex items-center justify-between gap-4">
              <span class="text-[#6c6a64] dark:text-gray-400">{{ t('admin.usage.cacheReadCost') }}</span>
              <span class="font-medium text-[#141413] dark:text-white">${{ tooltipData.cache_read_cost.toFixed(6) }}</span>
            </div>
          </div>
          <!-- Rate and Summary -->
          <div class="flex items-center justify-between gap-6">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.serviceTier') }}</span>
            <span class="font-semibold text-[#a9583e] dark:text-[#f0b89e]">{{ getUsageServiceTierLabel(tooltipData?.service_tier, t) }}</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.rate') }}</span>
            <span class="font-semibold text-[#a9583e] dark:text-[#f0b89e]">{{ formatMultiplier(tooltipData?.rate_multiplier || 1) }}x</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.original') }}</span>
            <span class="font-medium text-[#141413] dark:text-white">${{ tooltipData?.total_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.userBilled') }}</span>
            <span class="font-semibold text-[#a9583e] dark:text-[#f0b89e]">${{ tooltipData?.actual_cost?.toFixed(6) || '0.000000' }}</span>
          </div>
          <!-- Account billing (separated from user billing) -->
          <div class="flex items-center justify-between gap-6 border-t border-[#d8cec2] pt-1.5 dark:border-gray-700">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.accountMultiplier') }}</span>
            <span class="font-semibold text-[#a9583e] dark:text-[#f0b89e]">{{ formatMultiplier(tooltipData?.account_rate_multiplier ?? 1) }}x</span>
          </div>
          <div class="flex items-center justify-between gap-6">
            <span class="text-[#6c6a64] dark:text-gray-400">{{ t('usage.accountBilled') }}</span>
            <span class="font-semibold text-[#a9583e] dark:text-[#f0b89e]">
              ${{ accountBilled({
                total_cost: tooltipData?.total_cost,
                account_stats_cost: tooltipData?.account_stats_cost,
                account_rate_multiplier: tooltipData?.account_rate_multiplier,
              }).toFixed(6) }}
            </span>
          </div>
        </div>
        <div class="absolute right-full top-1/2 h-0 w-0 -translate-y-1/2 border-b-[6px] border-r-[6px] border-t-[6px] border-b-transparent border-r-[#fffaf5] border-t-transparent dark:border-r-gray-800"></div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatDateTime, formatReasoningEffort } from '@/utils/format'
import { formatCacheTokens, formatMultiplier } from '@/utils/formatters'
import { formatTokenPricePerMillion } from '@/utils/usagePricing'
import { getUsageServiceTierLabel } from '@/utils/usageServiceTier'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import {
  LATENCY_BAR_CLASSES,
  LATENCY_BAR_FROM_CLASSES,
  LATENCY_BAR_TO_CLASSES,
  LATENCY_TEXT_CLASSES,
  durationSeverity,
  firstTokenSeverity,
} from '@/utils/latencyHealth'
import { getBillingModeLabel, getBillingModeBadgeClass, BILLING_MODE_TOKEN, BILLING_MODE_PER_REQUEST, BILLING_MODE_IMAGE } from '@/utils/billingMode'
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
import { displayModelChain, displayModelLabel, displayModelWithReasoningEffort } from '@/utils/modelDisplay'

/** Compute the account-billed cost for display: (account_stats_cost ?? total_cost) * rate_multiplier */
function accountBilled(row: { total_cost?: number | null; account_stats_cost?: number | null; account_rate_multiplier?: number | null }): number {
  const base = row.account_stats_cost != null ? row.account_stats_cost : (row.total_cost ?? 0)
  const result = base * (row.account_rate_multiplier ?? 1)
  return Number.isNaN(result) ? 0 : result
}

function imageUnitPrice(row: AdminUsageLog | null): number {
  if (!row || row.image_count <= 0) return 0
  const total = row.total_cost ?? 0
  const price = total / row.image_count
  return Number.isFinite(price) ? price : 0
}

function getDisplayBillingMode(row: Pick<AdminUsageLog, 'billing_mode' | 'image_count'> | null | undefined): string | null | undefined {
  if (row?.billing_mode) {
    return row.billing_mode
  }
  if ((row?.image_count ?? 0) > 0) {
    return BILLING_MODE_IMAGE
  }
  return row?.billing_mode
}

function isImageUsage(row: Pick<AdminUsageLog, 'billing_mode' | 'image_count'> | null | undefined): boolean {
  const mode = getDisplayBillingMode(row)
  return mode === BILLING_MODE_IMAGE || mode === BILLING_MODE_PER_REQUEST
}

function formatCacheReadPercent(row: Pick<AdminUsageLog, 'input_tokens' | 'cache_creation_tokens' | 'cache_read_tokens'>): string {
  const total = (row.input_tokens ?? 0) + (row.cache_creation_tokens ?? 0) + (row.cache_read_tokens ?? 0)
  if (total <= 0) return '0%'
  const percent = ((row.cache_read_tokens ?? 0) / total) * 100
  if (percent >= 99.95 && percent < 100) return '99.9%'
  return `${percent.toFixed(1)}%`
}

import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import type { AdminUsageLog } from '@/types'
import type { Column } from '@/components/common/types'

interface Props {
  data: AdminUsageLog[]
  loading?: boolean
  columns: Column[]
  serverSideSort?: boolean
  defaultSortKey?: string
  defaultSortOrder?: 'asc' | 'desc'
}

withDefaults(defineProps<Props>(), {
  loading: false,
  serverSideSort: false,
  defaultSortKey: '',
  defaultSortOrder: 'asc'
})
defineEmits<{
  userClick: [userID: number, email?: string]
  sort: [key: string, order: 'asc' | 'desc']
}>()
const { t } = useI18n()

// Tooltip state - cost
const tooltipVisible = ref(false)
const tooltipPosition = ref({ x: 0, y: 0 })
const tooltipData = ref<AdminUsageLog | null>(null)

// Tooltip state - token
const tokenTooltipVisible = ref(false)
const tokenTooltipPosition = ref({ x: 0, y: 0 })
const tokenTooltipData = ref<AdminUsageLog | null>(null)

const getRequestTypeLabel = (row: AdminUsageLog): string => {
  const requestType = resolveUsageRequestType(row)
  if (requestType === 'live') return t('usage.live')
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
}

const getRequestTypeBadgeClass = (row: AdminUsageLog): string => {
  const requestType = resolveUsageRequestType(row)
  if (requestType === 'live') return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200'
  if (requestType === 'ws_v2') return 'bg-[#fffaf5] text-[#5f7f68] ring-1 ring-[#d8cec2] dark:bg-[#7f9d8a]/15 dark:text-[#9ab3a0] dark:ring-[#7f9d8a]/30'
  if (requestType === 'stream') return 'bg-[#f3e7df] text-[#504f49] dark:bg-[#cc785c]/15 dark:text-[#f0b89e]'
  if (requestType === 'sync') return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
  return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200'
}



const formatUserAgent = (ua: string): string => {
  return ua
}

const formatDuration = (ms: number | null | undefined): string => {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(2)}s`
  const totalSec = Math.round(ms / 1000)
  if (totalSec < 3600) return `${Math.floor(totalSec / 60)}m ${totalSec % 60}s`
  return `${Math.floor(totalSec / 3600)}h ${Math.floor((totalSec % 3600) / 60)}m`
}

// Cost tooltip functions
const showTooltip = (event: MouseEvent, row: AdminUsageLog) => {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  tooltipData.value = row
  tooltipPosition.value.x = rect.right + 8
  tooltipPosition.value.y = rect.top + rect.height / 2
  tooltipVisible.value = true
}

const hideTooltip = () => {
  tooltipVisible.value = false
  tooltipData.value = null
}

// Token tooltip functions
const showTokenTooltip = (event: MouseEvent, row: AdminUsageLog) => {
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
</script>
