import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import MonitorTimeline from '../MonitorTimeline.vue'
import type { MonitorTimelinePoint } from '@/api/channelMonitor'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const messages: Record<string, string> = {
        'monitorCommon.history60pts': `近 ${params?.n} 次记录`,
        'monitorCommon.nextUpdateIn': `${params?.n}s 后刷新`,
        'monitorCommon.past': 'PAST',
        'monitorCommon.now': 'NOW',
        'monitorCommon.latencyEmpty': '-',
        'monitorCommon.status.operational': '正常',
        'monitorCommon.status.failed': '失败',
        'monitorCommon.relativeSecondsAgo': `${params?.n} 秒前`,
      }
      return messages[key] ?? key
    },
  }),
}))

const timeline: MonitorTimelinePoint[] = [
  {
    status: 'failed',
    latency_ms: 4317,
    ping_latency_ms: 32,
    checked_at: new Date().toISOString(),
  },
  {
    status: 'operational',
    latency_ms: 2200,
    ping_latency_ms: 30,
    checked_at: new Date().toISOString(),
  },
  {
    status: 'degraded',
    latency_ms: 2800,
    ping_latency_ms: 31,
    checked_at: new Date().toISOString(),
  },
]

describe('MonitorTimeline', () => {
  it('keeps all 60 bars shrinkable inside the card content width', () => {
    const wrapper = mount(MonitorTimeline, {
      props: {
        buckets: timeline,
        countdownSeconds: 59,
      },
    })

    const root = wrapper.get('[data-testid="monitor-timeline"]')
    const bars = wrapper.findAll('[data-testid="monitor-timeline-bar"]')

    expect(root.classes()).toContain('min-w-0')
    expect(root.classes()).toContain('w-full')
    expect(bars).toHaveLength(60)
    expect(bars.every(bar => bar.classes().includes('min-w-0'))).toBe(true)
    expect(bars.every(bar => !bar.classes().includes('min-w-[3px]'))).toBe(true)
  })

  it('preserves newest-last ordering and status encoding', () => {
    const wrapper = mount(MonitorTimeline, {
      props: {
        buckets: timeline,
        countdownSeconds: 59,
      },
    })

    const bars = wrapper.findAll('[data-testid="monitor-timeline-bar"]')
    const realBars = bars.slice(-3)

    expect(realBars[0].classes()).toContain('bg-orange-500')
    expect(realBars[0].attributes('style')).toContain('height: 65%')
    expect(realBars[1].classes()).toContain('bg-emerald-500')
    expect(realBars[1].attributes('style')).toContain('height: 100%')
    expect(realBars[2].classes()).toContain('bg-red-500')
    expect(realBars[2].attributes('style')).toContain('height: 35%')
    expect(realBars[2].attributes('title')).toContain('失败')
    expect(wrapper.text()).toContain('近 60 次记录')
    expect(wrapper.text()).toContain('59s 后刷新')
    expect(wrapper.text()).toContain('PAST')
    expect(wrapper.text()).toContain('NOW')
  })
})
