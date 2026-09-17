import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const metadata = vi.hoisted(() => ({ enabled: false, displayName: '鹈鹕测试', loaded: true, load: vi.fn(async () => {}), reset: vi.fn() }))
const authAPI = vi.hoisted(() => ({ getCurrentUser: vi.fn(async () => ({ data: { id: 7, role: 'user' } })) }))

vi.mock('@/stores/pelicanMetadata', () => ({ usePelicanMetadataStore: () => metadata }))
vi.mock('@/api/setup', () => ({ getSetupStatus: vi.fn() }))
vi.mock('@/api', () => ({ authAPI, isTotp2FARequired: () => false, passkeyAPI: {} }))

describe('Pelican feature switch router guard', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.clearAllMocks()
    vi.stubGlobal('scrollTo', vi.fn())
    setActivePinia(createPinia())
    localStorage.clear()
    localStorage.setItem('auth_token', 'test-token')
    localStorage.setItem('auth_user', JSON.stringify({ id: 7, role: 'user' }))
    metadata.enabled = false
    metadata.loaded = true
  })

  it('redirects an ordinary user when disabled, permits admin inspection, and permits again after enable', async () => {
    const { default: router } = await import('@/router')
    await router.push('/pelican-tests')
    await router.isReady()
    expect(router.currentRoute.value.path).toBe('/dashboard')

    const { useAuthStore } = await import('@/stores/auth')
    const auth = useAuthStore()
    auth.user = { ...auth.user!, role: 'admin' }
    await router.push('/pelican-tests')
    expect(router.currentRoute.value.name).toBe('PelicanTests')

    auth.user = { ...auth.user!, role: 'user' }
    metadata.enabled = true
    await router.push('/dashboard')
    await router.push('/pelican-tests')
    expect(router.currentRoute.value.name).toBe('PelicanTests')
  })
})
