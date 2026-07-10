<template>
  <div
    ref="rootRef"
    class="public-reveal-backdrop"
    :class="[`is-${variant}`, { 'is-static': !revealEnabled }]"
    aria-hidden="true"
  >
    <div class="public-reveal-image"></div>
    <canvas ref="canvasRef" class="public-reveal-mask"></canvas>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

defineOptions({
  name: 'PublicRevealBackdrop'
})

const props = withDefaults(defineProps<{
  variant?: 'hero' | 'page'
  disabled?: boolean
}>(), {
  variant: 'page',
  disabled: false
})

const variant = computed(() => props.variant)
const rootRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const revealEnabled = ref(false)

let ctx: CanvasRenderingContext2D | null = null
let resizeObserver: ResizeObserver | null = null
let reducedMotionQuery: MediaQueryList | null = null
let hoverQuery: MediaQueryList | null = null
let fadeFrame = 0
let refillTimer = 0
let revealFrame = 0
let pendingPoint: { x: number; y: number } | null = null
let targetPoint: { x: number; y: number } | null = null
let easedPoint: { x: number; y: number } | null = null
let lastTrailPoint: { x: number; y: number } | null = null
let trailFrame = 0
let lastFrameTime = 0

type RevealRipple = {
  x: number
  y: number
  age: number
  life: number
  grow: number
  radius: number
}

const ripples: RevealRipple[] = []

function revealRadius(): number {
  return variant.value === 'hero' ? 218 : 112
}

function revealLife(): number {
  return variant.value === 'hero' ? 820 : 340
}

function revealGrowDuration(): number {
  return variant.value === 'hero' ? 220 : 140
}

function easeOutQuad(value: number): number {
  return 1 - (1 - value) * (1 - value)
}

function canUseDynamicReveal(): boolean {
  if (props.disabled || typeof window === 'undefined') return false
  if (typeof window.matchMedia !== 'function') return false

  reducedMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  hoverQuery = window.matchMedia('(hover: hover) and (pointer: fine)')
  return hoverQuery.matches && !reducedMotionQuery.matches
}

function fillMask(alpha = 1): void {
  const canvas = canvasRef.value
  if (!canvas || !ctx) return
  ctx.globalCompositeOperation = 'source-over'
  const baseAlpha = variant.value === 'hero' ? 0.88 : 1
  const resolvedAlpha = Math.min(alpha, baseAlpha)
  ctx.fillStyle = resolvedAlpha >= 1 ? '#faf9f5' : `rgba(250, 249, 245, ${resolvedAlpha})`
  ctx.clearRect(0, 0, canvas.width, canvas.height)
  ctx.fillRect(0, 0, canvas.width, canvas.height)
}

function resizeCanvas(): void {
  const root = rootRef.value
  const canvas = canvasRef.value
  if (!root || !canvas) return

  const rect = root.getBoundingClientRect()
  const ratio = Math.min(window.devicePixelRatio || 1, 2)
  canvas.width = Math.max(1, Math.round(rect.width * ratio))
  canvas.height = Math.max(1, Math.round(rect.height * ratio))
  canvas.style.width = `${rect.width}px`
  canvas.style.height = `${rect.height}px`

  ctx = canvas.getContext('2d')
  if (!ctx) return
  ctx.setTransform(ratio, 0, 0, ratio, 0, 0)
  ripples.length = 0
  targetPoint = null
  easedPoint = null
  lastTrailPoint = null
  fillMask()
}

function revealAt(clientX: number, clientY: number): void {
  const root = rootRef.value
  const canvas = canvasRef.value
  if (!root || !canvas || !ctx || !revealEnabled.value) return

  const rect = root.getBoundingClientRect()
  if (
    clientX < rect.left ||
    clientX > rect.right ||
    clientY < rect.top ||
    clientY > rect.bottom
  ) {
    return
  }

  if (fadeFrame) {
    window.cancelAnimationFrame(fadeFrame)
    fadeFrame = 0
  }

  if (refillTimer) {
    window.clearTimeout(refillTimer)
    refillTimer = 0
  }

  const x = clientX - rect.left
  const y = clientY - rect.top
  targetPoint = { x, y }
  if (!easedPoint) {
    easedPoint = { x, y }
    lastTrailPoint = { x, y }
    addRipple(x, y, 1.08)
  }
  startTrailLoop()

  refillTimer = window.setTimeout(softlyRefillMask, variant.value === 'hero' ? 320 : 220)
}

