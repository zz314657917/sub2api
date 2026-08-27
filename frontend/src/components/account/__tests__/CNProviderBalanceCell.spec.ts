import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { Account } from '@/types'

const { queryBalance } = vi.hoisted(() => ({ queryBalance: vi.fn() }))

vi.mock('@/api/admin', () => ({
  adminAPI: { cnProviders: { queryBalance } }
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

function makeAccount(mode: string): Account {
  return {
    id: 41,
    name: 'Kimi',
    platform: 'kimi',
    type: 'apikey',
    credentials: { account_mode: mode },
    extra: {
      kimi_balance: 12.3,
      kimi_balance_currency: 'CNY'
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

import CNProviderBalanceCell from '../CNProviderBalanceCell.vue'

describe('CNProviderBalanceCell', () => {
  beforeEach(() => queryBalance.mockReset())

  it('shows the last balance snapshot for a pay-as-you-go account', () => {
    const wrapper = mount(CNProviderBalanceCell, { props: { account: makeAccount('payg') } })
    expect(wrapper.get('[data-test="cn-provider-balance-value"]').text()).toContain('CNY 12.30')
    expect(wrapper.get('[data-test="cn-provider-balance-probe"]').text()).toBe('admin.accounts.cnProviders.probe')
  })

  it('uses the CN balance placeholder when no snapshot is available', () => {
    const account = makeAccount('payg')
    account.extra = {}

    const wrapper = mount(CNProviderBalanceCell, { props: { account } })

    expect(wrapper.text()).toContain('admin.accounts.cnProviders.balance')
    expect(wrapper.text()).not.toContain('admin.accounts.grokBalance')
  })

  it('retains the snapshot when a manual probe fails', async () => {
    queryBalance.mockResolvedValue({ success: false, error: 'provider unavailable' })
    const wrapper = mount(CNProviderBalanceCell, { props: { account: makeAccount('payg') } })

    await wrapper.get('[data-test="cn-provider-balance-probe"]').trigger('click')
    await flushPromises()

    expect(queryBalance).toHaveBeenCalledWith(41)
    expect(wrapper.text()).toContain('CNY 12.30')
    expect(wrapper.text()).toContain('provider unavailable')
  })

  it('hides balance controls for Coding Plan accounts', () => {
    const wrapper = mount(CNProviderBalanceCell, { props: { account: makeAccount('coding') } })
    expect(wrapper.text()).toBe('')
  })
})
