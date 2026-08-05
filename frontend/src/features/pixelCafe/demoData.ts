import type { CafeLobbyActivity, CafePublicOverview, CafePublicRoom, CafePublicZone } from '@/types/pixelCafe'

const demoRooms: CafePublicRoom[] = [
  {
    id: -101,
    code: 'C-018',
    name: 'Claude 深夜包间',
    zone_key: 'claude',
    theme_key: 'warm_wood',
    scene_slot_key: 'featured-room-01',
    featured: true,
    plan: { id: -11, title: 'Claude Max', description: '编程与长上下文的深夜包间。', price_per_seat: 99, price_label: '99 CNY', validity_days: 30, total_seats: 5 },
    round: { id: -1001, status: 'open', paid_seats: 3, reserved_seats: 1, remaining_seats: 1, deadline_at: '2026-12-31T15:59:59Z' },
    seat_visuals: [
      { seat_no: 1, state: 'paid', avatar_seed: 'ember-7f3a', is_mine: false },
      { seat_no: 2, state: 'paid', avatar_seed: 'aster-20c8', is_mine: false },
      { seat_no: 3, state: 'locked', avatar_seed: 'moss-9a52', is_mine: false },
      { seat_no: 4, state: 'empty', is_mine: false },
      { seat_no: 5, state: 'paid', avatar_seed: 'nova-42d1', is_mine: false },
    ],
    purchase_state: 'available',
  },
  {
    id: -102,
    code: 'O-022',
    name: 'OpenAI 编程包间',
    zone_key: 'openai',
    theme_key: 'green_terminal',
    scene_slot_key: 'openai-room-01',
    featured: true,
    plan: { id: -12, title: 'GPT Coding', description: '适合日常开发的协作座位。', price_per_seat: 79, price_label: '79 CNY', validity_days: 30, total_seats: 6 },
    round: { id: -1002, status: 'open', paid_seats: 2, reserved_seats: 0, remaining_seats: 4, deadline_at: '2026-12-31T15:59:59Z' },
    seat_visuals: [
      { seat_no: 1, state: 'paid', avatar_seed: 'leaf-14bc', is_mine: false },
      { seat_no: 2, state: 'paid', avatar_seed: 'opal-871d', is_mine: false },
      { seat_no: 3, state: 'empty', is_mine: false },
      { seat_no: 4, state: 'empty', is_mine: false },
      { seat_no: 5, state: 'empty', is_mine: false },
      { seat_no: 6, state: 'empty', is_mine: false },
    ],
    purchase_state: 'available',
  },
  {
    id: -103,
    code: 'G-007',
    name: 'Gemini 灵感包间',
    zone_key: 'gemini',
    theme_key: 'blue_neon',
    scene_slot_key: 'gemini-room-01',
    featured: true,
    plan: { id: -13, title: 'Gemini Pro', description: '多模态与长文档探索座位。', price_per_seat: 69, price_label: '69 CNY', validity_days: 30, total_seats: 4 },
    round: { id: -1003, status: 'open', paid_seats: 4, reserved_seats: 0, remaining_seats: 0, deadline_at: '2026-12-31T15:59:59Z' },
    seat_visuals: [
      { seat_no: 1, state: 'paid', avatar_seed: 'orbit-33af', is_mine: false },
      { seat_no: 2, state: 'paid', avatar_seed: 'cider-b244', is_mine: false },
      { seat_no: 3, state: 'paid', avatar_seed: 'lyra-9db0', is_mine: false },
      { seat_no: 4, state: 'paid', avatar_seed: 'coral-4e18', is_mine: false },
    ],
    purchase_state: 'full',
  },
  {
    id: -104,
    code: 'C-042',
    name: 'Claude 推理包间',
    zone_key: 'claude',
    theme_key: 'violet_night',
    scene_slot_key: 'claude-room-08',
    featured: true,
    plan: { id: -14, title: 'Claude Reasoning', description: '正在凑齐最后两席。', price_per_seat: 119, price_label: '119 CNY', validity_days: 30, total_seats: 8 },
    round: { id: -1004, status: 'open', paid_seats: 5, reserved_seats: 1, remaining_seats: 2, deadline_at: '2026-12-31T15:59:59Z' },
    seat_visuals: [
      { seat_no: 1, state: 'paid', avatar_seed: 'mist-320a', is_mine: false },
      { seat_no: 2, state: 'paid', avatar_seed: 'sprout-c6e1', is_mine: false },
      { seat_no: 3, state: 'paid', avatar_seed: 'pixel-88fe', is_mine: false },
      { seat_no: 4, state: 'paid', avatar_seed: 'wave-120c', is_mine: false },
      { seat_no: 5, state: 'paid', avatar_seed: 'dusk-62a5', is_mine: false },
      { seat_no: 6, state: 'locked', avatar_seed: 'mango-71e9', is_mine: false },
      { seat_no: 7, state: 'empty', is_mine: false },
      { seat_no: 8, state: 'empty', is_mine: false },
    ],
    purchase_state: 'available',
  },
  {
    id: -105,
    code: 'N-001',
    name: '通宵体验包间',
    zone_key: 'featured',
    theme_key: 'blue_neon',
    scene_slot_key: 'night-arcade-01',
    featured: true,
    plan: { id: -15, title: 'Night Shift', description: '满员后等待统一开通。', price_per_seat: 49, price_label: '49 CNY', validity_days: 7, total_seats: 4 },
    round: { id: -1005, status: 'activating', paid_seats: 4, reserved_seats: 0, remaining_seats: 0, deadline_at: '2026-12-31T15:59:59Z' },
    seat_visuals: [
      { seat_no: 1, state: 'paid', avatar_seed: 'solar-4bc0', is_mine: false },
      { seat_no: 2, state: 'paid', avatar_seed: 'petal-18ef', is_mine: false },
      { seat_no: 3, state: 'paid', avatar_seed: 'cedar-7b01', is_mine: false },
      { seat_no: 4, state: 'paid', avatar_seed: 'cloud-a89d', is_mine: false },
    ],
    purchase_state: 'activating',
  },
]

