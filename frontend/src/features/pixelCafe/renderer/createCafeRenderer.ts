import type { CafeLobbyAvatar, CafePublicRoom } from '@/types/pixelCafe'
import type { Container as PixiContainer, Graphics as PixiGraphics, Sprite as PixiSprite, Texture as PixiTexture } from 'pixi.js'
import { cafeSceneAssets } from './assetManifest'
import {
  CAFE_SCENE_DESIGN_HEIGHT,
  CAFE_SCENE_DESIGN_WIDTH,
  CAFE_SCENE_ROOM_LIMIT,
  CAFE_SCENE_WALKING_AVATAR_COUNT,
  type CafeWorkstationSlot,
  getCafeSceneCoverTransform,
  getAvatarToneIndex,
  getCafeWalkRoute,
  getLobbySeat,
  getRoomHotspot,
  resolveCafeWorkstationLayout,
} from './sceneLayout'

export interface CafeSceneData {
  rooms: CafePublicRoom[]
  lobbyAvatars: CafeLobbyAvatar[]
  workstations: CafeWorkstationSlot[]
  reducedMotion: boolean
  onRoomSelect: (room: CafePublicRoom) => void
}

export interface CafeSceneRenderer {
  update: (data: CafeSceneData) => void
  destroy: () => void
}

interface AnimatedAvatar {
  graphic: { x: number; y: number; zIndex: number }
  sprite: PixiSprite
  typingIndicator: { alpha: number }
  walkTextures: readonly PixiTexture[]
  route: CafeScenePoint[]
  phaseSeconds: number
  cycleSeconds: number
  walking: boolean
  baseScaleX: number
  frameIndex: number
}

interface CafeScenePoint {
  x: number
  y: number
}

type GraphicsConstructor = new () => PixiGraphics
type ContainerConstructor = new () => PixiContainer
type SpriteConstructor = new (texture?: PixiTexture) => PixiSprite

function getAvatarTextureIndex(seed: string): number {
  return getAvatarToneIndex(seed) % cafeSceneAssets.avatars.length
}

function createAvatar(
  Container: ContainerConstructor,
  Graphics: GraphicsConstructor,
  Sprite: SpriteConstructor,
  texture: PixiTexture,
  aspectRatio: number,
  scale: number,
): { graphic: PixiContainer; sprite: PixiSprite; typingIndicator: PixiGraphics } {
  const graphic = new Container()
  const sprite = new Sprite(texture)
  sprite.anchor.set(.5, 1)
  sprite.height = 48 * scale
  sprite.width = 48 * aspectRatio * scale
  graphic.addChild(sprite)
  const typingIndicator = new Graphics()
    .rect(8 * scale, -24 * scale, 9 * scale, 3 * scale)
    .fill({ color: 0x8ff0d2, alpha: 1 })
    .rect(10 * scale, -22 * scale, 3 * scale, 2 * scale)
    .fill({ color: 0xf4ce7c, alpha: 1 })
  graphic.addChild(typingIndicator)
  return { graphic, sprite, typingIndicator }
}

function pointOnRoute(route: CafeScenePoint[], progress: number): CafeScenePoint {
  if (route.length < 2) return route[0] ?? { x: 0, y: 0 }
  const segmentLengths = route.slice(1).map((point, index) => Math.hypot(point.x - route[index].x, point.y - route[index].y))
  const totalLength = segmentLengths.reduce((total, length) => total + length, 0)
  let remaining = Math.max(0, Math.min(1, progress)) * totalLength
  for (let index = 0; index < segmentLengths.length; index += 1) {
    const length = segmentLengths[index]
    if (remaining <= length || index === segmentLengths.length - 1) {
      const local = length === 0 ? 0 : remaining / length
      return {
        x: route[index].x + (route[index + 1].x - route[index].x) * local,
        y: route[index].y + (route[index + 1].y - route[index].y) * local,
      }
    }
    remaining -= length
  }
  return route[route.length - 1]
}

