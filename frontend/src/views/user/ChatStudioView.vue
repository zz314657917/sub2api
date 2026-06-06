<template>
  <AppLayout>
    <div
      class="chat-studio"
      :class="{ 'chat-studio-drawer-open': sessionsPanelOpen }"
      data-testid="chat-studio-view"
    >
      <button
        v-if="sessionsPanelOpen"
        type="button"
        class="chat-mobile-backdrop"
        :aria-label="t('chatStudio.closeSessions')"
        @click="sessionsPanelOpen = false"
      ></button>

      <section
        class="chat-sessions"
        :class="{ 'chat-sessions-mobile-open': sessionsPanelOpen }"
        data-testid="chat-sessions-panel"
      >
        <div class="chat-sessions-header">
          <div class="min-w-0">
            <h1 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('chatStudio.title') }}</h1>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-300">{{ t('chatStudio.localOnly') }}</p>
          </div>
          <div class="flex flex-shrink-0 items-center gap-2">
            <button
              type="button"
              class="btn btn-primary btn-sm"
              data-testid="chat-new-session"
              @click="startNewSession"
            >
              <Icon name="plus" size="sm" />
              <span>{{ t('chatStudio.newChat') }}</span>
            </button>
            <button
              type="button"
              class="btn btn-secondary btn-icon lg:hidden"
              :title="t('chatStudio.closeSessions')"
              @click="sessionsPanelOpen = false"
            >
              <Icon name="x" size="sm" />
            </button>
          </div>
        </div>

        <div class="chat-settings">
          <div>
            <div class="mb-1.5 flex items-center justify-between gap-2">
              <label class="input-label mb-0">{{ t('chatStudio.apiKey') }}</label>
              <button
                type="button"
                class="text-xs font-medium text-primary-600 hover:text-primary-700 disabled:cursor-not-allowed disabled:text-gray-400 dark:text-primary-400"
                :disabled="loadingKeys"
                @click="loadApiKeys"
              >
                {{ t('common.refresh') }}
              </button>
            </div>
            <div data-testid="chat-api-key-select">
              <Select
                v-model="selectedKeyId"
                :options="apiKeyOptions"
                :placeholder="loadingKeys ? t('chatStudio.loadingKeys') : t('chatStudio.selectKey')"
                :disabled="loadingKeys || apiKeyOptions.length === 0"
                searchable="auto"
              />
            </div>
            <p v-if="!loadingKeys && apiKeyOptions.length === 0" class="input-hint text-amber-600 dark:text-amber-300">
              {{ t('chatStudio.noUsableKey') }}
            </p>
          </div>

          <div>
            <div class="mb-1.5 flex items-center justify-between gap-2">
              <label class="input-label mb-0">{{ t('chatStudio.model') }}</label>
              <button
                type="button"
                class="text-xs font-medium text-primary-600 hover:text-primary-700 disabled:cursor-not-allowed disabled:text-gray-400 dark:text-primary-400"
                :disabled="loadingChannels"
                @click="loadModels"
              >
                {{ t('common.refresh') }}
              </button>
            </div>
            <Select
              v-model="model"
              :options="modelOptions"
              :placeholder="loadingChannels ? t('chatStudio.loadingModels') : t('chatStudio.selectModel')"
              searchable="auto"
              creatable
              :creatable-prefix="t('chatStudio.useCustomModel')"
              data-testid="chat-model-select"
            />
            <p class="input-hint">{{ t('chatStudio.modelHint') }}</p>
          </div>
        </div>

        <div class="chat-session-list">
          <div
            v-for="session in sessions"
            :key="session.id"
            class="chat-session-item"
            :class="{ 'chat-session-item-active': session.id === currentSessionId }"
            data-testid="chat-session-item"
          >
            <button
              type="button"
              class="chat-session-select"
              @click="selectSession(session.id)"
            >
              <Icon name="chatBubble" size="sm" class="mt-0.5 flex-shrink-0" />
              <span class="min-w-0 flex-1">
                <span class="block truncate text-sm font-medium">{{ session.title }}</span>
                <span class="mt-0.5 block truncate text-xs text-gray-500 dark:text-dark-300">
                  {{ sessionPreview(session) }}
                </span>
              </span>
              <span class="flex-shrink-0 text-[11px] text-gray-400 dark:text-dark-400">
                {{ messageCountLabel(session) }}
              </span>
            </button>
            <button
              type="button"
              class="chat-session-delete"
              :title="t('chatStudio.deleteChat')"
              :aria-label="t('chatStudio.deleteChat')"
              :disabled="sending"
              data-testid="chat-delete-session"
              @click="requestDeleteSession(session.id)"
            >
              <Icon name="trash" size="sm" />
            </button>
          </div>
        </div>
      </section>

      <section class="chat-main">
        <header class="chat-main-header">
          <button
            type="button"
            class="btn btn-secondary btn-icon lg:hidden"
            :title="t('chatStudio.sessions')"
            @click="sessionsPanelOpen = !sessionsPanelOpen"
          >
            <Icon name="menu" size="md" />
          </button>
          <div class="min-w-0 flex-1">
            <template v-if="renaming">
              <input
                ref="renameInputRef"
                v-model="renameDraft"
                type="text"
                class="input max-w-xl"
                data-testid="chat-rename-input"
                @keydown.enter.prevent="confirmRename"
                @keydown.esc.prevent="cancelRename"
                @blur="confirmRename"
              />
            </template>
            <template v-else>
              <h2 class="truncate text-lg font-semibold text-gray-900 dark:text-white">
                {{ currentSession?.title || t('chatStudio.defaultSessionTitle') }}
              </h2>
              <p class="chat-main-subtitle mt-1 truncate text-sm text-gray-500 dark:text-dark-300">
                {{ t('chatStudio.subtitle') }}
              </p>
            </template>
          </div>
          <div class="flex flex-shrink-0 items-center gap-2">
            <button
              type="button"
              class="btn btn-secondary btn-icon"
              :title="t('chatStudio.rename')"
              :disabled="!currentSession"
              @click="beginRename"
            >
              <Icon name="edit" size="sm" />
            </button>
            <button
              type="button"
              class="btn btn-secondary btn-icon"
              :title="t('chatStudio.clearChat')"
              :disabled="!currentSession || currentMessages.length === 0 || sending"
              data-testid="chat-clear-session"
              @click="clearCurrentSession"
            >
              <Icon name="x" size="sm" />
            </button>
            <button
              type="button"
              class="btn btn-secondary btn-icon text-red-600 hover:text-red-700 dark:text-red-300"
              :title="t('chatStudio.deleteChat')"
              :disabled="!currentSession || sending"
              data-testid="chat-delete-current-session"
              @click="deleteCurrentSession"
            >
              <Icon name="trash" size="sm" />
            </button>
          </div>
        </header>

        <div ref="messagesRef" class="chat-messages custom-scrollbar">
          <div v-if="currentMessages.length === 0" class="chat-empty">
            <div class="chat-empty-icon">
              <Icon name="chatBubble" size="xl" />
            </div>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('chatStudio.emptyTitle') }}</h3>
            <p class="mt-2 max-w-md text-sm leading-6 text-gray-500 dark:text-dark-300">
              {{ t('chatStudio.emptyDescription') }}
            </p>
          </div>

          <div v-else class="mx-auto flex w-full max-w-4xl flex-col gap-5 py-5">
            <article
              v-for="message in currentMessages"
              :key="message.id"
              class="chat-message"
              :class="message.role === 'user' ? 'chat-message-user' : 'chat-message-assistant'"
              :data-testid="message.role === 'assistant' ? 'chat-message-assistant' : 'chat-message-user'"
            >
              <template v-if="message.role === 'assistant'">
                <div class="chat-avatar chat-avatar-assistant">
                  <Icon name="sparkles" size="sm" />
                </div>
                <div class="chat-assistant-body">
                  <div class="chat-message-toolbar">
                    <span>{{ t('chatStudio.assistant') }}</span>
                    <button
                      v-if="message.content"
                      type="button"
                      class="chat-copy-button"
                      :title="t('chatStudio.copyReply')"
                      data-testid="chat-copy-reply"
                      @click="copyReply(message.content)"
                    >
                      <Icon name="copy" size="xs" />
                      <span>{{ t('common.copy') }}</span>
                    </button>
                  </div>
                  <div class="chat-message-content whitespace-pre-wrap break-words">
                    <template v-if="message.content">{{ message.content }}</template>
                    <span v-else class="chat-typing">{{ t('chatStudio.thinking') }}</span>
                  </div>
                </div>
              </template>

              <template v-else>
                <div class="chat-user-bubble">
                  <div class="chat-message-content whitespace-pre-wrap break-words">
                    {{ message.content }}
                  </div>
                </div>
              </template>
            </article>
          </div>
        </div>

        <footer class="chat-composer">
          <div class="chat-composer-shell">
            <div class="chat-composer-box">
              <textarea
                v-model="prompt"
                rows="2"
                class="chat-input"
                :placeholder="t('chatStudio.inputPlaceholder')"
                :disabled="sending"
                data-testid="chat-message-input"
                @keydown.enter.exact.prevent="sendMessage"
              ></textarea>
              <div class="chat-composer-actions">
                <span class="min-w-0 flex-1 truncate text-xs text-gray-500 dark:text-dark-300">
                  {{ selectedKey ? selectedKeyLabel : t('chatStudio.noKeySelected') }}
                </span>
                <div class="flex items-center gap-2">
                  <button
                    v-if="sending"
                    type="button"
                    class="btn btn-secondary btn-sm"
                    data-testid="chat-stop-button"
                    @click="stopGenerating"
                  >
                    <Icon name="x" size="sm" />
                    <span>{{ t('chatStudio.stop') }}</span>
                  </button>
                  <button
                    v-else
                    type="button"
                    class="btn btn-primary btn-sm"
                    :disabled="!canSend"
                    :title="t('chatStudio.send')"
                    data-testid="chat-send-button"
                    @click="sendMessage"
                  >
                    <Icon name="arrowUp" size="sm" />
                    <span>{{ t('chatStudio.send') }}</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </footer>
      </section>
    </div>

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('chatStudio.deleteChat')"
      :message="t('chatStudio.deleteConfirmMessage', { title: deletingSession?.title || t('chatStudio.defaultSessionTitle') })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDeleteSession"
      @cancel="cancelDeleteSession"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI, userChannelsAPI } from '@/api'
