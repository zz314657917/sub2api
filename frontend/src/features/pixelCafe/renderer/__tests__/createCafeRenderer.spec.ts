import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { CafePublicRoom } from '@/types/pixelCafe'
import { createCafeRenderer } from '../createCafeRenderer'

const apps = vi.hoisted(() => [] as MockApplication[])

class MockNode {
  children: MockNode[] = []
  y = 0
  x = 0
  eventMode = 'auto'
  cursor = ''
  handlers: Record<string, () => void> = {}

  addChild(...children: MockNode[]): void {
    this.children.push(...children)
  }

  removeChildren(): MockNode[] {
    const children = [...this.children]
    this.children = []
    return children
  }

  destroy(): void {
    this.children = []
  }

  rect(): this { return this }
  fill(): this { return this }
  on(event: string, handler: () => void): void { this.handlers[event] = handler }
  position = { set: (x: number, y: number) => { this.x = x; this.y = y } }
}

class MockApplication {
  canvas = document.createElement('canvas')
  stage = new MockNode()
  renderer = { resize: vi.fn() }
  ticker = { add: vi.fn(), remove: vi.fn() }
  init = vi.fn(async () => undefined)
  destroy = vi.fn()
}

vi.mock('pixi.js', () => ({
  Application: class extends MockApplication {
    constructor() {
      super()
      apps.push(this)
    }
  },
  Graphics: MockNode,
  Text: class extends MockNode {
    constructor() { super() }
  },
}))

const room = {
  id: 18,
  code: 'C-018',
  name: 'Claude 包间 18',
  zone_key: 'claude',
  theme_key: 'warm_wood',
  scene_slot_key: 'claude-room-08',
  featured: true,
  plan: { id: 3, title: 'Claude Max', description: '', price_per_seat: 99, price_label: '99 CNY', validity_days: 30, total_seats: 5 },
  round: { id: 1008, status: 'open', paid_seats: 4, reserved_seats: 0, remaining_seats: 1, deadline_at: '2026-08-03T12:00:00Z' },
  seat_visuals: [],
  purchase_state: 'available',
} satisfies CafePublicRoom

describe('createCafeRenderer', () => {
  beforeEach(() => {
    apps.length = 0
    vi.stubGlobal('ResizeObserver', class {
      observe = vi.fn()
      disconnect = vi.fn()
    })
  })

  it('initializes a bounded canvas, maps hotspots, updates data and destroys owned resources', async () => {
    const host = document.createElement('div')
    Object.defineProperty(host, 'clientWidth', { configurable: true, value: 960 })
    Object.defineProperty(host, 'clientHeight', { configurable: true, value: 540 })
    const onRoomSelect = vi.fn()
    const renderer = await createCafeRenderer(host, {
      rooms: [room],
      lobbyAvatars: [{ avatar_seed: 'opaque-seed', seat_index: 1, activity: 'recent' }],
      reducedMotion: true,
      onRoomSelect,
    })

    const app = apps[0]
    expect(app.init).toHaveBeenCalledWith(expect.objectContaining({ resolution: 1, antialias: false }))
    expect(host.querySelector('canvas')).not.toBeNull()
    const hotspot = app.stage.children.find(child => child.eventMode === 'static')
    hotspot?.handlers.pointertap()
    expect(onRoomSelect).toHaveBeenCalledWith(room)

    renderer.update({ rooms: [], lobbyAvatars: [], reducedMotion: false, onRoomSelect })
    expect(app.stage.children).toHaveLength(0)
    renderer.destroy()
    expect(app.ticker.remove).toHaveBeenCalledTimes(1)
    expect(app.destroy).toHaveBeenCalledWith({ removeView: true }, { children: true })
    expect(host.querySelector('canvas')).toBeNull()
  })
})
