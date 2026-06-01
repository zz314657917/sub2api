import { apiClient } from './client'
import type {
  CreateSupportTicketMessageRequest,
  CreateSupportTicketRequest,
  PaginatedResponse,
  SupportTicket,
  SupportTicketDetail,
  SupportTicketListParams,
  TicketUnreadSummary,
} from '@/types'

function buildListParams(params: SupportTicketListParams = {}) {
  return {
    page: params.page,
    page_size: params.page_size,
    status: params.status || undefined,
    ticket_type: params.ticket_type || undefined,
    search: params.search || undefined,
    unread_only: params.unread_only || undefined,
  }
}

export async function listTickets(
  params: SupportTicketListParams = {}
): Promise<PaginatedResponse<SupportTicket>> {
  const { data } = await apiClient.get<PaginatedResponse<SupportTicket>>('/user/tickets', {
    params: buildListParams(params),
  })
  return data
}

export async function getUnreadSummary(): Promise<TicketUnreadSummary> {
  const { data } = await apiClient.get<TicketUnreadSummary>('/user/tickets/unread-summary')
  return data
}

export async function createTicket(input: CreateSupportTicketRequest): Promise<SupportTicket> {
  const { data } = await apiClient.post<SupportTicket>('/user/tickets', input)
  return data
}

export async function getTicket(id: number): Promise<SupportTicketDetail> {
  const { data } = await apiClient.get<SupportTicketDetail>(`/user/tickets/${id}`)
  return data
}

export async function createMessage(
  id: number,
  input: CreateSupportTicketMessageRequest
): Promise<SupportTicketDetail> {
  const { data } = await apiClient.post<SupportTicketDetail>(`/user/tickets/${id}/messages`, input)
  return data
}

export async function markRead(id: number): Promise<void> {
  await apiClient.post(`/user/tickets/${id}/read`)
}

export async function closeTicket(id: number): Promise<void> {
  await apiClient.post(`/user/tickets/${id}/close`)
}

export const ticketsAPI = {
  list: listTickets,
  unreadSummary: getUnreadSummary,
  create: createTicket,
  get: getTicket,
  createMessage,
  markRead,
  close: closeTicket,
}

export default ticketsAPI