import type { UserAvailableChannel } from '@/api/channels'
import {
  CHAT_STUDIO_DEFAULT_MODEL,
  CHAT_STUDIO_STORAGE_KEY,
  createChatCompletionStream,
  isAbortError,
  listChatModels,
  type ChatStudioMessage,
  type ChatStudioModel,
  type ChatStudioRole,
} from '@/api/chatStudio'
import { useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import type { ApiKey } from '@/types'
import { apiKeyChatGroups, apiKeySupportsChat, primaryAPIKeyGroupName } from '@/utils/apiKeyCapabilities'

interface LocalChatMessage {
  id: string
  role: ChatStudioRole
  content: string
  createdAt: string
}

interface LocalChatSession {
  id: string
  title: string
  messages: LocalChatMessage[]
  createdAt: string
  updatedAt: string
}

interface ChatStoragePayload {
  sessions?: LocalChatSession[]
  currentSessionId?: string | null
}

const MAX_SESSIONS = 20
const MAX_MESSAGES_PER_SESSION = 100
const SESSION_TITLE_LIMIT = 28

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const apiKeys = ref<ApiKey[]>([])
const channels = ref<UserAvailableChannel[]>([])
const keyModels = ref<ChatStudioModel[]>([])
const selectedKeyId = ref<number | null>(null)
const model = ref(CHAT_STUDIO_DEFAULT_MODEL)
const prompt = ref('')
const sessions = ref<LocalChatSession[]>([])
const currentSessionId = ref<string | null>(null)
const loadingKeys = ref(false)
const loadingChannels = ref(false)
const sending = ref(false)
const sessionsPanelOpen = ref(false)
const renaming = ref(false)
const renameDraft = ref('')
const showDeleteDialog = ref(false)
const deletingSessionId = ref<string | null>(null)
const messagesRef = ref<HTMLElement | null>(null)
const renameInputRef = ref<HTMLInputElement | null>(null)

let abortController: AbortController | null = null
let loadModelsRequestId = 0

const currentSession = computed(() =>
  sessions.value.find((session) => session.id === currentSessionId.value) ?? null
)

const currentMessages = computed(() => currentSession.value?.messages ?? [])

const deletingSession = computed(() =>
  sessions.value.find((session) => session.id === deletingSessionId.value) ?? null
)

const selectedKey = computed(() =>
  apiKeys.value.find((key) => key.id === selectedKeyId.value) ?? null
)

const selectedKeyLabel = computed(() => selectedKey.value ? apiKeyLabel(selectedKey.value) : '')

const apiKeyOptions = computed<SelectOption[]>(() =>
  apiKeys.value.map((key) => ({
    value: key.id,
    label: apiKeyLabel(key),
  }))
)

const modelOptions = computed<SelectOption[]>(() => {
  const groupIds = new Set(
    (selectedKey.value ? apiKeyChatGroups(selectedKey.value).map((group) => group.id) : [])
      .filter((id): id is number => typeof id === 'number' && id > 0)
  )
  const names = new Set<string>()

  for (const keyModel of keyModels.value) {
    if (keyModel.id) names.add(keyModel.id)
  }

  for (const channel of channels.value) {
    for (const section of channel.platforms || []) {
      if (groupIds.size > 0 && !section.groups?.some((group) => groupIds.has(group.id))) continue
      for (const supportedModel of section.supported_models || []) {
        if (supportedModel.name) names.add(supportedModel.name)
      }
    }
  }

  if (!names.has(model.value)) names.add(model.value || CHAT_STUDIO_DEFAULT_MODEL)
  if (!names.has(CHAT_STUDIO_DEFAULT_MODEL)) names.add(CHAT_STUDIO_DEFAULT_MODEL)

  return Array.from(names)
    .filter(Boolean)
    .sort((a, b) => a.localeCompare(b))
    .map((name) => ({ value: name, label: name }))
})

const canSend = computed(() =>
  !sending.value &&
  !!selectedKey.value?.key &&
  model.value.trim().length > 0 &&
  prompt.value.trim().length > 0
)

watch([sessions, currentSessionId], () => {
  persistSessions()
}, { deep: true })

watch(selectedKeyId, () => {
  keyModels.value = []
  void loadModels()
})

onMounted(() => {
  restoreSessions()
  void loadApiKeys()
})

onBeforeUnmount(() => {
  abortController?.abort()
  abortController = null
})

function createId(prefix: string): string {
  return `${prefix}_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`
}

function nowIso(): string {
  return new Date().toISOString()
}

function createEmptySession(title = t('chatStudio.defaultSessionTitle')): LocalChatSession {
  const now = nowIso()
  return {
    id: createId('chat'),
    title,
    messages: [],
    createdAt: now,
    updatedAt: now,
  }
}

function restoreSessions(): void {
  try {
    const raw = localStorage.getItem(CHAT_STUDIO_STORAGE_KEY)
    if (!raw) {
      const session = createEmptySession()
      sessions.value = [session]
      currentSessionId.value = session.id
      return
    }

    const payload = JSON.parse(raw) as ChatStoragePayload
    const restored = normalizeSessions(payload.sessions || [])
    if (restored.length === 0) {
      const session = createEmptySession()
      sessions.value = [session]
      currentSessionId.value = session.id
      return
    }

    sessions.value = restored
    currentSessionId.value =
      restored.find((session) => session.id === payload.currentSessionId)?.id ??
      restored[0]?.id ??
      null
  } catch {
    const session = createEmptySession()
    sessions.value = [session]
    currentSessionId.value = session.id
  }
}

function normalizeSessions(input: LocalChatSession[]): LocalChatSession[] {
  return input
    .filter((session) => session && typeof session.id === 'string')
    .map((session) => ({
      ...session,
      title: String(session.title || t('chatStudio.defaultSessionTitle')),
      messages: Array.isArray(session.messages)
        ? session.messages
          .filter((message) =>
            message &&
            ['system', 'user', 'assistant'].includes(message.role) &&
            typeof message.content === 'string'
          )
          .slice(-MAX_MESSAGES_PER_SESSION)
        : [],
      createdAt: session.createdAt || nowIso(),
      updatedAt: session.updatedAt || nowIso(),
    }))
    .sort((a, b) => Date.parse(b.updatedAt) - Date.parse(a.updatedAt))
    .slice(0, MAX_SESSIONS)
}

function trimSessions(): void {
  const normalized = normalizeSessions(sessions.value)
  const sameOrder =
    normalized.length === sessions.value.length &&
    normalized.every((session, index) => session.id === sessions.value[index]?.id)
  if (!sameOrder) {
    sessions.value = normalized
  }
  if (!sessions.value.some((session) => session.id === currentSessionId.value)) {
    currentSessionId.value = sessions.value[0]?.id ?? null
  }
}

function persistSessions(): void {
  localStorage.setItem(CHAT_STUDIO_STORAGE_KEY, JSON.stringify({
    sessions: normalizeSessions(sessions.value),
    currentSessionId: currentSessionId.value,
  }))
}

function touchSession(session: LocalChatSession): void {
  session.updatedAt = nowIso()
  if (session.messages.length > MAX_MESSAGES_PER_SESSION) {
    session.messages = session.messages.slice(-MAX_MESSAGES_PER_SESSION)
  }
}

function ensureCurrentSession(): LocalChatSession {
  const existing = currentSession.value
  if (existing) return existing

  const session = createEmptySession()
  sessions.value.unshift(session)
  currentSessionId.value = session.id
  return session
}

function startNewSession(): void {
  if (sending.value) stopGenerating()
  const session = createEmptySession()
  sessions.value.unshift(session)
  currentSessionId.value = session.id
  trimSessions()
  sessionsPanelOpen.value = false
  prompt.value = ''
}

function selectSession(id: string): void {
  currentSessionId.value = id
  sessionsPanelOpen.value = false
  void scrollToBottom()
}

function requestDeleteSession(id: string): void {
  if (sending.value || !sessions.value.some((session) => session.id === id)) return
  deletingSessionId.value = id
  showDeleteDialog.value = true
}

function cancelDeleteSession(): void {
  showDeleteDialog.value = false
  deletingSessionId.value = null
}

function confirmDeleteSession(): void {
  const id = deletingSessionId.value
  cancelDeleteSession()
  if (!id) return
  deleteSession(id)
}

function deleteSession(id: string): void {
  const index = sessions.value.findIndex((session) => session.id === id)
  if (index < 0) return
  if (sending.value) stopGenerating()

  sessions.value.splice(index, 1)
  if (sessions.value.length === 0) {
    const session = createEmptySession()
    sessions.value = [session]
    currentSessionId.value = session.id
    sessionsPanelOpen.value = false
    return
  }

  if (id === currentSessionId.value) {
    currentSessionId.value = sessions.value[Math.max(0, index - 1)]?.id ?? sessions.value[0]?.id ?? null
  }
  sessionsPanelOpen.value = false
}

function deleteCurrentSession(): void {
  if (!currentSession.value) return
  requestDeleteSession(currentSession.value.id)
}

function clearCurrentSession(): void {
  if (!currentSession.value || sending.value) return
  currentSession.value.messages = []
  touchSession(currentSession.value)
}

function beginRename(): void {
  if (!currentSession.value) return
  renameDraft.value = currentSession.value.title
  renaming.value = true
  nextTick(() => {
    renameInputRef.value?.focus()
    renameInputRef.value?.select()
  })
}

function confirmRename(): void {
  if (!renaming.value || !currentSession.value) return
  const title = renameDraft.value.trim()
  currentSession.value.title = title || t('chatStudio.defaultSessionTitle')
  touchSession(currentSession.value)
  renaming.value = false
}

function cancelRename(): void {
  renaming.value = false
  renameDraft.value = ''
}

function sessionPreview(session: LocalChatSession): string {
  const last = [...session.messages].reverse().find((message) => message.content.trim())
  return last?.content.trim() || t('chatStudio.emptySession')
}

function messageCountLabel(session: LocalChatSession): string {
  return t('chatStudio.messageCount', { count: session.messages.length })
}

function apiKeyLabel(key: ApiKey): string {
  return [
    key.name,
    primaryAPIKeyGroupName(key),
    key.group?.platform,
  ].filter(Boolean).join(' · ')
}

function isUsableChatKey(key: ApiKey): boolean {
  return !!key.key && apiKeySupportsChat(key)
}

function pickDefaultKey(keys: ApiKey[]): ApiKey | null {
  const current = keys.find((key) => key.id === selectedKeyId.value)
  return current ?? keys[0] ?? null
}

async function loadApiKeys(): Promise<void> {
  loadingKeys.value = true
  try {
    const response = await keysAPI.list(1, 100, {
      status: 'active',
      sort_by: 'created_at',
      sort_order: 'desc',
    })
    apiKeys.value = response.items.filter(isUsableChatKey)
    const previousSelectedKeyId = selectedKeyId.value
    selectedKeyId.value = pickDefaultKey(apiKeys.value)?.id ?? null
    if (selectedKeyId.value === previousSelectedKeyId) {
      void loadModels()
    }
  } catch {
    appStore.showError(t('chatStudio.loadKeysFailed'))
  } finally {
    loadingKeys.value = false
  }
}

async function loadModels(): Promise<void> {
  loadingChannels.value = true
  const key = selectedKey.value?.key || ''
  const requestId = ++loadModelsRequestId
  try {
    const [models, availableChannels] = await Promise.all([
      key ? listChatModels(key).catch((): ChatStudioModel[] => []) : Promise.resolve<ChatStudioModel[]>([]),
      userChannelsAPI.getAvailable().catch((): UserAvailableChannel[] => []),
    ])
    if (requestId !== loadModelsRequestId) return
    keyModels.value = models
    channels.value = availableChannels
    ensureModelForSelectedKey()
  } catch {
    if (requestId !== loadModelsRequestId) return
    keyModels.value = []
    channels.value = []
    model.value = model.value.trim() || CHAT_STUDIO_DEFAULT_MODEL
    appStore.showWarning(t('chatStudio.loadModelsFailed'))
  } finally {
    if (requestId === loadModelsRequestId) {
      loadingChannels.value = false
    }
  }
}

function ensureModelForSelectedKey(): void {
  if (model.value.trim()) return
  model.value = modelOptions.value[0]?.value ? String(modelOptions.value[0].value) : CHAT_STUDIO_DEFAULT_MODEL
}

function buildConversationMessages(session: LocalChatSession): ChatStudioMessage[] {
  return session.messages
    .filter((message) => message.content.trim().length > 0)
    .map((message) => ({
      role: message.role,
      content: message.content,
    }))
}

function updateTitleFromPrompt(session: LocalChatSession, text: string): void {
  if (session.title !== t('chatStudio.defaultSessionTitle') && session.title.trim()) return
  const normalized = text.replace(/\s+/g, ' ').trim()
  if (!normalized) return
  session.title = normalized.length > SESSION_TITLE_LIMIT
    ? `${normalized.slice(0, SESSION_TITLE_LIMIT)}...`
    : normalized
}

async function sendMessage(): Promise<void> {
  const text = prompt.value.trim()
  if (!text) {
    appStore.showError(t('chatStudio.emptyMessage'))
    return
  }
  if (!selectedKey.value?.key) {
    appStore.showError(t('chatStudio.noApiKey'))
    return
  }
  if (!model.value.trim()) {
    appStore.showError(t('chatStudio.noModel'))
    return
  }

  const session = ensureCurrentSession()
  prompt.value = ''
  updateTitleFromPrompt(session, text)

  const userMessage: LocalChatMessage = {
    id: createId('msg'),
    role: 'user',
    content: text,
    createdAt: nowIso(),
  }
  const assistantMessage: LocalChatMessage = {
    id: createId('msg'),
    role: 'assistant',
    content: '',
    createdAt: nowIso(),
  }
  session.messages.push(userMessage, assistantMessage)
  touchSession(session)
  void scrollToBottom()

  abortController = new AbortController()
  sending.value = true
  const requestMessages = buildConversationMessages(session)

  try {
    await createChatCompletionStream({
      apiKey: selectedKey.value.key,
      model: model.value.trim(),
      messages: requestMessages,
      signal: abortController.signal,
      onDelta: (delta) => {
        assistantMessage.content += delta
        touchSession(session)
        void scrollToBottom()
      },
    })
    if (!assistantMessage.content.trim()) {
      assistantMessage.content = t('chatStudio.emptyAssistantReply')
    }
  } catch (error: unknown) {
    if (isAbortError(error)) {
      if (!assistantMessage.content.trim()) {
        session.messages = session.messages.filter((message) => message.id !== assistantMessage.id)
      }
      return
    }

    const message = error instanceof Error ? error.message : t('chatStudio.requestFailed')
    const displayMessage = t('chatStudio.requestFailedWithMessage', { message })
    if (!assistantMessage.content.trim()) {
      assistantMessage.content = displayMessage
    }
    appStore.showError(displayMessage)
  } finally {
    touchSession(session)
    sending.value = false
    abortController = null
    void scrollToBottom()
  }
}

function stopGenerating(): void {
  abortController?.abort()
}

async function scrollToBottom(): Promise<void> {
  await nextTick()
  const el = messagesRef.value
  if (!el) return
  if (typeof el.scrollTo === 'function') {
    el.scrollTo({ top: el.scrollHeight })
  } else {
    el.scrollTop = el.scrollHeight
  }
}

async function copyReply(text: string): Promise<void> {
  await copyToClipboard(text, t('chatStudio.replyCopied'))
}
</script>

<style scoped>
.chat-studio {
  --chat-studio-height: calc(100vh - 6rem);
  --chat-studio-height: calc(100dvh - 6rem);
  display: grid;
  position: relative;
  height: var(--chat-studio-height);
  min-height: 32rem;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 1rem;
}

.chat-mobile-backdrop {
  display: none;
}

.chat-sessions,
.chat-main {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  border: 1px solid rgb(229 231 235);
  background: rgb(255 255 255);
}

.dark .chat-sessions,
.dark .chat-main {
  border-color: rgb(55 65 81);
  background: rgb(31 41 55 / 0.8);
}

.chat-sessions {
  display: flex;
  flex-direction: column;
  border-radius: 0.75rem;
}

.chat-sessions-header,
.chat-main-header {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-bottom: 1px solid rgb(243 244 246);
  padding: 1rem;
}

.dark .chat-sessions-header,
.dark .chat-main-header {
  border-color: rgb(55 65 81);
}

.chat-settings {
  flex-shrink: 0;
  display: grid;
  gap: 0.875rem;
  border-bottom: 1px solid rgb(243 244 246);
  padding: 1rem;
}

.dark .chat-settings {
  border-color: rgb(55 65 81);
}

.chat-session-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 0.75rem;
}

