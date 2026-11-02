import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import UsageProgressBar from '../UsageProgressBar.vue'

enableAutoUnmount(afterEach)

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

describe('UsageProgressBar', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-03-17T00:00:00Z'))
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  it('showNowWhenIdle=true 且利用率为 0 但有未来 resetsAt 时显示倒计时', () => {
    const wrapper = mount(UsageProgressBar, {
      props: {
        label: '5h',
        utilization: 0,
        resetsAt: '2026-03-17T02:30:00Z',
        showNowWhenIdle: true,
        color: 'indigo'
      }
    })

    expect(wrapper.text()).toContain('2h 30m')
    expect(wrapper.text()).not.toContain('usage.resetNow')
  })

  it('showNowWhenIdle=true 但利用率大于 0 时显示倒计时', () => {
    const wrapper = mount(UsageProgressBar, {
      props: {
        label: '7d',
        utilization: 12,
        resetsAt: '2026-03-17T02:30:00Z',
        showNowWhenIdle: true,
        color: 'emerald'
      }
    })

    expect(wrapper.text()).toContain('2h 30m')
    expect(wrapper.text()).not.toContain('usage.resetNow')
    expect(wrapper.text()).not.toContain('usage.resetPending')
  })

  it('showNowWhenIdle=false 时保持原有倒计时行为', () => {
    const wrapper = mount(UsageProgressBar, {
      props: {
        label: '1d',
        utilization: 0,
        resetsAt: '2026-03-17T02:30:00Z',
        showNowWhenIdle: false,
        color: 'indigo'
      }
    })

    expect(wrapper.text()).toContain('2h 30m')
    expect(wrapper.text()).not.toContain('usage.resetNow')
  })

  it('resetsAt 已过期且利用率大于 0 时显示「待刷新」', () => {
    const wrapper = mount(UsageProgressBar, {
      props: {
        label: '5h',
        utilization: 53,
        // 早于 fake system time 2026-03-17T00:00:00Z
        resetsAt: '2026-03-16T22:00:00Z',
        color: 'indigo'
      }
    })

    expect(wrapper.text()).toContain('usage.resetPending')
    expect(wrapper.text()).not.toContain('usage.resetNow')
  })

  it('resetsAt 已过期且利用率为 0 时仍显示「现在」', () => {
    const wrapper = mount(UsageProgressBar, {
      props: {
        label: '5h',
        utilization: 0,
        resetsAt: '2026-03-16T22:00:00Z',
        color: 'indigo'
      }
    })

    expect(wrapper.text()).toContain('usage.resetNow')
    expect(wrapper.text()).not.toContain('usage.resetPending')
  })

  it('按账号计费除以已使用百分比估算窗口总额度', () => {
    const wrapper = mount(UsageProgressBar, {
      props: {
        label: '5h',
        utilization: 73,
        color: 'indigo',
        windowStats: { requests: 1, tokens: 10, cost: 1237.77 }
      }
    })

    expect(wrapper.text()).toContain('usage.estimatedQuota $1695.58')
    expect(wrapper.get('[title="usage.estimatedQuotaDetail"]').exists()).toBe(true)
  })

  it('refreshes immediately on visible without running the minute interval', async () => {
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('visible')
    const wrapper = mount(UsageProgressBar, {
      props: { label: '5h', utilization: 0, resetsAt: '2026-03-17T02:30:00Z', color: 'indigo' },
    })
    vi.setSystemTime(new Date('2026-03-17T01:00:00Z'))
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('2h 30m')
    document.dispatchEvent(new Event('visibilitychange'))
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('1h 30m')
  })

  it('does not refresh while hidden', async () => {
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('hidden')
    const wrapper = mount(UsageProgressBar, {
      props: { label: '5h', utilization: 12, resetsAt: '2026-03-17T02:30:00Z', color: 'indigo' },
    })
    vi.setSystemTime(new Date('2026-03-17T01:00:00Z'))
    document.dispatchEvent(new Event('visibilitychange'))
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('2h 30m')
  })

  it('keeps no idle clock and starts/stops it when resetsAt changes', async () => {
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('visible')
    const wrapper = mount(UsageProgressBar, {
      props: { label: '5h', utilization: 0, showNowWhenIdle: true, color: 'indigo' },
    })
    expect(vi.getTimerCount()).toBe(0)
    const initialClock = wrapper.vm.$.setupState.now
    vi.setSystemTime(new Date('2026-03-17T01:00:00Z'))
    document.dispatchEvent(new Event('visibilitychange'))
    await wrapper.vm.$nextTick()
    expect(wrapper.vm.$.setupState.now).toBe(initialClock)
    expect(wrapper.text()).toContain('usage.resetNow')
    expect(vi.getTimerCount()).toBe(0)
    await wrapper.setProps({ resetsAt: '2026-03-17T02:30:00Z' })
    expect(wrapper.text()).toContain('1h 30m')
    expect(vi.getTimerCount()).toBe(1)
    await wrapper.setProps({ resetsAt: null })
    expect(vi.getTimerCount()).toBe(0)
    expect(wrapper.text()).toContain('usage.resetNow')
    document.dispatchEvent(new Event('visibilitychange'))
    expect(vi.getTimerCount()).toBe(0)
  })

  it('removes its exact listener and timer on unmount', () => {
    const added = vi.spyOn(document, 'addEventListener')
    const removed = vi.spyOn(document, 'removeEventListener')
    const wrapper = mount(UsageProgressBar, {
      props: { label: '5h', utilization: 10, resetsAt: '2026-03-17T02:30:00Z', color: 'indigo' },
    })
    const listener = added.mock.calls.find(([name]) => name === 'visibilitychange')![1]
    wrapper.unmount()
    expect(removed).toHaveBeenCalledWith('visibilitychange', listener)
    expect(vi.getTimerCount()).toBe(0)
  })
})
