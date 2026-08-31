import { describe, expect, it } from 'vitest'

import { formatDateTimeLocalInput, parseDateTimeLocalInput } from '../format'

describe('formatDateTimeLocalInput', () => {
  it('formats a timestamp using local wall-clock fields', () => {
    const localDate = new Date(2026, 6, 23, 9, 5, 30)

    expect(formatDateTimeLocalInput(Math.floor(localDate.getTime() / 1000))).toBe('2026-07-23T09:05')
  })

  it('returns an empty string for an empty timestamp', () => {
    expect(formatDateTimeLocalInput(null)).toBe('')
  })
})

describe('parseDateTimeLocalInput', () => {
  it('converts local wall-clock components to seconds and truncates fractions', () => {
    const expected = new Date(0)
    expected.setFullYear(2026, 6, 23)
    expected.setHours(9, 5, 30, 999)

    expect(parseDateTimeLocalInput('2026-07-23T09:05:30.9999')).toBe(
      Math.floor(expected.getTime() / 1000)
    )
    expect(parseDateTimeLocalInput('2026-07-23T09:05')).toBe(
      Math.floor(new Date(2026, 6, 23, 9, 5).getTime() / 1000)
    )
  })

  it.each([
    '',
    '2026-07-23 09:05',
    '2026-07-23T09:05Z',
    '2026-07-23T09:05+08:00',
    '2026-02-29T09:05',
    '2024-02-30T09:05',
    '2026-13-01T09:05',
    '2026-07-23T24:00',
    '2026-07-23T09:60',
    '2026-07-23T09:05:60',
    '2026-07-23T09:05:30Z'
  ])('rejects malformed, timezone-bearing, or overflowing value %s', (value) => {
    expect(parseDateTimeLocalInput(value)).toBeNull()
  })
})
