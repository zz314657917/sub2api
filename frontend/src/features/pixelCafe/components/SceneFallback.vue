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
.pixel-cafe-room-navigator { position: relative; z-index: 3; padding: 1rem; background: rgba(255, 253, 248, .94); border-top: 1px solid rgba(69, 54, 44, .24); }
.pixel-cafe-room-navigator-label { margin: 0 0 .7rem; color: #6d6258; font: 700 .68rem/1 monospace; text-transform: uppercase; }
.pixel-cafe-room-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .65rem; }
.pixel-cafe-room { position: relative; min-height: 96px; padding: .7rem .75rem; border: 2px solid #504237; color: #2b2927; background: #cf9c73; box-shadow: 3px 3px 0 rgba(79, 59, 45, .22); text-align: left; cursor: pointer; }
.pixel-cafe-room:hover, .pixel-cafe-room:focus-visible, .pixel-cafe-room.active { transform: translate(-1px, -1px); box-shadow: 5px 5px 0 rgba(79, 59, 45, .24); outline: 2px solid #8f624f; outline-offset: 2px; }
.pixel-cafe-room.room-green { background: #9ebc9a; }.pixel-cafe-room.room-blue { background: #a3b5cb; }.pixel-cafe-room.room-night { background: #61667c; color: #f8f2e6; }
.pixel-cafe-room-sign { display: block; color: rgba(43, 41, 39, .7); font: 700 .68rem/1 monospace; }.room-night .pixel-cafe-room-sign { color: #e5dfd2; }
.pixel-cafe-room-name { display: block; min-width: 0; margin-top: .55rem; overflow: hidden; font-size: .82rem; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }.pixel-cafe-room-meta { display: block; margin-top: .3rem; overflow: hidden; font-size: .68rem; opacity: .8; text-overflow: ellipsis; white-space: nowrap; }
.pixel-cafe-room-lamp { position: absolute; top: .6rem; right: .6rem; width: .45rem; height: .45rem; background: #f3cf72; box-shadow: 0 0 0 2px rgba(92, 74, 61, .18); }
@media (max-width: 620px) { .pixel-cafe-room-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }.pixel-cafe-room { min-height: 98px; } }
</style>
