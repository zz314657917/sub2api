import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import PelicanPlansPanel from '../PelicanPlansPanel.vue'
import type { PelicanPlan, PelicanPlanAccount } from '@/api/pelicanTests'

const api = vi.hoisted(() => ({ models: vi.fn(), accounts: vi.fn(), get: vi.fn() }))
vi.mock('@/api/pelicanTests', () => ({ listPelicanPlanModels: api.models, listPelicanPlanAccounts: api.accounts }))
vi.mock('@/api/client', () => ({ apiClient: { get: api.get } }))
const groups = [{ id: 3, name: '分组 A' }, { id: 4, name: '分组 B' }]
const plan: PelicanPlan = { id: 7, group_id: 3, group_name: '分组 A', model_id: 'gpt-test', model_ids: ['gpt-test'], preferred_account_ids: [22, 11], reasoning_effort: '', interval_minutes: 60, enabled: false, max_results: 20, min_chars: 9366, daily_call_limit: 0, failure_pause_threshold: 0, retention_days: 0, daily_calls_used: 0, usage_day: null, consecutive_failed_runs: 0, pause_reason: '', estimated_calls_per_run: 1, last_run_calls: 0, created_at: '', updated_at: '' }
const accounts = [{ id: 11, name: '账号甲' }, { id: 22, name: '账号乙' }, { id: 33, name: '账号丙' }]
const wrappers: ReturnType<typeof mount>[] = []
function panel(value = plan) {
  const wrapper = mount(PelicanPlansPanel, { props: { plans: [value], groups, loading: false }, global: { stubs: { teleport: true } } })
  wrappers.push(wrapper)
  return wrapper
}
async function edit(wrapper: ReturnType<typeof panel>) {
  await wrapper.findAll('button').find(button => button.text() === '编辑')!.trigger('click')
  await flushPromises()
}
function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason: Error) => void
  const promise = new Promise<T>((yes, no) => { resolve = yes; reject = no })
  return { promise, resolve, reject }
}
const selected = (wrapper: ReturnType<typeof panel>) => wrapper.findAll('.preferred-accounts li span').map(row => row.text())

