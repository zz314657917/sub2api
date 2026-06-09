import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import LuoyeAILaunchView from '../LuoyeAILaunchView.vue'

const launch = vi.hoisted(() => vi.fn())
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

  it('launches Luoye AI through the studio bridge in the current tab', async () => {
    mountView()
    await flushPromises()

    expect(launch).toHaveBeenCalledWith()
    expect(locationState.assign).toHaveBeenCalledWith(
      'http://127.0.0.1:8081/auth/sub2api/launch?launch_token=one-time'
    )
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
