<template>
  <AppLayout>
    <div class="flex h-[calc(100vh-5rem)] min-h-[640px] overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
      <aside
        class="ticket-list-panel"
        :class="selectedTicketId && !showMobileList ? 'hidden md:flex' : 'flex'"
      >
        <div class="border-b border-gray-200 p-4 dark:border-dark-700">
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h1 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('tickets.title') }}</h1>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('tickets.description') }}</p>
            </div>
            <button class="btn btn-primary shrink-0 px-3" @click="showCreateForm = !showCreateForm">
              <Icon name="plus" size="sm" />
              <span class="hidden sm:inline">{{ t('tickets.newTicket') }}</span>
            </button>
          </div>

          <form v-if="showCreateForm" class="mb-4 space-y-3 rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-800" @submit.prevent="handleCreateTicket">
            <input v-model.trim="createForm.title" class="input" :placeholder="t('tickets.titlePlaceholder')" />
            <textarea v-model.trim="createForm.content" class="input min-h-[96px]" :placeholder="t('tickets.contentPlaceholder')"></textarea>
            <div class="flex justify-end gap-2">
              <button type="button" class="btn btn-secondary btn-sm" @click="showCreateForm = false">{{ t('common.cancel') }}</button>
              <button type="submit" class="btn btn-primary btn-sm" :disabled="creating">{{ creating ? t('common.saving') : t('tickets.newTicket') }}</button>
            </div>
          </form>

          <div class="space-y-3">
            <div class="relative">
              <Icon name="search" size="sm" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input v-model="filters.search" class="input pl-9" :placeholder="t('tickets.searchPlaceholder')" @input="handleSearchInput" />
            </div>
            <div class="grid grid-cols-[1fr_auto_auto] gap-2">
              <select v-model="filters.status" class="input" @change="reloadFromFirstPage">
                <option value="">{{ t('tickets.allStatus') }}</option>
                <option value="open">{{ t('tickets.open') }}</option>
                <option value="pending_admin">{{ t('tickets.pending_admin') }}</option>
                <option value="pending_user">{{ t('tickets.pending_user') }}</option>
                <option value="closed">{{ t('tickets.closed') }}</option>
              </select>
              <button
                type="button"
                class="btn px-3"
                :class="filters.unread_only ? 'btn-primary' : 'btn-secondary'"
                :title="t('tickets.unreadOnly')"
                @click="toggleUnreadOnly"
              >
                <Icon name="eye" size="sm" />
              </button>
              <button type="button" class="btn btn-secondary px-3" :title="t('tickets.refresh')" @click="loadTickets">
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
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <span class="truncate font-medium text-gray-900 dark:text-white">{{ ticket.title }}</span>
                  <span v-if="isSystemTicket(ticket)" class="badge badge-warning">{{ t('tickets.systemTicket') }}</span>
                  <span v-if="ticket.user_unread_count > 0" class="unread-pill">{{ ticket.user_unread_count }}</span>
                </div>
                <p class="mt-1 line-clamp-2 text-sm text-gray-500 dark:text-dark-400">{{ ticket.last_message_preview || '-' }}</p>
              </div>
              <span :class="statusClass(ticket.status, ticket.ticket_type)">{{ statusLabel(ticket) }}</span>
            </div>
            <div class="mt-3 flex items-center justify-between text-xs text-gray-400 dark:text-dark-500">
              <span>#{{ ticket.id }}</span>
              <span>{{ formatRelativeTime(ticket.last_message_at || ticket.updated_at) }}</span>
            </div>
          </button>

          <div v-if="!loading && tickets.length === 0" class="flex h-full min-h-[260px] flex-col items-center justify-center px-6 text-center text-gray-500 dark:text-dark-400">
            <Icon name="ticket" size="xl" class="mb-3 text-gray-300 dark:text-dark-600" />
            <p>{{ t('tickets.noTickets') }}</p>
          </div>
        </div>

        <div class="border-t border-gray-200 p-3 dark:border-dark-700">
          <Pagination
            :total="pagination.total"
            :page="pagination.page"
            :page-size="pagination.page_size"
            :show-page-size-selector="false"
            @update:page="handlePageChange"
            @update:page-size="handlePageSizeChange"
          />
        </div>
      </aside>

      <section class="flex min-w-0 flex-1 flex-col">
        <div v-if="selectedTicket" class="flex items-center justify-between gap-3 border-b border-gray-200 p-4 dark:border-dark-700">
          <div class="min-w-0">
            <button type="button" class="mb-2 inline-flex items-center gap-1 text-sm text-primary-600 md:hidden" @click="showMobileList = true">
              <Icon name="chevronLeft" size="sm" />
              {{ t('tickets.title') }}
            </button>
            <h2 class="truncate text-lg font-semibold text-gray-900 dark:text-white">{{ selectedTicket.title }}</h2>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">#{{ selectedTicket.id }} · {{ formatDateTime(selectedTicket.created_at) }}</p>
          </div>
          <div class="flex shrink-0 items-center gap-2">
            <button v-if="selectedTicket.user_unread_count > 0" class="btn btn-secondary btn-sm" :disabled="acting" @click="handleMarkRead">
              <Icon name="check" size="sm" />
              {{ t('tickets.markRead') }}
            </button>
            <button v-if="!selectedIsSystem && selectedTicket.status !== 'closed'" class="btn btn-secondary btn-sm" :disabled="acting" @click="handleCloseTicket">
              <Icon name="lock" size="sm" />
              {{ t('tickets.close') }}
            </button>
          </div>
        </div>

        <div v-if="selectedTicket" ref="messageListRef" class="min-h-0 flex-1 space-y-4 overflow-y-auto bg-gray-50 p-4 dark:bg-dark-950">
          <div
            v-for="message in messages"
            :key="message.id"
            class="flex"
            :class="message.sender_type === 'user' ? 'justify-end' : 'justify-start'"
          >
            <div :class="messageBubbleClass(message.sender_type)">
              <div class="mb-1 flex items-center justify-between gap-4 text-xs opacity-75">
                <span>{{ senderLabel(message.sender_type) }}</span>
                <span>{{ formatDateTime(message.created_at) }}</span>
              </div>
              <p class="whitespace-pre-wrap break-words text-sm leading-6">{{ message.content }}</p>
              <button
                v-if="systemActionForMessage(message)"
                type="button"
                class="system-action-link mt-3"
                @click="openSystemAction(systemActionForMessage(message)!.path)"
              >
                <Icon :name="systemActionForMessage(message)!.icon" size="sm" />
                {{ systemActionForMessage(message)!.label }}
                <Icon name="arrowRight" size="xs" />
              </button>
            </div>
          </div>
          <div v-if="!detailLoading && messages.length === 0" class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('tickets.noMessages') }}
          </div>
        </div>

        <form v-if="selectedTicket && !selectedIsSystem" class="border-t border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900" @submit.prevent="handleSendMessage">
          <div class="flex gap-3">
            <textarea
              v-model.trim="replyContent"
              class="input min-h-[76px] flex-1 resize-none"
              :placeholder="selectedTicket.status === 'closed' ? t('tickets.closed') : t('tickets.messagePlaceholder')"
              :disabled="selectedTicket.status === 'closed'"
            ></textarea>
            <button class="btn btn-primary self-end" :disabled="sending || selectedTicket.status === 'closed'">
              <Icon name="mail" size="sm" />
              {{ sending ? t('common.saving') : t('tickets.send') }}
            </button>
          </div>
        </form>

        <div v-else-if="selectedTicket" class="border-t border-gray-200 bg-white p-4 text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400">
          {{ t('tickets.systemReadOnly') }}
        </div>

        <div v-else class="flex flex-1 flex-col items-center justify-center bg-gray-50 p-6 text-center text-gray-500 dark:bg-dark-950 dark:text-dark-400">
          <Icon name="ticket" size="xl" class="mb-3 text-gray-300 dark:text-dark-600" />
          <p>{{ t('tickets.noTicketSelected') }}</p>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import { ticketsAPI } from '@/api/tickets'
