<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('myAccounts.summary.available') }}</p>
          <p class="mt-2 text-2xl font-semibold text-[#5f7f68] dark:text-[#9ab3a0]">
            {{ formatCreditAmount(shareSummary?.available_amount ?? 0) }}
          </p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('myAccounts.summary.frozen') }}</p>
          <p class="mt-2 text-2xl font-semibold text-amber-600 dark:text-amber-400">
            {{ formatCreditAmount(shareSummary?.frozen_amount ?? 0) }}
          </p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('myAccounts.summary.transferred') }}</p>
          <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
            {{ formatCreditAmount(shareSummary?.transferred_amount ?? 0) }}
          </p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('myAccounts.summary.total') }}</p>
          <div class="mt-2 flex items-center justify-between gap-3">
            <p class="text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatCreditAmount(shareSummary?.total_amount ?? 0) }}
            </p>
            <button
              class="btn btn-primary btn-sm"
              :disabled="transferring || (shareSummary?.available_amount ?? 0) <= 0"
              @click="transferShare"
            >
              <Icon v-if="transferring" name="refresh" size="sm" class="animate-spin" />
              <Icon v-else name="dollar" size="sm" />
              <span>{{ t('myAccounts.transfer') }}</span>
            </button>
          </div>
        </div>
      </div>

      <TablePageLayout>
        <template #actions>
          <div class="flex flex-wrap justify-end gap-3">
            <button class="btn btn-secondary" :disabled="loading" @click="loadAll">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
              <span>{{ t('common.refresh') }}</span>
            </button>
            <button class="btn btn-secondary" data-testid="my-accounts-open-import" @click="openImportModal">
              <Icon name="upload" size="md" />
              <span>{{ t('myAccounts.import.title') }}</span>
            </button>
            <button class="btn btn-secondary" data-testid="my-accounts-open-proxies" @click="openProxyModal">
              <Icon name="globe" size="md" />
              <span>我的代理</span>
            </button>
            <button class="btn btn-primary" data-testid="my-accounts-open-create" @click="openCreateModal">
              <Icon name="plus" size="md" />
              <span>{{ t('myAccounts.addAccount') }}</span>
            </button>
          </div>
        </template>

        <template #table>
          <div class="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-lg bg-primary-50 p-3 dark:bg-primary-900/20">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-sm font-medium text-primary-900 dark:text-primary-100">
                {{ selectedIds.length > 0 ? t('myAccounts.bulk.selected', { count: selectedIds.length }) : t('myAccounts.bulk.title') }}
              </span>
              <template v-if="selectedIds.length > 0">
                <button
                  type="button"
                  class="text-xs font-medium text-primary-700 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200"
                  @click="selectCurrentPage"
                >
                  {{ t('myAccounts.bulk.selectCurrentPage') }}
                </button>
                <span class="text-gray-300 dark:text-primary-800">•</span>
                <button
                  type="button"
                  class="text-xs font-medium text-primary-700 hover:text-primary-800 dark:text-primary-300 dark:hover:text-primary-200"
                  @click="clearSelection"
                >
                  {{ t('myAccounts.bulk.clear') }}
                </button>
              </template>
            </div>
            <div class="flex flex-wrap justify-end gap-2">
              <button
                type="button"
                data-testid="my-accounts-bulk-apply-public"
                class="btn btn-primary btn-sm"
                :disabled="bulkSharing || selectedShareableCount === 0"
                @click="bulkApplyPublic"
              >
                <Icon v-if="bulkSharing" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="upload" size="sm" />
                <span>{{ t('myAccounts.bulk.applyPublic', { count: selectedShareableCount }) }}</span>
              </button>
              <button
                type="button"
                data-testid="my-accounts-bulk-make-private"
                class="btn btn-secondary btn-sm"
                :disabled="bulkSharing || selectedPublicCount === 0"
                @click="bulkMakePrivate"
              >
                <Icon v-if="bulkSharing" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="lock" size="sm" />
                <span>{{ t('myAccounts.bulk.makePrivate', { count: selectedPublicCount }) }}</span>
              </button>
            </div>
          </div>
          <DataTable
            :columns="columns"
            :data="accounts"
            :loading="loading"
            :server-side-sort="true"
            default-sort-key="created_at"
            default-sort-order="desc"
            @sort="handleSort"
          >
            <template #header-select>
              <input
                type="checkbox"
                class="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                :checked="allVisibleSelected"
                :disabled="accounts.length === 0 || bulkSharing"
                :aria-label="t('myAccounts.bulk.selectCurrentPage')"
                @change="toggleSelectAllVisible"
                @click.stop
              />
            </template>

            <template #cell-select="{ row }">
              <input
                type="checkbox"
                class="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary-600 focus:ring-primary-500 disabled:cursor-not-allowed disabled:opacity-40"
                :checked="isSelected(row.id)"
                :disabled="bulkSharing"
                :aria-label="t('myAccounts.bulk.selectAccount', { name: row.name })"
                @change="toggleSelection(row.id)"
              />
            </template>

            <template #cell-name="{ row }">
              <div class="flex min-w-[180px] flex-col">
                <span class="font-medium text-gray-900 dark:text-white">{{ row.name }}</span>
                <span v-if="row.notes" class="max-w-[220px] truncate text-xs text-gray-500 dark:text-dark-400" :title="row.notes">
                  {{ row.notes }}
                </span>
              </div>
            </template>

            <template #cell-platform_type="{ row }">
              <PlatformTypeBadge
                :platform="row.platform"
                :type="row.type"
                :plan-type="row.share_display_tier || row.credentials?.plan_type"
                :privacy-mode="row.extra?.privacy_mode"
                :subscription-expires-at="row.credentials?.subscription_expires_at"
              />
            </template>

            <template #cell-share="{ row }">
              <div class="flex min-w-[140px] flex-col gap-1.5">
                <div class="flex flex-wrap items-center gap-1">
                  <span :class="['badge text-xs', row.share_mode === 'public' ? 'badge-primary' : 'badge-secondary']">
                    {{ formatShareMode(row.share_mode) }}
                  </span>
                  <span :class="['badge text-xs', shareStatusClass(row.share_status)]">
                    {{ formatShareStatus(row.share_status) }}
                  </span>
                </div>
                <button
                  class="btn btn-xs btn-secondary"
                  :disabled="shareUpdatingId === row.id || bulkSharing"
                  @click="toggleShareMode(row)"
                >
                  {{ row.share_mode === 'public' ? t('myAccounts.makePrivate') : t('myAccounts.applyPublic') }}
                </button>
              </div>
            </template>

            <template #cell-capacity="{ row }">
              <AccountCapacityCell :account="row" />
            </template>

            <template #cell-status="{ row }">
              <AccountStatusIndicator :account="row" />
            </template>

            <template #cell-schedulable="{ row }">
              <span :class="['badge text-xs', row.schedulable ? 'badge-success' : 'badge-secondary']">
                {{ row.schedulable ? t('myAccounts.schedulable.yes') : t('myAccounts.schedulable.no') }}
              </span>
            </template>

            <template #cell-today_stats="{ row }">
              <div class="text-xs text-gray-600 dark:text-gray-300">
                <div>{{ t('myAccounts.requests') }}: {{ formatNumber(row.current_rpm ?? 0) }}</div>
                <div>{{ t('myAccounts.cost') }}: {{ formatCurrency(row.current_window_cost ?? 0) }}</div>
              </div>
            </template>

            <template #cell-usage="{ row }">
              <AccountUsageCell :account="row" usage-api-scope="user" />
            </template>

            <template #cell-last_used_at="{ value }">
              <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatRelativeTime(value) }}</span>
            </template>

            <template #cell-expires_at="{ value }">
              <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatExpiresAt(value) }}</span>
            </template>

            <template #cell-earnings="{ row }">
              <div class="text-sm text-gray-700 dark:text-gray-300">
                <span v-if="row.share_mode === 'public' && row.share_status === 'active'">
                  {{ t('myAccounts.earningEnabled') }}
                </span>
                <span v-else>-</span>
              </div>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex items-center gap-1">
                <button class="rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400" @click="openEditModal(row)">
                  <Icon name="edit" size="sm" />
                </button>
                <button class="rounded-lg p-1.5 text-gray-500 hover:bg-[#f5f0e8] hover:text-[#5f7f68] dark:hover:bg-[#5f7f68]/20 dark:hover:text-[#9ab3a0]" @click="runTest(row)">
                  <Icon name="play" size="sm" />
                </button>
                <button class="rounded-lg p-1.5 text-gray-500 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400" @click="deleteOwnedAccount(row)">
                  <Icon name="trash" size="sm" />
                </button>
              </div>
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
    </div>

    <div v-if="showAccountModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="flex max-h-[92vh] w-full max-w-4xl flex-col overflow-hidden rounded-2xl bg-white shadow-xl dark:bg-dark-800">
        <div class="flex items-center justify-between border-b border-gray-200 px-6 py-5 dark:border-dark-700">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ editingAccount ? t('myAccounts.editAccount') : t('myAccounts.addAccount') }}
          </h3>
          <button class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 dark:hover:bg-dark-700" @click="closeAccountModal">
            <Icon name="x" size="md" />
          </button>
        </div>

        <div class="overflow-y-auto px-6 py-5">
          <div v-if="!editingAccount" class="mb-6 flex items-center justify-center">
            <div class="flex items-center gap-4">
              <div class="flex items-center">
                <div
                  :class="[
                    'flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold',
                    accountWizardStep >= 1 ? 'bg-primary-500 text-white' : 'bg-gray-200 text-gray-500 dark:bg-dark-600'
                  ]"
                >
                  1
                </div>
                <span class="ml-2 text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('myAccounts.steps.authMethod') }}</span>
              </div>
              <div class="h-0.5 w-8 bg-gray-300 dark:bg-dark-600" />
              <div class="flex items-center">
                <div
                  :class="[
                    'flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold',
                    accountWizardStep >= 2 ? 'bg-primary-500 text-white' : 'bg-gray-200 text-gray-500 dark:bg-dark-600'
                  ]"
                >
                  2
                </div>
                <span class="ml-2 text-sm font-medium text-gray-700 dark:text-gray-300">{{ accountAuthStepTitle }}</span>
              </div>
            </div>
          </div>

          <div v-if="editingAccount || accountWizardStep === 1" class="space-y-5">
            <div>
              <label class="input-label">{{ t('myAccounts.accountName') }}</label>
              <input
                v-model="form.name"
                data-testid="my-accounts-name"
                type="text"
                class="input"
                :placeholder="t('myAccounts.namePlaceholder')"
              />
            </div>

            <div>
              <label class="input-label">{{ t('myAccounts.notes') }}</label>
              <textarea
                v-model="form.notes"
                rows="3"
                class="input"
                :placeholder="t('myAccounts.notesPlaceholder')"
              ></textarea>
              <p class="input-hint">{{ t('myAccounts.notesOptional') }}</p>
            </div>

            <div v-if="!editingAccount">
              <label class="input-label">{{ t('myAccounts.shareMode.title') }}</label>
              <div class="mt-2 grid grid-cols-2 gap-3">
                <button
                  type="button"
                  @click="createShareMode = 'private'"
                  :class="[
                    'flex items-center justify-center gap-2 rounded-lg border px-4 py-3 text-sm font-medium transition-all',
                    createShareMode === 'private'
                      ? 'border-primary-500 bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-300'
                      : 'border-gray-200 text-gray-600 hover:border-primary-300 dark:border-dark-600 dark:text-gray-400'
                  ]"
                >
                  <Icon name="lock" size="sm" />
                  <span>{{ t('myAccounts.shareMode.private') }}</span>
                </button>
                <button
                  type="button"
                  @click="createShareMode = 'public'"
                  :class="[
                    'flex items-center justify-center gap-2 rounded-lg border px-4 py-3 text-sm font-medium transition-all',
                    createShareMode === 'public'
                      ? 'border-primary-500 bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-300'
                      : 'border-gray-200 text-gray-600 hover:border-primary-300 dark:border-dark-600 dark:text-gray-400'
                  ]"
                >
                  <Icon name="globe" size="sm" />
                  <span>{{ t('myAccounts.shareMode.public') }}</span>
                </button>
              </div>
              <p v-if="createShareMode === 'public'" class="mt-2 text-xs text-gray-500 dark:text-dark-400">
                {{ t('myAccounts.shareMode.publicCreateHint') }}
              </p>
            </div>

            <div v-if="!editingAccount">
              <label class="input-label">{{ t('myAccounts.platform') }}</label>
              <div class="mt-2 grid grid-cols-2 gap-2 rounded-lg bg-gray-100 p-1 dark:bg-dark-700 md:grid-cols-4">
                <button
                  v-for="platform in accountPlatformCards"
                  :key="platform.value"
                  type="button"
                  @click="selectAccountPlatform(platform.value)"
                  :class="[
                    'flex min-h-11 items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-all',
                    form.platform === platform.value
                      ? `${platform.activeClass} bg-white shadow-sm dark:bg-dark-600`
                      : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200'
                  ]"
                >
                  <Icon :name="platform.icon" size="sm" />
                  <span>{{ platform.label }}</span>
                </button>
              </div>
            </div>

            <div v-if="!editingAccount">
              <label class="input-label">{{ t('myAccounts.accountType') }}</label>
              <div class="mt-2 grid gap-3" :class="methodCardGridClass">
                <button
                  v-for="method in accountMethodCards"
                  :key="method.value"
                  type="button"
                  @click="selectAccountMethod(method.value)"
                  :class="[
                    'flex items-center gap-3 rounded-lg border-2 p-3 text-left transition-all',
                    form.method === method.value
                      ? `${method.borderClass} ${method.bgClass}`
                      : 'border-gray-200 hover:border-primary-300 dark:border-dark-600 dark:hover:border-primary-700'
                  ]"
                >
                  <div
                    :class="[
                      'flex h-8 w-8 shrink-0 items-center justify-center rounded-lg',
                      form.method === method.value
                        ? `${method.iconBgClass} text-white`
                        : 'bg-gray-100 text-gray-500 dark:bg-dark-600 dark:text-gray-400'
                    ]"
                  >
                    <Icon :name="method.icon" size="sm" />
                  </div>
                  <div>
                    <span class="block text-sm font-medium text-gray-900 dark:text-white">{{ method.label }}</span>
                    <span class="text-xs text-gray-500 dark:text-gray-400">{{ method.description }}</span>
                  </div>
                </button>
              </div>
            </div>

            <div>
              <label class="input-label">{{ t('myAccounts.proxy.title') }}</label>
              <Select v-model="form.proxyId" :options="proxyOptions" />
              <div v-if="!editingAccount" class="mt-2 flex items-center justify-between gap-3 text-xs text-gray-500 dark:text-dark-400">
                <span>{{ t('myAccounts.proxy.hint') }}</span>
                <button type="button" class="font-medium text-primary-600 hover:underline dark:text-primary-400" @click="openProxyModal">
                  {{ t('myAccounts.proxy.manage') }}
                </button>
              </div>
            </div>

            <div v-if="accountAdvancedConfigVisible" class="space-y-5 border-t border-gray-200 pt-5 dark:border-dark-700">
              <div v-if="form.platform === 'openai'">
                <label class="input-label">{{ t('myAccounts.advanced.planTier.title') }}</label>
                <div class="mt-2 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                  <button
                    v-for="tier in accountPlanTierOptions"
                    :key="tier.value"
                    type="button"
                    @click="selectAccountPlanTier(tier.value)"
                    :class="[
                      'rounded-lg border px-3 py-3 text-left transition-all',
                      accountPlanTier === tier.value
                        ? 'border-primary-500 bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-300'
                        : 'border-gray-200 text-gray-600 hover:border-primary-300 dark:border-dark-600 dark:text-gray-400'
                    ]"
                  >
                    <span class="block text-sm font-semibold">{{ tier.label }}</span>
                    <span class="mt-1 block text-xs">{{ tier.description }}</span>
                  </button>
                </div>
                <p class="input-hint">{{ t('myAccounts.advanced.planTier.hint') }}</p>
              </div>

              <div v-if="isModelLimitConfigVisible">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <label class="input-label">{{ t('myAccounts.advanced.modelLimit.title') }}</label>
                    <p class="input-hint">{{ t('myAccounts.advanced.modelLimit.hint') }}</p>
                  </div>
                  <div class="grid w-full grid-cols-2 gap-2 rounded-lg bg-gray-100 p-1 dark:bg-dark-700 sm:w-80">
                    <button
                      type="button"
                      @click="accountModelLimitMode = 'allowlist'"
                      :class="[
                        'rounded-md px-3 py-2 text-sm font-medium transition-all',
                        accountModelLimitMode === 'allowlist' ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-600 dark:text-primary-300' : 'text-gray-600 dark:text-gray-400'
                      ]"
                    >
                      {{ t('myAccounts.advanced.modelLimit.allowlist') }}
                    </button>
                    <button
                      type="button"
                      @click="accountModelLimitMode = 'mapping'"
                      :class="[
                        'rounded-md px-3 py-2 text-sm font-medium transition-all',
                        accountModelLimitMode === 'mapping' ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-600 dark:text-primary-300' : 'text-gray-600 dark:text-gray-400'
                      ]"
                    >
                      {{ t('myAccounts.advanced.modelLimit.mapping') }}
                    </button>
                  </div>
                </div>

                <div v-if="accountModelLimitMode === 'allowlist'" class="mt-3 rounded-lg border border-gray-200 p-3 dark:border-dark-700">
                  <div class="grid max-h-48 gap-2 overflow-y-auto sm:grid-cols-2">
                    <button
                      v-for="model in openAIModelSuggestions"
                      :key="model"
                      type="button"
                      @click="toggleSuggestedModel(model)"
                      :class="[
                        'flex min-h-9 items-center justify-between rounded-md px-3 py-2 text-left text-xs transition-all',
                        accountModelAllowlist.includes(model)
                          ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-300'
                          : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-dark-200 dark:hover:bg-dark-600'
                      ]"
                    >
                      <span class="truncate">{{ model }}</span>
                      <span v-if="accountModelAllowlist.includes(model)" class="ml-2">x</span>
                    </button>
                  </div>
                  <div class="mt-3 flex flex-col gap-2 sm:flex-row">
                    <button type="button" class="btn btn-secondary btn-sm" @click="fillSuggestedModels">
                      {{ t('myAccounts.advanced.modelLimit.fillSuggested') }}
                    </button>
                    <button type="button" class="btn btn-secondary btn-sm text-red-600 dark:text-red-400" @click="clearSelectedModels">
                      {{ t('myAccounts.advanced.modelLimit.clear') }}
                    </button>
                  </div>
                  <div class="mt-3 flex flex-col gap-2 sm:flex-row">
                    <input v-model="accountModelCustomInput" class="input flex-1" :placeholder="t('myAccounts.advanced.modelLimit.customPlaceholder')" @keyup.enter="addCustomModel" />
                    <button type="button" class="btn btn-secondary" @click="addCustomModel">
                      {{ t('myAccounts.advanced.modelLimit.add') }}
                    </button>
                  </div>
                  <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">
                    {{ t('myAccounts.advanced.modelLimit.selected', { count: selectedModelCount }) }}
                  </p>
                </div>

                <div v-else class="mt-3 rounded-lg border border-gray-200 p-3 dark:border-dark-700">
                  <div class="grid gap-2 sm:grid-cols-[1fr_1fr_auto]">
                    <input v-model="accountModelMappingSource" class="input" :placeholder="t('myAccounts.advanced.modelLimit.sourcePlaceholder')" />
                    <input v-model="accountModelMappingTarget" class="input" :placeholder="t('myAccounts.advanced.modelLimit.targetPlaceholder')" @keyup.enter="addModelMapping" />
                    <button type="button" class="btn btn-secondary" @click="addModelMapping">
                      {{ t('myAccounts.advanced.modelLimit.add') }}
                    </button>
                  </div>
                  <div v-if="accountModelMappings.length > 0" class="mt-3 space-y-2">
                    <div
                      v-for="row in accountModelMappings"
                      :key="row.id"
                      class="flex items-center justify-between gap-3 rounded-md bg-gray-50 px-3 py-2 text-sm dark:bg-dark-900"
                    >
                      <span class="min-w-0 flex-1 truncate">{{ row.source }} -> {{ row.target }}</span>
                      <button type="button" class="text-xs font-medium text-red-600 hover:underline dark:text-red-400" @click="removeModelMapping(row.id)">
                        {{ t('common.delete') }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <div class="grid gap-4 md:grid-cols-2">
                <div>
                  <label class="input-label">{{ t('myAccounts.advanced.concurrency') }}</label>
                  <input v-model.number="accountConcurrency" type="number" min="1" class="input" @input="accountConcurrency = normalizePositiveInteger(accountConcurrency, 3)" />
                </div>
                <div>
                  <label class="input-label">{{ t('myAccounts.advanced.expiresAt') }}</label>
                  <input v-model="accountExpiresAtInput" type="datetime-local" class="input" />
                  <label class="mt-2 flex items-center gap-2 text-xs text-gray-600 dark:text-dark-300">
                    <input v-model="accountAutoPauseOnExpired" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                    <span>{{ t('myAccounts.advanced.autoPauseOnExpired') }}</span>
                  </label>
                </div>
              </div>

              <div v-if="form.platform === 'openai'" class="rounded-lg border border-gray-200 p-4 dark:border-dark-700">
                <div class="flex items-center justify-between gap-3">
                  <div>
                    <label class="input-label mb-0">{{ t('myAccounts.advanced.codexLimit.title') }}</label>
                    <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('myAccounts.advanced.codexLimit.hint') }}</p>
                  </div>
                  <button
                    type="button"
                    @click="codexLimitProtectionEnabled = !codexLimitProtectionEnabled"
                    :class="[
                      'relative inline-flex h-6 w-11 shrink-0 rounded-full border-2 border-transparent transition-colors',
                      codexLimitProtectionEnabled ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
                    ]"
                  >
                    <span :class="['inline-block h-5 w-5 rounded-full bg-white shadow transition-transform', codexLimitProtectionEnabled ? 'translate-x-5' : 'translate-x-0']" />
                  </button>
                </div>
                <div v-if="codexLimitProtectionEnabled" class="mt-4 grid gap-4 md:grid-cols-2">
                  <div>
                    <label class="input-label">{{ t('myAccounts.advanced.codexLimit.fiveHour') }}</label>
                    <input v-model.number="codex5hLimitPercent" type="number" min="0" step="1" class="input" />
                  </div>
                  <div>
                    <label class="input-label">{{ t('myAccounts.advanced.codexLimit.sevenDay') }}</label>
                    <input v-model.number="codex7dLimitPercent" type="number" min="0" step="1" class="input" />
                  </div>
                </div>
              </div>

              <div class="space-y-3 rounded-lg border border-gray-200 p-4 dark:border-dark-700">
                <div>
                  <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('myAccounts.advanced.dispatch.title') }}</h4>
                  <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('myAccounts.advanced.dispatch.hint') }}</p>
                </div>
                <div class="flex items-center justify-between gap-3 rounded-lg border border-gray-100 p-3 dark:border-dark-700">
                  <div>
                    <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('myAccounts.advanced.dispatch.rpm') }}</p>
                    <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('myAccounts.advanced.dispatch.rpmHint') }}</p>
                  </div>
                  <button
                    type="button"
                    @click="rpmLimitEnabled = !rpmLimitEnabled"
                    :class="['relative inline-flex h-6 w-11 shrink-0 rounded-full border-2 border-transparent transition-colors', rpmLimitEnabled ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600']"
                  >
                    <span :class="['inline-block h-5 w-5 rounded-full bg-white shadow transition-transform', rpmLimitEnabled ? 'translate-x-5' : 'translate-x-0']" />
                  </button>
                </div>
                <div v-if="rpmLimitEnabled">
                  <label class="input-label">{{ t('myAccounts.advanced.dispatch.baseRpm') }}</label>
                  <input v-model.number="baseRpm" type="number" min="1" class="input" />
                </div>
                <div>
                  <label class="input-label">{{ t('myAccounts.advanced.dispatch.userMsgQueueMode') }}</label>
                  <div class="mt-2 flex flex-wrap gap-2">
                    <button
                      v-for="mode in userMsgQueueModeOptions"
                      :key="mode"
                      type="button"
                      @click="userMsgQueueMode = mode"
                      :class="[
                        'rounded-md border px-3 py-2 text-sm transition-all',
                        userMsgQueueMode === mode ? 'border-primary-500 bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-300' : 'border-gray-200 text-gray-600 dark:border-dark-600 dark:text-gray-400'
                      ]"
                    >
                      {{ t(`myAccounts.advanced.dispatch.queueModes.${mode}`) }}
                    </button>
                  </div>
                </div>
                <div class="grid gap-3 md:grid-cols-3">
                  <label class="flex items-center gap-2 rounded-lg border border-gray-100 p-3 text-sm dark:border-dark-700">
                    <input v-model="tlsFingerprintEnabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                    <span>{{ t('myAccounts.advanced.dispatch.tlsFingerprint') }}</span>
                  </label>
                  <label class="flex items-center gap-2 rounded-lg border border-gray-100 p-3 text-sm dark:border-dark-700">
                    <input v-model="sessionIdMaskingEnabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                    <span>{{ t('myAccounts.advanced.dispatch.sessionMasking') }}</span>
                  </label>
                  <label class="flex items-center gap-2 rounded-lg border border-gray-100 p-3 text-sm dark:border-dark-700">
                    <input v-model="cacheTtlOverrideEnabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                    <span>{{ t('myAccounts.advanced.dispatch.cacheTtl') }}</span>
                  </label>
                </div>
                <div v-if="cacheTtlOverrideEnabled">
                  <label class="input-label">{{ t('myAccounts.advanced.dispatch.cacheTtlTarget') }}</label>
                  <Select v-model="cacheTtlOverrideTarget" :options="[{ value: '5m', label: '5m' }, { value: '1h', label: '1h' }]" />
                </div>
              </div>
            </div>

            <div v-if="editingAccount && !isUserManagedKeyBackedType(editingAccount.type)">
              <label class="input-label">{{ t('myAccounts.import.credentials') }}</label>
              <textarea v-model="credentialsJson" class="input min-h-[160px] w-full font-mono text-xs" :placeholder="t('myAccounts.import.credentialsPlaceholder')"></textarea>
            </div>
          </div>

          <div v-else class="space-y-5">
            <div class="rounded-lg border border-gray-200 bg-gray-50 px-4 py-3 dark:border-dark-700 dark:bg-dark-900">
              <div class="flex flex-wrap items-center gap-2 text-sm text-gray-700 dark:text-dark-200">
                <span class="font-medium text-gray-900 dark:text-white">{{ selectedPlatformLabel }}</span>
                <span class="text-gray-400">/</span>
                <span>{{ selectedMethodLabel }}</span>
                <span v-if="selectedProxyLabel" class="text-gray-400">/</span>
                <span v-if="selectedProxyLabel">{{ selectedProxyLabel }}</span>
              </div>
            </div>

            <div v-if="form.method === 'oauth'" class="rounded-xl border border-gray-200 p-4 dark:border-dark-700">
              <div class="flex flex-col gap-3 md:flex-row md:items-end">
                <Input v-model="oauthCode" class="flex-1" :label="t('myAccounts.oauthCode')" :placeholder="t('myAccounts.oauthCodePlaceholder')" />
                <button class="btn btn-secondary" :disabled="authUrlLoading" @click="generateAuthUrl">
                  <Icon v-if="authUrlLoading" name="refresh" size="sm" class="animate-spin" />
                  <Icon v-else name="externalLink" size="sm" />
                  <span>{{ t('myAccounts.generateAuthUrl') }}</span>
                </button>
              </div>
              <div v-if="authUrl" class="mt-3 rounded-lg bg-gray-50 p-3 text-sm dark:bg-dark-900">
                <a :href="authUrl" target="_blank" rel="noopener noreferrer" class="break-all text-primary-600 hover:underline dark:text-primary-400">
                  {{ authUrl }}
                </a>
              </div>
              <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">{{ t('myAccounts.oauthHint') }}</p>
            </div>

            <div v-else-if="form.method === 'session-key' || form.method === 'setup-token'">
              <label class="input-label">{{ t('myAccounts.sessionKey') }}</label>
              <textarea v-model="sessionKey" class="input min-h-[160px] w-full" :placeholder="t('myAccounts.sessionKeyPlaceholder')"></textarea>
            </div>

            <div v-else>
              <label class="input-label">{{ t('myAccounts.import.credentials') }}</label>
              <textarea v-model="credentialsJson" class="input min-h-[180px] w-full font-mono text-xs" :placeholder="t('myAccounts.import.credentialsPlaceholder')"></textarea>
              <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">
                {{ t('myAccounts.tokenJsonHint') }}
              </p>
            </div>
          </div>

          <div v-if="isAccountQuotaConfigVisible" class="mt-5 space-y-4 rounded-xl border border-gray-200 p-4 dark:border-dark-700">
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.accounts.quotaControl.title') }}</h4>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('admin.accounts.quotaLimitHint') }}</p>
          </div>
          <div class="grid gap-4 md:grid-cols-4">
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaDailyLimit') }}</label>
              <input v-model.number="quotaDailyLimit" data-testid="my-accounts-quota-daily" type="number" min="0" step="0.01" class="input" />
            </div>
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaWeeklyLimit') }}</label>
              <input v-model.number="quotaWeeklyLimit" data-testid="my-accounts-quota-weekly" type="number" min="0" step="0.01" class="input" />
            </div>
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaMonthlyLimit') }}</label>
              <input v-model.number="quotaMonthlyLimit" data-testid="my-accounts-quota-monthly" type="number" min="0" step="0.01" class="input" />
            </div>
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaTotalLimit') }}</label>
              <input v-model.number="quotaTotalLimit" data-testid="my-accounts-quota-total" type="number" min="0" step="0.01" class="input" />
            </div>
          </div>
          <ShareDisplayCard
            v-if="form.platform === 'openai'"
            :enabled="shareDisplayEnabled"
            :display-name="shareDisplayName"
            :display-tier="shareDisplayTier"
            :percent-only="shareDisplayPercentOnly"
            :account-count="shareDisplayAccountCount"
            :display-5h-limit="shareDisplay5hLimit"
            :display-5h-used="shareDisplay5hUsed"
            :display-7d-limit="shareDisplay7dLimit"
            :display-7d-used="shareDisplay7dUsed"
            @update:enabled="shareDisplayEnabled = $event"
            @update:displayName="shareDisplayName = $event"
            @update:displayTier="shareDisplayTier = $event"
            @update:percentOnly="shareDisplayPercentOnly = $event"
            @update:accountCount="shareDisplayAccountCount = $event"
            @update:display5hLimit="shareDisplay5hLimit = $event"
            @update:display5hUsed="shareDisplay5hUsed = $event"
            @update:display7dLimit="shareDisplay7dLimit = $event"
            @update:display7dUsed="shareDisplay7dUsed = $event"
          />
        </div>

          <div class="mt-6 flex justify-end gap-3 border-t border-gray-200 pt-5 dark:border-dark-700">
            <button class="btn btn-secondary" @click="closeAccountModal">{{ t('common.cancel') }}</button>
            <button v-if="!editingAccount && accountWizardStep === 2" class="btn btn-secondary" @click="accountWizardStep = 1">
              {{ t('common.back') }}
            </button>
            <button
              v-if="!editingAccount && accountWizardStep === 1"
              class="btn btn-primary"
              data-testid="my-accounts-next"
              @click="goToAccountAuthStep"
            >
              <span>{{ t('common.next') }}</span>
            </button>
            <button v-else class="btn btn-primary" data-testid="my-accounts-save" :disabled="savingAccount" @click="saveAccount">
              <Icon v-if="savingAccount" name="refresh" size="sm" class="animate-spin" />
              <span>{{ t('common.save') }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showImportModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="w-full max-w-2xl rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-800">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('myAccounts.import.title') }}</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('myAccounts.import.description') }}</p>
          </div>
          <button class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 dark:hover:bg-dark-700" @click="showImportModal = false">
            <Icon name="x" size="md" />
          </button>
        </div>
        <div class="mt-5 grid gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('myAccounts.platform') }}</label>
            <Select v-model="importForm.platform" :options="platformOptions" />
          </div>
          <div>
            <label class="input-label">{{ t('myAccounts.import.format') }}</label>
            <Select v-model="importForm.format" :options="importFormatOptions" />
          </div>
        </div>
        <div class="mt-4">
          <label class="input-label">代理</label>
          <Select v-model="importForm.proxyId" :options="proxyOptions" />
        </div>
        <div class="mt-4">
          <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
            <label class="input-label mb-0">{{ t('myAccounts.import.content') }}</label>
            <div class="flex flex-wrap items-center gap-2">
              <input
                ref="importFileInput"
                data-testid="my-accounts-import-file-input"
                type="file"
                class="hidden"
                accept=".json,.txt,.token,.key,application/json,text/plain"
                @change="handleImportFileChange"
              />
              <input
                ref="importFolderInput"
                data-testid="my-accounts-import-folder-input"
                type="file"
                class="hidden"
                accept=".json,.txt,.token,.key,application/json,text/plain"
                multiple
                webkitdirectory
                directory
                @change="handleImportFolderChange"
              />
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="importFileReading"
                @click="openImportFilePicker"
              >
                <Icon v-if="importFileReading" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="upload" size="sm" />
                <span>{{ t('myAccounts.import.chooseFile') }}</span>
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="importFileReading"
                @click="openImportFolderPicker"
              >
                <Icon v-if="importFileReading" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="upload" size="sm" />
                <span>{{ t('myAccounts.import.chooseFolder') }}</span>
              </button>
              <span v-if="importFolderFiles.length" class="max-w-[260px] truncate text-xs text-gray-500 dark:text-dark-400">
                {{ t('myAccounts.import.folderSelected', { count: importFolderFiles.length }) }}
              </span>
              <span v-else-if="importFileName" class="max-w-[260px] truncate text-xs text-gray-500 dark:text-dark-400" :title="importFileName">
                {{ t('myAccounts.import.fileSelected', { name: importFileName }) }}
              </span>
            </div>
          </div>
          <textarea
            v-model="importContent"
            data-testid="my-accounts-import-content"
            class="input min-h-[220px] w-full font-mono text-xs"
            :placeholder="t('myAccounts.import.contentPlaceholder')"
          ></textarea>
        </div>
        <div class="mt-6 flex justify-end gap-3">
          <button class="btn btn-secondary" @click="showImportModal = false">{{ t('common.cancel') }}</button>
          <button data-testid="my-accounts-import-submit" class="btn btn-primary" :disabled="importing || importFileReading" @click="importFromContent">
            <Icon v-if="importing" name="refresh" size="sm" class="animate-spin" />
            <span>{{ t('myAccounts.import.submit') }}</span>
          </button>
        </div>
      </div>
    </div>

    <div v-if="showProxyModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-800">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">我的代理</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">这些代理只对你自己的共享账号可见。</p>
          </div>
          <button class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 dark:hover:bg-dark-700" @click="showProxyModal = false">
            <Icon name="x" size="md" />
          </button>
        </div>

        <div class="mt-5">
          <label class="input-label">{{ t('myAccounts.proxy.smartInputLabel') }}</label>
          <textarea
            v-model="proxyInput"
            data-testid="my-accounts-proxy-smart-input"
            rows="3"
            class="input font-mono text-sm"
            :placeholder="t('myAccounts.proxy.smartInputPlaceholder')"
          ></textarea>
          <div class="mt-2 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-dark-400">
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              data-testid="my-accounts-proxy-smart-parse"
              @click="applyProxyInput"
            >
              <Icon name="refresh" size="sm" />
              <span>{{ t('myAccounts.proxy.smartInputButton') }}</span>
            </button>
            <span>{{ t('myAccounts.proxy.smartInputHint') }}</span>
          </div>
          <div
            v-if="proxyBatch.length > 0"
            class="mt-3 flex flex-wrap items-center justify-between gap-3 rounded-lg border border-primary-200 bg-primary-50 px-3 py-2 text-sm text-primary-700 dark:border-primary-900/60 dark:bg-primary-950/30 dark:text-primary-300"
          >
            <span>
              {{ t('myAccounts.proxy.smartInputBatchDetected', { count: proxyBatch.length }) }}
              <span v-if="proxyBatchDuplicateCount > 0" class="text-xs opacity-80">
                {{ t('myAccounts.proxy.smartInputBatchDuplicates', { count: proxyBatchDuplicateCount }) }}
              </span>
            </span>
            <button
              type="button"
              class="btn btn-primary btn-sm"
              data-testid="my-accounts-proxy-smart-batch-save"
              :disabled="savingProxy"
              @click="saveProxyBatch"
            >
              <Icon v-if="savingProxy" name="refresh" size="sm" class="animate-spin" />
              <Icon v-else name="plus" size="sm" />
              <span>{{ t('myAccounts.proxy.smartInputBatchSave', { count: proxyBatch.length }) }}</span>
            </button>
          </div>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2">
          <Input v-model="proxyForm.name" label="名称" placeholder="美国住宅代理" />
          <div>
            <label class="input-label">协议</label>
            <Select v-model="proxyForm.protocol" :options="proxyProtocolOptions" />
          </div>
          <Input v-model="proxyForm.host" label="主机" placeholder="127.0.0.1" />
          <div>
            <label class="input-label">端口</label>
            <input v-model.number="proxyForm.port" type="number" min="1" max="65535" class="input" placeholder="7890" />
          </div>
          <Input v-model="proxyForm.username" label="用户名" placeholder="可选" />
          <Input v-model="proxyForm.password" label="密码" type="password" placeholder="新增可选，编辑留空不修改" />
        </div>

        <div class="mt-5 flex justify-end gap-3">
          <button class="btn btn-secondary" @click="resetProxyForm">清空</button>
          <button class="btn btn-primary" :disabled="savingProxy || proxyBatch.length > 0" @click="saveProxy">
            <Icon v-if="savingProxy" name="refresh" size="sm" class="animate-spin" />
            <span>{{ editingProxyId ? '保存代理' : '新增代理' }}</span>
          </button>
        </div>

        <div class="mt-6 overflow-hidden rounded-xl border border-gray-200 dark:border-dark-700">
          <div v-if="proxyLoading" class="p-5 text-sm text-gray-500 dark:text-dark-400">加载中...</div>
          <div v-else-if="userProxies.length === 0" class="p-5 text-sm text-gray-500 dark:text-dark-400">暂无代理</div>
          <table v-else class="w-full divide-y divide-gray-200 text-sm dark:divide-dark-700">
            <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-900 dark:text-dark-400">
              <tr>
                <th class="px-4 py-3">名称</th>
                <th class="px-4 py-3">地址</th>
                <th class="px-4 py-3">状态</th>
                <th class="px-4 py-3 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="proxy in userProxies" :key="proxy.id">
                <td class="px-4 py-3 font-medium text-gray-900 dark:text-white">{{ proxy.name }}</td>
                <td class="px-4 py-3 text-gray-600 dark:text-dark-300">{{ proxy.protocol }}://{{ proxy.host }}:{{ proxy.port }}</td>
                <td class="px-4 py-3">
                  <span :class="['badge text-xs', proxy.status === 'active' ? 'badge-success' : 'badge-secondary']">
                    {{ proxy.status === 'active' ? '启用' : '停用' }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <div class="flex justify-end gap-2">
                    <button class="btn btn-xs btn-secondary" @click="editProxy(proxy)">编辑</button>
                    <button class="btn btn-xs btn-danger" @click="removeProxy(proxy)">删除</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Input from '@/components/common/Input.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformTypeBadge from '@/components/common/PlatformTypeBadge.vue'
import AccountCapacityCell from '@/components/account/AccountCapacityCell.vue'
import AccountStatusIndicator from '@/components/account/AccountStatusIndicator.vue'
import AccountUsageCell from '@/components/account/AccountUsageCell.vue'
import ShareDisplayCard from '@/components/account/ShareDisplayCard.vue'
import userAPI from '@/api/user'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { formatCurrency, formatDateTime, formatDateTimeLocalInput, formatNumber, formatRelativeTime, parseDateTimeLocalInput } from '@/utils/format'
import { formatCreditAmount } from '@/utils/credits'
import { extractApiErrorMessage } from '@/utils/apiError'
import { parseOAuthCallbackInput } from '@/utils/oauthCallback'
import { parseProxyInput, type ParsedProxyInput } from '@/utils/proxyInput'
import { useTableSelection } from '@/composables/useTableSelection'
import type { Account, AccountPlatform, AccountShareMode, AccountShareStatus, Proxy, ProxyProtocol, UserAccountAuthURLRequest, UserAccountShareSummary } from '@/types'
import type { Column } from '@/components/common/types'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const accounts = ref<Account[]>([])
const loading = ref(false)
const transferring = ref(false)
const savingAccount = ref(false)
const importing = ref(false)
const importFileReading = ref(false)
const authUrlLoading = ref(false)
const bulkSharing = ref(false)
const proxyLoading = ref(false)
const savingProxy = ref(false)
const shareUpdatingId = ref<number | null>(null)
const shareSummary = ref<UserAccountShareSummary | null>(null)
const showAccountModal = ref(false)
const showImportModal = ref(false)
const showProxyModal = ref(false)
const editingAccount = ref<Account | null>(null)
const userProxies = ref<Proxy[]>([])
const editingProxyId = ref<number | null>(null)
const accountWizardStep = ref(1)
const createShareMode = ref<AccountShareMode>('private')
const authUrl = ref('')
const authSessionId = ref('')
const authState = ref('')
const oauthCode = ref('')
const sessionKey = ref('')
const credentialsJson = ref('')
const apiKeyBaseUrl = ref('https://api.openai.com')
const apiKeyValue = ref('')
const quotaDailyLimit = ref<number | null>(null)
const quotaWeeklyLimit = ref<number | null>(null)
const quotaMonthlyLimit = ref<number | null>(null)
const quotaTotalLimit = ref<number | null>(null)
const shareDisplayEnabled = ref(false)
const shareDisplayName = ref('')
const shareDisplayTier = ref('pro')
const shareDisplayPercentOnly = ref(true)
const shareDisplayAccountCount = ref(1)
const shareDisplay5hLimit = ref<number | null>(null)
const shareDisplay5hUsed = ref<number | null>(null)
const shareDisplay7dLimit = ref<number | null>(null)
const shareDisplay7dUsed = ref<number | null>(null)
const accountPlanTier = ref('plus')
const accountModelLimitMode = ref<'allowlist' | 'mapping'>('allowlist')
const accountModelAllowlist = ref<string[]>(['gpt-5.2', 'gpt-5.2-chat-latest', 'gpt-5.2-pro', 'gpt-5.5', 'gpt-5.4', 'gpt-5.4-mini', 'gpt-5.3-codex'])
const accountModelCustomInput = ref('')
const accountModelMappingSource = ref('')
const accountModelMappingTarget = ref('')
const accountModelMappings = ref<Array<{ id: number; source: string; target: string }>>([])
const accountConcurrency = ref(3)
const accountExpiresAt = ref<number | null>(null)
const accountAutoPauseOnExpired = ref(true)
const codexLimitProtectionEnabled = ref(true)
const codex5hLimitPercent = ref<number | null>(100)
const codex7dLimitPercent = ref<number | null>(100)
const rpmLimitEnabled = ref(false)
const baseRpm = ref<number | null>(15)
const userMsgQueueMode = ref<'off' | 'throttle' | 'serialize'>('off')
const tlsFingerprintEnabled = ref(false)
const sessionIdMaskingEnabled = ref(false)
const cacheTtlOverrideEnabled = ref(false)
const cacheTtlOverrideTarget = ref<'5m' | '1h'>('5m')
let modelMappingRowId = 1

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})
const sort = reactive<{ sort_by: string; sort_order: 'asc' | 'desc' }>({
  sort_by: 'created_at',
  sort_order: 'desc'
})

