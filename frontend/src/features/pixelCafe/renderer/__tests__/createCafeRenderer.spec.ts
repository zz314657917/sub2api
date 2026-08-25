import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { CafePublicRoom } from '@/types/pixelCafe'
import { createCafeRenderer } from '../createCafeRenderer'
import {
  CAFE_SCENE_WALK_ROUTES,
  CAFE_SCENE_WORKSTATIONS,
  createCafeWorkstationLayout,
  getCafeSceneCoverTransform,
  resizeCafeWorkstationLayout,
  resolveCafeWorkstationLayout,
} from '../sceneLayout'

const apps = vi.hoisted(() => [] as MockApplication[])

class MockNode {
  children: MockNode[] = []
  y = 0
  x = 0
  zIndex = 0
  width = 0
  height = 0
  sortableChildren = false
  eventMode = 'auto'
  cursor = ''
  texture: unknown
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
  anchor = { set: vi.fn() }
  scale = { x: 1, y: 1 }
}

class MockApplication {
  canvas = document.createElement('canvas')
  stage = new MockNode()
  renderer = { resize: vi.fn() }
  tickerCallback: (() => void) | undefined
  ticker = {
    autoStart: false,
    add: vi.fn((callback: () => void) => { this.tickerCallback = callback }),
    remove: vi.fn(),
    start: vi.fn(),
    stop: vi.fn(),
  }
  init = vi.fn(async () => undefined)
  destroy = vi.fn()
}

