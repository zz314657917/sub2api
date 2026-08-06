<template>
  <div class="pixel-cafe-scene-stage" :data-renderer-state="rendererState" data-testid="pixel-cafe-scene-stage">
    <div class="pixel-cafe-scene-visual">
      <img class="pixel-cafe-scene-art" :src="cafeSceneAssets.lobbyBackground" alt="" aria-hidden="true" />
      <div class="pixel-cafe-workstations" aria-hidden="true">
        <img
          v-for="station in workstationStyles"
          :key="station.id"
          class="pixel-cafe-workstation"
          :src="cafeSceneAssets.workstation"
          :style="station.style"
          alt=""
          data-testid="pixel-cafe-workstation"
        />
      </div>
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
import { CAFE_SCENE_DESIGN_HEIGHT, CAFE_SCENE_DESIGN_WIDTH, CAFE_SCENE_ROOM_LIMIT, CAFE_SCENE_WORKSTATIONS, getAvatarToneIndex, getLobbySeat } from '../renderer/sceneLayout'

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

const workstationStyles = CAFE_SCENE_WORKSTATIONS.map(station => ({
  id: station.id,
  style: {
    '--workstation-x': `${(station.x / CAFE_SCENE_DESIGN_WIDTH) * 100}%`,
    '--workstation-y': `${(station.y / CAFE_SCENE_DESIGN_HEIGHT) * 100}%`,
  },
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
.pixel-cafe-scene-stage { position: relative; min-height: clamp(520px, 47vw, 700px); overflow: hidden; background: #1f2837; }
.pixel-cafe-scene-visual { position: absolute; inset: 0; }
.pixel-cafe-scene-art { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; image-rendering: pixelated; pointer-events: none; }
.pixel-cafe-workstations { position: absolute; z-index: 1; inset: 0; pointer-events: none; }.pixel-cafe-workstation { position: absolute; top: var(--workstation-y); left: var(--workstation-x); width: 5.5%; height: auto; transform: translate(-50%, -50%); filter: drop-shadow(2px 3px 0 rgba(9, 13, 18, .48)); image-rendering: pixelated; user-select: none; }
.pixel-cafe-scene-canvas-host { position: absolute; inset: 0; z-index: 2; pointer-events: auto; }.pixel-cafe-scene-canvas-host :deep(canvas) { display: block; width: 100%; height: 100%; image-rendering: pixelated; }
.pixel-cafe-lobby-avatars { position: absolute; z-index: 2; inset: 0; pointer-events: none; }.pixel-cafe-lobby-avatar { position: absolute; top: var(--avatar-y); left: var(--avatar-x); width: 34px; height: 42px; transform: translate(-50%, -50%); filter: drop-shadow(2px 2px 0 rgba(4, 12, 21, .84)); image-rendering: pixelated; animation: pixel-cafe-avatar-bob 1.8s steps(2, end) infinite; }.pixel-cafe-lobby-avatar:nth-child(2n) { animation-delay: -.6s; }.pixel-cafe-lobby-avatar:nth-child(3n) { animation-delay: -1.1s; }.pixel-cafe-lobby-avatar-head { position: absolute; top: 0; left: 10px; width: 13px; height: 12px; background: #f8d6b9; box-shadow: 2px 0 0 #dfa777, 0 2px 0 #dfa777; }.pixel-cafe-lobby-avatar-body { position: absolute; top: 12px; left: 5px; width: 24px; height: 17px; background: var(--avatar-color); box-shadow: -3px 5px 0 var(--avatar-color), 24px 5px 0 var(--avatar-color), 4px 16px 0 #394252, 15px 16px 0 #394252; }.pixel-cafe-lobby-avatar-body::after { position: absolute; right: -7px; bottom: 4px; width: 12px; height: 3px; background: #f8d6b9; box-shadow: 3px 3px 0 #172336; content: ''; animation: pixel-cafe-typing .6s steps(2, end) infinite; }.pixel-cafe-lobby-avatar.tone-0 { --avatar-color: #e2846f; }.pixel-cafe-lobby-avatar.tone-1 { --avatar-color: #74b78d; }.pixel-cafe-lobby-avatar.tone-2 { --avatar-color: #7e9de0; }.pixel-cafe-lobby-avatar.tone-3 { --avatar-color: #e3ae54; }.pixel-cafe-lobby-avatar.tone-4 { --avatar-color: #b88bd0; }@keyframes pixel-cafe-avatar-bob { 50% { transform: translate(-50%, calc(-50% - 2px)); } }@keyframes pixel-cafe-typing { 50% { transform: translateX(3px); } }
.pixel-cafe-scene-state { position: absolute; z-index: 4; top: .65rem; left: .75rem; margin: 0; padding: .35rem .5rem; color: #fff6e5; background: rgba(37, 43, 57, .78); font-size: .72rem; }
@media (prefers-reduced-motion: reduce) { .pixel-cafe-lobby-avatar, .pixel-cafe-lobby-avatar-body::after { animation: none; } }@media (max-width: 900px) { .pixel-cafe-scene-stage { min-height: 0; overflow: visible; }.pixel-cafe-scene-visual { position: relative; inset: auto; aspect-ratio: 12 / 5; overflow: hidden; }.pixel-cafe-lobby-avatar { display: none; } }
</style>
