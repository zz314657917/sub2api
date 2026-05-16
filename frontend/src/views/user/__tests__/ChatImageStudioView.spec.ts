import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ChatImageStudioView from '../ChatImageStudioView.vue'

const keysList = vi.hoisted(() => vi.fn())
const getAvailable = vi.hoisted(() => vi.fn())
const createChatCompletionStream = vi.hoisted(() => vi.fn())
const listChatModels = vi.hoisted(() => vi.fn())
const createImageTask = vi.hoisted(() => vi.fn())
const downloadImageFile = vi.hoisted(() => vi.fn())
const listImageTasks = vi.hoisted(() => vi.fn())
const getImageTask = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())
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

vi.mock('@/api/imageCreator', () => ({
  createImageTask,
  downloadImageFile,
  listImageTasks,
  getImageTask,
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
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
    key: 'sk-studio',
    name: 'Studio token',
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
      allow_image_generation: true,
    },
    ...overrides,
  }
}

function makeImage(overrides: Record<string, unknown> = {}) {
  return {
    id: 9,
    task_id: 123,
    user_id: 42,
    url: '/api/v1/user/image-creator/images/9/file',
    output_format: 'png',
    mime_type: 'image/png',
    byte_size: 7,
    sha256: 'hash',
    revised_prompt: 'preview prompt',
    expires_at: '2026-05-17T00:00:00Z',
    created_at: '2026-05-10T00:00:00Z',
    ...overrides,
  }
}

function mockAnchorDownload() {
  const clicks: HTMLAnchorElement[] = []
  const originalCreateElement = document.createElement.bind(document)
  vi.spyOn(document, 'createElement').mockImplementation((tagName: string, options?: ElementCreationOptions) => {
    const element = originalCreateElement(tagName, options)
    if (tagName.toLowerCase() === 'a') {
      vi.spyOn(element as HTMLAnchorElement, 'click').mockImplementation(() => {
        clicks.push(element as HTMLAnchorElement)
      })
    }
    return element
  })
  return clicks
}

function makeTask(overrides: Record<string, unknown> = {}) {
  return {
    id: 123,
    user_id: 42,
    api_key_id: 1,
    status: 'succeeded',
    model: 'gpt-image-2',
    prompt: 'draw image',
    size: '1024x1024',
    quality: 'auto',
    output_format: 'png',
    background: 'auto',
    count: 1,
    expires_at: '2026-05-17T00:00:00Z',
    created_at: '2026-05-10T00:00:00Z',
    updated_at: '2026-05-10T00:00:00Z',
    images: [],
    ...overrides,
  }
}

function mountView() {
  return mount(ChatImageStudioView, {
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
      },
    },
  })
}

