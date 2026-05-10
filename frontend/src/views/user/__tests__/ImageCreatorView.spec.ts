import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ImageCreatorView from '../ImageCreatorView.vue'

const keysList = vi.hoisted(() => vi.fn())
const createImageTask = vi.hoisted(() => vi.fn())
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
          template: '<div />',
        },
        EmptyState: { template: '<div />' },
        Icon: { template: '<span />' },
      },
    },
  })
}

describe('ImageCreatorView', () => {
  beforeEach(() => {
    keysList.mockReset().mockResolvedValue({
      items: [
        {
          id: 1,
          key: 'sk-test',
          name: 'Image token',
          status: 'active',
          group: {
            name: 'OpenAI',
            platform: 'openai',
            allow_image_generation: true,
          },
        },
      ],
    })
    createImageTask.mockReset().mockResolvedValue(makeTask({ status: 'pending' }))
    listImageTasks.mockReset().mockResolvedValue({ tasks: [], images: [] })
    getImageTask.mockReset().mockResolvedValue(makeTask({ status: 'running' }))
    showError.mockReset()
    showSuccess.mockReset()
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
    })
    expect(payload).not.toHaveProperty('apiKey')
    expect(wrapper.text()).toContain('imageCreator.generatingTitle')
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
    expect(wrapper.find('img').attributes('src')).toBe('/api/v1/user/image-creator/images/9/file')
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
    expect(overlay.find('img').attributes('src')).toBe('/api/v1/user/image-creator/images/9/file')

    await wrapper.find('[data-testid="image-preview-close"]').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[data-testid="image-preview-overlay"]').exists()).toBe(false)
  })
})
