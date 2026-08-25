import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { CafePublicRoom } from '@/types/pixelCafe'
import CafeScene from '../CafeScene.vue'

const createCafeRenderer = vi.hoisted(() => vi.fn())
vi.mock('../../renderer/createCafeRenderer', () => ({ createCafeRenderer }))

const room = {
  id: 18,
  code: 'C-018',
  name: 'Claude 包间 18',
  zone_key: 'claude',
  theme_key: 'warm_wood',
  scene_slot_key: 'claude-18',
  featured: true,
  plan: { id: 3, title: 'Claude Max', description: '', price_per_seat: 99, price_label: '99 CNY', validity_days: 30, total_seats: 5 },
  round: { id: 1008, status: 'open', paid_seats: 4, reserved_seats: 0, remaining_seats: 1, deadline_at: '2026-08-03T12:00:00Z' },
  seat_visuals: [],
  purchase_state: 'available',
} satisfies CafePublicRoom

describe('CafeScene', () => {
  it('renders a static workstation and avatar fallback when Pixi is unavailable', async () => {
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: false, addEventListener: vi.fn(), removeEventListener: vi.fn() })))
    const wrapper = mount(CafeScene, {
      props: {
        rooms: [room],
        lobbyAvatars: [{ avatar_seed: 'opaque-seat-user', seat_index: 17, activity: 'recent' }],
      },
    })
    await flushPromises()

    expect(wrapper.find('.pixel-cafe-scene-art').exists()).toBe(true)
    expect(wrapper.find('[data-renderer-state="fallback"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="pixel-cafe-scene-fallback"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="pixel-cafe-fallback-workstation"]')).toHaveLength(10)
    expect(wrapper.findAll('[data-testid="pixel-cafe-fallback-avatar"]')).toHaveLength(1)
    expect(wrapper.find('[data-testid="pixel-cafe-fallback-workstation"]').attributes('src')).toContain('workstation')
    expect(wrapper.find('[data-testid="pixel-cafe-fallback-avatar"]').attributes('src')).toContain('avatar')
    expect(wrapper.find('[data-testid="pixel-cafe-room-navigator"]').exists()).toBe(false)
  })

  it('uses a variable saved workstation layout for furniture and seated/walking avatars', async () => {
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: false, addEventListener: vi.fn(), removeEventListener: vi.fn() })))
    const workstations = Array.from({ length: 3 }, (_, index) => ({
      id: index + 1,
      x: 300 + index * 20,
      y: 200 + index * 10,
    }))
    const lobbyAvatars = Array.from({ length: 10 }, (_, seat_index) => ({
      avatar_seed: `saved-layout-user-${seat_index}`,
      seat_index,
      activity: 'recent' as const,
    }))
    const wrapper = mount(CafeScene, {
      props: {
        rooms: [room],
        workstations,
        lobbyAvatars,
      },
    })
    await flushPromises()

    const workstation = wrapper.find('[data-testid="pixel-cafe-fallback-workstation"]')
    const avatar = wrapper.find('[data-testid="pixel-cafe-fallback-avatar"]')
    expect(wrapper.findAll('[data-testid="pixel-cafe-fallback-workstation"]')).toHaveLength(3)
    expect(wrapper.findAll('[data-testid="pixel-cafe-fallback-avatar"]')).toHaveLength(9)
    expect(wrapper.findAll('.pixel-cafe-fallback-avatar.walking')).toHaveLength(6)
    expect(workstation.attributes('style')).toContain('left: 31.25%')
    expect(workstation.attributes('style')).toContain('top: 37.037')
    expect(avatar.attributes('style')).toContain('left: 29.479')
    expect(avatar.attributes('style')).toContain('top: 37.777')
  })

  it('syncs avatars that arrive while the async renderer is still initializing', async () => {
    vi.stubGlobal('navigator', { userAgent: 'Chrome' })
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: false, addEventListener: vi.fn(), removeEventListener: vi.fn() })))
    const update = vi.fn()
    const destroy = vi.fn()
    let resolveRenderer: (value: { update: typeof update; destroy: typeof destroy }) => void = () => undefined
    createCafeRenderer.mockImplementationOnce(() => new Promise(resolve => { resolveRenderer = resolve }))

    const wrapper = mount(CafeScene, { props: { rooms: [room], lobbyAvatars: [] } })
    await flushPromises()
    const avatars = Array.from({ length: 50 }, (_, seat_index) => ({ avatar_seed: `late-avatar-${seat_index}`, seat_index, activity: 'recent' as const }))
    await wrapper.setProps({ lobbyAvatars: avatars })

    resolveRenderer({ update, destroy })
    await flushPromises()

    expect(update).toHaveBeenCalledWith(expect.objectContaining({ lobbyAvatars: avatars }))
    expect(wrapper.find('[data-renderer-state="ready"]').exists()).toBe(true)
    wrapper.unmount()
    expect(destroy).toHaveBeenCalledTimes(1)
    vi.unstubAllGlobals()
  })

  it('falls back after a stalled renderer and destroys a renderer that resolves late', async () => {
    vi.useFakeTimers()
    vi.stubGlobal('navigator', { userAgent: 'Chrome' })
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: false, addEventListener: vi.fn(), removeEventListener: vi.fn() })))
    const update = vi.fn()
    const destroy = vi.fn()
    let resolveRenderer: (value: { update: typeof update; destroy: typeof destroy }) => void = () => undefined
    createCafeRenderer.mockImplementationOnce(() => new Promise(resolve => { resolveRenderer = resolve }))

    const wrapper = mount(CafeScene, {
      props: {
        rooms: [room],
        lobbyAvatars: [{ avatar_seed: 'timeout-user', seat_index: 1, activity: 'recent' }],
      },
    })
    await flushPromises()
    expect(wrapper.find('[data-renderer-state="loading"]').exists()).toBe(true)

    await vi.advanceTimersByTimeAsync(8_000)
    await flushPromises()
    expect(wrapper.find('[data-renderer-state="fallback"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="pixel-cafe-fallback-avatar"]')).toHaveLength(1)

    resolveRenderer({ update, destroy })
    await flushPromises()
    expect(destroy).toHaveBeenCalledTimes(1)
    expect(update).not.toHaveBeenCalled()

    wrapper.unmount()
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })
})
