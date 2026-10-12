import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, shallowMount } from '@vue/test-utils'

vi.mock('vue-i18n', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key }),
}))
enableAutoUnmount(afterEach)

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise<T>((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}

import OpsSettingsDialog from '../OpsSettingsDialog.vue'
const { opsAPI, showError } = vi.hoisted(() => ({
  opsAPI: {
    getAlertRuntimeSettings: vi.fn(), getEmailNotificationConfig: vi.fn(),
    getAdvancedSettings: vi.fn(), getMetricThresholds: vi.fn(),
    updateAlertRuntimeSettings: vi.fn(), updateEmailNotificationConfig: vi.fn(),
    updateAdvancedSettings: vi.fn(), updateMetricThresholds: vi.fn(),
  }, showError: vi.fn(),
}))
vi.mock('@/api/admin/ops', () => ({ opsAPI }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError, showSuccess: vi.fn() }) }))
async function openDialog() {
  const wrapper = shallowMount(OpsSettingsDialog, {
    props: { show: false },
    global: { stubs: { BaseDialog: { template: '<div><slot name="footer" /></div>' } } },
  })
  await wrapper.setProps({ show: true })
  return wrapper
}
type View = {
  saveAllSettings: () => Promise<void>
  loading: boolean
  settingsLoaded: boolean
  metricThresholds: { sla_percent_min: number }
}
describe('ops settings load guard', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
    opsAPI.getAlertRuntimeSettings.mockResolvedValue({ evaluation_interval_seconds: 30 })
    opsAPI.getEmailNotificationConfig.mockResolvedValue({ alert: { enabled: false }, report: { enabled: false } })
    opsAPI.getAdvancedSettings.mockResolvedValue({ data_retention: { error_log_retention_days: 7, minute_metrics_retention_days: 7, hourly_metrics_retention_days: 7 } })
    opsAPI.getMetricThresholds.mockResolvedValue({ sla_percent_min: 98 })
  })
  afterEach(() => { vi.restoreAllMocks() })
  it('blocks saving while any settings request is still loading', async () => {
    const pending = deferred<object>()
    opsAPI.getMetricThresholds.mockReturnValueOnce(pending.promise)
    const wrapper = await openDialog()
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    await (wrapper.vm as unknown as View).saveAllSettings()
    expect(opsAPI.updateMetricThresholds).not.toHaveBeenCalled()
    pending.resolve({ sla_percent_min: 98 })
    await flushPromises()
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeUndefined()
    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()
    expect(opsAPI.updateMetricThresholds).toHaveBeenCalledWith(expect.objectContaining({ sla_percent_min: 98 }))
    expect(wrapper.emitted('saved')).toHaveLength(1)
  })
  it('does not save default thresholds after a failed initial load', async () => {
    opsAPI.getMetricThresholds.mockRejectedValueOnce(new Error('load failed'))
    const wrapper = await openDialog()
    await flushPromises()
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    await (wrapper.vm as unknown as View).saveAllSettings()
    expect(opsAPI.updateMetricThresholds).not.toHaveBeenCalled()
    expect(wrapper.emitted('saved')).toBeUndefined()
  })
  it('blocks saving stale settings after a failed reopen and allows a later retry', async () => {
    const wrapper = await openDialog()
    await flushPromises()
    await wrapper.setProps({ show: false })
    opsAPI.getMetricThresholds.mockRejectedValueOnce(new Error('refresh failed'))
    await wrapper.setProps({ show: true })
    await flushPromises()
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    await (wrapper.vm as unknown as View).saveAllSettings()
    expect(opsAPI.updateAlertRuntimeSettings).not.toHaveBeenCalled()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    await flushPromises()
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeUndefined()
  })

  it('rejects an old successful load after the reopened load has failed without unlocking save', async () => {
    const old = deferred<object>()
    opsAPI.getMetricThresholds.mockReturnValueOnce(old.promise).mockRejectedValueOnce(new Error('current load failed'))
    const wrapper = await openDialog()
    const view = wrapper.vm as unknown as View
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    await flushPromises()
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    const thresholds = { ...view.metricThresholds }
    old.resolve({ sla_percent_min: 97 })
    await flushPromises()
    expect(view.metricThresholds).toEqual(thresholds)
    expect(view.settingsLoaded).toBe(false)
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    await wrapper.get('button.btn-primary').trigger('click')
    await view.saveAllSettings()
    await flushPromises()
    expect(opsAPI.updateAlertRuntimeSettings).not.toHaveBeenCalled()
    expect(opsAPI.updateEmailNotificationConfig).not.toHaveBeenCalled()
    expect(opsAPI.updateAdvancedSettings).not.toHaveBeenCalled()
    expect(opsAPI.updateMetricThresholds).not.toHaveBeenCalled()
    expect(wrapper.emitted('saved')).toBeUndefined()
    expect(showError).toHaveBeenCalledTimes(1)
  })

  it('ignores old errors and finally while the reopened load remains pending', async () => {
    const old = deferred<object>()
    const current = deferred<object>()
    opsAPI.getMetricThresholds.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const wrapper = await openDialog()
    const view = wrapper.vm as unknown as View
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    old.reject(new Error('obsolete load failed'))
    await flushPromises()
    expect(showError).not.toHaveBeenCalled()
    expect(view.loading).toBe(true)
    expect(view.settingsLoaded).toBe(false)
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    current.resolve({ sla_percent_min: 98 })
    await flushPromises()
    expect(view.loading).toBe(false)
    expect(view.settingsLoaded).toBe(true)
    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()
    expect(opsAPI.updateMetricThresholds).toHaveBeenCalledWith(expect.objectContaining({ sla_percent_min: 98 }))
  })

  it('invalidates a pending load on unmount without emitting an obsolete error toast', async () => {
    const old = deferred<object>()
    opsAPI.getMetricThresholds.mockReturnValueOnce(old.promise)
    const wrapper = await openDialog()
    const view = wrapper.vm as unknown as View
    wrapper.unmount()
    old.reject(new Error('unmounted load failed'))
    await flushPromises()
    expect(showError).not.toHaveBeenCalled()
    expect(view.loading).toBe(false)
    expect(view.settingsLoaded).toBe(false)
    await view.saveAllSettings()
    expect(opsAPI.updateMetricThresholds).not.toHaveBeenCalled()
    expect(wrapper.emitted('saved')).toBeUndefined()
  })
})