vi.mock('pixi.js', () => ({
  Assets: { load: vi.fn(async (url: string) => ({ url })) },
  Application: class extends MockApplication {
    constructor() {
      super()
      apps.push(this)
    }
  },
  Container: MockNode,
  Graphics: MockNode,
  Sprite: class extends MockNode {
    constructor(texture?: unknown) { super(); this.texture = texture }
  },
  Text: class extends MockNode {
    constructor() { super() }
  },
}))
vi.mock('pixi.js/unsafe-eval', () => ({}))

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

  it('uses the lobby cover transform and accepts only bounded contiguous variable layouts', () => {
    expect(getCafeSceneCoverTransform(1000, 500)).toEqual({
      scale: 1000 / 960,
      offsetX: 0,
      offsetY: (500 - 540 * (1000 / 960)) / 2,
    })
    expect(resolveCafeWorkstationLayout([{ id: 1, x: 300, y: 200 }])).toEqual([{ id: 1, x: 300, y: 200 }])
    expect(resolveCafeWorkstationLayout([{ id: 2, x: 300, y: 200 }])).toEqual(CAFE_SCENE_WORKSTATIONS)

    const movedDefaults = CAFE_SCENE_WORKSTATIONS.map(slot => ({ ...slot }))
    movedDefaults[0].x = 401
    const fifty = resizeCafeWorkstationLayout(movedDefaults, 50)
    expect(fifty).toHaveLength(50)
    expect(fifty.slice(0, 10)).toEqual(movedDefaults)
    expect(new Set(fifty.map(slot => `${slot.x}:${slot.y}`)).size).toBe(50)
    expect(resolveCafeWorkstationLayout(fifty)).toEqual(fifty)
    expect(createCafeWorkstationLayout(1)).toEqual([CAFE_SCENE_WORKSTATIONS[0]])
  })

  it('keeps every slow walking segment inside the open aisles and outside furniture bounds', () => {
    const sampledPoints = CAFE_SCENE_WALK_ROUTES.flatMap(route => route.slice(1).flatMap((point, index) => {
      const start = route[index]
      return Array.from({ length: 21 }, (_, step) => ({
        x: start.x + (point.x - start.x) * step / 20,
        y: start.y + (point.y - start.y) * step / 20,
      }))
    }))

    for (const point of sampledPoints) {
      expect(point.x).toBeGreaterThanOrEqual(220)
      expect(point.x).toBeLessThanOrEqual(840)
      expect(point.y).toBeGreaterThanOrEqual(200)
      expect(point.y).toBeLessThanOrEqual(430)
      for (const workstation of CAFE_SCENE_WORKSTATIONS) {
        const intersectsFurniture = Math.abs(point.x - workstation.x) < 46
          && point.y > workstation.y - 76
          && point.y < workstation.y + 36
        expect(intersectsFurniture).toBe(false)
      }
    }
  })

  it('starts one ticker, advances a deterministic walking avatar, and destroys owned resources', async () => {
    const host = document.createElement('div')
    Object.defineProperty(host, 'clientWidth', { configurable: true, value: 960 })
    Object.defineProperty(host, 'clientHeight', { configurable: true, value: 540 })
    const onRoomSelect = vi.fn()
    const lobbyAvatars = Array.from({ length: 11 }, (_, seat_index) => ({
      avatar_seed: `scene-avatar-${seat_index}`,
      seat_index,
      activity: 'recent' as const,
    }))
    const renderer = await createCafeRenderer(host, {
      rooms: [room],
      lobbyAvatars,
      workstations: CAFE_SCENE_WORKSTATIONS,
      reducedMotion: false,
      onRoomSelect,
    })

    const app = apps[0]
    expect(app.init).toHaveBeenCalledWith(expect.objectContaining({ resolution: 1, antialias: false, autoStart: true, sharedTicker: false }))
    expect(host.querySelector('canvas')).not.toBeNull()
    expect(app.stage.children).toHaveLength(22)
    expect(app.ticker.add).toHaveBeenCalledTimes(1)
    expect(app.ticker.start).toHaveBeenCalledTimes(1)
    const hotspot = app.stage.children.find(child => child.eventMode === 'static')
    hotspot?.handlers.pointertap()
    expect(onRoomSelect).toHaveBeenCalledWith(room)

    const walkingAvatar = app.stage.children[20]
    const walkingSprite = walkingAvatar.children[0]
    expect(walkingSprite.texture).toEqual(expect.objectContaining({ url: expect.stringContaining('walk-1') }))
    const start = { x: walkingAvatar.x, y: walkingAvatar.y }
    const performanceNow = vi.spyOn(performance, 'now').mockReturnValue(15000)
    app.tickerCallback?.()
    expect({ x: walkingAvatar.x, y: walkingAvatar.y }).not.toEqual(start)
    expect(walkingSprite.texture).toEqual(expect.objectContaining({ url: expect.stringContaining('walk-2') }))
    expect(walkingSprite.scale.x).toBeGreaterThan(0)

    performanceNow.mockReturnValue(40000)
    app.tickerCallback?.()
    expect(walkingSprite.texture).toEqual(expect.objectContaining({ url: expect.stringContaining('walk-0') }))
    expect(walkingSprite.scale.x).toBeLessThan(0)

    renderer.update({ rooms: [], lobbyAvatars, workstations: CAFE_SCENE_WORKSTATIONS, reducedMotion: true, onRoomSelect })
    const reducedMotionAvatar = app.stage.children[20]
    expect(reducedMotionAvatar.children[0].texture).toEqual(expect.objectContaining({ url: expect.stringContaining('walk-1') }))
    const frozen = { x: reducedMotionAvatar.x, y: reducedMotionAvatar.y }
    performanceNow.mockReturnValue(2800)
    app.tickerCallback?.()
    expect({ x: reducedMotionAvatar.x, y: reducedMotionAvatar.y }).toEqual(frozen)
    renderer.destroy()
    expect(app.ticker.remove).toHaveBeenCalledTimes(1)
    expect(app.ticker.stop).toHaveBeenCalledTimes(1)
    expect(app.destroy).toHaveBeenCalledWith({ removeView: true }, { children: true })
    expect(host.querySelector('canvas')).toBeNull()
  })

  it('uses the saved workstation count as seated capacity and keeps six extra walkers at most', async () => {
    const host = document.createElement('div')
    Object.defineProperty(host, 'clientWidth', { configurable: true, value: 960 })
    Object.defineProperty(host, 'clientHeight', { configurable: true, value: 540 })
    const workstations = createCafeWorkstationLayout(3)
    const lobbyAvatars = Array.from({ length: 12 }, (_, seat_index) => ({
      avatar_seed: `dynamic-avatar-${seat_index}`,
      seat_index,
      activity: 'recent' as const,
    }))

    const renderer = await createCafeRenderer(host, {
      rooms: [room],
      lobbyAvatars,
      workstations,
      reducedMotion: false,
      onRoomSelect: vi.fn(),
    })

    const app = apps[0]
    expect(app.stage.children).toHaveLength(3 + 9 + 1)
    const firstWalkingAvatar = app.stage.children[3 + 3]
    expect(firstWalkingAvatar.children[0].texture).toEqual(expect.objectContaining({ url: expect.stringContaining('walk-1') }))
    renderer.destroy()
  })
})
