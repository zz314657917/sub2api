import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import UserDashboardModels from '../UserDashboardModels.vue'
import type { ModelStat } from '@/types'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/i18n', () => ({
  i18n: { global: { locale: { value: 'en' }, t: (key: string) => key } }
}))

const model = (name: string, requests: number, actualCost: number): ModelStat => ({
  model: name,
  requests,
  actual_cost: actualCost,
  cost: 50,
  total_tokens: 100,
  input_tokens: 50,
  output_tokens: 50,
  cache_creation_tokens: 0,
  cache_read_tokens: 0,
  account_cost: 99
})
const global = { stubs: { Icon: true, RouterLink: true, LoadingSpinner: true } }

describe('model spending ranking', () => {
  it('ranks high spend above high request volume and sizes bars by actual spend', () => {
    const wrapper = mount(UserDashboardModels, {
      props: {
        models: [model('many-cheap', 200, 2), model('few-expensive', 2, 10)],
        loading: false,
        error: false
      },
      global
    })
    const rows = wrapper.findAll('tbody tr')
    expect(rows[0].find('[title]').attributes('title')).toBe('few-expensive')
    expect(rows[1].find('[title]').attributes('title')).toBe('many-cheap')
    expect(rows[0].get('.dashboard-model-track i').attributes('style')).toContain('100%')
    expect(rows[1].get('.dashboard-model-track i').attributes('style')).toContain('20%')
    expect(rows[0].findAll('td')[3].text()).toContain('✪ 10')
    expect(wrapper.text()).not.toContain('99')
    wrapper.unmount()
  })
  it('uses request count to break zero-spend ties while every spend bar stays zero', () => {
    const wrapper = mount(UserDashboardModels, {
      props: {
        models: [model('few-free', 2, 0), model('many-free', 200, 0)],
        loading: false,
        error: false
      },
      global
    })
    const rows = wrapper.findAll('tbody tr')
    expect(rows[0].find('[title]').attributes('title')).toBe('many-free')
    expect(rows[1].find('[title]').attributes('title')).toBe('few-free')
    for (const row of rows) {
      expect(row.get('.dashboard-model-track i').attributes('style')).toBe('width: 0%;')
      expect(row.findAll('td')[3].text()).toContain('✪ 0')
    }
    wrapper.unmount()
  })
})