.chat-session-item {
  display: flex;
  width: 100%;
  align-items: flex-start;
  gap: 0.5rem;
  border-radius: 0.5rem;
  padding: 0.25rem;
  text-align: left;
  color: rgb(55 65 81);
  transition: background-color 0.15s ease, color 0.15s ease;
}

.chat-session-select {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.5rem;
  text-align: left;
}

.chat-session-delete {
  display: inline-flex;
  height: 2.25rem;
  width: 2.25rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  color: rgb(148 163 184);
  transition: background-color 0.15s ease, color 0.15s ease;
}

.chat-session-delete:hover:not(:disabled) {
  background: rgb(254 226 226);
  color: rgb(220 38 38);
}

.chat-session-delete:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.chat-session-item:hover,
.chat-session-item-active {
  background: rgb(239 246 255);
  color: rgb(29 78 216);
}

.dark .chat-session-item {
  color: rgb(209 213 219);
}

.dark .chat-session-delete {
  color: rgb(148 163 184);
}

.dark .chat-session-delete:hover:not(:disabled) {
  background: rgb(127 29 29 / 0.35);
  color: rgb(252 165 165);
}

.dark .chat-session-item:hover,
.dark .chat-session-item-active {
  background: rgb(30 64 175 / 0.22);
  color: rgb(191 219 254);
}

