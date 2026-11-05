import axios, { type InternalAxiosRequestConfig } from 'axios'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { OpenAIFastPolicy } from '@/types'

const { transport } = vi.hoisted(() => ({ transport: vi.fn() }))

// Keep Axios' real JSON transformation; intercept only the network boundary.
vi.mock('@/api/client', async () => {
  const { default: axios } = await import('axios')
  return {
    apiClient: axios.create({
      adapter: async (config) => ({
        config,
        status: 200,
        statusText: 'OK',
        headers: {},
        data: await transport(config)
      })
    })
  }
})

import { create, getById, list, toggleStatus, update } from '@/api/keys'

const policies: OpenAIFastPolicy[] = ['passthrough', 'force', 'filter']
const body = () => JSON.parse(transport.mock.lastCall![0].data as string)
const request = () => transport.mock.lastCall![0] as InternalAxiosRequestConfig

describe('API key OpenAI Fast policy serialization', () => {
  beforeEach(() => {
    transport.mockReset().mockResolvedValue({ id: 17, name: 'test-key' })
  })

  it.each(policies)('serializes explicit create policy %s at the appended argument', async (policy) => {
    const routes = [{ group_id: 3, priority: 1, weight: 1, cooldown_seconds: 30, enabled: true }]
    const result = await create(
      'test-key', 2, 'custom-key-12345678', ['127.0.0.1'], ['192.0.2.1'],
      10, 30, { rate_limit_5h: 1, rate_limit_1d: 2, rate_limit_7d: 3 },
      routes, 'private_first', 4, policy
    )

    expect(request()).toMatchObject({ method: 'post', url: '/keys' })
    expect(body()).toEqual({
      name: 'test-key',
      group_id: 2,
      custom_key: 'custom-key-12345678',
      ip_whitelist: ['127.0.0.1'],
      ip_blacklist: ['192.0.2.1'],
      quota: 10,
      expires_in_days: 30,
      rate_limit_5h: 1,
      rate_limit_1d: 2,
      rate_limit_7d: 3,
      multi_group_routes: routes,
      account_pool_strategy: 'private_first',
      token_multiplier_cap: 4,
      openai_fast_policy: policy
    })
    expect(result).toEqual({ id: 17, name: 'test-key' })
  })

  it('keeps short legacy create calls compatible with the server default', async () => {
    await create('legacy')
    expect(body()).toEqual({ name: 'legacy' })
  })

  it('preserves all eleven legacy create positions without injecting a policy', async () => {
    await create('legacy', 2, undefined, [], [], 0, undefined, {}, [], 'shared_only', 0)
    expect(body()).toEqual({
      name: 'legacy',
      group_id: 2,
      multi_group_routes: [],
      account_pool_strategy: 'shared_only',
      token_multiplier_cap: 0
    })
  })

  it.each(policies)('serializes an update to %s, including an explicit reset to passthrough', async (policy) => {
    await update(17, { name: 'renamed', openai_fast_policy: policy })
    expect(request()).toMatchObject({ method: 'put', url: '/keys/17' })
    expect(body()).toEqual({ name: 'renamed', openai_fast_policy: policy })
  })

  it('keeps omitted update policies absent rather than resetting the saved preference', async () => {
    await update(17, { name: 'renamed' })
    expect(body()).toEqual({ name: 'renamed' })
  })

  it('does not inject policy into status or quota-only requests', async () => {
    await toggleStatus(17, 'inactive')
    expect(body()).toEqual({ status: 'inactive' })
    await update(17, { reset_quota: true })
    expect(body()).toEqual({ reset_quota: true })
  })

  it('accepts old get/list responses with no policy field', async () => {
    const legacy = { id: 17, name: 'legacy' }
    transport.mockResolvedValueOnce(legacy)
    expect(await getById(17)).toEqual(legacy)
    transport.mockResolvedValueOnce({ items: [legacy], total: 1 })
    expect((await list()).items).toEqual([legacy])
  })

  it.each(policies)('returns policy %s from the server without altering it', async (policy) => {
    transport.mockResolvedValueOnce({ id: 17, openai_fast_policy: policy })
    expect((await getById(17)).openai_fast_policy).toBe(policy)
  })

  it('propagates a managed-key rejection without retrying another write endpoint', async () => {
    const error = new axios.AxiosError('CAFE_MANAGED_KEY_PROTECTED')
    transport.mockRejectedValueOnce(error)
    await expect(update(17, { openai_fast_policy: 'force' })).rejects.toBe(error)
    expect(transport).toHaveBeenCalledTimes(1)
    expect(request()).toMatchObject({ method: 'put', url: '/keys/17' })
  })
})
