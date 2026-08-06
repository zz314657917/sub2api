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
  graphic: { y: number }
  baseY: number
  phase: number
}

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
      const graphic = new Graphics()
        .rect(toX(seat.x - 8), toY(seat.y - 14), 16 * scale, 18 * scale)
        .fill({ color: avatarColor(avatar.avatar_seed) })
        .rect(toX(seat.x - 5), toY(seat.y - 23), 10 * scale, 10 * scale)
        .fill({ color: 0xf3d0b3 })
      app.stage.addChild(graphic)
      animatedAvatars.push({ graphic, baseY: graphic.y, phase: index * .7 })
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
    const time = performance.now() / 480
    for (const avatar of animatedAvatars) avatar.graphic.y = avatar.baseY + Math.sin(time + avatar.phase) * 1.5
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
