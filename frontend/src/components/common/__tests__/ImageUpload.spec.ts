import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import ImageUpload from '@/components/common/ImageUpload.vue'

describe('ImageUpload', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('accepts webp files even when the browser omits the MIME type', async () => {
    vi.stubGlobal('FileReader', class {
      onload: ((event: ProgressEvent<FileReader>) => void) | null = null
      onerror: (() => void) | null = null

      readAsDataURL() {
        this.onload?.({
          target: {
            result: 'data:application/octet-stream;base64,AQID',
          },
        } as ProgressEvent<FileReader>)
      }

      readAsText() {}
    })

    const wrapper = mount(ImageUpload, {
      props: {
        modelValue: '',
        mode: 'image',
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    const input = wrapper.find<HTMLInputElement>('input[type="file"]')
    const file = new File([new Uint8Array([1, 2, 3])], 'bot.webp', { type: '' })
    Object.defineProperty(input.element, 'files', {
      configurable: true,
      value: [file],
    })

    await input.trigger('change')
    await flushPromises()

    expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toMatch(/^data:image\/webp;base64,/)
  })
})
