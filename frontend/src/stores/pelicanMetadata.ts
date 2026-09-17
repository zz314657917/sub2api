import { ref } from 'vue'
import { defineStore } from 'pinia'
import { getPelicanTestMetadata } from '@/api/pelicanTests'

const defaultName = '鹈鹕测试'

export const usePelicanMetadataStore = defineStore('pelicanMetadata', () => {
  const displayName = ref(defaultName)
  const enabled = ref(false)
  const loaded = ref(false)
  const loading = ref(false)
  let requestVersion = 0
  let pending: Promise<void> | null = null

  async function load(force = false): Promise<void> {
    if (loaded.value && !force) return
    if (pending && !force) return pending
    if (force) { requestVersion++; pending = null }
    const version = ++requestVersion
    loading.value = true
    const task = getPelicanTestMetadata().then(metadata => {
      if (version === requestVersion && metadata?.display_name?.trim()) {
        displayName.value = metadata.display_name
        enabled.value = metadata.enabled === true
        loaded.value = true
      }
    }).catch(reason => {
      if (version === requestVersion) { enabled.value = false; loaded.value = false; loading.value = false }
      throw reason
    }).finally(() => {
      if (pending === task) {
        loading.value = false
        pending = null
      }
    })
    pending = task
    return task
  }

  function applySavedSettings(name: string, nextEnabled = true): void {
    requestVersion++
    pending = null
    loading.value = false
    if (name.trim()) displayName.value = name
    enabled.value = nextEnabled
    loaded.value = true
  }

  // Retained for callers compiled against the previous metadata-only API.
  function applySavedName(name: string): void { applySavedSettings(name, enabled.value) }

  function reset(): void {
    requestVersion++; pending = null; loading.value = false
    displayName.value = defaultName; enabled.value = false; loaded.value = false
  }

  return { displayName, enabled, loaded, loading, load, applySavedSettings, applySavedName, reset }
})
