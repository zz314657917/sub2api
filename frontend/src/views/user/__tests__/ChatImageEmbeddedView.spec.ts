import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ChatImageEmbeddedView from '../ChatImageEmbeddedView.vue'

const launch = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const appState = vi.hoisted(() => ({
  sidebarCollapsed: false,
  setSidebarCollapsed: vi.fn((collapsed: boolean) => {
    appState.sidebarCollapsed = collapsed
  }),
}))
const routeState = vi.hoisted(() => ({
  fullPath: '/chat-images',
  query: {} as Record<string, unknown>,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      locale: { value: 'zh' },
      t: (key: string, params?: Record<string, unknown>) => params ? `${key}:${JSON.stringify(params)}` : key,
    }),
  }
})

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
}))

vi.mock('@/api', () => ({
  openWebUIAPI: {
    launch,
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    get sidebarCollapsed() {
      return appState.sidebarCollapsed
    },
    setSidebarCollapsed: appState.setSidebarCollapsed,
    showError,
  }),
}))

function mountView() {
  return mount(ChatImageEmbeddedView, {
    attachTo: document.body,
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: { template: '<span />' },
      },
    },
  })
}

describe('ChatImageEmbeddedView', () => {
  beforeEach(() => {
    document.documentElement.className = ''
    appState.sidebarCollapsed = false
    appState.setSidebarCollapsed.mockClear()
    routeState.fullPath = '/chat-images'
    routeState.query = {}
    launch.mockReset().mockResolvedValue({
      launch_url: 'https://chat.example.com/api/v1/auths/sub2api/launch?token=one-time',
      expires_at: '2026-05-15T12:00:00Z',
    })
    showError.mockReset()
  })

  afterEach(() => {
    document.body.innerHTML = ''
    vi.restoreAllMocks()
  })

  it('opens the chat image workspace as a pure embedded iframe without preselecting an API key', async () => {
    routeState.query = { prompt: 'draw cat', mode: 'image' }

    mountView()
    await flushPromises()
    await flushPromises()

    expect(launch).toHaveBeenCalledWith()

    const frame = document.body.querySelector('[data-testid="chat-image-embedded-frame"]') as HTMLIFrameElement
    expect(frame).not.toBeNull()

    const url = new URL(frame.src)
    expect(url.origin).toBe('https://chat.example.com')
    expect(url.pathname).toBe('/api/v1/auths/sub2api/launch')
    expect(url.searchParams.get('token')).toBe('one-time')
    expect(url.searchParams.get('ui_mode')).toBe('embedded')
    expect(url.searchParams.get('theme')).toBe('light')
    expect(url.searchParams.get('lang')).toBe('zh')
    expect(url.searchParams.get('prompt')).toBe('draw cat')
    expect(url.searchParams.get('mode')).toBe('image')
  })

  it('does not allow route query to override reserved launch parameters', async () => {
    routeState.query = {
      prompt: 'draw cat',
      mode: 'image',
      token: 'bad-token',
      ui_mode: 'bad-mode',
      lang: 'en',
    }

    mountView()
    await flushPromises()
    await flushPromises()

    const frame = document.body.querySelector('[data-testid="chat-image-embedded-frame"]') as HTMLIFrameElement
    expect(frame).not.toBeNull()

    const url = new URL(frame.src)
    expect(url.searchParams.get('token')).toBe('one-time')
    expect(url.searchParams.get('ui_mode')).toBe('embedded')
    expect(url.searchParams.get('lang')).toBe('zh')
    expect(url.searchParams.get('prompt')).toBe('draw cat')
    expect(launch).toHaveBeenCalledTimes(1)
  })

  it('collapses the sidebar immediately while the embedded workspace is active', async () => {
    const wrapper = mountView()

    expect(appState.setSidebarCollapsed).toHaveBeenCalledWith(true)
    expect(appState.sidebarCollapsed).toBe(true)

    await wrapper.unmount()

    expect(appState.setSidebarCollapsed).toHaveBeenLastCalledWith(false)
    expect(appState.sidebarCollapsed).toBe(false)
  })

  it('shows an embedded error state when launch fails', async () => {
    launch.mockRejectedValue(new Error('launch failed'))

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('chatImageStudio.embeddedOpenFailedTitle')
    expect(showError).toHaveBeenCalledWith('openWebUI.launchFailed')
  })
})
