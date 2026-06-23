import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import LuoyeAILaunchView from '../LuoyeAILaunchView.vue'

const launch = vi.hoisted(() => vi.fn())
const keysList = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/api', () => ({
  keysAPI: {
    list: keysList,
  },
  studioBridgeAPI: {
    launch,
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
  }),
}))

function mountView() {
  return mount(LuoyeAILaunchView, {
    attachTo: document.body,
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: { template: '<span />' },
      },
    },
  })
}

function makeKey(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    user_id: 1,
    key: 'sk-default',
    name: '默认 API Key（勿删）',
    is_default: true,
    group_id: 10,
    multi_group_routes: [],
    account_pool_strategy: 'shared_only',
    status: 'active',
    ip_whitelist: [],
    ip_blacklist: [],
    last_used_at: null,
    quota: 0,
    quota_used: 0,
    expires_at: null,
    created_at: '2026-06-09T12:00:00Z',
    updated_at: '2026-06-09T12:00:00Z',
    rate_limit_5h: 0,
    rate_limit_1d: 0,
    rate_limit_7d: 0,
    usage_5h: 0,
    usage_1d: 0,
    usage_7d: 0,
    window_5h_start: null,
    window_1d_start: null,
    window_7d_start: null,
    reset_5h_at: null,
    reset_1d_at: null,
    reset_7d_at: null,
    group: {
      id: 10,
      name: 'Image',
      status: 'active',
      platform: 'openai',
      routing_scope: 'image',
      allow_image_generation: true,
    },
    ...overrides,
  }
}

describe('LuoyeAILaunchView', () => {
  let originalLocation: Location
  let locationState: { assign: ReturnType<typeof vi.fn> }

  beforeEach(() => {
    originalLocation = window.location
    locationState = { assign: vi.fn() }
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: locationState,
    })
    launch.mockReset().mockResolvedValue({
      launch_url: 'http://127.0.0.1:8081/auth/sub2api/launch?launch_token=one-time',
      expires_at: '2026-06-09T12:00:00Z',
    })
    keysList.mockReset().mockResolvedValue({
      items: [makeKey()],
    })
    showError.mockReset()
  })

  afterEach(() => {
    document.body.innerHTML = ''
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: originalLocation,
    })
    vi.restoreAllMocks()
  })

  it('launches Luoye Creative through the studio bridge in the current tab', async () => {
    mountView()
    await flushPromises()

    expect(keysList).toHaveBeenCalledWith(1, 100, {
      status: 'active',
      sort_by: 'created_at',
      sort_order: 'desc',
    })
    expect(launch).toHaveBeenCalledWith()
    expect(locationState.assign).toHaveBeenCalledWith(
      'http://127.0.0.1:8081/auth/sub2api/launch?launch_token=one-time'
    )
  })

  it('blocks launch when the default API key has no image group', async () => {
    keysList.mockResolvedValue({
      items: [
        makeKey({
          group: {
            id: 10,
            name: 'Chat',
            status: 'active',
            platform: 'openai',
            routing_scope: 'inference',
            allow_image_generation: false,
          },
        }),
        makeKey({
          id: 2,
          is_default: false,
          group: {
            id: 20,
            name: 'Image',
            status: 'active',
            platform: 'openai',
            routing_scope: 'image',
            allow_image_generation: true,
          },
        }),
      ],
    })

    const wrapper = mountView()
    await flushPromises()

    expect(launch).not.toHaveBeenCalled()
    expect(locationState.assign).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('chatImageStudio.launchFailedTitle')
    expect(wrapper.text()).toContain('chatImageStudio.defaultImageGroupMissing')
    expect(showError).toHaveBeenCalledWith('chatImageStudio.defaultImageGroupMissing')
  })

  it('shows a retry state when launch fails', async () => {
    launch.mockRejectedValue(new Error('launch failed'))

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('chatImageStudio.launchFailedTitle')
    expect(wrapper.text()).toContain('chatImageStudio.launchFailedDescription')
    expect(showError).toHaveBeenCalledWith('chatImageStudio.launchFailedDescription')
  })
})
