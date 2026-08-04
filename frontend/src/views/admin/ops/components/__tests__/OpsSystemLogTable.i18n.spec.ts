import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import enLocale from '@/i18n/locales/en'
import zhLocale from '@/i18n/locales/zh'
import OpsSystemLogTable from '../OpsSystemLogTable.vue'

const systemLogKeys = [
  'title',
  'description',
  'queue',
  'written',
  'dropped',
  'failed',
  'runtimeConfig',
  'all',
  'level',
  'stacktraceThreshold',
  'samplingInitial',
  'samplingThereafter',
  'retentionDays',
  'caller',
  'sampling',
  'saveAndApply',
  'resetDefaults',
  'latestWriteError',
  'timeRange',
  'startTime',
  'endTime',
  'component',
  'componentPlaceholder',
  'platform',
  'model',
  'keyword',
  'keywordPlaceholder',
  'search',
  'cleanCurrentFilters',
  'refreshHealth',
  'empty',
  'time',
  'logDetails',
  'loadFailed',
  'runtimeConfigActive',
  'runtimeConfigSaveFailed',
  'resetRuntimeConfigConfirm',
  'runtimeConfigReset',
  'runtimeConfigResetFailed',
  'cleanupConfirm',
  'cleanupSuccess',
  'cleanupFilterRequired',
  'cleanupFailed'
] as const

const {
  listSystemLogs,
  getSystemLogSinkHealth,
  getRuntimeLogConfig,
  cleanupSystemLogs,
  showError,
  showSuccess
} = vi.hoisted(() => ({
  listSystemLogs: vi.fn(),
  getSystemLogSinkHealth: vi.fn(),
  getRuntimeLogConfig: vi.fn(),
  cleanupSystemLogs: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    listSystemLogs,
    getSystemLogSinkHealth,
    getRuntimeLogConfig,
    cleanupSystemLogs
  }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError, showSuccess })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'admin.ops.systemLogs.cleanupFilterRequired': '清理需要至少一个筛选条件（起止时间或其他字段）',
    'admin.ops.systemLogs.cleanupFailed': '清理系统日志失败'
  }
  return { ...actual, useI18n: () => ({ t: (key: string) => messages[key] ?? key }) }
})

const health = {
  queue_depth: 0,
  queue_capacity: 10,
  dropped_count: 0,
  write_failed_count: 0,
  written_count: 1,
  avg_write_delay_ms: 1
}

const runtimeConfig = {
  level: 'info',
  enable_sampling: false,
  sampling_initial: 100,
  sampling_thereafter: 100,
  caller: true,
  stacktrace_level: 'error',
  retention_days: 30
}

function mountView() {
  return mount(OpsSystemLogTable, {
    global: {
      stubs: {
        Select: {
          props: ['modelValue', 'options'],
          template: '<select :value="modelValue" />'
        },
        Pagination: { template: '<div />' }
      }
    }
  })
}

describe('OpsSystemLogTable cleanup errors', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listSystemLogs.mockResolvedValue({ items: [], total: 0 })
    getSystemLogSinkHealth.mockResolvedValue(health)
    getRuntimeLogConfig.mockResolvedValue(runtimeConfig)
    cleanupSystemLogs.mockReset()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('localizes the stable filter-required error and preserves other backend details', async () => {
    cleanupSystemLogs.mockRejectedValueOnce({ reason: 'OPS_SYSTEM_LOG_CLEANUP_FILTER_REQUIRED' })
    const wrapper = mountView()
    await flushPromises()

    const cleanupButton = wrapper.find('button.btn-danger')
    expect(cleanupButton).toBeTruthy()
    await cleanupButton!.trigger('click')
    await flushPromises()
    expect(showError).toHaveBeenLastCalledWith('清理需要至少一个筛选条件（起止时间或其他字段）')

    cleanupSystemLogs.mockRejectedValueOnce({ response: { data: { detail: '后端拒绝原因' } } })
    await cleanupButton!.trigger('click')
    await flushPromises()
    expect(showError).toHaveBeenLastCalledWith('后端拒绝原因')
    expect(showSuccess).not.toHaveBeenCalled()
  })
})

describe('OpsSystemLogTable locale parity', () => {
  it('defines the same complete system-log key set in Chinese and English', () => {
    const zhKeys = Object.keys(zhLocale.admin.ops.systemLogs).sort()
    const enKeys = Object.keys(enLocale.admin.ops.systemLogs).sort()
    expect(zhKeys).toEqual([...systemLogKeys].sort())
    expect(enKeys).toEqual([...systemLogKeys].sort())

    for (const key of systemLogKeys) {
      expect(zhLocale.admin.ops.systemLogs[key]).not.toBe(key)
      expect(enLocale.admin.ops.systemLogs[key]).toBeTypeOf('string')
      expect(enLocale.admin.ops.systemLogs[key]).not.toBe('')
    }
    expect(zhLocale.admin.ops.systemLogs.cleanupSuccess).toContain('{count}')
    expect(enLocale.admin.ops.systemLogs.cleanupSuccess).toContain('{count}')
  })
})
