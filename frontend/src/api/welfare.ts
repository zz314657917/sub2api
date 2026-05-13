import { apiClient } from './client'
import type {
  WelfareDailyCheckin,
  WelfareDailyCheckinClaimResponse,
  WelfareMilestoneClaimResponse,
  WelfareOverview
} from '@/types'

export type {
  WelfareDailyCheckin,
  WelfareDailyCheckinClaimResponse,
  WelfareDailyCheckinMilestone,
  WelfareMilestoneClaimResponse,
  WelfareNewUserTrial,
  WelfareOverview
} from '@/types'

export async function getWelfareOverview(options?: { signal?: AbortSignal }): Promise<WelfareOverview> {
  const { data } = await apiClient.get<WelfareOverview>('/user/welfare/overview', {
    signal: options?.signal
  })
  return data
}

export async function getWelfareDailyCheckin(options?: { signal?: AbortSignal }): Promise<WelfareDailyCheckin> {
  const { data } = await apiClient.get<WelfareDailyCheckin>('/user/welfare/daily-checkin', {
    signal: options?.signal
  })
  return data
}

export async function claimWelfareDailyCheckin(): Promise<WelfareDailyCheckinClaimResponse> {
  const { data } = await apiClient.post<WelfareDailyCheckinClaimResponse>('/user/welfare/daily-checkin/claim')
  return data
}

export async function claimWelfareDailyCheckinMilestone(day: number): Promise<WelfareMilestoneClaimResponse> {
  const { data } = await apiClient.post<WelfareMilestoneClaimResponse>(`/user/welfare/daily-checkin/milestones/${day}/claim`)
  return data
}

export const welfareAPI = {
  getWelfareOverview,
  getWelfareDailyCheckin,
  claimWelfareDailyCheckin,
  claimWelfareDailyCheckinMilestone
}

export default welfareAPI