const form = reactive({
  name: '',
  notes: '',
  platform: 'openai',
  method: 'oauth',
  proxyId: null as number | null
})
const importForm = reactive({
  platform: 'openai',
  format: 'sub2api_oauth_json',
  proxyId: null as number | null
})
const proxyForm = reactive({
  name: '',
  protocol: 'http' as ProxyProtocol,
  host: '',
  port: null as number | null,
  username: '',
  password: '',
  status: 'active' as 'active' | 'inactive'
})
const proxyInput = ref('')
const proxyBatch = ref<ParsedProxyInput[]>([])
const proxyBatchDuplicateCount = ref(0)
const importContent = ref('')
const importFileInput = ref<HTMLInputElement | null>(null)
const importFolderInput = ref<HTMLInputElement | null>(null)
const importFileName = ref('')
const importFolderFiles = ref<ImportFileEntry[]>([])

type ImportFileEntry = {
  name: string
  content: string
  platform: AccountPlatform
  format: string
}

type UserAccountMethod = 'oauth' | 'setup-token' | 'session-key' | 'json'
type IconName = 'sparkles' | 'bolt' | 'cloud' | 'key' | 'lock' | 'document' | 'globe'

type AccountPlatformCard = {
  value: AccountPlatform
  label: string
  icon: IconName
  activeClass: string
}

