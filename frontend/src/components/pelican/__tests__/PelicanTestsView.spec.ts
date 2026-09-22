import { flushPromises, mount } from '@vue/test-utils'
import { nextTick, reactive } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import PelicanTestsView from '@/views/user/PelicanTestsView.vue'

const api = vi.hoisted(() => ({
  list: vi.fn(), history: vi.fn(), result: vi.fn(), plans: vi.fn(), groups: vi.fn(), models: vi.fn(), create: vi.fn(), update: vi.fn(), remove: vi.fn(), run: vi.fn(), resume: vi.fn(), cleanup: vi.fn(), settings: vi.fn(), saveSettings: vi.fn()
}))
const auth = reactive({ isAdmin: true, user: { id: 1 } })
const metadata = reactive({ displayName: '鹈鹕测试', load: vi.fn(() => Promise.resolve()), applySavedName: vi.fn((name: string) => { metadata.displayName = name }) })
const app = reactive({ siteName: 'Sub2API' })
vi.mock('@/api/pelicanTests', () => ({
  buildPelicanListParams: (group: number | '', account: string, page: number, pageSize: number) => ({ group_id: group || undefined, account_id: account ? Number(account) : undefined, page, page_size: pageSize }),
  listPelicanTests: api.list, listPelicanHistory: api.history, getPelicanResult: api.result,
  listPelicanPlans: api.plans, listPelicanPlanModels: api.models, createPelicanPlan: api.create, updatePelicanPlan: api.update, deletePelicanPlan: api.remove, runPelicanPlan: api.run, resumePelicanPlan: api.resume, cleanupPelicanPlan: api.cleanup,
  getPelicanTestSettings: api.settings, updatePelicanTestSettings: api.saveSettings
}))
vi.mock('@/api/admin/groups', () => ({ default: { getAll: api.groups }, getAll: api.groups }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => auth }))
vi.mock('@/stores/pelicanMetadata', () => ({ usePelicanMetadataStore: () => metadata }))
vi.mock('@/stores/app', () => ({ useAppStore: () => app }))

const item = (id: number, resultID = id) => ({ plan_id: 1, group_id: 3, group_name: '授权组', account_id: id, model_id: 'gpt-test', status: 'success', latency_ms: 100, char_count: 10000, min_chars: 9366, finished_at: '2026-09-15T00:00:00Z', history_count: 1, result_id: resultID, artwork_result_id: resultID })
const result = (id: number) => ({ id, plan_id: 1, group_id: 3, account_id: 27131, model_id: 'gpt-test', prompt_version: 'pelican-v1', status: 'success', latency_ms: 100, char_count: 10000, min_chars: 9366, started_at: '', finished_at: '', html: '<p>safe</p>' })
const stubs = { AppLayout: { template: '<div><slot /></div>' }, Icon: true, PelicanSandboxFrame: true, BaseDialog: { props: ['show'], template: '<div v-if="show" role="dialog"><button aria-label="close" @click="$emit(\'close\')">close</button><slot /></div>' } }

