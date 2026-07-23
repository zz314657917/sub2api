import { describe, expect, it } from 'vitest'

import { formatDateTimeLocalInput } from '../format'

describe('formatDateTimeLocalInput', () => {
  it('formats a timestamp using local wall-clock fields', () => {
    const localDate = new Date(2026, 6, 23, 9, 5, 30)

    expect(formatDateTimeLocalInput(Math.floor(localDate.getTime() / 1000))).toBe('2026-07-23T09:05')
  })

  it('returns an empty string for an empty timestamp', () => {
    expect(formatDateTimeLocalInput(null)).toBe('')
  })
})
