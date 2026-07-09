import type { CreateOrderResult, PaymentOrder } from './payment'

export type GroupBuyPlanStatus = 'active' | 'disabled'
export type GroupBuyRoundStatus = 'open' | 'activating' | 'active' | 'failed' | 'cancelled'
export type GroupBuySeatStatus = 'locked' | 'released' | 'paid' | 'active' | 'refund_pending' | 'refund_processing' | 'refunded' | 'cancelled'
export type GroupBuyRefundMode = 'balance_credit' | 'provider_refund'
export type GroupBuyLaunchMode = 'auto' | 'manual'
export type GroupBuyEntitlementStatus = 'active' | 'inactive'

export interface GroupBuyGroupView {
  id: number
  name: string
  platform: string
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
}

export interface GroupBuyRound {
  id: number
  plan_id: number
  status: GroupBuyRoundStatus
  total_shares: number
  paid_shares: number
  reserved_shares: number
  available_shares: number
  total_seats: number
  paid_seats: number
  reserved_seats: number
  available_seats: number
  deadline_at: string
  started_at?: string
  closed_at?: string
  close_reason?: string
  created_at: string
  updated_at: string
}

export interface GroupBuyTier {
  share_count?: number
  min_shares: number
  max_shares: number
  target_group_id: number
  label?: string
  target_group?: GroupBuyGroupView
}

export interface GroupBuyPlan {
  id: number
  title: string
  description: string
  product_key?: string
  total_shares: number
  seat_count: number
  price_per_share: number
  price_per_seat: number
  price_label: string
  quota_per_share_label: string
  quota_label: string
  max_shares_per_user: number
  target_group_id: number
  target_group?: GroupBuyGroupView
  tier_group_ids: Record<string, number>
  tier_groups: GroupBuyTier[]
  tier_rules: GroupBuyTier[]
  validity_days: number
  timeout_minutes: number
  launch_mode: GroupBuyLaunchMode
  refund_mode: GroupBuyRefundMode
  agreement_text: string
  status: GroupBuyPlanStatus
  sort_order: number
  current_round?: GroupBuyRound
  created_at: string
  updated_at: string
}

export interface GroupBuySeat {
  id: number
  round_id: number
  plan_id: number
  user_id: number
  order_id?: number
  status: GroupBuySeatStatus
  share_count: number
  subscription_id?: number
  bound_api_key_id?: number
  locked_until?: string
  paid_at?: string
  activated_at?: string
  expires_at?: string
  bound_at?: string
  refund_processed_at?: string
  refund_note?: string
  plan?: GroupBuyPlan
  round?: GroupBuyRound
  order?: PaymentOrderLite
  created_at: string
  updated_at: string
}

export interface GroupBuyEntitlement {
  id: number
  user_id: number
  product_key: string
  status: GroupBuyEntitlementStatus
  active_share_count: number
  target_group_id?: number
  target_group?: GroupBuyGroupView
  subscription_id?: number
  bound_api_key_id?: number
  entitlement_label?: string
  last_activated_at?: string
  expires_at?: string
  refreshed_at?: string
  deactivated_at?: string
  created_at: string
  updated_at: string
}

export interface GroupBuyMySeatsResponse {
  entitlement?: GroupBuyEntitlement
  seats: GroupBuySeat[]
}

export interface PaymentOrderLite extends Pick<PaymentOrder,
  'id' | 'amount' | 'pay_amount' | 'currency' | 'payment_type' | 'out_trade_no' | 'status' | 'order_type' | 'created_at' | 'expires_at' | 'paid_at' | 'completed_at'
> {}

export interface GroupBuyEvent {
  id: number
  plan_id?: number
  round_id?: number
  seat_id?: number
  user_id?: number
  event_type: string
  message: string
  metadata?: Record<string, unknown>
  created_at: string
}

export interface CreateGroupBuyOrderRequest {
  plan_id: number
  share_count: number
  payment_type: string
  openid?: string
  return_url?: string
  payment_source?: string
  is_mobile?: boolean
}

export interface CreateGroupBuyOrderResult extends CreateOrderResult {
  seat?: GroupBuySeat
  round?: GroupBuyRound
}

export interface BindGroupBuyKeyRequest {
  api_key_id: number
}
