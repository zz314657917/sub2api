import { describe, it, expect, beforeEach, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsSettingsDialog from '../OpsSettingsDialog.vue'
import { opsAPI } from '@/api/admin/ops'

const { mockShowError, mockShowSuccess } = vi.hoisted(() => ({
  mockShowError: vi.fn(),
  mockShowSuccess: vi.fn(),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: mockShowError,
    showSuccess: mockShowSuccess,
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getAlertRuntimeSettings: vi.fn(),
    getEmailNotificationConfig: vi.fn(),
    getAdvancedSettings: vi.fn(),
    getMetricThresholds: vi.fn(),
    updateAlertRuntimeSettings: vi.fn(),
    updateEmailNotificationConfig: vi.fn(),
    updateAdvancedSettings: vi.fn(),
    updateMetricThresholds: vi.fn(),
  },
}))

const BaseDialogStub = defineComponent({
  name: 'BaseDialog',
  props: {
    show: {
      type: Boolean,
      default: false,
    },
  },
  template: '<div v-if="show" class="base-dialog"><slot /><slot name="footer" /></div>',
})

const ToggleStub = defineComponent({
  name: 'Toggle',
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:modelValue'],
  template: '<input class="toggle-stub" type="checkbox" :checked="modelValue" />',
})

const SelectStub = defineComponent({
  name: 'Select',
  props: {
    modelValue: {
      type: [String, Number],
      default: '',
    },
  },
  emits: ['update:modelValue'],
  template: '<select class="select-stub" :value="modelValue" />',
})

function runtimeSettings() {
  return {
    evaluation_interval_seconds: 60,
    distributed_lock: {
      enabled: false,
      key: 'ops-alert-runtime-lock',
      ttl_seconds: 60,
    },
    silencing: {
      enabled: false,
      global_until_rfc3339: '',
      global_reason: '',
      entries: [],
    },
    thresholds: {},
  }
}

function emailConfig() {
  return {
    alert: {
      enabled: false,
      recipients: [],
      min_severity: '',
      rate_limit_per_hour: 10,
      batching_window_seconds: 60,
      include_resolved_alerts: false,
    },
    report: {
      enabled: false,
      recipients: [],
      daily_summary_enabled: false,
      daily_summary_schedule: '0 9 * * *',
      weekly_summary_enabled: false,
      weekly_summary_schedule: '0 9 * * 1',
      error_digest_enabled: false,
      error_digest_schedule: '0 * * * *',
      error_digest_min_count: 5,
      account_health_enabled: false,
      account_health_schedule: '0 9 * * *',
      account_health_error_rate_threshold: 5,
    },
  }
}

function legacyAdvancedSettings() {
  return {
    data_retention: {
      cleanup_enabled: false,
      cleanup_schedule: '0 2 * * *',
      error_log_retention_days: 30,
      minute_metrics_retention_days: 30,
      hourly_metrics_retention_days: 30,
    },
    aggregation: {
      aggregation_enabled: false,
    },
    ignore_count_tokens_errors: true,
    ignore_context_canceled: true,
    ignore_no_available_accounts: false,
    ignore_invalid_api_key_errors: false,
    ignore_insufficient_balance_errors: false,
    display_openai_token_stats: false,
    display_alert_events: true,
    auto_refresh_enabled: false,
    auto_refresh_interval_seconds: 30,
  }
}

describe('OpsSettingsDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(opsAPI.getAlertRuntimeSettings).mockResolvedValue(runtimeSettings() as any)
    vi.mocked(opsAPI.getEmailNotificationConfig).mockResolvedValue(emailConfig() as any)
    vi.mocked(opsAPI.getMetricThresholds).mockResolvedValue({})
    vi.mocked(opsAPI.updateAlertRuntimeSettings).mockResolvedValue(runtimeSettings() as any)
    vi.mocked(opsAPI.updateEmailNotificationConfig).mockResolvedValue(emailConfig() as any)
    vi.mocked(opsAPI.updateAdvancedSettings).mockImplementation(async (settings) => settings)
    vi.mocked(opsAPI.updateMetricThresholds).mockResolvedValue()
  })

  it('backfills missing OpenAI quota auto-pause settings from legacy responses before saving', async () => {
    vi.mocked(opsAPI.getAdvancedSettings).mockResolvedValue(legacyAdvancedSettings() as any)

    const wrapper = mount(OpsSettingsDialog, {
      props: {
        show: false,
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Toggle: ToggleStub,
          Select: SelectStub,
        },
      },
    })

    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(wrapper.text()).toContain('admin.ops.settings.openAIQuotaAutoPause')

    const saveButton = wrapper.findAll('button').find(button => button.text() === 'common.save')
    expect(saveButton).toBeTruthy()
    await saveButton!.trigger('click')
    await flushPromises()

    expect(mockShowError).not.toHaveBeenCalled()
    expect(opsAPI.updateAdvancedSettings).toHaveBeenCalledWith(expect.objectContaining({
      openai_account_quota_auto_pause: {
        default_threshold_5h: 0,
        default_threshold_7d: 0,
      },
    }))
  })
})