type AccountMethodCard = {
  value: UserAccountMethod
  label: string
  description: string
  icon: IconName
  borderClass: string
  bgClass: string
  iconBgClass: string
}

type AccountPlanTierOption = {
  value: string
  label: string
  description: string
}

const openAIModelSuggestions = [
  'gpt-5.2',
  'gpt-5.2-2025-12-11',
  'gpt-5.2-chat-latest',
  'gpt-5.2-pro',
  'gpt-5.2-pro-2025-12-11',
  'gpt-5.5',
  'gpt-5.4',
  'gpt-5.4-mini',
  'gpt-5.4-2026-03-05',
  'gpt-5.3-codex',
  'gpt-5.3-codex-spark',
  'codex-auto-review',
  'gpt-4o-audio-preview',
  'gpt-4o-realtime-preview',
  'gpt-image-1',
  'gpt-image-1.5',
  'gpt-image-2',
]

const accountPlanTierOptions: AccountPlanTierOption[] = [
  { value: 'free', label: 'Free', description: '无法确认等级时按 Free 处理' },
  { value: 'plus', label: 'Plus', description: '常规 ChatGPT Plus 账号' },
  { value: 'pro', label: 'Pro', description: '高配账号或 Pro 订阅' },
  { value: 'team', label: 'Team', description: '团队账号或 JSON 导入' },
]
const userMsgQueueModeOptions: Array<'off' | 'throttle' | 'serialize'> = ['off', 'throttle', 'serialize']

