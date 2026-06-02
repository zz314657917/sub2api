import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    post,
  },
}))

import { ticketsAPI } from '@/api/tickets'
import { adminTicketsAPI } from '@/api/admin/tickets'

describe('tickets api contract', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
  })

  it('uses the fixed user ticket endpoints', async () => {
    get.mockResolvedValueOnce({ data: { items: [], total: 0, page: 1, page_size: 20, pages: 0 } })
    get.mockResolvedValueOnce({ data: { support_unread: 1, system_unread: 2, total_unread: 3 } })
    get.mockResolvedValueOnce({ data: { ticket: { id: 7 }, messages: [] } })
    post.mockResolvedValue({ data: { id: 1 } })

    await ticketsAPI.list({ page: 1, page_size: 20, status: 'open', ticket_type: 'system', search: 'quota', unread_only: true })
    await ticketsAPI.unreadSummary()
    await ticketsAPI.create({ title: 'Need help', content: 'Plain text only' })
    await ticketsAPI.get(7)
    await ticketsAPI.createMessage(7, { content: 'Thanks' })
    await ticketsAPI.markRead(7)
    await ticketsAPI.close(7)

    expect(get).toHaveBeenNthCalledWith(1, '/user/tickets', {
      params: {
        page: 1,
        page_size: 20,
        status: 'open',
        ticket_type: 'system',
        search: 'quota',
        unread_only: true,
      },
    })
    expect(get).toHaveBeenNthCalledWith(2, '/user/tickets/unread-summary')
    expect(post).toHaveBeenNthCalledWith(1, '/user/tickets', {
      title: 'Need help',
      content: 'Plain text only',
    })
    expect(get).toHaveBeenNthCalledWith(3, '/user/tickets/7')
    expect(post).toHaveBeenNthCalledWith(2, '/user/tickets/7/messages', { content: 'Thanks' })
    expect(post).toHaveBeenNthCalledWith(3, '/user/tickets/7/read')
    expect(post).toHaveBeenNthCalledWith(4, '/user/tickets/7/close')
  })

  it('uses the fixed admin ticket endpoints', async () => {
    get.mockResolvedValueOnce({ data: { items: [], total: 0, page: 1, page_size: 20, pages: 0 } })
    get.mockResolvedValueOnce({ data: { ticket: { id: 9 }, messages: [] } })
    post.mockResolvedValue({ data: { id: 1 } })

    await adminTicketsAPI.list({
      page: 2,
      page_size: 30,
      status: 'closed',
      ticket_type: 'support',
      search: 'billing',
      user_id: 12,
      unread_only: true,
      sort_by: 'unread_first',
      sort_order: 'desc',
    })
    await adminTicketsAPI.get(9)
    await adminTicketsAPI.createMessage(9, { content: 'Admin reply' })
    await adminTicketsAPI.markRead(9)
    await adminTicketsAPI.close(9)
    await adminTicketsAPI.reopen(9)
    await adminTicketsAPI.createForUser(12, { title: 'Follow up', content: 'Please check this.' })

    expect(get).toHaveBeenNthCalledWith(1, '/admin/tickets', {
      params: {
        page: 2,
        page_size: 30,
        status: 'closed',
        ticket_type: 'support',
        search: 'billing',
        user_id: 12,
        unread_only: true,
        sort_by: 'unread_first',
        sort_order: 'desc',
      },
    })
    expect(get).toHaveBeenNthCalledWith(2, '/admin/tickets/9')
    expect(post).toHaveBeenNthCalledWith(1, '/admin/tickets/9/messages', { content: 'Admin reply' })
    expect(post).toHaveBeenNthCalledWith(2, '/admin/tickets/9/read')
    expect(post).toHaveBeenNthCalledWith(3, '/admin/tickets/9/close')
    expect(post).toHaveBeenNthCalledWith(4, '/admin/tickets/9/reopen')
    expect(post).toHaveBeenNthCalledWith(5, '/admin/users/12/tickets', {
      title: 'Follow up',
      content: 'Please check this.',
    })
  })
})
