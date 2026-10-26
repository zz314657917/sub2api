import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { DOMWrapper, enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import ScheduledTestsPanel from '../ScheduledTestsPanel.vue'

const api = vi.hoisted(() => ({
  listByAccount: vi.fn(), listResults: vi.fn(), create: vi.fn(), update: vi.fn(),
  showError: vi.fn(), showSuccess: vi.fn()
}))
vi.mock('@/api/admin', () => ({ adminAPI: { scheduledTests: api } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => api }))
vi.mock('vue-i18n', async original => ({
  ...await original<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key })
}))
enableAutoUnmount(afterEach)
const plan = (id: number) => ({
  id, account_id: 1, model_id: `model-${id}`, cron_expression: `*/${id} * * * *`,
  max_results: id * 10, enabled: true, auto_recover: id === 1
})
function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason: Error) => void
  const promise = new Promise<T>((yes, no) => { resolve = yes; reject = no })
  return { promise, resolve, reject }
}
async function openPanel() {
  const wrapper = mount(ScheduledTestsPanel, {
    props: { show: false, accountId: 1, modelOptions: [1, 2].map(id => ({ value: `model-${id}`, label: `model-${id}` })) },
    global: { stubs: { teleport: true, Icon: true } }
  })
  await wrapper.setProps({ show: true })
  await flushPromises()
  return wrapper
}

describe('scheduled plan editor entry', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    api.listByAccount.mockResolvedValue([plan(1), plan(2)])
    api.update.mockImplementation(async id => plan(id))
    api.create.mockResolvedValue(plan(3))
  })

  it('opens the first plan immediately and saves another selected plan without clicking edit', async () => {
    const wrapper = await openPanel()
    expect(wrapper.findAll('[role="dialog"]')).toHaveLength(2)
    expect((wrapper.get('select[aria-label="当前定时计划"]').element as HTMLSelectElement).value).toBe('1')
    await wrapper.get('select').setValue('2')
    expect((wrapper.get('input[placeholder="*/30 * * * *"]').element as HTMLInputElement).value).toBe('*/2 * * * *')
    await wrapper.get('input[placeholder="*/30 * * * *"]').setValue('0 9 * * *')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(api.update).toHaveBeenCalledWith(2, expect.objectContaining({ model_id: 'model-2', cron_expression: '0 9 * * *', max_results: 20 }))
    expect(wrapper.findAll('[role="dialog"]')).toHaveLength(1)
  })

  it('opens creation directly when no plan exists and can also create from the plan selector', async () => {
    api.listByAccount.mockResolvedValueOnce([])
    const wrapper = await openPanel()
    expect(wrapper.findAll('[role="dialog"]')[1].text()).toContain('admin.scheduledTests.addPlan')
    const modelPicker = wrapper.findAllComponents({ name: 'Select' })[0]
    modelPicker.vm.$emit('update:modelValue', 'model-1')
    await wrapper.get('input[placeholder="*/30 * * * *"]').setValue('0 * * * *')
    await wrapper.get('form').trigger('submit'); await flushPromises()
    expect(api.create).toHaveBeenCalledWith(expect.objectContaining({ account_id: 1, model_id: 'model-1', cron_expression: '0 * * * *' }))
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true }); await flushPromises()
    await wrapper.get('select').setValue('')
    expect((wrapper.get('input[placeholder="*/30 * * * *"]').element as HTMLInputElement).value).toBe('')
  })

  it('retains the failed draft, blocks duplicate submission and ignores all close paths while saving', async () => {
    const pending = deferred<ReturnType<typeof plan>>()
    api.update.mockReturnValueOnce(pending.promise)
    const wrapper = await openPanel()
    await wrapper.get('input[placeholder="*/30 * * * *"]').setValue('0 7 * * *')
    await wrapper.get('form').trigger('submit')
    await wrapper.get('form').trigger('submit')
    expect(api.update).toHaveBeenCalledTimes(1)
    for (const close of wrapper.findAll('.modal-header button')) await close.trigger('click')
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await flushPromises()
    expect(wrapper.findAll('[role="dialog"]')).toHaveLength(2)
    expect(wrapper.emitted('close')).toBeUndefined()
    pending.reject(new Error('save failed')); await flushPromises()
    expect(api.showError).toHaveBeenCalledWith('save failed')
    expect((wrapper.get('input[placeholder="*/30 * * * *"]').element as HTMLInputElement).value).toBe('0 7 * * *')
    await wrapper.get('form').trigger('submit'); await flushPromises()
    expect(api.update).toHaveBeenCalledTimes(2)
    expect(wrapper.findAll('[role="dialog"]')).toHaveLength(1)
  })

  it('returns to the list on Escape, allowing an edit button to open the same filled dialog', async () => {
    const wrapper = await openPanel()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await flushPromises()
    expect(wrapper.findAll('[role="dialog"]')).toHaveLength(1)
    expect(wrapper.emitted('close')).toBeUndefined()
    await wrapper.get('button[title="admin.scheduledTests.editPlan"]').trigger('click')
    expect(wrapper.findAll('[role="dialog"]')).toHaveLength(2)
    expect((wrapper.get('select').element as HTMLSelectElement).value).toBe('1')
  })

  it('ignores an old list after closing and reopening for another account', async () => {
    const pending = deferred<ReturnType<typeof plan>[]>()
    api.listByAccount.mockReturnValueOnce(pending.promise)
    const wrapper = await openPanel()
    expect(wrapper.findAll('[role="dialog"]')).toHaveLength(1)
    await wrapper.setProps({ show: false })
    api.listByAccount.mockResolvedValueOnce([plan(9)])
    await wrapper.setProps({ show: true, accountId: 9 }); await flushPromises()
    pending.resolve([plan(1)]); await flushPromises()
    expect((wrapper.get('select').element as HTMLSelectElement).value).toBe('9')
  })

  it('does not close a new account editor when the previous account save resolves', async () => {
    const pending = deferred<ReturnType<typeof plan>>()
    api.update.mockReturnValueOnce(pending.promise)
    const wrapper = await openPanel()
    await wrapper.get('form').trigger('submit')
    api.listByAccount.mockResolvedValueOnce([plan(9)])
    await wrapper.setProps({ accountId: 9 }); await flushPromises()
    pending.resolve(plan(1)); await flushPromises()
    expect((wrapper.get('select').element as HTMLSelectElement).value).toBe('9')
    expect(api.showSuccess).not.toHaveBeenCalled()
  })

  it('keeps the outer scroll lock and restores focus after cancelling a reopened editor', async () => {
    const host = document.createElement('div')
    document.body.append(host)
    const wrapper = mount(ScheduledTestsPanel, {
      attachTo: host,
      props: { show: false, accountId: 1, modelOptions: [] },
      global: { stubs: { Icon: true } }
    })
    try {
      await wrapper.setProps({ show: true }); await flushPromises()
      const body = new DOMWrapper(document.body)
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
      await flushPromises()
      const edit = body.get('button[title="admin.scheduledTests.editPlan"]')
      edit.element.focus()
      await edit.trigger('click'); await flushPromises()
      expect(body.findAll('.modal-overlay')[1].attributes('style')).toContain('z-index: 60')
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
      await flushPromises()
      expect(document.activeElement).toBe(edit.element)
      expect(document.body.classList.contains('modal-open')).toBe(true)
      await wrapper.setProps({ show: false }); await flushPromises()
      expect(document.body.classList.contains('modal-open')).toBe(false)
    } finally { wrapper.unmount(); host.remove() }
  })
})
