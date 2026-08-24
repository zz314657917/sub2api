import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { CafePublicRoom } from '@/types/pixelCafe'
import CafeScene from '../CafeScene.vue'

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
  it('renders the lobby background without the legacy workstation grid', async () => {
    vi.stubGlobal('matchMedia', vi.fn(() => ({ matches: false, addEventListener: vi.fn(), removeEventListener: vi.fn() })))
    const wrapper = mount(CafeScene, {
      props: {
        rooms: [room],
        lobbyAvatars: [{ avatar_seed: 'opaque-seat-user', seat_index: 17, activity: 'recent' }],
      },
    })
    await flushPromises()

    expect(wrapper.find('.pixel-cafe-scene-art').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="pixel-cafe-workstation"]')).toHaveLength(0)
    expect(wrapper.findAll('[data-testid="pixel-cafe-lobby-avatar"]')).toHaveLength(0)
    expect(wrapper.find('[data-renderer-state="fallback"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="pixel-cafe-room-navigator"]').exists()).toBe(false)
  })
})
