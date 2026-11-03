import { defineComponent, h } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import GroupsView from '../GroupsView.vue'

const api = vi.hoisted(() => ({
  list: vi.fn(),
  create: vi.fn(),
  update: vi.fn(),
  getModelsListCandidates: vi.fn(),
  getUsageSummary: vi.fn(),
  getCapacitySummary: vi.fn(),
  listChannels: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groups: {
      ...api,
      getAll: vi.fn(),
      delete: vi.fn(),
      duplicate: vi.fn(),
      updateSortOrder: vi.fn(),
    },
    accounts: { list: vi.fn(), getById: vi.fn() },
    channels: { list: api.listChannels },
  },
}))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: { server_utc_offset: 8 },
    showSuccess: vi.fn(),
    showWarning: vi.fn(),
    showError: api.showError,
  }),
}))
vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({ isCurrentStep: () => false, nextStep: vi.fn() }),
}))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key }),
}))

const Dialog = defineComponent({
  props: { show: Boolean },
  template: '<section v-if="show"><slot /><slot name="footer" /></section>',
})
const Select = defineComponent({
  inheritAttrs: false,
  props: {
    modelValue: { type: [String, Number, Boolean], default: '' },
    options: { type: Array, default: () => [] },
    disabled: Boolean,
  },
  emits: ['update:modelValue'],
  setup(props, { attrs, emit }) {
    return () => h('select', {
      ...attrs,
      value: props.modelValue,
      disabled: props.disabled,
      onChange: (event: Event) => emit('update:modelValue', (event.target as HTMLSelectElement).value),
    }, (props.options as Array<{ value: string; label: string }>).map(
      (option) => h('option', { value: option.value }, option.label),
    ))
  },
})

function group(platform = 'openai', force?: boolean) {
  return {
    id: 42, name: 'Fast group', description: '', platform,
    force_openai_fast: force, rate_multiplier: 1, is_exclusive: false,
    status: 'active', subscription_type: 'standard', access_mode: 'normal',
    daily_limit_usd: null, weekly_limit_usd: null, monthly_limit_usd: null,
    allow_image_generation: false, peak_rate_enabled: false,
    peak_start: '', peak_end: '', peak_rate_multiplier: 1,
    model_pricing: [], model_match_patterns: ['*'], model_routing: null,
    supported_model_scopes: [], account_count: 0,
  }
}

let wrapper: ReturnType<typeof mountView> | undefined
function mountView() {
  return mount(GroupsView, {
    attachTo: document.body,
    global: {
      stubs: {
        AppLayout: { template: '<main><slot /></main>' },
        TablePageLayout: { template: '<section><slot name="filters" /><slot name="table" /><slot name="pagination" /></section>' },
        DataTable: {
          props: ['data'],
          template: '<div><div v-for="row in data" :key="row.id"><slot name="cell-actions" :row="row" /></div></div>',
        },
        BaseDialog: Dialog, Select,
        Pagination: true, ConfirmDialog: true, EmptyState: true,
        PlatformIcon: true, Icon: true, GroupCapacityBadge: true,
        GroupRateMultipliersModal: true, GroupRPMOverridesModal: true, VueDraggable: true,
      },
    },
  })
}

async function openCreate() {
  wrapper = mountView()
  await flushPromises()
  await wrapper.get('[data-tour="groups-create-btn"]').trigger('click')
  await flushPromises()
  await wrapper.get('#create-group-form input[type="text"]').setValue('New Fast group')
  const areas = wrapper.findAll('#create-group-form textarea')
  await areas[areas.length - 1].setValue('*')
  await wrapper.get('#create-group-form [data-tour="group-form-platform"]').setValue('openai')
  await flushPromises()
  return wrapper
}

async function openEdit(source: ReturnType<typeof group>) {
  api.list.mockResolvedValue({ items: [source], total: 1, page: 1, page_size: 20, pages: 1 })
  wrapper = mountView()
  await flushPromises()
  const button = wrapper.findAll('button').find((item) => item.text() === 'common.edit')
  expect(button).toBeDefined()
  await button!.trigger('click')
  await flushPromises()
  return wrapper
}

async function save(view: ReturnType<typeof mountView>, mode: 'create' | 'edit') {
  // Native form-associated footer button: click must actually submit the mounted form.
  await view.get(`button[form="${mode}-group-form"]`).trigger('click')
  await flushPromises()
  expect(api.showError).not.toHaveBeenCalled()
}

beforeEach(() => {
  localStorage.clear()
  vi.clearAllMocks()
  api.list.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20, pages: 0 })
  api.create.mockResolvedValue(group())
  api.update.mockResolvedValue(group())
  api.getModelsListCandidates.mockResolvedValue([])
  api.getUsageSummary.mockResolvedValue([])
  api.getCapacitySummary.mockResolvedValue([])
  api.listChannels.mockResolvedValue({ items: [], total: 0 })
})
afterEach(() => {
  wrapper?.unmount()
  wrapper = undefined
  document.body.innerHTML = ''
})

