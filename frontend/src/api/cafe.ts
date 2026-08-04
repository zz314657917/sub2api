import { apiClient } from './client'
import type { PaginatedResponse } from '@/types'
import type {
  CafePublicOverview,
  CafeLobbyActivity,
  CafePublicRoom,
  CafePublicRoomDetail,
  CafeMyRoom,
  CreateCafeRoomOrderRequest,
  CreateCafeRoomOrderResult,
} from '@/types/pixelCafe'

export interface CafeRoomListParams {
  page?: number
  page_size?: number
  zone?: string
  featured?: boolean
}

export interface CafeMyRoomsListParams {
  page?: number
  page_size?: number
  status?: string
}

export const cafeAPI = {
  overview(params?: { room_limit?: number }) {
    return apiClient.get<CafePublicOverview>('/cafe/overview', { params })
  },

  lobbyActivity() {
    return apiClient.get<CafeLobbyActivity>('/cafe/lobby-activity')
  },

  listRooms(params?: CafeRoomListParams) {
    return apiClient.get<PaginatedResponse<CafePublicRoom>>('/cafe/rooms', { params })
  },

  listMyRooms(params?: CafeMyRoomsListParams) {
    return apiClient.get<PaginatedResponse<CafeMyRoom>>('/cafe/my-rooms', { params })
  },

  getRoom(id: number) {
    return apiClient.get<CafePublicRoomDetail>(`/cafe/rooms/${id}`)
  },

  createOrder(id: number, data: CreateCafeRoomOrderRequest, idempotencyKey: string) {
    return apiClient.post<CreateCafeRoomOrderResult>(`/cafe/rooms/${id}/orders`, data, {
      headers: { 'Idempotency-Key': idempotencyKey },
    })
  },
}

export default cafeAPI
