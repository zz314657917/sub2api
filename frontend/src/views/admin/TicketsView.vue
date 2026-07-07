<template>
  <AppLayout>
    <div class="flex h-[calc(100vh-5rem)] min-h-[640px] overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
      <aside
        class="ticket-list-panel"
        :class="selectedTicketId && !showMobileList ? 'hidden md:flex' : 'flex'"
      >
        <div class="border-b border-gray-200 p-4 dark:border-dark-700">
          <div class="mb-4">
            <h1 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.tickets.title') }}</h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.tickets.description') }}</p>
          </div>

          <div class="space-y-3">
            <div class="rounded-md border border-[#d8cec2] bg-[#fffaf5] px-3 py-2 text-xs leading-5 text-[#6c6a64] dark:border-[#cc785c]/30 dark:bg-[#cc785c]/10 dark:text-[#f0b89e]">
              {{ filters.ticket_type === 'system' ? t('admin.tickets.systemAuditHint') : t('admin.tickets.supportDefaultHint') }}
            </div>
            <div class="relative">
              <Icon name="search" size="sm" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input v-model="filters.search" class="input pl-9" :placeholder="t('admin.tickets.searchPlaceholder')" @input="handleSearchInput" />
            </div>
            <div class="grid grid-cols-2 gap-2">
              <select v-model="filters.status" class="input" @change="reloadFromFirstPage">
                <option value="">{{ t('admin.tickets.allStatus') }}</option>
                <option value="open">{{ t('admin.tickets.open') }}</option>
                <option value="pending_admin">{{ t('admin.tickets.pending_admin') }}</option>
                <option value="pending_user">{{ t('admin.tickets.pending_user') }}</option>
                <option value="closed">{{ t('admin.tickets.closed') }}</option>
              </select>
              <select v-model="filters.ticket_type" class="input" @change="reloadFromFirstPage">
                <option value="support">{{ t('admin.tickets.supportTicket') }}</option>
                <option value="system">{{ t('admin.tickets.systemTicket') }}</option>
                <option value="">{{ t('admin.tickets.allTypes') }}</option>
              </select>
            </div>
            <div class="grid grid-cols-1 gap-2">
              <input
                v-model.trim="filters.user_id"
                type="number"
                min="1"
                class="input"
                :placeholder="t('admin.tickets.userIdPlaceholder')"
                @input="handleSearchInput"
              />
            </div>
            <div v-if="filters.ticket_type === 'system'" class="grid grid-cols-2 gap-2">
              <input
                v-model.trim="filters.event_type"
                data-test="ticket-event-type-filter"
                class="input"
                :placeholder="t('admin.tickets.eventTypePlaceholder')"
                @input="handleSearchInput"
              />
              <input
                v-model.trim="filters.event_key"
                data-test="ticket-event-key-filter"
                class="input"
                :placeholder="t('admin.tickets.eventKeyPlaceholder')"
                @input="handleSearchInput"
              />
            </div>
            <div class="grid grid-cols-2 gap-2">
              <input
                v-model="filters.date_from"
                data-test="ticket-date-from-filter"
                type="date"
                class="input"
                :aria-label="t('admin.tickets.dateFrom')"
                @change="reloadFromFirstPage"
              />
              <input
                v-model="filters.date_to"
                data-test="ticket-date-to-filter"
                type="date"
                class="input"
                :aria-label="t('admin.tickets.dateTo')"
                @change="reloadFromFirstPage"
              />
            </div>
            <div class="grid grid-cols-[minmax(0,1fr)_auto_auto] gap-2">
              <select v-model="filters.sort_by" data-test="ticket-sort-filter" class="input" @change="reloadFromFirstPage">
                <option value="last_message_at">{{ t('admin.tickets.sortLatest') }}</option>
                <option value="unread_first">{{ t('admin.tickets.sortUnreadFirst') }}</option>
              </select>
              <button
                type="button"
                class="btn justify-center"
                :class="filters.unread_only ? 'btn-primary' : 'btn-secondary'"
                @click="toggleUnreadOnly"
              >
                <Icon name="eye" size="sm" />
                {{ t('admin.tickets.unreadOnly') }}
              </button>
              <button type="button" class="btn btn-secondary px-3" :title="t('admin.tickets.refresh')" @click="loadTickets">
                <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
              </button>
            </div>
          </div>
        </div>

        <div class="min-h-0 flex-1 overflow-y-auto">
          <button
            v-for="ticket in tickets"
            :key="ticket.id"
            type="button"
            class="ticket-list-item"
            :class="{ 'ticket-list-item-active': ticket.id === selectedTicketId }"
            @click="selectTicket(ticket.id)"
          >
            <div class="ticket-list-main">
              <div class="ticket-list-content">
                <div class="ticket-list-title-row">
                  <span class="ticket-title">{{ ticket.title }}</span>
                  <span v-if="ticket.admin_unread_count > 0" class="unread-pill">{{ ticket.admin_unread_count }}</span>
                </div>
                <p class="ticket-list-user">
                  {{ ticket.user?.email || ticket.user?.username || t('admin.tickets.userInfo', { id: ticket.user_id }) }}
                </p>
                <p class="ticket-list-preview">{{ formatTicketPreview(ticket.last_message_preview) }}</p>
              </div>
              <span v-if="!isSystemTicket(ticket)" :class="statusClass(ticket.status, ticket.ticket_type)">{{ statusLabel(ticket) }}</span>
            </div>
            <div class="mt-3 flex items-center justify-between text-xs text-gray-400 dark:text-dark-500">
              <span>#{{ ticket.id }}</span>
              <span>{{ formatRelativeTime(ticket.last_message_at || ticket.updated_at) }}</span>
            </div>
          </button>

          <div v-if="!loading && tickets.length === 0" class="flex h-full min-h-[260px] flex-col items-center justify-center px-6 text-center text-gray-500 dark:text-dark-400">
            <Icon name="ticket" size="xl" class="mb-3 text-gray-300 dark:text-dark-600" />
            <p>{{ t('admin.tickets.noTickets') }}</p>
          </div>
        </div>

        <div class="border-t border-gray-200 p-3 dark:border-dark-700">
          <Pagination
            :total="pagination.total"
            :page="pagination.page"
            :page-size="pagination.page_size"
            :show-page-size-selector="false"
            compact
            @update:page="handlePageChange"
            @update:page-size="handlePageSizeChange"
          />
        </div>
      </aside>

      <section class="min-w-0 flex-1 flex-col" :class="detailPanelClass">
        <div v-if="selectedTicket" class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200 p-4 dark:border-dark-700">
          <div class="min-w-0">
            <button type="button" class="mb-2 inline-flex items-center gap-1 text-sm text-primary-600 md:hidden" @click="showMobileList = true">
              <Icon name="chevronLeft" size="sm" />
              {{ t('admin.tickets.title') }}
            </button>
            <h2 class="truncate text-lg font-semibold text-gray-900 dark:text-white">{{ selectedTicket.title }}</h2>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
              #{{ selectedTicket.id }} · {{ selectedTicket.user?.email || selectedTicket.user?.username || t('admin.tickets.userInfo', { id: selectedTicket.user_id }) }} · {{ formatDateTime(selectedTicket.created_at) }}
            </p>
            <div v-if="selectedIsSystem" class="mt-2">
              <span class="badge badge-warning">{{ t('admin.tickets.readOnlyStatus') }}</span>
            </div>
          </div>
          <div class="flex shrink-0 flex-wrap items-center gap-2">
            <button v-if="selectedTicket.admin_unread_count > 0" class="btn btn-secondary btn-sm" :disabled="acting" @click="handleMarkRead">
              <Icon name="check" size="sm" />
              {{ t('admin.tickets.markRead') }}
            </button>
            <button v-if="!selectedIsSystem && selectedTicket.status !== 'closed'" class="btn btn-secondary btn-sm" :disabled="acting" @click="handleCloseTicket">
              <Icon name="lock" size="sm" />
              {{ t('admin.tickets.close') }}
            </button>
            <button v-else-if="!selectedIsSystem" class="btn btn-secondary btn-sm" :disabled="acting" @click="handleReopenTicket">
              <Icon name="refresh" size="sm" />
              {{ t('admin.tickets.reopen') }}
            </button>
          </div>
        </div>

        <div v-if="selectedTicket" ref="messageListRef" class="min-h-0 flex-1 space-y-4 overflow-y-auto bg-gray-50 p-4 dark:bg-dark-950">
          <div
            v-for="message in messages"
            :key="message.id"
            class="flex"
            :class="message.sender_type === 'admin' ? 'justify-end' : 'justify-start'"
          >
            <div :class="messageBubbleClass(message.sender_type)">
              <div class="message-meta">
                <span>{{ senderLabel(message.sender_type) }}</span>
                <span>{{ formatDateTime(message.created_at) }}</span>
              </div>
              <p class="whitespace-pre-wrap break-words text-sm leading-6">{{ formatMessageContent(message.content) }}</p>
              <SystemTicketMetadataDetails :message="message" />
            </div>
          </div>
          <div v-if="!detailLoading && messages.length === 0" class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('admin.tickets.noMessages') }}
          </div>
        </div>

        <form v-if="selectedTicket && !selectedIsSystem" class="border-t border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900" @submit.prevent="handleSendMessage">
          <div class="flex gap-3">
            <textarea
              v-model.trim="replyContent"
              class="input min-h-[76px] flex-1 resize-none"
              :placeholder="selectedTicket.status === 'closed' ? t('admin.tickets.closed') : t('admin.tickets.replyPlaceholder')"
              :disabled="selectedTicket.status === 'closed'"
            ></textarea>
            <button class="btn btn-primary self-end" :disabled="sending || selectedTicket.status === 'closed'">
              <Icon name="mail" size="sm" />
              {{ sending ? t('common.saving') : t('admin.tickets.send') }}
            </button>
          </div>
        </form>

        <div v-else-if="selectedTicket" class="border-t border-gray-200 bg-white p-4 text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400">
          {{ t('admin.tickets.systemReadOnly') }}
        </div>

        <div v-else class="flex flex-1 flex-col items-center justify-center bg-gray-50 p-6 text-center text-gray-500 dark:bg-dark-950 dark:text-dark-400">
          <Icon name="ticket" size="xl" class="mb-3 text-gray-300 dark:text-dark-600" />
          <p>{{ t('admin.tickets.noTicketSelected') }}</p>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import SystemTicketMetadataDetails from '@/components/tickets/SystemTicketMetadataDetails.vue'
