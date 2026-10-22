import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import { useRoute } from 'vue-router'
import AccountsView from '../AccountsView.vue'
import AccountTableFilters from '@/components/admin/account/AccountTableFilters.vue'

const { listAccounts, getAllProxies, getAllGroups, getBatchTodayStats, getProbeSettings, requests } = vi.hoisted(() => ({
  listAccounts: vi.fn(),
  getAllProxies: vi.fn(),
  getAllGroups: vi.fn(),
  getBatchTodayStats: vi.fn(),
  getProbeSettings: vi.fn(),
  requests: [] as Record<string, unknown>[]
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      list: listAccounts,
      getBatchTodayStats,
      getUpstreamBillingProbeSettings: getProbeSettings
    },
    proxies: { getAll: getAllProxies },
    groups: { getAll: getAllGroups }
  }
}))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn(), showInfo: vi.fn() })
}))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ token: 'test-token' }) }))
vi.mock('vue-router', async () => {
  const { reactive } = await import('vue')
  const route = reactive({ name: 'AdminAccounts' })
  return { useRoute: () => route }
})
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key })
}))

const MAIN = 'account-filter-values'
const SHARED = 'shared-account-filter-values'
const route = useRoute()
const nativeSelect = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue', 'change'],
  template: `<select :value="modelValue" @change="$emit('update:modelValue', $event.target.value); $emit('change')">
    <option v-for="option in options" :key="option.value" :value="option.value">{{ option.label }}</option>
  </select>`
}
const createWrapper = () => shallowMount(AccountsView, {
  global: {
    stubs: {
      AppLayout: { template: '<div><slot /></div>' },
      TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /></div>' },
      AccountTableFilters: false,
      Select: nativeSelect,
      SearchInput: {
        props: ['modelValue'],
        emits: ['update:modelValue', 'search'],
        template: `<input :value="modelValue" @input="$emit('update:modelValue', $event.target.value)" @change="$emit('search')" />`
      }
    }
  }
})
type Wrapper = ReturnType<typeof createWrapper>
let wrappers: Wrapper[] = []
const mountView = async () => {
  const count = requests.length
  const wrapper = createWrapper()
  wrappers.push(wrapper)
  await flushPromises()
  expect(requests).toHaveLength(count + 1)
  return wrapper
}
const settle = async () => {
  await vi.advanceTimersByTimeAsync(350)
  await flushPromises()
}
const changeFilter = async (wrapper: Wrapper, index: number, value: string) => {
  const count = requests.length
  await wrapper.findComponent(AccountTableFilters).findAll('select')[index].setValue(value)
  await settle()
  expect(requests).toHaveLength(count + 1)
}
const lastRequest = () => requests[requests.length - 1]
const expectRequest = (platform: string, status: string, shared = false) => {
  expect(lastRequest()).toMatchObject({
    platform, status,
    owner_filter: shared ? 'user' : 'system',
    share_mode: shared ? 'public' : '',
    ...(shared ? { group: '' } : {})
  })
}
const navigate = async (shared: boolean) => {
  const count = requests.length
  Reflect.set(route, 'name', shared ? 'AdminSharedAccounts' : 'AdminAccounts')
  await flushPromises()
  expect(requests).toHaveLength(count + 1)
}

