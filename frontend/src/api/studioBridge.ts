import { apiClient } from './client'

export interface StudioBridgeLaunchResponse {
  launch_url: string
  expires_at: string
}

export interface StudioBridgeSessionProbeResponse {
  user_id: number
}

export const studioBridgeAPI = {
  async launch(): Promise<StudioBridgeLaunchResponse> {
    const response = await apiClient.post<StudioBridgeLaunchResponse>('/user/studio-bridge/launch', {
      app_id: 'luoye-ai',
    })
    return response.data
  },
  async sessionProbe(parentOrigin: string): Promise<StudioBridgeSessionProbeResponse> {
    const response = await apiClient.get<StudioBridgeSessionProbeResponse>('/user/studio-bridge/session-probe', {
      params: {
        app_id: 'luoye-ai',
        parent_origin: parentOrigin,
      },
    })
    return response.data
  },
}
