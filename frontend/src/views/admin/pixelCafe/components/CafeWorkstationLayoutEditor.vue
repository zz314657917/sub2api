<template>
  <div class="cafe-layout-editor" data-testid="cafe-layout-editor">
    <div class="cafe-layout-toolbar">
      <p>{{ t('admin.pixelCafe.layout.hint') }}</p>
      <label class="cafe-layout-snap">
        <input v-model="snapEnabled" type="checkbox" />
        {{ t('admin.pixelCafe.layout.snap') }}
      </label>
    </div>

    <div class="cafe-layout-count-bar">
      <label>
        <span>{{ t('admin.pixelCafe.layout.count') }}</span>
        <span class="cafe-layout-count-stepper">
          <button
            type="button"
            :aria-label="t('admin.pixelCafe.layout.decreaseCount')"
            :disabled="resolvedLayout.length <= CAFE_SCENE_MIN_WORKSTATION_COUNT"
            data-testid="cafe-layout-count-decrease"
            @click="adjustWorkstationCount(-1)"
          >−</button>
          <input
            :value="resolvedLayout.length"
            type="number"
            :min="CAFE_SCENE_MIN_WORKSTATION_COUNT"
            :max="CAFE_SCENE_MAX_WORKSTATION_COUNT"
            step="1"
            data-testid="cafe-layout-count-input"
            @change="setWorkstationCountFromInput"
          />
          <button
            type="button"
            :aria-label="t('admin.pixelCafe.layout.increaseCount')"
            :disabled="resolvedLayout.length >= CAFE_SCENE_MAX_WORKSTATION_COUNT"
            data-testid="cafe-layout-count-increase"
            @click="adjustWorkstationCount(1)"
          >+</button>
        </span>
      </label>
      <span>{{ t('admin.pixelCafe.layout.countRange', { min: CAFE_SCENE_MIN_WORKSTATION_COUNT, max: CAFE_SCENE_MAX_WORKSTATION_COUNT }) }}</span>
    </div>

    <p class="cafe-layout-mobile-note">{{ t('admin.pixelCafe.layout.desktopOnly') }}</p>
    <div
      ref="stage"
      class="cafe-layout-stage"
      data-testid="cafe-layout-stage"
      @pointermove="moveWorkstation"
      @pointerup="endDrag"
      @pointercancel="endDrag"
    >
      <img :src="cafeSceneAssets.lobbyBackground" alt="" draggable="false" aria-hidden="true" />
      <button
        v-for="(slot, index) in resolvedLayout"
        :key="slot.id"
        type="button"
        :class="['cafe-layout-workstation', { selected: selectedID === slot.id }]"
        :style="workstationStyle(slot)"
        :aria-label="t('admin.pixelCafe.layout.workstation', { id: slot.id })"
        :data-workstation-id="slot.id"
        data-testid="cafe-layout-workstation"
        @pointerdown="startDrag($event, slot)"
        @keydown="nudgeWorkstation($event, slot)"
        @click="selectedID = slot.id"
      >
        <img :src="cafeSceneAssets.workstations[index % cafeSceneAssets.workstations.length].url" alt="" draggable="false" />
        <span>{{ slot.id }}</span>
      </button>
    </div>

    <div v-if="selectedWorkstation" class="cafe-layout-coordinate-panel">
      <strong>{{ t('admin.pixelCafe.layout.selected', { id: selectedWorkstation.id }) }}</strong>
      <label>
        X
        <input
          :value="selectedWorkstation.x"
          type="number"
          min="48"
          max="912"
          step="1"
          @change="setCoordinate('x', $event)"
        />
      </label>
      <label>
        Y
        <input
          :value="selectedWorkstation.y"
          type="number"
          min="72"
          max="520"
          step="1"
          @change="setCoordinate('y', $event)"
        />
      </label>
      <span>{{ t('admin.pixelCafe.layout.keyboardHint') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { CafeWorkstationPosition } from '@/types/pixelCafe'
import { cafeSceneAssets } from '@/features/pixelCafe/renderer/assetManifest'
import {
  CAFE_SCENE_DESIGN_HEIGHT,
  CAFE_SCENE_DESIGN_WIDTH,
  CAFE_SCENE_MAX_WORKSTATION_COUNT,
  CAFE_SCENE_MIN_WORKSTATION_COUNT,
  resolveCafeWorkstationLayout,
  resizeCafeWorkstationLayout,
} from '@/features/pixelCafe/renderer/sceneLayout'

const props = defineProps<{
  modelValue: CafeWorkstationPosition[]
}>()

const emit = defineEmits<{
  'update:modelValue': [layout: CafeWorkstationPosition[]]
}>()

const { t } = useI18n()
const stage = ref<HTMLElement>()
const selectedID = ref(1)
const snapEnabled = ref(true)
const drag = ref<{ id: number; pointerId: number; offsetX: number; offsetY: number } | null>(null)
const resolvedLayout = computed(() => resolveCafeWorkstationLayout(props.modelValue))
const selectedWorkstation = computed(() => resolvedLayout.value.find(slot => slot.id === selectedID.value))

function setWorkstationCount(count: number): void {
  const next = resizeCafeWorkstationLayout(resolvedLayout.value, count)
  selectedID.value = Math.min(selectedID.value, next.length)
  emit('update:modelValue', next)
}

function setWorkstationCountFromInput(event: Event): void {
  setWorkstationCount(Number((event.target as HTMLInputElement).value))
}

function adjustWorkstationCount(delta: number): void {
  setWorkstationCount(resolvedLayout.value.length + delta)
}

function workstationStyle(slot: CafeWorkstationPosition): Record<string, string | number> {
  return {
    left: `${slot.x / CAFE_SCENE_DESIGN_WIDTH * 100}%`,
    top: `${slot.y / CAFE_SCENE_DESIGN_HEIGHT * 100}%`,
    zIndex: Math.round(slot.y),
  }
}

function pointerToDesign(event: PointerEvent): { x: number; y: number } | null {
  const rect = stage.value?.getBoundingClientRect()
  if (!rect || rect.width <= 0 || rect.height <= 0) return null
  return {
    x: (event.clientX - rect.left) / rect.width * CAFE_SCENE_DESIGN_WIDTH,
    y: (event.clientY - rect.top) / rect.height * CAFE_SCENE_DESIGN_HEIGHT,
  }
}

function startDrag(event: PointerEvent, slot: CafeWorkstationPosition): void {
  const point = pointerToDesign(event)
  if (!point) return
  selectedID.value = slot.id
  drag.value = {
    id: slot.id,
    pointerId: event.pointerId,
    offsetX: point.x - slot.x,
    offsetY: point.y - slot.y,
  }
  ;(event.currentTarget as HTMLElement).setPointerCapture?.(event.pointerId)
  event.preventDefault()
}

function moveWorkstation(event: PointerEvent): void {
  if (!drag.value || drag.value.pointerId !== event.pointerId) return
  const point = pointerToDesign(event)
  if (!point) return
  updateWorkstation(drag.value.id, point.x - drag.value.offsetX, point.y - drag.value.offsetY)
}

function endDrag(event: PointerEvent): void {
  if (!drag.value || drag.value.pointerId !== event.pointerId) return
  drag.value = null
}

function clamp(value: number, min: number, max: number, snap = true): number {
  const bounded = Math.min(max, Math.max(min, value))
  const step = snap && snapEnabled.value ? 4 : 1
  return Math.round(bounded / step) * step
}

function updateWorkstation(id: number, x: number, y: number, snapX = true, snapY = true): void {
  const next = resolvedLayout.value.map(slot => slot.id === id
    ? { ...slot, x: clamp(x, 48, 912, snapX), y: clamp(y, 72, 520, snapY) }
    : { ...slot })
  emit('update:modelValue', next)
}

function nudgeWorkstation(event: KeyboardEvent, slot: CafeWorkstationPosition): void {
  const deltas: Record<string, [number, number]> = {
    ArrowLeft: [-1, 0],
    ArrowRight: [1, 0],
    ArrowUp: [0, -1],
    ArrowDown: [0, 1],
  }
  const delta = deltas[event.key]
  if (!delta) return
  event.preventDefault()
  selectedID.value = slot.id
  const step = event.shiftKey ? 10 : (snapEnabled.value ? 4 : 1)
  updateWorkstation(slot.id, slot.x + delta[0] * step, slot.y + delta[1] * step, delta[0] !== 0, delta[1] !== 0)
}

function setCoordinate(axis: 'x' | 'y', event: Event): void {
  const slot = selectedWorkstation.value
  if (!slot) return
  const value = Number((event.target as HTMLInputElement).value)
  if (!Number.isFinite(value)) return
  updateWorkstation(slot.id, axis === 'x' ? value : slot.x, axis === 'y' ? value : slot.y, axis === 'x', axis === 'y')
}
</script>

<style scoped>
.cafe-layout-editor { display: grid; gap: .8rem; }
.cafe-layout-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 1rem; color: #64748b; font-size: .82rem; }
.cafe-layout-toolbar p { margin: 0; }
.cafe-layout-snap { display: inline-flex; flex: 0 0 auto; align-items: center; gap: .4rem; }
.cafe-layout-count-bar { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: .65rem 1rem; padding: .55rem .7rem; border: 1px solid #d9e0e6; color: #64748b; background: #f8fafc; font-size: .78rem; }
.cafe-layout-count-bar > label { display: inline-flex; align-items: center; gap: .65rem; color: #26384a; font-weight: 700; }
.cafe-layout-count-stepper { display: inline-grid; grid-template-columns: 2rem 4.5rem 2rem; overflow: hidden; border: 1px solid #cbd5e1; background: #fff; }
.cafe-layout-count-stepper button, .cafe-layout-count-stepper input { min-height: 2rem; border: 0; color: #26384a; background: #fff; text-align: center; }
.cafe-layout-count-stepper button { font-size: 1rem; font-weight: 800; cursor: pointer; }
.cafe-layout-count-stepper button:disabled { color: #aab5c0; cursor: not-allowed; }
.cafe-layout-count-stepper input { width: 100%; border-right: 1px solid #cbd5e1; border-left: 1px solid #cbd5e1; font-weight: 700; }
.cafe-layout-mobile-note { display: none; margin: 0; padding: .75rem; border: 1px solid #f1c76b; color: #8a5c17; background: #fff8df; font-size: .8rem; }
.cafe-layout-stage { position: relative; width: 100%; overflow: hidden; aspect-ratio: 16 / 9; border: 1px solid #385369; background: #07111e; touch-action: none; user-select: none; }
.cafe-layout-stage::after { position: absolute; z-index: 1000; inset: 0; border: 1px solid rgba(255, 255, 255, .08); background-image: linear-gradient(rgba(255, 255, 255, .055) 1px, transparent 1px), linear-gradient(90deg, rgba(255, 255, 255, .055) 1px, transparent 1px); background-size: calc(100% / 60) calc(100% / 34); content: ''; pointer-events: none; }
.cafe-layout-stage > img { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; image-rendering: pixelated; pointer-events: none; }
.cafe-layout-workstation { position: absolute; width: 8.125%; padding: 0; border: 0; color: #15212c; background: transparent; cursor: grab; transform: translate(-50%, -92%); touch-action: none; }
.cafe-layout-workstation:active { cursor: grabbing; }
.cafe-layout-workstation img { display: block; width: 100%; height: auto; image-rendering: pixelated; pointer-events: none; }
.cafe-layout-workstation span { position: absolute; right: -.2rem; bottom: -.15rem; display: grid; width: 1.35rem; height: 1.35rem; border: 1px solid #fff4cf; color: #172235; background: #f4c76c; box-shadow: 2px 2px 0 rgba(0, 0, 0, .45); place-items: center; font: 800 .68rem/1 monospace; }
.cafe-layout-workstation.selected { filter: drop-shadow(0 0 5px #f8cc75); outline: 2px solid #f8cc75; outline-offset: 3px; }
.cafe-layout-coordinate-panel { display: flex; flex-wrap: wrap; align-items: center; gap: .7rem 1rem; padding: .7rem .8rem; border: 1px solid #d9e0e6; background: #f8fafc; font-size: .78rem; }
.cafe-layout-coordinate-panel label { display: inline-flex; align-items: center; gap: .35rem; }
.cafe-layout-coordinate-panel input { width: 5.5rem; min-height: 2rem; border: 1px solid #cbd5e1; padding: .25rem .4rem; background: #fff; }
.cafe-layout-coordinate-panel > span { color: #64748b; }
:global(.dark) .cafe-layout-toolbar, :global(.dark) .cafe-layout-coordinate-panel > span { color: #a9bac8; }
:global(.dark) .cafe-layout-count-bar { border-color: #42566a; color: #a9bac8; background: #102438; }
:global(.dark) .cafe-layout-count-bar > label { color: #e7eff5; }
:global(.dark) .cafe-layout-count-stepper, :global(.dark) .cafe-layout-count-stepper button, :global(.dark) .cafe-layout-count-stepper input { border-color: #536b80; color: #f1f5f9; background: #0b1928; }
:global(.dark) .cafe-layout-coordinate-panel { border-color: #42566a; color: #e7eff5; background: #102438; }
:global(.dark) .cafe-layout-coordinate-panel input { border-color: #536b80; color: #f1f5f9; background: #0b1928; }
@media (max-width: 639px) { .cafe-layout-toolbar, .cafe-layout-count-bar, .cafe-layout-stage, .cafe-layout-coordinate-panel { display: none; }.cafe-layout-mobile-note { display: block; } }
</style>
