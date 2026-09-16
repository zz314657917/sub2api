import { flushPromises, mount } from '@vue/test-utils'
import { nextTick, reactive } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import PelicanTestsView from '@/views/user/PelicanTestsView.vue'

const api = vi.hoisted(() => ({
  list: vi.fn(), history: vi.fn(), result: vi.fn(), plans: vi.fn(), groups: vi.fn(), create: vi.fn(), update: vi.fn(), remove: vi.fn(), run: vi.fn()
}))
const auth = reactive({ isAdmin: true, user: { id: 1 } })
vi.mock('@/api/pelicanTests', () => ({
  buildPelicanListParams: (group: number | '', account: string, page: number, pageSize: number) => ({ group_id: group || undefined, account_id: account ? Number(account) : undefined, page, page_size: pageSize }),
  listPelicanTests: api.list, listPelicanHistory: api.history, getPelicanResult: api.result,
  listPelicanPlans: api.plans, createPelicanPlan: api.create, updatePelicanPlan: api.update, deletePelicanPlan: api.remove, runPelicanPlan: api.run
}))
vi.mock('@/api/admin/groups', () => ({ default: { getAll: api.groups }, getAll: api.groups }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => auth }))

const item = (id: number, resultID = id) => ({ plan_id: 1, group_id: 3, group_name: '授权组', account_id: id, model_id: 'gpt-test', status: 'success', latency_ms: 100, char_count: 10000, min_chars: 9366, finished_at: '2026-09-15T00:00:00Z', history_count: 1, result_id: resultID, artwork_result_id: resultID })
const result = (id: number) => ({ id, plan_id: 1, group_id: 3, account_id: 27131, model_id: 'gpt-test', prompt_version: 'pelican-v1', status: 'success', latency_ms: 100, char_count: 10000, min_chars: 9366, started_at: '', finished_at: '', html: '<p>safe</p>' })
const stubs = { AppLayout: { template: '<div><slot /></div>' }, Icon: true, PelicanSandboxFrame: true, BaseDialog: { template: '<div role="dialog"><button aria-label="close" @click="$emit(\'close\')">close</button><slot /></div>' } }

describe('PelicanTestsView', () => {
  beforeEach(() => {
    vi.clearAllMocks(); auth.isAdmin = true
    api.list.mockResolvedValue({ items: [item(27131)], total: 7, page: 1, page_size: 6, groups: [{ id: 3, name: '授权组' }] })
    api.result.mockResolvedValue(result(27131)); api.history.mockResolvedValue([result(27131)])
    api.plans.mockResolvedValue([]); api.groups.mockResolvedValue([{ id: 3, name: '授权组' }]); api.create.mockResolvedValue({}); api.update.mockResolvedValue({}); api.remove.mockResolvedValue(undefined); api.run.mockResolvedValue({ queued: true })
  })
  async function view() { const wrapper = mount(PelicanTestsView, { global: { stubs } }); await flushPromises(); return wrapper }

  it('requests page two and resets to page one when filtering', async () => {
    api.list.mockImplementation(async (params: { page: number; page_size: number }) => ({ items: [item(27130 + params.page)], total: 12, page: params.page, page_size: params.page_size, groups: [{ id: 3, name: '授权组' }] }))
    const wrapper = await view()
    await wrapper.get('select[aria-label="每页条数"]').setValue('6')
    await wrapper.findAll('button').find(button => button.text() === '下一页')!.trigger('click')
    await flushPromises()
    expect(api.list).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2, page_size: 6 }), expect.any(AbortSignal))
    await wrapper.get('input[aria-label="搜索账号 ID"]').setValue('27131')
    await flushPromises()
    expect(api.list).toHaveBeenLastCalledWith(expect.objectContaining({ account_id: 27131, page: 1, page_size: 6 }), expect.any(AbortSignal))
  })

  it('keeps the history dialog closed when a delayed revision detail resolves', async () => {
    let resolve!: (value: ReturnType<typeof result>) => void
    const wrapper = await view()
    await wrapper.get('button[aria-label*="历史"]').trigger('click'); await flushPromises()
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
    await wrapper.get('button[aria-label="刷新动画列表"]').trigger('click')
    auth.user = { id: 2 }; await nextTick(); await flushPromises()
    resolveOld({ items: [item(99999)], total: 1, page: 1, page_size: 24, groups: [{ id: 3, name: '授权组' }] }); await flushPromises()
    expect(wrapper.text()).not.toContain('#99999')
  })

  it('keeps a failed admin form submission editable and supports history return', async () => {
    api.create.mockRejectedValueOnce(new Error('保存失败'))
    const wrapper = await view()
    await wrapper.get('.admin-toggle').trigger('click'); await flushPromises()
    await wrapper.get('select[aria-label="分组"]').setValue('3')
    await wrapper.get('input[aria-label="模型 ID"]').setValue('gpt-kept')
    await wrapper.get('form').trigger('submit'); await flushPromises()
    expect(wrapper.text()).toContain('保存失败')
    expect((wrapper.get('input[aria-label="模型 ID"]').element as HTMLInputElement).value).toBe('gpt-kept')
    await wrapper.get('button[aria-label*="历史"]').trigger('click'); await flushPromises()
    await wrapper.get('[role="dialog"] button:not([aria-label="close"])').trigger('click'); await flushPromises()
    expect(wrapper.text()).toContain('返回历史记录')
    await wrapper.get('.back-button').trigger('click')
    expect(wrapper.text()).toContain('历史记录')
  })
})