import { useAppStore } from '@/stores/app'
import { formatDateTime, formatRelativeTime } from '@/utils/format'
import type { SupportTicket, SupportTicketActionType, SupportTicketMessage, SupportTicketStatus } from '@/types'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const tickets = ref<SupportTicket[]>([])
const selectedTicketId = ref<number | null>(null)
const selectedTicket = ref<SupportTicket | null>(null)
const messages = ref<SupportTicketMessage[]>([])
const loading = ref(false)
const detailLoading = ref(false)
const creating = ref(false)
const sending = ref(false)
const acting = ref(false)
const showCreateForm = ref(false)
const showMobileList = ref(true)
const replyContent = ref('')
const messageListRef = ref<HTMLElement | null>(null)
let searchTimer: ReturnType<typeof setTimeout> | null = null
const TICKET_UNREAD_BADGE_REFRESH_EVENT = 'sub2api:ticket-unread-updated'

const filters = reactive({
  search: '',
  status: '' as SupportTicketStatus | '',
  unread_only: false,
})

const createForm = reactive({
  title: '',
  content: '',
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
  search: filters.search.trim() || undefined,
  unread_only: filters.unread_only,
}))

const selectedIsSystem = computed(() => selectedTicket.value?.ticket_type === 'system')

type SystemActionEntry = {
  path: string
  label: string
  icon: InstanceType<typeof Icon>['$props']['name']
}