describe('AccountsView platform/status persistence through real filter events', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    localStorage.clear()
    Reflect.set(route, 'name', 'AdminAccounts')
    requests.length = 0
    listAccounts.mockReset()
    listAccounts.mockImplementation(async (_page, _size, params) => {
      // useTableLoader passes a mutable object; capture the request at call time.
      requests.push({ ...params })
      return { items: [], total: 0, page: 1, page_size: 20, pages: 0 }
    })
    getAllProxies.mockResolvedValue([])
    getAllGroups.mockResolvedValue([{ id: 42, name: 'group-42' }])
    getBatchTodayStats.mockResolvedValue({ stats: {} })
    getProbeSettings.mockResolvedValue({ enabled: false, interval_minutes: 30 })
    wrappers = []
  })
  afterEach(() => {
    wrappers.forEach(wrapper => wrapper.unmount())
    vi.clearAllTimers()
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('restores only platform/status on mount and refresh, retaining fixed main ownership', async () => {
    localStorage.setItem(MAIN, JSON.stringify({
      platform: 'openai', status: 'error', owner_filter: 'user', share_mode: 'public',
      share_status: 'suspended', group: '42', search: 'secret', type: 'oauth', plan_type: 'pro'
    }))
    const wrapper = await mountView()
    expectRequest('openai', 'error')
    expect(lastRequest()).toMatchObject({ group: '', search: '', type: '', plan_type: '', share_status: '' })
    const selects = wrapper.findComponent(AccountTableFilters).findAll('select')
    expect((selects[0].element as HTMLSelectElement).value).toBe('openai')
    expect((selects[2].element as HTMLSelectElement).value).toBe('error')
    wrapper.findComponent({ name: 'AccountTableActions' }).vm.$emit('refresh')
    await flushPromises()
    expect(requests).toHaveLength(2)
    expectRequest('openai', 'error')
  })

  it('saves only platform/status and restores them across unmount/remount', async () => {
    const wrapper = await mountView()
    await changeFilter(wrapper, 0, 'grok')
    expectRequest('grok', '')
    await changeFilter(wrapper, 2, 'unschedulable')
    expectRequest('grok', 'unschedulable')
    await wrapper.findComponent(AccountTableFilters).get('input').setValue('live-search')
    await settle()
    await changeFilter(wrapper, 5, '42')
    expect(lastRequest()).toMatchObject({ search: 'live-search', group: '42' })
    expect(JSON.parse(localStorage.getItem(MAIN)!)).toEqual({ platform: 'grok', status: 'unschedulable' })
    expect(localStorage.getItem(SHARED)).toBeNull()
    wrapper.unmount()
    wrappers = []
    await mountView()
    expectRequest('grok', 'unschedulable')
    expect(lastRequest()).toMatchObject({ search: '', group: '' })
  })

  it('reloads separate saved values on same-instance main→shared→main transitions', async () => {
    localStorage.setItem(MAIN, JSON.stringify({ platform: 'anthropic', status: 'inactive' }))
    localStorage.setItem(SHARED, JSON.stringify({
      platform: 'gemini', status: 'rate_limited', owner_filter: 'system', share_mode: 'private',
      group: '42', share_status: 'rejected', search: 'injected'
    }))
    const wrapper = await mountView()
    expectRequest('anthropic', 'inactive')
    await changeFilter(wrapper, 5, '42')
    await navigate(true)
    expectRequest('gemini', 'rate_limited', true)
    expect(lastRequest()).toMatchObject({ search: '', share_status: '' })
    await changeFilter(wrapper, 0, 'antigravity')
    await changeFilter(wrapper, 2, 'temp_unschedulable')
    expectRequest('antigravity', 'temp_unschedulable', true)
    await navigate(false)
    expectRequest('anthropic', 'inactive')
    expect(JSON.parse(localStorage.getItem(SHARED)!)).toEqual({ platform: 'antigravity', status: 'temp_unschedulable' })
    await navigate(true)
    expectRequest('antigravity', 'temp_unschedulable', true)
    expect(wrapper.findComponent(AccountTableFilters).props('showGroupFilter')).toBe(false)
  })

  it.each([false, true])('falls back to empty values when the destination scope is missing (shared=%s)', async (startShared) => {
    const key = startShared ? SHARED : MAIN
    localStorage.setItem(key, JSON.stringify({ platform: 'openai', status: 'active' }))
    Reflect.set(route, 'name', startShared ? 'AdminSharedAccounts' : 'AdminAccounts')
    const wrapper = await mountView()
    expectRequest('openai', 'active', startShared)
    await navigate(!startShared)
    expectRequest('', '', !startShared)
    await changeFilter(wrapper, 0, 'gemini')
    await navigate(startShared)
    expectRequest('openai', 'active', startShared)
  })

  it.each([false, true])('clearing both real filters persists reset through scope changes and remount (shared=%s)', async (shared) => {
    Reflect.set(route, 'name', shared ? 'AdminSharedAccounts' : 'AdminAccounts')
    const key = shared ? SHARED : MAIN
    localStorage.setItem(key, JSON.stringify({ platform: 'openai', status: 'error' }))
    const wrapper = await mountView()
    await changeFilter(wrapper, 0, '')
    expectRequest('', 'error', shared)
    await changeFilter(wrapper, 2, '')
    expectRequest('', '', shared)
    expect(JSON.parse(localStorage.getItem(key)!)).toEqual({ platform: '', status: '' })
    await navigate(!shared)
    await navigate(shared)
    expectRequest('', '', shared)
    wrapper.unmount()
    wrappers = []
    await mountView()
    expectRequest('', '', shared)
  })

  it('accepts exactly the current filter UI platform and status options', async () => {
    let wrapper = await mountView()
    const selects = wrapper.findComponent(AccountTableFilters).findAll('select')
    for (const [index, key] of [[0, 'platform'], [2, 'status']] as const) {
      const options = selects[index].findAll('option').map(option => option.attributes('value'))
      for (const value of options) {
        await changeFilter(wrapper, index, value)
        expect(lastRequest()[key]).toBe(value)
        wrapper.unmount()
        wrappers = []
        wrapper = await mountView()
        expect(lastRequest()[key]).toBe(value)
      }
    }
  })

  it.each([
    ['malformed JSON', '{', '', ''],
    ['null', 'null', '', ''],
    ['array', '[]', '', ''],
    ['scalar', '"openai"', '', ''],
    ['invalid enum', '{"platform":"bedrock","status":"disabled"}', '', ''],
    ['wrong types', '{"platform":true,"status":7}', '', ''],
    ['valid platform only', '{"platform":"grok","status":"ACTIVE"}', 'grok', ''],
    ['valid status only', '{"platform":"OPENAI","status":"error"}', '', 'error']
  ])('handles %s safely on mount and destination scope reload', async (_label, stored, platform, status) => {
    localStorage.setItem(MAIN, stored)
    localStorage.setItem(SHARED, stored)
    await mountView()
    expectRequest(platform, status)
    await navigate(true)
    expectRequest(platform, status, true)
  })

  it('continues real filtering and scope requests when storage reads and writes throw', async () => {
    const error = new DOMException('Storage unavailable', 'SecurityError')
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => { throw error })
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => { throw error })
    vi.spyOn(console, 'error').mockImplementation(() => {})
    const wrapper = await mountView()
    expectRequest('', '')
    await changeFilter(wrapper, 0, 'gemini')
    await changeFilter(wrapper, 2, 'active')
    expectRequest('gemini', 'active')
    await navigate(true)
    expectRequest('', '', true)
    await changeFilter(wrapper, 0, 'grok')
    expectRequest('grok', '', true)
    await navigate(false)
    expectRequest('', '')
  })
})
