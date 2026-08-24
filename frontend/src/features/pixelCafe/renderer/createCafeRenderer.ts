import type { CafeLobbyAvatar, CafePublicRoom } from '@/types/pixelCafe'
import { CAFE_SCENE_DESIGN_HEIGHT, CAFE_SCENE_DESIGN_WIDTH, CAFE_SCENE_ROOM_LIMIT, getAvatarToneIndex, getLobbySeat, getRoomHotspot } from './sceneLayout'

export interface CafeSceneData {
  rooms: CafePublicRoom[]
  lobbyAvatars: CafeLobbyAvatar[]
  reducedMotion: boolean
  onRoomSelect: (room: CafePublicRoom) => void
}

export interface CafeSceneRenderer {
  update: (data: CafeSceneData) => void
  destroy: () => void
}

interface AnimatedAvatar {
  graphic: { x: number; y: number }
  typingIndicator: { alpha: number }
  start: CafeScenePoint
  target: CafeScenePoint
  phase: number
  pause: number
}

interface CafeScenePoint {
  x: number
  y: number
}

const lobbyWalkPoints: CafeScenePoint[] = [
  { x: 360, y: 245 },
  { x: 470, y: 275 },
  { x: 585, y: 235 },
  { x: 670, y: 305 },
  { x: 505, y: 335 },
]

function avatarColor(seed: string): number {
  return [0xb87565, 0x6f9a83, 0x7b91bb, 0xcb9d59, 0x9d7ab1][getAvatarToneIndex(seed)]
}

export async function createCafeRenderer(host: HTMLElement, initialData: CafeSceneData): Promise<CafeSceneRenderer> {
  const { Application, Graphics } = await import('pixi.js')
  const app = new Application()
  const resolution = Math.min(window.devicePixelRatio || 1, 2)
  const initialWidth = Math.max(host.clientWidth, CAFE_SCENE_DESIGN_WIDTH)
  const initialHeight = Math.max(host.clientHeight, Math.round(initialWidth * CAFE_SCENE_DESIGN_HEIGHT / CAFE_SCENE_DESIGN_WIDTH))
  await app.init({
    width: initialWidth,
    height: initialHeight,
    autoDensity: true,
    antialias: false,
    backgroundAlpha: 0,
    resolution,
  })

  app.canvas.className = 'pixel-cafe-pixi-canvas'
  host.replaceChildren(app.canvas)

  let current = initialData
  let destroyed = false
  let animatedAvatars: AnimatedAvatar[] = []

  const clearStage = (): void => {
    const children = app.stage.removeChildren()
    for (const child of children) child.destroy({ children: true })
    animatedAvatars = []
  }

  const render = (): void => {
    if (destroyed) return
    const width = Math.max(host.clientWidth, CAFE_SCENE_DESIGN_WIDTH)
    const height = Math.max(host.clientHeight, Math.round(width * CAFE_SCENE_DESIGN_HEIGHT / CAFE_SCENE_DESIGN_WIDTH))
    app.renderer.resize(width, height)
    clearStage()

    const scaleX = width / CAFE_SCENE_DESIGN_WIDTH
    const scaleY = height / CAFE_SCENE_DESIGN_HEIGHT
    const scale = Math.min(scaleX, scaleY)
    const offsetX = (width - CAFE_SCENE_DESIGN_WIDTH * scale) / 2
    const offsetY = (height - CAFE_SCENE_DESIGN_HEIGHT * scale) / 2
    const toX = (value: number): number => offsetX + value * scale
    const toY = (value: number): number => offsetY + value * scale

    current.lobbyAvatars.slice(0, CAFE_SCENE_ROOM_LIMIT).forEach((avatar, index) => {
      const seat = getLobbySeat(avatar.seat_index)
      const target = lobbyWalkPoints[index % lobbyWalkPoints.length]
      const graphic = new Graphics()
        .rect(-8 * scale, -10 * scale, 16 * scale, 18 * scale)
        .fill({ color: avatarColor(avatar.avatar_seed) })
        .rect(-5 * scale, -19 * scale, 10 * scale, 10 * scale)
        .fill({ color: 0xf3d0b3 })
      const typingIndicator = new Graphics()
        .rect(8 * scale, -2 * scale, 6 * scale, 2 * scale)
        .fill({ color: 0xf3d0b3 })
      graphic.addChild(typingIndicator)
      app.stage.addChild(graphic)
      graphic.position.set(toX(seat.x), toY(seat.y - 4))
      animatedAvatars.push({
        graphic,
        typingIndicator,
        start: { x: toX(seat.x), y: toY(seat.y - 4) },
        target: { x: toX(target.x), y: toY(target.y) },
        phase: index * .9,
        pause: index % 3 === 0 ? 1.4 : .4,
      })
    })

    current.rooms.slice(0, CAFE_SCENE_ROOM_LIMIT).forEach((room, index) => {
      const hotspot = getRoomHotspot(room.scene_slot_key, index)
      const width = hotspot.width * scale
      const height = hotspot.height * scale
      const graphic = new Graphics()
        .rect(toX(hotspot.x), toY(hotspot.y), width, height)
        .fill({ color: 0xffffff, alpha: .01 })
      graphic.eventMode = 'static'
      graphic.cursor = 'pointer'
      graphic.on('pointertap', () => current.onRoomSelect(room))
      app.stage.addChild(graphic)
    })
  }

  const animate = (): void => {
    if (destroyed || current.reducedMotion) return
    const time = performance.now() / 1000
    for (const avatar of animatedAvatars) {
      const cycle = 8 + avatar.pause
      const progress = ((time + avatar.phase) % cycle) / cycle
      let from = avatar.start
      let to = avatar.start
      let walkProgress = 0
      const isTyping = progress < .16
      if (progress < .16) {
        from = avatar.start
        to = avatar.start
      } else if (progress < .66) {
        from = avatar.start
        to = avatar.target
        walkProgress = (progress - .16) / .5
      } else if (progress < .84) {
        from = avatar.target
        to = avatar.target
      } else {
        from = avatar.target
        to = avatar.start
        walkProgress = (progress - .84) / .16
      }
      const eased = walkProgress * walkProgress * (3 - 2 * walkProgress)
      avatar.graphic.x = from.x + (to.x - from.x) * eased
      avatar.graphic.y = from.y + (to.y - from.y) * eased + Math.sin(time * 7 + avatar.phase) * 1.5
      avatar.typingIndicator.alpha = isTyping ? 1 : 0
    }
  }

  app.ticker.add(animate)
  const resizeObserver = typeof ResizeObserver === 'undefined' ? undefined : new ResizeObserver(render)
  resizeObserver?.observe(host)
  render()

  return {
    update(data: CafeSceneData): void {
      current = data
      render()
    },
    destroy(): void {
      if (destroyed) return
      destroyed = true
      resizeObserver?.disconnect()
      app.ticker.remove(animate)
      clearStage()
      app.destroy({ removeView: true }, { children: true })
      host.replaceChildren()
    },
  }
}
