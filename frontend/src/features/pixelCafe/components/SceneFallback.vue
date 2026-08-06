<template>
  <nav v-if="displayedRooms.length > 0" class="pixel-cafe-room-navigator" :aria-label="`${activeZoneLabel} 房间导航`" data-testid="pixel-cafe-room-navigator">
    <p class="pixel-cafe-room-navigator-label">房间导航</p>
    <div class="pixel-cafe-room-grid">
      <button
        v-for="room in displayedRooms"
        :key="room.id"
        type="button"
        class="pixel-cafe-room"
        :class="[`room-${roomTone(room)}`, { active: room.id === selectedRoomId }]"
        :aria-pressed="room.id === selectedRoomId"
        @click="$emit('select-room', room)"
      >
        <span class="pixel-cafe-room-sign">{{ room.code }}</span>
        <span class="pixel-cafe-room-name">{{ room.name }}</span>
        <span class="pixel-cafe-room-meta">{{ roomSeatLabel(room) }} · {{ roomProgressLabel(room) }}</span>
        <span class="pixel-cafe-room-lamp" aria-hidden="true"></span>
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { CafePublicRoom } from '@/types/pixelCafe'
import { CAFE_SCENE_ROOM_LIMIT } from '../renderer/sceneLayout'

const props = defineProps<{
  rooms: CafePublicRoom[]
  activeZoneLabel: string
  selectedRoomId?: number | null
}>()

defineEmits<{
  'select-room': [room: CafePublicRoom]
}>()

const displayedRooms = computed(() => props.rooms.slice(0, CAFE_SCENE_ROOM_LIMIT))

function roomSeatLabel(room: CafePublicRoom): string {
  if (!room.round) return `${room.plan.total_seats} 席 · 暂未开团`
  return `${room.round.remaining_seats}/${room.plan.total_seats} 空位`
}

function purchaseStateLabel(state: string): string {
  return ({
    available: '可购买',
    full: '已满',
    activating: '开通中',
    active: '已开通',
  } as Record<string, string>)[state] || '暂不可用'
}

function roomProgressLabel(room: CafePublicRoom): string {
  if (room.round?.status === 'open') return '等待拼团'
  if (room.round?.status === 'activating') return '开通中'
  if (room.round?.status === 'active') return '已开通'
  return purchaseStateLabel(room.purchase_state)
}

function roomTone(room: CafePublicRoom): string {
  if (room.theme_key.includes('blue') || room.zone_key === 'gemini') return 'blue'
  if (room.theme_key.includes('green') || room.zone_key === 'openai') return 'green'
  if (room.purchase_state === 'unavailable') return 'night'
  return 'wood'
}
</script>

<style scoped>
.pixel-cafe-room-navigator { position: absolute; z-index: 3; top: 4rem; right: 1rem; bottom: auto; left: auto; width: min(16rem, calc(100% - 2rem)); padding: 0; background: transparent; }
.pixel-cafe-room-navigator-label { margin: 0 0 .55rem; color: #f0c370; font: 700 .68rem/1 monospace; letter-spacing: .1em; text-shadow: 1px 1px 0 #050d15; text-transform: uppercase; }
.pixel-cafe-room-grid { display: grid; grid-template-columns: 1fr; gap: .45rem; }
.pixel-cafe-room { position: relative; min-width: 0; min-height: 58px; padding: .5rem .6rem; border: 1px solid rgba(193, 217, 233, .44); color: #eaf3fa; background: rgba(5, 16, 28, .86); box-shadow: 3px 3px 0 rgba(0, 0, 0, .28); text-align: left; cursor: pointer; backdrop-filter: blur(4px); }
.pixel-cafe-room:hover, .pixel-cafe-room:focus-visible, .pixel-cafe-room.active { border-color: #f2bd69; background: rgba(56, 39, 23, .92); box-shadow: 4px 4px 0 rgba(0, 0, 0, .3); outline: 1px solid #f2bd69; outline-offset: 2px; }.pixel-cafe-room.room-green { border-color: rgba(111, 198, 151, .64); }.pixel-cafe-room.room-blue { border-color: rgba(126, 176, 231, .64); }.pixel-cafe-room.room-night { border-color: rgba(201, 156, 237, .64); }
.pixel-cafe-room-sign { display: block; color: #f1c26f; font: 700 .68rem/1 monospace; }.pixel-cafe-room-name { display: block; min-width: 0; margin-top: .4rem; overflow: hidden; font-size: .78rem; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }.pixel-cafe-room-meta { display: block; margin-top: .25rem; overflow: hidden; color: #b7c9d7; font-size: .66rem; text-overflow: ellipsis; white-space: nowrap; }.pixel-cafe-room-lamp { position: absolute; top: .55rem; right: .55rem; width: .45rem; height: .45rem; background: #f3cf72; box-shadow: 0 0 0 2px rgba(243, 207, 114, .2); }
@media (max-width: 900px) { .pixel-cafe-room-navigator { position: relative; top: auto; right: auto; bottom: auto; left: auto; margin-top: .85rem; }.pixel-cafe-room-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); overflow: visible; padding-bottom: 0; }.pixel-cafe-room { min-width: 0; flex: initial; } }@media (max-width: 620px) { .pixel-cafe-room-navigator { top: auto; right: auto; bottom: auto; left: auto; margin-top: .75rem; }.pixel-cafe-room { min-height: 64px; }.pixel-cafe-room-grid { gap: .4rem; } }
</style>
