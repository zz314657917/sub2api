import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('audit log locales', () => {
  it('exposes audit messages at the keys used by the admin page', () => {
    expect(zh.admin.audit.title).toBe('操作日志')
    expect(zh.admin.audit.filters.q).toBe('关键字')
    expect(zh.admin.audit.columns.detail).toBe('详情')

    expect(en.admin.audit.title).toBe('Audit Logs')
    expect(en.admin.audit.filters.q).toBe('Keyword')
    expect(en.admin.audit.columns.detail).toBe('Detail')
  })

  it('does not add a duplicate audit namespace', () => {
    expect(zh.admin.audit).not.toHaveProperty('audit')
    expect(en.admin.audit).not.toHaveProperty('audit')
  })
})
