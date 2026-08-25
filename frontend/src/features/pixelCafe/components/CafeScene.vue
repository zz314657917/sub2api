<template>
  <div class="pixel-cafe-scene-stage" :data-renderer-state="rendererState" data-testid="pixel-cafe-scene-stage">
    <div class="pixel-cafe-scene-visual">
      <img class="pixel-cafe-scene-art" :src="cafeSceneAssets.lobbyBackground" alt="" aria-hidden="true" />
      <div ref="canvasHost" class="pixel-cafe-scene-canvas-host" aria-hidden="true"></div>
      <div v-if="rendererState === 'fallback'" class="pixel-cafe-scene-fallback" data-testid="pixel-cafe-scene-fallback" aria-label="像素网吧静态场景预览">
        <img
          v-for="(slot, index) in resolvedWorkstations"
          :key="`desk-${slot.id}`"
          class="pixel-cafe-fallback-workstation"
          data-testid="pixel-cafe-fallback-workstation"
          :src="cafeSceneAssets.workstations[index % cafeSceneAssets.workstations.length].url"
          alt=""
          :style="sceneStyle(slot.x, slot.y)"
          aria-hidden="true"
        />
        <img
          v-for="avatar in fallbackAvatars"
          :key="`${avatar.avatar_seed}-${avatar.seat_index}`"
          :class="['pixel-cafe-fallback-avatar', { walking: avatar.walking }]"
          data-testid="pixel-cafe-fallback-avatar"
          :src="avatar.asset.url"
          alt=""
          :style="sceneStyle(avatar.position.x, avatar.position.y)"
          aria-hidden="true"
        />
      </div>
    </div>
    <p v-if="rendererState === 'loading'" class="pixel-cafe-scene-state">场景加载中...</p>
    <p v-else-if="rendererState === 'fallback'" class="pixel-cafe-scene-state">场景暂不可用，房间列表仍可正常使用。</p>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import type { CafeLobbyAvatar, CafePublicRoom, CafeWorkstationPosition } from '@/types/pixelCafe'
import { cafeSceneAssets } from '../renderer/assetManifest'
import { createCafeRenderer, type CafeSceneRenderer } from '../renderer/createCafeRenderer'
import {
  CAFE_SCENE_DESIGN_HEIGHT,
  CAFE_SCENE_DESIGN_WIDTH,
  CAFE_SCENE_WALKING_AVATAR_COUNT,
  type CafeWorkstationSlot,
  getAvatarToneIndex,
  getCafeWalkRoute,
  getLobbySeat,
  resolveCafeWorkstationLayout,
} from '../renderer/sceneLayout'

const props = withDefaults(defineProps<{
  rooms: CafePublicRoom[]
  lobbyAvatars: CafeLobbyAvatar[]
  workstations?: CafeWorkstationPosition[]
}>(), { workstations: () => [] })

const emit = defineEmits<{
  'select-room': [room: CafePublicRoom]
}>()

const canvasHost = ref<HTMLElement>()
const rendererState = ref<'loading' | 'ready' | 'fallback'>('loading')
const reducedMotion = ref(false)
let renderer: CafeSceneRenderer | undefined
let initialization = 0
let mediaQuery: MediaQueryList | undefined
const RENDERER_INITIALIZATION_TIMEOUT_MS = 8_000
const resolvedWorkstations = computed<CafeWorkstationSlot[]>(() => resolveCafeWorkstationLayout(props.workstations))

const rendererData = computed(() => ({
  rooms: props.rooms,
  lobbyAvatars: props.lobbyAvatars,
  workstations: resolvedWorkstations.value,
  reducedMotion: reducedMotion.value,
  onRoomSelect: (room: CafePublicRoom) => emit('select-room', room),
}))

const fallbackAvatars = computed(() => props.lobbyAvatars.slice(0, resolvedWorkstations.value.length + CAFE_SCENE_WALKING_AVATAR_COUNT).map((avatar, index) => {
  const seatedAvatarCount = resolvedWorkstations.value.length
  const walking = index >= seatedAvatarCount
  return {
    ...avatar,
    walking,
    asset: cafeSceneAssets.avatars[getAvatarToneIndex(avatar.avatar_seed) % cafeSceneAssets.avatars.length],
    position: walking ? getCafeWalkRoute(index - seatedAvatarCount)[0] : getLobbySeat(avatar.seat_index, resolvedWorkstations.value),
  }
}))