export async function createCafeRenderer(host: HTMLElement, initialData: CafeSceneData): Promise<CafeSceneRenderer> {
  // Pixi's official compatibility module replaces the unsafe-eval-only paths.
  // It must evaluate before Application.init() in a strict Content Security Policy.
  // @ts-expect-error Pixi v8 exports this documented side-effect module without a d.ts entry.
  await import('pixi.js/unsafe-eval')
  const { Application, Assets, Container, Graphics, Sprite } = await import('pixi.js')
  const avatarTextures = await Promise.all(cafeSceneAssets.avatars.map(asset => Assets.load(asset.url)))
  const avatarWalkTextures = await Promise.all(cafeSceneAssets.avatars.map(asset => Promise.all(
    asset.walkFrames.map(frameUrl => Assets.load(frameUrl)),
  )))
  const workstationTextures = await Promise.all(cafeSceneAssets.workstations.map(asset => Assets.load(asset.url)))
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
    autoStart: true,
    sharedTicker: false,
  })

  app.canvas.className = 'pixel-cafe-pixi-canvas'
  host.replaceChildren(app.canvas)

  let current = initialData
  let destroyed = false
  let animatedAvatars: AnimatedAvatar[] = []
  app.stage.sortableChildren = true

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

    const { scale, offsetX, offsetY } = getCafeSceneCoverTransform(width, height)
    const toX = (value: number): number => offsetX + value * scale
    const toY = (value: number): number => offsetY + value * scale

    const workstations = resolveCafeWorkstationLayout(current.workstations)
    for (const [index, slot] of workstations.entries()) {
      const assetIndex = index % workstationTextures.length
      const workstation = new Sprite(workstationTextures[assetIndex])
      const asset = cafeSceneAssets.workstations[assetIndex]
      workstation.anchor.set(.5, .92)
      workstation.width = 78 * scale
      workstation.height = 78 / asset.aspectRatio * scale
      workstation.position.set(toX(slot.x), toY(slot.y))
      workstation.zIndex = toY(slot.y)
      app.stage.addChild(workstation)
    }

    const seatedAvatarCount = workstations.length
    const avatarLimit = seatedAvatarCount + CAFE_SCENE_WALKING_AVATAR_COUNT
    current.lobbyAvatars.slice(0, avatarLimit).forEach((avatar, index) => {
      const assetIndex = getAvatarTextureIndex(avatar.avatar_seed)
      const asset = cafeSceneAssets.avatars[assetIndex]
      const walking = index >= seatedAvatarCount
      const designRoute = walking ? [...getCafeWalkRoute(index - seatedAvatarCount)] : [getLobbySeat(avatar.seat_index, workstations)]
      const route = designRoute.map(point => ({ x: toX(point.x), y: toY(point.y) }))
      const walkTextures = avatarWalkTextures[assetIndex]
      const initialTexture = walking ? (walkTextures[1] ?? walkTextures[0]) : avatarTextures[assetIndex]
      const aspectRatio = walking ? asset.walkFrameAspectRatio : asset.aspectRatio
      const { graphic, sprite, typingIndicator } = createAvatar(Container, Graphics, Sprite, initialTexture, aspectRatio, scale)
      app.stage.addChild(graphic)
      graphic.position.set(route[0].x, route[0].y)
      graphic.zIndex = route[0].y + 1
      typingIndicator.alpha = walking ? 0 : 1
      animatedAvatars.push({
        graphic,
        sprite,
        typingIndicator,
        walkTextures,
        route,
        phaseSeconds: (index - seatedAvatarCount) * 6.5,
        cycleSeconds: 40 + (index % 3) * 4,
        walking,
        baseScaleX: Math.abs(sprite.scale.x),
        frameIndex: walking ? 1 : -1,
      })
    })

    current.rooms.slice(0, CAFE_SCENE_ROOM_LIMIT).forEach((room, index) => {
      const hotspot = getRoomHotspot(room.scene_slot_key, index)
      const width = hotspot.width * scale
      const height = hotspot.height * scale
      const graphic = new Graphics()
        .rect(toX(hotspot.x), toY(hotspot.y), width, height)
        .fill({ color: 0xffffff, alpha: .01 })
      graphic.zIndex = 10_000
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
      if (!avatar.walking) continue
      const progress = ((time + avatar.phaseSeconds) % avatar.cycleSeconds) / avatar.cycleSeconds
      let routeProgress = 0
      let isWalking = false
      if (progress >= .22 && progress < .56) {
        routeProgress = (progress - .22) / .34
        isWalking = true
      } else if (progress >= .72) {
        routeProgress = 1 - (progress - .72) / .28
        isWalking = true
      } else if (progress >= .56) {
        routeProgress = 1
      }
      const point = pointOnRoute(avatar.route, routeProgress)
      const deltaX = point.x - avatar.graphic.x
      avatar.graphic.x = point.x
      avatar.graphic.y = point.y
      avatar.graphic.zIndex = avatar.graphic.y + 1
      avatar.typingIndicator.alpha = 0
      if (Math.abs(deltaX) > .01) {
        avatar.sprite.scale.x = avatar.baseScaleX * (deltaX > 0 ? 1 : -1)
      }
      const nextFrameIndex = isWalking
        ? Math.floor((time + avatar.phaseSeconds) * 6) % avatar.walkTextures.length
        : Math.min(1, avatar.walkTextures.length - 1)
      if (nextFrameIndex >= 0 && nextFrameIndex !== avatar.frameIndex) {
        avatar.sprite.texture = avatar.walkTextures[nextFrameIndex]
        avatar.frameIndex = nextFrameIndex
      }
    }
  }

  // A private, explicit ticker avoids relying on page-level shared ticker state.
  app.ticker.autoStart = true
  app.ticker.add(animate)
  app.ticker.start()
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
      app.ticker.stop()
      clearStage()
      app.destroy({ removeView: true }, { children: true })
      host.replaceChildren()
    },
  }
}