.chat-main {
  display: flex;
  flex-direction: column;
  border-radius: 0.75rem;
}

.chat-messages {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 0 1rem 0.5rem;
}

.chat-empty {
  display: flex;
  min-height: 420px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.chat-empty-icon {
  margin-bottom: 1rem;
  display: flex;
  height: 4rem;
  width: 4rem;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  background: rgb(239 246 255);
  color: rgb(37 99 235);
}

.dark .chat-empty-icon {
  background: rgb(30 64 175 / 0.22);
  color: rgb(147 197 253);
}

.chat-message {
  display: flex;
  width: 100%;
  gap: 0.875rem;
}

.chat-message-user {
  justify-content: flex-end;
}

.chat-message-assistant {
  align-items: flex-start;
  justify-content: flex-start;
}

.chat-avatar {
  display: flex;
  height: 2rem;
  width: 2rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
}

.chat-avatar-user {
  background: rgb(229 231 235);
  color: rgb(55 65 81);
}

.chat-avatar-assistant {
  background: rgb(219 234 254);
  color: rgb(37 99 235);
}

.dark .chat-avatar-user {
  background: rgb(55 65 81);
  color: rgb(229 231 235);
}

.dark .chat-avatar-assistant {
  background: rgb(30 64 175 / 0.35);
  color: rgb(147 197 253);
}

.chat-assistant-body {
  min-width: 0;
  max-width: min(760px, calc(100% - 3rem));
  flex: 1;
  padding-top: 0.125rem;
}

.chat-user-bubble {
  max-width: min(720px, 78%);
  border-radius: 1rem;
  border-top-right-radius: 0.375rem;
  background: rgb(239 246 255);
  padding: 0.75rem 1rem;
  color: rgb(30 41 59);
}

.dark .chat-user-bubble {
  background: rgb(30 64 175 / 0.35);
  color: rgb(239 246 255);
}

.chat-message-toolbar {
  margin-bottom: 0.375rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: rgb(100 116 139);
}

.dark .chat-message-toolbar {
  color: rgb(148 163 184);
}

.chat-copy-button {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border-radius: 0.5rem;
  padding: 0.25rem 0.5rem;
  color: rgb(100 116 139);
  transition: background-color 0.15s ease, color 0.15s ease;
}

.chat-copy-button:hover {
  background: rgb(241 245 249);
  color: rgb(30 41 59);
}

.dark .chat-copy-button {
  color: rgb(148 163 184);
}

.dark .chat-copy-button:hover {
  background: rgb(51 65 85 / 0.65);
  color: rgb(226 232 240);
}

.chat-message-content {
  font-size: 0.925rem;
  line-height: 1.75;
  color: rgb(17 24 39);
}

.dark .chat-message-content {
  color: rgb(243 244 246);
}

.chat-typing {
  color: rgb(107 114 128);
}

.dark .chat-typing {
  color: rgb(156 163 175);
}

.chat-composer {
  flex-shrink: 0;
  padding: 0.75rem 1rem 1rem;
}

.chat-composer-shell {
  margin: 0 auto;
  width: 100%;
  max-width: 56rem;
}

.chat-composer-box {
  border-radius: 1rem;
  border: 1px solid rgb(209 213 219);
  background: rgb(255 255 255);
  padding: 0.75rem 0.875rem;
  box-shadow: 0 12px 34px rgb(15 23 42 / 0.08);
}

.dark .chat-composer-box {
  border-color: rgb(75 85 99);
  background: rgb(17 24 39 / 0.55);
}

.chat-input {
  width: 100%;
  min-height: 58px;
  resize: vertical;
  border: 0;
  background: transparent;
  font-size: 0.925rem;
  line-height: 1.6;
  color: rgb(17 24 39);
  outline: none;
}

.dark .chat-input {
  color: rgb(243 244 246);
}

.chat-composer-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding-top: 0.5rem;
}