import { adminTicketsAPI } from '@/api/admin/tickets'
import { useAppStore } from '@/stores/app'
import { formatDateTime, formatRelativeTime } from '@/utils/format'
import type { AdminSupportTicketSortBy, SupportTicket, SupportTicketMessage, SupportTicketStatus, SupportTicketType } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const TICKET_UNREAD_BADGE_REFRESH_EVENT = 'sub2api:ticket-unread-updated'

const tickets = ref<SupportTicket[]>([])
const selectedTicketId = ref<number | null>(null)
const selectedTicket = ref<SupportTicket | null>(null)
const messages = ref<SupportTicketMessage[]>([])
const loading = ref(false)
const detailLoading = ref(false)
const sending = ref(false)
const acting = ref(false)
const showMobileList = ref(true)
const replyContent = ref('')
const messageListRef = ref<HTMLElement | null>(null)
let searchTimer: ReturnType<typeof setTimeout> | null = null

const filters = reactive({
  search: '',
  status: '' as SupportTicketStatus | '',
  ticket_type: 'support' as SupportTicketType | '',
  user_id: '',
  event_type: '',
  event_key: '',
  date_from: '',
  date_to: '',
  unread_only: false,
  sort_by: 'last_message_at' as AdminSupportTicketSortBy,
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0,
})

