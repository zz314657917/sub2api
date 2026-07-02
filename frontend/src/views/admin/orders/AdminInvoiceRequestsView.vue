<template>
  <AppLayout>
    <div class="space-y-4">
      <div class="card p-4">
        <div class="flex flex-wrap items-center gap-3">
          <Select v-model="filters.status" :options="statusOptions" class="w-40" @change="loadInvoices" />
          <input v-model.number="filters.user_id" type="number" min="1" :placeholder="t('payment.admin.userIdFilter')" class="input w-40" @keyup.enter="loadInvoices" />
          <div class="flex flex-1 items-center justify-end gap-2">
            <button class="btn btn-secondary" :disabled="loading" :title="t('common.refresh')" @click="loadInvoices">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </div>

      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 text-sm dark:divide-dark-600">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr class="text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                <th class="px-4 py-3">{{ t('payment.invoices.requestId') }}</th>
                <th class="px-4 py-3">{{ t('payment.admin.colUser') }}</th>
                <th class="px-4 py-3">{{ t('payment.invoices.amount') }}</th>
                <th class="px-4 py-3">{{ t('payment.invoices.type') }}</th>
                <th class="px-4 py-3">{{ t('payment.invoices.title') }}</th>
                <th class="px-4 py-3">{{ t('payment.invoices.status') }}</th>
                <th class="px-4 py-3">{{ t('payment.orders.createdAt') }}</th>
                <th class="px-4 py-3 text-right">{{ t('payment.orders.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-if="loading">
                <td colspan="8" class="px-4 py-8 text-center text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</td>
              </tr>
              <tr v-else-if="items.length === 0">
                <td colspan="8" class="px-4 py-8 text-center text-gray-500 dark:text-gray-400">{{ t('payment.invoices.empty') }}</td>
              </tr>
              <tr v-for="item in items" v-else :key="item.id" class="text-gray-700 dark:text-gray-200">
                <td class="whitespace-nowrap px-4 py-3 font-mono">#{{ item.id }}</td>
                <td class="px-4 py-3">
                  <div class="min-w-0">
                    <div class="truncate font-medium text-gray-900 dark:text-white">{{ item.user_email || `#${item.user_id}` }}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">#{{ item.user_id }} {{ item.user_name || '' }}</div>
                  </div>
                </td>
                <td class="whitespace-nowrap px-4 py-3 font-medium">{{ formatInvoiceAmount(item.amount, item.currency) }}</td>
                <td class="whitespace-nowrap px-4 py-3">{{ invoiceTypeLabel(item.invoice_type) }}</td>
                <td class="max-w-[220px] truncate px-4 py-3">{{ item.title }}</td>
                <td class="whitespace-nowrap px-4 py-3"><span :class="invoiceStatusClass(item.status)">{{ invoiceStatusLabel(item.status) }}</span></td>
                <td class="whitespace-nowrap px-4 py-3 text-gray-500 dark:text-gray-400">{{ formatDateTime(item.created_at) }}</td>
                <td class="whitespace-nowrap px-4 py-3 text-right">
                  <div class="flex justify-end gap-1">
                    <button class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-dark-600" @click="openDetail(item)">
                      <Icon name="eye" size="sm" />
                      {{ t('common.view') }}
                    </button>
                    <button v-if="item.status === 'pending'" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-blue-600 hover:bg-blue-50 dark:text-blue-400 dark:hover:bg-blue-900/20" @click="approveInvoice(item)">
                      <Icon name="check" size="sm" />
                      {{ t('payment.invoices.approve') }}
                    </button>
                    <button v-if="item.status === 'pending' || item.status === 'approved'" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20" @click="openReject(item)">
                      <Icon name="x" size="sm" />
                      {{ t('payment.invoices.reject') }}
                    </button>
                    <button v-if="item.status === 'pending' || item.status === 'approved'" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-green-600 hover:bg-green-50 dark:text-green-400 dark:hover:bg-green-900/20" @click="openIssue(item)">
                      <Icon name="upload" size="sm" />
                      {{ t('payment.invoices.issue') }}
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <BaseDialog :show="!!detailTarget" :title="t('payment.invoices.detailTitle')" width="wide" @close="detailTarget = null">
      <div v-if="detailTarget" class="grid gap-4 sm:grid-cols-2">
        <InfoRow :label="t('payment.invoices.requestId')" :value="`#${detailTarget.id}`" />
        <InfoRow :label="t('payment.admin.colUser')" :value="detailTarget.user_email || `#${detailTarget.user_id}`" />
        <InfoRow :label="t('payment.invoices.amount')" :value="formatInvoiceAmount(detailTarget.amount, detailTarget.currency)" />
        <InfoRow :label="t('payment.invoices.type')" :value="invoiceTypeLabel(detailTarget.invoice_type)" />
        <InfoRow :label="t('payment.invoices.title')" :value="detailTarget.title" />
        <InfoRow :label="t('payment.invoices.taxNumber')" :value="detailTarget.tax_number" />
        <InfoRow :label="t('payment.invoices.status')" :value="invoiceStatusLabel(detailTarget.status)" />
        <InfoRow :label="t('payment.invoices.invoiceNo')" :value="detailTarget.invoice_no || '-'" />
        <InfoRow class="sm:col-span-2" :label="t('payment.invoices.remark')" :value="detailTarget.remark || '-'" />
        <InfoRow class="sm:col-span-2" :label="t('payment.invoices.adminNote')" :value="detailTarget.admin_note || '-'" />
        <InfoRow :label="t('payment.invoices.fileName')" :value="detailTarget.file_name || '-'" />
        <InfoRow :label="t('payment.invoices.issuedAt')" :value="detailTarget.issued_at ? formatDateTime(detailTarget.issued_at) : '-'" />
      </div>
    </BaseDialog>

    <BaseDialog :show="!!rejectTarget" :title="t('payment.invoices.reject')" @close="rejectTarget = null">
      <div class="space-y-3">
        <label class="input-label">{{ t('payment.invoices.adminNote') }}</label>
        <textarea v-model.trim="rejectNote" rows="3" class="input w-full" />
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="rejectTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="actionLoading || !rejectNote.trim()" @click="rejectInvoice">{{ actionLoading ? t('common.processing') : t('payment.invoices.reject') }}</button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="!!issueTarget" :title="t('payment.invoices.issue')" @close="closeIssue">
      <div class="space-y-4">
        <div>
          <label class="input-label">{{ t('payment.invoices.invoiceNo') }}</label>
          <input v-model.trim="issueForm.invoice_no" type="text" class="input mt-1 w-full" />
        </div>
        <div>
          <label class="input-label">{{ t('payment.invoices.adminNote') }}</label>
          <textarea v-model.trim="issueForm.admin_note" rows="2" class="input mt-1 w-full" />
        </div>
        <div>
          <label class="input-label">{{ t('payment.invoices.file') }}</label>
          <input ref="fileInput" type="file" accept=".pdf,.ofd,.jpg,.jpeg,.png" class="input mt-1 w-full" @change="handleFileChange" />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('payment.invoices.fileHint') }}</p>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="closeIssue">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="actionLoading || !issueForm.file" @click="issueInvoice">{{ actionLoading ? t('common.processing') : t('payment.invoices.issue') }}</button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminPaymentAPI } from '@/api/admin/payment'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatOrderDateTime } from '@/components/payment/orderUtils'
