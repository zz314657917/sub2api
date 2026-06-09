import { apiClient } from './client'

export interface StudioBridgeLaunchResponse {
  launch_url: string
  expires_at: string
}

export const studioBridgeAPI = {
  async launch(): Promise<StudioBridgeLaunchResponse> {
    const response = await apiClient.post<StudioBridgeLaunchResponse>('/user/studio-bridge/launch', {
      app_id: 'luoye-ai',
    })
    return response.data
  },
}
