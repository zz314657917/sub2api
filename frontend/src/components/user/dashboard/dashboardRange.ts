/** Calendar dates in the user's timezone, including today; safe across DST transitions. */
export function calendarRange(days: number, now = new Date()) {
  const start = new Date(now)
  start.setDate(start.getDate() - days + 1)
  return { startDate: localDate(start), endDate: localDate(now) }
}
export function localDate(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}
export function calendarDayCount(start: string, end: string): number {
  // UTC counts labels only; API requests carry the user's timezone.
  return (
    Math.round((Date.parse(`${end}T00:00:00Z`) - Date.parse(`${start}T00:00:00Z`)) / 86400000) + 1
  )
}