describe('PelicanTestsView', () => {
  beforeEach(() => {
    vi.clearAllMocks(); auth.isAdmin = true; auth.user = { id: 1 }
    api.list.mockResolvedValue({ items: [item(27131)], total: 7, page: 1, page_size: 6, groups: [{ id: 3, name: '授权组' }] })
    api.result.mockResolvedValue(result(27131)); api.history.mockResolvedValue([result(27131)])
    api.plans.mockResolvedValue([]); api.groups.mockResolvedValue([{ id: 3, name: '授权组', status: 'active' }]); api.models.mockResolvedValue(['gpt-kept']); api.create.mockResolvedValue({}); api.update.mockResolvedValue({}); api.remove.mockResolvedValue(undefined); api.run.mockResolvedValue({ queued: true }); api.resume.mockResolvedValue({}); api.cleanup.mockResolvedValue({ deleted_count: 2 }); api.settings.mockResolvedValue({ display_name: '鹈鹕测试', prompt: '默认提示词' }); api.saveSettings.mockResolvedValue({ display_name: '新页面名', prompt: '自定义提示词' }); metadata.displayName = '鹈鹕测试'
  })
  async function view() { const wrapper = mount(PelicanTestsView, { global: { stubs } }); await flushPromises(); return wrapper }

  it('keeps unsaved plans during profile refresh and pauses polling', async () => {
    vi.useFakeTimers()
    const wrapper = await view()
    try {
      await wrapper.get('select[aria-label="自动刷新间隔"]').setValue('15')
      await wrapper.findAll('button').find(button => button.text() === '管理测试计划')!.trigger('click'); await flushPromises()
      await wrapper.get('input[aria-label="间隔分钟"]').setValue('123')
      const calls = api.list.mock.calls.length
      auth.user = { id: 1 }; await flushPromises()
      await vi.advanceTimersByTimeAsync(30_000); await flushPromises()
      expect(wrapper.get('input[aria-label="间隔分钟"]').element).toHaveProperty('value', '123')
      expect(api.list).toHaveBeenCalledTimes(calls)
    } finally { wrapper.unmount(); vi.useRealTimers() }
  })

  it('preserves previews after network failures and new failed attempts', async () => {
    const wrapper = await view()
    const preview = wrapper.get('.artwork-preview').element
    const requests = api.result.mock.calls.length
    api.list.mockRejectedValueOnce(new Error('Network Error'))
    await wrapper.get('.refresh-button').trigger('click'); await flushPromises()
    expect(wrapper.get('.artwork-preview').element).toBe(preview)
    api.list.mockResolvedValueOnce({ items: [{ ...item(27131, 999), status: 'failed', artwork_result_id: 27131 }], total: 1, page: 1, page_size: 24, groups: [] })
    await wrapper.get('.refresh-button').trigger('click'); await flushPromises()
    expect(wrapper.get('.artwork-preview').element).toBe(preview)
    expect(api.result).toHaveBeenCalledTimes(requests)
    wrapper.unmount()
  })

  it('shows the recorded effort after the model without a label prefix', async () => {
    api.list.mockResolvedValue({ items: [{ ...item(1), reasoning_effort: 'high' }], total: 1, page: 1, page_size: 24, groups: [] })
    api.history.mockResolvedValue([{ ...result(1), reasoning_effort: 'low' }, { ...result(2), reasoning_effort: '' }, { ...result(3), reasoning_effort: null }])
    const wrapper = await view()
    expect(wrapper.get('.pelican-card .html-meta').text()).toBe('gpt-test · 高')
    await wrapper.get('.history-count').trigger('click'); await flushPromises()
    const history = wrapper.get('.history-grid')
    expect(history.text()).toContain('gpt-test · 低')
    expect(history.text()).toContain('gpt-test · 默认')
    expect(history.text()).toContain('gpt-test · 强度未记录')
    expect(history.text()).not.toContain('思考强度：')
    wrapper.unmount()
  })

  it('requests page two and resets to page one when filtering', async () => {
    api.list.mockImplementation(async (params: { page: number; page_size: number }) => ({ items: [item(27130 + params.page)], total: 12, page: params.page, page_size: params.page_size, groups: [{ id: 3, name: '授权组' }] }))
    const wrapper = await view()
    await wrapper.get('select[aria-label="每页条数"]').setValue('6')
    await wrapper.findAll('button').find(button => button.text() === '下一页')!.trigger('click')
    await flushPromises()
    expect(api.list).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2, page_size: 6 }), expect.any(AbortSignal))
    await wrapper.get('select[aria-label="筛选分组"]').setValue('3')
    await flushPromises()
    expect(api.list).toHaveBeenLastCalledWith(expect.objectContaining({ group_id: 3, account_id: undefined, page: 1, page_size: 6 }), expect.any(AbortSignal))
  })

  it('preserves cards and scroll position on automatic refresh, but resets both when the query changes', async () => {
    vi.useFakeTimers()
    const wrapper = await view()
    try {
      const region = wrapper.get('.results-region').element as HTMLElement
      Object.defineProperty(region, 'scrollTop', { configurable: true, writable: true, value: 84 })
      expect(wrapper.find('.artwork-preview').exists()).toBe(true)
      const artworkRequestsBeforeRefresh = api.result.mock.calls.length

      await wrapper.get('select[aria-label="自动刷新间隔"]').setValue('15')
      await vi.advanceTimersByTimeAsync(15_000); await flushPromises()
      expect(region.scrollTop).toBe(84)
      expect(wrapper.find('.artwork-preview').exists()).toBe(true)
      expect(api.result).toHaveBeenCalledTimes(artworkRequestsBeforeRefresh)

      api.list.mockResolvedValue({ items: [{ ...item(27131, 999), artwork_result_id: 888 }], total: 1, page: 1, page_size: 24, groups: [{ id: 3, name: '授权组' }] })
      await vi.advanceTimersByTimeAsync(15_000); await flushPromises()
      expect(region.scrollTop).toBe(84)
      expect(api.result).toHaveBeenLastCalledWith(888, expect.any(AbortSignal))

      await wrapper.get('select[aria-label="筛选分组"]').setValue('3'); await flushPromises()
      expect(region.scrollTop).toBe(0)
    } finally {
      wrapper.unmount()
      vi.useRealTimers()
    }
  })

  it('clears previously accessible artwork when a refresh fails', async () => {
    const wrapper = await view()
    expect(wrapper.find('.artwork-preview').exists()).toBe(true)
    api.list.mockRejectedValueOnce(Object.assign(new Error('权限已变更'), { status: 403 }))
    await wrapper.get('.refresh-button').trigger('click'); await flushPromises()
    expect(wrapper.find('.artwork-preview').exists()).toBe(false)
    expect(wrapper.text()).toContain('权限已变更')
  })

  it('keeps the history dialog closed when a delayed revision detail resolves', async () => {
    let resolve!: (value: ReturnType<typeof result>) => void
    const wrapper = await view()
    await wrapper.get('button[aria-label*="历史"]').trigger('click'); await flushPromises()
    expect(api.history).toHaveBeenLastCalledWith({ plan_id: 1 }, expect.any(AbortSignal))
    api.result.mockImplementationOnce(() => new Promise<ReturnType<typeof result>>(inner => { resolve = inner }))
    await wrapper.get('[role="dialog"] button:not([aria-label="close"])').trigger('click')
    await wrapper.get('[role="dialog"] button[aria-label="close"]').trigger('click')
    resolve(result(27131)); await flushPromises()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })

  it('drops an old list response after the authenticated user changes', async () => {
    let resolveOld!: (value: { items: ReturnType<typeof item>[]; total: number; page: number; page_size: number; groups: { id: number; name: string }[] }) => void
    const wrapper = await view()
    api.list.mockImplementationOnce(() => new Promise(inner => { resolveOld = inner }))
    await wrapper.get('.refresh-button').trigger('click')
    auth.user = { id: 2 }; await nextTick(); await flushPromises()
    resolveOld({ items: [item(99999)], total: 1, page: 1, page_size: 24, groups: [{ id: 3, name: '授权组' }] }); await flushPromises()
    expect(wrapper.text()).not.toContain('#99999')
  })

  it('keeps a failed admin form submission editable and supports history return', async () => {
    api.create.mockRejectedValueOnce(new Error('保存失败'))
    const wrapper = await view()
    await wrapper.findAll('.admin-toggle')[1].trigger('click'); await flushPromises()
    await wrapper.get('select[aria-label="分组"]').setValue('3')
    await flushPromises()
    await wrapper.get('select[aria-label="模型"]').setValue('gpt-kept')
    await wrapper.get('form').trigger('submit'); await flushPromises()
    expect(wrapper.text()).toContain('保存失败')
    expect((wrapper.get('select[aria-label="模型"]').element as HTMLSelectElement).value).toBe('gpt-kept')
    await wrapper.get('[role="dialog"] button[aria-label="close"]').trigger('click'); await flushPromises()
    await wrapper.get('button[aria-label*="历史"]').trigger('click'); await flushPromises()
    await wrapper.get('[role="dialog"] button:not([aria-label="close"])').trigger('click'); await flushPromises()
    expect(wrapper.text()).toContain('返回历史记录')
    await wrapper.get('.back-button').trigger('click')
    expect(wrapper.find('.history-grid').exists()).toBe(true)
  })

  it('keeps the settings draft after a failed save and propagates a successful name', async () => {
    const wrapper = await view()
    await wrapper.findAll('.admin-toggle')[0].trigger('click'); await flushPromises()
    await wrapper.get('input[aria-label="页面名称"]').setValue('保留名称')
    api.saveSettings.mockRejectedValueOnce(new Error('设置保存失败'))
    await wrapper.findAll('button').find(button => button.text() === '保存设置')!.trigger('click'); await flushPromises()
    expect((wrapper.get('input[aria-label="页面名称"]').element as HTMLInputElement).value).toBe('保留名称')
    expect(wrapper.text()).toContain('设置保存失败')
    await wrapper.findAll('button').find(button => button.text() === '保存设置')!.trigger('click'); await flushPromises()
    expect(metadata.displayName).toBe('新页面名')
  })

  it('keeps cards compact and shows Chinese skip reasons in history', async () => {
    api.list.mockResolvedValue({ items: [{ ...item(27131), status: 'skipped', error_message: 'account_scheduling_disabled' }, { ...item(27132), status: 'skipped', error_message: 'account unavailable' }, { ...item(27133), status: 'skipped', error_message: 'daily_call_limit_reached' }, { ...item(27134), status: 'skipped', error_message: 'quota_reservation_failed' }, { ...item(27135), status: 'skipped', error_message: 'model_unsupported' }, { ...item(27136), status: 'skipped', error_message: 'account_model_not_allowed' }, { ...item(27137), status: 'skipped', error_message: 'model_not_html_capable' }], total: 7, page: 1, page_size: 24, groups: [{ id: 3, name: '授权组' }] })
    const wrapper = await view()
    const card = wrapper.get('.pelican-card')
    expect(card.text()).toContain('gpt-test')
    expect(card.text()).toContain('测试于')
    expect(card.find('.result-row').exists()).toBe(false)
    expect(card.text()).not.toContain('阈值')
    expect(card.text()).not.toContain('账号已关闭调度')
    api.history.mockResolvedValue(['account_scheduling_disabled', 'account unavailable', 'daily_call_limit_reached', 'quota_reservation_failed', 'model_unsupported', 'account_model_not_allowed', 'model_not_html_capable'].map((reason, index) => ({ ...result(index + 1), status: 'skipped', error_message: reason })))
    await card.get('.history-count').trigger('click'); await flushPromises()
    expect(wrapper.text()).toContain('本轮已跳过（未执行）')
    expect(wrapper.text()).toContain('账号已关闭调度，已跳过测试')
    expect(wrapper.text()).toContain('账号当前不可用（历史记录未提供具体原因）')
    expect(wrapper.text()).toContain('今日调用次数已达到上限（按 UTC+8 日期重置）')
    expect(wrapper.text()).toContain('调用额度预留失败，已安全跳过本次测试')
    expect(wrapper.text()).toContain('本轮模型配置检查未通过，已跳过；请检查账号的模型配置')
    expect(wrapper.text()).toContain('该账号未配置此模型，已跳过')
    expect(wrapper.findAll('.history-record')).toHaveLength(6)
    expect(wrapper.find('.history-grid button').exists()).toBe(false)
    await wrapper.get('.history-pagination').findAll('button').find(button => button.text() === '下一页')!.trigger('click')
    expect(wrapper.text()).toContain('该账号映射后的模型不适合生成HTML，已跳过')
  })

  it('marks retained artwork as a previous success and opens a loaded preview', async () => {
    api.list.mockResolvedValue({ items: [{ ...item(27131), status: 'skipped', result_id: 101, artwork_result_id: 88 }], total: 1, page: 1, page_size: 24, groups: [{ id: 3, name: '授权组' }] })
    api.result.mockResolvedValue(result(88))
    const wrapper = await view()
    expect(wrapper.text()).toContain('上次成功作品')
    const preview = wrapper.get('.artwork-preview')
    await preview.trigger('click'); await flushPromises()
    expect(api.result).toHaveBeenLastCalledWith(88, expect.any(AbortSignal))
    expect(wrapper.text()).toContain('成功作品')
  })

  it('keeps plan controls hidden for ordinary users', async () => {
    auth.isAdmin = false
    const wrapper = await view()
    expect(wrapper.text()).not.toContain('管理测试计划')
    expect(api.plans).not.toHaveBeenCalled()
  })

  it('resumes only a paused plan and makes cleanup cancellable before confirming it', async () => {
    api.plans.mockResolvedValue([{ id: 7, group_id: 3, group_name: '授权组', model_id: 'gpt-kept', interval_minutes: 30, enabled: false, max_results: 20, min_chars: 9366, daily_call_limit: 5, failure_pause_threshold: 2, retention_days: 7, daily_calls_used: 2, usage_day: '2026-09-17', consecutive_failed_runs: 2, pause_reason: 'consecutive_failures', estimated_calls_per_run: 1, last_run_calls: 1, created_at: '', updated_at: '' }])
    const wrapper = await view()
    await wrapper.findAll('.admin-toggle')[1].trigger('click'); await flushPromises()
    await wrapper.findAll('button').find(button => button.text() === '恢复计划')!.trigger('click'); await flushPromises()
    expect(api.resume).toHaveBeenCalledWith(7)
    await wrapper.findAll('button').find(button => button.text() === '清理失败/跳过记录')!.trigger('click'); await flushPromises()
    expect(wrapper.text()).toContain('将永久删除该计划的失败和已跳过记录')
    await wrapper.findAll('button').find(button => button.text() === '取消')!.trigger('click'); await flushPromises()
    expect(api.cleanup).not.toHaveBeenCalled()
    await wrapper.findAll('button').find(button => button.text() === '按保留规则清理')!.trigger('click'); await flushPromises()
    await wrapper.findAll('button').find(button => button.text() === '确认清理')!.trigger('click'); await flushPromises()
    expect(api.cleanup).toHaveBeenCalledWith(7, 'expired')
    expect(wrapper.text()).toContain('已删除 2 条过期记录')
  })
})
