import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAnnouncementStore } from '../announcements'

const markRead = vi.hoisted(() => vi.fn())
vi.mock('@/api', () => ({ announcementsAPI: { markRead } }))

describe('mark all announcements read', () => {
  beforeEach(() => { setActivePinia(createPinia()); markRead.mockReset(); vi.spyOn(console, 'error').mockImplementation(() => {}) })

  it('retains successful reads when another bulk request fails', async () => {
    const store = useAnnouncementStore()
    store.announcements = [{ id: 1, title: 'one' }, { id: 2, title: 'two' }] as never
    markRead.mockImplementation((id: number) => id === 1 ? Promise.resolve() : Promise.reject(new Error('offline')))
    await expect(store.markAllAsRead()).rejects.toThrow('offline')
    expect(store.announcements[0].read_at).toBeTruthy()
    expect(store.announcements[1].read_at).toBeFalsy()
  })
})