const currentListParams = computed(() => ({
  page: pagination.page,
  page_size: pagination.page_size,
  status: filters.status,
  ticket_type: filters.ticket_type,
  search: filters.search.trim() || undefined,
  user_id: filters.user_id.trim() || undefined,
  event_type: filters.event_type.trim() || undefined,
  event_key: filters.event_key.trim() || undefined,
  date_from: filters.date_from || undefined,
  date_to: filters.date_to || undefined,
  unread_only: filters.unread_only,
  sort_by: filters.sort_by,
  sort_order: 'desc' as const,
}))

const selectedIsSystem = computed(() => selectedTicket.value?.ticket_type === 'system')
const detailPanelClass = computed(() => (selectedTicketId.value && !showMobileList.value ? 'flex' : 'hidden md:flex'))

function isSystemTicket(ticket: SupportTicket) {
  return ticket.ticket_type === 'system'
}

function statusLabel(ticket: SupportTicket) {
  if (isSystemTicket(ticket)) return t('admin.tickets.systemTicket')
  return t(`admin.tickets.${ticket.status}`)
}

function statusClass(status: SupportTicketStatus, ticketType: SupportTicket['ticket_type'] = 'support') {
  return [
    'badge',
    ticketType === 'system' ? 'badge-warning' : status === 'closed' ? 'badge-gray' : status === 'pending_admin' ? 'badge-warning' : 'badge-success',
  ]
}

function senderLabel(senderType: SupportTicketMessage['sender_type']) {
  if (senderType === 'admin') return t('admin.tickets.adminSender')
  if (senderType === 'system') return t('admin.tickets.systemMessage')
  return t('admin.tickets.userSender')
}