const {
  selectedIds,
  allVisibleSelected,
  isSelected,
  setSelectedIds,
  toggle: toggleSelection,
  clear: clearSelection,
  removeMany: removeSelectedAccounts,
  toggleVisible,
  selectVisible: selectCurrentPage
} = useTableSelection<Account>({
  rows: accounts,
  getId: (account) => account.id
})

const columns = computed<Column[]>(() => [
  { key: 'select', label: '', sortable: false, class: 'w-12' },
  { key: 'name', label: t('myAccounts.columns.name'), sortable: true },
  { key: 'platform_type', label: t('myAccounts.columns.platformType'), sortable: false },
  { key: 'share', label: t('myAccounts.columns.share'), sortable: false },
  { key: 'capacity', label: t('myAccounts.columns.capacity'), sortable: false },
  { key: 'status', label: t('myAccounts.columns.status'), sortable: true },
  { key: 'schedulable', label: t('myAccounts.columns.schedulable'), sortable: true },
  { key: 'today_stats', label: t('myAccounts.columns.todayStats'), sortable: false },
  { key: 'usage', label: t('myAccounts.columns.usageWindows'), sortable: false },
  { key: 'last_used_at', label: t('myAccounts.columns.lastUsed'), sortable: true },
  { key: 'expires_at', label: t('myAccounts.columns.expiresAt'), sortable: true },
  { key: 'earnings', label: t('myAccounts.columns.earnings'), sortable: false },
  { key: 'actions', label: t('myAccounts.columns.actions'), sortable: false }
])

const platformOptions = computed(() => [
  { value: 'openai', label: 'OpenAI / ChatGPT' },
  { value: 'anthropic', label: 'Claude' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'antigravity', label: 'Antigravity' }
])

const accountPlatformCards = computed<AccountPlatformCard[]>(() => [
  { value: 'anthropic', label: 'Anthropic', icon: 'sparkles', activeClass: 'text-orange-600 dark:text-orange-400' },
  { value: 'openai', label: 'OpenAI', icon: 'bolt', activeClass: 'text-green-600 dark:text-green-400' },
  { value: 'gemini', label: 'Gemini', icon: 'sparkles', activeClass: 'text-blue-600 dark:text-blue-400' },
  { value: 'antigravity', label: 'Antigravity', icon: 'cloud', activeClass: 'text-purple-600 dark:text-purple-400' }
])

const accountMethodCards = computed<AccountMethodCard[]>(() => {
  const palette = platformMethodPalette.value
  if (form.platform === 'anthropic') {
    return [
      {
        value: 'oauth',
        label: 'Claude Code',
        description: 'OAuth',
        icon: 'sparkles',
        ...palette
      },
      {
        value: 'setup-token',
        label: 'Setup Token',
        description: t('myAccounts.methodDescriptions.setupToken'),
        icon: 'key',
        ...palette
      },
      {
        value: 'session-key',
        label: 'Session Key',
        description: t('myAccounts.methodDescriptions.sessionKey'),
        icon: 'lock',
        ...palette
      },
      {
        value: 'json',
        label: t('myAccounts.import.jsonToken'),
        description: t('myAccounts.methodDescriptions.json'),
        icon: 'document',
        ...palette
      }
    ]
  }
  return [
    {
      value: 'oauth',
      label: selectedPlatformLabel.value,
      description: 'OAuth',
      icon: form.platform === 'antigravity' ? 'cloud' : 'key',
      ...palette
    },
    {
      value: 'json',
      label: t('myAccounts.import.jsonToken'),
      description: t('myAccounts.methodDescriptions.json'),
      icon: 'document',
      ...palette
    }
  ]
})

