import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const copyToClipboard = vi.fn().mockResolvedValue(true)
const { showError, syncUpstreamModels } = vi.hoisted(() => ({
  showError: vi.fn(), syncUpstreamModels: vi.fn(),
}))
vi.mock('@/api/admin/accounts', () => ({ default: { syncUpstreamModels }, accountsAPI: { syncUpstreamModels } }))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: { message?: string }) => (
        key === 'common.copy' ? 'Copy' : params?.message ? `${key}: ${params.message}` : key
      )
    })
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess: vi.fn(),
    showInfo: vi.fn()
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard })
}))

import ModelWhitelistSelector from '../ModelWhitelistSelector.vue'

function mountSelector(platform: 'openai' | 'kimi' = 'openai') {
  return mount(ModelWhitelistSelector, {
    props: {
      modelValue: [],
      platform
    },
    global: {
      stubs: {
        ModelIcon: true
      }
    }
  })
}

function findModelRow(wrapper: ReturnType<typeof mountSelector>, modelId: string) {
  const row = wrapper
    .findAll('[data-testid="model-option"]')
    .find(candidate => candidate.text().includes(modelId))
  if (!row) throw new Error(`Model row not found: ${modelId}`)
  return row
}

describe('ModelWhitelistSelector', () => {
  beforeEach(() => {
    copyToClipboard.mockClear()
    showError.mockReset()
    syncUpstreamModels.mockReset()
  })

  it('copies a model ID without selecting it', async () => {
    const wrapper = mountSelector()
    await wrapper.get('div.cursor-pointer').trigger('click')
    const row = findModelRow(wrapper, 'gpt-5.6-sol')

    const copyButton = row.get('[data-testid="copy-model-id"]')
    expect(copyButton.attributes('aria-label')).toBe('Copy gpt-5.6-sol')
    await copyButton.trigger('click')
    await flushPromises()

    expect(copyToClipboard).toHaveBeenCalledWith('gpt-5.6-sol')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('keeps the existing model selection behavior', async () => {
    const wrapper = mountSelector()
    await wrapper.get('div.cursor-pointer').trigger('click')
    const row = findModelRow(wrapper, 'gpt-5.6-sol')

    await row.get('[data-testid="select-model"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([[['gpt-5.6-sol']]])
    expect(copyToClipboard).not.toHaveBeenCalled()
  })

  it('exposes Kimi coding models without cross-provider suggestions', async () => {
    const wrapper = mountSelector('kimi')
    await wrapper.get('div.cursor-pointer').trigger('click')

    expect(findModelRow(wrapper, 'kimi-for-coding').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="model-option"]').some(row => row.text().includes('gpt-5.6-sol'))).toBe(false)
  })

  it.each([
    [{ status: 502, message: 'upstream unavailable' }, 'upstream unavailable'],
    [new Error('network failure'), 'network failure'],
    [{ status: 502 }, 'admin.accounts.syncUpstreamModelsFailed'],
  ])('extracts sync errors with fallback: %j', async (error, message) => {
    syncUpstreamModels.mockRejectedValue(error)
    const wrapper = mountSelector()
    await wrapper.setProps({ accountId: 42 })
    const button = wrapper.findAll('button').find(button => button.text() === 'admin.accounts.syncUpstreamModels')!
    await button.trigger('click')
    await flushPromises()
    expect(syncUpstreamModels).toHaveBeenCalledWith(42)
    expect(showError).toHaveBeenCalledWith(`admin.accounts.syncUpstreamModelsError: ${message}`)
    expect(button.attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })
})
