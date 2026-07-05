/**
 * Payment System Type Definitions
 */

// ==================== Enums / Union Types ====================

export type OrderStatus =
  | 'PENDING'
  | 'PAID'
  | 'RECHARGING'
  | 'COMPLETED'
  | 'EXPIRED'
  | 'CANCELLED'
  | 'FAILED'
  | 'REFUND_REQUESTED'
  | 'REFUNDING'
  | 'REFUND_PENDING'
  | 'PARTIALLY_REFUNDED'
  | 'REFUNDED'
  | 'REFUND_FAILED'

export type PaymentType = 'alipay' | 'wxpay' | 'alipay_direct' | 'wxpay_direct' | 'stripe' | 'easypay' | 'airwallex'

export type OrderType = 'balance' | 'subscription'

export type InvoiceType = 'vat_general' | 'vat_special'

export type InvoiceStatus = 'pending' | 'approved' | 'rejected' | 'issued'

// ==================== Configuration ====================

export interface PaymentConfig {
  payment_enabled: boolean
  min_amount: number
  max_amount: number
  daily_limit: number
  max_pending_orders: number
  order_timeout_minutes: number
  balance_disabled: boolean
  balance_recharge_multiplier: number
  enabled_payment_types: PaymentType[]
  help_image_url: string
  help_text: string
  stripe_publishable_key: string
}

export interface MethodLimit {
  currency?: string
  daily_limit: number
  daily_used: number
  daily_remaining: number
  single_min: number
  single_max: number
  fee_rate: number
  available: boolean
}

/** Response from /payment/limits API */
export interface MethodLimitsResponse {
  methods: Record<string, MethodLimit>
  global_min: number  // widest min across all methods; 0 = no minimum
  global_max: number  // widest max across all methods; 0 = no maximum
}

/** Response from /payment/checkout-info API — single call for the payment page */
export interface CheckoutInfoResponse {
  methods: Record<string, MethodLimit>
  global_min: number
  global_max: number
  plans: SubscriptionPlan[]
  recharge_packages: RechargePackage[]
  monthly_recharge_bonus_claimed: boolean
  monthly_recharge_bonus_claimed_at?: string
  balance_disabled: boolean
  balance_recharge_multiplier: number
  recharge_fee_rate: number
  help_text: string
  help_image_url: string
  stripe_publishable_key: string
}

export interface RechargePackage {
  id: string
  label: string
  pay_amount: number
  credited_amount: number
  bonus_amount: number
  effective_credited_amount: number
  effective_bonus_amount: number
  sort_order: number
}

export type MembershipTierLevel = 'normal' | 'vip' | 'svip'

export interface MembershipTierConfig {
  level: MembershipTierLevel
  label: string
  threshold_amount: number
  rate_multiplier: number
  rpm_limit: number
  tpm_limit: number
  image_active_tasks: number
  subscription_group_id: number
}

export interface MembershipSettings {
  enabled: boolean
  validity_days: number
  tiers: MembershipTierConfig[]
}

export interface MembershipGrant {
  id: number
  user_id: number
  tier: MembershipTierLevel
  source: string
  period_key: string
  qualified_amount: number
  starts_at: string
  expires_at: string
  status: string
  subscription_id?: number
  subscription_group_id?: number
}

export interface MembershipStatus {
  enabled: boolean
  current_tier: MembershipTierLevel
  current_tier_label: string
  benefits: MembershipTierConfig
  expires_at?: string
  current_month_paid: number
  month_period_start: string
  month_period_end: string
  next_tier?: MembershipTierConfig
  amount_to_next: number
  tiers: MembershipTierConfig[]
  grant?: MembershipGrant
}

// ==================== Orders ====================

