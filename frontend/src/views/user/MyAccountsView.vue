<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('myAccounts.summary.available') }}</p>
          <p class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">
            {{ formatCurrency(shareSummary?.available_amount ?? 0) }}
          </p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('myAccounts.summary.frozen') }}</p>
          <p class="mt-2 text-2xl font-semibold text-amber-600 dark:text-amber-400">
            {{ formatCurrency(shareSummary?.frozen_amount ?? 0) }}
          </p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('myAccounts.summary.transferred') }}</p>
          <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
            {{ formatCurrency(shareSummary?.transferred_amount ?? 0) }}
          </p>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('myAccounts.summary.total') }}</p>
          <div class="mt-2 flex items-center justify-between gap-3">
            <p class="text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatCurrency(shareSummary?.total_amount ?? 0) }}
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
                <button class="rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 hover:text-emerald-600 dark:hover:bg-dark-700 dark:hover:text-emerald-400" @click="runTest(row)">
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
      <div class="max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-800">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ editingAccount ? t('myAccounts.editAccount') : t('myAccounts.addAccount') }}
            </h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('myAccounts.wizardHint') }}</p>
          </div>
          <button class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 dark:hover:bg-dark-700" @click="closeAccountModal">
            <Icon name="x" size="md" />
          </button>
        </div>

        <div v-if="!editingAccount" class="mt-6 grid gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('myAccounts.platform') }}</label>
            <Select v-model="form.platform" :options="platformOptions" />
          </div>
          <div>
            <label class="input-label">{{ t('myAccounts.method') }}</label>
            <Select v-model="form.method" :options="methodOptions" />
          </div>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2">
          <Input v-model="form.name" :label="t('myAccounts.name')" :placeholder="t('myAccounts.namePlaceholder')" data-testid="my-accounts-name" />
          <Input v-model="form.notes" :label="t('myAccounts.notes')" :placeholder="t('myAccounts.notesPlaceholder')" />
        </div>

        <div v-if="!editingAccount && form.method === 'oauth'" class="mt-5 rounded-xl border border-gray-200 p-4 dark:border-dark-700">
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

        <div v-else-if="!editingAccount && (form.method === 'session-key' || form.method === 'setup-token')" class="mt-5">
          <label class="input-label">{{ t('myAccounts.sessionKey') }}</label>
          <textarea v-model="sessionKey" class="input min-h-[120px] w-full" :placeholder="t('myAccounts.sessionKeyPlaceholder')"></textarea>
        </div>

        <div v-else-if="!editingAccount && form.method === 'apikey'" class="mt-5 grid gap-4 md:grid-cols-2">
          <Input v-model="apiKeyBaseUrl" :label="t('admin.accounts.baseUrl')" placeholder="https://api.openai.com" data-testid="my-accounts-apikey-base-url" />
          <Input v-model="apiKeyValue" :label="t('admin.accounts.apiKeyRequired')" type="password" placeholder="sk-proj-..." data-testid="my-accounts-apikey-value" />
        </div>

        <div v-else class="mt-5">
          <label class="input-label">{{ t('myAccounts.import.credentials') }}</label>
          <textarea v-model="credentialsJson" class="input min-h-[160px] w-full font-mono text-xs" :placeholder="t('myAccounts.import.credentialsPlaceholder')"></textarea>
          <p v-if="!editingAccount && form.method === 'json'" class="mt-2 text-xs text-gray-500 dark:text-dark-400">
            {{ t('myAccounts.tokenJsonHint') }}
          </p>
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
            @update:enabled="shareDisplayEnabled = $event"
            @update:displayName="shareDisplayName = $event"
            @update:displayTier="shareDisplayTier = $event"
            @update:percentOnly="shareDisplayPercentOnly = $event"
          />
        </div>

        <div class="mt-6 flex justify-end gap-3">
          <button class="btn btn-secondary" @click="closeAccountModal">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" data-testid="my-accounts-save" :disabled="savingAccount" @click="saveAccount">
            <Icon v-if="savingAccount" name="refresh" size="sm" class="animate-spin" />
            <span>{{ t('common.save') }}</span>
          </button>
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
import { formatCurrency, formatDateTime, formatNumber, formatRelativeTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'
import { parseOAuthCallbackInput } from '@/utils/oauthCallback'
import { useTableSelection } from '@/composables/useTableSelection'
import type { Account, AccountPlatform, AccountShareMode, AccountShareStatus, UserAccountShareSummary } from '@/types'
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
const shareUpdatingId = ref<number | null>(null)
const shareSummary = ref<UserAccountShareSummary | null>(null)
const showAccountModal = ref(false)
const showImportModal = ref(false)
const editingAccount = ref<Account | null>(null)
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
  method: 'oauth'
})
const importForm = reactive({
  platform: 'openai',
  format: 'sub2api_oauth_json'
})
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
    ...(form.platform === 'openai' ? [{ value: 'apikey', label: 'OpenAI API Key' }] : []),
    { value: 'json', label: t('myAccounts.import.jsonToken') }
  ]
})

