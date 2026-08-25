import type { GroupBuyPlan } from './groupBuy'
import type { CreateOrderResult } from './payment'

export type CafeRoomStatus = 'draft' | 'enabled' | 'maintenance' | 'disabled'
export type CafeRoundStatus = 'open' | 'awaiting_account' | 'activating' | 'active' | 'completed' | 'refunding' | 'refunded' | 'failed' | 'cancelled'
export type CafeRoomUpdateInput = Partial<CafeRoomInput>
export type CafePublicPurchaseState = 'available' | 'buyers_full' | 'awaiting_account' | 'activating' | 'active' | 'refunding' | 'refunded' | 'unavailable'
export type CafeMyRoomFilter = 'active,waiting' | 'history'

export interface CafeRoomPlan extends Pick<GroupBuyPlan, 'id' | 'title' | 'target_group_id' | 'total_shares' | 'timeout_minutes' | 'validity_days'> {
  fulfillment_mode: 'room_subscription' | 'aggregate_tier' | string
  group_platform: string
  group_access_mode: 'room_managed' | 'normal' | string
  subscription_tier?: 'plus' | 'pro'
  max_buyers?: number
  max_shares_per_user?: number
  fulfillment_timeout_minutes?: number
}
export interface CafeRound {
  id: number; plan_id: number; status: CafeRoundStatus | string; total_shares: number
  paid_shares?: number; reserved_shares?: number; remaining_shares?: number
  max_buyers?: number; joined_buyers?: number; remaining_buyer_slots?: number
  deadline_at: string; created_at: string; updated_at: string
}
export interface CafeRoom { id: number; code: string; name: string; plan_id: number; account_id?: number | null; zone_key: string; theme_key: string; scene_slot_key: string; status: CafeRoomStatus; featured: boolean; sort_order: number; plan?: CafeRoomPlan; created_at: string; updated_at: string }
export interface CafeRoomInput { code: string; name: string; plan_id: number; zone_key: string; theme_key: string; scene_slot_key?: string; status: CafeRoomStatus; featured: boolean; sort_order: number }
export interface CafeRoomBulkInput { plan_id: number; quantity: number; code_prefix: string; start_number: number; zone_key: string; theme_key: string; create_open_round: boolean }
export interface CafeRoomBulkResult { created: Array<{ room?: CafeRoom; round?: CafeRound }>; failed: Array<{ index?: number; error_code: string; message: string }> }
export interface CafePublicPlan { id: number; title: string; description: string; price_per_share: number; price_label: string; validity_days: number; subscription_tier: 'plus' | 'pro'; total_shares: number; max_buyers: number; max_shares_per_user: number }
export interface CafePublicRound { id: number; status: CafeRoundStatus | string; paid_shares: number; reserved_shares: number; remaining_shares: number; max_buyers: number; joined_buyers: number; remaining_buyer_slots: number; deadline_at: string; fulfillment_deadline_at?: string | null }
export interface CafePublicMemberAvatar { avatar_seed: string }
export interface CafePublicRoom { id: number; code: string; name: string; zone_key: string; theme_key: string; scene_slot_key: string; featured: boolean; plan: CafePublicPlan; round?: CafePublicRound | null; member_avatars: CafePublicMemberAvatar[]; purchase_state: CafePublicPurchaseState | string; my_paid_shares?: number }
export interface CafePublicZone { key: string; name: string; room_count: number; open_share_count: number }
export interface CafeLobbyAvatar { avatar_seed: string; seat_index: number; activity: 'recent' | 'today' | string }
export interface CafeWorkstationPosition { id: number; x: number; y: number }
export interface CafeLobbyActivity { available: boolean; date: string; timezone: string; label: string; unique_users: number; successful_requests: number; display_max: number; avatars: CafeLobbyAvatar[] }
export interface CafePublicOverview { api_version: string; server_time: string; zones: CafePublicZone[]; rooms: CafePublicRoom[]; lobby: CafeLobbyActivity }
export interface CafePublicRoomDetail { api_version: string; room: CafePublicRoom; rules: { activation: string; refund: string; one_key_per_member: boolean }; server_time: string }
export interface CafeMyRoomManagedKey { id: number; name: string; status: string; quota: number; quota_used: number; rate_limit_5h: number; rate_limit_1d: number; rate_limit_7d: number; usage_5h: number; usage_7d: number; protected: true }
export interface CafeMyRoom { membership_id: number; status?: string; paid_shares: number; room: { id: number; code: string; name: string; zone_key: string; theme_key: string }; account?: { name: string; platform: string; email_masked?: string } | null; plan: { id: number; title: string; subscription_tier?: 'plus' | 'pro'; validity_days: number }; round: { id: number; status: CafeRoundStatus | string; paid_shares: number; total_shares: number }; managed_api_key: CafeMyRoomManagedKey | null }
export interface CreateCafeRoomOrderRequest { share_count: number; payment_type: string; openid?: string; return_url?: string; payment_source?: string; is_mobile?: boolean; agreement_accepted: boolean }
export interface CreateCafeRoomOrderResult extends CreateOrderResult { room_id: number; round_id: number; share_count: number; membership_id?: number }