import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import type { InvoiceRequest, InvoiceStatus, InvoiceType } from '@/types/payment'

const InfoRow = defineComponent({
  name: 'InfoRow',
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
  },
  setup(props) {
    return () => h('div', { class: 'rounded-lg bg-gray-50 p-3 dark:bg-dark-800' }, [
      h('p', { class: 'text-xs text-gray-500 dark:text-gray-400' }, props.label),
      h('p', { class: 'mt-1 break-words text-sm font-medium text-gray-900 dark:text-white' }, props.value),
    ])
  },
})

const { t, locale } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const actionLoading = ref(false)
const items = ref<InvoiceRequest[]>([])
const detailTarget = ref<InvoiceRequest | null>(null)
const rejectTarget = ref<InvoiceRequest | null>(null)
const issueTarget = ref<InvoiceRequest | null>(null)
const rejectNote = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const filters = reactive<{ status: InvoiceStatus | ''; user_id: number | null }>({ status: '', user_id: null })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const issueForm = reactive<{ invoice_no: string; admin_note: string; file: File | null }>({
  invoice_no: '',
  admin_note: '',
  file: null,
})

const statusOptions = computed(() => [
  { value: '', label: t('payment.admin.allStatuses') },
  { value: 'pending', label: t('payment.invoices.statuses.pending') },
  { value: 'approved', label: t('payment.invoices.statuses.approved') },
  { value: 'rejected', label: t('payment.invoices.statuses.rejected') },
  { value: 'issued', label: t('payment.invoices.statuses.issued') },
])

