import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ImageCreatorView from '../ImageCreatorView.vue'

const keysList = vi.hoisted(() => vi.fn())
const createImageTask = vi.hoisted(() => vi.fn())
const downloadImageFile = vi.hoisted(() => vi.fn())
const listImageTasks = vi.hoisted(() => vi.fn())
const getImageTask = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())

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
}))

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
  }),
}))

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

function mountView() {
  return mount(ImageCreatorView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Select: {
          props: ['modelValue', 'options', 'placeholder', 'disabled', 'searchable'],
          emits: ['update:modelValue'],
          template: `
            <select
              :value="modelValue"
              :disabled="disabled"
              @change="$emit('update:modelValue', $event.target.value)"
            >
              <option v-for="option in options" :key="String(option.value)" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          `,
        },
        EmptyState: { template: '<div />' },
        Icon: { template: '<span />' },
      },
    },
  })
}

describe('ImageCreatorView', () => {
  beforeEach(() => {
    vi.useRealTimers()
    keysList.mockReset().mockResolvedValue({
      items: [
        {
          id: 1,
          key: 'sk-test',
          name: 'Image token',
          status: 'active',
          group: {
            id: 10,
            name: 'OpenAI',
            platform: 'openai',
            status: 'active',
            routing_scope: 'image',
            allow_image_generation: true,
          },
        },
      ],
    })
    createImageTask.mockReset().mockResolvedValue(makeTask({ status: 'pending' }))
    downloadImageFile.mockReset().mockResolvedValue(new Blob(['pngdata'], { type: 'image/png' }))
    listImageTasks.mockReset().mockResolvedValue({ tasks: [], images: [] })
    getImageTask.mockReset().mockResolvedValue(makeTask({ status: 'running' }))
    showError.mockReset()
    showSuccess.mockReset()
    Object.defineProperty(URL, 'createObjectURL', {
      configurable: true,
      writable: true,
      value: vi.fn(),
    })
    Object.defineProperty(URL, 'revokeObjectURL', {
      configurable: true,
      writable: true,
      value: vi.fn(),
    })
    vi.spyOn(URL, 'createObjectURL').mockImplementation(() => 'blob:image-preview')
    vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined)
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('creates a server task with the selected api key id and enters waiting state', async () => {
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('textarea').setValue('draw a persisted image')
    await wrapper.find('button.btn-primary').trigger('click')
    await flushPromises()

    expect(createImageTask).toHaveBeenCalledTimes(1)
    const payload = createImageTask.mock.calls[0][0]
    expect(payload).toMatchObject({
      apiKeyId: 1,
      model: 'gpt-image-2',
      prompt: 'draw a persisted image',
      count: 1,
      outputFormat: 'webp',
    })
    expect(payload).not.toHaveProperty('apiKey')
    expect(wrapper.text()).toContain('imageCreator.generatingTitle')
  })

  it('clamps image count to the frontend maximum before creating a task', async () => {
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('textarea').setValue('draw many persisted images')
    await wrapper.find('input[type="number"]').setValue('100')
    await wrapper.find('button.btn-primary').trigger('click')
    await flushPromises()

    expect(createImageTask).toHaveBeenCalledTimes(1)
    expect(createImageTask.mock.calls[0][0]).toMatchObject({
      count: 8,
    })
  })

  it('normalizes transparent background when switching to gpt-image-1.5', async () => {
    const wrapper = mountView()

    await flushPromises()
    const selects = wrapper.findAll('select')
    await selects[5].setValue('transparent')
    await selects[1].setValue('gpt-image-1.5')
    await wrapper.find('textarea').setValue('draw without transparent background')
    await wrapper.find('button.btn-primary').trigger('click')
    await flushPromises()

    expect(createImageTask).toHaveBeenCalledTimes(1)
    expect(createImageTask.mock.calls[0][0]).toMatchObject({
      model: 'gpt-image-1.5',
      background: 'auto',
    })
  })

  it('restores a running task on page load and shows the completed image after polling', async () => {
    const image = makeImage()
    listImageTasks.mockResolvedValue({
      tasks: [makeTask({ id: 321, status: 'running' })],
      images: [],
    })
    getImageTask.mockResolvedValue(makeTask({ id: 321, status: 'succeeded', images: [image] }))

    const wrapper = mountView()
    await flushPromises()

    expect(getImageTask).toHaveBeenCalledWith(321)
    expect(wrapper.find('[data-testid="image-result-preview"]').exists()).toBe(true)
    expect(downloadImageFile).toHaveBeenCalledWith('/api/v1/user/image-creator/images/9/file')
    expect(wrapper.find('img').attributes('src')).toBe('blob:image-preview')
  })

  it('keeps elapsed waiting time when restoring an unfinished task after refresh', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-05-10T00:01:20Z'))
    listImageTasks.mockResolvedValue({
      tasks: [makeTask({
        id: 321,
        status: 'running',
        started_at: '2026-05-10T00:00:20Z',
      })],
      images: [],
    })
    getImageTask.mockResolvedValue(makeTask({ id: 321, status: 'running' }))

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('imageCreator.elapsedSeconds:{"seconds":60}')
  })

  it('opens and closes a large preview for a stored recent image', async () => {
    listImageTasks.mockResolvedValue({
      tasks: [],
      images: [makeImage()],
    })
    const wrapper = mountView()

    await flushPromises()

    const previewButton = wrapper.find('[data-testid="image-result-preview"]')
    expect(previewButton.exists()).toBe(true)

    await previewButton.trigger('dblclick')
    await wrapper.vm.$nextTick()

    const overlay = wrapper.find('[data-testid="image-preview-overlay"]')
    expect(overlay.exists()).toBe(true)
    expect(overlay.find('img').attributes('src')).toBe('blob:image-preview')

    await wrapper.find('[data-testid="image-preview-close"]').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[data-testid="image-preview-overlay"]').exists()).toBe(false)
  })
})
