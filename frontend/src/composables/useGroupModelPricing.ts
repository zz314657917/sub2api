import { ref } from 'vue'
import { userChannelsAPI } from '@/api'
import type { UserAvailableChannel, UserSupportedModel } from '@/api/channels'

/** Index supported models (with channel pricing) by the group ids that can reach them. */
export function indexModelsByGroup(channels: UserAvailableChannel[]): Map<number, UserSupportedModel[]> {
  const byGroup = new Map<number, Map<string, UserSupportedModel>>()
  for (const channel of channels) {
    for (const section of channel.platforms ?? []) {
      for (const group of section.groups ?? []) {
        const models = byGroup.get(group.id) ?? new Map<string, UserSupportedModel>()
        for (const model of section.supported_models ?? []) {
          const key = `${model.platform}:${model.name}`
          if (!models.has(key)) models.set(key, model)
        }
        byGroup.set(group.id, models)
      }
    }
  }
  const out = new Map<number, UserSupportedModel[]>()
  for (const [groupId, models] of byGroup) {
    out.set(groupId, [...models.values()].sort((a, b) => a.name.localeCompare(b.name)))
  }
  return out
}

/**
 * Lazily loads per-group model pricing from the user "available channels" API.
 * Loads at most once; failures and a disabled feature both yield no data.
 */
export function useGroupModelPricing() {
  const modelsByGroup = ref(new Map<number, UserSupportedModel[]>())
  const loading = ref(false)
  const loaded = ref(false)
  let pending: Promise<void> | null = null

  const ensureLoaded = () => {
    if (loaded.value || pending) return pending ?? Promise.resolve()
    loading.value = true
    pending = userChannelsAPI.getAvailable()
      .then((channels) => {
        modelsByGroup.value = indexModelsByGroup(channels ?? [])
      })
      .catch((error) => {
        console.error('Failed to load group model pricing:', error)
      })
      .finally(() => {
        loading.value = false
        loaded.value = true
        pending = null
      })
    return pending
  }

  const modelsFor = (groupId: number | null | undefined) =>
    groupId ? modelsByGroup.value.get(groupId) ?? [] : []

  return { ensureLoaded, modelsFor, loading, loaded }
}