function messageBubbleClass(senderType: SupportTicketMessage['sender_type']) {
  return [
    'message-bubble',
    senderType === 'admin'
      ? 'message-bubble-admin'
      : senderType === 'user'
        ? 'message-bubble-peer'
        : 'message-bubble-system',
  ]
}

function formatTicketPreview(preview: string | null | undefined): string {
  const text = (preview || '-').trim()
  return formatInlineArrows(text)
}

function formatMessageContent(content: string): string {
  return formatInlineArrows(content)
}

function formatInlineArrows(text: string): string {
  return text.replace(/\s*(?:->|→)\s*/g, '\u00a0→\u00a0')
}

function shouldAutoSelectInitialTicket(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return true
  return window.matchMedia('(min-width: 768px)').matches
}

function applyTicketReadState(ticketId: number) {
  tickets.value = tickets.value.map((ticket) => (
    ticket.id === ticketId ? { ...ticket, admin_unread_count: 0 } : ticket
  ))
  if (selectedTicket.value?.id === ticketId) {
    selectedTicket.value = { ...selectedTicket.value, admin_unread_count: 0 }
  }
}

function dispatchTicketUnreadBadgeRefresh() {
  window.dispatchEvent(new Event(TICKET_UNREAD_BADGE_REFRESH_EVENT))
}

async function markTicketRead(ticketId: number, silent = false) {
  try {
    await adminTicketsAPI.markRead(ticketId)
    applyTicketReadState(ticketId)
    dispatchTicketUnreadBadgeRefresh()
    if (!silent) appStore.showSuccess(t('admin.tickets.readSuccess'))
  } catch (error: any) {
    if (!silent) {
      appStore.showError(error.message || t('admin.tickets.readFailed'))
    }
  }
}

