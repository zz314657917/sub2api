import { afterEach, describe, it, expect, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import { defineComponent, ref } from 'vue'
enableAutoUnmount(afterEach)
import {
  useStepUp,
  isStepUpRequired,
  isStepUpBlocked,
  isStepUpCancelled,
  stepUpBlockReason,
  StepUpCancelledError
} from '../useStepUp'

describe('useStepUp error classification', () => {
  it('detects STEP_UP_REQUIRED from code field', () => {
    expect(isStepUpRequired({ status: 403, code: 'STEP_UP_REQUIRED' })).toBe(true)
    expect(isStepUpRequired({ status: 403, reason: 'STEP_UP_REQUIRED' })).toBe(true)
    expect(isStepUpRequired({ status: 500, code: 'INTERNAL' })).toBe(false)
    expect(isStepUpRequired(null)).toBe(false)
  })

  it('detects blocked (non-retryable) step-up errors', () => {
    expect(isStepUpBlocked({ code: 'STEP_UP_TOTP_NOT_ENABLED' })).toBe(true)
    expect(isStepUpBlocked({ reason: 'STEP_UP_ADMIN_API_KEY_FORBIDDEN' })).toBe(true)
    expect(isStepUpBlocked({ code: 'STEP_UP_REQUIRED' })).toBe(false)
  })

  it('surfaces the block reason marker', () => {
    expect(stepUpBlockReason({ reason: 'STEP_UP_ADMIN_API_KEY_FORBIDDEN' })).toBe('STEP_UP_ADMIN_API_KEY_FORBIDDEN')
    expect(stepUpBlockReason({ code: 'OTHER' })).toBe('')
  })
})

describe('useStepUp concurrent prompts', () => {
  it.each([true, false])('settles every waiter when verification resolves to %s', async (verified) => {
    const stepUp = useStepUp()
    const firstDone = vi.fn()
    const secondDone = vi.fn()
    const first = stepUp.prompt()
    first.then(firstDone)
    const second = stepUp.prompt()
    expect(second).toBe(first)
    second.then(secondDone)
    if (verified) stepUp.onVerified()
    else stepUp.onCancel()
    await Promise.resolve()
    expect(firstDone).toHaveBeenCalledWith(verified)
    expect(secondDone).toHaveBeenCalledWith(verified)
    expect(stepUp.visible.value).toBe(false)
    const next = stepUp.prompt()
    expect(next).not.toBe(first)
    stepUp.onCancel()
    await expect(next).resolves.toBe(false)
  })
})

describe('v0215 mounted concurrent step-up actions', () => {
  it.each([true, false])('settles both mounted callers after verified=%s', async verified => {
    const calls = [0, 0]
    const Harness = defineComponent({
      setup() {
        const stepUp = useStepUp()
        const outcomes = ref<string[]>([])
        async function start() {
          await Promise.all([0, 1].map(async index => {
            try {
              const result = await stepUp.run(async () => {
                if (++calls[index] === 1) throw { code: 'STEP_UP_REQUIRED' }
                return `verified-${index}`
              })
              outcomes.value.push(result)
            } catch (error) {
              outcomes.value.push(error instanceof StepUpCancelledError ? `cancelled-${index}` : 'unexpected')
            }
          }))
        }
        return { start, visible: stepUp.visible, onVerified: stepUp.onVerified, onCancel: stepUp.onCancel, outcomes }
      },
      template: '<div><button class="start" @click="start">Start</button><div v-if="visible" class="prompt"><button class="verify" @click="onVerified">Verify</button><button class="cancel" @click="onCancel">Cancel</button></div><span class="outcomes">{{ outcomes.join(",") }}</span></div>',
    })
    const wrapper = mount(Harness)
    await wrapper.get('.start').trigger('click')
    await flushPromises()
    expect(wrapper.findAll('.prompt')).toHaveLength(1)
    await wrapper.get(verified ? '.verify' : '.cancel').trigger('click')
    await flushPromises()
    expect(wrapper.get('.outcomes').text()).toContain(verified ? 'verified-0' : 'cancelled-0')
    expect(wrapper.get('.outcomes').text()).toContain(verified ? 'verified-1' : 'cancelled-1')
    expect(wrapper.find('.prompt').exists()).toBe(false)
    expect(calls).toEqual(verified ? [2, 2] : [1, 1])
  })
})

describe('useStepUp.run', () => {
  it('returns the action result directly on success', async () => {
    const stepUp = useStepUp()
    const result = await stepUp.run(async () => 42)
    expect(result).toBe(42)
    expect(stepUp.visible.value).toBe(false)
  })

  it('re-throws non-step-up errors without prompting', async () => {
    const stepUp = useStepUp()
    const err = { status: 500, code: 'INTERNAL' }
    await expect(stepUp.run(async () => { throw err })).rejects.toBe(err)
    expect(stepUp.visible.value).toBe(false)
  })

  it('re-throws blocked errors without prompting', async () => {
    const stepUp = useStepUp()
    const err = { status: 403, code: 'STEP_UP_TOTP_NOT_ENABLED' }
    await expect(stepUp.run(async () => { throw err })).rejects.toBe(err)
    expect(stepUp.visible.value).toBe(false)
  })

  it('prompts on STEP_UP_REQUIRED and retries after verification', async () => {
    const stepUp = useStepUp()
    let calls = 0
    const action = async () => {
      calls++
      if (calls === 1) throw { status: 403, code: 'STEP_UP_REQUIRED' }
      return 'ok'
    }
    const promise = stepUp.run(action)
    // The dialog should now be open, awaiting verification (after the first rejection is handled).
    await vi.waitFor(() => expect(stepUp.visible.value).toBe(true))
    stepUp.onVerified()
    await expect(promise).resolves.toBe('ok')
    expect(calls).toBe(2)
    expect(stepUp.visible.value).toBe(false)
  })

  it('throws a cancellation sentinel (not the original error) if the user cancels the prompt', async () => {
    const stepUp = useStepUp()
    const err = { status: 403, code: 'STEP_UP_REQUIRED' }
    const promise = stepUp.run(async () => { throw err })
    await vi.waitFor(() => expect(stepUp.visible.value).toBe(true))
    stepUp.onCancel()
    await expect(promise).rejects.toBeInstanceOf(StepUpCancelledError)
    expect(stepUp.visible.value).toBe(false)
  })

  it('classifies the cancellation sentinel distinctly from step-up errors', () => {
    const cancelled = new StepUpCancelledError()
    expect(isStepUpCancelled(cancelled)).toBe(true)
    expect(isStepUpRequired(cancelled)).toBe(false)
    expect(isStepUpBlocked(cancelled)).toBe(false)
    expect(isStepUpCancelled({ status: 403, code: 'STEP_UP_REQUIRED' })).toBe(false)
  })
})