async function loadInvoices() {
  loading.value = true
  try {
    const res = await adminPaymentAPI.getInvoices({
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
      user_id: filters.user_id || undefined,
    })
    items.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) { pagination.page = page; loadInvoices() }
function handlePageSizeChange(size: number) { pagination.page_size = size; pagination.page = 1; loadInvoices() }

async function openDetail(item: InvoiceRequest) {
  detailTarget.value = item
  try {
    const res = await adminPaymentAPI.getInvoice(item.id)
    detailTarget.value = res.data
  } catch {
    // keep list item as fallback
  }
}

async function approveInvoice(item: InvoiceRequest) {
  actionLoading.value = true
  try {
    await adminPaymentAPI.approveInvoice(item.id, {})
    appStore.showSuccess(t('common.success'))
    await loadInvoices()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function openReject(item: InvoiceRequest) {
  rejectTarget.value = item
  rejectNote.value = item.admin_note || ''
}

async function rejectInvoice() {
  if (!rejectTarget.value || !rejectNote.value.trim()) return
  actionLoading.value = true
  try {
    await adminPaymentAPI.rejectInvoice(rejectTarget.value.id, { admin_note: rejectNote.value.trim() })
    appStore.showSuccess(t('common.success'))
    rejectTarget.value = null
    await loadInvoices()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function openIssue(item: InvoiceRequest) {
  issueTarget.value = item
  issueForm.invoice_no = item.invoice_no || ''
  issueForm.admin_note = item.admin_note || ''
  issueForm.file = null
  if (fileInput.value) fileInput.value.value = ''
}

function closeIssue() {
  issueTarget.value = null
  issueForm.file = null
  if (fileInput.value) fileInput.value.value = ''
}

function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  issueForm.file = target.files?.[0] || null
}

async function issueInvoice() {
  if (!issueTarget.value || !issueForm.file) return
  const form = new FormData()
  form.append('file', issueForm.file)
  form.append('invoice_no', issueForm.invoice_no.trim())
  form.append('admin_note', issueForm.admin_note.trim())
  actionLoading.value = true
  try {
    await adminPaymentAPI.issueInvoice(issueTarget.value.id, form)
    appStore.showSuccess(t('common.success'))
    closeIssue()
    await loadInvoices()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function invoiceTypeLabel(type: InvoiceType): string {
  return t(`payment.invoices.types.${type}`)
}

function invoiceStatusLabel(status: InvoiceStatus): string {
  return t(`payment.invoices.statuses.${status}`)
}

function invoiceStatusClass(status: InvoiceStatus): string {
  const base = 'rounded-full px-2 py-0.5 text-xs font-medium'
  if (status === 'issued') return `${base} bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300`
  if (status === 'approved') return `${base} bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300`
  if (status === 'rejected') return `${base} bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300`
  return `${base} bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300`
}

function localeCode(): string | undefined {
  const raw = locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
}

function formatInvoiceAmount(value: number, currency?: string): string {
  return formatPaymentAmount(value, normalizePaymentCurrency(currency), localeCode())
}

function formatDateTime(dateStr: string): string {
  return formatOrderDateTime(dateStr)
}

onMounted(() => loadInvoices())
</script>