function addRipple(x: number, y: number, pressure = 1): void {
  const baseRadius = revealRadius()
  const resolvedPressure = Math.min(1.12, Math.max(0.26, pressure))
  ripples.push({
    x,
    y,
    age: 0,
    life: revealLife(),
    grow: revealGrowDuration(),
    radius: baseRadius * (0.42 + resolvedPressure * 0.58)
  })

  const maxRipples = variant.value === 'hero' ? 52 : 16
  if (ripples.length > maxRipples) {
    ripples.splice(0, ripples.length - maxRipples)
  }
}

function drawRevealSpot(x: number, y: number, radius: number, strength = 1): void {
  if (!ctx) return

  if (strength <= 0) return

  const gradient = ctx.createRadialGradient(x, y, 0, x, y, radius)
  gradient.addColorStop(0, `rgba(0, 0, 0, ${strength})`)
  gradient.addColorStop(0.26, `rgba(0, 0, 0, ${0.82 * strength})`)
  gradient.addColorStop(0.62, `rgba(0, 0, 0, ${0.22 * strength})`)
  gradient.addColorStop(1, 'rgba(0, 0, 0, 0)')

  ctx.globalCompositeOperation = 'destination-out'
  ctx.fillStyle = gradient
  ctx.beginPath()
  ctx.arc(x, y, radius, 0, Math.PI * 2)
  ctx.fill()
  ctx.globalCompositeOperation = 'source-over'
}

function renderRipples(delta: number): void {
  fillMask()

  for (let index = ripples.length - 1; index >= 0; index -= 1) {
    const ripple = ripples[index]
    ripple.age += delta
    const progress = Math.min(ripple.age / ripple.life, 1)
    const growProgress = Math.min(ripple.age / ripple.grow, 1)
    const growEase = easeOutQuad(growProgress)
    const remaining = 1 - progress
    const strength = Math.pow(remaining, variant.value === 'hero' ? 0.82 : 1.1) * (0.34 + growEase * 0.66)
    const shrinkRadius = 0.08 + Math.pow(remaining, variant.value === 'hero' ? 1.35 : 1.95) * 0.92
    const growRadius = 0.14 + growEase * 0.86
    const radius = ripple.radius * shrinkRadius * growRadius
    if (strength <= 0.018 || radius <= 5) {
      ripples.splice(index, 1)
      continue
    }
    drawRevealSpot(ripple.x, ripple.y, radius, strength)
  }
}

function startTrailLoop(): void {
  if (trailFrame) return
  lastFrameTime = performance.now()
  trailFrame = window.requestAnimationFrame(runTrailFrame)
}

function runTrailFrame(time: number): void {
  trailFrame = 0
  if (!revealEnabled.value || !ctx) return

  const delta = Math.min(48, Math.max(16, time - lastFrameTime || 16))
  lastFrameTime = time

  if (targetPoint) {
    if (!easedPoint) {
      easedPoint = { ...targetPoint }
    } else {
      const easing = variant.value === 'hero' ? 0.14 : 0.22
      easedPoint.x += (targetPoint.x - easedPoint.x) * easing
      easedPoint.y += (targetPoint.y - easedPoint.y) * easing
    }

    const dx = !lastTrailPoint ? 0 : easedPoint.x - lastTrailPoint.x
    const dy = !lastTrailPoint ? 0 : easedPoint.y - lastTrailPoint.y
    const distance = Math.hypot(dx, dy)
    const spacing = variant.value === 'hero' ? 6 : 5
    if (!lastTrailPoint || distance > spacing) {
      const steps = !lastTrailPoint ? 1 : Math.min(8, Math.max(1, Math.floor(distance / spacing)))
      const pressure = Math.min(1, Math.max(0.32, distance / (variant.value === 'hero' ? 84 : 54)))
      for (let step = 1; step <= steps; step += 1) {
        const progress = step / steps
        const rippleX = lastTrailPoint ? lastTrailPoint.x + dx * progress : easedPoint.x
        const rippleY = lastTrailPoint ? lastTrailPoint.y + dy * progress : easedPoint.y
        addRipple(rippleX, rippleY, pressure)
      }
      lastTrailPoint = { ...easedPoint }
    }
  }

  renderRipples(delta)

  if (ripples.length > 0 || targetPoint) {
    trailFrame = window.requestAnimationFrame(runTrailFrame)
    return
  }

  fillMask()
}