watch(
  () => form.platform,
  (platform) => {
    if (platform !== 'openai' && form.method === 'apikey') {
      form.method = 'oauth'
    }
    if (!editingAccount.value) {
      apiKeyBaseUrl.value = platform === 'openai' ? 'https://api.openai.com' : ''
    }
  }
)

const isAccountQuotaConfigVisible = computed(() => {
  if (editingAccount.value) {
    return editingAccount.value.type === 'apikey' || editingAccount.value.type === 'bedrock'
  }
  return form.method === 'apikey'
})

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

async function loadAll(): Promise<void> {
  await Promise.all([loadAccounts(), loadShareSummary()])
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
  form.name = account.name
  form.notes = account.notes ?? ''
  form.platform = account.platform
  form.method = 'json'
  credentialsJson.value = JSON.stringify(account.credentials ?? {}, null, 2)
  const credentials = account.credentials ?? {}
  apiKeyBaseUrl.value = String(credentials.base_url || (account.platform === 'openai' ? 'https://api.openai.com' : ''))
  apiKeyValue.value = ''
  loadAccountExtraForm(account.extra ?? {})
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
  showImportModal.value = true
}

function closeAccountModal(): void {
  showAccountModal.value = false
  editingAccount.value = null
}

function parseJsonObject(raw: string): Record<string, unknown> {
  const parsed = JSON.parse(raw)
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error(t('myAccounts.import.invalidJson'))
  }
  return parsed as Record<string, unknown>
}

function inferTypeFromForm(): string {
  if (form.method === 'apikey') return 'apikey'
  if (form.method === 'setup-token') return 'setup-token'
  return 'oauth'
}

function normalizePositiveNumber(value: number | null): number | null {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : null
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
}

function loadAccountExtraForm(extra: Record<string, unknown>): void {
  quotaDailyLimit.value = typeof extra.quota_daily_limit === 'number' && extra.quota_daily_limit > 0 ? extra.quota_daily_limit : null
  quotaWeeklyLimit.value = typeof extra.quota_weekly_limit === 'number' && extra.quota_weekly_limit > 0 ? extra.quota_weekly_limit : null
  quotaMonthlyLimit.value = typeof extra.quota_monthly_limit === 'number' && extra.quota_monthly_limit > 0 ? extra.quota_monthly_limit : null
  quotaTotalLimit.value = typeof extra.quota_limit === 'number' && extra.quota_limit > 0 ? extra.quota_limit : null
  shareDisplayEnabled.value = typeof extra.share_display_name === 'string' || typeof extra.share_display_tier === 'string' || extra.share_display_percent_only === true
  shareDisplayName.value = typeof extra.share_display_name === 'string' ? extra.share_display_name : ''
  shareDisplayTier.value = typeof extra.share_display_tier === 'string' && extra.share_display_tier ? extra.share_display_tier : 'pro'
  shareDisplayPercentOnly.value = extra.share_display_percent_only !== false
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

  if (form.platform === 'openai' && shareDisplayEnabled.value) {
    const displayName = shareDisplayName.value.trim()
    if (displayName) extra.share_display_name = displayName
    else delete extra.share_display_name
    extra.share_display_tier = shareDisplayTier.value || 'pro'
    extra.share_display_percent_only = shareDisplayPercentOnly.value
  } else {
    delete extra.share_display_name
    delete extra.share_display_tier
    delete extra.share_display_percent_only
  }
  return Object.keys(extra).length > 0 ? extra : undefined
}

