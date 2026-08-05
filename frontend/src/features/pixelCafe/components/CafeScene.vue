<template>
  <div class="pixel-cafe-scene-stage" :data-renderer-state="rendererState" data-testid="pixel-cafe-scene-stage">
    <img class="pixel-cafe-scene-art" :src="cafeSceneAssets.lobbyBackground" alt="" aria-hidden="true" />
    <div ref="canvasHost" class="pixel-cafe-scene-canvas-host" aria-hidden="true"></div>
    <div v-if="displayedLobbyAvatars.length" class="pixel-cafe-lobby-avatars" aria-hidden="true">
      <span
        v-for="avatar in displayedLobbyAvatars"
        :key="avatar.key"
        class="pixel-cafe-lobby-avatar"
        :class="avatar.tone"
        :style="avatar.style"
        data-testid="pixel-cafe-lobby-avatar"
      >
        <span class="pixel-cafe-lobby-avatar-head"></span>
        <span class="pixel-cafe-lobby-avatar-body"></span>
      </span>
    </div>
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
import { CAFE_SCENE_DESIGN_HEIGHT, CAFE_SCENE_DESIGN_WIDTH, CAFE_SCENE_ROOM_LIMIT, getAvatarToneIndex, getLobbySeat } from '../renderer/sceneLayout'

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

const displayedLobbyAvatars = computed(() => props.lobbyAvatars.slice(0, CAFE_SCENE_ROOM_LIMIT).map((avatar, index) => {
  const seat = getLobbySeat(avatar.seat_index)
  const mobileColumn = index % 5
  const mobileRow = Math.floor(index / 5)
  return {
    key: `${avatar.avatar_seed}:${avatar.seat_index}`,
    tone: `tone-${getAvatarToneIndex(avatar.avatar_seed)}`,
    style: {
      '--avatar-x': `${(seat.x / CAFE_SCENE_DESIGN_WIDTH) * 100}%`,
      '--avatar-y': `${(seat.y / CAFE_SCENE_DESIGN_HEIGHT) * 100}%`,
      '--avatar-mobile-x': `${12 + mobileColumn * 18}%`,
      '--avatar-mobile-y': `${87 + mobileRow * 7}%`,
    },
  }
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
.pixel-cafe-lobby-avatars { position: absolute; z-index: 2; inset: 0; pointer-events: none; }.pixel-cafe-lobby-avatar { position: absolute; top: var(--avatar-y); left: var(--avatar-x); width: 22px; height: 29px; transform: translate(-50%, -50%); filter: drop-shadow(2px 2px 0 rgba(31, 40, 55, .58)); image-rendering: pixelated; }.pixel-cafe-lobby-avatar-head { position: absolute; top: 0; left: 6px; width: 10px; height: 10px; background: #f3d0b3; box-shadow: 2px 0 0 #e6b990, 0 2px 0 #e6b990; }.pixel-cafe-lobby-avatar-body { position: absolute; top: 10px; left: 3px; width: 16px; height: 13px; background: var(--avatar-color); box-shadow: -3px 5px 0 var(--avatar-color), 16px 5px 0 var(--avatar-color), 3px 13px 0 #394252, 10px 13px 0 #394252; }.pixel-cafe-lobby-avatar.tone-0 { --avatar-color: #b87565; }.pixel-cafe-lobby-avatar.tone-1 { --avatar-color: #6f9a83; }.pixel-cafe-lobby-avatar.tone-2 { --avatar-color: #7b91bb; }.pixel-cafe-lobby-avatar.tone-3 { --avatar-color: #cb9d59; }.pixel-cafe-lobby-avatar.tone-4 { --avatar-color: #9d7ab1; }
.pixel-cafe-scene-state { position: absolute; z-index: 4; top: .65rem; left: .75rem; margin: 0; padding: .35rem .5rem; color: #fff6e5; background: rgba(37, 43, 57, .78); font-size: .72rem; }
@media (max-width: 900px) { .pixel-cafe-scene-stage { min-height: 430px; }.pixel-cafe-lobby-avatar { top: var(--avatar-mobile-y); left: var(--avatar-mobile-x); } }
</style>
