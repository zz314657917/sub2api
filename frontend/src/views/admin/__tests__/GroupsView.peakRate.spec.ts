import { defineComponent, h } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import GroupsView from '@/views/admin/GroupsView.vue'

const {
  listGroups,
  createGroup,
  getModelsListCandidates,
  getUsageSummary,
  getCapacitySummary,
  showSuccess,
  showWarning,
  showError
} = vi.hoisted(() => ({
  listGroups: vi.fn(),
  createGroup: vi.fn(),
  getModelsListCandidates: vi.fn(),
  getUsageSummary: vi.fn(),
  getCapacitySummary: vi.fn(),
  showSuccess: vi.fn(),
  showWarning: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groups: {
      list: listGroups,
      create: createGroup,
      update: vi.fn(),
      delete: vi.fn(),
      duplicate: vi.fn(),
      updateSortOrder: vi.fn(),
      getModelsListCandidates,
      getUsageSummary,
      getCapacitySummary,
      getAll: vi.fn()
    },
    accounts: {
      list: vi.fn(),
      getById: vi.fn()
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: { server_utc_offset: 8 },
    showSuccess,
    showWarning,
    showError
  })
}))

vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({
    isCurrentStep: vi.fn(() => false),
    nextStep: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() })
}))

const AppLayoutStub = defineComponent({
  template: '<main><slot /></main>'
})

const TablePageLayoutStub = defineComponent({
  template: '<section><slot name="filters" /><slot name="table" /><slot name="pagination" /></section>'
})

const DataTableStub = defineComponent({
  props: {
    data: { type: Array, default: () => [] },
    columns: { type: Array, default: () => [] },
    loading: { type: Boolean, default: false }
  },
  template: '<div><slot name="empty" /></div>'
})

const BaseDialogStub = defineComponent({
  props: { show: { type: Boolean, default: false } },
  template: '<section v-if="show"><slot /><slot name="footer" /></section>'
})

const SelectStub = defineComponent({
  inheritAttrs: false,
  props: {
    modelValue: { type: [String, Number, Boolean], default: '' },
    options: { type: Array, default: () => [] },
    disabled: { type: Boolean, default: false }
  },
  emits: ['update:modelValue'],
  setup(props, { attrs, emit }) {
    return () => h(
      'select',
      {
        ...attrs,
        disabled: props.disabled,
        value: props.modelValue,
        onChange: (event: Event) => emit(
          'update:modelValue',
          (event.target as HTMLSelectElement).value
        )
      },
      (props.options as Array<{ value: string; label: string }>).map((option) =>
        h('option', { value: option.value }, option.label)
      )
    )
  }
})

function mountView() {
  return mount(GroupsView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        TablePageLayout: TablePageLayoutStub,
        DataTable: DataTableStub,
        BaseDialog: BaseDialogStub,
        Select: SelectStub,
        Pagination: true,
        ConfirmDialog: true,
        EmptyState: true,
        PlatformIcon: true,
        Icon: true,
        GroupCapacityBadge: true,
        GroupRateMultipliersModal: true,
        GroupRPMOverridesModal: true,
        VueDraggable: true
      }
    }
  })
}

async function openCreateForm(wrapper: ReturnType<typeof mountView>) {
  await flushPromises()
  await wrapper.get('[data-tour="groups-create-btn"]').trigger('click')
  await flushPromises()
  await wrapper.get('#create-group-form input[type="text"]').setValue('Timed standard group')
  const textareas = wrapper.findAll('#create-group-form textarea')
  await textareas[textareas.length - 1].setValue('*')
}

async function fillPeakWindow(wrapper: ReturnType<typeof mountView>, multiplier: string) {
  await wrapper.get('[data-testid="create-peak-rate-enabled"]').setValue(true)
  await wrapper.get('[data-testid="create-peak-start"]').setValue('18:00')
  await wrapper.get('[data-testid="create-peak-end"]').setValue('23:00')
  await wrapper.get('[data-testid="create-peak-multiplier"]').setValue(multiplier)
}

