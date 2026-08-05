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

function roomColor(state: string): number {
  return ({ available: 0x6f9a83, full: 0xbb8065, activating: 0xc6a35a, active: 0x778bb2 } as Record<string, number>)[state] || 0x77757c
}

export async function createCafeRenderer(host: HTMLElement, initialData: CafeSceneData): Promise<CafeSceneRenderer> {
  const { Application, Graphics, Text } = await import('pixi.js')
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
        .fill({ color: roomColor(room.purchase_state), alpha: .86 })
        .rect(toX(hotspot.x + 4), toY(hotspot.y + 4), width - 8 * scale, height - 8 * scale)
        .fill({ color: 0x2e3038, alpha: .88 })
      graphic.eventMode = 'static'
      graphic.cursor = 'pointer'
      graphic.on('pointertap', () => current.onRoomSelect(room))
      app.stage.addChild(graphic)

      const label = new Text({
        text: room.code,
        style: { fill: 0xfff5df, fontFamily: 'monospace', fontSize: Math.max(10, 13 * scale), fontWeight: '700' },
      })
      label.position.set(toX(hotspot.x + 10), toY(hotspot.y + 12))
      app.stage.addChild(label)
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
