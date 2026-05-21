import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import OpenWebUILaunchView from '../OpenWebUILaunchView.vue'

const keysList = vi.hoisted(() => vi.fn())
const launch = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => params ? `${key}:${JSON.stringify(params)}` : key,
    }),
  }
})

vi.mock('@/api', () => ({
  keysAPI: {
    list: keysList,
  },
  openWebUIAPI: {
    launch,
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
  }),
}))

vi.mock('vue-router', () => ({
  RouterLink: {
    props: ['to'],
    template: '<a :data-to="typeof to === `string` ? to : JSON.stringify(to)"><slot /></a>',
  },
  useRoute: () => routeState,
}))

function mountView() {
  return mount(OpenWebUILaunchView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Select: {
          props: ['modelValue', 'options', 'placeholder', 'disabled', 'searchable'],
          emits: ['update:modelValue'],
          template: `
            <select
              :value="modelValue"
              :disabled="disabled"
              @change="$emit('update:modelValue', Number($event.target.value))"
            >
              <option v-for="option in options" :key="String(option.value)" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          `,
        },
        Icon: { template: '<span />' },
      },
    },
  })
}

describe('OpenWebUILaunchView', () => {
  beforeEach(() => {
    routeState.query = {}
    keysList.mockReset().mockResolvedValue({
      items: [
        {
          id: 1,
          key: 'sk-test',
          name: 'mc',
          status: 'active',
          quota: 0,
          quota_used: 0,
          expires_at: null,
          group_id: 10,
          group: {
            id: 10,
            name: 'codex',
            platform: 'openai',
            allow_image_generation: true,
          },
        },
      ],
    })
    launch.mockReset().mockResolvedValue({
      launch_url: 'https://chat.example.com/auth/sub2api/callback?launch_token=one-time',
      expires_at: '2026-05-15T12:00:00Z',
    })
    showError.mockReset()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('navigates a user-opened new tab without changing the current page', async () => {
    const openedTab = {
      opener: {},
      location: {
        replace: vi.fn(),
      },
      close: vi.fn(),
    }
    const windowOpen = vi.spyOn(window, 'open').mockReturnValue(openedTab as unknown as Window)
    const originalHref = window.location.href
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('button.btn-primary').trigger('click')
    await flushPromises()

    expect(windowOpen).toHaveBeenCalledWith('about:blank', '_blank')
    expect(launch).toHaveBeenCalledWith(1)
    expect(openedTab.opener).toBeNull()
    expect(openedTab.location.replace).toHaveBeenCalledWith('https://chat.example.com/auth/sub2api/callback?launch_token=one-time')
    expect(window.location.href).toBe(originalHref)
  })

  it('does not request a launch token when the browser blocks the new tab', async () => {
    vi.spyOn(window, 'open').mockReturnValue(null)
    const wrapper = mountView()

    await flushPromises()
    await wrapper.find('button.btn-primary').trigger('click')
    await flushPromises()

    expect(launch).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('openWebUI.popupBlocked')
  })

  it('auto-launches in the current tab when auto_launch is requested', async () => {
    routeState.query = { auto_launch: '1' }
    const originalLocation = window.location
    const locationState = { href: 'http://localhost/open-webui/launch?auto_launch=1' }
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: locationState,
    })

    mountView()
    await flushPromises()
    await flushPromises()

    expect(launch).toHaveBeenCalledWith(1)
    expect(locationState.href).toBe('https://chat.example.com/auth/sub2api/callback?launch_token=one-time')

    Object.defineProperty(window, 'location', {
      configurable: true,
      value: originalLocation,
    })
  })

  it('keeps the launch intent when the user needs to create a key first', async () => {
    routeState.query = { auto_launch: '1' }
    keysList.mockResolvedValue({ items: [] })

    const wrapper = mountView()
    await flushPromises()

    expect(launch).not.toHaveBeenCalled()
    expect(wrapper.find('a.btn-primary').attributes('data-to')).toBe(
      JSON.stringify({
        path: '/keys',
        query: { redirect: '/open-webui/launch?auto_launch=1' },
      })
    )
  })
})
