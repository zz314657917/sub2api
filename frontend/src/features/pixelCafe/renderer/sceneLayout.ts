export const CAFE_SCENE_DESIGN_WIDTH = 960
export const CAFE_SCENE_DESIGN_HEIGHT = 400
export const CAFE_SCENE_ROOM_LIMIT = 12
export const CAFE_SCENE_WORKSTATION_COUNT = 50
export const CAFE_SCENE_WORKSTATION_COLUMNS = 10

export interface CafeScenePoint {
  x: number
  y: number
}

export interface CafeRoomHotspot extends CafeScenePoint {
  width: number
  height: number
}

const defaultRoomHotspots: CafeRoomHotspot[] = Array.from({ length: CAFE_SCENE_ROOM_LIMIT }, (_, index) => ({
  x: 752 + (index % 2) * 102,
  y: 58 + Math.floor(index / 2) * 50,
  width: 94,
  height: 42,
}))

export interface CafeWorkstationSlot extends CafeScenePoint {
  id: number
}

export const CAFE_SCENE_WORKSTATIONS: CafeWorkstationSlot[] = Array.from(
  { length: CAFE_SCENE_WORKSTATION_COUNT },
  (_, index) => {
    const row = Math.floor(index / CAFE_SCENE_WORKSTATION_COLUMNS)
    const column = index % CAFE_SCENE_WORKSTATION_COLUMNS
    return {
      id: index + 1,
      // Keep the full 10-column grid inside the unobstructed floor; the room rail
      // occupies the right side of the visual scene on desktop.
      x: 110 + column * 66 + (row % 2) * 8,
      y: 88 + row * 42,
    }
  },
)

const explicitRoomSlots: Record<string, CafeRoomHotspot> = {
  'featured-room-01': defaultRoomHotspots[0],
  'claude-room-08': defaultRoomHotspots[1],
  'openai-room-01': defaultRoomHotspots[2],
  'gemini-room-01': defaultRoomHotspots[3],
}

function stableIndex(value: string, length: number): number {
  let hash = 0
  for (const char of value) hash = (hash * 31 + char.charCodeAt(0)) >>> 0
  return hash % length
}

export function getRoomHotspot(sceneSlotKey: string, index: number): CafeRoomHotspot {
  const known = explicitRoomSlots[sceneSlotKey]
  if (known) return known
  return defaultRoomHotspots[stableIndex(`${sceneSlotKey}:${index}`, defaultRoomHotspots.length)]
}

export function getLobbySeat(seatIndex: number): CafeScenePoint {
  return CAFE_SCENE_WORKSTATIONS[Math.abs(seatIndex) % CAFE_SCENE_WORKSTATIONS.length]
}

export function getAvatarToneIndex(seed: string): number {
  let hash = 0
  for (const char of seed) hash = (hash * 31 + char.charCodeAt(0)) >>> 0
  return hash % 5
}
