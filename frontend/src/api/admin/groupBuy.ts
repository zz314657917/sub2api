import { apiClient } from '../client'
import type { BasePaginationResponse } from '@/types'
import type { GroupBuyAdminSeat, GroupBuyFulfillmentMode, GroupBuyPlan, GroupBuyRefundBatchResult, GroupBuyRound, GroupBuyTier } from '@/types/groupBuy'

export interface GroupBuyPlanPayload {
  title: string
  description?: string
  product_key?: string
  total_shares: number
  seat_count: number
  price_per_share: number
  price_per_seat: number
  price_label?: string
  quota_per_share_label: string
  quota_label: string
  max_shares_per_user: number
  target_group_id: number
  fulfillment_mode: GroupBuyFulfillmentMode
  room_key_quota_usd: number
  room_key_rate_limit_5h: number
  room_key_rate_limit_1d: number
  room_key_rate_limit_7d: number
  auto_create_room_key: boolean
  tier_group_ids: Record<string, number>
  tier_rules: GroupBuyTier[]
  launch_mode: 'auto' | 'manual'
  validity_days: number
  timeout_minutes: number
  refund_mode: 'balance_credit' | 'provider_refund'
  agreement_text?: string
  status: 'active' | 'disabled'
  sort_order?: number
}

const adminGroupBuyAPI = {
  listPlans() {
    return apiClient.get<GroupBuyPlan[]>('/admin/group-buy/plans')
  },

  createPlan(data: GroupBuyPlanPayload) {
    return apiClient.post<GroupBuyPlan>('/admin/group-buy/plans', data)
  },

  updatePlan(id: number, data: GroupBuyPlanPayload) {
    return apiClient.put<GroupBuyPlan>(`/admin/group-buy/plans/${id}`, data)
  },

  deletePlan(id: number) {
    return apiClient.delete(`/admin/group-buy/plans/${id}`)
  },

  createRound(planId: number) {
    return apiClient.post<GroupBuyRound>(`/admin/group-buy/plans/${planId}/rounds`)
  },

  listRounds(params?: { page?: number; page_size?: number; status?: string }) {
    return apiClient.get<BasePaginationResponse<GroupBuyRound>>('/admin/group-buy/rounds', { params })
  },

  listRoundSeats(id: number) {
    return apiClient.get<GroupBuyAdminSeat[]>(`/admin/group-buy/rounds/${id}/seats`)
  },

  closeRound(id: number, reason: string) {
    return apiClient.post(`/admin/group-buy/rounds/${id}/close`, { reason })
  },

  retryActivation(id: number) {
    return apiClient.post(`/admin/group-buy/rounds/${id}/retry-activation`)
  },

  processRefunds(id: number) {
    return apiClient.post<GroupBuyRefundBatchResult>(`/admin/group-buy/rounds/${id}/process-refunds`)
  },
}

export default adminGroupBuyAPI
