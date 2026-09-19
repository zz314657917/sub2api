import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import PelicanPlansPanel from '../PelicanPlansPanel.vue'

const models = vi.hoisted(() => vi.fn())
vi.mock('@/api/pelicanTests', () => ({ listPelicanPlanModels: models }))

const groups = [{ id: 3, name: '授权分组' }, { id: 4, name: '备用分组' }]
const plan = { id: 7, group_id: 3, group_name: '授权分组', model_id: 'gpt-test', interval_minutes: 30, enabled: false, max_results: 20, min_chars: 9366, daily_call_limit: 0, failure_pause_threshold: 0, retention_days: 0, daily_calls_used: 0, usage_day: '2026-09-17', consecutive_failed_runs: 0, pause_reason: '', estimated_calls_per_run: 1, last_run_calls: 0, created_at: '', updated_at: '' }
const wrappers: ReturnType<typeof mount>[] = []
const mountPanel = () => { const wrapper = mount(PelicanPlansPanel, { props: { loading: false, groups, plans: [plan] } }); wrappers.push(wrapper); return wrapper }

describe('PelicanPlansPanel', () => {
  beforeEach(() => models.mockReset())
  afterEach(() => { wrappers.splice(0).forEach(wrapper => wrapper.unmount()) })
  it('renders an actual plan and only submits a selected configured model', async () => {
    models.mockResolvedValue(['gpt-test'])
    const wrapper = mountPanel()
    expect(wrapper.text()).toContain('授权分组 · gpt-test')
    await wrapper.findAll('button').find(button => button.text() === '立即运行')!.trigger('click')
    expect(wrapper.emitted('run')?.[0]).toEqual([7])
    await wrapper.get('select[aria-label="分组"]').setValue('3'); await flushPromises()
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
    await wrapper.get('select[aria-label="模型"]').setValue('gpt-test')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[0]).toEqual([null, expect.objectContaining({ group_id: 3, model_id: 'gpt-test', enabled: false, interval_minutes: 60, max_results: 20, min_chars: 9366, daily_call_limit: 0, failure_pause_threshold: 0, retention_days: 0 })])
    expect(wrapper.text()).toContain('每日用量按 UTC+8 重置')
    expect(wrapper.text()).toContain('每轮由分组调度选择一个账号，最多测试一次')
    expect((wrapper.get('select[aria-label="思考强度"]').element as HTMLSelectElement).value).toBe('')
  })

  it('saves a custom timeout and restores it when editing', async () => {
    models.mockResolvedValue(['gpt-test'])
    const wrapper = mountPanel()
    expect((wrapper.get('input[aria-label="测试超时秒"]').element as HTMLInputElement).value).toBe('180')
    await wrapper.get('select[aria-label="分组"]').setValue('3'); await flushPromises()
    await wrapper.get('select[aria-label="模型"]').setValue('gpt-test')
    await wrapper.get('input[aria-label="测试超时秒"]').setValue('1800')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[0]).toEqual([null, expect.objectContaining({ timeout_seconds: 1800 })])
    await wrapper.setProps({ plans: [{ ...plan, timeout_seconds: 3600 }] })
    await wrapper.findAll('button').find(button => button.text() === '编辑')!.trigger('click'); await flushPromises()
    expect((wrapper.get('input[aria-label="测试超时秒"]').element as HTMLInputElement).value).toBe('3600')
  })

  it('round-trips the selected reasoning effort through create and edit', async () => {
    models.mockResolvedValue(['gpt-test'])
    const wrapper = mountPanel()
    await wrapper.get('select[aria-label="分组"]').setValue('3'); await flushPromises()
    await wrapper.get('select[aria-label="模型"]').setValue('gpt-test')
    await wrapper.get('select[aria-label="思考强度"]').setValue('xhigh')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[0]).toEqual([null, expect.objectContaining({ reasoning_effort: 'xhigh' })])

    await wrapper.setProps({ plans: [{ ...plan, reasoning_effort: 'minimal' }] })
    await wrapper.findAll('button').find(button => button.text() === '编辑')!.trigger('click'); await flushPromises()
    expect((wrapper.get('select[aria-label="思考强度"]').element as HTMLSelectElement).value).toBe('minimal')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[1]).toEqual([7, expect.objectContaining({ reasoning_effort: 'minimal' })])
  })

  it('includes cost and retention settings in edited plan payloads without resuming an auto-paused plan', async () => {
    models.mockResolvedValue(['gpt-test'])
    const wrapper = mountPanel()
    await wrapper.findAll('button').find(button => button.text() === '编辑')!.trigger('click'); await flushPromises()
    await wrapper.get('input[aria-label="每日调用上限"]').setValue('50')
    await wrapper.get('input[aria-label="连续失败暂停阈值"]').setValue('3')
    await wrapper.get('select[aria-label="按天保留"]').setValue('30')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[0]).toEqual([7, expect.objectContaining({ daily_call_limit: 50, failure_pause_threshold: 3, retention_days: 30 })])
    await wrapper.setProps({ plans: [{ ...plan, pause_reason: 'consecutive_failures', enabled: false }] })
    await wrapper.findAll('button').find(button => button.text() === '编辑')!.trigger('click'); await flushPromises()
    expect(wrapper.text()).toContain('请使用下方“恢复计划”')
    expect(wrapper.get('input[type="checkbox"]').attributes('disabled')).toBeDefined()
  })

  it('emits explicit resume and named cleanup scopes', async () => {
    const wrapper = mountPanel()
    await wrapper.setProps({ plans: [{ ...plan, pause_reason: 'consecutive_failures' }] })
    await wrapper.findAll('button').find(button => button.text() === '恢复计划')!.trigger('click')
    await wrapper.findAll('button').find(button => button.text() === '清理失败/跳过记录')!.trigger('click')
    await wrapper.findAll('button').find(button => button.text() === '按保留规则清理')!.trigger('click')
    expect(wrapper.emitted('resume')?.[0]).toEqual([7])
    expect(wrapper.emitted('cleanup-request')).toEqual([[expect.objectContaining({ id: 7 }), 'failed'], [expect.objectContaining({ id: 7 }), 'expired']])
  })

  it('ignores stale group model responses and clears the previous selection immediately', async () => {
    models.mockImplementation((groupID: number) => new Promise<string[]>(resolve => setTimeout(() => resolve(groupID === 3 ? ['gpt-three'] : ['gpt-four']), groupID === 3 ? 20 : 0)))
    const wrapper = mountPanel()
    await wrapper.get('select[aria-label="分组"]').setValue('3')
    await wrapper.get('select[aria-label="分组"]').setValue('4')
    await new Promise(resolve => setTimeout(resolve, 30)); await flushPromises()
    expect((wrapper.get('select[aria-label="分组"]').element as HTMLSelectElement).value).toBe('4')
    expect(wrapper.text()).toContain('gpt-four')
    expect(wrapper.text()).not.toContain('gpt-three')
    expect((wrapper.get('select[aria-label="模型"]').element as HTMLSelectElement).value).toBe('')
  })

  it('keeps a valid edited model, and makes an invalid edited model require reselection', async () => {
    models.mockResolvedValueOnce(['gpt-test']).mockResolvedValueOnce(['gpt-new'])
    const wrapper = mountPanel()
    await wrapper.findAll('button').find(button => button.text() === '编辑')!.trigger('click'); await flushPromises()
    expect((wrapper.get('select[aria-label="模型"]').element as HTMLSelectElement).value).toBe('gpt-test')
    await wrapper.setProps({ plans: [{ ...plan, id: 8, model_id: 'gpt-old' }] })
    await wrapper.findAll('button').find(button => button.text() === '编辑')!.trigger('click'); await flushPromises()
    expect((wrapper.get('select[aria-label="模型"]').element as HTMLSelectElement).value).toBe('')
    expect(wrapper.text()).toContain('原模型已不在该分组的可选模型中，请重新选择。')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('shows a retryable model-load error and blocks submission', async () => {
    models.mockRejectedValueOnce(new Error('目录暂不可用')).mockResolvedValueOnce(['gpt-retry'])
    const wrapper = mountPanel()
    await wrapper.get('select[aria-label="分组"]').setValue('3'); await flushPromises()
    expect(wrapper.text()).toContain('无法加载模型：目录暂不可用')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
    await wrapper.findAll('button').find(button => button.text() === '重试')!.trigger('click'); await flushPromises()
    expect(wrapper.text()).toContain('gpt-retry')
  })
})
