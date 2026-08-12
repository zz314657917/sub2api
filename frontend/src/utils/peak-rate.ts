/**
 * 高峰时段倍率的共享展示逻辑。
 *
 * 高峰窗口由后端按服务器全局时区判定（Group.PeakMultiplierAt），
 * 前端展示必须带上服务器时区标注（来自公共设置 server_utc_offset），
 * 避免用户按浏览器本地时间误读计费窗口。
 */

export interface PeakRateFields {
  peak_rate_enabled?: boolean
  peak_start?: string
  peak_end?: string
  peak_rate_multiplier?: number
}

export function hasPeakRate(fields?: PeakRateFields | null): boolean {
  return Boolean(fields?.peak_rate_enabled && fields.peak_start && fields.peak_end)
}

/** "+08:00" -> "UTC+08:00"; missing offsets should still show that server time applies. */
export function serverTimezoneLabel(utcOffset?: string | null, fallback = 'Server time'): string {
  return utcOffset ? `UTC${utcOffset}` : fallback
}

export function timeOfDayMinutes(value?: string | null): number | null {
  const match = value?.trim().match(/^(\d{1,2}):(\d{2})$/)
  if (!match) return null
  const hour = Number(match[1])
  const minute = Number(match[2])
  if (!Number.isInteger(hour) || !Number.isInteger(minute) || hour > 23 || minute > 59) {
    return null
  }
  return hour * 60 + minute
}

export function normalizeTimeInputValue(value?: string | null): string {
  const minutes = timeOfDayMinutes(value)
  if (minutes === null) return value?.trim() || ''
  const hour = Math.floor(minutes / 60).toString().padStart(2, '0')
  const minute = (minutes % 60).toString().padStart(2, '0')
  return `${hour}:${minute}`
}

/** "14:00-18:00 ×2 (UTC+08:00)"，tzLabel 为空时省略括号部分 */
export function formatPeakRateWindow(
  fields: PeakRateFields | null | undefined,
  tzLabel?: string
): string {
  if (!hasPeakRate(fields) || !fields) return ''
  const base = `${fields.peak_start}-${fields.peak_end} ×${fields.peak_rate_multiplier ?? 1}`
  return tzLabel ? `${base} (${tzLabel})` : base
}