const platformMethodPalette = computed(() => {
  switch (form.platform) {
    case 'anthropic':
      return {
        borderClass: 'border-orange-500',
        bgClass: 'bg-orange-50 dark:bg-orange-900/20',
        iconBgClass: 'bg-orange-500'
      }
    case 'gemini':
      return {
        borderClass: 'border-blue-500',
        bgClass: 'bg-blue-50 dark:bg-blue-900/20',
        iconBgClass: 'bg-blue-500'
      }
    case 'antigravity':
      return {
        borderClass: 'border-purple-500',
        bgClass: 'bg-purple-50 dark:bg-purple-900/20',
        iconBgClass: 'bg-purple-500'
      }
    default:
      return {
        borderClass: 'border-green-500',
        bgClass: 'bg-green-50 dark:bg-green-900/20',
        iconBgClass: 'bg-green-500'
      }
  }
})

const methodCardGridClass = computed(() => accountMethodCards.value.length > 2 ? 'sm:grid-cols-2' : 'grid-cols-1 sm:grid-cols-2')

const selectedPlatformLabel = computed(() => (
  accountPlatformCards.value.find(platform => platform.value === form.platform)?.label ??
  platformOptions.value.find(platform => platform.value === form.platform)?.label ??
  form.platform
))

const selectedMethodLabel = computed(() => (
  accountMethodCards.value.find(method => method.value === form.method)?.label ??
  methodOptions.value.find(method => method.value === form.method)?.label ??
  form.method
))

const selectedProxyLabel = computed(() => {
  const id = selectedProxyIdPayload(form.proxyId)
  if (id == null) return ''
  return userProxies.value.find(proxy => proxy.id === id)?.name ?? ''
})

const accountAdvancedConfigVisible = computed(() => {
  return editingAccount.value !== null || accountWizardStep.value === 1
})

const isModelLimitConfigVisible = computed(() => form.platform === 'openai')

const accountExpiresAtInput = computed({
  get: () => formatDateTimeLocalInput(accountExpiresAt.value),
  set: (value: string) => {
    accountExpiresAt.value = parseDateTimeLocalInput(value)
  }
})

const accountAuthStepTitle = computed(() => {
  if (form.platform === 'anthropic') {
    return t('myAccounts.steps.claudeAuth')
  }
  return t('myAccounts.steps.accountAuth', { platform: selectedPlatformLabel.value })
})

const methodOptions = computed(() => {
  if (form.platform === 'anthropic') {
    return [
      { value: 'oauth', label: 'Claude OAuth' },
      { value: 'setup-token', label: 'Claude Setup Token' },
      { value: 'session-key', label: 'Claude Session Key' },
      { value: 'json', label: t('myAccounts.import.jsonToken') }
    ]
  }
  return [
    { value: 'oauth', label: t('myAccounts.oauth') },
    { value: 'json', label: t('myAccounts.import.jsonToken') }
  ]
})

const proxyOptions = computed(() => [
  { value: null, label: '不使用代理' },
  ...userProxies.value
    .filter(proxy => proxy.status === 'active' || proxy.id === form.proxyId || proxy.id === importForm.proxyId)
    .map(proxy => ({
      value: proxy.id,
      label: `${proxy.name} (${proxy.protocol}://${proxy.host}:${proxy.port})`
    }))
])

const proxyProtocolOptions = computed(() => [
  { value: 'http', label: 'HTTP' },
  { value: 'https', label: 'HTTPS' },
  { value: 'socks5', label: 'SOCKS5' },
  { value: 'socks5h', label: 'SOCKS5H' }
])

watch(
  () => form.platform,
  (platform) => {
    if (platform !== 'anthropic' && (form.method === 'setup-token' || form.method === 'session-key')) {
      form.method = 'oauth'
    }
    if (!accountMethodCards.value.some(method => method.value === form.method)) {
      form.method = 'oauth'
    }
    if (!editingAccount.value) {
      apiKeyBaseUrl.value = platform === 'openai' ? 'https://api.openai.com' : ''
    }
  }
)

const isAccountQuotaConfigVisible = computed(() => {
  if (editingAccount.value) {
    return isUserManagedKeyBackedType(editingAccount.value.type) || editingAccount.value.type === 'bedrock'
  }
  return false
})

const selectedModelCount = computed(() => accountModelAllowlist.value.length)

const importFormatOptions = computed(() => [
  { value: 'sub2api_oauth_json', label: 'Sub2API OAuth JSON' },
  { value: 'codex_manager_chatgpt_token_json', label: 'Codex-Manager ChatGPT Token JSON' },
  { value: 'openai_refresh_token', label: 'OpenAI Refresh Token' },
  { value: 'claude_session_key', label: 'Claude Session Key' },
  { value: 'advanced_json', label: t('myAccounts.import.advancedJson') }
])

const selectedAccounts = computed(() => {
  const selected = new Set(selectedIds.value)
  return accounts.value.filter(account => selected.has(account.id))
})

const selectedShareableAccounts = computed(() =>
  selectedAccounts.value.filter(account => account.share_mode !== 'public')
)

const selectedShareableCount = computed(() => selectedShareableAccounts.value.length)

const selectedPublicAccounts = computed(() =>
  selectedAccounts.value.filter(account => account.share_mode === 'public')
)

const selectedPublicCount = computed(() => selectedPublicAccounts.value.length)

async function loadAccounts(): Promise<void> {
  loading.value = true
  try {
    const result = await userAPI.listAccounts(pagination.page, pagination.page_size, sort)
    accounts.value = result.items || []
    pagination.total = result.total || 0
    pagination.pages = result.pages || 0
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('myAccounts.failedToLoad')))
  } finally {
    loading.value = false
  }
}

async function loadShareSummary(): Promise<void> {
  try {
    shareSummary.value = await userAPI.getAccountShareSummary()
  } catch {
    shareSummary.value = null
  }
}

async function loadProxies(): Promise<void> {
  proxyLoading.value = true
  try {
    userProxies.value = await userAPI.listProxies()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '代理加载失败'))
  } finally {
    proxyLoading.value = false
  }
}

async function loadAll(): Promise<void> {
  await Promise.all([loadAccounts(), loadShareSummary(), loadProxies()])
}

function handlePageChange(page: number): void {
  pagination.page = page
  clearSelection()
  loadAccounts()
}

function handlePageSizeChange(size: number): void {
  pagination.page_size = size
  pagination.page = 1
  clearSelection()
  loadAccounts()
}

function handleSort(key: string, order: 'asc' | 'desc'): void {
  sort.sort_by = key
  sort.sort_order = order
  pagination.page = 1
  clearSelection()
  loadAccounts()
}

function toggleSelectAllVisible(event: Event): void {
  const target = event.target as HTMLInputElement
  toggleVisible(target.checked)
}

function resetForm(): void {
  form.name = ''
  form.notes = ''
  form.platform = 'openai'
  form.method = 'oauth'
  form.proxyId = null
  accountWizardStep.value = 1
  createShareMode.value = 'private'
  oauthCode.value = ''
  sessionKey.value = ''
  credentialsJson.value = ''
  apiKeyBaseUrl.value = form.platform === 'openai' ? 'https://api.openai.com' : ''
  apiKeyValue.value = ''
  resetAccountExtraForm()
  authUrl.value = ''
  authSessionId.value = ''
  authState.value = ''
}

function openCreateModal(): void {
  editingAccount.value = null
  resetForm()
  showAccountModal.value = true
}

function openEditModal(account: Account): void {
  editingAccount.value = account
  accountWizardStep.value = 1
  form.name = account.name
  form.notes = account.notes ?? ''
  form.platform = account.platform
  form.method = 'json'
  form.proxyId = account.proxy_id ?? null
  credentialsJson.value = JSON.stringify(account.credentials ?? {}, null, 2)
  const credentials = account.credentials ?? {}
  apiKeyBaseUrl.value = String(credentials.base_url || (account.platform === 'openai' ? 'https://api.openai.com' : ''))
  apiKeyValue.value = ''
  loadAccountExtraForm(account.extra ?? {}, account)
  showAccountModal.value = true
}

function openImportModal(): void {
  importContent.value = ''
  importFileName.value = ''
  importFolderFiles.value = []
  if (importFileInput.value) importFileInput.value.value = ''
  if (importFolderInput.value) importFolderInput.value.value = ''
  importForm.platform = 'openai'
  importForm.format = 'sub2api_oauth_json'
  importForm.proxyId = null
  showImportModal.value = true
}

function closeAccountModal(): void {
  showAccountModal.value = false
  editingAccount.value = null
  accountWizardStep.value = 1
}

function selectAccountPlatform(platform: AccountPlatform): void {
  form.platform = platform
  if (!accountMethodCards.value.some(method => method.value === form.method)) {
    form.method = 'oauth'
  }
  authUrl.value = ''
  authSessionId.value = ''
  authState.value = ''
  oauthCode.value = ''
}

function selectAccountMethod(method: UserAccountMethod): void {
  form.method = method
  authUrl.value = ''
  authSessionId.value = ''
  authState.value = ''
  oauthCode.value = ''
  sessionKey.value = ''
  credentialsJson.value = ''
}

function goToAccountAuthStep(): void {
  if (!form.name.trim()) {
    form.name = defaultAccountName()
  }
  accountWizardStep.value = 2
}

function resetProxyForm(): void {
  editingProxyId.value = null
  proxyForm.name = ''
  proxyForm.protocol = 'http'
  proxyForm.host = ''
  proxyForm.port = null
  proxyForm.username = ''
  proxyForm.password = ''
  proxyForm.status = 'active'
  proxyInput.value = ''
  proxyBatch.value = []
  proxyBatchDuplicateCount.value = 0
}

async function openProxyModal(): Promise<void> {
  resetProxyForm()
  showProxyModal.value = true
  await loadProxies()
}

function editProxy(proxy: Proxy): void {
  editingProxyId.value = proxy.id
  proxyForm.name = proxy.name
  proxyForm.protocol = proxy.protocol
  proxyForm.host = proxy.host
  proxyForm.port = proxy.port
  proxyForm.username = proxy.username ?? ''
  proxyForm.password = ''
  proxyForm.status = proxy.status
  proxyInput.value = ''
  proxyBatch.value = []
  proxyBatchDuplicateCount.value = 0
}

function proxyInputLines(raw: string): string[] {
  return raw
    .split(/\r?\n/)
    .map(line => line.trim())
    .filter(Boolean)
}

function proxyDisplayName(parsed: ParsedProxyInput): string {
  const displayHost = parsed.host.includes(':') ? `[${parsed.host}]` : parsed.host
  return `${displayHost}:${parsed.port}`
}

function proxyBatchKey(parsed: ParsedProxyInput): string {
  return `${parsed.protocol}|${parsed.host}|${parsed.port}|${parsed.username}|${parsed.password}`
}

function applyProxyInput(): void {
  const lines = proxyInputLines(proxyInput.value)
  proxyBatch.value = []
  proxyBatchDuplicateCount.value = 0

  if (lines.length === 0) {
    appStore.showError(t('myAccounts.proxy.smartInputInvalid'))
    return
  }

  if (lines.length > 1) {
    if (editingProxyId.value) {
      appStore.showError(t('myAccounts.proxy.smartInputBatchEditInvalid'))
      return
    }

    const parsedBatch: ParsedProxyInput[] = []
    const seen = new Set<string>()
    let invalidCount = 0
    let duplicateCount = 0
    for (const line of lines) {
      const parsed = parseProxyInput(line, { defaultProtocol: proxyForm.protocol })
      if (!parsed) {
        invalidCount++
        continue
      }
      const key = proxyBatchKey(parsed)
      if (seen.has(key)) {
        duplicateCount++
        continue
      }
      seen.add(key)
      parsedBatch.push(parsed)
    }

    if (invalidCount > 0 || parsedBatch.length === 0) {
      appStore.showError(t('myAccounts.proxy.smartInputBatchInvalid', { count: invalidCount || lines.length }))
      return
    }

    proxyBatch.value = parsedBatch
    proxyBatchDuplicateCount.value = duplicateCount
    proxyInput.value = ''
    appStore.showSuccess(t('myAccounts.proxy.smartInputBatchDetected', { count: parsedBatch.length }))
    return
  }

  const parsed = parseProxyInput(lines[0], { defaultProtocol: proxyForm.protocol })
  if (!parsed) {
    appStore.showError(t('myAccounts.proxy.smartInputInvalid'))
    return
  }

  proxyForm.protocol = parsed.protocol
  proxyForm.host = parsed.host
  proxyForm.port = parsed.port
  proxyForm.username = parsed.username
  proxyForm.password = parsed.password
  if (!proxyForm.name.trim()) {
    proxyForm.name = proxyDisplayName(parsed)
  }
  proxyInput.value = ''
  appStore.showSuccess(t('myAccounts.proxy.smartInputSuccess'))
}

