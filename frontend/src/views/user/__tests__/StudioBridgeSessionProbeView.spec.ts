import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import StudioBridgeSessionProbeView from '../StudioBridgeSessionProbeView.vue'

const getAuthToken = vi.hoisted(() => vi.fn())
const sessionProbe = vi.hoisted(() => vi.fn())

vi.mock('@/api', () => ({
  authAPI: {
    getAuthToken,
  },
  studioBridgeAPI: {
    sessionProbe,
  },
}))

function mountProbe() {
  window.history.pushState(
    {},
    '',
    '/studio-bridge/session-probe?app_id=luoye-ai&parent_origin=http%3A%2F%2F127.0.0.1%3A8081',
  )
  return mount(StudioBridgeSessionProbeView, {
    attachTo: document.body,
  })
}

describe('StudioBridgeSessionProbeView', () => {
  let originalParent: WindowProxy
  let postMessage: ReturnType<typeof vi.fn>

  beforeEach(() => {
    originalParent = window.parent
    postMessage = vi.fn()
    Object.defineProperty(window, 'parent', {
      configurable: true,
      value: { postMessage },
    })
    getAuthToken.mockReset()
    sessionProbe.mockReset()
  })

  afterEach(() => {
    document.body.innerHTML = ''
    Object.defineProperty(window, 'parent', {
      configurable: true,
      value: originalParent,
    })
    window.history.pushState({}, '', '/')
    vi.restoreAllMocks()
  })

  it('posts an unauthenticated bridge session when Sub2API has no local token', async () => {
    getAuthToken.mockReturnValue(null)

    mountProbe()
    await flushPromises()

    expect(sessionProbe).not.toHaveBeenCalled()
    expect(postMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'sub2api:studio-bridge-session',
        app_id: 'luoye-ai',
        authenticated: false,
      }),
      'http://127.0.0.1:8081',
    )
  })

  it('posts the current Sub2API user id when authenticated', async () => {
    getAuthToken.mockReturnValue('token')
    sessionProbe.mockResolvedValue({ user_id: 42 })

    mountProbe()
    await flushPromises()

    expect(sessionProbe).toHaveBeenCalledWith('http://127.0.0.1:8081')
    expect(postMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'sub2api:studio-bridge-session',
        app_id: 'luoye-ai',
        authenticated: true,
        user_id: '42',
      }),
      'http://127.0.0.1:8081',
    )
  })

  it('does not expose a user id when the parent origin is rejected', async () => {
    getAuthToken.mockReturnValue('token')
    sessionProbe.mockRejectedValue({ status: 400 })

    mountProbe()
    await flushPromises()

    expect(postMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'sub2api:studio-bridge-session',
        app_id: 'luoye-ai',
        authenticated: false,
        error: 'probe_failed',
      }),
      'http://127.0.0.1:8081',
    )
  })
})
