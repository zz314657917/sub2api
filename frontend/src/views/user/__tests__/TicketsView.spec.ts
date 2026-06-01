import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import TicketsView from '../TicketsView.vue'

const push = vi.fn()

const { list, get, createMessage, close, markRead } = vi.hoisted(() => ({
  list: vi.fn(),
  get: vi.fn(),
  createMessage: vi.fn(),
  close: vi.fn(),
  markRead: vi.fn(),
}))

const { showError, showSuccess } = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/tickets', () => ({
  ticketsAPI: {
    list,
    get,
    create: vi.fn(),
    createMessage,
    markRead,
    close,
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
      t: (key: string) => key,
    }),
  }
})

vi.mock('vue-router', () => ({
  useRouter: () => ({ push }),
}))

vi.mock('@/utils/format', () => ({
  formatDateTime: (value: string) => value,
  formatRelativeTime: (value: string) => value,
}))

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

function supportTicket(overrides: Record<string, unknown> = {}) {
  return {
    id: 7,
    user_id: 1,
    title: '支付已确认',
    status: 'pending_user',
    ticket_type: 'support',
    last_message_preview: '管理员回复',
    last_message_at: '2026-06-01T12:00:00Z',
    user_unread_count: 1,
    admin_unread_count: 0,
    created_at: '2026-06-01T11:00:00Z',
    updated_at: '2026-06-01T12:00:00Z',
    ...overrides,
  }
}

function systemTicket() {
  return supportTicket({
    id: 3,
    title: '系统通知',
    status: 'open',
    ticket_type: 'system',
    system_key: 'system',
    last_message_preview: '充值已到账',
  })
}

describe('user TicketsView', () => {
  beforeEach(() => {
    list.mockReset().mockResolvedValue({
      items: [
        supportTicket(),
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    get.mockReset().mockResolvedValue({
      ticket: supportTicket(),
      messages: [
        {
          id: 1,
          ticket_id: 7,
          sender_type: 'admin',
          sender_user_id: 99,
          content: '管理员回复',
          created_at: '2026-06-01T12:00:00Z',
        },
      ],
    })
    createMessage.mockReset().mockResolvedValue({ ticket: { id: 7 }, messages: [] })
    close.mockReset().mockResolvedValue(undefined)
    markRead.mockReset().mockResolvedValue(undefined)
    push.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('loads tickets and renders the selected conversation', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(list).toHaveBeenCalledWith(expect.objectContaining({ page: 1, page_size: 20 }))
    expect(get).toHaveBeenCalledWith(7)
    expect(wrapper.text()).toContain('支付已确认')
    expect(wrapper.text()).toContain('管理员回复')
  })

  it('marks an unread conversation as read when opened', async () => {
    const wrapper = mountView()
    await flushPromises()
    await flushPromises()

    expect(markRead).toHaveBeenCalledWith(7)
    expect(wrapper.find('.unread-pill').exists()).toBe(false)
  })

  it('sends a user reply and refreshes detail/list', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('form.border-t textarea').setValue('收到，谢谢')
    await wrapper.find('form.border-t').trigger('submit')
    await flushPromises()

    expect(createMessage).toHaveBeenCalledWith(7, { content: '收到，谢谢' })
    expect(showSuccess).toHaveBeenCalledWith('tickets.sent')
    expect(get).toHaveBeenCalledTimes(2)
    expect(list).toHaveBeenCalledTimes(2)
  })

  it('renders system ticket as pinned read-only conversation', async () => {
    list.mockResolvedValue({
      items: [systemTicket(), supportTicket()],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    get.mockResolvedValue({
      ticket: systemTicket(),
      messages: [
        {
          id: 2,
          ticket_id: 3,
          sender_type: 'system',
          content: '充值已到账',
          event_type: 'payment_completed',
          event_key: 'payment_completed:order-1',
          metadata: { action_type: 'payment_completed', order_id: 'order-1' },
          created_at: '2026-06-01T12:00:00Z',
        },
      ],
    })

    const wrapper = mountView()
    await flushPromises()

    expect(get).toHaveBeenCalledWith(3)
    expect(wrapper.text()).toContain('tickets.systemTicket')
    expect(wrapper.text()).toContain('充值已到账')
    expect(wrapper.find('form.border-t').exists()).toBe(false)
    expect(wrapper.text()).toContain('tickets.systemReadOnly')
    expect(wrapper.text()).toContain('tickets.actions.paymentCompleted')

    await wrapper.find('.system-action-link').trigger('click')
    expect(push).toHaveBeenCalledWith('/orders')
  })
})
