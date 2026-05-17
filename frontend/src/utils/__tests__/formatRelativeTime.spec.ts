import { afterEach, describe, expect, it } from 'vitest'

import { i18n } from '@/i18n'
import { formatRelativeTime } from '../format'

const originalLocale = i18n.global.locale.value

afterEach(() => {
  i18n.global.locale.value = originalLocale
})

describe('formatRelativeTime', () => {
  it('falls back to readable Chinese text when locale messages are not loaded', () => {
    i18n.global.locale.value = 'zh'

    expect(formatRelativeTime(null)).toBe('从未')

    const fiveHoursAgo = new Date(Date.now() - 5 * 60 * 60 * 1000)
    expect(formatRelativeTime(fiveHoursAgo)).toBe('5小时前')
  })

  it('falls back to readable English text when locale messages are not loaded', () => {
    i18n.global.locale.value = 'en'

    expect(formatRelativeTime(null)).toBe('Never')

    const fiveHoursAgo = new Date(Date.now() - 5 * 60 * 60 * 1000)
    expect(formatRelativeTime(fiveHoursAgo)).toBe('5h ago')
  })
})