async function saveProxy(): Promise<void> {
  if (proxyBatch.value.length > 0) return

  const port = Number(proxyForm.port)
  if (!proxyForm.name.trim() || !proxyForm.host.trim() || !Number.isInteger(port) || port < 1 || port > 65535) {
    appStore.showError('请填写有效的代理名称、主机和端口')
    return
  }

  savingProxy.value = true
  try {
    const payload = {
      name: proxyForm.name.trim(),
      protocol: proxyForm.protocol,
      host: proxyForm.host.trim(),
      port,
      username: proxyForm.username.trim() || null,
      status: proxyForm.status
    }
    const password = proxyForm.password.trim()
    const saved = editingProxyId.value
      ? await userAPI.updateProxy(editingProxyId.value, {
          ...payload,
          ...(password ? { password } : {})
        })
      : await userAPI.createProxy({
          ...payload,
          password: password || null
        })
    const index = userProxies.value.findIndex(proxy => proxy.id === saved.id)
    if (index >= 0) {
      const next = [...userProxies.value]
      next[index] = saved
      userProxies.value = next
    } else {
      userProxies.value = [saved, ...userProxies.value]
    }
    appStore.showSuccess(t('common.success'))
    resetProxyForm()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '代理保存失败'))
  } finally {
    savingProxy.value = false
  }
}

async function saveProxyBatch(): Promise<void> {
  if (proxyBatch.value.length === 0 || editingProxyId.value) return

  savingProxy.value = true
  const remaining: ParsedProxyInput[] = []
  let created = 0
  try {
    for (const parsed of proxyBatch.value) {
      try {
        const saved = await userAPI.createProxy({
          name: proxyDisplayName(parsed),
          protocol: parsed.protocol,
          host: parsed.host,
          port: parsed.port,
          username: parsed.username.trim() || null,
          password: parsed.password.trim() || null
        })
        userProxies.value = [saved, ...userProxies.value]
        created++
      } catch {
        remaining.push(parsed)
      }
    }

    if (remaining.length === 0) {
      appStore.showSuccess(t('myAccounts.proxy.smartInputBatchSuccess', { count: created }))
      resetProxyForm()
    } else {
      proxyBatch.value = remaining
      appStore.showError(
        t('myAccounts.proxy.smartInputBatchPartial', {
          created,
          failed: remaining.length
        })
      )
    }
  } finally {
    savingProxy.value = false
  }
}

async function removeProxy(proxy: Proxy): Promise<void> {
  if (!window.confirm(`确定删除代理「${proxy.name}」吗？`)) return
  try {
    await userAPI.deleteProxy(proxy.id)
    userProxies.value = userProxies.value.filter(item => item.id !== proxy.id)
    if (form.proxyId === proxy.id) form.proxyId = null
    if (importForm.proxyId === proxy.id) importForm.proxyId = null
    appStore.showSuccess(t('common.success'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '代理删除失败'))
  }
}

function parseJsonObject(raw: string): Record<string, unknown> {
  const parsed = JSON.parse(raw)
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error(t('myAccounts.import.invalidJson'))
  }
  return parsed as Record<string, unknown>
}

function inferTypeFromForm(): string {
  if (form.method === 'setup-token') return 'setup-token'
  return 'oauth'
}

function isUserManagedKeyBackedType(type?: string | null): boolean {
  const normalized = String(type || '').trim().toLowerCase()
  return normalized === 'apikey' || normalized === 'upstream'
}

function containsUserManagedApiKeyCredential(value: unknown): boolean {
  if (!value || typeof value !== 'object') return false
  if (Array.isArray(value)) {
    return value.some(item => containsUserManagedApiKeyCredential(item))
  }
  return Object.entries(value as Record<string, unknown>).some(([key, nested]) =>
    isUserManagedApiKeyCredentialKey(key) || containsUserManagedApiKeyCredential(nested)
  )
}

function isUserManagedApiKeyCredentialKey(key: string): boolean {
  const normalized = key.trim().toLowerCase().replace(/[\s-]+/g, '_')
  return normalized === 'api_key' || normalized === 'apikey' || normalized === 'x_api_key' || normalized === 'xapikey'
}

function normalizePositiveNumber(value: number | null): number | null {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : null
}

function normalizeNonNegativeNumber(value: number | null): number | null {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 ? value : null
}

function normalizePositiveInteger(value: number | null | undefined, fallback = 3): number {
  return typeof value === 'number' && Number.isFinite(value) && value >= 1 ? Math.trunc(value) : fallback
}

function resetAccountExtraForm(): void {
  quotaDailyLimit.value = null
  quotaWeeklyLimit.value = null
  quotaMonthlyLimit.value = null
  quotaTotalLimit.value = null
  shareDisplayEnabled.value = false
  shareDisplayName.value = ''
  shareDisplayTier.value = 'pro'
  shareDisplayPercentOnly.value = true
  shareDisplayAccountCount.value = 1
  shareDisplay5hLimit.value = null
  shareDisplay5hUsed.value = null
  shareDisplay7dLimit.value = null
  shareDisplay7dUsed.value = null
  accountPlanTier.value = 'plus'
  accountModelLimitMode.value = 'allowlist'
  accountModelAllowlist.value = ['gpt-5.2', 'gpt-5.2-chat-latest', 'gpt-5.2-pro', 'gpt-5.5', 'gpt-5.4', 'gpt-5.4-mini', 'gpt-5.3-codex']
  accountModelCustomInput.value = ''
  accountModelMappingSource.value = ''
  accountModelMappingTarget.value = ''
  accountModelMappings.value = []
  accountConcurrency.value = 3
  accountExpiresAt.value = null
  accountAutoPauseOnExpired.value = true
  codexLimitProtectionEnabled.value = true
  codex5hLimitPercent.value = 100
  codex7dLimitPercent.value = 100
  rpmLimitEnabled.value = false
  baseRpm.value = 15
  userMsgQueueMode.value = 'off'
  tlsFingerprintEnabled.value = false
  sessionIdMaskingEnabled.value = false
  cacheTtlOverrideEnabled.value = false
  cacheTtlOverrideTarget.value = '5m'
}

function selectAccountPlanTier(tier: string): void {
  accountPlanTier.value = tier
  shareDisplayTier.value = tier
}

function loadAccountExtraForm(extra: Record<string, unknown>, account?: Account): void {
  quotaDailyLimit.value = typeof extra.quota_daily_limit === 'number' && extra.quota_daily_limit > 0 ? extra.quota_daily_limit : null
  quotaWeeklyLimit.value = typeof extra.quota_weekly_limit === 'number' && extra.quota_weekly_limit > 0 ? extra.quota_weekly_limit : null
  quotaMonthlyLimit.value = typeof extra.quota_monthly_limit === 'number' && extra.quota_monthly_limit > 0 ? extra.quota_monthly_limit : null
  quotaTotalLimit.value = typeof extra.quota_limit === 'number' && extra.quota_limit > 0 ? extra.quota_limit : null
  shareDisplayEnabled.value = typeof extra.share_display_name === 'string' || typeof extra.share_display_tier === 'string' || extra.share_display_percent_only === true || typeof extra.share_display_account_count === 'number' || typeof extra.share_display_5h_limit === 'number' || typeof extra.share_display_7d_limit === 'number'
  shareDisplayName.value = typeof extra.share_display_name === 'string' ? extra.share_display_name : ''
  shareDisplayTier.value = typeof extra.share_display_tier === 'string' && extra.share_display_tier ? extra.share_display_tier : 'pro'
  shareDisplayPercentOnly.value = extra.share_display_percent_only !== false
  shareDisplayAccountCount.value = typeof extra.share_display_account_count === 'number' && extra.share_display_account_count > 0
    ? Math.trunc(extra.share_display_account_count)
    : 1
  shareDisplay5hLimit.value = typeof extra.share_display_5h_limit === 'number' && extra.share_display_5h_limit > 0 ? extra.share_display_5h_limit : null
  shareDisplay5hUsed.value = typeof extra.share_display_5h_used === 'number' && extra.share_display_5h_used >= 0 ? extra.share_display_5h_used : null
  shareDisplay7dLimit.value = typeof extra.share_display_7d_limit === 'number' && extra.share_display_7d_limit > 0 ? extra.share_display_7d_limit : null
  shareDisplay7dUsed.value = typeof extra.share_display_7d_used === 'number' && extra.share_display_7d_used >= 0 ? extra.share_display_7d_used : null
  accountPlanTier.value = typeof extra.share_display_tier === 'string' && extra.share_display_tier ? extra.share_display_tier : stringValue(account?.credentials?.plan_type) || 'plus'
  shareDisplayTier.value = accountPlanTier.value
  accountConcurrency.value = normalizePositiveInteger(account?.concurrency, 3)
  accountExpiresAt.value = typeof account?.expires_at === 'number' ? account.expires_at : null
  accountAutoPauseOnExpired.value = account?.auto_pause_on_expired !== false
  const codex5hLimit = typeof extra.share_display_5h_limit === 'number' ? extra.share_display_5h_limit : null
  const codex7dLimit = typeof extra.share_display_7d_limit === 'number' ? extra.share_display_7d_limit : null
  codexLimitProtectionEnabled.value = codex5hLimit != null || codex7dLimit != null
  codex5hLimitPercent.value = codex5hLimit ?? 100
  codex7dLimitPercent.value = codex7dLimit ?? 100
  rpmLimitEnabled.value = typeof extra.base_rpm === 'number' && extra.base_rpm > 0
  baseRpm.value = typeof extra.base_rpm === 'number' && extra.base_rpm > 0 ? extra.base_rpm : 15
  userMsgQueueMode.value = normalizeUserMsgQueueMode(extra.user_msg_queue_mode)
  tlsFingerprintEnabled.value = extra.enable_tls_fingerprint === true
  sessionIdMaskingEnabled.value = extra.session_id_masking_enabled === true
  cacheTtlOverrideEnabled.value = extra.cache_ttl_override_enabled === true
  cacheTtlOverrideTarget.value = extra.cache_ttl_override_target === '1h' ? '1h' : '5m'
  loadAccountModelMapping(account?.credentials?.model_mapping)
}

function normalizeUserMsgQueueMode(value: unknown): 'off' | 'throttle' | 'serialize' {
  if (value === 'throttle' || value === 'soft') {
    return 'throttle'
  }
  if (value === 'serialize' || value === 'serial') {
    return 'serialize'
  }
  return 'off'
}

function loadAccountModelMapping(raw: unknown): void {
  const mapping = asRecord(raw)
  if (!mapping || Object.keys(mapping).length === 0) {
    accountModelLimitMode.value = 'allowlist'
    return
  }
  const entries = Object.entries(mapping)
    .map(([source, target]) => ({ source, target: String(target ?? '') }))
    .filter(row => row.source.trim() && row.target.trim())
  if (entries.length === 0) {
    accountModelLimitMode.value = 'allowlist'
    return
  }
  const allowlist = entries.every(row => row.source === row.target)
  accountModelLimitMode.value = allowlist ? 'allowlist' : 'mapping'
  if (allowlist) {
    accountModelAllowlist.value = entries.map(row => row.source)
    return
  }
  accountModelMappings.value = entries.map(row => ({
    id: modelMappingRowId++,
    source: row.source,
    target: row.target,
  }))
}

function toggleSuggestedModel(model: string): void {
  if (accountModelAllowlist.value.includes(model)) {
    accountModelAllowlist.value = accountModelAllowlist.value.filter(item => item !== model)
    return
  }
  accountModelAllowlist.value = [...accountModelAllowlist.value, model]
}

function addCustomModel(): void {
  const model = accountModelCustomInput.value.trim()
  if (!model || accountModelAllowlist.value.includes(model)) {
    accountModelCustomInput.value = ''
    return
  }
  accountModelAllowlist.value = [...accountModelAllowlist.value, model]
  accountModelCustomInput.value = ''
}

