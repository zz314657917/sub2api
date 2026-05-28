import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ImageManagerView from '../ImageManagerView.vue'

const listManagedImages = vi.hoisted(() => vi.fn())
const deleteManagedImages = vi.hoisted(() => vi.fn())
const downloadImageFile = vi.hoisted(() => vi.fn())
const push = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())
const copyToClipboard = vi.hoisted(() => vi.fn())

vi.mock('vue-router', () => ({
  RouterLink: {
    props: ['to'],
    template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
  },
  useRouter: () => ({ push }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => params ? `${key}:${JSON.stringify(params)}` : key,
    }),
  }
})

vi.mock('@/api/imageCreator', () => ({
  deleteManagedImages,
  downloadImageFile,
  listManagedImages,
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

function makeImage(overrides: Record<string, unknown> = {}) {
  return {
    id: 9,
    task_id: 123,
    user_id: 42,
    url: '/api/v1/user/image-creator/images/9/file',
    output_format: 'png',
    mime_type: 'image/png',
    byte_size: 2048,
    sha256: 'hash',
    revised_prompt: 'revised prompt',
    task_prompt: 'draw a useful image',
    task_model: 'gpt-image-2',
    task_size: '1024x1024',
    task_quality: 'auto',
    expires_at: '2026-05-17T00:00:00Z',
    created_at: '2026-05-10T00:00:00Z',
    width: 1920,
    height: 1080,
    resolution: '1920x1080',
    aspect_ratio: '16:9',
    ...overrides,
  }
}

function mountView() {
  return mount(ImageManagerView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: { template: '<span />' },
      },
    },
  })
}

describe('ImageManagerView', () => {
  beforeEach(() => {
    listManagedImages.mockReset().mockResolvedValue({ items: [makeImage()], total: 1, limit: 40, offset: 0 })
    deleteManagedImages.mockReset().mockResolvedValue({ deleted: 1 })
    downloadImageFile.mockReset().mockResolvedValue(new Blob(['pngdata'], { type: 'image/png' }))
    push.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
    copyToClipboard.mockReset()
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

  it('loads managed images and hydrates authenticated image URLs', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(listManagedImages).toHaveBeenCalledWith({ limit: 40, offset: 0 })
    expect(downloadImageFile).toHaveBeenCalledWith('/api/v1/user/image-creator/images/9/file')
    expect(wrapper.find('[data-testid="image-manager-card"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('draw a useful image')
    expect(wrapper.find('img').attributes('src')).toBe('blob:image-preview')
  })

  it('copies prompt and reuses it in the embedded studio', async () => {
    const wrapper = mountView()
    await flushPromises()

    const actionButtons = wrapper.findAll('.image-manager-icon-button')
    await actionButtons[1].trigger('click')
    expect(copyToClipboard).toHaveBeenCalledWith('draw a useful image')

    await actionButtons[2].trigger('click')
    expect(push).toHaveBeenCalledWith({
      path: '/chat-images',
      query: {
        prompt: 'draw a useful image',
        mode: 'image',
      },
    })

    await actionButtons[3].trigger('click')
    expect(push).toHaveBeenCalledWith({
      path: '/chat-images',
      query: {
        mode: 'image',
        reference_image_id: '9',
        prompt: 'draw a useful image',
      },
    })
  })

  it('applies image library filters to the list request', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('input[type="search"]').setValue('useful')
    const selects = wrapper.findAll('select')
    await selects[0].setValue('webp')
    await selects[1].setValue('landscape')
    await selects[2].setValue('2k')
    await selects[3].setValue('16:9')
    await wrapper.find('[data-testid="image-manager-filters"] .btn-primary').trigger('click')
    await flushPromises()

    expect(listManagedImages).toHaveBeenLastCalledWith(expect.objectContaining({
      limit: 40,
      offset: 0,
      q: 'useful',
      format: 'webp',
      orientation: 'landscape',
      resolution: '2k',
      aspect_ratio: '16:9',
    }))
  })

  it('deletes selected images and updates the gallery', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('.image-manager-select').trigger('click')
    await wrapper.find('.image-manager-text-button.danger').trigger('click')
    await flushPromises()

    expect(deleteManagedImages).toHaveBeenCalledWith([9])
    expect(showSuccess).toHaveBeenCalledWith('imageManager.deleteSuccess:{"count":1}')
    expect(wrapper.find('[data-testid="image-manager-card"]').exists()).toBe(false)
  })
})
