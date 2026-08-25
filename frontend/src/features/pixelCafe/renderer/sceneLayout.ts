export const CAFE_SCENE_DESIGN_WIDTH = 960
export const CAFE_SCENE_DESIGN_HEIGHT = 540
export const CAFE_SCENE_ROOM_LIMIT = 12
export const CAFE_SCENE_DEFAULT_WORKSTATION_COUNT = 10
export const CAFE_SCENE_MIN_WORKSTATION_COUNT = 1
export const CAFE_SCENE_MAX_WORKSTATION_COUNT = 50
export const CAFE_SCENE_WALKING_AVATAR_COUNT = 6
export const CAFE_SCENE_WORKSTATION_COUNT = CAFE_SCENE_DEFAULT_WORKSTATION_COUNT

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

// Two staggered rows leave a clear aisle around the generated furniture. One visible
// worker maps to one workstation; extra active users take one of the safe walk routes.
export const CAFE_SCENE_WORKSTATIONS: CafeWorkstationSlot[] = [
  { id: 1, x: 340, y: 250 },
  { id: 2, x: 445, y: 250 },
  { id: 3, x: 550, y: 250 },
  { id: 4, x: 655, y: 250 },
  { id: 5, x: 760, y: 250 },
  { id: 6, x: 360, y: 362 },
  { id: 7, x: 465, y: 362 },
  { id: 8, x: 570, y: 362 },
  { id: 9, x: 675, y: 362 },
  { id: 10, x: 780, y: 362 },
]

const autoPlacementCandidates: readonly CafeScenePoint[] = Array.from({ length: 6 }, (_, row) =>
  Array.from({ length: 10 }, (_, column) => ({
    x: 280 + column * 60,
    y: 160 + row * 60,
  })),
).flat()

export const CAFE_SCENE_WALK_ROUTES: readonly (readonly CafeScenePoint[])[] = [
  [{ x: 300, y: 420 }, { x: 405, y: 420 }, { x: 515, y: 420 }, { x: 620, y: 420 }],
  [{ x: 280, y: 405 }, { x: 280, y: 330 }, { x: 280, y: 260 }, { x: 280, y: 200 }],
  [{ x: 620, y: 420 }, { x: 515, y: 420 }, { x: 405, y: 420 }, { x: 300, y: 420 }],
  [{ x: 290, y: 405 }, { x: 290, y: 340 }, { x: 290, y: 280 }, { x: 290, y: 215 }],
  [{ x: 320, y: 430 }, { x: 420, y: 430 }, { x: 520, y: 430 }, { x: 610, y: 430 }],
  [{ x: 275, y: 200 }, { x: 275, y: 270 }, { x: 275, y: 340 }, { x: 275, y: 400 }],
]

const WORKSTATION_MIN_X = 48
const WORKSTATION_MAX_X = 912
const WORKSTATION_MIN_Y = 72
const WORKSTATION_MAX_Y = 520

function normalizeWorkstationCount(value: number): number {
  if (!Number.isFinite(value)) return CAFE_SCENE_DEFAULT_WORKSTATION_COUNT
  return Math.min(CAFE_SCENE_MAX_WORKSTATION_COUNT, Math.max(CAFE_SCENE_MIN_WORKSTATION_COUNT, Math.trunc(value)))
}

function nextAutoPlacement(existing: readonly CafeWorkstationSlot[]): CafeScenePoint {
  let best = autoPlacementCandidates[0]
  let bestDistance = -1
  for (const candidate of autoPlacementCandidates) {
    const nearestDistance = existing.reduce((nearest, slot) => Math.min(
      nearest,
      (candidate.x - slot.x) ** 2 + (candidate.y - slot.y) ** 2,
    ), Number.POSITIVE_INFINITY)
    if (nearestDistance > bestDistance) {
      best = candidate
      bestDistance = nearestDistance
    }
  }
  return best
}

