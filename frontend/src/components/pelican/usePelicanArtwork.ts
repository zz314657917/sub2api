import { inject, provide, ref, type InjectionKey } from 'vue'
import { getPelicanResult, type PelicanResult } from '@/api/pelicanTests'

type Cached = { result: PelicanResult; bytes: number; height?: number }
type Consumer = { resolve: (result: PelicanResult) => void; reject: (reason: unknown) => void; cleanup: () => void }
type Work = { id: number; generation: number; controller: AbortController; consumers: Set<Consumer>; started: boolean }
const canceled = () => new DOMException('Artwork request canceled', 'AbortError')
export function artworkStatus(reason: unknown): number | undefined {
  return (reason as { status?: number })?.status ?? (reason as { response?: { status?: number } })?.response?.status
}

// One instance belongs to one mounted page, never to a persisted/shared auth store.
export function createPelicanArtwork(onDenied: () => void = () => {}, fetchResult = getPelicanResult) {
  const revision = ref(0)
  const allowed = ref(true)
  const cache = new Map<number, Cached>()
  const pending = new Map<number, Work>()
  let generation = 0
  let bytes = 0
  let running = 0
  let paused = false
  let disposed = false
  const peek = (id: number) => {
    const item = cache.get(id)
    if (item) { cache.delete(id); cache.set(id, item) }
    return item
  }
  function retain(result: PelicanResult) {
    const size = JSON.stringify(result).length * 2
    if (!result.html || size > 6 * 1024 * 1024) return
    const old = cache.get(result.id)
    if (old) { bytes -= old.bytes; cache.delete(result.id) }
    cache.set(result.id, { result, bytes: size, height: old?.height })
    bytes += size
    while (cache.size > 60 || bytes > 6 * 1024 * 1024) {
      const id = cache.keys().next().value!
      bytes -= cache.get(id)!.bytes
      cache.delete(id)
    }
  }
  function finish(work: Work, reason?: unknown, result?: PelicanResult) {
    for (const consumer of work.consumers) {
      consumer.cleanup()
      if (reason) consumer.reject(reason)
      else consumer.resolve(result!)
    }
    work.consumers.clear()
  }
  function pump() {
    if (paused || disposed || !allowed.value) return
    for (const work of pending.values()) {
      if (running >= 3) break
      if (work.started || !work.consumers.size) continue
      work.started = true; running++
      void fetchResult(work.id, work.controller.signal).then(result => {
        if (work.generation !== generation || work.controller.signal.aborted || disposed) return
        if (!result.html || result.status !== 'success') throw new Error('作品当前不可访问')
        retain(result)
        finish(work, undefined, result)
      }).catch(reason => {
        if (work.generation !== generation || work.controller.signal.aborted || disposed) return
        if ([401, 403].includes(artworkStatus(reason) ?? 0)) { invalidate(); onDenied() }
        else finish(work, reason)
      }).finally(() => {
        // Aborted transports still occupy a slot until they actually settle.
        running--
        if (pending.get(work.id) === work) pending.delete(work.id)
        finish(work, canceled())
        pump()
      })
    }
  }
  function request(id: number, signal?: AbortSignal, fresh = false): Promise<PelicanResult> {
    if (disposed || !allowed.value || signal?.aborted) return Promise.reject(canceled())
    const cached = !fresh && peek(id)
    if (cached) return Promise.resolve(cached.result)
    let work = pending.get(id)
    if (!work || work.controller.signal.aborted) {
      work = { id, generation, controller: new AbortController(), consumers: new Set(), started: false }
      pending.set(id, work)
    }
    const target = work
    const promise = new Promise<PelicanResult>((resolve, reject) => {
      const abort = () => {
        target.consumers.delete(consumer); consumer.cleanup(); reject(canceled())
        if (!target.consumers.size) {
          target.controller.abort()
          if (pending.get(id) === target) pending.delete(id)
        }
      }
      const consumer: Consumer = { resolve, reject, cleanup: () => signal?.removeEventListener('abort', abort) }
      target.consumers.add(consumer)
      signal?.addEventListener('abort', abort, { once: true })
    })
    pump()
    return promise
  }
  function cancelPending() {
    generation++
    for (const work of pending.values()) { work.controller.abort(); finish(work, canceled()) }
    pending.clear()
    revision.value++
  }
  function invalidate() {
    allowed.value = false
    cache.clear(); bytes = 0
    cancelPending()
  }
  return {
    revision, allowed, request, peek,
    height: (id: number) => peek(id)?.height,
    setHeight(id: number, height: number) { const item = cache.get(id); if (item) item.height = Math.max(1, Math.min(580, height)) },
    cancelPending, invalidate,
    resume() { allowed.value = true; pump() },
    setPaused(value: boolean) { paused = value; if (value) cancelPending(); else pump() },
    dispose() { disposed = true; invalidate() },
    stats: () => ({ entries: cache.size, bytes, running, pending: pending.size })
  }
}
export type PelicanArtwork = ReturnType<typeof createPelicanArtwork>
const key: InjectionKey<PelicanArtwork> = Symbol('pelican-artwork')
export function providePelicanArtwork(value: PelicanArtwork) { provide(key, value) }
export function usePelicanArtwork(): PelicanArtwork {
  const value = inject(key)
  if (!value) throw new Error('Pelican artwork page context missing')
  return value
}