function softlyRefillMask(): void {
  if (!revealEnabled.value || !ctx) return

  if (fadeFrame) window.cancelAnimationFrame(fadeFrame)
  if (refillTimer) {
    window.clearTimeout(refillTimer)
    refillTimer = 0
  }
  targetPoint = null
  easedPoint = null
  lastTrailPoint = null
  startTrailLoop()
}

function handlePointerMove(event: PointerEvent): void {
  pendingPoint = { x: event.clientX, y: event.clientY }
  if (revealFrame) return
  revealFrame = window.requestAnimationFrame(() => {
    revealFrame = 0
    const point = pendingPoint
    pendingPoint = null
    if (!point) return
    revealAt(point.x, point.y)
  })
}

function handlePointerLeave(): void {
  pendingPoint = null
  if (revealFrame) {
    window.cancelAnimationFrame(revealFrame)
    revealFrame = 0
  }
  targetPoint = null
  softlyRefillMask()
}

function setupReveal(): void {
  revealEnabled.value = canUseDynamicReveal()
  resizeCanvas()

  if (!revealEnabled.value) return

  window.addEventListener('pointermove', handlePointerMove, { passive: true })
  document.addEventListener('pointerleave', handlePointerLeave)
  window.addEventListener('blur', handlePointerLeave)

  const root = rootRef.value
  if (root && 'ResizeObserver' in window) {
    resizeObserver = new ResizeObserver(resizeCanvas)
    resizeObserver.observe(root)
  } else {
    window.addEventListener('resize', resizeCanvas)
  }
}

onMounted(setupReveal)

onBeforeUnmount(() => {
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('resize', resizeCanvas)
  window.removeEventListener('blur', handlePointerLeave)
  document.removeEventListener('pointerleave', handlePointerLeave)
  resizeObserver?.disconnect()
  if (fadeFrame) window.cancelAnimationFrame(fadeFrame)
  if (revealFrame) window.cancelAnimationFrame(revealFrame)
  if (trailFrame) window.cancelAnimationFrame(trailFrame)
  if (refillTimer) window.clearTimeout(refillTimer)
  ripples.length = 0
  targetPoint = null
  easedPoint = null
  lastTrailPoint = null
  reducedMotionQuery = null
  hoverQuery = null
})
</script>

<style scoped>
.public-reveal-backdrop {
  position: absolute;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
  background: #faf9f5;
}

.public-reveal-image,
.public-reveal-mask {
  position: absolute;
  inset: 0;
  height: 100%;
  width: 100%;
}

.public-reveal-image {
  background-image: url('/public/mimo-relief-bg.webp');
  background-position: center top;
  background-repeat: no-repeat;
  background-size: min(1540px, 128vw) auto;
  opacity: 0.42;
  filter: saturate(0.86) sepia(0.12) contrast(1.04);
}

.public-reveal-backdrop.is-hero .public-reveal-image {
  inset: 0;
  height: 100%;
  background-position: center center;
  background-size: cover;
  opacity: 1;
  image-rendering: auto;
  filter: saturate(0.9) sepia(0.12) contrast(1.12);
}

.public-reveal-mask {
  display: block;
}

.public-reveal-backdrop.is-page .public-reveal-image {
  background-size: min(1480px, 140vw) auto;
  opacity: 0.16;
}

.public-reveal-backdrop.is-static .public-reveal-mask {
  display: none;
}

.public-reveal-backdrop.is-static .public-reveal-image {
  opacity: 0.11;
}

.public-reveal-backdrop.is-hero.is-static .public-reveal-image {
  opacity: 0.12;
}

@media (max-width: 768px), (hover: none), (pointer: coarse), (prefers-reduced-motion: reduce) {
  .public-reveal-mask {
    display: none;
  }

  .public-reveal-image {
    background-size: 68rem auto;
    opacity: 0.1;
  }

  .public-reveal-backdrop.is-hero .public-reveal-image {
    inset: 0;
    height: 100%;
    background-position: center center;
    background-size: cover;
    opacity: 0.12;
  }
}
</style>
