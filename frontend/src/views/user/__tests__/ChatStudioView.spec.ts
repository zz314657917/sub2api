import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import ChatStudioView from '../ChatStudioView.vue'
import { CHAT_STUDIO_STORAGE_KEY } from '@/api/chatStudio'

const chatStudioSource = readFileSync(resolve(process.cwd(), 'src/views/user/ChatStudioView.vue'), 'utf8')

const keysList = vi.hoisted(() => vi.fn())
const getAvailable = vi.hoisted(() => vi.fn())
const createChatCompletionStream = vi.hoisted(() => vi.fn())
const listChatModels = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showWarning = vi.hoisted(() => vi.fn())
const copyToClipboard = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => params ? `${key}:${JSON.stringify(params)}` : key,
    }),
  }
})

vi.mock('@/api', () => ({
  keysAPI: {
    list: keysList,
  },
  userChannelsAPI: {
    getAvailable,
  },
}))

vi.mock('@/api/chatStudio', async () => {
  const actual = await vi.importActual<typeof import('@/api/chatStudio')>('@/api/chatStudio')
  return {
    ...actual,
    createChatCompletionStream,
    listChatModels,
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showWarning,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

function makeKey(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    user_id: 1,
    key: 'sk-chat',
    name: 'Chat token',
    group_id: 10,
    status: 'active',
    ip_whitelist: [],
    ip_blacklist: [],
    last_used_at: null,
    quota: 0,
    quota_used: 0,
    expires_at: null,
    created_at: '2026-05-10T00:00:00Z',
    updated_at: '2026-05-10T00:00:00Z',
    rate_limit_5h: 0,
    rate_limit_1d: 0,
    rate_limit_7d: 0,
    usage_5h: 0,
    usage_1d: 0,
    usage_7d: 0,
    window_5h_start: null,
    window_1d_start: null,
    window_7d_start: null,
    group: {
      id: 10,
      name: 'OpenAI',
      platform: 'openai',
      status: 'active',
      routing_scope: 'inference',
      allow_image_generation: false,
    },
    ...overrides,
  }
}

function mountView() {
  return mount(ChatStudioView, {
    attachTo: document.body,
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Select: {
          props: ['modelValue', 'options', 'placeholder', 'disabled'],
          emits: ['update:modelValue'],
          methods: {
            parseSelectValue(value: string) {
              const numeric = Number(value)
              return Number.isNaN(numeric) ? value : numeric
            },
          },
          template: `
            <select
              :value="modelValue"
              :disabled="disabled"
              @change="$emit('update:modelValue', parseSelectValue($event.target.value))"
            >
              <option v-for="option in options" :key="String(option.value)" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          `,
        },
        Icon: { template: '<span />' },
        ConfirmDialog: {
          props: ['show', 'title', 'message', 'confirmText', 'cancelText', 'danger'],
          emits: ['confirm', 'cancel'],
          template: `
            <div v-if="show" data-testid="chat-delete-confirm-dialog">
              <p>{{ title }}</p>
              <p>{{ message }}</p>
              <button type="button" data-testid="chat-delete-confirm-cancel" @click="$emit('cancel')">
                {{ cancelText }}
              </button>
              <button type="button" data-testid="chat-delete-confirm-submit" @click="$emit('confirm')">
                {{ confirmText }}
              </button>
            </div>
          `,
        },
      },
    },
  })
}