function fillSuggestedModels(): void {
  accountModelAllowlist.value = [...openAIModelSuggestions]
}

function clearSelectedModels(): void {
  accountModelAllowlist.value = []
}

function addModelMapping(): void {
  const source = accountModelMappingSource.value.trim()
  const target = accountModelMappingTarget.value.trim()
  if (!source || !target) return
  accountModelMappings.value = [
    ...accountModelMappings.value,
    { id: modelMappingRowId++, source, target },
  ]
  accountModelMappingSource.value = ''
  accountModelMappingTarget.value = ''
}

function removeModelMapping(id: number): void {
  accountModelMappings.value = accountModelMappings.value.filter(row => row.id !== id)
}

function buildModelMapping(): Record<string, string> | undefined {
  const mapping: Record<string, string> = {}
  if (accountModelLimitMode.value === 'allowlist') {
    for (const model of accountModelAllowlist.value) {
      const normalized = model.trim()
      if (normalized) mapping[normalized] = normalized
    }
  } else {
    for (const row of accountModelMappings.value) {
      const source = row.source.trim()
      const target = row.target.trim()
      if (source && target) mapping[source] = target
    }
  }
  return Object.keys(mapping).length > 0 ? mapping : undefined
}

function buildAccountCredentialExtras(): Record<string, unknown> | undefined {
  const extras: Record<string, unknown> = {}
  if (form.platform === 'openai') {
    extras.plan_type = accountPlanTier.value
    const modelMapping = buildModelMapping()
    if (modelMapping) {
      extras.model_mapping = modelMapping
    }
  }
  return Object.keys(extras).length > 0 ? extras : undefined
}

function buildAccountCredentials(base?: Record<string, unknown>): Record<string, unknown> | undefined {
  const credentials: Record<string, unknown> = { ...(base || {}) }
  const extras = buildAccountCredentialExtras()
  if (extras) {
    Object.assign(credentials, extras)
  } else {
    delete credentials.model_mapping
  }
  return Object.keys(credentials).length > 0 ? credentials : undefined
}

function buildAccountExtra(base?: Record<string, unknown>): Record<string, unknown> | undefined {
  const extra: Record<string, unknown> = { ...(base || {}) }
  const daily = normalizePositiveNumber(quotaDailyLimit.value)
  const weekly = normalizePositiveNumber(quotaWeeklyLimit.value)
  const monthly = normalizePositiveNumber(quotaMonthlyLimit.value)
  const total = normalizePositiveNumber(quotaTotalLimit.value)

  if (daily != null) extra.quota_daily_limit = daily
  else {
    delete extra.quota_daily_limit
    delete extra.quota_daily_used
    delete extra.quota_daily_start
  }
  if (weekly != null) extra.quota_weekly_limit = weekly
  else {
    delete extra.quota_weekly_limit
    delete extra.quota_weekly_used
    delete extra.quota_weekly_start
  }
  if (monthly != null) extra.quota_monthly_limit = monthly
  else {
    delete extra.quota_monthly_limit
    delete extra.quota_monthly_used
    delete extra.quota_monthly_start
  }
  if (total != null) extra.quota_limit = total
  else delete extra.quota_limit

  if (form.platform === 'openai') {
    extra.share_display_tier = accountPlanTier.value || 'plus'
    extra.share_display_percent_only = true
    if (codexLimitProtectionEnabled.value) {
      const fiveHour = normalizeNonNegativeNumber(codex5hLimitPercent.value)
      const sevenDay = normalizeNonNegativeNumber(codex7dLimitPercent.value)
      extra.share_display_5h_limit = fiveHour ?? 100
      extra.share_display_7d_limit = sevenDay ?? 100
    } else {
      delete extra.share_display_5h_limit
      delete extra.share_display_5h_used
      delete extra.share_display_5h_start
      delete extra.share_display_7d_limit
      delete extra.share_display_7d_used
      delete extra.share_display_7d_start
    }
  } else {
    delete extra.share_display_tier
    delete extra.share_display_percent_only
    delete extra.share_display_5h_limit
    delete extra.share_display_5h_used
    delete extra.share_display_5h_start
    delete extra.share_display_7d_limit
    delete extra.share_display_7d_used
    delete extra.share_display_7d_start
  }

  if (rpmLimitEnabled.value) {
    extra.base_rpm = normalizePositiveNumber(baseRpm.value) ?? 15
    extra.rpm_strategy = 'tiered'
  } else {
    delete extra.base_rpm
    delete extra.rpm_strategy
  }
  if (userMsgQueueMode.value !== 'off') {
    extra.user_msg_queue_mode = userMsgQueueMode.value
  } else {
    delete extra.user_msg_queue_mode
  }
  if (tlsFingerprintEnabled.value) extra.enable_tls_fingerprint = true
  else delete extra.enable_tls_fingerprint
  if (sessionIdMaskingEnabled.value) extra.session_id_masking_enabled = true
  else delete extra.session_id_masking_enabled
  if (cacheTtlOverrideEnabled.value) {
    extra.cache_ttl_override_enabled = true
    extra.cache_ttl_override_target = cacheTtlOverrideTarget.value
  } else {
    delete extra.cache_ttl_override_enabled
    delete extra.cache_ttl_override_target
  }

  if (form.platform === 'openai' && shareDisplayEnabled.value) {
    const displayName = shareDisplayName.value.trim()
    if (displayName) extra.share_display_name = displayName
    else delete extra.share_display_name
    extra.share_display_tier = shareDisplayTier.value || 'pro'
    extra.share_display_percent_only = shareDisplayPercentOnly.value
    extra.share_display_account_count = Math.max(1, Math.trunc(shareDisplayAccountCount.value || 1))
    writeOptionalShareDisplayNumber(extra, 'share_display_5h_limit', shareDisplay5hLimit.value, true)
    writeOptionalShareDisplayNumber(extra, 'share_display_5h_used', shareDisplay5hUsed.value, false)
    writeOptionalShareDisplayNumber(extra, 'share_display_7d_limit', shareDisplay7dLimit.value, true)
    writeOptionalShareDisplayNumber(extra, 'share_display_7d_used', shareDisplay7dUsed.value, false)
  } else {
    delete extra.share_display_name
    delete extra.share_display_account_count
    if (form.platform !== 'openai') {
      delete extra.share_display_tier
      delete extra.share_display_percent_only
      delete extra.share_display_5h_limit
      delete extra.share_display_5h_used
      delete extra.share_display_5h_start
      delete extra.share_display_7d_limit
      delete extra.share_display_7d_used
      delete extra.share_display_7d_start
    }
  }
  return Object.keys(extra).length > 0 ? extra : undefined
}

function writeOptionalShareDisplayNumber(extra: Record<string, unknown>, key: string, value: number | null, requirePositive: boolean): void {
  if (typeof value === 'number' && Number.isFinite(value) && (requirePositive ? value > 0 : value >= 0)) {
    extra[key] = value
  } else {
    delete extra[key]
  }
}

function selectedProxyIdPayload(value: number | null): number | null {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : null
}

function withSelectedProxy<T extends Record<string, unknown>>(payload: T, value: number | null): T {
  const proxyId = selectedProxyIdPayload(value)
  if (proxyId != null) {
    return { ...payload, proxy_id: proxyId }
  }
  return payload
}

function selectedProxyIdUpdatePayload(value: number | null): number {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : 0
}

async function saveAccount(): Promise<void> {
  savingAccount.value = true
  try {
    const accountExtra = buildAccountExtra()
    const credentialExtras = buildAccountCredentialExtras()
    const accountConfigPayload = {
      extra: accountExtra,
      credential_extras: credentialExtras,
      concurrency: normalizePositiveInteger(accountConcurrency.value, 3),
      expires_at: accountExpiresAt.value,
      auto_pause_on_expired: accountAutoPauseOnExpired.value,
    }
    if (editingAccount.value) {
      const baseExtra = (editingAccount.value.extra ?? {}) as Record<string, unknown>
      const parsedCredentials = credentialsJson.value.trim()
        ? parseJsonObject(credentialsJson.value)
        : (editingAccount.value.credentials ?? {})
      const payload = {
        name: form.name.trim(),
        notes: form.notes.trim() || null,
        credentials: isUserManagedKeyBackedType(editingAccount.value.type) ? undefined : buildAccountCredentials(parsedCredentials),
        extra: buildAccountExtra(baseExtra) ?? {},
        proxy_id: selectedProxyIdUpdatePayload(form.proxyId),
        concurrency: normalizePositiveInteger(accountConcurrency.value, 3),
        expires_at: accountExpiresAt.value ?? 0,
        auto_pause_on_expired: accountAutoPauseOnExpired.value
      }
      const updated = await userAPI.updateAccount(editingAccount.value.id, payload)
      patchAccount(updated)
      appStore.showSuccess(t('common.success'))
      closeAccountModal()
      return
    }

    let createdAccount: Account | null = null
    if (form.method === 'oauth') {
      const callback = parseOAuthCallbackInput(oauthCode.value)
      if (!authSessionId.value || !callback.code) {
        appStore.showError(t('myAccounts.oauthMissing'))
        return
      }
      createdAccount = await userAPI.exchangeAccountOAuthCode(withSelectedProxy({
        platform: form.platform,
        method: form.method,
        session_id: authSessionId.value,
        code: callback.code,
        state: callback.state || authState.value || undefined,
        name: form.name.trim(),
        notes: form.notes.trim() || null,
        ...accountConfigPayload,
      }, form.proxyId))
    } else if (form.method === 'session-key' || form.method === 'setup-token') {
      createdAccount = await userAPI.importAccountSession(withSelectedProxy({
        platform: form.platform,
        method: form.method,
        session_key: sessionKey.value.trim(),
        name: form.name.trim(),
        notes: form.notes.trim() || null,
        ...accountConfigPayload,
      }, form.proxyId))
    } else if (form.method === 'apikey') {
      if (!apiKeyValue.value.trim()) {
        appStore.showError(t('admin.accounts.pleaseEnterApiKey'))
        return
      }
      createdAccount = await userAPI.createAccount(withSelectedProxy({
        name: form.name.trim() || defaultAccountName(),
        notes: form.notes.trim() || null,
        platform: form.platform,
        type: 'apikey',
        credentials: {
          base_url: apiKeyBaseUrl.value.trim() || 'https://api.openai.com',
          api_key: apiKeyValue.value.trim(),
          ...(buildAccountCredentialExtras() ?? {})
        },
        extra: accountExtra,
        concurrency: normalizePositiveInteger(accountConcurrency.value, 3),
        expires_at: accountExpiresAt.value,
        auto_pause_on_expired: accountAutoPauseOnExpired.value,
      }, form.proxyId))
    } else {
      const credentials = buildAccountCredentials(parseJsonObject(credentialsJson.value))
      createdAccount = await userAPI.createAccount(withSelectedProxy({
        name: form.name.trim() || defaultAccountName(),
        notes: form.notes.trim() || null,
        platform: form.platform,
        type: inferTypeFromForm(),
        credentials: credentials ?? {},
        extra: accountExtra,
        concurrency: normalizePositiveInteger(accountConcurrency.value, 3),
        expires_at: accountExpiresAt.value,
        auto_pause_on_expired: accountAutoPauseOnExpired.value,
      }, form.proxyId))
    }
    if (createdAccount && createShareMode.value === 'public') {
      createdAccount = await userAPI.updateAccountShareMode(createdAccount.id, 'public')
    }
    if (createdAccount) {
      accounts.value = [createdAccount, ...accounts.value]
    }
    appStore.showSuccess(t('common.success'))
    closeAccountModal()
    await loadAll()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('myAccounts.saveFailed')))
  } finally {
    savingAccount.value = false
  }
}

function defaultAccountName(): string {
  const platform = platformOptions.value.find(item => item.value === form.platform)?.label ?? form.platform
  return `${platform} ${t('myAccounts.account')}`
}

async function generateAuthUrl(): Promise<void> {
  authUrlLoading.value = true
  try {
    const payload: UserAccountAuthURLRequest = {
      platform: form.platform,
      method: form.method
    }
    if (form.platform !== 'openai' && typeof window !== 'undefined') {
      payload.redirect_uri = `${window.location.origin}/auth/callback`
    }
    const proxyID = selectedProxyIdPayload(form.proxyId)
    if (proxyID != null) {
      payload.proxy_id = proxyID
    }
    const result = await userAPI.generateAccountAuthURL(payload)
    authUrl.value = String(result.auth_url || result.url || '')
    authSessionId.value = String(result.session_id || '')
    authState.value = String(result.state || '')
    if (!authUrl.value) {
      appStore.showError(t('myAccounts.authUrlMissing'))
    }
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('myAccounts.authUrlFailed')))
  } finally {
    authUrlLoading.value = false
  }
}