export interface PaymentOrder {
  id: number
  user_id: number
  user_email?: string
  user_name?: string
  user_notes?: string | null
  amount: number
  pay_amount: number
  currency?: string
  fee_rate: number
  payment_type: string
  out_trade_no: string
  status: OrderStatus
  order_type: OrderType
  created_at: string
  expires_at: string
  paid_at?: string
  completed_at?: string
  refund_amount: number
  refund_reason?: string
  refund_requested_at?: string
  refund_requested_by?: number
  refund_request_reason?: string
  plan_id?: number
  subscription_group_id?: number
  subscription_days?: number
  provider_instance_id?: string
}

// ==================== Invoices ====================

export interface InvoiceSummary {
  currency: string
  eligible_amount: number
  requested_amount: number
  available_amount: number
}

export interface InvoiceClaimSummary {
  claimable_count: number
}

export interface InvoiceRequest {
  id: number
  user_id: number
  user_email?: string
  user_name?: string
  amount: number
  currency: string
  invoice_type: InvoiceType
  title: string
  tax_number: string
  remark?: string
  status: InvoiceStatus
  admin_note?: string
  invoice_no?: string
  file_name?: string
  file_size?: number
  file_content_type?: string
  reviewed_by?: number
  reviewed_at?: string
  issued_at?: string
  downloaded_at?: string
  download_count: number
  created_at: string
  updated_at: string
  downloadable: boolean
  claimable: boolean
}

export interface CreateInvoiceRequest {
  amount: number
  invoice_type: InvoiceType
  title: string
  tax_number: string
  remark?: string
}

// ==================== Plans & Channels ====================

export interface SubscriptionPlan {
  id: number
  group_id: number
  group_platform?: string
  group_name?: string
  rate_multiplier?: number
  peak_rate_enabled?: boolean
  peak_start?: string
  peak_end?: string
  peak_rate_multiplier?: number
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
  supported_model_scopes?: string[]
  name: string
  description: string
  price: number
  original_price?: number
  validity_days: number
  validity_unit: string
  /** Stored as JSON string in backend; API layer should parse before use */
  features: string[]
  for_sale: boolean
  sort_order: number
}

export interface PaymentChannel {
  id: number
  group_id?: number
  name: string
  platform: string
  rate_multiplier: number
  description: string
  models: string[]
  features: string[]
  enabled: boolean
}

// ==================== Providers ====================

export interface ProviderInstance {
  id: number
  provider_key: string
  name: string
  config: Record<string, string>
  supported_types: string[]
  enabled: boolean
  payment_mode: string
  refund_enabled: boolean
  allow_user_refund: boolean
  limits: string
  sort_order: number
}

// ==================== Request / Response ====================

export interface CreateOrderRequest {
  amount: number
  payment_type: string
  order_type: string
  plan_id?: number
  recharge_package_id?: string
  return_url?: string
  payment_source?: string
  openid?: string
  wechat_resume_token?: string
  is_mobile?: boolean
}

export type CreateOrderResultType = 'order_created' | 'oauth_required' | 'jsapi_ready'

export interface WechatOAuthInfo {
  authorize_url?: string
  appid?: string
  openid?: string
  scope?: string
  state?: string
  redirect_url?: string
}

export interface WechatJSAPIPayload {
  appId?: string
  timeStamp?: string
  nonceStr?: string
  package?: string
  signType?: string
  paySign?: string
}

export interface CreateOrderResult {
  order_id: number
  amount: number
  pay_url?: string
  qr_code?: string
  client_secret?: string
  intent_id?: string
  currency?: string
  country_code?: string
  payment_env?: string
  pay_amount: number
  fee_rate: number
  expires_at: string
  result_type?: CreateOrderResultType
  payment_type?: string
  out_trade_no?: string
  payment_mode?: string
  resume_token?: string
  oauth?: WechatOAuthInfo
  jsapi?: WechatJSAPIPayload
  jsapi_payload?: WechatJSAPIPayload
}

export interface DashboardStats {
  today_amount: number
  total_amount: number
  today_count: number
  total_count: number
  avg_amount: number
  daily_series: { date: string; amount: number; count: number }[]
  payment_methods: { type: string; amount: number; count: number }[]
  top_users: { user_id: number; email: string; amount: number }[]
}