describe('GroupsView standard group time-window billing', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.spyOn(console, 'error').mockImplementation(() => {})
    for (const fn of [
      listGroups,
      createGroup,
      getModelsListCandidates,
      getUsageSummary,
      getCapacitySummary,
      showSuccess,
      showWarning,
      showError
    ]) {
      fn.mockReset()
    }
    listGroups.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
      pages: 0
    })
    getModelsListCandidates.mockResolvedValue([])
    getUsageSummary.mockResolvedValue([])
    getCapacitySummary.mockResolvedValue([])
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('shows the time-window controls for the default standard group', async () => {
    const wrapper = mountView()
    await openCreateForm(wrapper)
    await wrapper.get('[data-testid="create-peak-rate-enabled"]').setValue(true)

    expect(wrapper.get('[data-testid="create-peak-rate-section"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="create-peak-multiplier"]').attributes('min')).toBe('0.001')
    wrapper.unmount()
  })

  it('rejects zero for a standard group and submits a positive factor', async () => {
    const wrapper = mountView()
    await openCreateForm(wrapper)
    await fillPeakWindow(wrapper, '0')

    await wrapper.get('#create-group-form').trigger('submit')
    await flushPromises()
    expect(createGroup).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('admin.groups.peakRate.standardMultiplierPositive')

    createGroup.mockRejectedValueOnce(new Error('stop after payload capture'))
    await wrapper.get('[data-testid="create-peak-multiplier"]').setValue('0.7')
    await wrapper.get('#create-group-form').trigger('submit')
    await flushPromises()

    expect(createGroup).toHaveBeenCalledWith(expect.objectContaining({
      subscription_type: 'standard',
      peak_rate_enabled: true,
      peak_start: '18:00',
      peak_end: '23:00',
      peak_rate_multiplier: 0.7
    }))
    wrapper.unmount()
  })

  it('keeps the configured values across type switches and allows subscription zero', async () => {
    const wrapper = mountView()
    await openCreateForm(wrapper)
    await fillPeakWindow(wrapper, '0')

    const typeSelect = wrapper.get('[data-testid="create-subscription-type"]')
    await typeSelect.setValue('subscription')
    expect(wrapper.get('[data-testid="create-peak-multiplier"]').attributes('min')).toBe('0')

    createGroup.mockRejectedValueOnce(new Error('stop after payload capture'))
    await wrapper.get('#create-group-form').trigger('submit')
    await flushPromises()
    expect(createGroup).toHaveBeenCalledWith(expect.objectContaining({
      subscription_type: 'subscription',
      peak_rate_multiplier: 0
    }))

    await typeSelect.setValue('standard')
    expect((wrapper.get('[data-testid="create-peak-rate-enabled"]').element as HTMLInputElement).checked).toBe(true)
    expect((wrapper.get('[data-testid="create-peak-start"]').element as HTMLInputElement).value).toBe('18:00')
    expect((wrapper.get('[data-testid="create-peak-end"]').element as HTMLInputElement).value).toBe('23:00')
    expect((wrapper.get('[data-testid="create-peak-multiplier"]').element as HTMLInputElement).value).toBe('0')
    wrapper.unmount()
  })

  it('submits room_managed only after the administrator selects subscription mode', async () => {
    const wrapper = mountView()
    await openCreateForm(wrapper)
    await wrapper.get('[data-testid="create-subscription-type"]').setValue('subscription')
    await wrapper.get('[data-testid="create-room-managed"]').setValue(true)

    createGroup.mockRejectedValueOnce(new Error('stop after payload capture'))
    await wrapper.get('#create-group-form').trigger('submit')
    await flushPromises()

    expect(createGroup).toHaveBeenCalledWith(expect.objectContaining({
      subscription_type: 'subscription',
      access_mode: 'room_managed',
    }))
    wrapper.unmount()
  })

  it('rejects an empty subscription multiplier instead of silently normalizing it', async () => {
    const wrapper = mountView()
    await openCreateForm(wrapper)
    await fillPeakWindow(wrapper, '0')
    await wrapper.get('[data-testid="create-subscription-type"]').setValue('subscription')
    await wrapper.get('[data-testid="create-peak-multiplier"]').setValue('')

    await wrapper.get('#create-group-form').trigger('submit')
    await flushPromises()

    expect(createGroup).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('admin.groups.peakRate.multiplierInvalid')
    wrapper.unmount()
  })
})