function buildImportPayload(format: string, platform: string, content: string): { type: string; name: string; credentials: Record<string, unknown>; extra: Record<string, unknown> | null } {
  const trimmed = content.trim()
  if (!trimmed) throw new Error(t('myAccounts.import.emptyContent'))
  if (format === 'openai_refresh_token') {
    return { type: 'oauth', name: '', credentials: { refresh_token: trimmed }, extra: null }
  }
  if (format === 'claude_session_key') {
    return { type: 'oauth', name: '', credentials: { session_key: trimmed }, extra: null }
  }
  const parsed = parseJsonObject(trimmed)
  const credentials = (parsed.credentials && typeof parsed.credentials === 'object')
    ? parsed.credentials as Record<string, unknown>
    : parsed
  const type = typeof parsed.type === 'string' ? parsed.type : (platform === 'anthropic' && format.includes('setup') ? 'setup-token' : 'oauth')
  if (isUserManagedKeyBackedType(type) || containsUserManagedApiKeyCredential(credentials)) {
    throw new Error(t('myAccounts.apiKeyUploadDisabled'))
  }
  const extra = asRecord(parsed.extra)
  return {
    type,
    name: stringValue(parsed.name),
    credentials,
    extra: extra ? { ...extra } : null,
  }
}

function openImportFilePicker(): void {
  importFileInput.value?.click()
}

function openImportFolderPicker(): void {
  importFolderInput.value?.click()
}

async function handleImportFileChange(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  importFileReading.value = true
  try {
    const content = await file.text()
    importContent.value = content
    importFileName.value = file.name
    importFolderFiles.value = []
    inferImportFileSelection(file.name, content)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('myAccounts.import.fileReadFailed')))
  } finally {
    importFileReading.value = false
    input.value = ''
  }
}

async function handleImportFolderChange(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  if (files.length === 0) return
  importFileReading.value = true
  try {
    const entries: ImportFileEntry[] = []
    for (const file of files) {
      if (!isSupportedImportFileName(file.name)) {
        continue
      }
      const content = await file.text()
      const selection = inferImportFileSelectionValues(file.name, content)
      entries.push({
        name: file.webkitRelativePath || file.name,
        content,
        platform: selection.platform,
        format: selection.format,
      })
    }
    if (entries.length === 0) {
      appStore.showError(t('myAccounts.import.folderNoSupportedFiles'))
      return
    }
    importFolderFiles.value = entries
    importFileName.value = ''
    importContent.value = entries.map(entry => `// ${entry.name}\n${entry.content}`).join('\n\n')
    importForm.platform = entries[0].platform
    importForm.format = entries[0].format
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('myAccounts.import.folderReadFailed')))
  } finally {
    importFileReading.value = false
    input.value = ''
  }
}

function inferImportFileSelection(fileName: string, content: string): void {
  const selection = inferImportFileSelectionValues(fileName, content)
  importForm.platform = selection.platform
  importForm.format = selection.format
}

function inferImportFileSelectionValues(fileName: string, content: string): { platform: AccountPlatform, format: string } {
  const lowerName = fileName.toLowerCase()
  const trimmed = content.trim()
  const parsed = tryParseImportObject(trimmed)
  if (parsed) {
    const credentials = asRecord(parsed.credentials) ?? parsed
    const platform = normalizeImportPlatform(
      stringValue(parsed.platform) ||
      stringValue(parsed.provider) ||
      stringValue(parsed.account_platform) ||
      inferPlatformFromCredentials(credentials)
    )
    return {
      platform: platform || 'openai',
      format: isLikelyCodexManagerToken(lowerName, parsed, credentials)
        ? 'codex_manager_chatgpt_token_json'
        : 'sub2api_oauth_json',
    }
  }

  if (isLikelyClaudeSession(lowerName, trimmed)) {
    return { platform: 'anthropic', format: 'claude_session_key' }
  }
  return { platform: 'openai', format: 'openai_refresh_token' }
}

function isSupportedImportFileName(fileName: string): boolean {
  return /\.(json|txt|token|key)$/i.test(fileName)
}

function tryParseImportObject(raw: string): Record<string, unknown> | null {
  if (!raw) return null
  try {
    return parseJsonObject(raw)
  } catch {
    return null
  }
}

function asRecord(value: unknown): Record<string, unknown> | null {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? value as Record<string, unknown>
    : null
}

function stringValue(value: unknown): string {
  return typeof value === 'string' ? value.trim() : ''
}

function normalizeImportPlatform(value: string): AccountPlatform | '' {
  const normalized = value.toLowerCase()
  if (normalized === 'claude' || normalized === 'anthropic') return 'anthropic'
  if (normalized === 'chatgpt' || normalized === 'openai') return 'openai'
  if (normalized === 'gemini' || normalized === 'antigravity') return normalized as AccountPlatform
  return ''
}

function inferPlatformFromCredentials(credentials: Record<string, unknown>): string {
  if (stringValue(credentials.session_key) || stringValue(credentials.sessionKey)) return 'anthropic'
  if (stringValue(credentials.refresh_token) || stringValue(credentials.refreshToken) || stringValue(credentials.access_token) || stringValue(credentials.accessToken)) {
    return 'openai'
  }
  return ''
}

function isLikelyCodexManagerToken(fileName: string, root: Record<string, unknown>, credentials: Record<string, unknown>): boolean {
  if (fileName.includes('codex') || fileName.includes('chatgpt')) return true
  return Boolean(
    stringValue(root.sessionToken) ||
    stringValue(root.accessToken) ||
    stringValue(root.refreshToken) ||
    stringValue(credentials.sessionToken) ||
    stringValue(credentials.accessToken) ||
    stringValue(credentials.refreshToken)
  )
}

function isLikelyClaudeSession(fileName: string, content: string): boolean {
  const lower = content.toLowerCase()
  return fileName.includes('claude') ||
    fileName.includes('anthropic') ||
    fileName.includes('session') ||
    lower.startsWith('sk-ant') ||
    lower.includes('sessionkey')
}

async function importFromContent(): Promise<void> {
  importing.value = true
  try {
    if (importFolderFiles.value.length > 0) {
      const createdAccounts: Account[] = []
      const failedNames: string[] = []
      for (const entry of importFolderFiles.value) {
        try {
          const built = buildImportPayload(entry.format, entry.platform, entry.content)
          const created = await userAPI.importAccount(withSelectedProxy({
            format: entry.format,
            name: built.name,
            platform: entry.platform,
            type: built.type,
            credentials: built.credentials,
            extra: built.extra ?? undefined,
          }, importForm.proxyId))
          createdAccounts.push(created)
        } catch {
          failedNames.push(entry.name)
        }
      }
      if (createdAccounts.length > 0) {
        accounts.value = [...createdAccounts, ...accounts.value]
      }
      if (failedNames.length > 0) {
        appStore.showError(t('myAccounts.import.folderImportPartialFailed', {
          success: createdAccounts.length,
          failed: failedNames.length,
        }))
        if (createdAccounts.length === 0) {
          return
        }
      } else {
        appStore.showSuccess(t('myAccounts.import.folderImportSuccess', { count: createdAccounts.length }))
      }
      showImportModal.value = false
      await loadAll()
      return
    }

    const built = buildImportPayload(importForm.format, importForm.platform, importContent.value)
    const created = await userAPI.importAccount(withSelectedProxy({
      format: importForm.format,
      name: built.name,
      platform: importForm.platform as AccountPlatform,
      type: built.type,
      credentials: built.credentials,
      extra: built.extra ?? undefined,
    }, importForm.proxyId))
    accounts.value = [created, ...accounts.value]
    appStore.showSuccess(t('common.success'))
    showImportModal.value = false
    await loadAll()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('myAccounts.import.failed')))
  } finally {
    importing.value = false
  }
}

function patchAccount(updated: Account): void {
  const index = accounts.value.findIndex(account => account.id === updated.id)
  if (index >= 0) {
    const next = [...accounts.value]
    next[index] = { ...next[index], ...updated }
    accounts.value = next
  }
}

async function bulkUpdateShareMode(
  targets: Account[],
  shareMode: AccountShareMode,
  messages: {
    empty: string
    partial: string
    success: string
  }
): Promise<void> {
  if (targets.length === 0) {
    appStore.showError(t(messages.empty))
    return
  }

  bulkSharing.value = true
  const succeeded: Account[] = []
  const failedIds: number[] = []
  try {
    for (const account of targets) {
      try {
        const updated = await userAPI.updateAccountShareMode(account.id, shareMode)
        succeeded.push(updated)
        patchAccount(updated)
      } catch {
        failedIds.push(account.id)
      }
    }

    if (succeeded.length > 0) {
      removeSelectedAccounts(succeeded.map(account => account.id))
    }

    if (failedIds.length > 0) {
      setSelectedIds(failedIds)
      appStore.showError(t(messages.partial, {
        success: succeeded.length,
        failed: failedIds.length
      }))
      return
    }

    clearSelection()
    appStore.showSuccess(t(messages.success, { count: succeeded.length }))
    await loadShareSummary()
  } finally {
    bulkSharing.value = false
  }
}

async function bulkApplyPublic(): Promise<void> {
  await bulkUpdateShareMode([...selectedShareableAccounts.value], 'public', {
    empty: 'myAccounts.bulk.noShareableSelection',
    partial: 'myAccounts.bulk.applyPublicPartial',
    success: 'myAccounts.bulk.applyPublicSuccess'
  })
}

async function bulkMakePrivate(): Promise<void> {
  await bulkUpdateShareMode([...selectedPublicAccounts.value], 'private', {
    empty: 'myAccounts.bulk.noPublicSelection',
    partial: 'myAccounts.bulk.makePrivatePartial',
    success: 'myAccounts.bulk.makePrivateSuccess'
  })
}

async function toggleShareMode(account: Account): Promise<void> {
  const nextMode: AccountShareMode = account.share_mode === 'public' ? 'private' : 'public'
  shareUpdatingId.value = account.id
  try {
    const updated = await userAPI.updateAccountShareMode(account.id, nextMode)
    patchAccount(updated)
    appStore.showSuccess(nextMode === 'public' ? t('myAccounts.publicRequested') : t('myAccounts.privateSet'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('myAccounts.shareFailed')))
  } finally {
    shareUpdatingId.value = null
  }
}

async function runTest(account: Account): Promise<void> {
  try {
    await userAPI.testAccount(account.id)
    appStore.showSuccess(t('myAccounts.testStarted'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('myAccounts.testFailed')))
  }
}

async function deleteOwnedAccount(account: Account): Promise<void> {
  if (!window.confirm(t('myAccounts.deleteConfirm', { name: account.name }))) return
  try {
    await userAPI.deleteAccount(account.id)
    accounts.value = accounts.value.filter(item => item.id !== account.id)
    await loadShareSummary()
    appStore.showSuccess(t('myAccounts.deleted'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('myAccounts.deleteFailed')))
  }
}

async function transferShare(): Promise<void> {
  transferring.value = true
  try {
    const result = await userAPI.transferAccountShareToBalance()
    appStore.showSuccess(t('myAccounts.transferSuccess', { amount: formatCreditAmount(result.transferred_amount) }))
    await Promise.all([loadShareSummary(), authStore.refreshUser().catch(() => undefined)])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('myAccounts.transferFailed')))
  } finally {
    transferring.value = false
  }
}

function formatShareMode(mode: string | null | undefined): string {
  return mode === 'public' ? t('myAccounts.shareMode.public') : t('myAccounts.shareMode.private')
}

function formatShareStatus(status: string | null | undefined): string {
  switch (status as AccountShareStatus) {
    case 'pending_review':
      return t('myAccounts.shareStatus.pendingReview')
    case 'active':
      return t('myAccounts.shareStatus.active')
    case 'rejected':
      return t('myAccounts.shareStatus.rejected')
    case 'suspended':
      return t('myAccounts.shareStatus.suspended')
    default:
      return t('myAccounts.shareStatus.notShared')
  }
}

function shareStatusClass(status: string | null | undefined): string {
  switch (status as AccountShareStatus) {
    case 'pending_review':
      return 'badge-warning'
    case 'active':
      return 'badge-success'
    case 'rejected':
    case 'suspended':
      return 'badge-danger'
    default:
      return 'badge-secondary'
  }
}

function formatExpiresAt(value: number | null): string {
  return value ? formatDateTime(new Date(value * 1000)) : '-'
}

onMounted(() => {
  loadAll()
})
</script>