function isSystemTicket(ticket: SupportTicket) {
  return ticket.ticket_type === 'system'
}

function statusLabel(ticket: SupportTicket) {
  if (isSystemTicket(ticket)) return t('tickets.systemTicket')
  return t(`tickets.${ticket.status}`)
}

function statusClass(status: SupportTicketStatus, ticketType: SupportTicket['ticket_type'] = 'support') {
  return [
    'badge',
    ticketType === 'system' ? 'badge-warning' : status === 'closed' ? 'badge-gray' : status === 'pending_admin' ? 'badge-warning' : 'badge-success',
  ]
}

function senderLabel(senderType: SupportTicketMessage['sender_type']) {
  if (senderType === 'admin') return t('tickets.adminSender')
  if (senderType === 'system') return t('tickets.systemMessage')
  return t('tickets.userSender')
}

function messageBubbleClass(senderType: SupportTicketMessage['sender_type']) {
  return [
    'max-w-[min(720px,85%)] rounded-lg px-4 py-3 shadow-sm',
    senderType === 'user'
      ? 'bg-primary-600 text-white'
      : senderType === 'admin'
        ? 'bg-white text-gray-900 dark:bg-dark-800 dark:text-gray-100'
        : 'bg-amber-50 text-amber-900 dark:bg-amber-950 dark:text-amber-100',
  ]
}

function resolveSystemActionType(message: SupportTicketMessage): SupportTicketActionType | null {
  const rawActionType = message.metadata?.action_type
  const actionType = typeof rawActionType === 'string' ? rawActionType : message.event_type
  switch (actionType) {
    case 'payment_completed':
    case 'affiliate_first_api_reward':
    case 'welfare_first_api_unclaimed':
    case 'group_changed':
      return actionType
    default:
      return null
  }
}

function systemActionForMessage(message: SupportTicketMessage): SystemActionEntry | null {
  if (message.sender_type !== 'system') return null
  switch (resolveSystemActionType(message)) {
    case 'payment_completed':
      return { path: '/orders', label: t('tickets.actions.paymentCompleted'), icon: 'clipboard' }
    case 'affiliate_first_api_reward':
      return { path: '/affiliate', label: t('tickets.actions.affiliateReward'), icon: 'userPlus' }
    case 'welfare_first_api_unclaimed':
      return { path: '/welfare', label: t('tickets.actions.welfareReward'), icon: 'sparkles' }
    default:
      return null
  }
}

function openSystemAction(path: string) {
  void router.push(path)
}

function notifyTicketUnreadBadgeChanged() {
  window.dispatchEvent(new Event(TICKET_UNREAD_BADGE_REFRESH_EVENT))
}

function applyTicketReadState(ticketId: number) {
  tickets.value = tickets.value.map((ticket) => (
    ticket.id === ticketId ? { ...ticket, user_unread_count: 0 } : ticket
  ))
  if (selectedTicket.value?.id === ticketId) {
    selectedTicket.value = { ...selectedTicket.value, user_unread_count: 0 }
  }
}

