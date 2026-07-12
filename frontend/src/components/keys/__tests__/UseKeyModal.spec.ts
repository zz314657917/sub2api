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

  it('generates OpenCode config for bare and explicit GPT-5.6 max variants', async () => {
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

    const parsed = JSON.parse(wrapper.find('pre code').text())
    const models = parsed.provider.openai.models
    const expectedNames = {
      'gpt-5.6': 'GPT-5.6 (Sol)',
      'gpt-5.6-sol': 'GPT-5.6 Sol',
      'gpt-5.6-terra': 'GPT-5.6 Terra',
      'gpt-5.6-luna': 'GPT-5.6 Luna'
    }
    for (const [model, name] of Object.entries(expectedNames)) {
      expect(models[model].name).toBe(name)
      expect(models[model].limit).toEqual({ context: 1050000, output: 128000 })
      expect(Object.keys(models[model].variants)).toEqual(['low', 'medium', 'high', 'xhigh', 'max'])
    }
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

  it('renders unified access examples when enabled', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-unified',
        baseUrl: 'https://ai.3zapi.top/v1',
        platform: 'openai',
        unifiedAccess: true
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

    expect(wrapper.text()).toContain('keys.useKeyModal.unifiedAccessTitle')
    const codeText = wrapper.findAll('pre code').map((block) => block.text()).join('\n')
    expect(codeText).toContain('/v1/chat/completions')
    expect(codeText).toContain('/v1/images/generations')
    expect(codeText).toContain('/v1/videos/generations')
    expect(codeText).toContain('doubao-seedance-2-0-fast-480p')
    expect(codeText).toContain('sk-unified')
  })

  it('marks video as unavailable for unified keys without video routing', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-unified',
        baseUrl: 'https://ai.3zapi.top/v1',
        platform: 'openai',
        unifiedAccess: true,
        unifiedCapabilities: {
          chat: true,
          image: true,
          video: false
        }
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

    expect(wrapper.text()).toContain('keys.useKeyModal.unifiedAccessVideoUnavailable')
    const codeText = wrapper.findAll('pre code').map((block) => block.text()).join('\n')
    expect(codeText).toContain('/v1/chat/completions')
    expect(codeText).toContain('/v1/images/generations')
    expect(codeText).not.toContain('/v1/videos/generations')
  })

  it.each([
    ['https://ai.3zapi.top', 'https://ai.3zapi.top/v1/chat/completions'],
    ['https://ai.3zapi.top/v1', 'https://ai.3zapi.top/v1/chat/completions'],
    ['https://ai.3zapi.top/v1/', 'https://ai.3zapi.top/v1/chat/completions'],
    ['', 'http://localhost:3000/v1/chat/completions']
  ])('normalizes unified access API base URL from %s', (baseUrl, expectedChatURL) => {
    const originalLocation = window.location
    Object.defineProperty(window, 'location', {
      value: { ...originalLocation, origin: 'http://localhost:3000' },
      configurable: true
    })

    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-unified',
        baseUrl,
        platform: 'openai',
        unifiedAccess: true
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

    const codeText = wrapper.findAll('pre code').map((block) => block.text()).join('\n')
    expect(codeText).toContain(`curl ${expectedChatURL}`)
    expect(codeText).not.toContain('/v1/v1/')

    Object.defineProperty(window, 'location', {
      value: originalLocation,
      configurable: true
    })
  })

  it('renders Claude Fable 5 OpenCode config with adaptive thinking', async () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'antigravity'
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

    const claudeConfig = wrapper.findAll('pre code')
      .map((code) => code.text())
      .find((content) => content.includes('"antigravity-claude"'))

    expect(claudeConfig).toBeDefined()
    const parsed = JSON.parse(claudeConfig!)
    const fable = parsed.provider['antigravity-claude'].models['claude-fable-5']

    expect(fable.name).toBe('Claude Fable 5')
    expect(fable.limit).toEqual({ context: 1048576, output: 128000 })
    expect(fable.options.thinking).toEqual({ type: 'adaptive' })
    expect(fable.options.thinking).not.toHaveProperty('budgetTokens')
  })
})
