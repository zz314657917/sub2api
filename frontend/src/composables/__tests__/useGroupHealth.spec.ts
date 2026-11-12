import { describe, expect, it, vi } from 'vitest'

const { listMonitors, getSnapshot } = vi.hoisted(() => ({
  listMonitors: vi.fn(),
  getSnapshot: vi.fn()
}))

vi.mock('@/api', () => ({
  channelMonitorUserAPI: { list: listMonitors },
  usageAPI: { getDashboardSnapshotV2: getSnapshot }
}))

import { useGroupHealth } from '../useGroupHealth'

describe('useGroupHealth', () => {
  it('averages monitor availability by group name and computes own cache hit rate by group id', async () => {
    listMonitors.mockResolvedValue({
      items: [
        { group_name: 'pro max', availability_7d: 100, timeline: [{ status: 'operational' }] },
        { group_name: ' pro max ', availability_7d: 98, timeline: [] },
        { group_name: '', availability_7d: 10, timeline: [] }
      ]
    })
    getSnapshot.mockResolvedValue({
      groups: [
        { group_id: 7, input_tokens: 1000, cache_read_tokens: 3000 },
        { group_id: 8, input_tokens: 0, cache_read_tokens: 0 }
      ]
    })

    const { load, healthFor } = useGroupHealth()
    await load()

    expect(healthFor(7, 'pro max')).toEqual({
      availability: 99,
      timeline: [{ status: 'operational' }],
      cacheRate: 75
    })
    // No traffic and no monitor: hide instead of showing a fake 0%.
    expect(healthFor(8, 'unknown')).toBeNull()
    expect(getSnapshot).toHaveBeenCalledWith(expect.objectContaining({ include_group_stats: true, include_trend: false }))
  })

  it('degrades to the remaining source when one request fails', async () => {
    listMonitors.mockRejectedValue(new Error('down'))
    getSnapshot.mockResolvedValue({ groups: [{ group_id: 3, input_tokens: 1, cache_read_tokens: 1 }] })

    const { load, healthFor } = useGroupHealth()
    await load()

    expect(healthFor(3, 'x')).toEqual({ availability: null, timeline: [], cacheRate: 50 })
  })
})
