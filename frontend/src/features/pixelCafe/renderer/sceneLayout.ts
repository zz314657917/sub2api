export const CAFE_SCENE_DESIGN_WIDTH = 960
export const CAFE_SCENE_DESIGN_HEIGHT = 400
export const CAFE_SCENE_ROOM_LIMIT = 12
export const CAFE_SCENE_WORKSTATION_COUNT = 8

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

// These are visual anchors for ambient activity, not a rendered workstation grid.
export const CAFE_SCENE_WORKSTATIONS: CafeWorkstationSlot[] = [
  { id: 1, x: 250, y: 145 },
  { id: 2, x: 330, y: 178 },
  { id: 3, x: 430, y: 145 },
  { id: 4, x: 535, y: 185 },
  { id: 5, x: 640, y: 145 },
  { id: 6, x: 730, y: 205 },
  { id: 7, x: 420, y: 300 },
  { id: 8, x: 590, y: 320 },
]

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
