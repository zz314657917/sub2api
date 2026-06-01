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

import { adminTicketsAPI } from '@/api/admin/tickets'

describe('admin tickets api contract', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
  })

  it('uses the fixed admin ticket endpoints', async () => {
    get.mockResolvedValueOnce({ data: { items: [], total: 0, page: 1, page_size: 20, pages: 0 } })
    get.mockResolvedValueOnce({ data: { ticket: { id: 9 }, messages: [] } })
    post.mockResolvedValue({ data: { id: 1 } })

    await adminTicketsAPI.list({
      page: 2,
      page_size: 30,
      status: 'closed',
      ticket_type: 'system',
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
        ticket_type: 'system',
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
