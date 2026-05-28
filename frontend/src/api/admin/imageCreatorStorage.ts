import { apiClient } from '../client'

export type ImageCreatorStorageGovernanceAction =
  | 'expired_images'
  | 'orphan_files'
  | 'preview_cache'
  | 'thumb_cache'

export interface ImageCreatorStorageGovernanceItem {
  count: number
  byte_size: number
  unsupported?: boolean
  reason?: string
}

export interface ImageCreatorStorageGovernanceStats {
  storage_backend: string
  storage_dir?: string
  scanned_at: string
  images: ImageCreatorStorageGovernanceItem
  expired_images: ImageCreatorStorageGovernanceItem
  orphan_files: ImageCreatorStorageGovernanceItem
  preview_cache: ImageCreatorStorageGovernanceItem
  thumb_cache: ImageCreatorStorageGovernanceItem
}

export interface ImageCreatorStorageGovernanceCleanupResult {
  action: ImageCreatorStorageGovernanceAction
  deleted: number
  deleted_bytes: number
  unsupported?: boolean
  reason?: string
}

export async function getStorageGovernanceStats(): Promise<ImageCreatorStorageGovernanceStats> {
  const { data } = await apiClient.get<ImageCreatorStorageGovernanceStats>(
    '/admin/image-creator/storage-governance',
  )
  return data
}

export async function cleanupStorageGovernance(
  action: ImageCreatorStorageGovernanceAction,
): Promise<ImageCreatorStorageGovernanceCleanupResult> {
  const { data } = await apiClient.post<ImageCreatorStorageGovernanceCleanupResult>(
    '/admin/image-creator/storage-governance',
    { action },
  )
  return data
}

export const imageCreatorStorageAPI = {
  getStorageGovernanceStats,
  cleanupStorageGovernance,
}

export default imageCreatorStorageAPI
