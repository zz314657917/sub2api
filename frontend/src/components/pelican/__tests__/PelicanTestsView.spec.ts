import { DOMWrapper, flushPromises, mount } from '@vue/test-utils'
import { nextTick, reactive } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import PelicanTestsView from '@/views/user/PelicanTestsView.vue'
import PelicanPlansPanel from '../PelicanPlansPanel.vue'

const api = vi.hoisted(() => ({
  list: vi.fn(), history: vi.fn(), result: vi.fn(), plans: vi.fn(), groups: vi.fn(), models: vi.fn(), accounts: vi.fn(), create: vi.fn(), update: vi.fn(), remove: vi.fn(), run: vi.fn(), resume: vi.fn(), cleanup: vi.fn(), settings: vi.fn(), saveSettings: vi.fn(), reference: vi.fn(), catalog: vi.fn()
}))
const auth = reactive({ isAdmin: true, user: { id: 1 } })
const metadata = reactive({ displayName: '鹈鹕测试', load: vi.fn(() => Promise.resolve()), applySavedName: vi.fn((name: string) => { metadata.displayName = name }) })
const app = reactive({ siteName: 'Sub2API' })
vi.mock('@/api/pelicanTests', () => ({
  buildPelicanListParams: (group: number | '', account: string, page: number, pageSize: number) => ({ group_id: group || undefined, account_id: account ? Number(account) : undefined, page, page_size: pageSize }),
  listPelicanTests: api.list, listPelicanHistory: api.history, getPelicanResult: api.result,
  listPelicanPlans: api.plans, listPelicanPlanModels: api.models, listPelicanPlanAccounts: api.accounts, createPelicanPlan: api.create, updatePelicanPlan: api.update, deletePelicanPlan: api.remove, runPelicanPlan: api.run, resumePelicanPlan: api.resume, cleanupPelicanPlan: api.cleanup,
  getPelicanTestSettings: api.settings, updatePelicanTestSettings: api.saveSettings,
  getPelicanReferenceScores: api.reference, getPelicanReferenceCatalog: api.catalog
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
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({ top: 100, bottom: 600, left: 0, right: 700, width: 700, height: 500, x: 0, y: 100, toJSON: () => ({}) })
    vi.clearAllMocks(); auth.isAdmin = true; auth.user = { id: 1 }
    api.accounts.mockResolvedValue([])
    api.reference.mockResolvedValue({ enabled: false, status: 'disabled', items: [] })
    api.catalog.mockResolvedValue({ models: [], status: 'unavailable', synced_at: null })
    api.list.mockResolvedValue({ items: [item(27131)], total: 7, page: 1, page_size: 6, groups: [{ id: 3, name: '授权组' }] })
    api.result.mockResolvedValue(result(27131)); api.history.mockResolvedValue([result(27131)])
    api.plans.mockResolvedValue([]); api.groups.mockResolvedValue([{ id: 3, name: '授权组', status: 'active' }]); api.models.mockResolvedValue(['gpt-kept']); api.create.mockResolvedValue({}); api.update.mockResolvedValue({}); api.remove.mockResolvedValue(undefined); api.run.mockResolvedValue({ queued: true }); api.resume.mockResolvedValue({}); api.cleanup.mockResolvedValue({ deleted_count: 2 }); api.settings.mockResolvedValue({ display_name: '鹈鹕测试', prompt: '默认提示词' }); api.saveSettings.mockResolvedValue({ display_name: '新页面名', prompt: '自定义提示词' }); metadata.displayName = '鹈鹕测试'
  })
  afterEach(() => vi.restoreAllMocks())
  async function view() { const wrapper = mount(PelicanTestsView, { global: { stubs } }); await flushPromises(); return wrapper }

  it('reuses route metadata on mount and refreshes at most once on returning visible without overlapping a pending list', async () => {
    vi.useFakeTimers()
    let hidden = false
    vi.spyOn(document, 'hidden', 'get').mockImplementation(() => hidden)
    const wrapper = await view()
    try {
      expect(metadata.load).toHaveBeenCalledWith()
      await wrapper.get('select[aria-label="自动刷新间隔"]').setValue('15')
      const calls = api.list.mock.calls.length
      hidden = true; document.dispatchEvent(new Event('visibilitychange'))
      await vi.advanceTimersByTimeAsync(60_000)
      expect(api.list).toHaveBeenCalledTimes(calls)
      let finish!: (value: unknown) => void
      api.list.mockReturnValueOnce(new Promise(resolve => { finish = resolve }))
      hidden = false; document.dispatchEvent(new Event('visibilitychange'))
      await flushPromises()
      expect(api.list).toHaveBeenCalledTimes(calls + 1)
      hidden = true; document.dispatchEvent(new Event('visibilitychange'))
      hidden = false; document.dispatchEvent(new Event('visibilitychange'))
      await vi.advanceTimersByTimeAsync(30_000)
      expect(api.list).toHaveBeenCalledTimes(calls + 1)
      finish({ items: [item(27131)], total: 1, page: 1, groups: [] })
      await flushPromises()
    } finally { wrapper.unmount(); vi.useRealTimers() }
  })

  it('does not request offscreen results and releases a removed result before reusing a newly authorized list', async () => {
    const geometry = vi.mocked(HTMLElement.prototype.getBoundingClientRect)
    geometry.mockImplementation(function () {
      return { top: this.classList.contains('lazy-artwork') ? 1400 : 100, bottom: this.classList.contains('lazy-artwork') ? 1620 : 600, left: 0, right: 700, width: 700, height: 500, x: 0, y: 100, toJSON: () => ({}) }
    })
    const wrapper = await view()
    try {
      expect(wrapper.text()).toContain('gpt-test')
      expect(api.result).not.toHaveBeenCalled()
      geometry.mockReturnValue({ top: 100, bottom: 600, left: 0, right: 700, width: 700, height: 500, x: 0, y: 100, toJSON: () => ({}) })
      await wrapper.get('.results-region').trigger('scroll'); await flushPromises()
      expect(api.result).toHaveBeenCalledTimes(1)
      api.list.mockResolvedValueOnce({ items: [], total: 0, page: 1, groups: [] })
      await wrapper.get('.refresh-button').trigger('click'); await flushPromises()
      expect(wrapper.find('.artwork-preview').exists()).toBe(false)
      await wrapper.get('.refresh-button').trigger('click'); await flushPromises()
      expect(api.result).toHaveBeenCalledTimes(2)
    } finally { wrapper.unmount() }
  })

  it('shows complete generation timestamps with seconds and timezone for success and fallback artwork', async () => {
    api.list.mockResolvedValue({
      items: [
        { ...item(1), finished_at: '2026-10-07T10:00:37+08:00' },
        { ...item(2), status: 'failed', finished_at: '2026-10-07T11:00:59+08:00', artwork_finished_at: '2026-10-07T10:00:12+08:00' },
        { ...item(3), status: 'skipped', artwork_result_id: null, finished_at: 'invalid' }
      ],
      total: 1, page: 1, page_size: 24, groups: []
    })
    const wrapper = await view()
    try {
      const timestamps = wrapper.findAll('.generation-time')
      expect(timestamps.map(node => node.attributes('title'))).toEqual([
        '2026-10-07T02:00:37.000Z (UTC)',
        '2026-10-07T02:00:12.000Z (UTC)',
        '时间未知'
      ])
      expect(timestamps[0]!.text()).not.toContain('(UTC)')
      expect(timestamps[1]!.attributes('title')).not.toContain('03:00:59')
    } finally { wrapper.unmount() }
  })

  it('stacks models from separate plans in one group and keeps model-scoped history after refresh', async () => {
    const entries = [
      { ...item(1), model_id: 'gpt-a' },
      { ...item(2), model_id: 'gpt-b' },
      { ...item(3), plan_id: 2, model_id: 'gpt-a' }
    ]
    api.list.mockResolvedValue({ items: entries, total: 1, page: 1, page_size: 24, groups: [] })
    const wrapper = await view()
    try {
      expect(wrapper.findAll('.pelican-card')).toHaveLength(1)
      expect(wrapper.findAll('.model-artwork')).toHaveLength(3)
      expect(wrapper.findAll('.model-heading strong').map(node => node.text())).toEqual(['gpt-a', 'gpt-b', 'gpt-a'])
      await wrapper.findAll('.history-count')[1]!.trigger('click'); await flushPromises()
      expect(api.history).toHaveBeenLastCalledWith({ plan_id: 1, model_id: 'gpt-b' }, expect.any(AbortSignal))
      api.list.mockResolvedValue({ items: entries.map(entry => ({ ...entry, result_id: entry.result_id + 100 })), total: 1, page: 1, page_size: 24, groups: [] })
      await wrapper.get('.refresh-button').trigger('click'); await flushPromises()
      expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
      expect(wrapper.get('[role="dialog"]').attributes('title')).toContain('gpt-b')
      api.list.mockResolvedValue({ items: entries.filter(entry => entry.model_id !== 'gpt-b'), total: 1, page: 1, page_size: 24, groups: [] })
      await wrapper.get('.refresh-button').trigger('click'); await flushPromises()
      expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    } finally { wrapper.unmount() }
  })

  it('keeps unsaved plans during profile refresh and pauses polling', async () => {
    vi.useFakeTimers()
    const wrapper = await view()
    try {
      await wrapper.get('select[aria-label="自动刷新间隔"]').setValue('15')
      await wrapper.findAll('button').find(button => button.text() === '管理测试计划')!.trigger('click'); await flushPromises()
      await wrapper.findAll('button').find(button => button.text() === '新建计划')!.trigger('click')
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

  it('keeps model headings and recorded effort in history without duplicate footer metadata', async () => {
    api.list.mockResolvedValue({ items: [{ ...item(1), reasoning_effort: 'high' }], total: 1, page: 1, page_size: 24, groups: [] })
    api.history.mockResolvedValue([{ ...result(1), reasoning_effort: 'low' }, { ...result(2), reasoning_effort: '' }, { ...result(3), reasoning_effort: null }])
    const wrapper = await view()
    expect(wrapper.get('.pelican-card .model-heading').text()).toContain('gpt-test')
    expect(wrapper.find('.pelican-card .html-meta').exists()).toBe(false)
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
    expect(api.history).toHaveBeenLastCalledWith({ plan_id: 1, model_id: 'gpt-test' }, expect.any(AbortSignal))
    api.result.mockImplementationOnce(() => new Promise<ReturnType<typeof result>>(inner => { resolve = inner }))
    await wrapper.get('[role="dialog"] button:not([aria-label="close"])').trigger('click')
    await wrapper.get('[role="dialog"] button[aria-label="close"]').trigger('click')
    resolve(result(27131)); await flushPromises()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })

  it.each([false, true])('restores the gallery opener after historical detail closes (return first: %s)', async (returnFirst) => {
    const host = document.createElement('div')
    document.body.append(host)
    const wrapper = mount(PelicanTestsView, { attachTo: host, global: { stubs } })
    try {
      await flushPromises()
      const opener = wrapper.get<HTMLButtonElement>('.history-count')
      opener.element.focus()
      await opener.trigger('click'); await flushPromises()
      const revision = wrapper.get<HTMLButtonElement>('.history-grid button')
      revision.element.focus()
      await revision.trigger('click'); await flushPromises()
      expect(wrapper.find('.back-button').exists()).toBe(true)
      expect(revision.element.isConnected).toBe(false)
      if (returnFirst) {
        await wrapper.get('.back-button').trigger('click'); await flushPromises()
        expect(wrapper.find('.history-grid').exists()).toBe(true)
      }
      await wrapper.get('[role="dialog"] button[aria-label="close"]').trigger('click'); await flushPromises()
      expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
      expect(document.activeElement).toBe(opener.element)
    } finally {
      wrapper.unmount()
      host.remove()
    }
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
    await wrapper.findAll('button').find(button => button.text() === '新建计划')!.trigger('click')
    await wrapper.get('select[aria-label="分组"]').setValue('3')
    await flushPromises()
    await wrapper.get('input[aria-label="gpt-kept"]').setValue(true)
    await wrapper.get('form').trigger('submit'); await flushPromises()
    expect(wrapper.text()).toContain('保存失败')
    expect((wrapper.get('input[aria-label="gpt-kept"]').element as HTMLInputElement).checked).toBe(true)
    await wrapper.get('.editor-panel').findAll('button').find(button => button.text() === '取消')!.trigger('click'); await flushPromises()
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

  it('defaults old settings without radar to off and keeps the original toolbar and gallery', async () => {
    const wrapper = await view()
    try {
      expect(wrapper.find('.reference-scores').exists()).toBe(false)
      expect(wrapper.find('.toolbar').exists()).toBe(true)
      expect(wrapper.findAll('.pelican-card')).toHaveLength(1)
      expect(api.catalog).not.toHaveBeenCalled()
      await wrapper.findAll('.admin-toggle')[0]!.trigger('click'); await flushPromises()
      expect(wrapper.get('input[aria-label="启用第三方参考评分"]').element).toHaveProperty('checked', false)
      expect(wrapper.findAll('.reference-selection-list li')).toHaveLength(0)
      expect(api.catalog).toHaveBeenCalledOnce()
    } finally { wrapper.unmount() }
  })

  it('preserves radar draft on save failure, preserves radar on restore defaults, and refreshes scores only after save succeeds', async () => {
    const radar = { enabled: true, selections: [{ model: 'unknown-custom-model', effort: 'ultra' }] }
    api.settings.mockResolvedValue({ display_name: '原页面', prompt: '原提示词', enabled: false, radar })
    api.saveSettings.mockResolvedValue({ display_name: '鹈鹕测试', prompt: '默认提示词', enabled: true, radar })
    api.reference.mockResolvedValue({ enabled: true, status: 'ready', items: [{ model: 'unknown-custom-model', display_name: 'unknown-custom-model', effort: 'ultra', status: 'not_recorded', score: null }] })
    const wrapper = await view()
    try {
      const section = wrapper.get('.reference-scores').element
      expect(section.compareDocumentPosition(wrapper.get('.toolbar').element) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
      await wrapper.findAll('.admin-toggle')[0]!.trigger('click'); await flushPromises()
      await wrapper.get('input[aria-label="参考模型 1"]').setValue('new-custom-model')
      const referenceCalls = api.reference.mock.calls.length
      api.saveSettings.mockRejectedValueOnce(new Error('设置保存失败'))
      await wrapper.findAll('button').find(button => button.text() === '保存设置')!.trigger('click'); await flushPromises()
      expect(wrapper.get('input[aria-label="参考模型 1"]').element).toHaveProperty('value', 'new-custom-model')
      expect(api.reference).toHaveBeenCalledTimes(referenceCalls)
      await wrapper.findAll('button').find(button => button.text() === '恢复默认')!.trigger('click')
      expect(wrapper.get('input[aria-label="页面名称"]').element).toHaveProperty('value', '鹈鹕测试')
      expect(wrapper.get('input[aria-label="参考模型 1"]').element).toHaveProperty('value', 'new-custom-model')
      expect(wrapper.get('input[aria-label="启用第三方参考评分"]').element).toHaveProperty('checked', true)
      await wrapper.findAll('button').find(button => button.text() === '保存设置')!.trigger('click'); await flushPromises()
      expect(api.saveSettings).toHaveBeenLastCalledWith(expect.objectContaining({ enabled: true, radar: { enabled: true, selections: [{ model: 'new-custom-model', effort: 'ultra' }] } }))
      expect(api.reference).toHaveBeenCalledTimes(referenceCalls + 1)
    } finally { wrapper.unmount() }
  })

  it('keeps ordinary results usable when reference requests fail and blocks late scores after role change', async () => {
    let resolveReference!: (value: unknown) => void
    api.reference.mockReturnValueOnce(new Promise(resolve => { resolveReference = resolve }))
    const wrapper = await view()
    try {
      auth.isAdmin = false
      await flushPromises()
      resolveReference({ enabled: true, status: 'ready', items: [{ model: 'late-private', effort: 'high', display_name: 'late-private', status: 'available', score: 100 }] })
      await flushPromises()
      expect(wrapper.text()).not.toContain('late-private')
      expect(api.catalog).not.toHaveBeenCalled()
      auth.isAdmin = true
      api.reference.mockRejectedValue(new Error('internal origin error'))
      await flushPromises()
      expect(wrapper.findAll('.pelican-card')).toHaveLength(1)
      expect(wrapper.text()).toContain('参考评分暂时不可用')
      expect(wrapper.text()).not.toContain('internal origin error')
    } finally { wrapper.unmount() }
  })

  it('keeps cards compact and shows Chinese skip reasons in history', async () => {
    api.list.mockResolvedValue({ items: [{ ...item(27131), status: 'skipped', error_message: 'account_scheduling_disabled' }, { ...item(27132), status: 'skipped', error_message: 'account unavailable' }, { ...item(27133), status: 'skipped', error_message: 'daily_call_limit_reached' }, { ...item(27134), status: 'skipped', error_message: 'quota_reservation_failed' }, { ...item(27135), status: 'skipped', error_message: 'model_unsupported' }, { ...item(27136), status: 'skipped', error_message: 'account_model_not_allowed' }, { ...item(27137), status: 'skipped', error_message: 'model_not_html_capable' }], total: 7, page: 1, page_size: 24, groups: [{ id: 3, name: '授权组' }] })
    const wrapper = await view()
    const card = wrapper.get('.pelican-card')
    expect(card.text()).toContain('gpt-test')
    expect(card.text()).not.toContain('测试于')
    expect(card.text()).not.toContain('本次耗时')
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

  const editablePlan = { id: 7, group_id: 3, group_name: '授权组', model_id: 'gpt-kept', preferred_account_ids: [22, 11], interval_minutes: 30, enabled: false, max_results: 20, min_chars: 9366, daily_call_limit: 0, failure_pause_threshold: 0, retention_days: 0, daily_calls_used: 0, usage_day: '', consecutive_failed_runs: 0, pause_reason: '', estimated_calls_per_run: 1, last_run_calls: 0, created_at: '', updated_at: '' }
  function deferred<T>() {
    let resolve!: (value: T) => void
    let reject!: (reason: Error) => void
    const promise = new Promise<T>((yes, no) => { resolve = yes; reject = no })
    return { promise, resolve, reject }
  }
  async function openPlanEditor(wrapper: ReturnType<typeof mount>) {
    await wrapper.findAll('button').find(button => button.text() === '管理测试计划')!.trigger('click')
    await flushPromises()
    expect(wrapper.find('.editor-panel').exists()).toBe(false)
    await wrapper.findAll('button').find(button => button.text() === '编辑')!.trigger('click')
    await flushPromises()
    expect(wrapper.find('.editor-panel').exists()).toBe(true)
  }
  function configureEditablePlan() {
    api.plans.mockResolvedValue([editablePlan])
    api.accounts.mockResolvedValue([{ id: 11, name: '账号甲' }, { id: 22, name: '账号乙' }])
  }

  it('opens the plan list on every visit and only creates or edits after an explicit action', async () => {
    const pending = deferred<object[]>()
    api.plans.mockReturnValueOnce(pending.promise)
    const wrapper = await view()
    try {
      const open = async () => {
        await wrapper.findAll('button').find(button => button.text() === '管理测试计划')!.trigger('click')
        await flushPromises()
      }
      await open()
      expect(wrapper.find('.editor-panel').exists()).toBe(false)
      expect(wrapper.text()).toContain('加载中')
      pending.resolve([]); await flushPromises()
      expect(wrapper.find('.editor-panel').exists()).toBe(false)
      expect(wrapper.text()).toContain('还没有测试计划')
      await wrapper.findAll('button').find(button => button.text() === '新建计划')!.trigger('click')
      expect(wrapper.get('button[type="submit"]').text()).toBe('新建计划')
      await wrapper.findAll('button').find(button => button.text() === '返回计划列表')!.trigger('click')
      await wrapper.get('[role="dialog"] > button[aria-label="close"]').trigger('click')
      configureEditablePlan()
      await open()
      expect(wrapper.find('.editor-panel').exists()).toBe(false)
      expect(wrapper.get('.plan-card').text()).toContain('授权组')
      await wrapper.findAll('button').find(button => button.text() === '编辑')!.trigger('click')
      await flushPromises()
      expect((wrapper.get('select[aria-label="当前计划"]').element as HTMLSelectElement).value).toBe('7')
      expect(wrapper.get('button[type="submit"]').text()).toBe('保存计划')
      expect(api.create).not.toHaveBeenCalled()
      expect(api.update).not.toHaveBeenCalled()
    } finally { wrapper.unmount() }
  })

  it('blocks duplicate submits and every close path during save, retaining the failed draft inside the editor', async () => {
    configureEditablePlan()
    const pending = deferred<object>()
    api.update.mockReturnValueOnce(pending.promise)
    const wrapper = await view()
    try {
      await openPlanEditor(wrapper)
      await wrapper.get('input[aria-label="间隔分钟"]').setValue('123')
      const panel = wrapper.getComponent(PelicanPlansPanel)
      // Two events before Vue propagates the busy prop exercise the parent's synchronous guard.
      panel.vm.$emit('save', 7, { ...editablePlan, preferred_account_ids: [22, 11] })
      panel.vm.$emit('save', 7, { ...editablePlan, preferred_account_ids: [11] })
      await nextTick()
      expect(api.update).toHaveBeenCalledTimes(1)
      expect(api.update).toHaveBeenCalledWith(7, expect.objectContaining({ preferred_account_ids: [22, 11] }))
      for (const close of wrapper.findAll('[role="dialog"] > button[aria-label="close"]')) await close.trigger('click')
      await wrapper.get('.editor-panel').findAll('button').find(button => button.text() === '取消')!.trigger('click')
      expect(wrapper.findAll('[role="dialog"]')).toHaveLength(2)
      expect(wrapper.get('.form-fields').attributes('disabled')).toBeDefined()
      pending.reject(new Error('计划保存失败'))
      await flushPromises()
      expect(wrapper.get('.editor-panel [role="alert"]').text()).toContain('计划保存失败')
      expect((wrapper.get('input[aria-label="间隔分钟"]').element as HTMLInputElement).value).toBe('123')
      expect(wrapper.findAll('.preferred-accounts li span').map(row => row.text())).toEqual(['1. 账号乙 · #22', '2. 账号甲 · #11'])
      expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
      await wrapper.get('form').trigger('submit'); await flushPromises()
      expect(api.update).toHaveBeenCalledTimes(2)
      expect(wrapper.find('.editor-panel').exists()).toBe(false)
    } finally { wrapper.unmount() }
  })

  it('closes the matching editor after save even if list refresh fails and shows refresh failure separately', async () => {
    configureEditablePlan()
    const wrapper = await view()
    try {
      await openPlanEditor(wrapper)
      api.plans.mockRejectedValueOnce(new Error('列表断线'))
      await wrapper.get('form').trigger('submit'); await flushPromises()
      expect(wrapper.find('.editor-panel').exists()).toBe(false)
      expect(wrapper.get('.pelican-plans [role="alert"]').text()).toBe('计划已保存，但列表刷新失败：列表断线')
      expect(wrapper.findAll('[role="dialog"]')).toHaveLength(1)
      await wrapper.findAll('button').find(button => button.text() === '编辑')!.trigger('click'); await flushPromises()
      expect(wrapper.find('.editor-panel [role="alert"]').exists()).toBe(false)
    } finally { wrapper.unmount() }
  })

  it.each(['success', 'failure'])('ignores old mutation callbacks after identity change and same-group reopen (%s)', async outcome => {
    configureEditablePlan()
    const pending = deferred<object>()
    api.update.mockReturnValueOnce(pending.promise)
    const wrapper = await view()
    try {
      await openPlanEditor(wrapper)
      await wrapper.get('form').trigger('submit')
      auth.user = { id: 2 }; await flushPromises()
      await openPlanEditor(wrapper)
      await wrapper.get('input[aria-label="间隔分钟"]').setValue('345')
      const requests = api.plans.mock.calls.length
      if (outcome === 'success') pending.resolve({})
      else pending.reject(new Error('旧保存失败'))
      await flushPromises()
      expect(api.plans).toHaveBeenCalledTimes(requests)
      expect(wrapper.find('.editor-panel').exists()).toBe(true)
      expect((wrapper.get('input[aria-label="间隔分钟"]').element as HTMLInputElement).value).toBe('345')
      expect(wrapper.text()).not.toContain('旧保存失败')
      expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
    } finally { wrapper.unmount() }
  })

  it.each(['success', 'failure'])('ignores stale plan-list callbacks after closing and reopening (%s)', async outcome => {
    const pending = deferred<object[]>()
    api.plans.mockReturnValueOnce(pending.promise)
    const wrapper = await view()
    try {
      await wrapper.findAll('button').find(button => button.text() === '管理测试计划')!.trigger('click')
      await flushPromises()
      await wrapper.get('[role="dialog"] > button[aria-label="close"]').trigger('click')
      configureEditablePlan()
      await openPlanEditor(wrapper)
      if (outcome === 'success') pending.resolve([{ ...editablePlan, group_name: '旧列表' }])
      else pending.reject(new Error('旧列表失败'))
      await flushPromises()
      expect(wrapper.text()).not.toContain('旧列表')
      expect(wrapper.get('.plan-card').text()).toContain('授权组')
      expect(wrapper.find('.editor-panel').exists()).toBe(true)
    } finally { wrapper.unmount() }
  })

  it('ignores a mutation that resolves after view unmount', async () => {
    configureEditablePlan()
    const pending = deferred<object>()
    api.update.mockReturnValueOnce(pending.promise)
    const wrapper = await view()
    await openPlanEditor(wrapper)
    await wrapper.get('form').trigger('submit')
    const requests = api.plans.mock.calls.length
    wrapper.unmount()
    pending.resolve({}); await flushPromises()
    expect(api.plans).toHaveBeenCalledTimes(requests)
  })

  it('does not let an old failed save clear the busy state or error of a newer session', async () => {
    configureEditablePlan()
    const oldSave = deferred<object>()
    const newSave = deferred<object>()
    api.update.mockReturnValueOnce(oldSave.promise).mockReturnValueOnce(newSave.promise)
    const wrapper = await view()
    try {
      await openPlanEditor(wrapper)
      await wrapper.get('form').trigger('submit')
      auth.user = { id: 2 }; await flushPromises()
      await openPlanEditor(wrapper)
      await wrapper.get('input[aria-label="间隔分钟"]').setValue('234')
      await wrapper.get('form').trigger('submit')
      oldSave.reject(new Error('旧会话失败')); await flushPromises()
      expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
      expect(wrapper.find('.editor-panel [role="alert"]').exists()).toBe(false)
      await wrapper.get('[role="dialog"] > button[aria-label="close"]').trigger('click')
      expect(wrapper.findAll('[role="dialog"]')).toHaveLength(2)
      newSave.reject(new Error('本次保存失败')); await flushPromises()
      expect(wrapper.get('.editor-panel [role="alert"]').text()).toBe('本次保存失败')
      expect((wrapper.get('input[aria-label="间隔分钟"]').element as HTMLInputElement).value).toBe('234')
      expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
    } finally { wrapper.unmount() }
  })

  it('uses real nested BaseDialogs for Escape, z-index, focus return and the surviving body scroll lock', async () => {
    configureEditablePlan()
    const pending = deferred<object>()
    api.update.mockReturnValueOnce(pending.promise)
    const host = document.createElement('div')
    document.body.append(host)
    const wrapper = mount(PelicanTestsView, { attachTo: host, global: { stubs: { ...stubs, BaseDialog: false, teleport: false } } })
    try {
      await flushPromises()
      const outerOpener = wrapper.findAll('button').find(button => button.text() === '管理测试计划')!
      outerOpener.element.focus()
      await outerOpener.trigger('click'); await flushPromises()
      const body = new DOMWrapper(document.body)
      expect(body.findAll('.modal-overlay')).toHaveLength(1)
      expect(body.find('.editor-panel').exists()).toBe(false)
      const outerBody = body.get('.modal-body').element
      Object.defineProperty(outerBody, 'scrollTop', { configurable: true, writable: true, value: 160 })
      const editButton = body.findAll('button').find(button => button.text() === '编辑')!
      editButton.element.focus()
      await editButton.trigger('click'); await flushPromises()
      expect(body.findAll('.modal-overlay')).toHaveLength(2)
      expect((body.get('select[aria-label="当前计划"]').element as HTMLSelectElement).value).toBe('7')
      expect(body.findAll('.modal-overlay')[1]!.attributes('style')).toContain('z-index: 60')
      expect(body.get('.editor-panel').classes()).toContain('pelican-plans')
      expect(outerBody.scrollTop).toBe(160)
      await body.findAll('.modal-header button')[0]!.trigger('click')
      expect(body.findAll('.modal-overlay')).toHaveLength(2)
      await body.get('form').trigger('submit')
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
      await body.findAll('.modal-header button')[1]!.trigger('click')
      await flushPromises()
      expect(body.findAll('.modal-overlay')).toHaveLength(2)
      pending.reject(new Error('保存被拒绝')); await flushPromises()
      expect(body.get('.editor-panel [role="alert"]').text()).toContain('保存被拒绝')
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
      await flushPromises()
      expect(body.findAll('.modal-overlay')).toHaveLength(1)
      expect(body.find('.editor-panel').exists()).toBe(false)
      expect(document.body.classList.contains('modal-open')).toBe(true)
      expect(document.activeElement).toBe(editButton.element)
      expect(outerBody.scrollTop).toBe(160)
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
      await flushPromises()
      expect(body.findAll('.modal-overlay')).toHaveLength(0)
      expect(document.body.classList.contains('modal-open')).toBe(false)
      expect(document.activeElement).toBe(outerOpener.element)
    } finally { wrapper.unmount(); host.remove() }
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
