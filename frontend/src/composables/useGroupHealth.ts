import { ref } from 'vue'
import { channelMonitorUserAPI, usageAPI } from '@/api'
import type { MonitorTimelinePoint } from '@/api/channelMonitor'

export interface GroupHealth {
  /** 7-day probe availability (0-100) from channel monitors tagged with this group name. */
  availability: number | null
  /** Newest-first probe points of the first matching monitor. */
  timeline: MonitorTimelinePoint[]
  /** Caller's own 7-day cache hit rate in this group (0-100): read / (input + read). */
  cacheRate: number | null
}

const HEALTH_WINDOW_DAYS = 7

const formatDate = (date: Date) => {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

/**
 * Best-effort group health for the key editor. Both sources are optional:
 * failures leave the map empty so the UI simply hides the strip.
 */
export function useGroupHealth() {
  const availabilityByName = ref(new Map<string, { availability: number; timeline: MonitorTimelinePoint[] }>())
  const cacheRateById = ref(new Map<number, number>())

  const loadMonitors = async () => {
    const { items } = await channelMonitorUserAPI.list()
    const buckets = new Map<string, { sum: number; count: number; timeline: MonitorTimelinePoint[] }>()
    for (const item of items ?? []) {
      const name = item.group_name?.trim()
      if (!name) continue
      const bucket = buckets.get(name) ?? { sum: 0, count: 0, timeline: item.timeline ?? [] }
      bucket.sum += item.availability_7d
      bucket.count += 1
      buckets.set(name, bucket)
    }
    const next = new Map<string, { availability: number; timeline: MonitorTimelinePoint[] }>()
    for (const [name, bucket] of buckets) {
      next.set(name, { availability: bucket.sum / bucket.count, timeline: bucket.timeline })
    }
    availabilityByName.value = next
  }

  const loadCacheRates = async () => {
    const end = new Date()
    const start = new Date(end)
    start.setDate(end.getDate() - (HEALTH_WINDOW_DAYS - 1))
    const snapshot = await usageAPI.getDashboardSnapshotV2({
      start_date: formatDate(start),
      end_date: formatDate(end),
      include_trend: false,
      include_model_stats: false,
      include_group_stats: true
    })
    const next = new Map<number, number>()
    for (const stat of snapshot.groups ?? []) {
      const input = stat.input_tokens ?? 0
      const read = stat.cache_read_tokens ?? 0
      if (stat.group_id > 0 && input + read > 0) {
        next.set(stat.group_id, (read / (input + read)) * 100)
      }
    }
    cacheRateById.value = next
  }

  const load = async () => {
    await Promise.allSettled([loadMonitors(), loadCacheRates()])
  }

  const healthFor = (groupId: number | null | undefined, groupName: string | null | undefined): GroupHealth | null => {
    const monitor = groupName ? availabilityByName.value.get(groupName.trim()) : undefined
    const cacheRate = groupId ? cacheRateById.value.get(groupId) ?? null : null
    if (!monitor && cacheRate === null) return null
    return {
      availability: monitor?.availability ?? null,
      timeline: monitor?.timeline ?? [],
      cacheRate
    }
  }

  return { load, healthFor }
}
