<template>
  <div class="pixel-cafe-scene-stage" :data-renderer-state="rendererState" data-testid="pixel-cafe-scene-stage">
    <img class="pixel-cafe-scene-art" :src="cafeSceneAssets.lobbyBackground" alt="" aria-hidden="true" />
    <div ref="canvasHost" class="pixel-cafe-scene-canvas-host" aria-hidden="true"></div>
    <p v-if="rendererState === 'loading'" class="pixel-cafe-scene-state">场景加载中...</p>
    <p v-else-if="rendererState === 'fallback'" class="pixel-cafe-scene-state">场景暂不可用，房间导航仍可正常使用。</p>
    <SceneFallback
      :rooms="rooms"
      :active-zone-label="activeZoneLabel"
      :selected-room-id="selectedRoomId"
      @select-room="$emit('select-room', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import type { CafeLobbyAvatar, CafePublicRoom } from '@/types/pixelCafe'
import SceneFallback from './SceneFallback.vue'
import { cafeSceneAssets } from '../renderer/assetManifest'
import { createCafeRenderer, type CafeSceneRenderer } from '../renderer/createCafeRenderer'

const props = defineProps<{
  rooms: CafePublicRoom[]
  lobbyAvatars: CafeLobbyAvatar[]
  activeZoneLabel: string
  selectedRoomId?: number | null
}>()

const emit = defineEmits<{
  'select-room': [room: CafePublicRoom]
}>()

const canvasHost = ref<HTMLElement>()
const rendererState = ref<'loading' | 'ready' | 'fallback'>('loading')
const reducedMotion = ref(false)
let renderer: CafeSceneRenderer | undefined
let initialization = 0
let mediaQuery: MediaQueryList | undefined

const rendererData = computed(() => ({
  rooms: props.rooms,
  lobbyAvatars: props.lobbyAvatars,
  reducedMotion: reducedMotion.value,
  onRoomSelect: (room: CafePublicRoom) => emit('select-room', room),
}))

function updateMotionPreference(): void {
  reducedMotion.value = mediaQuery?.matches ?? false
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
    const nextRenderer = await createCafeRenderer(host, rendererData.value)
    if (attempt !== initialization) {
      nextRenderer.destroy()
      return
    }
    renderer = nextRenderer
    rendererState.value = 'ready'
  } catch {
    if (attempt === initialization) rendererState.value = 'fallback'
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
.pixel-cafe-scene-stage { position: relative; min-height: 500px; overflow: hidden; background: #1f2837; }
.pixel-cafe-scene-art { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; image-rendering: pixelated; pointer-events: none; }
.pixel-cafe-scene-canvas-host { position: absolute; inset: 0; z-index: 2; pointer-events: auto; }.pixel-cafe-scene-canvas-host :deep(canvas) { display: block; width: 100%; height: 100%; image-rendering: pixelated; }
.pixel-cafe-scene-state { position: absolute; z-index: 4; top: .65rem; left: .75rem; margin: 0; padding: .35rem .5rem; color: #fff6e5; background: rgba(37, 43, 57, .78); font-size: .72rem; }
@media (max-width: 900px) { .pixel-cafe-scene-stage { min-height: 430px; } }
</style>
