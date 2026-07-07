<template>
  <AppLayout>
    <div class="space-y-4">
      <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.35fr)]">
        <div class="card p-4">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.invoices.availableAmount') }}</p>
              <p class="mt-2 text-3xl font-semibold text-gray-900 dark:text-white">{{ formatInvoiceAmount(invoiceSummary.available_amount, invoiceSummary.currency) }}</p>
              <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                {{ t('payment.invoices.summaryHint', {
                  eligible: formatInvoiceAmount(invoiceSummary.eligible_amount, invoiceSummary.currency),
                  requested: formatInvoiceAmount(invoiceSummary.requested_amount, invoiceSummary.currency),
                }) }}
              </p>
            </div>
            <button class="btn btn-primary shrink-0" :disabled="invoiceSummaryLoading || invoiceSummary.available_amount <= 0" @click="openInvoiceDialog">
              {{ t('payment.invoices.apply') }}
            </button>
          </div>
        </div>

        <div class="card flex min-h-0 flex-col p-4">
          <div class="mb-3 flex items-center justify-between gap-3">
            <div>
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.invoices.myRequests') }}</h3>
            </div>
            <button @click="fetchInvoices" :disabled="invoiceLoading" class="btn btn-secondary btn-sm" :title="t('common.refresh')">
              <Icon name="refresh" size="sm" :class="invoiceLoading ? 'animate-spin' : ''" />
            </button>
          </div>
          <div class="invoice-request-scroll">
            <table class="min-w-full divide-y divide-gray-200 text-sm dark:divide-dark-600">
              <thead>
                <tr class="text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  <th class="px-3 py-2">{{ t('payment.invoices.amount') }}</th>
                  <th class="px-3 py-2">{{ t('payment.invoices.type') }}</th>
                  <th class="px-3 py-2">{{ t('payment.invoices.status') }}</th>
                  <th class="px-3 py-2">{{ t('payment.orders.createdAt') }}</th>
                  <th class="px-3 py-2 text-right">{{ t('payment.orders.actions') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-if="invoiceLoading">
                  <td colspan="5" class="px-3 py-6 text-center text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</td>
                </tr>
                <tr v-else-if="invoices.length === 0">
                  <td colspan="5" class="px-3 py-6 text-center text-gray-500 dark:text-gray-400">{{ t('payment.invoices.empty') }}</td>
                </tr>
                <tr v-for="item in invoices" v-else :key="item.id" class="text-gray-700 dark:text-gray-200">
                  <td class="whitespace-nowrap px-3 py-2 font-medium">{{ formatInvoiceAmount(item.amount, item.currency) }}</td>
                  <td class="whitespace-nowrap px-3 py-2">{{ invoiceTypeLabel(item.invoice_type) }}</td>
                  <td class="whitespace-nowrap px-3 py-2"><span :class="invoiceStatusClass(item.status)">{{ invoiceStatusLabel(item.status) }}</span></td>
                  <td class="whitespace-nowrap px-3 py-2 text-gray-500 dark:text-gray-400">{{ formatDateTime(item.created_at) }}</td>
                  <td class="whitespace-nowrap px-3 py-2 text-right">
                    <button v-if="item.downloadable" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-[#a9583e] hover:bg-[#f3e7df] dark:text-[#f0b89e] dark:hover:bg-[#cc785c]/20" @click="downloadInvoice(item)">
                      <Icon name="download" size="sm" />
                      {{ item.downloaded_at ? t('payment.invoices.downloaded') : t('payment.invoices.download') }}
                    </button>
                    <span v-else class="text-xs text-gray-400">{{ t('payment.invoices.noFile') }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Filters -->
      <div class="card p-4">
        <div class="flex flex-wrap items-center gap-3">
          <Select v-model="currentFilter" :options="statusFilters" class="w-36" @change="fetchOrders" />
          <div class="flex flex-1 items-center justify-end gap-2">
            <button @click="fetchOrders" :disabled="loading" class="btn btn-secondary" :title="t('common.refresh')">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button class="btn btn-primary" @click="router.push('/purchase')">{{ t('payment.result.backToRecharge') }}</button>
          </div>
        </div>
      </div>

      <!-- Table -->
      <div class="orders-table-scroll">
        <OrderTable :orders="orders" :loading="loading">
          <template #actions="{ row }">
            <div class="flex items-center gap-2">
              <button v-if="row.status === 'PENDING'" @click="handleCancel(row.id)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-yellow-600 hover:bg-yellow-50 dark:text-yellow-400 dark:hover:bg-yellow-900/20">
                <Icon name="x" size="sm" />
                <span>{{ t('payment.orders.cancel') }}</span>
              </button>
              <button v-if="canRequestRefund(row)" @click="openRefundDialog(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-[#a9583e] hover:bg-[#f3e7df] dark:text-[#f0b89e] dark:hover:bg-[#cc785c]/20">
                <Icon name="dollar" size="sm" />
                <span>{{ t('payment.orders.requestRefund') }}</span>
              </button>
            </div>
          </template>
        </OrderTable>
      </div>

      <!-- Pagination -->
      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <!-- Invoice Dialog -->
    <BaseDialog :show="showInvoiceDialog" :title="t('payment.invoices.applyTitle')" @close="closeInvoiceDialog">
      <div class="space-y-4">
        <div>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.invoices.availableAmount') }}</p>
          <p class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">{{ formatInvoiceAmount(invoiceSummary.available_amount, invoiceSummary.currency) }}</p>
          <p class="mt-2 text-sm text-[#a9583e] dark:text-[#f0b89e]">{{ t('payment.invoices.applyHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('payment.invoices.amount') }}</label>
          <input v-model.number="invoiceForm.amount" type="number" min="0" step="0.01" class="input mt-1 w-full" />
        </div>
        <div>
          <label class="input-label">{{ t('payment.invoices.type') }}</label>
          <div class="mt-2 flex flex-wrap gap-2">
            <button
              v-for="option in invoiceTypeOptions"
              :key="option.value"
              type="button"
              class="rounded-md px-3 py-2 text-sm font-medium"
              :class="invoiceForm.invoice_type === option.value ? 'bg-[#141413] text-white dark:bg-white dark:text-gray-900' : 'bg-[#f5f0e8] text-[#504f49] hover:bg-[#efe9de] dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
              @click="invoiceForm.invoice_type = option.value"
            >
              {{ option.label }}
            </button>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('payment.invoices.title') }}</label>
          <input v-model.trim="invoiceForm.title" type="text" class="input mt-1 w-full" />
        </div>
        <div>
          <label class="input-label">{{ t('payment.invoices.taxNumber') }}</label>
          <input v-model.trim="invoiceForm.tax_number" type="text" class="input mt-1 w-full" />
        </div>
        <div>
          <label class="input-label">{{ t('payment.invoices.remark') }}</label>
          <input v-model.trim="invoiceForm.remark" type="text" class="input mt-1 w-full" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="closeInvoiceDialog">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="invoiceSubmitting || !canSubmitInvoice" @click="submitInvoice">{{ invoiceSubmitting ? t('common.processing') : t('common.submit') }}</button>
        </div>
      </template>
    </BaseDialog>

    <!-- Cancel Confirm Dialog -->
    <BaseDialog :show="!!cancelTargetId" :title="t('payment.orders.cancel')" width="narrow" @close="cancelTargetId = null">
      <p class="text-sm text-gray-600 dark:text-gray-300">{{ t('payment.confirmCancel') }}</p>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="cancelTargetId = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="actionLoading" @click="confirmCancel">{{ actionLoading ? t('common.processing') : t('payment.orders.cancel') }}</button>
        </div>
      </template>
    </BaseDialog>

    <!-- Refund Dialog -->
    <BaseDialog :show="!!refundTarget" :title="t('payment.orders.requestRefund')" @close="refundTarget = null">
      <div v-if="refundTarget" class="space-y-4">
        <div class="rounded-lg border border-[#d8cec2] bg-[#f5f0e8] p-4 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</span>
            <span class="font-mono text-gray-900 dark:text-white">#{{ refundTarget.id }}</span>
          </div>
          <div class="mt-2 flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.amount') }}</span>
            <span class="text-gray-900 dark:text-white">${{ refundTarget.amount.toFixed(2) }}</span>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('payment.refundReason') }}</label>
          <textarea v-model="refundReason" rows="3" class="input mt-1 w-full" :placeholder="t('payment.refundReasonPlaceholder')" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="refundTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="actionLoading || !refundReason.trim()" @click="confirmRefund">{{ actionLoading ? t('common.processing') : t('payment.orders.requestRefund') }}</button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatOrderDateTime } from '@/components/payment/orderUtils'
import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import type { CreateInvoiceRequest, InvoiceRequest, InvoiceStatus, InvoiceSummary, InvoiceType, PaymentOrder } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderTable from '@/components/payment/OrderTable.vue'

const { t, locale } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const TICKET_UNREAD_BADGE_REFRESH_EVENT = 'sub2api:ticket-unread-updated'

const loading = ref(false)
const actionLoading = ref(false)
const invoiceSummaryLoading = ref(false)
const invoiceLoading = ref(false)
const invoiceSubmitting = ref(false)
const orders = ref<PaymentOrder[]>([])
const invoices = ref<InvoiceRequest[]>([])
const refundEligibleProviders = ref<Set<string>>(new Set())
const currentFilter = ref('')
const cancelTargetId = ref<number | null>(null)
const refundTarget = ref<PaymentOrder | null>(null)
const refundReason = ref('')
const showInvoiceDialog = ref(false)
const invoiceSummary = reactive<InvoiceSummary>({ currency: 'CNY', eligible_amount: 0, requested_amount: 0, available_amount: 0 })
const invoiceForm = reactive<CreateInvoiceRequest>({
  amount: 0,
  invoice_type: 'vat_general',
  title: '',
  tax_number: '',
  remark: '',
})
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const statusFilters = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
])

const invoiceTypeOptions = computed<Array<{ value: InvoiceType; label: string }>>(() => [
  { value: 'vat_general', label: t('payment.invoices.types.vat_general') },
  { value: 'vat_special', label: t('payment.invoices.types.vat_special') },
])

const canSubmitInvoice = computed(() => {
  return invoiceForm.amount > 0
    && invoiceForm.amount <= invoiceSummary.available_amount
    && invoiceForm.title.trim() !== ''
    && invoiceForm.tax_number.trim() !== ''
})

async function fetchOrders() {
  loading.value = true
  try {
    const res = await paymentAPI.getMyOrders({
      page: pagination.page,
      page_size: pagination.page_size,
      status: currentFilter.value || undefined,
    })
    orders.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function fetchInvoiceSummary() {
  invoiceSummaryLoading.value = true
  try {
    const res = await paymentAPI.getInvoiceSummary()
    Object.assign(invoiceSummary, res.data)
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    invoiceSummaryLoading.value = false
  }
}

async function fetchInvoices() {
  invoiceLoading.value = true
  try {
    const res = await paymentAPI.getMyInvoices({ page: 1, page_size: 20 })
    invoices.value = res.data.items || []
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    invoiceLoading.value = false
  }
}

function handlePageChange(page: number) { pagination.page = page; fetchOrders() }
function handlePageSizeChange(size: number) { pagination.page_size = size; pagination.page = 1; fetchOrders() }

function handleCancel(orderId: number) { cancelTargetId.value = orderId }

async function confirmCancel() {
  if (!cancelTargetId.value) return
  actionLoading.value = true
  try {
    await paymentAPI.cancelOrder(cancelTargetId.value)
    appStore.showSuccess(t('common.success'))
    cancelTargetId.value = null
    await fetchOrders()
    await fetchInvoiceSummary()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function openRefundDialog(order: PaymentOrder) { refundTarget.value = order; refundReason.value = '' }

async function confirmRefund() {
  if (!refundTarget.value || !refundReason.value.trim()) return
  actionLoading.value = true
  try {
    await paymentAPI.requestRefund(refundTarget.value.id, { reason: refundReason.value.trim() })
    appStore.showSuccess(t('common.success'))
    refundTarget.value = null
    refundReason.value = ''
    await fetchOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function canRequestRefund(order: PaymentOrder): boolean {
  if (order.status !== 'COMPLETED') return false
  if (!order.provider_instance_id) return false
  return refundEligibleProviders.value.has(order.provider_instance_id)
}

async function loadRefundEligibility() {
  try {
    const res = await paymentAPI.getRefundEligibleProviders()
    refundEligibleProviders.value = new Set(res.data.provider_instance_ids || [])
  } catch { /* ignore, default to hiding refund button */ }
}

function openInvoiceDialog() {
  invoiceForm.amount = Number(invoiceSummary.available_amount.toFixed(2))
  invoiceForm.invoice_type = 'vat_general'
  invoiceForm.title = ''
  invoiceForm.tax_number = ''
  invoiceForm.remark = ''
  showInvoiceDialog.value = true
}

function closeInvoiceDialog() {
  showInvoiceDialog.value = false
}

async function submitInvoice() {
  if (!canSubmitInvoice.value) return
  invoiceSubmitting.value = true
  try {
    await paymentAPI.createInvoice({
      amount: Number(invoiceForm.amount),
      invoice_type: invoiceForm.invoice_type,
      title: invoiceForm.title.trim(),
      tax_number: invoiceForm.tax_number.trim(),
      remark: invoiceForm.remark?.trim() || undefined,
    })
    appStore.showSuccess(t('payment.invoices.submitSuccess'))
    closeInvoiceDialog()
    await Promise.all([fetchInvoiceSummary(), fetchInvoices()])
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    invoiceSubmitting.value = false
  }
}

async function downloadInvoice(item: InvoiceRequest) {
  try {
    const res = await paymentAPI.downloadInvoice(item.id)
    const blob = res.data
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = item.file_name || `invoice-${item.id}`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    window.dispatchEvent(new Event(TICKET_UNREAD_BADGE_REFRESH_EVENT))
    await fetchInvoices()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
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
  if (status === 'issued') return `${base} bg-[#9ab3a0]/20 text-[#5f7f68] dark:bg-[#9ab3a0]/10 dark:text-[#9ab3a0]`
  if (status === 'approved') return `${base} bg-[#f5f0e8] text-[#a9583e] dark:bg-[#cc785c]/20 dark:text-[#f0b89e]`
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

onMounted(() => {
  fetchOrders()
  loadRefundEligibility()
  fetchInvoiceSummary()
  fetchInvoices()
})
</script>

<style scoped>
.invoice-request-scroll {
  max-height: 13.5rem;
  overflow: auto;
  scrollbar-gutter: stable;
}

.orders-table-scroll {
  display: flex;
  min-height: 18rem;
  max-height: min(32rem, calc(100vh - 24rem));
  overflow: auto;
  scrollbar-gutter: stable;
}

.orders-table-scroll :deep(.table-wrapper) {
  width: 100%;
}

@media (max-width: 767px) {
  .orders-table-scroll {
    display: block;
    max-height: 34rem;
    overflow-y: auto;
  }
}
</style>