describe('Pelican preferred accounts', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    api.models.mockResolvedValue(['gpt-test'])
    api.accounts.mockResolvedValue(accounts)
  })
  afterEach(() => wrappers.splice(0).forEach(wrapper => wrapper.unmount()))

  it('uses the fixed account endpoint with group and abort signal', async () => {
    const actual = await vi.importActual<typeof import('@/api/pelicanTests')>('@/api/pelicanTests')
    api.get.mockResolvedValue({ data: { accounts } })
    const controller = new AbortController()
    expect(await actual.listPelicanPlanAccounts(3, controller.signal)).toEqual(accounts)
    expect(api.get).toHaveBeenCalledWith('/admin/pelican-test-plans/accounts', { params: { group_id: 3 }, signal: controller.signal })
  })

  it('round-trips saved IDs and appends, deduplicates, removes and renumbers without sorting', async () => {
    const wrapper = panel()
    await edit(wrapper)
    expect(selected(wrapper)).toEqual(['1. 账号乙 · #22', '2. 账号甲 · #11'])
    const select = wrapper.get('select[aria-label="追加优先账号"]')
    expect(select.find('option[value="22"]').attributes('disabled')).toBeDefined()
    await select.setValue('33')
    await select.setValue('33')
    expect((select.element as HTMLSelectElement).value).toBe('')
    await wrapper.get('button[aria-label="移除优先账号 #22"]').trigger('click')
    expect(selected(wrapper)).toEqual(['1. 账号甲 · #11', '2. 账号丙 · #33'])
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[0]).toEqual([7, expect.objectContaining({ preferred_account_ids: [11, 33] })])
    expect(plan.preferred_account_ids).toEqual([22, 11])
  })

  it('supports legacy automatic scheduling and explicit clearing', async () => {
    const wrapper = panel({ ...plan, preferred_account_ids: undefined })
    await edit(wrapper)
    expect(selected(wrapper)).toEqual([])
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[0]).toEqual([7, expect.objectContaining({ preferred_account_ids: [] })])
    await wrapper.get('select[aria-label="追加优先账号"]').setValue('11')
    await wrapper.get('button[aria-label="移除优先账号 #11"]').trigger('click')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[1]).toEqual([7, expect.objectContaining({ preferred_account_ids: [] })])
  })

  it('creates a disabled plan with accounts in dropdown selection order', async () => {
    const wrapper = panel()
    await wrapper.findAll('button').find(button => button.text() === '新建计划')!.trigger('click')
    await wrapper.get('select[aria-label="分组"]').setValue('3'); await flushPromises()
    await wrapper.get('input[aria-label="gpt-test"]').setValue(true)
    await wrapper.get('select[aria-label="追加优先账号"]').setValue('33')
    await wrapper.get('select[aria-label="追加优先账号"]').setValue('11')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[0]).toEqual([null, expect.objectContaining({ enabled: false, preferred_account_ids: [33, 11] })])
  })

  it('limits to 50 preferences and permits replacement after removal', async () => {
    api.accounts.mockResolvedValue(Array.from({ length: 51 }, (_, i) => ({ id: i + 1, name: `账号 ${i + 1}` })))
    const wrapper = panel({ ...plan, preferred_account_ids: Array.from({ length: 50 }, (_, i) => i + 1) })
    await edit(wrapper)
    expect(wrapper.get('select[aria-label="追加优先账号"]').attributes('disabled')).toBeDefined()
    await wrapper.get('button[aria-label="移除优先账号 #1"]').trigger('click')
    await wrapper.get('select[aria-label="追加优先账号"]').setValue('51')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[0]).toEqual([7, expect.objectContaining({ preferred_account_ids: Array.from({ length: 50 }, (_, i) => i + 2) })])
  })

  it('retains missing saved accounts and blocks saving until explicitly removed', async () => {
    api.accounts.mockResolvedValue([accounts[0]])
    const wrapper = panel()
    await edit(wrapper)
    expect(selected(wrapper)).toEqual(['1. 不可用账号 · #22', '2. 账号甲 · #11'])
    expect(wrapper.get('.editor-panel').text()).toContain('请移除')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')).toBeUndefined()
    await wrapper.get('button[aria-label="移除优先账号 #22"]').trigger('click')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[0]).toEqual([7, expect.objectContaining({ preferred_account_ids: [11] })])
  })

  it('keeps preference order on fetch failure and retries without losing the draft', async () => {
    api.accounts.mockRejectedValueOnce(new Error('账号目录暂不可用'))
    const wrapper = panel()
    await edit(wrapper)
    expect(wrapper.get('.editor-panel').text()).toContain('无法加载账号：账号目录暂不可用')
    expect(selected(wrapper)).toEqual(['1. 已保存账号 · #22', '2. 已保存账号 · #11'])
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')).toBeUndefined()
    await wrapper.findAll('button').find(button => button.text() === '重试账号加载')!.trigger('click')
    await flushPromises()
    expect(selected(wrapper)).toEqual(['1. 账号乙 · #22', '2. 账号甲 · #11'])
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('save')?.[0]).toEqual([7, expect.objectContaining({ preferred_account_ids: [22, 11] })])
  })

  it.each(['success', 'failure'])('ignores an old A response after A-B-A switching (%s)', async outcome => {
    const old = deferred<PelicanPlanAccount[]>()
    api.accounts.mockReturnValueOnce(old.promise)
    const wrapper = panel()
    await edit(wrapper)
    const oldSignal = api.accounts.mock.calls[0]![1] as AbortSignal
    await wrapper.get('select[aria-label="分组"]').setValue('4')
    await wrapper.get('select[aria-label="分组"]').setValue('3')
    await flushPromises()
    expect(oldSignal.aborted).toBe(true)
    expect(selected(wrapper)).toEqual([])
    await wrapper.get('select[aria-label="追加优先账号"]').setValue('33')
    if (outcome === 'success') old.resolve([{ id: 999, name: '旧账号' }])
    else old.reject(new Error('旧请求失败'))
    await flushPromises()
    expect(selected(wrapper)).toEqual(['1. 账号丙 · #33'])
    expect(wrapper.get('.editor-panel').text()).not.toContain('旧')
    expect(wrapper.findAll('.model-options input:checked')).toHaveLength(0)
  })

  it.each(['success', 'failure'])('ignores old account and model responses after same-group reopen (%s)', async outcome => {
    const oldAccounts = deferred<PelicanPlanAccount[]>()
    const oldModels = deferred<string[]>()
    api.accounts.mockReturnValueOnce(oldAccounts.promise)
    api.models.mockReturnValueOnce(oldModels.promise)
    const wrapper = panel()
    await edit(wrapper)
    const accountSignal = api.accounts.mock.calls[0]![1] as AbortSignal
    const modelSignal = api.models.mock.calls[0]![1] as AbortSignal
    await wrapper.findAll('button').find(button => button.text() === '取消')!.trigger('click')
    expect(accountSignal.aborted).toBe(true)
    expect(modelSignal.aborted).toBe(true)
    await edit(wrapper)
    if (outcome === 'success') { oldAccounts.resolve([]); oldModels.resolve(['old-model']) }
    else { oldAccounts.reject(new Error('旧账号失败')); oldModels.reject(new Error('旧模型失败')) }
    await flushPromises()
    expect(selected(wrapper)).toEqual(['1. 账号乙 · #22', '2. 账号甲 · #11'])
    expect(wrapper.get('.editor-panel').text()).not.toMatch(/旧|old-model/)
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })

  it('aborts both directories on unmount and safely ignores eventual rejection', async () => {
    const pendingAccounts = deferred<PelicanPlanAccount[]>()
    const pendingModels = deferred<string[]>()
    api.accounts.mockReturnValueOnce(pendingAccounts.promise); api.models.mockReturnValueOnce(pendingModels.promise)
    const wrapper = panel()
    await edit(wrapper)
    const accountSignal = api.accounts.mock.calls[0]![1] as AbortSignal
    const modelSignal = api.models.mock.calls[0]![1] as AbortSignal
    wrappers.splice(wrappers.indexOf(wrapper), 1)
    wrapper.unmount()
    expect(accountSignal.aborted).toBe(true); expect(modelSignal.aborted).toBe(true)
    pendingAccounts.reject(new Error('aborted')); pendingModels.reject(new Error('aborted'))
    await flushPromises()
    expect(wrapper.emitted('save')).toBeUndefined()
  })

  it('cancels without a save and reopens the original persisted order', async () => {
    const wrapper = panel()
    await edit(wrapper)
    await wrapper.get('button[aria-label="移除优先账号 #22"]').trigger('click')
    await wrapper.findAll('button').find(button => button.text() === '取消')!.trigger('click')
    expect(wrapper.find('form').exists()).toBe(false)
    expect(wrapper.emitted('save')).toBeUndefined()
    await edit(wrapper)
    expect(selected(wrapper)).toEqual(['1. 账号乙 · #22', '2. 账号甲 · #11'])
  })
})
