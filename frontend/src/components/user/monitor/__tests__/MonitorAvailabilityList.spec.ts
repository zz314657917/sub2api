import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import MonitorAvailabilityList from '../MonitorAvailabilityList.vue'
import type { UserMonitorView } from '@/api/channelMonitor'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const messages: Record<string, string> = {
        'channelStatus.availabilityPanel.title': '服务可用性',
        'channelStatus.availabilityPanel.searchPlaceholder': '搜索服务名称',
        'channelStatus.availabilityPanel.groupSuffix': '稳定性监听',
        'channelStatus.availabilityPanel.availabilityLabel': '可用率',
        'channelStatus.availabilityPanel.modelCount': `${params?.count} 个模型`,
        'channelStatus.availabilityPanel.noResultsTitle': '没有匹配的服务',
        'channelStatus.availabilityPanel.noResultsDescription': '换个关键词试试。',
        'channelStatus.availabilityPanel.legend.abnormal': '异常',
        'channelStatus.availabilityPanel.legend.normal': '正常',
        'channelStatus.availabilityPanel.legend.highLatency': '高延迟',
        'channelStatus.availabilityPanel.legend.maintenance': '维护中',
        'channelStatus.empty.title': '暂无可显示的渠道',
        'channelStatus.empty.description': '管理员尚未配置可监控的渠道。',
        'monitorCommon.providers.gemini': 'Gemini',
      }
      return messages[key] ?? key
    },
  }),
}))

const items: UserMonitorView[] = [
  {
    id: 1,
    name: 'gemini image channel',
    provider: 'gemini',
    group_name: '绘图模型',
    primary_model: 'gemini-3.1-flash-image-preview',
    primary_status: 'operational',
    primary_latency_ms: 120,
    primary_ping_latency_ms: 40,
    availability_7d: 99.31,
    extra_models: [
      {
        model: 'gemini-3-pro-image-preview',
        status: 'degraded',
        latency_ms: 320,
        availability_7d: 97.66,
      },
      {
        model: 'gemini-2.5-flash-image',
        status: 'operational',
        latency_ms: 180,
      },
    ],
    timeline: [],
  },
  {
    id: 2,
    name: 'openai chat channel',
    provider: 'openai',
    group_name: '对话模型',
    primary_model: 'gpt-5-mini',
    primary_status: 'operational',
    primary_latency_ms: 210,
    primary_ping_latency_ms: 55,
    availability_7d: 99.9,
    extra_models: [],
    timeline: [],
  },
]

function mountList() {
  return mount(MonitorAvailabilityList, {
    props: {
      items,
      window: '7d',
      loading: false,
      detailCache: {
        1: {
          id: 1,
          name: 'gemini image channel',
          provider: 'gemini',
          group_name: '绘图模型',
          models: [
            {
              model: 'gemini-2.5-flash-image',
              latest_status: 'operational',
              latest_latency_ms: 180,
              availability_7d: 98.64,
              availability_15d: 99,
              availability_30d: 99.5,
              avg_latency_7d_ms: 180,
            },
          ],
        },
      },
    },
  })
}

describe('MonitorAvailabilityList', () => {
  it('renders grouped availability rows and emits row clicks', async () => {
    const wrapper = mountList()

    expect(wrapper.text()).toContain('服务可用性')
    expect(wrapper.text()).toContain('绘图模型 稳定性监听')
    expect(wrapper.text()).toContain('对话模型 稳定性监听')
    expect(wrapper.text()).toContain('3 个模型')
    expect(wrapper.text()).toContain('1 个模型')
    expect(wrapper.text()).toContain('gemini-3.1-flash-image-preview')
    expect(wrapper.text()).toContain('gemini-3-pro-image-preview')
    expect(wrapper.text()).toContain('gemini-2.5-flash-image')
    expect(wrapper.text()).toContain('gpt-5-mini')
    expect(wrapper.text()).toContain('99.31%')
    expect(wrapper.text()).toContain('97.66%')
    expect(wrapper.text()).toContain('98.64%')
    expect(wrapper.text()).toContain('99.90%')
    expect(wrapper.text()).toContain('异常')
    expect(wrapper.text()).toContain('正常')
    expect(wrapper.findAll('[data-testid="monitor-availability-group-card"]')).toHaveLength(2)
    expect(wrapper.findAll('[data-testid="monitor-availability-row"]')).toHaveLength(4)

    const progressBars = wrapper.findAll('[role="progressbar"]')
    expect(progressBars[0].attributes('aria-valuenow')).toBe('99.31')
    expect(progressBars[0].find('span').attributes('style')).toContain('width: 99.31%')

    await wrapper.find('[data-testid="monitor-availability-row"]').trigger('click')
    expect(wrapper.emitted('rowClick')?.[0]?.[0]).toEqual(items[0])
  })

  it('filters rows by model name', async () => {
    const wrapper = mountList()

    await wrapper.get('[data-testid="monitor-availability-search"]').setValue('2.5')

    expect(wrapper.findAll('[data-testid="monitor-availability-row"]')).toHaveLength(1)
    expect(wrapper.findAll('[data-testid="monitor-availability-group-card"]')).toHaveLength(1)
    expect(wrapper.text()).not.toContain('gemini-3.1-flash-image-preview')
    expect(wrapper.text()).not.toContain('gemini-3-pro-image-preview')
    expect(wrapper.text()).not.toContain('gpt-5-mini')
    expect(wrapper.text()).toContain('gemini-2.5-flash-image')
  })

  it('keeps long model lists inside a scroll area', () => {
    const wrapper = mount(MonitorAvailabilityList, {
      props: {
        items: [
          {
            ...items[0],
            extra_models: Array.from({ length: 12 }, (_, index) => ({
              model: `gemini-extra-${index + 1}`,
              status: 'operational',
              latency_ms: 160 + index,
              availability_7d: 99 - index * 0.1,
            })),
          },
        ],
        window: '7d',
        loading: false,
        detailCache: {},
      },
    })

    expect(wrapper.findAll('[data-testid="monitor-availability-row"]')).toHaveLength(13)
    expect(wrapper.get('[data-testid="monitor-availability-scroll"]').classes()).toContain(
      'monitor-availability-list__scroll',
    )
  })
})
