import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import CafeWorkstationLayoutEditor from '../CafeWorkstationLayoutEditor.vue'
import {
  CAFE_SCENE_MAX_WORKSTATION_COUNT,
  CAFE_SCENE_WORKSTATIONS,
} from '@/features/pixelCafe/renderer/sceneLayout'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string, params?: Record<string, number>) => `${key}${params?.id ?? ''}` }) }
})

function mountEditor() {
  return mount(CafeWorkstationLayoutEditor, {
    props: { modelValue: CAFE_SCENE_WORKSTATIONS.map(slot => ({ ...slot })) },
  })
}

describe('CafeWorkstationLayoutEditor', () => {
  it('renders ten numbered workstations on the shared 16:9 design space', () => {
    const wrapper = mountEditor()
    const workstations = wrapper.findAll('[data-testid="cafe-layout-workstation"]')
    expect(workstations).toHaveLength(10)
    expect(workstations[0].attributes('style')).toContain('left: 35.416')
    expect(workstations[0].attributes('style')).toContain('top: 46.296')
    expect(workstations[0].text()).toContain('1')
  })

  it('emits a snapped drag position and supports keyboard nudging', async () => {
    const wrapper = mountEditor()
    const stage = wrapper.get('[data-testid="cafe-layout-stage"]')
    vi.spyOn(stage.element, 'getBoundingClientRect').mockReturnValue({
      x: 0, y: 0, left: 0, top: 0, right: 960, bottom: 540, width: 960, height: 540, toJSON: () => ({}),
    })
    const first = wrapper.findAll('[data-testid="cafe-layout-workstation"]')[0]
    await first.trigger('pointerdown', { pointerId: 1, clientX: 340, clientY: 250 })
    await stage.trigger('pointermove', { pointerId: 1, clientX: 402, clientY: 301 })

    const dragLayout = wrapper.emitted('update:modelValue')?.[0]?.[0] as Array<{ id: number; x: number; y: number }>
    expect(dragLayout[0]).toEqual({ id: 1, x: 404, y: 300 })

    await first.trigger('keydown', { key: 'ArrowRight' })
    const keyboardLayout = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as Array<{ id: number; x: number; y: number }>
    expect(keyboardLayout[0]).toEqual({ id: 1, x: 344, y: 250 })
  })

  it('grows to fifty contiguous editable workstations and can shrink to one', async () => {
    const wrapper = mountEditor()
    const countInput = wrapper.get('[data-testid="cafe-layout-count-input"]')
    await countInput.setValue(CAFE_SCENE_MAX_WORKSTATION_COUNT)

    const grown = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as Array<{ id: number; x: number; y: number }>
    expect(grown).toHaveLength(50)
    expect(grown.slice(0, 10)).toEqual(CAFE_SCENE_WORKSTATIONS)
    expect(grown.map(slot => slot.id)).toEqual(Array.from({ length: 50 }, (_, index) => index + 1))
    expect(grown[49].x).toBeGreaterThanOrEqual(48)
    expect(grown[49].x).toBeLessThanOrEqual(912)
    expect(grown[49].y).toBeGreaterThanOrEqual(72)
    expect(grown[49].y).toBeLessThanOrEqual(520)

    await wrapper.setProps({ modelValue: grown })
    await countInput.setValue(1)
    const shrunk = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as Array<{ id: number; x: number; y: number }>
    expect(shrunk).toEqual([CAFE_SCENE_WORKSTATIONS[0]])
  })
})
