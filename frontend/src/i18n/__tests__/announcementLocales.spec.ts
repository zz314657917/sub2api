import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('announcement empty-state locale copy', () => {
  it('provides a creation prompt instead of an error message', () => {
    expect(en.admin.announcements.createFirstAnnouncement).toBe(
      'No announcements yet. Create your first one.'
    )
    expect(zh.admin.announcements.createFirstAnnouncement).toBe(
      '还没有公告，创建您的第一条公告。'
    )
    expect(en.admin.announcements.createFirstAnnouncement).not.toContain('Failed')
    expect(zh.admin.announcements.createFirstAnnouncement).not.toContain('失败')
  })
})
