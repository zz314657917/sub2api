import { describe, expect, it } from 'vitest'
import { createI18n } from 'vue-i18n'

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

  it('keeps display labels for audit roles, auth methods, and known actions', () => {
    expect(zh.admin.audit.roles.admin).toBe('管理员')
    expect(zh.admin.audit.authMethods.adminApiKey).toBe('管理员 API Key')
    expect(zh.admin.audit.actions['auth.login']).toBe('登录')
    expect(zh.admin.audit.actions['admin.accounts.export']).toBe('导出账号')
    expect(zh.admin.audit.actionParts.delete).toBe('删除')

    expect(en.admin.audit.roles.admin).toBe('Administrator')
    expect(en.admin.audit.authMethods.adminApiKey).toBe('Admin API Key')
    expect(en.admin.audit.actions['auth.login']).toBe('Login')
    expect(en.admin.audit.actions['admin.accounts.export']).toBe('Export accounts')
    expect(en.admin.audit.actionParts.delete).toBe('Delete')
  })

  it('keeps exact actions addressable by their raw dotted values', () => {
    expect('auth.login' in zh.admin.audit.actions).toBe(true)
    expect('auth.login' in en.admin.audit.actions).toBe(true)

    const i18n = createI18n({
      legacy: false,
      locale: 'zh',
      messages: { zh }
    })
    const actions = i18n.global.tm('admin.audit.actions') as Record<string, unknown>
    expect(actions['auth.login']).toBe('登录')
    expect(actions['auth_login']).toBeUndefined()
  })
})
