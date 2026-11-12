import { describe, expect, it, vi } from 'vitest'

vi.mock('@/api', () => ({ userChannelsAPI: { getAvailable: vi.fn() } }))

import { indexModelsByGroup } from '../useGroupModelPricing'
import type { UserAvailableChannel } from '@/api/channels'

const model = (name: string, platform = 'anthropic') => ({ name, platform, pricing: null })
const group = (id: number) => ({ id, name: `g${id}`, platform: 'anthropic' }) as UserAvailableChannel['platforms'][number]['groups'][number]

describe('indexModelsByGroup', () => {
  it('merges models across channels per group, dedupes and sorts by name', () => {
    const channels = [
      { name: 'a', description: '', platforms: [{ platform: 'anthropic', groups: [group(1), group(2)], supported_models: [model('opus'), model('haiku')] }] },
      { name: 'b', description: '', platforms: [{ platform: 'anthropic', groups: [group(1)], supported_models: [model('haiku'), model('sonnet')] }] }
    ] as UserAvailableChannel[]

    const index = indexModelsByGroup(channels)

    expect(index.get(1)?.map((m) => m.name)).toEqual(['haiku', 'opus', 'sonnet'])
    expect(index.get(2)?.map((m) => m.name)).toEqual(['haiku', 'opus'])
    expect(index.get(3)).toBeUndefined()
  })
})
