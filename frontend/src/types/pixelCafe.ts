import type { GroupBuyPlan } from './groupBuy'
import type { CreateOrderResult } from './payment'

export type CafeRoomStatus = 'draft' | 'enabled' | 'maintenance' | 'disabled'
export type CafeRoundStatus = 'open' | 'activating' | 'active' | 'completed' | 'failed' | 'cancelled'

export interface CafeRoomPlan extends Pick<
  GroupBuyPlan,
  'id' | 'title' | 'target_group_id' | 'total_shares' | 'seat_count' | 'timeout_minutes' | 'validity_days'
> {
  fulfillment_mode: 'room_subscription' | 'aggregate_tier' | string
  group_platform: string
  group_access_mode: 'room_managed' | 'normal' | string
}

export interface CafeRound {
  id: number
  plan_id: number
  cafe_room_id?: number | null
  assigned_account_id?: number | null
  room_code_snapshot?: string | null
  room_name_snapshot?: string | null
  status: CafeRoundStatus | string
  total_shares: number
  total_seats: number
  deadline_at: string
  created_at: string
  updated_at: string
}

export interface CafeRoom {
  id: number
  code: string
  name: string
  plan_id: number
  account_id?: number | null
  zone_key: string
  theme_key: string
  scene_slot_key: string
  status: CafeRoomStatus
  featured: boolean
  sort_order: number
  metadata?: Record<string, unknown>
  plan?: CafeRoomPlan
  created_at: string
  updated_at: string
}

export interface CafeRoomInput {
  code: string
  name: string
  plan_id: number
  account_id: number
  zone_key: string
  theme_key: string
  scene_slot_key?: string
  status: CafeRoomStatus
  featured: boolean
  sort_order: number
  metadata?: Record<string, unknown>
}

export interface CafeRoomUpdateInput {
  code?: string
  name?: string
  plan_id?: number
  account_id?: number
  zone_key?: string
  theme_key?: string
  scene_slot_key?: string
  status?: CafeRoomStatus
  featured?: boolean
  sort_order?: number
  metadata?: Record<string, unknown>
}

export interface CafeRoomBulkInput {
  plan_id: number
  account_ids: number[]
  code_prefix: string
  start_number: number
  zone_key: string
  theme_key: string
  create_open_round: boolean
}

export interface CafeRoomBulkResult {
  created: Array<{ account_id: number; room?: CafeRoom; round?: CafeRound }>
  failed: Array<{ account_id: number; error_code: string; message: string }>
}

export type CafePublicPurchaseState = 'available' | 'full' | 'activating' | 'active' | 'unavailable'
export type CafePublicSeatState = 'empty' | 'locked' | 'paid' | 'active'

export interface CafePublicPlan {
  id: number
  title: string
  description: string
  price_per_seat: number
  price_label: string
  validity_days: number
  total_seats: number
}

export interface CafePublicRound {
  id: number
  status: CafeRoundStatus | string
  paid_seats: number
  reserved_seats: number
  remaining_seats: number
  deadline_at: string
  activated_at?: string | null
  expires_at?: string | null
}

export interface CafePublicSeatVisual {
  seat_no: number
  state: CafePublicSeatState | string
  avatar_seed?: string
  is_mine: boolean
}

export interface CafePublicRoom {
  id: number
  code: string
  name: string
  zone_key: string
  theme_key: string
  scene_slot_key: string
  featured: boolean
  plan: CafePublicPlan
  round?: CafePublicRound | null
  seat_visuals: CafePublicSeatVisual[]
  purchase_state: CafePublicPurchaseState | string
}

export interface CafePublicZone {
  key: string
  name: string
  room_count: number
  open_seat_count: number
}

export interface CafeLobbyAvatar {
  avatar_seed: string
  seat_index: number
  activity: 'recent' | 'today' | string
}

export interface CafeLobbyActivity {
  available: boolean
  date: string
  timezone: string
  label: '今日使用用户' | string
  unique_users: number
  successful_requests: number
  display_max: number
  avatars: CafeLobbyAvatar[]
}

export interface CafePublicOverview {
  api_version: 'cafe.v1' | string
  server_time: string
  zones: CafePublicZone[]
  rooms: CafePublicRoom[]
  lobby: CafeLobbyActivity
}

export interface CafePublicRoomDetail {
  api_version: 'cafe.v1' | string
  room: CafePublicRoom
  rules: {
    activation: string
    refund: string
    one_seat_per_user: boolean
  }
  server_time: string
}

export type CafeMyRoomFilter = 'active,waiting' | 'history'

export interface CafeMyRoomManagedKey {
  id: number
  name: string
  status: string
  quota: number
  quota_used: number
  rate_limit_5h: number
  rate_limit_1d: number
  rate_limit_7d: number
  protected: true
}

export interface CafeMyRoom {
  membership_id: number
  room: {
    id: number
    code: string
    name: string
    zone_key: string
    theme_key: string
  }
  plan: {
    id: number
    title: string
    validity_days: number
  }
  round: {
    id: number
    status: CafeRoundStatus | string
    paid_seats: number
    total_seats: number
  }
  seat: {
    id: number
    seat_no: number | null
    status: string
    activated_at: string | null
    expires_at: string | null
  }
  managed_api_key: CafeMyRoomManagedKey | null
}

export interface CreateCafeRoomOrderRequest {
  seat_no: number
  payment_type: string
  openid?: string
  return_url?: string
  payment_source?: string
  is_mobile?: boolean
  agreement_accepted: boolean
}

export interface CreateCafeRoomOrderResult extends CreateOrderResult {
  room_id: number
  round_id: number
  seat_no: number
}