describe('mounted group Fast create/save', () => {
  it('defaults off and clicks Save to send false', async () => {
    const view = await openCreate()
    expect(view.get('[data-testid="create-force-openai-fast"]').attributes('aria-checked')).toBe('false')
    expect(view.text()).toContain('admin.groups.openaiFast.hint')
    await save(view, 'create')
    expect(api.create).toHaveBeenCalledWith(expect.objectContaining({ force_openai_fast: false }))
  })

  it('clicks the switch then Save to send true', async () => {
    const view = await openCreate()
    await view.get('[data-testid="create-force-openai-fast"]').trigger('click')
    expect(view.get('[data-testid="create-force-openai-fast"]').attributes('aria-checked')).toBe('true')
    await save(view, 'create')
    expect(api.create).toHaveBeenCalledWith(expect.objectContaining({ platform: 'openai', force_openai_fast: true }))
  })

  it('clears enabled state after switching away and back', async () => {
    const view = await openCreate()
    await view.get('[data-testid="create-force-openai-fast"]').trigger('click')
    const platform = view.get('#create-group-form [data-tour="group-form-platform"]')
    await platform.setValue('gemini')
    expect(view.find('[data-testid="create-force-openai-fast"]').exists()).toBe(false)
    await platform.setValue('openai')
    expect(view.get('[data-testid="create-force-openai-fast"]').attributes('aria-checked')).toBe('false')
    await save(view, 'create')
    expect(api.create).toHaveBeenCalledWith(expect.objectContaining({ force_openai_fast: false }))
  })

  it('saves unsupported platforms with false after enabling on OpenAI', async () => {
    const view = await openCreate()
    await view.get('[data-testid="create-force-openai-fast"]').trigger('click')
    await view.get('#create-group-form [data-tour="group-form-platform"]').setValue('gemini')
    await save(view, 'create')
    expect(api.create).toHaveBeenCalledWith(expect.objectContaining({ platform: 'gemini', force_openai_fast: false }))
  })

  it('resets the switch when create is closed and reopened', async () => {
    const view = await openCreate()
    await view.get('[data-testid="create-force-openai-fast"]').trigger('click')
    const cancel = view.findAll('button').find((item) => item.text() === 'common.cancel')
    await cancel!.trigger('click')
    await view.get('[data-tour="groups-create-btn"]').trigger('click')
    await view.get('#create-group-form [data-tour="group-form-platform"]').setValue('openai')
    expect(view.get('[data-testid="create-force-openai-fast"]').attributes('aria-checked')).toBe('false')
  })
})

describe('mounted group Fast edit/load/save', () => {
  it.each(['openai', 'composite'])('loads and saves enabled %s group', async (platform) => {
    const view = await openEdit(group(platform, true))
    expect(view.get('[data-testid="edit-force-openai-fast"]').attributes('aria-checked')).toBe('true')
    expect(view.text()).toContain('admin.groups.openaiFast.hint')
    await save(view, 'edit')
    expect(api.update).toHaveBeenCalledWith(42, expect.objectContaining({ platform, force_openai_fast: true }))
  })

  it('clicks the edit switch then Save to disable Fast', async () => {
    const view = await openEdit(group('openai', true))
    await view.get('[data-testid="edit-force-openai-fast"]').trigger('click')
    await save(view, 'edit')
    expect(api.update).toHaveBeenCalledWith(42, expect.objectContaining({ force_openai_fast: false }))
  })

  it('clicks the edit switch then Save to enable Fast', async () => {
    const view = await openEdit(group('openai', false))
    await view.get('[data-testid="edit-force-openai-fast"]').trigger('click')
    await save(view, 'edit')
    expect(api.update).toHaveBeenCalledWith(42, expect.objectContaining({ force_openai_fast: true }))
  })

  it('clears edit state on platform updates while preserving the disabled selector', async () => {
    const view = await openEdit(group('openai', true))
    const platform = view.get('#edit-group-form').findComponent(Select)
    expect(platform.props('disabled')).toBe(true)
    // Editing a platform is disabled in the existing UI; exercise its watcher through the model event.
    platform.vm.$emit('update:modelValue', 'gemini')
    await flushPromises()
    expect(view.find('[data-testid="edit-force-openai-fast"]').exists()).toBe(false)
    platform.vm.$emit('update:modelValue', 'openai')
    await flushPromises()
    expect(view.get('[data-testid="edit-force-openai-fast"]').attributes('aria-checked')).toBe('false')
    await save(view, 'edit')
    expect(api.update).toHaveBeenCalledWith(42, expect.objectContaining({ force_openai_fast: false }))
  })

  it.each([false, undefined])('loads legacy/disabled flag %s as false and saves false', async (flag) => {
    const view = await openEdit(group('openai', flag))
    expect(view.get('[data-testid="edit-force-openai-fast"]').attributes('aria-checked')).toBe('false')
    await save(view, 'edit')
    expect(api.update).toHaveBeenCalledWith(42, expect.objectContaining({ force_openai_fast: false }))
  })

  it('normalizes incompatible persisted data to false on edit save', async () => {
    const view = await openEdit(group('gemini', true))
    expect(view.find('[data-testid="edit-force-openai-fast"]').exists()).toBe(false)
    await save(view, 'edit')
    expect(api.update).toHaveBeenCalledWith(42, expect.objectContaining({ platform: 'gemini', force_openai_fast: false }))
  })

  it('discards an unsaved edit toggle when reopening the group', async () => {
    const view = await openEdit(group('openai', false))
    await view.get('[data-testid="edit-force-openai-fast"]').trigger('click')
    const cancel = view.findAll('button').find((item) => item.text() === 'common.cancel')
    await cancel!.trigger('click')
    const edit = view.findAll('button').find((item) => item.text() === 'common.edit')
    await edit!.trigger('click')
    await flushPromises()
    expect(view.get('[data-testid="edit-force-openai-fast"]').attributes('aria-checked')).toBe('false')
    await save(view, 'edit')
    expect(api.update).toHaveBeenCalledWith(42, expect.objectContaining({ force_openai_fast: false }))
  })
})
