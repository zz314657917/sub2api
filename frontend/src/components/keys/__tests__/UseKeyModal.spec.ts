import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true)
  })
}))

import UseKeyModal from '../UseKeyModal.vue'

describe('UseKeyModal', () => {
  it('normalizes trailing slashes in Codex config.toml', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://ai.3zapi.top/',
        platform: 'openai'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const codeBlocks = wrapper.findAll('pre code')
    expect(codeBlocks[0].text()).toContain('base_url = "https://ai.3zapi.top"')
    expect(codeBlocks[0].text()).not.toContain('base_url = "https://ai.3zapi.top/"')
  })

  it('renders GPT-5.4 mini entry in OpenCode config', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'openai'
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.exists()).toBe(true)
    expect(codeBlock.text()).toContain('"name": "GPT-5.4 Mini"')
    expect(codeBlock.text()).not.toContain('"name": "GPT-5.4 Nano"')
  })

  it('renders OpenAI-compatible config for smart-routed keys without a group', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-smart',
        baseUrl: 'https://ai.3zapi.top',
        platform: null
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    expect(wrapper.text()).toContain('keys.useKeyModal.smartRoutingTitle')
    expect(wrapper.text()).not.toContain('keys.useKeyModal.noGroupTitle')

    const codeBlocks = wrapper.findAll('pre code')
    expect(codeBlocks.length).toBeGreaterThan(0)
    const codeText = codeBlocks.map((block) => block.text()).join('\n')
    expect(codeText).toContain('base_url = "https://ai.3zapi.top"')
    expect(codeText).toContain('sk-smart')
  })
})