const demoLobby: CafeLobbyActivity = {
  available: true,
  date: '2026-08-05',
  timezone: 'Asia/Shanghai',
  label: '本地演示用户',
  unique_users: 10,
  successful_requests: 128,
  display_max: 12,
  avatars: [
    'ember-7f3a', 'aster-20c8', 'moss-9a52', 'nova-42d1', 'leaf-14bc',
    'opal-871d', 'orbit-33af', 'cider-b244', 'lyra-9db0', 'coral-4e18',
  ].map((avatar_seed, seat_index) => ({ avatar_seed, seat_index, activity: seat_index % 3 === 0 ? 'today' : 'recent' })),
}

function createDemoZones(rooms: CafePublicRoom[]): CafePublicZone[] {
  const definitions = [
    { key: 'featured', name: '精选大厅' },
    { key: 'claude', name: 'Claude 区' },
    { key: 'openai', name: 'OpenAI 区' },
    { key: 'gemini', name: 'Gemini 区' },
  ]

  return definitions.map((zone) => {
    const zoneRooms = zone.key === 'featured' ? rooms : rooms.filter(room => room.zone_key === zone.key)
    return {
      ...zone,
      room_count: zoneRooms.length,
      open_seat_count: zoneRooms.reduce((total, room) => total + room.seat_visuals.filter(seat => seat.state === 'empty').length, 0),
    }
  })
}

export function createPixelCafeDemoOverview(): CafePublicOverview {
  const rooms = demoRooms.map(room => ({
    ...room,
    plan: { ...room.plan },
    round: room.round ? { ...room.round } : room.round,
    seat_visuals: room.seat_visuals.map(seat => ({ ...seat })),
  }))

  return {
    api_version: 'cafe.v1',
    server_time: '2026-08-05T12:00:00+08:00',
    zones: createDemoZones(rooms),
    rooms,
    lobby: {
      ...demoLobby,
      avatars: demoLobby.avatars.map(avatar => ({ ...avatar })),
    },
  }
}

export function isLocalPixelCafeDemo(query: Record<string, unknown>): boolean {
  if (typeof window === 'undefined') return false
  const host = window.location.hostname.toLowerCase()
  const localHost = host === '127.0.0.1' || host === 'localhost' || host === '::1'
  return localHost && (query.demo === '1' || query.demo === true)
}