describe('ChatImageStudioView', () => {
  beforeEach(() => {
    localStorage.clear()
    keysList.mockReset().mockResolvedValue({ items: [makeKey()] })
    getAvailable.mockReset().mockResolvedValue([])
    createChatCompletionStream.mockReset().mockImplementation(async ({ onDelta }) => {
      onDelta?.('你好')
      return { content: '你好' }
    })
    listChatModels.mockReset().mockResolvedValue([{ id: 'gpt-5.4-mini' }])
    createImageTask.mockReset().mockResolvedValue(makeTask({ id: 123, status: 'pending' }))
    listImageTasks.mockReset().mockResolvedValue({ tasks: [], images: [] })
    getImageTask.mockReset().mockResolvedValue(makeTask({
      id: 123,
      status: 'succeeded',
      images: [makeImage()],
    }))
    downloadImageFile.mockReset().mockResolvedValue(new Blob(['pngdata'], { type: 'image/png' }))
    showError.mockReset()
    showSuccess.mockReset()
    showWarning.mockReset()
    copyToClipboard.mockReset().mockResolvedValue(true)
    Object.defineProperty(URL, 'createObjectURL', {
      configurable: true,
      writable: true,
      value: vi.fn(() => 'blob:image-preview'),
    })
    Object.defineProperty(URL, 'revokeObjectURL', {
      configurable: true,
      writable: true,
      value: vi.fn(),
    })
  })

  afterEach(() => {
    vi.useRealTimers()
    document.body.innerHTML = ''
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('keeps the API key picker as a rail dropdown and labels keys with API prefix', async () => {
    const wrapper = mountView()

    await flushPromises()

    expect(wrapper.find('.studio-rail-header').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('chatImageStudio.localOnly')
    expect(wrapper.findAll('[data-testid="studio-chat-model-select"]')).toHaveLength(0)

    const apiKeySelect = wrapper.find('[data-testid="studio-api-key-select"]')
    const newChatButton = wrapper.find('[data-testid="studio-new-chat-button"]')

    expect(apiKeySelect.exists()).toBe(true)
    expect(apiKeySelect.text()).toContain('API· Studio token')
    expect(newChatButton.exists()).toBe(true)
    expect(wrapper.find('.studio-rail').element.contains(apiKeySelect.element)).toBe(true)
    expect(wrapper.find('.studio-topbar').element.contains(apiKeySelect.element)).toBe(false)
    expect(wrapper.find('.studio-rail').element.contains(newChatButton.element)).toBe(true)
  })

  it('shows chat model controls only in chat mode', async () => {
    const wrapper = mountView()

    await flushPromises()

    expect(wrapper.find('[data-testid="studio-image-controls"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="studio-image-model-select"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="studio-model-market-link"]').exists()).toBe(false)
    expect(wrapper.find('.studio-mode-cluster').element.contains(wrapper.find('[data-testid="studio-image-model-select"]').element)).toBe(true)
    expect(wrapper.find('[data-testid="studio-reference-upload-button"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="studio-image-params-popover"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('mc · codex · openai')
    expect(wrapper.find('[data-testid="studio-chat-controls"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="studio-chat-model-select"]').exists()).toBe(false)
    expect(wrapper.find('.studio-refresh-action').exists()).toBe(false)

    await wrapper.find('[data-testid="studio-image-params-button"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="studio-image-params-popover"]').exists()).toBe(true)
    expect(wrapper.findAll('.studio-control-field')).toHaveLength(4)
    expect(wrapper.find('.studio-control-field-count input[type="number"]').exists()).toBe(false)
    expect(wrapper.findAll('.studio-control-field-count option').map((option) => option.text())).toEqual([
      '1',
      '2',
      '3',
      '4',
      '5',
      '6',
      '7',
      '8',
    ])
    expect(wrapper.text()).not.toContain('chatImageStudio.referenceField')

    await wrapper.find('button.studio-mode-button').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="studio-chat-controls"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="studio-chat-model-select"]').exists()).toBe(true)
    expect(wrapper.find('.studio-mode-cluster').element.contains(wrapper.find('[data-testid="studio-chat-model-select"]').element)).toBe(true)
    expect(wrapper.find('[data-testid="studio-image-controls"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="studio-reference-upload-button"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="studio-image-params-popover"]').exists()).toBe(false)
    expect(wrapper.find('.studio-refresh-action').exists()).toBe(false)
  })

  it('attaches a pasted image as the reference image and switches to image mode', async () => {
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('button.studio-mode-button').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="studio-chat-controls"]').exists()).toBe(true)

    const file = new File([new Uint8Array([1, 2, 3])], 'clipboard.webp', { type: 'image/webp' })
    const pasteEvent = new Event('paste', { bubbles: true, cancelable: true }) as ClipboardEvent
    Object.defineProperty(pasteEvent, 'clipboardData', {
      configurable: true,
      value: {
        files: [file],
        items: [],
      },
    })
    const prevented = vi.spyOn(pasteEvent, 'preventDefault')

    wrapper.find('[data-testid="studio-message-input"]').element.dispatchEvent(pasteEvent)
    await flushPromises()

    expect(prevented).toHaveBeenCalled()
    expect(wrapper.find('[data-testid="studio-image-controls"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="studio-reference-bubble"]').attributes('title')).toBe('clipboard.webp')
    expect(wrapper.find('[data-testid="studio-reference-bubble"] img').attributes('src')).toBe('blob:image-preview')
    expect(showSuccess).toHaveBeenCalledWith('chatImageStudio.referenceAttachedFromPaste')

    await wrapper.find('[data-testid="studio-reference-remove"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="studio-reference-bubble"]').exists()).toBe(false)
    expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:image-preview')
  })

  it('shows elapsed time while an image task is running', async () => {
    vi.useFakeTimers()
    getImageTask.mockReset().mockResolvedValue(makeTask({ id: 123, status: 'running', images: [] }))
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('[data-testid="studio-message-input"]').setValue('生成图片')
    await wrapper.find('[data-testid="studio-submit-button"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('chatImageStudio.elapsedTime:{"time":"00:00"}')

    await vi.advanceTimersByTimeAsync(65_000)
    await flushPromises()

    expect(wrapper.text()).toContain('chatImageStudio.elapsedTime:{"time":"01:05"}')
    expect(wrapper.text()).toContain('chatImageStudio.queueLaterHint')
    vi.useRealTimers()
  })

  it('opens the queue popup from the standalone queue button', async () => {
    listImageTasks.mockResolvedValueOnce({
      tasks: [makeTask({ id: 456, status: 'succeeded', prompt: 'queued image' })],
      images: [],
    })
    const wrapper = mountView()

    await flushPromises()

    const queueButton = wrapper.find('[data-testid="studio-queue-button"]')
    expect(queueButton.exists()).toBe(true)
    expect(queueButton.text()).toContain('chatImageStudio.queue')
    expect(queueButton.text()).toContain('1')
    expect(wrapper.find('[data-testid="studio-queue-overlay"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="studio-queue"]').exists()).toBe(false)

    await queueButton.trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="studio-queue-overlay"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="studio-queue"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('queued image')

    await wrapper.find('[data-testid="studio-queue-close"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="studio-queue-overlay"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="studio-message-input"]').exists()).toBe(true)
  })

  it('renders the gallery as image-only masonry cards with click preview', async () => {
    listImageTasks.mockResolvedValueOnce({
      tasks: [],
      images: [
        makeImage({
          id: 77,
          url: '/api/v1/user/image-creator/images/77/file',
          output_format: 'webp',
          revised_prompt: 'hidden gallery prompt',
        }),
      ],
    })
    const wrapper = mountView()

    await flushPromises()
    await wrapper.findAll('.studio-tab')[1].trigger('click')
    await flushPromises()

    const gallery = wrapper.find('[data-testid="studio-gallery"]')
    expect(gallery.exists()).toBe(true)
    expect(gallery.find('.studio-gallery-grid').exists()).toBe(true)
    expect(gallery.find('.studio-gallery-column').exists()).toBe(true)
    expect(gallery.find('.studio-gallery-preview img').exists()).toBe(true)
    expect(gallery.find('.studio-gallery-download').exists()).toBe(true)
    expect(gallery.find('.studio-gallery-meta').exists()).toBe(false)
    expect(gallery.text()).not.toContain('hidden gallery prompt')
    expect(gallery.text()).not.toContain('WEBP')
    expect(wrapper.find('.studio-composer').exists()).toBe(false)

    await gallery.find('.studio-gallery-preview').trigger('click')
    await flushPromises()

    expect(document.body.querySelector('[data-testid="studio-image-preview-overlay"]')).not.toBeNull()
  })

  it('sends a chat message from the unified studio', async () => {
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('[data-testid="studio-message-input"]').setValue('你好')
    await wrapper.find('button.studio-mode-button').trigger('click')
    await wrapper.find('[data-testid="studio-submit-button"]').trigger('click')
    await flushPromises()

    expect(createChatCompletionStream).toHaveBeenCalledWith(expect.objectContaining({
      apiKey: 'sk-studio',
      model: 'gpt-5.4-mini',
      messages: [{ role: 'user', content: '你好' }],
    }))
    expect(wrapper.text()).toContain('你好')
  })

  it('edits and resends chat messages from message actions', async () => {
    createChatCompletionStream.mockReset().mockImplementation(async ({ onDelta }) => {
      onDelta?.('reply')
      return { content: 'reply' }
    })
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('button.studio-mode-button').trigger('click')
    await wrapper.find('[data-testid="studio-message-input"]').setValue('first prompt')
    await wrapper.find('[data-testid="studio-submit-button"]').trigger('click')
    await flushPromises()

    expect(createChatCompletionStream).toHaveBeenCalledTimes(1)
    await wrapper.find('[data-testid="studio-edit-message"]').trigger('click')
    await flushPromises()

    const input = wrapper.find('[data-testid="studio-message-input"]')
    expect((input.element as HTMLTextAreaElement).value).toBe('first prompt')

    await input.setValue('edited prompt')
    await wrapper.find('[data-testid="studio-submit-button"]').trigger('click')
    await flushPromises()

    expect(createChatCompletionStream).toHaveBeenCalledTimes(2)
    expect(createChatCompletionStream).toHaveBeenLastCalledWith(expect.objectContaining({
      messages: [{ role: 'user', content: 'edited prompt' }],
    }))
    expect(wrapper.text()).toContain('edited prompt')
    expect(wrapper.findAll('[data-testid="studio-message-user"]').map((item) => item.text())).toEqual([
      expect.stringContaining('edited prompt'),
    ])

    await wrapper.find('[data-testid="studio-resend-assistant"]').trigger('click')
    await flushPromises()

    expect(createChatCompletionStream).toHaveBeenCalledTimes(3)
    expect(createChatCompletionStream).toHaveBeenLastCalledWith(expect.objectContaining({
      messages: [{ role: 'user', content: 'edited prompt' }],
    }))
  })

  it('creates an image task and renders the completed image in the conversation', async () => {
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('[data-testid="studio-message-input"]').setValue('生成一张雪王大战金刚')
    await wrapper.find('[data-testid="studio-submit-button"]').trigger('click')
    await flushPromises()

    expect(createImageTask).toHaveBeenCalledWith(expect.objectContaining({
      apiKeyId: 1,
      model: 'gpt-image-2',
      prompt: '生成一张雪王大战金刚',
      outputFormat: 'webp',
    }))
    expect(getImageTask).toHaveBeenCalledWith(123)
    expect(downloadImageFile).toHaveBeenCalledWith('/api/v1/user/image-creator/images/9/file')
    expect(wrapper.find('.studio-image-card img').attributes('src')).toBe('blob:image-preview')
  })

  it('restores an active image message and applies the final failed task state', async () => {
    localStorage.setItem('sub2api:chat-image-studio:v1', JSON.stringify({
      currentSessionId: 'studio_restore',
      sessions: [{
        id: 'studio_restore',
        title: 'Restore image',
        createdAt: '2026-05-10T00:00:00Z',
        updatedAt: '2026-05-10T00:00:00Z',
        messages: [
          {
            id: 'msg_restore',
            role: 'user',
            kind: 'text',
            content: 'draw a restored image',
            createdAt: '2026-05-10T00:00:00Z',
          },
          {
            id: 'img_restore',
            role: 'assistant',
            kind: 'image',
            content: 'chatImageStudio.generatingHint',
            createdAt: '2026-05-10T00:00:01Z',
            taskId: 321,
            status: 'running',
            images: [],
          },
        ],
      }],
    }))
    listImageTasks.mockResolvedValueOnce({
      tasks: [makeTask({ id: 321, status: 'running', prompt: 'draw a restored image', images: [] })],
      images: [],
    })
    getImageTask.mockResolvedValueOnce(makeTask({
      id: 321,
      status: 'failed',
      prompt: 'draw a restored image',
      error_message: 'image gateway returned HTTP 502',
      images: [],
    }))

    const wrapper = mountView()

    await flushPromises()
    await flushPromises()

    expect(getImageTask).toHaveBeenCalledWith(321)
    expect(wrapper.text()).toContain('chatImageStudio.status.failed')
    expect(wrapper.text()).toContain('image gateway returned HTTP 502')
  })

  it('does not render protected image file URLs directly when preview hydration fails', async () => {
    downloadImageFile.mockRejectedValueOnce(new Error('unauthorized'))
    listImageTasks.mockResolvedValueOnce({
      tasks: [],
      images: [makeImage({ id: 91, url: '/api/v1/user/image-creator/images/91/file' })],
    })
    const wrapper = mountView()

    await flushPromises()
    await wrapper.findAll('.studio-tab')[1].trigger('click')
    await flushPromises()

    const src = wrapper.find('.studio-gallery-preview img').attributes('src')
    expect(downloadImageFile).toHaveBeenCalledWith('/api/v1/user/image-creator/images/91/file')
    expect(src).toContain('data:image/gif;base64')
    expect(src).not.toContain('/api/v1/user/image-creator/images/91/file')
  })

  it('uses the count dropdown value when creating image tasks', async () => {
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('[data-testid="studio-image-params-button"]').trigger('click')
    await flushPromises()

    await wrapper.find('.studio-control-field-count select').setValue('8')
    await wrapper.find('[data-testid="studio-message-input"]').setValue('生成八张图片')
    await wrapper.find('[data-testid="studio-submit-button"]').trigger('click')
    await flushPromises()

    expect(createImageTask).toHaveBeenCalledWith(expect.objectContaining({
      count: 8,
    }))
  })

  it('selects generated images and downloads the selected batch', async () => {
    const clicks = mockAnchorDownload()
    getImageTask.mockResolvedValue(makeTask({
      id: 123,
      status: 'succeeded',
      images: [
        makeImage({ id: 9, url: '/api/v1/user/image-creator/images/9/file' }),
        makeImage({ id: 10, url: '/api/v1/user/image-creator/images/10/file' }),
      ],
    }))
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('[data-testid="studio-message-input"]').setValue('生成两张图')
    await wrapper.find('[data-testid="studio-submit-button"]').trigger('click')
    await flushPromises()

    const selectors = wrapper.findAll('[data-testid="studio-image-select"]')
    expect(selectors).toHaveLength(2)
    await selectors[0].trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('chatImageStudio.selectedImages:{"count":1}')
    await wrapper.find('[data-testid="studio-download-selected-result"]').trigger('click')
    await flushPromises()

    expect(downloadImageFile).toHaveBeenCalledWith('/api/v1/user/image-creator/images/9/file')
    expect(clicks).toHaveLength(1)
    expect(showSuccess).toHaveBeenCalledWith('chatImageStudio.preparingDownloads:{"count":1}')
  })

  it('previews image metadata with zoom and download controls', async () => {
    const clicks = mockAnchorDownload()
    getImageTask.mockResolvedValue(makeTask({
      id: 123,
      status: 'succeeded',
      images: [
        makeImage({ id: 9, byte_size: 2_076_508, url: '/api/v1/user/image-creator/images/9/file' }),
      ],
    }))
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('[data-testid="studio-message-input"]').setValue('鐢熸垚涓€寮犲浘')
    await wrapper.find('[data-testid="studio-submit-button"]').trigger('click')
    await flushPromises()

    await wrapper.find('.studio-image-preview').trigger('click')
    await flushPromises()

    const previewImg = document.body.querySelector('[data-testid="studio-image-preview-img"]') as HTMLImageElement
    expect(previewImg).not.toBeNull()
    Object.defineProperty(previewImg, 'naturalWidth', { configurable: true, value: 1254 })
    Object.defineProperty(previewImg, 'naturalHeight', { configurable: true, value: 1254 })
    previewImg.dispatchEvent(new Event('load'))
    await flushPromises()

    expect(document.body.querySelector('[data-testid="studio-image-preview-meta"]')?.textContent).toBe('1.98 MB · 1254 x 1254 · 1:1 · 1.57MP')
    expect(document.body.querySelector('[data-testid="studio-image-preview-counter"]')?.textContent).toBe('1 / 1')
    expect(document.body.querySelector('[data-testid="studio-image-preview-zoom"]')?.textContent).toBe('100%')

    ;(document.body.querySelector('[data-testid="studio-image-preview-zoom-in"]') as HTMLButtonElement).click()
    await flushPromises()
    expect(document.body.querySelector('[data-testid="studio-image-preview-zoom"]')?.textContent).toBe('125%')

    ;(document.body.querySelector('[data-testid="studio-image-preview-reset-zoom"]') as HTMLButtonElement).click()
    await flushPromises()
    expect(document.body.querySelector('[data-testid="studio-image-preview-zoom"]')?.textContent).toBe('100%')

    ;(document.body.querySelector('[data-testid="studio-image-preview-download"]') as HTMLButtonElement).click()
    await flushPromises()
    expect(clicks.length).toBeGreaterThan(0)
  })

  it('navigates between preview images with previous and next controls', async () => {
    getImageTask.mockResolvedValue(makeTask({
      id: 123,
      status: 'succeeded',
      images: [
        makeImage({ id: 9, url: '/api/v1/user/image-creator/images/9/file' }),
        makeImage({ id: 10, url: '/api/v1/user/image-creator/images/10/file', byte_size: 12 }),
      ],
    }))
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('[data-testid="studio-message-input"]').setValue('生成两张图')
    await wrapper.find('[data-testid="studio-submit-button"]').trigger('click')
    await flushPromises()

    await wrapper.find('.studio-image-preview').trigger('click')
    await flushPromises()

    const previousButton = document.body.querySelector('[data-testid="studio-image-preview-prev"]') as HTMLButtonElement
    const nextButton = document.body.querySelector('[data-testid="studio-image-preview-next"]') as HTMLButtonElement

    expect(document.body.querySelector('[data-testid="studio-image-preview-counter"]')?.textContent).toBe('1 / 2')
    expect(previousButton.disabled).toBe(true)
    expect(nextButton.disabled).toBe(false)

    nextButton.click()
    await flushPromises()

    expect(document.body.querySelector('[data-testid="studio-image-preview-counter"]')?.textContent).toBe('2 / 2')
    expect(previousButton.disabled).toBe(false)
    expect(nextButton.disabled).toBe(true)

    previousButton.click()
    await flushPromises()

    expect(document.body.querySelector('[data-testid="studio-image-preview-counter"]')?.textContent).toBe('1 / 2')
    expect(previousButton.disabled).toBe(true)
    expect(nextButton.disabled).toBe(false)
  })
})