@media (min-width: 768px) {
  .chat-studio {
    --chat-studio-height: calc(100vh - 6.7rem);
    --chat-studio-height: calc(100dvh - 6.7rem);
  }
}

@media (min-width: 1024px) {
  .chat-studio {
    --chat-studio-height: calc(100vh - 7.2rem);
    --chat-studio-height: calc(100dvh - 7.2rem);
  }
}

@media (max-width: 1023px) {
  .chat-studio {
    grid-template-columns: 1fr;
  }

  .chat-mobile-backdrop {
    position: fixed;
    inset: 0;
    z-index: 58;
    display: block;
    background: rgb(15 23 42 / 0.42);
    backdrop-filter: blur(2px);
  }

  .chat-sessions {
    position: fixed;
    inset: 0 auto 0 0;
    z-index: 59;
    display: flex;
    width: min(88vw, 360px);
    max-width: 360px;
    transform: translateX(-105%);
    border-radius: 0;
    box-shadow: 0 24px 60px rgb(15 23 42 / 0.28);
    transition: transform 0.2s ease;
  }

  .chat-sessions-mobile-open {
    transform: translateX(0);
  }

  .chat-main-header {
    padding: 0.875rem;
  }

  .chat-main-subtitle {
    display: none;
  }

  .chat-messages {
    padding: 0 0.75rem;
  }

  .chat-message {
    gap: 0.75rem;
  }

  .chat-assistant-body {
    max-width: calc(100% - 2.75rem);
  }

  .chat-user-bubble {
    max-width: 88%;
    padding: 0.7rem 0.875rem;
  }

  .chat-composer {
    padding: 0.75rem;
  }

  .chat-composer-actions {
    align-items: flex-start;
    flex-direction: column;
  }

  .chat-composer-actions > div {
    align-self: flex-end;
  }
}
</style>
