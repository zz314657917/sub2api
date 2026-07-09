import { apiClient } from './client'
import type { BasePaginationResponse } from '@/types'
import type {
  BindGroupBuyKeyRequest,
  CreateGroupBuyOrderRequest,
  CreateGroupBuyOrderResult,
  GroupBuyEvent,
  GroupBuyMySeatsResponse,
  GroupBuyPlan,
  GroupBuySeat,
  PaymentOrderLite,
} from '@/types/groupBuy'

export const groupBuyAPI = {
  listPlans() {
    return apiClient.get<GroupBuyPlan[]>('/group-buy/plans')
  },

  activity(limit = 20) {
    return apiClient.get<GroupBuyEvent[]>('/group-buy/activity', { params: { limit } })
  },

  createOrder(data: CreateGroupBuyOrderRequest) {
    return apiClient.post<CreateGroupBuyOrderResult>('/group-buy/orders', data)
  },

  mySeats() {
    return apiClient.get<GroupBuyMySeatsResponse>('/group-buy/my/seats')
  },

  myOrders(params?: { page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<PaymentOrderLite>>('/group-buy/my/orders', { params })
  },

  bindKey(seatId: number, data: BindGroupBuyKeyRequest) {
    return apiClient.post<GroupBuySeat>(`/group-buy/seats/${seatId}/bind-key`, data)
  },
}
