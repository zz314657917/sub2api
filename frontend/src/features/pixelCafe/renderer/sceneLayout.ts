export const CAFE_SCENE_DESIGN_WIDTH = 960
export const CAFE_SCENE_DESIGN_HEIGHT = 540
export const CAFE_SCENE_ROOM_LIMIT = 12

export interface CafeScenePoint {
  x: number
  y: number
}

export interface CafeRoomHotspot extends CafeScenePoint {
  width: number
  height: number
}

const defaultRoomHotspots: CafeRoomHotspot[] = [
  { x: 88, y: 122, width: 148, height: 78 },
  { x: 302, y: 114, width: 148, height: 78 },
  { x: 516, y: 122, width: 148, height: 78 },
  { x: 730, y: 114, width: 148, height: 78 },
  { x: 142, y: 268, width: 148, height: 78 },
  { x: 356, y: 260, width: 148, height: 78 },
  { x: 570, y: 268, width: 148, height: 78 },
  { x: 784, y: 260, width: 112, height: 78 },
  { x: 248, y: 402, width: 148, height: 76 },
  { x: 462, y: 394, width: 148, height: 76 },
  { x: 676, y: 402, width: 148, height: 76 },
  { x: 70, y: 402, width: 112, height: 76 },
]

const lobbySeats: CafeScenePoint[] = [
  { x: 102, y: 342 },
  { x: 164, y: 366 },
  { x: 228, y: 340 },
  { x: 292, y: 368 },
  { x: 356, y: 344 },
  { x: 420, y: 370 },
  { x: 484, y: 346 },
  { x: 548, y: 370 },
  { x: 612, y: 344 },
  { x: 676, y: 368 },
  { x: 740, y: 342 },
  { x: 804, y: 366 },
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
  return lobbySeats[Math.abs(seatIndex) % lobbySeats.length]
}