async function markTicketRead(ticketId: number, silent = false) {
  try {
    await ticketsAPI.markRead(ticketId)
    applyTicketReadState(ticketId)
    notifyTicketUnreadBadgeChanged()
    if (!silent) appStore.showSuccess(t('tickets.readSuccess'))
  } catch (error: any) {
    if (!silent) {
      appStore.showError(error.message || t('tickets.readFailed'))
    }
  }
}

async function loadTickets() {
  loading.value = true
  try {
    const response = await ticketsAPI.list(currentListParams.value)
    tickets.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
    if (!selectedTicketId.value && response.items.length > 0) {
      await selectTicket(response.items[0].id, false)
    } else if (selectedTicketId.value && !response.items.some((item) => item.id === selectedTicketId.value)) {
      selectedTicketId.value = null
      selectedTicket.value = null
      messages.value = []
    } else if (selectedTicketId.value) {
      selectedTicket.value = response.items.find((item) => item.id === selectedTicketId.value) ?? selectedTicket.value
    }
  } catch (error: any) {
    appStore.showError(error.message || t('tickets.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function loadDetail(ticketId: number) {
  detailLoading.value = true
  try {
    const detail = await ticketsAPI.get(ticketId)
    selectedTicket.value = detail.ticket
    messages.value = detail.messages
    await nextTick()
    messageListRef.value?.scrollTo?.({ top: messageListRef.value.scrollHeight })
    if (detail.ticket.user_unread_count > 0) {
      await markTicketRead(ticketId, true)
    }
  } catch (error: any) {
    appStore.showError(error.message || t('tickets.detailLoadFailed'))
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

async function handleCreateTicket() {
  if (!createForm.title.trim()) {
    appStore.showError(t('tickets.titleRequired'))
    return
  }
  if (!createForm.content.trim()) {
    appStore.showError(t('tickets.contentRequired'))
    return
  }
  creating.value = true
  try {
    const ticket = await ticketsAPI.create({
      title: createForm.title.trim(),
      content: createForm.content.trim(),
    })
    appStore.showSuccess(t('tickets.created'))
    createForm.title = ''
    createForm.content = ''
    showCreateForm.value = false
    await loadTickets()
    await selectTicket(ticket.id)
  } catch (error: any) {
    appStore.showError(error.message || t('tickets.createFailed'))
  } finally {
    creating.value = false
  }
}

async function handleSendMessage() {
  if (selectedIsSystem.value) {
    appStore.showError(t('tickets.systemReadOnly'))
    return
  }
  if (!selectedTicketId.value || !replyContent.value.trim()) {
    appStore.showError(t('tickets.contentRequired'))
    return
  }
  sending.value = true
  try {
    await ticketsAPI.createMessage(selectedTicketId.value, { content: replyContent.value.trim() })
    replyContent.value = ''
    appStore.showSuccess(t('tickets.sent'))
    await loadDetail(selectedTicketId.value)
    await loadTickets()
  } catch (error: any) {
    appStore.showError(error.message || t('tickets.sendFailed'))
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
    await ticketsAPI.close(selectedTicketId.value)
    appStore.showSuccess(t('tickets.closedSuccess'))
    await loadDetail(selectedTicketId.value)
    await loadTickets()
  } catch (error: any) {
    appStore.showError(error.message || t('tickets.closeFailed'))
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
    width: 22rem;
    flex: 0 0 22rem;
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

.ticket-list-item:hover,
.ticket-list-item-active {
  background: rgb(249 250 251);
}

.dark .ticket-list-item {
  border-bottom-color: rgb(55 65 81);
}

.dark .ticket-list-item:hover,
.dark .ticket-list-item-active {
  background: rgb(31 41 55);
}

.unread-pill {
  display: inline-flex;
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

.system-action-link {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border-radius: 0.375rem;
  border: 1px solid rgb(251 191 36);
  background: rgb(255 251 235);
  padding: 0.375rem 0.625rem;
  color: rgb(146 64 14);
  font-size: 0.8125rem;
  font-weight: 600;
}

.dark .system-action-link {
  border-color: rgba(251, 191, 36, 0.35);
  background: rgba(251, 191, 36, 0.1);
  color: rgb(253 230 138);
}
</style>
