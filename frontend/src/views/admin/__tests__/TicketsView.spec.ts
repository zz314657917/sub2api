import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import TicketsView from '../TicketsView.vue'

const { list, get, createMessage, close, reopen, markRead } = vi.hoisted(() => ({
  list: vi.fn(),
  get: vi.fn(),
  createMessage: vi.fn(),
  close: vi.fn(),
  reopen: vi.fn(),
  markRead: vi.fn(),
}))

const { showError, showSuccess } = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin/tickets', () => ({
  default: {
    list,
    get,
    createMessage,
    markRead,
    close,
    reopen,
  },
  adminTicketsAPI: {
    list,
    get,
    createMessage,
    markRead,
    close,
    reopen,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key),
    }),
  }
})

vi.mock('@/utils/format', () => ({
  formatDateTime: (value: string) => value,
  formatRelativeTime: (value: string) => value,
}))

function ticket(status = 'pending_admin') {
  return {
    id: 9,
    user_id: 12,
    title: '账单问题',
    status,
    ticket_type: 'support',
    last_message_preview: '用户留言',
    last_message_at: '2026-06-01T12:00:00Z',
    user_unread_count: 0,
    admin_unread_count: status === 'pending_admin' ? 1 : 0,
    created_at: '2026-06-01T11:00:00Z',
    updated_at: '2026-06-01T12:00:00Z',
    user: {
      id: 12,
      email: 'user@example.com',
      username: 'user',
    },
  }
}

function systemTicket() {
  return {
    ...ticket('open'),
    id: 5,
    title: '系统通知',
    ticket_type: 'system',
    system_key: 'system',
    last_message_preview: '分组倍率已调整',
    admin_unread_count: 0,
  }
}

function mockDetail(status = 'pending_admin') {
  get.mockResolvedValue({
    ticket: ticket(status),
    messages: [
      {
        id: 1,
        ticket_id: 9,
        sender_type: 'user',
        sender_user_id: 12,
        content: '用户留言',
        created_at: '2026-06-01T12:00:00Z',
      },
    ],
  })
}

function mountView() {
  return mount(TicketsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: { template: '<span />' },
        Pagination: { template: '<div />' },
      },
    },
  })
}

