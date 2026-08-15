import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import OpenAIQuotaResetCell from '../OpenAIQuotaResetCell.vue'
import type { Account } from '@/types'

const { refreshOpenAIQuota, resetOpenAIQuota } = vi.hoisted(() => ({
  refreshOpenAIQuota: vi.fn(),
  resetOpenAIQuota: vi.fn()
}))

vi.mock('@/api/admin/accounts', () => ({
  refreshOpenAIQuota,
  resetOpenAIQuota
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) =>
        params ? `${key}:${JSON.stringify(params)}` : key
    })
  }
})

function makeAccount(overrides: Partial<Account>): Account {
  return {
    id: 1,
    name: 'account',
    platform: 'openai',
    type: 'oauth',
    proxy_id: null,
    concurrency: 1,
    priority: 1,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: true,
    created_at: '2026-03-15T00:00:00Z',
    updated_at: '2026-03-15T00:00:00Z',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    ...overrides
  }
}

describe('OpenAIQuotaResetCell', () => {
  beforeEach(() => {
    refreshOpenAIQuota.mockReset()
    resetOpenAIQuota.mockReset()
  })

  it('只在 OpenAI OAuth 账号显示', () => {
    const wrapper = mount(OpenAIQuotaResetCell, {
      props: {
        account: makeAccount({ platform: 'anthropic', type: 'oauth' })
      }
    })

    expect(wrapper.text()).toBe('')
  })

  it('查询 reset credits 后显示次数并启用重置', async () => {
    refreshOpenAIQuota.mockResolvedValue({
      fetched_at: 1,
      rate_limit_reset_credits: { available_count: 2 }
    })

    const wrapper = mount(OpenAIQuotaResetCell, {
      props: {
        account: makeAccount({ id: 22 })
      }
    })

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(refreshOpenAIQuota).toHaveBeenCalledWith(22)
    expect(wrapper.findAll('button')[0].text()).toContain('admin.accounts.openaiQuotaReset.count')
    expect(wrapper.findAll('button')[0].text()).toContain('2')
    expect(wrapper.findAll('button')[1].attributes('disabled')).toBeUndefined()
  })

  it('重置成功后不会自动再次查询额度', async () => {
    refreshOpenAIQuota.mockResolvedValueOnce({
        fetched_at: 1,
        rate_limit_reset_credits: { available_count: 1 }
      })
    resetOpenAIQuota.mockResolvedValue({
      code: 'success',
      windows_reset: 2
    })

    const wrapper = mount(OpenAIQuotaResetCell, {
      props: {
        account: makeAccount({ id: 23 })
      }
    })

    await wrapper.findAll('button')[0].trigger('click')
    await flushPromises()
    await wrapper.findAll('button')[1].trigger('click')
    await flushPromises()
    expect(resetOpenAIQuota).not.toHaveBeenCalled()
    await wrapper.getComponent({ name: 'ConfirmDialog' }).vm.$emit('confirm')
    await flushPromises()

    expect(resetOpenAIQuota).toHaveBeenCalledWith(23)
    expect(refreshOpenAIQuota).toHaveBeenCalledTimes(1)
    expect(wrapper.findAll('button')[0].text()).toContain('admin.accounts.openaiQuotaReset.count')
    expect(wrapper.text()).toContain('admin.accounts.openaiQuotaReset.resetSuccess')
    expect(wrapper.text()).toContain('"windows":2')
  })

  it('查询失败时显示压缩后的错误', async () => {
    refreshOpenAIQuota.mockRejectedValue({
      message: 'x'.repeat(100)
    })

    const wrapper = mount(OpenAIQuotaResetCell, {
      props: {
        account: makeAccount({ id: 24 })
      }
    })

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('x'.repeat(80))
    expect(wrapper.text()).toContain('...')
  })

  it('账号行复用时会重置本地状态', async () => {
    refreshOpenAIQuota.mockResolvedValue({
      fetched_at: 1,
      rate_limit_reset_credits: { available_count: 3 }
    })

    const wrapper = mount(OpenAIQuotaResetCell, {
      props: {
        account: makeAccount({ id: 25 })
      }
    })

    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.findAll('button')[0].text()).toContain('admin.accounts.openaiQuotaReset.count')
    expect(wrapper.findAll('button')[0].text()).toContain('3')

    await wrapper.setProps({ account: makeAccount({ id: 26 }) })
    await flushPromises()

    expect(wrapper.findAll('button')[0].text()).toContain('admin.accounts.openaiQuotaReset.count')
    expect(wrapper.findAll('button')[0].text()).not.toContain('3')
  })
})
