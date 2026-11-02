import { defineComponent } from 'vue'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { Channel } from '@/api/admin/channels'
import ChannelsView from '@/views/admin/ChannelsView.vue'

const { get, put, post, getAllGroups, getWebSearchConfig, showError, showSuccess } = vi.hoisted(() => ({
  get: vi.fn(),
  put: vi.fn(),
  post: vi.fn(),
  getAllGroups: vi.fn(),
  getWebSearchConfig: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/client', () => ({
  apiClient: { get, put, post, delete: vi.fn() }
}))

// Keep the real channels API adapter so assertions cover the actual PUT/POST body.
vi.mock('@/api/admin', async () => ({
  adminAPI: {
    channels: await vi.importActual('@/api/admin/channels'),
    groups: { getAll: getAllGroups },
    settings: { getWebSearchEmulationConfig: getWebSearchConfig },
    accounts: {
      list: vi.fn().mockResolvedValue({ items: [] }),
      getAvailableModels: vi.fn().mockResolvedValue([]),
      getById: vi.fn()
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess })
}))

vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key })
}))

const LayoutStub = defineComponent({ template: '<main><slot /></main>' })
const TableLayoutStub = defineComponent({
  template: '<section><slot name="filters" /><slot name="table" /><slot name="pagination" /></section>'
})
const DataTableStub = defineComponent({
  props: ['data'],
  template: '<div><div v-for="row in data" :key="row.id"><slot name="cell-actions" :row="row" /></div></div>'
})
const DialogStub = defineComponent({
  props: ['show'],
  template: '<section v-if="show" data-testid="channel-dialog"><slot /><slot name="footer" /></section>'
})

let source: Channel
const wrappers: VueWrapper[] = []

function sourceChannel(): Channel {
  return {
    id: 42,
    name: 'Preserve hidden search dimensions',
    description: 'existing channel',
    status: 'active',
    billing_model_source: 'channel_mapped',
    restrict_models: false,
    group_ids: [1, 2],
    model_pricing: [],
    model_mapping: {},
    features_config: {
      web_search_emulation: { anthropic: true, openai: true, future: { mode: 'keep' } },
      codex_image_generation_bridge: { openai: true },
      unrelated: { nested: ['keep', false] }
    },
    apply_pricing_to_account_stats: false,
    account_stats_pricing_rules: [],
    created_at: '2026-10-30T00:00:00Z',
    updated_at: '2026-10-30T00:00:00Z'
  }
}

async function mountView() {
  const wrapper = mount(ChannelsView, {
    global: {
      stubs: {
        AppLayout: LayoutStub,
        TablePageLayout: TableLayoutStub,
        DataTable: DataTableStub,
        BaseDialog: DialogStub,
        Pagination: true,
        ConfirmDialog: true,
        EmptyState: true,
        Select: true,
        Icon: true,
        PlatformIcon: true,
        PricingEntryCard: true
      }
    }
  })
  wrappers.push(wrapper)
  await flushPromises()
  return wrapper
}

async function clickText(wrapper: VueWrapper, text: string) {
  const button = wrapper.findAll('button').find(item => item.text() === text)
  expect(button, `button ${text}`).toBeDefined()
  await button!.trigger('click')
  await flushPromises()
}

async function setPlatform(wrapper: VueWrapper, platform: string, enabled: boolean) {
  const label = wrapper.findAll('label').find(item => item.text() === `admin.groups.platforms.${platform}`)
  expect(label, `platform checkbox ${platform}`).toBeDefined()
  await label!.get('input').setValue(enabled)
  await flushPromises()
}

function searchToggle(wrapper: VueWrapper) {
  const toggles = wrapper.findAll('button[role="switch"]').filter(item =>
    item.element.parentElement?.textContent?.includes('admin.channels.form.webSearchEmulation')
  )
  expect(toggles).toHaveLength(1) // The WebSearch UI must remain Anthropic-only.
  return toggles[0]!
}

async function save(wrapper: VueWrapper) {
  await wrapper.get('#channel-form').trigger('submit')
  await flushPromises()
  expect(showError).not.toHaveBeenCalled()
}