describe('admin TicketsView', () => {
  beforeEach(() => {
    window.matchMedia = vi.fn().mockReturnValue({
      matches: true,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }) as unknown as typeof window.matchMedia
    list.mockReset().mockResolvedValue({
      items: [ticket()],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    get.mockReset()
    mockDetail()
    createMessage.mockReset().mockResolvedValue({ ticket: ticket('pending_user'), messages: [] })
    close.mockReset().mockResolvedValue(undefined)
    reopen.mockReset().mockResolvedValue(undefined)
    markRead.mockReset().mockResolvedValue(undefined)
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('loads tickets and renders user context', async () => {
    const dispatchSpy = vi.spyOn(window, 'dispatchEvent')
    const wrapper = mountView()
    await flushPromises()
    await flushPromises()

    expect(list).toHaveBeenCalledWith(expect.objectContaining({ page: 1, page_size: 20, ticket_type: 'support' }))
    expect(list).toHaveBeenCalledWith(expect.objectContaining({ sort_by: 'last_message_at', sort_order: 'desc' }))
    expect(get).toHaveBeenCalledWith(9)
    expect(markRead).toHaveBeenCalledWith(9)
    expect(wrapper.find('.unread-pill').exists()).toBe(false)
    expect(wrapper.text()).toContain('账单问题')
    expect(wrapper.text()).toContain('user@example.com')
    expect(wrapper.text()).toContain('用户留言')
    expect(wrapper.text()).toContain('admin.tickets.supportDefaultHint')
    expect(dispatchSpy).toHaveBeenCalledWith(expect.objectContaining({ type: 'sub2api:ticket-unread-updated' }))
    dispatchSpy.mockRestore()
  })

  it('keeps admin unread list state on mobile until the admin opens a conversation', async () => {
    window.matchMedia = vi.fn().mockReturnValue({
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }) as unknown as typeof window.matchMedia

    const wrapper = mountView()
    await flushPromises()

    expect(get).not.toHaveBeenCalled()
    expect(markRead).not.toHaveBeenCalled()
    expect(wrapper.find('.unread-pill').exists()).toBe(true)

    await wrapper.find('.ticket-list-item').trigger('click')
    await flushPromises()
    await flushPromises()

    expect(get).toHaveBeenCalledWith(9)
    expect(markRead).toHaveBeenCalledWith(9)
    expect(wrapper.find('.unread-pill').exists()).toBe(false)
  })

  it('sends an admin reply and refreshes detail/list', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('form.border-t textarea').setValue('已处理')
    await wrapper.find('form.border-t').trigger('submit')
    await flushPromises()

    expect(createMessage).toHaveBeenCalledWith(9, { content: '已处理' })
    expect(showSuccess).toHaveBeenCalledWith('admin.tickets.sent')
    expect(get).toHaveBeenCalledTimes(2)
    expect(list).toHaveBeenCalledTimes(2)
  })

  it('reopens a closed ticket', async () => {
    list.mockResolvedValueOnce({
      items: [ticket('closed')],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    get.mockReset()
    mockDetail('closed')

    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('admin.tickets.reopen'))?.trigger('click')
    await flushPromises()

    expect(reopen).toHaveBeenCalledWith(9)
    expect(showSuccess).toHaveBeenCalledWith('admin.tickets.reopenedSuccess')
  })

  it('filters ticket type and renders system ticket as view-only', async () => {
    list.mockResolvedValue({
      items: [systemTicket()],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    get.mockResolvedValue({
      ticket: systemTicket(),
      messages: [
        {
          id: 2,
          ticket_id: 5,
          sender_type: 'system',
          content: '分组倍率已调整',
          event_type: 'group_changed',
          event_key: 'group_changed:12:20260601120000',
          metadata: {
            action_type: 'group_changed',
            group_id: 5,
            old_rate_multiplier: 0.06,
            new_rate_multiplier: 0.08,
            old_rpm_limit: 60,
            new_rpm_limit: 120,
          },
          created_at: '2026-06-01T12:00:00Z',
        },
      ],
    })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('select')[1].setValue('system')
    await wrapper.get('[data-test="ticket-sort-filter"]').setValue('unread_first')
    await flushPromises()

    expect(list).toHaveBeenLastCalledWith(expect.objectContaining({ ticket_type: 'system', sort_by: 'unread_first', sort_order: 'desc' }))
    expect(wrapper.text()).toContain('系统通知')
    expect(wrapper.text()).toContain('admin.tickets.systemAuditHint')
    expect(wrapper.text()).toContain('admin.tickets.readOnlyStatus')
    expect(wrapper.text()).toContain('分组倍率已调整')
    expect(wrapper.text()).toContain('tickets.metadata.changeDetails')
    expect(wrapper.text()).toContain('0.06x')
    expect(wrapper.text()).toContain('0.08x')
    expect(wrapper.text()).toContain('120')
    expect(wrapper.find('form.border-t').exists()).toBe(false)
    expect(wrapper.text()).toContain('admin.tickets.systemReadOnly')
    expect(wrapper.findAll('button').some((button) => button.text().includes('admin.tickets.close'))).toBe(false)
    expect(wrapper.findAll('button').some((button) => button.text().includes('admin.tickets.reopen'))).toBe(false)
  })

  it('passes system notification audit filters to ticket list params', async () => {
    vi.useFakeTimers()
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('select')[1].setValue('system')
    await flushPromises()
    await wrapper.get('[data-test="ticket-event-type-filter"]').setValue('group_changed')
    await wrapper.get('[data-test="ticket-event-key-filter"]').setValue('group_changed:12:20260601120000')
    vi.advanceTimersByTime(300)
    await flushPromises()

    await wrapper.get('[data-test="ticket-date-from-filter"]').setValue('2026-06-01')
    await wrapper.get('[data-test="ticket-date-to-filter"]').setValue('2026-06-02')
    await flushPromises()

    expect(list).toHaveBeenLastCalledWith(expect.objectContaining({
      ticket_type: 'system',
      event_type: 'group_changed',
      event_key: 'group_changed:12:20260601120000',
      date_from: '2026-06-01',
      date_to: '2026-06-02',
    }))
    vi.useRealTimers()
  })

  it('clears system audit filters when switching back to support tickets', async () => {
    vi.useFakeTimers()
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('select')[1].setValue('system')
    await flushPromises()
    await wrapper.get('[data-test="ticket-event-type-filter"]').setValue('group_changed')
    await wrapper.get('[data-test="ticket-event-key-filter"]').setValue('group_changed:12')
    vi.advanceTimersByTime(300)
    await flushPromises()

    await wrapper.findAll('select')[1].setValue('support')
    await flushPromises()

    expect(list).toHaveBeenLastCalledWith(expect.objectContaining({
      ticket_type: 'support',
      event_type: undefined,
      event_key: undefined,
    }))
    vi.useRealTimers()
  })
})