function sceneStyle(x: number, y: number): Record<string, string> {
  return {
    left: `${x / CAFE_SCENE_DESIGN_WIDTH * 100}%`,
    top: `${y / CAFE_SCENE_DESIGN_HEIGHT * 100}%`,
  }
}

function updateMotionPreference(): void {
  reducedMotion.value = mediaQuery?.matches ?? false
}

function waitForRenderer(rendererPromise: Promise<CafeSceneRenderer>): Promise<CafeSceneRenderer> {
  return new Promise((resolve, reject) => {
    let timedOut = false
    const timeout = window.setTimeout(() => {
      timedOut = true
      reject(new Error(`scene renderer initialization exceeded ${RENDERER_INITIALIZATION_TIMEOUT_MS}ms`))
    }, RENDERER_INITIALIZATION_TIMEOUT_MS)

    rendererPromise.then((nextRenderer) => {
      if (timedOut) {
        nextRenderer.destroy()
        return
      }
      window.clearTimeout(timeout)
      resolve(nextRenderer)
    }, (error) => {
      if (timedOut) return
      window.clearTimeout(timeout)
      reject(error)
    })
  })
}

async function initializeRenderer(): Promise<void> {
  const host = canvasHost.value
  if (!host) return
  const attempt = ++initialization
  rendererState.value = 'loading'
  if (typeof navigator !== 'undefined' && /jsdom/i.test(navigator.userAgent)) {
    rendererState.value = 'fallback'
    return
  }
  renderer?.destroy()
  renderer = undefined
  try {
    const nextRenderer = await waitForRenderer(createCafeRenderer(host, rendererData.value))
    if (attempt !== initialization) {
      nextRenderer.destroy()
      return
    }
    renderer = nextRenderer
    // Props can change while Pixi and WebGL initialize. The watcher has no
    // renderer to update during that interval, so synchronise the latest data
    // before exposing the ready scene instead of retaining the empty first frame.
    renderer.update(rendererData.value)
    rendererState.value = 'ready'
  } catch (error) {
    if (attempt === initialization) {
      rendererState.value = 'fallback'
      if (import.meta.env.DEV) {
        const summary = error instanceof Error ? error.message.replace(/\s+/g, ' ').slice(0, 180) : 'unknown renderer initialization failure'
        console.warn(`[PixelCafe] scene renderer fallback: ${summary}`)
      }
    }
  }
}

watch(rendererData, (value) => {
  if (renderer) renderer.update(value)
}, { deep: true })

onMounted(() => {
  mediaQuery = window.matchMedia?.('(prefers-reduced-motion: reduce)')
  updateMotionPreference()
  mediaQuery?.addEventListener?.('change', updateMotionPreference)
  void initializeRenderer()
})

onUnmounted(() => {
  initialization += 1
  mediaQuery?.removeEventListener?.('change', updateMotionPreference)
  renderer?.destroy()
  renderer = undefined
})
</script>

<style scoped>
.pixel-cafe-scene-stage { position: relative; min-height: 0; overflow: hidden; aspect-ratio: 16 / 9; background: #1f2837; }
.pixel-cafe-scene-visual { position: absolute; inset: 0; }
.pixel-cafe-scene-art { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; image-rendering: pixelated; pointer-events: none; }
.pixel-cafe-scene-canvas-host { position: absolute; inset: 0; z-index: 2; pointer-events: auto; }:global(.pixel-cafe-pixi-canvas) { display: block; width: 100% !important; height: 100% !important; image-rendering: pixelated; }
.pixel-cafe-scene-fallback { position: absolute; z-index: 1; inset: 0; pointer-events: none; }.pixel-cafe-fallback-workstation, .pixel-cafe-fallback-avatar { position: absolute; display: block; height: auto; image-rendering: pixelated; object-fit: contain; }.pixel-cafe-fallback-workstation { z-index: 1; width: 8.125%; transform: translate(-50%, -92%); }.pixel-cafe-fallback-avatar { z-index: 2; width: auto; height: 12%; transform: translate(-50%, -100%); }.pixel-cafe-fallback-avatar.walking { z-index: 3; }
.pixel-cafe-scene-state { position: absolute; z-index: 4; top: .65rem; left: .75rem; margin: 0; padding: .35rem .5rem; color: #fff6e5; background: rgba(37, 43, 57, .78); font-size: .72rem; }
@media (max-width: 900px) { .pixel-cafe-scene-stage { overflow: visible; }.pixel-cafe-scene-visual { position: relative; inset: auto; aspect-ratio: 16 / 9; overflow: hidden; } }
</style>
