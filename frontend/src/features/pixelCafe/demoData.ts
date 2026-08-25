import type { CafeLobbyActivity, CafePublicOverview, CafePublicRoom, CafePublicZone, CafeRoundStatus } from '@/types/pixelCafe'

function demoRoom(id: number, code: string, name: string, tier: 'plus' | 'pro', paid: number, buyers: number, status: CafeRoundStatus = 'open'): CafePublicRoom {
  const total = 10
  return { id, code, name, zone_key: 'featured', theme_key: id % 2 ? 'warm_wood' : 'green_terminal', scene_slot_key: `demo-room-${Math.abs(id)}`, featured: true,
    plan: { id: id - 1000, title: `ChatGPT ${tier === 'pro' ? 'Pro' : 'Plus'}`, description: '本地大厅预览的份额制包间。', price_per_share: tier === 'pro' ? 99 : 59, price_label: `${tier === 'pro' ? 99 : 59} CNY`, validity_days: 30, subscription_tier: tier, total_shares: total, max_buyers: 4, max_shares_per_user: total },
    round: { id: id - 2000, status, paid_shares: paid, reserved_shares: 0, remaining_shares: Math.max(0, total - paid), max_buyers: 4, joined_buyers: buyers, remaining_buyer_slots: Math.max(0, 4 - buyers), deadline_at: '2026-12-31T15:59:59Z', fulfillment_deadline_at: status === 'awaiting_account' ? '2026-12-31T15:59:59Z' : null },
    member_avatars: Array.from({ length: buyers }, (_, index) => ({ avatar_seed: `demo-member-${Math.abs(id)}-${index + 1}` })),
    purchase_state: status === 'awaiting_account' ? 'awaiting_account' : status === 'active' ? 'active' : status === 'refunding' ? 'refunding' : paid >= total ? 'buyers_full' : 'available',
  }
}
const demoRooms: CafePublicRoom[] = [
  demoRoom(-101, 'P-101', '晨光 Pro 包间', 'pro', 6, 3), demoRoom(-102, 'P-102', '夜航 Plus 包间', 'plus', 4, 2),
  demoRoom(-103, 'P-103', '代码 Pro 包间', 'pro', 10, 4, 'awaiting_account'), demoRoom(-104, 'P-104', '协作 Plus 包间', 'plus', 8, 4),
  demoRoom(-105, 'P-105', '创作 Pro 包间', 'pro', 10, 4, 'active'), demoRoom(-106, 'P-106', '周末 Plus 包间', 'plus', 1, 1),
  demoRoom(-107, 'P-107', '深夜 Pro 包间', 'pro', 7, 2), demoRoom(-108, 'P-108', '轻量 Plus 包间', 'plus', 5, 3),
  demoRoom(-109, 'P-109', '研究 Pro 包间', 'pro', 9, 4), demoRoom(-110, 'P-110', '新手 Plus 包间', 'plus', 0, 0),
]
const demoLobby: CafeLobbyActivity = { available: true, date: '2026-08-05', timezone: 'Asia/Shanghai', label: '本地演示用户', unique_users: 50, successful_requests: 1380, display_max: 50, avatars: Array.from({ length: 50 }, (_, seat_index) => ({ avatar_seed: `demo-active-${seat_index + 1}`, seat_index, activity: seat_index % 4 === 0 ? 'today' : 'recent' })) }
function createDemoZones(rooms: CafePublicRoom[]): CafePublicZone[] { return [{ key: 'featured', name: '大厅', room_count: rooms.length, open_share_count: rooms.reduce((sum, room) => sum + (room.round?.remaining_shares || 0), 0) }] }
export function createPixelCafeDemoOverview(): CafePublicOverview { const rooms = demoRooms.map(room => ({ ...room, plan: { ...room.plan }, round: room.round ? { ...room.round } : room.round, member_avatars: room.member_avatars.map(member => ({ ...member })) })); return { api_version: 'cafe.v1', server_time: '2026-08-05T12:00:00+08:00', zones: createDemoZones(rooms), rooms, lobby: { ...demoLobby, avatars: demoLobby.avatars.map(avatar => ({ ...avatar })) } } }
export function isLocalPixelCafeDemo(query: Record<string, unknown>): boolean { if (typeof window === 'undefined') return false; const host = window.location.hostname.toLowerCase(); return (host === '127.0.0.1' || host === 'localhost' || host === '::1') && (query.demo === '1' || query.demo === true) }
