import { describe, expect, it, vi } from 'vitest'

const { get } = vi.hoisted(() => ({ get: vi.fn() }))

vi.mock('../client', () => ({
  apiClient: { get }
}))

import { getUsageSummary } from '../admin/groups'

describe('admin group usage summary API', () => {
  it('does not send the browser timezone', async () => {
    get.mockResolvedValue({ data: [] })

    await getUsageSummary()

    expect(get).toHaveBeenCalledWith('/admin/groups/usage-summary')
  })
})
