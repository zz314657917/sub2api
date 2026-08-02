import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserTokenRanking from '../UserTokenRanking.vue'

const getUserBreakdown = vi.fn()

vi.mock('@/api/admin/dashboard', () => ({
  getUserBreakdown: (...args: unknown[]) => getUserBreakdown(...args),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

const makeItem = (id: number, totalTokens: number) => ({
  user_id: id,
  email: `user-${id}@test.com`,
  requests: id,
  input_tokens: totalTokens - 30,
  output_tokens: 20,
  cache_tokens: 10,
  total_tokens: totalTokens,
  cost: 1,
  actual_cost: 0.5,
  account_cost: 0.25,
})

const mountRanking = (filters: Record<string, unknown> = {}) => mount(UserTokenRanking, {
  props: {
    startDate: '2026-07-01',
    endDate: '2026-07-08',
    filters,
    model: 'gpt-5.6',
  },
  global: {
    stubs: {
      Select: { props: ['modelValue', 'options'], template: '<div data-test="limit-select">{{ modelValue }}</div>' },
      LoadingSpinner: true,
    },
  },
})

describe('UserTokenRanking', () => {
  beforeEach(() => {
    getUserBreakdown.mockReset().mockResolvedValue({ users: [makeItem(1, 100), makeItem(2, 50)] })
  })

  it('loads with shared filters and drills into a selected user', async () => {
    const wrapper = mountRanking({ group_id: 3, request_type: 'stream' })
    await flushPromises()

    expect(getUserBreakdown).toHaveBeenCalledWith(expect.objectContaining({
      start_date: '2026-07-01',
      end_date: '2026-07-08',
      group_id: 3,
      request_type: 'stream',
      model: 'gpt-5.6',
      sort_by: 'total_tokens',
      limit: 50,
    }))

    await wrapper.find('tbody tr').trigger('click')
    expect(wrapper.emitted('select-user')?.[0]).toEqual([1, 'user-1@test.com'])
  })

  it('requests another allowlisted ranking when a sortable column is selected', async () => {
    const wrapper = mountRanking()
    await flushPromises()

    const inputSort = wrapper.findAll('thead button').find((button) => button.text().includes('inputTokens'))
    expect(inputSort).toBeDefined()
    await inputSort!.trigger('click')
    await flushPromises()

    expect(getUserBreakdown).toHaveBeenLastCalledWith(expect.objectContaining({ sort_by: 'input_tokens' }))
  })

  it('ignores a stale response after shared filters change', async () => {
    let resolveFirst!: (value: unknown) => void
    getUserBreakdown
      .mockReturnValueOnce(new Promise((resolve) => { resolveFirst = resolve }))
      .mockResolvedValueOnce({ users: [makeItem(9, 900)] })

    const wrapper = mountRanking()
    await wrapper.setProps({ filters: { user_id: 9 } })
    await flushPromises()
    resolveFirst({ users: [makeItem(1, 100)] })
    await flushPromises()

    expect(wrapper.text()).toContain('user-9@test.com')
    expect(wrapper.text()).not.toContain('user-1@test.com')
  })
})
