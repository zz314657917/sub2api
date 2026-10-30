import { describe, expect, it } from 'vitest'
import { calendarDayCount, calendarRange } from '../dashboardRange'

describe('dashboard calendar ranges', () => {
  it.each([7, 30, 90])('includes exactly %i local calendar dates across a DST boundary', (days) => {
    const result = calendarRange(days, new Date(2026, 10, 3, 0, 15))
    expect(result.endDate).toBe('2026-11-03')
    expect(calendarDayCount(result.startDate, result.endDate)).toBe(days)
  })
  it('does not shift the end date to UTC near midnight', () => {
    const result = calendarRange(7, new Date(2026, 0, 1, 0, 5))
    expect(result).toEqual({ startDate: '2025-12-26', endDate: '2026-01-01' })
  })
})