export function createCafeWorkstationLayout(count = CAFE_SCENE_DEFAULT_WORKSTATION_COUNT): CafeWorkstationSlot[] {
  const targetCount = normalizeWorkstationCount(count)
  const layout = CAFE_SCENE_WORKSTATIONS
    .slice(0, Math.min(targetCount, CAFE_SCENE_DEFAULT_WORKSTATION_COUNT))
    .map(slot => ({ ...slot }))
  while (layout.length < targetCount) {
    const position = nextAutoPlacement(layout)
    layout.push({ id: layout.length + 1, ...position })
  }
  return layout
}

export function resolveCafeWorkstationLayout(value?: readonly CafeWorkstationSlot[] | null): CafeWorkstationSlot[] {
  if (!Array.isArray(value)
    || value.length < CAFE_SCENE_MIN_WORKSTATION_COUNT
    || value.length > CAFE_SCENE_MAX_WORKSTATION_COUNT) {
    return createCafeWorkstationLayout()
  }
  const workstationCount = value.length
  const seen = new Set<number>()
  const normalized = value.map((slot) => {
    const id = Number(slot?.id)
    const x = Number(slot?.x)
    const y = Number(slot?.y)
    if (!Number.isInteger(id) || id < 1 || id > workstationCount || seen.has(id)
      || !Number.isFinite(x) || x < WORKSTATION_MIN_X || x > WORKSTATION_MAX_X
      || !Number.isFinite(y) || y < WORKSTATION_MIN_Y || y > WORKSTATION_MAX_Y) {
      return null
    }
    seen.add(id)
    return { id, x, y }
  })
  if (normalized.some(slot => slot === null)) return createCafeWorkstationLayout()
  const sorted = (normalized as CafeWorkstationSlot[]).sort((a, b) => a.id - b.id)
  if (sorted.some((slot, index) => slot.id !== index + 1)) return createCafeWorkstationLayout()
  return sorted
}

export function resizeCafeWorkstationLayout(value: readonly CafeWorkstationSlot[], count: number): CafeWorkstationSlot[] {
  const targetCount = normalizeWorkstationCount(count)
  const layout = resolveCafeWorkstationLayout(value).slice(0, targetCount).map(slot => ({ ...slot }))
  while (layout.length < targetCount) {
    const position = nextAutoPlacement(layout)
    layout.push({ id: layout.length + 1, ...position })
  }
  return layout
}

export function getCafeSceneCoverTransform(width: number, height: number): { scale: number; offsetX: number; offsetY: number } {
  const scale = Math.max(width / CAFE_SCENE_DESIGN_WIDTH, height / CAFE_SCENE_DESIGN_HEIGHT)
  const offsetX = (width - CAFE_SCENE_DESIGN_WIDTH * scale) / 2
  const offsetY = (height - CAFE_SCENE_DESIGN_HEIGHT * scale) / 2
  return {
    scale,
    offsetX: Math.abs(offsetX) < 1e-9 ? 0 : offsetX,
    offsetY: Math.abs(offsetY) < 1e-9 ? 0 : offsetY,
  }
}

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

export function getLobbySeat(seatIndex: number, workstations: readonly CafeWorkstationSlot[] = CAFE_SCENE_WORKSTATIONS): CafeScenePoint {
  const normalized = Math.abs(seatIndex)
  const resolvedWorkstations = resolveCafeWorkstationLayout(workstations)
  const workstation = resolvedWorkstations[normalized % resolvedWorkstations.length]
  return { x: workstation.x + (normalized % 2 === 0 ? -17 : 17), y: workstation.y + 4 }
}

export function getCafeWalkRoute(index: number): readonly CafeScenePoint[] {
  return CAFE_SCENE_WALK_ROUTES[Math.abs(index) % CAFE_SCENE_WALK_ROUTES.length]
}

export function getAvatarToneIndex(seed: string): number {
  let hash = 0
  for (const char of seed) hash = (hash * 31 + char.charCodeAt(0)) >>> 0
  return hash % 5
}
