import { describe, expect, it } from 'vitest'
import {
  hasImageInputCost,
  hasImageInputTokens,
  hasImageOutputCost,
  hasImageOutputTokens,
  textInputTokens,
  textOutputTokens,
} from '@/utils/imageUsage'

describe('imageUsage token helpers', () => {
  it('splits image tokens from total token counters', () => {
    expect(textInputTokens({ input_tokens: 371, image_input_tokens: 352 })).toBe(19)
    expect(textOutputTokens({ output_tokens: 439, image_output_tokens: 400 })).toBe(39)
    expect(hasImageInputTokens({ input_tokens: 371, image_input_tokens: 352 })).toBe(true)
    expect(hasImageOutputTokens({ output_tokens: 439, image_output_tokens: 400 })).toBe(true)
  })

  it('clamps malformed image token counters and detects costs', () => {
    expect(textInputTokens({ input_tokens: 10, image_input_tokens: 20 })).toBe(0)
    expect(textOutputTokens({ output_tokens: 10, image_output_tokens: 20 })).toBe(0)
    expect(hasImageInputCost({ image_input_cost: 0.002816 })).toBe(true)
    expect(hasImageOutputCost({ image_output_cost: 0.012 })).toBe(true)
    expect(hasImageInputCost({ image_input_cost: 0 })).toBe(false)
    expect(hasImageOutputCost({ image_output_cost: 0 })).toBe(false)
  })
})