describe('ChatStudioView', () => {
  beforeEach(() => {
    localStorage.clear()
    keysList.mockReset().mockResolvedValue({
      items: [makeKey()],
    })
    getAvailable.mockReset().mockResolvedValue([
      {
        name: 'OpenAI',
        description: '',
        platforms: [
          {
            platform: 'openai',
            groups: [{ id: 10, name: 'OpenAI', platform: 'openai' }],
            supported_models: [
              { name: 'gpt-5.4-mini', platform: 'openai', pricing: null, reference_pricing: null },
              { name: 'gpt-5.4', platform: 'openai', pricing: null, reference_pricing: null },
            ],
          },
        ],
      },
    ])
    createChatCompletionStream.mockReset().mockImplementation(async ({ onDelta }) => {
      onDelta?.('你')
      onDelta?.('好')
      return { content: '你好' }
    })
    listChatModels.mockReset().mockResolvedValue([
      { id: 'gpt-5.5', display_name: 'GPT-5.5' },
      { id: 'gpt-5.4', display_name: 'GPT-5.4' },
    ])
    showError.mockReset()
    showWarning.mockReset()
    copyToClipboard.mockReset().mockResolvedValue(true)
  })

  afterEach(() => {
    document.body.innerHTML = ''
    localStorage.clear()
  })

  it('keeps the chat workspace height constrained so the composer stays visible', () => {
    expect(chatStudioSource).toContain('height: var(--chat-studio-height)')
    expect(chatStudioSource).toMatch(/\.chat-messages\s*\{[\s\S]*flex:\s*1;[\s\S]*overflow-y:\s*auto/)
    expect(chatStudioSource).toMatch(/\.chat-composer\s*\{[\s\S]*flex-shrink:\s*0/)
  })

  it('loads keys, sends a message, appends streamed assistant text, and stores local history', async () => {
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('[data-testid="chat-message-input"]').setValue('你好')
    await wrapper.find('[data-testid="chat-send-button"]').trigger('click')
    await flushPromises()

    expect(createChatCompletionStream).toHaveBeenCalledTimes(1)
    expect(createChatCompletionStream.mock.calls[0][0]).toMatchObject({
      apiKey: 'sk-chat',
      model: 'gpt-5.4-mini',
      messages: [{ role: 'user', content: '你好' }],
    })
    expect(wrapper.text()).toContain('你好')

    const stored = JSON.parse(localStorage.getItem(CHAT_STUDIO_STORAGE_KEY) || '{}')
    expect(stored.sessions[0].messages).toEqual(expect.arrayContaining([
      expect.objectContaining({ role: 'user', content: '你好' }),
      expect.objectContaining({ role: 'assistant', content: '你好' }),
    ]))
  })

  it('loads model options from the selected API key even when available channels are empty', async () => {
    getAvailable.mockResolvedValue([])
    const wrapper = mountView()

    await flushPromises()

    expect(listChatModels).toHaveBeenCalledWith('sk-chat')
    const modelOptions = wrapper.findAll('[data-testid="chat-model-select"] option').map((option) => option.text())
    expect(modelOptions).toEqual(expect.arrayContaining([
      'gpt-5.5',
      'gpt-5.4',
      'gpt-5.4-mini',
    ]))
  })

  it('loads route-group channel models for a smart-routed API key without a default group', async () => {
    keysList.mockResolvedValue({
      items: [
        makeKey({
          group_id: null,
          group: undefined,
          multi_group_routes: [
            { group_id: 20, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true },
          ],
          route_groups: [
            {
              id: 20,
              name: 'Gemini route',
              platform: 'gemini',
              status: 'active',
              routing_scope: 'inference',
              allow_image_generation: false,
            },
          ],
        }),
      ],
    })
    getAvailable.mockResolvedValue([
      {
        name: 'Available',
        description: '',
        platforms: [
          {
            platform: 'openai',
            groups: [{ id: 10, name: 'OpenAI', platform: 'openai' }],
            supported_models: [
              { name: 'gpt-hidden', platform: 'openai', pricing: null, reference_pricing: null },
            ],
          },
          {
            platform: 'gemini',
            groups: [{ id: 20, name: 'Gemini route', platform: 'gemini' }],
            supported_models: [
              { name: 'gemini-3-pro', platform: 'gemini', pricing: null, reference_pricing: null },
            ],
          },
        ],
      },
    ])

    const wrapper = mountView()
    await flushPromises()

    const modelOptions = wrapper.findAll('[data-testid="chat-model-select"] option').map((option) => option.text())
    expect(modelOptions).toContain('gemini-3-pro')
    expect(modelOptions).not.toContain('gpt-hidden')
    expect(wrapper.text()).toContain('Gemini route')
  })

  it('restores local sessions from browser storage', async () => {
    localStorage.setItem(CHAT_STUDIO_STORAGE_KEY, JSON.stringify({
      currentSessionId: 'chat_old',
      sessions: [
        {
          id: 'chat_old',
          title: '历史会话',
          messages: [{ id: 'msg_1', role: 'assistant', content: '旧回复', createdAt: '2026-05-10T00:00:00Z' }],
          createdAt: '2026-05-10T00:00:00Z',
          updatedAt: '2026-05-10T00:00:00Z',
        },
      ],
    }))

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('历史会话')
    expect(wrapper.text()).toContain('旧回复')
  })

  it('deletes a local session from the session list and keeps another session selected', async () => {
    localStorage.setItem(CHAT_STUDIO_STORAGE_KEY, JSON.stringify({
      currentSessionId: 'chat_two',
      sessions: [
        {
          id: 'chat_one',
          title: 'Older chat',
          messages: [{ id: 'msg_1', role: 'assistant', content: 'kept', createdAt: '2026-05-10T00:00:00Z' }],
          createdAt: '2026-05-10T00:00:00Z',
          updatedAt: '2026-05-10T00:00:00Z',
        },
        {
          id: 'chat_two',
          title: 'Current chat',
          messages: [{ id: 'msg_2', role: 'user', content: 'remove me', createdAt: '2026-05-11T00:00:00Z' }],
          createdAt: '2026-05-11T00:00:00Z',
          updatedAt: '2026-05-11T00:00:00Z',
        },
      ],
    }))
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-testid="chat-delete-session"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="chat-delete-confirm-dialog"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('chatStudio.deleteConfirmMessage')
    expect(wrapper.text()).toContain('Current chat')

    await wrapper.find('[data-testid="chat-delete-confirm-cancel"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="chat-delete-confirm-dialog"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Current chat')
    expect(JSON.parse(localStorage.getItem(CHAT_STUDIO_STORAGE_KEY) || '{}').sessions).toHaveLength(2)

    await wrapper.find('[data-testid="chat-delete-session"]').trigger('click')
    await wrapper.find('[data-testid="chat-delete-confirm-submit"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('Current chat')
    expect(wrapper.text()).toContain('Older chat')

    const stored = JSON.parse(localStorage.getItem(CHAT_STUDIO_STORAGE_KEY) || '{}')
    expect(stored.currentSessionId).toBe('chat_one')
    expect(stored.sessions).toHaveLength(1)
    expect(stored.sessions[0]).toMatchObject({ id: 'chat_one', title: 'Older chat' })
  })

  it('does not send when no usable key is available', async () => {
    keysList.mockResolvedValue({ items: [] })
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('[data-testid="chat-message-input"]').setValue('hello')

    expect(wrapper.find('[data-testid="chat-send-button"]').attributes('disabled')).toBeDefined()
    expect(createChatCompletionStream).not.toHaveBeenCalled()
  })

  it('aborts an active stream when stopping generation', async () => {
    let capturedSignal: AbortSignal | undefined
    let resolveStream: ((value: unknown) => void) | undefined
    createChatCompletionStream.mockImplementation(({ signal, onDelta }) => {
      capturedSignal = signal
      onDelta?.('半')
      return new Promise((resolve) => {
        resolveStream = resolve
      })
    })
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('[data-testid="chat-message-input"]').setValue('stop test')
    await wrapper.find('[data-testid="chat-send-button"]').trigger('click')
    await wrapper.vm.$nextTick()

    await wrapper.find('[data-testid="chat-stop-button"]').trigger('click')

    expect(capturedSignal?.aborted).toBe(true)
    resolveStream?.({ content: '半' })
    await flushPromises()
  })
})
