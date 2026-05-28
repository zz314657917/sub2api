import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import HeaderAnnouncementCarousel from '../HeaderAnnouncementCarousel.vue'
import type { UserAnnouncement } from '@/types'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => (key === 'announcements.title' ? '公告' : key),
  }),
}))

const announcements: UserAnnouncement[] = [
  {
    id: 1,
    title: '维护通知',
    content: '今晚 23:00 开始维护，预计 10 分钟。',
    notify_mode: 'silent',
    created_at: '2026-05-28T00:00:00Z',
    updated_at: '2026-05-28T00:00:00Z',
  },
  {
    id: 2,
    title: '新模型上线',
    content: '支持更多模型路由。',
    notify_mode: 'popup',
    read_at: '2026-05-28T01:00:00Z',
    created_at: '2026-05-28T01:00:00Z',
    updated_at: '2026-05-28T01:00:00Z',
  },
]

const globalStubs = {
  Icon: true,
  Transition: { template: '<slot />' },
}

describe('HeaderAnnouncementCarousel', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('does not render when there are no announcements', () => {
    const wrapper = mount(HeaderAnnouncementCarousel, {
      props: { announcements: [] },
      global: {
        stubs: globalStubs,
      },
    })

    expect(wrapper.find('[data-testid="header-announcement-carousel"]').exists()).toBe(false)
  })

  it('renders the current announcement and emits it when selected', async () => {
    const wrapper = mount(HeaderAnnouncementCarousel, {
      props: { announcements },
      global: {
        stubs: globalStubs,
      },
    })

    expect(wrapper.text()).toContain('公告')
    expect(wrapper.text()).toContain('新模型上线')
    expect(wrapper.text()).not.toContain('支持更多模型路由')
    expect(wrapper.text()).not.toContain('1/2')

    await wrapper.get('[data-testid="header-announcement-carousel"]').trigger('click')

    expect(wrapper.emitted('select')?.[0]).toEqual([announcements[1]])
  })

  it('rotates through announcements and pauses while hovered', async () => {
    const wrapper = mount(HeaderAnnouncementCarousel, {
      props: { announcements },
      global: {
        stubs: globalStubs,
      },
    })

    vi.advanceTimersByTime(5000)
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('维护通知')

    await wrapper.get('[data-testid="header-announcement-carousel"]').trigger('mouseenter')
    vi.advanceTimersByTime(5000)
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('维护通知')

    await wrapper.get('[data-testid="header-announcement-carousel"]').trigger('mouseleave')
    vi.advanceTimersByTime(5000)
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('新模型上线')
  })

  it('keeps only the latest three announcements in the header carousel', async () => {
    const wrapper = mount(HeaderAnnouncementCarousel, {
      props: {
        announcements: [
          ...announcements,
          {
            id: 3,
            title: '最新活动',
            content: '活动说明',
            notify_mode: 'silent',
            created_at: '2026-05-28T03:00:00Z',
            updated_at: '2026-05-28T03:00:00Z',
          },
          {
            id: 4,
            title: '较早公告',
            content: '较早内容',
            notify_mode: 'silent',
            created_at: '2026-05-27T03:00:00Z',
            updated_at: '2026-05-27T03:00:00Z',
          },
        ],
      },
      global: {
        stubs: globalStubs,
      },
    })

    expect(wrapper.text()).toContain('最新活动')

    vi.advanceTimersByTime(5000)
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('新模型上线')

    vi.advanceTimersByTime(5000)
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('维护通知')

    vi.advanceTimersByTime(5000)
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('最新活动')
    expect(wrapper.text()).not.toContain('较早公告')
  })
})
