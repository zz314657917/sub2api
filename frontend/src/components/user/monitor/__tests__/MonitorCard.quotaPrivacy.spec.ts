import { mount, enableAutoUnmount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import MonitorCard from '../MonitorCard.vue'
import ProviderIcon from '../ProviderIcon.vue'
import { PROVIDERS } from '@/constants/channelMonitor'
import type { UserMonitorView } from '@/api/channelMonitor'

enableAutoUnmount(afterEach)
vi.mock('vue-i18n', async importOriginal => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key }),
}))

describe('monitor provider icons and public quota projection', () => {
  it.each(PROVIDERS)('renders a real SVG mark for %s', provider => {
    const wrapper = mount(ProviderIcon, { props: { provider } })
    expect(wrapper.find('svg').exists()).toBe(true)
    expect(wrapper.find('path').attributes('d')?.length).toBeGreaterThan(20)
    expect(wrapper.find('span').exists()).toBe(false)
  })

  it('localizes the quota model without showing admin snapshots or account identity', () => {
    const item = {
      id: 1, name: 'Public quota monitor', provider: 'antigravity',
      primary_model: 'quota', primary_status: 'operational', primary_latency_ms: null,
      primary_ping_latency_ms: null, group_name: '', availability_7d: 99,
      extra_models: [], timeline: [],
      latest_quota: { success: true, balance: 123.45, currency: 'USD', plan_level: 'private-plan' },
      account_id: 98765,
    } as UserMonitorView
    const wrapper = mount(MonitorCard, {
      props: { item, window: '7d', availabilityValue: 99, countdownSeconds: 10 },
      global: { stubs: { MonitorTimeline: true } },
    })
    expect(wrapper.text()).toContain('monitorCommon.checkMode.quota')
    expect(wrapper.text()).not.toContain('private-plan')
    expect(wrapper.text()).not.toContain('123.45')
    expect(wrapper.text()).not.toContain('98765')
    expect(wrapper.find('[data-testid="monitor-quota-view"]').exists()).toBe(false)
    expect(wrapper.find('monitor-quota-view-stub').exists()).toBe(false)
  })
})