async function loadTickets() {
  loading.value = true
  try {
    const response = await adminTicketsAPI.list(currentListParams.value)
    tickets.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
    if (!selectedTicketId.value && response.items.length > 0 && shouldAutoSelectInitialTicket()) {
      await selectTicket(response.items[0].id, false)
    } else if (selectedTicketId.value && !response.items.some((item) => item.id === selectedTicketId.value)) {
      selectedTicketId.value = null
      selectedTicket.value = null
      messages.value = []
      showMobileList.value = true
    } else if (selectedTicketId.value) {
      selectedTicket.value = response.items.find((item) => item.id === selectedTicketId.value) ?? selectedTicket.value
    }
  } catch (error: any) {
    appStore.showError(error.message || t('admin.tickets.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function loadDetail(ticketId: number) {
  detailLoading.value = true
  try {
    const detail = await adminTicketsAPI.get(ticketId)
    selectedTicket.value = detail.ticket
    messages.value = detail.messages
    await nextTick()
    messageListRef.value?.scrollTo?.({ top: messageListRef.value.scrollHeight })
    if (detail.ticket.admin_unread_count > 0) {
      await markTicketRead(ticketId, true)
    }
  } catch (error: any) {
    appStore.showError(error.message || t('admin.tickets.detailLoadFailed'))
  } finally {
    detailLoading.value = false
  }
}

async function selectTicket(ticketId: number, closeMobileList = true) {
  selectedTicketId.value = ticketId
  if (closeMobileList) showMobileList.value = false
  await loadDetail(ticketId)
}

function reloadFromFirstPage() {
  pagination.page = 1
  if (filters.ticket_type !== 'system') {
    filters.event_type = ''
    filters.event_key = ''
  }
  void loadTickets()
}

function handleSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(reloadFromFirstPage, 300)
}

function toggleUnreadOnly() {
  filters.unread_only = !filters.unread_only
  reloadFromFirstPage()
}

function handlePageChange(page: number) {
  pagination.page = page
  void loadTickets()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadTickets()
}

async function handleSendMessage() {
  if (selectedIsSystem.value) {
    appStore.showError(t('admin.tickets.systemReadOnly'))
    return
  }
  if (!selectedTicketId.value || !replyContent.value.trim()) {
    appStore.showError(t('admin.tickets.contentRequired'))
    return
  }
  sending.value = true
  try {
    await adminTicketsAPI.createMessage(selectedTicketId.value, { content: replyContent.value.trim() })
    replyContent.value = ''
    appStore.showSuccess(t('admin.tickets.sent'))
    await loadDetail(selectedTicketId.value)
    await loadTickets()
  } catch (error: any) {
    appStore.showError(error.message || t('admin.tickets.sendFailed'))
  } finally {
    sending.value = false
  }
}

async function handleMarkRead() {
  if (!selectedTicketId.value) return
  acting.value = true
  try {
    await markTicketRead(selectedTicketId.value)
    await loadTickets()
  } finally {
    acting.value = false
  }
}

async function handleCloseTicket() {
  if (selectedIsSystem.value) return
  if (!selectedTicketId.value) return
  acting.value = true
  try {
    await adminTicketsAPI.close(selectedTicketId.value)
    appStore.showSuccess(t('admin.tickets.closedSuccess'))
    await loadDetail(selectedTicketId.value)
    await loadTickets()
  } catch (error: any) {
    appStore.showError(error.message || t('admin.tickets.closeFailed'))
  } finally {
    acting.value = false
  }
}

async function handleReopenTicket() {
  if (selectedIsSystem.value) return
  if (!selectedTicketId.value) return
  acting.value = true
  try {
    await adminTicketsAPI.reopen(selectedTicketId.value)
    appStore.showSuccess(t('admin.tickets.reopenedSuccess'))
    await loadDetail(selectedTicketId.value)
    await loadTickets()
  } catch (error: any) {
    appStore.showError(error.message || t('admin.tickets.reopenFailed'))
  } finally {
    acting.value = false
  }
}

onMounted(() => {
  void loadTickets()
})
</script>

<style scoped>
.ticket-list-panel {
  width: 100%;
  min-width: 0;
  flex-direction: column;
  border-right: 1px solid rgb(229 231 235);
}

@media (min-width: 768px) {
  .ticket-list-panel {
    width: 23rem;
    flex: 0 0 23rem;
  }
}

.dark .ticket-list-panel {
  border-right-color: rgb(55 65 81);
}

.ticket-list-item {
  display: block;
  width: 100%;
  border-bottom: 1px solid rgb(229 231 235);
  padding: 1rem;
  text-align: left;
  transition: background-color 0.15s ease;
}

.ticket-list-main {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.75rem;
  align-items: start;
}

.ticket-list-content {
  min-width: 0;
}

.ticket-list-title-row {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.375rem;
}

.ticket-title {
  min-width: 0;
  max-width: 100%;
  overflow-wrap: anywhere;
  color: rgb(17 24 39);
  font-weight: 600;
  line-height: 1.375rem;
}

.ticket-list-user {
  margin-top: 0.25rem;
  overflow: hidden;
  color: rgb(107 114 128);
  font-size: 0.75rem;
  line-height: 1.125rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ticket-list-preview {
  margin-top: 0.375rem;
  display: -webkit-box;
  overflow: hidden;
  color: rgb(107 114 128);
  font-size: 0.875rem;
  line-height: 1.35rem;
  overflow-wrap: anywhere;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.ticket-list-item:hover,
.ticket-list-item-active {
  background: rgb(249 250 251);
}

.dark .ticket-list-item {
  border-bottom-color: rgb(55 65 81);
}

.dark .ticket-title {
  color: white;
}

.dark .ticket-list-user,
.dark .ticket-list-preview {
  color: rgb(156 163 175);
}

.dark .ticket-list-item:hover,
.dark .ticket-list-item-active {
  background: rgb(31 41 55);
}

.unread-pill {
  display: inline-flex;
  flex: 0 0 auto;
  min-width: 1.25rem;
  height: 1.25rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgb(239 68 68);
  padding: 0 0.35rem;
  font-size: 0.75rem;
  font-weight: 700;
  color: white;
}

.message-bubble {
  width: fit-content;
  max-width: min(720px, 86%);
  border-radius: 0.5rem;
  padding: 0.75rem 1rem;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.06);
}

.message-bubble-system {
  width: min(720px, 100%);
  max-width: 100%;
  border: 1px solid rgb(216 206 194);
  background: rgb(255 250 245);
  color: rgb(80 79 73);
}

.message-bubble-admin {
  background: rgb(20 20 19);
  color: white;
}

.message-bubble-peer {
  background: white;
  color: rgb(17 24 39);
}

.message-meta {
  margin-bottom: 0.25rem;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.25rem 1rem;
  font-size: 0.75rem;
  line-height: 1rem;
  opacity: 0.75;
}

.dark .message-bubble-system {
  border-color: rgba(204, 120, 92, 0.32);
  background: rgba(204, 120, 92, 0.12);
  color: rgb(240 184 158);
}

.dark .message-bubble-peer {
  background: rgb(31 41 55);
  color: rgb(243 244 246);
}

@media (max-width: 767px) {
  .message-bubble {
    max-width: 100%;
  }
}
</style>
