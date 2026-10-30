import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import PelicanReferenceScores from '../PelicanReferenceScores.vue'
import type { PelicanReferenceScores as Scores } from '@/api/pelicanTests'

function scores(overrides: Partial<Scores> = {}): Scores {
  return {
    enabled: true, status: 'ready', source_url: 'https://codexradar.com/#stationmaster-round',
    test_started_at: '2026-10-29T16:08:11+08:00', test_completed_at: '2026-10-29T20:00:24+08:00',
    source_snapshot_at: '2026-10-29T20:00:25+08:00', synced_at: '2026-10-30T08:00:00Z',
    round_id: 'codex-four-261-10c-20261029', score_max: 150,
    items: [{ model: 'gpt-6.1-sol', display_name: 'GPT-6.1 Sol', effort: 'low', status: 'available', score: 106.165760869565 }],
    ...overrides
  }
}
describe('PelicanReferenceScores', () => {
  it('marks each model maximum using exact scores, includes ties and excludes unavailable entries', async () => {
    const wrapper = mount(PelicanReferenceScores, { props: { data: scores({ items: [
      { model: 'a', display_name: 'A', effort: 'low', status: 'available', score: 100.01 },
      { model: 'a', display_name: 'A', effort: 'medium', status: 'available', score: 100.04 },
      { model: 'a', display_name: 'A', effort: 'high', status: 'available', score: 100.04 },
      { model: 'a', display_name: 'A', effort: 'ultra', status: 'not_recorded', score: 150 },
      { model: 'b', display_name: 'B', effort: 'low', status: 'available', score: 0 },
      { model: 'c', display_name: 'C', effort: 'max', status: 'unavailable', score: null }
    ] }) } })
    expect(wrapper.findAll('.reference-effort-best').map(row => row.attributes('aria-label'))).toEqual(['A · medium', 'A · high', 'B · low'])
    expect(wrapper.findAll('.reference-best')).toHaveLength(3)
    expect(wrapper.get('.reference-best').attributes('aria-label')).toBe('本模型已展示档位最高分')
    expect(wrapper.findAll('.reference-card')[2]!.find('.reference-best').exists()).toBe(false)
    await wrapper.setProps({ data: scores({ score_max: null }) })
    expect(wrapper.find('.reference-best').exists()).toBe(false)
  })

  it('groups interleaved exact model IDs once and preserves each selected effort, score and status in order', () => {
    const data = scores({ items: [
      { model: 'a', display_name: 'Model A', effort: 'low', status: 'available', score: 19.04 },
      { model: 'b', display_name: 'Model B', effort: 'max', status: 'unsupported', score: null },
      { model: 'a', display_name: 'Model A', effort: 'high', status: 'available', score: 123.46 },
      { model: 'a', display_name: 'Model A', effort: 'ultra', status: 'not_recorded', score: null },
      { model: 'b', display_name: 'Model B', effort: 'low', status: 'unavailable', score: null },
      { model: 'A', display_name: 'Distinct A', effort: 'high', status: 'available', score: 55 }
    ] })
    const before = structuredClone(data)
    const wrapper = mount(PelicanReferenceScores, { props: { data } })
    const cards = wrapper.findAll('.reference-card')
    expect(cards.map(card => card.get('.reference-model strong').text())).toEqual(['Model A', 'Model B', 'Distinct A'])
    expect(wrapper.findAllComponents({ name: 'ModelIcon' })).toHaveLength(3)
    expect(cards[0]!.findAll('.reference-effort-row').map(row => row.text())).toEqual(['low19', 'high123.5', 'ultra来源未收录此组合'])
    expect(cards[1]!.findAll('.reference-effort-row').map(row => row.text())).toEqual(['max来源不支持此档位', 'low参考评分暂不可用'])
    expect(cards[2]!.get('.reference-effort-row').text()).toBe('high55')
    expect(wrapper.findAll('.reference-effort-row')).toHaveLength(data.items.length)
    expect(data).toEqual(before)
  })

  it('shows exact selected model/effort, display-only rounded score, real interval, batch and sync metadata', () => {
    const data = scores()
    const wrapper = mount(PelicanReferenceScores, { props: { data } })
    expect(wrapper.text()).toContain('第三方 Codex 订阅测试，不代表本站分组或账号的实测成绩。')
    expect(wrapper.get('.reference-model').text()).toContain('GPT-6.1 Sol')
    expect(wrapper.get('.reference-effort').text()).toBe('low')
    expect(wrapper.get('.reference-value').text()).toBe('106.2')
    expect(wrapper.get('.reference-heading .reference-max').text()).toBe('满分 150')
    expect(wrapper.findAll('.reference-max')).toHaveLength(1)
    expect(wrapper.get('.reference-value').element.firstElementChild?.classList.contains('reference-best')).toBe(true)
    expect(wrapper.get('.reference-best').element.nextElementSibling?.tagName).toBe('STRONG')
    expect(data.items[0]!.score).toBe(106.165760869565)
    expect(wrapper.get('.reference-metadata').text()).toContain('2026/10/29 16:08 (UTC+8)')
    expect(wrapper.text()).toContain('来源批次：codex-four-261-10c-20261029')
    expect(wrapper.text()).toContain('最近成功同步：2026/10/30 16:00 (UTC+8)')
    expect(wrapper.text()).not.toContain('轮次数')
    expect(wrapper.get('.reference-strip').attributes()).toMatchObject({ tabindex: '0', role: 'region', 'aria-label': '第三方模型参考评分，可横向滚动' })
    expect(wrapper.findComponent({ name: 'ModelIcon' }).exists()).toBe(true)
  })

  it('uses the fixed safe external source link even if DTO carries an unsafe URL', () => {
    const wrapper = mount(PelicanReferenceScores, { props: { data: scores({ source_url: 'javascript:alert(1)' }) } })
    expect(wrapper.get('a').attributes()).toMatchObject({ href: 'https://codexradar.com/#stationmaster-round', target: '_blank', rel: 'noopener noreferrer' })
  })

  it('preserves configured order and distinguishes missing, unsupported and unavailable without zero scores', () => {
    const wrapper = mount(PelicanReferenceScores, { props: { data: scores({ items: [
      { model: 'custom', display_name: 'Custom', effort: 'ultra', status: 'not_recorded', score: null },
      { model: 'gpt-6-luna', display_name: 'Luna', effort: 'ultra', status: 'unsupported', score: null },
      { model: 'gpt-6.1-sol', display_name: 'Sol', effort: 'max', status: 'unavailable', score: null }
    ] }) } })
    expect(wrapper.findAll('.reference-model strong').map(node => node.text())).toEqual(['Custom', 'Luna', 'Sol'])
    expect(wrapper.findAll('.reference-unrecorded').map(node => node.text())).toEqual(['来源未收录此组合', '来源不支持此档位', '参考评分暂不可用'])
    expect(wrapper.find('.reference-value').exists()).toBe(false)
  })

  it('handles disabled, empty, loading, unavailable and stale states independently', async () => {
    const wrapper = mount(PelicanReferenceScores, { props: { loading: true } })
    expect(wrapper.text()).toContain('正在加载参考评分')
    await wrapper.setProps({ loading: false, data: scores({ enabled: false, status: 'disabled', items: [] }) })
    expect(wrapper.find('section').exists()).toBe(false)
    await wrapper.setProps({ data: scores({ items: [] }) })
    expect(wrapper.text()).toContain('管理员尚未选择')
    await wrapper.setProps({ data: scores({ status: 'unavailable', synced_at: null, items: [] }) })
    expect(wrapper.text()).toContain('参考评分暂时不可用')
    expect(wrapper.find('.reference-metadata').exists()).toBe(false)
    await wrapper.setProps({ data: scores({ status: 'stale' }) })
    expect(wrapper.text()).toContain('保留上次成功同步')
    expect(wrapper.get('.reference-value').text()).toBe('106.2')
    await wrapper.setProps({ data: scores(), failed: true })
    expect(wrapper.text()).toContain('来源更新延迟')
  })

  it('uses source-defined max and never fabricates a score for missing totals', async () => {
    const wrapper = mount(PelicanReferenceScores, { props: { data: scores({ score_max: 200 }) } })
    expect(wrapper.get('.reference-heading .reference-max').text()).toBe('满分 200')
    expect(wrapper.get('.reference-value').text()).toBe('106.2')
    expect(wrapper.findAll('.reference-effort-row').every(row => !row.text().includes('/'))).toBe(true)
    await wrapper.setProps({ data: scores({ score_max: null }) })
    expect(wrapper.find('.reference-max').exists()).toBe(false)
    expect(wrapper.find('.reference-value').exists()).toBe(false)
  })
})
