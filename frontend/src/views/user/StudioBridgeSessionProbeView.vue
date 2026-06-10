<template>
  <main class="studio-bridge-session-probe" aria-hidden="true" />
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import { authAPI, studioBridgeAPI } from '@/api'

const MESSAGE_TYPE = 'sub2api:studio-bridge-session'
const PROBE_REQUEST_TYPE = 'sub2api:studio-bridge-session-probe'
const DEFAULT_APP_ID = 'luoye-ai'

type ProbePayload = {
  type: typeof MESSAGE_TYPE
  app_id: string
  authenticated: boolean
  user_id?: string
  error?: string
  checked_at: string
}

function queryString(name: string): string {
  return String(new URLSearchParams(window.location.search).get(name) || '').trim()
}

function normalizeOrigin(value: string): string {
  try {
    const url = new URL(value)
    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      return ''
    }
    return url.origin
  } catch {
    return ''
  }
}

function resolveAppID(): string {
  return queryString('app_id') || DEFAULT_APP_ID
}

function targetParentOrigin(): string {
  return normalizeOrigin(queryString('parent_origin'))
}

function postSession(payload: Omit<ProbePayload, 'type' | 'app_id' | 'checked_at'>): void {
  const parentOrigin = targetParentOrigin()
  if (!parentOrigin || window.parent === window) {
    return
  }
  window.parent.postMessage(
    {
      type: MESSAGE_TYPE,
      app_id: resolveAppID(),
      checked_at: new Date().toISOString(),
      ...payload,
    } satisfies ProbePayload,
    parentOrigin,
  )
}

function shouldAcceptParentMessage(event: MessageEvent): boolean {
  const parentOrigin = targetParentOrigin()
  if (!parentOrigin || event.origin !== parentOrigin) {
    return false
  }
  const data = event.data as { type?: unknown; app_id?: unknown } | null
  return Boolean(
    data
    && typeof data === 'object'
    && data.type === PROBE_REQUEST_TYPE
    && String(data.app_id || '').trim() === resolveAppID(),
  )
}

function classifyProbeError(error: unknown): string {
  const status = Number((error as { status?: unknown })?.status)
  if (status === 401 || status === 403) {
    return 'invalid_session'
  }
  return 'probe_failed'
}

async function probeCurrentSession(): Promise<void> {
  if (!authAPI.getAuthToken()) {
    postSession({ authenticated: false })
    return
  }

  try {
    const parentOrigin = targetParentOrigin()
    if (!parentOrigin) {
      postSession({ authenticated: false, error: 'invalid_parent_origin' })
      return
    }
    const response = await studioBridgeAPI.sessionProbe(parentOrigin)
    const userID = String(response.user_id || '').trim()
    if (!userID) {
      postSession({ authenticated: false, error: 'missing_user_id' })
      return
    }
    postSession({ authenticated: true, user_id: userID })
  } catch (error) {
    const errorCode = classifyProbeError(error)
    postSession({
      authenticated: false,
      ...(errorCode === 'probe_failed' ? { error: errorCode } : {}),
    })
  }
}

function handleParentMessage(event: MessageEvent): void {
  if (!shouldAcceptParentMessage(event)) {
    return
  }
  void probeCurrentSession()
}

onMounted(() => {
  window.addEventListener('message', handleParentMessage)
  void probeCurrentSession()
})

onBeforeUnmount(() => {
  window.removeEventListener('message', handleParentMessage)
})
</script>

<style scoped>
.studio-bridge-session-probe {
  display: none;
}
</style>