async function saveAccount(): Promise<void> {
  savingAccount.value = true
  try {
    if (editingAccount.value) {
      const baseExtra = (editingAccount.value.extra ?? {}) as Record<string, unknown>
      const payload = {
        name: form.name.trim(),
        notes: form.notes.trim() || null,
        credentials: credentialsJson.value.trim() ? parseJsonObject(credentialsJson.value) : undefined,
        extra: isAccountQuotaConfigVisible.value ? buildAccountExtra(baseExtra) ?? {} : undefined
      }
      const updated = await userAPI.updateAccount(editingAccount.value.id, payload)
      patchAccount(updated)
      appStore.showSuccess(t('common.success'))
      closeAccountModal()
      return
    }

    if (form.method === 'oauth') {
      const callback = parseOAuthCallbackInput(oauthCode.value)
      if (!authSessionId.value || !callback.code) {
        appStore.showError(t('myAccounts.oauthMissing'))
        return
      }
      const created = await userAPI.exchangeAccountOAuthCode({
        platform: form.platform,
        method: form.method,
        session_id: authSessionId.value,
        code: callback.code,
        state: callback.state || authState.value || undefined,
        name: form.name.trim(),
        notes: form.notes.trim() || null
      })
      accounts.value = [created, ...accounts.value]
    } else if (form.method === 'session-key' || form.method === 'setup-token') {
      const created = await userAPI.importAccountSession({
        platform: form.platform,
        method: form.method,
        session_key: sessionKey.value.trim(),
        name: form.name.trim(),
        notes: form.notes.trim() || null
      })
      accounts.value = [created, ...accounts.value]
    } else if (form.method === 'apikey') {
      if (!apiKeyValue.value.trim()) {
        appStore.showError(t('admin.accounts.pleaseEnterApiKey'))
        return
      }
      const created = await userAPI.createAccount({
        name: form.name.trim() || defaultAccountName(),
        notes: form.notes.trim() || null,
        platform: form.platform,
        type: 'apikey',
        credentials: {
          base_url: apiKeyBaseUrl.value.trim() || 'https://api.openai.com',
          api_key: apiKeyValue.value.trim()
        },
        extra: buildAccountExtra()
      })
      accounts.value = [created, ...accounts.value]
    } else {
      const created = await userAPI.createAccount({
        name: form.name.trim() || defaultAccountName(),
        notes: form.notes.trim() || null,
        platform: form.platform,
        type: inferTypeFromForm(),
        credentials: parseJsonObject(credentialsJson.value),
        extra: isAccountQuotaConfigVisible.value ? buildAccountExtra() : undefined
      })
      accounts.value = [created, ...accounts.value]
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
    const result = await userAPI.generateAccountAuthURL({
      platform: form.platform,
      method: form.method,
      redirect_uri: typeof window === 'undefined' ? undefined : `${window.location.origin}/auth/callback`
    })
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

function buildImportCredentials(format: string, platform: string, content: string): { type: string; credentials: Record<string, unknown> } {
  const trimmed = content.trim()
  if (!trimmed) throw new Error(t('myAccounts.import.emptyContent'))
  if (format === 'openai_refresh_token') {
    return { type: 'oauth', credentials: { refresh_token: trimmed } }
  }
  if (format === 'claude_session_key') {
    return { type: 'oauth', credentials: { session_key: trimmed } }
  }
  const parsed = parseJsonObject(trimmed)
  const credentials = (parsed.credentials && typeof parsed.credentials === 'object')
    ? parsed.credentials as Record<string, unknown>
    : parsed
  const type = typeof parsed.type === 'string' ? parsed.type : (platform === 'anthropic' && format.includes('setup') ? 'setup-token' : 'oauth')
  return { type, credentials }
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
          const built = buildImportCredentials(entry.format, entry.platform, entry.content)
          const created = await userAPI.importAccount({
            format: entry.format,
            name: '',
            platform: entry.platform,
            type: built.type,
            credentials: built.credentials
          })
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

    const built = buildImportCredentials(importForm.format, importForm.platform, importContent.value)
    const created = await userAPI.importAccount({
      format: importForm.format,
      name: '',
      platform: importForm.platform as AccountPlatform,
      type: built.type,
      credentials: built.credentials
    })
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
    appStore.showSuccess(t('myAccounts.transferSuccess', { amount: formatCurrency(result.transferred_amount) }))
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