describe('ChannelsView upstream priority save payload', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    source = sourceChannel()
    get.mockImplementation(async (url: string, options?: { params?: { page_size?: number } }) => {
      if (url !== '/admin/channels') throw new Error(`Unexpected GET ${url}`)
      const items = options?.params?.page_size === 1000 ? [] : [source]
      return { data: { items, total: items.length } }
    })
    put.mockResolvedValue({ data: source })
    post.mockResolvedValue({ data: source })
    getAllGroups.mockResolvedValue([
      { id: 1, name: 'Anthropic group', platform: 'anthropic', rate_multiplier: 1, status: 'active' },
      { id: 2, name: 'OpenAI group', platform: 'openai', rate_multiplier: 1, status: 'active' }
    ])
    getWebSearchConfig.mockResolvedValue({ enabled: true, providers: [{}] })
  })

  afterEach(() => {
    for (const wrapper of wrappers.splice(0)) wrapper.unmount()
    vi.restoreAllMocks()
  })

  it.each([true, false])('keeps unknown/OpenAI dimensions and saves explicit Anthropic=%s without mutating source', async enabled => {
    const original = JSON.parse(JSON.stringify(source))
    const wrapper = await mountView()
    await clickText(wrapper, 'common.edit')
    const toggle = searchToggle(wrapper)
    expect(toggle.attributes('aria-checked')).toBe('true')
    if (!enabled) await toggle.trigger('click')
    await save(wrapper)
    expect(put).toHaveBeenCalledTimes(1)
    expect(put).toHaveBeenCalledWith('/admin/channels/42', expect.objectContaining({
      features_config: {
        ...original.features_config,
        web_search_emulation: { ...original.features_config.web_search_emulation, anthropic: enabled }
      },
      group_ids: [1, 2]
    }))
    expect(source).toEqual(original)
    expect(post).not.toHaveBeenCalled()
  })

  it('writes Anthropic false when its platform section is disabled and retains other dimensions/features', async () => {
    const original = JSON.parse(JSON.stringify(source))
    const wrapper = await mountView()
    await clickText(wrapper, 'common.edit')
    await setPlatform(wrapper, 'anthropic', false)
    await save(wrapper)
    expect(put).toHaveBeenCalledWith('/admin/channels/42', expect.objectContaining({
      group_ids: [2],
      features_config: {
        ...original.features_config,
        web_search_emulation: { ...original.features_config.web_search_emulation, anthropic: false }
      }
    }))
    expect(source).toEqual(original)
  })

  it('writes Anthropic false when the saved channel has no Anthropic section', async () => {
    source.group_ids = [2]
    const wrapper = await mountView()
    await clickText(wrapper, 'common.edit')
    await save(wrapper)
    expect(put.mock.calls[0]?.[1].features_config.web_search_emulation).toEqual({
      anthropic: false, openai: true, future: { mode: 'keep' }
    })
  })

  it('preserves search dimensions while the global setting hides the toggle', async () => {
    getWebSearchConfig.mockResolvedValue({ enabled: false, providers: [] })
    const wrapper = await mountView()
    await clickText(wrapper, 'common.edit')
    expect(wrapper.findAll('button[role="switch"]').filter(item =>
      item.element.parentElement?.textContent?.includes('admin.channels.form.webSearchEmulation')
    )).toHaveLength(0)
    await save(wrapper)
    expect(put.mock.calls[0]?.[1].features_config.web_search_emulation).toEqual(source.features_config?.web_search_emulation)
  })

  it.each([true, false])('creates a channel with explicit Anthropic=%s through the real POST adapter', async enabled => {
    const wrapper = await mountView()
    await clickText(wrapper, 'admin.channels.createChannel')
    await wrapper.get('input[required]').setValue('New channel')
    await setPlatform(wrapper, 'anthropic', true)
    const group = wrapper.findAll('label').find(item => item.text().includes('Anthropic group'))
    expect(group).toBeDefined()
    await group!.get('input').setValue(true)
    if (enabled) await searchToggle(wrapper).trigger('click')
    await save(wrapper)
    expect(post).toHaveBeenCalledTimes(1)
    expect(post).toHaveBeenCalledWith('/admin/channels', expect.objectContaining({
      name: 'New channel',
      group_ids: [1],
      features_config: { web_search_emulation: { anthropic: enabled } }
    }))
    expect(put).not.toHaveBeenCalled()
  })

  it('creates an empty channel with explicit Anthropic false and no inherited dimensions', async () => {
    const wrapper = await mountView()
    await clickText(wrapper, 'admin.channels.createChannel')
    await wrapper.get('input[required]').setValue('Empty channel')
    await save(wrapper)
    expect(post).toHaveBeenCalledWith('/admin/channels', expect.objectContaining({
      group_ids: [],
      features_config: { web_search_emulation: { anthropic: false } }
    }))
  })
})
