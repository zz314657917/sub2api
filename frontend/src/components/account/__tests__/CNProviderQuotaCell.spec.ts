import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { Account } from '@/types'

const { queryQuota } = vi.hoisted(() => ({ queryQuota: vi.fn() }))

vi.mock('@/api/admin', () => ({
  adminAPI: { cnProviders: { queryQuota } }
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

function makeAccount(mode: string): Account {
  return {
    id: 42,
    name: 'Kimi Coding',
    platform: 'kimi',
    type: 'apikey',
    credentials: { account_mode: mode },
    extra: {
      kimi_5h_used_percent: 45,
      kimi_weekly_used_percent: 80,
      kimi_5h_reset_at: '2099-03-15T05:00:00Z',
      kimi_weekly_reset_at: '2099-03-20T00:00:00Z',
      kimi_usage_updated_at: new Date().toISOString()
    },
    proxy_id: null,
    concurrency: 1,
    priority: 1,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: false,
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
    session_window_status: null
  }
}

import CNProviderQuotaCell from '../CNProviderQuotaCell.vue'

describe('CNProviderQuotaCell', () => {
  beforeEach(() => queryQuota.mockReset())

  it('renders fresh Coding Plan snapshots without an automatic probe', async () => {
    const wrapper = mount(CNProviderQuotaCell, { props: { account: makeAccount('coding') } })
    await flushPromises()

    expect(queryQuota).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('45%')
    expect(wrapper.text()).toContain('80%')
  })

  it('retains the snapshot when a manual probe fails', async () => {
    queryQuota.mockResolvedValue({ success: false, error: 'quota unavailable' })
    const wrapper = mount(CNProviderQuotaCell, { props: { account: makeAccount('coding') } })

    await wrapper.get('[data-test="cn-provider-quota-probe"]').trigger('click')
    await flushPromises()

    expect(queryQuota).toHaveBeenCalledWith(42)
    expect(wrapper.text()).toContain('45%')
    expect(wrapper.text()).toContain('quota unavailable')
  })

  it('labels the refresh control as an explicit action', async () => {
    const wrapper = mount(CNProviderQuotaCell, { props: { account: makeAccount('coding') } })
    await flushPromises()

    expect(queryQuota).not.toHaveBeenCalled()
    expect(wrapper.get('[data-test="cn-provider-quota-probe"]').text()).toBe('admin.accounts.cnProviders.probe')
  })

  it('hides quota controls for pay-as-you-go accounts', () => {
    const wrapper = mount(CNProviderQuotaCell, { props: { account: makeAccount('payg') } })
    expect(wrapper.text()).toBe('')
  })
})
