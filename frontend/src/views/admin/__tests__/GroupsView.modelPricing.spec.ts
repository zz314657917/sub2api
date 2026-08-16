import { describe, expect, it } from 'vitest'

import {
  groupPricingFromAPI,
  groupPricingToAPI,
} from '../GroupsView.vue'

describe('group model pricing form', () => {
  it('converts API token prices to form values and clears token intervals in create payloads', () => {
    const form = groupPricingFromAPI([{
      platform: 'openai',
      models: ['gpt-5'],
      billing_mode: 'token',
      input_price: 0.0000025,
      output_price: 0.00001,
      cache_write_price: null,
      cache_read_price: null,
      image_input_price: null,
      image_output_price: null,
      per_request_price: null,
      intervals: [{
        min_tokens: 0,
        max_tokens: 200000,
        tier_label: '',
        input_price: 0.000004,
        output_price: null,
        cache_write_price: null,
        cache_read_price: null,
        per_request_price: null,
        sort_order: 0,
      }],
    }])

    expect(form[0].input_price).toBe(2.5)
    expect(form[0].output_price).toBe(10)
    expect(form[0].intervals).toHaveLength(1)

    expect(groupPricingToAPI(form, 'openai')).toEqual([expect.objectContaining({
      platform: 'openai',
      models: ['gpt-5'],
      billing_mode: 'token',
      input_price: 0.0000025,
      output_price: 0.00001,
      intervals: [],
    })])
  })

  it('uses the same typed pricing payload for edit and excludes empty model entries', () => {
    const payload = groupPricingToAPI([
      {
        models: ['  gpt-image-2  '],
        billing_mode: 'image',
        input_price: null,
        output_price: null,
        cache_write_price: null,
        cache_read_price: null,
        image_input_price: null,
        image_output_price: null,
        per_request_price: '0.12',
        intervals: [],
      },
      {
        models: [' '],
        billing_mode: 'token',
        input_price: null,
        output_price: null,
        cache_write_price: null,
        cache_read_price: null,
        image_input_price: null,
        image_output_price: null,
        per_request_price: null,
        intervals: [],
      },
    ], 'openai')

    expect(payload).toEqual([expect.objectContaining({
      platform: 'openai',
      models: ['gpt-image-2'],
      billing_mode: 'image',
      per_request_price: 0.12,
    })])
  })
})
